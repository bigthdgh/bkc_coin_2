package payments

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HeliusWebhookHandler - обработчик вебхуков от Helius
type HeliusWebhookHandler struct {
	paymentManager *MultiChainPaymentManager
	helius        *HeliusIntegration
}

// HeliusWebhookData - структура данных от вебхука Helius
type HeliusWebhookData struct {
	Signature    string      `json:"signature"`
	Slot         uint64      `json:"slot"`
	BlockTime    int64       `json:"blockTime"`
	Transaction  interface{} `json:"transaction"`
	Meta         interface{} `json:"meta"`
	Type         string      `json:"type"`
	Source       string      `json:"source"`
	Timestamp    time.Time   `json:"timestamp"`
}

// HeliusWebhookResponse - ответ на вебхук
type HeliusWebhookResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewHeliusWebhookHandler - создание обработчика вебхуков
func NewHeliusWebhookHandler(paymentManager *MultiChainPaymentManager, helius *HeliusIntegration) *HeliusWebhookHandler {
	return &HeliusWebhookHandler{
		paymentManager: paymentManager,
		helius:        helius,
	}
}

// HandleWebhook - обработка входящего вебхука от Helius
func (hwh *HeliusWebhookHandler) HandleWebhook(c *gin.Context) {
	var webhookData HeliusWebhookData

	// Декодируем JSON из тела запроса
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		log.Printf("Failed to decode webhook data: %v", err)
		c.JSON(http.StatusBadRequest, HeliusWebhookResponse{
			Status:  "error",
			Message: "Invalid JSON data",
		})
		return
	}

	log.Printf("Received Helius webhook: %s", webhookData.Signature)

	// Обрабатываем транзакцию
	err := hwh.processWebhookTransaction(webhookData)
	if err != nil {
		log.Printf("Failed to process webhook transaction: %v", err)
		c.JSON(http.StatusInternalServerError, HeliusWebhookResponse{
			Status:  "error",
			Message: "Failed to process transaction",
		})
		return
	}

	// Отправляем успешный ответ
	c.JSON(http.StatusOK, HeliusWebhookResponse{
		Status:  "success",
		Message: "Transaction processed successfully",
	})
}

// processWebhookTransaction - обработка транзакции из вебхука
func (hwh *HeliusWebhookHandler) processWebhookTransaction(webhookData HeliusWebhookData) error {
	// Извлекаем OrderID из транзакции
	orderID, err := hwh.extractOrderIDFromWebhookData(webhookData)
	if err != nil {
		return fmt.Errorf("failed to extract OrderID: %w", err)
	}

	if orderID == "" {
		log.Printf("No OrderID found in webhook transaction %s", webhookData.Signature)
		return nil // Не ошибка, просто не наша транзакция
	}

	// Проверяем что это платеж на наш кошелек
	isOurTransaction, err := hwh.validateTransaction(webhookData)
	if err != nil {
		return fmt.Errorf("failed to validate transaction: %w", err)
	}

	if !isOurTransaction {
		log.Printf("Transaction %s is not for our wallet", webhookData.Signature)
		return nil
	}

	// Подтверждаем платеж
	err = hwh.paymentManager.confirmPayment(c.Request.Context(), orderID, webhookData.Signature)
	if err != nil {
		return fmt.Errorf("failed to confirm payment: %w", err)
	}

	log.Printf("✅ Payment confirmed via webhook: OrderID=%s, Signature=%s", orderID, webhookData.Signature)
	return nil
}

// extractOrderIDFromWebhookData - извлечение OrderID из данных вебхука
func (hwh *HeliusWebhookHandler) extractOrderIDFromWebhookData(webhookData HeliusWebhookData) (string, error) {
	// Конвертируем транзакцию в JSON для анализа
	txData, err := json.Marshal(webhookData.Transaction)
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction: %w", err)
	}

	var txStruct struct {
		Message struct {
			Instructions []json.RawMessage `json:"instructions"`
		} `json:"message"`
	}

	err = json.Unmarshal(txData, &txStruct)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal transaction structure: %w", err)
	}

	// Ищем инструкцию Memo
	for _, instruction := range txStruct.Message.Instructions {
		var memoInstruction struct {
			ProgramID string `json:"programId"`
			Parsed    struct {
				Type string `json:"type"`
				Info string `json:"info"`
			} `json:"parsed"`
		}

		err = json.Unmarshal(instruction, &memoInstruction)
		if err != nil {
			continue
		}

		// Проверяем что это Memo программ
		if memoInstruction.ProgramID == "Memo1UhkJRfHyvLMcVucJwxXeuDx28UQ" && memoInstruction.Parsed.Type == "memo" {
			return memoInstruction.Parsed.Info, nil
		}
	}

	return "", nil
}

// validateTransaction - валидация транзакции
func (hwh *HeliusWebhookHandler) validateTransaction(webhookData HeliusWebhookData) (bool, error) {
	// Проверяем что транзакция успешна
	if webhookData.Meta == nil {
		return false, fmt.Errorf("no transaction metadata")
	}

	// TODO: Добавить проверку что транзакция на наш кошелек
	// Это требует анализа инструкций транзакции

	// Проверяем что транзакция не старше 5 минут
	if webhookData.BlockTime > 0 {
		txTime := time.Unix(webhookData.BlockTime, 0)
		if time.Since(txTime) > 5*time.Minute {
			return false, fmt.Errorf("transaction too old")
		}
	}

	return true, nil
}

// SetupRoutes - настройка роутов для вебхуков
func (hwh *HeliusWebhookHandler) SetupRoutes(router *gin.Engine) {
	// Основной эндпоинт для вебхуков Helius
	router.POST("/webhook/solana", hwh.HandleWebhook)

	// Тестовый эндпоинт для проверки вебхука
	router.POST("/webhook/solana/test", hwh.HandleTestWebhook)

	// Эндпоинт для проверки статуса вебхука
	router.GET("/webhook/solana/status", hwh.GetWebhookStatus)
}

// HandleTestWebhook - обработчик тестового вебхука
func (hwh *HeliusWebhookHandler) HandleTestWebhook(c *gin.Context) {
	log.Printf("🧪 Received test webhook from Helius")

	// Создаем тестовые данные
	testData := HeliusWebhookData{
		Signature: "test_signature_" + fmt.Sprintf("%d", time.Now().Unix()),
		Slot:      123456789,
		BlockTime: time.Now().Unix(),
		Type:      "test",
		Source:    "helius_test",
		Timestamp: time.Now(),
	}

	// Обрабатываем тестовые данные
	err := hwh.processWebhookTransaction(testData)
	if err != nil {
		log.Printf("Test webhook processing failed: %v", err)
		c.JSON(http.StatusInternalServerError, HeliusWebhookResponse{
			Status:  "error",
			Message: "Test webhook processing failed",
		})
		return
	}

	c.JSON(http.StatusOK, HeliusWebhookResponse{
		Status:  "success",
		Message: "Test webhook processed successfully",
	})
}

// GetWebhookStatus - получение статуса вебхука
func (hwh *HeliusWebhookHandler) GetWebhookStatus(c *gin.Context) {
	status := map[string]interface{}{
		"status": "active",
		"service": "helius_webhook",
		"timestamp": time.Now(),
		"endpoints": map[string]string{
			"webhook": "/webhook/solana",
			"test": "/webhook/solana/test",
			"status": "/webhook/solana/status",
		},
		"helius_config": map[string]interface{}{
			"admin_wallet": hwh.helius.adminWallet.String(),
			"ws_connected": hwh.helius.wsClient != nil,
		},
	}

	c.JSON(http.StatusOK, status)
}

// WebhookStats - статистика вебхуков
type WebhookStats struct {
	TotalReceived    int64     `json:"total_received"`
	SuccessfulProcessed int64    `json:"successful_processed"`
	FailedProcessed  int64     `json:"failed_processed"`
	LastReceived     time.Time  `json:"last_received"`
	LastSuccess      time.Time  `json:"last_success"`
	LastError       time.Time  `json:"last_error"`
	LastErrorMsg     string     `json:"last_error_msg"`
}

// GetWebhookStats - получение статистики вебхуков
func (hwh *HeliusWebhookHandler) GetWebhookStats(c *gin.Context) {
	// TODO: Реализовать сбор статистики
	stats := WebhookStats{
		TotalReceived:      0,
		SuccessfulProcessed: 0,
		FailedProcessed:    0,
		LastReceived:      time.Time{},
		LastSuccess:       time.Time{},
		LastError:        time.Time{},
		LastErrorMsg:      "",
	}

	c.JSON(http.StatusOK, stats)
}
