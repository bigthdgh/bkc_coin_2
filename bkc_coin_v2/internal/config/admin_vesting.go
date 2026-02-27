package config

import (
	"fmt"
	"time"
)

// ==========================================
// 💰 АДМИНСКАЯ ВЕСТИНГ (ТВОИ 300МЛН BKC)
// ==========================================

const (
	// Твои параметры вестинга
	ADMIN_VESTING_TOTAL_AMOUNT      = 300000000 // 300 млн BKC
	ADMIN_VESTING_PERIOD_MONTHS     = 5         // 5 месяцев
	ADMIN_VESTING_PERCENT_PER_MONTH = 10.0      // 10% в месяц

	// Расчетные значения
	ADMIN_VESTING_MONTHLY_AMOUNT = int64(float64(ADMIN_VESTING_TOTAL_AMOUNT) * ADMIN_VESTING_PERCENT_PER_MONTH / 100) // 30 млн BKC в месяц
	ADMIN_VESTING_START_DATE     = "2024-01-15"                                                                       // Дата начала вестинга
)

// Структура вестинга
type AdminVestingSchedule struct {
	TotalAmount     int64   `json:"total_amount"`      // 300,000,000 BKC
	PeriodMonths    int     `json:"period_months"`     // 5 месяцев
	PercentPerMonth float64 `json:"percent_per_month"` // 10% в месяц
	MonthlyAmount   int64   `json:"monthly_amount"`    // 30,000,000 BKC
	StartDate       string  `json:"start_date"`        // Дата начала
	EndDate         string  `json:"end_date"`          // Дата окончания

	// Текущее состояние
	CurrentMonth     int    `json:"current_month"`      // Текущий месяц (1-5)
	TotalWithdrawn   int64  `json:"total_withdrawn"`    // Всего выведено
	RemainingAmount  int64  `json:"remaining_amount"`   // Осталось вывести
	NextWithdrawDate string `json:"next_withdraw_date"` // Следующая дата вывода
	IsCompleted      bool   `json:"is_completed"`       // Завершен?
	LastWithdrawDate string `json:"last_withdraw_date"` // Последний вывод
}

// Инициализация вестинга
func InitializeAdminVesting() *AdminVestingSchedule {
	startDate, _ := time.Parse("2006-01-02", ADMIN_VESTING_START_DATE)
	endDate := startDate.AddDate(0, ADMIN_VESTING_PERIOD_MONTHS, 0)

	return &AdminVestingSchedule{
		TotalAmount:      ADMIN_VESTING_TOTAL_AMOUNT,
		PeriodMonths:     ADMIN_VESTING_PERIOD_MONTHS,
		PercentPerMonth:  ADMIN_VESTING_PERCENT_PER_MONTH,
		MonthlyAmount:    ADMIN_VESTING_MONTHLY_AMOUNT,
		StartDate:        ADMIN_VESTING_START_DATE,
		EndDate:          endDate.Format("2006-01-02"),
		CurrentMonth:     1,
		TotalWithdrawn:   0,
		RemainingAmount:  ADMIN_VESTING_TOTAL_AMOUNT,
		NextWithdrawDate: startDate.AddDate(0, 1, 0).Format("2006-01-02"),
		IsCompleted:      false,
		LastWithdrawDate: "",
	}
}

// Получение текущего статуса вестинга
func GetAdminVestingStatus() *AdminVestingSchedule {
	vesting := InitializeAdminVesting()

	// TODO: Получить реальные данные из базы
	// dbVesting := getAdminVestingFromDB()
	// if dbVesting != nil {
	//     return dbVesting
	// }

	// Расчет текущего месяца
	now := time.Now()
	startDate, _ := time.Parse("2006-01-02", vesting.StartDate)
	monthsPassed := calculateMonthsPassed(startDate, now)

	if monthsPassed >= ADMIN_VESTING_PERIOD_MONTHS {
		vesting.CurrentMonth = ADMIN_VESTING_PERIOD_MONTHS
		vesting.IsCompleted = true
		vesting.RemainingAmount = 0
	} else {
		vesting.CurrentMonth = monthsPassed + 1
		vesting.TotalWithdrawn = int64(monthsPassed) * ADMIN_VESTING_MONTHLY_AMOUNT
		vesting.RemainingAmount = ADMIN_VESTING_TOTAL_AMOUNT - vesting.TotalWithdrawn

		// Расчет следующей даты вывода
		nextDate := time.Date(now.Year(), now.Month()+1, 15, 0, 0, 0, 0, time.UTC)
		vesting.NextWithdrawDate = nextDate.Format("2006-01-02")
	}

	return vesting
}

// Проверка возможности вывода
func CanWithdrawAmount(amount int64, vesting *AdminVestingSchedule) (bool, string) {
	if vesting.IsCompleted {
		return false, "Vesting already completed"
	}

	if amount > vesting.RemainingAmount {
		return false, fmt.Sprintf("Insufficient vested amount. Available: %d BKC", vesting.RemainingAmount)
	}

	// Проверяем что наступила дата вывода
	now := time.Now()
	currentDay := now.Day()

	if currentDay < 15 {
		return false, fmt.Sprintf("Withdrawal available from 15th day. Current day: %d", currentDay)
	}

	return true, "Withdrawal approved"
}

// Расчет количества прошедших месяцев
func calculateMonthsPassed(startDate, endDate time.Time) int {
	years := endDate.Year() - startDate.Year()
	months := endDate.Month() - startDate.Month()

	// Корректировка если день в endDate меньше дня в startDate
	if endDate.Day() < startDate.Day() {
		months--
	}

	return years*12 + int(months)
}

// Получение детальной информации о выводах
func GetVestingWithdrawalHistory() []map[string]interface{} {
	// TODO: Получить из базы
	// withdrawals := getVestingWithdrawalsFromDB()

	// Временные данные для примера
	withdrawals := []map[string]interface{}{
		{
			"month":            1,
			"withdraw_date":    "2024-02-15",
			"amount":           30000000,
			"percentage":       10.0,
			"status":           "completed",
			"transaction_hash": "0x1234567890abcdef",
		},
		{
			"month":            2,
			"withdraw_date":    "2024-03-15",
			"amount":           30000000,
			"percentage":       10.0,
			"status":           "completed",
			"transaction_hash": "0x1234567891abcdef",
		},
	}

	return withdrawals
}

// Форматирование суммы BKC для админ-вестинга
func FormatBKCAdmin(amount int64) string {
	if amount >= 1000000000 {
		return fmt.Sprintf("%.1f млрд", float64(amount)/1000000000)
	} else if amount >= 1000000 {
		return fmt.Sprintf("%.1f млн", float64(amount)/1000000)
	} else if amount >= 1000 {
		return fmt.Sprintf("%.1f тыс", float64(amount)/1000)
	}
	return fmt.Sprintf("%d", amount)
}

// Получение следующей даты вывода
func GetNextWithdrawalDate(vesting *AdminVestingSchedule) string {
	if vesting.IsCompleted {
		return "Vesting completed"
	}

	return fmt.Sprintf("Доступно: %s (каждое 15-е число месяца)", vesting.NextWithdrawDate)
}

// Статистика вестинга
func GetVestingStats(vesting *AdminVestingSchedule) map[string]interface{} {
	return map[string]interface{}{
		"total_amount":      vesting.TotalAmount,
		"formatted_total":   FormatBKC(vesting.TotalAmount),
		"period_months":     vesting.PeriodMonths,
		"percent_per_month": vesting.PercentPerMonth,
		"monthly_amount":    vesting.MonthlyAmount,
		"formatted_monthly": FormatBKC(vesting.MonthlyAmount),
		"current_status": map[string]interface{}{
			"current_month":       vesting.CurrentMonth,
			"total_withdrawn":     vesting.TotalWithdrawn,
			"formatted_withdrawn": FormatBKC(vesting.TotalWithdrawn),
			"remaining_amount":    vesting.RemainingAmount,
			"formatted_remaining": FormatBKC(vesting.RemainingAmount),
			"next_withdraw_date":  vesting.NextWithdrawDate,
			"is_completed":        vesting.IsCompleted,
			"progress_percent":    float64(vesting.TotalWithdrawn) / float64(vesting.TotalAmount) * 100,
		},
		"schedule": map[string]interface{}{
			"start_date":        vesting.StartDate,
			"end_date":          vesting.EndDate,
			"withdrawal_rule":   "Каждое 15-е число месяца",
			"total_withdrawals": vesting.PeriodMonths,
		},
	}
}
