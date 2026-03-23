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

func TestGetNotifications_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	GetNotifications(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetNotifications_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	c.Set("user_id", "invalid")
	GetNotifications(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUnreadCount_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	GetUnreadCount(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUnreadCount_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	c.Set("user_id", "invalid")
	GetUnreadCount(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMarkNotificationRead_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/notifications/1/read", nil)
	MarkNotificationRead(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMarkNotificationRead_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/notifications/1/read", nil)
	c.Set("user_id", "invalid")
	MarkNotificationRead(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMarkAllNotificationsRead_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/notifications/read-all", nil)
	MarkAllNotificationsRead(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMarkAllNotificationsRead_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/notifications/read-all", nil)
	c.Set("user_id", "invalid")
	MarkAllNotificationsRead(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegisterDeviceToken_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := map[string]string{"token": "test-token", "platform": "ios"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/notifications/device-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	RegisterDeviceToken(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegisterDeviceToken_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := map[string]string{"token": "test-token", "platform": "ios"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/notifications/device-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "invalid")
	RegisterDeviceToken(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegisterDeviceToken_MissingToken(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := map[string]string{"platform": "ios"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/notifications/device-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)
	RegisterDeviceToken(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterDeviceToken_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/notifications/device-token", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)
	RegisterDeviceToken(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUnregisterDeviceToken_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := map[string]string{"token": "test-token"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/notifications/device-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	UnregisterDeviceToken(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUnregisterDeviceToken_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := map[string]string{"token": "test-token"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/notifications/device-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "invalid")
	UnregisterDeviceToken(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUnregisterDeviceToken_MissingToken(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/notifications/device-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)
	UnregisterDeviceToken(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUnregisterDeviceToken_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/notifications/device-token", bytes.NewBuffer([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)
	UnregisterDeviceToken(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetNotifications_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	c.Set("user_id", 1)
	GetNotifications(c)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetUnreadCount_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/notifications/unread-count", nil)
	c.Set("user_id", 1)
	GetUnreadCount(c)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestMarkNotificationRead_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/notifications/1/read", nil)
	c.Set("user_id", 1)
	MarkNotificationRead(c)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestMarkAllNotificationsRead_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/notifications/read-all", nil)
	c.Set("user_id", 1)
	MarkAllNotificationsRead(c)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestRegisterDeviceToken_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := map[string]string{"token": "test-device-token", "platform": "ios"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/notifications/device-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)
	RegisterDeviceToken(c)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestUnregisterDeviceToken_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := map[string]string{"token": "test-device-token"}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/notifications/device-token", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)
	UnregisterDeviceToken(c)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestGetNotifications_WithPagination(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/notifications?page=1&per_page=10", nil)
	c.Set("user_id", 1)
	GetNotifications(c)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}
