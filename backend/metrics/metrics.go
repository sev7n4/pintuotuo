package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP Metrics
var (
	// HTTP request counter
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP request duration histogram
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.001, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// HTTP request size histogram
	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint"},
	)

	// HTTP response size histogram
	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint", "status"},
	)

	// Active connections gauge
	ActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "Number of active HTTP connections",
		},
		[]string{"endpoint"},
	)
)

// Database Metrics
var (
	// Database query duration
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"query_type", "table"},
	)

	// Database query errors
	DatabaseQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_errors_total",
			Help: "Total number of database query errors",
		},
		[]string{"query_type", "table", "error_type"},
	)

	// Database connection pool size
	DatabaseConnectionPoolSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connection_pool_size",
			Help: "Database connection pool size",
		},
		[]string{"pool"},
	)

	// Database open connections
	DatabaseOpenConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_open_connections",
			Help: "Number of open database connections",
		},
		[]string{"pool"},
	)

	// Database transaction duration
	DatabaseTransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_transaction_duration_seconds",
			Help:    "Database transaction duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"transaction_type"},
	)

	// Database transaction rollbacks
	DatabaseTransactionRollbacks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_transaction_rollbacks_total",
			Help: "Total number of database transaction rollbacks",
		},
		[]string{"transaction_type"},
	)
)

// Cache Metrics
var (
	// Cache hit/miss counter
	CacheHitsMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_misses_total",
			Help: "Total number of cache hits and misses",
		},
		[]string{"cache_name", "result"},
	)

	// Cache operation duration
	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Cache operation duration in seconds",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .05, .1},
		},
		[]string{"cache_name", "operation"},
	)

	// Cache size gauge
	CacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_size_bytes",
			Help: "Cache size in bytes",
		},
		[]string{"cache_name"},
	)

	// Cache evictions counter
	CacheEvictions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_evictions_total",
			Help: "Total number of cache evictions",
		},
		[]string{"cache_name", "reason"},
	)
)

// Business Metrics
var (
	// User registrations counter
	UserRegistrations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations",
		},
		[]string{"status"},
	)

	// Active users gauge
	ActiveUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
		[]string{},
	)

	// Orders created counter
	OrdersCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		},
		[]string{"status"},
	)

	// Order value histogram
	OrderValue = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_value_cents",
			Help:    "Order value in cents",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"currency"},
	)

	// Groups created counter
	GroupsCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "groups_created_total",
			Help: "Total number of groups created",
		},
		[]string{"status"},
	)

	// Group completion rate gauge
	GroupCompletionRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "group_completion_rate",
			Help: "Group completion rate (0-1)",
		},
		[]string{},
	)

	// Payments processed counter
	PaymentsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_processed_total",
			Help: "Total number of payments processed",
		},
		[]string{"method", "status"},
	)

	// Payment value histogram
	PaymentValue = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_value_cents",
			Help:    "Payment value in cents",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"method"},
	)
)

// Error Metrics
var (
	// Application errors counter
	ApplicationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "application_errors_total",
			Help: "Total number of application errors",
		},
		[]string{"error_code", "severity"},
	)

	// Panic counter
	ApplicationPanics = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "application_panics_total",
			Help: "Total number of application panics",
		},
		[]string{"component"},
	)
)

// System Metrics
var (
	// Goroutines gauge
	Goroutines = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "application_goroutines",
			Help: "Number of active goroutines",
		},
		[]string{},
	)

	// Memory usage gauge
	MemoryUsageBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "application_memory_usage_bytes",
			Help: "Application memory usage in bytes",
		},
		[]string{"type"},
	)
)

// Helper functions for metrics recording

// RecordHTTPRequest records an HTTP request
func RecordHTTPRequest(method, endpoint string, status int, durationSeconds float64, requestSizeBytes, responseSizeBytes int64) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, string(rune(status))).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(durationSeconds)
	HTTPRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSizeBytes))
	HTTPResponseSize.WithLabelValues(method, endpoint, string(rune(status))).Observe(float64(responseSizeBytes))
}

// RecordDatabaseQuery records a database query
func RecordDatabaseQuery(queryType, table string, durationSeconds float64, isError bool, errorType string) {
	DatabaseQueryDuration.WithLabelValues(queryType, table).Observe(durationSeconds)
	if isError {
		DatabaseQueryErrors.WithLabelValues(queryType, table, errorType).Inc()
	}
}

// RecordCacheOperation records a cache operation
func RecordCacheOperation(cacheName string, operation string, durationSeconds float64, isHit bool) {
	result := "miss"
	if isHit {
		result = "hit"
	}
	CacheHitsMisses.WithLabelValues(cacheName, result).Inc()
	CacheOperationDuration.WithLabelValues(cacheName, operation).Observe(durationSeconds)
}

// RecordOrderCreation records an order creation
func RecordOrderCreation(status string, valueInCents int64, currency string) {
	OrdersCreated.WithLabelValues(status).Inc()
	OrderValue.WithLabelValues(currency).Observe(float64(valueInCents))
}

// RecordPaymentProcessed records a payment
func RecordPaymentProcessed(method, status string, valueInCents int64) {
	PaymentsProcessed.WithLabelValues(method, status).Inc()
	PaymentValue.WithLabelValues(method).Observe(float64(valueInCents))
}

// RecordApplicationError records an application error
func RecordApplicationError(errorCode, severity string) {
	ApplicationErrors.WithLabelValues(errorCode, severity).Inc()
}
