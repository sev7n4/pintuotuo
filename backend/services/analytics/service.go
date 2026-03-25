package analytics

import (
  "context"
  "database/sql"
  "fmt"
  "log"
  "os"
  "time"

  "github.com/pintuotuo/backend/cache"
)

// Service defines the analytics service interface
type Service interface {
  // Consumption Analytics
  GetUserConsumption(ctx context.Context, req *ConsumptionAnalyticsRequest) (*UserConsumptionSummary, error)
  GetUserConsumptionRecords(ctx context.Context, userID int, startDate, endDate time.Time, limit, offset int) ([]ConsumptionRecord, error)
  GetSpendingPattern(ctx context.Context, userID int) (*SpendingPattern, error)

  // Revenue Analytics
  GetRevenueData(ctx context.Context, req *RevenueAnalyticsRequest) (*RevenueData, error)
  GetProductRevenue(ctx context.Context, productID int, startDate, endDate time.Time) (*RevenueData, error)
  GetMerchantRevenue(ctx context.Context, merchantID int, startDate, endDate time.Time) (*RevenueData, error)

  // Top Performers
  GetTopSpenders(ctx context.Context, req *TopSpendersRequest) ([]TopSpender, error)

  // Platform Metrics
  GetPlatformMetrics(ctx context.Context) (*TokenMetrics, error)
}

// service implements the Service interface
type service struct {
  db  *sql.DB
  log *log.Logger
}

// NewService creates a new analytics service
func NewService(db *sql.DB, logger *log.Logger) Service {
  if logger == nil {
    logger = log.New(os.Stderr, "[AnalyticsService] ", log.LstdFlags)
  }

  return &service{
    db:  db,
    log: logger,
  }
}

// GetUserConsumption retrieves consumption summary for a user
func (s *service) GetUserConsumption(ctx context.Context, req *ConsumptionAnalyticsRequest) (*UserConsumptionSummary, error) {
  if req.UserID <= 0 {
    return nil, ErrUserNotFound
  }

  if req.StartDate.IsZero() || req.EndDate.IsZero() {
    return nil, ErrInvalidDateRange
  }

  if req.StartDate.After(req.EndDate) {
    return nil, ErrInvalidDateRange
  }

  // Try cache first
  cacheKey := fmt.Sprintf("analytics:consumption:%d:%s:%s", req.UserID, req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02"))
  if cachedData, err := cache.Get(ctx, cacheKey); err == nil {
    _ = cachedData // Cache hit - would unmarshal JSON here in real implementation
    s.log.Printf("Cache hit for %s", cacheKey)
  }

  // Query database
  var totalSpent, totalEarned float64
  var txnCount int

  err := s.db.QueryRowContext(
    ctx,
    `SELECT
      COALESCE(SUM(CASE WHEN type IN ('consume', 'transfer_out') THEN amount ELSE 0 END), 0) as total_spent,
      COALESCE(SUM(CASE WHEN type IN ('recharge', 'transfer_in') THEN amount ELSE 0 END), 0) as total_earned,
      COUNT(*) as txn_count
    FROM token_transactions
    WHERE user_id = $1 AND created_at BETWEEN $2 AND $3`,
    req.UserID, req.StartDate, req.EndDate,
  ).Scan(&totalSpent, &totalEarned, &txnCount)

  if err != nil && err != sql.ErrNoRows {
    s.log.Printf("GetUserConsumption query failed: %v", err)
    return nil, wrapError("GetUserConsumption", "query", err)
  }

  summary := &UserConsumptionSummary{
    UserID:          req.UserID,
    TotalSpent:      totalSpent,
    TotalEarned:     totalEarned,
    TransactionCount: txnCount,
    StartDate:       req.StartDate,
    EndDate:         req.EndDate,
    Period:          "custom",
  }

  return summary, nil
}

// GetUserConsumptionRecords retrieves detailed consumption records
func (s *service) GetUserConsumptionRecords(ctx context.Context, userID int, startDate, endDate time.Time, limit, offset int) ([]ConsumptionRecord, error) {
  if userID <= 0 {
    return nil, ErrUserNotFound
  }

  if startDate.IsZero() || endDate.IsZero() {
    return nil, ErrInvalidDateRange
  }

  if limit <= 0 || limit > 500 {
    return nil, ErrInvalidLimit
  }

  rows, err := s.db.QueryContext(
    ctx,
    `SELECT id, user_id, amount, reason, type, order_id, created_at
    FROM token_transactions
    WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
    ORDER BY created_at DESC
    LIMIT $4 OFFSET $5`,
    userID, startDate, endDate, limit, offset,
  )

  if err != nil {
    s.log.Printf("GetUserConsumptionRecords query failed: %v", err)
    return nil, wrapError("GetUserConsumptionRecords", "query", err)
  }
  defer rows.Close()

  var records []ConsumptionRecord
  for rows.Next() {
    var record ConsumptionRecord
    err := rows.Scan(&record.ID, &record.UserID, &record.Amount, &record.Reason, &record.Type, &record.OrderID, &record.CreatedAt)
    if err != nil {
      s.log.Printf("GetUserConsumptionRecords scan failed: %v", err)
      return nil, wrapError("GetUserConsumptionRecords", "scan", err)
    }
    records = append(records, record)
  }

  return records, nil
}

// GetSpendingPattern analyzes user spending patterns
func (s *service) GetSpendingPattern(ctx context.Context, userID int) (*SpendingPattern, error) {
  if userID <= 0 {
    return nil, ErrUserNotFound
  }

  // Last 30 days
  endDate := time.Now()
  startDate := endDate.AddDate(0, 0, -30)

  var avgDailySpend, maxDailySpend, minDailySpend, last30DaysSpent float64
  var frequentType string

  err := s.db.QueryRowContext(
    ctx,
    `SELECT
      COALESCE(AVG(daily_spend), 0) as avg_daily,
      COALESCE(MAX(daily_spend), 0) as max_daily,
      COALESCE(MIN(daily_spend), 0) as min_daily,
      COALESCE(SUM(daily_spend), 0) as total_30days,
      (SELECT type FROM token_transactions WHERE user_id = $1 AND created_at > $2 GROUP BY type ORDER BY COUNT(*) DESC LIMIT 1) as freq_type
    FROM (
      SELECT DATE(created_at), SUM(amount) as daily_spend
      FROM token_transactions
      WHERE user_id = $1 AND created_at > $2 AND type IN ('consume', 'transfer_out')
      GROUP BY DATE(created_at)
    ) daily_stats`,
    userID, startDate,
  ).Scan(&avgDailySpend, &maxDailySpend, &minDailySpend, &last30DaysSpent, &frequentType)

  if err != nil && err != sql.ErrNoRows {
    s.log.Printf("GetSpendingPattern query failed: %v", err)
    return nil, wrapError("GetSpendingPattern", "query", err)
  }

  // Calculate trend (simplified)
  var last15DaysSpent float64
  startDate15 := endDate.AddDate(0, 0, -15)
  s.db.QueryRowContext(
    ctx,
    `SELECT COALESCE(SUM(amount), 0)
    FROM token_transactions
    WHERE user_id = $1 AND created_at BETWEEN $2 AND $3 AND type IN ('consume', 'transfer_out')`,
    userID, startDate15, endDate,
  ).Scan(&last15DaysSpent)

  trendPercent := 0.0
  if last30DaysSpent > 0 {
    trendPercent = ((last15DaysSpent - (last30DaysSpent - last15DaysSpent)) / (last30DaysSpent - last15DaysSpent)) * 100
  }

  pattern := &SpendingPattern{
    UserID:                  userID,
    AverageDailySpend:       avgDailySpend,
    MaxDailySpend:           maxDailySpend,
    MinDailySpend:           minDailySpend,
    FrequentTransactionType: frequentType,
    Last30DaysSpent:         last30DaysSpent,
    TrendPercentage:         trendPercent,
  }

  return pattern, nil
}

// GetRevenueData retrieves revenue information based on query
func (s *service) GetRevenueData(ctx context.Context, req *RevenueAnalyticsRequest) (*RevenueData, error) {
  if req.StartDate.IsZero() || req.EndDate.IsZero() {
    return nil, ErrInvalidDateRange
  }

  var query string
  var args []interface{}

  if req.MerchantID != nil && *req.MerchantID > 0 {
    query = `SELECT
      COALESCE(SUM(p.amount), 0) as total_revenue,
      COALESCE(COUNT(p.id), 0) as txn_count,
      COALESCE(AVG(o.total_price), 0) as avg_order_value,
      COALESCE(SUM(p.amount), 0) as total_tokens
    FROM payments p
    JOIN orders o ON p.order_id = o.id
    JOIN products prod ON o.product_id = prod.id
    WHERE prod.merchant_id = $1 AND p.created_at BETWEEN $2 AND $3 AND p.status = 'success'`
    args = []interface{}{*req.MerchantID, req.StartDate, req.EndDate}
  } else if req.ProductID != nil && *req.ProductID > 0 {
    query = `SELECT
      COALESCE(SUM(p.amount), 0) as total_revenue,
      COALESCE(COUNT(p.id), 0) as txn_count,
      COALESCE(AVG(o.total_price), 0) as avg_order_value,
      COALESCE(SUM(p.amount), 0) as total_tokens
    FROM payments p
    JOIN orders o ON p.order_id = o.id
    WHERE o.product_id = $1 AND p.created_at BETWEEN $2 AND $3 AND p.status = 'success'`
    args = []interface{}{*req.ProductID, req.StartDate, req.EndDate}
  } else {
    return nil, ErrInvalidMetricsQuery
  }

  var totalRevenue, totalTokens, avgOrderValue float64
  var txnCount int

  err := s.db.QueryRowContext(ctx, query, args...).Scan(&totalRevenue, &txnCount, &avgOrderValue, &totalTokens)
  if err != nil && err != sql.ErrNoRows {
    s.log.Printf("GetRevenueData query failed: %v", err)
    return nil, wrapError("GetRevenueData", "query", err)
  }

  data := &RevenueData{
    MerchantID:       0,
    ProductID:        0,
    PeriodStartDate:  req.StartDate,
    PeriodEndDate:    req.EndDate,
    TotalTokensSold:  totalTokens,
    TotalRevenue:     totalRevenue,
    TransactionCount: txnCount,
    AverageOrderValue: avgOrderValue,
  }

  if req.MerchantID != nil {
    data.MerchantID = *req.MerchantID
  }
  if req.ProductID != nil {
    data.ProductID = *req.ProductID
  }

  return data, nil
}

// GetProductRevenue retrieves revenue for a specific product
func (s *service) GetProductRevenue(ctx context.Context, productID int, startDate, endDate time.Time) (*RevenueData, error) {
  return s.GetRevenueData(ctx, &RevenueAnalyticsRequest{
    ProductID: &productID,
    StartDate: startDate,
    EndDate:   endDate,
  })
}

// GetMerchantRevenue retrieves revenue for a specific merchant
func (s *service) GetMerchantRevenue(ctx context.Context, merchantID int, startDate, endDate time.Time) (*RevenueData, error) {
  return s.GetRevenueData(ctx, &RevenueAnalyticsRequest{
    MerchantID: &merchantID,
    StartDate:  startDate,
    EndDate:    endDate,
  })
}

// GetTopSpenders retrieves top spending users
func (s *service) GetTopSpenders(ctx context.Context, req *TopSpendersRequest) ([]TopSpender, error) {
  if req.Limit <= 0 || req.Limit > 500 {
    req.Limit = 10
  }

  var startDate time.Time
  switch req.Period {
  case "last_7_days":
    startDate = time.Now().AddDate(0, 0, -7)
  case "last_30_days":
    startDate = time.Now().AddDate(0, 0, -30)
  case "all_time":
    startDate = time.Unix(0, 0)
  default:
    startDate = time.Now().AddDate(0, 0, -30)
  }

  rows, err := s.db.QueryContext(
    ctx,
    `SELECT u.id, u.email, u.name,
      COALESCE(SUM(tt.amount), 0) as total_spent,
      COUNT(tt.id) as txn_count,
      MAX(tt.created_at) as last_txn
    FROM users u
    LEFT JOIN token_transactions tt ON u.id = tt.user_id AND tt.created_at > $1 AND tt.type IN ('consume', 'transfer_out')
    GROUP BY u.id, u.email, u.name
    HAVING COALESCE(SUM(tt.amount), 0) >= $2
    ORDER BY total_spent DESC
    LIMIT $3`,
    startDate, req.MinSpend, req.Limit,
  )

  if err != nil {
    s.log.Printf("GetTopSpenders query failed: %v", err)
    return nil, wrapError("GetTopSpenders", "query", err)
  }
  defer rows.Close()

  var spenders []TopSpender
  for rows.Next() {
    var spender TopSpender
    var lastTxn sql.NullTime
    err := rows.Scan(&spender.UserID, &spender.Email, &spender.Name, &spender.TotalTokensSpent, &spender.TransactionCount, &lastTxn)
    if err != nil {
      s.log.Printf("GetTopSpenders scan failed: %v", err)
      return nil, wrapError("GetTopSpenders", "scan", err)
    }
    if lastTxn.Valid {
      spender.LastTransactionAt = lastTxn.Time
    }
    spenders = append(spenders, spender)
  }

  return spenders, nil
}

// GetPlatformMetrics retrieves overall platform metrics
func (s *service) GetPlatformMetrics(ctx context.Context) (*TokenMetrics, error) {
  var totalIssued, totalConsumed float64
  var activeUsers, txnCount int

  // Get metrics from database
  err := s.db.QueryRowContext(
    ctx,
    `SELECT
      COALESCE(SUM(CASE WHEN type IN ('recharge', 'transfer_in') THEN amount ELSE 0 END), 0) as issued,
      COALESCE(SUM(CASE WHEN type IN ('consume', 'transfer_out') THEN amount ELSE 0 END), 0) as consumed,
      COUNT(DISTINCT user_id) as active_users,
      COUNT(*) as txn_count
    FROM token_transactions
    WHERE created_at > NOW() - INTERVAL '30 days'`,
  ).Scan(&totalIssued, &totalConsumed, &activeUsers, &txnCount)

  if err != nil {
    s.log.Printf("GetPlatformMetrics query failed: %v", err)
    return nil, wrapError("GetPlatformMetrics", "query", err)
  }

  var avgBalance float64
  s.db.QueryRowContext(
    ctx,
    "SELECT COALESCE(AVG(balance), 0) FROM tokens",
  ).Scan(&avgBalance)

  metrics := &TokenMetrics{
    TotalTokensIssued:   totalIssued,
    TotalTokensConsumed: totalConsumed,
    ActiveUsers:         activeUsers,
    TransactionCount:    txnCount,
    AverageUserBalance:  avgBalance,
    Timestamp:           time.Now(),
  }

  return metrics, nil
}
