package economy

import (
	"math"
	"time"
)

// MathematicalModels - математические модели экономики
type MathematicalModels struct {
	// Параметры модели
	alpha float64 // Базовый спрос
	beta  float64 // Эластичность спроса
	gamma float64 // Базовое предложение
	delta float64 // Эластичность предложения
	
	// Макроэкономические параметры
	inflationTarget    float64
	volatilityTarget   float64
	growthTarget       float64
	
	// Калбровочные коэффициенты
	emissionCalibration float64
	burnCalibration     float64
	stakingCalibration  float64
}

// NewMathematicalModels создает новые математические модели
func NewMathematicalModels() *MathematicalModels {
	return &MathematicalModels{
		alpha:              100000,    // Базовый спрос 100K пользователей
		beta:               50000,     // Эластичность спроса
		gamma:              1000000000, // Базовое предложение 1B токенов
		delta:              100000,    // Эластичность предложения
		inflationTarget:    0.03,      // 3% годовая инфляция
		volatilityTarget:   0.15,      // 15% волатильность
		growthTarget:       0.10,      // 10% рост
		emissionCalibration: 0.1,     // Калибровка эмиссии
		burnCalibration:     0.05,    // Калибровка сжигания
		stakingCalibration:  0.15,    // Калибровка стейкинга
	}
}

// SupplyDemandModel - модель спроса и предложения
func (m *MathematicalModels) SupplyDemandModel(price, users, tokens float64) (float64, float64) {
	// Спрос: Qd = α × (U/100000) - β × P
	demand := m.alpha * (users / 100000) - m.beta * price
	
	// Предложение: Qs = γ × (T/1000000000) + δ × P
	supply := m.gamma * (tokens / 1000000000) + m.delta * price
	
	// Ограничения
	if demand < 0 {
		demand = 0
	}
	if supply < 0 {
		supply = 0
	}
	
	return demand, supply
}

// EquilibriumPrice - равновесная цена
func (m *MathematicalModels) EquilibriumPrice(users, tokens float64) float64 {
	// P* = (α × U/100000 - γ × T/1000000000) / (β + δ)
	numerator := m.alpha*(users/100000) - m.gamma*(tokens/1000000000)
	denominator := m.beta + m.delta
	
	if denominator == 0 {
		return 0
	}
	
	price := numerator / denominator
	if price < 0 {
		price = 0
	}
	
	return price
}

// InflationModel - модель инфляции
func (m *MathematicalModels) InflationModel(currentMoneySupply, targetMoneySupply, lastInflation float64) float64 {
	// πt = πt-1 + λ × (Mt - Mt*) / Mt*
	
	lambda := m.emissionCalibration
	moneyGap := (currentMoneySupply - targetMoneySupply) / targetMoneySupply
	
	currentInflation := lastInflation + lambda*moneyGap
	
	// Ограничение инфляции
	if currentInflation > 0.5 { // Максимум 50%
		currentInflation = 0.5
	}
	if currentInflation < -0.5 { // Минимум -50%
		currentInflation = -0.5
	}
	
	return currentInflation
}

// VolatilityModel - модель волатильности
func (m *MathematicalModels) VolatilityModel(prices []float64, window int) float64 {
	if len(prices) < window {
		return 0
	}
	
	// Берем последние window цен
	recent := prices[len(prices)-window:]
	
	// Расчет логарифмических доходностей
	var returns []float64
	for i := 1; i < len(recent); i++ {
		if recent[i-1] > 0 {
			returns = append(returns, math.Log(recent[i]/recent[i-1]))
		}
	}
	
	if len(returns) < 2 {
		return 0
	}
	
	// Расчет среднего и дисперсии
	var sum, sumSq float64
	for _, r := range returns {
		sum += r
		sumSq += r * r
	}
	
	n := float64(len(returns))
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	
	// Годовая волатильность
	volatility := math.Sqrt(variance) * math.Sqrt(365)
	
	return volatility
}

// StakingModel - модель стейкинга
func (m *MathematicalModels) StakingModel(principal, rate, timeYears float64, compoundFrequency int) float64 {
	// Сложный процент: A = P(1 + r/n)^(nt)
	
	if compoundFrequency <= 0 {
		compoundFrequency = 1
	}
	
	amount := principal * math.Pow(1+rate/float64(compoundFrequency), float64(compoundFrequency)*timeYears)
	
	return amount - principal // Только доход
}

// OptimalStakingRate - оптимальная ставка стейкинга
func (m *MathematicalModels) OptimalStakingRate(inflation, volatility, targetAPY float64) float64 {
	// r* = targetAPY + inflation - riskPremium
	
	riskPremium := volatility * 0.5 // Премия за риск
	
	optimalRate := targetAPY + inflation - riskPremium
	
	// Ограничения
	if optimalRate < 0.05 { // Минимум 5%
		optimalRate = 0.05
	}
	if optimalRate > 0.50 { // Максимум 50%
		optimalRate = 0.50
	}
	
	return optimalRate
}

// BurnRateModel - модель ставки сжигания
func (m *MathematicalModels) BurnRateModel(transactionVolume, price, supply float64) float64 {
	// B = V × r × (P/Ptarget) × (S/Smax)
	
	// Базовая ставка
	baseRate := m.burnCalibration
	
	// Корректировка по цене
	priceFactor := price / 0.001 // Целевая цена $0.001
	
	// Корректировка по поставке
	supplyFactor := supply / 1000000000 // Нормализация на 1B
	
	burnRate := baseRate * priceFactor * supplyFactor
	
	// Ограничения
	if burnRate > 0.1 { // Максимум 10%
		burnRate = 0.1
	}
	if burnRate < 0.01 { // Минимум 1%
		burnRate = 0.01
	}
	
	return burnRate
}

// EmissionModel - модель эмиссии
func (m *MathematicalModels) EmissionModel(activeUsers, maxUsers, price, targetPrice, time float64) float64 {
	// E = B × (U/Umax) × (Ptarget/P) × Tfactor
	
	baseEmission := 1000000.0 // Базовая эмиссия 1M BKC в день
	
	// Фактор пользователей
	userFactor := activeUsers / maxUsers
	
	// Фактор цены (обратная зависимость)
	priceFactor := targetPrice / price
	
	// Временной фактор (снижение со временем)
	timeFactor := math.Exp(-time / (365 * 2)) // Половинное снижение за 2 года
	
	emission := baseEmission * userFactor * priceFactor * timeFactor
	
	// Ограничения
	if emission > 5000000 { // Максимум 5M в день
		emission = 5000000
	}
	if emission < 100000 { // Минимум 100K в день
		emission = 100000
	}
	
	return emission
}

// LiquidityModel - модель ликвидности
func (m *MathematicalModels) LiquidityModel(bids, asks, price float64) float64 {
	// LI = (Bids × Asks) / (Price × Volume)
	
	// Расчет объема
	totalVolume := bids + asks
	
	if totalVolume == 0 || price == 0 {
		return 0
	}
	
	// Коэффициент ликвидности
	liquidityIndex := (bids * asks) / (price * totalVolume)
	
	// Нормализация
	normalizedLiquidity := liquidityIndex / 1000000
	
	return normalizedLiquidity
}

// MarketImpactModel - модель влияния на рынок
func (m *MathematicalModels) MarketImpactModel(tradeSize, liquidity, price float64) float64 {
	// ΔP = k × (V/L)^α
	
	// Коэффициент влияния
	k := 0.01
	
	// Эластичность влияния
	alpha := 0.5
	
	// Относительный размер сделки
	relativeSize := tradeSize / liquidity
	
	if relativeSize <= 0 {
		return 0
	}
	
	priceImpact := k * math.Pow(relativeSize, alpha)
	
	return priceImpact
}

// RiskModel - модель риска
func (m *MathematicalModels) RiskModel(volatility, correlation, concentration float64) float64 {
	// Risk = √(σ² + ρ² + c²)
	
	risk := math.Sqrt(volatility*volatility + correlation*correlation + concentration*concentration)
	
	// Нормализация риска (0-1)
	normalizedRisk := risk / 3.0 // Максимум 3.0
	
	if normalizedRisk > 1.0 {
		normalizedRisk = 1.0
	}
	
	return normalizedRisk
}

// UtilityModel - модель полезности
func (m *MathematicalModels) UtilityModel(wealth, riskAversion, expectedReturn float64) float64 {
	// U = E(R) - 0.5 × A × σ²
	
	// Ожидаемая доходность
	expectedReturnPct := expectedReturn / wealth
	
	// Дисперсия (используем волатильность как прокси)
	variance := 0.15 * 0.15 // 15% волатильность
	
	// Функция полезности Карвера-Мехры
	utility := expectedReturnPct - 0.5*riskAversion*variance
	
	return utility
}

// GameTheoryModel - модель теории игр
func (m *MathematicalModels) GameTheoryModel(playerCount, cooperationRate, reward float64) float64 {
	// Nash Equilibrium для кооперативной игры
	
	// Базовая награда
	baseReward := reward
	
	// Бонус за кооперацию
	cooperationBonus := cooperationRate * baseReward * 0.5
	
	// Штраф за большое количество игроков (конкуренция)
	competitionPenalty := math.Log(playerCount) * baseReward * 0.1
	
	// Итоговая награда
	finalReward := baseReward + cooperationBonus - competitionPenalty
	
	if finalReward < 0 {
		finalReward = 0
	}
	
	return finalReward
}

// NetworkEffectModel - модель сетевого эффекта
func (m *MathematicalModels) NetworkEffectModel(userCount, maxUsers, baseValue float64) float64 {
	// Metcalfe's Law: V = n²
	
	// Нормализованное количество пользователей
	normalizedUsers := userCount / maxUsers
	
	// Сетевой эффект
	networkEffect := math.Pow(normalizedUsers, 2)
	
	// Итоговая ценность
	networkValue := baseValue * networkEffect
	
	return networkValue
}

// AdoptionCurveModel - модель кривой принятия
func (m *MathematicalModels) AdoptionCurveModel(time, marketSize, saturationRate float64) float64 {
	// Logistic Growth: P(t) = K / (1 + Ae^(-rt))
	
	// Максимальный размер рынка
	K := marketSize
	
	// Начальное принятие
	A := (K / 1000) - 1 // Начинаем с 0.1% рынка
	
	// Скорость принятия
	r := saturationRate
	
	// Кривая принятия
	adoption := K / (1 + A*math.Exp(-r*time))
	
	return adoption
}

// TokenVelocityModel - модель скорости токена
func (m *MathematicalModels) TokenVelocityModel(transactionVolume, totalSupply, price float64) float64 {
	// V = (P × Q) / M
	
	// Общий объем транзакций в долларах
	totalValue := transactionVolume * price
	
	// Средняя скорость
	velocity := totalValue / (totalSupply * price)
	
	// Годовая скорость
	annualVelocity := velocity * 365
	
	return annualVelocity
}

// PriceDiscoveryModel - модель ценообразования
func (m *MathematicalModels) PriceDiscoveryModel(demand, supply, liquidity, sentiment float64) float64 {
	// P = (D/S) × L × S
	
	// Базовая цена от спроса/предложения
	basePrice := demand / supply
	
	// Корректировка на ликвидность
	liquidityFactor := 1.0 + (liquidity - 0.5) * 0.2
	
	// Корректировка на сентимент
	sentimentFactor := 1.0 + sentiment*0.3
	
	// Итоговая цена
	discoveredPrice := basePrice * liquidityFactor * sentimentFactor
	
	return discoveredPrice
}

// EconomicEquilibrium - экономическое равновесие
func (m *MathematicalModels) EconomicEquilibrium(users, tokens, liquidity, sentiment float64) EquilibriumResult {
	// Расчет равновесной цены
	equilibriumPrice := m.EquilibriumPrice(users, tokens)
	
	// Расчет спроса и предложения
	demand, supply := m.SupplyDemandModel(equilibriumPrice, users, tokens)
	
	// Расчет волатильности
	prices := []float64{equilibriumPrice}
	volatility := m.VolatilityModel(prices, 1)
	
	// Расчет оптимальной ставки стейкинга
	optimalStakingRate := m.OptimalStakingRate(m.inflationTarget, volatility, 0.15)
	
	// Расчет оптимальной эмиссии
	time := float64(time.Now().Unix()) / (365 * 24 * 3600) // Годы
	optimalEmission := m.EmissionModel(users, 1000000, equilibriumPrice, 0.001, time)
	
	// Расчет оптимальной ставки сжигания
	burnRate := m.BurnRateModel(1000000, equilibriumPrice, tokens)
	
	return EquilibriumResult{
		Price:              equilibriumPrice,
		Demand:             demand,
		Supply:             supply,
		Volatility:         volatility,
		OptimalStakingRate: optimalStakingRate,
		OptimalEmission:    optimalEmission,
		OptimalBurnRate:    burnRate,
		HealthScore:        m.calculateHealthScore(equilibriumPrice, volatility, liquidity),
	}
}

// EquilibriumResult - результат равновесия
type EquilibriumResult struct {
	Price              float64 `json:"price"`
	Demand             float64 `json:"demand"`
	Supply             float64 `json:"supply"`
	Volatility         float64 `json:"volatility"`
	OptimalStakingRate float64 `json:"optimal_staking_rate"`
	OptimalEmission    float64 `json:"optimal_emission"`
	OptimalBurnRate    float64 `json:"optimal_burn_rate"`
	HealthScore        float64 `json:"health_score"`
}

// calculateHealthScore рассчитывает здоровье экономики
func (m *MathematicalModels) calculateHealthScore(price, volatility, liquidity float64) float64 {
	score := 100.0
	
	// Штраф за отклонение цены от целевой
	priceDeviation := math.Abs(price-0.001) / 0.001
	score -= priceDeviation * 50
	
	// Штраф за высокую волатильность
	if volatility > m.volatilityTarget {
		score -= (volatility - m.volatilityTarget) * 100
	}
	
	// Штраф за низкую ликвидность
	if liquidity < 0.3 {
		score -= (0.3 - liquidity) * 50
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
