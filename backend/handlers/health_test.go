package handlers

import (
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

func TestHealthCheck(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
	HealthCheck(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDBStats_MissingDB(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/db", nil)
	DBStats(c)
	if config.GetDB() == nil {
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	} else {
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestReadyCheck_MissingDB(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	ReadyCheck(c)
	if config.GetDB() == nil {
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	} else {
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestLiveCheck(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/live", nil)
	LiveCheck(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDBStats_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/db", nil)
	DBStats(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReadyCheck_WithDatabase(t *testing.T) {
	if config.GetDB() == nil {
		t.Skip("跳过需要数据库连接的测试")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	ReadyCheck(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
