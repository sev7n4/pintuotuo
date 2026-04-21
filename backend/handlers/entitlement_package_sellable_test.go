package handlers

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrichEntitlementPackageItems_FuelPackOnly(t *testing.T) {
	items := []EntitlementPackageItem{
		{
			SKUID:           1,
			SKUType:         "token_pack",
			DefaultQuantity: 1,
			SKUStatus:       merchantStatusActive,
			SPUStatus:       productStatusActive,
			Stock:           -1,
			ModelProvider:   "internal",
			ModelName:       "fuel",
		},
	}

	ok, reason := enrichEntitlementPackageItems(items)
	assert.False(t, ok)
	assert.Contains(t, reason, "加油包不可单独购买")
}

func TestValidateEntitlementPackageBundlePolicy_FuelPackOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	defer func() {
		mock.ExpectRollback()
		_ = tx.Rollback()
	}()

	mock.ExpectQuery(`SELECT s\.sku_type, COALESCE\(sp\.model_provider, ''\), COALESCE\(sp\.model_name, ''\), COALESCE\(sp\.provider_model_id, ''\)`).
		WithArgs(1001).
		WillReturnRows(sqlmock.NewRows([]string{"sku_type", "model_provider", "model_name", "provider_model_id"}).
			AddRow("token_pack", "internal", "fuel", ""))

	err = validateEntitlementPackageBundlePolicy(tx, []struct {
		SKUID           int
		DefaultQuantity int
	}{
		{SKUID: 1001, DefaultQuantity: 1},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "加油包不可单独购买")
	require.NoError(t, mock.ExpectationsWereMet())
}
