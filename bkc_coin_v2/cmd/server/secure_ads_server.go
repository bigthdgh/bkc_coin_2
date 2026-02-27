package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
	"github.com/gorilla/websocket"
	"bkc_coin_v2/internal/monetization"
)

// SecureAdsServer handles secure advertisement server
type SecureAdsServer struct {
	adsIntegration *monetization.SecureAdsIntegration
	upgrader       websocket.Upgrader
}

// NewSecureAdsServer creates a new secure ads server
func NewSecureAdsServer(redisClient *redis.Client) *SecureAdsServer {
	// Initialize secure ads integration
	adsgramSecret := "your-adsgram-webhook-secret" // Should be from env
	rewardURL := "https://your-domain.com/webhook/adsgram"
	
	adsIntegration := monetization.NewSecureAdsIntegration(redisClient, adsgramSecret, rewardURL)
	adsIntegration.InitializeSecureConfigs()
	
	// Start WebSocket server
	adsIntegration.StartWebSocketServer(8081)
	
	return &SecureAdsServer{
		adsIntegration: adsIntegration,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Configure properly for production
			},
		},
	}
}

// SetupRoutes sets up the secure ads routes
func (s *SecureAdsServer) SetupRoutes() chi.Router {
	r := chi.NewRouter()
	
	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure properly for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Ads API routes
	r.Route("/api/v1/ads", func(r chi.Router) {
		r.Get("/rewards", s.GetAvailableRewards)
		r.Get("/stats", s.GetAdStats)
		r.Post("/session/start", s.StartAdSession)
		r.Get("/next-time/{trigger}", s.GetNextAdTime)
	})

	// Webhook endpoint for Adsgram
	r.Post("/webhook/adsgram", s.HandleAdsgramWebhook)

	// WebSocket endpoint
	r.Get("/ws", s.HandleWebSocket)

	return r
}

// StartAdSession starts a new ad session
func (s *SecureAdsServer) StartAdSession(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Trigger string `json:"trigger"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get user ID from context or JWT
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Convert trigger
	trigger := monetization.AdTrigger(request.Trigger)
	
	// Start ad session
	config, err := s.adsIntegration.StartAdSession(r.Context(), userID, trigger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"placement_id": config.PlacementID,
		"type":         config.Type,
		"reward":       config.Reward,
		"reward_type":  config.RewardType,
		"trigger":      config.Trigger,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAvailableRewards returns available rewards for user
func (s *SecureAdsServer) GetAvailableRewards(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	rewards, err := s.adsIntegration.GetAvailableRewards(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"rewards": rewards,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAdStats returns ad statistics for user
func (s *SecureAdsServer) GetAdStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	stats, err := s.adsIntegration.GetAdStats(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetNextAdTime returns when user can watch next ad
func (s *SecureAdsServer) GetNextAdTime(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	triggerStr := chi.URLParam(r, "trigger")
	trigger := monetization.AdTrigger(triggerStr)

	nextTime, err := s.adsIntegration.GetNextAdTime(r.Context(), userID, trigger)
	if err != nil {
		response := map[string]interface{}{
			"success":   false,
			"available": true,
			"message":   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"available":   false,
		"next_time":   nextTime.Unix(),
		"next_time_formatted": nextTime.Format("2006-01-02 15:04:05"),
		"cooldown_minutes": int(time.Until(nextTime).Minutes()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAdsgramWebhook handles webhook from Adsgram
func (s *SecureAdsServer) HandleAdsgramWebhook(w http.ResponseWriter, r *http.Request) {
	s.adsIntegration.HandleAdsgramWebhook(w, r)
}

// HandleWebSocket handles WebSocket connections
func (s *SecureAdsServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	s.adsIntegration.HandleWebSocket(w, r)
}

// getUserIDFromContext extracts user ID from context
func getUserIDFromContext(ctx context.Context) int64 {
	// This should extract user ID from JWT token or session
	// For now, return a mock user ID
	return 12345 // Mock user ID - replace with actual authentication
}

func main() {
	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password
		DB:       0,  // Default DB
	})

	// Test Redis connection
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Create secure ads server
	secureAdsServer := NewSecureAdsServer(redisClient)

	// Setup routes
	router := secureAdsServer.SetupRoutes()

	// Start HTTP server
	port := ":8080"
	log.Printf("Starting Secure Ads Server on port %s", port)
	log.Printf("WebSocket server running on port 8081")
	log.Printf("Available endpoints:")
	log.Printf("  GET  /api/v1/ads/rewards")
	log.Printf("  GET  /api/v1/ads/stats")
	log.Printf("  POST /api/v1/ads/session/start")
	log.Printf("  GET  /api/v1/ads/next-time/{trigger}")
	log.Printf("  POST /webhook/adsgram")
	log.Printf("  GET  /ws")

	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
