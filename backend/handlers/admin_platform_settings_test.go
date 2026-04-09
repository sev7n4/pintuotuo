package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetAdminPlatformSettings_ForbiddenNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/admin/platform-settings", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_role", "user")
	GetAdminPlatformSettings(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestUpdateAdminPlatformSettings_ForbiddenNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(map[string]interface{}{
		"health_scheduler_enabled":          false,
		"health_scheduler_interval_seconds": 3600,
		"health_scheduler_batch":            2,
	})
	req := httptest.NewRequest(http.MethodPut, "/admin/platform-settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user_role", "merchant")
	UpdateAdminPlatformSettings(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
