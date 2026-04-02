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

func TestCreateProduct_ReturnsGone(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"name":        "Test Product",
		"description": "Test Description",
		"price":       10.0,
		"stock":       100,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/catalog/merchants", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateProduct(c)

	assert.Equal(t, http.StatusGone, w.Code)
	assert.Contains(t, w.Body.String(), "DEPRECATED")
	assert.Contains(t, w.Body.String(), "merchants/skus")
}
