package token

import (
  "context"
  "database/sql"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "os"

  "github.com/pintuotuo/backend/cache"
  apperrors "github.com/pintuotuo/backend/errors"
)

// Service defines the token service interface
type Service interface {
  // Balance operations
  GetBalance(ctx context.Context, userID int) (*TokenBalance, error)
  GetTotalBalance(ctx context.Context, userID int) (float64, error)

  // Consumption tracking
  GetConsumption(ctx context.Context, userID int, params *GetConsumptionParams) (*ConsumptionResult, error)
  GetTransactions(ctx context.Context, userID int, params *TransactionParams) ([]TokenTransaction, error)

  // Token operations
  RechargeTokens(ctx context.Context, req *RechargeTokensRequest) (*TokenBalance, error)
  ConsumeTokens(ctx context.Context, req *ConsumeTokensRequest) (*TokenBalance, error)
  TransferTokens(ctx context.Context, req *TransferTokensRequest) error

  // Internal operations
  InitializeUserTokens(ctx context.Context, userID int) (*TokenBalance, error)
  AdjustBalance(ctx context.Context, userID int, delta float64, reason string) (*TokenBalance, error)
  IsBalanceSufficient(ctx context.Context, userID int, requiredAmount float64) (bool, error)
}

// service implements the Service interface
type service struct {
  db  *sql.DB
  log *log.Logger
}

// NewService creates a new token service
func NewService(db *sql.DB, logger *log.Logger) Service {
  if logger == nil {
    logger = log.New(os.Stderr, "[TokenService] ", log.LstdFlags)
  }

  return &service{
    db:  db,
    log: logger,
  }
}

// GetBalance retrieves user token balance with caching
func (s *service) GetBalance(ctx context.Context, userID int) (*TokenBalance, error) {
  if userID <= 0 {
    return nil, ErrInvalidUserID
  }

  // Try cache first
  cacheKey := cache.TokenBalanceKey(userID)
  if cachedBalance, err := cache.Get(ctx, cacheKey); err == nil {
    var balance TokenBalance
    if err := json.Unmarshal([]byte(cachedBalance), &balance); err == nil {
      return &balance, nil
    }
  }

  // Query database
  var balance TokenBalance
  err := s.db.QueryRowContext(
    ctx,
    `SELECT id, user_id, balance, total_used, total_earned, created_at, updated_at
     FROM tokens WHERE user_id = $1`,
    userID,
  ).Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.TotalUsed, &balance.TotalEarned, &balance.CreatedAt, &balance.UpdatedAt)

  if err != nil {
    if err == sql.ErrNoRows {
      return nil, ErrTokenNotFound
    }
    s.log.Printf("GetBalance query failed: %v", err)
    return nil, wrapError("GetBalance", "query", err)
  }

  // Cache the result
  if balanceJSON, err := json.Marshal(balance); err == nil {
    cache.Set(ctx, cacheKey, string(balanceJSON), cache.TokenBalanceTTL)
  }

  return &balance, nil
}

// GetTotalBalance returns the total balance amount for a user
func (s *service) GetTotalBalance(ctx context.Context, userID int) (float64, error) {
  balance, err := s.GetBalance(ctx, userID)
  if err != nil {
    return 0, err
  }
  return balance.Balance, nil
}

// GetConsumption retrieves user token consumption with pagination
func (s *service) GetConsumption(ctx context.Context, userID int, params *GetConsumptionParams) (*ConsumptionResult, error) {
  if userID <= 0 {
    return nil, ErrInvalidUserID
  }

  if params == nil {
    params = &GetConsumptionParams{PageSize: 20, Page: 1}
  }

  if params.PageSize <= 0 {
    params.PageSize = 20
  }
  if params.Page < 1 {
    params.Page = 1
  }

  offset := (params.Page - 1) * params.PageSize

  // Get total count
  var total int
  err := s.db.QueryRowContext(
    ctx,
    "SELECT COUNT(*) FROM token_transactions WHERE user_id = $1",
    userID,
  ).Scan(&total)

  if err != nil {
    s.log.Printf("GetConsumption count failed: %v", err)
    return nil, wrapError("GetConsumption", "count", err)
  }

  // Query transactions with pagination
  query := `SELECT id, user_id, type, amount, reason, order_id, created_at
    FROM token_transactions
    WHERE user_id = $1
    ORDER BY created_at DESC
    LIMIT $2 OFFSET $3`

  rows, err := s.db.QueryContext(ctx, query, userID, params.PageSize, offset)
  if err != nil {
    s.log.Printf("GetConsumption query failed: %v", err)
    return nil, wrapError("GetConsumption", "query", err)
  }
  defer rows.Close()

  var transactions []TokenTransaction
  for rows.Next() {
    var tx TokenTransaction
    err := rows.Scan(&tx.ID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Reason, &tx.OrderID, &tx.CreatedAt)
    if err != nil {
      s.log.Printf("GetConsumption scan failed: %v", err)
      return nil, wrapError("GetConsumption", "scan", err)
    }
    transactions = append(transactions, tx)
  }

  if err = rows.Err(); err != nil {
    s.log.Printf("GetConsumption rows error: %v", err)
    return nil, wrapError("GetConsumption", "rows", err)
  }

  return &ConsumptionResult{
    Total:        total,
    Page:         params.Page,
    PageSize:     params.PageSize,
    Transactions: transactions,
  }, nil
}

// GetTransactions retrieves transactions with optional filtering
func (s *service) GetTransactions(ctx context.Context, userID int, params *TransactionParams) ([]TokenTransaction, error) {
  if userID <= 0 {
    return nil, ErrInvalidUserID
  }

  if params == nil {
    params = &TransactionParams{Limit: 100}
  }

  if params.Limit <= 0 {
    params.Limit = 100
  }

  query := "SELECT id, user_id, type, amount, reason, order_id, created_at FROM token_transactions WHERE user_id = $1"
  args := []interface{}{userID}

  if params.Type != "" {
    query += " AND type = $2"
    args = append(args, params.Type)
  }

  query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1)
  args = append(args, params.Limit)

  rows, err := s.db.QueryContext(ctx, query, args...)
  if err != nil {
    s.log.Printf("GetTransactions query failed: %v", err)
    return nil, wrapError("GetTransactions", "query", err)
  }
  defer rows.Close()

  var transactions []TokenTransaction
  for rows.Next() {
    var tx TokenTransaction
    err := rows.Scan(&tx.ID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Reason, &tx.OrderID, &tx.CreatedAt)
    if err != nil {
      s.log.Printf("GetTransactions scan failed: %v", err)
      return nil, wrapError("GetTransactions", "scan", err)
    }
    transactions = append(transactions, tx)
  }

  if err = rows.Err(); err != nil {
    s.log.Printf("GetTransactions rows error: %v", err)
    return nil, wrapError("GetTransactions", "rows", err)
  }

  return transactions, nil
}

// RechargeTokens adds tokens to user balance with transaction logging
func (s *service) RechargeTokens(ctx context.Context, req *RechargeTokensRequest) (*TokenBalance, error) {
  if req == nil {
    return nil, errInvalidRequest()
  }

  if req.UserID <= 0 {
    return nil, ErrInvalidUserID
  }

  if req.Amount <= 0 {
    return nil, ErrInvalidAmount
  }

  if req.Reason == "" {
    return nil, ErrInvalidReason
  }

  // Start transaction for atomicity
  tx, err := s.db.BeginTx(ctx, nil)
  if err != nil {
    s.log.Printf("RechargeTokens begin transaction failed: %v", err)
    return nil, ErrTransactionFailed
  }
  defer tx.Rollback()

  // Update balance (using SELECT FOR UPDATE for concurrency safety)
  var balance TokenBalance
  err = tx.QueryRowContext(
    ctx,
    `UPDATE tokens SET balance = balance + $1, total_earned = total_earned + $2, updated_at = NOW()
     WHERE user_id = $3
     RETURNING id, user_id, balance, total_used, total_earned, created_at, updated_at`,
    req.Amount, req.Amount, req.UserID,
  ).Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.TotalUsed, &balance.TotalEarned, &balance.CreatedAt, &balance.UpdatedAt)

  if err != nil {
    if err == sql.ErrNoRows {
      tx.Rollback()
      return nil, ErrTokenNotFound
    }
    s.log.Printf("RechargeTokens update failed: %v", err)
    return nil, ErrTransactionFailed
  }

  // Record transaction
  _, err = tx.ExecContext(
    ctx,
    `INSERT INTO token_transactions (user_id, type, amount, reason, created_at)
     VALUES ($1, $2, $3, $4, NOW())`,
    req.UserID, "recharge", req.Amount, req.Reason,
  )

  if err != nil {
    s.log.Printf("RechargeTokens insert transaction failed: %v", err)
    return nil, ErrTransactionFailed
  }

  if err := tx.Commit(); err != nil {
    s.log.Printf("RechargeTokens commit failed: %v", err)
    return nil, ErrTransactionFailed
  }

  // Invalidate cache
  cacheKey := cache.TokenBalanceKey(req.UserID)
  cache.Delete(ctx, cacheKey)

  return &balance, nil
}

// ConsumeTokens deducts tokens from user balance
func (s *service) ConsumeTokens(ctx context.Context, req *ConsumeTokensRequest) (*TokenBalance, error) {
  if req == nil {
    return nil, errInvalidRequest()
  }

  if req.UserID <= 0 {
    return nil, ErrInvalidUserID
  }

  if req.Amount <= 0 {
    return nil, ErrInvalidAmount
  }

  if req.Reason == "" {
    return nil, ErrInvalidReason
  }

  // Start transaction
  tx, err := s.db.BeginTx(ctx, nil)
  if err != nil {
    s.log.Printf("ConsumeTokens begin transaction failed: %v", err)
    return nil, ErrTransactionFailed
  }
  defer tx.Rollback()

  // Check balance with SELECT FOR UPDATE for concurrency safety
  var currentBalance float64
  err = tx.QueryRowContext(
    ctx,
    "SELECT balance FROM tokens WHERE user_id = $1 FOR UPDATE",
    req.UserID,
  ).Scan(&currentBalance)

  if err != nil {
    if err == sql.ErrNoRows {
      tx.Rollback()
      return nil, ErrTokenNotFound
    }
    s.log.Printf("ConsumeTokens select failed: %v", err)
    return nil, ErrTransactionFailed
  }

  if currentBalance < req.Amount {
    tx.Rollback()
    return nil, ErrInsufficientBalance
  }

  // Update balance
  var balance TokenBalance
  err = tx.QueryRowContext(
    ctx,
    `UPDATE tokens SET balance = balance - $1, total_used = total_used + $2, updated_at = NOW()
     WHERE user_id = $3
     RETURNING id, user_id, balance, total_used, total_earned, created_at, updated_at`,
    req.Amount, req.Amount, req.UserID,
  ).Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.TotalUsed, &balance.TotalEarned, &balance.CreatedAt, &balance.UpdatedAt)

  if err != nil {
    s.log.Printf("ConsumeTokens update failed: %v", err)
    return nil, ErrTransactionFailed
  }

  // Record transaction
  _, err = tx.ExecContext(
    ctx,
    `INSERT INTO token_transactions (user_id, type, amount, reason, created_at)
     VALUES ($1, $2, $3, $4, NOW())`,
    req.UserID, "consume", req.Amount, req.Reason,
  )

  if err != nil {
    s.log.Printf("ConsumeTokens insert transaction failed: %v", err)
    return nil, ErrTransactionFailed
  }

  if err := tx.Commit(); err != nil {
    s.log.Printf("ConsumeTokens commit failed: %v", err)
    return nil, ErrTransactionFailed
  }

  // Invalidate cache
  cacheKey := cache.TokenBalanceKey(req.UserID)
  cache.Delete(ctx, cacheKey)

  return &balance, nil
}

// TransferTokens transfers tokens between users atomically
func (s *service) TransferTokens(ctx context.Context, req *TransferTokensRequest) error {
  if req == nil {
    return errInvalidRequest()
  }

  if req.SenderID <= 0 || req.RecipientID <= 0 {
    return ErrInvalidUserID
  }

  if req.Amount <= 0 {
    return ErrInvalidAmount
  }

  if req.SenderID == req.RecipientID {
    return ErrTransferToSelf
  }

  // Start transaction
  tx, err := s.db.BeginTx(ctx, nil)
  if err != nil {
    s.log.Printf("TransferTokens begin transaction failed: %v", err)
    return ErrTransactionFailed
  }
  defer tx.Rollback()

  // Check sender balance with SELECT FOR UPDATE
  var senderBalance float64
  err = tx.QueryRowContext(
    ctx,
    "SELECT balance FROM tokens WHERE user_id = $1 FOR UPDATE",
    req.SenderID,
  ).Scan(&senderBalance)

  if err != nil {
    if err == sql.ErrNoRows {
      tx.Rollback()
      return ErrTokenNotFound
    }
    s.log.Printf("TransferTokens select sender failed: %v", err)
    return ErrTransactionFailed
  }

  if senderBalance < req.Amount {
    tx.Rollback()
    return ErrInsufficientBalance
  }

  // Verify recipient exists and lock for update
  var recipientExists bool
  err = tx.QueryRowContext(
    ctx,
    "SELECT EXISTS(SELECT 1 FROM tokens WHERE user_id = $1) FOR UPDATE",
    req.RecipientID,
  ).Scan(&recipientExists)

  if err != nil || !recipientExists {
    tx.Rollback()
    return ErrRecipientNotFound
  }

  // Deduct from sender
  _, err = tx.ExecContext(
    ctx,
    "UPDATE tokens SET balance = balance - $1, total_used = total_used + $2, updated_at = NOW() WHERE user_id = $3",
    req.Amount, req.Amount, req.SenderID,
  )

  if err != nil {
    s.log.Printf("TransferTokens deduct sender failed: %v", err)
    return ErrTransactionFailed
  }

  // Add to recipient
  _, err = tx.ExecContext(
    ctx,
    "UPDATE tokens SET balance = balance + $1, total_earned = total_earned + $2, updated_at = NOW() WHERE user_id = $3",
    req.Amount, req.Amount, req.RecipientID,
  )

  if err != nil {
    s.log.Printf("TransferTokens add recipient failed: %v", err)
    return ErrTransactionFailed
  }

  // Record sender transaction
  _, err = tx.ExecContext(
    ctx,
    "INSERT INTO token_transactions (user_id, type, amount, reason, created_at) VALUES ($1, $2, $3, $4, NOW())",
    req.SenderID, "transfer_out", req.Amount, fmt.Sprintf("Transfer to user %d", req.RecipientID),
  )

  if err != nil {
    s.log.Printf("TransferTokens insert sender transaction failed: %v", err)
    return ErrTransactionFailed
  }

  // Record recipient transaction
  _, err = tx.ExecContext(
    ctx,
    "INSERT INTO token_transactions (user_id, type, amount, reason, created_at) VALUES ($1, $2, $3, $4, NOW())",
    req.RecipientID, "transfer_in", req.Amount, fmt.Sprintf("Transfer from user %d", req.SenderID),
  )

  if err != nil {
    s.log.Printf("TransferTokens insert recipient transaction failed: %v", err)
    return ErrTransactionFailed
  }

  if err := tx.Commit(); err != nil {
    s.log.Printf("TransferTokens commit failed: %v", err)
    return ErrTransactionFailed
  }

  // Invalidate caches
  cache.Delete(ctx, cache.TokenBalanceKey(req.SenderID))
  cache.Delete(ctx, cache.TokenBalanceKey(req.RecipientID))

  return nil
}

// InitializeUserTokens creates a token record for a new user
func (s *service) InitializeUserTokens(ctx context.Context, userID int) (*TokenBalance, error) {
  if userID <= 0 {
    return nil, ErrInvalidUserID
  }

  var balance TokenBalance
  err := s.db.QueryRowContext(
    ctx,
    `INSERT INTO tokens (user_id, balance, total_used, total_earned, created_at, updated_at)
     VALUES ($1, 0, 0, 0, NOW(), NOW())
     ON CONFLICT (user_id) DO UPDATE SET updated_at = NOW()
     RETURNING id, user_id, balance, total_used, total_earned, created_at, updated_at`,
    userID,
  ).Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.TotalUsed, &balance.TotalEarned, &balance.CreatedAt, &balance.UpdatedAt)

  if err != nil {
    s.log.Printf("InitializeUserTokens failed: %v", err)
    return nil, wrapError("InitializeUserTokens", "insert", err)
  }

  return &balance, nil
}

// AdjustBalance adjusts user balance by a delta amount
func (s *service) AdjustBalance(ctx context.Context, userID int, delta float64, reason string) (*TokenBalance, error) {
  if userID <= 0 {
    return nil, ErrInvalidUserID
  }

  if reason == "" {
    return nil, ErrInvalidReason
  }

  tx, err := s.db.BeginTx(ctx, nil)
  if err != nil {
    s.log.Printf("AdjustBalance begin transaction failed: %v", err)
    return nil, ErrTransactionFailed
  }
  defer tx.Rollback()

  var balance TokenBalance
  var totalEarned float64
  var totalUsed float64

  // Adjust based on delta sign
  if delta > 0 {
    totalEarned = delta
    totalUsed = 0
  } else {
    totalEarned = 0
    totalUsed = -delta
  }

  err = tx.QueryRowContext(
    ctx,
    `UPDATE tokens SET balance = balance + $1, total_earned = total_earned + $2, total_used = total_used + $3, updated_at = NOW()
     WHERE user_id = $4
     RETURNING id, user_id, balance, total_used, total_earned, created_at, updated_at`,
    delta, totalEarned, totalUsed, userID,
  ).Scan(&balance.ID, &balance.UserID, &balance.Balance, &balance.TotalUsed, &balance.TotalEarned, &balance.CreatedAt, &balance.UpdatedAt)

  if err != nil {
    if err == sql.ErrNoRows {
      tx.Rollback()
      return nil, ErrTokenNotFound
    }
    s.log.Printf("AdjustBalance update failed: %v", err)
    return nil, ErrTransactionFailed
  }

  // Record transaction
  txType := "adjustment"
  _, err = tx.ExecContext(
    ctx,
    `INSERT INTO token_transactions (user_id, type, amount, reason, created_at)
     VALUES ($1, $2, $3, $4, NOW())`,
    userID, txType, delta, reason,
  )

  if err != nil {
    s.log.Printf("AdjustBalance insert transaction failed: %v", err)
    return nil, ErrTransactionFailed
  }

  if err := tx.Commit(); err != nil {
    s.log.Printf("AdjustBalance commit failed: %v", err)
    return nil, ErrTransactionFailed
  }

  // Invalidate cache
  cache.Delete(ctx, cache.TokenBalanceKey(userID))

  return &balance, nil
}

// IsBalanceSufficient checks if user has sufficient balance
func (s *service) IsBalanceSufficient(ctx context.Context, userID int, requiredAmount float64) (bool, error) {
  if userID <= 0 {
    return false, ErrInvalidUserID
  }

  if requiredAmount <= 0 {
    return false, ErrInvalidAmount
  }

  balance, err := s.GetBalance(ctx, userID)
  if err != nil {
    return false, err
  }

  return balance.Balance >= requiredAmount, nil
}

// errInvalidRequest is a helper function for invalid request errors
func errInvalidRequest() *apperrors.AppError {
  return apperrors.NewAppError(
    "INVALID_REQUEST",
    "Invalid request",
    http.StatusBadRequest,
    nil,
  )
}
