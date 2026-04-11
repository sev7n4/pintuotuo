package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPricingVersionCodeBaseline(t *testing.T) {
	assert.Equal(t, "baseline", pricingVersionCodeBaseline)
}

func TestCostFromPer1KRates(t *testing.T) {
	got := CostFromPer1KRates(0.01, 0.03, 1000, 500)
	assert.InDelta(t, 0.025, got, 1e-9)
}
