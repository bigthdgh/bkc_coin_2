package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// 🐛 CockroachManager для управления профилей
type CockroachManager struct {
	URL string
	DB  *sql.DB
}

// UserProfile структура пользователя
type UserProfile struct {
	UserID         int64     `json:"user_id" db:"user_id"`
	Username       string    `json:"username" db:"username"`
	FirstName      string    `json:"first_name" db:"first_name"`
	ReferralID     *int64    `json:"referral_id" db:"referral_id"`
	ReferralCount  int       `json:"referral_count" db:"referral_count"`
	TotalEarned    int64     `json:"total_earned" db:"total_earned"`
	VerificationLevel int       `json:"verification_level" db:"verification_level"`
	IsBanned       bool      `json:"is_banned" db:"is_banned"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Referral структура реферала
type Referral struct {
	ID         int64     `json:"id" db:"id"`
	ReferrerID int64     `json:"referrer_id" db:"referrer_id"`
	ReferredID int64     `json:"referred_id" db:"referred_id"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// UserStats структура статистики
type UserStats struct {
	UserID          int64     `json:"user_id" db:"user_id"`
	TotalTaps       int64     `json:"total_taps" db:"total_taps"`
	TotalSessions    int       `json:"total_sessions" db:"total_sessions"`
	AvgSessionTime  int       `json:"avg_session_time" db:"avg_session_time"`
	LastActive      time.Time `json:"last_active" db:"last_active"`
}

// NewCockroachManager создает менеджер CockroachDB
func NewCockroachManager(url string) *CockroachManager {
	return &CockroachManager{
		URL: url,
	}
}

// Initialize инициализирует CockroachDB
func (c *CockroachManager) Initialize() error {
	db, err := sql.Open("pgx", c.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to CockroachDB: %w", err)
	}
	c.DB = db

	// Создание таблиц
	if err := c.createTables(); err != nil {
		return fmt.Errorf("failed to create CockroachDB tables: %w", err)
	}

	log.Printf("🐛 CockroachDB initialized: %s", c.URL)
	return nil
}

// createTables создает таблицы для CockroachDB
func (c *CockroachManager) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user_profiles (
			user_id BIGINT PRIMARY KEY,
			username TEXT,
			first_name TEXT,
			referral_id BIGINT,
			referral_count INTEGER DEFAULT 0,
			total_earned BIGINT DEFAULT 0,
			verification_level INTEGER DEFAULT 0,
			is_banned BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS referrals (
			id BIGSERIAL PRIMARY KEY,
			referrer_id BIGINT NOT NULL,
			referred_id BIGINT NOT NULL,
			status TEXT DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX(referrer_id),
			INDEX(referred_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_stats (
			user_id BIGINT PRIMARY KEY,
			total_taps BIGINT DEFAULT 0,
			total_sessions INTEGER DEFAULT 0,
			avg_session_time INTEGER DEFAULT 0,
			last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX(last_active)
		)`,
		`CREATE TABLE IF NOT EXISTS p2p_orders (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			order_type TEXT NOT NULL,
			amount BIGINT NOT NULL,
			price DECIMAL(20,8) NOT NULL,
			status TEXT DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX(user_id, status),
			INDEX(created_at)
		)`,
		`CREATE TABLE IF NOT EXISTS bank_loans (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			amount BIGINT NOT NULL,
			interest_rate DECIMAL(5,2) DEFAULT 5.00,
			status TEXT DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			due_date TIMESTAMP,
			INDEX(user_id, status)
		)`,
	}

	for _, query := range queries {
		if _, err := c.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// CreateUserProfile создает профиль пользователя
func (c *CockroachManager) CreateUserProfile(profile *UserProfile) error {
	_, err := c.DB.Exec(
		`INSERT INTO user_profiles (user_id, username, first_name, referral_id, verification_level, is_banned, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		profile.UserID, profile.Username, profile.FirstName, profile.ReferralID, 
		profile.VerificationLevel, profile.IsBanned, time.Now(), time.Now())
	return err
}

// GetUserProfile получает профиль пользователя
func (c *CockroachManager) GetUserProfile(userID int64) (*UserProfile, error) {
	var profile UserProfile
	err := c.DB.QueryRow(
		"SELECT user_id, username, first_name, referral_id, referral_count, total_earned, verification_level, is_banned, created_at, updated_at FROM user_profiles WHERE user_id = ?",
		userID).Scan(&profile.UserID, &profile.Username, &profile.FirstName, &profile.ReferralID, 
		&profile.ReferralCount, &profile.TotalEarned, &profile.VerificationLevel, &profile.IsBanned, 
		&profile.CreatedAt, &profile.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &profile, nil
}

// UpdateUserProfile обновляет профиль пользователя
func (c *CockroachManager) UpdateUserProfile(profile *UserProfile) error {
	_, err := c.DB.Exec(
		`UPDATE user_profiles SET username = ?, first_name = ?, verification_level = ?, is_banned = ?, updated_at = ? 
		 WHERE user_id = ?`,
		profile.Username, profile.FirstName, profile.VerificationLevel, profile.IsBanned, time.Now(), profile.UserID)
	return err
}

// CreateReferral создает реферала
func (c *CockroachManager) CreateReferral(referrerID, referredID int64) error {
	_, err := c.DB.Exec(
		"INSERT INTO referrals (referrer_id, referred_id, status, created_at) VALUES (?, ?, 'pending', ?)",
		referrerID, referredID, time.Now())
	return err
}

// GetReferrals получает рефералов пользователя
func (c *CockroachManager) GetReferrals(userID int64) ([]*Referral, error) {
	rows, err := c.DB.Query(
		"SELECT id, referrer_id, referred_id, status, created_at FROM referrals WHERE referrer_id = ? ORDER BY created_at DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var referrals []*Referral
	for rows.Next() {
		var referral Referral
		err := rows.Scan(&referral.ID, &referral.ReferrerID, &referral.ReferredID, &referral.Status, &referral.CreatedAt)
		if err != nil {
			return nil, err
		}
		referrals = append(referrals, &referral)
	}

	return referrals, nil
}

// UpdateReferralCount обновляет количество рефералов
func (c *CockroachManager) UpdateReferralCount(userID int64) error {
	_, err := c.DB.Exec(
		"UPDATE user_profiles SET referral_count = referral_count + 1 WHERE user_id = ?",
		userID)
	return err
}

// GetUserStats получает статистику пользователя
func (c *CockroachManager) GetUserStats(userID int64) (*UserStats, error) {
	var stats UserStats
	err := c.DB.QueryRow(
		"SELECT user_id, total_taps, total_sessions, avg_session_time, last_active FROM user_stats WHERE user_id = ?",
		userID).Scan(&stats.UserID, &stats.TotalTaps, &stats.TotalSessions, &stats.AvgSessionTime, &stats.LastActive)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return &UserStats{
				UserID:         userID,
				TotalTaps:       0,
				TotalSessions:    0,
				AvgSessionTime:  0,
				LastActive:      time.Now(),
			}, nil
		}
		return nil, err
	}
	
	return &stats, nil
}

// UpdateUserStats обновляет статистику пользователя
func (c *CockroachManager) UpdateUserStats(userID int64, taps int64, sessionTime int) error {
	_, err := c.DB.Exec(
		`INSERT INTO user_stats (user_id, total_taps, total_sessions, avg_session_time, last_active) 
		 VALUES (?, ?, ?, ?, ?) 
		 ON CONFLICT(user_id) DO UPDATE SET 
		 total_taps = user_stats.total_taps + ?, 
		 total_sessions = user_stats.total_sessions + 1,
		 avg_session_time = (user_stats.avg_session_time * user_stats.total_sessions + ?) / (user_stats.total_sessions + 1),
		 last_active = ?`,
		userID, taps, 1, sessionTime, time.Now(), taps, sessionTime, time.Now())
	return err
}

// CreateP2POrder создает P2P ордер
func (c *CockroachManager) CreateP2POrder(userID int64, orderType string, amount int64, price float64) error {
	_, err := c.DB.Exec(
		"INSERT INTO p2p_orders (user_id, order_type, amount, price, status, created_at) VALUES (?, ?, ?, ?, 'pending', ?)",
		userID, orderType, amount, price, time.Now())
	return err
}

// GetP2POrders получает P2P ордера
func (c *CockroachManager) GetP2POrders(limit int) ([]map[string]interface{}, error) {
	rows, err := c.DB.Query(
		"SELECT id, user_id, order_type, amount, price, status, created_at FROM p2p_orders WHERE status = 'pending' ORDER BY created_at DESC LIMIT ?",
		limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []map[string]interface{}
	for rows.Next() {
		var order struct {
			ID        int64     `json:"id"`
			UserID    int64     `json:"user_id"`
			OrderType string    `json:"order_type"`
			Amount    int64     `json:"amount"`
			Price     float64    `json:"price"`
			Status    string    `json:"status"`
			CreatedAt time.Time `json:"created_at"`
		}
		err := rows.Scan(&order.ID, &order.UserID, &order.OrderType, &order.Amount, &order.Price, &order.Status, &order.CreatedAt)
		if err != nil {
			return nil, err
		}
		
		orders = append(orders, map[string]interface{}{
			"id":         order.ID,
			"user_id":    order.UserID,
			"order_type": order.OrderType,
			"amount":     order.Amount,
			"price":      order.Price,
			"status":     order.Status,
			"created_at": order.CreatedAt,
		})
	}

	return orders, nil
}

// CreateBankLoan создает банковский кредит
func (c *CockroachManager) CreateBankLoan(userID int64, amount int64, interestRate float64, dueDate time.Time) error {
	_, err := c.DB.Exec(
		"INSERT INTO bank_loans (user_id, amount, interest_rate, status, created_at, due_date) VALUES (?, ?, ?, 'active', ?, ?)",
		userID, amount, interestRate, time.Now(), dueDate)
	return err
}

// GetUserLoans получает кредиты пользователя
func (c *CockroachManager) GetUserLoans(userID int64) ([]map[string]interface{}, error) {
	rows, err := c.DB.Query(
		"SELECT id, user_id, amount, interest_rate, status, created_at, due_date FROM bank_loans WHERE user_id = ? ORDER BY created_at DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []map[string]interface{}
	for rows.Next() {
		var loan struct {
			ID           int64     `json:"id"`
			UserID       int64     `json:"user_id"`
			Amount       int64     `json:"amount"`
			InterestRate float64    `json:"interest_rate"`
			Status       string    `json:"status"`
			CreatedAt    time.Time `json:"created_at"`
			DueDate      time.Time `json:"due_date"`
		}
		err := rows.Scan(&loan.ID, &loan.UserID, &loan.Amount, &loan.InterestRate, &loan.Status, &loan.CreatedAt, &loan.DueDate)
		if err != nil {
			return nil, err
		}
		
		loans = append(loans, map[string]interface{}{
			"id":            loan.ID,
			"user_id":       loan.UserID,
			"amount":        loan.Amount,
			"interest_rate":  loan.InterestRate,
			"status":        loan.Status,
			"created_at":    loan.CreatedAt,
			"due_date":      loan.DueDate,
		})
	}

	return loans, nil
}

// Ping проверяет соединение с CockroachDB
func (c *CockroachManager) Ping() error {
	if c.DB == nil {
		return fmt.Errorf("CockroachDB not initialized")
	}
	return c.DB.Ping()
}

// Close закрывает соединение с CockroachDB
func (c *CockroachManager) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
