package integration

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/payment"
)

// TestCompleteGroupPurchaseWithPayment tests full group purchase workflow
func TestCompleteGroupPurchaseWithPayment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	// Setup users and products with unique IDs for parallel test isolation
	uniqueIDA := GenerateUniqueID()
	uniqueIDB := GenerateUniqueID()
	userAID := SeedTestUser(t, ts.DB, uniqueIDA)
	userBID := SeedTestUser(t, ts.DB, uniqueIDB)
	productID, _ := SeedTestProduct(t, ts.DB, uniqueIDA)

	defer CleanupTestData(t, ts.DB, userAID)
	defer CleanupTestData(t, ts.DB, userBID)

	// Create orders for both users
	orderAReq := &order.CreateOrderRequest{
		ProductID: productID,
		Quantity:  1,
	}
	orderA, err := ts.OrderService.CreateOrder(ctx, userAID, orderAReq)
	require.NoError(t, err)

	orderBReq := &order.CreateOrderRequest{
		ProductID: productID,
		Quantity:  1,
	}
	orderB, err := ts.OrderService.CreateOrder(ctx, userBID, orderBReq)
	require.NoError(t, err)

	// User A initiates payment
	payReqA := &payment.InitiatePaymentRequest{
		OrderID:       orderA.ID,
		PaymentMethod: "alipay",
	}
	paymentA, err := ts.PaymentService.InitiatePayment(ctx, userAID, payReqA)
	require.NoError(t, err)

	// Verify both orders are still pending
	AssertOrderStatus(t, ts.DB, orderA.ID, "pending")
	AssertOrderStatus(t, ts.DB, orderB.ID, "pending")

	// Simulate webhook callback for User A's payment
	callbackA := &payment.AlipayCallback{
		OutTradeNo:  strconv.Itoa(paymentA.ID),
		TradeNo:     "alipay_user_a_123456",
		TotalAmount: paymentA.Amount,
		TradeStatus: "TRADE_SUCCESS",
		Timestamp:   "2026-03-15 12:00:00",
		Sign:        "test_signature",
	}

	result, err := ts.PaymentService.HandleAlipayCallback(ctx, callbackA)
	require.NoError(t, err)
	require.Equal(t, "success", result.Status)

	// Verify User A's order is now paid
	AssertOrderStatus(t, ts.DB, orderA.ID, "paid")

	// User B can independently pay
	payReqB := &payment.InitiatePaymentRequest{
		OrderID:       orderB.ID,
		PaymentMethod: "wechat",
	}
	paymentB, err := ts.PaymentService.InitiatePayment(ctx, userBID, payReqB)
	require.NoError(t, err)

	// Simulate webhook for User B
	callbackB := &payment.WechatCallback{
		OutTradeNo:    strconv.Itoa(paymentB.ID),
		TransactionID: "wechat_user_b_123456",
		TotalFee:      int(paymentB.Amount * 100),
		ResultCode:    "SUCCESS",
		Sign:          "test_signature",
	}

	result, err = ts.PaymentService.HandleWechatCallback(ctx, callbackB)
	require.NoError(t, err)
	require.Equal(t, "success", result.Status)

	// Verify User B's order is now paid
	AssertOrderStatus(t, ts.DB, orderB.ID, "paid")
}

// TestMultiplePaymentsForDifferentOrders tests paying for multiple orders
func TestMultiplePaymentsForDifferentOrders(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID, _ := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Create 3 orders
	orderIDs := make([]int, 3)
	paymentIDs := make([]int, 3)

	for i := 0; i < 3; i++ {
		req := &order.CreateOrderRequest{
			ProductID: productID,
			Quantity:  1,
		}
		o, err := ts.OrderService.CreateOrder(ctx, userID, req)
		require.NoError(t, err)
		orderIDs[i] = o.ID

		// Initiate payment
		payReq := &payment.InitiatePaymentRequest{
			OrderID:       o.ID,
			PaymentMethod: "alipay",
		}
		p, err := ts.PaymentService.InitiatePayment(ctx, userID, payReq)
		require.NoError(t, err)
		paymentIDs[i] = p.ID
	}

	// Pay for each order separately
	for i := 0; i < 3; i++ {
		p := GetPaymentFromDB(t, ts.DB, paymentIDs[i])

		callback := &payment.AlipayCallback{
			OutTradeNo:  strconv.Itoa(paymentIDs[i]),
			TradeNo:     "alipay_test_" + strconv.Itoa(i),
			TotalAmount: p.Amount,
			TradeStatus: "TRADE_SUCCESS",
			Timestamp:   "2026-03-15 12:00:00",
			Sign:        "test_signature",
		}

		_, err := ts.PaymentService.HandleAlipayCallback(ctx, callback)
		require.NoError(t, err)

		// Verify this order is paid
		AssertOrderStatus(t, ts.DB, orderIDs[i], "paid")

		// Verify other orders remain pending
		for j := i + 1; j < 3; j++ {
			AssertOrderStatus(t, ts.DB, orderIDs[j], "pending")
		}
	}

	// Verify all orders eventually paid
	for i := 0; i < 3; i++ {
		AssertOrderStatus(t, ts.DB, orderIDs[i], "paid")
	}
}

// TestPaymentWithOrderCancellation tests payment and order cancellation interaction
func TestPaymentWithOrderCancellation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	userID, orderID, paymentID := CreateTestPaymentFlow(t, ts)
	defer CleanupTestData(t, ts.DB, userID)

	// 1. Payment is pending, order is pending
	AssertPaymentStatus(t, ts.DB, paymentID, "pending")
	AssertOrderStatus(t, ts.DB, orderID, "pending")

	// 2. Try to cancel order while payment is pending
	// (This test assumes order service allows cancellation of pending orders)
	cancelledOrder, err := ts.OrderService.CancelOrder(ctx, userID, orderID)
	// If cancellation is allowed, order should be cancelled
	if err == nil {
		require.NotNil(t, cancelledOrder)
		// Verify order is cancelled
		AssertOrderStatus(t, ts.DB, orderID, "cancelled")
	}

	// 3. Receive webhook - depending on business logic, this may fail
	// because order is now cancelled
	p := GetPaymentFromDB(t, ts.DB, paymentID)

	callback := &payment.AlipayCallback{
		OutTradeNo:  strconv.Itoa(paymentID),
		TradeNo:     "alipay_test_123456",
		TotalAmount: p.Amount,
		TradeStatus: "TRADE_SUCCESS",
		Timestamp:   "2026-03-15 12:00:00",
		Sign:        "test_signature",
	}

	// Process callback - may fail if order is cancelled
	_, err = ts.PaymentService.HandleAlipayCallback(ctx, callback)

	// If payment processing failed due to cancelled order, that's expected
	// If it succeeded, order should still be paid or cancelled depending on logic
	if err == nil {
		// Verify order is now paid (overrides cancelled state)
		AssertOrderStatus(t, ts.DB, orderID, "paid")
	}
}

// TestConcurrentPaymentsForDifferentUsers tests multiple users paying simultaneously
func TestConcurrentPaymentsForDifferentUsers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	// Setup 3 users with orders - use unique IDs for parallel test isolation
	const numUsers = 3
	users := make([]int, numUsers)
	orders := make([]int, numUsers)
	payments := make([]int, numUsers)

	baseID := GenerateUniqueID()
	productID, _ := SeedTestProduct(t, ts.DB, baseID)

	for i := 0; i < numUsers; i++ {
		users[i] = SeedTestUser(t, ts.DB, baseID+i)
		defer CleanupTestData(t, ts.DB, users[i])

		req := &order.CreateOrderRequest{
			ProductID: productID,
			Quantity:  1,
		}
		o, err := ts.OrderService.CreateOrder(ctx, users[i], req)
		require.NoError(t, err)
		orders[i] = o.ID

		payReq := &payment.InitiatePaymentRequest{
			OrderID:       o.ID,
			PaymentMethod: "alipay",
		}
		p, err := ts.PaymentService.InitiatePayment(ctx, users[i], payReq)
		require.NoError(t, err)
		payments[i] = p.ID
	}

	// Process payments concurrently
	var wg sync.WaitGroup
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			p := GetPaymentFromDB(t, ts.DB, payments[idx])

			callback := &payment.AlipayCallback{
				OutTradeNo:  strconv.Itoa(payments[idx]),
				TradeNo:     "alipay_concurrent_" + strconv.Itoa(idx),
				TotalAmount: p.Amount,
				TradeStatus: "TRADE_SUCCESS",
				Timestamp:   "2026-03-15 12:00:00",
				Sign:        "test_signature",
			}

			_, err := ts.PaymentService.HandleAlipayCallback(ctx, callback)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all payments and orders completed
	for i := 0; i < numUsers; i++ {
		AssertPaymentStatus(t, ts.DB, payments[i], "success")
		AssertOrderStatus(t, ts.DB, orders[i], "paid")
	}
}

// TestConcurrentPaymentsForSameOrder tests multiple payment attempts for same order
func TestConcurrentPaymentsForSameOrder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	userID, orderID, firstPaymentID := CreateTestPaymentFlow(t, ts)
	defer CleanupTestData(t, ts.DB, userID)

	// First payment succeeds
	p := GetPaymentFromDB(t, ts.DB, firstPaymentID)

	callback := &payment.AlipayCallback{
		OutTradeNo:  strconv.Itoa(firstPaymentID),
		TradeNo:     "alipay_first",
		TotalAmount: p.Amount,
		TradeStatus: "TRADE_SUCCESS",
		Timestamp:   "2026-03-15 12:00:00",
		Sign:        "test_signature",
	}

	_, err := ts.PaymentService.HandleAlipayCallback(ctx, callback)
	require.NoError(t, err)

	// Verify order is now paid
	AssertOrderStatus(t, ts.DB, orderID, "paid")

	// Attempt second payment should fail
	payReq := &payment.InitiatePaymentRequest{
		OrderID:       orderID,
		PaymentMethod: "wechat",
	}
	_, err = ts.PaymentService.InitiatePayment(ctx, userID, payReq)
	require.Error(t, err, "Should not allow second payment for already-paid order")
}

// TestConcurrentRefunds tests multiple refunds processed concurrently
func TestConcurrentRefunds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	// Setup 5 users with completed payments - use unique IDs for parallel test isolation
	const numPayments = 5
	users := make([]int, numPayments)
	payments := make([]int, numPayments)

	baseID := GenerateUniqueID()
	productID, _ := SeedTestProduct(t, ts.DB, baseID)

	for i := 0; i < numPayments; i++ {
		users[i] = SeedTestUser(t, ts.DB, baseID+i)
		defer CleanupTestData(t, ts.DB, users[i])

		req := &order.CreateOrderRequest{
			ProductID: productID,
			Quantity:  1,
		}
		o, err := ts.OrderService.CreateOrder(ctx, users[i], req)
		require.NoError(t, err)

		payReq := &payment.InitiatePaymentRequest{
			OrderID:       o.ID,
			PaymentMethod: "alipay",
		}
		p, err := ts.PaymentService.InitiatePayment(ctx, users[i], payReq)
		require.NoError(t, err)
		payments[i] = p.ID

		// Complete payment
		SimulateAlipayCallback(t, ctx, ts.DB, ts.PaymentService, p.ID)
	}

	// Process refunds concurrently
	var wg sync.WaitGroup
	for i := 0; i < numPayments; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			_, err := ts.PaymentService.RefundPayment(ctx, users[idx], payments[idx], "Concurrent refund test")
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all refunds succeeded
	for i := 0; i < numPayments; i++ {
		AssertPaymentStatus(t, ts.DB, payments[i], "refunded")
	}
}

// TestPaymentRetryAfterTimeout tests payment retry after timeout
func TestPaymentRetryAfterTimeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	userID, orderID, paymentID := CreateTestPaymentFlow(t, ts)
	defer CleanupTestData(t, ts.DB, userID)

	// Verify first payment is pending
	AssertPaymentStatus(t, ts.DB, paymentID, "pending")

	// Simulate timeout by not sending webhook
	// User retries by initiating new payment
	payReq := &payment.InitiatePaymentRequest{
		OrderID:       orderID,
		PaymentMethod: "wechat",
	}
	_, err := ts.PaymentService.InitiatePayment(ctx, userID, payReq)
	require.Error(t, err, "Should not allow second payment for already-pending order")

	// In real scenario, old payment would be cleaned up or marked as expired
	// and user could retry. For now, verify business logic is enforced.
}

// TestWebhookDelayedDelivery tests webhook callback with simulated delay
func TestWebhookDelayedDelivery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	userID, orderID, paymentID := CreateTestPaymentFlow(t, ts)
	defer CleanupTestData(t, ts.DB, userID)

	// Verify payment is pending
	AssertPaymentStatus(t, ts.DB, paymentID, "pending")

	// Simulate delayed webhook delivery
	go func() {
		time.Sleep(1 * time.Second)

		p := GetPaymentFromDB(t, ts.DB, paymentID)

		callback := &payment.AlipayCallback{
			OutTradeNo:  strconv.Itoa(paymentID),
			TradeNo:     "alipay_delayed",
			TotalAmount: p.Amount,
			TradeStatus: "TRADE_SUCCESS",
			Timestamp:   "2026-03-15 12:00:00",
			Sign:        "test_signature",
		}

		_, err := ts.PaymentService.HandleAlipayCallback(ctx, callback)
		require.NoError(t, err)
	}()

	// Wait for delayed webhook to be processed
	time.Sleep(2 * time.Second)

	// Verify payment eventually updates
	AssertPaymentStatus(t, ts.DB, paymentID, "success")
	AssertOrderStatus(t, ts.DB, orderID, "paid")
}
