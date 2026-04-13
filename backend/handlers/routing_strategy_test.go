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

func TestAdminGetRoutingStrategies_MissingRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/routing-strategies", nil)

	AdminGetRoutingStrategies(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminGetRoutingStrategies_NonAdminRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/routing-strategies", nil)
	c.Set("user_role", "user")

	AdminGetRoutingStrategies(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminGetRoutingStrategies_MerchantRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/routing-strategies", nil)
	c.Set("user_role", "merchant")

	AdminGetRoutingStrategies(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminGetRoutingStrategy_MissingRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/routing-strategies/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	AdminGetRoutingStrategy(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminGetRoutingStrategy_NonAdminRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/routing-strategies/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_role", "user")

	AdminGetRoutingStrategy(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminGetRoutingStrategy_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/routing-strategies/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set("user_role", "admin")

	AdminGetRoutingStrategy(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminCreateRoutingStrategy_MissingRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"name":                      "Test Strategy",
		"code":                      "test_strategy",
		"price_weight":              0.33,
		"latency_weight":            0.34,
		"reliability_weight":        0.33,
		"max_retry_count":           3,
		"retry_backoff_base":        1000,
		"circuit_breaker_threshold": 5,
		"circuit_breaker_timeout":   60,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/routing-strategies", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	AdminCreateRoutingStrategy(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminCreateRoutingStrategy_NonAdminRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"name":               "Test Strategy",
		"code":               "test_strategy",
		"price_weight":       0.33,
		"latency_weight":     0.34,
		"reliability_weight": 0.33,
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/routing-strategies", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_role", "user")

	AdminCreateRoutingStrategy(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminCreateRoutingStrategy_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/routing-strategies", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_role", "admin")

	AdminCreateRoutingStrategy(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminUpdateRoutingStrategy_MissingRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"name": "Updated Strategy",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/routing-strategies/1", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	AdminUpdateRoutingStrategy(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminUpdateRoutingStrategy_NonAdminRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"name": "Updated Strategy",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/routing-strategies/1", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_role", "merchant")

	AdminUpdateRoutingStrategy(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminUpdateRoutingStrategy_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"name": "Updated Strategy",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/routing-strategies/invalid", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set("user_role", "admin")

	AdminUpdateRoutingStrategy(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminUpdateRoutingStrategy_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/routing-strategies/1", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_role", "admin")

	AdminUpdateRoutingStrategy(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminDeleteRoutingStrategy_MissingRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/admin/routing-strategies/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	AdminDeleteRoutingStrategy(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminDeleteRoutingStrategy_NonAdminRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/admin/routing-strategies/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_role", "user")

	AdminDeleteRoutingStrategy(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminDeleteRoutingStrategy_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodDelete, "/api/admin/routing-strategies/invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set("user_role", "admin")

	AdminDeleteRoutingStrategy(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminGetRoutingStrategies_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/routing-strategies?page=1&page_size=10", nil)
	c.Set("user_role", "admin")

	AdminGetRoutingStrategies(c)

	assert.NotEqual(t, http.StatusForbidden, w.Code)
	if w.Code == http.StatusOK {
		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Contains(t, body, "strategies")
	}
}

func TestAdminGetRoutingStrategy_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/routing-strategies/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_role", "admin")

	AdminGetRoutingStrategy(c)

	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

func TestAdminRoutingStrategies_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		method         string
		path           string
		body           interface{}
		params         gin.Params
		expectedStatus int
	}{
		{
			name:           "获取列表-缺少角色",
			userRole:       "",
			method:         http.MethodGet,
			path:           "/api/admin/routing-strategies",
			body:           nil,
			params:         nil,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "获取列表-非管理员",
			userRole:       "user",
			method:         http.MethodGet,
			path:           "/api/admin/routing-strategies",
			body:           nil,
			params:         nil,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "创建-缺少角色",
			userRole:       "",
			method:         http.MethodPost,
			path:           "/api/admin/routing-strategies",
			body:           map[string]interface{}{"name": "Test"},
			params:         nil,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "更新-缺少角色",
			userRole:       "",
			method:         http.MethodPut,
			path:           "/api/admin/routing-strategies/1",
			body:           map[string]interface{}{"name": "Test"},
			params:         gin.Params{{Key: "id", Value: "1"}},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "删除-缺少角色",
			userRole:       "",
			method:         http.MethodDelete,
			path:           "/api/admin/routing-strategies/1",
			body:           nil,
			params:         gin.Params{{Key: "id", Value: "1"}},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body bytes.Buffer
			if tt.body != nil {
				b, _ := json.Marshal(tt.body)
				body = *bytes.NewBuffer(b)
			}

			c.Request = httptest.NewRequest(tt.method, tt.path, &body)
			c.Request.Header.Set("Content-Type", "application/json")
			if tt.params != nil {
				c.Params = tt.params
			}

			if tt.userRole != "" {
				c.Set("user_role", tt.userRole)
			}

			switch tt.method {
			case http.MethodGet:
				if len(tt.params) > 0 {
					AdminGetRoutingStrategy(c)
				} else {
					AdminGetRoutingStrategies(c)
				}
			case http.MethodPost:
				AdminCreateRoutingStrategy(c)
			case http.MethodPut:
				AdminUpdateRoutingStrategy(c)
			case http.MethodDelete:
				AdminDeleteRoutingStrategy(c)
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
