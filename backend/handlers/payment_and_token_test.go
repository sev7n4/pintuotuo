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

func TestInitiatePayment_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"order_id": 1,
		"method":   "alipay",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	InitiatePayment(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInitiatePayment_MissingOrderID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"method": "alipay",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	InitiatePayment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInitiatePayment_MissingMethod(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"order_id": 1,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	InitiatePayment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInitiatePayment_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	InitiatePayment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPaymentByID_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/payments/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	GetPaymentByID(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetPaymentByID_InvalidID(t *testing.T) {
	t.Skip("Skipping - requires database connection for proper validation")
}

func TestRefundPayment_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/payments/1/refund", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	RefundPayment(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefundPayment_InvalidID(t *testing.T) {
	t.Skip("Skipping - requires database connection for proper validation")
}

func TestGetTokenBalance_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/tokens/balance", nil)

	GetTokenBalance(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTokenConsumption_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/tokens/consumption", nil)

	GetTokenConsumption(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTransferTokens_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"recipient_id": 2,
		"amount":       100.0,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/transfer", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	TransferTokens(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTransferTokens_MissingRecipientID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"amount": 100.0,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/transfer", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	TransferTokens(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransferTokens_MissingAmount(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"recipient_id": 2,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/transfer", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	TransferTokens(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransferTokens_ZeroAmount(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"recipient_id": 2,
		"amount":       0.0,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/transfer", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	TransferTokens(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransferTokens_NegativeAmount(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"recipient_id": 2,
		"amount":       -100.0,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/transfer", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	TransferTokens(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransferTokens_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/tokens/transfer", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	TransferTokens(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTokenBalance_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/tokens/balance", nil)
	c.Set("user_id", 1)

	GetTokenBalance(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetTokenConsumption_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/tokens/consumption", nil)
	c.Set("user_id", 1)

	GetTokenConsumption(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}
