package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCreatePaymentV2_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"order_id": 1,
		"method":   "alipay",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v2/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	CreatePayment(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreatePaymentV2_MissingOrderID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"method": "alipay",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v2/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreatePayment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePaymentV2_MissingMethod(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"order_id": 1,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v2/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreatePayment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePaymentV2_InvalidMethod(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"order_id": 1,
		"method":   "invalid_method",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v2/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreatePayment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePaymentV2_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/v2/payments", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreatePayment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPaymentStatus_MissingUserID(t *testing.T) {
	t.Skip("Skipping - GetPaymentStatus does not require user_id")
}

func TestGetPaymentStatus_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/v2/payments/invalid/status", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set("user_id", 1)

	GetPaymentStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAlipayNotify_InvalidContentType(t *testing.T) {
	t.Skip("Skipping - requires payment service initialization")
}

func TestWechatNotify_InvalidContentType(t *testing.T) {
	t.Skip("Skipping - requires payment service initialization")
}

func TestCreatePaymentV2_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"order_id": 1,
		"method":   "alipay",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v2/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreatePayment(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestCreatePaymentV2_BalanceMethod(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"order_id": 1,
		"method":   "balance",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v2/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreatePayment(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestCreatePaymentV2_WechatMethod(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"order_id": 1,
		"method":   "wechat",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v2/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreatePayment(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestGetPaymentStatus_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/v2/payments/1/status", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", 1)

	GetPaymentStatus(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}
