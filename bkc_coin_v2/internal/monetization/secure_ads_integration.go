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
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// SecureAdsIntegration handles secure advertisement integration with webhook verification
type SecureAdsIntegration struct {
	redisClient   *redis.Client
	adsgramSecret string // Secret key for webhook verification
	rewardURL     string // Webhook URL for Adsgram
	websocketHub  *WebSocketHub
	configs       map[AdTrigger][]AdConfig
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

// AdsgramWebhook represents webhook payload from Adsgram
type AdsgramWebhook struct {
	Event       string  `json:"event"`
	PlacementID string  `json:"placement_id"`
	UserID      int64   `json:"user_id"`
	Reward      float64 `json:"reward"`
	Timestamp   int64   `json:"timestamp"`
	Signature   string  `json:"signature"`
}

// AdSession represents an active ad watching session
type AdSession struct {
	UserID      int64     `json:"user_id"`
	PlacementID string    `json:"placement_id"`
	Trigger     AdTrigger `json:"trigger"`
	StartTime   time.Time `json:"start_time"`
	Status      string    `json:"status"` // "started", "completed", "expired"
}

// NewSecureAdsIntegration creates a new secure ads integration system
func NewSecureAdsIntegration(redisClient *redis.Client, adsgramSecret, rewardURL string) *SecureAdsIntegration {
	return &SecureAdsIntegration{
		redisClient:   redisClient,
		adsgramSecret: adsgramSecret,
		rewardURL:     rewardURL,
		websocketHub:  NewWebSocketHub(),
		configs:       make(map[AdTrigger][]AdConfig),
	}
}

// InitializeSecureConfigs initializes secure ad configurations
func (sai *SecureAdsIntegration) InitializeSecureConfigs() {
	sai.configs = map[AdTrigger][]AdConfig{
		TriggerEnergyDepleted: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerEnergyDepleted,
				PlacementID: "energy_restore_1",
				Reward:      0, // Energy restoration, no coins
				RewardType:  "energy_restore",
				Cooldown:    60, // 1 hour cooldown
				MaxPerDay:   24, // Max once per hour theoretically
				Enabled:     true,
			},
		},
		TriggerDailyBonus: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerDailyBonus,
				PlacementID: "daily_bonus_1",
				Reward:      30, // 30 BKC coins
				RewardType:  "coins",
				Cooldown:    60, // 1 hour cooldown
				MaxPerDay:   24, // Max once per hour
				Enabled:     true,
			},
		},
		TriggerAchievement: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerAchievement,
				PlacementID: "achievement_1",
				Reward:      30, // 30 BKC coins
				RewardType:  "coins",
				Cooldown:    60, // 1 hour cooldown
				MaxPerDay:   24, // Max once per hour
				Enabled:     true,
			},
		},
		TriggerLevelUp: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeInterstitial,
				Trigger:     TriggerLevelUp,
				PlacementID: "level_up_1",
				Reward:      0, // No reward
				RewardType:  "",
				Cooldown:    60, // 1 hour cooldown
				MaxPerDay:   24, // Max once per hour
				Enabled:     true,
			},
		},
		TriggerTournamentEntry: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerTournamentEntry,
				PlacementID: "tournament_1",
				Reward:      30, // 30 BKC coins
				RewardType:  "coins",
				Cooldown:    60, // 1 hour cooldown
				MaxPerDay:   24, // Max once per hour
				Enabled:     true,
			},
		},
		TriggerNFTPurchase: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerNFTPurchase,
				PlacementID: "nft_discount_1",
				Reward:      30, // 30 BKC coins
				RewardType:  "coins",
				Cooldown:    60, // 1 hour cooldown
				MaxPerDay:   24, // Max once per hour
				Enabled:     true,
			},
		},
	}
}

// StartAdSession starts a new ad watching session
func (sai *SecureAdsIntegration) StartAdSession(ctx context.Context, userID int64, trigger AdTrigger) (*AdConfig, error) {
	configs, exists := sai.configs[trigger]
	if !exists {
		return nil, fmt.Errorf("no configuration for trigger: %s", trigger)
	}

	for _, config := range configs {
		if !config.Enabled {
			continue
		}

		// Check 1-hour cooldown
		if sai.isOnCooldown(ctx, userID, config) {
			continue
		}

		// Check daily limit (max 24 per day = once per hour)
		if sai.isDailyLimitReached(ctx, userID, config) {
			continue
		}

		// Create ad session
		session := AdSession{
			UserID:      userID,
			PlacementID: config.PlacementID,
			Trigger:     config.Trigger,
			StartTime:   time.Now(),
			Status:      "started",
		}

		// Store session in Redis with 5-minute expiration
		sessionKey := fmt.Sprintf("ad_session:%d:%s", userID, config.PlacementID)
		sessionJSON, _ := json.Marshal(session)
		sai.redisClient.Set(ctx, sessionKey, sessionJSON, 5*time.Minute)

		// Set cooldown immediately (prevent duplicate starts)
		cooldownKey := fmt.Sprintf("ad_cooldown:%d:%s", userID, config.PlacementID)
		sai.redisClient.Set(ctx, cooldownKey, "1", time.Hour)

		return &config, nil
	}

	return nil, fmt.Errorf("no available ads for trigger: %s", trigger)
}

// isOnCooldown checks if user is on 1-hour cooldown
func (sai *SecureAdsIntegration) isOnCooldown(ctx context.Context, userID int64, config AdConfig) bool {
	key := fmt.Sprintf("ad_cooldown:%d:%s", userID, config.PlacementID)
	_, err := sai.redisClient.Get(ctx, key).Result()
	return err == nil
}

// isDailyLimitReached checks if user reached daily limit (24 per day)
func (sai *SecureAdsIntegration) isDailyLimitReached(ctx context.Context, userID int64, config AdConfig) bool {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("ad_daily:%s:%d:%s", today, userID, config.PlacementID)

	count, err := sai.redisClient.Get(ctx, key).Int()
	if err == redis.Nil {
		return false
	}

	return count >= config.MaxPerDay
}

// HandleAdsgramWebhook handles webhook from Adsgram
func (sai *SecureAdsIntegration) HandleAdsgramWebhook(w http.ResponseWriter, r *http.Request) {
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
	if !sai.verifyWebhookSignature(webhook) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Process the webhook
	ctx := r.Context()
	err := sai.processAdCompletion(ctx, webhook)
	if err != nil {
		log.Printf("Error processing ad completion: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// verifyWebhookSignature verifies the webhook signature from Adsgram
func (sai *SecureAdsIntegration) verifyWebhookSignature(webhook AdsgramWebhook) bool {
	// Create signature string
	signatureString := fmt.Sprintf("%s:%s:%d:%d",
		webhook.Event,
		webhook.PlacementID,
		webhook.UserID,
		webhook.Timestamp,
	)

	// Calculate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(sai.adsgramSecret))
	h.Write([]byte(signatureString))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(webhook.Signature), []byte(expectedSignature))
}

// processAdCompletion processes ad completion from webhook
func (sai *SecureAdsIntegration) processAdCompletion(ctx context.Context, webhook AdsgramWebhook) error {
	// Check if session exists and is valid
	sessionKey := fmt.Sprintf("ad_session:%d:%s", webhook.UserID, webhook.PlacementID)
	sessionJSON, err := sai.redisClient.Get(ctx, sessionKey).Result()
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
		sai.redisClient.Del(ctx, sessionKey)
		return fmt.Errorf("session expired")
	}

	// Mark session as completed
	session.Status = "completed"
	updatedSessionJSON, _ := json.Marshal(session)
	sai.redisClient.Set(ctx, sessionKey, updatedSessionJSON, time.Hour)

	// Get ad config
	var config *AdConfig
	for _, configs := range sai.configs {
		for _, cfg := range configs {
			if cfg.PlacementID == webhook.PlacementID {
				config = &cfg
				break
			}
		}
		if config != nil {
			break
		}
	}

	if config == nil {
		return fmt.Errorf("no configuration found for placement %s", webhook.PlacementID)
	}

	// Grant reward based on type
	var reward float64
	var rewardType string
	var rewardData map[string]interface{}

	switch config.RewardType {
	case "energy_restore":
		// Restore energy to full
		energyKey := fmt.Sprintf("user_energy:%d", webhook.UserID)
		sai.redisClient.Set(ctx, energyKey, 1000, 0) // Full energy
		reward = 1000
		rewardType = "energy"
		rewardData = map[string]interface{}{
			"energy":  1000,
			"message": "Энергия полностью восстановлена!",
		}

	case "coins":
		// Grant 30 BKC coins
		coinsKey := fmt.Sprintf("user_coins:%d", webhook.UserID)
		sai.redisClient.IncrByFloat(ctx, coinsKey, 30)
		reward = 30
		rewardType = "coins"
		rewardData = map[string]interface{}{
			"coins":   30,
			"message": "Получено 30 BKC!",
		}

	default:
		// No reward (interstitial)
		reward = 0
		rewardType = "none"
		rewardData = map[string]interface{}{
			"message": "Спасибо за просмотр!",
		}
	}

	// Increment daily count
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("ad_daily:%s:%d:%s", today, webhook.UserID, webhook.PlacementID)
	sai.redisClient.Incr(ctx, dailyKey)
	sai.redisClient.Expire(ctx, dailyKey, 24*time.Hour)

	// Record ad event
	event := AdEvent{
		UserID:      webhook.UserID,
		AdType:      config.Type,
		Trigger:     config.Trigger,
		Reward:      reward,
		Completed:   true,
		Timestamp:   time.Now(),
		PlacementID: webhook.PlacementID,
	}
	sai.recordAdEvent(ctx, event)

	// Send WebSocket notification to client
	sai.sendWebSocketNotification(webhook.UserID, map[string]interface{}{
		"type":       "ad_completed",
		"trigger":    config.Trigger,
		"reward":     reward,
		"rewardType": rewardType,
		"data":       rewardData,
		"timestamp":  time.Now().Unix(),
	})

	return nil
}

// sendWebSocketNotification sends notification to specific user via WebSocket
func (sai *SecureAdsIntegration) sendWebSocketNotification(userID int64, data map[string]interface{}) {
	if sai.websocketHub == nil {
		return
	}

	notification, _ := json.Marshal(data)

	// Send to specific user
	if client, exists := sai.websocketHub.clients[userID]; exists {
		select {
		case client.send <- notification:
			log.Printf("WebSocket notification sent to user %d", userID)
		default:
			log.Printf("WebSocket client %d not ready for messages", userID)
		}
	}
}

// recordAdEvent records an ad event
func (sai *SecureAdsIntegration) recordAdEvent(ctx context.Context, event AdEvent) error {
	key := fmt.Sprintf("ad_events:%d", event.UserID)

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal ad event: %w", err)
	}

	// Store in list (keep last 100 events)
	err = sai.redisClient.LPush(ctx, key, eventJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to store ad event: %w", err)
	}

	sai.redisClient.LTrim(ctx, key, 0, 99)
	sai.redisClient.Expire(ctx, key, 30*24*time.Hour) // 30 days

	return nil
}

// GetAdStats returns advertisement statistics for a user
func (sai *SecureAdsIntegration) GetAdStats(ctx context.Context, userID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get today's ad count
	today := time.Now().Format("2006-01-02")
	todayKey := fmt.Sprintf("ad_daily:%s:%d:*", today, userID)

	keys, err := sai.redisClient.Keys(ctx, todayKey).Result()
	if err != nil {
		return stats, fmt.Errorf("failed to get daily ad keys: %w", err)
	}

	stats["ads_today"] = len(keys)

	// Get total ads watched
	eventsKey := fmt.Sprintf("ad_events:%d", userID)
	events, err := sai.redisClient.LRange(ctx, eventsKey, 0, -1).Result()
	if err != nil {
		return stats, fmt.Errorf("failed to get ad events: %w", err)
	}

	stats["total_ads"] = len(events)

	// Calculate total rewards
	var totalRewards float64
	completedAds := 0

	for _, eventJSON := range events {
		var event AdEvent
		err := json.Unmarshal([]byte(eventJSON), &event)
		if err != nil {
			continue
		}

		if event.Completed {
			completedAds++
			totalRewards += event.Reward
		}
	}

	stats["completed_ads"] = completedAds
	stats["total_rewards"] = totalRewards

	return stats, nil
}

// GetAvailableRewards returns available rewards for user
func (sai *SecureAdsIntegration) GetAvailableRewards(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	var rewards []map[string]interface{}

	for trigger, configs := range sai.configs {
		for _, config := range configs {
			if !config.Enabled {
				continue
			}

			if sai.isOnCooldown(ctx, userID, config) {
				continue
			}

			if sai.isDailyLimitReached(ctx, userID, config) {
				continue
			}

			// Check trigger-specific conditions
			if trigger == TriggerEnergyDepleted {
				energyKey := fmt.Sprintf("user_energy:%d", userID)
				energy, err := sai.redisClient.Get(ctx, energyKey).Int64()
				if err == nil && energy > 100 {
					continue
				}
			}

			rewards = append(rewards, map[string]interface{}{
				"trigger":      config.Trigger,
				"type":         config.Type,
				"reward":       config.Reward,
				"reward_type":  config.RewardType,
				"placement_id": config.PlacementID,
				"provider":     config.Provider,
				"cooldown":     config.Cooldown,
			})
		}
	}

	return rewards, nil
}

// GetNextAdTime returns when user can watch next ad
func (sai *SecureAdsIntegration) GetNextAdTime(ctx context.Context, userID int64, trigger AdTrigger) (time.Time, error) {
	configs, exists := sai.configs[trigger]
	if !exists {
		return time.Time{}, fmt.Errorf("no configuration for trigger: %s", trigger)
	}

	var nextTime time.Time
	var hasCooldown bool

	for _, config := range configs {
		if !config.Enabled {
			continue
		}

		cooldownKey := fmt.Sprintf("ad_cooldown:%d:%s", userID, config.PlacementID)
		ttl, err := sai.redisClient.TTL(ctx, cooldownKey).Result()
		if err == nil && ttl > 0 {
			cooldownEnd := time.Now().Add(ttl)
			if !hasCooldown || cooldownEnd.After(nextTime) {
				nextTime = cooldownEnd
				hasCooldown = true
			}
		}
	}

	if hasCooldown {
		return nextTime, nil
	}

	return time.Time{}, fmt.Errorf("no cooldown for trigger: %s", trigger)
}

// CancelExpiredSessions cancels expired ad sessions
func (sai *SecureAdsIntegration) CancelExpiredSessions(ctx context.Context) error {
	pattern := "ad_session:*"
	keys, err := sai.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	for _, key := range keys {
		sessionJSON, err := sai.redisClient.Get(ctx, key).Result()
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
			sai.redisClient.Set(ctx, key, updatedSessionJSON, time.Hour)

			// Remove cooldown (allow user to try again)
			cooldownKey := fmt.Sprintf("ad_cooldown:%d:%s", session.UserID, session.PlacementID)
			sai.redisClient.Del(ctx, cooldownKey)

			// Notify user via WebSocket
			sai.sendWebSocketNotification(session.UserID, map[string]interface{}{
				"type":      "ad_expired",
				"trigger":   session.Trigger,
				"message":   "Время просмотра рекламы истекло",
				"timestamp": time.Now().Unix(),
			})
		}
	}

	return nil
}

// StartWebSocketServer starts WebSocket server for real-time notifications
func (sai *SecureAdsIntegration) StartWebSocketServer(port int) {
	http.HandleFunc("/ws", sai.handleWebSocket)

	log.Printf("Starting WebSocket server on port %d", port)
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			sai.CancelExpiredSessions(ctx)
		}
	}()
}

// handleWebSocket handles WebSocket connections
func (sai *SecureAdsIntegration) handleWebSocket(w http.ResponseWriter, r *http.Request) {
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
	sai.websocketHub.register <- client

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
