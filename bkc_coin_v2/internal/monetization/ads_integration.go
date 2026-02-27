package monetization

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// AdsProvider represents different ad providers
type AdsProvider string

const (
	AdsProviderAdsgram  AdsProvider = "adsgram"
	AdsProviderUnity    AdsProvider = "unity"
	AdsProviderAppLovin AdsProvider = "applovin"
)

// AdType represents different types of ads
type AdType string

const (
	AdTypeRewarded     AdType = "rewarded"     // Rewarded video ads
	AdTypeInterstitial AdType = "interstitial" // Full-screen ads
	AdTypeBanner       AdType = "banner"       // Banner ads
	AdTypeNative       AdType = "native"       // Native ads
)

// AdTrigger represents different trigger conditions
type AdTrigger string

const (
	TriggerEnergyDepleted  AdTrigger = "energy_depleted"
	TriggerDailyBonus      AdTrigger = "daily_bonus"
	TriggerAchievement     AdTrigger = "achievement"
	TriggerLevelUp         AdTrigger = "level_up"
	TriggerTournamentEntry AdTrigger = "tournament_entry"
	TriggerNFTPurchase     AdTrigger = "nft_purchase"
)

// AdConfig represents advertisement configuration
type AdConfig struct {
	Provider    AdsProvider `json:"provider"`
	Type        AdType      `json:"type"`
	Trigger     AdTrigger   `json:"trigger"`
	PlacementID string      `json:"placement_id"`
	Reward      float64     `json:"reward"`
	RewardType  string      `json:"reward_type"`
	Cooldown    int         `json:"cooldown_minutes"`
	MaxPerDay   int         `json:"max_per_day"`
	Enabled     bool        `json:"enabled"`
}

// AdEvent represents an ad event
type AdEvent struct {
	UserID      int64     `json:"user_id"`
	AdType      AdType    `json:"ad_type"`
	Trigger     AdTrigger `json:"trigger"`
	Reward      float64   `json:"reward"`
	Completed   bool      `json:"completed"`
	Timestamp   time.Time `json:"timestamp"`
	PlacementID string    `json:"placement_id"`
}

// AdsIntegration handles advertisement integration
type AdsIntegration struct {
	redisClient *redis.Client
	configs     map[AdTrigger][]AdConfig
}

// NewAdsIntegration creates a new ads integration system
func NewAdsIntegration(redisClient *redis.Client) *AdsIntegration {
	return &AdsIntegration{
		redisClient: redisClient,
		configs:     make(map[AdTrigger][]AdConfig),
	}
}

// InitializeConfigs initializes default ad configurations
func (ai *AdsIntegration) InitializeConfigs() {
	ai.configs = map[AdTrigger][]AdConfig{
		TriggerEnergyDepleted: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerEnergyDepleted,
				PlacementID: "energy_reward_1",
				Reward:      500, // 500 energy points
				RewardType:  "energy",
				Cooldown:    5,  // 5 minutes cooldown
				MaxPerDay:   10, // Max 10 per day
				Enabled:     true,
			},
		},
		TriggerDailyBonus: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerDailyBonus,
				PlacementID: "daily_bonus_1",
				Reward:      50, // 50 BKC coins
				RewardType:  "coins",
				Cooldown:    60, // 1 hour cooldown
				MaxPerDay:   3,  // Max 3 per day
				Enabled:     true,
			},
		},
		TriggerAchievement: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerAchievement,
				PlacementID: "achievement_1",
				Reward:      25, // 25 BKC coins
				RewardType:  "coins",
				Cooldown:    30, // 30 minutes cooldown
				MaxPerDay:   5,  // Max 5 per day
				Enabled:     true,
			},
		},
		TriggerLevelUp: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeInterstitial,
				Trigger:     TriggerLevelUp,
				PlacementID: "level_up_1",
				Reward:      0, // No reward, just show ad
				RewardType:  "",
				Cooldown:    0, // No cooldown
				MaxPerDay:   1, // Max 1 per level up
				Enabled:     true,
			},
		},
		TriggerTournamentEntry: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerTournamentEntry,
				PlacementID: "tournament_1",
				Reward:      100, // Free tournament entry
				RewardType:  "tournament_entry",
				Cooldown:    120, // 2 hours cooldown
				MaxPerDay:   2,   // Max 2 per day
				Enabled:     true,
			},
		},
		TriggerNFTPurchase: {
			{
				Provider:    AdsProviderAdsgram,
				Type:        AdTypeRewarded,
				Trigger:     TriggerNFTPurchase,
				PlacementID: "nft_discount_1",
				Reward:      0.1, // 10% discount
				RewardType:  "discount",
				Cooldown:    180, // 3 hours cooldown
				MaxPerDay:   3,   // Max 3 per day
				Enabled:     true,
			},
		},
	}
}

// CheckAdTrigger checks if an ad should be triggered
func (ai *AdsIntegration) CheckAdTrigger(ctx context.Context, userID int64, trigger AdTrigger) (*AdConfig, error) {
	configs, exists := ai.configs[trigger]
	if !exists {
		return nil, fmt.Errorf("no configuration for trigger: %s", trigger)
	}

	for _, config := range configs {
		if !config.Enabled {
			continue
		}

		// Check cooldown
		if ai.isOnCooldown(ctx, userID, config) {
			continue
		}

		// Check daily limit
		if ai.isDailyLimitReached(ctx, userID, config) {
			continue
		}

		// Return the first available config
		return &config, nil
	}

	return nil, fmt.Errorf("no available ads for trigger: %s", trigger)
}

// isOnCooldown checks if user is on cooldown for this ad
func (ai *AdsIntegration) isOnCooldown(ctx context.Context, userID int64, config AdConfig) bool {
	if config.Cooldown == 0 {
		return false
	}

	key := fmt.Sprintf("ad_cooldown:%d:%s", userID, config.PlacementID)
	_, err := ai.redisClient.Get(ctx, key).Result()
	return err == nil
}

// isDailyLimitReached checks if user reached daily limit
func (ai *AdsIntegration) isDailyLimitReached(ctx context.Context, userID int64, config AdConfig) bool {
	if config.MaxPerDay == 0 {
		return false
	}

	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("ad_daily:%s:%d:%s", today, userID, config.PlacementID)

	count, err := ai.redisClient.Get(ctx, key).Int()
	if err == redis.Nil {
		return false
	}

	return count >= config.MaxPerDay
}

// RecordAdStart records when user starts watching an ad
func (ai *AdsIntegration) RecordAdStart(ctx context.Context, userID int64, config AdConfig) error {
	// Set cooldown
	if config.Cooldown > 0 {
		key := fmt.Sprintf("ad_cooldown:%d:%s", userID, config.PlacementID)
		ai.redisClient.Set(ctx, key, "1", time.Duration(config.Cooldown)*time.Minute)
	}

	// Increment daily count
	if config.MaxPerDay > 0 {
		today := time.Now().Format("2006-01-02")
		key := fmt.Sprintf("ad_daily:%s:%d:%s", today, userID, config.PlacementID)
		ai.redisClient.Incr(ctx, key)
		ai.redisClient.Expire(ctx, key, 24*time.Hour)
	}

	// Record ad start event
	event := AdEvent{
		UserID:      userID,
		AdType:      config.Type,
		Trigger:     config.Trigger,
		Reward:      config.Reward,
		Completed:   false,
		Timestamp:   time.Now(),
		PlacementID: config.PlacementID,
	}

	return ai.recordAdEvent(ctx, event)
}

// RecordAdCompletion records when user completes watching an ad
func (ai *AdsIntegration) RecordAdCompletion(ctx context.Context, userID int64, config AdConfig) error {
	// Record completion event
	event := AdEvent{
		UserID:      userID,
		AdType:      config.Type,
		Trigger:     config.Trigger,
		Reward:      config.Reward,
		Completed:   true,
		Timestamp:   time.Now(),
		PlacementID: config.PlacementID,
	}

	err := ai.recordAdEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to record ad completion: %w", err)
	}

	// Grant reward
	return ai.grantReward(ctx, userID, config)
}

// recordAdEvent records an ad event
func (ai *AdsIntegration) recordAdEvent(ctx context.Context, event AdEvent) error {
	key := fmt.Sprintf("ad_events:%d", event.UserID)

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal ad event: %w", err)
	}

	// Store in list (keep last 100 events)
	err = ai.redisClient.LPush(ctx, key, eventJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to store ad event: %w", err)
	}

	ai.redisClient.LTrim(ctx, key, 0, 99)
	ai.redisClient.Expire(ctx, key, 30*24*time.Hour) // 30 days

	return nil
}

// grantReward grants the reward for completing an ad
func (ai *AdsIntegration) grantReward(ctx context.Context, userID int64, config AdConfig) error {
	switch config.RewardType {
	case "energy":
		// Grant energy reward
		energyKey := fmt.Sprintf("user_energy:%d", userID)
		ai.redisClient.IncrBy(ctx, energyKey, int64(config.Reward))

	case "coins":
		// Grant coin reward
		coinsKey := fmt.Sprintf("user_coins:%d", userID)
		ai.redisClient.IncrByFloat(ctx, coinsKey, config.Reward)

	case "tournament_entry":
		// Grant free tournament entry
		entryKey := fmt.Sprintf("free_entries:%d", userID)
		ai.redisClient.Incr(ctx, entryKey)

	case "discount":
		// Grant discount
		discountKey := fmt.Sprintf("nft_discount:%d", userID)
		ai.redisClient.Set(ctx, discountKey, config.Reward, time.Hour)
	}

	return nil
}

// GetAdStats returns advertisement statistics for a user
func (ai *AdsIntegration) GetAdStats(ctx context.Context, userID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get today's ad count
	today := time.Now().Format("2006-01-02")
	todayKey := fmt.Sprintf("ad_daily:%s:%d:*", today, userID)

	keys, err := ai.redisClient.Keys(ctx, todayKey).Result()
	if err != nil {
		return stats, fmt.Errorf("failed to get daily ad keys: %w", err)
	}

	stats["ads_today"] = len(keys)

	// Get total ads watched
	eventsKey := fmt.Sprintf("ad_events:%d", userID)
	events, err := ai.redisClient.LRange(ctx, eventsKey, 0, -1).Result()
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

// GetAdConfig returns ad configuration for frontend
func (ai *AdsIntegration) GetAdConfig(ctx context.Context, userID int64, trigger AdTrigger) (map[string]interface{}, error) {
	config, err := ai.CheckAdTrigger(ctx, userID, trigger)
	if err != nil {
		return nil, err
	}

	// Check if user is eligible for energy restoration
	if trigger == TriggerEnergyDepleted {
		energyKey := fmt.Sprintf("user_energy:%d", userID)
		energy, err := ai.redisClient.Get(ctx, energyKey).Int64()
		if err == nil && energy > 100 { // Only show if energy is low
			return nil, fmt.Errorf("energy not depleted enough")
		}
	}

	return map[string]interface{}{
		"provider":     config.Provider,
		"type":         config.Type,
		"placement_id": config.PlacementID,
		"reward":       config.Reward,
		"reward_type":  config.RewardType,
		"trigger":      config.Trigger,
	}, nil
}

// GenerateAdsgramScript generates Adsgram integration script
func (ai *AdsIntegration) GenerateAdsgramScript(placementID string) string {
	return fmt.Sprintf(`
// Adsgram integration
window.adsgram = {
    showRewardedAd: function(placementId, onReward, onError) {
        // Check if Adsgram SDK is loaded
        if (typeof window.Adsgram === 'undefined') {
            console.error('Adsgram SDK not loaded');
            onError('SDK not loaded');
            return;
        }
        
        // Show rewarded ad
        window.Adsgram.showRewardedAd({
            placementId: placementId,
            onReward: function(reward) {
                console.log('Ad completed, reward:', reward);
                onReward(reward);
            },
            onError: function(error) {
                console.error('Ad error:', error);
                onError(error);
            },
            onAdClosed: function() {
                console.log('Ad closed by user');
            }
        });
    },
    
    showInterstitial: function(placementId, onClosed, onError) {
        if (typeof window.Adsgram === 'undefined') {
            console.error('Adsgram SDK not loaded');
            onError('SDK not loaded');
            return;
        }
        
        window.Adsgram.showInterstitial({
            placementId: placementId,
            onAdClosed: function() {
                console.log('Interstitial ad closed');
                onClosed();
            },
            onError: function(error) {
                console.error('Interstitial ad error:', error);
                onError(error);
            }
        });
    }
};

// Load Adsgram SDK
(function() {
    const script = document.createElement('script');
    script.src = 'https://static.adsgram.com/js/adsgram.js';
    script.async = true;
    script.onload = function() {
        console.log('Adsgram SDK loaded');
    };
    script.onerror = function() {
        console.error('Failed to load Adsgram SDK');
    };
    document.head.appendChild(script);
})();
`, placementID)
}

// GetAvailableRewards returns available rewards for user
func (ai *AdsIntegration) GetAvailableRewards(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	var rewards []map[string]interface{}

	for trigger, configs := range ai.configs {
		for _, config := range configs {
			if !config.Enabled {
				continue
			}

			if ai.isOnCooldown(ctx, userID, config) {
				continue
			}

			if ai.isDailyLimitReached(ctx, userID, config) {
				continue
			}

			// Check trigger-specific conditions
			if trigger == TriggerEnergyDepleted {
				energyKey := fmt.Sprintf("user_energy:%d", userID)
				energy, err := ai.redisClient.Get(ctx, energyKey).Int64()
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
			})
		}
	}

	return rewards, nil
}

// UpdateConfig updates ad configuration
func (ai *AdsIntegration) UpdateConfig(trigger AdTrigger, configs []AdConfig) {
	ai.configs[trigger] = configs
}

// DisableAd disables a specific ad placement
func (ai *AdsIntegration) DisableAd(placementID string) {
	for trigger, configs := range ai.configs {
		for i, config := range configs {
			if config.PlacementID == placementID {
				configs[i].Enabled = false
				ai.configs[trigger] = configs
				return
			}
		}
	}
}

// EnableAd enables a specific ad placement
func (ai *AdsIntegration) EnableAd(placementID string) {
	for trigger, configs := range ai.configs {
		for i, config := range configs {
			if config.PlacementID == placementID {
				configs[i].Enabled = true
				ai.configs[trigger] = configs
				return
			}
		}
	}
}

// GetAdRevenue returns estimated revenue from ads
func (ai *AdsIntegration) GetAdRevenue(ctx context.Context, fromDate, toDate time.Time) (map[string]float64, error) {
	revenue := make(map[string]float64)

	// This would integrate with ad provider APIs
	// For now, return estimated revenue based on completed ads

	pattern := "ad_events:*"
	keys, err := ai.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return revenue, fmt.Errorf("failed to get ad event keys: %w", err)
	}

	totalCompleted := 0
	for _, key := range keys {
		events, err := ai.redisClient.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}

		for _, eventJSON := range events {
			var event AdEvent
			err := json.Unmarshal([]byte(eventJSON), &event)
			if err != nil {
				continue
			}

			if event.Completed && event.Timestamp.After(fromDate) && event.Timestamp.Before(toDate) {
				totalCompleted++

				// Estimate revenue (e.g., $0.01 per completed ad)
				revenue["estimated_revenue"] += 0.01
				revenue["completed_ads"]++
			}
		}
	}

	return revenue, nil
}
