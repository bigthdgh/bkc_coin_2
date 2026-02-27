package i18n

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// I18nManager - менеджер интернационализации
type I18nManager struct {
	SupportedLanguages []string                     `json:"supported_languages"`
	DefaultLanguage    string                       `json:"default_language"`
	Translations       map[string]map[string]string `json:"translations"`
	CurrencyRates      map[string]float64           `json:"currency_rates"`
	RegionalSettings   map[string]RegionConfig      `json:"regional_settings"`
	mutex              sync.RWMutex
}

// RegionConfig - конфигурация региона
type RegionConfig struct {
	Country      string   `json:"country"`
	Currency     string   `json:"currency"`
	Timezone     string   `json:"timezone"`
	DateFormat   string   `json:"date_format"`
	NumberFormat string   `json:"number_format"`
	Language     string   `json:"language"`
	Features     []string `json:"features"`
	Restrictions []string `json:"restrictions"`
}

// LanguageData - данные языка
type LanguageData struct {
	Code         string            `json:"code"`
	Name         string            `json:"name"`
	NativeName   string            `json:"native_name"`
	Direction    string            `json:"direction"` // ltr, rtl
	Translations map[string]string `json:"translations"`
}

// NewI18nManager - создание менеджера интернационализации
func NewI18nManager() *I18nManager {
	i18n := &I18nManager{
		SupportedLanguages: []string{"en", "ru"},
		DefaultLanguage:    "en",
		Translations:       make(map[string]map[string]string),
		CurrencyRates:      make(map[string]float64),
		RegionalSettings:   make(map[string]RegionConfig),
	}

	// Инициализируем переводы
	i18n.initializeTranslations()

	// Инициализируем курсы валют
	i18n.initializeCurrencyRates()

	// Инициализируем региональные настройки
	i18n.initializeRegionalSettings()

	return i18n
}

// initializeTranslations - инициализация переводов
func (i18n *I18nManager) initializeTranslations() {
	i18n.mutex.Lock()
	defer i18n.mutex.Unlock()

	// Английские переводы
	i18n.Translations["en"] = map[string]string{
		// Общие
		"app_name": "BKC Coin",
		"welcome":  "Welcome to BKC Coin",
		"loading":  "Loading...",
		"error":    "Error",
		"success":  "Success",
		"cancel":   "Cancel",
		"confirm":  "Confirm",
		"save":     "Save",
		"delete":   "Delete",
		"edit":     "Edit",
		"back":     "Back",
		"next":     "Next",
		"previous": "Previous",
		"close":    "Close",
		"search":   "Search",
		"filter":   "Filter",
		"refresh":  "Refresh",
		"settings": "Settings",
		"profile":  "Profile",
		"logout":   "Logout",

		// Меню и навигация
		"home":        "Home",
		"games":       "Games",
		"marketplace": "Marketplace",
		"nft":         "NFT",
		"wallet":      "Wallet",
		"referrals":   "Referrals",
		"leaderboard": "Leaderboard",
		"support":     "Support",

		// Игра
		"tap_to_earn":  "Tap to Earn",
		"energy":       "Energy",
		"balance":      "Balance",
		"level":        "Level",
		"experience":   "Experience",
		"tap_reward":   "Tap Reward",
		"energy_regen": "Energy Regeneration",
		"daily_limit":  "Daily Limit",
		"taps_today":   "Taps Today",
		"earn_today":   "Earned Today",
		"best_day":     "Best Day",

		// NFT
		"my_nfts":      "My NFTs",
		"nft_market":   "NFT Market",
		"nft_auction":  "NFT Auction",
		"buy_now":      "Buy Now",
		"place_bid":    "Place Bid",
		"current_bid":  "Current Bid",
		"starting_bid": "Starting Bid",
		"buyout_price": "Buyout Price",
		"time_left":    "Time Left",

		// Кошелек
		"withdraw":            "Withdraw",
		"deposit":             "Deposit",
		"transaction_history": "Transaction History",
		"send":                "Send",
		"receive":             "Receive",
		"exchange":            "Exchange",

		// Рефералы
		"invite_friends":  "Invite Friends",
		"referral_link":   "Referral Link",
		"referrals_count": "Referrals Count",
		"referral_bonus":  "Referral Bonus",
		"copy_link":       "Copy Link",
		"share":           "Share",

		// Ошибки
		"insufficient_balance": "Insufficient Balance",
		"insufficient_energy":  "Insufficient Energy",
		"daily_limit_reached":  "Daily Limit Reached",
		"network_error":        "Network Error",
		"server_error":         "Server Error",
		"invalid_request":      "Invalid Request",
		"unauthorized":         "Unauthorized",
		"forbidden":            "Forbidden",
		"not_found":            "Not Found",

		// Уведомления
		"tap_success":          "Tap Successful!",
		"level_up":             "Level Up!",
		"achievement_unlocked": "Achievement Unlocked!",
		"new_referral":         "New Referral!",
		"bonus_received":       "Bonus Received!",
		"payment_received":     "Payment Received",
		"payment_sent":         "Payment Sent",

		// Время и даты
		"now":         "Now",
		"today":       "Today",
		"yesterday":   "Yesterday",
		"this_week":   "This Week",
		"this_month":  "This Month",
		"minutes_ago": "minutes ago",
		"hours_ago":   "hours ago",
		"days_ago":    "days ago",
		"weeks_ago":   "weeks ago",
		"months_ago":  "months ago",
	}

	// Русские переводы
	i18n.Translations["ru"] = map[string]string{
		// Общие
		"app_name": "BKC Coin",
		"welcome":  "Добро пожаловать в BKC Coin",
		"loading":  "Загрузка...",
		"error":    "Ошибка",
		"success":  "Успех",
		"cancel":   "Отмена",
		"confirm":  "Подтвердить",
		"save":     "Сохранить",
		"delete":   "Удалить",
		"edit":     "Редактировать",
		"back":     "Назад",
		"next":     "Далее",
		"previous": "Предыдущий",
		"close":    "Закрыть",
		"search":   "Поиск",
		"filter":   "Фильтр",
		"refresh":  "Обновить",
		"settings": "Настройки",
		"profile":  "Профиль",
		"logout":   "Выйти",

		// Меню и навигация
		"home":        "Главная",
		"games":       "Игры",
		"marketplace": "Маркетплейс",
		"nft":         "NFT",
		"wallet":      "Кошелек",
		"referrals":   "Рефералы",
		"leaderboard": "Таблица лидеров",
		"support":     "Поддержка",

		// Игра
		"tap_to_earn":  "Тапай чтобы зарабатывать",
		"energy":       "Энергия",
		"balance":      "Баланс",
		"level":        "Уровень",
		"experience":   "Опыт",
		"tap_reward":   "Награда за тап",
		"energy_regen": "Восстановление энергии",
		"daily_limit":  "Дневной лимит",
		"taps_today":   "Тапов сегодня",
		"earn_today":   "Заработано сегодня",
		"best_day":     "Лучший день",

		// NFT
		"my_nfts":      "Мои NFT",
		"nft_market":   "Рынок NFT",
		"nft_auction":  "Аукцион NFT",
		"buy_now":      "Купить сейчас",
		"place_bid":    "Сделать ставку",
		"current_bid":  "Текущая ставка",
		"starting_bid": "Начальная ставка",
		"buyout_price": "Цена выкупа",
		"time_left":    "Осталось времени",

		// Кошелек
		"withdraw":            "Вывести",
		"deposit":             "Пополнить",
		"transaction_history": "История транзакций",
		"send":                "Отправить",
		"receive":             "Получить",
		"exchange":            "Обменять",

		// Рефералы
		"invite_friends":  "Пригласить друзей",
		"referral_link":   "Реферальная ссылка",
		"referrals_count": "Количество рефералов",
		"referral_bonus":  "Реферальный бонус",
		"copy_link":       "Копировать ссылку",
		"share":           "Поделиться",

		// Ошибки
		"insufficient_balance": "Недостаточно средств",
		"insufficient_energy":  "Недостаточно энергии",
		"daily_limit_reached":  "Дневной лимит достигнут",
		"network_error":        "Ошибка сети",
		"server_error":         "Ошибка сервера",
		"invalid_request":      "Неверный запрос",
		"unauthorized":         "Не авторизован",
		"forbidden":            "Доступ запрещен",
		"not_found":            "Не найдено",

		// Уведомления
		"tap_success":          "Тап успешный!",
		"level_up":             "Новый уровень!",
		"achievement_unlocked": "Достижение разблокировано!",
		"new_referral":         "Новый реферал!",
		"bonus_received":       "Бонус получен!",
		"payment_received":     "Платеж получен",
		"payment_sent":         "Платеж отправлен",

		// Время и даты
		"now":         "Сейчас",
		"today":       "Сегодня",
		"yesterday":   "Вчера",
		"this_week":   "Эта неделя",
		"this_month":  "Этот месяц",
		"minutes_ago": "минут назад",
		"hours_ago":   "часов назад",
		"days_ago":    "дней назад",
		"weeks_ago":   "недель назад",
		"months_ago":  "месяцев назад",
	}
}

// initializeCurrencyRates - инициализация курсов валют
func (i18n *I18nManager) initializeCurrencyRates() {
	i18n.mutex.Lock()
	defer i18n.mutex.Unlock()

	// Базовая валюта - BKC (только криптовалюты, фиаты удалены)
	i18n.CurrencyRates = map[string]float64{
		"BKC":         1.0,    // Базовая валюта
		"TON":         0.0008, // 1250 BKC = 1 TON
		"USDT":        0.001,  // 1000 BKC = 1 USDT (универсальный курс)
		"USDT_TON":    0.001,  // 1000 BKC = 1 USDT (TON)
		"USDT_SOLANA": 0.001,  // 1000 BKC = 1 USDT (Solana)
	}
}

// initializeRegionalSettings - инициализация региональных настроек
func (i18n *I18nManager) initializeRegionalSettings() {
	i18n.mutex.Lock()
	defer i18n.mutex.Unlock()

	i18n.RegionalSettings = map[string]RegionConfig{
		"US": {
			Country:      "US",
			Currency:     "USDT",
			Timezone:     "America/New_York",
			DateFormat:   "MM/DD/YYYY",
			NumberFormat: "1,234.56",
			Language:     "en",
			Features:     []string{"ton", "solana_usdt"},
			Restrictions: []string{},
		},
		"RU": {
			Country:      "RU",
			Currency:     "USDT",
			Timezone:     "Europe/Moscow",
			DateFormat:   "DD.MM.YYYY",
			NumberFormat: "1 234,56",
			Language:     "ru",
			Features:     []string{"ton", "ton_usdt", "solana_usdt"},
			Restrictions: []string{},
		},
		"GB": {
			Country:      "GB",
			Currency:     "USDT",
			Timezone:     "Europe/London",
			DateFormat:   "DD/MM/YYYY",
			NumberFormat: "1,234.56",
			Language:     "en",
			Features:     []string{"ton", "solana_usdt"},
			Restrictions: []string{},
		},
		"DE": {
			Country:      "DE",
			Currency:     "USDT",
			Timezone:     "Europe/Berlin",
			DateFormat:   "DD.MM.YYYY",
			NumberFormat: "1.234,56",
			Language:     "en",
			Features:     []string{"ton", "solana_usdt"},
			Restrictions: []string{},
		},
		"FR": {
			Country:      "FR",
			Currency:     "USDT",
			Timezone:     "Europe/Paris",
			DateFormat:   "DD/MM/YYYY",
			NumberFormat: "1 234,56",
			Language:     "en",
			Features:     []string{"ton", "solana_usdt"},
			Restrictions: []string{},
		},
		"CN": {
			Country:      "CN",
			Currency:     "USDT",
			Timezone:     "Asia/Shanghai",
			DateFormat:   "YYYY-MM-DD",
			NumberFormat: "1,234.56",
			Language:     "en",
			Features:     []string{"ton"},
			Restrictions: []string{"solana_usdt"},
		},
		"IN": {
			Country:      "IN",
			Currency:     "USDT",
			Timezone:     "Asia/Kolkata",
			DateFormat:   "DD/MM/YYYY",
			NumberFormat: "1,23,456.78",
			Language:     "en",
			Features:     []string{"ton"},
			Restrictions: []string{"solana_usdt"},
		},
		"BR": {
			Country:      "BR",
			Currency:     "USDT",
			Timezone:     "America/Sao_Paulo",
			DateFormat:   "DD/MM/YYYY",
			NumberFormat: "1.234,56",
			Language:     "en",
			Features:     []string{"ton", "solana_usdt"},
			Restrictions: []string{},
		},
		"JP": {
			Country:      "JP",
			Currency:     "USDT",
			Timezone:     "Asia/Tokyo",
			DateFormat:   "YYYY/MM/DD",
			NumberFormat: "1,234.56",
			Language:     "en",
			Features:     []string{"ton", "solana_usdt"},
			Restrictions: []string{},
		},
		"KR": {
			Country:      "KR",
			Currency:     "USDT",
			Timezone:     "Asia/Seoul",
			DateFormat:   "YYYY. MM. DD.",
			NumberFormat: "1,234.56",
			Language:     "en",
			Features:     []string{"ton"},
			Restrictions: []string{"solana_usdt"},
		},
		"KZ": {
			Country:      "KZ",
			Currency:     "USDT",
			Timezone:     "Asia/Almaty",
			DateFormat:   "DD.MM.YYYY",
			NumberFormat: "1 234,56",
			Language:     "ru",
			Features:     []string{"ton", "ton_usdt", "solana_usdt"},
			Restrictions: []string{},
		},
		"UA": {
			Country:      "UA",
			Currency:     "USDT",
			Timezone:     "Europe/Kiev",
			DateFormat:   "DD.MM.YYYY",
			NumberFormat: "1 234,56",
			Language:     "ru",
			Features:     []string{"ton", "ton_usdt"},
			Restrictions: []string{"solana_usdt"},
		},
		"BY": {
			Country:      "BY",
			Currency:     "USDT",
			Timezone:     "Europe/Minsk",
			DateFormat:   "DD.MM.YYYY",
			NumberFormat: "1 234,56",
			Language:     "ru",
			Features:     []string{"ton", "ton_usdt"},
			Restrictions: []string{"solana_usdt"},
		},
	}
}

// GetText - получение текста на указанном языке
func (i18n *I18nManager) GetText(lang, key string) string {
	i18n.mutex.RLock()
	defer i18n.mutex.RUnlock()

	// Проверяем поддерживаемый язык
	if !i18n.isLanguageSupported(lang) {
		lang = i18n.DefaultLanguage
	}

	// Ищем перевод
	if translations, exists := i18n.Translations[lang]; exists {
		if text, exists := translations[key]; exists {
			return text
		}
	}

	// Если перевод не найден, пробуем язык по умолчанию
	if lang != i18n.DefaultLanguage {
		if translations, exists := i18n.Translations[i18n.DefaultLanguage]; exists {
			if text, exists := translations[key]; exists {
				return text
			}
		}
	}

	// Возвращаем ключ если перевод не найден
	return "[" + key + "]"
}

// isLanguageSupported - проверка поддержки языка
func (i18n *I18nManager) isLanguageSupported(lang string) bool {
	for _, supported := range i18n.SupportedLanguages {
		if supported == lang {
			return true
		}
	}
	return false
}

// ConvertCurrency - конвертация валюты
func (i18n *I18nManager) ConvertCurrency(amount int64, fromCurrency, toCurrency string) (float64, error) {
	i18n.mutex.RLock()
	defer i18n.mutex.RUnlock()

	fromRate, fromExists := i18n.CurrencyRates[fromCurrency]
	toRate, toExists := i18n.CurrencyRates[toCurrency]

	if !fromExists || !toExists {
		return 0, fmt.Errorf("currency not supported")
	}

	// Конвертируем BKC -> целевая валюта
	if fromCurrency == "BKC" {
		return float64(amount) * toRate, nil
	}

	// Конвертируем из исходной валюты в BKC, затем в целевую
	bkcAmount := float64(amount) / fromRate
	return bkcAmount * toRate, nil
}

// GetRegionalSettings - получение региональных настроек
func (i18n *I18nManager) GetRegionalSettings(countryCode string) (RegionConfig, error) {
	i18n.mutex.RLock()
	defer i18n.mutex.RUnlock()

	if settings, exists := i18n.RegionalSettings[countryCode]; exists {
		return settings, nil
	}

	// Возвращаем настройки по умолчанию
	return i18n.RegionalSettings["US"], nil
}

// FormatCurrency - форматирование валюты
func (i18n *I18nManager) FormatCurrency(amount float64, currency string) string {
	settings, err := i18n.GetRegionalSettings("US") // По умолчанию US
	if err != nil {
		return fmt.Sprintf("%.2f %s", amount, currency)
	}

	// Форматируем в зависимости от региона
	switch settings.Country {
	case "RU", "KZ", "UA", "BY":
		return fmt.Sprintf("%.2f %s", amount, currency)
	case "DE", "FR":
		return fmt.Sprintf("%.2f %s", amount, currency)
	case "IN":
		return fmt.Sprintf("%.2f %s", amount, currency)
	default:
		return fmt.Sprintf("%.2f %s", amount, currency)
	}
}

// FormatNumber - форматирование чисел
func (i18n *I18nManager) FormatNumber(number float64, countryCode string) string {
	settings, err := i18n.GetRegionalSettings(countryCode)
	if err != nil {
		return fmt.Sprintf("%.2f", number)
	}

	// Форматируем в зависимости от региона
	switch settings.NumberFormat {
	case "1,234.56":
		return fmt.Sprintf("%.2f", number)
	case "1 234,56":
		return fmt.Sprintf("%.2f", number)
	case "1.234,56":
		return fmt.Sprintf("%.2f", number)
	case "1,23,456.78":
		return fmt.Sprintf("%.2f", number)
	default:
		return fmt.Sprintf("%.2f", number)
	}
}

// GetSupportedLanguages - получение поддерживаемых языков
func (i18n *I18nManager) GetSupportedLanguages() []LanguageData {
	i18n.mutex.RLock()
	defer i18n.mutex.RUnlock()

	languages := []LanguageData{
		{
			Code:       "en",
			Name:       "English",
			NativeName: "English",
			Direction:  "ltr",
		},
		{
			Code:       "ru",
			Name:       "Russian",
			NativeName: "Русский",
			Direction:  "ltr",
		},
	}

	return languages
}

// AddTranslation - добавление перевода
func (i18n *I18nManager) AddTranslation(lang, key, text string) error {
	i18n.mutex.Lock()
	defer i18n.mutex.Unlock()

	if !i18n.isLanguageSupported(lang) {
		return fmt.Errorf("language not supported: %s", lang)
	}

	if _, exists := i18n.Translations[lang]; !exists {
		i18n.Translations[lang] = make(map[string]string)
	}

	i18n.Translations[lang][key] = text
	log.Printf("Translation added: %s.%s = %s", lang, key, text)
	return nil
}

// UpdateCurrencyRate - обновление курса валюты
func (i18n *I18nManager) UpdateCurrencyRate(currency string, rate float64) error {
	i18n.mutex.Lock()
	defer i18n.mutex.Unlock()

	if _, exists := i18n.CurrencyRates[currency]; !exists {
		return fmt.Errorf("currency not supported: %s", currency)
	}

	i18n.CurrencyRates[currency] = rate
	log.Printf("Currency rate updated: %s = %.6f", currency, rate)
	return nil
}

// GetAvailableCurrencies - получение доступных валют
func (i18n *I18nManager) GetAvailableCurrencies() []string {
	i18n.mutex.RLock()
	defer i18n.mutex.RUnlock()

	currencies := make([]string, 0, len(i18n.CurrencyRates))
	for currency := range i18n.CurrencyRates {
		currencies = append(currencies, currency)
	}

	return currencies
}

// DetectLanguage - определение языка по заголовкам HTTP
func (i18n *I18nManager) DetectLanguage(acceptLanguage string) string {
	// Простая логика определения языка
	// В реальном приложении здесь будет более сложный парсинг Accept-Language

	if acceptLanguage == "" {
		return i18n.DefaultLanguage
	}

	// Проверяем русский
	if contains(acceptLanguage, "ru") {
		return "ru"
	}

	// По умолчанию английский
	return "en"
}

// DetectCountry - определение страны по IP
func (i18n *I18nManager) DetectCountry(ip string) string {
	// В реальном приложении здесь будет запрос к GeoIP сервису
	// Для примера возвращаем US по умолчанию
	return "US"
}

// contains - проверка наличия подстроки
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// toJSON - конвертация в JSON
func (i18n *I18nManager) toJSON(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}
