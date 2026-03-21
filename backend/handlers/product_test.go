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

func TestCreateProduct_NegativePrice_ReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"name":        "Test Product",
		"description": "Test Description",
		"price":       -10.0,
		"stock":       100,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateProduct(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "价格必须大于0")
}

func TestCreateProduct_ZeroPrice_ReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"name":        "Test Product",
		"description": "Test Description",
		"price":       0,
		"stock":       100,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateProduct(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "价格必须大于0")
}

func TestCreateProduct_NegativeStock_ReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"name":        "Test Product",
		"description": "Test Description",
		"price":       99.9,
		"stock":       -10,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateProduct(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "库存不能为负数")
}
