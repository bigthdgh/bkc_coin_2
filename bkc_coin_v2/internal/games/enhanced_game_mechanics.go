package games

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// GameMechanicType represents different game mechanics
type GameMechanicType string

const (
	MechanicTypeDailyBonus  GameMechanicType = "daily_bonus"
	MechanicTypeLuckyTap    GameMechanicType = "lucky_tap"
	MechanicTypeCombo       GameMechanicType = "combo"
	MechanicTypeBonusRound  GameMechanicType = "bonus_round"
	MechanicTypeTournament  GameMechanicType = "tournament"
	MechanicTypeLeaderboard GameMechanicType = "leaderboard"
	MechanicTypeSeasonal    GameMechanicType = "seasonal"
	MechanicTypeAchievement GameMechanicType = "achievement"
)

// GameState represents the current game state
type GameState struct {
	UserID           int64     `json:"user_id"`
	CurrentLevel     int       `json:"current_level"`
	Experience       int64     `json:"experience"`
	TotalTaps        int64     `json:"total_taps"`
	CurrentCombo     int       `json:"current_combo"`
	MaxCombo         int       `json:"max_combo"`
	DailyStreak      int       `json:"daily_streak"`
	LastActive       time.Time `json:"last_active"`
	BonusMultiplier  float64   `json:"bonus_multiplier"`
	UnlockedFeatures []string  `json:"unlocked_features"`
}

// DailyBonus represents daily bonus configuration
type DailyBonus struct {
	Day       int       `json:"day"`
	Reward    float64   `json:"reward"`
	Required  int       `json:"required"`
	Claimed   bool      `json:"claimed"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Combo represents a combo configuration
type Combo struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	RequiredTaps int     `json:"required_taps"`
	Reward       float64 `json:"reward"`
	Multiplier   float64 `json:"multiplier"`
	Description  string  `json:"description"`
}

// Tournament represents a tournament
type Tournament struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	PrizePool      float64   `json:"prize_pool"`
	EntryFee       float64   `json:"entry_fee"`
	MaxPlayers     int       `json:"max_players"`
	CurrentPlayers int       `json:"current_players"`
	Status         string    `json:"status"`
	Rules          string    `json:"rules"`
}

// EnhancedGameMechanics handles advanced game mechanics
type EnhancedGameMechanics struct {
	db          *sqlx.DB
	redisClient *redis.Client
}

// NewEnhancedGameMechanics creates a new enhanced game mechanics system
func NewEnhancedGameMechanics(db *sqlx.DB, redisClient *redis.Client) *EnhancedGameMechanics {
	return &EnhancedGameMechanics{
		db:          db,
		redisClient: redisClient,
	}
}

// ProcessLuckyTap processes a lucky tap event
func (egm *EnhancedGameMechanics) ProcessLuckyTap(ctx context.Context, userID int64, baseReward float64) (float64, bool, error) {
	// 1% chance for lucky tap
	if rand.Float64() > 0.01 {
		return baseReward, false, nil
	}

	// Lucky tap multiplier (2x to 10x)
	multiplier := 2.0 + rand.Float64()*8.0
	luckyReward := baseReward * multiplier

	// Store lucky tap event
	event := map[string]interface{}{
		"user_id":      userID,
		"type":         "lucky_tap",
		"base_reward":  baseReward,
		"multiplier":   multiplier,
		"final_reward": luckyReward,
		"timestamp":    time.Now(),
	}

	eventJSON, _ := json.Marshal(event)
	key := fmt.Sprintf("lucky_taps:%d", userID)
	err := egm.redisClient.LPush(ctx, key, eventJSON).Err()
	if err != nil {
		log.Printf("Failed to store lucky tap: %v", err)
	}

	// Keep only last 100 lucky taps
	egm.redisClient.LTrim(ctx, key, 0, 99)
	egm.redisClient.Expire(ctx, key, 30*24*time.Hour)

	return luckyReward, true, nil
}

// ProcessCombo processes combo mechanics
func (egm *EnhancedGameMechanics) ProcessCombo(ctx context.Context, userID int64, tapCount int) (*Combo, error) {
	// Get current combo state
	comboKey := fmt.Sprintf("combo:%d", userID)
	currentCombo, err := egm.redisClient.Get(ctx, comboKey).Int()
	if err != nil && err.Error() != "redis: nil" {
		return nil, fmt.Errorf("failed to get combo: %w", err)
	}

	// Increment combo
	newCombo := currentCombo + 1

	// Check for combo milestones
	combo := egm.getComboByTaps(newCombo)
	if combo != nil {
		// Store combo achievement
		comboData := map[string]interface{}{
			"user_id":   userID,
			"combo_id":  combo.ID,
			"taps":      newCombo,
			"reward":    combo.Reward,
			"timestamp": time.Now(),
		}

		comboJSON, _ := json.Marshal(comboData)
		achievementKey := fmt.Sprintf("combo_achievements:%d", userID)
		egm.redisClient.LPush(ctx, achievementKey, comboJSON)
		egm.redisClient.Expire(ctx, achievementKey, 30*24*time.Hour)
	}

	// Update combo counter with expiration (5 seconds without tap)
	egm.redisClient.Set(ctx, comboKey, newCombo, 5*time.Second)

	return combo, nil
}

// getComboByTaps returns combo configuration based on tap count
func (egm *EnhancedGameMechanics) getComboByTaps(taps int) *Combo {
	combos := []Combo{
		{ID: 1, Name: "Double Tap", RequiredTaps: 2, Reward: 0.5, Multiplier: 1.2, Description: "2 taps in a row"},
		{ID: 2, Name: "Triple Tap", RequiredTaps: 3, Reward: 1.0, Multiplier: 1.5, Description: "3 taps in a row"},
		{ID: 3, Name: "Speed Demon", RequiredTaps: 5, Reward: 2.0, Multiplier: 2.0, Description: "5 taps in a row"},
		{ID: 4, Name: "Tap Master", RequiredTaps: 10, Reward: 5.0, Multiplier: 2.5, Description: "10 taps in a row"},
		{ID: 5, Name: "Legendary Tapper", RequiredTaps: 25, Reward: 15.0, Multiplier: 3.0, Description: "25 taps in a row"},
		{ID: 6, Name: "God Mode", RequiredTaps: 50, Reward: 50.0, Multiplier: 5.0, Description: "50 taps in a row"},
	}

	for _, combo := range combos {
		if taps == combo.RequiredTaps {
			return &combo
		}
	}

	return nil
}

// ProcessDailyBonus processes daily bonus
func (egm *EnhancedGameMechanics) ProcessDailyBonus(ctx context.Context, userID int64) (*DailyBonus, error) {
	// Get user's daily bonus progress
	bonusKey := fmt.Sprintf("daily_bonus:%d", userID)

	// Check if already claimed
	claimed, err := egm.redisClient.Get(ctx, bonusKey+":claimed").Bool()
	if err == nil && claimed {
		return nil, fmt.Errorf("daily bonus already claimed")
	}

	// Calculate daily bonus based on streak
	streak, _ := egm.redisClient.Get(ctx, fmt.Sprintf("streak:%d", userID)).Int()

	day := (streak % 30) + 1
	bonus := egm.calculateDailyBonus(day)

	// Create daily bonus object
	dailyBonus := &DailyBonus{
		Day:       day,
		Reward:    bonus,
		Required:  10, // Minimum 10 taps
		Claimed:   false,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return dailyBonus, nil
}

// calculateDailyBonus calculates bonus based on day
func (egm *EnhancedGameMechanics) calculateDailyBonus(day int) float64 {
	baseBonus := 10.0

	// Bonus increases with streak
	if day <= 7 {
		return baseBonus + float64(day)*2
	} else if day <= 14 {
		return baseBonus + 20 + float64(day-7)*3
	} else if day <= 21 {
		return baseBonus + 50 + float64(day-14)*4
	} else {
		return baseBonus + 100 + float64(day-21)*5
	}
}

// ClaimDailyBonus claims daily bonus for user
func (egm *EnhancedGameMechanics) ClaimDailyBonus(ctx context.Context, userID int64) error {
	bonusKey := fmt.Sprintf("daily_bonus:%d", userID)

	// Mark as claimed
	err := egm.redisClient.Set(ctx, bonusKey+":claimed", true, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to claim daily bonus: %w", err)
	}

	// Get bonus amount
	bonus, err := egm.ProcessDailyBonus(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get daily bonus: %w", err)
	}

	// Log bonus claim
	claimEvent := map[string]interface{}{
		"user_id":   userID,
		"type":      "daily_bonus_claim",
		"bonus":     bonus.Reward,
		"day":       bonus.Day,
		"timestamp": time.Now(),
	}

	eventJSON, _ := json.Marshal(claimEvent)
	logKey := fmt.Sprintf("bonus_claims:%d", userID)
	egm.redisClient.LPush(ctx, logKey, eventJSON)
	egm.redisClient.Expire(ctx, logKey, 30*24*time.Hour)

	return nil
}

// CreateTournament creates a new tournament
func (egm *EnhancedGameMechanics) CreateTournament(ctx context.Context, name string, startTime, endTime time.Time, entryFee float64, prizePool float64) (*Tournament, error) {
	tournament := &Tournament{
		ID:             time.Now().UnixNano(),
		Name:           name,
		StartTime:      startTime,
		EndTime:        endTime,
		PrizePool:      prizePool,
		EntryFee:       entryFee,
		MaxPlayers:     1000,
		CurrentPlayers: 0,
		Status:         "upcoming",
		Rules:          "Most taps wins!",
	}

	// Store tournament
	tournamentJSON, _ := json.Marshal(tournament)
	tournamentKey := fmt.Sprintf("tournament:%d", tournament.ID)
	err := egm.redisClient.Set(ctx, tournamentKey, tournamentJSON, 7*24*time.Hour).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to create tournament: %w", err)
	}

	// Add to tournaments list
	egm.redisClient.LPush(ctx, "tournaments", tournamentJSON)

	return tournament, nil
}

// JoinTournament allows user to join a tournament
func (egm *EnhancedGameMechanics) JoinTournament(ctx context.Context, userID int64, tournamentID int64) error {
	tournamentKey := fmt.Sprintf("tournament:%d", tournamentID)

	// Get tournament
	tournamentJSON, err := egm.redisClient.Get(ctx, tournamentKey).Result()
	if err != nil {
		return fmt.Errorf("tournament not found: %w", err)
	}

	var tournament Tournament
	err = json.Unmarshal([]byte(tournamentJSON), &tournament)
	if err != nil {
		return fmt.Errorf("failed to parse tournament: %w", err)
	}

	// Check if tournament is active
	if tournament.Status != "active" {
		return fmt.Errorf("tournament is not active")
	}

	// Check if user already joined
	playersKey := fmt.Sprintf("tournament_players:%d", tournamentID)
	isJoined, err := egm.redisClient.SIsMember(ctx, playersKey, userID).Result()
	if err != nil {
		return fmt.Errorf("failed to check tournament membership: %w", err)
	}

	if isJoined {
		return fmt.Errorf("user already joined tournament")
	}

	// Add user to tournament
	egm.redisClient.SAdd(ctx, playersKey, userID)
	egm.redisClient.Expire(ctx, playersKey, tournament.EndTime.Sub(time.Now()))

	// Update player count
	tournament.CurrentPlayers++
	updatedJSON, _ := json.Marshal(tournament)
	egm.redisClient.Set(ctx, tournamentKey, updatedJSON, 7*24*time.Hour)

	return nil
}

// GetLeaderboard returns tournament leaderboard
func (egm *EnhancedGameMechanics) GetLeaderboard(ctx context.Context, tournamentID int64, limit int) ([]map[string]interface{}, error) {
	leaderboardKey := fmt.Sprintf("tournament_leaderboard:%d", tournamentID)

	// Get top players
	leaders, err := egm.redisClient.ZRevRangeWithScores(ctx, leaderboardKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	var leaderboard []map[string]interface{}
	for _, leader := range leaders {
		var entry map[string]interface{}
		err := json.Unmarshal([]byte(leader.Member.(string)), &entry)
		if err != nil {
			continue
		}
		entry["score"] = leader.Score
		leaderboard = append(leaderboard, entry)
	}

	return leaderboard, nil
}

// UpdateTournamentScore updates user's score in tournament
func (egm *EnhancedGameMechanics) UpdateTournamentScore(ctx context.Context, userID int64, tournamentID int64, score float64) error {
	leaderboardKey := fmt.Sprintf("tournament_leaderboard:%d", tournamentID)

	// Update user score
	userData := map[string]interface{}{
		"user_id":   userID,
		"score":     score,
		"timestamp": time.Now(),
	}

	userJSON, _ := json.Marshal(userData)

	// Add to sorted set (leaderboard)
	err := egm.redisClient.ZAdd(ctx, leaderboardKey, redis.Z{
		Score:  score,
		Member: userJSON,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to update tournament score: %w", err)
	}

	// Set expiration
	egm.redisClient.Expire(ctx, leaderboardKey, 24*time.Hour)

	return nil
}

// ProcessBonusRound processes a bonus round
func (egm *EnhancedGameMechanics) ProcessBonusRound(ctx context.Context, userID int64) (float64, error) {
	// Check if user is eligible for bonus round
	eligibilityKey := fmt.Sprintf("bonus_round_eligibility:%d", userID)

	eligible, err := egm.redisClient.Get(ctx, eligibilityKey).Bool()
	if err != nil || !eligible {
		return 0, fmt.Errorf("user not eligible for bonus round")
	}

	// Random bonus (10x to 100x base reward)
	multiplier := 10.0 + rand.Float64()*90.0
	bonusReward := 1.0 * multiplier // Base reward of 1 BKC

	// Mark bonus round as used
	egm.redisClient.Del(ctx, eligibilityKey)

	// Log bonus round
	bonusEvent := map[string]interface{}{
		"user_id":    userID,
		"type":       "bonus_round",
		"multiplier": multiplier,
		"reward":     bonusReward,
		"timestamp":  time.Now(),
	}

	eventJSON, _ := json.Marshal(bonusEvent)
	bonusKey := fmt.Sprintf("bonus_rounds:%d", userID)
	egm.redisClient.LPush(ctx, bonusKey, eventJSON)
	egm.redisClient.Expire(ctx, bonusKey, 30*24*time.Hour)

	return bonusReward, nil
}

// CheckBonusRoundEligibility checks if user is eligible for bonus round
func (egm *EnhancedGameMechanics) CheckBonusRoundEligibility(ctx context.Context, userID int64, totalTaps int64) bool {
	// Eligible after every 1000 taps
	if totalTaps%1000 != 0 {
		return false
	}

	// Check cooldown (24 hours)
	eligibilityKey := fmt.Sprintf("bonus_round_eligibility:%d", userID)

	_, err := egm.redisClient.Get(ctx, eligibilityKey).Result()
	if err == nil {
		return false // Already used
	}

	// Mark as eligible
	egm.redisClient.Set(ctx, eligibilityKey, true, 24*time.Hour)
	return true
}

// GetGameState returns user's current game state
func (egm *EnhancedGameMechanics) GetGameState(ctx context.Context, userID int64) (*GameState, error) {
	stateKey := fmt.Sprintf("game_state:%d", userID)

	stateJSON, err := egm.redisClient.Get(ctx, stateKey).Result()
	if err != nil {
		// Create initial state
		state := &GameState{
			UserID:           userID,
			CurrentLevel:     1,
			Experience:       0,
			TotalTaps:        0,
			CurrentCombo:     0,
			MaxCombo:         0,
			DailyStreak:      0,
			LastActive:       time.Now(),
			BonusMultiplier:  1.0,
			UnlockedFeatures: []string{},
		}

		egm.saveGameState(ctx, state)
		return state, nil
	}

	var state GameState
	err = json.Unmarshal([]byte(stateJSON), &state)
	if err != nil {
		return nil, fmt.Errorf("failed to parse game state: %w", err)
	}

	return &state, nil
}

// saveGameState saves user's game state
func (egm *EnhancedGameMechanics) saveGameState(ctx context.Context, state *GameState) error {
	stateKey := fmt.Sprintf("game_state:%d", state.UserID)

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}

	err = egm.redisClient.Set(ctx, stateKey, stateJSON, 30*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to save game state: %w", err)
	}

	return nil
}

// UpdateGameState updates user's game state
func (egm *EnhancedGameMechanics) UpdateGameState(ctx context.Context, userID int64, updates map[string]interface{}) error {
	state, err := egm.GetGameState(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "experience":
			if exp, ok := value.(int64); ok {
				state.Experience += exp
				// Check for level up
				newLevel := egm.calculateLevel(state.Experience)
				if newLevel > state.CurrentLevel {
					state.CurrentLevel = newLevel
					// Unlock new features
					egm.unlockFeatures(state, newLevel)
				}
			}
		case "total_taps":
			if taps, ok := value.(int64); ok {
				state.TotalTaps += taps
			}
		case "combo":
			if combo, ok := value.(int); ok {
				state.CurrentCombo = combo
				if combo > state.MaxCombo {
					state.MaxCombo = combo
				}
			}
		case "daily_streak":
			if streak, ok := value.(int); ok {
				state.DailyStreak = streak
			}
		}
	}

	state.LastActive = time.Now()
	return egm.saveGameState(ctx, state)
}

// calculateLevel calculates level based on experience
func (egm *EnhancedGameMechanics) calculateLevel(experience int64) int {
	// Level formula: 100 * level^2 experience required
	level := 1
	for {
		requiredExp := int64(100 * level * level)
		if experience < requiredExp {
			break
		}
		level++
	}
	return level - 1
}

// unlockFeatures unlocks features based on level
func (egm *EnhancedGameMechanics) unlockFeatures(state *GameState, level int) {
	features := map[int][]string{
		5:  {"lucky_tap"},
		10: {"combo_multiplier"},
		15: {"daily_bonus_plus"},
		20: {"tournament_access"},
		25: {"bonus_round"},
		30: {"premium_features"},
		50: {"god_mode"},
	}

	if newFeatures, exists := features[level]; exists {
		for _, feature := range newFeatures {
			// Check if already unlocked
			found := false
			for _, unlocked := range state.UnlockedFeatures {
				if unlocked == feature {
					found = true
					break
				}
			}
			if !found {
				state.UnlockedFeatures = append(state.UnlockedFeatures, feature)
			}
		}
	}
}

// GetActiveTournaments returns active tournaments
func (egm *EnhancedGameMechanics) GetActiveTournaments(ctx context.Context) ([]Tournament, error) {
	tournamentsJSON, err := egm.redisClient.LRange(ctx, "tournaments", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get tournaments: %w", err)
	}

	var tournaments []Tournament
	for _, tournamentJSON := range tournamentsJSON {
		var tournament Tournament
		err := json.Unmarshal([]byte(tournamentJSON), &tournament)
		if err != nil {
			continue
		}

		// Only include active tournaments
		if tournament.Status == "active" {
			tournaments = append(tournaments, tournament)
		}
	}

	return tournaments, nil
}

// CreateGameMechanicsTables creates necessary database tables
func (egm *EnhancedGameMechanics) CreateGameMechanicsTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tournaments (
			id BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			prize_pool DECIMAL(20,8) NOT NULL,
			entry_fee DECIMAL(20,8) NOT NULL,
			max_players INTEGER DEFAULT 1000,
			current_players INTEGER DEFAULT 0,
			status VARCHAR(50) DEFAULT 'upcoming',
			rules TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tournament_participants (
			id SERIAL PRIMARY KEY,
			tournament_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			score DECIMAL(20,8) DEFAULT 0,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tournament_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS game_sessions (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			session_start TIMESTAMP NOT NULL,
			session_end TIMESTAMP,
			taps INTEGER DEFAULT 0,
			experience_earned INTEGER DEFAULT 0,
			bonus_earned DECIMAL(20,8) DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		_, err := egm.db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create game mechanics table: %w", err)
		}
	}

	return nil
}
