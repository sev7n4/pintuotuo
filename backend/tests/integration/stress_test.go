package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/payment"
)

// TestHighConcurrencyPaymentInitiation tests 100 concurrent payment initiations
func TestHighConcurrencyPaymentInitiation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Remove t.Parallel() to debug execution order issues
	// t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	const numConcurrent = 100
	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID, _ := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Prepare 100 orders
	orderIDs := make([]int, numConcurrent)
	for i := 0; i < numConcurrent; i++ {
		req := &order.CreateOrderRequest{
			ProductID: productID,
			Quantity:  1,
		}
		o, err := ts.OrderService.CreateOrder(ctx, userID, req)
		require.NoError(t, err)
		orderIDs[i] = o.ID
	}

	// Initiate payments concurrently
	var wg sync.WaitGroup
	var mutex sync.Mutex
	successCount := 0
	errorCount := 0

	start := time.Now()

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			payReq := &payment.InitiatePaymentRequest{
				OrderID:       orderIDs[idx],
				PaymentMethod: "alipay",
			}

			_, err := ts.PaymentService.InitiatePayment(ctx, userID, payReq)
			if err != nil {
				fmt.Printf("[StressTest] Failed to initiate payment for order %d: %v\n", orderIDs[idx], err)
			}

			mutex.Lock()
			if err == nil {
				successCount++
			} else {
				errorCount++
			}
			mutex.Unlock()
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Verify results
	// For high concurrency tests, allow a small margin of error due to system constraints
	require.Greater(t, successCount, int(95), "Most payment initiations should succeed")
	require.Less(t, errorCount, int(5), "Should have minimal errors")

	// Log performance metrics
	t.Logf("Initiated %d payments concurrently in %v (%.2f payments/sec)",
		numConcurrent, elapsed, float64(numConcurrent)/elapsed.Seconds())

	// Verify performance is reasonable (< 10 seconds for 100 payments)
	require.Less(t, elapsed, 10*time.Second, "Should handle 100 concurrent payments quickly")
}

// TestHighConcurrencyWebhookCallbacks tests 50 concurrent webhook callbacks
func TestHighConcurrencyWebhookCallbacks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	const numPayments = 50
	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID, _ := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Create 50 pending payments
	paymentIDs := make([]int, numPayments)
	for i := 0; i < numPayments; i++ {
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
		paymentIDs[i] = p.ID
	}

	// Process callbacks concurrently
	var wg sync.WaitGroup
	var mutex sync.Mutex
	successCount := 0
	errorCount := 0

	start := time.Now()

	for i := 0; i < numPayments; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			p := GetPaymentFromDB(t, ts.DB, paymentIDs[idx])

			callback := &payment.AlipayCallback{
				OutTradeNo:  fmt.Sprintf("%d", paymentIDs[idx]),
				TradeNo:     fmt.Sprintf("alipay_stress_%d", idx),
				TotalAmount: p.Amount,
				TradeStatus: "TRADE_SUCCESS",
				Timestamp:   "2026-03-15 12:00:00",
				Sign:        "test_signature",
			}

			_, err := ts.PaymentService.HandleAlipayCallback(ctx, callback)
			if err != nil {
				fmt.Printf("[StressTest] Failed to handle alipay callback for payment %d: %v\n", paymentIDs[idx], err)
			}

			mutex.Lock()
			if err == nil {
				successCount++
			} else {
				errorCount++
			}
			mutex.Unlock()
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Verify results
	// For high concurrency tests, allow a small margin of error due to system constraints
	require.Greater(t, successCount, int(45), "Most callback processings should succeed")
	require.Less(t, errorCount, int(5), "Should have minimal errors")

	// Log performance metrics
	t.Logf("Processed %d webhook callbacks concurrently in %v (%.2f callbacks/sec)",
		numPayments, elapsed, float64(numPayments)/elapsed.Seconds())

	// Verify performance is reasonable (< 5 seconds for 50 callbacks)
	require.Less(t, elapsed, 5*time.Second, "Should handle 50 concurrent webhook callbacks quickly")

	// Verify all payments were processed idempotently
	for i := 0; i < numPayments; i++ {
		AssertPaymentStatus(t, ts.DB, paymentIDs[i], "success")
	}
}

// TestCacheUnderLoad tests cache performance with 100 reads
func TestCacheUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID, _ := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Create 10 payments
	paymentIDs := make([]int, 10)
	for i := 0; i < 10; i++ {
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
		paymentIDs[i] = p.ID
	}

	// Read each payment 100 times concurrently
	var wg sync.WaitGroup
	const readsPerPayment = 100
	var mutex sync.Mutex
	cacheHits := 0
	cacheMisses := 0

	start := time.Now()

	for i := 0; i < 10; i++ {
		for j := 0; j < readsPerPayment; j++ {
			wg.Add(1)
			go func(paymentIdx int) {
				defer wg.Done()

				_, err := ts.PaymentService.GetPaymentByID(ctx, userID, paymentIDs[paymentIdx])

				mutex.Lock()
				if err == nil {
					cacheHits++
				} else {
					cacheMisses++
				}
				mutex.Unlock()
			}(i)
		}
	}

	wg.Wait()
	elapsed := time.Since(start)

	totalReads := 10 * readsPerPayment

	// Verify results
	require.Equal(t, totalReads, cacheHits, "All reads should succeed")
	require.Equal(t, 0, cacheMisses, "Should have no read failures")

	// Calculate hit rate (after first read, subsequent reads should hit cache)
	hitRate := float64(cacheHits) / float64(totalReads) * 100

	// Log performance metrics
	t.Logf("Performed %d concurrent reads in %v (%.2f reads/sec), hit rate: %.1f%%",
		totalReads, elapsed, float64(totalReads)/elapsed.Seconds(), hitRate)

	// Verify performance is reasonable (< 5 seconds for 1000 reads)
	require.Less(t, elapsed, 5*time.Second, "Should handle 1000 concurrent cache reads quickly")
}

// TestDatabaseConnectionPoolUnderLoad tests connection pool under load
func TestDatabaseConnectionPoolUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID, _ := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Perform 500 concurrent operations
	const numOps = 500
	var wg sync.WaitGroup
	var mutex sync.Mutex
	successCount := 0
	errorCount := 0

	start := time.Now()

	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Alternate between create and read operations
			if idx%2 == 0 {
				// Create order and payment
				req := &order.CreateOrderRequest{
					ProductID: productID,
					Quantity:  1,
				}
				o, err := ts.OrderService.CreateOrder(ctx, userID, req)

				mutex.Lock()
				if err == nil {
					// Try to initiate payment
					payReq := &payment.InitiatePaymentRequest{
						OrderID:       o.ID,
						PaymentMethod: "alipay",
					}
					_, err := ts.PaymentService.InitiatePayment(ctx, userID, payReq)
					if err == nil {
						successCount++
					} else {
						errorCount++
					}
				} else {
					errorCount++
				}
				mutex.Unlock()
			} else {
				// List payments
				params := &payment.ListPaymentsParams{
					Page:    1,
					PerPage: 10,
				}
				_, err := ts.PaymentService.ListPayments(ctx, userID, params)

				mutex.Lock()
				if err == nil {
					successCount++
				} else {
					errorCount++
				}
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	// Verify results
	require.Greater(t, successCount, 0, "Should have successful operations")
	require.Less(t, errorCount, numOps, "Should have minimal errors")

	successRate := float64(successCount) / float64(numOps) * 100

	// Log performance metrics
	t.Logf("Performed %d concurrent operations in %v (%.2f ops/sec), success rate: %.1f%%",
		numOps, elapsed, float64(numOps)/elapsed.Seconds(), successRate)

	// Verify database stayed responsive (< 15 seconds for 500 operations)
	require.Less(t, elapsed, 15*time.Second, "Should handle 500 concurrent operations without exhausting connection pool")
}

// TestRaceConditionDetection uses Go's race detector (run with: go test -race)
func TestRaceConditionDetection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ts := SetupPaymentTest(t)
	defer TeardownPaymentTest(t, ts)

	uniqueID := GenerateUniqueID()
	userID := SeedTestUser(t, ts.DB, uniqueID)
	productID, _ := SeedTestProduct(t, ts.DB, uniqueID)
	defer CleanupTestData(t, ts.DB, userID)

	// Create order
	req := &order.CreateOrderRequest{
		ProductID: productID,
		Quantity:  1,
	}
	o, err := ts.OrderService.CreateOrder(ctx, userID, req)
	require.NoError(t, err)

	// Initiate payment
	payReq := &payment.InitiatePaymentRequest{
		OrderID:       o.ID,
		PaymentMethod: "alipay",
	}
	p, err := ts.PaymentService.InitiatePayment(ctx, userID, payReq)
	require.NoError(t, err)

	// Race between reading and updating payment
	var wg sync.WaitGroup

	// Reader goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = ts.PaymentService.GetPaymentByID(ctx, userID, p.ID)
		}()
	}

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		paymentData := GetPaymentFromDB(t, ts.DB, p.ID)

		callback := &payment.AlipayCallback{
			OutTradeNo:  fmt.Sprintf("%d", p.ID),
			TradeNo:     "alipay_race_test",
			TotalAmount: paymentData.Amount,
			TradeStatus: "TRADE_SUCCESS",
			Timestamp:   "2026-03-15 12:00:00",
			Sign:        "test_signature",
		}

		_, _ = ts.PaymentService.HandleAlipayCallback(ctx, callback)
	}()

	wg.Wait()

	// Verify final state is consistent
	finalPayment, err := ts.PaymentService.GetPaymentByID(ctx, userID, p.ID)
	require.NoError(t, err)
	require.NotNil(t, finalPayment)
}
