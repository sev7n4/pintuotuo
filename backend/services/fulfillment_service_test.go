package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateSubscriptionEndFrom_Monthly(t *testing.T) {
	base := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	end := CalculateSubscriptionEndFrom(base, "monthly", 0, 0, "subscription")
	assert.Equal(t, time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC), end.UTC())
}

func TestCalculateSubscriptionEndFrom_ValidDaysWhenPeriodEmpty(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := CalculateSubscriptionEndFrom(base, "", 30, 0, "subscription")
	assert.Equal(t, time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC), end.UTC())
}

func TestCalculateSubscriptionEndFrom_Trial(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := CalculateSubscriptionEndFrom(base, "", 0, 14, "trial")
	assert.Equal(t, time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), end.UTC())
}

func TestValidateSKUForOrder_TokenPack(t *testing.T) {
	assert.NoError(t, ValidateSKUForOrder("token_pack", 100, 0, "", 0, 0))
	assert.Error(t, ValidateSKUForOrder("token_pack", 0, 0, "", 0, 0))
}

func TestValidateSKUForOrder_ComputePoints(t *testing.T) {
	assert.NoError(t, ValidateSKUForOrder("compute_points", 0, 10, "", 0, 0))
	assert.Error(t, ValidateSKUForOrder("compute_points", 0, 0, "", 0, 0))
}

func TestValidateSKUForOrder_Subscription(t *testing.T) {
	assert.NoError(t, ValidateSKUForOrder("subscription", 0, 0, "monthly", 0, 0))
	assert.NoError(t, ValidateSKUForOrder("subscription", 0, 0, "", 365, 0))
	assert.Error(t, ValidateSKUForOrder("subscription", 0, 0, "", 0, 0))
}

func TestEffectiveTokenAmountForFulfillment_PrefersOrder(t *testing.T) {
	order := sql.NullInt64{Valid: true, Int64: 500}
	sku := sql.NullInt64{Valid: true, Int64: 1000}
	got := EffectiveTokenAmountForFulfillment(order, sku)
	assert.True(t, got.Valid)
	assert.Equal(t, int64(500), got.Int64)
}

func TestEffectiveTokenAmountForFulfillment_FallbackSKU(t *testing.T) {
	order := sql.NullInt64{Valid: false}
	sku := sql.NullInt64{Valid: true, Int64: 1000}
	got := EffectiveTokenAmountForFulfillment(order, sku)
	assert.True(t, got.Valid)
	assert.Equal(t, int64(1000), got.Int64)
}

func TestEffectiveSKUTypeForFulfillment_PrefersOrder(t *testing.T) {
	ot := sql.NullString{Valid: true, String: " token_pack "}
	assert.Equal(t, "token_pack", EffectiveSKUTypeForFulfillment(ot, "compute_points"))
}

func TestEffectiveComputePointsForFulfillment_PrefersOrder(t *testing.T) {
	o := sql.NullFloat64{Valid: true, Float64: 12.5}
	s := sql.NullFloat64{Valid: true, Float64: 99}
	got := EffectiveComputePointsForFulfillment(o, s)
	assert.True(t, got.Valid)
	assert.InDelta(t, 12.5, got.Float64, 1e-9)
}
