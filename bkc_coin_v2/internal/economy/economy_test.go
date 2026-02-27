package economy

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestEconomySystem tests the economy system
func TestEconomySystem(t *testing.T) {
	// Setup test database
	db, err := sqlx.Connect("postgres", "postgres://test:test@localhost/bkc_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
		return
	}
	defer db.Close()

	// Create test tables
	err = createTestTables(db)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Initialize economy system
	economy := NewOptimizedEconomySystem(db)

	// Test user tap processing
	t.Run("ProcessUserTap", func(t *testing.T) {
		userID := int64(12345)
		
		// Create test user
		err := createTestUser(db, userID)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		// Process tap
		result, err := economy.ProcessUserTap(userID, 1, 0)
		if err != nil {
			t.Fatalf("ProcessUserTap failed: %v", err)
		}

		// Verify result
		if result.Success != true {
			t.Error("Expected success")
		}

		if result.Balance <= 0 {
			t.Error("Expected positive balance")
		}

		if result.Energy < 0 {
			t.Error("Expected non-negative energy")
		}
	})

	// Test referral processing
	t.Run("ProcessReferral", func(t *testing.T) {
		referrerID := int64(12345)
		referralID := int64(67890)

		// Process referral
		result, err := economy.ProcessReferral(referrerID, referralID)
		if err != nil {
			t.Fatalf("ProcessReferral failed: %v", err)
		}

		// Verify result
		if result.Success != true {
			t.Error("Expected success")
		}

		if result.ReferrerReward <= 0 {
			t.Error("Expected positive referrer reward")
		}

		if result.ReferralBonus <= 0 {
			t.Error("Expected positive referral bonus")
		}
	})

	// Test burn calculation
	t.Run("CalculateBurnAmount", func(t *testing.T) {
		amount := 1000.0
		supply := 100000000.0
		price := 0.001
		volatility := 0.1

		burnAmount := economy.CalculateBurnAmount(amount, supply, price, volatility)

		if burnAmount <= 0 {
			t.Error("Expected positive burn amount")
		}

		if burnAmount > amount*0.05 { // Max 5%
			t.Error("Burn amount exceeds maximum")
		}
	})

	// Test emission calculation
	t.Run("CalculateEmission", func(t *testing.T) {
		activeUsers := 10000
		currentPrice := 0.001
		targetPrice := 0.001
		currentSupply := 50000000.0

		emission := economy.CalculateEmission(activeUsers, currentPrice, targetPrice, currentSupply)

		if emission <= 0 {
			t.Error("Expected positive emission")
		}

		if emission > 300000 { // Daily max
			t.Error("Emission exceeds daily maximum")
		}
	})

	// Test stabilization fund
	t.Run("StabilizationFund", func(t *testing.T) {
		currentPrice := 0.0008 // Below target
		targetPrice := 0.001
		volatility := 0.2

		action := economy.CalculateStabilizationAction(currentPrice, targetPrice, volatility)

		if action.Type == "sell" {
			t.Error("Expected buy action when price is below target")
		}

		if action.Amount <= 0 {
			t.Error("Expected positive action amount")
		}
	})

	// Test time to exhaustion
	t.Run("TimeToExhaustion", func(t *testing.T) {
		currentSupply := 50000000.0
		dailyEmission := 50000.0
		dailyBurn := 5000.0

		days := economy.CalculateTimeToExhaustion(currentSupply, dailyEmission, dailyBurn)

		if days <= 0 {
			t.Error("Expected positive days to exhaustion")
		}

		expectedDays := (1000000000 - currentSupply) / (dailyEmission - dailyBurn)
		if abs(days-expectedDays) > 1 {
			t.Errorf("Expected %.0f days, got %.0f", expectedDays, days)
		}
	})
}

// TestEconomyEdgeCases tests edge cases
func TestEconomyEdgeCases(t *testing.T) {
	db, err := sqlx.Connect("postgres", "postgres://test:test@localhost/bkc_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
		return
	}
	defer db.Close()

	err = createTestTables(db)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	economy := NewOptimizedEconomySystem(db)

	// Test zero energy
	t.Run("ZeroEnergy", func(t *testing.T) {
		userID := int64(12346)
		err := createTestUserWithEnergy(db, userID, 0)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		result, err := economy.ProcessUserTap(userID, 1, 0)
		if err != nil {
			t.Fatalf("ProcessUserTap failed: %v", err)
		}

		if result.Success {
			t.Error("Expected failure with zero energy")
		}
	})

	// Test maximum daily limit
	t.Run("DailyLimit", func(t *testing.T) {
		userID := int64(12347)
		err := createTestUserWithDailyEarned(db, userID, 300)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		result, err := economy.ProcessUserTap(userID, 1, 0)
		if err != nil {
			t.Fatalf("ProcessUserTap failed: %v", err)
		}

		if result.Success {
			t.Error("Expected failure at daily limit")
		}
	})

	// Test negative amounts
	t.Run("NegativeAmounts", func(t *testing.T) {
		burnAmount := economy.CalculateBurnAmount(-1000, 100000000, 0.001, 0.1)
		if burnAmount != 0 {
			t.Error("Expected zero burn amount for negative input")
		}

		emission := economy.CalculateEmission(-1000, 0.001, 0.001, 50000000)
		if emission != 0 {
			t.Error("Expected zero emission for negative user count")
		}
	})
}

// TestEconomyPerformance tests performance
func TestEconomyPerformance(t *testing.T) {
	db, err := sqlx.Connect("postgres", "postgres://test:test@localhost/bkc_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
		return
	}
	defer db.Close()

	err = createTestTables(db)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	economy := NewOptimizedEconomySystem(db)

	// Test concurrent tap processing
	t.Run("ConcurrentTaps", func(t *testing.T) {
		userID := int64(12348)
		err := createTestUser(db, userID)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		const numGoroutines = 100
		const tapsPerGoroutine = 10

		start := time.Now()

		done := make(chan bool, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < tapsPerGoroutine; j++ {
					_, err := economy.ProcessUserTap(userID, 1, 0)
					if err != nil {
						t.Errorf("ProcessUserTap failed: %v", err)
					}
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		duration := time.Since(start)
		t.Logf("Processed %d taps in %v (%.2f taps/sec)", 
			numGoroutines*tapsPerGoroutine, duration, 
			float64(numGoroutines*tapsPerGoroutine)/duration.Seconds())

		if duration > 5*time.Second {
			t.Error("Performance too slow")
		}
	})
}

// TestEconomyIntegration tests integration with other systems
func TestEconomyIntegration(t *testing.T) {
	db, err := sqlx.Connect("postgres", "postgres://test:test@localhost/bkc_test?sslmode=disable")
	if err != nil {
		t.Skip("Skipping test: database not available")
		return
	}
	defer db.Close()

	err = createTestTables(db)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	economy := NewOptimizedEconomySystem(db)

	// Test full user journey
	t.Run("UserJourney", func(t *testing.T) {
		userID := int64(12349)
		referrerID := int64(12350)

		// Create users
		err := createTestUser(db, userID)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		err = createTestUser(db, referrerID)
		if err != nil {
			t.Fatalf("Failed to create referrer: %v", err)
		}

		// Process referral
		referralResult, err := economy.ProcessReferral(referrerID, userID)
		if err != nil {
			t.Fatalf("ProcessReferral failed: %v", err)
		}

		if !referralResult.Success {
			t.Error("Referral should succeed")
		}

		// Process taps
		for i := 0; i < 10; i++ {
			result, err := economy.ProcessUserTap(userID, 1, 0)
			if err != nil {
				t.Fatalf("ProcessUserTap failed: %v", err)
			}

			if !result.Success {
				t.Error("Tap should succeed")
			}
		}

		// Check final state
		status, err := economy.GetUserStatus(userID)
		if err != nil {
			t.Fatalf("GetUserStatus failed: %v", err)
		}

		if status.Balance <= 0 {
			t.Error("User should have positive balance")
		}

		if status.Energy < 0 {
			t.Error("Energy should not be negative")
		}
	})
}

// Helper functions

func createTestTables(db *sqlx.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY,
			balance DECIMAL(20,8) DEFAULT 0,
			energy INTEGER DEFAULT 3000,
			daily_earned DECIMAL(20,8) DEFAULT 0,
			last_energy_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS referrals (
			id SERIAL PRIMARY KEY,
			referrer_id BIGINT NOT NULL,
			referral_id BIGINT NOT NULL,
			is_active BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(referrer_id, referral_id)
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			type VARCHAR(50) NOT NULL,
			amount DECIMAL(20,8) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func createTestUser(db *sqlx.DB, userID int64) error {
	query := `
		INSERT INTO users (id, balance, energy, daily_earned, last_energy_update)
		VALUES ($1, 1000.0, 3000, 0, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := db.Exec(query, userID)
	return err
}

func createTestUserWithEnergy(db *sqlx.DB, userID int64, energy int) error {
	query := `
		INSERT INTO users (id, balance, energy, daily_earned, last_energy_update)
		VALUES ($1, 1000.0, $2, 0, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := db.Exec(query, userID, energy)
	return err
}

func createTestUserWithDailyEarned(db *sqlx.DB, userID int64, dailyEarned float64) error {
	query := `
		INSERT INTO users (id, balance, energy, daily_earned, last_energy_update)
		VALUES ($1, 1000.0, 3000, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := db.Exec(query, userID, dailyEarned)
	return err
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// BenchmarkEconomySystem benchmarks economy operations
func BenchmarkEconomySystem(b *testing.B) {
	db, err := sqlx.Connect("postgres", "postgres://test:test@localhost/bkc_test?sslmode=disable")
	if err != nil {
		b.Skip("Skipping benchmark: database not available")
		return
	}
	defer db.Close()

	err = createTestTables(db)
	if err != nil {
		b.Fatalf("Failed to create test tables: %v", err)
	}

	economy := NewOptimizedEconomySystem(db)
	userID := int64(12351)

	err = createTestUser(db, userID)
	if err != nil {
		b.Fatalf("Failed to create test user: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := economy.ProcessUserTap(userID, 1, 0)
			if err != nil {
				b.Errorf("ProcessUserTap failed: %v", err)
			}
		}
	})
}
