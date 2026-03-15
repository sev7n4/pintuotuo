package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/payment"
	"github.com/pintuotuo/backend/tests/integration"
)

// SetupPaymentRouter creates a test router with payment endpoints
func SetupPaymentRouter(t *testing.T, ts *integration.TestServices) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(middleware.ErrorHandlingMiddleware())

	// Setup payment endpoints
	paymentGroup := router.Group("/api/v1")

	paymentGroup.POST("/payments", func(c *gin.Context) {
		// Mock auth middleware - set user ID
		c.Set("userID", 9999) // Will be set by test

		var req payment.InitiatePaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetInt("userID")
		p, err := ts.PaymentService.InitiatePayment(c.Request.Context(), userID, &req)
		if err != nil {
			if appErr, ok := err.(*apperrors.AppError); ok {
				c.JSON(appErr.Status, appErr)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, p)
	})

	paymentGroup.GET("/payments/:id", func(c *gin.Context) {
		c.Set("userID", 9999)

		paymentID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
			return
		}

		userID := c.GetInt("userID")
		p, err := ts.PaymentService.GetPaymentByID(c.Request.Context(), userID, paymentID)
		if err != nil {
			if appErr, ok := err.(*apperrors.AppError); ok {
				c.JSON(appErr.Status, appErr)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, p)
	})

	paymentGroup.GET("/payments", func(c *gin.Context) {
		c.Set("userID", 9999)

		page := 1
		perPage := 20
		if p := c.Query("page"); p != "" {
			page, _ = strconv.Atoi(p)
		}
		if pp := c.Query("per_page"); pp != "" {
			perPage, _ = strconv.Atoi(pp)
		}

		params := &payment.ListPaymentsParams{
			Page:    page,
			PerPage: perPage,
			Status:  c.Query("status"),
			Method:  c.Query("method"),
		}

		userID := c.GetInt("userID")
		result, err := ts.PaymentService.ListPayments(c.Request.Context(), userID, params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	paymentGroup.POST("/payments/:id/refund", func(c *gin.Context) {
		c.Set("userID", 9999)

		paymentID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
			return
		}

		var req payment.RefundRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetInt("userID")
		p, err := ts.PaymentService.RefundPayment(c.Request.Context(), userID, paymentID, req.Reason)
		if err != nil {
			if appErr, ok := err.(*apperrors.AppError); ok {
				c.JSON(appErr.Status, appErr)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, p)
	})

	// Webhook endpoints (no auth required)
	paymentGroup.POST("/webhooks/alipay", func(c *gin.Context) {
		var callback payment.AlipayCallback
		if err := c.ShouldBindJSON(&callback); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		p, err := ts.PaymentService.HandleAlipayCallback(c.Request.Context(), &callback)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook processing failed"})
			return
		}

		c.JSON(http.StatusOK, p)
	})

	paymentGroup.POST("/webhooks/wechat", func(c *gin.Context) {
		var callback payment.WechatCallback
		if err := c.ShouldBindJSON(&callback); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		p, err := ts.PaymentService.HandleWechatCallback(c.Request.Context(), &callback)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook processing failed"})
			return
		}

		c.JSON(http.StatusOK, p)
	})

	return router
}

// TestInitiatePaymentEndpoint tests POST /api/v1/payments
func TestInitiatePaymentEndpoint(t *testing.T) {
	t.Parallel()
	ts := integration.SetupPaymentTest(t)
	defer integration.TeardownPaymentTest(t, ts)

	router := SetupPaymentRouter(t, ts)

	userID := integration.SeedTestUser(t, ts.DB, 10)
	productID := integration.SeedTestProduct(t, ts.DB, 10)
	orderID := integration.SeedTestOrder(t, ts.DB, userID, productID)

	defer integration.CleanupTestData(t, ts.DB, userID)

	// Create request
	reqBody := payment.InitiatePaymentRequest{
		OrderID:       orderID,
		PaymentMethod: "alipay",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code, "Should return 201 Created")

	var resp payment.Payment
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Greater(t, resp.ID, 0, "Response should contain payment ID")
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, orderID, resp.OrderID)
	assert.Equal(t, "pending", resp.Status)

	// Verify database record created
	integration.AssertPaymentStatus(t, ts.DB, resp.ID, "pending")
}

// TestGetPaymentEndpoint tests GET /api/v1/payments/:id
func TestGetPaymentEndpoint(t *testing.T) {
	t.Parallel()
	ts := integration.SetupPaymentTest(t)
	defer integration.TeardownPaymentTest(t, ts)

	router := SetupPaymentRouter(t, ts)

	userID, _, paymentID := integration.CreateTestPaymentFlow(t, ts)
	defer integration.CleanupTestData(t, ts.DB, userID)

	// Make request
	req := httptest.NewRequest("GET", "/api/v1/payments/"+strconv.Itoa(paymentID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

	var resp payment.Payment
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, paymentID, resp.ID)
	assert.Equal(t, "pending", resp.Status)
}

// TestListPaymentsEndpoint tests GET /api/v1/payments with pagination
func TestListPaymentsEndpoint(t *testing.T) {
	t.Parallel()
	ts := integration.SetupPaymentTest(t)
	defer integration.TeardownPaymentTest(t, ts)

	router := SetupPaymentRouter(t, ts)

	userID := integration.SeedTestUser(t, ts.DB, 11)
	productID := integration.SeedTestProduct(t, ts.DB, 11)
	defer integration.CleanupTestData(t, ts.DB, userID)

	// Create 5 test payments
	for i := 0; i < 5; i++ {
		orderID := integration.SeedTestOrder(t, ts.DB, userID, productID)
		req := &payment.InitiatePaymentRequest{
			OrderID:       orderID,
			PaymentMethod: "alipay",
		}
		_, err := ts.PaymentService.InitiatePayment(context.Background(), userID, req)
		require.NoError(t, err)
	}

	// Make list request
	httpReq := httptest.NewRequest("GET", "/api/v1/payments?page=1&per_page=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp payment.PaymentListResult
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PerPage)
	assert.GreaterOrEqual(t, resp.Total, 5, "Should have at least 5 payments")
	assert.GreaterOrEqual(t, len(resp.Data), 5, "Should return at least 5 payments")
}

// TestRefundPaymentEndpoint tests POST /api/v1/payments/:id/refund
func TestRefundPaymentEndpoint(t *testing.T) {
	t.Parallel()
	ts := integration.SetupPaymentTest(t)
	defer integration.TeardownPaymentTest(t, ts)

	router := SetupPaymentRouter(t, ts)

	userID, _, paymentID := integration.CreateTestPaymentFlow(t, ts)
	defer integration.CleanupTestData(t, ts.DB, userID)

	// First, complete the payment
	ctx := context.Background()
	integration.SimulateAlipayCallback(t, ctx, ts.DB, ts.PaymentService, paymentID)

	// Create refund request
	refundReq := payment.RefundRequest{
		PaymentID: paymentID,
		Reason:    "Customer requested refund",
	}
	body, _ := json.Marshal(refundReq)

	req := httptest.NewRequest("POST", "/api/v1/payments/"+strconv.Itoa(paymentID)+"/refund", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

	var resp payment.Payment
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "refunded", resp.Status)

	// Verify database updated
	integration.AssertPaymentStatus(t, ts.DB, paymentID, "refunded")
}

// TestAlipayCallbackEndpoint tests POST /api/v1/webhooks/alipay
func TestAlipayCallbackEndpoint(t *testing.T) {
	t.Parallel()
	ts := integration.SetupPaymentTest(t)
	defer integration.TeardownPaymentTest(t, ts)

	router := SetupPaymentRouter(t, ts)

	userID, orderID, paymentID := integration.CreateTestPaymentFlow(t, ts)
	defer integration.CleanupTestData(t, ts.DB, userID)

	p := integration.GetPaymentFromDB(t, ts.DB, paymentID)

	// Create callback
	callback := payment.AlipayCallback{
		OutTradeNo:  strconv.Itoa(paymentID),
		TradeNo:     "alipay_test_123456",
		TotalAmount: p.Amount,
		TradeStatus: "TRADE_SUCCESS",
		Timestamp:   "2026-03-15 12:00:00",
		Sign:        "test_signature",
	}
	body, _ := json.Marshal(callback)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/alipay", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp payment.Payment
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "success", resp.Status)

	// Verify database updated
	integration.AssertPaymentStatus(t, ts.DB, paymentID, "success")
	integration.AssertOrderStatus(t, ts.DB, orderID, "paid")
}

// TestWechatCallbackEndpoint tests POST /api/v1/webhooks/wechat
func TestWechatCallbackEndpoint(t *testing.T) {
	t.Parallel()
	ts := integration.SetupPaymentTest(t)
	defer integration.TeardownPaymentTest(t, ts)

	router := SetupPaymentRouter(t, ts)

	userID, orderID, paymentID := integration.CreateTestPaymentFlow(t, ts)
	defer integration.CleanupTestData(t, ts.DB, userID)

	p := integration.GetPaymentFromDB(t, ts.DB, paymentID)

	// Create callback
	callback := payment.WechatCallback{
		OutTradeNo:    strconv.Itoa(paymentID),
		TransactionID: "wechat_test_123456",
		TotalFee:      int(p.Amount * 100),
		ResultCode:    "SUCCESS",
		Sign:          "test_signature",
	}
	body, _ := json.Marshal(callback)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/wechat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp payment.Payment
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "success", resp.Status)

	// Verify database updated
	integration.AssertPaymentStatus(t, ts.DB, paymentID, "success")
	integration.AssertOrderStatus(t, ts.DB, orderID, "paid")
}

// TestPaymentNotFoundError tests 404 error handling
func TestPaymentNotFoundError(t *testing.T) {
	t.Parallel()
	ts := integration.SetupPaymentTest(t)
	defer integration.TeardownPaymentTest(t, ts)

	router := SetupPaymentRouter(t, ts)

	req := httptest.NewRequest("GET", "/api/v1/payments/99999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestOrderAlreadyPaidError tests 409 conflict error
func TestOrderAlreadyPaidError(t *testing.T) {
	t.Parallel()
	ts := integration.SetupPaymentTest(t)
	defer integration.TeardownPaymentTest(t, ts)

	router := SetupPaymentRouter(t, ts)

	userID, orderID, paymentID := integration.CreateTestPaymentFlow(t, ts)
	defer integration.CleanupTestData(t, ts.DB, userID)

	// Complete payment
	integration.SimulateAlipayCallback(t, context.Background(), ts.DB, ts.PaymentService, paymentID)

	// Try to create another payment for same order
	reqBody := payment.InitiatePaymentRequest{
		OrderID:       orderID,
		PaymentMethod: "wechat",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// TestInvalidPaymentMethodError tests 400 validation error
func TestInvalidPaymentMethodError(t *testing.T) {
	t.Parallel()
	ts := integration.SetupPaymentTest(t)
	defer integration.TeardownPaymentTest(t, ts)

	router := SetupPaymentRouter(t, ts)

	userID := integration.SeedTestUser(t, ts.DB, 12)
	productID := integration.SeedTestProduct(t, ts.DB, 12)
	orderID := integration.SeedTestOrder(t, ts.DB, userID, productID)

	defer integration.CleanupTestData(t, ts.DB, userID)

	// Create request with invalid method
	reqBody := payment.InitiatePaymentRequest{
		OrderID:       orderID,
		PaymentMethod: "invalid_method",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
