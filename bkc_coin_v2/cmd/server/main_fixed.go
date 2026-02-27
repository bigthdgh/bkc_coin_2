package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	
	"bkc_coin_v2/internal/economy"
	"bkc_coin_v2/internal/security"
	"bkc_coin_v2/internal/monitoring"
	"bkc_coin_v2/internal/achievements"
	"bkc_coin_v2/internal/notifications"
	"bkc_coin_v2/internal/games"
	"bkc_coin_v2/internal/ton"
	"bkc_coin_v2/internal/monetization"
)

// Server представляет основной сервер
type Server struct {
	router           *gin.Engine
	redisClient      *redis.Client
	db               *sqlx.DB
	economySystem    *economy.BalancedEconomySystem
	securitySystem   *security.EnhancedSecurity
	monitoringSystem *monitoring.AnalyticsSystem
	achievementsSystem *achievements.AchievementsSystem
	notificationsSystem *notifications.NotificationsSystem
	gamesSystem      *games.EnhancedGameMechanics
	tonClient        *ton.TonClient
	adsSystem        *monetization.MinimalSecureAds
}

// Config представляет конфигурацию сервера
type Config struct {
	ServerPort      string `json:"server_port"`
	RedisAddr       string `json:"redis_addr"`
	DatabaseURL     string `json:"database_url"`
	JWTSecret       string `json:"jwt_secret"`
	AdsgramSecret   string `json:"adsgram_secret"`
}

func main() {
	// Загрузка конфигурации
	config := &Config{
		ServerPort:    ":8080",
		RedisAddr:     "localhost:6379",
		DatabaseURL:   "postgres://user:password@localhost/bkc_coin?sslmode=disable",
		JWTSecret:     "your-super-secret-jwt-key",
		AdsgramSecret: "your-adsgram-webhook-secret",
	}

	// Создание сервера
	server, err := NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Запуск сервера
	go func() {
		log.Printf("Starting BKC Coin server on port %s", config.ServerPort)
		if err := server.router.Run(config.ServerPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Запуск WebSocket сервера для рекламы
	go func() {
		server.adsSystem.StartWebSocketServer(8081)
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// NewServer создает новый сервер
func NewServer(config *Config) (*Server, error) {
	// Инициализация Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "",
		DB:       0,
	})

	// Проверка соединения с Redis
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Инициализация базы данных
	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Инициализация систем
	economySystem := economy.NewBalancedEconomySystem(redisClient)
	securitySystem := security.NewEnhancedSecurity(config.JWTSecret)
	monitoringSystem := monitoring.NewAnalyticsSystem(redisClient)
	achievementsSystem := achievements.NewAchievementsSystem(redisClient)
	notificationsSystem := notifications.NewNotificationsSystem(redisClient)
	gamesSystem := games.NewEnhancedGameMechanics(redisClient)
	tonClient := ton.NewTonClient()
	adsSystem := monetization.NewMinimalSecureAds(redisClient, config.AdsgramSecret)

	// Создание сервера
	server := &Server{
		router:             gin.Default(),
		redisClient:        redisClient,
		db:                 db,
		economySystem:      economySystem,
		securitySystem:     securitySystem,
		monitoringSystem:   monitoringSystem,
		achievementsSystem: achievementsSystem,
		notificationsSystem: notificationsSystem,
		gamesSystem:        gamesSystem,
		tonClient:          tonClient,
		adsSystem:          adsSystem,
	}

	// Настройка маршрутов
	server.setupRoutes()

	return server, nil
}

// setupRoutes настраивает маршруты API
func (s *Server) setupRoutes() {
	// Middleware
	s.router.Use(s.corsMiddleware())
	s.router.Use(s.loggingMiddleware())

	// API v1
	v1 := s.router.Group("/api/v1")
	{
		// Экономика
		economy := v1.Group("/economy")
		{
			economy.GET("/user/:user_id", s.getUserEconomy)
			economy.POST("/tap", s.processTap)
			economy.POST("/referral", s.processReferral)
			economy.GET("/daily-bonus/:user_id", s.getDailyBonus)
			economy.GET("/leaderboard", s.getLeaderboard)
		}

		// Безопасность
		auth := v1.Group("/auth")
		{
			auth.POST("/login", s.login)
			auth.POST("/register", s.register)
			auth.POST("/refresh", s.refreshToken)
		}

		// Мониторинг
		monitoring := v1.Group("/monitoring")
		{
			monitoring.GET("/stats", s.getSystemStats)
			monitoring.GET("/health", s.healthCheck)
		}

		// Достижения
		achievements := v1.Group("/achievements")
		{
			achievements.GET("/:user_id", s.getUserAchievements)
			achievements.POST("/unlock", s.unlockAchievement)
		}

		// Уведомления
		notifications := v1.Group("/notifications")
		{
			notifications.GET("/:user_id", s.getUserNotifications)
			notifications.POST("/send", s.sendNotification)
		}

		// Игры
		games := v1.Group("/games")
		{
			games.POST("/tournament", s.createTournament)
			games.GET("/tournament/:id", s.getTournament)
			games.POST("/tournament/:id/join", s.joinTournament)
		}

		// TON
		ton := v1.Group("/ton")
		{
			ton.GET("/balance/:user_id", s.getTonBalance)
			ton.POST("/transfer", s.makeTonTransfer)
		}

		// Реклама (минимальная система)
		ads := v1.Group("/ads")
		{
			ads.GET("/available", s.getAvailableAds)
			ads.POST("/session/start", s.startAdSession)
			ads.GET("/stats", s.getAdStats)
			ads.GET("/next-time/:type", s.getNextAdTime)
		}
	}

	// Webhook для Adsgram
	s.router.POST("/webhook/adsgram", s.handleAdsgramWebhook)

	// WebSocket
	s.router.GET("/ws", s.handleWebSocket)

	// Статические файлы
	s.router.Static("/static", "./webapp")
	s.router.StaticFile("/", "./webapp/index_enhanced.html")
}

// CORS middleware
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logging middleware
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("API | %s | %s | %d | %v | %s",
			method,
			path,
			statusCode,
			latency,
			clientIP,
		)
	}
}

// Экономические эндпоинты
func (s *Server) getUserEconomy(c *gin.Context) {
	userID := parseInt64Param(c.Param("user_id"))
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	economy, err := s.economySystem.GetUserEconomy(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": economy})
}

func (s *Server) processTap(c *gin.Context) {
	var request struct {
		UserID int64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := s.economySystem.ProcessTap(c.Request.Context(), request.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (s *Server) processReferral(c *gin.Context) {
	var request struct {
		ReferrerID int64 `json:"referrer_id" binding:"required"`
		ReferredID int64 `json:"referred_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.economySystem.ProcessReferral(c.Request.Context(), request.ReferrerID, request.ReferredID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *Server) getDailyBonus(c *gin.Context) {
	userID := parseInt64Param(c.Param("user_id"))
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	bonus, err := s.economySystem.GetDailyBonus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": bonus})
}

func (s *Server) getLeaderboard(c *gin.Context) {
	limit := parseInt64Query(c, "limit", 50)

	leaderboard, err := s.economySystem.GetLeaderboard(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": leaderboard})
}

// Рекламные эндпоинты
func (s *Server) getAvailableAds(c *gin.Context) {
	userID := s.getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ads, err := s.adsSystem.GetAvailableAds(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": ads})
}

func (s *Server) startAdSession(c *gin.Context) {
	var request struct {
		AdType string `json:"ad_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := s.getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	config, err := s.adsSystem.StartAdSession(c.Request.Context(), userID, request.AdType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": config})
}

func (s *Server) getAdStats(c *gin.Context) {
	userID := s.getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	stats, err := s.adsSystem.GetAdStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (s *Server) getNextAdTime(c *gin.Context) {
	userID := s.getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	adType := c.Param("type")
	nextTime, err := s.adsSystem.GetNextAdTime(c.Request.Context(), userID, adType)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "available": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"available":   false,
		"next_time":   nextTime.Unix(),
		"next_time_formatted": nextTime.Format("2006-01-02 15:04:05"),
		"cooldown_minutes": int(time.Until(nextTime).Minutes()),
	})
}

func (s *Server) handleAdsgramWebhook(c *gin.Context) {
	s.adsSystem.HandleAdsgramWebhook(c.Writer, c.Request)
}

func (s *Server) handleWebSocket(c *gin.Context) {
	s.adsSystem.HandleWebSocket(c.Writer, c.Request)
}

// Другие эндпоинты (заглушки)
func (s *Server) login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Login endpoint"})
}

func (s *Server) register(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Register endpoint"})
}

func (s *Server) refreshToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Refresh token endpoint"})
}

func (s *Server) getSystemStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "System stats endpoint"})
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "status": "healthy"})
}

func (s *Server) getUserAchievements(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User achievements endpoint"})
}

func (s *Server) unlockAchievement(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Unlock achievement endpoint"})
}

func (s *Server) getUserNotifications(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User notifications endpoint"})
}

func (s *Server) sendNotification(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Send notification endpoint"})
}

func (s *Server) createTournament(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Create tournament endpoint"})
}

func (s *Server) getTournament(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Get tournament endpoint"})
}

func (s *Server) joinTournament(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Join tournament endpoint"})
}

func (s *Server) getTonBalance(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "TON balance endpoint"})
}

func (s *Server) makeTonTransfer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "TON transfer endpoint"})
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Close database connection
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}

	// Close Redis connection
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		}
	}

	return nil
}

// Вспомогательные функции
func parseInt64Param(param string) int64 {
	var result int64
	fmt.Sscanf(param, "%d", &result)
	return result
}

func parseInt64Query(c *gin.Context, key string, defaultValue int64) int64 {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	
	var result int64
	fmt.Sscanf(value, "%d", &result)
	return result
}

func (s *Server) getUserIDFromContext(c *gin.Context) int64 {
	// В реальном приложении здесь будет извлечение user ID из JWT токена
	// Пока возвращаем тестовый ID
	return 12345
}
