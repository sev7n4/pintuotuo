package analytics

import "time"

// ConsumptionRecord represents a single consumption record
type ConsumptionRecord struct {
  ID        int       `json:"id"`
  UserID    int       `json:"user_id"`
  Amount    float64   `json:"amount"`
  Reason    string    `json:"reason"`
  Type      string    `json:"type"`
  OrderID   *int      `json:"order_id,omitempty"`
  CreatedAt time.Time `json:"created_at"`
}

// UserConsumptionSummary represents total consumption for a user in a period
type UserConsumptionSummary struct {
  UserID          int       `json:"user_id"`
  TotalSpent      float64   `json:"total_spent"`
  TotalEarned     float64   `json:"total_earned"`
  TransactionCount int      `json:"transaction_count"`
  StartDate       time.Time `json:"start_date"`
  EndDate         time.Time `json:"end_date"`
  Period          string    `json:"period"` // daily, weekly, monthly
}

// SpendingPattern represents user spending pattern analysis
type SpendingPattern struct {
  UserID              int       `json:"user_id"`
  AverageDailySpend   float64   `json:"average_daily_spend"`
  MaxDailySpend       float64   `json:"max_daily_spend"`
  MinDailySpend       float64   `json:"min_daily_spend"`
  FrequentTransactionType string `json:"frequent_transaction_type"`
  Last30DaysSpent     float64   `json:"last_30_days_spent"`
  TrendPercentage     float64   `json:"trend_percentage"` // positive = increasing
}

// RevenueData represents revenue information
type RevenueData struct {
  MerchantID        int       `json:"merchant_id,omitempty"`
  ProductID         int       `json:"product_id,omitempty"`
  PeriodStartDate   time.Time `json:"period_start_date"`
  PeriodEndDate     time.Time `json:"period_end_date"`
  TotalTokensSold   float64   `json:"total_tokens_sold"`
  TotalRevenue      float64   `json:"total_revenue"`
  TransactionCount  int       `json:"transaction_count"`
  AverageOrderValue float64   `json:"average_order_value"`
}

// TopSpender represents a top spending user
type TopSpender struct {
  UserID           int       `json:"user_id"`
  Email            string    `json:"email"`
  Name             string    `json:"name"`
  TotalTokensSpent float64   `json:"total_tokens_spent"`
  TransactionCount int       `json:"transaction_count"`
  LastTransactionAt time.Time `json:"last_transaction_at"`
}

// TokenMetrics represents overall platform token metrics
type TokenMetrics struct {
  TotalTokensIssued   float64 `json:"total_tokens_issued"`
  TotalTokensConsumed float64 `json:"total_tokens_consumed"`
  ActiveUsers         int     `json:"active_users"`
  TransactionCount    int     `json:"transaction_count"`
  AverageUserBalance  float64 `json:"average_user_balance"`
  Timestamp           time.Time `json:"timestamp"`
}

// ConsumptionAnalyticsRequest represents analytics query parameters
type ConsumptionAnalyticsRequest struct {
  UserID    int
  StartDate time.Time
  EndDate   time.Time
  Period    string // daily, weekly, monthly
  Limit     int
  Offset    int
}

// RevenueAnalyticsRequest represents revenue query parameters
type RevenueAnalyticsRequest struct {
  MerchantID *int
  ProductID  *int
  StartDate  time.Time
  EndDate    time.Time
  Period     string
}

// TopSpendersRequest represents top spenders query parameters
type TopSpendersRequest struct {
  Limit     int
  Period    string // all_time, last_30_days, last_7_days
  MinSpend  float64
}
