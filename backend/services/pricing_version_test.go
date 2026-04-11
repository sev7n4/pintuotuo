package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPricingVersionCodeBaseline(t *testing.T) {
	assert.Equal(t, "baseline", pricingVersionCodeBaseline)
}
