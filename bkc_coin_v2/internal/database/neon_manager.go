package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// 🌟 NeonManager для управления логами
type NeonManager struct {
	URL string
	DB  *sql.DB
}

// ActivityLog структура лога активности
type ActivityLog struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Action     string    `json:"action" db:"action"`
	Details    string    `json:"details" db:"details"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	ServiceType string    `json:"service_type" db:"service_type"`
}

// TransactionHistory структура истории транзакций
type TransactionHistory struct {
	ID            int64     `json:"id" db:"id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	TransactionType string    `json:"transaction_type" db:"transaction_type"`
	Amount        int64     `json:"amount" db:"amount"`
	BalanceBefore int64     `json:"balance_before" db:"balance_before"`
	BalanceAfter  int64     `json:"balance_after" db:"balance_after"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	ServiceType   string    `json:"service_type" db:"service_type"`
}

// SystemMetrics структура системных метрик
type SystemMetrics struct {
	ID          int64     `json:"id" db:"id"`
	MetricName  string    `json:"metric_name" db:"metric_name"`
	MetricValue float64   `json:"metric_value" db:"metric_value"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	ServiceType string    `json:"service_type" db:"service_type"`
}

// Task структура задачи
type Task struct {
	ID          int64     `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Reward      int64     `json:"reward" db:"reward"`
	Type        string    `json:"type" db:"type"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserTask структура выполнения задачи
type UserTask struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	TaskID     int64     `json:"task_id" db:"task_id"`
	Status     string    `json:"status" db:"status"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// GameResult структура результата игры
type GameResult struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	GameType  string    `json:"game_type" db:"game_type"`
	BetAmount int64     `json:"bet_amount" db:"bet_amount"`
	Result    string    `json:"result" db:"result"`
	WinAmount int64     `json:"win_amount" db:"win_amount"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NewNeonManager создает менеджер Neon
func NewNeonManager(url string) *NeonManager {
	return &NeonManager{
		URL: url,
	}
}

// Initialize инициализирует Neon
func (n *NeonManager) Initialize() error {
	db, err := sql.Open("pgx", n.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to Neon: %w", err)
	}
	n.DB = db

	// Создание таблиц
	if err := n.createTables(); err != nil {
		return fmt.Errorf("failed to create Neon tables: %w", err)
	}

	log.Printf("🌟 Neon initialized: %s", n.URL)
	return nil
}

// createTables создает таблицы для Neon
func (n *NeonManager) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			action TEXT NOT NULL,
			details JSONB,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			service_type TEXT,
			INDEX(user_id, timestamp),
			INDEX(action, timestamp)
		)`,
		`CREATE TABLE IF NOT EXISTS transaction_history (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			transaction_type TEXT NOT NULL,
			amount BIGINT NOT NULL,
			balance_before BIGINT NOT NULL,
			balance_after BIGINT NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			service_type TEXT,
			INDEX(user_id, timestamp)
		)`,
		`CREATE TABLE IF NOT EXISTS system_metrics (
			id BIGSERIAL PRIMARY KEY,
			metric_name TEXT NOT NULL,
			metric_value NUMERIC NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			service_type TEXT,
			INDEX(metric_name, timestamp)
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id BIGSERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			reward BIGINT NOT NULL DEFAULT 0,
			type TEXT NOT NULL DEFAULT 'daily',
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_tasks (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			task_id BIGINT NOT NULL,
			status TEXT DEFAULT 'pending',
			completed_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX(user_id, task_id),
			INDEX(status)
		)`,
		`CREATE TABLE IF NOT EXISTS game_results (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			game_type TEXT NOT NULL,
			bet_amount BIGINT NOT NULL DEFAULT 0,
			result TEXT NOT NULL,
			win_amount BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX(user_id, game_type),
			INDEX(created_at)
		)`,
	}

	for _, query := range queries {
		if _, err := n.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	return nil
}

// LogActivity логирует активность пользователя
func (n *NeonManager) LogActivity(userID int64, action, details, serviceType string) error {
	_, err := n.DB.Exec(
		"INSERT INTO activity_logs (user_id, action, details, timestamp, service_type) VALUES (?, ?, ?, ?, ?)",
		userID, action, details, time.Now(), serviceType)
	return err
}

// LogTransaction логирует транзакцию
func (n *NeonManager) LogTransaction(userID int64, transactionType string, amount, balanceBefore, balanceAfter int64, serviceType string) error {
	_, err := n.DB.Exec(
		"INSERT INTO transaction_history (user_id, transaction_type, amount, balance_before, balance_after, timestamp, service_type) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, transactionType, amount, balanceBefore, balanceAfter, time.Now(), serviceType)
	return err
}

// LogMetric логирует системную метрику
func (n *NeonManager) LogMetric(metricName string, metricValue float64, serviceType string) error {
	_, err := n.DB.Exec(
		"INSERT INTO system_metrics (metric_name, metric_value, timestamp, service_type) VALUES (?, ?, ?, ?)",
		metricName, metricValue, time.Now(), serviceType)
	return err
}

// CreateTask создает задачу
func (n *NeonManager) CreateTask(title, description string, reward int64, taskType string) error {
	_, err := n.DB.Exec(
		"INSERT INTO tasks (title, description, reward, type, is_active, created_at) VALUES (?, ?, ?, ?, TRUE, ?)",
		title, description, reward, taskType, time.Now())
	return err
}

// GetTasks получает список задач
func (n *NeonManager) GetTasks(limit int) ([]*Task, error) {
	rows, err := n.DB.Query(
		"SELECT id, title, description, reward, type, is_active, created_at FROM tasks WHERE is_active = TRUE ORDER BY created_at DESC LIMIT ?",
		limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Reward, &task.Type, &task.IsActive, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// CreateUserTask создает выполнение задачи
func (n *NeonManager) CreateUserTask(userID, taskID int64) error {
	_, err := n.DB.Exec(
		"INSERT INTO user_tasks (user_id, task_id, status, created_at) VALUES (?, ?, 'pending', ?)",
		userID, taskID, time.Now())
	return err
}

// CompleteTask завершает задачу
func (n *NeonManager) CompleteTask(userID, taskID int64) error {
	_, err := n.DB.Exec(
		"UPDATE user_tasks SET status = 'completed', completed_at = ? WHERE user_id = ? AND task_id = ?",
		time.Now(), userID, taskID)
	return err
}

// GetUserTasks получает задачи пользователя
func (n *NeonManager) GetUserTasks(userID int64) ([]*UserTask, error) {
	rows, err := n.DB.Query(
		"SELECT id, user_id, task_id, status, completed_at, created_at FROM user_tasks WHERE user_id = ? ORDER BY created_at DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userTasks []*UserTask
	for rows.Next() {
		var userTask UserTask
		err := rows.Scan(&userTask.ID, &userTask.UserID, &userTask.TaskID, &userTask.Status, &userTask.CompletedAt, &userTask.CreatedAt)
		if err != nil {
			return nil, err
		}
		userTasks = append(userTasks, &userTask)
	}

	return userTasks, nil
}

// LogGameResult логирует результат игры
func (n *NeonManager) LogGameResult(userID int64, gameType string, betAmount, winAmount int64, result string) error {
	_, err := n.DB.Exec(
		"INSERT INTO game_results (user_id, game_type, bet_amount, result, win_amount, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		userID, gameType, betAmount, result, winAmount, time.Now())
	return err
}

// GetGameHistory получает историю игр
func (n *NeonManager) GetGameHistory(userID int64, limit int) ([]*GameResult, error) {
	rows, err := n.DB.Query(
		"SELECT id, user_id, game_type, bet_amount, result, win_amount, created_at FROM game_results WHERE user_id = ? ORDER BY created_at DESC LIMIT ?",
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*GameResult
	for rows.Next() {
		var result GameResult
		err := rows.Scan(&result.ID, &result.UserID, &result.GameType, &result.BetAmount, &result.Result, &result.WinAmount, &result.CreatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, &result)
	}

	return results, nil
}

// GetTransactionHistory получает историю транзакций
func (n *NeonManager) GetTransactionHistory(userID int64, limit int) ([]*TransactionHistory, error) {
	rows, err := n.DB.Query(
		"SELECT id, user_id, transaction_type, amount, balance_before, balance_after, timestamp, service_type FROM transaction_history WHERE user_id = ? ORDER BY timestamp DESC LIMIT ?",
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*TransactionHistory
	for rows.Next() {
		var transaction TransactionHistory
		err := rows.Scan(&transaction.ID, &transaction.UserID, &transaction.TransactionType, &transaction.Amount, &transaction.BalanceBefore, &transaction.BalanceAfter, &transaction.Timestamp, &transaction.ServiceType)
		if err != nil {
			return nil, err
		}
		history = append(history, &transaction)
	}

	return history, nil
}

// Ping проверяет соединение с Neon
func (n *NeonManager) Ping() error {
	if n.DB == nil {
		return fmt.Errorf("Neon database not initialized")
	}
	return n.DB.Ping()
}

// Close закрывает соединение с Neon
func (n *NeonManager) Close() error {
	if n.DB != nil {
		return n.DB.Close()
	}
	return nil
}
