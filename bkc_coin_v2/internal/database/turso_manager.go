package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// 🎯 TursoManager для управления SQLite базой
type TursoManager struct {
	URL      string
	Token    string
	DB       *sql.DB
	Buffer   []TapRecord
	bufferMu sync.Mutex
	stopChan chan bool
}

// 📊 TapRecord для буферизации тапов
type TapRecord struct {
	UserID    int64     `json:"user_id"`
	Amount    int64     `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"session_id"`
}

// NewTursoManager создает менеджер Turso
func NewTursoManager(url, token string) *TursoManager {
	return &TursoManager{
		URL:      url,
		Token:    token,
		Buffer:   make([]TapRecord, 0, 1000),
		stopChan: make(chan bool, 1),
	}
}

// Initialize инициализирует Turso базу данных
func (t *TursoManager) Initialize() error {
	db, err := sql.Open("libsql", t.URL+"?authToken="+t.Token)
	if err != nil {
		return fmt.Errorf("failed to connect to Turso: %w", err)
	}
	t.DB = db

	// Создание таблиц
	if err := t.createTables(); err != nil {
		return fmt.Errorf("failed to create Turso tables: %w", err)
	}

	// Запуск буферизации
	go t.startBuffer()

	log.Printf("🎯 Turso initialized: %s", t.URL)
	return nil
}

// createTables создает таблицы для Turso
func (t *TursoManager) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user_balances (
			user_id INTEGER PRIMARY KEY,
			balance INTEGER NOT NULL DEFAULT 0,
			energy INTEGER NOT NULL DEFAULT 1000,
			max_energy INTEGER NOT NULL DEFAULT 1000,
			tap_value INTEGER NOT NULL DEFAULT 10,
			last_tap TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			daily_sessions INTEGER NOT NULL DEFAULT 0,
			last_session_date DATE DEFAULT CURRENT_DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tap_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			amount INTEGER NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			session_id TEXT,
			INDEX(user_id, timestamp)
		)`,
		`CREATE TABLE IF NOT EXISTS nft_inventory (
			user_id INTEGER PRIMARY KEY,
			nft_basic INTEGER DEFAULT 0,
			nft_pro INTEGER DEFAULT 0,
			nft_ultra INTEGER DEFAULT 0,
			purchase_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS power_upgrades (
			user_id INTEGER PRIMARY KEY,
			multiplier REAL DEFAULT 1.0,
			max_daily_earnings INTEGER DEFAULT 300,
			instant_regen_count INTEGER DEFAULT 3,
			last_upgrade TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := t.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// startBuffer запускает буферизацию тапов
func (t *TursoManager) startBuffer() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.flushBuffer()
		case <-t.stopChan:
			t.flushBuffer()
			return
		}
	}
}

// flushBuffer сбрасывает буфер тапов в базу
func (t *TursoManager) flushBuffer() {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	if len(t.Buffer) == 0 {
		return
	}

	// Копируем буфер
	buffer := make([]TapRecord, len(t.Buffer))
	copy(buffer, t.Buffer)

	// Очищаем буфер
	t.Buffer = t.Buffer[:0]

	// Вставляем пачкой
	if len(buffer) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tx, err := t.DB.BeginTx(ctx, nil)
		if err != nil {
			log.Printf("Failed to begin transaction for Turso buffer: %v", err)
			return
		}

		for _, record := range buffer {
			_, err = tx.ExecContext(ctx,
				"INSERT INTO tap_history (user_id, amount, timestamp, session_id) VALUES (?, ?, ?, ?)",
				record.UserID, record.Amount, record.Timestamp, record.SessionID)
			if err != nil {
				log.Printf("Failed to insert tap record: %v", err)
				tx.Rollback()
				return
			}
		}

		if err = tx.Commit(); err != nil {
			log.Printf("Failed to commit Turso buffer: %v", err)
		} else {
			log.Printf("🎯 Turso: Flushed %d tap records", len(buffer))
		}
	}
}

// AddTapToBuffer добавляет тап в буфер
func (t *TursoManager) AddTapToBuffer(userID int64, amount int64, sessionID string) {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	record := TapRecord{
		UserID:    userID,
		Amount:    amount,
		Timestamp: time.Now(),
		SessionID: sessionID,
	}

	t.Buffer = append(t.Buffer, record)

	// Если буфер переполнен, сбрасываем немедленно
	if len(t.Buffer) >= 1000 {
		t.flushBuffer()
	}
}

// GetUserBalance получает баланс пользователя
func (t *TursoManager) GetUserBalance(userID int64) (int64, error) {
	var balance int64
	err := t.DB.QueryRow("SELECT balance FROM user_balances WHERE user_id = ?", userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}

// UpdateUserBalance обновляет баланс пользователя
func (t *TursoManager) UpdateUserBalance(userID int64, balance int64) error {
	_, err := t.DB.Exec(
		"INSERT INTO user_balances (user_id, balance) VALUES (?, ?) ON CONFLICT(user_id) DO UPDATE SET balance = ?",
		userID, balance, balance)
	return err
}

// GetUserEnergy получает энергию пользователя
func (t *TursoManager) GetUserEnergy(userID int64) (int64, int64, error) {
	var energy, maxEnergy int64
	err := t.DB.QueryRow("SELECT energy, max_energy FROM user_balances WHERE user_id = ?", userID).Scan(&energy, &maxEnergy)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1000, 1000, nil
		}
		return 0, 0, err
	}
	return energy, maxEnergy, nil
}

// UpdateUserEnergy обновляет энергию пользователя
func (t *TursoManager) UpdateUserEnergy(userID int64, energy int64, maxEnergy int64) error {
	_, err := t.DB.Exec(
		"INSERT INTO user_balances (user_id, energy, max_energy) VALUES (?, ?, ?) ON CONFLICT(user_id) DO UPDATE SET energy = ?, max_energy = ?",
		userID, energy, maxEnergy, energy, maxEnergy)
	return err
}

// GetUserNFTInventory получает NFT инвентарь пользователя
func (t *TursoManager) GetUserNFTInventory(userID int64) (map[string]int64, error) {
	var basic, pro, ultra int64
	err := t.DB.QueryRow("SELECT nft_basic, nft_pro, nft_ultra FROM nft_inventory WHERE user_id = ?", userID).Scan(&basic, &pro, &ultra)
	if err != nil {
		if err == sql.ErrNoRows {
			return map[string]int64{
				"basic": 0,
				"pro":   0,
				"ultra": 0,
			}, nil
		}
		return nil, err
	}

	return map[string]int64{
		"basic": basic,
		"pro":   pro,
		"ultra": ultra,
	}, nil
}

// UpdateUserNFTInventory обновляет NFT инвентарь пользователя
func (t *TursoManager) UpdateUserNFTInventory(userID int64, nftType string, count int64) error {
	query := fmt.Sprintf("UPDATE nft_inventory SET nft_%s = nft_%s + ? WHERE user_id = ?", nftType, nftType)
	_, err := t.DB.Exec(query, count, userID)
	return err
}

// GetUserPowerUpgrades получает улучшения пользователя
func (t *TursoManager) GetUserPowerUpgrades(userID int64) (float64, int64, int64, error) {
	var multiplier float64
	var maxDaily, instantRegen int64
	err := t.DB.QueryRow("SELECT multiplier, max_daily_earnings, instant_regen_count FROM power_upgrades WHERE user_id = ?", userID).Scan(&multiplier, &maxDaily, &instantRegen)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1.0, 300, 3, nil
		}
		return 0, 0, 0, err
	}
	return multiplier, maxDaily, instantRegen, nil
}

// UpdateUserPowerUpgrades обновляет улучшения пользователя
func (t *TursoManager) UpdateUserPowerUpgrades(userID int64, multiplier float64, maxDaily int64, instantRegen int64) error {
	_, err := t.DB.Exec(
		"INSERT INTO power_upgrades (user_id, multiplier, max_daily_earnings, instant_regen_count) VALUES (?, ?, ?, ?) ON CONFLICT(user_id) DO UPDATE SET multiplier = ?, max_daily_earnings = ?, instant_regen_count = ?",
		userID, multiplier, maxDaily, instantRegen, multiplier, maxDaily, instantRegen)
	return err
}

// Ping проверяет соединение с Turso
func (t *TursoManager) Ping() error {
	if t.DB == nil {
		return fmt.Errorf("Turso database not initialized")
	}
	return t.DB.Ping()
}

// Close закрывает соединение с Turso
func (t *TursoManager) Close() error {
	// Останавливаем буферизацию
	close(t.stopChan)

	// Сбрасываем буфер перед закрытием
	t.flushBuffer()

	if t.DB != nil {
		return t.DB.Close()
	}
	return nil
}
