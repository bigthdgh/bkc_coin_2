package main

import (
	"fmt"
	"log"
	"net/http"

	"bkc_coin_v2/internal/api"
	"bkc_coin_v2/internal/ton"
)

func main() {
	// Инициализация handlers
	p2pHandler := api.NewP2PHandler()
	nftHandler := api.NewNFTHandler()
	gamesHandler := api.NewGamesHandler(nil) // TODO: передать gamesManager
	exchangeHandler := api.NewExchangeHandler()
	creditsHandler := api.NewCreditsHandler(nil)           // TODO: передать creditsManager
	subscriptionHandler := api.NewSubscriptionHandler(nil) // TODO: передать subscriptionManager

	// Регистрация роутов
	mux := http.NewServeMux()
	p2pHandler.RegisterRoutes(mux)
	nftHandler.RegisterRoutes(mux)
	gamesHandler.RegisterRoutes(mux)
	exchangeHandler.RegisterRoutes(mux)
	creditsHandler.RegisterRoutes(mux)
	subscriptionHandler.RegisterRoutes(mux)

	// Добавляем health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintf(w, "BKC Coin API is running")
	})

	// Запуск сервера
	port := ":8080"
	fmt.Printf("🚀 BKC Coin API Server starting on port %s\n", port)
	fmt.Printf("📊 P2P Market: /api/v1/p2p/*\n")
	fmt.Printf("💎 NFT Shop: /api/v1/nft/*\n")
	fmt.Printf("🎮 Games: /api/v1/games/*\n")
	fmt.Printf("💱 Exchange: /api/v1/exchange/*\n")
	fmt.Printf("💳 Credits: /api/v1/credits/*\n")
	fmt.Printf("👑 Subscriptions: /api/v1/subscription/*\n")
	fmt.Printf("🪙 TON Webhooks: /webhook/*\n")
	fmt.Printf("💰 Your TON Wallet: %s\n", ton.COMMISSION_WALLET)

	log.Fatal(http.ListenAndServe(port, mux))
}
