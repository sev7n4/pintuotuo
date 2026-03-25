package token

import "time"

// RechargeTokensRequest represents a token recharge request
type RechargeTokensRequest struct {
  UserID int     `json:"user_id" binding:"required,gt=0"`
  Amount float64 `json:"amount" binding:"required,gt=0"`
  Reason string  `json:"reason" binding:"required"`
}

// ConsumeTokensRequest represents a token consumption request
type ConsumeTokensRequest struct {
  UserID int     `json:"user_id" binding:"required,gt=0"`
  Amount float64 `json:"amount" binding:"required,gt=0"`
  Reason string  `json:"reason" binding:"required"`
  Source string  `json:"source" binding:"required"` // e.g., "api_call", "payment", "adjustment"
}

// TransferTokensRequest represents a token transfer request
type TransferTokensRequest struct {
  SenderID    int     `json:"sender_id" binding:"required,gt=0"`
  RecipientID int     `json:"recipient_id" binding:"required,gt=0"`
  Amount      float64 `json:"amount" binding:"required,gt=0"`
}

// GetConsumptionParams represents consumption query parameters
type GetConsumptionParams struct {
  UserID    int
  StartDate *time.Time
  EndDate   *time.Time
  PageSize  int
  Page      int
}

// TransactionParams represents transaction query parameters
type TransactionParams struct {
  UserID int
  Type   string        // Optional filter: "recharge", "consume", "transfer_out", "transfer_in", "adjustment"
  Limit  int
  Offset int
}

// TokenBalance represents user token balance information
type TokenBalance struct {
  ID          int       `json:"id"`
  UserID      int       `json:"user_id"`
  Balance     float64   `json:"balance"`
  TotalUsed   float64   `json:"total_used"`
  TotalEarned float64   `json:"total_earned"`
  CreatedAt   time.Time `json:"created_at"`
  UpdatedAt   time.Time `json:"updated_at"`
}

// TokenTransaction represents a single token transaction
type TokenTransaction struct {
  ID        int       `json:"id"`
  UserID    int       `json:"user_id"`
  Type      string    `json:"type"` // "recharge", "consume", "transfer_out", "transfer_in", "adjustment"
  Amount    float64   `json:"amount"`
  Reason    string    `json:"reason"`
  OrderID   *int      `json:"order_id,omitempty"`
  CreatedAt time.Time `json:"created_at"`
}

// ConsumptionResult represents paginated consumption query result
type ConsumptionResult struct {
  Total        int                 `json:"total"`
  Page         int                 `json:"page"`
  PageSize     int                 `json:"page_size"`
  Transactions []TokenTransaction `json:"transactions"`
}

// MerchantRevenue represents merchant token revenue information
type MerchantRevenue struct {
  MerchantID        int       `json:"merchant_id"`
  TotalTokenEarned  float64   `json:"total_token_earned"`
  TotalTokenConsumed float64   `json:"total_token_consumed"`
  NetTokenBalance   float64   `json:"net_token_balance"`
  TransactionCount  int       `json:"transaction_count"`
  UpdatedAt         time.Time `json:"updated_at"`
}

// TokenBalanceUpdate represents an atomic balance update
type TokenBalanceUpdate struct {
  UserID int
  Delta  float64
  Reason string
}
