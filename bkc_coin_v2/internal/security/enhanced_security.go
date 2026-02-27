package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret         string        `json:"jwt_secret"`
	JWTExpiration     time.Duration `json:"jwt_expiration"`
	MaxLoginAttempts  int           `json:"max_login_attempts"`
	LockoutDuration   time.Duration `json:"lockout_duration"`
	RateLimitRequests int           `json:"rate_limit_requests"`
	RateLimitWindow   time.Duration `json:"rate_limit_window"`
	PasswordMinLength int           `json:"password_min_length"`
	SessionTimeout    time.Duration `json:"session_timeout"`
}

// EnhancedSecurity provides comprehensive security features
type EnhancedSecurity struct {
	config      *SecurityConfig
	redisClient *redis.Client
	jwtSecret   []byte
	mu          sync.RWMutex
}

// NewEnhancedSecurity creates a new enhanced security system
func NewEnhancedSecurity(config *SecurityConfig, redisClient *redis.Client) *EnhancedSecurity {
	return &EnhancedSecurity{
		config:      config,
		redisClient: redisClient,
		jwtSecret:   []byte(config.JWTSecret),
	}
}

// GenerateSecureToken generates a secure random token
func (es *EnhancedSecurity) GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashPassword securely hashes a password
func (es *EnhancedSecurity) HashPassword(password string) (string, error) {
	if len(password) < es.config.PasswordMinLength {
		return "", fmt.Errorf("password must be at least %d characters", es.config.PasswordMinLength)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (es *EnhancedSecurity) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates a JWT token
func (es *EnhancedSecurity) GenerateJWT(userID int64, claims map[string]interface{}) (string, error) {
	now := time.Now()

	// Create claims
	jwtClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     now.Add(es.config.JWTExpiration).Unix(),
		"iat":     now.Unix(),
		"iss":     "bkc-coin",
		"aud":     "bkc-coin-users",
	}

	// Add custom claims
	for key, value := range claims {
		jwtClaims[key] = value
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	// Sign token
	tokenString, err := token.SignedString(es.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}

// VerifyJWT verifies a JWT token
func (es *EnhancedSecurity) VerifyJWT(tokenString string) (map[string]interface{}, error) {
	// Parse and verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return es.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid JWT token")
}

// CheckRateLimit checks if a user has exceeded rate limits
func (es *EnhancedSecurity) CheckRateLimit(ctx context.Context, identifier string) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s", identifier)

	// Get current count
	count, err := es.redisClient.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	// Check if limit exceeded
	if count >= es.config.RateLimitRequests {
		return false, nil
	}

	// Increment count
	pipe := es.redisClient.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, es.config.RateLimitWindow)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to update rate limit: %w", err)
	}

	return true, nil
}

// RecordLoginAttempt records a login attempt
func (es *EnhancedSecurity) RecordLoginAttempt(ctx context.Context, identifier string, success bool) error {
	key := fmt.Sprintf("login_attempts:%s", identifier)

	if success {
		// Clear failed attempts on successful login
		err := es.redisClient.Del(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("failed to clear login attempts: %w", err)
		}
		return nil
	}

	// Increment failed attempts
	pipe := es.redisClient.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, es.config.LockoutDuration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record login attempt: %w", err)
	}

	return nil
}

// IsAccountLocked checks if an account is locked
func (es *EnhancedSecurity) IsAccountLocked(ctx context.Context, identifier string) (bool, error) {
	key := fmt.Sprintf("login_attempts:%s", identifier)

	count, err := es.redisClient.Get(ctx, key).Int()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get login attempts: %w", err)
	}

	return count >= es.config.MaxLoginAttempts, nil
}

// ValidateInput validates user input against common attacks
func (es *EnhancedSecurity) ValidateInput(input string) error {
	// Check for SQL injection patterns
	sqlPatterns := []string{
		"'|\"|;|--|/*|*/|xp_|sp_|exec|insert|select|delete|update|drop|create|alter",
		"union.*select",
		"select.*from",
		"insert.*into",
		"delete.*from",
		"update.*set",
		"drop.*table",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(strings.ToLower(input), pattern) {
			return fmt.Errorf("potential SQL injection detected")
		}
	}

	// Check for XSS patterns
	xssPatterns := []string{
		"<script|</script|javascript:|vbscript:|onload=|onerror=|onclick=|onmouseover=",
		"<iframe|</iframe|<object|</object|<embed|</embed",
		"eval\\(|alert\\(|confirm\\(|prompt\\(",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(strings.ToLower(input), pattern) {
			return fmt.Errorf("potential XSS detected")
		}
	}

	// Check length limits
	if len(input) > 1000 {
		return fmt.Errorf("input too long")
	}

	return nil
}

// GenerateCSRFToken generates a CSRF token
func (es *EnhancedSecurity) GenerateCSRFToken(ctx context.Context, sessionID string) (string, error) {
	token, err := es.GenerateSecureToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	key := fmt.Sprintf("csrf:%s", sessionID)
	err = es.redisClient.Set(ctx, key, token, es.config.SessionTimeout).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store CSRF token: %w", err)
	}

	return token, nil
}

// VerifyCSRFToken verifies a CSRF token
func (es *EnhancedSecurity) VerifyCSRFToken(ctx context.Context, sessionID, token string) bool {
	key := fmt.Sprintf("csrf:%s", sessionID)

	storedToken, err := es.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		log.Printf("Failed to get CSRF token: %v", err)
		return false
	}

	return storedToken == token
}

// EncryptData encrypts sensitive data
func (es *EnhancedSecurity) EncryptData(data string) (string, error) {
	// This is a simplified encryption
	// In production, use proper encryption like AES-256-GCM

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

// DecryptData decrypts sensitive data
func (es *EnhancedSecurity) DecryptData(encryptedData string) (string, error) {
	// This is a simplified decryption
	// In production, use proper decryption matching your encryption

	// For demo purposes, we'll just return the encrypted data
	// In a real implementation, this would properly decrypt
	return encryptedData, nil
}

// SecurityMiddleware provides security middleware
func (es *EnhancedSecurity) SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Validate request size
		if r.ContentLength > 10*1024*1024 { // 10MB
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}

		// Check for suspicious patterns
		userAgent := r.Header.Get("User-Agent")
		if userAgent != "" {
			err := es.ValidateInput(userAgent)
			if err != nil {
				log.Printf("Suspicious User-Agent: %s - %v", userAgent, err)
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware provides rate limiting middleware
func (es *EnhancedSecurity) RateLimitMiddleware(identifierFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identifier := identifierFunc(r)

			// Check rate limit
			allowed, err := es.CheckRateLimit(r.Context(), identifier)
			if err != nil {
				log.Printf("Rate limit check failed: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", es.config.RateLimitWindow.String())
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware provides authentication middleware
func (es *EnhancedSecurity) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		// Extract Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]

		// Verify JWT
		claims, err := es.VerifyJWT(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "user_claims", claims)
		ctx = context.WithValue(ctx, "user_id", claims["user_id"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value("user_id").(int64)
	return userID, ok
}

// GetUserClaimsFromContext extracts user claims from context
func GetUserClaimsFromContext(ctx context.Context) (map[string]interface{}, bool) {
	claims, ok := ctx.Value("user_claims").(map[string]interface{})
	return claims, ok
}

// AuditLog records security events
func (es *EnhancedSecurity) AuditLog(ctx context.Context, eventType, userID, description string) error {
	auditEntry := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"event_type":  eventType,
		"user_id":     userID,
		"description": description,
		"ip_address":  getClientIP(ctx),
		"user_agent":  getUserAgent(ctx),
	}

	// Store audit log
	key := fmt.Sprintf("audit:%d", time.Now().Unix())

	auditJSON, err := json.Marshal(auditEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	err = es.redisClient.Set(ctx, key, auditJSON, 90*24*time.Hour).Err() // 90 days retention
	if err != nil {
		return fmt.Errorf("failed to store audit entry: %w", err)
	}

	return nil
}

// Helper functions

func getClientIP(ctx context.Context) string {
	// This would extract IP from request context
	// Implementation depends on how you store request in context
	return "unknown"
}

func getUserAgent(ctx context.Context) string {
	// This would extract User-Agent from request context
	// Implementation depends on how you store request in context
	return "unknown"
}

// SecurityEvent represents a security event
type SecurityEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	UserID      string    `json:"user_id"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
}

// GetSecurityEvents retrieves security events
func (es *EnhancedSecurity) GetSecurityEvents(ctx context.Context, from, to time.Time) ([]SecurityEvent, error) {
	pattern := "audit:*"

	keys, err := es.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get audit keys: %w", err)
	}

	var events []SecurityEvent
	for _, key := range keys {
		auditJSON, err := es.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var event SecurityEvent
		err = json.Unmarshal([]byte(auditJSON), &event)
		if err != nil {
			continue
		}

		// Filter by time range
		if event.Timestamp.After(from) && event.Timestamp.Before(to) {
			events = append(events, event)
		}
	}

	return events, nil
}

// ValidatePasswordStrength validates password strength
func (es *EnhancedSecurity) ValidatePasswordStrength(password string) error {
	if len(password) < es.config.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters", es.config.PasswordMinLength)
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		JWTSecret:         "your-super-secret-jwt-key-change-in-production",
		JWTExpiration:     24 * time.Hour,
		MaxLoginAttempts:  5,
		LockoutDuration:   15 * time.Minute,
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
		PasswordMinLength: 8,
		SessionTimeout:    24 * time.Hour,
	}
}
