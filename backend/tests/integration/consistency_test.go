package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/payment"
)

// TestPaymentDatabaseConsistency verifies database consistency
func TestPaymentDatabaseConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	userID, _, paymentID := CreateTestPaymentFlow(t, ts)
	defer CleanupTestData(t, ts.DB, userID)

	// 1. Query via service layer
	servicePayment, err := ts.PaymentService.GetPaymentByID(ctx, userID, paymentID)
	require.NoError(t, err)

	// 2. Query database directly via SQL
	dbPayment := GetPaymentFromDB(t, ts.DB, paymentID)

	// 3. Verify both return identical data
	assert.Equal(t, servicePayment.ID, dbPayment.ID)
	assert.Equal(t, servicePayment.UserID, dbPayment.UserID)
	assert.Equal(t, servicePayment.OrderID, dbPayment.OrderID)
	assert.Equal(t, servicePayment.Amount, dbPayment.Amount)
	assert.Equal(t, servicePayment.Method, dbPayment.Method)
	assert.Equal(t, servicePayment.Status, dbPayment.Status)

	// 4. Verify all required fields present
	assert.NotNil(t, dbPayment.CreatedAt)
	assert.NotNil(t, dbPayment.UpdatedAt)
	assert.Greater(t, dbPayment.ID, 0)
	assert.Greater(t, dbPayment.UserID, 0)
}

// TestPaymentOrderSyncConsistency verifies payment and order status sync
func TestPaymentOrderSyncConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	userID, orderID, paymentID := CreateTestPaymentFlow(t, ts)
	defer CleanupTestData(t, ts.DB, userID)

	// 1. Initial state - both pending
	AssertPaymentStatus(t, ts.DB, paymentID, "pending")
	AssertOrderStatus(t, ts.DB, orderID, "pending")

	// 2. Complete payment
	SimulateAlipayCallback(t, ctx, ts.DB, ts.PaymentService, paymentID)

	// 3. Query both via service layer
	payment, err := ts.PaymentService.GetPaymentByID(ctx, userID, paymentID)
	require.NoError(t, err)

	order, err := ts.OrderService.GetOrderByID(ctx, userID, orderID)
	require.NoError(t, err)

	// 4. Verify statuses match
	assert.Equal(t, "success", payment.Status)
	assert.Equal(t, "paid", order.Status)

	// 5. Verify timestamps are approximately synchronized (within 1 second)
	timeDiff := payment.UpdatedAt.Sub(order.UpdatedAt)
	assert.True(t, timeDiff.Abs().Seconds() < 1.0,
		"Payment and order timestamps should be within 1 second of each other")
}

// TestRevenueCalculationConsistency verifies commission calculations
func TestRevenueCalculationConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	// Setup merchant with unique IDs for parallel test isolation
	merchantID := TestMerchantID
	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Create and pay for 3 orders
	totalAmount := 0.0
	for i := 0; i < 3; i++ {
		req := &order.CreateOrderRequest{
			ProductID: productID,
			Quantity:  1,
		}
		o, err := ts.OrderService.CreateOrder(ctx, userID, req)
		require.NoError(t, err)

		payReq := &payment.InitiatePaymentRequest{
			OrderID:       o.ID,
			PaymentMethod: "alipay",
		}
		p, err := ts.PaymentService.InitiatePayment(ctx, userID, payReq)
		require.NoError(t, err)

		// Complete payment
		SimulateAlipayCallback(t, ctx, ts.DB, ts.PaymentService, p.ID)
		totalAmount += p.Amount
	}

	// 1. Calculate commission via service
	commissionRate := 0.30 // 30%
	expectedCommission := ts.PaymentService.CalculateCommission(totalAmount, commissionRate)

	// 2. Manually calculate expected values
	expectedMerchantEarnings := totalAmount - expectedCommission

	// 3. Verify calculations with floating point tolerance (0.01 = 1 cent)
	const tolerance = 0.01
	assert.InDelta(t, totalAmount*commissionRate, expectedCommission, tolerance)
	assert.InDelta(t, totalAmount*(1.0-commissionRate), expectedMerchantEarnings, tolerance)

	// 4. Query merchant revenue
	period := "2026-03"
	revenue, err := ts.PaymentService.GetMerchantRevenue(ctx, merchantID, period)
	require.NoError(t, err)

	// 5. Verify revenue calculations match with tolerance
	assert.GreaterOrEqual(t, revenue.TotalSales, totalAmount*0.99) // Allow small variance
	assert.Greater(t, revenue.PlatformCommission, 0.0)
	assert.Greater(t, revenue.MerchantEarnings, 0.0)
}

// TestCacheConsistencyAfterFailure verifies cache handles failures gracefully
func TestCacheConsistencyAfterFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	userID, _, paymentID := CreateTestPaymentFlow(t, ts)
	defer CleanupTestData(t, ts.DB, userID)

	// 1. Get payment (populates cache)
	payment1, err := ts.PaymentService.GetPaymentByID(ctx, userID, paymentID)
	require.NoError(t, err)

	cacheKey := cache.PaymentKey(paymentID)

	// 2. Verify cache entry exists
	exists, _ := cache.Exists(ctx, cacheKey)
	assert.True(t, exists, "Payment should be cached")

	// 3. Simulate cache failure by invalidating manually
	_ = cache.Delete(ctx, cacheKey)

	// 4. Verify cache key no longer exists
	exists, _ = cache.Exists(ctx, cacheKey)
	assert.False(t, exists, "Cache should be cleared")

	// 5. Request payment again (should fetch from DB and repopulate cache)
	payment2, err := ts.PaymentService.GetPaymentByID(ctx, userID, paymentID)
	require.NoError(t, err)

	// 6. Verify data is identical
	assert.Equal(t, payment1.ID, payment2.ID)
	assert.Equal(t, payment1.Amount, payment2.Amount)
	assert.Equal(t, payment1.Status, payment2.Status)

	// 7. Verify cache is repopulated
	exists, _ = cache.Exists(ctx, cacheKey)
	assert.True(t, exists, "Cache should be repopulated after fetch")
}

// TestPaymentAmountConsistency verifies amounts are correct
func TestPaymentAmountConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Create order and payment
	req := &order.CreateOrderRequest{
		ProductID: productID,
		Quantity:  1,
	}
	o, err := ts.OrderService.CreateOrder(ctx, userID, req)
	require.NoError(t, err)

	payReq := &payment.InitiatePaymentRequest{
		OrderID:       o.ID,
		PaymentMethod: "alipay",
	}
	p, err := ts.PaymentService.InitiatePayment(ctx, userID, payReq)
	require.NoError(t, err)

	// 1. Verify payment amount matches order amount
	order, err := ts.OrderService.GetOrderByID(ctx, userID, o.ID)
	require.NoError(t, err)

	assert.Equal(t, order.TotalPrice, p.Amount, "Payment amount should match order total")

	// 2. Verify all payment fields are consistent
	payment := GetPaymentFromDB(t, ts.DB, p.ID)
	assert.Equal(t, p.Amount, payment.Amount)
	assert.Equal(t, p.Method, payment.Method)
	assert.Equal(t, p.Status, payment.Status)
}

// TestOrderStatusTransitionConsistency verifies status transitions
func TestOrderStatusTransitionConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	userID, orderID, paymentID := CreateTestPaymentFlow(t, ts)
	defer CleanupTestData(t, ts.DB, userID)

	// Track status transitions
	statuses := []string{}

	// 1. Initial state
	order, _ := ts.OrderService.GetOrderByID(ctx, userID, orderID)
	statuses = append(statuses, order.Status)
	assert.Equal(t, "pending", statuses[0])

	// 2. After payment success
	SimulateAlipayCallback(t, ctx, ts.DB, ts.PaymentService, paymentID)

	order, _ = ts.OrderService.GetOrderByID(ctx, userID, orderID)
	statuses = append(statuses, order.Status)
	assert.Equal(t, "paid", statuses[1])

	// 3. Verify valid state transition (pending → paid)
	assert.Equal(t, "pending", statuses[0])
	assert.Equal(t, "paid", statuses[1])

	// 4. Verify timestamps are monotonic increasing
	order1 := GetOrderFromDB(t, ts.DB, orderID)
	assert.True(t, order1.UpdatedAt.After(order1.CreatedAt) || order1.UpdatedAt.Equal(order1.CreatedAt))
}

// TestPaymentListConsistency verifies payment list returns correct data
func TestPaymentListConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Create 5 payments
	createdPaymentIDs := make([]int, 5)
	for i := 0; i < 5; i++ {
		req := &order.CreateOrderRequest{
			ProductID: productID,
			Quantity:  1,
		}
		o, err := ts.OrderService.CreateOrder(ctx, userID, req)
		require.NoError(t, err)

		payReq := &payment.InitiatePaymentRequest{
			OrderID:       o.ID,
			PaymentMethod: "alipay",
		}
		p, err := ts.PaymentService.InitiatePayment(ctx, userID, payReq)
		require.NoError(t, err)
		createdPaymentIDs[i] = p.ID
	}

	// 1. List payments via service
	params := &payment.ListPaymentsParams{
		Page:    1,
		PerPage: 10,
	}
	result, err := ts.PaymentService.ListPayments(ctx, userID, params)
	require.NoError(t, err)

	// 2. Verify all created payments are in list
	listPaymentIDs := make(map[int]bool)
	for _, p := range result.Data {
		listPaymentIDs[p.ID] = true
	}

	// Debug: Log what we got
	if len(result.Data) == 0 {
		// Check if payments actually exist in database
		var dbCount int
		err := ts.DB.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM payments WHERE user_id = $1",
			userID).Scan(&dbCount)
		if err == nil {
			t.Logf("ERROR: ListPayments returned 0 results but database has %d payments for user %d. Created: %v",
				dbCount, userID, createdPaymentIDs)
		}
	} else {
		t.Logf("ListPayments returned %d results (total=%d) for user %d",
			len(result.Data), result.Total, userID)
	}

	for _, id := range createdPaymentIDs {
		assert.True(t, listPaymentIDs[id], "Payment %d should be in list (got %d items: %v)",
			id, len(result.Data), result.Data)
	}

	// 3. Verify list count is consistent
	assert.GreaterOrEqual(t, result.Total, 5, "Total should include all created payments")

	// 4. Verify list is sorted by created_at descending (newest first)
	if len(result.Data) > 1 {
		for i := 0; i < len(result.Data)-1; i++ {
			assert.True(t, result.Data[i].CreatedAt.After(result.Data[i+1].CreatedAt) || result.Data[i].CreatedAt.Equal(result.Data[i+1].CreatedAt),
				"List should be sorted by created_at descending")
		}
	}
}

// TestMultiplePaymentsForSameOrderConsistency verifies only one payment succeeds per order
func TestMultiplePaymentsForSameOrderConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	userID, orderID, paymentID1 := CreateTestPaymentFlow(t, ts)
	defer CleanupTestData(t, ts.DB, userID)

	// Complete first payment
	SimulateAlipayCallback(t, ctx, ts.DB, ts.PaymentService, paymentID1)

	// Verify order is paid
	AssertOrderStatus(t, ts.DB, orderID, "paid")

	// Try to create second payment for same order
	payReq := &payment.InitiatePaymentRequest{
		OrderID:       orderID,
		PaymentMethod: "wechat",
	}
	_, err := ts.PaymentService.InitiatePayment(ctx, userID, payReq)
	require.Error(t, err, "Should not allow second payment for already-paid order")

	// Verify only one payment exists for order
	payments, err := ts.PaymentService.GetPaymentsByOrder(ctx, userID, orderID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(payments), "Should only have one payment for order")
	assert.Equal(t, "success", payments[0].Status)
}

// TestPaymentFilterConsistency verifies payment list filtering works correctly
func TestPaymentFilterConsistency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Create 3 alipay payments and 2 wechat payments
	for i := 0; i < 3; i++ {
		req := &order.CreateOrderRequest{
			ProductID: productID,
			Quantity:  1,
		}
		o, err := ts.OrderService.CreateOrder(ctx, userID, req)
		require.NoError(t, err)

		payReq := &payment.InitiatePaymentRequest{
			OrderID:       o.ID,
			PaymentMethod: "alipay",
		}
		_, err = ts.PaymentService.InitiatePayment(ctx, userID, payReq)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		req := &order.CreateOrderRequest{
			ProductID: productID,
			Quantity:  1,
		}
		o, err := ts.OrderService.CreateOrder(ctx, userID, req)
		require.NoError(t, err)

		payReq := &payment.InitiatePaymentRequest{
			OrderID:       o.ID,
			PaymentMethod: "wechat",
		}
		_, err = ts.PaymentService.InitiatePayment(ctx, userID, payReq)
		require.NoError(t, err)
	}

	// 1. Filter by alipay method
	alipayParams := &payment.ListPaymentsParams{
		Page:    1,
		PerPage: 10,
		Method:  "alipay",
	}
	alipayResult, err := ts.PaymentService.ListPayments(ctx, userID, alipayParams)
	require.NoError(t, err)

	// Count alipay payments
	alipayCount := 0
	for _, p := range alipayResult.Data {
		if p.Method == "alipay" {
			alipayCount++
		}
	}
	assert.GreaterOrEqual(t, alipayCount, 3, "Should have at least 3 alipay payments")

	// 2. Filter by wechat method
	wechatParams := &payment.ListPaymentsParams{
		Page:    1,
		PerPage: 10,
		Method:  "wechat",
	}
	wechatResult, err := ts.PaymentService.ListPayments(ctx, userID, wechatParams)
	require.NoError(t, err)

	// Count wechat payments
	wechatCount := 0
	for _, p := range wechatResult.Data {
		if p.Method == "wechat" {
			wechatCount++
		}
	}
	assert.GreaterOrEqual(t, wechatCount, 2, "Should have at least 2 wechat payments")
}
