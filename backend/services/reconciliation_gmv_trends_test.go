package services

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGMVTrends_Day(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, 1, 31, 23, 59, 59, 999999999, time.Local)

	mock.ExpectQuery(`SELECT to_char\(\(o\.created_at\)::date`).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"period", "count", "sum"}).
			AddRow("2026-01-15", 3, 199.5))

	pts, err := GetGMVTrends(db, "day", start, end)
	require.NoError(t, err)
	require.Len(t, pts, 1)
	assert.Equal(t, "2026-01-15", pts[0].Period)
	assert.Equal(t, 3, pts[0].OrderCount)
	assert.InDelta(t, 199.5, pts[0].GMVCNY, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}
