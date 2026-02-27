package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/jmoiron/sqlx"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeCounter     MetricType = "counter"
	MetricTypeGauge       MetricType = "gauge"
	MetricTypeHistogram    MetricType = "histogram"
	MetricTypeTimer       MetricType = "timer"
)

// EventType represents different event types
type EventType string

const (
	EventTypeUserTap        EventType = "user_tap"
	EventTypeUserLogin      EventType = "user_login"
	EventTypeUserLogout     EventType = "user_logout"
	EventTypeTransaction    EventType = "transaction"
	EventTypeNFTPurchase   EventType = "nft_purchase"
	EventTypeReferral      EventType = "referral"
	EventTypeAchievement   EventType = "achievement"
	EventTypeError         EventType = "error"
	EventTypePageView      EventType = "page_view"
	EventTypeAPICall       EventType = "api_call"
)

// Metric represents a single metric
type Metric struct {
	Name      string      `json:"name"`
	Type      MetricType `json:"type"`
	Value     float64    `json:"value"`
	Tags      map[string]string `json:"tags,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Event represents a tracked event
type Event struct {
	Type      EventType         `json:"type"`
	UserID    int64            `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// AnalyticsSystem handles analytics and monitoring
type AnalyticsSystem struct {
	redisClient *redis.Client
	db          *sqlx.DB
}

// NewAnalyticsSystem creates a new analytics system
func NewAnalyticsSystem(redisClient *redis.Client, db *sqlx.DB) *AnalyticsSystem {
	return &AnalyticsSystem{
		redisClient: redisClient,
		db:          db,
	}
}

// RecordMetric records a metric
func (as *AnalyticsSystem) RecordMetric(ctx context.Context, metric *Metric) error {
	metric.Timestamp = time.Now()
	
	// Store in Redis for real-time analytics
	key := fmt.Sprintf("metrics:%s:%d", metric.Name, metric.Timestamp.Unix())
	
	metricJSON, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	// Store with expiration (7 days)
	err = as.redisClient.Set(ctx, key, metricJSON, 7*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store metric in Redis: %w", err)
	}

	// Add to time series for aggregation
	timeSeriesKey := fmt.Sprintf("timeseries:%s", metric.Name)
	err = as.redisClient.ZAdd(ctx, timeSeriesKey, redis.Z{
		Score:  float64(metric.Timestamp.Unix()),
		Member: string(metricJSON),
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add to time series: %w", err)
	}

	// Keep only last 1000 entries per metric
	err = as.redisClient.ZRemRangeByRank(ctx, timeSeriesKey, 0, -1001).Err()
	if err != nil {
		log.Printf("Failed to trim time series: %v", err)
	}

	// Set expiration for time series
	err = as.redisClient.Expire(ctx, timeSeriesKey, 7*24*time.Hour).Err()
	if err != nil {
		log.Printf("Failed to set expiration for time series: %v", err)
	}

	return nil
}

// RecordEvent records an event
func (as *AnalyticsSystem) RecordEvent(ctx context.Context, event *Event) error {
	event.Timestamp = time.Now()
	
	// Store in Redis for real-time analytics
	key := fmt.Sprintf("events:%s:%d", event.Type, event.Timestamp.Unix())
	
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Store with expiration (30 days)
	err = as.redisClient.Set(ctx, key, eventJSON, 30*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store event in Redis: %w", err)
	}

	// Add to event time series
	timeSeriesKey := fmt.Sprintf("events:%s", event.Type)
	err = as.redisClient.ZAdd(ctx, timeSeriesKey, redis.Z{
		Score:  float64(event.Timestamp.Unix()),
		Member: string(eventJSON),
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add to event time series: %w", err)
	}

	// Keep only last 1000 entries per event type
	err = as.redisClient.ZRemRangeByRank(ctx, timeSeriesKey, 0, -1001).Err()
	if err != nil {
		log.Printf("Failed to trim event time series: %v", err)
	}

	// Set expiration
	err = as.redisClient.Expire(ctx, timeSeriesKey, 30*24*time.Hour).Err()
	if err != nil {
		log.Printf("Failed to set expiration for event time series: %v", err)
	}

	// Store in database for long-term analytics
	err = as.storeEventInDB(ctx, event)
	if err != nil {
		log.Printf("Failed to store event in database: %v", err)
		// Don't return error as Redis storage succeeded
	}

	return nil
}

// GetMetrics retrieves metrics for a given time range
func (as *AnalyticsSystem) GetMetrics(ctx context.Context, metricName string, from, to time.Time) ([]*Metric, error) {
	timeSeriesKey := fmt.Sprintf("timeseries:%s", metricName)
	
	// Get metrics from time series
	minScore := float64(from.Unix())
	maxScore := float64(to.Unix())
	
	results, err := as.redisClient.ZRangeByScore(ctx, timeSeriesKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", minScore),
		Max: fmt.Sprintf("%f", maxScore),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics from time series: %w", err)
	}

	var metrics []*Metric
	for _, result := range results {
		var metric Metric
		err := json.Unmarshal([]byte(result), &metric)
		if err != nil {
			log.Printf("Failed to unmarshal metric: %v", err)
			continue
		}
		metrics = append(metrics, &metric)
	}

	return metrics, nil
}

// GetEvents retrieves events for a given time range
func (as *AnalyticsSystem) GetEvents(ctx context.Context, eventType EventType, from, to time.Time) ([]*Event, error) {
	timeSeriesKey := fmt.Sprintf("events:%s", eventType)
	
	// Get events from time series
	minScore := float64(from.Unix())
	maxScore := float64(to.Unix())
	
	results, err := as.redisClient.ZRangeByScore(ctx, timeSeriesKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", minScore),
		Max: fmt.Sprintf("%f", maxScore),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get events from time series: %w", err)
	}

	var events []*Event
	for _, result := range results {
		var event Event
		err := json.Unmarshal([]byte(result), &event)
		if err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}

// GetAggregatedMetrics returns aggregated metrics for a time period
func (as *AnalyticsSystem) GetAggregatedMetrics(ctx context.Context, metricName string, period string) (map[string]float64, error) {
	var from time.Time
	var to time.Time = time.Now()

	switch period {
	case "1h":
		from = to.Add(-1 * time.Hour)
	case "24h":
		from = to.Add(-24 * time.Hour)
	case "7d":
		from = to.Add(-7 * 24 * time.Hour)
	case "30d":
		from = to.Add(-30 * 24 * time.Hour)
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	metrics, err := as.GetMetrics(ctx, metricName, from, to)
	if err != nil {
		return nil, err
	}

	return as.aggregateMetrics(metrics), nil
}

// aggregateMetrics aggregates metrics by type
func (as *AnalyticsSystem) aggregateMetrics(metrics []*Metric) map[string]float64 {
	result := make(map[string]float64)
	
	if len(metrics) == 0 {
		return result
	}

	var sum, min, max, count float64
	min = metrics[0].Value
	max = metrics[0].Value

	for _, metric := range metrics {
		sum += metric.Value
		if metric.Value < min {
			min = metric.Value
		}
		if metric.Value > max {
			max = metric.Value
		}
		count++
	}

	result["count"] = count
	result["sum"] = sum
	result["avg"] = sum / count
	result["min"] = min
	result["max"] = max

	return result
}

// GetRealTimeStats returns real-time statistics
func (as *AnalyticsSystem) GetRealTimeStats(ctx context.Context) (*RealTimeStats, error) {
	stats := &RealTimeStats{
		Timestamp: time.Now(),
	}

	// Get active users (last 5 minutes)
	activeUsersKey := "stats:active_users"
	activeUsers, err := as.redisClient.SCard(ctx, activeUsersKey).Result()
	if err != nil {
		log.Printf("Failed to get active users: %v", err)
	} else {
		stats.ActiveUsers = int(activeUsers)
	}

	// Get total taps in last hour
	taps, err := as.GetAggregatedMetrics(ctx, "taps_per_second", "1h")
	if err != nil {
		log.Printf("Failed to get taps metrics: %v", err)
	} else {
		stats.TapsPerHour = taps["sum"]
		stats.AverageTapsPerSecond = taps["avg"]
	}

	// Get transactions in last hour
	transactions, err := as.GetAggregatedMetrics(ctx, "transactions", "1h")
	if err != nil {
		log.Printf("Failed to get transactions metrics: %v", err)
	} else {
		stats.TransactionsPerHour = transactions["sum"]
	}

	// Get current price
	price, err := as.redisClient.Get(ctx, "current_price").Result()
	if err != nil {
		log.Printf("Failed to get current price: %v", err)
	} else {
		stats.CurrentPrice = price
	}

	// Get system health
	health, err := as.redisClient.Get(ctx, "system_health").Result()
	if err != nil {
		log.Printf("Failed to get system health: %v", err)
	} else {
		stats.SystemHealth = health
	}

	return stats, nil
}

// RealTimeStats represents real-time statistics
type RealTimeStats struct {
	Timestamp              time.Time `json:"timestamp"`
	ActiveUsers            int       `json:"active_users"`
	TapsPerHour           float64   `json:"taps_per_hour"`
	AverageTapsPerSecond  float64   `json:"average_taps_per_second"`
	TransactionsPerHour    float64   `json:"transactions_per_hour"`
	CurrentPrice          string    `json:"current_price"`
	SystemHealth          string    `json:"system_health"`
}

// GetUserAnalytics returns analytics for a specific user
func (as *AnalyticsSystem) GetUserAnalytics(ctx context.Context, userID int64) (*UserAnalytics, error) {
	analytics := &UserAnalytics{
		UserID:    userID,
		Timestamp: time.Now(),
	}

	// Get user events from database
	query := `
		SELECT 
			COUNT(*) as total_events,
			COUNT(CASE WHEN type = 'user_tap' THEN 1 END) as total_taps,
			COUNT(CASE WHEN type = 'transaction' THEN 1 END) as total_transactions,
			COUNT(CASE WHEN type = 'nft_purchase' THEN 1 END) as nft_purchases,
			COUNT(CASE WHEN type = 'achievement' THEN 1 END) as achievements,
			MIN(created_at) as first_seen,
			MAX(created_at) as last_seen
		FROM user_events 
		WHERE user_id = $1 AND created_at > NOW() - INTERVAL '30 days'
	`

	err := as.db.Get(analytics, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user analytics: %w", err)
	}

	// Calculate session duration
	if analytics.FirstSeen != nil && analytics.LastSeen != nil {
		duration := analytics.LastSeen.Sub(*analytics.FirstSeen)
		analytics.TotalSessionTime = int(duration.Seconds())
	}

	return analytics, nil
}

// UserAnalytics represents user-specific analytics
type UserAnalytics struct {
	UserID            int64      `json:"user_id" db:"user_id"`
	Timestamp         time.Time   `json:"timestamp"`
	TotalEvents       int         `json:"total_events" db:"total_events"`
	TotalTaps        int         `json:"total_taps" db:"total_taps"`
	TotalTransactions int         `json:"total_transactions" db:"total_transactions"`
	NFTPurchases     int         `json:"nft_purchases" db:"nft_purchases"`
	Achievements      int         `json:"achievements" db:"achievements"`
	FirstSeen        *time.Time  `json:"first_seen" db:"first_seen"`
	LastSeen         *time.Time  `json:"last_seen" db:"last_seen"`
	TotalSessionTime int         `json:"total_session_time"`
}

// storeEventInDB stores event in database for long-term storage
func (as *AnalyticsSystem) storeEventInDB(ctx context.Context, event *Event) error {
	// This would store events in a database table for long-term analytics
	// For now, we'll just log it
	log.Printf("Storing event in DB: %+v", event)
	return nil
}

// TrackActiveUser tracks an active user
func (as *AnalyticsSystem) TrackActiveUser(ctx context.Context, userID int64) error {
	activeUsersKey := "stats:active_users"
	
	// Add user to active set
	err := as.redisClient.SAdd(ctx, activeUsersKey, userID).Err()
	if err != nil {
		return fmt.Errorf("failed to add active user: %w", err)
	}

	// Set expiration (5 minutes)
	err = as.redisClient.Expire(ctx, activeUsersKey, 5*time.Minute).Err()
	if err != nil {
		log.Printf("Failed to set expiration for active users: %v", err)
	}

	return nil
}

// RecordAPICall records an API call
func (as *AnalyticsSystem) RecordAPICall(ctx context.Context, endpoint string, method string, statusCode int, duration time.Duration) error {
	metric := &Metric{
		Name:  "api_calls",
		Type:  MetricTypeCounter,
		Value: 1,
		Tags: map[string]string{
			"endpoint":    endpoint,
			"method":      method,
			"status_code": fmt.Sprintf("%d", statusCode),
		},
	}

	err := as.RecordMetric(ctx, metric)
	if err != nil {
		return fmt.Errorf("failed to record API call metric: %w", err)
	}

	// Record response time
	durationMetric := &Metric{
		Name:  "api_response_time",
		Type:  MetricTypeTimer,
		Value: duration.Seconds(),
		Tags: map[string]string{
			"endpoint": endpoint,
			"method":   method,
		},
	}

	err = as.RecordMetric(ctx, durationMetric)
	if err != nil {
		return fmt.Errorf("failed to record response time metric: %w", err)
	}

	return nil
}

// RecordError records an error
func (as *AnalyticsSystem) RecordError(ctx context.Context, errorType string, message string, userID int64) error {
	event := &Event{
		Type: EventTypeError,
		Data: map[string]interface{}{
			"error_type": errorType,
			"message":    message,
		},
		Timestamp: time.Now(),
	}

	if userID > 0 {
		event.UserID = userID
	}

	return as.RecordEvent(ctx, event)
}

// GetSystemHealth returns system health metrics
func (as *AnalyticsSystem) GetSystemHealth(ctx context.Context) (*SystemHealth, error) {
	health := &SystemHealth{
		Timestamp: time.Now(),
	}

	// Check Redis connection
	_, err := as.redisClient.Ping(ctx).Result()
	if err != nil {
		health.RedisStatus = "error"
		health.RedisError = err.Error()
	} else {
		health.RedisStatus = "healthy"
	}

	// Check database connection
	err = as.db.Ping()
	if err != nil {
		health.DatabaseStatus = "error"
		health.DatabaseError = err.Error()
	} else {
		health.DatabaseStatus = "healthy"
	}

	// Get memory usage
	// This would require additional system monitoring
	health.MemoryUsage = "N/A"
	health.CPUUsage = "N/A"

	// Determine overall health
	if health.RedisStatus == "healthy" && health.DatabaseStatus == "healthy" {
		health.OverallStatus = "healthy"
	} else {
		health.OverallStatus = "degraded"
	}

	return health, nil
}

// SystemHealth represents system health metrics
type SystemHealth struct {
	Timestamp      time.Time `json:"timestamp"`
	RedisStatus    string    `json:"redis_status"`
	RedisError     string    `json:"redis_error,omitempty"`
	DatabaseStatus string    `json:"database_status"`
	DatabaseError  string    `json:"database_error,omitempty"`
	MemoryUsage    string    `json:"memory_usage"`
	CPUUsage      string    `json:"cpu_usage"`
	OverallStatus  string    `json:"overall_status"`
}

// CreateAnalyticsTables creates necessary database tables
func (as *AnalyticsSystem) CreateAnalyticsTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user_events (
			id SERIAL PRIMARY KEY,
			user_id BIGINT,
			type VARCHAR(50) NOT NULL,
			data JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_user_events_user_id (user_id),
			INDEX idx_user_events_type (type),
			INDEX idx_user_events_created_at (created_at)
		)`,
		`CREATE TABLE IF NOT EXISTS metrics_log (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			type VARCHAR(20) NOT NULL,
			value DECIMAL(20,8) NOT NULL,
			tags JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_metrics_log_name (name),
			INDEX idx_metrics_log_created_at (created_at)
		)`,
	}

	for _, query := range queries {
		_, err := as.db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create analytics table: %w", err)
		}
	}

	return nil
}
