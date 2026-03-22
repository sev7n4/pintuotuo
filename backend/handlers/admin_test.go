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

func TestGetAdminUsers_MissingRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)

	GetAdminUsers(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAdminUsers_NonAdminRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	c.Set("user_role", "user")

	GetAdminUsers(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAdminUsers_MerchantRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	c.Set("user_role", "merchant")

	GetAdminUsers(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateAdminUser_MissingRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"name":     "Test User",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateAdminUser(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateAdminUser_NonAdminRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"name":     "Test User",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_role", "user")

	CreateAdminUser(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateAdminUser_MissingEmail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"name":     "Test User",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_role", "admin")

	CreateAdminUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateAdminUser_MissingName(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_role", "admin")

	CreateAdminUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateAdminUser_MissingPassword(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"email": "test@example.com",
		"name":  "Test User",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_role", "admin")

	CreateAdminUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateAdminUser_ShortPassword(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"name":     "Test User",
		"password": "short",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_role", "admin")

	CreateAdminUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateAdminUser_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_role", "admin")

	CreateAdminUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAdminStats_MissingRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/stats", nil)

	GetAdminStats(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAdminStats_NonAdminRole(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/stats", nil)
	c.Set("user_role", "user")

	GetAdminStats(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetAdminUsers_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	c.Set("user_role", "admin")

	GetAdminUsers(c)

	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

func TestCreateAdminUser_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]string{
		"email":    "newadmin@example.com",
		"name":     "New Admin",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_role", "admin")

	CreateAdminUser(c)

	assert.NotEqual(t, http.StatusForbidden, w.Code)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestGetAdminStats_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/stats", nil)
	c.Set("user_role", "admin")

	GetAdminStats(c)

	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

func TestCreateAdminUser_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		requestBody    map[string]string
		expectedStatus int
	}{
		{
			name:     "缺少角色",
			userRole: "",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"name":     "Test",
				"password": "password123",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:     "非管理员角色",
			userRole: "user",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"name":     "Test",
				"password": "password123",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:     "缺少邮箱",
			userRole: "admin",
			requestBody: map[string]string{
				"name":     "Test",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "缺少密码",
			userRole: "admin",
			requestBody: map[string]string{
				"email": "test@example.com",
				"name":  "Test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "密码太短",
			userRole: "admin",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"name":     "Test",
				"password": "short",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.userRole != "" {
				c.Set("user_role", tt.userRole)
			}

			CreateAdminUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
