package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Enhanced server with RateManager integration
type EnhancedServer struct {
	serviceType string
	httpServer  *http.Server
	rateManager *RateManager
}

// RateManager for Binance API
type RateManager struct {
	currentRate float64
	lastUpdate  time.Time
	client      *http.Client
}

// Binance API response
type BinanceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// NewRateManager creates new rate manager
func NewRateManager() *RateManager {
	return &RateManager{
		currentRate: 5.0, // Default rate
		lastUpdate:  time.Now(),
		client:      &http.Client{Timeout: 30 * time.Second},
	}
}

// UpdateTONRate updates TON/USD rate from Binance
func (rm *RateManager) UpdateTONRate() error {
	url := "https://api.binance.com/api/v3/ticker/price?symbol=TONUSDT"

	resp, err := rm.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch rate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("binance API returned status: %d", resp.StatusCode)
	}

	var binanceResp BinanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&binanceResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	newRate, err := strconv.ParseFloat(binanceResp.Price, 64)
	if err != nil {
		return fmt.Errorf("failed to parse rate: %w", err)
	}

	rm.currentRate = newRate
	rm.lastUpdate = time.Now()

	log.Printf("TON/USD rate updated: %.6f", newRate)
	return nil
}

// GetTONRate returns current TON/USD rate
func (rm *RateManager) GetTONRate() float64 {
	return rm.currentRate
}

// ConvertTONtoBKC converts TON to BKC
func (rm *RateManager) ConvertTONtoBKC(tonAmount float64) float64 {
	// 1 BKC = $0.001
	return tonAmount * rm.currentRate / 0.001
}

// ConvertBKCtoTON converts BKC to TON
func (rm *RateManager) ConvertBKCtoTON(bkcAmount float64) float64 {
	return (bkcAmount * 0.001) / rm.currentRate
}

// UpdateNFTPrices updates NFT prices based on TON rate
func (rm *RateManager) UpdateNFTPrices() error {
	// This would update NFT prices in database
	// For now, just log the action
	log.Printf("Updating NFT prices based on TON rate: %.6f", rm.currentRate)
	return nil
}

// Enhanced request/response structures
type EnhancedTapRequest struct {
	UserID    int64 `json:"user_id"`
	TapCount  int   `json:"tap_count"`
	Timestamp int64 `json:"timestamp"`
}

type EnhancedTapResponse struct {
	Success    bool    `json:"success"`
	Balance    float64 `json:"balance"`
	Energy     int     `json:"energy"`
	Reward     float64 `json:"reward"`
	Experience int     `json:"experience"`
	Message    string  `json:"message"`
}

type RatesResponse struct {
	TON_USD    float64 `json:"ton_usd"`
	BKC_USD    float64 `json:"bkc_usd"`
	TON_BKC    float64 `json:"ton_bkc"`
	LastUpdate string  `json:"last_update"`
	Source     string  `json:"source"`
}

type NFTItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	PriceBKC    float64 `json:"price_bkc"`
	PriceTON    float64 `json:"price_ton"`
	Rarity      string  `json:"rarity"`
}

type Transaction struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"user_id"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

// NewEnhancedServer creates new enhanced server
func NewEnhancedServer(serviceType string) *EnhancedServer {
	return &EnhancedServer{
		serviceType: serviceType,
		rateManager: NewRateManager(),
	}
}

// SetupRoutes configures all routes
func (es *EnhancedServer) SetupRoutes() chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", es.HealthHandler)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// User routes
		r.Route("/user", func(r chi.Router) {
			r.Get("/{id}", es.GetUserHandler)
			r.Post("/{id}/tap", es.TapHandler)
		})

		// Rates routes
		r.Route("/rates", func(r chi.Router) {
			r.Get("/current", es.GetCurrentRatesHandler)
			r.Post("/update", es.UpdateRatesHandler)
			r.Get("/history", es.GetRatesHistoryHandler)
		})

		// NFT routes
		r.Route("/nfts", func(r chi.Router) {
			r.Get("/", es.GetNFTsHandler)
			r.Post("/{id}/list", es.ListNFTHandler)
			r.Get("/{id}", es.GetNFTHandler)
		})

		// Transaction routes
		r.Route("/transactions", func(r chi.Router) {
			r.Get("/{id}", es.GetTransactionsHandler)
			r.Post("/", es.CreateTransactionHandler)
		})

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(es.AdminMiddleware)
			r.Post("/rates/set", es.SetRateHandler)
			r.Get("/stats", es.GetAdminStatsHandler)
		})
	})

	return r
}

// Health handler
func (es *EnhancedServer) HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":       "healthy",
		"service_type": es.serviceType,
		"timestamp":    time.Now().Unix(),
		"version":      "2.0.0",
		"rates": map[string]interface{}{
			"ton_usd":     es.rateManager.GetTONRate(),
			"last_update": es.rateManager.lastUpdate.Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Tap handler with enhanced validation
func (es *EnhancedServer) TapHandler(w http.ResponseWriter, r *http.Request) {
	var req EnhancedTapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Enhanced validation
	if req.UserID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if req.TapCount <= 0 || req.TapCount > 100 {
		http.Error(w, "Invalid tap count", http.StatusBadRequest)
		return
	}

	// Simulate tap processing
	reward := float64(req.TapCount) * 0.1 // 0.1 BKC per tap
	experience := req.TapCount * 5

	response := EnhancedTapResponse{
		Success:    true,
		Balance:    1000 + reward,
		Energy:     1000 - req.TapCount,
		Reward:     reward,
		Experience: experience,
		Message:    fmt.Sprintf("Successfully tapped %d times", req.TapCount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get current rates handler
func (es *EnhancedServer) GetCurrentRatesHandler(w http.ResponseWriter, r *http.Request) {
	tonRate := es.rateManager.GetTONRate()

	response := RatesResponse{
		TON_USD:    tonRate,
		BKC_USD:    0.001,
		TON_BKC:    es.rateManager.ConvertTONtoBKC(1),
		LastUpdate: es.rateManager.lastUpdate.Format(time.RFC3339),
		Source:     "binance",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Update rates handler
func (es *EnhancedServer) UpdateRatesHandler(w http.ResponseWriter, r *http.Request) {
	if err := es.rateManager.UpdateTONRate(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update rates: %v", err), http.StatusInternalServerError)
		return
	}

	// Update NFT prices based on new rate
	if err := es.rateManager.UpdateNFTPrices(); err != nil {
		log.Printf("Failed to update NFT prices: %v", err)
	}

	response := map[string]interface{}{
		"success":     true,
		"message":     "Rates updated successfully",
		"ton_usd":     es.rateManager.GetTONRate(),
		"ton_bkc":     es.rateManager.ConvertTONtoBKC(1),
		"last_update": es.rateManager.lastUpdate.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get rates history handler
func (es *EnhancedServer) GetRatesHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Mock history data
	history := []map[string]interface{}{
		{
			"rate":      es.rateManager.GetTONRate(),
			"timestamp": es.rateManager.lastUpdate.Format(time.RFC3339),
			"source":    "binance",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// Get NFTs handler
func (es *EnhancedServer) GetNFTsHandler(w http.ResponseWriter, r *http.Request) {
	tonRate := es.rateManager.GetTONRate()

	// Mock NFT data with dynamic pricing
	nfts := []NFTItem{
		{
			ID:          1,
			Name:        "Golden Dragon",
			Description: "Legendary NFT with +50% tap bonus",
			Image:       "/assets/nft1.png",
			PriceBKC:    10000,
			PriceTON:    10000 * 0.001 / tonRate,
			Rarity:      "Legendary",
		},
		{
			ID:          2,
			Name:        "Silver Phoenix",
			Description: "Epic NFT with +25% tap bonus",
			Image:       "/assets/nft2.png",
			PriceBKC:    5000,
			PriceTON:    5000 * 0.001 / tonRate,
			Rarity:      "Epic",
		},
		{
			ID:          3,
			Name:        "Bronze Tiger",
			Description: "Rare NFT with +10% tap bonus",
			Image:       "/assets/nft3.png",
			PriceBKC:    2000,
			PriceTON:    2000 * 0.001 / tonRate,
			Rarity:      "Rare",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nfts)
}

// List NFT handler
func (es *EnhancedServer) ListNFTHandler(w http.ResponseWriter, r *http.Request) {
	// Mock listing response
	response := map[string]interface{}{
		"success":    true,
		"message":    "NFT listed successfully",
		"listing_id": fmt.Sprintf("list_%d", time.Now().Unix()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get NFT handler
func (es *EnhancedServer) GetNFTHandler(w http.ResponseWriter, r *http.Request) {
	// Mock single NFT response
	nft := NFTItem{
		ID:          1,
		Name:        "Golden Dragon",
		Description: "Legendary NFT with +50% tap bonus",
		Image:       "/assets/nft1.png",
		PriceBKC:    10000,
		PriceTON:    10000 * 0.001 / es.rateManager.GetTONRate(),
		Rarity:      "Legendary",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nft)
}

// Get transactions handler
func (es *EnhancedServer) GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Mock transactions data
	transactions := []Transaction{
		{
			ID:        1,
			UserID:    12345,
			Type:      "tap",
			Amount:    10,
			Currency:  "BKC",
			Status:    "completed",
			Timestamp: time.Now().Format(time.RFC3339),
		},
		{
			ID:        2,
			UserID:    12345,
			Type:      "nft_purchase",
			Amount:    5000,
			Currency:  "BKC",
			Status:    "completed",
			Timestamp: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// Create transaction handler
func (es *EnhancedServer) CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	// Mock transaction creation
	response := map[string]interface{}{
		"success":        true,
		"transaction_id": fmt.Sprintf("tx_%d", time.Now().Unix()),
		"message":        "Transaction created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Admin middleware
func (es *EnhancedServer) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple admin check (in production, use proper authentication)
		adminID := r.Header.Get("X-Admin-ID")
		if adminID != "8425434588" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Set rate handler (admin only)
func (es *EnhancedServer) SetRateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Rate   float64 `json:"rate"`
		Reason string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Rate <= 0 {
		http.Error(w, "Invalid rate", http.StatusBadRequest)
		return
	}

	es.rateManager.currentRate = req.Rate
	es.rateManager.lastUpdate = time.Now()

	// Update NFT prices
	es.rateManager.UpdateNFTPrices()

	response := map[string]interface{}{
		"success":     true,
		"message":     "Rate set successfully",
		"rate":        req.Rate,
		"reason":      req.Reason,
		"last_update": es.rateManager.lastUpdate.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get admin stats handler
func (es *EnhancedServer) GetAdminStatsHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"total_users":        10000,
		"active_users":       2500,
		"total_transactions": 50000,
		"total_volume":       1000000,
		"current_rate":       es.rateManager.GetTONRate(),
		"last_update":        es.rateManager.lastUpdate.Format(time.RFC3339),
		"service_uptime":     time.Since(time.Now().Add(-24 * time.Hour)).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get user handler
func (es *EnhancedServer) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Mock user data
	user := map[string]interface{}{
		"id":           userID,
		"balance":      1000.0,
		"energy":       1000,
		"max_energy":   1000,
		"tap_power":    1,
		"level":        1,
		"experience":   0,
		"referrals":    0,
		"achievements": []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Start background tasks
func (es *EnhancedServer) StartBackgroundTasks(ctx context.Context) {
	// Update rates every 5 minutes
	rateUpdater := time.NewTicker(5 * time.Minute)
	defer rateUpdater.Stop()

	// Initial rate update
	es.rateManager.UpdateTONRate()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rateUpdater.C:
			if err := es.rateManager.UpdateTONRate(); err != nil {
				log.Printf("Failed to update rates: %v", err)
			} else {
				// Update NFT prices after rate update
				es.rateManager.UpdateNFTPrices()
			}
		}
	}
}

// Start server
func (es *EnhancedServer) Start() error {
	router := es.SetupRoutes()

	es.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start background tasks
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go es.StartBackgroundTasks(ctx)

	log.Printf("Starting enhanced server on port 8080")
	return es.httpServer.ListenAndServe()
}

// Graceful shutdown
func (es *EnhancedServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return es.httpServer.Shutdown(ctx)
}

func main() {
	serviceType := os.Getenv("SERVICE_TYPE")
	if serviceType == "" {
		serviceType = "enhanced"
	}

	server := NewEnhancedServer(serviceType)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		if err := server.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server stopped")
}
