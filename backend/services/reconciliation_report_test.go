package services

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalUsageLedgerMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT\s+COALESCE`).
		WillReturnRows(sqlmock.NewRows([]string{"a", "b"}).AddRow(100.5, 100.5))

	a, b, err := GlobalUsageLedgerMatch(db)
	require.NoError(t, err)
	assert.InDelta(t, 100.5, a, 1e-9)
	assert.InDelta(t, 100.5, b, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCountUsageDriftUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM`).
		WithArgs(usageReconcileEpsilon).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	n, err := CountUsageDriftUsers(db)
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGMVReportSummary_AllTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT\(\*\), COALESCE\(SUM\(total_price\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count", "sum"}).AddRow(10, 999.5))

	s, err := GetGMVReportSummary(db, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 10, s.OrderCount)
	assert.InDelta(t, 999.5, s.GMVCNY, 1e-9)
	assert.Equal(t, "CNY", s.Currency)
	require.NoError(t, mock.ExpectationsWereMet())
}
