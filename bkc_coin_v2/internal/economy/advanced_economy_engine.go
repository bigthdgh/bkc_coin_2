package economy

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// AdvancedEconomyEngine - продвинутый экономический движок
type AdvancedEconomyEngine struct {
	mu sync.RWMutex
	
	// Основные параметры
	totalSupply        float64
	currentSupply      float64
	targetPrice        float64
	currentPrice       float64
	
	// Экономические метрики
	activeUsers        int64
	totalTransactions  float64
	burnedTokens       float64
	stakedTokens       float64
	
	// Настройки экономики
	baseEmission       float64
	maxDailyEmission   float64
	minDailyEmission   float64
	burnRate           float64
	stakingAPY         float64
	
	// Стабилизационный фонд
	stabilizationFund  float64
	fundAllocation     float64
	
	// Исторические данные
	priceHistory       []PricePoint
	emissionHistory    []EmissionPoint
	userActivity       []ActivityPoint
	
	// Конфигурация
	config             EconomyConfig
	
	// Контекст
	ctx                context.Context
	cancel             context.CancelFunc
}

// EconomyConfig - конфигурация экономики
type EconomyConfig struct {
	TotalSupply       float64  `json:"total_supply"`
	TargetPrice       float64  `json:"target_price"`
	BaseEmission      float64  `json:"base_emission"`
	MaxDailyEmission  float64  `json:"max_daily_emission"`
	MinDailyEmission  float64  `json:"min_daily_emission"`
	BurnRate          float64  `json:"burn_rate"`
	StakingAPY        float64  `json:"staking_apy"`
	FundAllocation    float64  `json:"fund_allocation"`
	PriceThreshold    float64  `json:"price_threshold"`
	VolatilityLimit   float64  `json:"volatility_limit"`
}

// PricePoint - точка цены
type PricePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Price     float64    `json:"price"`
	Volume    float64    `json:"volume"`
	Supply    float64    `json:"supply"`
}

// EmissionPoint - точка эмиссии
type EmissionPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Amount    float64    `json:"amount"`
	Source    string     `json:"source"`
	Reason    string     `json:"reason"`
}

// ActivityPoint - точка активности
type ActivityPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	ActiveUsers   int64     `json:"active_users"`
	Transactions  float64   `json:"transactions"`
	Volume        float64   `json:"volume"`
}

// EconomicMetrics - экономические метрики
type EconomicMetrics struct {
	CurrentPrice       float64   `json:"current_price"`
	TargetPrice        float64   `json:"target_price"`
	CurrentSupply      float64   `json:"current_supply"`
	TotalSupply        float64   `json:"total_supply"`
	InflationRate      float64   `json:"inflation_rate"`
	DeflationRate      float64   `json:"deflation_rate"`
	Volatility         float64   `json:"volatility"`
	StakingAPY         float64   `json:"staking_apy"`
	BurnRate           float64   `json:"burn_rate"`
	ActiveUsers        int64     `json:"active_users"`
	MarketCap          float64   `json:"market_cap"`
	TradingVolume      float64   `json:"trading_volume"`
	StabilizationFund  float64   `json:"stabilization_fund"`
	HealthScore        float64   `json:"health_score"`
}

// NewAdvancedEconomyEngine создает новый экономический движок
func NewAdvancedEconomyEngine(config EconomyConfig) *AdvancedEconomyEngine {
	ctx, cancel := context.WithCancel(context.Background())
	
	engine := &AdvancedEconomyEngine{
		totalSupply:       config.TotalSupply,
		currentSupply:     config.TotalSupply * 0.7, // 70% выпущено
		targetPrice:       config.TargetPrice,
		currentPrice:      config.TargetPrice,
		baseEmission:      config.BaseEmission,
		maxDailyEmission:  config.MaxDailyEmission,
		minDailyEmission:  config.MinDailyEmission,
		burnRate:          config.BurnRate,
		stakingAPY:        config.StakingAPY,
		fundAllocation:    config.FundAllocation,
		stabilizationFund: config.TotalSupply * config.FundAllocation,
		config:           config,
		ctx:              ctx,
		cancel:           cancel,
	}
	
	// Запуск фоновых процессов
	go engine.startBackgroundProcesses()
	
	return engine
}

// CalculateDailyEmission рассчитывает дневную эмиссию
func (e *AdvancedEconomyEngine) CalculateDailyEmission() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// Базовая формула эмиссии
	// E = B × (U / Umax) × (P / Ptarget) × Vfactor
	
	userFactor := float64(e.activeUsers) / 1000000.0 // Нормализация на 1M пользователей
	priceFactor := e.targetPrice / e.currentPrice
	volatilityFactor := e.calculateVolatilityFactor()
	
	emission := e.baseEmission * userFactor * priceFactor * volatilityFactor
	
	// Ограничение эмиссии
	if emission > e.maxDailyEmission {
		emission = e.maxDailyEmission
	}
	if emission < e.minDailyEmission {
		emission = e.minDailyEmission
	}
	
	return emission
}

// calculateVolatilityFactor рассчитывает фактор волатильности
func (e *AdvancedEconomyEngine) calculateVolatilityFactor() float64 {
	if len(e.priceHistory) < 2 {
		return 1.0
	}
	
	// Расчет волатильности за последние 7 дней
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
	
	// Фактор волатильности: чем выше волатильность, тем ниже эмиссия
	if volatility > e.config.VolatilityLimit {
		return 0.8 // Снижаем эмиссию на 20%
	}
	
	return 1.0
}

// CalculateBurnAmount рассчитывает количество токенов для сжигания
func (e *AdvancedEconomyEngine) CalculateBurnAmount(transactionVolume float64) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// Формула сжигания: B = V × r × (P / Ptarget)
	burnAmount := transactionVolume * e.burnRate * (e.currentPrice / e.targetPrice)
	
	// Ограничение сжигания
	maxBurn := e.currentSupply * 0.01 // Максимум 1% от текущей поставки
	if burnAmount > maxBurn {
		burnAmount = maxBurn
	}
	
	return burnAmount
}

// UpdatePrice обновляет цену токена
func (e *AdvancedEconomyEngine) UpdatePrice(newPrice, volume float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.currentPrice = newPrice
	
	// Добавление в историю
	e.priceHistory = append(e.priceHistory, PricePoint{
		Timestamp: time.Now(),
		Price:     newPrice,
		Volume:    volume,
		Supply:    e.currentSupply,
	})
	
	// Ограничение истории
	if len(e.priceHistory) > 1000 {
		e.priceHistory = e.priceHistory[len(e.priceHistory)-1000:]
	}
	
	// Автоматическая стабилизация
	e.autoStabilize()
}

// autoStabilize автоматически стабилизирует цену
func (e *AdvancedEconomyEngine) autoStabilize() {
	priceDeviation := math.Abs(e.currentPrice-e.targetPrice) / e.targetPrice
	
	// Если отклонение больше 20%, запускаем стабилизацию
	if priceDeviation > 0.2 {
		if e.currentPrice < e.targetPrice {
			// Цена слишком низкая - покупаем из стабфонда
			e.buyFromStabilizationFund()
		} else {
			// Цена слишком высокая - продаем в стабфонд
			e.sellToStabilizationFund()
		}
	}
}

// buyFromStabilizationFund покупает токены из стабфонда
func (e *AdvancedEconomyEngine) buyFromStabilizationFund() {
	if e.stabilizationFund <= 0 {
		return
	}
	
	// Покупаем 1% от стабфонда
	buyAmount := e.stabilizationFund * 0.01
	e.stabilizationFund -= buyAmount
	e.currentSupply -= buyAmount
	
	log.Printf("Покупка из стабфонда: %.2f BKC", buyAmount)
}

// sellToStabilizationFund продает токены в стабфонд
func (e *AdvancedEconomyEngine) sellToStabilizationFund() {
	// Продаем 0.5% от текущей поставки
	sellAmount := e.currentSupply * 0.005
	e.stabilizationFund += sellAmount
	e.currentSupply += sellAmount
	
	log.Printf("Продажа в стабфонд: %.2f BKC", sellAmount)
}

// ProcessTransaction обрабатывает транзакцию
func (e *AdvancedEconomyEngine) ProcessTransaction(amount float64, txType string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.totalTransactions += amount
	
	// Сжигание токенов
	if txType == "transfer" || txType == "market" || txType == "staking" {
		burnAmount := e.CalculateBurnAmount(amount)
		e.currentSupply -= burnAmount
		e.burnedTokens += burnAmount
		
		log.Printf("Сожжено токенов: %.2f BKC", burnAmount)
	}
	
	// Обновление активности
	e.updateActivity()
	
	return nil
}

// updateActivity обновляет данные об активности
func (e *AdvancedEconomyEngine) updateActivity() {
	e.userActivity = append(e.userActivity, ActivityPoint{
		Timestamp:    time.Now(),
		ActiveUsers:  e.activeUsers,
		Transactions: e.totalTransactions,
		Volume:       e.calculateDailyVolume(),
	})
	
	// Ограничение истории
	if len(e.userActivity) > 100 {
		e.userActivity = e.userActivity[len(e.userActivity)-100:]
	}
}

// calculateDailyVolume рассчитывает дневной объем
func (e *AdvancedEconomyEngine) calculateDailyVolume() float64 {
	if len(e.userActivity) == 0 {
		return 0
	}
	
	today := time.Now().Truncate(24 * time.Hour)
	var volume float64
	
	for _, activity := range e.userActivity {
		if activity.Timestamp.After(today) {
			volume += activity.Volume
		}
	}
	
	return volume
}

// CalculateStakingRewards рассчитывает награды за стейкинг
func (e *AdvancedEconomyEngine) CalculateStakingRewards(stakedAmount, lockPeriodDays float64) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// Базовая годовая ставка
	baseAPY := e.stakingAPY
	
	// Бонус за период блокировки
	lockBonus := 1.0 + (lockPeriodDays/365.0)*0.5 // Максимум +50%
	
	// Бонус за сумму стейкинга
	amountBonus := 1.0
	if stakedAmount >= 1000000 { // 1M+ BKC
		amountBonus = 1.25 // +25%
	} else if stakedAmount >= 100000 { // 100K+ BKC
		amountBonus = 1.15 // +15%
	} else if stakedAmount >= 10000 { // 10K+ BKC
		amountBonus = 1.10 // +10%
	}
	
	// Итоговая годовая ставка
	finalAPY := baseAPY * lockBonus * amountBonus
	
	// Расчет награды за период
	reward := stakedAmount * finalAPY * (lockPeriodDays / 365.0)
	
	return reward
}

// GetMetrics возвращает экономические метрики
func (e *AdvancedEconomyEngine) GetMetrics() EconomicMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	metrics := EconomicMetrics{
		CurrentPrice:      e.currentPrice,
		TargetPrice:       e.targetPrice,
		CurrentSupply:     e.currentSupply,
		TotalSupply:       e.totalSupply,
		StakingAPY:        e.stakingAPY,
		BurnRate:          e.burnRate,
		ActiveUsers:       e.activeUsers,
		MarketCap:         e.currentSupply * e.currentPrice,
		TradingVolume:     e.calculateDailyVolume(),
		StabilizationFund: e.stabilizationFund,
	}
	
	// Расчет инфляции/дефляции
	metrics.InflationRate = e.calculateInflationRate()
	metrics.DeflationRate = e.calculateDeflationRate()
	
	// Расчет волатильности
	metrics.Volatility = e.calculateCurrentVolatility()
	
	// Расчет здоровья экономики
	metrics.HealthScore = e.calculateHealthScore()
	
	return metrics
}

// calculateInflationRate рассчитывает инфляцию
func (e *AdvancedEconomyEngine) calculateInflationRate() float64 {
	if len(e.emissionHistory) < 2 {
		return 0
	}
	
	recent := e.emissionHistory[len(e.emissionHistory)-1]
	previous := e.emissionHistory[len(e.emissionHistory)-2]
	
	dailyEmission := recent.Amount - previous.Amount
	annualEmission := dailyEmission * 365
	
	inflationRate := (annualEmission / e.currentSupply) * 100
	return inflationRate
}

// calculateDeflationRate рассчитывает дефляцию
func (e *AdvancedEconomyEngine) calculateDeflationRate() float64 {
	if len(e.emissionHistory) < 2 {
		return 0
	}
	
	recent := e.emissionHistory[len(e.emissionHistory)-1]
	previous := e.emissionHistory[len(e.emissionHistory)-2]
	
	dailyBurn := (previous.Amount - recent.Amount) - (recent.Amount - previous.Amount)
	annualBurn := dailyBurn * 365
	
	deflationRate := (annualBurn / e.currentSupply) * 100
	return deflationRate
}

// calculateCurrentVolatility рассчитывает текущую волатильность
func (e *AdvancedEconomyEngine) calculateCurrentVolatility() float64 {
	if len(e.priceHistory) < 2 {
		return 0
	}
	
	// Волатильность за последние 24 часа
	recent := e.priceHistory
	cutoff := time.Now().Add(-24 * time.Hour)
	
	var prices []float64
	for _, point := range recent {
		if point.Timestamp.After(cutoff) {
			prices = append(prices, point.Price)
		}
	}
	
	if len(prices) < 2 {
		return 0
	}
	
	var sum, sumSq float64
	for i := 1; i < len(prices); i++ {
		returns := (prices[i] - prices[i-1]) / prices[i-1]
		sum += returns
		sumSq += returns * returns
	}
	
	n := float64(len(prices) - 1)
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	volatility := math.Sqrt(variance) * 100 // В процентах
	
	return volatility
}

// calculateHealthScore рассчитывает здоровье экономики
func (e *AdvancedEconomyEngine) calculateHealthScore() float64 {
	score := 100.0
	
	// Штраф за отклонение цены
	priceDeviation := math.Abs(e.currentPrice-e.targetPrice) / e.targetPrice
	score -= priceDeviation * 50
	
	// Штраф за высокую волатильность
	volatility := e.calculateCurrentVolatility()
	if volatility > 20 {
		score -= (volatility - 20) * 2
	}
	
	// Штраф за низкую ликвидность
	if e.stabilizationFund < e.totalSupply*0.05 {
		score -= 20
	}
	
	// Бонус за здоровую инфляцию
	inflation := e.calculateInflationRate()
	if inflation > 0 && inflation < 5 {
		score += 5
	}
	
	// Бонус за высокий стейкинг
	stakingRatio := e.stakedTokens / e.currentSupply
	if stakingRatio > 0.3 {
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

// startBackgroundProcesses запускает фоновые процессы
func (e *AdvancedEconomyEngine) startBackgroundProcesses() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.updateMetrics()
		}
	}
}

// updateMetrics обновляет метрики
func (e *AdvancedEconomyEngine) updateMetrics() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// Обновление эмиссии
	dailyEmission := e.CalculateDailyEmission()
	e.emissionHistory = append(e.emissionHistory, EmissionPoint{
		Timestamp: time.Now(),
		Amount:    dailyEmission,
		Source:    "daily_emission",
		Reason:    "automated",
	})
	
	// Ограничение истории
	if len(e.emissionHistory) > 100 {
		e.emissionHistory = e.emissionHistory[len(e.emissionHistory)-100:]
	}
	
	// Логирование здоровья экономики
	health := e.calculateHealthScore()
	if health < 50 {
		log.Printf("ВНИМАНИЕ: Здоровье экономики: %.1f", health)
	}
}

// UpdateActiveUsers обновляет количество активных пользователей
func (e *AdvancedEconomyEngine) UpdateActiveUsers(count int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.activeUsers = count
}

// UpdateStakedTokens обновляет количество застейканых токенов
func (e *AdvancedEconomyEngine) UpdateStakedTokens(amount float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stakedTokens = amount
}

// Shutdown останавливает движок
func (e *AdvancedEconomyEngine) Shutdown() {
	e.cancel()
}

// GetEconomicRecommendations возвращает экономические рекомендации
func (e *AdvancedEconomyEngine) GetEconomicRecommendations() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	var recommendations []string
	metrics := e.GetMetrics()
	
	// Рекомендации по цене
	if metrics.CurrentPrice < metrics.TargetPrice*0.8 {
		recommendations = append(recommendations, "Цена слишком низкая. Рекомендуется увеличить эмиссию на 20%")
	} else if metrics.CurrentPrice > metrics.TargetPrice*1.2 {
		recommendations = append(recommendations, "Цена слишком высокая. Рекомендуется уменьшить эмиссию на 20%")
	}
	
	// Рекомендации по волатильности
	if metrics.Volatility > 20 {
		recommendations = append(recommendations, "Высокая волатильность. Рекомендуется увеличить стабфонд")
	}
	
	// Рекомендации по инфляции
	if metrics.InflationRate > 10 {
		recommendations = append(recommendations, "Высокая инфляция. Рекомендуется увеличить ставку сжигания")
	}
	
	// Рекомендации по стейкингу
	if metrics.StakingAPY < 10 {
		recommendations = append(recommendations, "Низкая доходность стейкинга. Рекомендуется увеличить APY")
	}
	
	// Рекомендации по здоровью
	if metrics.HealthScore < 50 {
		recommendations = append(recommendations, "Низкое здоровье экономики. Требуются срочные меры")
	}
	
	return recommendations
}
