package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pintuotuo/backend/logger"
)

type BillingService struct {
	db *sql.DB
}

func NewBillingService(db *sql.DB) *BillingService {
	return &BillingService{db: db}
}

type BillingRecordDetail struct {
	ID           int       `json:"id"`
	MerchantID   int       `json:"merchant_id"`
	CompanyName  string    `json:"company_name"`
	UserID       *int      `json:"user_id,omitempty"`
	Username     *string   `json:"username,omitempty"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	TotalTokens  int       `json:"total_tokens"`
	Cost         float64   `json:"cost"`
	RequestTime  float64   `json:"request_time"`
	StatusCode   int       `json:"status_code"`
	CreatedAt    time.Time `json:"created_at"`
}

type BillingStats struct {
	// TotalCost is sum(api_usage_logs.cost): same internal Token ledger units as tokens.balance deductions (not CNY GMV).
	TotalCost         float64                  `json:"total_cost"`
	TotalRequests     int                      `json:"total_requests"`
	TotalTokens       int64                    `json:"total_tokens"`
	AverageLatency    float64                  `json:"average_latency"`
	SuccessRate       float64                  `json:"success_rate"`
	ProviderBreakdown map[string]ProviderStats `json:"provider_breakdown"`
}

type ProviderStats struct {
	Provider      string  `json:"provider"`
	TotalCost     float64 `json:"total_cost"`
	TotalRequests int     `json:"total_requests"`
	Percentage    float64 `json:"percentage"`
}

type BillingFilter struct {
	MerchantID *int
	UserID     *int
	Provider   *string
	Model      *string
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}

func (s *BillingService) GetMerchantBillings(filter *BillingFilter) ([]BillingRecordDetail, int, error) {
	ctx := context.Background()

	logger.LogInfo(ctx, "billing_service", "Getting merchant billings", map[string]interface{}{
		"merchant_id": filter.MerchantID,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
	})

	query := `
		SELECT aul.id, mak.merchant_id, m.company_name, aul.user_id, u.name,
			   aul.provider, aul.model, aul.input_tokens, aul.output_tokens,
			   (aul.input_tokens + aul.output_tokens) as total_tokens,
			   aul.cost, aul.latency_ms, aul.status_code, aul.created_at
		FROM api_usage_logs aul
		JOIN merchant_api_keys mak ON mak.id = aul.key_id
		JOIN merchants m ON m.id = mak.merchant_id
		LEFT JOIN users u ON u.id = aul.user_id
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.MerchantID != nil {
		query += fmt.Sprintf(" AND mak.merchant_id = $%d", argIndex)
		args = append(args, *filter.MerchantID)
		argIndex++
	}

	if filter.Provider != nil {
		query += fmt.Sprintf(" AND aul.provider = $%d", argIndex)
		args = append(args, *filter.Provider)
		argIndex++
	}

	if filter.Model != nil {
		query += fmt.Sprintf(" AND aul.model = $%d", argIndex)
		args = append(args, *filter.Model)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND aul.created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND aul.created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) FROM (" + query + ") as subquery"
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		logger.LogError(ctx, "billing_service", "Failed to count billings", err, nil)
		return nil, 0, fmt.Errorf("failed to count billings: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	query += fmt.Sprintf(" ORDER BY aul.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		logger.LogError(ctx, "billing_service", "Failed to query billings", err, nil)
		return nil, 0, fmt.Errorf("failed to query billings: %w", err)
	}
	defer rows.Close()

	var billings []BillingRecordDetail
	for rows.Next() {
		var b BillingRecordDetail
		err := rows.Scan(
			&b.ID, &b.MerchantID, &b.CompanyName, &b.UserID, &b.Username,
			&b.Provider, &b.Model, &b.InputTokens, &b.OutputTokens,
			&b.TotalTokens, &b.Cost, &b.RequestTime, &b.StatusCode, &b.CreatedAt,
		)
		if err != nil {
			logger.LogError(ctx, "billing_service", "Failed to scan billing record", err, nil)
			continue
		}
		billings = append(billings, b)
	}

	if billings == nil {
		billings = []BillingRecordDetail{}
	}

	logger.LogInfo(ctx, "billing_service", "Merchant billings retrieved", map[string]interface{}{
		"total": len(billings),
	})

	return billings, total, nil
}

func (s *BillingService) GetUserBillings(filter *BillingFilter) ([]BillingRecordDetail, int, error) {
	ctx := context.Background()

	logger.LogInfo(ctx, "billing_service", "Getting user billings", map[string]interface{}{
		"user_id": filter.UserID,
		"page":    filter.Page,
	})

	filter.MerchantID = nil

	query := `
		SELECT aul.id, mak.merchant_id, m.company_name, aul.user_id, u.name,
			   aul.provider, aul.model, aul.input_tokens, aul.output_tokens,
			   (aul.input_tokens + aul.output_tokens) as total_tokens,
			   aul.cost, aul.latency_ms, aul.status_code, aul.created_at
		FROM api_usage_logs aul
		JOIN merchant_api_keys mak ON mak.id = aul.key_id
		JOIN merchants m ON m.id = mak.merchant_id
		LEFT JOIN users u ON u.id = aul.user_id
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND aul.user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND aul.created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND aul.created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) FROM (" + query + ") as subquery"
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user billings: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize
	query += fmt.Sprintf(" ORDER BY aul.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user billings: %w", err)
	}
	defer rows.Close()

	var billings []BillingRecordDetail
	for rows.Next() {
		var b BillingRecordDetail
		err := rows.Scan(
			&b.ID, &b.MerchantID, &b.CompanyName, &b.UserID, &b.Username,
			&b.Provider, &b.Model, &b.InputTokens, &b.OutputTokens,
			&b.TotalTokens, &b.Cost, &b.RequestTime, &b.StatusCode, &b.CreatedAt,
		)
		if err != nil {
			continue
		}
		billings = append(billings, b)
	}

	if billings == nil {
		billings = []BillingRecordDetail{}
	}

	return billings, total, nil
}

func (s *BillingService) GetBillingStats(filter *BillingFilter) (*BillingStats, error) {
	ctx := context.Background()

	logger.LogInfo(ctx, "billing_service", "Getting billing stats", map[string]interface{}{
		"merchant_id": filter.MerchantID,
	})

	query := `
		SELECT
			COUNT(*) as total_requests,
			COALESCE(SUM(aul.cost), 0) as total_cost,
			COALESCE(SUM(aul.input_tokens + aul.output_tokens), 0) as total_tokens,
			COALESCE(AVG(aul.latency_ms), 0) as average_latency,
			COALESCE(100.0 * COUNT(CASE WHEN aul.status_code = 200 THEN 1 END) / NULLIF(COUNT(*), 0), 0) as success_rate
		FROM api_usage_logs aul
		JOIN merchant_api_keys mak ON mak.id = aul.key_id
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.MerchantID != nil {
		query += fmt.Sprintf(" AND mak.merchant_id = $%d", argIndex)
		args = append(args, *filter.MerchantID)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND aul.created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND aul.created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
	}

	var stats BillingStats
	err := s.db.QueryRow(query, args...).Scan(
		&stats.TotalRequests, &stats.TotalCost, &stats.TotalTokens,
		&stats.AverageLatency, &stats.SuccessRate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing stats: %w", err)
	}

	providerQuery := `
		SELECT aul.provider, COUNT(*) as count, COALESCE(SUM(aul.cost), 0) as cost
		FROM api_usage_logs aul
		JOIN merchant_api_keys mak ON mak.id = aul.key_id
		WHERE 1=1
	`
	providerArgs := []interface{}{}
	providerArgIndex := 1

	if filter.MerchantID != nil {
		providerQuery += fmt.Sprintf(" AND mak.merchant_id = $%d", providerArgIndex)
		providerArgs = append(providerArgs, *filter.MerchantID)
		providerArgIndex++
	}

	if filter.StartDate != nil {
		providerQuery += fmt.Sprintf(" AND aul.created_at >= $%d", providerArgIndex)
		providerArgs = append(providerArgs, *filter.StartDate)
		providerArgIndex++
	}

	if filter.EndDate != nil {
		providerQuery += fmt.Sprintf(" AND aul.created_at <= $%d", providerArgIndex)
		providerArgs = append(providerArgs, *filter.EndDate)
	}

	providerQuery += " GROUP BY aul.provider ORDER BY cost DESC"

	rows, err := s.db.Query(providerQuery, providerArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider breakdown: %w", err)
	}
	defer rows.Close()

	stats.ProviderBreakdown = make(map[string]ProviderStats)
	for rows.Next() {
		var ps ProviderStats
		var count int
		err := rows.Scan(&ps.Provider, &count, &ps.TotalCost)
		if err != nil {
			continue
		}
		ps.TotalRequests = count
		if stats.TotalCost > 0 {
			ps.Percentage = (ps.TotalCost / stats.TotalCost) * 100
		}
		stats.ProviderBreakdown[ps.Provider] = ps
	}

	return &stats, nil
}

type BillingTrend struct {
	Date          string  `json:"date"`
	TotalCost     float64 `json:"total_cost"`
	TotalTokens   int64   `json:"total_tokens"`
	TotalRequests int     `json:"total_requests"`
	AvgLatency    float64 `json:"avg_latency"`
}

func (s *BillingService) GetBillingTrends(filter *BillingFilter, granularity string) ([]BillingTrend, error) {
	ctx := context.Background()

	logger.LogInfo(ctx, "billing_service", "Getting billing trends", map[string]interface{}{
		"merchant_id": filter.MerchantID,
		"granularity": granularity,
	})

	var groupBy string
	switch granularity {
	case "day":
		groupBy = "DATE(aul.created_at)"
	case "week":
		groupBy = "TO_CHAR(aul.created_at, 'YYYY-\"W\"WW')"
	case "month":
		groupBy = "TO_CHAR(aul.created_at, 'YYYY-MM')"
	default:
		groupBy = "DATE(aul.created_at)"
	}

	query := fmt.Sprintf(`
		SELECT
			%s as date,
			COALESCE(SUM(aul.cost), 0) as total_cost,
			COALESCE(SUM(aul.input_tokens + aul.output_tokens), 0) as total_tokens,
			COUNT(*) as total_requests,
			COALESCE(AVG(aul.latency_ms), 0) as avg_latency
		FROM api_usage_logs aul
		JOIN merchant_api_keys mak ON mak.id = aul.key_id
		WHERE 1=1
	`, groupBy)

	args := []interface{}{}
	argIndex := 1

	if filter.MerchantID != nil {
		query += fmt.Sprintf(" AND mak.merchant_id = $%d", argIndex)
		args = append(args, *filter.MerchantID)
		argIndex++
	}

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND aul.user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND aul.created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND aul.created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
	}

	query += fmt.Sprintf(" GROUP BY %s ORDER BY date ASC", groupBy)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		logger.LogError(ctx, "billing_service", "Failed to get billing trends", err, nil)
		return nil, fmt.Errorf("failed to get billing trends: %w", err)
	}
	defer rows.Close()

	var trends []BillingTrend
	for rows.Next() {
		var t BillingTrend
		err := rows.Scan(&t.Date, &t.TotalCost, &t.TotalTokens, &t.TotalRequests, &t.AvgLatency)
		if err != nil {
			continue
		}
		trends = append(trends, t)
	}

	if trends == nil {
		trends = []BillingTrend{}
	}

	return trends, nil
}

func (s *BillingService) ExportBillingsToCSV(filter *BillingFilter) ([]byte, error) {
	billings, _, err := s.GetMerchantBillings(filter)
	if err != nil {
		return nil, err
	}

	var csv string
	csv = "ID,商户ID,公司名称,用户ID,用户名,Provider,Model,输入Token,输出Token,总Token,费用,请求时间(ms),状态码,创建时间\n"

	for _, b := range billings {
		username := ""
		if b.Username != nil {
			username = *b.Username
		}
		userID := ""
		if b.UserID != nil {
			userID = fmt.Sprintf("%d", *b.UserID)
		}
		csv += fmt.Sprintf("%d,%d,%s,%s,%s,%s,%s,%d,%d,%d,%.6f,%.2f,%d,%s\n",
			b.ID, b.MerchantID, b.CompanyName, userID, username,
			b.Provider, b.Model, b.InputTokens, b.OutputTokens, b.TotalTokens,
			b.Cost, b.RequestTime, b.StatusCode, b.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	return []byte(csv), nil
}

func (s *BillingService) GetProviders(filter *BillingFilter) ([]string, error) {
	query := `
                SELECT DISTINCT provider
                FROM api_usage_logs aul
                WHERE 1=1
        `
	args := []interface{}{}

	if filter != nil {
		if filter.MerchantID != nil {
			query += fmt.Sprintf(" AND aul.key_id IN (SELECT id FROM merchant_api_keys WHERE merchant_id = $%d)", len(args)+1)
			args = append(args, *filter.MerchantID)
		}
		if filter.StartDate != nil {
			query += fmt.Sprintf(" AND aul.created_at >= $%d", len(args)+1)
			args = append(args, *filter.StartDate)
		}
		if filter.EndDate != nil {
			query += fmt.Sprintf(" AND aul.created_at <= $%d", len(args)+1)
			args = append(args, *filter.EndDate)
		}
	}

	query += " ORDER BY provider"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	defer rows.Close()

	var providers []string
	for rows.Next() {
		var provider string
		err := rows.Scan(&provider)
		if err != nil {
			continue
		}
		providers = append(providers, provider)
	}

	return providers, nil
}

func (s *BillingService) GetModels(provider string) ([]string, error) {
	query := `
                SELECT DISTINCT model
                FROM api_usage_logs
                WHERE 1=1
        `
	args := []interface{}{}

	if provider != "" {
		query += fmt.Sprintf(" AND provider = $%d", len(args)+1)
		args = append(args, provider)
	}

	query += " ORDER BY model"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer rows.Close()

	var models []string
	for rows.Next() {
		var model string
		err := rows.Scan(&model)
		if err != nil {
			continue
		}
		models = append(models, model)
	}

	return models, nil
}

func (s *BillingService) ExportUserBillingsToCSV(filter *BillingFilter) ([]byte, error) {
	billings, _, err := s.GetUserBillings(filter)
	if err != nil {
		return nil, err
	}

	var csv string
	csv = "ID,商户ID,公司名称,用户ID,用户名,Provider,Model,输入Token,输出Token,总Token,费用,请求时间(ms),状态码,创建时间\n"

	for _, b := range billings {
		username := ""
		if b.Username != nil {
			username = *b.Username
		}
		userID := ""
		if b.UserID != nil {
			userID = fmt.Sprintf("%d", *b.UserID)
		}
		csv += fmt.Sprintf("%d,%d,%s,%s,%s,%s,%s,%d,%d,%d,%.6f,%.2f,%d,%s\n",
			b.ID, b.MerchantID, b.CompanyName, userID, username,
			b.Provider, b.Model, b.InputTokens, b.OutputTokens, b.TotalTokens,
			b.Cost, b.RequestTime, b.StatusCode, b.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	return []byte(csv), nil
}
