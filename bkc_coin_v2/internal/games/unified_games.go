package games

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	mathrand "math/rand"
	"time"

	"bkc_coin_v2/internal/database"
)

// UnifiedGamesManager - унифицированный менеджер игр
type UnifiedGamesManager struct {
	db *database.UnifiedDB
}

// NewUnifiedGamesManager - создание нового менеджера игр
func NewUnifiedGamesManager(database *database.UnifiedDB) *UnifiedGamesManager {
	return &UnifiedGamesManager{db: database}
}

// UnifiedCrashGame - игра "Ракетка"
type UnifiedCrashGame struct {
	ID           int64      `json:"id"`
	GameID       string     `json:"game_id"`
	Hash         string     `json:"hash"`
	Salt         string     `json:"salt"`
	CrashPoint   float64    `json:"crash_point"`
	Status       string     `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	CrashedAt    *time.Time `json:"crashed_at"`
	TotalBets    int64      `json:"total_bets"`
	TotalWinners int64      `json:"total_winners"`
	SystemProfit int64      `json:"system_profit"`
}

// UnifiedCrashBet - ставка в игре Ракетка
type UnifiedCrashBet struct {
	BetID       string    `json:"bet_id"`
	GameID      string    `json:"game_id"`
	UserID      int64     `json:"user_id"`
	Amount      int64     `json:"amount"`
	AutoCashout float64   `json:"auto_cashout"`
	Status      string    `json:"status"`
	CashedOutAt float64   `json:"cashed_out_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// UnifiedExchangePrice - цена обмена
type UnifiedExchangePrice struct {
	Pair      string    `json:"pair"`
	Price     float64   `json:"price"`
	Change24h float64   `json:"change_24h"`
	CreatedAt time.Time `json:"created_at"`
}

// GameResult - результат игры
type GameResult struct {
	Success    bool      `json:"success"`
	GameID     string    `json:"game_id"`
	PlayerID   int64     `json:"player_id"`
	Amount     int64     `json:"amount"`
	Winnings   int64     `json:"winnings"`
	Multiplier float64   `json:"multiplier"`
	PlayedAt   time.Time `json:"played_at"`
}

// GenerateUnifiedCrashHash - генерация хеша для игры Ракетка
func GenerateUnifiedCrashHash() (hash, salt string, crashPoint float64, err error) {
	// Генерируем случайную соль
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", "", 0, fmt.Errorf("failed to generate salt: %w", err)
	}
	salt = hex.EncodeToString(saltBytes)

	// Генерируем случайную точку краха (1.01 - 10.00)
	crashPoint = 1.01 + mathrand.Float64()*8.99

	// Создаем хеш
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s-%.2f-%d", salt, crashPoint, time.Now().UnixNano())))
	hash = hex.EncodeToString(h.Sum(nil))

	return hash, salt, crashPoint, nil
}

// StartUnifiedCrashGame - начало игры Ракетка
func (ugm *UnifiedGamesManager) StartUnifiedCrashGame(ctx context.Context) (*UnifiedCrashGame, error) {
	hash, salt, crashPoint, err := GenerateUnifiedCrashHash()
	if err != nil {
		return nil, fmt.Errorf("failed to generate crash hash: %w", err)
	}

	gameID := fmt.Sprintf("crash_%d", time.Now().Unix())
	game := &UnifiedCrashGame{
		GameID:     gameID,
		Hash:       hash,
		Salt:       salt,
		CrashPoint: crashPoint,
		Status:     "waiting",
		StartedAt:  time.Now(),
	}

	// Сохраняем игру в базу данных
	err = ugm.saveUnifiedCrashGame(ctx, game)
	if err != nil {
		return nil, fmt.Errorf("failed to save crash game: %w", err)
	}

	log.Printf("Crash game %s started with crash point %.2f", gameID, crashPoint)
	return game, nil
}

// saveUnifiedCrashGame - сохранение игры в базу данных
func (ugm *UnifiedGamesManager) saveUnifiedCrashGame(ctx context.Context, game *UnifiedCrashGame) error {
	// Здесь должна быть реальная запись в базу данных
	// Для примера просто логируем
	log.Printf("Saving crash game: %+v", game)
	return nil
}

// PlaceUnifiedCrashBet - размещение ставки в игре Ракетка
func (ugm *UnifiedGamesManager) PlaceUnifiedCrashBet(ctx context.Context, userID int64, gameID string, amount int64, autoCashout float64) (*UnifiedCrashBet, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("bet amount must be positive")
	}

	if autoCashout < 1.01 || autoCashout > 10.00 {
		return nil, fmt.Errorf("auto cashout must be between 1.01 and 10.00")
	}

	// Проверяем баланс пользователя
	userState, err := ugm.db.GetUserState(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user state: %w", err)
	}

	if userState.Balance < amount {
		return nil, fmt.Errorf("insufficient balance: need %d, have %d", amount, userState.Balance)
	}

	// Создаем ставку
	betID := fmt.Sprintf("bet_%d_%d", userID, time.Now().Unix())
	bet := &CrashBet{
		BetID:       betID,
		GameID:      gameID,
		UserID:      userID,
		Amount:      amount,
		AutoCashout: autoCashout,
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	// Списываем монеты
	err = ugm.db.UpdateUserBalance(ctx, userID, -amount)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct balance: %w", err)
	}

	// Сохраняем ставку
	err = ugm.saveCrashBet(ctx, bet)
	if err != nil {
		// Возвращаем монеты при ошибке
		ugm.db.UpdateUserBalance(ctx, userID, amount)
		return nil, fmt.Errorf("failed to save bet: %w", err)
	}

	log.Printf("Bet placed: user %d, amount %d, auto cashout %.2f", userID, amount, autoCashout)
	return bet, nil
}

// saveCrashBet - сохранение ставки в базу данных
func (ugm *UnifiedGamesManager) saveCrashBet(ctx context.Context, bet *CrashBet) error {
	// Здесь должна быть реальная запись в базу данных
	log.Printf("Saving crash bet: %+v", bet)
	return nil
}

// CashoutUnifiedBet - вывод ставки из игры
func (ugm *UnifiedGamesManager) CashoutUnifiedBet(ctx context.Context, betID string, currentMultiplier float64) (*GameResult, error) {
	if currentMultiplier < 1.01 {
		return nil, fmt.Errorf("multiplier too low for cashout")
	}

	// Получаем информацию о ставке (в реальном приложении из БД)
	bet := &UnifiedCrashBet{
		BetID:       betID,
		UserID:      12345, // Заглушка
		Amount:      1000,
		AutoCashout: 2.0,
		Status:      "active",
	}

	// Рассчитываем выигрыш
	winAmount := int64(float64(bet.Amount) * currentMultiplier)

	// Обновляем баланс пользователя
	err := ugm.db.UpdateUserBalance(ctx, bet.UserID, winAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to credit winnings: %w", err)
	}

	// Создаем результат
	result := &GameResult{
		Success:    true,
		GameID:     bet.GameID,
		PlayerID:   bet.UserID,
		Amount:     bet.Amount,
		Winnings:   winAmount - bet.Amount,
		Multiplier: currentMultiplier,
		PlayedAt:   time.Now(),
	}

	log.Printf("Bet cashed out: bet %s, multiplier %.2f, winnings %d", betID, currentMultiplier, winAmount-bet.Amount)
	return result, nil
}

// CrashGame - завершение игры Ракетка
func (ugm *UnifiedGamesManager) CrashGame(ctx context.Context, gameID string) error {
	// Получаем информацию об игре (в реальном приложении из БД)
	game := &CrashGame{
		GameID:     gameID,
		CrashPoint: 2.5,
		Status:     "active",
		TotalBets:  5000,
	}

	// Рассчитаем системную прибыль
	systemProfit := int64(float64(game.TotalBets) * 0.05) // 5% комиссия

	game.Status = "crashed"
	game.SystemProfit = systemProfit
	now := time.Now()
	game.CrashedAt = &now

	// Обновляем игру в базе данных
	err := ugm.saveCrashGame(ctx, game)
	if err != nil {
		return fmt.Errorf("failed to update crashed game: %w", err)
	}

	log.Printf("Crash game %s crashed at %.2fx: system profit %d BKC", gameID, game.CrashPoint, systemProfit)
	return nil
}

// UpdateExchangePrice - обновление цены обмена
func (ugm *UnifiedGamesManager) UpdateExchangePrice(ctx context.Context, pair string, price float64) error {
	// Получаем предыдущую цену для расчета изменения
	prevPrice := ugm.getPreviousPrice(ctx, pair)
	change24h := 0.0
	if prevPrice > 0 {
		change24h = ((price - prevPrice) / prevPrice) * 100
	}

	// Создаем запись о цене
	exchangePrice := &ExchangePrice{
		Pair:      pair,
		Price:     price,
		Change24h: change24h,
		CreatedAt: time.Now(),
	}

	// Сохраняем в базу данных
	err := ugm.saveExchangePrice(ctx, exchangePrice)
	if err != nil {
		return fmt.Errorf("failed to save exchange price: %w", err)
	}

	log.Printf("Exchange price updated: %s = %.6f (24h change: %.2f%%)", pair, price, change24h)
	return nil
}

// getPreviousPrice - получение предыдущей цены
func (ugm *UnifiedGamesManager) getPreviousPrice(ctx context.Context, pair string) float64 {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем заглушку
	return 1.0
}

// saveExchangePrice - сохранение цены обмена
func (ugm *UnifiedGamesManager) saveExchangePrice(ctx context.Context, price *ExchangePrice) error {
	// Здесь должна быть реальная запись в базу данных
	log.Printf("Saving exchange price: %+v", price)
	return nil
}

// GetExchangePrices - получение цен обмена
func (ugm *UnifiedGamesManager) GetExchangePrices(ctx context.Context) ([]ExchangePrice, error) {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем тестовые данные
	prices := []ExchangePrice{
		{Pair: "BKC/USD", Price: 0.001, Change24h: 2.5, CreatedAt: time.Now()},
		{Pair: "BKC/TON", Price: 0.0008, Change24h: -1.2, CreatedAt: time.Now()},
		{Pair: "BKC/BTC", Price: 0.00000002, Change24h: 5.0, CreatedAt: time.Now()},
	}

	return prices, nil
}

// GetGameHistory - получение истории игр
func (ugm *UnifiedGamesManager) GetGameHistory(ctx context.Context, gameType string, limit int) ([]GameResult, error) {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем тестовые данные
	history := make([]GameResult, 0, limit)

	for i := 0; i < limit; i++ {
		history = append(history, GameResult{
			Success:    mathrand.Float64() > 0.3, // 70% выигрышей
			GameID:     fmt.Sprintf("%s_%d", gameType, i),
			PlayerID:   int64(mathrand.Int63n(10000) + 1),
			Amount:     int64(mathrand.Int63n(5000) + 100),
			Winnings:   int64(mathrand.Int63n(1000)),
			Multiplier: 1.0 + mathrand.Float64()*5.0,
			PlayedAt:   time.Now().Add(-time.Duration(i) * time.Hour),
		})
	}

	return history, nil
}

// GetPlayerStats - получение статистики игрока
func (ugm *UnifiedGamesManager) GetPlayerStats(ctx context.Context, playerID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Получаем историю игр игрока
	history, err := ugm.GetGameHistory(ctx, "crash", 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get game history: %w", err)
	}

	// Считаем статистику
	var totalGames, wins, losses int
	var totalWagered, totalWon int64

	for _, game := range history {
		if game.PlayerID == playerID {
			totalGames++
			totalWagered += game.Amount
			if game.Success {
				wins++
				totalWon += game.Winnings
			} else {
				losses++
			}
		}
	}

	winRate := 0.0
	if totalGames > 0 {
		winRate = float64(wins) / float64(totalGames) * 100
	}

	roi := 0.0
	if totalWagered > 0 {
		roi = float64(totalWon-totalWagered) / float64(totalWagered) * 100
	}

	stats["total_games"] = totalGames
	stats["wins"] = wins
	stats["losses"] = losses
	stats["win_rate"] = winRate
	stats["total_wagered"] = totalWagered
	stats["total_won"] = totalWon
	stats["roi"] = roi

	return stats, nil
}

// GetLeaderboard - получение таблицы лидеров
func (ugm *UnifiedGamesManager) GetLeaderboard(ctx context.Context, metric string, limit int) ([]map[string]interface{}, error) {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем тестовые данные
	leaderboard := make([]map[string]interface{}, 0, limit)

	for i := 0; i < limit; i++ {
		player := map[string]interface{}{
			"rank":      i + 1,
			"player_id": int64(mathrand.Int63n(10000) + 1),
			"username":  fmt.Sprintf("Player%d", i+1),
			"score":     int64(mathrand.Int63n(100000) + 1000),
			"wins":      int64(mathrand.Int63n(100) + 10),
			"win_rate":  mathrand.Float64() * 100,
		}
		leaderboard = append(leaderboard, player)
	}

	return leaderboard, nil
}

// ValidateGame - валидация игры
func (ugm *UnifiedGamesManager) ValidateGame(ctx context.Context, gameID string) (bool, error) {
	// В реальном приложении здесь будет проверка в БД
	// Для примера просто проверяем формат
	if len(gameID) < 5 {
		return false, fmt.Errorf("invalid game ID format")
	}

	return true, nil
}

// GetActiveGames - получение активных игр
func (ugm *UnifiedGamesManager) GetActiveGames(ctx context.Context) ([]CrashGame, error) {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем тестовые данные
	activeGames := []CrashGame{
		{
			ID:        1,
			GameID:    "crash_1234567890",
			Status:    "active",
			TotalBets: 2500,
			StartedAt: time.Now().Add(-5 * time.Minute),
		},
		{
			ID:        2,
			GameID:    "crash_1234567891",
			Status:    "waiting",
			TotalBets: 0,
			StartedAt: time.Now().Add(-1 * time.Minute),
		},
	}

	return activeGames, nil
}
