package microservices

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"bkc_coin_v2/internal/database"
)

// UserService - микросервис для управления пользователями
type UserService struct {
	db          *database.UnifiedDB
	server      *http.Server
	port        int
	userCache   map[int64]*UserProfile
	cacheMutex  sync.RWMutex
	metrics     *ServiceMetrics
}

// UserProfile - профиль пользователя
type UserProfile struct {
	UserID          int64     `json:"user_id"`
	Username        string    `json:"username"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Avatar          string    `json:"avatar"`
	Balance         int64     `json:"balance"`
	FrozenBalance   int64     `json:"frozen_balance"`
	Level           int       `json:"level"`
	Experience      int64     `json:"experience"`
	Energy          float64   `json:"energy"`
	EnergyMax       float64   `json:"energy_max"`
	ReferralsCount  int64     `json:"referrals_count"`
	ReferralBonus   int64     `json:"referral_bonus"`
	IsVIP           bool      `json:"is_vip"`
	VIPLevel        int       `json:"vip_level"`
	Language        string    `json:"language"`
	Country         string    `json:"country"`
	Timezone        string    `json:"timezone"`
	Currency        string    `json:"currency"`
	Preferences     UserPrefs `json:"preferences"`
	Stats           UserStats `json:"stats"`
	CreatedAt       time.Time `json:"created_at"`
	LastActive      time.Time `json:"last_active"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserPrefs - настройки пользователя
type UserPrefs struct {
	NotificationsEnabled bool     `json:"notifications_enabled"`
	EmailNotifications   bool     `json:"email_notifications"`
	PushNotifications    bool     `json:"push_notifications"`
	Theme                string   `json:"theme"`
	Language             string   `json:"language"`
	Currency             string   `json:"currency"`
	Timezone             string   `json:"timezone"`
	Privacy              Privacy  `json:"privacy"`
}

// Privacy - настройки приватности
type Privacy struct {
	ProfileVisible    bool `json:"profile_visible"`
	StatsVisible      bool `json:"stats_visible"`
	FriendsVisible    bool `json:"friends_visible"`
	ReferralsVisible  bool `json:"referrals_visible"`
}

// UserStats - статистика пользователя
type UserStats struct {
	TotalTaps        int64     `json:"total_taps"`
	DailyTaps        int64     `json:"daily_taps"`
	TotalEarned      int64     `json:"total_earned"`
	DailyEarned      int64     `json:"daily_earned"`
	BestDayTaps      int64     `json:"best_day_taps"`
	BestDayEarned    int64     `json:"best_day_earned"`
	ConsecutiveDays  int       `json:"consecutive_days"`
	FirstTap         time.Time `json:"first_tap"`
	LastTap          time.Time `json:"last_tap"`
	AvgTapsPerHour   float64   `json:"avg_taps_per_hour"`
	AvgEarnPerHour   float64   `json:"avg_earn_per_hour"`
}

// ServiceMetrics - метрики сервиса
type ServiceMetrics struct {
	RequestsTotal    int64     `json:"requests_total"`
	RequestsPerSecond float64   `json:"requests_per_second"`
	ResponseTime     time.Duration `json:"response_time"`
	ErrorRate        float64   `json:"error_rate"`
	ActiveUsers      int64     `json:"active_users"`
	CacheHitRate     float64   `json:"cache_hit_rate"`
	LastUpdated      time.Time `json:"last_updated"`
	mutex            sync.RWMutex
}

// NewUserService - создание UserService
func NewUserService(db *database.UnifiedDB, port int) *UserService {
	us := &UserService{
		db:        db,
		port:      port,
		userCache: make(map[int64]*UserProfile),
		metrics:   &ServiceMetrics{},
	}

	// Запускаем HTTP сервер
	go us.startServer()
	
	// Запускаем очистку кэша
	go us.cleanupCache()
	
	return us
}

// startServer - запуск HTTP сервера
func (us *UserService) startServer() {
	mux := http.NewServeMux()
	
	// Регистрируем обработчики
	mux.HandleFunc("/api/user/", us.handleUserRequest)
	mux.HandleFunc("/api/user/profile", us.handleProfileRequest)
	mux.HandleFunc("/api/user/stats", us.handleStatsRequest)
	mux.HandleFunc("/api/user/preferences", us.handlePreferencesRequest)
	mux.HandleFunc("/api/health", us.handleHealthCheck)
	mux.HandleFunc("/api/metrics", us.handleMetrics)

	us.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", us.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("UserService starting on port %d", us.port)
	if err := us.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("UserService server error: %v", err)
	}
}

// handleUserRequest - обработка запросов пользователя
func (us *UserService) handleUserRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	us.metrics.mutex.Lock()
	us.metrics.RequestsTotal++
	us.metrics.mutex.Unlock()

	defer func() {
		us.metrics.mutex.Lock()
		us.metrics.ResponseTime = time.Since(start)
		us.metrics.LastUpdated = time.Now()
		us.metrics.mutex.Unlock()
	}()

	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case http.MethodGet:
		us.getUser(w, r)
	case http.MethodPost:
		us.createUser(w, r)
	case http.MethodPut:
		us.updateUser(w, r)
	case http.MethodDelete:
		us.deleteUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getUser - получение пользователя
func (us *UserService) getUser(w http.ResponseWriter, r *http.Request) {
	userID := extractUserIDFromPath(r.URL.Path)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Проверяем кэш
	us.cacheMutex.RLock()
	if profile, exists := us.userCache[userID]; exists {
		us.cacheMutex.RUnlock()
		us.updateCacheHitRate(true)
		json.NewEncoder(w).Encode(profile)
		return
	}
	us.cacheMutex.RUnlock()

	us.updateCacheHitRate(false)

	// Получаем из базы данных
	profile, err := us.getUserFromDB(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user: %v", err), http.StatusInternalServerError)
		return
	}

	// Сохраняем в кэш
	us.cacheMutex.Lock()
	us.userCache[userID] = profile
	us.cacheMutex.Unlock()

	json.NewEncoder(w).Encode(profile)
}

// createUser - создание пользователя
func (us *UserService) createUser(w http.ResponseWriter, r *http.Request) {
	var user UserProfile
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Создаем пользователя в базе данных
	err := us.createUserInDB(&user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	// Сохраняем в кэш
	us.cacheMutex.Lock()
	us.userCache[user.UserID] = &user
	us.cacheMutex.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// updateUser - обновление пользователя
func (us *UserService) updateUser(w http.ResponseWriter, r *http.Request) {
	userID := extractUserIDFromPath(r.URL.Path)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user UserProfile
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user.UserID = userID
	user.UpdatedAt = time.Now()

	// Обновляем в базе данных
	err := us.updateUserInDB(&user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user: %v", err), http.StatusInternalServerError)
		return
	}

	// Обновляем кэш
	us.cacheMutex.Lock()
	us.userCache[userID] = &user
	us.cacheMutex.Unlock()

	json.NewEncoder(w).Encode(user)
}

// deleteUser - удаление пользователя
func (us *UserService) deleteUser(w http.ResponseWriter, r *http.Request) {
	userID := extractUserIDFromPath(r.URL.Path)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Удаляем из базы данных
	err := us.deleteUserFromDB(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete user: %v", err), http.StatusInternalServerError)
		return
	}

	// Удаляем из кэша
	us.cacheMutex.Lock()
	delete(us.userCache, userID)
	us.cacheMutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

// handleProfileRequest - обработка запросов профиля
func (us *UserService) handleProfileRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		us.getProfile(w, r)
	case http.MethodPut:
		us.updateProfile(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getProfile - получение профиля
func (us *UserService) getProfile(w http.ResponseWriter, r *http.Request) {
	userID := extractUserIDFromQuery(r)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	profile, err := us.getUserFromDB(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get profile: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(profile)
}

// updateProfile - обновление профиля
func (us *UserService) updateProfile(w http.ResponseWriter, r *http.Request) {
	userID := extractUserIDFromQuery(r)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := us.updateProfileInDB(userID, updates)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update profile: %v", err), http.StatusInternalServerError)
		return
	}

	// Обновляем кэш
	us.cacheMutex.Lock()
	if profile, exists := us.userCache[userID]; exists {
		us.updateProfileFromMap(profile, updates)
	}
	us.cacheMutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

// handleStatsRequest - обработка запросов статистики
func (us *UserService) handleStatsRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := extractUserIDFromQuery(r)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	stats, err := us.getUserStats(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// handlePreferencesRequest - обработка запросов настроек
func (us *UserService) handlePreferencesRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		us.getPreferences(w, r)
	case http.MethodPut:
		us.updatePreferences(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getPreferences - получение настроек
func (us *UserService) getPreferences(w http.ResponseWriter, r *http.Request) {
	userID := extractUserIDFromQuery(r)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	prefs, err := us.getUserPreferences(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get preferences: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(prefs)
}

// updatePreferences - обновление настроек
func (us *UserService) updatePreferences(w http.ResponseWriter, r *http.Request) {
	userID := extractUserIDFromQuery(r)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var prefs UserPrefs
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := us.updateUserPreferences(userID, prefs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update preferences: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleHealthCheck - проверка здоровья
func (us *UserService) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "UserService",
		"port":      us.port,
	}

	json.NewEncoder(w).Encode(health)
}

// handleMetrics - метрики сервиса
func (us *UserService) handleMetrics(w http.ResponseWriter, r *http.Request) {
	us.metrics.mutex.RLock()
	metrics := *us.metrics
	us.metrics.mutex.RUnlock()

	json.NewEncoder(w).Encode(metrics)
}

// Вспомогательные функции

func extractUserIDFromPath(path string) int64 {
	// Извлекает ID из пути /api/user/123
	// В реальном приложении здесь будет парсинг пути
	return 123 // Заглушка
}

func extractUserIDFromQuery(r *http.Request) int64 {
	// Извлекает ID из query параметров
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		return 0
	}
	// Конвертация string в int64
	return 123 // Заглушка
}

func (us *UserService) updateCacheHitRate(hit bool) {
	us.metrics.mutex.Lock()
	defer us.metrics.mutex.Unlock()
	
	if hit {
		us.metrics.CacheHitRate = (us.metrics.CacheHitRate*0.9 + 0.1)
	} else {
		us.metrics.CacheHitRate = (us.metrics.CacheHitRate * 0.9)
	}
}

func (us *UserService) updateProfileFromMap(profile *UserProfile, updates map[string]interface{}) {
	// Обновляет профиль из карты
	for key, value := range updates {
		switch key {
		case "username":
			if v, ok := value.(string); ok {
				profile.Username = v
			}
		case "email":
			if v, ok := value.(string); ok {
				profile.Email = v
			}
		case "language":
			if v, ok := value.(string); ok {
				profile.Language = v
			}
		}
	}
}

func (us *UserService) cleanupCache() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		us.cacheMutex.Lock()
		now := time.Now()
		
		for userID, profile := range us.userCache {
			// Удаляем старые записи из кэша
			if now.Sub(profile.LastActive) > 24*time.Hour {
				delete(us.userCache, userID)
			}
		}
		
		us.cacheMutex.Unlock()
		log.Printf("UserService cache cleanup completed")
	}
}

// Функции работы с базой данных (заглушки)

func (us *UserService) getUserFromDB(userID int64) (*UserProfile, error) {
	// В реальном приложении здесь будет запрос к БД
	return &UserProfile{
		UserID:    userID,
		Username:  fmt.Sprintf("User%d", userID),
		Balance:   10000,
		Level:     1,
		Language:  "en",
		Country:   "US",
		Currency:  "USD",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		LastActive: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (us *UserService) createUserInDB(user *UserProfile) error {
	// В реальном приложении здесь будет запрос к БД
	log.Printf("Creating user: %d", user.UserID)
	return nil
}

func (us *UserService) updateUserInDB(user *UserProfile) error {
	// В реальном приложении здесь будет запрос к БД
	log.Printf("Updating user: %d", user.UserID)
	return nil
}

func (us *UserService) deleteUserFromDB(userID int64) error {
	// В реальном приложении здесь будет запрос к БД
	log.Printf("Deleting user: %d", userID)
	return nil
}

func (us *UserService) updateProfileInDB(userID int64, updates map[string]interface{}) error {
	// В реальном приложении здесь будет запрос к БД
	log.Printf("Updating profile: %d", userID)
	return nil
}

func (us *UserService) getUserStats(userID int64) (*UserStats, error) {
	// В реальном приложении здесь будет запрос к БД
	return &UserStats{
		TotalTaps:       10000,
		DailyTaps:       500,
		TotalEarned:     1000000,
		DailyEarned:     50000,
		BestDayTaps:     1000,
		BestDayEarned:   100000,
		ConsecutiveDays: 7,
		FirstTap:        time.Now().Add(-30 * 24 * time.Hour),
		LastTap:         time.Now(),
		AvgTapsPerHour:  20.5,
		AvgEarnPerHour:  1025.0,
	}, nil
}

func (us *UserService) getUserPreferences(userID int64) (*UserPrefs, error) {
	// В реальном приложении здесь будет запрос к БД
	return &UserPrefs{
		NotificationsEnabled: true,
		EmailNotifications:   true,
		PushNotifications:    true,
		Theme:                "light",
		Language:             "en",
		Currency:             "USD",
		Timezone:             "UTC",
		Privacy: Privacy{
			ProfileVisible:   true,
			StatsVisible:     true,
			FriendsVisible:   true,
			ReferralsVisible: true,
		},
	}, nil
}

func (us *UserService) updateUserPreferences(userID int64, prefs UserPrefs) error {
	// В реальном приложении здесь будет запрос к БД
	log.Printf("Updating preferences: %d", userID)
	return nil
}

// Shutdown - остановка сервиса
func (us *UserService) Shutdown(ctx context.Context) error {
	log.Printf("UserService shutting down...")
	return us.server.Shutdown(ctx)
}
