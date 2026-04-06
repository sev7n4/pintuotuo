package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSettlementService_FinanceApprove_TransactionProtection(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("should use transaction and row-level lock", func(t *testing.T) {
		settlementID := 1
		financeUserID := 10

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id = \$1 FOR UPDATE`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_confirmed", "finance_approved"}).
				AddRow(true, false))

		mock.ExpectExec(`UPDATE merchant_settlements SET finance_approved`).
			WithArgs(true, financeUserID, settlementID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := service.FinanceApprove(settlementID, financeUserID)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should rollback on error", func(t *testing.T) {
		settlementID := 1
		financeUserID := 10

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id = \$1 FOR UPDATE`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_confirmed", "finance_approved"}).
				AddRow(false, false))

		mock.ExpectRollback()

		err := service.FinanceApprove(settlementID, financeUserID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchant confirmation required")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should prevent concurrent approval with FOR UPDATE lock", func(t *testing.T) {
		settlementID := 1
		financeUserID := 10

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id = \$1 FOR UPDATE`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_confirmed", "finance_approved"}).
				AddRow(true, true))

		mock.ExpectRollback()

		err := service.FinanceApprove(settlementID, financeUserID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already approved")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSettlementService_MarkAsPaid_TransactionProtection(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("should use transaction and row-level lock", func(t *testing.T) {
		settlementID := 1
		userID := 10

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT finance_approved FROM merchant_settlements WHERE id = \$1 FOR UPDATE`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"finance_approved"}).
				AddRow(true))

		mock.ExpectExec(`UPDATE merchant_settlements SET marked_paid_at`).
			WithArgs(userID, settlementID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := service.MarkAsPaid(settlementID, userID)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should rollback on missing finance approval", func(t *testing.T) {
		settlementID := 1
		userID := 10

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT finance_approved FROM merchant_settlements WHERE id = \$1 FOR UPDATE`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"finance_approved"}).
				AddRow(false))

		mock.ExpectRollback()

		err := service.MarkAsPaid(settlementID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "finance approval required")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSettlementService_ProcessDispute_TransactionProtection(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("should use transaction for dispute processing", func(t *testing.T) {
		disputeID := 1
		handlerID := 10
		resolution := "Adjusted after review"
		adjustedAmount := 9500.00
		settlementID := 100

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT id, settlement_id, merchant_id, dispute_type, dispute_reason, evidence_urls, original_amount, disputed_amount, status FROM settlement_disputes WHERE id = \$1 FOR UPDATE`).
			WithArgs(disputeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "settlement_id", "merchant_id", "dispute_type", "dispute_reason",
				"evidence_urls", "original_amount", "disputed_amount", "status",
			}).AddRow(
				disputeID, settlementID, 1, "amount_error", "Incorrect amount",
				nil, 10000.00, 9500.00, "pending",
			))

		mock.ExpectExec(`UPDATE settlement_disputes SET status = 'resolved'`).
			WithArgs(handlerID, resolution, adjustedAmount, disputeID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`UPDATE merchant_settlements SET settlement_amount`).
			WithArgs(adjustedAmount, settlementID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := service.ProcessDispute(disputeID, handlerID, resolution, adjustedAmount)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should rollback on dispute not found", func(t *testing.T) {
		disputeID := 999
		handlerID := 10
		resolution := "Adjusted after review"
		adjustedAmount := 9500.00

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT id, settlement_id, merchant_id, dispute_type, dispute_reason, evidence_urls, original_amount, disputed_amount, status FROM settlement_disputes WHERE id = \$1 FOR UPDATE`).
			WithArgs(disputeID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "settlement_id", "merchant_id", "dispute_type", "dispute_reason",
				"evidence_urls", "original_amount", "disputed_amount", "status",
			}))

		mock.ExpectRollback()

		err := service.ProcessDispute(disputeID, handlerID, resolution, adjustedAmount)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "dispute not found")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSettlementService_GenerateMonthlySettlements_TransactionProtection(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("should use transaction for each settlement", func(t *testing.T) {
		periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)
		merchantID := 1

		mock.ExpectQuery(`SELECT mak\.merchant_id, COUNT`).
			WithArgs(periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_id", "total_requests", "total_tokens", "total_cost"}).
				AddRow(merchantID, 100, 50000, 10000.00))

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT id FROM merchant_settlements WHERE merchant_id = \$1 AND period_start = \$2 AND period_end = \$3 FOR UPDATE`).
			WithArgs(merchantID, periodStart, periodEnd).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery(`INSERT INTO merchant_settlements`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectExec(`INSERT INTO settlement_items`).
			WillReturnResult(sqlmock.NewResult(0, 100))

		mock.ExpectCommit()

		settlements, err := service.GenerateMonthlySettlements(periodStart, periodEnd)
		assert.NoError(t, err)
		assert.Len(t, settlements, 1)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should rollback and skip on existing settlement", func(t *testing.T) {
		periodStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)
		merchantID := 1

		mock.ExpectQuery(`SELECT mak\.merchant_id, COUNT`).
			WithArgs(periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_id", "total_requests", "total_tokens", "total_cost"}).
				AddRow(merchantID, 100, 50000, 10000.00))

		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT id FROM merchant_settlements WHERE merchant_id = \$1 AND period_start = \$2 AND period_end = \$3 FOR UPDATE`).
			WithArgs(merchantID, periodStart, periodEnd).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectRollback()

		settlements, err := service.GenerateMonthlySettlements(periodStart, periodEnd)
		assert.NoError(t, err)
		assert.Len(t, settlements, 0)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSettlementService_ConcurrentOperations(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("should prevent concurrent finance approval", func(t *testing.T) {
		settlementID := 1
		financeUserID1 := 10
		financeUserID2 := 11

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id = \$1 FOR UPDATE`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_confirmed", "finance_approved"}).
				AddRow(true, false))
		mock.ExpectExec(`UPDATE merchant_settlements SET finance_approved`).
			WithArgs(true, financeUserID1, settlementID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := service.FinanceApprove(settlementID, financeUserID1)
		assert.NoError(t, err)

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id = \$1 FOR UPDATE`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_confirmed", "finance_approved"}).
				AddRow(true, true))
		mock.ExpectRollback()

		err = service.FinanceApprove(settlementID, financeUserID2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already approved")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSettlementService_DataConsistency(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	service := &SettlementService{db: db}

	t.Run("should ensure finance approval before marking as paid", func(t *testing.T) {
		settlementID := 1
		userID := 10

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT finance_approved FROM merchant_settlements WHERE id = \$1 FOR UPDATE`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"finance_approved"}).
				AddRow(false))
		mock.ExpectRollback()

		err := service.MarkAsPaid(settlementID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "finance approval required")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should ensure merchant confirmation before finance approval", func(t *testing.T) {
		settlementID := 1
		financeUserID := 10

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id = \$1 FOR UPDATE`).
			WithArgs(settlementID).
			WillReturnRows(sqlmock.NewRows([]string{"merchant_confirmed", "finance_approved"}).
				AddRow(false, false))
		mock.ExpectRollback()

		err := service.FinanceApprove(settlementID, financeUserID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchant confirmation required")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
