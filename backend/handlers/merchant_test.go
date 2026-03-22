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

func TestRegisterMerchant_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"company_name": "Test Company",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	RegisterMerchant(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegisterMerchant_MissingCompanyName(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	RegisterMerchant(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterMerchant_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/register", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	RegisterMerchant(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMerchantProfile_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/profile", nil)

	GetMerchantProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantProfile_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/profile", nil)
	c.Set("user_id", "invalid")

	GetMerchantProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMerchantProfile_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{"company_name": "Updated Company"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/merchants/profile", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateMerchantProfile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMerchantProfile_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPut, "/api/merchants/profile", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	UpdateMerchantProfile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMerchantStats_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/stats", nil)

	GetMerchantStats(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantStats_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/stats", nil)
	c.Set("user_id", "invalid")

	GetMerchantStats(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantProducts_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/products", nil)

	GetMerchantProducts(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantProducts_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/products", nil)
	c.Set("user_id", "invalid")

	GetMerchantProducts(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantOrders_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/orders", nil)

	GetMerchantOrders(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantOrders_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/orders", nil)
	c.Set("user_id", "invalid")

	GetMerchantOrders(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantSettlements_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/settlements", nil)

	GetMerchantSettlements(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantSettlements_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/settlements", nil)
	c.Set("user_id", "invalid")

	GetMerchantSettlements(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegisterMerchant_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"company_name":     "Test Company",
		"business_license": "12345678",
		"contact_name":     "Test Contact",
		"contact_phone":    "13800138000",
		"contact_email":    "test@example.com",
		"address":          "Test Address",
		"description":      "Test Description",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/merchants/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	RegisterMerchant(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantProfile_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/profile", nil)
	c.Set("user_id", 1)

	GetMerchantProfile(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateMerchantProfile_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"company_name":  "Updated Company",
		"contact_name":  "Updated Contact",
		"contact_phone": "13900139000",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/merchants/profile", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	UpdateMerchantProfile(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantStats_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/stats", nil)
	c.Set("user_id", 1)

	GetMerchantStats(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantProducts_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/products", nil)
	c.Set("user_id", 1)

	GetMerchantProducts(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantOrders_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/orders", nil)
	c.Set("user_id", 1)

	GetMerchantOrders(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantSettlements_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/settlements", nil)
	c.Set("user_id", 1)

	GetMerchantSettlements(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantProducts_WithPagination(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/products?page=1&per_page=10", nil)
	c.Set("user_id", 1)

	GetMerchantProducts(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetMerchantOrders_WithStatusFilter(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/merchants/orders?status=paid", nil)
	c.Set("user_id", 1)

	GetMerchantOrders(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}
