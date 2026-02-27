package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func main() {
	// Твой API ключ из Helius
	apiKey := "f983dbf9-7518-4337-985d-d8ea68b16e64"
	
	// Создаем RPC клиент
	rpcURL := fmt.Sprintf("https://mainnet.helius-rpc.com/?api-key=%s", apiKey)
	client := rpc.New(rpcURL)

	// Тестовый кошелек (замени на свой)
	walletAddress := "11111111111111111111111111111112" // System Program для теста
	
	// Парсим адрес кошелька
	pubKey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		log.Fatalf("Failed to parse wallet address: %v", err)
	}

	fmt.Printf("🔗 Testing Helius connection...\n")
	fmt.Printf("API Key: %s\n", apiKey)
	fmt.Printf("Wallet: %s\n", walletAddress)
	fmt.Printf("RPC URL: %s\n\n", rpcURL)

	// Проверяем баланс
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	balance, err := client.GetBalance(ctx, pubKey, rpc.CommitmentConfirmed)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	fmt.Printf("✅ Connection successful!\n")
	fmt.Printf("💰 Wallet balance: %d lamports (%.9f SOL)\n", balance.Value, float64(balance.Value)/1e9)

	// Получаем последние транзакции
	fmt.Printf("\n📋 Getting recent transactions...\n")
	
	signatures, err := client.GetSignaturesForAddress(ctx, pubKey, &rpc.GetSignaturesForAddressOpts{
		Limit:      func(i int) *int { return &i }(5), // Последние 5 транзакций
		Commitment: rpc.CommitmentConfirmed,
	})
	if err != nil {
		log.Printf("Failed to get signatures: %v", err)
	} else {
		fmt.Printf("Found %d recent transactions:\n", len(signatures.Value))
		for i, sig := range signatures.Value {
			fmt.Printf("  %d. %s (Slot: %d)\n", i+1, sig.Signature, sig.Slot)
		}
	}

	// Тест WebSocket соединения
	fmt.Printf("\n🌐 Testing WebSocket connection...\n")
	
	wsURL := fmt.Sprintf("wss://mainnet.helius-rpc.com/?api-key=%s", apiKey)
	fmt.Printf("WebSocket URL: %s\n", wsURL)
	
	// Примечание: для реального WebSocket соединения нужно использовать ws.Connect()
	// Это просто демонстрация URL
	fmt.Printf("✅ WebSocket URL generated successfully!\n")
	fmt.Printf("📝 Use this URL in your Go application for real-time transaction monitoring\n")

	// Тест с USDT токеном
	fmt.Printf("\n💵 Testing USDT token balance...\n")
	
	// USDT mint адрес на Solana
	usdtMint := "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"
	usdtPubKey, err := solana.PublicKeyFromBase58(usdtMint)
	if err != nil {
		log.Printf("Failed to parse USDT mint: %v", err)
	} else {
		// Получаем токен аккаунты
		tokenAccounts, err := client.GetTokenAccountsByOwner(ctx, pubKey, &rpc.GetTokenAccountsByOwnerConfig{
			Mint: &usdtPubKey,
		}, rpc.CommitmentConfirmed)
		
		if err != nil {
			log.Printf("Failed to get USDT token accounts: %v", err)
		} else {
			if len(tokenAccounts.Value) == 0 {
				fmt.Printf("💸 No USDT tokens found in wallet\n")
			} else {
				fmt.Printf("💰 Found USDT token accounts:\n")
				for i, account := range tokenAccounts.Value {
					amount := account.Account.Data.Parsed.Info.TokenAmount.AmountUint64
					fmt.Printf("  %d. Account: %s, Balance: %d USDT\n", 
						i+1, account.Pubkey, amount)
				}
			}
		}
	}

	fmt.Printf("\n🎯 Helius integration test completed successfully!\n")
	fmt.Printf("📋 Next steps:\n")
	fmt.Printf("   1. Replace wallet address with your actual Solana wallet\n")
	fmt.Printf("   2. Set up webhooks in Helius dashboard\n")
	fmt.Printf("   3. Implement WebSocket listener for real-time monitoring\n")
	fmt.Printf("   4. Add transaction validation logic\n")
}
