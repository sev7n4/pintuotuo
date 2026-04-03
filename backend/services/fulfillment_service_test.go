package services

import (
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
