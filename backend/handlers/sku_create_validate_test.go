package handlers

import (
	"net/http"
	"testing"

	"github.com/pintuotuo/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateSKUCreateRequest_TokenPack(t *testing.T) {
	ok := &models.SKUCreateRequest{SKUType: "token_pack", TokenAmount: 100, ComputePoints: 1}
	assert.Nil(t, validateSKUCreateRequest(ok))

	bad := &models.SKUCreateRequest{SKUType: "token_pack", TokenAmount: 0, ComputePoints: 1}
	e := validateSKUCreateRequest(bad)
	assert.NotNil(t, e)
	assert.Equal(t, http.StatusBadRequest, e.Status)
	assert.Equal(t, "TOKEN_AMOUNT_REQUIRED", e.Code)
}

func TestValidateSKUCreateRequest_Subscription(t *testing.T) {
	bad := &models.SKUCreateRequest{SKUType: "subscription", SubscriptionPeriod: ""}
	e := validateSKUCreateRequest(bad)
	assert.NotNil(t, e)
	assert.Equal(t, "SUBSCRIPTION_PERIOD_REQUIRED", e.Code)

	ok := &models.SKUCreateRequest{SKUType: "subscription", SubscriptionPeriod: "monthly"}
	assert.Nil(t, validateSKUCreateRequest(ok))
}

func TestValidateSKUCreateRequest_Concurrent(t *testing.T) {
	bad := &models.SKUCreateRequest{SKUType: "concurrent", ConcurrentReqs: 0}
	assert.Equal(t, "CONCURRENT_REQUESTS_REQUIRED", validateSKUCreateRequest(bad).Code)
	assert.Nil(t, validateSKUCreateRequest(&models.SKUCreateRequest{SKUType: "concurrent", ConcurrentReqs: 3}))
}

func TestValidateSKUCreateRequest_Trial(t *testing.T) {
	assert.Nil(t, validateSKUCreateRequest(&models.SKUCreateRequest{SKUType: "trial"}))
}

func TestSqlNullableString(t *testing.T) {
	assert.Nil(t, sqlNullableString(""))
	assert.Nil(t, sqlNullableString("   "))
	assert.Equal(t, "monthly", sqlNullableString(" monthly "))
}
