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

func TestCreateOrder_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"product_id": 1,
		"quantity":   2,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateOrder(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateOrder_MissingProductID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"quantity": 2,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_MissingQuantity(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"product_id": 1,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_ZeroQuantity(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"product_id": 1,
		"quantity":   0,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_NegativeQuantity(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"product_id": 1,
		"quantity":   -1,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateOrder_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListOrders_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/orders", nil)

	ListOrders(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetOrderByID_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/orders/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	GetOrderByID(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetOrderByID_InvalidID(t *testing.T) {
	t.Skip("Skipping - requires database connection for proper validation")
}

func TestCancelOrder_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPut, "/api/orders/1/cancel", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	CancelOrder(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCancelOrder_InvalidID(t *testing.T) {
	t.Skip("Skipping - requires database connection for proper validation")
}

func TestCreateGroup_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"product_id":   1,
		"target_count": 5,
		"deadline":     "2025-12-31T23:59:59Z",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateGroup(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateGroup_MissingProductID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"target_count": 5,
		"deadline":     "2025-12-31T23:59:59Z",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateGroup_MissingTargetCount(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"product_id": 1,
		"deadline":   "2025-12-31T23:59:59Z",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateGroup_ZeroTargetCount(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"product_id":   1,
		"target_count": 0,
		"deadline":     "2025-12-31T23:59:59Z",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateGroup_MissingDeadline(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"product_id":   1,
		"target_count": 5,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/groups", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetGroupByID_InvalidID(t *testing.T) {
	t.Skip("Skipping - requires database connection for proper validation")
}

func TestJoinGroup_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/groups/1/join", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	JoinGroup(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJoinGroup_InvalidID(t *testing.T) {
	t.Skip("Skipping - requires database connection for proper validation")
}

func TestCancelGroup_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/groups/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	CancelGroup(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCancelGroup_InvalidID(t *testing.T) {
	t.Skip("Skipping - requires database connection for proper validation")
}

func TestGetGroupProgress_InvalidID(t *testing.T) {
	t.Skip("Skipping - requires database connection for proper validation")
}

func TestCreateOrder_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"product_id": 1,
		"quantity":   2,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	CreateOrder(c)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestListOrders_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/orders?page=1&per_page=20", nil)
	c.Set("user_id", 1)

	ListOrders(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "data")
}

func TestListGroups_DefaultParams(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/groups?page=1&per_page=20&status=active", nil)

	ListGroups(c)

	if config.GetDB() == nil {
		assert.NotEqual(t, http.StatusOK, w.Code)
	} else {
		assert.Equal(t, http.StatusOK, w.Code)
	}
}
