package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics - система метрик Prometheus
type PrometheusMetrics struct {
	registry *prometheus.Registry
	server   *http.Server
	port     int

	// Business metrics
	bkcTotalSupply       prometheus.Gauge
	bkcCirculatingSupply prometheus.Gauge
	activeUsers          prometheus.Gauge
	totalTransactions    prometheus.Counter
	dailyActiveUsers     prometheus.Gauge
	totalNFTs            prometheus.Gauge
	activeAuctions       prometheus.Gauge

	// Performance metrics
	requestDuration   *prometheus.HistogramVec
	requestCount      *prometheus.CounterVec
	errorCount        *prometheus.CounterVec
	responseSize      *prometheus.HistogramVec
	activeConnections prometheus.Gauge

	// Database metrics
	dbConnections   prometheus.Gauge
	dbQueryDuration *prometheus.HistogramVec
	dbQueryCount    *prometheus.CounterVec
	dbErrors        *prometheus.CounterVec

	// Game metrics
	gameSessions prometheus.Counter
	gameBets     prometheus.Counter
	gameWinnings prometheus.Counter
	crashGames   prometheus.Counter

	// Payment metrics
	paymentAttempts *prometheus.CounterVec
	paymentSuccess  *prometheus.CounterVec
	paymentAmount   *prometheus.HistogramVec
	paymentDuration *prometheus.HistogramVec

	// NFT metrics
	nftMinted      prometheus.Counter
	nftTransferred prometheus.Counter
	nftUpgrades    prometheus.Counter

	// System metrics
	memoryUsage    prometheus.Gauge
	cpuUsage       prometheus.Gauge
	goroutineCount prometheus.Gauge
	gcDuration     prometheus.Histogram

	mutex sync.RWMutex
}

// NewPrometheusMetrics - создание системы метрик
func NewPrometheusMetrics(port int) *PrometheusMetrics {
	pm := &PrometheusMetrics{
		registry: prometheus.NewRegistry(),
		port:     port,
	}

	pm.initializeMetrics()
	pm.registerMetrics()

	return pm
}

// initializeMetrics - инициализация метрик
func (pm *PrometheusMetrics) initializeMetrics() {
	// Business metrics
	pm.bkcTotalSupply = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_total_supply",
		Help: "Total BKC token supply",
	})

	pm.bkcCirculatingSupply = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_circulating_supply",
		Help: "Circulating BKC token supply",
	})

	pm.activeUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_active_users",
		Help: "Number of active users",
	})

	pm.totalTransactions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bkc_total_transactions",
		Help: "Total number of transactions",
	})

	pm.dailyActiveUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_daily_active_users",
		Help: "Number of daily active users",
	})

	pm.totalNFTs = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_total_nfts",
		Help: "Total number of NFTs",
	})

	pm.activeAuctions = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_active_auctions",
		Help: "Number of active NFT auctions",
	})

	// Performance metrics
	pm.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bkc_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	pm.requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bkc_requests_total",
			Help: "Total number of requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	pm.errorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bkc_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "endpoint"},
	)

	pm.responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bkc_response_size_bytes",
			Help:    "Response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"endpoint"},
	)

	pm.activeConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_active_connections",
		Help: "Number of active connections",
	})

	// Database metrics
	pm.dbConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_db_connections",
		Help: "Number of database connections",
	})

	pm.dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bkc_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	pm.dbQueryCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bkc_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	pm.dbErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bkc_db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "table"},
	)

	// Game metrics
	pm.gameSessions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bkc_game_sessions_total",
		Help: "Total number of game sessions",
	})

	pm.gameBets = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bkc_game_bets_total",
		Help: "Total number of game bets",
	})

	pm.gameWinnings = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bkc_game_winnings_total",
		Help: "Total game winnings",
	})

	pm.crashGames = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bkc_crash_games_total",
		Help: "Total number of crash games",
	})

	// Payment metrics
	pm.paymentAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bkc_payment_attempts_total",
			Help: "Total number of payment attempts",
		},
		[]string{"chain", "currency", "status"},
	)

	pm.paymentSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bkc_payment_success_total",
			Help: "Total number of successful payments",
		},
		[]string{"chain", "currency"},
	)

	pm.paymentAmount = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bkc_payment_amount_bkc",
			Help:    "Payment amount in BKC",
			Buckets: []float64{100, 1000, 5000, 10000, 50000, 100000, 500000},
		},
		[]string{"chain", "currency"},
	)

	pm.paymentDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bkc_payment_duration_seconds",
			Help:    "Payment processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"chain"},
	)

	// NFT metrics
	pm.nftMinted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bkc_nft_minted_total",
		Help: "Total number of NFTs minted",
	})

	pm.nftTransferred = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bkc_nft_transferred_total",
		Help: "Total number of NFT transfers",
	})

	pm.nftUpgrades = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bkc_nft_upgrades_total",
		Help: "Total number of NFT upgrades",
	})

	// System metrics
	pm.memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_memory_usage_bytes",
		Help: "Memory usage in bytes",
	})

	pm.cpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_cpu_usage_percent",
		Help: "CPU usage percentage",
	})

	pm.goroutineCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bkc_goroutines",
		Help: "Number of goroutines",
	})

	pm.gcDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "bkc_gc_duration_seconds",
		Help:    "Garbage collection duration in seconds",
		Buckets: prometheus.DefBuckets,
	})
}

// registerMetrics - регистрация метрик
func (pm *PrometheusMetrics) registerMetrics() {
	// Business metrics
	pm.registry.MustRegister(pm.bkcTotalSupply)
	pm.registry.MustRegister(pm.bkcCirculatingSupply)
	pm.registry.MustRegister(pm.activeUsers)
	pm.registry.MustRegister(pm.totalTransactions)
	pm.registry.MustRegister(pm.dailyActiveUsers)
	pm.registry.MustRegister(pm.totalNFTs)
	pm.registry.MustRegister(pm.activeAuctions)

	// Performance metrics
	pm.registry.MustRegister(pm.requestDuration)
	pm.registry.MustRegister(pm.requestCount)
	pm.registry.MustRegister(pm.errorCount)
	pm.registry.MustRegister(pm.responseSize)
	pm.registry.MustRegister(pm.activeConnections)

	// Database metrics
	pm.registry.MustRegister(pm.dbConnections)
	pm.registry.MustRegister(pm.dbQueryDuration)
	pm.registry.MustRegister(pm.dbQueryCount)
	pm.registry.MustRegister(pm.dbErrors)

	// Game metrics
	pm.registry.MustRegister(pm.gameSessions)
	pm.registry.MustRegister(pm.gameBets)
	pm.registry.MustRegister(pm.gameWinnings)
	pm.registry.MustRegister(pm.crashGames)

	// Payment metrics
	pm.registry.MustRegister(pm.paymentAttempts)
	pm.registry.MustRegister(pm.paymentSuccess)
	pm.registry.MustRegister(pm.paymentAmount)
	pm.registry.MustRegister(pm.paymentDuration)

	// NFT metrics
	pm.registry.MustRegister(pm.nftMinted)
	pm.registry.MustRegister(pm.nftTransferred)
	pm.registry.MustRegister(pm.nftUpgrades)

	// System metrics
	pm.registry.MustRegister(pm.memoryUsage)
	pm.registry.MustRegister(pm.cpuUsage)
	pm.registry.MustRegister(pm.goroutineCount)
	pm.registry.MustRegister(pm.gcDuration)

	// Default Go metrics
	pm.registry.MustRegister(prometheus.NewGoCollector())
	pm.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
}

// StartServer - запуск сервера метрик
func (pm *PrometheusMetrics) StartServer() error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(pm.registry, promhttp.HandlerOpts{}))

	pm.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", pm.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Prometheus metrics server starting on port %d", pm.port)
	return pm.server.ListenAndServe()
}

// Shutdown - остановка сервера
func (pm *PrometheusMetrics) Shutdown(ctx context.Context) error {
	if pm.server != nil {
		return pm.server.Shutdown(ctx)
	}
	return nil
}

// Business metrics update methods

func (pm *PrometheusMetrics) UpdateBKCSupply(total, circulating float64) {
	pm.bkcTotalSupply.Set(total)
	pm.bkcCirculatingSupply.Set(circulating)
}

func (pm *PrometheusMetrics) UpdateActiveUsers(count float64) {
	pm.activeUsers.Set(count)
}

func (pm *PrometheusMetrics) UpdateDailyActiveUsers(count float64) {
	pm.dailyActiveUsers.Set(count)
}

func (pm *PrometheusMetrics) UpdateTotalNFTs(count float64) {
	pm.totalNFTs.Set(count)
}

func (pm *PrometheusMetrics) UpdateActiveAuctions(count float64) {
	pm.activeAuctions.Set(count)
}

func (pm *PrometheusMetrics) IncrementTransactions() {
	pm.totalTransactions.Inc()
}

// Performance metrics update methods

func (pm *PrometheusMetrics) RecordRequest(method, endpoint, status string, duration time.Duration, size int64) {
	pm.requestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
	pm.requestCount.WithLabelValues(method, endpoint, status).Inc()
	pm.responseSize.WithLabelValues(endpoint).Observe(float64(size))
}

func (pm *PrometheusMetrics) RecordError(errorType, endpoint string) {
	pm.errorCount.WithLabelValues(errorType, endpoint).Inc()
}

func (pm *PrometheusMetrics) UpdateActiveConnections(count float64) {
	pm.activeConnections.Set(count)
}

// Database metrics update methods

func (pm *PrometheusMetrics) UpdateDBConnections(count float64) {
	pm.dbConnections.Set(count)
}

func (pm *PrometheusMetrics) RecordDBQuery(operation, table string, duration time.Duration) {
	pm.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	pm.dbQueryCount.WithLabelValues(operation, table).Inc()
}

func (pm *PrometheusMetrics) RecordDBError(operation, table string) {
	pm.dbErrors.WithLabelValues(operation, table).Inc()
}

// Game metrics update methods

func (pm *PrometheusMetrics) IncrementGameSessions() {
	pm.gameSessions.Inc()
}

func (pm *PrometheusMetrics) IncrementGameBets() {
	pm.gameBets.Inc()
}

func (pm *PrometheusMetrics) AddGameWinnings(amount float64) {
	pm.gameWinnings.Add(amount)
}

func (pm *PrometheusMetrics) IncrementCrashGames() {
	pm.crashGames.Inc()
}

// Payment metrics update methods

func (pm *PrometheusMetrics) RecordPaymentAttempt(chain, currency, status string) {
	pm.paymentAttempts.WithLabelValues(chain, currency, status).Inc()
}

func (pm *PrometheusMetrics) RecordPaymentSuccess(chain, currency string) {
	pm.paymentSuccess.WithLabelValues(chain, currency).Inc()
}

func (pm *PrometheusMetrics) RecordPaymentAmount(chain, currency string, amount float64) {
	pm.paymentAmount.WithLabelValues(chain, currency).Observe(amount)
}

func (pm *PrometheusMetrics) RecordPaymentDuration(chain string, duration time.Duration) {
	pm.paymentDuration.WithLabelValues(chain).Observe(duration.Seconds())
}

// NFT metrics update methods

func (pm *PrometheusMetrics) IncrementNFTMinted() {
	pm.nftMinted.Inc()
}

func (pm *PrometheusMetrics) IncrementNFTTransferred() {
	pm.nftTransferred.Inc()
}

func (pm *PrometheusMetrics) IncrementNFTUpgrades() {
	pm.nftUpgrades.Inc()
}

// System metrics update methods

func (pm *PrometheusMetrics) UpdateMemoryUsage(bytes float64) {
	pm.memoryUsage.Set(bytes)
}

func (pm *PrometheusMetrics) UpdateCPUUsage(percent float64) {
	pm.cpuUsage.Set(percent)
}

func (pm *PrometheusMetrics) UpdateGoroutineCount(count float64) {
	pm.goroutineCount.Set(count)
}

func (pm *PrometheusMetrics) RecordGCDuration(duration time.Duration) {
	pm.gcDuration.Observe(duration.Seconds())
}

// Custom metrics collection

func (pm *PrometheusMetrics) CollectSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pm.UpdateMemoryUsage(float64(m.Alloc))
	pm.UpdateGoroutineCount(float64(runtime.NumGoroutine()))
}

// MetricsMiddleware - middleware для HTTP метрик
func (pm *PrometheusMetrics) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Capture response size
		wrapped := &responseWriter{ResponseWriter: w, status: 200}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		status := fmt.Sprintf("%d", wrapped.status)

		pm.RecordRequest(r.Method, r.URL.Path, status, duration, wrapped.written)

		// Record errors for 4xx and 5xx status codes
		if wrapped.status >= 400 {
			errorType := "client_error"
			if wrapped.status >= 500 {
				errorType = "server_error"
			}
			pm.RecordError(errorType, r.URL.Path)
		}
	})
}

// responseWriter - обертка для захвата статуса и размера ответа
type responseWriter struct {
	http.ResponseWriter
	status  int
	written int64
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// GetMetricsSummary - получение сводки метрик
func (pm *PrometheusMetrics) GetMetricsSummary() (map[string]interface{}, error) {
	// Gather metrics
	metricFamilies, err := pm.registry.Gather()
	if err != nil {
		return nil, err
	}

	summary := make(map[string]interface{})

	for _, mf := range metricFamilies {
		for _, m := range mf.Metric {
			name := mf.GetName()

			switch {
			case name == "bkc_active_users":
				summary["active_users"] = m.GetGauge().GetValue()
			case name == "bkc_total_transactions":
				summary["total_transactions"] = m.GetCounter().GetValue()
			case name == "bkc_total_nfts":
				summary["total_nfts"] = m.GetGauge().GetValue()
			case name == "bkc_active_auctions":
				summary["active_auctions"] = m.GetGauge().GetValue()
			case name == "bkc_requests_total":
				summary["total_requests"] = m.GetCounter().GetValue()
			case name == "bkc_memory_usage_bytes":
				summary["memory_usage_mb"] = m.GetGauge().GetValue() / 1024 / 1024
			case name == "bkc_goroutines":
				summary["goroutines"] = m.GetGauge().GetValue()
			}
		}
	}

	summary["last_updated"] = time.Now()

	return summary, nil
}

// ExportMetrics - экспорт метрик в формате JSON
func (pm *PrometheusMetrics) ExportMetrics() (string, error) {
	summary, err := pm.GetMetricsSummary()
	if err != nil {
		return "", err
	}

	bytes, err := json.Marshal(summary)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
