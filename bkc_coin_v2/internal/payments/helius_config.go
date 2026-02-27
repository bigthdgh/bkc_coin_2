package payments

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// HeliusConfigManager - менеджер конфигурации Helius
type HeliusConfigManager struct {
	APIKey      string `json:"api_key"`
	AdminWallet  string `json:"admin_wallet"`
	USDTMint     string `json:"usdt_mint"`
	WebhookURL   string `json:"webhook_url"`
	ServerURL    string `json:"server_url"`
}

// NewHeliusConfigManager - создание менеджера конфигурации
func NewHeliusConfigManager() *HeliusConfigManager {
	return &HeliusConfigManager{
		APIKey:     "f983dbf9-7518-4337-985d-d8ea68b16e64", // Твой API ключ
		USDTMint:   "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB", // USDT на Solana
		ServerURL:  "https://your-domain.com", // Замени на свой домен
	}
}

// LoadFromFile - загрузка конфигурации из файла
func (hcm *HeliusConfigManager) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Config file %s not found, using defaults", filename)
			return hcm.SaveToFile(filename)
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = json.Unmarshal(data, hcm)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	log.Printf("Helius configuration loaded from %s", filename)
	return nil
}

// SaveToFile - сохранение конфигурации в файл
func (hcm *HeliusConfigManager) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(hcm, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("Helius configuration saved to %s", filename)
	return nil
}

// GetRPCURL - получение RPC URL
func (hcm *HeliusConfigManager) GetRPCURL() string {
	return fmt.Sprintf("https://mainnet.helius-rpc.com/?api-key=%s", hcm.APIKey)
}

// GetWebSocketURL - получение WebSocket URL
func (hcm *HeliusConfigManager) GetWebSocketURL() string {
	return fmt.Sprintf("wss://mainnet.helius-rpc.com/?api-key=%s", hcm.APIKey)
}

// GetWebhookURL - получение URL для вебхука
func (hcm *HeliusConfigManager) GetWebhookURL() string {
	if hcm.WebhookURL != "" {
		return hcm.WebhookURL
	}
	return fmt.Sprintf("%s/webhook/solana", hcm.ServerURL)
}

// Validate - валидация конфигурации
func (hcm *HeliusConfigManager) Validate() error {
	if hcm.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if hcm.AdminWallet == "" {
		return fmt.Errorf("admin wallet is required")
	}

	if hcm.USDTMint == "" {
		return fmt.Errorf("USDT mint address is required")
	}

	if hcm.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}

	return nil
}

// PrintConfig - вывод конфигурации в лог
func (hcm *HeliusConfigManager) PrintConfig() {
	log.Printf("🔧 Helius Configuration:")
	log.Printf("   API Key: %s***", hcm.APIKey[:8])
	log.Printf("   Admin Wallet: %s", hcm.AdminWallet)
	log.Printf("   USDT Mint: %s", hcm.USDTMint)
	log.Printf("   Server URL: %s", hcm.ServerURL)
	log.Printf("   Webhook URL: %s", hcm.GetWebhookURL())
	log.Printf("   RPC URL: %s***", hcm.GetRPCURL()[:50])
	log.Printf("   WebSocket URL: %s***", hcm.GetWebSocketURL()[:50])
}

// GetWebhookConfig - получение конфигурации для создания вебхука в Helius
func (hcm *HeliusConfigManager) GetWebhookConfig() map[string]interface{} {
	return map[string]interface{}{
		"webhookURL":       hcm.GetWebhookURL(),
		"accountAddresses": []string{hcm.AdminWallet},
		"webhookType":     "enhanced",
		"txnType":         "any",
	}
}

// ExampleWebhookPayload - пример payload от вебхука
type ExampleWebhookPayload struct {
	Signature string `json:"signature"`
	Slot      uint64 `json:"slot"`
	BlockTime int64  `json:"blockTime"`
	Transaction struct {
		Message struct {
			Instructions []json.RawMessage `json:"instructions"`
		} `json:"message"`
	} `json:"transaction"`
	Meta interface{} `json:"meta"`
	Type string `json:"type"`
	Source string `json:"source"`
}

// PrintExamplePayload - вывод примера payload
func (hcm *HeliusConfigManager) PrintExamplePayload() {
	example := ExampleWebhookPayload{
		Signature: "5j7sL...",
		Slot:      123456789,
		BlockTime:  1640995200,
		Type:      "confirmed",
		Source:    "helius",
	}

	payload, _ := json.MarshalIndent(example, "", "  ")
	log.Printf("📋 Example Helius Webhook Payload:")
	log.Printf("%s", string(payload))
}

// SetupInstructions - инструкции по настройке
func (hcm *HeliusConfigManager) SetupInstructions() {
	log.Printf("🚀 Helius Setup Instructions:")
	log.Printf("1. Go to https://console.helius.dev/")
	log.Printf("2. Create a new project or use existing one")
	log.Printf("3. Copy your API key: %s***", hcm.APIKey[:8])
	log.Printf("4. Set up webhooks:")
	log.Printf("   - URL: %s", hcm.GetWebhookURL())
	log.Printf("   - Type: Enhanced")
	log.Printf("   - Account: %s", hcm.AdminWallet)
	log.Printf("5. Test the connection using: go run cmd/helius_test/main.go")
	log.Printf("6. Start the payment system with your admin wallet")
}

// SecurityNotes - заметки по безопасности
func (hcm *HeliusConfigManager) SecurityNotes() {
	log.Printf("🛡️ Security Notes:")
	log.Printf("1. Never expose your API key in public repositories")
	log.Printf("2. Use environment variables for production deployment")
	log.Printf("3. Monitor your API usage in Helius dashboard")
	log.Printf("4. Set up rate limiting for webhook endpoints")
	log.Printf("5. Validate all incoming webhook signatures")
	log.Printf("6. Use HTTPS for all webhook URLs")
}
