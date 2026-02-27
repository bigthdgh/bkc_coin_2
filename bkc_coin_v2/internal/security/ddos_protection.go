package security

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// DDoSProtection - защита от DDoS атак
type DDoSProtection struct {
	requestCounts map[string]*RequestCounter
	mutex         sync.RWMutex
	config        DDoSConfig
	blockList     map[string]time.Time
	blockMutex    sync.RWMutex
}

// RequestCounter - счетчик запросов
type RequestCounter struct {
	Count      int64
	Window     time.Time
	LastReset  time.Time
	IsBlocked  bool
	BlockUntil time.Time
}

// DDoSConfig - конфигурация защиты
type DDoSConfig struct {
	MaxRequestsPerMinute    int           `json:"max_requests_per_minute"`
	MaxRequestsPerHour      int           `json:"max_requests_per_hour"`
	MaxRequestsPerDay       int           `json:"max_requests_per_day"`
	BlockDuration           time.Duration `json:"block_duration"`
	Whitelist               []string      `json:"whitelist"`
	Blacklist               []string      `json:"blacklist"`
	EnableRateLimiting      bool          `json:"enable_rate_limiting"`
	EnableIPBlocking        bool          `json:"enable_ip_blocking"`
	EnableGeoBlocking       bool          `json:"enable_geo_blocking"`
	BlockedCountries        []string      `json:"blocked_countries"`
	EnableAdvancedDetection bool          `json:"enable_advanced_detection"`
}

// NewDDoSProtection - создание защиты от DDoS
func NewDDoSProtection(config DDoSConfig) *DDoSProtection {
	if config.MaxRequestsPerMinute == 0 {
		config.MaxRequestsPerMinute = 60
	}
	if config.MaxRequestsPerHour == 0 {
		config.MaxRequestsPerHour = 1000
	}
	if config.MaxRequestsPerDay == 0 {
		config.MaxRequestsPerDay = 10000
	}
	if config.BlockDuration == 0 {
		config.BlockDuration = time.Hour
	}

	ddos := &DDoSProtection{
		requestCounts: make(map[string]*RequestCounter),
		config:        config,
		blockList:     make(map[string]time.Time),
	}

	// Запускаем очистку счетчиков
	go ddos.cleanupCounters()
	
	return ddos
}

// Middleware - middleware для Gin
func (ddos *DDoSProtection) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := ddos.getClientIP(c)
		
		// Проверяем в whitelist
		if ddos.isWhitelisted(clientIP) {
			c.Next()
			return
		}

		// Проверяем в blacklist
		if ddos.isBlacklisted(clientIP) {
			ddos.blockIP(clientIP, "Blacklisted IP")
			c.JSON(http.StatusForbidden, gin.H{"error": "IP blocked"})
			c.Abort()
			return
		}

		// Проверяем гео-блокировку
		if ddos.config.EnableGeoBlocking {
			country := ddos.getCountryByIP(clientIP)
			if ddos.isCountryBlocked(country) {
				ddos.blockIP(clientIP, fmt.Sprintf("Country %s blocked", country))
				c.JSON(http.StatusForbidden, gin.H{"error": "Region blocked"})
				c.Abort()
				return
			}
		}

		// Проверяем rate limiting
		if ddos.config.EnableRateLimiting {
			if ddos.isRateLimited(clientIP) {
				ddos.blockIP(clientIP, "Rate limit exceeded")
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
				c.Abort()
				return
			}
		}

		// Продвинутая детекция
		if ddos.config.EnableAdvancedDetection {
			if ddos.detectSuspiciousActivity(c) {
				ddos.blockIP(clientIP, "Suspicious activity detected")
				c.JSON(http.StatusForbidden, gin.H{"error": "Suspicious activity"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// getClientIP - получение IP клиента
func (ddos *DDoSProtection) getClientIP(c *gin.Context) string {
	// Проверяем заголовки
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	
	// Возвращаем удаленный адрес
	return c.ClientIP()
}

// isWhitelisted - проверка в whitelist
func (ddos *DDoSProtection) isWhitelisted(ip string) bool {
	for _, whitelistIP := range ddos.config.Whitelist {
		if ip == whitelistIP {
			return true
		}
	}
	return false
}

// isBlacklisted - проверка в blacklist
func (ddos *DDoSProtection) isBlacklisted(ip string) bool {
	for _, blacklistIP := range ddos.config.Blacklist {
		if ip == blacklistIP {
			return true
		}
	}
	return false
}

// getCountryByIP - получение страны по IP (заглушка)
func (ddos *DDoSProtection) getCountryByIP(ip string) string {
	// В реальном приложении здесь будет запрос к GeoIP сервису
	// Для примера возвращаем тестовые данные
	return "US"
}

// isCountryBlocked - проверка заблокированных стран
func (ddos *DDoSProtection) isCountryBlocked(country string) bool {
	for _, blockedCountry := range ddos.config.BlockedCountries {
		if country == blockedCountry {
			return true
		}
	}
	return false
}

// isRateLimited - проверка rate limiting
func (ddos *DDoSProtection) isRateLimited(ip string) bool {
	ddos.mutex.Lock()
	defer ddos.mutex.Unlock()

	now := time.Now()
	counter, exists := ddos.requestCounts[ip]
	
	if !exists {
		counter = &RequestCounter{
			Window:    now,
			LastReset: now,
		}
		ddos.requestCounts[ip] = counter
	}

	// Сбрасываем счетчики при необходимости
	if now.Sub(counter.LastReset) >= time.Minute {
		counter.Count = 0
		counter.LastReset = now
	}

	counter.Count++

	// Проверяем лимиты
	if counter.Count > ddos.config.MaxRequestsPerMinute {
		return true
	}

	// Проверяем часовой лимит
	if now.Sub(counter.Window) >= time.Hour {
		counter.Window = now
	}
	
	if counter.Count > ddos.config.MaxRequestsPerHour {
		return true
	}

	return false
}

// detectSuspiciousActivity - детекция подозрительной активности
func (ddos *DDoSProtection) detectSuspiciousActivity(c *gin.Context) bool {
	ip := ddos.getClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	
	// Проверяем User-Agent
	if userAgent == "" || userAgent == "curl/7.68.0" {
		log.Printf("Suspicious User-Agent from %s: %s", ip, userAgent)
		return true
	}

	// Проверяем частые запросы к одному endpoint
	ddos.mutex.RLock()
	counter, exists := ddos.requestCounts[ip]
	ddos.mutex.RUnlock()
	
	if exists && counter.Count > 10 {
		// Проверяем запрошенные URL
		if ddos.isEndpointAbuse(ip, c.Request.URL.Path) {
			log.Printf("Endpoint abuse detected from %s: %s", ip, c.Request.URL.Path)
			return true
		}
	}

	return false
}

// isEndpointAbuse - проверка злоупотребления endpoint
func (ddos *DDoSProtection) isEndpointAbuse(ip string, endpoint string) bool {
	// В реальном приложении здесь будет отслеживание запросов к endpoint
	// Для примера блокируем частые запросы к API
	if endpoint == "/api/tap" || endpoint == "/api/claim" {
		return true
	}
	return false
}

// blockIP - блокировка IP
func (ddos *DDoSProtection) blockIP(ip string, reason string) {
	if !ddos.config.EnableIPBlocking {
		return
	}

	ddos.blockMutex.Lock()
	defer ddos.blockMutex.Unlock()

	ddos.blockList[ip] = time.Now().Add(ddos.config.BlockDuration)
	
	// Обновляем счетчик
	ddos.mutex.Lock()
	if counter, exists := ddos.requestCounts[ip]; exists {
		counter.IsBlocked = true
		counter.BlockUntil = time.Now().Add(ddos.config.BlockDuration)
	}
	ddos.mutex.Unlock()

	log.Printf("IP %s blocked for %v. Reason: %s", ip, ddos.config.BlockDuration, reason)
}

// isIPBlocked - проверка заблокированного IP
func (ddos *DDoSProtection) isIPBlocked(ip string) bool {
	ddos.blockMutex.RLock()
	defer ddos.blockMutex.RUnlock()

	blockTime, exists := ddos.blockList[ip]
	if !exists {
		return false
	}

	if time.Now().After(blockTime) {
		// Разблокируем IP
		ddos.blockMutex.Lock()
		delete(ddos.blockList, ip)
		ddos.blockMutex.Unlock()
		return false
	}

	return true
}

// cleanupCounters - очистка старых счетчиков
func (ddos *DDoSProtection) cleanupCounters() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ddos.mutex.Lock()
		now := time.Now()
		
		for ip, counter := range ddos.requestCounts {
			// Удаляем старые счетчики
			if now.Sub(counter.LastReset) > 24*time.Hour {
				delete(ddos.requestCounts, ip)
			}
			
			// Разблокируем IP если время блокировки истекло
			if counter.IsBlocked && now.After(counter.BlockUntil) {
				counter.IsBlocked = false
				counter.Count = 0
			}
		}
		
		ddos.mutex.Unlock()
	}
}

// GetStats - получение статистики DDoS защиты
func (ddos *DDoSProtection) GetStats() map[string]interface{} {
	ddos.mutex.RLock()
	ddos.blockMutex.RLock()
	defer ddos.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	stats["active_ips"] = len(ddos.requestCounts)
	stats["blocked_ips"] = len(ddos.blockList)
	
	totalRequests := int64(0)
	for _, counter := range ddos.requestCounts {
		totalRequests += counter.Count
	}
	stats["total_requests"] = totalRequests
	
	blockedCount := int64(0)
	for _, counter := range ddos.requestCounts {
		if counter.IsBlocked {
			blockedCount++
		}
	}
	stats["currently_blocked"] = blockedCount
	
	return stats
}

// UnblockIP - ручная разблокировка IP
func (ddos *DDoSProtection) UnblockIP(ip string) error {
	ddos.blockMutex.Lock()
	defer ddos.blockMutex.Unlock()

	delete(ddos.blockList, ip)
	
	ddos.mutex.Lock()
	if counter, exists := ddos.requestCounts[ip]; exists {
		counter.IsBlocked = false
		counter.Count = 0
	}
	ddos.mutex.Unlock()

	log.Printf("IP %s manually unblocked", ip)
	return nil
}

// AddToWhitelist - добавление в whitelist
func (ddos *DDoSProtection) AddToWhitelist(ip string) {
	ddos.config.Whitelist = append(ddos.config.Whitelist, ip)
	log.Printf("IP %s added to whitelist", ip)
}

// AddToBlacklist - добавление в blacklist
func (ddos *DDoSProtection) AddToBlacklist(ip string) {
	ddos.config.Blacklist = append(ddos.config.Blacklist, ip)
	ddos.blockIP(ip, "Manually blacklisted")
	log.Printf("IP %s added to blacklist", ip)
}

// UpdateConfig - обновление конфигурации
func (ddos *DDoSProtection) UpdateConfig(config DDoSConfig) {
	ddos.mutex.Lock()
	defer ddos.mutex.Unlock()
	
	ddos.config = config
	log.Printf("DDoS protection config updated")
}

// GetBlockedIPs - получение списка заблокированных IP
func (ddos *DDoSProtection) GetBlockedIPs() map[string]time.Time {
	ddos.blockMutex.RLock()
	defer ddos.blockMutex.RUnlock()
	
	blocked := make(map[string]time.Time)
	for ip, blockTime := range ddos.blockList {
		blocked[ip] = blockTime
	}
	
	return blocked
}

// TestLoad - тестирование нагрузки
func (ddos *DDoSProtection) TestLoad(ctx context.Context, requests int, duration time.Duration) error {
	log.Printf("Starting load test: %d requests in %v", requests, duration)
	
	interval := duration / time.Duration(requests)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	successful := 0
	blocked := 0
	
	for i := 0; i < requests; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Симулируем запрос с разных IP
			ip := fmt.Sprintf("192.168.1.%d", (i%254)+1)
			
			if ddos.isRateLimited(ip) {
				blocked++
			} else {
				successful++
			}
		}
	}
	
	log.Printf("Load test completed: %d successful, %d blocked", successful, blocked)
	return nil
}
