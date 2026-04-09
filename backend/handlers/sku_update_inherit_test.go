package handlers

import (
	"testing"

	"github.com/pintuotuo/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestSKUUpdateRequest_InheritSPUCost_Nil(t *testing.T) {
	req := models.SKUUpdateRequest{}
	assert.Nil(t, req.InheritSPUCost, "InheritSPUCost should be nil when not provided")
}

func TestSKUUpdateRequest_InheritSPUCost_False(t *testing.T) {
	falsy := false
	req := models.SKUUpdateRequest{InheritSPUCost: &falsy}
	assert.NotNil(t, req.InheritSPUCost)
	assert.False(t, *req.InheritSPUCost)
}

func TestSKUUpdateRequest_InheritSPUCost_True(t *testing.T) {
	truthy := true
	req := models.SKUUpdateRequest{InheritSPUCost: &truthy}
	assert.NotNil(t, req.InheritSPUCost)
	assert.True(t, *req.InheritSPUCost)
}

func TestResolveSKUUpdateInherit_Default(t *testing.T) {
	req := models.SKUUpdateRequest{}
	dbInherit := true
	inherit, inputRate, outputRate := resolveSKUUpdateInherit(req, dbInherit, 0.01, 0.02, 0.03, 0.04)
	assert.True(t, inherit, "Should keep DB value when not provided in request")
	assert.Equal(t, 0.01, inputRate)
	assert.Equal(t, 0.02, outputRate)
}

func TestResolveSKUUpdateInherit_ExplicitTrue(t *testing.T) {
	truthy := true
	req := models.SKUUpdateRequest{InheritSPUCost: &truthy}
	dbInherit := false
	inherit, inputRate, outputRate := resolveSKUUpdateInherit(req, dbInherit, 0.01, 0.02, 0.03, 0.04)
	assert.True(t, inherit, "Should use SPU rates when explicitly set to true")
	assert.Equal(t, 0.03, inputRate, "Should use SPU input rate")
	assert.Equal(t, 0.04, outputRate, "Should use SPU output rate")
}

func TestResolveSKUUpdateInherit_ExplicitFalse(t *testing.T) {
	falsy := false
	req := models.SKUUpdateRequest{InheritSPUCost: &falsy, CostInputRate: 0.05, CostOutputRate: 0.06}
	dbInherit := true
	inherit, inputRate, outputRate := resolveSKUUpdateInherit(req, dbInherit, 0.01, 0.02, 0.03, 0.04)
	assert.False(t, inherit, "Should use request value when explicitly set to false")
	assert.Equal(t, 0.05, inputRate, "Should use request input rate")
	assert.Equal(t, 0.06, outputRate, "Should use request output rate")
}

func TestResolveSKUUpdateInherit_ExplicitFalseWithDBRates(t *testing.T) {
	falsy := false
	req := models.SKUUpdateRequest{InheritSPUCost: &falsy}
	dbInherit := false
	inherit, inputRate, outputRate := resolveSKUUpdateInherit(req, dbInherit, 0.01, 0.02, 0.03, 0.04)
	assert.False(t, inherit)
	assert.Equal(t, 0.01, inputRate, "Should keep DB rates when not provided in request")
	assert.Equal(t, 0.02, outputRate)
}
