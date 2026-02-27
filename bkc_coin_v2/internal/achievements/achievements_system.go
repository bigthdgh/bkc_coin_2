package achievements

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// AchievementType represents different types of achievements
type AchievementType string

const (
	AchievementTypeFirstTap      AchievementType = "first_tap"
	AchievementTypeTapMaster     AchievementType = "tap_master"
	AchievementTypeEnergySaver   AchievementType = "energy_saver"
	AchievementTypeReferralKing  AchievementType = "referral_king"
	AchievementTypeNFTCollector  AchievementType = "nft_collector"
	AchievementTypeDailyStreak   AchievementType = "daily_streak"
	AchievementTypeWealthHolder  AchievementType = "wealth_holder"
	AchievementTypeSocialButterfly AchievementType = "social_butterfly"
	AchievementTypeGameChampion  AchievementType = "game_champion"
	AchievementTypeEarlyAdopter  AchievementType = "early_adopter"
)

// Achievement represents a single achievement
type Achievement struct {
	ID          int64           `json:"id" db:"id"`
	Type        AchievementType `json:"type" db:"type"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Icon        string          `json:"icon" db:"icon"`
	Reward      float64         `json:"reward" db:"reward"`
	XP          int             `json:"xp" db:"xp"`
	Requirement json.RawMessage `json:"requirement" db:"requirement"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// UserAchievement represents a user's achievement progress
type UserAchievement struct {
	ID            int64     `json:"id" db:"id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	AchievementID int64     `json:"achievement_id" db:"achievement_id"`
	Progress      float64   `json:"progress" db:"progress"`
	IsCompleted   bool      `json:"is_completed" db:"is_completed"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	RewardClaimed bool      `json:"reward_claimed" db:"reward_claimed"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// AchievementProgress represents achievement progress data
type AchievementProgress struct {
	Current float64            `json:"current"`
	Target  float64            `json:"target"`
	Data    map[string]float64 `json:"data,omitempty"`
}

// AchievementSystem manages all achievements
type AchievementSystem struct {
	db *sqlx.DB
}

// NewAchievementSystem creates a new achievement system
func NewAchievementSystem(db *sqlx.DB) *AchievementSystem {
	return &AchievementSystem{
		db: db,
	}
}

// InitializeDefaultAchievements creates default achievements
func (as *AchievementSystem) InitializeDefaultAchievements() error {
	defaultAchievements := []Achievement{
		{
			Type:        AchievementTypeFirstTap,
			Name:        "Первый шаг",
			Description: "Сделайте свой первый тап",
			Icon:        "🎯",
			Reward:      10.0,
			XP:          10,
			Requirement: json.RawMessage(`{"taps": 1}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeTapMaster,
			Name:        "Мастер тапов",
			Description: "Сделайте 1000 тапов",
			Icon:        "⚡",
			Reward:      100.0,
			XP:          100,
			Requirement: json.RawMessage(`{"taps": 1000}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeTapMaster,
			Name:        "Легенда тапов",
			Description: "Сделайте 10000 тапов",
			Icon:        "👑",
			Reward:      1000.0,
			XP:          1000,
			Requirement: json.RawMessage(`{"taps": 10000}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeEnergySaver,
			Name:        "Энергосберегатель",
			Description: "Используйте всю энергию за день 7 раз",
			Icon:        "🔋",
			Reward:      50.0,
			XP:          50,
			Requirement: json.RawMessage(`{"full_energy_days": 7}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeReferralKing,
			Name:        "Король рефералов",
			Description: "Пригласите 10 активных рефералов",
			Icon:        "👥",
			Reward:      500.0,
			XP:          500,
			Requirement: json.RawMessage(`{"active_referrals": 10}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeNFTCollector,
			Name:        "Коллекционер NFT",
			Description: "Купите 5 NFT",
			Icon:        "🎨",
			Reward:      200.0,
			XP:          200,
			Requirement: json.RawMessage(`{"nfts_owned": 5}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeDailyStreak,
			Name:        "Ежедневная активность",
			Description: "Входите в игру 7 дней подряд",
			Icon:        "📅",
			Reward:      100.0,
			XP:          100,
			Requirement: json.RawMessage(`{"daily_streak": 7}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeWealthHolder,
			Name:        "Состоятельный игрок",
			Description: "Накопите 10000 BKC",
			Icon:        "💰",
			Reward:      500.0,
			XP:          500,
			Requirement: json.RawMessage(`{"balance": 10000}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeSocialButterfly,
			Name:        "Социальная бабочка",
			Description: "Поделитесь игрой 10 раз",
			Icon:        "🦋",
			Reward:      50.0,
			XP:          50,
			Requirement: json.RawMessage(`{"shares": 10}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeGameChampion,
			Name:        "Чемпион игры",
			Description: "Достигните 50 уровня",
			Icon:        "🏆",
			Reward:      2000.0,
			XP:          2000,
			Requirement: json.RawMessage(`{"level": 50}`),
			IsActive:    true,
		},
		{
			Type:        AchievementTypeEarlyAdopter,
			Name:        "Ранний адепт",
			Description: "Зарегистрируйтесь в первый месяц",
			Icon:        "🌟",
			Reward:      100.0,
			XP:          100,
			Requirement: json.RawMessage(`{"early_adopter": true}`),
			IsActive:    true,
		},
	}

	for _, achievement := range defaultAchievements {
		err := as.createAchievement(&achievement)
		if err != nil {
			log.Printf("Error creating achievement %s: %v", achievement.Name, err)
		}
	}

	return nil
}

// createAchievement creates a new achievement
func (as *AchievementSystem) createAchievement(achievement *Achievement) error {
	query := `
		INSERT INTO achievements (type, name, description, icon, reward, xp, requirement, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (type, name) DO NOTHING
		RETURNING id
	`

	err := as.db.QueryRow(
		query,
		achievement.Type,
		achievement.Name,
		achievement.Description,
		achievement.Icon,
		achievement.Reward,
		achievement.XP,
		achievement.Requirement,
		achievement.IsActive,
		time.Now(),
	).Scan(&achievement.ID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to create achievement: %w", err)
	}

	return nil
}

// UpdateProgress updates user's achievement progress
func (as *AchievementSystem) UpdateProgress(ctx context.Context, userID int64, achievementType AchievementType, progress float64, data map[string]float64) error {
	// Get achievement by type
	achievement, err := as.getAchievementByType(achievementType)
	if err != nil {
		return fmt.Errorf("failed to get achievement: %w", err)
	}

	// Get or create user achievement
	userAchievement, err := as.getOrCreateUserAchievement(userID, achievement.ID)
	if err != nil {
		return fmt.Errorf("failed to get user achievement: %w", err)
	}

	// Update progress
	if progress > userAchievement.Progress {
		userAchievement.Progress = progress
		userAchievement.UpdatedAt = time.Now()

		// Check if achievement is completed
		requirement, err := as.parseRequirement(achievement.Requirement)
		if err != nil {
			return fmt.Errorf("failed to parse requirement: %w", err)
		}

		if progress >= requirement.Target {
			userAchievement.IsCompleted = true
			completedAt := time.Now()
			userAchievement.CompletedAt = &completedAt
		}

		// Save progress
		err = as.saveUserAchievement(userAchievement)
		if err != nil {
			return fmt.Errorf("failed to save user achievement: %w", err)
		}

		// If newly completed, trigger reward
		if userAchievement.IsCompleted && !userAchievement.RewardClaimed {
			return as.triggerAchievementReward(ctx, userID, userAchievement, achievement)
		}
	}

	return nil
}

// getAchievementByType retrieves achievement by type
func (as *AchievementSystem) getAchievementByType(achievementType AchievementType) (*Achievement, error) {
	var achievement Achievement
	query := `
		SELECT id, type, name, description, icon, reward, xp, requirement, is_active, created_at
		FROM achievements
		WHERE type = $1 AND is_active = true
		ORDER BY created_at ASC
		LIMIT 1
	`

	err := as.db.Get(&achievement, query, achievementType)
	if err != nil {
		return nil, err
	}

	return &achievement, nil
}

// getOrCreateUserAchievement gets or creates user achievement
func (as *AchievementSystem) getOrCreateUserAchievement(userID, achievementID int64) (*UserAchievement, error) {
	var userAchievement UserAchievement
	query := `
		SELECT id, user_id, achievement_id, progress, is_completed, completed_at, reward_claimed, created_at, updated_at
		FROM user_achievements
		WHERE user_id = $1 AND achievement_id = $2
	`

	err := as.db.Get(&userAchievement, query, userID, achievementID)
	if err == sql.ErrNoRows {
		// Create new user achievement
		userAchievement = UserAchievement{
			UserID:        userID,
			AchievementID: achievementID,
			Progress:      0,
			IsCompleted:   false,
			RewardClaimed: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		insertQuery := `
			INSERT INTO user_achievements (user_id, achievement_id, progress, is_completed, reward_claimed, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`

		err = as.db.QueryRow(
			insertQuery,
			userAchievement.UserID,
			userAchievement.AchievementID,
			userAchievement.Progress,
			userAchievement.IsCompleted,
			userAchievement.RewardClaimed,
			userAchievement.CreatedAt,
			userAchievement.UpdatedAt,
		).Scan(&userAchievement.ID)

		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &userAchievement, nil
}

// saveUserAchievement saves user achievement progress
func (as *AchievementSystem) saveUserAchievement(userAchievement *UserAchievement) error {
	query := `
		UPDATE user_achievements
		SET progress = $1, is_completed = $2, completed_at = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := as.db.Exec(
		query,
		userAchievement.Progress,
		userAchievement.IsCompleted,
		userAchievement.CompletedAt,
		userAchievement.UpdatedAt,
		userAchievement.ID,
	)

	return err
}

// parseRequirement parses achievement requirement
func (as *AchievementSystem) parseRequirement(requirement json.RawMessage) (*AchievementProgress, error) {
	var progress AchievementProgress
	err := json.Unmarshal(requirement, &progress)
	if err != nil {
		return nil, err
	}

	// Extract target from data if not set
	if progress.Target == 0 && len(progress.Data) > 0 {
		for _, value := range progress.Data {
			if value > progress.Target {
				progress.Target = value
			}
		}
	}

	return &progress, nil
}

// triggerAchievementReward triggers achievement reward
func (as *AchievementSystem) triggerAchievementReward(ctx context.Context, userID int64, userAchievement *UserAchievement, achievement *Achievement) error {
	// This would integrate with the economy system to give rewards
	// For now, we'll just mark as claimed
	userAchievement.RewardClaimed = true
	userAchievement.UpdatedAt = time.Now()

	err := as.saveUserAchievement(userAchievement)
	if err != nil {
		return fmt.Errorf("failed to mark reward as claimed: %w", err)
	}

	// Log achievement completion
	log.Printf("Achievement completed: User %d completed %s, reward: %.2f BKC, XP: %d",
		userID, achievement.Name, achievement.Reward, achievement.XP)

	// Here you would:
	// 1. Add BKC reward to user's balance
	// 2. Add XP to user's level
	// 3. Send notification
	// 4. Update statistics

	return nil
}

// GetUserAchievements gets all user achievements
func (as *AchievementSystem) GetUserAchievements(userID int64) ([]UserAchievementWithDetails, error) {
	query := `
		SELECT 
			ua.id, ua.user_id, ua.achievement_id, ua.progress, ua.is_completed, 
			ua.completed_at, ua.reward_claimed, ua.created_at, ua.updated_at,
			a.type, a.name, a.description, a.icon, a.reward, a.xp, a.requirement
		FROM user_achievements ua
		JOIN achievements a ON ua.achievement_id = a.id
		WHERE ua.user_id = $1
		ORDER BY ua.created_at DESC
	`

	rows, err := as.db.Queryx(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []UserAchievementWithDetails
	for rows.Next() {
		var achievement UserAchievementWithDetails
		err := rows.StructScan(&achievement)
		if err != nil {
			return nil, err
		}
		achievements = append(achievements, achievement)
	}

	return achievements, nil
}

// UserAchievementWithDetails includes achievement details
type UserAchievementWithDetails struct {
	ID            int64           `json:"id" db:"id"`
	UserID        int64           `json:"user_id" db:"user_id"`
	AchievementID int64           `json:"achievement_id" db:"achievement_id"`
	Progress      float64         `json:"progress" db:"progress"`
	IsCompleted   bool            `json:"is_completed" db:"is_completed"`
	CompletedAt   *time.Time      `json:"completed_at" db:"completed_at"`
	RewardClaimed bool            `json:"reward_claimed" db:"reward_claimed"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	Type          AchievementType `json:"type" db:"type"`
	Name          string          `json:"name" db:"name"`
	Description   string          `json:"description" db:"description"`
	Icon          string          `json:"icon" db:"icon"`
	Reward        float64         `json:"reward" db:"reward"`
	XP            int             `json:"xp" db:"xp"`
	Requirement   json.RawMessage `json:"requirement" db:"requirement"`
}

// GetAvailableAchievements gets all available achievements
func (as *AchievementSystem) GetAvailableAchievements() ([]Achievement, error) {
	query := `
		SELECT id, type, name, description, icon, reward, xp, requirement, is_active, created_at
		FROM achievements
		WHERE is_active = true
		ORDER BY created_at ASC
	`

	var achievements []Achievement
	err := as.db.Select(&achievements, query)
	if err != nil {
		return nil, err
	}

	return achievements, nil
}

// GetAchievementStats gets achievement statistics
func (as *AchievementSystem) GetAchievementStats(userID int64) (*AchievementStats, error) {
	stats := &AchievementStats{}

	// Total achievements
	err := as.db.Get(&stats.TotalAchievements, "SELECT COUNT(*) FROM achievements WHERE is_active = true")
	if err != nil {
		return nil, err
	}

	// Completed achievements
	err = as.db.Get(&stats.CompletedAchievements, 
		"SELECT COUNT(*) FROM user_achievements WHERE user_id = $1 AND is_completed = true", userID)
	if err != nil {
		return nil, err
	}

	// Total rewards earned
	err = as.db.Get(&stats.TotalRewards, 
		`SELECT COALESCE(SUM(a.reward), 0) 
		FROM user_achievements ua 
		JOIN achievements a ON ua.achievement_id = a.id 
		WHERE ua.user_id = $1 AND ua.is_completed = true`, userID)
	if err != nil {
		return nil, err
	}

	// Total XP earned
	err = as.db.Get(&stats.TotalXP, 
		`SELECT COALESCE(SUM(a.xp), 0) 
		FROM user_achievements ua 
		JOIN achievements a ON ua.achievement_id = a.id 
		WHERE ua.user_id = $1 AND ua.is_completed = true`, userID)
	if err != nil {
		return nil, err
	}

	// Completion percentage
	if stats.TotalAchievements > 0 {
		stats.CompletionPercentage = float64(stats.CompletedAchievements) / float64(stats.TotalAchievements) * 100
	}

	return stats, nil
}

// AchievementStats represents achievement statistics
type AchievementStats struct {
	TotalAchievements     int     `json:"total_achievements"`
	CompletedAchievements int     `json:"completed_achievements"`
	TotalRewards         float64 `json:"total_rewards"`
	TotalXP              int     `json:"total_xp"`
	CompletionPercentage  float64 `json:"completion_percentage"`
}

// CreateAchievementTables creates necessary database tables
func (as *AchievementSystem) CreateAchievementTables() error {
	// Create achievements table
	achievementsTable := `
		CREATE TABLE IF NOT EXISTS achievements (
			id SERIAL PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			icon VARCHAR(10),
			reward DECIMAL(20,8) DEFAULT 0,
			xp INTEGER DEFAULT 0,
			requirement JSONB,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(type, name)
		);
	`

	// Create user_achievements table
	userAchievementsTable := `
		CREATE TABLE IF NOT EXISTS user_achievements (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			achievement_id INTEGER NOT NULL REFERENCES achievements(id),
			progress DECIMAL(20,8) DEFAULT 0,
			is_completed BOOLEAN DEFAULT false,
			completed_at TIMESTAMP,
			reward_claimed BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, achievement_id)
		);
	`

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_user_achievements_user_id ON user_achievements(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_achievements_completed ON user_achievements(is_completed);",
		"CREATE INDEX IF NOT EXISTS idx_achievements_type ON achievements(type);",
		"CREATE INDEX IF NOT EXISTS idx_achievements_active ON achievements(is_active);",
	}

	// Execute table creation
	_, err := as.db.Exec(achievementsTable)
	if err != nil {
		return fmt.Errorf("failed to create achievements table: %w", err)
	}

	_, err = as.db.Exec(userAchievementsTable)
	if err != nil {
		return fmt.Errorf("failed to create user_achievements table: %w", err)
	}

	// Create indexes
	for _, index := range indexes {
		_, err = as.db.Exec(index)
		if err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}

	return nil
}
