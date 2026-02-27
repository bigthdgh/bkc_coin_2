package config

import "fmt"

// ==========================================
// 🪙 ГЛОБАЛЬНАЯ ТОКЕНОМИКА (ЖЕЛЕЗНЫЕ ПРАВИЛА)
// ==========================================

// Жесткие константы - НЕЛЬЗЯ МЕНЯТЬ!
const (
	// ОБЩИЙ ОБЪЕМ - 1 МИЛЛИАРД BKC (ЖЕЛЕЗНОЕ ПРАВИЛО)
	TOTAL_GLOBAL_SUPPLY = 1000000000 // 1,000,000,000 BKC

	// РАСПРЕДЕЛЕНИЕ ПУЛОВ (ЖЕСТКОЕ)
	MINING_POOL_PERCENT = 70 // 70% = 700,000,000 BKC
	ADMIN_VAULT_PERCENT = 30 // 30% = 300,000,000 BKC

	// РАСЧЕТНЫЕ ЗНАЧЕНИЯ
	MINING_POOL_AMOUNT = TOTAL_GLOBAL_SUPPLY * MINING_POOL_PERCENT / 100 // 700,000,000 BKC
	ADMIN_VAULT_AMOUNT = TOTAL_GLOBAL_SUPPLY * ADMIN_VAULT_PERCENT / 100 // 300,000,000 BKC

	// АДМИНСКИЙ РАСПРЕДЕЛ ИЗ VAULT
	LIQUIDITY_ALLOCATION = 15 // 15% на листинг биржи
	MARKETING_ALLOCATION = 10 // 10% на маркетинг и блогеров
	TEAM_ALLOCATION      = 5  // 5% команде

	// КОНКРЕТНЫЕ СУММЫ
	LIQUIDITY_AMOUNT = ADMIN_VAULT_AMOUNT * LIQUIDITY_ALLOCATION / 100 // 45,000,000 BKC
	MARKETING_AMOUNT = ADMIN_VAULT_AMOUNT * MARKETING_ALLOCATION / 100 // 30,000,000 BKC
	TEAM_AMOUNT      = ADMIN_VAULT_AMOUNT * TEAM_ALLOCATION / 100      // 15,000,000 BKC
)

// Глобальное состояние экономики
type GlobalEconomyState struct {
	TotalSupply      int64 `json:"total_supply"`       // 1,000,000,000
	MiningPoolAmount int64 `json:"mining_pool_amount"` // 700,000,000
	AdminVaultAmount int64 `json:"admin_vault_amount"` // 300,000,000

	// Текущее состояние пулов
	MiningPoolMined  int64 `json:"mining_pool_mined"`  // Сколько добыто из пула
	AdminVaultLocked int64 `json:"admin_vault_locked"` // Сколько заморожено в админском хранилище

	// Распределение админского пула
	LiquidityAllocated int64 `json:"liquidity_allocated"` // 45,000,000
	MarketingAllocated int64 `json:"marketing_allocated"` // 30,000,000
	TeamAllocated      int64 `json:"team_allocated"`      // 15,000,000

	// Текущие балансы
	UsersCirculating int64 `json:"users_circulating"` // BKC на руках у юзеров
	TotalBurned      int64 `json:"total_burned"`      // Всего сожжено

	LastUpdated string `json:"last_updated"`
}

// Инициализация глобальной экономики
func InitializeGlobalEconomy() *GlobalEconomyState {
	return &GlobalEconomyState{
		TotalSupply:        TOTAL_GLOBAL_SUPPLY,
		MiningPoolAmount:   MINING_POOL_AMOUNT,
		AdminVaultAmount:   ADMIN_VAULT_AMOUNT,
		MiningPoolMined:    0,
		AdminVaultLocked:   ADMIN_VAULT_AMOUNT,
		LiquidityAllocated: LIQUIDITY_AMOUNT,
		MarketingAllocated: MARKETING_AMOUNT,
		TeamAllocated:      TEAM_AMOUNT,
		UsersCirculating:   0,
		TotalBurned:        0,
		LastUpdated:        "2024-01-15 00:00:00",
	}
}

// Проверка возможности выдачи награды
func CanMintReward(amount int64, currentMined int64) bool {
	availableInPool := MINING_POOL_AMOUNT - currentMined
	return amount <= availableInPool
}

// Расчет оставшихся в пуле добычи
func GetRemainingMiningPool(currentMined int64) int64 {
	return MINING_POOL_AMOUNT - currentMined
}

// Проверка общего предложения (не должно превышать 1 млрд)
func ValidateTotalSupply(usersBalance, burned, adminVault, miningPool int64) bool {
	total := usersBalance + burned + adminVault + miningPool
	return total <= TOTAL_GLOBAL_SUPPLY
}

// Получение процента добычи из пула
func GetMiningPoolProgress(currentMined int64) float64 {
	return float64(currentMined) / float64(MINING_POOL_AMOUNT) * 100
}

// Проверка админского распределения
func GetAdminVaultAllocationStatus(liquidity, marketing, team int64) map[string]interface{} {
	return map[string]interface{}{
		"allocated": map[string]interface{}{
			"liquidity": LIQUIDITY_AMOUNT,
			"marketing": MARKETING_AMOUNT,
			"team":      TEAM_AMOUNT,
		},
		"used": map[string]interface{}{
			"liquidity": liquidity,
			"marketing": marketing,
			"team":      team,
		},
		"remaining": map[string]interface{}{
			"liquidity": LIQUIDITY_AMOUNT - liquidity,
			"marketing": MARKETING_AMOUNT - marketing,
			"team":      TEAM_AMOUNT - team,
		},
	}
}

// Форматирование больших чисел
func FormatBKC(amount int64) string {
	if amount >= 1000000000 {
		return fmt.Sprintf("%.1f млрд", float64(amount)/1000000000)
	} else if amount >= 1000000 {
		return fmt.Sprintf("%.1f млн", float64(amount)/1000000)
	} else if amount >= 1000 {
		return fmt.Sprintf("%.1f тыс", float64(amount)/1000)
	}
	return fmt.Sprintf("%d", amount)
}

// Валидация транзакции с учетом глобальных лимитов
func ValidateTransactionWithGlobalLimits(fromBalance, toBalance, amount int64, isBurn bool) error {
	// Проверяем что общее предложение не превысит лимит
	if isBurn {
		// При сжигании общее предложение уменьшается - это ок
		return nil
	}

	// При переводе между юзерами общее предложение не меняется
	// Но нужно проверить что у отправителя достаточно средств
	if fromBalance < amount {
		return fmt.Errorf("insufficient balance")
	}

	return nil
}

// Получение статистики глобальной экономики
func GetGlobalEconomyStats() map[string]interface{} {
	return map[string]interface{}{
		"total_supply": TOTAL_GLOBAL_SUPPLY,
		"mining_pool": map[string]interface{}{
			"total":      MINING_POOL_AMOUNT,
			"percentage": MINING_POOL_PERCENT,
			"formatted":  FormatBKC(MINING_POOL_AMOUNT),
		},
		"admin_vault": map[string]interface{}{
			"total":      ADMIN_VAULT_AMOUNT,
			"percentage": ADMIN_VAULT_PERCENT,
			"formatted":  FormatBKC(ADMIN_VAULT_AMOUNT),
			"allocations": map[string]interface{}{
				"liquidity": map[string]interface{}{
					"amount":     LIQUIDITY_AMOUNT,
					"percentage": LIQUIDITY_ALLOCATION,
					"formatted":  FormatBKC(LIQUIDITY_AMOUNT),
				},
				"marketing": map[string]interface{}{
					"amount":     MARKETING_AMOUNT,
					"percentage": MARKETING_ALLOCATION,
					"formatted":  FormatBKC(MARKETING_AMOUNT),
				},
				"team": map[string]interface{}{
					"amount":     TEAM_AMOUNT,
					"percentage": TEAM_ALLOCATION,
					"formatted":  FormatBKC(TEAM_AMOUNT),
				},
			},
		},
		"rules": map[string]interface{}{
			"hard_cap":            "1,000,000,000 BKC (нельзя изменить)",
			"mining_distribution": "70% юзерам через тапы",
			"admin_distribution":  "30% админскому фонду",
			"no_inflation":        "Дополнительная эмиссия запрещена",
		},
	}
}
