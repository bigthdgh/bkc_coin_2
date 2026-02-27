package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"bkc_coin_v2/internal/config"
	"bkc_coin_v2/internal/database"
	"bkc_coin_v2/internal/games"
	"bkc_coin_v2/internal/monitoring"
	"bkc_coin_v2/internal/payments"
	"bkc_coin_v2/internal/security"
	"bkc_coin_v2/internal/i18n"
	"bkc_coin_v2/internal/loadbalancer"
)

func main() {
	// Загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Инициализация конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация базы данных
	db, err := database.NewUnifiedDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Инициализация игровых систем
	gameManager := games.NewUnifiedGameManager(db, cfg.Games)

	// Инициализация платежной системы
	paymentManager := payments.NewMultiChainPaymentManager(db, cfg.Payments)

	// Инициализация Helius
	heliusConfig := payments.HeliusConfig{
		APIKey:     "f983dbf9-7518-4337-985d-d8ea68b16e64",
		AdminWallet: os.Getenv("SOLANA_ADMIN_WALLET"),
	}
	
	helius, err := payments.NewHeliusIntegration(heliusConfig, paymentManager)
	if err != nil {
		log.Printf("Warning: Failed to initialize Helius: %v", err)
	} else {
		// Запускаем WebSocket слушатель
		go func() {
			if err := helius.StartWebSocketListener(); err != nil {
				log.Printf("Helius WebSocket error: %v", err)
			}
		}()
	}

	// Инициализация системы мониторинга
	prometheusMetrics := monitoring.NewPrometheusMetrics(cfg.Metrics.Port)
	go func() {
		if err := prometheusMetrics.StartServer(); err != nil {
			log.Printf("Prometheus server error: %v", err)
		}
	}()

	// Инициализация интернационализации
	i18nManager := i18n.NewI18nManager()
	i18nManager.LoadTranslations()
	i18nManager.LoadCurrencyRates()
	i18nManager.LoadRegionalSettings()

	// Инициализация балансировщика нагрузки
	loadBalancer := loadbalancer.NewLoadBalancer(cfg.LoadBalancer)
	
	// Добавляем серверы в балансировщик
	for _, server := range cfg.LoadBalancer.Servers {
		loadBalancer.AddServer(server.URL, server.Weight)
	}

	// Настройка Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// DDoS защита
	ddosProtection := security.NewDDoSProtection(cfg.Security)
	router.Use(ddosProtection.Middleware())

	// Prometheus метрики
	router.Use(prometheusMetrics.MetricsMiddleware())

	// API роуты
	setupAPIRoutes(router, db, gameManager, paymentManager, helius, i18nManager, prometheusMetrics)

	// Запуск сервера
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("🚀 BKC Coin server started on port %d", cfg.Server.Port)
	log.Printf("📊 Prometheus metrics available on port %d", cfg.Metrics.Port)
	log.Printf("🌐 API documentation: http://localhost:%d/docs", cfg.Server.Port)

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🔄 Shutting down server...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Останавливаем Helius
	if helius != nil {
		helius.Shutdown()
	}

	// Останавливаем Prometheus
	prometheusMetrics.Shutdown(ctx)

	log.Println("✅ Server shutdown completed")
}

func setupAPIRoutes(
	router *gin.Engine,
	db *database.UnifiedDB,
	gameManager *games.UnifiedGameManager,
	paymentManager *payments.MultiChainPaymentManager,
	helius *payments.HeliusIntegration,
	i18nManager *i18n.I18nManager,
	prometheusMetrics *monitoring.PrometheusMetrics,
) {
	// API v1
	v1 := router.Group("/api/v1")

	// Пользовательские роуты
	setupUserRoutes(v1, db, i18nManager)

	// Игровые роуты
	setupGameRoutes(v1, gameManager)

	// Платежные роуты
	setupPaymentRoutes(v1, paymentManager, helius)

	// Маркетплейс роуты
	setupMarketplaceRoutes(v1, db)

	// Мониторинг роуты
	setupMonitoringRoutes(v1, prometheusMetrics)

	// I18n роуты
	setupI18nRoutes(v1, i18nManager)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "2.0.0",
		})
	})

	// Статические файлы
	router.Static("/static", "./webapp/static")
	router.StaticFile("/", "./webapp/payment.html")
	router.StaticFile("/payment", "./webapp/payment.html")
}

func setupUserRoutes(router *gin.RouterGroup, db *database.UnifiedDB, i18nManager *i18n.I18nManager) {
	users := router.Group("/users")
	{
		users.GET("/:id", getUserHandler(db))
		users.POST("/", createUserHandler(db))
		users.PUT("/:id", updateUserHandler(db))
		users.GET("/:id/balance", getUserBalanceHandler(db))
		users.GET("/:id/stats", getUserStatsHandler(db))
	}
}

func setupGameRoutes(router *gin.RouterGroup, gameManager *games.UnifiedGameManager) {
	games := router.Group("/games")
	{
		games.GET("/crash", getCrashGameHandler(gameManager))
		games.POST("/crash/bet", placeCrashBetHandler(gameManager))
		games.GET("/crash/history", getCrashHistoryHandler(gameManager))
		games.GET("/exchange", getExchangeRateHandler(gameManager))
	}
}

func setupPaymentRoutes(router *gin.RouterGroup, paymentManager *payments.MultiChainPaymentManager, helius *payments.HeliusIntegration) {
	payments := router.Group("/payments")
	{
		payments.POST("/create", paymentManager.CreatePaymentOrder)
		payments.GET("/status/:id", paymentManager.GetPaymentStatus)
		payments.GET("/history", paymentManager.GetPaymentHistory)
		payments.POST("/cancel/:id", paymentManager.CancelPaymentOrder)
		payments.GET("/chains", paymentManager.GetSupportedChains)
		payments.GET("/commissions", paymentManager.GetCommissionInfo)
		payments.POST("/estimate", paymentManager.EstimatePayment)
	}

	// Вебхуки
	if helius != nil {
		webhookHandler := payments.NewHeliusWebhookHandler(paymentManager, helius)
		webhookHandler.SetupRoutes(router)
	}
}

func setupMarketplaceRoutes(router *gin.RouterGroup, db *database.UnifiedDB) {
	marketplace := router.Group("/marketplace")
	{
		marketplace.GET("/nfts", getNFTListingsHandler(db))
		marketplace.POST("/nfts", createNFTListingHandler(db))
		marketplace.GET("/auctions", getAuctionsHandler(db))
		marketplace.POST("/auctions", createAuctionHandler(db))
		marketplace.POST("/auctions/:id/bid", placeBidHandler(db))
	}
}

func setupMonitoringRoutes(router *gin.RouterGroup, prometheusMetrics *monitoring.PrometheusMetrics) {
	monitoring := router.Group("/monitoring")
	{
		monitoring.GET("/metrics", func(c *gin.Context) {
			metrics, _ := prometheusMetrics.GetMetricsSummary()
			c.JSON(http.StatusOK, metrics)
		})
		monitoring.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
				"timestamp": time.Now(),
			})
		})
	}
}

func setupI18nRoutes(router *gin.RouterGroup, i18nManager *i18n.I18nManager) {
	i18n := router.Group("/i18n")
	{
		i18n.GET("/translations/:lang", func(c *gin.Context) {
			lang := c.Param("lang")
			translations := i18nManager.GetTranslations(lang)
			c.JSON(http.StatusOK, translations)
		})
		i18n.GET("/currencies", func(c *gin.Context) {
			currencies := i18nManager.GetSupportedCurrencies()
			c.JSON(http.StatusOK, currencies)
		})
		i18n.GET("/regions", func(c *gin.Context) {
			regions := i18nManager.GetSupportedRegions()
			c.JSON(http.StatusOK, regions)
		})
	}
}

// Временные обработчики (заглушки)
func getUserHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "User endpoint"})
	}
}

func createUserHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Create user endpoint"})
	}
}

func updateUserHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Update user endpoint"})
	}
}

func getUserBalanceHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "User balance endpoint"})
	}
}

func getUserStatsHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "User stats endpoint"})
	}
}

func getCrashGameHandler(gameManager *games.UnifiedGameManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Crash game endpoint"})
	}
}

func placeCrashBetHandler(gameManager *games.UnifiedGameManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Place crash bet endpoint"})
	}
}

func getCrashHistoryHandler(gameManager *games.UnifiedGameManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Crash history endpoint"})
	}
}

func getExchangeRateHandler(gameManager *games.UnifiedGameManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Exchange rate endpoint"})
	}
}

func getNFTListingsHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "NFT listings endpoint"})
	}
}

func createNFTListingHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Create NFT listing endpoint"})
	}
}

func getAuctionsHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Get auctions endpoint"})
	}
}

func createAuctionHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Create auction endpoint"})
	}
}

func placeBidHandler(db *database.UnifiedDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Place bid endpoint"})
	}
}
