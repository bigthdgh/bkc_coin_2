package config

// Новая конфигурация системы тапов
const (
	// Базовая экономика (1 млрд коинов)
	TOTAL_SUPPLY         = 1000000000 // 1 миллиард BKC
	ADMIN_ALLOCATION_PCT = 10         // 10% админская аллокация
	BURN_TAX_PCT         = 10         // 10% налог на сжигание
	HALVING_THRESHOLD    = 100000000  // 100 миллионов для халвинга

	// Новая система тапов (дробные монеты)
	TAP_REWARD       = 0.1   // 1 тап = 0.1 BKC
	MAX_DAILY_TAPS   = 3000  // Максимум 3000 тапов в день
	MAX_DAILY_REWARD = 300.0 // Максимум 300 BKC в день

	// Энергия
	DEFAULT_ENERGY_MAX   = 1000 // Бак на 1000 единиц
	DEFAULT_ENERGY_REGEN = 2    // 2 энергии в секунду (полный бак за 8.3 минуты)
	ENERGY_PER_TAP       = 1    // 1 тап = 1 энергия

	// NFT множители и бонусы (будут настраиваться через админ панель)
	DEFAULT_TAP_MULTIPLIER = 1.0 // Базовый множитель
	DEFAULT_ENERGY_BONUS   = 0   // Базовый бонус энергии
	DEFAULT_DAILY_BONUS    = 0   // Базовый бонус дневных тапов
)

// NFT конфигурация
type NFTConfig struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	PriceBKC       float64 `json:"price_bkc"`
	PriceTON       float64 `json:"price_ton"`
	TapMultiplier  float64 `json:"tap_multiplier"`
	EnergyBonus    int     `json:"energy_bonus"`
	DailyTapBonus  int     `json:"daily_tap_bonus"`
	IsActive       bool    `json:"is_active"`
	CreatedByAdmin int64   `json:"created_by_admin"`
	CreatedAt      string  `json:"created_at"`
}

// Получить множитель тапов для пользователя (из базы данных)
func GetUserTapMultiplier(userNFTs []string) float64 {
	// TODO: Загружать множители из базы данных по NFT ID
	// Временно возвращаем базовый множитель
	return DEFAULT_TAP_MULTIPLIER
}

// Получить бонус к максимальной энергии (из базы данных)
func GetUserEnergyBonus(userNFTs []string) int {
	// TODO: Загружать бонусы из базы данных по NFT ID
	// Временно возвращаем базовый бонус
	return DEFAULT_ENERGY_BONUS
}

// Получить бонус к дневному лимиту тапов (из базы данных)
func GetUserDailyTapBonus(userNFTs []string) int {
	// TODO: Загружать бонусы из базы данных по NFT ID
	// Временно возвращаем базовый бонус
	return DEFAULT_DAILY_BONUS
}

// Расчет награды за тап
func CalculateTapReward(baseReward float64, multiplier float64) float64 {
	return baseReward * multiplier
}

// Проверка дневного лимита
func CheckDailyLimit(currentTaps int, userBonus int) bool {
	maxTaps := MAX_DAILY_TAPS + userBonus
	return currentTaps < maxTaps
}

// Расчет времени восстановления энергии
func CalculateEnergyRegenTime(currentEnergy, maxEnergy int) int {
	needed := maxEnergy - currentEnergy
	secondsNeeded := needed / DEFAULT_ENERGY_REGEN
	return secondsNeeded
}
