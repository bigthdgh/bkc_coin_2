package economy

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// BalancedEconomySystem - сбалансированная экономическая система
type BalancedEconomySystem struct {
	mu sync.RWMutex

	// Основные параметры
	totalSupply   float64
	currentSupply float64
	targetPrice   float64
	currentPrice  float64

	// Энергетическая система
	maxEnergy        int64
	energyPerHour    int64
	dailyEnergyLimit int64
	tapCost          float64
	tapReward        float64

	// Реферальная система
	referralRequirement int64
	referralReward      float64
	referralBonus       float64
	newUserBonus        float64

	// Эмиссия и сжигание
	dailyEmissionCap float64
	burnRate         float64
	stakingAPY       float64

	// Стабилизационный фонд
	stabilizationFund float64
	fundAllocation    float64

	// Админские функции
	adminLockedFunds map[int64]AdminLock
	adminUnlockRate  float64

	// Защита от массовых продаж
	saleProtectionEnabled bool
	maxDailySaleAmount    float64
	saleTaxRate           float64

	// Динамическая система
	priceElasticity  float64
	userGrowthTarget float64
	maxUsers         int64

	// Исторические данные
	dailyMetrics []DailyMetrics
	priceHistory []PricePoint

	// Конфигурация
	config BalancedEconomyConfig

	// Контекст
	ctx    context.Context
	cancel context.CancelFunc
}

// BalancedEconomyConfig - конфигурация сбалансированной экономики
type BalancedEconomyConfig struct {
	TotalSupply         float64 `json:"total_supply"`
	TargetPrice         float64 `json:"target_price"`
	MaxEnergy           int64   `json:"max_energy"`
	EnergyPerHour       int64   `json:"energy_per_hour"`
	DailyEnergyLimit    int64   `json:"daily_energy_limit"`
	TapCost             float64 `json:"tap_cost"`
	TapReward           float64 `json:"tap_reward"`
	ReferralRequirement int64   `json:"referral_requirement"`
	ReferralReward      float64 `json:"referral_reward"`
	ReferralBonus       float64 `json:"referral_bonus"`
	NewUserBonus        float64 `json:"new_user_bonus"`
	DailyEmissionCap    float64 `json:"daily_emission_cap"`
	BurnRate            float64 `json:"burn_rate"`
	StakingAPY          float64 `json:"staking_apy"`
	FundAllocation      float64 `json:"fund_allocation"`
	AdminUnlockRate     float64 `json:"admin_unlock_rate"`
	MaxDailySaleAmount  float64 `json:"max_daily_sale_amount"`
	SaleTaxRate         float64 `json:"sale_tax_rate"`
	PriceElasticity     float64 `json:"price_elasticity"`
	UserGrowthTarget    float64 `json:"user_growth_target"`
	MaxUsers            int64   `json:"max_users"`
}

// AdminLock - блокировка админских средств
type AdminLock struct {
	Amount       float64   `json:"amount"`
	LockDate     time.Time `json:"lock_date"`
	UnlockDate   time.Time `json:"unlock_date"`
	MonthlyLimit float64   `json:"monthly_limit"`
	Withdrawn    float64   `json:"withdrawn"`
	Reason       string    `json:"reason"`
}

// DailyMetrics - дневные метрики
type DailyMetrics struct {
	Date          time.Time `json:"date"`
	ActiveUsers   int64     `json:"active_users"`
	NewUsers      int64     `json:"new_users"`
	TotalTaps     int64     `json:"total_taps"`
	EnergyUsed    int64     `json:"energy_used"`
	TokensEmitted float64   `json:"tokens_emitted"`
	TokensBurned  float64   `json:"tokens_burned"`
	ReferralCount int64     `json:"referral_count"`
	TotalSales    float64   `json:"total_sales"`
	MarketCap     float64   `json:"market_cap"`
	Price         float64   `json:"price"`
	Volatility    float64   `json:"volatility"`
}

// PricePoint - точка цены
type PricePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Supply    float64   `json:"supply"`
}

// UserTapResult - результат тапа
type UserTapResult struct {
	Success        bool      `json:"success"`
	NewBalance     float64   `json:"new_balance"`
	NewEnergy      int64     `json:"new_energy"`
	EnergyUsed     int64     `json:"energy_used"`
	TapsMade       int64     `json:"taps_made"`
	TokensEarned   float64   `json:"tokens_earned"`
	Message        string    `json:"message"`
	CanTapMore     bool      `json:"can_tap_more"`
	NextEnergyTime time.Time `json:"next_energy_time"`
}

// ReferralResult - результат реферала
type ReferralResult struct {
	Success      bool    `json:"success"`
	ReferrerID   int64   `json:"referrer_id"`
	ReferralID   int64   `json:"referral_id"`
	RewardAmount float64 `json:"reward_amount"`
	BonusAmount  float64 `json:"bonus_amount"`
	Message      string  `json:"message"`
}

// NewBalancedEconomySystem создает новую сбалансированную экономическую систему
func NewBalancedEconomySystem(config BalancedEconomyConfig) *BalancedEconomySystem {
	ctx, cancel := context.WithCancel(context.Background())

	system := &BalancedEconomySystem{
		totalSupply:           config.TotalSupply,
		currentSupply:         config.TotalSupply * 0.7, // 70% выпущено
		targetPrice:           config.TargetPrice,
		currentPrice:          config.TargetPrice,
		maxEnergy:             config.MaxEnergy,
		energyPerHour:         config.EnergyPerHour,
		dailyEnergyLimit:      config.DailyEnergyLimit,
		tapCost:               config.TapCost,
		tapReward:             config.TapReward,
		referralRequirement:   config.ReferralRequirement,
		referralReward:        config.ReferralReward,
		referralBonus:         config.ReferralBonus,
		newUserBonus:          config.NewUserBonus,
		dailyEmissionCap:      config.DailyEmissionCap,
		burnRate:              config.BurnRate,
		stakingAPY:            config.StakingAPY,
		fundAllocation:        config.FundAllocation,
		stabilizationFund:     config.TotalSupply * config.FundAllocation,
		adminLockedFunds:      make(map[int64]AdminLock),
		adminUnlockRate:       config.AdminUnlockRate,
		saleProtectionEnabled: true,
		maxDailySaleAmount:    config.MaxDailySaleAmount,
		saleTaxRate:           config.SaleTaxRate,
		priceElasticity:       config.PriceElasticity,
		userGrowthTarget:      config.UserGrowthTarget,
		maxUsers:              config.MaxUsers,
		config:                config,
		ctx:                   ctx,
		cancel:                cancel,
	}

	// Запуск фоновых процессов
	go system.startBackgroundProcesses()

	return system
}

// ProcessUserTap обрабатывает тап пользователя
func (e *BalancedEconomySystem) ProcessUserTap(userID int64, currentBalance float64, currentEnergy int64, tapsRequested int64) UserTapResult {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := UserTapResult{
		Success:      false,
		NewBalance:   currentBalance,
		NewEnergy:    currentEnergy,
		TapsMade:     0,
		TokensEarned: 0,
	}

	// Проверка лимитов
	if currentEnergy <= 0 {
		result.Message = "Недостаточно энергии. Подождите восстановления."
		result.NextEnergyTime = time.Now().Add(time.Hour / time.Duration(e.energyPerHour))
		return result
	}

	// Расчет доступных тапов
	maxTaps := currentEnergy / int64(e.tapCost)
	if tapsRequested > maxTaps {
		tapsRequested = maxTaps
	}

	// Проверка дневного лимита энергии - 3000 тапов в день
	today := time.Now().Truncate(24 * time.Hour)
	dailyEnergyUsed := e.getDailyEnergyUsed(userID, today)
	if dailyEnergyUsed+int64(e.tapCost)*tapsRequested > 3000 {
		remainingEnergy := 3000 - dailyEnergyUsed
		maxTapsFromLimit := remainingEnergy / int64(e.tapCost)
		if tapsRequested > maxTapsFromLimit {
			tapsRequested = maxTapsFromLimit
		}
	}

	if tapsRequested <= 0 {
		result.Message = "Достигнут дневной лимит энергии."
		return result
	}

	// Расчет награды - 0.1 BKC за тап
	dynamicReward := 0.1
	totalReward := float64(tapsRequested) * dynamicReward
	energyCost := float64(tapsRequested) * e.tapCost

	// Проверка эмиссионного лимита
	dailyEmission := e.getDailyEmission(today)
	if dailyEmission+totalReward > e.dailyEmissionCap {
		availableEmission := e.dailyEmissionCap - dailyEmission
		maxTapsFromEmission := int64(availableEmission / dynamicReward)
		if tapsRequested > maxTapsFromEmission {
			tapsRequested = maxTapsFromEmission
			totalReward = float64(tapsRequested) * dynamicReward
			energyCost = float64(tapsRequested) * e.tapCost
		}
	}

	// Применение тапа
	result.Success = true
	result.NewBalance = currentBalance + totalReward
	result.NewEnergy = currentEnergy - int64(energyCost)
	result.TapsMade = tapsRequested
	result.TokensEarned = totalReward
	result.EnergyUsed = int64(energyCost)
	result.Message = fmt.Sprintf("Успешно! Заработано %.2f BKC", totalReward)
	result.CanTapMore = result.NewEnergy > 0
	result.NextEnergyTime = time.Now().Add(time.Hour / time.Duration(e.energyPerHour))

	// Обновление статистики
	e.updateDailyMetrics(userID, tapsRequested, int64(energyCost), totalReward)

	// Обновление предложения
	e.currentSupply += totalReward

	log.Printf("User %d tapped %d times, earned %.2f BKC", userID, tapsRequested, totalReward)

	return result
}

// calculateDynamicTapReward рассчитывает динамическую награду за тап
func (e *BalancedEconomySystem) calculateDynamicTapReward() float64 {
	// Базовая награда
	baseReward := e.tapReward

	// Корректировка по цене
	priceFactor := e.targetPrice / e.currentPrice

	// Корректировка по количеству пользователей
	userFactor := e.calculateUserGrowthFactor()

	// Корректировка по предложению токенов
	supplyFactor := e.currentSupply / e.totalSupply

	// Итоговая награда
	dynamicReward := baseReward * priceFactor * userFactor * (1.0 - supplyFactor*0.5)

	// Ограничения
	if dynamicReward > baseReward*2.0 {
		dynamicReward = baseReward * 2.0 // Максимум +100%
	}
	if dynamicReward < baseReward*0.1 {
		dynamicReward = baseReward * 0.1 // Минимум -90%
	}

	return dynamicReward
}

// calculateUserGrowthFactor рассчитывает фактор роста пользователей
func (e *BalancedEconomySystem) calculateUserGrowthFactor() float64 {
	// Целевой рост пользователей
	targetGrowth := e.userGrowthTarget

	// Текущий рост (упрощенный)
	currentGrowth := e.getCurrentUserGrowth()

	// Фактор роста
	growthFactor := 1.0 + (targetGrowth-currentGrowth)*0.5

	// Ограничения
	if growthFactor > 1.5 {
		growthFactor = 1.5 // Максимум +50%
	}
	if growthFactor < 0.5 {
		growthFactor = 0.5 // Минимум -50%
	}

	return growthFactor
}

// getCurrentUserGrowth получает текущий рост пользователей
func (e *BalancedEconomySystem) getCurrentUserGrowth() float64 {
	if len(e.dailyMetrics) < 7 {
		return e.userGrowthTarget
	}

	// Рост за последние 7 дней
	recent := e.dailyMetrics[len(e.dailyMetrics)-7:]
	if len(recent) < 2 {
		return e.userGrowthTarget
	}

	first := recent[0]
	last := recent[len(recent)-1]

	if first.ActiveUsers == 0 {
		return e.userGrowthTarget
	}

	growth := float64(last.ActiveUsers-first.ActiveUsers) / float64(first.ActiveUsers)

	return growth
}

// ProcessReferral обрабатывает реферала
func (e *BalancedEconomySystem) ProcessReferral(referrerID, referralID int64) ReferralResult {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := ReferralResult{
		Success:    false,
		ReferrerID: referrerID,
		ReferralID: referralID,
	}

	// Проверка требований - 4 реферала для 100 BKC
	referralCount := e.getReferralCount(referrerID)

	if referralCount < 4 {
		result.Message = fmt.Sprintf("Приведите еще %d рефералов для получения награды",
			4-referralCount)
		return result
	}

	// Расчет награды - 100 BKC за 4 рефералов
	rewardAmount := 100.0
	bonusAmount := 0.0

	// Бонус за превышение требований
	if referralCount > 4 {
		extraReferrals := referralCount - 4
		bonusAmount = float64(extraReferrals) * e.referralBonus
	}

	// Проверка эмиссионного лимита
	today := time.Now().Truncate(24 * time.Hour)
	dailyEmission := e.getDailyEmission(today)
	totalReward := rewardAmount + bonusAmount

	if dailyEmission+totalReward > e.dailyEmissionCap {
		availableEmission := e.dailyEmissionCap - dailyEmission
		if totalReward > availableEmission {
			// Пропорциональное сокращение наград
			ratio := availableEmission / totalReward
			rewardAmount *= ratio
			bonusAmount *= ratio
			totalReward = rewardAmount + bonusAmount
		}
	}

	result.Success = true
	result.RewardAmount = rewardAmount
	result.BonusAmount = bonusAmount
	result.Message = fmt.Sprintf("Получено %.2f BKC реферальной награды + %.2f BKC бонуса",
		rewardAmount, bonusAmount)

	// Обновление статистики
	e.updateReferralMetrics(referrerID, totalReward)

	// Обновление предложения
	e.currentSupply += totalReward

	log.Printf("Referral: %d -> %d, reward: %.2f BKC", referrerID, referralID, totalReward)

	return result
}

// ProcessNewUserBonus обрабатывает бонус для нового пользователя
func (e *BalancedEconomySystem) ProcessNewUserBonus(userID int64) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Проверка эмиссионного лимита
	today := time.Now().Truncate(24 * time.Hour)
	dailyEmission := e.getDailyEmission(today)

	if dailyEmission+10.0 > e.dailyEmissionCap {
		availableEmission := e.dailyEmissionCap - dailyEmission
		if availableEmission <= 0 {
			return 0
		}
		return availableEmission
	}

	// Обновление предложения
	e.currentSupply += 10.0

	log.Printf("New user bonus: %d, amount: 10.0 BKC", userID)

	return 10.0
}

// getReferralCount получает количество рефералов
func (e *BalancedEconomySystem) getReferralCount(userID int64) int64 {
	// В реальной системе здесь будет запрос к БД
	// Для примера возвращаем случайное значение
	return int64(time.Now().Unix() % 10)
}

// getDailyEnergyUsed получает использованную энергию за день
func (e *BalancedEconomySystem) getDailyEnergyUsed(userID int64, date time.Time) int64 {
	// В реальной системе здесь будет запрос к БД
	// Для примера возвращаем случайное значение
	return int64(date.Unix() % 3000)
}

// getDailyEmission получает дневную эмиссию
func (e *BalancedEconomySystem) getDailyEmission(date time.Time) float64 {
	for _, metric := range e.dailyMetrics {
		if metric.Date.Truncate(24 * time.Hour).Equal(date.Truncate(24 * time.Hour)) {
			return metric.TokensEmitted
		}
	}
	return 0
}

// updateDailyMetrics обновляет дневные метрики
func (e *BalancedEconomySystem) updateDailyMetrics(userID int64, taps, energyUsed int64, tokensEarned float64) {
	today := time.Now().Truncate(24 * time.Hour)

	var metric *DailyMetrics
	for i, m := range e.dailyMetrics {
		if m.Date.Truncate(24 * time.Hour).Equal(today) {
			metric = &e.dailyMetrics[i]
			break
		}
	}

	if metric == nil {
		metric = &DailyMetrics{
			Date:          today,
			ActiveUsers:   1,
			NewUsers:      0,
			TotalTaps:     taps,
			EnergyUsed:    energyUsed,
			TokensEmitted: tokensEarned,
		}
		e.dailyMetrics = append(e.dailyMetrics, *metric)

		// Ограничение истории
		if len(e.dailyMetrics) > 365 {
			e.dailyMetrics = e.dailyMetrics[len(e.dailyMetrics)-365:]
		}
	} else {
		metric.TotalTaps += taps
		metric.EnergyUsed += energyUsed
		metric.TokensEmitted += tokensEarned
	}
}

// updateReferralMetrics обновляет реферальные метрики
func (e *BalancedEconomySystem) updateReferralMetrics(userID int64, amount float64) {
	today := time.Now().Truncate(24 * time.Hour)

	var metric *DailyMetrics
	for i, m := range e.dailyMetrics {
		if m.Date.Truncate(24 * time.Hour).Equal(today) {
			metric = &e.dailyMetrics[i]
			break
		}
	}

	if metric != nil {
		metric.ReferralCount++
	}
}

// LockAdminFunds блокирует админские средства
func (e *BalancedEconomySystem) LockAdminFunds(adminID int64, amount float64, reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Проверка существующей блокировки
	if _, exists := e.adminLockedFunds[adminID]; exists {
		return fmt.Errorf("администратор %d уже имеет заблокированные средства", adminID)
	}

	// Создание блокировки
	lock := AdminLock{
		Amount:       amount,
		LockDate:     time.Now(),
		UnlockDate:   time.Now().Add(5 * 30 * 24 * time.Hour), // 5 месяцев
		MonthlyLimit: amount * e.adminUnlockRate,              // 20% в месяц
		Withdrawn:    0,
		Reason:       reason,
	}

	e.adminLockedFunds[adminID] = lock

	log.Printf("Admin %d locked %.2f BKC for 5 months", adminID, amount)

	return nil
}

// WithdrawAdminFunds выводит админские средства
func (e *BalancedEconomySystem) WithdrawAdminFunds(adminID int64, requestedAmount float64) (float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	lock, exists := e.adminLockedFunds[adminID]
	if !exists {
		return 0, fmt.Errorf("у администратора %d нет заблокированных средств", adminID)
	}

	// Проверка даты разблокировки
	if time.Now().Before(lock.UnlockDate) {
		// Проверка месячного лимита
		currentMonth := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -time.Now().Day()+1)
		monthlyWithdrawn := lock.Withdrawn

		if monthlyWithdrawn+requestedAmount > lock.MonthlyLimit {
			available := lock.MonthlyLimit - monthlyWithdrawn
			if available <= 0 {
				return 0, fmt.Errorf("достигнут месячный лимит вывода")
			}
			requestedAmount = available
		}
	} else {
		// Полная разблокировка
		requestedAmount = lock.Amount - lock.Withdrawn
	}

	// Обновление информации
	lock.Withdrawn += requestedAmount
	e.adminLockedFunds[adminID] = lock

	// Удаление если все выведено
	if lock.Withdrawn >= lock.Amount {
		delete(e.adminLockedFunds, adminID)
	}

	log.Printf("Admin %d withdrew %.2f BKC", adminID, requestedAmount)

	return requestedAmount, nil
}

// ProcessSale обрабатывает продажу с защитой от обвала
func (e *BalancedEconomySystem) ProcessSale(userID int64, amount float64) (float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Проверка дневного лимита продаж
	today := time.Now().Truncate(24 * time.Hour)
	dailySales := e.getDailySales(today)

	if dailySales+amount > e.maxDailySaleAmount {
		available := e.maxDailySaleAmount - dailySales
		if available <= 0 {
			return 0, fmt.Errorf("достигнут дневной лимит продаж")
		}
		amount = available
	}

	// Расчет налога на продажу
	taxAmount := amount * e.saleTaxRate
	netAmount := amount - taxAmount

	// Сжигание части налога
	burnAmount := taxAmount * 0.5 // 50% от налога сжигается
	e.currentSupply -= burnAmount

	// Остальная часть в стабфонд
	stabilizationAmount := taxAmount * 0.5
	e.stabilizationFund += stabilizationAmount

	// Обновление статистики
	e.updateSaleMetrics(today, amount)

	log.Printf("Sale: user %d, amount %.2f BKC, tax %.2f BKC, net %.2f BKC",
		userID, amount, taxAmount, netAmount)

	return netAmount, nil
}

// getDailySales получает дневные продажи
func (e *BalancedEconomySystem) getDailySales(date time.Time) float64 {
	for _, metric := range e.dailyMetrics {
		if metric.Date.Truncate(24 * time.Hour).Equal(date.Truncate(24 * time.Hour)) {
			return metric.TotalSales
		}
	}
	return 0
}

// updateSaleMetrics обновляет метрики продаж
func (e *BalancedEconomySystem) updateSaleMetrics(date time.Time, amount float64) {
	for i, metric := range e.dailyMetrics {
		if metric.Date.Truncate(24 * time.Hour).Equal(date.Truncate(24 * time.Hour)) {
			e.dailyMetrics[i].TotalSales += amount
			return
		}
	}
}

// GetSystemStatus получает статус системы
func (e *BalancedEconomySystem) GetSystemStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Расчет здоровья экономики
	healthScore := e.calculateEconomicHealth()

	// Расчет инфляции
	inflationRate := e.calculateInflationRate()

	// Расчет волатильности
	volatility := e.calculateVolatility()

	status := map[string]interface{}{
		"total_supply":         e.totalSupply,
		"current_supply":       e.currentSupply,
		"target_price":         e.targetPrice,
		"current_price":        e.currentPrice,
		"max_energy":           e.maxEnergy,
		"energy_per_hour":      e.energyPerHour,
		"daily_energy_limit":   e.dailyEnergyLimit,
		"tap_cost":             e.tapCost,
		"tap_reward":           e.tapReward,
		"referral_requirement": e.referralRequirement,
		"referral_reward":      e.referralReward,
		"daily_emission_cap":   e.dailyEmissionCap,
		"burn_rate":            e.burnRate,
		"staking_apy":          e.stakingAPY,
		"stabilization_fund":   e.stabilizationFund,
		"health_score":         healthScore,
		"inflation_rate":       inflationRate,
		"volatility":           volatility,
		"sale_protection":      e.saleProtectionEnabled,
		"max_daily_sale":       e.maxDailySaleAmount,
		"sale_tax_rate":        e.saleTaxRate,
		"admin_unlock_rate":    e.adminUnlockRate,
		"max_users":            e.maxUsers,
	}

	return status
}

// calculateEconomicHealth рассчитывает здоровье экономики
func (e *BalancedEconomySystem) calculateEconomicHealth() float64 {
	score := 100.0

	// Штраф за отклонение цены
	priceDeviation := math.Abs(e.currentPrice-e.targetPrice) / e.targetPrice
	score -= priceDeviation * 30

	// Штраф за высокую инфляцию
	inflation := e.calculateInflationRate()
	if inflation > 0.05 { // 5% в месяц
		score -= (inflation - 0.05) * 200
	}

	// Штраф за высокую волатильность
	volatility := e.calculateVolatility()
	if volatility > 0.20 { // 20%
		score -= (volatility - 0.20) * 100
	}

	// Бонус за здоровое предложение
	supplyRatio := e.currentSupply / e.totalSupply
	if supplyRatio > 0.5 && supplyRatio < 0.9 {
		score += 10
	}

	// Бонус за наличие стабфонда
	if e.stabilizationFund > e.totalSupply*0.05 {
		score += 15
	}

	// Ограничение счета
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// calculateInflationRate рассчитывает инфляцию
func (e *BalancedEconomySystem) calculateInflationRate() float64 {
	if len(e.dailyMetrics) < 30 {
		return 0
	}

	// Инфляция за последние 30 дней
	recent := e.dailyMetrics[len(e.dailyMetrics)-30:]
	if len(recent) < 2 {
		return 0
	}

	first := recent[0]
	last := recent[len(recent)-1]

	days := float64(len(recent) - 1)
	dailyInflation := (last.TokensEmitted - first.TokensEmitted) / first.TokensEmitted

	monthlyInflation := dailyInflation * days

	return monthlyInflation
}

// calculateVolatility рассчитывает волатильность
func (e *BalancedEconomySystem) calculateVolatility() float64 {
	if len(e.priceHistory) < 2 {
		return 0
	}

	// Волатильность за последние 7 дней
	recent := e.priceHistory
	if len(recent) > 7 {
		recent = e.priceHistory[len(recent)-7:]
	}

	var sum, sumSq float64
	for i := 1; i < len(recent); i++ {
		returns := (recent[i].Price - recent[i-1].Price) / recent[i-1].Price
		sum += returns
		sumSq += returns * returns
	}

	n := float64(len(recent) - 1)
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	volatility := math.Sqrt(variance)

	return volatility
}

// startBackgroundProcesses запускает фоновые процессы
func (e *BalancedEconomySystem) startBackgroundProcesses() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.performBackgroundTasks()
		}
	}
}

// performBackgroundTasks выполняет фоновые задачи
func (e *BalancedEconomySystem) performBackgroundTasks() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Обновление цены (в реальной системе здесь будет запрос к API)
	e.updatePrice()

	// Проверка здоровья экономики
	health := e.calculateEconomicHealth()
	if health < 30 {
		log.Printf("КРИТИЧЕСКО: Здоровье экономики: %.1f", health)
	} else if health < 50 {
		log.Printf("ВНИМАНИЕ: Здоровье экономики: %.1f", health)
	}

	// Автоматическая стабилизация
	e.autoStabilize()
}

// updatePrice обновляет цену
func (e *BalancedEconomySystem) updatePrice() {
	// В реальной системе здесь будет запрос к Binance API
	// Для примера используем случайную флуктуацию
	change := (math.Sin(float64(time.Now().Unix())/1000.0) * 0.02) // ±2%
	e.currentPrice = e.targetPrice * (1 + change)

	// Добавление в историю
	e.priceHistory = append(e.priceHistory, PricePoint{
		Timestamp: time.Now(),
		Price:     e.currentPrice,
		Volume:    1000000, // Пример
		Supply:    e.currentSupply,
	})

	// Ограничение истории
	if len(e.priceHistory) > 1000 {
		e.priceHistory = e.priceHistory[len(e.priceHistory)-1000:]
	}
}

// autoStabilize автоматически стабилизирует экономику
func (e *BalancedEconomySystem) autoStabilize() {
	priceDeviation := math.Abs(e.currentPrice-e.targetPrice) / e.targetPrice

	// Если отклонение больше 15%, запускаем стабилизацию
	if priceDeviation > 0.15 {
		if e.currentPrice < e.targetPrice {
			// Цена слишком низкая - используем стабфонд
			e.useStabilizationFund()
		} else {
			// Цена слишком высокая - увеличиваем сжигание
			e.increaseBurnRate()
		}
	}
}

// useStabilizationFund использует стабфонд
func (e *BalancedEconomySystem) useStabilizationFund() {
	if e.stabilizationFund <= 0 {
		return
	}

	// Используем 0.1% от стабфонда
	amount := e.stabilizationFund * 0.001
	e.stabilizationFund -= amount
	e.currentSupply -= amount

	log.Printf("Использовано %.2f BKC из стабфонда", amount)
}

// increaseBurnRate увеличивает ставку сжигания
func (e *BalancedEconomySystem) increaseBurnRate() {
	e.burnRate *= 1.1 // Увеличиваем на 10%

	if e.burnRate > 0.1 { // Максимум 10%
		e.burnRate = 0.1
	}

	log.Printf("Увеличена ставка сжигания до %.4f", e.burnRate)
}

// Shutdown останавливает систему
func (e *BalancedEconomySystem) Shutdown() {
	e.cancel()
}
