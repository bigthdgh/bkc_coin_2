package ton

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Обновленный TON клиент с управлением курсами
type TonClientUpdated struct {
	apiKey      string
	client      *http.Client
	rateManager *RateManager
	wallet      string
}

// NewTonClientUpdated создает новый TON клиент с управлением курсами
func NewTonClientUpdated() *TonClientUpdated {
	return &TonClientUpdated{
		apiKey:      "AHIRWAHVAEPU57IAAAAHPR2HMUAFO3SOIHX5UQJKP47OYHBJXH2ZUISDGIKVIAMDIJJTNUI",
		client:      &http.Client{Timeout: 30 * time.Second},
		rateManager: NewRateManager(),
		wallet:      "UQAEWAnkzkh2dUrTkVN0irMYoYgXInIGb5A_wBmqPJOVVKwX",
	}
}

// SetTONRate устанавливает курс TON/USD (для админа)
func (tc *TonClientUpdated) SetTONRate(rate float64) error {
	return tc.rateManager.SetTONRate(rate)
}

// GetTONRate получает текущий курс TON/USD
func (tc *TonClientUpdated) GetTONRate() (float64, error) {
	return tc.rateManager.GetTONRate()
}

// GetRatesInfo получает полную информацию о курсах
func (tc *TonClientUpdated) GetRatesInfo() map[string]interface{} {
	return tc.rateManager.GetCurrentRatesInfo()
}

// ConvertTONtoBKC конвертирует TON в BKC с текущим курсом
func (tc *TonClientUpdated) ConvertTONtoBKC(tonAmount float64) float64 {
	return tc.rateManager.ConvertTONtoBKC(tonAmount)
}

// ConvertBKCtoTON конвертирует BKC в TON с текущим курсом
func (tc *TonClientUpdated) ConvertBKCtoTON(bkcAmount float64) float64 {
	return tc.rateManager.ConvertBKCtoTON(bkcAmount)
}

// GetBalance получает баланс TON кошелька
func (tc *TonClientUpdated) GetBalance() (float64, error) {
	url := fmt.Sprintf("https://tonapi.io/v2/blockchain/account/address/%s", tc.wallet)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tc.apiKey)

	resp, err := tc.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Balance int64 `json:"balance"`
	}

	json.Unmarshal(body, &result)

	return float64(result.Balance) / 1e9, nil // Конвертация из нанотонов
}

// ValidateTONAddress валидирует TON адрес
func (tc *TonClientUpdated) ValidateTONAddress(address string) bool {
	// Базовая проверка формата TON адреса
	if len(address) < 48 || len(address) > 66 {
		return false
	}

	// Проверка префикса
	if !((address[0] == '0' && address[1] == ':') ||
		(address[0:2] == "UQ") ||
		(address[0:2] == "EQ") ||
		(address[0:2] == "k0")) {
		return false
	}

	return true
}

// CreateTransaction создает транзакцию для пополнения
func (tc *TonClientUpdated) CreateTransaction(amountTON float64, userID int64) (map[string]interface{}, error) {
	// Конвертация в нанотоны
	amountNano := int64(amountTON * 1e9)

	// Создание payload
	payload := fmt.Sprintf("BKC_TOPUP_USER_%d_%d", userID, time.Now().Unix())

	transaction := map[string]interface{}{
		"address": tc.wallet,
		"amount":  amountNano,
		"payload": payload,
		"message": fmt.Sprintf("Пополнение BKC для пользователя %d", userID),
	}

	return transaction, nil
}

// GetTransactionHistory получает историю транзакций
func (tc *TonClientUpdated) GetTransactionHistory(limit int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://tonapi.io/v2/blockchain/account/address/%s/transactions?limit=%d", tc.wallet, limit)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+tc.apiKey)

	resp, err := tc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Transactions []map[string]interface{} `json:"transactions"`
	}

	json.Unmarshal(body, &result)

	return result.Transactions, nil
}

// SendNotification отправляет уведомление в TON
func (tc *TonClientUpdated) SendNotification(address, message string) error {
	reqBody := map[string]interface{}{
		"address": address,
		"message": message,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://tonapi.io/v2/message/send",
		bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+tc.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := tc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetWalletInfo получает информацию о кошельке
func (tc *TonClientUpdated) GetWalletInfo() map[string]interface{} {
	return map[string]interface{}{
		"address":  tc.wallet,
		"balance":  "Запросить через GetBalance()",
		"qr_code":  fmt.Sprintf("ton://transfer/%s", tc.wallet),
		"explorer": fmt.Sprintf("https://tonscan.org/address/%s", tc.wallet),
	}
}
