package loadbalancer

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// LoadBalancer - балансировщик нагрузки
type LoadBalancer struct {
	servers       []*Server
	currentIndex  int
	mutex         sync.RWMutex
	strategy      Strategy
	healthChecker *HealthChecker
	config        Config
	metrics       *LoadBalancerMetrics
}

// Server - сервер в пуле
type Server struct {
	URL                *url.URL
	Weight             int
	MaxConnections     int
	CurrentConnections int
	IsHealthy          bool
	ResponseTime       time.Duration
	ErrorCount         int64
	LastChecked        time.Time
	mutex              sync.RWMutex
}

// Strategy - стратегия балансировки
type Strategy int

const (
	RoundRobin Strategy = iota
	WeightedRoundRobin
	LeastConnections
	ResponseTime
	Random
)

// Config - конфигурация балансировщика
type Config struct {
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
	MaxRetries          int           `json:"max_retries"`
	RetryDelay          time.Duration `json:"retry_delay"`
	Strategy            string        `json:"strategy"`
	EnableMetrics       bool          `json:"enable_metrics"`
}

// LoadBalancerMetrics - метрики балансировщика
type LoadBalancerMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	HealthyServers      int           `json:"healthy_servers"`
	TotalServers        int           `json:"total_servers"`
	LastUpdated         time.Time     `json:"last_updated"`
	mutex               sync.RWMutex
}

// HealthChecker - проверка здоровья серверов
type HealthChecker struct {
	interval time.Duration
	timeout  time.Duration
	servers  []*Server
	stopCh   chan struct{}
}

// NewLoadBalancer - создание балансировщика нагрузки
func NewLoadBalancer(config Config) *LoadBalancer {
	lb := &LoadBalancer{
		servers:      make([]*Server, 0),
		currentIndex: 0,
		strategy:     parseStrategy(config.Strategy),
		config:       config,
		metrics:      &LoadBalancerMetrics{},
	}

	// Создаем health checker
	lb.healthChecker = &HealthChecker{
		interval: config.HealthCheckInterval,
		timeout:  config.HealthCheckTimeout,
		servers:  lb.servers,
		stopCh:   make(chan struct{}),
	}

	// Запускаем проверку здоровья
	go lb.healthChecker.start()

	return lb
}

// AddServer - добавление сервера в пул
func (lb *LoadBalancer) AddServer(serverURL string, weight int) error {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	server := &Server{
		URL:            parsedURL,
		Weight:         weight,
		MaxConnections: 1000, // По умолчанию
		IsHealthy:      true,
		LastChecked:    time.Now(),
	}

	lb.mutex.Lock()
	lb.servers = append(lb.servers, server)
	lb.mutex.Unlock()

	// Обновляем health checker
	lb.healthChecker.updateServers(lb.servers)

	log.Printf("Server added to load balancer: %s (weight: %d)", serverURL, weight)
	return nil
}

// RemoveServer - удаление сервера из пула
func (lb *LoadBalancer) RemoveServer(serverURL string) error {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	for i, server := range lb.servers {
		if server.URL.String() == parsedURL.String() {
			lb.servers = append(lb.servers[:i], lb.servers[i+1:]...)
			lb.healthChecker.updateServers(lb.servers)
			log.Printf("Server removed from load balancer: %s", serverURL)
			return nil
		}
	}

	return fmt.Errorf("server not found: %s", serverURL)
}

// GetServer - получение сервера по стратегии
func (lb *LoadBalancer) GetServer() (*Server, error) {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	healthyServers := lb.getHealthyServers()
	if len(healthyServers) == 0 {
		return nil, fmt.Errorf("no healthy servers available")
	}

	switch lb.strategy {
	case RoundRobin:
		return lb.roundRobin(healthyServers), nil
	case WeightedRoundRobin:
		return lb.weightedRoundRobin(healthyServers), nil
	case LeastConnections:
		return lb.leastConnections(healthyServers), nil
	case ResponseTime:
		return lb.responseTimeBased(healthyServers), nil
	case Random:
		return lb.random(healthyServers), nil
	default:
		return lb.roundRobin(healthyServers), nil
	}
}

// getHealthyServers - получение списка здоровых серверов
func (lb *LoadBalancer) getHealthyServers() []*Server {
	healthy := make([]*Server, 0)
	for _, server := range lb.servers {
		if server.IsHealthy {
			healthy = append(healthy, server)
		}
	}
	return healthy
}

// roundRobin - Round Robin стратегия
func (lb *LoadBalancer) roundRobin(servers []*Server) *Server {
	server := servers[lb.currentIndex%len(servers)]
	lb.currentIndex++
	return server
}

// weightedRoundRobin - Weighted Round Robin стратегия
func (lb *LoadBalancer) weightedRoundRobin(servers []*Server) *Server {
	totalWeight := 0
	for _, server := range servers {
		totalWeight += server.Weight
	}

	if totalWeight == 0 {
		return lb.roundRobin(servers)
	}

	target := rand.Intn(totalWeight)
	currentWeight := 0

	for _, server := range servers {
		currentWeight += server.Weight
		if currentWeight > target {
			return server
		}
	}

	return servers[0]
}

// leastConnections - Least Connections стратегия
func (lb *LoadBalancer) leastConnections(servers []*Server) *Server {
	var selectedServer *Server
	minConnections := int(^uint(0) >> 1) // Max int

	for _, server := range servers {
		server.mutex.RLock()
		connections := server.CurrentConnections
		server.mutex.RUnlock()

		if connections < minConnections {
			minConnections = connections
			selectedServer = server
		}
	}

	return selectedServer
}

// responseTimeBased - стратегия на основе времени ответа
func (lb *LoadBalancer) responseTimeBased(servers []*Server) *Server {
	var selectedServer *Server
	minResponseTime := time.Hour

	for _, server := range servers {
		server.mutex.RLock()
		responseTime := server.ResponseTime
		server.mutex.RUnlock()

		if responseTime < minResponseTime {
			minResponseTime = responseTime
			selectedServer = server
		}
	}

	return selectedServer
}

// random - случайная стратегия
func (lb *LoadBalancer) random(servers []*Server) *Server {
	return servers[rand.Intn(len(servers))]
}

// ServeHTTP - обработчик HTTP запросов
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Обновляем метрики
	lb.metrics.mutex.Lock()
	lb.metrics.TotalRequests++
	lb.metrics.LastUpdated = time.Now()
	lb.metrics.mutex.Unlock()

	var lastErr error
	maxRetries := lb.config.MaxRetries

	for attempt := 0; attempt <= maxRetries; attempt++ {
		server, err := lb.GetServer()
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				time.Sleep(lb.config.RetryDelay)
				continue
			}
			break
		}

		// Увеличиваем счетчик соединений
		server.mutex.Lock()
		server.CurrentConnections++
		server.mutex.Unlock()

		// Создаем reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(server.URL)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error for server %s: %v", server.URL.String(), err)

			// Увеличиваем счетчик ошибок
			server.mutex.Lock()
			server.ErrorCount++
			server.mutex.Unlock()

			// Отмечаем сервер как нездоровый
			server.mutex.Lock()
			server.IsHealthy = false
			server.mutex.Unlock()

			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		}

		// Обрабатываем запрос
		proxy.ServeHTTP(w, r)

		// Уменьшаем счетчик соединений
		server.mutex.Lock()
		server.CurrentConnections--
		server.mutex.Unlock()

		// Обновляем время ответа
		responseTime := time.Since(start)
		server.mutex.Lock()
		server.ResponseTime = responseTime
		server.mutex.Unlock()

		// Обновляем метрики успеха
		lb.metrics.mutex.Lock()
		lb.metrics.SuccessfulRequests++
		lb.metrics.AverageResponseTime = time.Duration(
			(int64(lb.metrics.AverageResponseTime) + int64(responseTime)) / 2,
		)
		lb.metrics.mutex.Unlock()

		return
	}

	// Все попытки неудачны
	lb.metrics.mutex.Lock()
	lb.metrics.FailedRequests++
	lb.metrics.mutex.Unlock()

	log.Printf("All servers failed after %d attempts. Last error: %v", maxRetries+1, lastErr)
	http.Error(w, "All services unavailable", http.StatusServiceUnavailable)
}

// GetMetrics - получение метрик
func (lb *LoadBalancer) GetMetrics() LoadBalancerMetrics {
	lb.metrics.mutex.RLock()
	defer lb.metrics.mutex.RUnlock()

	lb.mutex.RLock()
	healthyCount := 0
	for _, server := range lb.servers {
		if server.IsHealthy {
			healthyCount++
		}
	}
	totalCount := len(lb.servers)
	lb.mutex.RUnlock()

	metrics := *lb.metrics
	metrics.HealthyServers = healthyCount
	metrics.TotalServers = totalCount

	return metrics
}

// GetServerStatus - получение статуса серверов
func (lb *LoadBalancer) GetServerStatus() []ServerStatus {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	status := make([]ServerStatus, len(lb.servers))
	for i, server := range lb.servers {
		server.mutex.RLock()
		status[i] = ServerStatus{
			URL:                server.URL.String(),
			IsHealthy:          server.IsHealthy,
			CurrentConnections: server.CurrentConnections,
			MaxConnections:     server.MaxConnections,
			ResponseTime:       server.ResponseTime,
			ErrorCount:         server.ErrorCount,
			LastChecked:        server.LastChecked,
		}
		server.mutex.RUnlock()
	}

	return status
}

// ServerStatus - статус сервера
type ServerStatus struct {
	URL                string        `json:"url"`
	IsHealthy          bool          `json:"is_healthy"`
	CurrentConnections int           `json:"current_connections"`
	MaxConnections     int           `json:"max_connections"`
	ResponseTime       time.Duration `json:"response_time"`
	ErrorCount         int64         `json:"error_count"`
	LastChecked        time.Time     `json:"last_checked"`
}

// HealthChecker методы

func (hc *HealthChecker) start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAllServers()
		case <-hc.stopCh:
			return
		}
	}
}

func (hc *HealthChecker) checkAllServers() {
	for _, server := range hc.servers {
		go hc.checkServer(server)
	}
}

func (hc *HealthChecker) checkServer(server *Server) {
	client := &http.Client{
		Timeout: hc.timeout,
	}

	// Проверяем здоровье сервера
	resp, err := client.Get(server.URL.String() + "/health")

	server.mutex.Lock()
	defer server.mutex.Unlock()

	server.LastChecked = time.Now()

	if err != nil {
		server.IsHealthy = false
		log.Printf("Health check failed for server %s: %v", server.URL.String(), err)
		return
	}
	defer resp.Body.Close()

	server.IsHealthy = resp.StatusCode == http.StatusOK

	if !server.IsHealthy {
		log.Printf("Server %s is unhealthy (status: %d)", server.URL.String(), resp.StatusCode)
	}
}

func (hc *HealthChecker) updateServers(servers []*Server) {
	hc.servers = servers
}

func (hc *HealthChecker) stop() {
	close(hc.stopCh)
}

// Вспомогательные функции

func parseStrategy(strategy string) Strategy {
	switch strategy {
	case "weighted_round_robin":
		return WeightedRoundRobin
	case "least_connections":
		return LeastConnections
	case "response_time":
		return ResponseTime
	case "random":
		return Random
	default:
		return RoundRobin
	}
}

// UpdateStrategy - обновление стратегии балансировки
func (lb *LoadBalancer) UpdateStrategy(strategy string) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.strategy = parseStrategy(strategy)
	log.Printf("Load balancer strategy updated to: %s", strategy)
}

// Shutdown - остановка балансировщика
func (lb *LoadBalancer) Shutdown() {
	if lb.healthChecker != nil {
		lb.healthChecker.stop()
	}
	log.Println("Load balancer shutdown completed")
}

// ExportMetrics - экспорт метрик в JSON
func (lb *LoadBalancer) ExportMetrics() (string, error) {
	metrics := lb.GetMetrics()
	status := lb.GetServerStatus()

	data := map[string]interface{}{
		"metrics":   metrics,
		"servers":   status,
		"config":    lb.config,
		"timestamp": time.Now(),
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// CreateLoadBalancerHandler - создание HTTP обработчика для балансировщика
func CreateLoadBalancerHandler(lb *LoadBalancer) http.Handler {
	mux := http.NewServeMux()

	// Основной обработчик всех запросов
	mux.HandleFunc("/", lb.ServeHTTP)

	// Эндпоинты для мониторинга
	mux.HandleFunc("/lb/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		metrics := lb.GetMetrics()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	})

	mux.HandleFunc("/lb/servers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := lb.GetServerStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/lb/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		healthyServers := 0
		lb.mutex.RLock()
		for _, server := range lb.servers {
			if server.IsHealthy {
				healthyServers++
			}
		}
		totalServers := len(lb.servers)
		lb.mutex.RUnlock()

		status := "healthy"
		if healthyServers == 0 {
			status = "unhealthy"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else if healthyServers < totalServers/2 {
			status = "degraded"
			w.WriteHeader(http.StatusOK) // Но с предупреждением
		}

		response := map[string]interface{}{
			"status":          status,
			"healthy_servers": healthyServers,
			"total_servers":   totalServers,
			"timestamp":       time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return mux
}
