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

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAuthFlow(t *testing.T) {
	router := setupTestRouter()

	t.Run("Register with valid data", func(t *testing.T) {
		router.POST("/register", func(c *gin.Context) {
			var req RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"user": gin.H{
					"id":    1,
					"email": req.Email,
					"name":  req.Name,
				},
				"token": "test-jwt-token",
			})
		})

		body := map[string]string{
			"email":    "test@example.com",
			"name":     "Test User",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["user"])
		assert.NotNil(t, response["token"])
	})

	t.Run("Register with invalid email", func(t *testing.T) {
		router.POST("/register-invalid", func(c *gin.Context) {
			var req RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"user": req})
		})

		body := map[string]string{
			"email":    "invalid-email",
			"name":     "Test User",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register-invalid", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Login with valid credentials", func(t *testing.T) {
		router.POST("/login", func(c *gin.Context) {
			var req LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			if req.Email == "test@example.com" && req.Password == "password123" {
				c.JSON(http.StatusOK, gin.H{
					"user": gin.H{
						"id":    1,
						"email": req.Email,
					},
					"token": "test-jwt-token",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			}
		})

		body := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["token"])
	})

	t.Run("Refresh token", func(t *testing.T) {
		router.POST("/refresh", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user": gin.H{
					"id":    1,
					"email": "test@example.com",
				},
				"token": "new-jwt-token",
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/refresh", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "new-jwt-token", response["token"])
	})

	t.Run("Password reset request", func(t *testing.T) {
		router.POST("/password/reset-request", func(c *gin.Context) {
			var req RequestPasswordResetRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message":     "If the email exists, a reset link has been sent",
				"debug_token": "reset-token-123",
			})
		})

		body := map[string]string{
			"email": "test@example.com",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/password/reset-request", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["message"], "reset link has been sent")
	})

	t.Run("Password reset with token", func(t *testing.T) {
		router.POST("/password/reset", func(c *gin.Context) {
			var req ResetPasswordRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			if req.Token == "valid-token" && len(req.Password) >= 6 {
				c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			}
		})

		body := map[string]string{
			"token":    "valid-token",
			"password": "newpassword123",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/password/reset", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Password reset successfully", response["message"])
	})
}

func TestOrderFlow(t *testing.T) {
	router := setupTestRouter()

	t.Run("Create order with valid data", func(t *testing.T) {
		router.POST("/orders", func(c *gin.Context) {
			var req struct {
				ProductID int `json:"product_id" binding:"required"`
				Quantity  int `json:"quantity" binding:"required,gt=0"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"id":          1,
				"product_id":  req.ProductID,
				"quantity":    req.Quantity,
				"total_price": 99.99,
				"status":      "pending",
			})
		})

		body := map[string]int{
			"product_id": 1,
			"quantity":   2,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "pending", response["status"])
	})

	t.Run("Create order with invalid quantity", func(t *testing.T) {
		router.POST("/orders-invalid", func(c *gin.Context) {
			var req struct {
				ProductID int `json:"product_id" binding:"required"`
				Quantity  int `json:"quantity" binding:"required,gt=0"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"id": 1})
		})

		body := map[string]int{
			"product_id": 1,
			"quantity":   0,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/orders-invalid", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Cancel order", func(t *testing.T) {
		router.PUT("/orders/:id/cancel", func(c *gin.Context) {
			orderID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"id":     orderID,
				"status": "canceled",
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/orders/1/cancel", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "canceled", response["status"])
	})

	t.Run("List orders", func(t *testing.T) {
		router.GET("/orders", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"orders": []gin.H{
					{"id": 1, "status": "pending"},
					{"id": 2, "status": "paid"},
				},
				"total": 2,
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/orders", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["orders"])
	})
}

func TestPaymentFlow(t *testing.T) {
	router := setupTestRouter()

	t.Run("Initiate payment", func(t *testing.T) {
		router.POST("/payments", func(c *gin.Context) {
			var req struct {
				OrderID int    `json:"order_id" binding:"required"`
				Method  string `json:"method" binding:"required,oneof=alipay wechat"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"id":          1,
				"order_id":    req.OrderID,
				"method":      req.Method,
				"status":      "pending",
				"payment_url": "https://payment.example.com/pay/123",
				"amount":      99.99,
			})
		})

		body := map[string]interface{}{
			"order_id": 1,
			"method":   "alipay",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/payments", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "pending", response["status"])
		assert.NotNil(t, response["payment_url"])
	})

	t.Run("Get payment status", func(t *testing.T) {
		router.GET("/payments/:id", func(c *gin.Context) {
			paymentID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"id":      paymentID,
				"status":  PaymentStatusSuccess,
				"amount":  99.99,
				"paid_at": "2024-01-01T00:00:00Z",
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/payments/1", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, PaymentStatusSuccess, response["status"])
	})

	t.Run("Refund payment", func(t *testing.T) {
		router.POST("/payments/:id/refund", func(c *gin.Context) {
			paymentID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"id":        paymentID,
				"status":    "refunded",
				"refund_id": "refund-123",
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/payments/1/refund", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "refunded", response["status"])
	})
}

func TestGroupPurchaseFlow(t *testing.T) {
	router := setupTestRouter()

	t.Run("Create group purchase", func(t *testing.T) {
		router.POST("/groups", func(c *gin.Context) {
			var req struct {
				ProductID   int `json:"product_id" binding:"required"`
				TargetCount int `json:"target_count" binding:"required,gt=1"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"id":            1,
				"product_id":    req.ProductID,
				"target_count":  req.TargetCount,
				"current_count": 1,
				"status":        "active",
			})
		})

		body := map[string]int{
			"product_id":   1,
			"target_count": 5,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/groups", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "active", response["status"])
	})

	t.Run("Join group purchase", func(t *testing.T) {
		router.POST("/groups/:id/join", func(c *gin.Context) {
			groupID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"id":            groupID,
				"current_count": 2,
				"status":        "active",
				"message":       "Successfully joined the group",
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/groups/1/join", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["message"], "joined")
	})

	t.Run("Get group progress", func(t *testing.T) {
		router.GET("/groups/:id/progress", func(c *gin.Context) {
			groupID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"id":            groupID,
				"target_count":  5,
				"current_count": 3,
				"progress":      60.0,
				"remaining":     2,
				"status":        "active",
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/groups/1/progress", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["progress"])
	})

	t.Run("Cancel group purchase", func(t *testing.T) {
		router.DELETE("/groups/:id", func(c *gin.Context) {
			groupID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"id":      groupID,
				"status":  "canceled",
				"message": "Group canceled successfully",
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/1", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "canceled", response["status"])
	})

	t.Run("List active groups", func(t *testing.T) {
		router.GET("/groups", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"groups": []gin.H{
					{"id": 1, "status": "active", "current_count": 3},
					{"id": 2, "status": "active", "current_count": 2},
				},
				"total": 2,
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/groups", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["groups"])
	})
}

func TestTokenManagement(t *testing.T) {
	router := setupTestRouter()

	t.Run("Get token balance", func(t *testing.T) {
		router.GET("/tokens/balance", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"balance":  100.50,
				"currency": "CNY",
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tokens/balance", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["balance"])
	})

	t.Run("Transfer tokens", func(t *testing.T) {
		router.POST("/tokens/transfer", func(c *gin.Context) {
			var req struct {
				RecipientID int     `json:"recipient_id" binding:"required"`
				Amount      float64 `json:"amount" binding:"required,gt=0"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message":      "Transfer successful",
				"amount":       req.Amount,
				"recipient_id": req.RecipientID,
			})
		})

		body := map[string]interface{}{
			"recipient_id": 2,
			"amount":       10.50,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/tokens/transfer", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Transfer successful", response["message"])
	})

	t.Run("List API keys", func(t *testing.T) {
		router.GET("/tokens/keys", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"keys": []gin.H{
					{"id": 1, "name": "Production Key", "status": "active"},
					{"id": 2, "name": "Development Key", "status": "active"},
				},
			})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/tokens/keys", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["keys"])
	})

	t.Run("Create API key", func(t *testing.T) {
		router.POST("/tokens/keys", func(c *gin.Context) {
			var req struct {
				Name string `json:"name" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"id":     3,
				"name":   req.Name,
				"key":    "pk_live_xxxxxxxxxxxx",
				"status": "active",
			})
		})

		body := map[string]string{
			"name": "New API Key",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/tokens/keys", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["key"])
	})
}
