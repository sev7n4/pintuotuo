package services

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconcileUserUsage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT\s+COALESCE`).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"usage_log_sum", "usage_tx_sum"}).AddRow(10.5, 10.5))

	gotLog, gotTx, err := ReconcileUserUsage(db, 42)
	require.NoError(t, err)
	assert.InDelta(t, 10.5, gotLog, 1e-9)
	assert.InDelta(t, 10.5, gotTx, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageReconcileOK(t *testing.T) {
	assert.True(t, UsageReconcileOK(1.0, 1.0+5e-7))
	assert.False(t, UsageReconcileOK(1.0, 1.01))
}
