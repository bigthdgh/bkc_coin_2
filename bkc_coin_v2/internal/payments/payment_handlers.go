package payments

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PaymentHandlers - обработчики HTTP запросов для платежей
type PaymentHandlers struct {
	paymentManager *MultiChainPaymentManager
}

// NewPaymentHandlers - создание обработчиков платежей
func NewPaymentHandlers(pm *MultiChainPaymentManager) *PaymentHandlers {
	return &PaymentHandlers{
		paymentManager: pm,
	}
}

// CreatePaymentRequest - создание запроса на оплату
func (ph *PaymentHandlers) CreatePaymentRequest(c *gin.Context) {
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Получаем user ID из контекста (из JWT токена)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	req.UserID = userID.(int64)

	// Создаем заказ на оплату
	response, err := ph.paymentManager.CreatePaymentOrder(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPaymentStatus - получение статуса платежа
func (ph *PaymentHandlers) GetPaymentStatus(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID required"})
		return
	}

	order, err := ph.paymentManager.GetPaymentStatus(context.Background(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetUserPaymentHistory - история платежей пользователя
func (ph *PaymentHandlers) GetUserPaymentHistory(c *gin.Context) {
	// Получаем user ID из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Получаем лимит из query параметров
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	history, err := ph.paymentManager.GetUserPaymentHistory(context.Background(), userID.(int64), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payment history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

// GetPaymentStats - статистика платежей
func (ph *PaymentHandlers) GetPaymentStats(c *gin.Context) {
	// Проверяем права администратора
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	stats, err := ph.paymentManager.GetPaymentStats(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payment stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CancelPayment - отмена платежа
func (ph *PaymentHandlers) CancelPayment(c *gin.Context) {
	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID required"})
		return
	}

	// Получаем user ID из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Получаем заказ для проверки прав
	order, err := ph.paymentManager.GetPaymentStatus(context.Background(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Проверяем, что заказ принадлежит пользователю
	if order.UserID != userID.(int64) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	err = ph.paymentManager.CancelOrder(context.Background(), orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment cancelled successfully"})
}

// GetSupportedChains - получение поддерживаемых цепочек (только TON и USDT)
func (ph *PaymentHandlers) GetSupportedChains(c *gin.Context) {
	chains := map[string]interface{}{
		"chains": []map[string]interface{}{
			{
				"id":          "ton",
				"name":        "TON",
				"currency":    "TON",
				"decimals":    9,
				"description": "Native TON cryptocurrency",
				"icon":        "/icons/ton.png",
			},
			{
				"id":          "ton_usdt",
				"name":        "USDT (TON)",
				"currency":    "USDT",
				"decimals":    6,
				"description": "USDT on TON blockchain",
				"icon":        "/icons/usdt-ton.png",
			},
			{
				"id":          "solana_usdt",
				"name":        "USDT (Solana)",
				"currency":    "USDT",
				"decimals":    6,
				"description": "USDT on Solana blockchain",
				"icon":        "/icons/usdt-sol.png",
			},
		},
		"rates": map[string]float64{
			"TON":         1000.0, // 1 TON = 1000 BKC
			"USDT_TON":    1000.0, // 1 USDT = 1000 BKC
			"USDT_SOLANA": 1000.0, // 1 USDT = 1000 BKC
		},
		"notice": "Only TON and USDT are supported. No fiat currencies are available.",
	}

	c.JSON(http.StatusOK, chains)
}

// GetCommissionInfo - информация о комиссиях
func (ph *PaymentHandlers) GetCommissionInfo(c *gin.Context) {
	commission := map[string]interface{}{
		"rates": map[string]float64{
			"platform": ph.paymentManager.commissionRates.PlatformCommission,
			"nft":      ph.paymentManager.commissionRates.NFTCommission,
			"market":   ph.paymentManager.commissionRates.MarketCommission,
		},
		"referral": map[string]interface{}{
			"rate": ph.paymentManager.commissionRates.ReferralCommission,
			"min":  ph.paymentManager.commissionRates.MinCommission,
		},
		"examples": map[string]interface{}{
			"purchase": map[string]interface{}{
				"amount_bkc": 10000,
				"commission": 250, // 2.5%
				"net_amount": 9750,
				"referral":   25, // 10% от комиссии
			},
			"nft": map[string]interface{}{
				"amount_bkc": 50000,
				"commission": 2500, // 5%
				"net_amount": 47500,
				"referral":   250, // 10% от комиссии
			},
			"market": map[string]interface{}{
				"amount_bkc": 30000,
				"commission": 900, // 3%
				"net_amount": 29100,
				"referral":   90, // 10% от комиссии
			},
		},
	}

	c.JSON(http.StatusOK, commission)
}

// WebhookTON - вебхук для TON платежей
func (ph *PaymentHandlers) WebhookTON(c *gin.Context) {
	var webhookData struct {
		OrderID         string `json:"order_id"`
		TransactionHash string `json:"transaction_hash"`
		Amount          string `json:"amount"`
		Currency        string `json:"currency"`
		Status          string `json:"status"`
		Timestamp       int64  `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
		return
	}

	// Валидация вебхука (в реальном приложении здесь будет проверка подписи)

	// Подтверждаем платеж
	if webhookData.Status == "confirmed" {
		err := ph.paymentManager.confirmPayment(context.Background(), webhookData.OrderID, webhookData.TransactionHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm payment"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// WebhookSolana - вебхук для Solana платежей
func (ph *PaymentHandlers) WebhookSolana(c *gin.Context) {
	var webhookData struct {
		OrderID         string  `json:"order_id"`
		TransactionHash string  `json:"signature"`
		Amount          float64 `json:"amount"`
		Token           string  `json:"token"`
		Memo            string  `json:"memo"`
		Status          string  `json:"status"`
		Slot            uint64  `json:"slot"`
	}

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook data"})
		return
	}

	// Валидация вебхука

	// Подтверждаем платеж
	if webhookData.Status == "confirmed" {
		err := ph.paymentManager.confirmPayment(context.Background(), webhookData.OrderID, webhookData.TransactionHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm payment"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// ValidatePayment - валидация платежа перед созданием
func (ph *PaymentHandlers) ValidatePayment(c *gin.Context) {
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Получаем user ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	req.UserID = userID.(int64)

	// Валидация
	err := ph.paymentManager.validatePaymentRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	// Конвертация и расчет комиссии
	bkcAmount, err := ph.paymentManager.convertToBKC(req.Amount, req.Currency, req.Chain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": "Conversion failed",
		})
		return
	}

	commission := ph.paymentManager.calculateCommission(bkcAmount, req.Type)
	netAmount := bkcAmount - commission

	response := map[string]interface{}{
		"valid":         true,
		"bkc_amount":    bkcAmount,
		"commission":    commission,
		"net_amount":    netAmount,
		"exchange_rate": 1000.0, // 1 USDT/TON = 1000 BKC
	}

	c.JSON(http.StatusOK, response)
}

// EstimatePayment - оценка стоимости платежа
func (ph *PaymentHandlers) EstimatePayment(c *gin.Context) {
	chain := c.Query("chain")
	amountStr := c.Query("amount")
	paymentType := c.Query("type")

	if chain == "" || amountStr == "" || paymentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	// Определяем валюту
	currency := "USDT"
	if chain == "ton" {
		currency = "TON"
	}

	// Создаем тестовый запрос
	req := PaymentRequest{
		UserID:   0, // Не важен для оценки
		Type:     paymentType,
		Chain:    chain,
		Amount:   amount,
		Currency: currency,
	}

	// Валидация
	err = ph.paymentManager.validatePaymentRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	// Расчет
	bkcAmount, err := ph.paymentManager.convertToBKC(amount, currency, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Conversion failed"})
		return
	}

	commission := ph.paymentManager.calculateCommission(bkcAmount, paymentType)
	netAmount := bkcAmount - commission

	response := map[string]interface{}{
		"input": map[string]interface{}{
			"chain":    chain,
			"amount":   amount,
			"currency": currency,
			"type":     paymentType,
		},
		"output": map[string]interface{}{
			"bkc_amount":         bkcAmount,
			"commission":         commission,
			"net_amount":         netAmount,
			"commission_percent": ph.paymentManager.getCommissionRate(paymentType),
		},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// getCommissionRate - получение процента комиссии по типу
func (pm *MultiChainPaymentManager) getCommissionRate(paymentType string) float64 {
	switch paymentType {
	case "nft":
		return pm.commissionRates.NFTCommission
	case "market":
		return pm.commissionRates.MarketCommission
	default:
		return pm.commissionRates.PlatformCommission
	}
}

// RegisterRoutes - регистрация роутов
func (ph *PaymentHandlers) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payments")
	{
		// Публичные эндпоинты
		payments.POST("/create", ph.CreatePaymentRequest)
		payments.GET("/status/:order_id", ph.GetPaymentStatus)
		payments.GET("/history", ph.GetUserPaymentHistory)
		payments.POST("/cancel/:order_id", ph.CancelPayment)

		// Информационные эндпоинты
		payments.GET("/chains", ph.GetSupportedChains)
		payments.GET("/commission", ph.GetCommissionInfo)
		payments.POST("/validate", ph.ValidatePayment)
		payments.GET("/estimate", ph.EstimatePayment)

		// Вебхуки
		payments.POST("/webhook/ton", ph.WebhookTON)
		payments.POST("/webhook/solana", ph.WebhookSolana)

		// Административные эндпоинты
		payments.GET("/stats", ph.GetPaymentStats)
	}
}

// PaymentMiddleware - middleware для проверки платежных токенов
func PaymentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// В реальном приложении здесь будет проверка JWT токена
		// Для примера пропускаем все запросы

		// Устанавливаем тестовый user ID
		c.Set("user_id", int64(12345))
		c.Set("is_admin", false)

		c.Next()
	}
}

// AdminMiddleware - middleware для администраторов
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем права администратора
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
