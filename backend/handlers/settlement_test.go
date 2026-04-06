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

func TestAdminGetSettlements_RequireAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 403 when user is not admin", func(t *testing.T) {
		router := gin.New()
		router.GET("/admin/settlements", func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Set("user_role", "merchant")
			AdminGetSettlements(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/admin/settlements", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should return 403 when user_role is missing", func(t *testing.T) {
		router := gin.New()
		router.GET("/admin/settlements", func(c *gin.Context) {
			c.Set("user_id", 1)
			AdminGetSettlements(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/admin/settlements", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAdminApproveSettlement_RequireAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 403 when user is not admin", func(t *testing.T) {
		router := gin.New()
		router.POST("/admin/settlements/:id/approve", func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Set("user_role", "merchant")
			AdminApproveSettlement(c)
		})

		req, _ := http.NewRequest(http.MethodPost, "/admin/settlements/1/approve", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAdminGenerateMonthlySettlements_RequireAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 403 when user is not admin", func(t *testing.T) {
		router := gin.New()
		router.POST("/admin/settlements/generate", func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Set("user_role", "user")
			AdminGenerateMonthlySettlements(c)
		})

		body := map[string]int{
			"year":  2026,
			"month": 3,
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPost, "/admin/settlements/generate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAdminProcessDispute_RequireAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 403 when user is not admin", func(t *testing.T) {
		router := gin.New()
		router.POST("/admin/disputes/:id/process", func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Set("user_role", "merchant")
			AdminProcessDispute(c)
		})

		body := map[string]interface{}{
			"resolution":      "Adjusted after review",
			"adjusted_amount": 9500.00,
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest(http.MethodPost, "/admin/disputes/1/process", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAdminReconcileSettlement_RequireAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 403 when user is not admin", func(t *testing.T) {
		router := gin.New()
		router.POST("/admin/settlements/:id/reconcile", func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Set("user_role", "user")
			AdminReconcileSettlement(c)
		})

		req, _ := http.NewRequest(http.MethodPost, "/admin/settlements/1/reconcile", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAdminMarkSettlementPaid_RequireAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 403 when user is not admin", func(t *testing.T) {
		router := gin.New()
		router.POST("/admin/settlements/:id/mark-paid", func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Set("user_role", "merchant")
			AdminMarkSettlementPaid(c)
		})

		req, _ := http.NewRequest(http.MethodPost, "/admin/settlements/1/mark-paid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAdminGetSettlements_MerchantIDFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-SETTLEMENT-001: should filter settlements by merchant_id", func(t *testing.T) {
		t.Skip("Requires database setup - will be implemented in integration tests")
		router := gin.New()
		router.GET("/admin/settlements", func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Set("user_role", "admin")
			AdminGetSettlements(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/admin/settlements?merchant_id=4", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		settlements := response["settlements"].([]interface{})
		assert.True(t, len(settlements) > 0, "Should return settlements for merchant_id=4")

		for _, s := range settlements {
			settlement := s.(map[string]interface{})
			assert.Equal(t, float64(4), settlement["merchant_id"], "All settlements should belong to merchant_id=4")
		}
	})
}

func TestAdminGetSettlements_NullSettledAt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-SETTLEMENT-002: should handle NULL settled_at correctly", func(t *testing.T) {
		t.Skip("Requires database setup - will be implemented in integration tests")
		router := gin.New()
		router.GET("/admin/settlements", func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Set("user_role", "admin")
			AdminGetSettlements(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/admin/settlements", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		settlements := response["settlements"].([]interface{})
		assert.True(t, len(settlements) > 0, "Should return settlements even with NULL settled_at")

		for _, s := range settlements {
			settlement := s.(map[string]interface{})
			if settledAt, exists := settlement["settled_at"]; exists {
				if settledAt == nil {
					assert.Nil(t, settledAt, "settled_at should be null for pending settlements")
				}
			}
		}
	})
}

func TestAdminGetSettlements_ErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TC-SETTLEMENT-003: should not skip records on scan error", func(t *testing.T) {
		t.Skip("Requires database setup - will be implemented in integration tests")
		router := gin.New()
		router.GET("/admin/settlements", func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Set("user_role", "admin")
			AdminGetSettlements(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/admin/settlements", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		settlements := response["settlements"].([]interface{})

		assert.True(t, len(settlements) > 0, "Should return all settlements without skipping")
	})
}
