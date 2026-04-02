package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCreateRechargeOrder_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"amount": 100.0,
		"method": "alipay",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/recharge", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateRechargeOrder(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateRechargeOrder_MissingAmount(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"method": "alipay",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/recharge", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateRechargeOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRechargeOrder_ZeroAmount(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"amount": 0.0,
		"method": "alipay",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/recharge", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateRechargeOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRechargeOrder_NegativeAmount(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"amount": -100.0,
		"method": "alipay",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/recharge", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateRechargeOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRechargeOrder_MissingMethod(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"amount": 100.0,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/recharge", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateRechargeOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRechargeOrder_InvalidMethod(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"amount": 100.0,
		"method": "invalid",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/recharge", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateRechargeOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRechargeOrder_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/recharge", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateRechargeOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetRechargeOrders_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/tokens/recharge/orders", nil)

	GetRechargeOrders(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetRechargeOrder_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/tokens/recharge/orders/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	GetRechargeOrder(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetRechargeOrder_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/tokens/recharge/orders/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set("user_id", 1)

	GetRechargeOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRechargeCallback_MissingPaymentID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"status": "success",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/recharge/callback", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleRechargeCallback(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRechargeCallback_MissingStatus(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"payment_id": 1,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/recharge/callback", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleRechargeCallback(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMockCompleteRechargeOrder_DisabledWithoutEnv(t *testing.T) {
	t.Setenv("ALLOW_TEST_RECHARGE", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", 1)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/tokens/recharge/orders/1/mock-pay", nil)

	MockCompleteRechargeOrder(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
