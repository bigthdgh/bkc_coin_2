package security

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"bkc_coin_v2/internal/cache"
)

// 🛡️ MultiServiceGuard для защиты и rate limiting
type MultiServiceGuard struct {
	secretKey  string
	redis      *cache.UpstashManager
	rateLimits map[string]int
	mu         sync.RWMutex
}

// NewMultiServiceGuard создает MultiService Guard
func NewMultiServiceGuard(secretKey string, redis *cache.UpstashManager) *MultiServiceGuard {
	return &MultiServiceGuard{
		secretKey:  secretKey,
		redis:      redis,
		rateLimits: map[string]int{
			"render":  800,  // Основной сервис
			"koyeb":   600,  // Резервный
			"render2": 500,  // Дополнительный
			"render3": 400,  // Финальный
		},
	}
}

// MultiServiceRateLimit создает middleware для rate limiting
func (g *MultiServiceGuard) MultiServiceRateLimit(serviceType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем rate limit для сервиса
			rateLimit := g.getRateLimit(serviceType)
			
			// Проверяем rate limit по IP
			clientIP := g.getClientIP(r)
			allowed, err := g.checkRateLimit(clientIP, rateLimit)
			if err != nil {
				http.Error(w, "Rate limit error", http.StatusInternalServerError)
				return
			}
			
			if !allowed {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Проверяем rate limit по пользователю
			userID := g.getUserID(r)
			if userID > 0 {
				allowed, err := g.checkUserRateLimit(userID, rateLimit)
				if err != nil {
					http.Error(w, "Rate limit error", http.StatusInternalServerError)
					return
				}
				
				if !allowed {
					w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimit))
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
					http.Error(w, "User rate limit exceeded", http.StatusTooManyRequests)
					return
				}
			}

			// Устанавливаем заголовки rate limit
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rateLimit-1))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))

			next.ServeHTTP(w, r)
		})
	}
}

// CheckAntiFraud проверяет анти-фрод
func (g *MultiServiceGuard) CheckAntiFraud(userID int64, r *http.Request) (bool, string) {
	clientIP := g.getClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	
	// Получаем анти-фрод данные
	fraudData, found, err := g.redis.GetAntiFraudData(userID)
	if err != nil {
		return false, "Anti-fraud check error"
	}
	
	if !found {
		// Создаем новые анти-фрод данные
		fraudData = map[string]interface{}{
			"first_seen":    time.Now(),
			"last_seen":     time.Now(),
			"request_count":  1,
			"ips":           []string{clientIP},
			"user_agents":    []string{userAgent},
			"suspicious":    false,
		}
		
		err := g.redis.SetAntiFraudData(userID, fraudData, 24*time.Hour)
		if err != nil {
			return false, "Anti-fraud data error"
		}
		
		return true, "OK"
	}
	
	// Обновляем существующие данные
	data, ok := fraudData.(map[string]interface{})
	if !ok {
		return false, "Invalid fraud data"
	}
	
	// Обновляем счетчик запросов
	requestCount := 1
	if rc, exists := data["request_count"]; exists {
		if rc, ok := rc.(float64); ok {
			requestCount = int(rc) + 1
		}
	}
	
	// Обновляем IP и User-Agent
	ips := []string{clientIP}
	if ipsList, exists := data["ips"]; exists {
		if ipsList, ok := ipsList.([]interface{}); ok {
			for _, ip := range ipsList {
				if ipStr, ok := ip.(string); ok && ipStr == clientIP {
					// IP уже в списке
					ips = append(ips, ipStr)
					break
				}
			}
		}
	}
	
	userAgents := []string{userAgent}
	if uaList, exists := data["user_agents"]; exists {
		if uaList, ok := uaList.([]interface{}); ok {
			for _, ua := range uaList {
				if uaStr, ok := ua.(string); ok && uaStr == userAgent {
					// User-Agent уже в списке
					userAgents = append(userAgents, uaStr)
					break
				}
			}
		}
	}
	
	// Проверяем на подозрительную активность
	suspicious := false
	if requestCount > 1000 { // Более 1000 запросов в 24 часа
		suspicious = true
	}
	
	if len(ips) > 10 { // Более 10 разных IP
		suspicious = true
	}
	
	if len(userAgents) > 5 { // Более 5 разных User-Agent
		suspicious = true
	}
	
	// Обновляем данные
	updatedData := map[string]interface{}{
		"first_seen":     data["first_seen"],
		"last_seen":      time.Now(),
		"request_count":  requestCount,
		"ips":           ips,
		"user_agents":    userAgents,
		"suspicious":     suspicious,
	}
	
	err = g.redis.SetAntiFraudData(userID, updatedData, 24*time.Hour)
	if err != nil {
		return false, "Anti-fraud update error"
	}
	
	if suspicious {
		return false, "Suspicious activity detected"
	}
	
	return true, "OK"
}

// ValidateToken проверяет JWT токен
func (g *MultiServiceGuard) ValidateToken(token string) (bool, int64, error) {
	// Простая валидация токена (в реальном проекте здесь будет JWT валидация)
	if token == "" {
		return false, 0, fmt.Errorf("empty token")
	}
	
	// Для demo purposes - в реальном проекте здесь будет JWT парсинг
	// и проверка подписи с помощью g.secretKey
	if len(token) < 10 {
		return false, 0, fmt.Errorf("invalid token format")
	}
	
	// Возвращаем demo userID (в реальном проекте будет из JWT claims)
	return true, 12345, nil
}

// getClientIP получает IP адрес клиента
func (g *MultiServiceGuard) getClientIP(r *http.Request) string {
	// Проверяем X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Берем первый IP из списка
		if idx := len(xff); idx > 0 {
			if commaIdx := 0; commaIdx < len(xff); commaIdx++ {
				if xff[commaIdx] == ',' {
					break
				}
			}
			return xff[:commaIdx]
		}
		return xff
	}
	
	// Проверяем X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Проверяем X-Forwarded header
	if xf := r.Header.Get("X-Forwarded"); xf != "" {
		return xf
	}
	
	// Возвращаем RemoteAddr
	return r.RemoteAddr
}

// getUserID получает ID пользователя из запроса
func (g *MultiServiceGuard) getUserID(r *http.Request) int64 {
	// Проверяем Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Парсим Bearer token
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			valid, userID, err := g.ValidateToken(token)
			if valid && err == nil {
				return userID
			}
		}
	}
	
	// Проверяем query параметр
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			return userID
		}
	}
	
	// Проверяем сессию в Redis
	sessionID := g.getSessionID(r)
	if sessionID != "" {
		if sessionData, found, err := g.redis.GetUserSession(12345); err == nil && found {
			if userID, ok := sessionData["user_id"]; ok {
				if uid, ok := userID.(float64); ok {
					return int64(uid)
				}
			}
		}
	}
	
	return 0
}

// getSessionID получает ID сессии
func (g *MultiServiceGuard) getSessionID(r *http.Request) string {
	// Проверяем cookie
	if cookie, err := r.Cookie("session_id"); err == nil && cookie != "" {
		return cookie.Value
	}
	
	// Проверяем header
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		return sessionID
	}
	
	// Проверяем query параметр
	return r.URL.Query().Get("session_id")
}

// getRateLimit получает rate limit для сервиса
func (g *MultiServiceGuard) getRateLimit(serviceType string) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	if limit, exists := g.rateLimits[serviceType]; exists {
		return limit
	}
	
	return 500 // Default rate limit
}

// checkRateLimit проверяет rate limit по IP
func (g *MultiServiceGuard) checkRateLimit(identifier string, maxRequests int) (bool, error) {
	return g.redis.SetRateLimit(identifier, time.Minute, maxRequests)
}

// checkUserRateLimit проверяет rate limit по пользователю
func (g *MultiServiceGuard) checkUserRateLimit(userID int64, maxRequests int) (bool, error) {
	identifier := fmt.Sprintf("user:%d", userID)
	return g.redis.SetRateLimit(identifier, time.Minute, maxRequests)
}

// BlockUser блокирует пользователя
func (g *MultiServiceGuard) BlockUser(userID int64, reason string, duration time.Duration) error {
	blockData := map[string]interface{}{
		"reason":   reason,
		"blocked_at": time.Now(),
		"duration":  duration,
	}
	
	key := fmt.Sprintf("blocked_user:%d", userID)
	return g.redis.Set(key, blockData, duration)
}

// IsUserBlocked проверяет, заблокирован ли пользователь
func (g *MultiServiceGuard) IsUserBlocked(userID int64) (bool, error) {
	key := fmt.Sprintf("blocked_user:%d", userID)
	_, found, err := g.redis.Get(key)
	if err != nil {
		return false, err
	}
	
	return found, nil
}

// UnblockUser разблокирует пользователя
func (g *MultiServiceGuard) UnblockUser(userID int64) error {
	key := fmt.Sprintf("blocked_user:%d", userID)
	return g.redis.Delete(key)
}

// GetSecurityStats получает статистику безопасности
func (g *MultiServiceGuard) GetSecurityStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"rate_limits": g.rateLimits,
		"redis_stats": map[string]interface{}{},
	}
	
	// Получаем статистику Redis
	if redisStats, err := g.redis.GetStats(); err == nil {
		stats["redis_stats"] = redisStats
	}
	
	return stats, nil
}
