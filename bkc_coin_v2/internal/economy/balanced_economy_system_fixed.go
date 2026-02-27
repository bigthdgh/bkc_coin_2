package economy

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

// BalancedEconomySystem - оптимизированная экономическая система
type BalancedEconomySystem struct {
	redisClient *redis.Client
}

// NewBalancedEconomySystem создает новую экономическую систему
func NewBalancedEconomySystem(redisClient *redis.Client) *BalancedEconomySystem {
	return &BalancedEconomySystem{
		redisClient: redisClient,
	}
}

// UserEconomy представляет экономическое состояние пользователя
type UserEconomy struct {
	UserID          int64   `json:"user_id"`
	Balance         float64 `json:"balance"`
	Energy          int64   `json:"energy"`
	MaxEnergy       int64   `json:"max_energy"`
	TapsToday       int64   `json:"taps_today"`
	LastTapTime     int64   `json:"last_tap_time"`
	ReferralCount   int64   `json:"referral_count"`
	ReferralIncome  float64 `json:"referral_income"`
	Level           int64   `json:"level"`
	Experience      int64   `json:"experience"`
	LastDailyBonus  int64   `json:"last_daily_bonus"`
	StreakDays      int64   `json:"streak_days"`
}

// TapResult представляет результат тапа
type TapResult struct {
	Success        bool    `json:"success"`
	EnergyGained   int64   `json:"energy_gained"`
	CoinsEarned    float64 `json:"coins_earned"`
	NewBalance     float64 `json:"new_balance"`
	NewEnergy      int64   `json:"new_energy"`
	TapsRemaining  int64   `json:"taps_remaining"`
	ComboMultiplier float64 `json:"combo_multiplier"`
	LevelUp        bool    `json:"level_up"`
}

// ReferralBonus представляет реферальный бонус
type ReferralBonus struct {
	ReferrerID int64   `json:"referrer_id"`
	Bonus      float64 `json:"bonus"`
	Level      int64   `json:"level"`
	Timestamp  int64   `json:"timestamp"`
}

// DailyBonus представляет ежедневный бонус
type DailyBonus struct {
	Day      int64   `json:"day"`
	Bonus    float64 `json:"bonus"`
	Streak   int64   `json:"streak"`
	Claimed  bool    `json:"claimed"`
}

// GetUserEconomy получает экономическое состояние пользователя
func (bes *BalancedEconomySystem) GetUserEconomy(ctx context.Context, userID int64) (*UserEconomy, error) {
	key := fmt.Sprintf("user_economy:%d", userID)
	
	data, err := bes.redisClient.HGetAll(ctx, key).Result()
	if err == redis.Nil {
		// Создаем нового пользователя
		return bes.createNewUser(ctx, userID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user economy: %w", err)
	}

	economy := &UserEconomy{
		UserID: userID,
	}

	// Парсим данные
	if balance, ok := data["balance"]; ok {
		economy.Balance = parseFloat(balance)
	}
	if energy, ok := data["energy"]; ok {
		economy.Energy = parseInt64(energy)
	}
	if maxEnergy, ok := data["max_energy"]; ok {
		economy.MaxEnergy = parseInt64(maxEnergy)
	}
	if tapsToday, ok := data["taps_today"]; ok {
		economy.TapsToday = parseInt64(tapsToday)
	}
	if lastTapTime, ok := data["last_tap_time"]; ok {
		economy.LastTapTime = parseInt64(lastTapTime)
	}
	if referralCount, ok := data["referral_count"]; ok {
		economy.ReferralCount = parseInt64(referralCount)
	}
	if referralIncome, ok := data["referral_income"]; ok {
		economy.ReferralIncome = parseFloat(referralIncome)
	}
	if level, ok := data["level"]; ok {
		economy.Level = parseInt64(level)
	}
	if experience, ok := data["experience"]; ok {
		economy.Experience = parseInt64(experience)
	}
	if lastDailyBonus, ok := data["last_daily_bonus"]; ok {
		economy.LastDailyBonus = parseInt64(lastDailyBonus)
	}
	if streakDays, ok := data["streak_days"]; ok {
		economy.StreakDays = parseInt64(streakDays)
	}

	// Восстанавливаем энергию
	bes.restoreEnergy(ctx, economy)

	return economy, nil
}

// createNewUser создает нового пользователя с начальными параметрами
func (bes *BalancedEconomySystem) createNewUser(ctx context.Context, userID int64) (*UserEconomy, error) {
	economy := &UserEconomy{
		UserID:         userID,
		Balance:        0,
		Energy:         1000, // Начальная энергия
		MaxEnergy:      1000, // Максимальная энергия
		TapsToday:      0,
		LastTapTime:    time.Now().Unix(),
		ReferralCount:  0,
		ReferralIncome: 0,
		Level:          1,
		Experience:     0,
		LastDailyBonus: 0,
		StreakDays:     0,
	}

	// Сохраняем в Redis
	err := bes.saveUserEconomy(ctx, economy)
	if err != nil {
		return nil, fmt.Errorf("failed to save new user: %w", err)
	}

	return economy, nil
}

// restoreEnergy восстанавливает энергию пользователя
func (bes *BalancedEconomySystem) restoreEnergy(ctx context.Context, economy *UserEconomy) {
	now := time.Now()
	lastTap := time.Unix(economy.LastTapTime, 0)
	
	// Восстановление 1 энергии в секунду
	secondsPassed := now.Sub(lastTap).Seconds()
	energyToRestore := int64(math.Floor(secondsPassed))
	
	if energyToRestore > 0 {
		economy.Energy = min(economy.Energy+energyToRestore, economy.MaxEnergy)
		economy.LastTapTime = now.Unix()
		
		// Сохраняем обновленное время и энергию
		key := fmt.Sprintf("user_economy:%d", economy.UserID)
		bes.redisClient.HSet(ctx, key, "energy", economy.Energy)
		bes.redisClient.HSet(ctx, key, "last_tap_time", economy.LastTapTime)
	}
}

// ProcessTap обрабатывает тап пользователя
func (bes *BalancedEconomySystem) ProcessTap(ctx context.Context, userID int64) (*TapResult, error) {
	economy, err := bes.GetUserEconomy(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user economy: %w", err)
	}

	// Проверяем дневной лимит
	if economy.TapsToday >= 300 { // Ограничили до 300 тапов в день
		return &TapResult{
			Success:       false,
			TapsRemaining: 0,
		}, nil
	}

	// Проверяем энергию
	if economy.Energy <= 0 {
		return &TapResult{
			Success:       false,
			TapsRemaining: economy.TapsToday,
			NewEnergy:     economy.Energy,
		}, nil
	}

	// Рассчитываем награду
	baseReward := 1.0 // Базовая награда 1 BKC
	
	// NFT бонусы (пока не реализованы)
	nftBonus := 1.0
	
	// Реферальные бонусы (пока не реализованы)
	referralBonus := 1.0
	
	// Комбо система
	comboMultiplier := bes.calculateComboMultiplier(economy)
	
	// Итоговая награда
	finalReward := baseReward * nftBonus * referralBonus * comboMultiplier
	
	// Обновляем экономику
	economy.Energy--
	economy.TapsToday++
	economy.Balance += finalReward
	economy.Experience++
	economy.LastTapTime = time.Now().Unix()

	// Проверяем повышение уровня
	levelUp := bes.checkLevelUp(economy)

	// Сохраняем изменения
	err = bes.saveUserEconomy(ctx, economy)
	if err != nil {
		return nil, fmt.Errorf("failed to save user economy: %w", err)
	}

	return &TapResult{
		Success:         true,
		EnergyGained:    0,
		CoinsEarned:     finalReward,
		NewBalance:     economy.Balance,
		NewEnergy:      economy.Energy,
		TapsRemaining:  300 - economy.TapsToday,
		ComboMultiplier: comboMultiplier,
		LevelUp:        levelUp,
	}, nil
}

// calculateComboMultiplier рассчитывает множитель комбо
func (bes *BalancedEconomySystem) calculateComboMultiplier(economy *UserEconomy) float64 {
	// Простая комбо система: чем больше тапов подряд, тем выше множитель
	if economy.TapsToday < 50 {
		return 1.0
	} else if economy.TapsToday < 100 {
		return 1.1
	} else if economy.TapsToday < 200 {
		return 1.2
	} else {
		return 1.3 // Максимум 1.3x для 300 тапов
	}
}

// checkLevelUp проверяет повышение уровня
func (bes *BalancedEconomySystem) checkLevelUp(economy *UserEconomy) bool {
	experienceNeeded := economy.Level * 100 // 100 опыта на уровень
	
	if economy.Experience >= experienceNeeded {
		economy.Level++
		economy.Experience = economy.Experience - experienceNeeded
		economy.MaxEnergy = 1000 + (economy.Level-1)*100 // +100 энергии за уровень
		return true
	}
	
	return false
}

// ProcessReferral обрабатывает реферала (ОДНОУРОВНЕВАЯ СИСТЕМА)
func (bes *BalancedEconomySystem) ProcessReferral(ctx context.Context, referrerID, referredID int64) error {
	// Проверяем, не превышен ли лимит рефералов
	referrerEconomy, err := bes.GetUserEconomy(ctx, referrerID)
	if err != nil {
		return fmt.Errorf("failed to get referrer economy: %w", err)
	}

	// Ограничиваем количество рефералов
	if referrerEconomy.ReferralCount >= 10 { // Максимум 10 рефералов
		return fmt.Errorf("referral limit reached")
	}

	// Начисляем бонус рефереру
	referralBonus := 10.0 // Фиксированный бонус 10 BKC
	
	referrerEconomy.ReferralCount++
	referrerEconomy.ReferralIncome += referralBonus
	referrerEconomy.Balance += referralBonus

	// Сохраняем изменения
	err = bes.saveUserEconomy(ctx, referrerEconomy)
	if err != nil {
		return fmt.Errorf("failed to save referrer economy: %w", err)
	}

	// Записываем реферальную связь
	referralKey := fmt.Sprintf("referral:%d", referredID)
	referralData := map[string]interface{}{
		"referrer_id": referrerID,
		"bonus":       referralBonus,
		"timestamp":   time.Now().Unix(),
	}
	
	err = bes.redisClient.HMSet(ctx, referralKey, referralData).Err()
	if err != nil {
		return fmt.Errorf("failed to save referral: %w", err)
	}

	// Устанавливаем TTL на 30 дней
	bes.redisClient.Expire(ctx, referralKey, 30*24*time.Hour)

	return nil
}

// GetDailyBonus получает ежедневный бонус
func (bes *BalancedEconomySystem) GetDailyBonus(ctx context.Context, userID int64) (*DailyBonus, error) {
	economy, err := bes.GetUserEconomy(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user economy: %w", err)
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	
	// Проверяем, получал ли пользователь бонус сегодня
	lastBonusTime := time.Unix(economy.LastDailyBonus, 0)
	if lastBonusTime.Format("2006-01-02") == today {
		return &DailyBonus{
			Day:     economy.StreakDays,
			Bonus:   0,
			Streak:  economy.StreakDays,
			Claimed: true,
		}, nil
	}

	// Рассчитываем бонус
	economy.StreakDays++
	
	// Прогрессивный бонус
	var bonus float64
	switch economy.StreakDays {
	case 1:
		bonus = 10
	case 2:
		bonus = 15
	case 3:
		bonus = 20
	case 4:
		bonus = 25
	case 5:
		bonus = 30
	case 6:
		bonus = 35
	case 7:
		bonus = 50 // Бонус за неделю
	default:
		bonus = 30 // После недели - 30 BKC
	}

	// Начисляем бонус
	economy.Balance += bonus
	economy.LastDailyBonus = now.Unix()

	// Сохраняем изменения
	err = bes.saveUserEconomy(ctx, economy)
	if err != nil {
		return nil, fmt.Errorf("failed to save user economy: %w", err)
	}

	return &DailyBonus{
		Day:     economy.StreakDays,
		Bonus:   bonus,
		Streak:  economy.StreakDays,
		Claimed: true,
	}, nil
}

// ResetDailyStats сбрасывает дневную статистику
func (bes *BalancedEconomySystem) ResetDailyStats(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("user_economy:%d", userID)
	
	// Сбрасываем тапы
	err := bes.redisClient.HSet(ctx, key, "taps_today", 0).Err()
	if err != nil {
		return fmt.Errorf("failed to reset daily stats: %w", err)
	}

	return nil
}

// GetReferralStats получает статистику рефералов
func (bes *BalancedEconomySystem) GetReferralStats(ctx context.Context, userID int64) (map[string]interface{}, error) {
	economy, err := bes.GetUserEconomy(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user economy: %w", err)
	}

	stats := map[string]interface{}{
		"referral_count":  economy.ReferralCount,
		"referral_income": economy.ReferralIncome,
		"max_referrals":   10, // Максимум 10 рефералов
		"bonus_per_ref":   10, // 10 BKC за реферала
	}

	return stats, nil
}

// GetLeaderboard получает таблицу лидеров
func (bes *BalancedEconomySystem) GetLeaderboard(ctx context.Context, limit int64) ([]map[string]interface{}, error) {
	// Получаем топ пользователей по балансу
	keys, err := bes.redisClient.Keys(ctx, "user_economy:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user keys: %w", err)
	}

	type LeaderboardEntry struct {
		UserID  int64   `json:"user_id"`
		Balance float64 `json:"balance"`
		Level   int64   `json:"level"`
	}

	var entries []LeaderboardEntry

	for _, key := range keys {
		data, err := bes.redisClient.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		entry := LeaderboardEntry{}
		if userID, ok := data["user_id"]; ok {
			entry.UserID = parseInt64(userID)
		}
		if balance, ok := data["balance"]; ok {
			entry.Balance = parseFloat(balance)
		}
		if level, ok := data["level"]; ok {
			entry.Level = parseInt64(level)
		}

		entries = append(entries, entry)
	}

	// Сортируем по балансу
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Balance > entries[i].Balance {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Ограничиваем количество
	if int64(len(entries)) > limit {
		entries = entries[:limit]
	}

	// Конвертируем в результат
	result := make([]map[string]interface{}, len(entries))
	for i, entry := range entries {
		result[i] = map[string]interface{}{
			"rank":    i + 1,
			"user_id": entry.UserID,
			"balance": entry.Balance,
			"level":   entry.Level,
		}
	}

	return result, nil
}

// saveUserEconomy сохраняет экономическое состояние пользователя
func (bes *BalancedEconomySystem) saveUserEconomy(ctx context.Context, economy *UserEconomy) error {
	key := fmt.Sprintf("user_economy:%d", economy.UserID)
	
	data := map[string]interface{}{
		"user_id":         economy.UserID,
		"balance":         economy.Balance,
		"energy":          economy.Energy,
		"max_energy":      economy.MaxEnergy,
		"taps_today":      economy.TapsToday,
		"last_tap_time":    economy.LastTapTime,
		"referral_count":  economy.ReferralCount,
		"referral_income": economy.ReferralIncome,
		"level":           economy.Level,
		"experience":      economy.Experience,
		"last_daily_bonus": economy.LastDailyBonus,
		"streak_days":     economy.StreakDays,
	}
	
	err := bes.redisClient.HMSet(ctx, key, data).Err()
	if err != nil {
		return fmt.Errorf("failed to save user economy: %w", err)
	}

	// Устанавливаем TTL на 30 дней
	bes.redisClient.Expire(ctx, key, 30*24*time.Hour)

	return nil
}

// Вспомогательные функции
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
