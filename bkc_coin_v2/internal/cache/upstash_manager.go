package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// 🚀 UpstashManager для управления Redis кэшем
type UpstashManager struct {
	client *redis.Client
}

// CacheEntry структура кэша
type CacheEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// RateLimitEntry структура для rate limiting
type RateLimitEntry struct {
	Count     int       `json:"count"`
	WindowStart time.Time `json:"window_start"`
}

// NewUpstashManager создает менеджер Upstash Redis
func NewUpstashManager(url, token string) *UpstashManager {
	opt := &redis.Options{
		Addr:     url,
		Password: token,
		DB:       0,
		PoolSize: 10,
		MinIdleConns: 5,
		MaxIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	}

	return &UpstashManager{
		client: redis.NewClient(opt),
	}
}

// Initialize инициализирует Upstash Redis
func (u *UpstashManager) Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Проверяем соединение
	_, err := u.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Upstash Redis: %w", err)
	}

	log.Printf("🚀 Upstash Redis initialized: %s", u.client.Options().Addr)
	return nil
}

// Set сохраняет значение в кэш
func (u *UpstashManager) Set(key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entry := CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(expiration),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	return u.client.Set(ctx, key, data, expiration).Err()
}

// Get получает значение из кэша
func (u *UpstashManager) Get(key string) (interface{}, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	data, err := u.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to get cache key: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Проверяем, не истекло ли значение
	if time.Now().After(entry.ExpiresAt) {
		// Удаляем истекший ключ
		u.Delete(key)
		return nil, false, nil
	}

	return entry.Value, true, nil
}

// Delete удаляет значение из кэша
func (u *UpstashManager) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return u.client.Del(ctx, key).Err()
}

// SetUserSession сохраняет сессию пользователя
func (u *UpstashManager) SetUserSession(userID int64, sessionData map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%d", userID)
	return u.Set(key, sessionData, expiration)
}

// GetUserSession получает сессию пользователя
func (u *UpstashManager) GetUserSession(userID int64) (map[string]interface{}, bool, error) {
	key := fmt.Sprintf("session:%d", userID)
	value, found, err := u.Get(key)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	sessionData, ok := value.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("invalid session data type")
	}

	return sessionData, true, nil
}

// DeleteUserSession удаляет сессию пользователя
func (u *UpstashManager) DeleteUserSession(userID int64) error {
	key := fmt.Sprintf("session:%d", userID)
	return u.Delete(key)
}

// SetRateLimit устанавливает rate limit
func (u *UpstashManager) SetRateLimit(identifier string, window time.Duration, maxRequests int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("rate_limit:%s", identifier)
	
	// Получаем текущие данные
	data, err := u.client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("failed to get rate limit: %w", err)
	}

	var entry RateLimitEntry
	if err != redis.Nil {
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			return false, fmt.Errorf("failed to unmarshal rate limit entry: %w", err)
		}
	}

	now := time.Now()
	
	// Если окно истекло, сбрасываем счетчик
	if now.Sub(entry.WindowStart) >= window {
		entry.Count = 0
		entry.WindowStart = now
	}

	// Проверяем лимит
	if entry.Count >= maxRequests {
		return false, nil // Превышен лимит
	}

	// Увеличиваем счетчик
	entry.Count++
	
	// Сохраняем обновленные данные
	entryData, err := json.Marshal(entry)
	if err != nil {
		return false, fmt.Errorf("failed to marshal rate limit entry: %w", err)
	}

	err = u.client.Set(ctx, key, entryData, window).Err()
	if err != nil {
		return false, fmt.Errorf("failed to set rate limit: %w", err)
	}

	return true, nil
}

// GetRateLimit получает текущий rate limit
func (u *UpstashManager) GetRateLimit(identifier string) (int, time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("rate_limit:%s", identifier)
	
	data, err := u.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, time.Time{}, nil
		}
		return 0, time.Time{}, fmt.Errorf("failed to get rate limit: %w", err)
	}

	var entry RateLimitEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to unmarshal rate limit entry: %w", err)
	}

	return entry.Count, entry.WindowStart, nil
}

// SetAntiFraudData сохраняет анти-фрод данные
func (u *UpstashManager) SetAntiFraudData(userID int64, fraudData map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("anti_fraud:%d", userID)
	return u.Set(key, fraudData, expiration)
}

// GetAntiFraudData получает анти-фрод данные
func (u *UpstashManager) GetAntiFraudData(userID int64) (map[string]interface{}, bool, error) {
	key := fmt.Sprintf("anti_fraud:%d", userID)
	value, found, err := u.Get(key)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	fraudData, ok := value.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("invalid anti-fraud data type")
	}

	return fraudData, true, nil
}

// SetUserCache сохраняет данные пользователя в кэш
func (u *UpstashManager) SetUserCache(userID int64, userData map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("user:%d", userID)
	return u.Set(key, userData, expiration)
}

// GetUserCache получает данные пользователя из кэша
func (u *UpstashManager) GetUserCache(userID int64) (map[string]interface{}, bool, error) {
	key := fmt.Sprintf("user:%d", userID)
	value, found, err := u.Get(key)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	userData, ok := value.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("invalid user data type")
	}

	return userData, true, nil
}

// InvalidateUserCache удаляет кэш пользователя
func (u *UpstashManager) InvalidateUserCache(userID int64) error {
	keys := []string{
		fmt.Sprintf("user:%d", userID),
		fmt.Sprintf("session:%d", userID),
		fmt.Sprintf("rate_limit:%d", userID),
		fmt.Sprintf("anti_fraud:%d", userID),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return u.client.Del(ctx, keys...).Err()
}

// Ping проверяет соединение с Upstash Redis
func (u *UpstashManager) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return u.client.Ping(ctx).Err()
}

// Close закрывает соединение с Upstash Redis
func (u *UpstashManager) Close() error {
	return u.client.Close()
}

// GetStats получает статистику Redis
func (u *UpstashManager) GetStats() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	info, err := u.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	stats := map[string]interface{}{
		"info": info,
		"connected_clients": u.client.PoolStats().TotalConns,
		"idle_connections": u.client.PoolStats().IdleConns,
		"stale_connections": u.client.PoolStats().StaleConns,
	}

	return stats, nil
}
