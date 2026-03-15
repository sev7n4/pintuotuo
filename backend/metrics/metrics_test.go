package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMetricsInitialization verifies all metrics are initialized
func TestMetricsInitialization(t *testing.T) {
	// Verify HTTP metrics exist
	assert.NotNil(t, HTTPRequestsTotal, "HTTPRequestsTotal should be initialized")
	assert.NotNil(t, HTTPRequestDuration, "HTTPRequestDuration should be initialized")
	assert.NotNil(t, HTTPRequestSize, "HTTPRequestSize should be initialized")
	assert.NotNil(t, HTTPResponseSize, "HTTPResponseSize should be initialized")
	assert.NotNil(t, ActiveConnections, "ActiveConnections should be initialized")

	// Verify Database metrics exist
	assert.NotNil(t, DatabaseQueryDuration, "DatabaseQueryDuration should be initialized")
	assert.NotNil(t, DatabaseQueryErrors, "DatabaseQueryErrors should be initialized")
	assert.NotNil(t, DatabaseConnectionPoolSize, "DatabaseConnectionPoolSize should be initialized")
	assert.NotNil(t, DatabaseOpenConnections, "DatabaseOpenConnections should be initialized")
	assert.NotNil(t, DatabaseTransactionDuration, "DatabaseTransactionDuration should be initialized")
	assert.NotNil(t, DatabaseTransactionRollbacks, "DatabaseTransactionRollbacks should be initialized")

	// Verify Cache metrics exist
	assert.NotNil(t, CacheHitsMisses, "CacheHitsMisses should be initialized")
	assert.NotNil(t, CacheOperationDuration, "CacheOperationDuration should be initialized")
	assert.NotNil(t, CacheSize, "CacheSize should be initialized")
	assert.NotNil(t, CacheEvictions, "CacheEvictions should be initialized")

	// Verify Business metrics exist
	assert.NotNil(t, UserRegistrations, "UserRegistrations should be initialized")
	assert.NotNil(t, ActiveUsers, "ActiveUsers should be initialized")
	assert.NotNil(t, OrdersCreated, "OrdersCreated should be initialized")
	assert.NotNil(t, OrderValue, "OrderValue should be initialized")
	assert.NotNil(t, GroupsCreated, "GroupsCreated should be initialized")
	assert.NotNil(t, GroupCompletionRate, "GroupCompletionRate should be initialized")
	assert.NotNil(t, PaymentsProcessed, "PaymentsProcessed should be initialized")
	assert.NotNil(t, PaymentValue, "PaymentValue should be initialized")

	// Verify Error metrics exist
	assert.NotNil(t, ApplicationErrors, "ApplicationErrors should be initialized")
	assert.NotNil(t, ApplicationPanics, "ApplicationPanics should be initialized")

	// Verify System metrics exist
	assert.NotNil(t, Goroutines, "Goroutines should be initialized")
	assert.NotNil(t, MemoryUsageBytes, "MemoryUsageBytes should be initialized")
}

// TestRecordHTTPRequest verifies HTTP request recording
func TestRecordHTTPRequest(t *testing.T) {
	// This should not panic or error
	RecordHTTPRequest(
		"GET",
		"/api/v1/products",
		200,
		0.125,
		1024,
		2048,
	)

	// Test with error status
	RecordHTTPRequest(
		"POST",
		"/api/v1/auth/register",
		400,
		0.050,
		512,
		256,
	)

	// Test with various status codes
	statusCodes := []int{200, 201, 400, 401, 403, 404, 409, 500}
	for _, status := range statusCodes {
		RecordHTTPRequest(
			"GET",
			"/test",
			status,
			0.01,
			100,
			200,
		)
	}
}

// TestRecordDatabaseQuery verifies database query recording
func TestRecordDatabaseQuery(t *testing.T) {
	// Test successful query
	RecordDatabaseQuery("SELECT", "users", 0.025, false, "")

	// Test failed query
	RecordDatabaseQuery("INSERT", "orders", 0.050, true, "CONSTRAINT_VIOLATION")

	// Test various query types
	queryTypes := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
	tables := []string{"users", "orders", "products", "payments"}

	for _, queryType := range queryTypes {
		for _, table := range tables {
			RecordDatabaseQuery(queryType, table, 0.01, false, "")
		}
	}
}

// TestRecordCacheOperation verifies cache operation recording
func TestRecordCacheOperation(t *testing.T) {
	// Test cache hit
	RecordCacheOperation("redis", "get", 0.001, true)

	// Test cache miss
	RecordCacheOperation("redis", "get", 0.002, false)

	// Test various operations
	operations := []string{"get", "set", "delete", "exists"}
	for _, op := range operations {
		RecordCacheOperation("redis", op, 0.001, true)
		RecordCacheOperation("redis", op, 0.002, false)
	}
}

// TestRecordOrderCreation verifies order recording
func TestRecordOrderCreation(t *testing.T) {
	statuses := []string{"created", "paid", "completed", "cancelled"}
	currencies := []string{"CNY", "USD", "EUR"}

	for _, status := range statuses {
		for _, currency := range currencies {
			RecordOrderCreation(status, 10000, currency)
		}
	}
}

// TestRecordPaymentProcessed verifies payment recording
func TestRecordPaymentProcessed(t *testing.T) {
	methods := []string{"alipay", "wechat"}
	statuses := []string{"success", "failed", "pending"}

	for _, method := range methods {
		for _, status := range statuses {
			RecordPaymentProcessed(method, status, 50000)
		}
	}
}

// TestRecordApplicationError verifies error recording
func TestRecordApplicationError(t *testing.T) {
	errorCodes := []string{
		"USER_NOT_FOUND",
		"INVALID_CREDENTIALS",
		"INSUFFICIENT_STOCK",
		"ORDER_NOT_FOUND",
		"PAYMENT_FAILED",
	}

	severities := []string{"error", "warning", "critical"}

	for _, code := range errorCodes {
		for _, severity := range severities {
			RecordApplicationError(code, severity)
		}
	}
}

// TestMetricsFormatting verifies metric values are properly formatted
func TestMetricsFormatting(t *testing.T) {
	// Test various duration values
	durations := []float64{
		0.0001, // 0.1ms
		0.001,  // 1ms
		0.01,   // 10ms
		0.1,    // 100ms
		1.0,    // 1s
	}

	for _, duration := range durations {
		RecordHTTPRequest("GET", "/test", 200, duration, 100, 200)
	}

	// Test various size values
	sizes := []int64{
		0,
		256,
		1024,
		10240,
		1048576, // 1MB
	}

	for _, size := range sizes {
		RecordHTTPRequest("POST", "/upload", 200, 0.1, size, size*2)
	}

	// Test various numeric values
	values := []int64{
		100,      // $1.00
		10000,    // $100.00
		100000,   // $1000.00
		1000000,  // $10000.00
	}

	for _, value := range values {
		RecordOrderCreation("created", value, "CNY")
	}
}

// TestConcurrentMetricsRecording verifies metrics can be recorded concurrently
func TestConcurrentMetricsRecording(t *testing.T) {
	done := make(chan bool, 10)

	// Simulate concurrent requests
	for i := 0; i < 10; i++ {
		go func() {
			RecordHTTPRequest(
				"GET",
				"/api/v1/products",
				200,
				0.05,
				512,
				1024,
			)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
