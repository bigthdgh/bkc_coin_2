package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// HeliusIntegration - интеграция с Helius для Solana
type HeliusIntegration struct {
	apiKey      string
	wsURL       string
	rpcClient   *rpc.Client
	wsClient    *ws.Client
	adminWallet solana.PublicKey
	paymentManager *MultiChainPaymentManager
}

// HeliusConfig - конфигурация Helius
type HeliusConfig struct {
	APIKey      string `json:"api_key"`
	AdminWallet string `json:"admin_wallet"`
}

// HeliusTransaction - структура транзакции от Helius
type HeliusTransaction struct {
	Signature    string      `json:"signature"`
	Slot         uint64      `json:"slot"`
	BlockTime    int64       `json:"blockTime"`
	Transaction  interface{} `json:"transaction"`
	Meta         interface{} `json:"meta"`
	Type         string      `json:"type"`
	Source       string      `json:"source"`
}

// HeliusLogEntry - запись лога транзакции
type HeliusLogEntry struct {
	Signature string `json:"signature"`
	Err       error  `json:"error"`
}

// NewHeliusIntegration - создание интеграции с Helius
func NewHeliusIntegration(config HeliusConfig, paymentManager *MultiChainPaymentManager) (*HeliusIntegration, error) {
	h := &HeliusIntegration{
		apiKey:         config.APIKey,
		wsURL:          fmt.Sprintf("wss://mainnet.helius-rpc.com/?api-key=%s", config.APIKey),
		paymentManager: paymentManager,
	}

	// Парсим админский кошелек
	adminWallet, err := solana.PublicKeyFromBase58(config.AdminWallet)
	if err != nil {
		return nil, fmt.Errorf("invalid admin wallet: %w", err)
	}
	h.adminWallet = adminWallet

	// Создаем RPC клиент
	rpcURL := fmt.Sprintf("https://mainnet.helius-rpc.com/?api-key=%s", config.APIKey)
	h.rpcClient = rpc.New(rpcURL)

	// Проверяем соединение
	err = h.testConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Helius: %w", err)
	}

	log.Printf("Helius integration initialized with admin wallet: %s", config.AdminWallet)
	return h, nil
}

// testConnection - проверка соединения с Helius
func (h *HeliusIntegration) testConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Проверяем баланс админского кошелька
	balance, err := h.rpcClient.GetBalance(ctx, h.adminWallet, rpc.CommitmentConfirmed)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	log.Printf("Helius connection successful. Admin wallet balance: %d lamports", balance.Value)
	return nil
}

// StartWebSocketListener - запуск WebSocket слушателя
func (h *HeliusIntegration) StartWebSocketListener() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Подключаемся к WebSocket
	client, err := ws.Connect(ctx, h.wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	h.wsClient = client

	log.Println("WebSocket listener started. Listening for Solana transactions...")

	// Подписываемся на логи админского кошелька
	sub, err := client.LogsSubscribeMentions(h.adminWallet, "confirmed")
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	// Запускаем обработчик в отдельной горутине
	go h.handleWebSocketMessages(sub)

	return nil
}

// handleWebSocketMessages - обработка WebSocket сообщений
func (h *HeliusIntegration) handleWebSocketMessages(sub *ws.LogSubscription) {
	for {
		msg, err := sub.Recv()
		if err != nil {
			log.Printf("WebSocket error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Обрабатываем полученное сообщение
		go h.processTransactionMessage(msg)
	}
}

// processTransactionMessage - обработка сообщения о транзакции
func (h *HeliusIntegration) processTransactionMessage(msg interface{}) {
	// Конвертируем сообщение в JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	var transaction HeliusTransaction
	err = json.Unmarshal(jsonData, &transaction)
	if err != nil {
		log.Printf("Failed to unmarshal transaction: %v", err)
		return
	}

	log.Printf("Received transaction: %s", transaction.Signature)

	// Получаем детальную информацию о транзакции
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := h.rpcClient.GetTransaction(ctx, solana.MustSignatureFromBase58(transaction.Signature), &rpc.GetTransactionOpts{
		Encoding: solana.EncodingJSON,
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		log.Printf("Failed to get transaction details: %v", err)
		return
	}

	// Извлекаем OrderID из Memo
	orderID, err := h.extractOrderIDFromTransaction(tx)
	if err != nil {
		log.Printf("Failed to extract OrderID from transaction %s: %v", transaction.Signature, err)
		return
	}

	if orderID == "" {
		log.Printf("No OrderID found in transaction %s", transaction.Signature)
		return
	}

	// Проверяем и подтверждаем платеж
	err = h.paymentManager.confirmPayment(context.Background(), orderID, transaction.Signature)
	if err != nil {
		log.Printf("Failed to confirm payment for order %s: %v", orderID, err)
		return
	}

	log.Printf("Payment confirmed for order %s with transaction %s", orderID, transaction.Signature)
}

// extractOrderIDFromTransaction - извлечение OrderID из Memo транзакции
func (h *HeliusIntegration) extractOrderIDFromTransaction(tx *rpc.GetTransactionResult) (string, error) {
	if tx == nil || tx.Meta == nil || tx.Transaction == nil {
		return "", fmt.Errorf("invalid transaction data")
	}

	// Ищем инструкцию Memo в транзакции
	transaction, ok := tx.Transaction.(solana.Transaction)
	if !ok {
		// Пробуем распарсить из JSON
		txData, err := json.Marshal(tx.Transaction)
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

		return "", fmt.Errorf("no memo instruction found")
	}

	// Для распарсенной транзакции ищем инструкции
	for _, instruction := range transaction.Message.Instructions {
		// Проверяем что это инструкция Memo
		if memo, ok := instruction.(solana.CompiledInstruction); ok {
			// Program ID для Memo программы
			memoProgramID := solana.MustPublicKeyFromBase58("Memo1UhkJRfHyvLMcVucJwxXeuDx28UQ")
			
			if memo.ProgramIDIndex == 0 && len(transaction.Message.AccountKeys) > 0 && transaction.Message.AccountKeys[0] == memoProgramID {
				// Извлекаем данные из инструкции
				if len(memo.Data) > 0 {
					return string(memo.Data), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no memo instruction found")
}

// GetAccountBalance - получение баланса аккаунта
func (h *HeliusIntegration) GetAccountBalance(walletAddress string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubKey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return 0, fmt.Errorf("invalid wallet address: %w", err)
	}

	balance, err := h.rpcClient.GetBalance(ctx, pubKey, rpc.CommitmentConfirmed)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance.Value, nil
}

// GetTokenBalance - получение баланса токена (USDT)
func (h *HeliusIntegration) GetTokenBalance(walletAddress, tokenMint string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubKey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return 0, fmt.Errorf("invalid wallet address: %w", err)
	}

	mint, err := solana.PublicKeyFromBase58(tokenMint)
	if err != nil {
		return 0, fmt.Errorf("invalid token mint: %w", err)
	}

	// Получаем все токен аккаунты
	tokenAccounts, err := h.rpcClient.GetTokenAccountsByOwner(ctx, pubKey, &rpc.GetTokenAccountsByOwnerConfig{
		Mint: &mint,
	}, rpc.CommitmentConfirmed)
	if err != nil {
		return 0, fmt.Errorf("failed to get token accounts: %w", err)
	}

	if len(tokenAccounts.Value) == 0 {
		return 0, nil
	}

	// Возвращаем баланс первого найденного аккаунта
	return tokenAccounts.Value[0].Account.Data.Parsed.Info.TokenAmount.AmountUint64, nil
}

// GetRecentTransactions - получение последних транзакций кошелька
func (h *HeliusIntegration) GetRecentTransactions(walletAddress string, limit int) ([]HeliusTransaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pubKey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid wallet address: %w", err)
	}

	// Получаем сигнатуры транзакций
	signatures, err := h.rpcClient.GetSignaturesForAddress(ctx, pubKey, &rpc.GetSignaturesForAddressOpts{
		Limit:      &limit,
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures: %w", err)
	}

	var transactions []HeliusTransaction

	for _, sig := range signatures.Value {
		tx, err := h.rpcClient.GetTransaction(ctx, sig.Signature, &rpc.GetTransactionOpts{
			Encoding: solana.EncodingJSON,
			Commitment: rpc.CommitmentConfirmed,
		})
		if err != nil {
			log.Printf("Failed to get transaction %s: %v", sig.Signature, err)
			continue
		}

		transaction := HeliusTransaction{
			Signature: sig.Signature.String(),
			Slot:     tx.Slot,
			BlockTime: *tx.BlockTime,
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// Shutdown - остановка WebSocket соединения
func (h *HeliusIntegration) Shutdown() error {
	if h.wsClient != nil {
		return h.wsClient.Close()
	}
	return nil
}

// CreateWebhookURL - создание URL для вебхука
func (h *HeliusIntegration) CreateWebhookURL(serverURL string) string {
	return fmt.Sprintf("%s/webhook/solana", serverURL)
}

// GetWebhookConfig - получение конфигурации для вебхука
func (h *HeliusIntegration) GetWebhookConfig(serverURL string) map[string]interface{} {
	return map[string]interface{}{
		"webhookURL":    h.CreateWebhookURL(serverURL),
		"accountAddresses": []string{h.adminWallet.String()},
		"webhookType":   "enhanced",
		"txnType":       "any",
	}
}

// ValidateUSDTTransaction - валидация USDT транзакции
func (h *HeliusIntegration) ValidateUSDTTransaction(signature, expectedAmount string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature: %w", err)
	}

	tx, err := h.rpcClient.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingJSON,
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Проверяем что транзакция успешна
	if tx.Meta == nil || tx.Meta.Err != nil {
		return false, fmt.Errorf("transaction failed")
	}

	// TODO: Добавить проверку суммы USDT
	// Это требует парсинга инструкций SPL токена

	return true, nil
}
