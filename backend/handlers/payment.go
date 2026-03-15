package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/payment"
)

// Initialize payment service
var paymentService payment.Service

func initPaymentService() {
	if paymentService == nil {
		logger := log.New(os.Stderr, "[PaymentHandler] ", log.LstdFlags)
		orderSvc := order.NewService(config.GetDB(), logger)
		paymentService = payment.NewService(config.GetDB(), orderSvc, logger)
	}
}

// InitiatePayment initiates a payment for an order
// POST /v1/payments
func InitiatePayment(c *gin.Context) {
	initPaymentService()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req payment.InitiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	ctx := context.Background()
	p, err := paymentService.InitiatePayment(ctx, userIDInt, &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusCreated, p)
}

// GetPaymentByID retrieves a payment by ID
// GET /v1/payments/:id
func GetPaymentByID(c *gin.Context) {
	initPaymentService()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	paymentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	ctx := context.Background()
	p, err := paymentService.GetPaymentByID(ctx, userIDInt, paymentID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, p)
}

// ListPayments retrieves user's payments with pagination
// GET /v1/payments
func ListPayments(c *gin.Context) {
	initPaymentService()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 20
	if pp := c.Query("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	status := c.Query("status")
	method := c.Query("method")

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	ctx := context.Background()
	params := &payment.ListPaymentsParams{
		Page:    page,
		PerPage: perPage,
		Status:  status,
		Method:  method,
	}

	result, err := paymentService.ListPayments(ctx, userIDInt, params)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// RefundPayment processes a refund for a payment
// POST /v1/payments/:id/refund
func RefundPayment(c *gin.Context) {
	initPaymentService()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	paymentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	ctx := context.Background()
	p, err := paymentService.RefundPayment(ctx, userIDInt, paymentID, req.Reason)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, p)
}

// HandleAlipayCallback handles Alipay payment callback
// POST /v1/webhooks/alipay
func HandleAlipayCallback(c *gin.Context) {
	initPaymentService()

	var payload payment.AlipayCallback
	if err := c.ShouldBindJSON(&payload); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	ctx := context.Background()
	p, err := paymentService.HandleAlipayCallback(ctx, &payload)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alipay callback processed",
		"payment": p,
	})
}

// HandleWechatCallback handles WeChat payment callback
// POST /v1/webhooks/wechat
func HandleWechatCallback(c *gin.Context) {
	initPaymentService()

	var payload payment.WechatCallback
	if err := c.ShouldBindJSON(&payload); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	ctx := context.Background()
	p, err := paymentService.HandleWechatCallback(ctx, &payload)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "WeChat callback processed",
		"payment": p,
	})
}

// GetMerchantRevenue retrieves merchant revenue information
// GET /v1/merchants/:merchant_id/revenue
func GetMerchantRevenue(c *gin.Context) {
	initPaymentService()

	merchantID, err := strconv.Atoi(c.Param("merchant_id"))
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "2026-03" // Default to current month
	}

	ctx := context.Background()
	revenue, err := paymentService.GetMerchantRevenue(ctx, merchantID, period)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.RespondWithError(c, appErr)
		} else {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		}
		return
	}

	c.JSON(http.StatusOK, revenue)
}
