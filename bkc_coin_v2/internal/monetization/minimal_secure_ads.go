package monetization

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/gorilla/websocket"
)

// MinimalSecureAds handles ONLY energy restoration and 30 BKC rewards
type MinimalSecureAds struct {
	redisClient    *redis.Client
	adsgramSecret  string
	websocketHub   *WebSocketHub
}

// AdConfig for minimal ads (only 2 types)
type AdConfig struct {
	PlacementID  string `json:"placement_id"`
	Reward       float64 `json:"reward"`
	RewardType   string `json:"reward_type"`
	Cooldown     int    `json:"cooldown_minutes"`
	MaxPerDay    int    `json:"max_per_day"`
	Enabled      bool   `json:"enabled"`
}

// AdsgramWebhook represents webhook payload from Adsgram
type AdsgramWebhook struct {
	Event       string `json:"event"`
	PlacementID string `json:"placement_id"`
	UserID      int64  `json:"user_id"`
	Reward      float64 `json:"reward"`
	Timestamp   int64  `json:"timestamp"`
	Signature   string `json:"signature"`
}

// AdSession represents an active ad watching session
type AdSession struct {
	UserID      int64     `json:"user_id"`
	PlacementID string    `json:"placement_id"`
	StartTime   time.Time `json:"start_time"`
	Status      string    `json:"status"` // "started", "completed", "expired"
}

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	clients    map[int64]*WebSocketClient
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan []byte
}

// WebSocketClient represents a WebSocket client
type WebSocketClient struct {
	userID int64
	conn   *websocket.Conn
	send   chan []byte
}

// NewMinimalSecureAds creates minimal secure ads system
func NewMinimalSecureAds(redisClient *redis.Client, adsgramSecret string) *MinimalSecureAds {
	return &MinimalSecureAds{
		redisClient:   redisClient,
		adsgramSecret: adsgramSecret,
		websocketHub:  NewWebSocketHub(),
	}
}

// GetAdConfigs returns ONLY 2 ad configurations
func (msa *MinimalSecureAds) GetAdConfigs() map[string]AdConfig {
	return map[string]AdConfig{
		"energy_restore": {
			PlacementID:  "energy_restore_1",
			Reward:       0,      // No coins, just energy
			RewardType:   "energy",
			Cooldown:     60,     // 1 hour
			MaxPerDay:    24,     // Max once per hour
			Enabled:      true,
		},
		"coins_reward": {
			PlacementID:  "coins_reward_1",
			Reward:       30,     // 30 BKC coins
			RewardType:   "coins",
			Cooldown:     60,     // 1 hour
			MaxPerDay:    24,     // Max once per hour
			Enabled:      true,
		},
	}
}

// StartAdSession starts a new ad session
func (msa *MinimalSecureAds) StartAdSession(ctx context.Context, userID int64, adType string) (*AdConfig, error) {
	configs := msa.GetAdConfigs()
	config, exists := configs[adType]
	if !exists || !config.Enabled {
		return nil, fmt.Errorf("ad type %s not available", adType)
	}

	// Check 1-hour cooldown
	if msa.isOnCooldown(ctx, userID, config.PlacementID) {
		return nil, fmt.Errorf("ad type %s is on cooldown", adType)
	}

	// Check daily limit (max 24 per day = once per hour)
	if msa.isDailyLimitReached(ctx, userID, config.PlacementID, config.MaxPerDay) {
		return nil, fmt.Errorf("daily limit reached for ad type %s", adType)
	}

	// Create ad session
	session := AdSession{
		UserID:      userID,
		PlacementID:  config.PlacementID,
		StartTime:    time.Now(),
		Status:       "started",
	}

	// Store session in Redis with 5-minute expiration
	sessionKey := fmt.Sprintf("ad_session:%d:%s", userID, config.PlacementID)
	sessionJSON, _ := json.Marshal(session)
	msa.redisClient.Set(ctx, sessionKey, sessionJSON, 5*time.Minute)

	// Set cooldown immediately (prevent duplicate starts)
	cooldownKey := fmt.Sprintf("ad_cooldown:%d:%s", userID, config.PlacementID)
	msa.redisClient.Set(ctx, cooldownKey, "1", time.Hour)

	return &config, nil
}

// isOnCooldown checks if user is on 1-hour cooldown
func (msa *MinimalSecureAds) isOnCooldown(ctx context.Context, userID int64, placementID string) bool {
	key := fmt.Sprintf("ad_cooldown:%d:%s", userID, placementID)
	_, err := msa.redisClient.Get(ctx, key).Result()
	return err == nil
}

// isDailyLimitReached checks if user reached daily limit
func (msa *MinimalSecureAds) isDailyLimitReached(ctx context.Context, userID int64, placementID string, maxPerDay int) bool {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("ad_daily:%s:%d:%s", today, userID, placementID)
	
	count, err := msa.redisClient.Get(ctx, key).Int()
	if err == redis.Nil {
		return false
	}
	
	return count >= maxPerDay
}

// HandleAdsgramWebhook handles webhook from Adsgram
func (msa *MinimalSecureAds) HandleAdsgramWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var webhook AdsgramWebhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	if !msa.verifyWebhookSignature(webhook) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Process the webhook
	ctx := r.Context()
	err := msa.processAdCompletion(ctx, webhook)
	if err != nil {
		log.Printf("Error processing ad completion: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// verifyWebhookSignature verifies the webhook signature from Adsgram
func (msa *MinimalSecureAds) verifyWebhookSignature(webhook AdsgramWebhook) bool {
	// Create signature string
	signatureString := fmt.Sprintf("%s:%s:%d:%d", 
		webhook.Event, 
		webhook.PlacementID, 
		webhook.UserID, 
		webhook.Timestamp,
	)

	// Calculate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(msa.adsgramSecret))
	h.Write([]byte(signatureString))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(webhook.Signature), []byte(expectedSignature))
}

// processAdCompletion processes ad completion from webhook
func (msa *MinimalSecureAds) processAdCompletion(ctx context.Context, webhook AdsgramWebhook) error {
	// Check if session exists and is valid
	sessionKey := fmt.Sprintf("ad_session:%d:%s", webhook.UserID, webhook.PlacementID)
	sessionJSON, err := msa.redisClient.Get(ctx, sessionKey).Result()
	if err == redis.Nil {
		return fmt.Errorf("no active session found for user %d, placement %s", webhook.UserID, webhook.PlacementID)
	}

	var session AdSession
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return fmt.Errorf("invalid session data: %w", err)
	}

	// Check if session is already completed
	if session.Status == "completed" {
		return fmt.Errorf("session already completed")
	}

	// Check if session expired (5 minutes max)
	if time.Since(session.StartTime) > 5*time.Minute {
		msa.redisClient.Del(ctx, sessionKey)
		return fmt.Errorf("session expired")
	}

	// Mark session as completed
	session.Status = "completed"
	updatedSessionJSON, _ := json.Marshal(session)
	msa.redisClient.Set(ctx, sessionKey, updatedSessionJSON, time.Hour)

	// Get ad config
	configs := msa.GetAdConfigs()
	var config *AdConfig
	var adType string
	
	if webhook.PlacementID == configs["energy_restore"].PlacementID {
		config = &configs["energy_restore"]
		adType = "energy"
	} else if webhook.PlacementID == configs["coins_reward"].PlacementID {
		config = &configs["coins_reward"]
		adType = "coins"
	} else {
		return fmt.Errorf("unknown placement ID: %s", webhook.PlacementID)
	}

	// Grant reward based on type
	var reward float64
	var rewardType string
	var rewardData map[string]interface{}

	switch adType {
	case "energy":
		// Restore energy to full
		energyKey := fmt.Sprintf("user_energy:%d", webhook.UserID)
		msa.redisClient.Set(ctx, energyKey, 1000, 0) // Full energy
		reward = 1000
		rewardType = "energy"
		rewardData = map[string]interface{}{
			"energy": 1000,
			"message": "Энергия полностью восстановлена!",
		}
		
	case "coins":
		// Grant 30 BKC coins
		coinsKey := fmt.Sprintf("user_coins:%d", webhook.UserID)
		msa.redisClient.IncrByFloat(ctx, coinsKey, 30)
		reward = 30
		rewardType = "coins"
		rewardData = map[string]interface{}{
			"coins": 30,
			"message": "Получено 30 BKC!",
		}
	}

	// Increment daily count
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("ad_daily:%s:%d:%s", today, webhook.UserID, webhook.PlacementID)
	msa.redisClient.Incr(ctx, dailyKey)
	msa.redisClient.Expire(ctx, dailyKey, 24*time.Hour)

	// Record ad event
	event := map[string]interface{}{
		"user_id":      webhook.UserID,
		"placement_id": webhook.PlacementID,
		"reward":       reward,
		"reward_type":  rewardType,
		"completed":    true,
		"timestamp":    time.Now(),
	}
	
	eventJSON, _ := json.Marshal(event)
	eventKey := fmt.Sprintf("ad_events:%d", webhook.UserID)
	msa.redisClient.LPush(ctx, eventKey, eventJSON)
	msa.redisClient.LTrim(ctx, eventKey, 0, 99)
	msa.redisClient.Expire(ctx, eventKey, 30*24*time.Hour)

	// Send WebSocket notification to client
	msa.sendWebSocketNotification(webhook.UserID, map[string]interface{}{
		"type":       "ad_completed",
		"ad_type":    adType,
		"reward":     reward,
		"rewardType": rewardType,
		"data":       rewardData,
		"timestamp":  time.Now().Unix(),
	})

	return nil
}

// sendWebSocketNotification sends notification to specific user via WebSocket
func (msa *MinimalSecureAds) sendWebSocketNotification(userID int64, data map[string]interface{}) {
	if msa.websocketHub == nil {
		return
	}

	notification, _ := json.Marshal(data)
	
	// Send to specific user
	if client, exists := msa.websocketHub.clients[userID]; exists {
		select {
		case client.send <- notification:
			log.Printf("WebSocket notification sent to user %d", userID)
		default:
			log.Printf("WebSocket client %d not ready for messages", userID)
		}
	}
}

// GetAvailableAds returns available ads for user
func (msa *MinimalSecureAds) GetAvailableAds(ctx context.Context, userID int64) (map[string]interface{}, error) {
	configs := msa.GetAdConfigs()
	available := make(map[string]interface{})

	// Check energy ads
	energyConfig := configs["energy_restore"]
	if energyConfig.Enabled && !msa.isOnCooldown(ctx, userID, energyConfig.PlacementID) {
		// Check if energy is actually depleted
		energyKey := fmt.Sprintf("user_energy:%d", userID)
		energy, err := msa.redisClient.Get(ctx, energyKey).Int64()
		if err == nil && energy <= 100 { // Only show if energy is low
			available["energy"] = map[string]interface{}{
				"type":         "energy",
				"placement_id": energyConfig.PlacementID,
				"reward":       energyConfig.Reward,
				"reward_type":  energyConfig.RewardType,
				"cooldown":     energyConfig.Cooldown,
			}
		}
	}

	// Check coin ads (always available if not on cooldown)
	coinsConfig := configs["coins_reward"]
	if coinsConfig.Enabled && !msa.isOnCooldown(ctx, userID, coinsConfig.PlacementID) {
		if !msa.isDailyLimitReached(ctx, userID, coinsConfig.PlacementID, coinsConfig.MaxPerDay) {
			available["coins"] = map[string]interface{}{
				"type":         "coins",
				"placement_id": coinsConfig.PlacementID,
				"reward":       coinsConfig.Reward,
				"reward_type":  coinsConfig.RewardType,
				"cooldown":     coinsConfig.Cooldown,
			}
		}
	}

	return available, nil
}

// GetAdStats returns advertisement statistics for a user
func (msa *MinimalSecureAds) GetAdStats(ctx context.Context, userID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get today's ad count
	today := time.Now().Format("2006-01-02")
	todayKey := fmt.Sprintf("ad_daily:%s:%d:*", today, userID)
	
	keys, err := msa.redisClient.Keys(ctx, todayKey).Result()
	if err != nil {
		return stats, fmt.Errorf("failed to get daily ad keys: %w", err)
	}

	stats["ads_today"] = len(keys)
	
	// Get total ads watched
	eventsKey := fmt.Sprintf("ad_events:%d", userID)
	events, err := msa.redisClient.LRange(ctx, eventsKey, 0, -1).Result()
	if err != nil {
		return stats, fmt.Errorf("failed to get ad events: %w", err)
	}

	stats["total_ads"] = len(events)
	
	// Calculate total rewards
	var totalRewards float64
	completedAds := 0
	
	for _, eventJSON := range events {
		var event map[string]interface{}
		err := json.Unmarshal([]byte(eventJSON), &event)
		if err != nil {
			continue
		}
		
		if completed, ok := event["completed"].(bool); ok && completed {
			completedAds++
			if reward, ok := event["reward"].(float64); ok {
				totalRewards += reward
			}
		}
	}
	
	stats["completed_ads"] = completedAds
	stats["total_rewards"] = totalRewards

	return stats, nil
}

// GetNextAdTime returns when user can watch next ad
func (msa *MinimalSecureAds) GetNextAdTime(ctx context.Context, userID int64, adType string) (time.Time, error) {
	configs := msa.GetAdConfigs()
	var config *AdConfig
	
	if adType == "energy" {
		config = &configs["energy_restore"]
	} else if adType == "coins" {
		config = &configs["coins_reward"]
	} else {
		return time.Time{}, fmt.Errorf("unknown ad type: %s", adType)
	}

	cooldownKey := fmt.Sprintf("ad_cooldown:%d:%s", userID, config.PlacementID)
	ttl, err := msa.redisClient.TTL(ctx, cooldownKey).Result()
	if err == nil && ttl > 0 {
		return time.Now().Add(ttl), nil
	}

	return time.Time{}, fmt.Errorf("no cooldown for ad type: %s", adType)
}

// CancelExpiredSessions cancels expired ad sessions
func (msa *MinimalSecureAds) CancelExpiredSessions(ctx context.Context) error {
	pattern := "ad_session:*"
	keys, err := msa.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	for _, key := range keys {
		sessionJSON, err := msa.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var session AdSession
		if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
			continue
		}

		// Check if session expired (older than 5 minutes)
		if time.Since(session.StartTime) > 5*time.Minute && session.Status == "started" {
			// Cancel session
			session.Status = "expired"
			updatedSessionJSON, _ := json.Marshal(session)
			msa.redisClient.Set(ctx, key, updatedSessionJSON, time.Hour)

			// Remove cooldown (allow user to try again)
			cooldownKey := fmt.Sprintf("ad_cooldown:%d:%s", session.UserID, session.PlacementID)
			msa.redisClient.Del(ctx, cooldownKey)

			// Notify user via WebSocket
			adType := "coins"
			if session.PlacementID == "energy_restore_1" {
				adType = "energy"
			}

			msa.sendWebSocketNotification(session.UserID, map[string]interface{}{
				"type":      "ad_expired",
				"ad_type":   adType,
				"message":   "Время просмотра рекламы истекло",
				"timestamp": time.Now().Unix(),
			})
		}
	}

	return nil
}

// StartWebSocketServer starts WebSocket server for real-time notifications
func (msa *MinimalSecureAds) StartWebSocketServer(port int) {
	http.HandleFunc("/ws", msa.handleWebSocket)
	
	log.Printf("Starting Minimal WebSocket server on port %d", port)
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		
		for range ticker.C {
			ctx := context.Background()
			msa.CancelExpiredSessions(ctx)
		}
	}()
}

// handleWebSocket handles WebSocket connections
func (msa *MinimalSecureAds) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Get user ID from query parameter
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		conn.Close()
		return
	}

	// Create client
	client := &WebSocketClient{
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, 256),
	}

	// Register client
	msa.websocketHub.register <- client

	// Start goroutines
	go client.writePump()
	go client.readPump()
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[int64]*WebSocketClient),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan []byte),
	}
}

// StartHub starts the WebSocket hub
func (hub *WebSocketHub) StartHub() {
	for {
		select {
		case client := <-hub.register:
			hub.clients[client.userID] = client
			log.Printf("Client %d connected", client.userID)

		case client := <-hub.unregister:
			if _, ok := hub.clients[client.userID]; ok {
				delete(hub.clients, client.userID)
				close(client.send)
				log.Printf("Client %d disconnected", client.userID)
			}

		case message := <-hub.broadcast:
			for _, client := range hub.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(hub.clients, client.userID)
				}
			}
		}
	}
}

// WebSocket client methods
func (c *WebSocketClient) readPump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *WebSocketClient) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		}
	}
}
