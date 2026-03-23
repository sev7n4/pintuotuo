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

func TestCreateMerchantAPIKey_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{"name": "Test Key"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/api-keys", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateMerchantAPIKey(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateMerchantAPIKey_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{"name": "Test Key"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/api-keys", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "invalid")

	CreateMerchantAPIKey(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateMerchantAPIKey_MissingName(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/api-keys", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateMerchantAPIKey(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateMerchantAPIKey_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/api-keys", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateMerchantAPIKey(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListMerchantAPIKeys_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/api-keys", nil)

	ListMerchantAPIKeys(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListMerchantAPIKeys_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/api-keys", nil)
	c.Set("user_id", "invalid")

	ListMerchantAPIKeys(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMerchantAPIKey_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{"name": "Updated Key"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/merchants/api-keys/1", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateMerchantAPIKey(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMerchantAPIKey_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPut, "/api/merchants/api-keys/1", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	UpdateMerchantAPIKey(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteMerchantAPIKey_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/merchants/api-keys/1", nil)

	DeleteMerchantAPIKey(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteMerchantAPIKey_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/merchants/api-keys/1", nil)
	c.Set("user_id", "invalid")

	DeleteMerchantAPIKey(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantAPIKeyUsage_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/api-keys/1/usage", nil)

	GetMerchantAPIKeyUsage(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantAPIKeyUsage_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/api-keys/1/usage", nil)
	c.Set("user_id", "invalid")

	GetMerchantAPIKeyUsage(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequestSettlement_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{"amount": 100.0}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/settlements/request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	RequestSettlement(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequestSettlement_InvalidJSON(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/settlements/request", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	RequestSettlement(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetSettlementDetail_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/settlements/1", nil)

	GetSettlementDetail(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSettlementDetail_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/settlements/1", nil)
	c.Set("user_id", "invalid")

	GetSettlementDetail(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateMerchantAPIKey_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{"name": "Test API Key"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/api-keys", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateMerchantAPIKey(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestListMerchantAPIKeys_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/api-keys", nil)
	c.Set("user_id", 1)

	ListMerchantAPIKeys(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMerchantAPIKey_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{"name": "Updated Key", "is_active": true}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/merchants/api-keys/1", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	UpdateMerchantAPIKey(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteMerchantAPIKey_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/merchants/api-keys/1", nil)
	c.Set("user_id", 1)

	DeleteMerchantAPIKey(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantAPIKeyUsage_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/api-keys/1/usage", nil)
	c.Set("user_id", 1)

	GetMerchantAPIKeyUsage(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestRequestSettlement_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{"amount": 100.0}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/settlements/request", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	RequestSettlement(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetSettlementDetail_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/settlements/1", nil)
	c.Set("user_id", 1)

	GetSettlementDetail(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}
