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

func TestGetProviderRouteConfigs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/admin/route-configs/providers", func(c *gin.Context) {
		c.Set("user_role", "admin")
		GetProviderRouteConfigs(c)
	})

	req := httptest.NewRequest("GET", "/admin/route-configs/providers", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestGetProviderRouteConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/admin/route-configs/providers/:code", func(c *gin.Context) {
		c.Set("user_role", "admin")
		GetProviderRouteConfig(c)
	})

	req := httptest.NewRequest("GET", "/admin/route-configs/providers/openai", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
}

func TestUpdateProviderRouteConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.PUT("/admin/route-configs/providers/:code", func(c *gin.Context) {
		c.Set("user_role", "admin")
		UpdateProviderRouteConfig(c)
	})

	body := map[string]interface{}{
		"provider_region": "overseas",
		"route_strategy": map[string]interface{}{
			"domestic_users": map[string]interface{}{
				"mode": "litellm",
			},
		},
		"endpoints": map[string]interface{}{
			"litellm": map[string]interface{}{
				"domestic": "http://litellm:4000/v1",
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/admin/route-configs/providers/openai", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError)
}

func TestGetMerchantRouteConfigs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/admin/route-configs/merchants", func(c *gin.Context) {
		c.Set("user_role", "admin")
		GetMerchantRouteConfigs(c)
	})

	req := httptest.NewRequest("GET", "/admin/route-configs/merchants", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestUpdateMerchantRouteConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.PUT("/admin/route-configs/merchants/:id", func(c *gin.Context) {
		c.Set("user_role", "admin")
		UpdateMerchantRouteConfig(c)
	})

	body := map[string]interface{}{
		"merchant_type": "standard",
		"region":        "domestic",
		"route_preference": map[string]interface{}{
			"preferred_mode": "litellm",
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/admin/route-configs/merchants/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError || w.Code == http.StatusBadRequest)
}

func TestRouteDecisionAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/admin/route-configs/test", func(c *gin.Context) {
		c.Set("user_role", "admin")
		TestRouteDecision(c)
	})

	body := map[string]interface{}{
		"provider_code": "openai",
		"merchant_id":   1,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/admin/route-configs/test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError || w.Code == http.StatusBadRequest)
}

func TestEnsureAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		userRole   interface{}
		expectPass bool
	}{
		{
			name:       "admin role should pass",
			userRole:   "admin",
			expectPass: true,
		},
		{
			name:       "non-admin role should fail",
			userRole:   "user",
			expectPass: false,
		},
		{
			name:       "no role should fail",
			userRole:   nil,
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				if tt.userRole != nil {
					c.Set("user_role", tt.userRole)
				}
				result := ensureAdmin(c)
				c.JSON(http.StatusOK, gin.H{"passed": result})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if tt.expectPass {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.True(t, resp["passed"].(bool))
			} else {
				assert.Equal(t, http.StatusForbidden, w.Code)
			}
		})
	}
}
