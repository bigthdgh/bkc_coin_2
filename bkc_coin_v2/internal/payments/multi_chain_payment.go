package payments

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"bkc_coin_v2/internal/database"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// MultiChainPaymentManager - менеджер мультицепочечных платежей
type MultiChainPaymentManager struct {
	db              *database.UnifiedDB
	config          PaymentConfig
	tonWallet       string
	solanaWallet    string
	activeOrders    map[string]*PaymentOrder
	orderMutex      sync.RWMutex
	solanaClient    *rpc.Client
	commissionRates CommissionConfig
}

// PaymentConfig - конфигурация платежей (только TON и USDT)
type PaymentConfig struct {
	EnabledChains       []string `json:"enabled_chains"`
	TONMasterAddress    string   `json:"ton_master_address"`
	SolanaMasterAddress string   `json:"solana_master_address"`
	USDTContractTON     string   `json:"usdt_contract_ton"`
	USDTMintSolana      string   `json:"usdt_mint_solana"`
	MinAmount           float64  `json:"min_amount"`
	MaxAmount           float64  `json:"max_amount"`
	OrderTimeout        int      `json:"order_timeout"` // в минутах
}

// CommissionConfig - конфигурация комиссий
type CommissionConfig struct {
	PlatformCommission float64 `json:"platform_commission"` // % комиссии платформы
	NFTCommission      float64 `json:"nft_commission"`      // % комиссии за NFT
	MarketCommission   float64 `json:"market_commission"`   // % комиссии за маркетплейс
	ReferralCommission float64 `json:"referral_commission"` // % комиссии за рефералов
	MinCommission      int64   `json:"min_commission"`      // минимальная комиссия в BKC
}

// PaymentOrder - заказ на оплату
type PaymentOrder struct {
	OrderID         string                 `json:"order_id"`
	UserID          int64                  `json:"user_id"`
	Type            string                 `json:"type"`  // purchase, withdrawal, nft, market
	Chain           string                 `json:"chain"` // ton, ton_usdt, solana_usdt
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	BKCAmount       int64                  `json:"bkc_amount"`
	Recipient       string                 `json:"recipient"`
	Memo            string                 `json:"memo"`
	Status          string                 `json:"status"` // pending, confirmed, expired, cancelled
	Commission      int64                  `json:"commission"`
	NetAmount       int64                  `json:"net_amount"`
	CreatedAt       time.Time              `json:"created_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	ConfirmedAt     *time.Time             `json:"confirmed_at,omitempty"`
	TransactionHash string                 `json:"transaction_hash,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentRequest - запрос на оплату
type PaymentRequest struct {
	UserID    int64                  `json:"user_id"`
	Type      string                 `json:"type"`
	Chain     string                 `json:"chain"`
	Amount    float64                `json:"amount"`
	Currency  string                 `json:"currency"`
	Recipient string                 `json:"recipient,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentResponse - ответ на создание платежа
type PaymentResponse struct {
	OrderID      string            `json:"order_id"`
	PaymentURL   string            `json:"payment_url"`
	QRCode       string            `json:"qr_code"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Instructions map[string]string `json:"instructions"`
}

// NewMultiChainPaymentManager - создание менеджера платежей
func NewMultiChainPaymentManager(db *database.UnifiedDB, config PaymentConfig) *MultiChainPaymentManager {
	mpm := &MultiChainPaymentManager{
		db:           db,
		config:       config,
		tonWallet:    config.TONMasterAddress,
		solanaWallet: config.SolanaMasterAddress,
		activeOrders: make(map[string]*PaymentOrder),
		commissionRates: CommissionConfig{
			PlatformCommission: 2.5,  // 2.5% комиссии платформы
			NFTCommission:      5.0,  // 5% за NFT транзакции
			MarketCommission:   3.0,  // 3% за маркетплейс
			ReferralCommission: 10.0, // 10% от комиссии идет рефералам
			MinCommission:      100,  // минимальная комиссия 100 BKC
		},
	}

	// Инициализируем Solana клиент
	if contains(config.EnabledChains, "solana_usdt") {
		mpm.solanaClient = rpc.New(rpc.MainNetBeta_RPC)
	}

	// Запускаем мониторинг платежей
	go mpm.startPaymentMonitoring()

	return mpm
}

// CreatePaymentOrder - создание заказа на оплату
func (mpm *MultiChainPaymentManager) CreatePaymentOrder(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	// Валидация запроса
	if err := mpm.validatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Генерируем OrderID
	orderID := mpm.generateOrderID()

	// Конвертируем в BKC
	bkcAmount, err := mpm.convertToBKC(req.Amount, req.Currency, req.Chain)
	if err != nil {
		return nil, fmt.Errorf("conversion failed: %w", err)
	}

	// Рассчитываем комиссию
	commission := mpm.calculateCommission(bkcAmount, req.Type)
	netAmount := bkcAmount - commission

	// Определяем получателя
	recipient := req.Recipient
	if recipient == "" {
		recipient = mpm.getRecipientByChain(req.Chain)
	}

	// Создаем заказ
	order := &PaymentOrder{
		OrderID:    orderID,
		UserID:     req.UserID,
		Type:       req.Type,
		Chain:      req.Chain,
		Amount:     req.Amount,
		Currency:   req.Currency,
		BKCAmount:  bkcAmount,
		Recipient:  recipient,
		Memo:       mpm.generateMemo(orderID),
		Status:     "pending",
		Commission: commission,
		NetAmount:  netAmount,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Duration(mpm.config.OrderTimeout) * time.Minute),
		Metadata:   req.Metadata,
	}

	// Сохраняем заказ
	err = mpm.savePaymentOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Добавляем в активные заказы
	mpm.orderMutex.Lock()
	mpm.activeOrders[orderID] = order
	mpm.orderMutex.Unlock()

	// Генерируем URL для оплаты
	paymentURL, qrCode, instructions := mpm.generatePaymentURL(order)

	response := &PaymentResponse{
		OrderID:      orderID,
		PaymentURL:   paymentURL,
		QRCode:       qrCode,
		ExpiresAt:    order.ExpiresAt,
		Instructions: instructions,
	}

	log.Printf("Payment order created: %s, User: %d, Chain: %s, Amount: %.2f %s",
		orderID, req.UserID, req.Chain, req.Amount, req.Currency)

	return response, nil
}

// validatePaymentRequest - валидация запроса на оплату (только TON и USDT)
func (mpm *MultiChainPaymentManager) validatePaymentRequest(req PaymentRequest) error {
	if req.UserID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	// Проверяем только поддерживаемые цепочки
	supportedChains := []string{"ton", "ton_usdt", "solana_usdt"}
	if !contains(supportedChains, req.Chain) {
		return fmt.Errorf("unsupported chain: %s. Only TON and USDT (Solana) are supported", req.Chain)
	}

	if req.Amount < mpm.config.MinAmount || req.Amount > mpm.config.MaxAmount {
		return fmt.Errorf("amount out of range: %.2f", req.Amount)
	}

	validTypes := []string{"purchase", "withdrawal", "nft", "market"}
	if !contains(validTypes, req.Type) {
		return fmt.Errorf("invalid payment type: %s", req.Type)
	}

	return nil
}

// convertToBKC - конвертация валюты в BKC
func (mpm *MultiChainPaymentManager) convertToBKC(amount float64, currency, chain string) (int64, error) {
	// Курсы конвертации (в реальном приложении берутся из API)
	var rate float64

	switch chain {
	case "ton":
		// TON -> BKC (1 TON = 1000 BKC)
		rate = 1000.0
	case "ton_usdt":
		// USDT -> BKC (1 USDT = 1000 BKC)
		rate = 1000.0
	case "solana_usdt":
		// USDT -> BKC (1 USDT = 1000 BKC)
		rate = 1000.0
	default:
		return 0, fmt.Errorf("unsupported chain for conversion: %s", chain)
	}

	bkcAmount := int64(amount * rate)
	return bkcAmount, nil
}

// calculateCommission - расчет комиссии
func (mpm *MultiChainPaymentManager) calculateCommission(bkcAmount int64, paymentType string) int64 {
	var commissionRate float64

	switch paymentType {
	case "nft":
		commissionRate = mpm.commissionRates.NFTCommission
	case "market":
		commissionRate = mpm.commissionRates.MarketCommission
	default:
		commissionRate = mpm.commissionRates.PlatformCommission
	}

	commission := int64(float64(bkcAmount) * commissionRate / 100.0)

	// Минимальная комиссия
	if commission < mpm.commissionRates.MinCommission {
		commission = mpm.commissionRates.MinCommission
	}

	return commission
}

// getRecipientByChain - получение адреса получателя по цепочке
func (mpm *MultiChainPaymentManager) getRecipientByChain(chain string) string {
	switch chain {
	case "ton", "ton_usdt":
		return mpm.tonWallet
	case "solana_usdt":
		return mpm.solanaWallet
	default:
		return mpm.tonWallet
	}
}

// generateOrderID - генерация ID заказа
func (mpm *MultiChainPaymentManager) generateOrderID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.StdEncoding.EncodeToString(bytes)[:12]
}

// generateMemo - генерация мемо для транзакции
func (mpm *MultiChainPaymentManager) generateMemo(orderID string) string {
	return fmt.Sprintf("BKC_%s_%d", orderID, time.Now().Unix())
}

// generatePaymentURL - генерация URL для оплаты
func (mpm *MultiChainPaymentManager) generatePaymentURL(order *PaymentOrder) (string, string, map[string]string) {
	var paymentURL, qrCode string
	instructions := make(map[string]string)

	switch order.Chain {
	case "ton":
		paymentURL, qrCode, instructions = mpm.generateTONPaymentURL(order)
	case "ton_usdt":
		paymentURL, qrCode, instructions = mpm.generateTONUSDTURL(order)
	case "solana_usdt":
		paymentURL, qrCode, instructions = mpm.generateSolanaUSDTURL(order)
	}

	return paymentURL, qrCode, instructions
}

// generateTONPaymentURL - генерация URL для оплаты в TON
func (mpm *MultiChainPaymentManager) generateTONPaymentURL(order *PaymentOrder) (string, string, map[string]string) {
	amount := int64(order.Amount * 1000000000) // конвертация в нанотоны

	// Формируем deep link для TON
	paymentURL := fmt.Sprintf("ton://transfer/%s?amount=%d&text=%s",
		order.Recipient, amount, order.Memo)

	qrCode := paymentURL

	instructions := map[string]string{
		"step1": "Click the payment button or scan QR code",
		"step2": "Confirm transaction in your TON wallet",
		"step3": "Wait for confirmation (usually 10-30 seconds)",
		"note":  fmt.Sprintf("Amount: %.2f TON", order.Amount),
	}

	return paymentURL, qrCode, instructions
}

// generateTONUSDTURL - генерация URL для оплаты в USDT (TON)
func (mpm *MultiChainPaymentManager) generateTONUSDTURL(order *PaymentOrder) (string, string, map[string]string) {
	// Для USDT в TON используем jetton transfer
	amount := int64(order.Amount * 1000000) // USDT имеет 6 decimals

	// Формируем deep link для TON с jetton transfer
	paymentURL := fmt.Sprintf("ton://transfer/%s?amount=%d&text=%s&jetton=%s",
		order.Recipient, amount, order.Memo, mpm.config.USDTContractTON)

	qrCode := paymentURL

	instructions := map[string]string{
		"step1": "Click the payment button or scan QR code",
		"step2": "Confirm USDT transfer in your TON wallet",
		"step3": "Wait for confirmation (usually 10-30 seconds)",
		"note":  fmt.Sprintf("Amount: %.2f USDT", order.Amount),
	}

	return paymentURL, qrCode, instructions
}

// generateSolanaUSDTURL - генерация URL для оплаты в USDT (Solana)
func (mpm *MultiChainPaymentManager) generateSolanaUSDTURL(order *PaymentOrder) (string, string, map[string]string) {
	recipient := order.Recipient
	usdtMint := mpm.config.USDTMintSolana
	amount := order.Amount // USDT в Solana имеет 6 decimals, но для Solana Pay используем прямое значение

	// Формируем Solana Pay URL
	paymentURL := fmt.Sprintf("solana:%s?amount=%.6f&spl-token=%s&memo=%s&label=BKC%%20Purchase&reference=%s",
		recipient, amount, usdtMint, order.Memo, order.OrderID)

	qrCode := paymentURL

	instructions := map[string]string{
		"step1": "Click the payment button or scan QR code",
		"step2": "Choose your wallet (Phantom, Trust, MetaMask, etc.)",
		"step3": "Confirm USDT transfer in your Solana wallet",
		"step4": "Wait for confirmation (usually 2-5 seconds)",
		"note":  fmt.Sprintf("Amount: %.2f USDT", order.Amount),
	}

	return paymentURL, qrCode, instructions
}

// startPaymentMonitoring - запуск мониторинга платежей
func (mpm *MultiChainPaymentManager) startPaymentMonitoring() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mpm.checkPendingPayments()
	}
}

// checkPendingPayments - проверка ожидающих платежей
func (mpm *MultiChainPaymentManager) checkPendingPayments() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mpm.orderMutex.RLock()
	pendingOrders := make([]*PaymentOrder, 0)
	for _, order := range mpm.activeOrders {
		if order.Status == "pending" && time.Now().Before(order.ExpiresAt) {
			pendingOrders = append(pendingOrders, order)
		}
	}
	mpm.orderMutex.RUnlock()

	for _, order := range pendingOrders {
		switch order.Chain {
		case "ton", "ton_usdt":
			go mpm.checkTONPayment(ctx, order)
		case "solana_usdt":
			go mpm.checkSolanaPayment(ctx, order)
		}
	}
}

// checkTONPayment - проверка платежа в TON
func (mpm *MultiChainPaymentManager) checkTONPayment(ctx context.Context, order *PaymentOrder) {
	// В реальном приложении здесь будет проверка через TON API
	// Для примера симулируем проверку

	// Симуляция: с вероятностью 10% платеж найден
	if time.Now().Sub(order.CreatedAt) > 30*time.Second {
		// Подтверждаем платеж
		mpm.confirmPayment(ctx, order.OrderID, "simulated_ton_hash")
	}
}

// checkSolanaPayment - проверка платежа в Solana
func (mpm *MultiChainPaymentManager) checkSolanaPayment(ctx context.Context, order *PaymentOrder) {
	if mpm.solanaClient == nil {
		return
	}

	pubKey := solana.MustPublicKeyFromBase58(mpm.solanaWallet)

	// Получаем последние подписи
	sigs, err := mpm.solanaClient.GetSignaturesForAddress(ctx, pubKey, &rpc.GetSignaturesForAddressOpts{
		Limit: 10,
	})

	if err != nil {
		log.Printf("Failed to get Solana signatures: %v", err)
		return
	}

	for _, sig := range sigs {
		// Проверяем, не слишком ли старая транзакция
		if sig.BlockTime != nil && time.Now().Unix()-*sig.BlockTime > 300 {
			continue
		}

		// Получаем детальную информацию о транзакции
		tx, err := mpm.solanaClient.GetTransaction(ctx, sig.Signature, &rpc.GetTransactionOpts{
			Encoding: solana.EncodingJSON,
		})

		if err != nil {
			continue
		}

		// Ищем мемо с нашим OrderID
		if tx != nil && tx.Meta != nil {
			for _, logMsg := range tx.Meta.LogMessages {
				if strings.Contains(logMsg, order.Memo) {
					// Подтверждаем платеж
					mpm.confirmPayment(ctx, order.OrderID, sig.Signature.String())
					return
				}
			}
		}
	}
}

// confirmPayment - подтверждение платежа
func (mpm *MultiChainPaymentManager) confirmPayment(ctx context.Context, orderID, transactionHash string) error {
	mpm.orderMutex.Lock()
	order, exists := mpm.activeOrders[orderID]
	mpm.orderMutex.Unlock()

	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status != "pending" {
		return fmt.Errorf("order already processed: %s", orderID)
	}

	// Обновляем статус заказа
	order.Status = "confirmed"
	order.TransactionHash = transactionHash
	confirmedAt := time.Now()
	order.ConfirmedAt = &confirmedAt

	// Начисляем BKC пользователю (временное решение)
	bkcAmount := order.NetAmount
	// TODO: Создать метод UpdateUserBalance в UnifiedDB
	log.Printf("Adding %d BKC to user %d", bkcAmount, order.UserID)

	// Обновляем в базе данных (временное решение)
	// TODO: Создать метод UpdatePaymentOrder в UnifiedDB
	log.Printf("Payment order %s updated in memory", orderID)

	// Удаляем из активных заказов
	mpm.orderMutex.Lock()
	delete(mpm.activeOrders, orderID)
	mpm.orderMutex.Unlock()

	log.Printf("Payment confirmed: OrderID=%s, UserID=%d, Amount=%.2f BKC, TxHash=%s",
		orderID, order.UserID, bkcAmount, transactionHash)

	return nil
}

// creditUserAccount - начисление BKC на счет пользователя
func (mpm *MultiChainPaymentManager) creditUserAccount(ctx context.Context, order *PaymentOrder) error {
	// Обновляем баланс пользователя
	err := mpm.db.UpdateUserBalance(ctx, order.UserID, order.NetAmount)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	// Записываем транзакцию в историю
	err = mpm.recordTransaction(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	return nil
}

// recordTransaction - запись транзакции в историю
func (mpm *MultiChainPaymentManager) recordTransaction(ctx context.Context, order *PaymentOrder) error {
	// В реальном приложении здесь будет запись в таблицу транзакций
	log.Printf("Transaction recorded: Order %s, User %d, Amount %d BKC, Type %s",
		order.OrderID, order.UserID, order.NetAmount, order.Type)
	return nil
}

// processReferralCommission - обработка реферальной комиссии
func (mpm *MultiChainPaymentManager) processReferralCommission(ctx context.Context, order *PaymentOrder) {
	// Получаем реферала пользователя
	referrerID, err := mpm.getUserReferrer(ctx, order.UserID)
	if err != nil || referrerID == 0 {
		return
	}

	// Рассчитываем реферальную комиссию (10% от нашей комиссии)
	referralCommission := int64(float64(order.Commission) * mpm.commissionRates.ReferralCommission / 100.0)

	if referralCommission > 0 {
		err = mpm.db.UpdateUserBalance(ctx, referrerID, referralCommission)
		if err != nil {
			log.Printf("Failed to credit referral commission: %v", err)
			return
		}

		log.Printf("Referral commission credited: Referrer %d, Order %s, Amount %d BKC",
			referrerID, order.OrderID, referralCommission)
	}
}

// getUserReferrer - получение реферала пользователя
func (mpm *MultiChainPaymentManager) getUserReferrer(ctx context.Context, userID int64) (int64, error) {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем 0 (нет реферала)
	return 0, nil
}

// savePaymentOrder - сохранение заказа на оплату
func (mpm *MultiChainPaymentManager) savePaymentOrder(ctx context.Context, order *PaymentOrder) error {
	// В реальном приложении здесь будет запись в БД
	log.Printf("Payment order saved: %s", order.OrderID)
	return nil
}

// GetPaymentStatus - получение статуса платежа
func (mpm *MultiChainPaymentManager) GetPaymentStatus(ctx context.Context, orderID string) (*PaymentOrder, error) {
	mpm.orderMutex.RLock()
	order, exists := mpm.activeOrders[orderID]
	mpm.orderMutex.RUnlock()

	if exists {
		return order, nil
	}

	// Ищем в базе данных
	return mpm.getPaymentOrderFromDB(ctx, orderID)
}

// getPaymentOrderFromDB - получение заказа из БД
func (mpm *MultiChainPaymentManager) getPaymentOrderFromDB(ctx context.Context, orderID string) (*PaymentOrder, error) {
	// В реальном приложении здесь будет запрос к БД
	return nil, fmt.Errorf("order not found")
}

// GetUserPaymentHistory - история платежей пользователя
func (mpm *MultiChainPaymentManager) GetUserPaymentHistory(ctx context.Context, userID int64, limit int) ([]PaymentOrder, error) {
	// В реальном приложении здесь будет запрос к БД
	return []PaymentOrder{}, nil
}

// GetPaymentStats - статистика платежей
func (mpm *MultiChainPaymentManager) GetPaymentStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	mpm.orderMutex.RLock()
	pendingCount := 0
	totalVolume := int64(0)

	for _, order := range mpm.activeOrders {
		if order.Status == "pending" {
			pendingCount++
		}
		if order.Status == "confirmed" {
			totalVolume += order.NetAmount
		}
	}
	mpm.orderMutex.RUnlock()

	stats["pending_orders"] = pendingCount
	stats["total_volume_bkc"] = totalVolume
	stats["supported_chains"] = mpm.config.EnabledChains
	stats["commission_rates"] = mpm.commissionRates

	return stats, nil
}

// CancelOrder - отмена заказа
func (mpm *MultiChainPaymentManager) CancelOrder(ctx context.Context, orderID string) error {
	mpm.orderMutex.Lock()
	defer mpm.orderMutex.Unlock()

	order, exists := mpm.activeOrders[orderID]
	if !exists {
		return fmt.Errorf("order not found")
	}

	if order.Status != "pending" {
		return fmt.Errorf("order cannot be cancelled")
	}

	order.Status = "cancelled"
	delete(mpm.activeOrders, orderID)

	log.Printf("Order cancelled: %s", orderID)
	return nil
}

// contains - проверка наличия элемента в слайсе
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// toJSON - конвертация в JSON
func (mpm *MultiChainPaymentManager) toJSON(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}
