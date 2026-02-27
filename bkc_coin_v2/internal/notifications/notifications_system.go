package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationTypeAchievement   NotificationType = "achievement"
	NotificationTypeReward        NotificationType = "reward"
	NotificationTypeEnergyFull    NotificationType = "energy_full"
	NotificationTypeReferral      NotificationType = "referral"
	NotificationTypeNFTPurchase   NotificationType = "nft_purchase"
	NotificationTypePriceAlert    NotificationType = "price_alert"
	NotificationTypeSystemUpdate  NotificationType = "system_update"
	NotificationTypeDailyBonus    NotificationType = "daily_bonus"
	NotificationTypeLevelUp       NotificationType = "level_up"
	NotificationTypeStakingReward NotificationType = "staking_reward"
)

// NotificationPriority represents notification priority
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityMedium NotificationPriority = "medium"
	PriorityHigh   NotificationPriority = "high"
	PriorityUrgent NotificationPriority = "urgent"
)

// Notification represents a push notification
type Notification struct {
	ID        int64                  `json:"id"`
	UserID    int64                  `json:"user_id"`
	Type      NotificationType       `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Priority  NotificationPriority   `json:"priority"`
	IsRead    bool                   `json:"is_read"`
	CreatedAt time.Time              `json:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}

// PushNotificationRequest represents a push notification request
type PushNotificationRequest struct {
	UserID   int64                  `json:"user_id"`
	Type     NotificationType       `json:"type"`
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Priority NotificationPriority   `json:"priority"`
	TTL      int                    `json:"ttl,omitempty"` // Time to live in seconds
}

// NotificationSystem manages push notifications
type NotificationSystem struct {
	redisClient *redis.Client
	// In production, you would integrate with:
	// - Firebase Cloud Messaging (FCM) for Android
	// - Apple Push Notification Service (APNS) for iOS
	// - Web Push API for browsers
	// - Telegram Bot API for in-app notifications
}

// NewNotificationSystem creates a new notification system
func NewNotificationSystem(redisClient *redis.Client) *NotificationSystem {
	return &NotificationSystem{
		redisClient: redisClient,
	}
}

// SendNotification sends a push notification
func (ns *NotificationSystem) SendNotification(ctx context.Context, req *PushNotificationRequest) error {
	notification := &Notification{
		ID:        time.Now().UnixNano(), // Simple ID generation
		UserID:    req.UserID,
		Type:      req.Type,
		Title:     req.Title,
		Message:   req.Message,
		Data:      req.Data,
		Priority:  req.Priority,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	// Store notification in Redis for in-app display
	err := ns.storeNotification(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to store notification: %w", err)
	}

	// Send push notification based on platform
	err = ns.sendPushNotification(ctx, notification, req.TTL)
	if err != nil {
		log.Printf("Failed to send push notification: %v", err)
		// Don't return error here as notification is stored
	}

	// Send Telegram notification if user has linked account
	err = ns.sendTelegramNotification(ctx, notification)
	if err != nil {
		log.Printf("Failed to send Telegram notification: %v", err)
		// Don't return error here as other channels may work
	}

	return nil
}

// storeNotification stores notification in Redis
func (ns *NotificationSystem) storeNotification(ctx context.Context, notification *Notification) error {
	key := fmt.Sprintf("notifications:user:%d", notification.UserID)

	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Store in Redis list (LPUSH)
	err = ns.redisClient.LPush(ctx, key, notificationJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to store notification in Redis: %w", err)
	}

	// Keep only last 100 notifications per user
	err = ns.redisClient.LTrim(ctx, key, 0, 99).Err()
	if err != nil {
		log.Printf("Failed to trim notifications: %v", err)
	}

	// Set expiration for the key (30 days)
	err = ns.redisClient.Expire(ctx, key, 30*24*time.Hour).Err()
	if err != nil {
		log.Printf("Failed to set expiration for notifications: %v", err)
	}

	return nil
}

// sendPushNotification sends push notification to mobile/web
func (ns *NotificationSystem) sendPushNotification(ctx context.Context, notification *Notification, ttl int) error {
	// In production, this would integrate with:
	// - Firebase Cloud Messaging (FCM) for Android
	// - Apple Push Notification Service (APNS) for iOS
	// - Web Push API for browsers

	log.Printf("Sending push notification to user %d: %s - %s",
		notification.UserID, notification.Title, notification.Message)

	// For now, we'll just log the notification
	// In production, you would:
	// 1. Get user's device tokens from database
	// 2. Send to FCM/APNS/Web Push
	// 3. Handle delivery status
	// 4. Update notification status

	return nil
}

// sendTelegramNotification sends notification via Telegram bot
func (ns *NotificationSystem) sendTelegramNotification(ctx context.Context, notification *Notification) error {
	// In production, this would integrate with Telegram Bot API
	// You would need:
	// 1. User's Telegram chat ID
	// 2. Bot token
	// 3. Message formatting

	log.Printf("Sending Telegram notification to user %d: %s - %s",
		notification.UserID, notification.Title, notification.Message)

	// For now, we'll just log
	// In production, you would:
	// 1. Get user's Telegram chat ID from database
	// 2. Send message via Telegram Bot API
	// 3. Handle delivery status

	return nil
}

// GetUserNotifications gets user's notifications
func (ns *NotificationSystem) GetUserNotifications(ctx context.Context, userID int64, limit int) ([]Notification, error) {
	key := fmt.Sprintf("notifications:user:%d", userID)

	// Get notifications from Redis list
	notificationJSONs, err := ns.redisClient.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications from Redis: %w", err)
	}

	var notifications []Notification
	for _, notificationJSON := range notificationJSONs {
		var notification Notification
		err := json.Unmarshal([]byte(notificationJSON), &notification)
		if err != nil {
			log.Printf("Failed to unmarshal notification: %v", err)
			continue
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkAsRead marks notification as read
func (ns *NotificationSystem) MarkAsRead(ctx context.Context, userID int64, notificationID int64) error {
	key := fmt.Sprintf("notifications:user:%d", userID)

	// Get all notifications
	notificationJSONs, err := ns.redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get notifications: %w", err)
	}

	// Find and update the notification
	for i, notificationJSON := range notificationJSONs {
		var notification Notification
		err := json.Unmarshal([]byte(notificationJSON), &notification)
		if err != nil {
			continue
		}

		if notification.ID == notificationID && !notification.IsRead {
			notification.IsRead = true
			now := time.Now()
			notification.ReadAt = &now

			updatedJSON, err := json.Marshal(notification)
			if err != nil {
				continue
			}

			// Update the notification in the list
			err = ns.redisClient.LSet(ctx, key, int64(i), updatedJSON).Err()
			if err != nil {
				log.Printf("Failed to update notification: %v", err)
			}
		}
	}

	return nil
}

// MarkAllAsRead marks all user notifications as read
func (ns *NotificationSystem) MarkAllAsRead(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("notifications:user:%d", userID)

	// Get all notifications
	notificationJSONs, err := ns.redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get notifications: %w", err)
	}

	// Update all notifications
	for _, notificationJSON := range notificationJSONs {
		var notification Notification
		err := json.Unmarshal([]byte(notificationJSON), &notification)
		if err != nil {
			continue
		}

		if !notification.IsRead {
			notification.IsRead = true
			now := time.Now()
			notification.ReadAt = &now

			updatedJSON, err := json.Marshal(notification)
			if err != nil {
				continue
			}

			// Find the index of this notification and update it
			for i, notifJSON := range notificationJSONs {
				var notif Notification
				err := json.Unmarshal([]byte(notifJSON), &notif)
				if err != nil {
					continue
				}
				if notif.ID == notification.ID {
					// Update the notification in the list
					err = ns.redisClient.LSet(ctx, key, int64(i), updatedJSON).Err()
					if err != nil {
						log.Printf("Failed to update notification: %v", err)
					}
					break
				}
			}
		}
	}

	return nil
}

// GetUnreadCount gets unread notifications count
func (ns *NotificationSystem) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	key := fmt.Sprintf("notifications:user:%d", userID)

	// Get all notifications
	notificationJSONs, err := ns.redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get notifications: %w", err)
	}

	unreadCount := 0
	for _, notificationJSON := range notificationJSONs {
		var notification Notification
		err := json.Unmarshal([]byte(notificationJSON), &notification)
		if err != nil {
			continue
		}

		if !notification.IsRead {
			unreadCount++
		}
	}

	return unreadCount, nil
}

// DeleteNotification deletes a notification
func (ns *NotificationSystem) DeleteNotification(ctx context.Context, userID int64, notificationID int64) error {
	key := fmt.Sprintf("notifications:user:%d", userID)

	// Get all notifications
	notificationJSONs, err := ns.redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get notifications: %w", err)
	}

	// Find and remove the notification
	for i, notificationJSON := range notificationJSONs {
		var notification Notification
		err := json.Unmarshal([]byte(notificationJSON), &notification)
		if err != nil {
			continue
		}

		if notification.ID == notificationID {
			// Remove from list
			err = ns.redisClient.LRem(ctx, key, 1, notificationJSON).Err()
			if err != nil {
				return fmt.Errorf("failed to remove notification: %w", err)
			}
			break
		}
	}

	return nil
}

// ClearAllNotifications clears all user notifications
func (ns *NotificationSystem) ClearAllNotifications(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("notifications:user:%d", userID)

	err := ns.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to clear notifications: %w", err)
	}

	return nil
}

// SendAchievementNotification sends achievement completion notification
func (ns *NotificationSystem) SendAchievementNotification(ctx context.Context, userID int64, achievementName string, reward float64) error {
	req := &PushNotificationRequest{
		UserID:   userID,
		Type:     NotificationTypeAchievement,
		Title:    "🏆 Достижение получено!",
		Message:  fmt.Sprintf("Поздравляем! Вы получили достижение \"%s\" и награду %.2f BKC", achievementName, reward),
		Priority: PriorityHigh,
		Data: map[string]interface{}{
			"achievement_name": achievementName,
			"reward":           reward,
		},
	}

	return ns.SendNotification(ctx, req)
}

// SendEnergyFullNotification sends energy full notification
func (ns *NotificationSystem) SendEnergyFullNotification(ctx context.Context, userID int64) error {
	req := &PushNotificationRequest{
		UserID:   userID,
		Type:     NotificationTypeEnergyFull,
		Title:    "⚡ Энергия восстановлена!",
		Message:  "Ваша энергия полностью восстановлена. Продолжайте играть!",
		Priority: PriorityMedium,
		Data: map[string]interface{}{
			"energy_full": true,
		},
	}

	return ns.SendNotification(ctx, req)
}

// SendReferralNotification sends referral notification
func (ns *NotificationSystem) SendReferralNotification(ctx context.Context, userID int64, referralCount int, reward float64) error {
	req := &PushNotificationRequest{
		UserID:   userID,
		Type:     NotificationTypeReferral,
		Title:    "👥 Новый реферал!",
		Message:  fmt.Sprintf("У вас %d активных рефералов! Получено %.2f BKC", referralCount, reward),
		Priority: PriorityHigh,
		Data: map[string]interface{}{
			"referral_count": referralCount,
			"reward":         reward,
		},
	}

	return ns.SendNotification(ctx, req)
}

// SendDailyBonusNotification sends daily bonus notification
func (ns *NotificationSystem) SendDailyBonusNotification(ctx context.Context, userID int64, bonus float64, streak int) error {
	req := &PushNotificationRequest{
		UserID:   userID,
		Type:     NotificationTypeDailyBonus,
		Title:    "🎁 Ежедневный бонус!",
		Message:  fmt.Sprintf("Ваш ежедневный бонус: %.2f BKC! Серия дней: %d", bonus, streak),
		Priority: PriorityMedium,
		Data: map[string]interface{}{
			"bonus":  bonus,
			"streak": streak,
		},
	}

	return ns.SendNotification(ctx, req)
}

// SendPriceAlertNotification sends price alert notification
func (ns *NotificationSystem) SendPriceAlertNotification(ctx context.Context, userID int64, currentPrice, targetPrice float64, isIncrease bool) error {
	direction := "выросла"
	if !isIncrease {
		direction = "упала"
	}

	req := &PushNotificationRequest{
		UserID:   userID,
		Type:     NotificationTypePriceAlert,
		Title:    "📊 Оповещение о цене!",
		Message:  fmt.Sprintf("Цена BKC %s до %.6f USD (цель: %.6f)", direction, currentPrice, targetPrice),
		Priority: PriorityMedium,
		Data: map[string]interface{}{
			"current_price": currentPrice,
			"target_price":  targetPrice,
			"is_increase":   isIncrease,
		},
	}

	return ns.SendNotification(ctx, req)
}

// GetNotificationStats gets notification statistics
func (ns *NotificationSystem) GetNotificationStats(ctx context.Context, userID int64) (*NotificationStats, error) {
	key := fmt.Sprintf("notifications:user:%d", userID)

	// Get all notifications
	notificationJSONs, err := ns.redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	stats := &NotificationStats{
		Total:      len(notificationJSONs),
		Unread:     0,
		ByType:     make(map[NotificationType]int),
		ByPriority: make(map[NotificationPriority]int),
	}

	for _, notificationJSON := range notificationJSONs {
		var notification Notification
		err := json.Unmarshal([]byte(notificationJSON), &notification)
		if err != nil {
			continue
		}

		if !notification.IsRead {
			stats.Unread++
		}

		stats.ByType[notification.Type]++
		stats.ByPriority[notification.Priority]++
	}

	return stats, nil
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	Total      int                          `json:"total"`
	Unread     int                          `json:"unread"`
	ByType     map[NotificationType]int     `json:"by_type"`
	ByPriority map[NotificationPriority]int `json:"by_priority"`
}
