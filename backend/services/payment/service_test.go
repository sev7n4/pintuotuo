package payment

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/token"
)

var testService Service

func init() {
	// Initialize test database
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to init test DB: %v", err)
	}

	// Initialize cache
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to init cache: %v", err)
	}

	// Clean database and seed for CI environment
	config.TruncateAndSeed()

	logger := log.New(os.Stderr, "[TestPaymentService] ", log.LstdFlags)
	tokenSvc := token.NewService(config.GetDB(), logger)
	orderService := order.NewService(config.GetDB(), logger)
	testService = NewService(config.GetDB(), orderService, logger, tokenSvc)
}

// ============================================================================
// Payment Initiation Tests
// ============================================================================

// TestInitiatePaymentValid tests valid payment initiation
func TestInitiatePaymentValid(t *testing.T) {
	ctx := context.Background()

	// Create an order first
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)
	require.NotNil(t, createdOrder)

	// Now initiate payment
	req := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "alipay",
	}

	payment, err := testService.InitiatePayment(ctx, 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, createdOrder.ID, payment.OrderID)
	assert.Equal(t, "alipay", payment.Method)
	assert.Equal(t, "pending", payment.Status)
	assert.Equal(t, 1, payment.UserID)
	assert.True(t, payment.ID > 0)
	assert.True(t, payment.Amount > 0)
}

// TestInitiatePaymentInvalidMethod tests payment initiation with invalid method
func TestInitiatePaymentInvalidMethod(t *testing.T) {
	ctx := context.Background()

	req := &InitiatePaymentRequest{
		OrderID:       1,
		PaymentMethod: "invalid_method",
	}

	payment, err := testService.InitiatePayment(ctx, 1, req)

	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Equal(t, ErrInvalidPaymentMethod, err)
}

// TestInitiatePaymentOrderNotFound tests payment for non-existent order
func TestInitiatePaymentOrderNotFound(t *testing.T) {
	ctx := context.Background()

	req := &InitiatePaymentRequest{
		OrderID:       99999,
		PaymentMethod: "alipay",
	}

	payment, err := testService.InitiatePayment(ctx, 1, req)

	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Equal(t, ErrOrderNotFound, err)
}

// TestInitiatePaymentOrderAlreadyPaid tests payment for already paid order
func TestInitiatePaymentOrderAlreadyPaid(t *testing.T) {
	ctx := context.Background()

	// Create and pay an order
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	// Update order status to paid
	_, err = order.NewService(config.GetDB(), nil).UpdateOrderStatus(ctx, 1, createdOrder.ID, "paid")
	require.NoError(t, err)

	// Try to initiate payment
	req := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "alipay",
	}

	payment, err := testService.InitiatePayment(ctx, 1, req)

	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Equal(t, ErrOrderAlreadyPaid, err)
}

// TestInitiatePaymentWechat tests payment initiation with WeChat
func TestInitiatePaymentWechat(t *testing.T) {
	ctx := context.Background()

	// Create an order
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  2,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	// Initiate WeChat payment
	req := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "wechat",
	}

	payment, err := testService.InitiatePayment(ctx, 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, "wechat", payment.Method)
	assert.Equal(t, "pending", payment.Status)
}

// ============================================================================
// Payment Retrieval Tests
// ============================================================================

// TestGetPaymentByIDValid tests retrieving a payment
func TestGetPaymentByIDValid(t *testing.T) {
	ctx := context.Background()

	// Create order and payment
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	paymentReq := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "alipay",
	}
	createdPayment, err := testService.InitiatePayment(ctx, 1, paymentReq)
	require.NoError(t, err)

	// Retrieve payment
	payment, err := testService.GetPaymentByID(ctx, 1, createdPayment.ID)

	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, createdPayment.ID, payment.ID)
	assert.Equal(t, 1, payment.UserID)
}

// TestGetPaymentByIDNotFound tests retrieving non-existent payment
func TestGetPaymentByIDNotFound(t *testing.T) {
	ctx := context.Background()

	payment, err := testService.GetPaymentByID(ctx, 1, 99999)

	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Equal(t, ErrPaymentNotFound, err)
}

// TestGetPaymentsByOrderValid tests retrieving all payments for an order
func TestGetPaymentsByOrderValid(t *testing.T) {
	ctx := context.Background()

	// Create order
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	// Create payment
	paymentReq := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "alipay",
	}
	_, err = testService.InitiatePayment(ctx, 1, paymentReq)
	require.NoError(t, err)

	// Retrieve payments
	payments, err := testService.GetPaymentsByOrder(ctx, 1, createdOrder.ID)

	assert.NoError(t, err)
	assert.NotEmpty(t, payments)
	assert.True(t, len(payments) > 0)
}

// ============================================================================
// Alipay Webhook Tests
// ============================================================================

// TestHandleAlipayCallbackValid tests valid Alipay callback
func TestHandleAlipayCallbackValid(t *testing.T) {
	ctx := context.Background()

	// Create order and payment
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	paymentReq := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "alipay",
	}
	createdPayment, err := testService.InitiatePayment(ctx, 1, paymentReq)
	require.NoError(t, err)

	// Handle callback
	callback := &AlipayCallback{
		OutTradeNo:  strconv.Itoa(createdPayment.ID),
		TradeNo:     "alipay_trade_123456",
		TotalAmount: float64(createdPayment.Amount),
		TradeStatus: "TRADE_SUCCESS",
		Timestamp:   "2026-03-15T10:00:00Z",
		Sign:        "test_signature",
	}

	payment, err := testService.HandleAlipayCallback(ctx, callback)

	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, "success", payment.Status)
	assert.NotNil(t, payment.TransactionID)
	assert.Equal(t, "alipay_trade_123456", *payment.TransactionID)
}

// TestHandleAlipayCallbackInvalidSignature tests Alipay callback with invalid signature
func TestHandleAlipayCallbackInvalidSignature(t *testing.T) {
	ctx := context.Background()

	callback := &AlipayCallback{
		OutTradeNo:  "invalid_id",
		TradeNo:     "alipay_trade_123456",
		TotalAmount: 100.00,
		TradeStatus: "TRADE_SUCCESS",
		Timestamp:   "2026-03-15T10:00:00Z",
		Sign:        "invalid_signature",
	}

	payment, err := testService.HandleAlipayCallback(ctx, callback)

	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Equal(t, ErrInvalidSignature, err)
}

// TestHandleAlipayCallbackIdempotency tests Alipay callback idempotency
func TestHandleAlipayCallbackIdempotency(t *testing.T) {
	ctx := context.Background()

	// Create order and payment
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	paymentReq := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "alipay",
	}
	createdPayment, err := testService.InitiatePayment(ctx, 1, paymentReq)
	require.NoError(t, err)

	callback := &AlipayCallback{
		OutTradeNo:  strconv.Itoa(createdPayment.ID),
		TradeNo:     "alipay_trade_789",
		TotalAmount: float64(createdPayment.Amount),
		TradeStatus: "TRADE_SUCCESS",
	}

	// First callback
	payment1, err := testService.HandleAlipayCallback(ctx, callback)
	assert.NoError(t, err)
	assert.Equal(t, "success", payment1.Status)

	// Second identical callback (should be idempotent)
	payment2, err := testService.HandleAlipayCallback(ctx, callback)
	assert.NoError(t, err)
	assert.Equal(t, "success", payment2.Status)
	assert.Equal(t, payment1.ID, payment2.ID)
}

// ============================================================================
// WeChat Webhook Tests
// ============================================================================

// TestHandleWechatCallbackValid tests valid WeChat callback
func TestHandleWechatCallbackValid(t *testing.T) {
	ctx := context.Background()

	// Create order and payment
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	paymentReq := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "wechat",
	}
	createdPayment, err := testService.InitiatePayment(ctx, 1, paymentReq)
	require.NoError(t, err)

	// Handle callback
	callback := &WechatCallback{
		OutTradeNo:    strconv.Itoa(createdPayment.ID),
		TransactionID: "wechat_trans_456789",
		TotalFee:      int(createdPayment.Amount * 100), // in cents
		ResultCode:    "SUCCESS",
		Sign:          "test_signature",
	}

	payment, err := testService.HandleWechatCallback(ctx, callback)

	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, "success", payment.Status)
	assert.NotNil(t, payment.TransactionID)
	assert.Equal(t, "wechat_trans_456789", *payment.TransactionID)
}

// TestHandleWechatCallbackInvalidSignature tests WeChat callback with invalid signature
func TestHandleWechatCallbackInvalidSignature(t *testing.T) {
	ctx := context.Background()

	callback := &WechatCallback{
		OutTradeNo:    "invalid_id",
		TransactionID: "wechat_trans_456789",
		TotalFee:      10000,
		ResultCode:    "SUCCESS",
		Sign:          "invalid_signature",
	}

	payment, err := testService.HandleWechatCallback(ctx, callback)

	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Equal(t, ErrInvalidSignature, err)
}

// TestHandleWechatCallbackIdempotency tests WeChat callback idempotency
func TestHandleWechatCallbackIdempotency(t *testing.T) {
	ctx := context.Background()

	// Create order and payment
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	paymentReq := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "wechat",
	}
	createdPayment, err := testService.InitiatePayment(ctx, 1, paymentReq)
	require.NoError(t, err)

	callback := &WechatCallback{
		OutTradeNo:    strconv.Itoa(createdPayment.ID),
		TransactionID: "wechat_trans_999",
		TotalFee:      int(createdPayment.Amount * 100),
		ResultCode:    "SUCCESS",
	}

	// First callback
	payment1, err := testService.HandleWechatCallback(ctx, callback)
	assert.NoError(t, err)
	assert.Equal(t, "success", payment1.Status)

	// Second identical callback (should be idempotent)
	payment2, err := testService.HandleWechatCallback(ctx, callback)
	assert.NoError(t, err)
	assert.Equal(t, "success", payment2.Status)
	assert.Equal(t, payment1.ID, payment2.ID)
}

// ============================================================================
// Refund Tests
// ============================================================================

// TestRefundPaymentValid tests valid refund
func TestRefundPaymentValid(t *testing.T) {
	ctx := context.Background()

	// Create, initiate, and complete a payment
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	paymentReq := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "alipay",
	}
	createdPayment, err := testService.InitiatePayment(ctx, 1, paymentReq)
	require.NoError(t, err)

	// Simulate successful payment
	callback := &AlipayCallback{
		OutTradeNo:  strconv.Itoa(createdPayment.ID),
		TradeNo:     "alipay_refund_test",
		TotalAmount: float64(createdPayment.Amount),
		TradeStatus: "TRADE_SUCCESS",
	}
	_, err = testService.HandleAlipayCallback(ctx, callback)
	require.NoError(t, err)

	// Process refund
	refund, err := testService.RefundPayment(ctx, 1, createdPayment.ID, "Customer requested")

	assert.NoError(t, err)
	assert.NotNil(t, refund)
	assert.Equal(t, "refunded", refund.Status)
}

// TestRefundPaymentPendingPayment tests refund on pending payment
func TestRefundPaymentPendingPayment(t *testing.T) {
	ctx := context.Background()

	// Create order and payment
	orderReq := &order.CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
	}
	createdOrder, err := order.NewService(config.GetDB(), nil).CreateOrder(ctx, 1, orderReq)
	require.NoError(t, err)

	paymentReq := &InitiatePaymentRequest{
		OrderID:       createdOrder.ID,
		PaymentMethod: "alipay",
	}
	createdPayment, err := testService.InitiatePayment(ctx, 1, paymentReq)
	require.NoError(t, err)

	// Try to refund pending payment
	refund, err := testService.RefundPayment(ctx, 1, createdPayment.ID, "Mistake")

	assert.Error(t, err)
	assert.Nil(t, refund)
	assert.Equal(t, ErrCannotRefundPendingPayment, err)
}

// TestRefundPaymentNotFound tests refund on non-existent payment
func TestRefundPaymentNotFound(t *testing.T) {
	ctx := context.Background()

	refund, err := testService.RefundPayment(ctx, 1, 99999, "Test")

	assert.Error(t, err)
	assert.Nil(t, refund)
	assert.Equal(t, ErrPaymentNotFound, err)
}

// ============================================================================
// Revenue Tests
// ============================================================================

// TestCalculateCommission tests commission calculation
func TestCalculateCommission(t *testing.T) {
	commission := testService.CalculateCommission(100.0, 0.30)
	assert.Equal(t, 30.0, commission)

	commission = testService.CalculateCommission(500.0, 0.30)
	assert.Equal(t, 150.0, commission)

	commission = testService.CalculateCommission(1000.0, 0.25)
	assert.Equal(t, 250.0, commission)
}

// TestGetMerchantRevenueValid tests merchant revenue calculation
func TestGetMerchantRevenueValid(t *testing.T) {
	ctx := context.Background()

	// Get merchant revenue
	revenue, err := testService.GetMerchantRevenue(ctx, 1, "2026-03")

	assert.NoError(t, err)
	assert.NotNil(t, revenue)
	assert.Equal(t, 1, revenue.MerchantID)
	assert.Equal(t, 0.30, revenue.CommissionRate)
	assert.True(t, revenue.TotalSales >= 0)
}

// ============================================================================
// Listing Tests
// ============================================================================

// TestListPaymentsValid tests payment listing
func TestListPaymentsValid(t *testing.T) {
	ctx := context.Background()

	params := &ListPaymentsParams{
		Page:    1,
		PerPage: 20,
	}

	result, err := testService.ListPayments(ctx, 1, params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Total >= 0)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PerPage)
}

// TestListPaymentsWithStatus tests payment listing with status filter
func TestListPaymentsWithStatus(t *testing.T) {
	ctx := context.Background()

	params := &ListPaymentsParams{
		Page:    1,
		PerPage: 20,
		Status:  "pending",
	}

	result, err := testService.ListPayments(ctx, 1, params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify all results are pending
	for _, payment := range result.Data {
		assert.Equal(t, "pending", payment.Status)
	}
}

// TestListPaymentsWithMethod tests payment listing with method filter
func TestListPaymentsWithMethod(t *testing.T) {
	ctx := context.Background()

	params := &ListPaymentsParams{
		Page:    1,
		PerPage: 20,
		Method:  "alipay",
	}

	result, err := testService.ListPayments(ctx, 1, params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify all results are alipay
	for _, payment := range result.Data {
		assert.Equal(t, "alipay", payment.Method)
	}
}

// TestListPaymentsPagination tests payment listing pagination
func TestListPaymentsPagination(t *testing.T) {
	ctx := context.Background()

	// Get first page
	params := &ListPaymentsParams{
		Page:    1,
		PerPage: 5,
	}

	result, err := testService.ListPayments(ctx, 1, params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 5, result.PerPage)
}
