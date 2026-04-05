package services

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSettlementService_GenerateMonthlySettlements(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("successful generation", func(t *testing.T) {
		merchantID := 1
		periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

		// Mock the first query to get merchant sales data
		mock.ExpectQuery(`SELECT m\.id, COALESCE`).
			WithArgs(periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_sales", "total_orders"}).
				AddRow(merchantID, 10000.00, 100))

		// Mock the check for existing settlement (returns no rows, meaning settlement doesn't exist)
		mock.ExpectQuery(`SELECT id FROM merchant_settlements WHERE merchant_id`).
			WithArgs(merchantID, periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Mock the insert with QueryRow (not Exec!)
		mock.ExpectQuery(`INSERT INTO merchant_settlements`).
			WithArgs(merchantID, periodStart, periodEnd, 10000.00, 500.00, 9500.00, "pending").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		settlements, err := service.GenerateMonthlySettlements(periodStart, periodEnd)
		assert.NoError(t, err)
		assert.Len(t, settlements, 1)
		assert.Equal(t, merchantID, settlements[0].MerchantID)
		assert.Equal(t, 10000.00, settlements[0].TotalSales)
		assert.Equal(t, 500.00, settlements[0].PlatformFee)
		assert.Equal(t, 9500.00, settlements[0].SettlementAmount)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("settlement already exists", func(t *testing.T) {
		merchantID := 1
		periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

		mock.ExpectQuery(`SELECT m\.id, COALESCE`).
			WithArgs(periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_sales", "total_orders"}).
				AddRow(merchantID, 10000.00, 100))

		mock.ExpectQuery(`SELECT id FROM merchant_settlements WHERE merchant_id`).
			WithArgs(merchantID, periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		settlements, err := service.GenerateMonthlySettlements(periodStart, periodEnd)
		assert.NoError(t, err)
		assert.Len(t, settlements, 0)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSettlementService_MerchantConfirm(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("successful confirmation", func(t *testing.T) {
		settlementID := 1
		merchantID := 1

		mock.ExpectQuery(`SELECT merchant_id, status FROM merchant_settlements WHERE id`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_id", "status"}).
				AddRow(merchantID, "pending"))

		mock.ExpectExec(`UPDATE merchant_settlements SET merchant_confirmed`).
			WithArgs(true, settlementID, merchantID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := service.MerchantConfirm(settlementID, merchantID)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("wrong merchant", func(t *testing.T) {
		settlementID := 1
		merchantID := 1

		mock.ExpectQuery(`SELECT merchant_id, status FROM merchant_settlements WHERE id`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_id", "status"}).
				AddRow(2, "pending"))

		err := service.MerchantConfirm(settlementID, merchantID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authorized")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSettlementService_FinanceApprove(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("successful approval", func(t *testing.T) {
		settlementID := 1
		financeUserID := 10

		mock.ExpectQuery(`SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_confirmed", "finance_approved"}).
				AddRow(true, false))

		mock.ExpectExec(`UPDATE merchant_settlements SET finance_approved`).
			WithArgs(true, financeUserID, settlementID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := service.FinanceApprove(settlementID, financeUserID)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("merchant not confirmed", func(t *testing.T) {
		settlementID := 1
		financeUserID := 10

		mock.ExpectQuery(`SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_confirmed", "finance_approved"}).
				AddRow(false, false))

		err := service.FinanceApprove(settlementID, financeUserID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchant confirmation required")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSettlementService_SubmitDispute(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("successful submission", func(t *testing.T) {
		settlementID := 1
		merchantID := 1
		disputeType := "amount_error"
		reason := "Incorrect order count"
		originalAmount := 10000.00
		disputedAmount := 9500.00

		mock.ExpectQuery(`SELECT merchant_id FROM merchant_settlements WHERE id`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_id"}).
				AddRow(merchantID))

		mock.ExpectExec(`INSERT INTO settlement_disputes`).
			WithArgs(settlementID, merchantID, disputeType, reason, originalAmount, disputedAmount, "pending").
			WillReturnResult(sqlmock.NewResult(1, 1))

		dispute, err := service.SubmitDispute(settlementID, merchantID, disputeType, reason, originalAmount, disputedAmount)
		assert.NoError(t, err)
		assert.NotNil(t, dispute)
		assert.Equal(t, settlementID, dispute.SettlementID)
		assert.Equal(t, disputeType, dispute.DisputeType)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSettlementService_ReconcileOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("successful reconciliation", func(t *testing.T) {
		settlementID := 1
		periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)
		merchantID := 1

		mock.ExpectQuery(`SELECT merchant_id, period_start, period_end FROM merchant_settlements WHERE id`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_id", "period_start", "period_end"}).
				AddRow(merchantID, periodStart, periodEnd))

		mock.ExpectQuery(`SELECT COUNT\(\*\), COALESCE\(SUM`).
			WithArgs(merchantID, periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"count", "total"}).
				AddRow(100, 10000.00))

		mock.ExpectQuery(`SELECT COUNT\(\*\), COALESCE\(SUM`).
			WithArgs(merchantID, periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"count", "total"}).
				AddRow(100, 10000.00))

		mock.ExpectExec(`INSERT INTO settlement_reconciliations`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		reconciliation, err := service.ReconcileOrders(settlementID)
		assert.NoError(t, err)
		assert.NotNil(t, reconciliation)
		assert.Equal(t, 100, reconciliation.OrderCountExpected)
		assert.Equal(t, 100, reconciliation.OrderCountActual)
		assert.Equal(t, 0, reconciliation.OrderCountDiff)
		assert.False(t, reconciliation.HasAnomalies)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("anomaly detected", func(t *testing.T) {
		settlementID := 1
		periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)
		merchantID := 1

		mock.ExpectQuery(`SELECT merchant_id, period_start, period_end FROM merchant_settlements WHERE id`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_id", "period_start", "period_end"}).
				AddRow(merchantID, periodStart, periodEnd))

		mock.ExpectQuery(`SELECT COUNT\(\*\), COALESCE\(SUM`).
			WithArgs(merchantID, periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"count", "total"}).
				AddRow(100, 10000.00))

		mock.ExpectQuery(`SELECT COUNT\(\*\), COALESCE\(SUM`).
			WithArgs(merchantID, periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"count", "total"}).
				AddRow(95, 9500.00))

		mock.ExpectExec(`INSERT INTO settlement_reconciliations`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		reconciliation, err := service.ReconcileOrders(settlementID)
		assert.NoError(t, err)
		assert.NotNil(t, reconciliation)
		assert.Equal(t, 100, reconciliation.OrderCountExpected)
		assert.Equal(t, 95, reconciliation.OrderCountActual)
		assert.Equal(t, 5, reconciliation.OrderCountDiff)
		assert.True(t, reconciliation.HasAnomalies)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
