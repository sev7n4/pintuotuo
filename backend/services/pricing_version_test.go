package services

import (
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPricingVersionCodeBaseline(t *testing.T) {
	assert.Equal(t, "baseline", pricingVersionCodeBaseline)
}

func TestCostFromPer1KRates(t *testing.T) {
	got := CostFromPer1KRates(0.01, 0.03, 1000, 500)
	assert.InDelta(t, 0.025, got, 1e-9)
}

func TestLatestUserPricingVersionID_ReturnsRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT pricing_version_id FROM orders`).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"pricing_version_id"}).AddRow(42))

	vid := LatestUserPricingVersionID(db, 7)
	assert.True(t, vid.Valid)
	assert.Equal(t, int64(42), vid.Int64)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLatestUserPricingVersionID_NoOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT pricing_version_id FROM orders`).
		WithArgs(7).
		WillReturnError(sql.ErrNoRows)

	vid := LatestUserPricingVersionID(db, 7)
	assert.False(t, vid.Valid)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalculateCostFromPricingVersion_Hit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT r.provider_input_rate, r.provider_output_rate`).
		WithArgs(5, "openai", "gpt-4o").
		WillReturnRows(sqlmock.NewRows([]string{"provider_input_rate", "provider_output_rate"}).
			AddRow(0.01, 0.03))

	cost, ok := CalculateCostFromPricingVersion(db, 5, "openai", "gpt-4o", 2000, 1000)
	assert.True(t, ok)
	want := CostFromPer1KRates(0.01, 0.03, 2000, 1000)
	assert.InDelta(t, want, cost, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalculateCostFromPricingVersion_Miss(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT r.provider_input_rate, r.provider_output_rate`).
		WithArgs(5, "openai", "unknown-model-xyz").
		WillReturnError(sql.ErrNoRows)

	_, ok := CalculateCostFromPricingVersion(db, 5, "openai", "unknown-model-xyz", 100, 100)
	assert.False(t, ok)
	require.NoError(t, mock.ExpectationsWereMet())
}
