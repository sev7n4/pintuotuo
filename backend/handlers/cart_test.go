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

func TestGetCart_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/cart", nil)

	GetCart(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAddToCart_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"sku_id":   1,
		"quantity": 2,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	AddToCart(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAddToCart_MissingProductID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"quantity": 2,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	AddToCart(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddToCart_MissingQuantity(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"sku_id": 1,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	AddToCart(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddToCart_ZeroQuantity(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"sku_id":   1,
		"quantity": 0,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	AddToCart(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddToCart_NegativeQuantity(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"sku_id":   1,
		"quantity": -1,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	AddToCart(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddToCart_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	AddToCart(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateCartItem_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"quantity": 3,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/cart/items/1", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	UpdateCartItem(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateCartItem_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"quantity": 3,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/cart/items/invalid", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set("user_id", 1)

	UpdateCartItem(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateCartItem_MissingQuantity(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/cart/items/1", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", 1)

	UpdateCartItem(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemoveFromCart_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/cart/items/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	RemoveFromCart(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRemoveFromCart_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/cart/items/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set("user_id", 1)

	RemoveFromCart(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestClearCart_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/cart", nil)

	ClearCart(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAddToCart_WithGroupID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"sku_id":   1,
		"quantity": 2,
		"group_id": 5,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	AddToCart(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}
