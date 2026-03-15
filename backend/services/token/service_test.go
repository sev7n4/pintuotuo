package token

import (
  "context"
  "database/sql"
  "fmt"
  "math/rand"
  "os"
  "testing"
  "time"

  "github.com/pintuotuo/backend/cache"
  _ "github.com/lib/pq"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
)

var testDB *sql.DB
var cacheInitialized bool

func init() {
  // Initialize cache for all tests (runs once)
  os.Setenv("REDIS_URL", "redis://localhost:6380")
  err := cache.Init()
  if err == nil {
    cacheInitialized = true
  }
  rand.Seed(time.Now().UnixNano())
}

// generateTestEmail creates a unique email for test isolation
func generateTestEmail() string {
  return fmt.Sprintf("test_%d_%d@example.com", time.Now().UnixNano(), rand.Intn(100000))
}

// setupTestDB initializes test database
func setupTestDB(t *testing.T) *sql.DB {
  if testDB != nil {
    return testDB
  }

  // For testing, we use a real database connection
  // In production CI/CD, use a test database
  connStr := "postgres://pintuotuo:dev_password_123@localhost:5433/pintuotuo_db?sslmode=disable"
  db, err := sql.Open("postgres", connStr)
  require.NoError(t, err, "Failed to connect to test database")

  err = db.Ping()
  require.NoError(t, err, "Failed to ping test database")

  testDB = db
  return testDB
}

// TestGetBalance - balance operations
func TestGetBalance(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("Get balance for valid user", func(t *testing.T) {
    // Create test user and token record
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    // Initialize tokens
    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    balance, err := service.GetBalance(ctx, userID)
    require.NoError(t, err)
    assert.NotNil(t, balance)
    assert.Equal(t, userID, balance.UserID)
    assert.Equal(t, 0.0, balance.Balance)
  })

  t.Run("Get balance for non-existent user", func(t *testing.T) {
    balance, err := service.GetBalance(ctx, 99999)
    assert.Error(t, err)
    assert.Nil(t, balance)
    assert.Equal(t, ErrTokenNotFound, err)
  })

  t.Run("Get balance with invalid user ID", func(t *testing.T) {
    balance, err := service.GetBalance(ctx, -1)
    assert.Error(t, err)
    assert.Nil(t, balance)
    assert.Equal(t, ErrInvalidUserID, err)
  })
}

// TestGetTotalBalance
func TestGetTotalBalance(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("Get total balance for valid user", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    total, err := service.GetTotalBalance(ctx, userID)
    require.NoError(t, err)
    assert.Equal(t, 0.0, total)
  })
}

// TestRechargeTokens
func TestRechargeTokens(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("Valid recharge", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    req := &RechargeTokensRequest{
      UserID: userID,
      Amount: 100.0,
      Reason: "Payment for order",
    }

    balance, err := service.RechargeTokens(ctx, req)
    require.NoError(t, err)
    assert.NotNil(t, balance)
    assert.Equal(t, 100.0, balance.Balance)
    assert.Equal(t, 100.0, balance.TotalEarned)
  })

  t.Run("Recharge with zero amount", func(t *testing.T) {
    req := &RechargeTokensRequest{
      UserID: 1,
      Amount: 0,
      Reason: "Test",
    }
    balance, err := service.RechargeTokens(ctx, req)
    assert.Error(t, err)
    assert.Nil(t, balance)
    assert.Equal(t, ErrInvalidAmount, err)
  })

  t.Run("Recharge with empty reason", func(t *testing.T) {
    req := &RechargeTokensRequest{
      UserID: 1,
      Amount: 100,
      Reason: "",
    }
    _, err := service.RechargeTokens(ctx, req)
    assert.Error(t, err)
    assert.Equal(t, ErrInvalidReason, err)
  })

  t.Run("Recharge creates transaction log", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    _, err = service.RechargeTokens(ctx, &RechargeTokensRequest{
      UserID: userID,
      Amount: 50.0,
      Reason: "Test recharge",
    })
    require.NoError(t, err)

    transactions, err := service.GetTransactions(ctx, userID, nil)
    require.NoError(t, err)
    assert.Len(t, transactions, 1)
    assert.Equal(t, "recharge", transactions[0].Type)
    assert.Equal(t, 50.0, transactions[0].Amount)
  })
}

// TestConsumeTokens
func TestConsumeTokens(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("Valid consume", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    // Recharge first
    service.RechargeTokens(ctx, &RechargeTokensRequest{
      UserID: userID,
      Amount: 100.0,
      Reason: "Initial balance",
    })

    req := &ConsumeTokensRequest{
      UserID: userID,
      Amount: 30.0,
      Reason: "API call",
      Source: "api_call",
    }

    balance, err := service.ConsumeTokens(ctx, req)
    require.NoError(t, err)
    assert.Equal(t, 70.0, balance.Balance)
    assert.Equal(t, 30.0, balance.TotalUsed)
  })

  t.Run("Consume with insufficient balance", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    req := &ConsumeTokensRequest{
      UserID: userID,
      Amount: 100.0,
      Reason: "API call",
      Source: "api_call",
    }

    balance, err := service.ConsumeTokens(ctx, req)
    assert.Error(t, err)
    assert.Nil(t, balance)
    assert.Equal(t, ErrInsufficientBalance, err)
  })
}

// TestTransferTokens
func TestTransferTokens(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("Valid transfer", func(t *testing.T) {
    // Create sender
    var senderID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Sender", "hash", "user", "active").Scan(&senderID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", senderID)

    // Create recipient
    var recipientID int
    err = db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Recipient", "hash", "user", "active").Scan(&recipientID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", recipientID)

    // Initialize both
    service.InitializeUserTokens(ctx, senderID)
    service.InitializeUserTokens(ctx, recipientID)

    // Recharge sender
    service.RechargeTokens(ctx, &RechargeTokensRequest{
      UserID: senderID,
      Amount: 100.0,
      Reason: "Initial balance",
    })

    req := &TransferTokensRequest{
      SenderID:    senderID,
      RecipientID: recipientID,
      Amount:      50.0,
    }

    err = service.TransferTokens(ctx, req)
    require.NoError(t, err)

    // Verify sender balance
    sender, err := service.GetBalance(ctx, senderID)
    require.NoError(t, err)
    assert.Equal(t, 50.0, sender.Balance)

    // Verify recipient balance
    recipient, err := service.GetBalance(ctx, recipientID)
    require.NoError(t, err)
    assert.Equal(t, 50.0, recipient.Balance)
  })

  t.Run("Transfer to self", func(t *testing.T) {
    req := &TransferTokensRequest{
      SenderID:    1,
      RecipientID: 1,
      Amount:      50.0,
    }
    err := service.TransferTokens(ctx, req)
    assert.Error(t, err)
    assert.Equal(t, ErrTransferToSelf, err)
  })

  t.Run("Transfer with insufficient balance", func(t *testing.T) {
    var senderID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Sender", "hash", "user", "active").Scan(&senderID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", senderID)

    var recipientID int
    err = db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Recipient", "hash", "user", "active").Scan(&recipientID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", recipientID)

    service.InitializeUserTokens(ctx, senderID)
    service.InitializeUserTokens(ctx, recipientID)

    req := &TransferTokensRequest{
      SenderID:    senderID,
      RecipientID: recipientID,
      Amount:      100.0,
    }
    err = service.TransferTokens(ctx, req)
    assert.Error(t, err)
    assert.Equal(t, ErrInsufficientBalance, err)
  })
}

// TestGetConsumption
func TestGetConsumption(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("Get consumption with pagination", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    // Create multiple transactions
    for i := 0; i < 5; i++ {
      service.RechargeTokens(ctx, &RechargeTokensRequest{
        UserID: userID,
        Amount: 10.0,
        Reason: "Test",
      })
    }

    result, err := service.GetConsumption(ctx, userID, &GetConsumptionParams{
      PageSize: 2,
      Page:     1,
    })
    require.NoError(t, err)
    assert.Equal(t, 5, result.Total)
    assert.Len(t, result.Transactions, 2)
  })
}

// TestInitializeUserTokens
func TestInitializeUserTokens(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("Initialize new user tokens", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    balance, err := service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)
    assert.NotNil(t, balance)
    assert.Equal(t, userID, balance.UserID)
    assert.Equal(t, 0.0, balance.Balance)
    assert.Equal(t, 0.0, balance.TotalUsed)
    assert.Equal(t, 0.0, balance.TotalEarned)
  })
}

// TestAdjustBalance
func TestAdjustBalance(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("Adjust balance positive", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    balance, err := service.AdjustBalance(ctx, userID, 50.0, "Admin adjustment")
    require.NoError(t, err)
    assert.Equal(t, 50.0, balance.Balance)
    assert.Equal(t, 50.0, balance.TotalEarned)
  })

  t.Run("Adjust balance negative", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    _, err = service.AdjustBalance(ctx, userID, 100.0, "Initial")
    require.NoError(t, err)

    balance, err := service.AdjustBalance(ctx, userID, -30.0, "Penalty")
    require.NoError(t, err)
    assert.Equal(t, 70.0, balance.Balance)
    assert.Equal(t, 30.0, balance.TotalUsed)
  })
}

// TestIsBalanceSufficient
func TestIsBalanceSufficient(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("Sufficient balance returns true", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    service.RechargeTokens(ctx, &RechargeTokensRequest{
      UserID: userID,
      Amount: 100.0,
      Reason: "Test",
    })

    sufficient, err := service.IsBalanceSufficient(ctx, userID, 50.0)
    require.NoError(t, err)
    assert.True(t, sufficient)
  })

  t.Run("Insufficient balance returns false", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    sufficient, err := service.IsBalanceSufficient(ctx, userID, 100.0)
    require.NoError(t, err)
    assert.False(t, sufficient)
  })
}

// TestCaching
func TestCaching(t *testing.T) {
  db := setupTestDB(t)
  service := NewService(db, nil)
  ctx := context.Background()

  t.Run("GetBalance uses cache after first query", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    // First call - reads from DB
    balance1, err := service.GetBalance(ctx, userID)
    require.NoError(t, err)

    // Second call - should read from cache
    balance2, err := service.GetBalance(ctx, userID)
    require.NoError(t, err)

    assert.Equal(t, balance1.Balance, balance2.Balance)

    // Verify cache was set
    cacheKey := cache.TokenBalanceKey(userID)
    cached, err := cache.Get(ctx, cacheKey)
    assert.NoError(t, err)
    assert.NotEmpty(t, cached)
  })

  t.Run("Cache is invalidated after balance update", func(t *testing.T) {
    var userID int
    err := db.QueryRowContext(ctx,
      "INSERT INTO users (email, name, password_hash, role, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id",
      generateTestEmail(), "Test User", "hash", "user", "active").Scan(&userID)
    require.NoError(t, err)
    defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

    _, err = service.InitializeUserTokens(ctx, userID)
    require.NoError(t, err)

    // Get balance to populate cache
    service.GetBalance(ctx, userID)

    // Update balance
    service.RechargeTokens(ctx, &RechargeTokensRequest{
      UserID: userID,
      Amount: 100.0,
      Reason: "Test",
    })

    // Cache should be invalidated and fresh data returned
    balance, err := service.GetBalance(ctx, userID)
    require.NoError(t, err)
    assert.Equal(t, 100.0, balance.Balance)
  })
}
