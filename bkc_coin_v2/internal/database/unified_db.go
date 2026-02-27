package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UnifiedDB - унифицированная база данных
type UnifiedDB struct {
	Pool *pgxpool.Pool
}

// SystemState - состояние системы
type SystemState struct {
	TotalSupply       int64     `json:"total_supply"`
	ReserveSupply     int64     `json:"reserve_supply"`
	ReservedSupply    int64     `json:"reserved_supply"`
	InitialReserve    int64     `json:"initial_reserve"`
	AdminUserID       int64     `json:"admin_user_id"`
	AdminAllocated    int64     `json:"admin_allocated"`
	StartRateCoinsUSD int64     `json:"start_rate_coins_usd"`
	MinRateCoinsUSD   int64     `json:"min_rate_coins_usd"`
	ReferralStep      int64     `json:"referral_step"`
	ReferralBonus     int64     `json:"referral_bonus"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UserState - состояние пользователя
type UserState struct {
	UserID                     int64     `json:"user_id"`
	Username                   string    `json:"username"`
	FirstName                  string    `json:"first_name"`
	Balance                    int64     `json:"balance"`
	FrozenBalance              int64     `json:"frozen_balance"`
	TapsTotal                  int64     `json:"taps_total"`
	Energy                     float64   `json:"energy"`
	EnergyMax                  float64   `json:"energy_max"`
	EnergyUpdatedAt            time.Time `json:"energy_updated_at"`
	EnergyBoostUntil           time.Time `json:"energy_boost_until"`
	EnergyBoostRegenMultiplier float64   `json:"energy_boost_regen_multiplier"`
	EnergyBoostMaxMultiplier   float64   `json:"energy_boost_max_multiplier"`
	ReferralsCount             int64     `json:"referrals_count"`
	ReferralBonusTotal         int64     `json:"referral_bonus_total"`
}

// NFT - информация о NFT
type NFT struct {
	NFTID       int64     `json:"nft_id"`
	Title       string    `json:"title"`
	ImageURL    string    `json:"image_url"`
	PriceCoins  int64     `json:"price_coins"`
	SupplyLeft  int64     `json:"supply_left"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserNFT - NFT пользователя
type UserNFT struct {
	NFTID  string `json:"nft_id"`
	Title  string `json:"title"`
	ImageURL string `json:"image_url"`
	Qty    int64  `json:"qty"`
}

// TapEvent - событие тапа
type TapEvent struct {
	UserID    int64     `json:"user_id"`
	Req       string    `json:"req"`
	Timestamp time.Time `json:"timestamp"`
}

// UserTapAggregate - агрегат тапов пользователя
type UserTapAggregate struct {
	UserID      int64 `json:"user_id"`
	Taps        int64 `json:"taps"`
	EnergySpent int64 `json:"energy_spent"`
}

// DailyTapAggregate - дневной агрегат тапов
type DailyTapAggregate struct {
	Date        time.Time `json:"date"`
	Taps        int64     `json:"taps"`
	EnergySpent int64     `json:"energy_spent"`
}

// Deposit - депозит
type Deposit struct {
	DepositID   int64     `json:"deposit_id"`
	UserID      int64     `json:"user_id"`
	TxHash      string    `json:"tx_hash"`
	AmountUSD   float64   `json:"amount_usd"`
	Currency    string    `json:"currency"`
	Coins       int64     `json:"coins"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ApprovedAt  *time.Time `json:"approved_at"`
	ApprovedBy  *int64    `json:"approved_by"`
}

// BankLoan - банковский кредит
type BankLoan struct {
	LoanID     int64     `json:"loan_id"`
	UserID     int64     `json:"user_id"`
	Principal  int64     `json:"principal"`
	Interest   int64     `json:"interest"`
	TotalDue   int64     `json:"total_due"`
	TermDays   int64     `json:"term_days"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	DueAt      time.Time `json:"due_at"`
	ClosedAt   *time.Time `json:"closed_at"`
}

// P2PLoan - P2P кредит
type P2PLoan struct {
	LoanID       int64     `json:"loan_id"`
	LenderID     int64     `json:"lender_id"`
	BorrowerID   int64     `json:"borrower_id"`
	Principal    int64     `json:"principal"`
	Interest     int64     `json:"interest"`
	TotalDue     int64     `json:"total_due"`
	InterestBP   int64     `json:"interest_bp"`
	TermDays     int64     `json:"term_days"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	AcceptedAt   *time.Time `json:"accepted_at"`
	DueAt        time.Time `json:"due_at"`
	ClosedAt     *time.Time `json:"closed_at"`
}

// Connect - подключение к базе данных
func Connect(ctx context.Context, databaseURL string) (*UnifiedDB, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}
	
	cfg.MaxConns = 20
	cfg.MinConns = 5
	cfg.MaxConnLifetime = time.Hour
	cfg.HealthCheckPeriod = time.Minute * 5
	
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	
	return &UnifiedDB{Pool: pool}, nil
}

// Close - закрытие соединения
func (db *UnifiedDB) Close() {
	db.Pool.Close()
}

// WithTx - выполнение транзакции
func (db *UnifiedDB) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	
	if err := fn(tx); err != nil {
		return err
	}
	
	return tx.Commit(ctx)
}

// GetUserState - получение состояния пользователя
func (db *UnifiedDB) GetUserState(ctx context.Context, userID int64) (*UserState, error) {
	var user UserState
	err := db.Pool.QueryRow(ctx, `
		SELECT user_id, username, first_name, balance, frozen_balance, 
		       taps_total, energy, energy_max, energy_updated_at,
		       energy_boost_until, energy_boost_regen_multiplier, energy_boost_max_multiplier,
		       referrals_count, referral_bonus_total
		FROM users WHERE user_id = $1
	`, userID).Scan(
		&user.UserID, &user.Username, &user.FirstName, &user.Balance, &user.FrozenBalance,
		&user.TapsTotal, &user.Energy, &user.EnergyMax, &user.EnergyUpdatedAt,
		&user.EnergyBoostUntil, &user.EnergyBoostRegenMultiplier, &user.EnergyBoostMaxMultiplier,
		&user.ReferralsCount, &user.ReferralBonusTotal,
	)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user state: %w", err)
	}
	
	return &user, nil
}

// UpdateUserBalance - обновление баланса пользователя
func (db *UnifiedDB) UpdateUserBalance(ctx context.Context, userID int64, delta int64) error {
	result, err := db.Pool.Exec(ctx, 
		"UPDATE users SET balance = balance + $1 WHERE user_id = $2", 
		delta, userID)
	
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// UpdateUserEnergy - обновление энергии пользователя
func (db *UnifiedDB) UpdateUserEnergy(ctx context.Context, userID int64, energy float64) error {
	result, err := db.Pool.Exec(ctx, 
		"UPDATE users SET energy = $1, energy_updated_at = NOW() WHERE user_id = $2", 
		energy, userID)
	
	if err != nil {
		return fmt.Errorf("failed to update energy: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// CreateUser - создание нового пользователя
func (db *UnifiedDB) CreateUser(ctx context.Context, user *UserState) error {
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO users (user_id, username, first_name, balance, energy, energy_max, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING user_id
	`, user.UserID, user.Username, user.FirstName, user.Balance, user.Energy, user.EnergyMax).Scan(&user.UserID)
	
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetSystemState - получение состояния системы
func (db *UnifiedDB) GetSystemState(ctx context.Context) (*SystemState, error) {
	var state SystemState
	err := db.Pool.QueryRow(ctx, `
		SELECT total_supply, reserve_supply, reserved_supply, initial_reserve,
		       admin_user_id, admin_allocated, start_rate_coins_usd, min_rate_coins_usd,
		       referral_step, referral_bonus, created_at, updated_at
		FROM system_state
	`).Scan(
		&state.TotalSupply, &state.ReserveSupply, &state.ReservedSupply, &state.InitialReserve,
		&state.AdminUserID, &state.AdminAllocated, &state.StartRateCoinsUSD, &state.MinRateCoinsUSD,
		&state.ReferralStep, &state.ReferralBonus, &state.CreatedAt, &state.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("system state not found")
		}
		return nil, fmt.Errorf("failed to get system state: %w", err)
	}
	
	return &state, nil
}

// ApplyTapEvents - применение событий тапов
func (db *UnifiedDB) ApplyTapEvents(ctx context.Context, events []TapEvent) error {
	if len(events) == 0 {
		return nil
	}
	
	userIDs := make([]int64, 0, len(events))
	for _, ev := range events {
		userIDs = append(userIDs, ev.UserID)
	}
	
	return db.WithTx(ctx, func(tx pgx.Tx) error {
		for _, ev := range events {
			_, err := tx.Exec(ctx, 
				"INSERT INTO tap_events (user_id, req, timestamp) VALUES ($1, $2, $3)",
				ev.UserID, ev.Req, ev.Timestamp)
			if err != nil {
				return fmt.Errorf("failed to insert tap event: %w", err)
			}
		}
		return nil
	})
}

// GetNFTs - получение списка NFT
func (db *UnifiedDB) GetNFTs(ctx context.Context, limit int) ([]NFT, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT nft_id, title, image_url, price_coins, supply_left, created_at
		FROM nfts
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get NFTs: %w", err)
	}
	defer rows.Close()
	
	var nfts []NFT
	for rows.Next() {
		var nft NFT
		if err := rows.Scan(&nft.NFTID, &nft.Title, &nft.ImageURL, &nft.PriceCoins, &nft.SupplyLeft, &nft.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan NFT: %w", err)
		}
		nfts = append(nfts, nft)
	}
	
	return nfts, nil
}

// GetUserNFTs - получение NFT пользователя
func (db *UnifiedDB) GetUserNFTs(ctx context.Context, userID int64, limit int) ([]UserNFT, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT n.nft_id, n.title, n.image_url, u.qty
		FROM user_nfts u
		JOIN nfts n ON u.nft_id = n.nft_id
		WHERE u.user_id = $1
		ORDER BY u.qty DESC, u.nft_id DESC
		LIMIT $2
	`, userID, limit)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user NFTs: %w", err)
	}
	defer rows.Close()
	
	var userNFTs []UserNFT
	for rows.Next() {
		var unft UserNFT
		if err := rows.Scan(&unft.NFTID, &unft.Title, &unft.ImageURL, &unft.Qty); err != nil {
			return nil, fmt.Errorf("failed to scan user NFT: %w", err)
		}
		userNFTs = append(userNFTs, unft)
	}
	
	return userNFTs, nil
}

// CreateDeposit - создание депозита
func (db *UnifiedDB) CreateDeposit(ctx context.Context, userID int64, txHash string, amountUSD float64, currency string, coins int64) (int64, error) {
	txHash = strings.TrimSpace(txHash)
	currency = strings.ToUpper(strings.TrimSpace(currency))
	
	if userID <= 0 || txHash == "" || amountUSD <= 0 || coins <= 0 || currency == "" {
		return 0, fmt.Errorf("invalid parameters")
	}
	
	var depositID int64
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO deposits (user_id, tx_hash, amount_usd, currency, coins, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', NOW())
		RETURNING deposit_id
	`, userID, txHash, amountUSD, currency, coins).Scan(&depositID)
	
	if err != nil {
		return 0, fmt.Errorf("failed to create deposit: %w", err)
	}
	
	return depositID, nil
}

// GetDeposits - получение депозитов
func (db *UnifiedDB) GetDeposits(ctx context.Context, status string, limit int) ([]Deposit, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT deposit_id, user_id, tx_hash, amount_usd, currency, coins, status, created_at, approved_at, approved_by
		FROM deposits
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, status, limit)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get deposits: %w", err)
	}
	defer rows.Close()
	
	var deposits []Deposit
	for rows.Next() {
		var deposit Deposit
		if err := rows.Scan(&deposit.DepositID, &deposit.UserID, &deposit.TxHash, &deposit.AmountUSD, 
			&deposit.Currency, &deposit.Coins, &deposit.Status, &deposit.CreatedAt, 
			&deposit.ApprovedAt, &deposit.ApprovedBy); err != nil {
			return nil, fmt.Errorf("failed to scan deposit: %w", err)
		}
		deposits = append(deposits, deposit)
	}
	
	return deposits, nil
}

// HealthCheck - проверка здоровья базы данных
func (db *UnifiedDB) HealthCheck(ctx context.Context) error {
	var result int
	err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	
	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}
	
	return nil
}

// GetStats - получение статистики
func (db *UnifiedDB) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Количество пользователей
	var userCount int64
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}
	stats["users"] = userCount
	
	// Общий баланс
	var totalBalance int64
	err = db.Pool.QueryRow(ctx, "SELECT COALESCE(SUM(balance), 0) FROM users").Scan(&totalBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to get total balance: %w", err)
	}
	stats["total_balance"] = totalBalance
	
	// Количество NFT
	var nftCount int64
	err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM nfts").Scan(&nftCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT count: %w", err)
	}
	stats["nfts"] = nftCount
	
	return stats, nil
}

// toJSON - конвертация в JSON
func toJSON(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}
