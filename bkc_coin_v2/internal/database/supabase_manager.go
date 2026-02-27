package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// 🔥 SupabaseManager для управления авторизацией
type SupabaseManager struct {
	URL  string
	Key  string
	DB   *sql.DB
}

// AuthUser структура пользователя авторизации
type AuthUser struct {
	ID          int64     `json:"id" db:"id"`
	TelegramID  int64     `json:"telegram_id" db:"telegram_id"`
	Username    string    `json:"username" db:"username"`
	FirstName   string    `json:"first_name" db:"first_name"`
	AuthToken   string    `json:"auth_token" db:"auth_token"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	LastLogin   time.Time `json:"last_login" db:"last_login"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserSession структура сессии
type UserSession struct {
	SessionID string    `json:"session_id" db:"session_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// AccessToken структура токена доступа
type AccessToken struct {
	TokenID     string    `json:"token_id" db:"token_id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	TokenType   string    `json:"token_type" db:"token_type"`
	Permissions string    `json:"permissions" db:"permissions"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
}

// NewSupabaseManager создает менеджер Supabase
func NewSupabaseManager(url, key string) *SupabaseManager {
	return &SupabaseManager{
		URL:  url,
		Key:  key,
	}
}

// Initialize инициализирует Supabase
func (s *SupabaseManager) Initialize() error {
	db, err := sql.Open("pgx", s.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to Supabase: %w", err)
	}
	s.DB = db

	// Создание таблиц
	if err := s.createTables(); err != nil {
		return fmt.Errorf("failed to create Supabase tables: %w", err)
	}

	log.Printf("🔥 Supabase initialized: %s", s.URL)
	return nil
}

// createTables создает таблицы для Supabase
func (s *SupabaseManager) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS auth_users (
			id BIGINT PRIMARY KEY,
			telegram_id BIGINT UNIQUE NOT NULL,
			username TEXT,
			first_name TEXT,
			auth_token TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_sessions (
			session_id TEXT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			INDEX(user_id),
			INDEX(expires_at)
		)`,
		`CREATE TABLE IF NOT EXISTS access_tokens (
			token_id TEXT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			token_type TEXT NOT NULL,
			permissions JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			INDEX(user_id),
			INDEX(expires_at)
		)`,
	}

	for _, query := range queries {
		if _, err := s.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// CreateAuthUser создает пользователя авторизации
func (s *SupabaseManager) CreateAuthUser(telegramID int64, username, firstName, authToken string) error {
	_, err := s.DB.Exec(
		"INSERT INTO auth_users (telegram_id, username, first_name, auth_token, is_active, last_login, created_at) VALUES (?, ?, ?, ?, TRUE, ?, ?)",
		telegramID, username, firstName, authToken, time.Now(), time.Now())
	return err
}

// GetAuthUserByTelegramID получает пользователя по Telegram ID
func (s *SupabaseManager) GetAuthUserByTelegramID(telegramID int64) (*AuthUser, error) {
	var user AuthUser
	err := s.DB.QueryRow(
		"SELECT id, telegram_id, username, first_name, auth_token, is_active, last_login, created_at FROM auth_users WHERE telegram_id = ?",
		telegramID).Scan(&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.AuthToken, &user.IsActive, &user.LastLogin, &user.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &user, nil
}

// GetAuthUserByID получает пользователя по ID
func (s *SupabaseManager) GetAuthUserByID(userID int64) (*AuthUser, error) {
	var user AuthUser
	err := s.DB.QueryRow(
		"SELECT id, telegram_id, username, first_name, auth_token, is_active, last_login, created_at FROM auth_users WHERE id = ?",
		userID).Scan(&user.ID, &user.TelegramID, &user.Username, &user.FirstName, &user.AuthToken, &user.IsActive, &user.LastLogin, &user.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &user, nil
}

// UpdateAuthUser обновляет пользователя авторизации
func (s *SupabaseManager) UpdateAuthUser(userID int64, username, firstName string) error {
	_, err := s.DB.Exec(
		"UPDATE auth_users SET username = ?, first_name = ?, last_login = ? WHERE id = ?",
		username, firstName, time.Now(), userID)
	return err
}

// UpdateAuthToken обновляет токен авторизации
func (s *SupabaseManager) UpdateAuthToken(userID int64, authToken string) error {
	_, err := s.DB.Exec(
		"UPDATE auth_users SET auth_token = ?, last_login = ? WHERE id = ?",
		authToken, time.Now(), userID)
	return err
}

// DeactivateUser деактивирует пользователя
func (s *SupabaseManager) DeactivateUser(userID int64) error {
	_, err := s.DB.Exec(
		"UPDATE auth_users SET is_active = FALSE WHERE id = ?",
		userID)
	return err
}

// CreateUserSession создает сессию пользователя
func (s *SupabaseManager) CreateUserSession(sessionID string, userID int64, expiresAt time.Time) error {
	_, err := s.DB.Exec(
		"INSERT INTO user_sessions (session_id, user_id, created_at, expires_at, is_active) VALUES (?, ?, ?, ?, TRUE)",
		sessionID, userID, time.Now(), expiresAt)
	return err
}

// GetUserSession получает сессию пользователя
func (s *SupabaseManager) GetUserSession(sessionID string) (*UserSession, error) {
	var session UserSession
	err := s.DB.QueryRow(
		"SELECT session_id, user_id, created_at, expires_at, is_active FROM user_sessions WHERE session_id = ? AND is_active = TRUE AND expires_at > NOW()",
		sessionID).Scan(&session.SessionID, &session.UserID, &session.CreatedAt, &session.ExpiresAt, &session.IsActive)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &session, nil
}

// InvalidateSession делает сессию неактивной
func (s *SupabaseManager) InvalidateSession(sessionID string) error {
	_, err := s.DB.Exec(
		"UPDATE user_sessions SET is_active = FALSE WHERE session_id = ?",
		sessionID)
	return err
}

// CleanupExpiredSessions очищает истекшие сессии
func (s *SupabaseManager) CleanupExpiredSessions() error {
	_, err := s.DB.Exec(
		"UPDATE user_sessions SET is_active = FALSE WHERE expires_at < NOW() OR is_active = FALSE")
	return err
}

// CreateAccessToken создает токен доступа
func (s *SupabaseManager) CreateAccessToken(tokenID, userID, tokenType, permissions string, expiresAt time.Time) error {
	_, err := s.DB.Exec(
		"INSERT INTO access_tokens (token_id, user_id, token_type, permissions, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)",
		tokenID, userID, tokenType, permissions, time.Now(), expiresAt)
	return err
}

// GetAccessToken получает токен доступа
func (s *SupabaseManager) GetAccessToken(tokenID string) (*AccessToken, error) {
	var token AccessToken
	err := s.DB.QueryRow(
		"SELECT token_id, user_id, token_type, permissions, created_at, expires_at FROM access_tokens WHERE token_id = ? AND expires_at > NOW()",
		tokenID).Scan(&token.TokenID, &token.UserID, &token.TokenType, &token.Permissions, &token.CreatedAt, &token.ExpiresAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &token, nil
}

// RevokeAccessToken отзывает токен доступа
func (s *SupabaseManager) RevokeAccessToken(tokenID string) error {
	_, err := s.DB.Exec(
		"DELETE FROM access_tokens WHERE token_id = ?",
		tokenID)
	return err
}

// CleanupExpiredTokens очищает истекшие токены
func (s *SupabaseManager) CleanupExpiredTokens() error {
	_, err := s.DB.Exec(
		"DELETE FROM access_tokens WHERE expires_at < NOW()")
	return err
}

// Ping проверяет соединение с Supabase
func (s *SupabaseManager) Ping() error {
	if s.DB == nil {
		return fmt.Errorf("Supabase database not initialized")
	}
	return s.DB.Ping()
}

// Close закрывает соединение с Supabase
func (s *SupabaseManager) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}
