package ton

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// TON API конфигурация
const (
	TON_API_KEY       = "AHIRWAHVAEPU57IAAAAHPR2HMUAFO3SOIHX5UQJKP47OYHBJXH2ZUISDGIKVIAMDIJJTNUI"
	TON_API_URL       = "https://tonapi.io/v2"
	BOT_WALLET        = "UQAEWAnkzkh2dUrTkVN0irMYoYgXInIGb5A_wBmqPJOVVKwX"
	COMMISSION_WALLET = "UQAEWAnkzkh2dUrTkVN0irMYoYgXInIGb5A_wBmqPJOVVKwX" // Твой кошелек для комиссий
)

// Курсы
const (
	TON_RATE_PER_USD  = 5.0    // 1 TON = $5
	BKC_RATE_PER_USD  = 1000.0 // 1000 BKC = $1
	BKC_PER_TON       = 5000.0 // 1 TON = 5000 BKC
	MIN_EXCHANGE_RATE = 0.70   // Минимум 70% от рыночного курса
)

// Структуры для TON API
type TonMessage struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
	Payload string `json:"payload"`
}

type TonConnectPayload struct {
	Messages []TonMessage `json:"messages"`
}

type WebhookData struct {
	Event   string `json:"event"`
	Payload struct {
		Comment string `json:"comment"`
		Amount  int64  `json:"amount"`
		From    string `json:"from"`
		To      string `json:"to"`
	} `json:"payload"`
}

type P2POrder struct {
	ID             string    `json:"id"`
	SellerID       int64     `json:"seller_id"`
	SellerWallet   string    `json:"seller_wallet"`
	AmountBKC      float64   `json:"amount_bkc"`
	PricePerBKCTON float64   `json:"price_per_bkc_ton"`
	TotalTONNeeded float64   `json:"total_ton_needed"`
	CommissionTON  float64   `json:"commission_ton"`
	SellerReceives float64   `json:"seller_receives"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// TON клиент
type TonClient struct {
	apiKey string
	client *http.Client
}

func NewTonClient() *TonClient {
	return &TonClient{
		apiKey: TON_API_KEY,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Проверка минимального курса
func (tc *TonClient) ValidateExchangeRate(pricePerBKCTON float64) error {
	marketRate := BKC_RATE_PER_USD / TON_RATE_PER_USD // 1000/5 = 200 BKC per TON
	minAllowedRate := marketRate * MIN_EXCHANGE_RATE  // 200 * 0.7 = 140 BKC per TON

	if pricePerBKCTON < minAllowedRate {
		return fmt.Errorf("курс слишком низкий. Минимальный: %.6f TON per BKC", minAllowedRate)
	}
	return nil
}

// Создание P2P ордера на продажу
func (tc *TonClient) CreateSellOrder(sellerID int64, amountBKC float64, pricePerBKCTON, sellerWallet string) (*P2POrder, error) {
	// Конвертация цены в float64
	pricePerBKCTONFloat, _ := strconv.ParseFloat(pricePerBKCTON, 64)

	// Проверка минимального курса
	if err := tc.ValidateExchangeRate(pricePerBKCTONFloat); err != nil {
		return nil, err
	}

	// Расчеты
	totalTONNeeded := amountBKC * pricePerBKCTONFloat
	commissionTON := totalTONNeeded * 0.05 // 5% комиссия
	sellerReceives := amountBKC + commissionTON

	marketRateTON := BKC_RATE_PER_USD / TON_RATE_PER_USD // 1000/5 = 200 BKC per TON
	_ = (pricePerBKCTONFloat / marketRateTON) * 100      // rateVsMarket calculation

	// Создание ордера
	order := &P2POrder{
		ID:             fmt.Sprintf("sell_%d", time.Now().Unix()),
		SellerID:       sellerID,
		SellerWallet:   sellerWallet,
		AmountBKC:      amountBKC,
		PricePerBKCTON: pricePerBKCTONFloat,
		TotalTONNeeded: totalTONNeeded,
		CommissionTON:  commissionTON,
		SellerReceives: sellerReceives,
		Status:         "active",
		CreatedAt:      time.Now(),
	}

	// TODO: Сохранить в базу данных
	// db.SaveP2POrder(order)

	return order, nil
}

// Создание TonConnect payload для покупки
func (tc *TonClient) CreateBuyPayload(order *P2POrder, buyerID int64, amountBKC float64) (*TonConnectPayload, error) {
	// Расчет платежа
	sellerReceives := amountBKC * order.PricePerBKCTON
	commissionTON := sellerReceives * 0.05
	_ = sellerReceives + commissionTON // totalPay calculation

	// Создание payload
	payload := &TonConnectPayload{
		Messages: []TonMessage{
			{
				Address: order.SellerWallet,
				Amount:  fmt.Sprintf("%.0f", sellerReceives*1e9), // В нанотонах
				Payload: fmt.Sprintf("ORDER_%s_BUYER_%d", order.ID, buyerID),
			},
			{
				Address: COMMISSION_WALLET,
				Amount:  fmt.Sprintf("%.0f", commissionTON*1e9),
				Payload: fmt.Sprintf("FEE_%s_BUYER_%d", order.ID, buyerID),
			},
		},
	}

	return payload, nil
}

// Отправка уведомления в TON
func (tc *TonClient) SendNotification(address, message string) error {
	reqBody := map[string]interface{}{
		"address": address,
		"message": message,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", TON_API_URL+"/message/send", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+tc.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := tc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Проверка баланса кошелька
func (tc *TonClient) GetBalance(address string) (float64, error) {
	url := fmt.Sprintf("%s/blockchain/account/address/%s", TON_API_URL, address)

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

// Валидация TON адреса
func (tc *TonClient) ValidateTONAddress(address string) bool {
	// Базовая проверка формата TON адреса
	if len(address) < 48 || len(address) > 66 {
		return false
	}

	// Проверка префиксов
	validPrefixes := []string{"0:", "EQ", "UQ"}
	for _, prefix := range validPrefixes {
		if len(address) > len(prefix) && address[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// Расчет курса BKC в TON
func (tc *TonClient) GetBKCToTONRate() float64 {
	return TON_RATE_PER_USD / BKC_RATE_PER_USD // 0.00143
}

// Расчет минимального разрешенного курса
func (tc *TonClient) GetMinBKCToTONRate() float64 {
	marketRate := tc.GetBKCToTONRate()
	return marketRate * MIN_EXCHANGE_RATE // 0.00100
}

// Форматирование суммы в TON
func FormatTON(amount float64) string {
	return fmt.Sprintf("%.6f", amount)
}

// Форматирование суммы в нанотонах
func FormatNanoTON(amount float64) string {
	return fmt.Sprintf("%.0f", amount*1e9)
}
