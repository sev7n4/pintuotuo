package services

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestBillingService_GetMerchantBillings(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	merchantID := 1
	filter := &BillingFilter{
		MerchantID: &merchantID,
		Page:       1,
		PageSize:   20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "merchant_id", "company_name", "user_id", "username",
		"provider", "model", "input_tokens", "output_tokens", "total_tokens",
		"cost", "request_time", "status_code", "created_at",
	}).AddRow(
		1, 1, "Test Company", nil, nil,
		"openai", "gpt-4", 100, 200, 300,
		0.05, 150.5, 200, time.Now(),
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM ( SELECT aul.id, mak.merchant_id, m.company_name, aul.user_id, u.name, aul.provider, aul.model, aul.input_tokens, aul.output_tokens, (aul.input_tokens + aul.output_tokens) as total_tokens, aul.cost, aul.latency_ms, aul.status_code, aul.created_at FROM api_usage_logs aul JOIN merchant_api_keys mak ON mak.id = aul.key_id JOIN merchants m ON m.id = mak.merchant_id LEFT JOIN users u ON u.id = aul.user_id WHERE 1=1 AND mak.merchant_id = $1) as subquery`)).
		WithArgs(merchantID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT aul.id, mak.merchant_id, m.company_name, aul.user_id, u.name, aul.provider, aul.model, aul.input_tokens, aul.output_tokens, (aul.input_tokens + aul.output_tokens) as total_tokens, aul.cost, aul.latency_ms, aul.status_code, aul.created_at FROM api_usage_logs aul JOIN merchant_api_keys mak ON mak.id = aul.key_id JOIN merchants m ON m.id = mak.merchant_id LEFT JOIN users u ON u.id = aul.user_id WHERE 1=1 AND mak.merchant_id = $1 ORDER BY aul.created_at DESC LIMIT $2 OFFSET $3`)).
		WithArgs(merchantID, 20, 0).
		WillReturnRows(rows)

	billings, total, err := service.GetMerchantBillings(filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, billings, 1)
	assert.Equal(t, "openai", billings[0].Provider)
	assert.Equal(t, "gpt-4", billings[0].Model)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_GetMerchantBillings_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	merchantID := 1
	provider := "openai"
	model := "gpt-4"
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)

	filter := &BillingFilter{
		MerchantID: &merchantID,
		Provider:   &provider,
		Model:      &model,
		StartDate:  &startDate,
		EndDate:    &endDate,
		Page:       1,
		PageSize:   20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "merchant_id", "company_name", "user_id", "username",
		"provider", "model", "input_tokens", "output_tokens", "total_tokens",
		"cost", "request_time", "status_code", "created_at",
	})

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT aul\.id, mak\.merchant_id`).
		WillReturnRows(rows)

	billings, total, err := service.GetMerchantBillings(filter)
	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, billings, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_GetUserBillings(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	userID := 1
	filter := &BillingFilter{
		UserID:   &userID,
		Page:     1,
		PageSize: 20,
	}

	rows := sqlmock.NewRows([]string{
		"id", "merchant_id", "company_name", "user_id", "username",
		"provider", "model", "input_tokens", "output_tokens", "total_tokens",
		"cost", "request_time", "status_code", "created_at",
	}).AddRow(
		1, 1, "Test Company", 1, "testuser",
		"openai", "gpt-4", 100, 200, 300,
		0.05, 150.5, 200, time.Now(),
	)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT aul\.id, mak\.merchant_id`).
		WillReturnRows(rows)

	billings, total, err := service.GetUserBillings(filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, billings, 1)
	assert.Equal(t, 1, *billings[0].UserID)
	assert.Equal(t, "testuser", *billings[0].Username)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_GetBillingStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	merchantID := 1
	filter := &BillingFilter{
		MerchantID: &merchantID,
	}

	statsRows := sqlmock.NewRows([]string{
		"total_requests", "total_cost", "total_tokens", "average_latency", "success_rate",
	}).AddRow(100, 50.5, 10000, 150.5, 99.5)

	mock.ExpectQuery(`SELECT COUNT\(\*\) as total_requests`).WillReturnRows(statsRows)

	providerRows := sqlmock.NewRows([]string{
		"provider", "count", "cost",
	}).AddRow("openai", 60, 30.0).AddRow("anthropic", 40, 20.5)

	mock.ExpectQuery(`SELECT aul\.provider, COUNT\(\*\) as count`).WillReturnRows(providerRows)

	stats, err := service.GetBillingStats(filter)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 100, stats.TotalRequests)
	assert.Equal(t, 50.5, stats.TotalCost)
	assert.Equal(t, int64(10000), stats.TotalTokens)
	assert.Equal(t, 150.5, stats.AverageLatency)
	assert.Equal(t, 99.5, stats.SuccessRate)
	assert.Len(t, stats.ProviderBreakdown, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_GetBillingTrends(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	merchantID := 1
	filter := &BillingFilter{
		MerchantID: &merchantID,
	}

	trendRows := sqlmock.NewRows([]string{
		"date", "total_cost", "total_tokens", "total_requests", "avg_latency",
	}).AddRow("2026-01-01", 10.5, 1000, 50, 120.5).
		AddRow("2026-01-02", 15.0, 1500, 75, 130.0)

	mock.ExpectQuery(`SELECT DATE\(aul\.created_at\) as date`).WillReturnRows(trendRows)

	trends, err := service.GetBillingTrends(filter, "day")
	assert.NoError(t, err)
	assert.Len(t, trends, 2)
	assert.Equal(t, "2026-01-01", trends[0].Date)
	assert.Equal(t, 10.5, trends[0].TotalCost)
	assert.Equal(t, int64(1000), trends[0].TotalTokens)
	assert.Equal(t, 50, trends[0].TotalRequests)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_GetBillingTrends_ByMonth(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	filter := &BillingFilter{}

	trendRows := sqlmock.NewRows([]string{
		"date", "total_cost", "total_tokens", "total_requests", "avg_latency",
	}).AddRow("2026-01", 100.5, 10000, 500, 125.5).
		AddRow("2026-02", 120.0, 12000, 600, 130.0)

	mock.ExpectQuery(`SELECT TO_CHAR\(aul\.created_at, 'YYYY-MM'\) as date`).WillReturnRows(trendRows)

	trends, err := service.GetBillingTrends(filter, "month")
	assert.NoError(t, err)
	assert.Len(t, trends, 2)
	assert.Equal(t, "2026-01", trends[0].Date)
	assert.Equal(t, "2026-02", trends[1].Date)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_ExportBillingsToCSV(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	filter := &BillingFilter{
		Page:     1,
		PageSize: 100,
	}

	username := "testuser"
	userID := 1

	rows := sqlmock.NewRows([]string{
		"id", "merchant_id", "company_name", "user_id", "username",
		"provider", "model", "input_tokens", "output_tokens", "total_tokens",
		"cost", "request_time", "status_code", "created_at",
	}).AddRow(
		1, 1, "Test Company", userID, username,
		"openai", "gpt-4", 100, 200, 300,
		0.05, 150.5, 200, time.Now(),
	)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT aul\.id, mak\.merchant_id`).
		WillReturnRows(rows)

	csvData, err := service.ExportBillingsToCSV(filter)
	assert.NoError(t, err)
	assert.NotNil(t, csvData)
	assert.Contains(t, string(csvData), "ID,商户ID,公司名称")
	assert.Contains(t, string(csvData), "openai")
	assert.Contains(t, string(csvData), "gpt-4")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_ExportUserBillingsToCSV(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	userID := 1
	filter := &BillingFilter{
		UserID:   &userID,
		Page:     1,
		PageSize: 100,
	}

	username := "testuser"

	rows := sqlmock.NewRows([]string{
		"id", "merchant_id", "company_name", "user_id", "username",
		"provider", "model", "input_tokens", "output_tokens", "total_tokens",
		"cost", "request_time", "status_code", "created_at",
	}).AddRow(
		1, 1, "Test Company", userID, username,
		"openai", "gpt-4", 100, 200, 300,
		0.05, 150.5, 200, time.Now(),
	)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT aul\.id, mak\.merchant_id`).
		WillReturnRows(rows)

	csvData, err := service.ExportUserBillingsToCSV(filter)
	assert.NoError(t, err)
	assert.NotNil(t, csvData)
	assert.Contains(t, string(csvData), "ID,商户ID,公司名称")
	assert.Contains(t, string(csvData), "testuser")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_GetMerchantBillings_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	merchantID := 1
	filter := &BillingFilter{
		MerchantID: &merchantID,
		Page:       1,
		PageSize:   20,
	}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT aul\.id, mak\.merchant_id`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "merchant_id", "company_name", "user_id", "username",
			"provider", "model", "input_tokens", "output_tokens", "total_tokens",
			"cost", "request_time", "status_code", "created_at",
		}))

	billings, total, err := service.GetMerchantBillings(filter)
	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, billings, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_GetBillingStats_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	filter := &BillingFilter{}

	statsRows := sqlmock.NewRows([]string{
		"total_requests", "total_cost", "total_tokens", "average_latency", "success_rate",
	}).AddRow(0, 0.0, 0, 0.0, 0.0)

	mock.ExpectQuery(`SELECT COUNT\(\*\) as total_requests`).WillReturnRows(statsRows)
	mock.ExpectQuery(`SELECT aul\.provider, COUNT\(\*\) as count`).WillReturnRows(sqlmock.NewRows([]string{
		"provider", "count", "cost",
	}))

	stats, err := service.GetBillingStats(filter)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalRequests)
	assert.Equal(t, 0.0, stats.TotalCost)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_GetBillingTrends_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	filter := &BillingFilter{}

	mock.ExpectQuery(`SELECT DATE\(aul\.created_at\) as date`).
		WillReturnRows(sqlmock.NewRows([]string{
			"date", "total_cost", "total_tokens", "total_requests", "avg_latency",
		}))

	trends, err := service.GetBillingTrends(filter, "day")
	assert.NoError(t, err)
	assert.Len(t, trends, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBillingService_GetBillingTrends_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := NewBillingService(db)

	filter := &BillingFilter{}

	mock.ExpectQuery(`SELECT DATE\(aul\.created_at\) as date`).
		WillReturnError(sql.ErrConnDone)

	trends, err := service.GetBillingTrends(filter, "day")
	assert.Error(t, err)
	assert.Nil(t, trends)

	assert.NoError(t, mock.ExpectationsWereMet())
}
