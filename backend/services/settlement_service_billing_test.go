package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/lib/pq"
)

func TestGenerateMonthlySettlements_BillingBased(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	service := &SettlementService{db: db}

	periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"merchant_id", "total_requests", "total_tokens", "total_cost"}).
		AddRow(1, 100, 50000, 100.50)

	mock.ExpectQuery(`SELECT mak\.merchant_id, COUNT\(aul\.id\)`).
		WithArgs(periodStart, periodEnd).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM merchant_settlements WHERE merchant_id`).
		WithArgs(1, periodStart, periodEnd).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO merchant_settlements`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO settlement_items`).
		WillReturnResult(sqlmock.NewResult(0, 100))
	mock.ExpectCommit()

	settlements, err := service.GenerateMonthlySettlements(periodStart, periodEnd)
	if err != nil {
		t.Errorf("GenerateMonthlySettlements failed: %v", err)
	}

	if len(settlements) == 0 {
		t.Error("Expected at least 1 settlement")
		return
	}

	s := settlements[0]
	if s.TotalSales != 100.50 {
		t.Errorf("Expected total_sales 100.50, got %f", s.TotalSales)
	}
	if s.SettlementAmount <= 0 {
		t.Errorf("Settlement has invalid settlement_amount: %f", s.SettlementAmount)
	}
	if s.PlatformFee <= 0 {
		t.Errorf("Settlement has invalid platform_fee: %f", s.PlatformFee)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGenerateSettlementForMerchant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	service := &SettlementService{db: db}

	periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)
	merchantID := 1

	rows := sqlmock.NewRows([]string{"total_requests", "total_tokens", "total_cost"}).
		AddRow(100, 50000, 100.50)

	mock.ExpectQuery(`SELECT COUNT\(aul\.id\)`).
		WithArgs(merchantID, periodStart, periodEnd).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM merchant_settlements WHERE merchant_id`).
		WithArgs(merchantID, periodStart, periodEnd).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO merchant_settlements`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO settlement_items`).
		WillReturnResult(sqlmock.NewResult(0, 100))
	mock.ExpectCommit()

	settlement, err := service.GenerateSettlementForMerchant(merchantID, periodStart, periodEnd)
	if err != nil {
		t.Errorf("GenerateSettlementForMerchant failed: %v", err)
	}

	if settlement == nil {
		t.Error("Settlement should not be nil")
		return
	}

	if settlement.MerchantID != merchantID {
		t.Errorf("Expected merchant_id %d, got %d", merchantID, settlement.MerchantID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGetBillingRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	service := &SettlementService{db: db}

	startDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "merchant_id", "provider", "model",
		"input_tokens", "output_tokens", "token_usage", "cost", "request_id",
		"status_code", "latency_ms", "created_at",
	}).
		AddRow(1, 10, 1, "openai", "gpt-4", 1000, 500, 1500, 0.05, "req-123", 200, 150, time.Now())

	mock.ExpectQuery(`SELECT aul\.id, aul\.user_id`).
		WithArgs(startDate, endDate, 20, 0).
		WillReturnRows(rows)

	records, err := service.GetBillingRecords(0, startDate, endDate, 1, 20)
	if err != nil {
		t.Errorf("GetBillingRecords failed: %v", err)
	}

	if records == nil {
		t.Error("Records should not be nil")
	}

	if len(records) == 0 {
		t.Error("Expected at least 1 record")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestSettlementItems(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	service := &SettlementService{db: db}

	settlementID := 1

	rows := sqlmock.NewRows([]string{
		"id", "settlement_id", "api_usage_log_id", "user_id",
		"merchant_id", "provider", "model",
		"input_tokens", "output_tokens", "cost", "created_at",
	}).
		AddRow(1, settlementID, 100, 10, 1, "openai", "gpt-4", 1000, 500, 0.05, time.Now())

	mock.ExpectQuery(`SELECT id, settlement_id, api_usage_log_id`).
		WithArgs(settlementID).
		WillReturnRows(rows)

	items, err := service.GetSettlementItems(settlementID)
	if err != nil {
		t.Errorf("GetSettlementItems failed: %v", err)
	}

	if len(items) == 0 {
		t.Error("Expected at least 1 item")
		return
	}

	item := items[0]
	if item.SettlementID != settlementID {
		t.Errorf("Expected settlement_id %d, got %d", settlementID, item.SettlementID)
	}
	if item.Cost <= 0 {
		t.Errorf("Item has invalid cost: %f", item.Cost)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
