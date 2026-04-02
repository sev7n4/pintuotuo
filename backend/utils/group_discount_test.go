package utils

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeGroupDiscountRate(t *testing.T) {
	assert.Equal(t, 0.0, NormalizeGroupDiscountRate(0))
	assert.Equal(t, 0.0, NormalizeGroupDiscountRate(-1))
	assert.InDelta(t, 0.2, NormalizeGroupDiscountRate(0.2), 1e-9)
	assert.InDelta(t, 0.2, NormalizeGroupDiscountRate(20), 1e-9)
	assert.InDelta(t, 1.0, NormalizeGroupDiscountRate(100), 1e-9)
	assert.InDelta(t, 1.0, NormalizeGroupDiscountRate(200), 1e-9)
}

func TestNormalizeGroupDiscountRateNull(t *testing.T) {
	assert.Equal(t, 0.0, NormalizeGroupDiscountRateNull(sql.NullFloat64{Valid: false}))
	assert.InDelta(t, 0.25, NormalizeGroupDiscountRateNull(sql.NullFloat64{Float64: 25, Valid: true}), 1e-9)
}
