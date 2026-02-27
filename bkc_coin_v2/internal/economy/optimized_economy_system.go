package economy

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// OptimizedEconomySystem - оптимизированная экономическая система
type OptimizedEconomySystem struct {
	mu sync.RWMutex
	
	// Основные параметры
	totalSupply      float64
	currentSupply    float64
	targetPrice      float64
	currentPrice     float64
	
	// Энергетическая система
	maxEnergy        int64
	energyPerHour    int64
	dailyEnergyLimit int64
	tapCost          float64
	tapReward        float64
	dailyMaxCoins    float64
	
	// Реферальная система
	referralRequirement int64
	referralReward     float64
	newUserBonus      float64
	activityThreshold int64
	
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
	maxDailySaleAmount float64
	saleTaxRate       float64
	
	// Динамическая система
	priceElasticity     float64
	userGrowthTarget   float64
	maxUsers           int64
	
	// NFT система
	nftBonusMultiplier float64
	nftBurnReduction   float64
	
	// Исторические данные
	dailyMetrics       []DailyMetrics
	priceHistory       []PricePoint
	
	// Конфигурация
	config OptimizedEconomyConfig
	
	// Контекст
	ctx    context.Context
	cancel context.CancelFunc
}

// OptimizedEconomyConfig - конфигурация оптимизированной экономики
type OptimizedEconomyConfig struct {
	TotalSupply         float64 `json:"total_supply"`
	TargetPrice         float64 `json:"target_price"`
	MaxEnergy          int64   `json:"max_energy"`
	EnergyPerHour      int64   `json:"energy_per_hour"`
	DailyEnergyLimit   int64   `json:"daily_energy_limit"`
	TapCost            float64 `json:"tap_cost"`
	TapReward          float64 `json:"tap_reward"`
	DailyMaxCoins      float64 `json:"daily_max_coins"`
	ReferralRequirement int64   `json:"referral_requirement"`
	ReferralReward     float64 `json:"referral_reward"`
	NewUserBonus       float64 `json:"new_user_bonus"`
	ActivityThreshold  int64   `json:"activity_threshold"`
	DailyEmissionCap  float64 `json:"daily_emission_cap"`
	BurnRate           float64 `json:"burn_rate"`
	StakingAPY         float64 `json:"staking_apy"`
	FundAllocation     float64 `json:"fund_allocation"`
	AdminUnlockRate    float64 `json:"admin_unlock_rate"`
	MaxDailySaleAmount float64 `json:"max_daily_sale_amount"`
	SaleTaxRate        float64 `json:"sale_tax_rate"`
	PriceElasticity    float64 `json:"price_elasticity"`
	UserGrowthTarget   float64 `json:"user_growth_target"`
	MaxUsers           int64   `json:"max_users"`
	NFTBonusMultiplier float64 `json:"nft_bonus_multiplier"`
	NFTBurnReduction   float64 `json:"nft_burn_reduction"`
}

// UserTapResult - результат тапа
type UserTapResult struct {
	Success       bool    `json:"success"`
	NewBalance    float64 `json:"new_balance"`
	NewEnergy     int64   `json:"new_energy"`
	EnergyUsed    int64   `json:"energy_used"`
	TapsMade      int64   `json:"taps_made"`
	TokensEarned  float64 `json:"tokens_earned"`
	Message       string  `json:"message"`
	CanTapMore    bool    `json:"can_tap_more"`
	NextEnergyTime time.Time `json:"next_energy_time"`
	DailyLimitReached bool `json:"daily_limit_reached"`
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

// NewOptimizedEconomySystem создает новую оптимизированную экономическую систему
func NewOptimizedEconomySystem(config OptimizedEconomyConfig) *OptimizedEconomySystem {
	ctx, cancel := context.WithCancel(context.Background())
	
	system := &OptimizedEconomySystem{
		totalSupply:        config.TotalSupply,
		currentSupply:      config.TotalSupply * 0.7, // 70% выпущено
		targetPrice:        config.TargetPrice,
		currentPrice:       config.TargetPrice,
		maxEnergy:         config.MaxEnergy,
		energyPerHour:     config.EnergyPerHour,
		dailyEnergyLimit:   config.DailyEnergyLimit,
		tapCost:           config.TapCost,
		tapReward:         config.TapReward,
		dailyMaxCoins:     config.DailyMaxCoins,
		referralRequirement: config.ReferralRequirement,
		referralReward:     config.ReferralReward,
		newUserBonus:      config.NewUserBonus,
		activityThreshold: config.ActivityThreshold,
		dailyEmissionCap:  config.DailyEmissionCap,
		burnRate:          config.BurnRate,
		stakingAPY:        config.StakingAPY,
		fundAllocation:    config.FundAllocation,
		stabilizationFund: config.TotalSupply * config.FundAllocation,
		adminLockedFunds:   make(map[int64]AdminLock),
		adminUnlockRate:   config.AdminUnlockRate,
		saleProtectionEnabled: true,
		maxDailySaleAmount: config.MaxDailySaleAmount,
		saleTaxRate:       config.SaleTaxRate,
		priceElasticity:    config.PriceElasticity,
		userGrowthTarget:  config.UserGrowthTarget,
		maxUsers:          config.MaxUsers,
		nftBonusMultiplier: config.NFTBonusMultiplier,
		nftBurnReduction:   config.NFTBurnReduction,
		config:            config,
		ctx:               ctx,
		cancel:            cancel,
	}
	
	// Запуск фоновых процессов
	go system.startBackgroundProcesses()
	
	return system
}

// ProcessUserTap обрабатывает тап пользователя с учетом NFT
func (e *OptimizedEconomySystem) ProcessUserTap(userID int64, currentBalance float64, currentEnergy int64, tapsRequested int64, hasNFT bool, nftLevel int) UserTapResult {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	result := UserTapResult{
		Success:    false,
		NewBalance: currentBalance,
		NewEnergy:  currentEnergy,
		TapsMade:   0,
		TokensEarned: 0,
	}
	
	// Проверка дневного лимита монет
	today := time.Now().Truncate(24 * time.Hour)
	dailyEarned := e.getDailyEarned(userID, today)
	if dailyEarned >= e.dailyMaxCoins {
		result.Message = "Достигнут дневной лимит заработка"
		result.DailyLimitReached = true
		return result
	}
	
	// Проверка энергии
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
	
	// Расчет доступных монет до лимита
	remainingCoins := e.dailyMaxCoins - dailyEarned
	maxTapsFromCoins := int64(remainingCoins / e.calculateDynamicTapReward(hasNFT, nftLevel))
	if tapsRequested > maxTapsFromCoins {
		tapsRequested = maxTapsFromCoins
	}
	
	if tapsRequested <= 0 {
		result.Message = "Достигнут дневной лимит монет"
		result.DailyLimitReached = true
		return result
	}
	
	// Расчет награды с учетом NFT
	dynamicReward := e.calculateDynamicTapReward(hasNFT, nftLevel)
	totalReward := float64(tapsRequested) * dynamicReward
	energyCost := float64(tapsRequested) * e.tapCost
	
	// Применение тапа
	result.Success = true
	result.NewBalance = currentBalance + totalReward
	result.NewEnergy = currentEnergy - int64(energyCost)
	result.TapsMade = tapsRequested
	result.TokensEarned = totalReward
	result.EnergyUsed = int64(energyCost)
	result.Message = fmt.Sprintf("Успешно! Заработано %.2f BKC", totalReward)
	result.CanTapMore = result.NewEnergy > 0 && (dailyEarned+totalReward) < e.dailyMaxCoins
	result.NextEnergyTime = time.Now().Add(time.Hour / time.Duration(e.energyPerHour))
	result.DailyLimitReached = (dailyEarned + totalReward) >= e.dailyMaxCoins
	
	// Обновление статистики
	e.updateDailyMetrics(userID, tapsRequested, int64(energyCost), totalReward)
	
	// Обновление предложения
	e.currentSupply += totalReward
	
	log.Printf("User %d tapped %d times, earned %.2f BKC (NFT: %v, Level: %d)", 
		userID, tapsRequested, totalReward, hasNFT, nftLevel)
	
	return result
}

// calculateDynamicTapReward рассчитывает динамическую награду за тап
func (e *OptimizedEconomySystem) calculateDynamicTapReward(hasNFT bool, nftLevel int) float64 {
	// Базовая награда
	baseReward := e.tapReward
	
	// Корректировка по цене
	priceFactor := e.targetPrice / e.currentPrice
	
	// Корректировка по количеству пользователей
	userFactor := e.calculateUserGrowthFactor()
	
	// Корректировка по предложению токенов
	supplyFactor := e.calculateSupplyFactor()
	
	// NFT бонусы
	nftBonus := 1.0
	if hasNFT {
		nftBonus = 1.0 + (float64(nftLevel) * e.nftBonusMultiplier)
	}
	
	// Итоговая награда
	dynamicReward := baseReward * priceFactor * userFactor * supplyFactor * nftBonus
	
	// Ограничения
	if dynamicReward > baseReward*3.0 {
		dynamicReward = baseReward * 3.0 // Максимум +200% с NFT
	}
	if dynamicReward < baseReward*0.1 {
		dynamicReward = baseReward * 0.1 // Минимум -90%
	}
	
	return dynamicReward
}

// calculateSupplyFactor рассчитывает фактор предложения
func (e *OptimizedEconomySystem) calculateSupplyFactor() float64 {
	supplyRatio := e.currentSupply / e.totalSupply
	
	// Чем больше предложение, тем ниже награда
	if supplyRatio < 0.3 {
		return 1.2 // +20% при низком предложении
	} else if supplyRatio < 0.5 {
		return 1.0 // Нормально
	} else if supplyRatio < 0.7 {
		return 0.8 // -20% при высоком предложении
	} else {
		return 0.6 // -40% при очень высоком предложении
	}
}

// ProcessReferral обрабатывает реферала с оптимизированными бонусами
func (e *OptimizedEconomySystem) ProcessReferral(referrerID, referralID int64) ReferralResult {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	result := ReferralResult{
		Success:    false,
		ReferrerID: referrerID,
		ReferralID: referralID,
	}
	
	// Проверка требований
	referralCount := e.getActiveReferralCount(referrerID)
	
	if referralCount < e.referralRequirement {
		result.Message = fmt.Sprintf("Приведите еще %d активных рефералов для получения награды", 
			e.referralRequirement-referralCount)
		return result
	}
	
	// Расчет награды
	rewardAmount := e.referralReward // 250 BKC за 4 реферала
	newUserBonus := e.newUserBonus  // 30 BKC новому пользователю
	
	// Проверка эмиссионного лимита
	today := time.Now().Truncate(24 * time.Hour)
	dailyEmission := e.getDailyEmission(today)
	totalReward := rewardAmount + newUserBonus
	
	if dailyEmission+totalReward > e.dailyEmissionCap {
		availableEmission := e.dailyEmissionCap - dailyEmission
		if totalReward > availableEmission {
			// Пропорциональное сокращение наград
			ratio := availableEmission / totalReward
			rewardAmount *= ratio
			newUserBonus *= ratio
			totalReward = rewardAmount + newUserBonus
		}
	}
	
	result.Success = true
	result.RewardAmount = rewardAmount
	result.BonusAmount = newUserBonus
	result.Message = fmt.Sprintf("Получено %.2f BKC реферальной награды, новому пользователю %.2f BKC", 
		rewardAmount, newUserBonus)
	
	// Обновление статистики
	e.updateReferralMetrics(referrerID, totalReward)
	
	// Обновление предложения
	e.currentSupply += totalReward
	
	log.Printf("Referral: %d -> %d, referrer reward: %.2f BKC, new user bonus: %.2f BKC", 
		referrerID, referralID, rewardAmount, newUserBonus)
	
	return result
}

// getActiveReferralCount получает количество активных рефералов
func (e *OptimizedEconomySystem) getActiveReferralCount(userID int64) int64 {
	// В реальной системе здесь будет запрос к БД с проверкой активности
	// Активность = 1 тап в неделю
	return int64(time.Now().Unix() % 10) // Для примера
}

// getDailyEarned получает заработанное за день
func (e *OptimizedEconomySystem) getDailyEarned(userID int64, date time.Time) float64 {
	// В реальной системе здесь будет запрос к БД
	return float64(date.Unix() % int64(e.dailyMaxCoins))
}

// ProcessSale обрабатывает продажу с оптимизированным сжиганием
func (e *OptimizedEconomySystem) ProcessSale(userID int64, amount float64) (float64, error) {
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
	
	// Оптимизированное сжигание
	burnAmount := e.calculateOptimizedBurnAmount(amount)
	stabilizationAmount := taxAmount - burnAmount
	
	// Применение сжигания
	e.currentSupply -= burnAmount
	e.stabilizationFund += stabilizationAmount
	
	// Обновление статистики
	e.updateSaleMetrics(today, amount)
	
	log.Printf("Sale: user %d, amount %.2f BKC, tax %.2f BKC, burned %.2f BKC, net %.2f BKC", 
		userID, amount, taxAmount, burnAmount, netAmount)
	
	return netAmount, nil
}

// calculateOptimizedBurnAmount рассчитывает оптимизированное количество для сжигания
func (e *OptimizedEconomySystem) calculateOptimizedBurnAmount(saleAmount float64) float64 {
	// Базовая ставка сжигания
	baseBurnRate := e.burnRate
	
	// Корректировка по предложению
	supplyRatio := e.currentSupply / e.totalSupply
	
	// Корректировка по цене
	priceRatio := e.currentPrice / e.targetPrice
	
	// Корректировка по волатильности
	volatility := e.calculateVolatility()
	
	// Итоговая ставка сжигания
	burnRate := baseBurnRate
	
	// Увеличиваем сжигание при высоком предложении
	if supplyRatio > 0.7 {
		burnRate *= 1.5
	}
	
	// Увеличиваем сжигание при низкой цене
	if priceRatio < 0.9 {
		burnRate *= 1.2
	}
	
	// Уменьшаем сжигание при высокой волатильности
	if volatility > 0.15 {
		burnRate *= 0.8
	}
	
	// Ограничения
	if burnRate > 0.05 { // Максимум 5%
		burnRate = 0.05
	}
	if burnRate < 0.001 { // Минимум 0.1%
		burnRate = 0.001
	}
	
	burnAmount := saleAmount * burnRate
	
	// Дополнительное ограничение: не сжигаем больше 0.1% от дневного объема
	maxDailyBurn := e.dailyEmissionCap * 0.001
	if burnAmount > maxDailyBurn {
		burnAmount = maxDailyBurn
	}
	
	return burnAmount
}

// calculateVolatility рассчитывает волатильность
func (e *OptimizedEconomySystem) calculateVolatility() float64 {
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

// GetSystemStatus получает статус системы
func (e *OptimizedEconomySystem) GetSystemStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// Расчет здоровья экономики
	healthScore := e.calculateEconomicHealth()
	
	// Расчет инфляции
	inflationRate := e.calculateInflationRate()
	
	// Расчет волатильности
	volatility := e.calculateVolatility()
	
	// Расчет времени до исчерпания токенов
	timeToExhaustion := e.calculateTimeToExhaustion()
	
	status := map[string]interface{}{
		"total_supply":        e.totalSupply,
		"current_supply":      e.currentSupply,
		"target_price":        e.targetPrice,
		"current_price":       e.currentPrice,
		"max_energy":         e.maxEnergy,
		"energy_per_hour":     e.energyPerHour,
		"daily_energy_limit":  e.dailyEnergyLimit,
		"daily_max_coins":     e.dailyMaxCoins,
		"tap_cost":           e.tapCost,
		"tap_reward":         e.tapReward,
		"referral_requirement": e.referralRequirement,
		"referral_reward":     e.referralReward,
		"new_user_bonus":      e.newUserBonus,
		"daily_emission_cap": e.dailyEmissionCap,
		"burn_rate":          e.burnRate,
		"staking_apy":        e.stakingAPY,
		"stabilization_fund": e.stabilizationFund,
		"health_score":       healthScore,
		"inflation_rate":      inflationRate,
		"volatility":        volatility,
		"sale_protection":    e.saleProtectionEnabled,
		"max_daily_sale":    e.maxDailySaleAmount,
		"sale_tax_rate":      e.saleTaxRate,
		"admin_unlock_rate":  e.adminUnlockRate,
		"max_users":          e.maxUsers,
		"nft_bonus_multiplier": e.nftBonusMultiplier,
		"nft_burn_reduction":  e.nftBurnReduction,
		"time_to_exhaustion": timeToExhaustion,
	}
	
	return status
}

// calculateTimeToExhaustion рассчитывает время до исчерпания токенов
func (e *OptimizedEconomySystem) calculateTimeToExhaustion() string {
	// Текущие темпы эмиссии и сжигания
	dailyEmission := e.dailyEmissionCap
	dailyBurn := dailyEmission * e.burnRate * 0.5 // Усредненное сжигание
	
	netDailyChange := dailyEmission - dailyBurn
	
	if netDailyChange <= 0 {
		return "Бесконечно (чистая дефляция)"
	}
	
	remainingTokens := e.totalSupply - e.currentSupply
	daysToExhaustion := remainingTokens / netDailyChange
	
	if daysToExhaustion > 365*10 {
		return "Более 10 лет"
	} else if daysToExhaustion > 365 {
		years := int(daysToExhaustion / 365)
		return fmt.Sprintf("%d лет", years)
	} else {
		return fmt.Sprintf("%.0f дней", daysToExhaustion)
	}
}

// calculateEconomicHealth рассчитывает здоровье экономики
func (e *OptimizedEconomySystem) calculateEconomicHealth() float64 {
	score := 100.0
	
	// Штраф за отклонение цены
	priceDeviation := math.Abs(e.currentPrice-e.targetPrice) / e.targetPrice
	score -= priceDeviation * 20
	
	// Штраф за высокую инфляцию
	inflation := e.calculateInflationRate()
	if inflation > 0.03 { // 3% в месяц
		score -= (inflation - 0.03) * 100
	}
	
	// Штраф за высокую волатильность
	volatility := e.calculateVolatility()
	if volatility > 0.15 { // 15%
		score -= (volatility - 0.15) * 50
	}
	
	// Бонус за здоровое предложение
	supplyRatio := e.currentSupply / e.totalSupply
	if supplyRatio > 0.4 && supplyRatio < 0.8 {
		score += 15
	}
	
	// Бонус за наличие стабфонда
	if e.stabilizationFund > e.totalSupply*0.05 {
		score += 10
	}
	
	// Бонус за устойчивое время до исчерпания
	remainingTokens := e.totalSupply - e.currentSupply
	if remainingTokens > e.totalSupply*0.3 {
		score += 10
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
func (e *OptimizedEconomySystem) calculateInflationRate() float64 {
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

// Вспомогательные функции (аналогично balanced_economy_system.go)
func (e *OptimizedEconomySystem) calculateUserGrowthFactor() float64 {
	if len(e.dailyMetrics) < 7 {
		return e.userGrowthTarget
	}
	
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
	
	growthFactor := 1.0 + (e.userGrowthTarget - growth)*0.5
	
	if growthFactor > 1.3 {
		growthFactor = 1.3
	}
	if growthFactor < 0.7 {
		growthFactor = 0.7
	}
	
	return growthFactor
}

func (e *OptimizedEconomySystem) getDailyEmission(date time.Time) float64 {
	for _, metric := range e.dailyMetrics {
		if metric.Date.Truncate(24 * time.Hour).Equal(date.Truncate(24 * time.Hour)) {
			return metric.TokensEmitted
		}
	}
	return 0
}

func (e *OptimizedEconomySystem) getDailySales(date time.Time) float64 {
	for _, metric := range e.dailyMetrics {
		if metric.Date.Truncate(24 * time.Hour).Equal(date.Truncate(24 * time.Hour)) {
			return metric.TotalSales
		}
	}
	return 0
}

func (e *OptimizedEconomySystem) updateDailyMetrics(userID int64, taps, energyUsed int64, tokensEarned float64) {
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
			Date:        today,
			ActiveUsers: 1,
			NewUsers:    0,
			TotalTaps:   taps,
			EnergyUsed:  energyUsed,
			TokensEmitted: tokensEarned,
		}
		e.dailyMetrics = append(e.dailyMetrics, *metric)
		
		if len(e.dailyMetrics) > 365 {
			e.dailyMetrics = e.dailyMetrics[len(e.dailyMetrics)-365:]
		}
	} else {
		metric.TotalTaps += taps
		metric.EnergyUsed += energyUsed
		metric.TokensEmitted += tokensEarned
	}
}

func (e *OptimizedEconomySystem) updateReferralMetrics(userID int64, amount float64) {
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

func (e *OptimizedEconomySystem) updateSaleMetrics(date time.Time, amount float64) {
	for i, metric := range e.dailyMetrics {
		if metric.Date.Truncate(24 * time.Hour).Equal(date.Truncate(24 * time.Hour)) {
			e.dailyMetrics[i].TotalSales += amount
			return
		}
	}
}

func (e *OptimizedEconomySystem) startBackgroundProcesses() {
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

func (e *OptimizedEconomySystem) performBackgroundTasks() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.updatePrice()
	
	health := e.calculateEconomicHealth()
	if health < 30 {
		log.Printf("КРИТИЧЕСКО: Здоровье экономики: %.1f", health)
	} else if health < 50 {
		log.Printf("ВНИМАНИЕ: Здоровье экономики: %.1f", health)
	}
	
	e.autoStabilize()
}

func (e *OptimizedEconomySystem) updatePrice() {
	change := (math.Sin(float64(time.Now().Unix())/1000.0) * 0.01) // ±1%
	e.currentPrice = e.targetPrice * (1 + change)
	
	e.priceHistory = append(e.priceHistory, PricePoint{
		Timestamp: time.Now(),
		Price:     e.currentPrice,
		Volume:    1000000,
		Supply:    e.currentSupply,
	})
	
	if len(e.priceHistory) > 1000 {
		e.priceHistory = e.priceHistory[len(e.priceHistory)-1000:]
	}
}

func (e *OptimizedEconomySystem) autoStabilize() {
	priceDeviation := math.Abs(e.currentPrice-e.targetPrice) / e.targetPrice
	
	if priceDeviation > 0.15 {
		if e.currentPrice < e.targetPrice {
			e.useStabilizationFund()
		} else {
			e.increaseBurnRate()
		}
	}
}

func (e *OptimizedEconomySystem) useStabilizationFund() {
	if e.stabilizationFund <= 0 {
		return
	}
	
	amount := e.stabilizationFund * 0.001
	e.stabilizationFund -= amount
	e.currentSupply -= amount
	
	log.Printf("Использовано %.2f BKC из стабфонда", amount)
}

func (e *OptimizedEconomySystem) increaseBurnRate() {
	e.burnRate *= 1.05 // Увеличиваем на 5%
	
	if e.burnRate > 0.03 { // Максимум 3%
		e.burnRate = 0.03
	}
	
	log.Printf("Увеличена ставка сжигания до %.4f", e.burnRate)
}

func (e *OptimizedEconomySystem) Shutdown() {
	e.cancel()
}
