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

func TestGetBYOKRoutingList_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/byok-routing", nil)

	GetBYOKRoutingList(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetBYOKRoutingList_InvalidUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/byok-routing", nil)
	c.Set("user_id", "invalid")

	GetBYOKRoutingList(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateBYOKRouteConfig_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"route_mode": "direct",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/byok-routing/1/config", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateBYOKRouteConfig(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateBYOKRouteConfig_InvalidKeyID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"route_mode": "direct",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/byok-routing/invalid/config", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	UpdateBYOKRouteConfig(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLightVerifyBYOK_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/byok-routing/1/light-verify", nil)

	LightVerifyBYOK(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLightVerifyBYOK_InvalidKeyID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/byok-routing/invalid/light-verify", nil)
	c.Set("user_id", 1)

	LightVerifyBYOK(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeepVerifyBYOK_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/byok-routing/1/deep-verify", nil)

	DeepVerifyBYOK(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeepVerifyBYOK_InvalidKeyID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/byok-routing/invalid/deep-verify", nil)
	c.Set("user_id", 1)

	DeepVerifyBYOK(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTriggerBYOKProbe_MissingUserID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/byok-routing/1/probe", nil)

	TriggerBYOKProbe(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTriggerBYOKProbe_InvalidKeyID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/byok-routing/invalid/probe", nil)
	c.Set("user_id", 1)

	TriggerBYOKProbe(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateBYOKRouteConfig_InvalidRouteMode(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"route_mode": "invalid_mode",
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/byok-routing/1/config", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	UpdateBYOKRouteConfig(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateBYOKRouteConfig_ValidRouteModes(t *testing.T) {
	validModes := []string{"direct", "litellm", "proxy", "auto"}

	for _, mode := range validModes {
		t.Run("mode_"+mode, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			reqBody := map[string]interface{}{
				"route_mode": mode,
			}
			body, _ := json.Marshal(reqBody)
			c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/byok-routing/1/config", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("user_id", 1)

			UpdateBYOKRouteConfig(c)

			assert.NotEqual(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestUpdateBYOKRouteConfig_ValidRouteConfig(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := map[string]interface{}{
		"route_mode": "direct",
		"route_config": map[string]interface{}{
			"endpoint_url": "https://api.openai.com",
			"endpoints": map[string]interface{}{
				"direct": map[string]interface{}{
					"overseas": "https://api.openai.com",
					"domestic": "https://api.openai-proxy.com",
				},
			},
		},
	}
	body, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/byok-routing/1/config", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)

	UpdateBYOKRouteConfig(c)

	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}
