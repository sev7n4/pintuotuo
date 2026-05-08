package metrics

import (
	"fmt"

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

	// Fuel pack restriction counter
	FuelPackRestrictionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fuel_pack_restriction_total",
			Help: "Total count of blocked fuel-pack-only purchase attempts",
		},
		[]string{"source", "reason_code"},
	)
)

// Pricing Service Metrics
var (
	// Pricing cache hits counter
	PricingCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pricing_cache_hits_total",
			Help: "Total number of pricing cache hits",
		},
		[]string{"provider", "model"},
	)

	// Pricing cache misses counter
	PricingCacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pricing_cache_misses_total",
			Help: "Total number of pricing cache misses",
		},
		[]string{"provider", "model"},
	)

	// Pricing calculation duration histogram
	PricingCalculationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pricing_calculation_duration_seconds",
			Help:    "Pricing calculation duration in seconds",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .005, .01},
		},
		[]string{"provider", "model"},
	)

	// Pricing cache reload counter
	PricingCacheReloads = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pricing_cache_reloads_total",
			Help: "Total number of pricing cache reloads",
		},
		[]string{"source"},
	)

	// Pricing cache size gauge
	PricingCacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pricing_cache_size",
			Help: "Number of items in pricing cache",
		},
		[]string{},
	)
)

// API Key Verification Metrics
var (
	// Verification total counter
	VerificationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_key_verification_total",
			Help: "Total number of API key verifications",
		},
		[]string{"provider", "verification_type", "status"},
	)

	// Verification duration histogram
	VerificationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_key_verification_duration_seconds",
			Help:    "API key verification duration in seconds",
			Buckets: []float64{.1, .5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"provider", "verification_type"},
	)

	// Verification cache hits counter
	VerificationCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_key_verification_cache_hits_total",
			Help: "Total number of verification cache hits",
		},
		[]string{"provider"},
	)

	// Verification cache misses counter
	VerificationCacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_key_verification_cache_misses_total",
			Help: "Total number of verification cache misses",
		},
		[]string{"provider"},
	)

	// Verification retry counter
	VerificationRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_key_verification_retries_total",
			Help: "Total number of verification retries",
		},
		[]string{"provider", "attempt"},
	)

	// Active verifications gauge
	ActiveVerifications = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "api_key_active_verifications",
			Help: "Number of active API key verifications",
		},
		[]string{"provider"},
	)

	// Verification connection latency histogram
	VerificationConnectionLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_key_verification_connection_latency_ms",
			Help:    "API key verification connection latency in milliseconds",
			Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000, 5000},
		},
		[]string{"provider"},
	)

	// Models discovered counter
	ModelsDiscovered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_key_models_discovered_total",
			Help: "Total number of models discovered during verification",
		},
		[]string{"provider"},
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
	HTTPRequestsTotal.WithLabelValues(method, endpoint, fmt.Sprintf("%d", status)).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(durationSeconds)
	HTTPRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSizeBytes))
	HTTPResponseSize.WithLabelValues(method, endpoint, fmt.Sprintf("%d", status)).Observe(float64(responseSizeBytes))
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

// RecordFuelPackRestriction records blocked fuel-pack-only attempts.
func RecordFuelPackRestriction(source, reasonCode string) {
	FuelPackRestrictionTotal.WithLabelValues(source, reasonCode).Inc()
}

// RecordApplicationError records an application error
func RecordApplicationError(errorCode, severity string) {
	ApplicationErrors.WithLabelValues(errorCode, severity).Inc()
}

// RecordPricingCacheHit records a pricing cache hit
func RecordPricingCacheHit(provider, model string) {
	PricingCacheHits.WithLabelValues(provider, model).Inc()
}

// RecordPricingCacheMiss records a pricing cache miss
func RecordPricingCacheMiss(provider, model string) {
	PricingCacheMisses.WithLabelValues(provider, model).Inc()
}

// RecordPricingCalculation records a pricing calculation
func RecordPricingCalculation(provider, model string, durationSeconds float64) {
	PricingCalculationDuration.WithLabelValues(provider, model).Observe(durationSeconds)
}

// RecordPricingCacheReload records a pricing cache reload
func RecordPricingCacheReload(source string) {
	PricingCacheReloads.WithLabelValues(source).Inc()
}

// SetPricingCacheSize sets the pricing cache size
func SetPricingCacheSize(size int) {
	PricingCacheSize.WithLabelValues().Set(float64(size))
}

// Routing Decision Metrics
var (
	RoutingDecisionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "routing_decision_total",
			Help: "Total number of routing decisions made",
		},
		[]string{"provider", "merchant_region", "merchant_type", "mode"},
	)

	RoutingDecisionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "routing_decision_duration_seconds",
			Help:    "Routing decision duration in seconds",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .005, .01},
		},
		[]string{"provider"},
	)

	RoutingFallbackTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "routing_fallback_total",
			Help: "Total number of routing fallbacks triggered",
		},
		[]string{"provider", "original_mode", "fallback_mode", "reason"},
	)

	RoutingFallbackActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "routing_fallback_active",
			Help: "Number of active routing fallbacks",
		},
		[]string{"provider"},
	)

	EndpointRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "endpoint_request_duration_seconds",
			Help:    "Endpoint request duration in seconds",
			Buckets: []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{"provider", "endpoint_type"},
	)

	EndpointRequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "endpoint_request_errors_total",
			Help: "Total number of endpoint request errors",
		},
		[]string{"provider", "endpoint_type", "error_type"},
	)

	EndpointHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "endpoint_health_status",
			Help: "Endpoint health status (1=healthy, 0.5=degraded, 0=unhealthy)",
		},
		[]string{"provider", "endpoint_type"},
	)

	APIKeyRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_key_requests_total",
			Help: "Total number of API proxy requests by provider and status",
		},
		[]string{"provider", "status"},
	)

	APIKeyRequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_key_request_errors_total",
			Help: "Total number of API proxy request errors by provider and error_type",
		},
		[]string{"provider", "error_type"},
	)

	APIKeyRequestLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_key_request_latency_seconds",
			Help:    "API proxy request latency in seconds by provider",
			Buckets: []float64{.05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
		},
		[]string{"provider"},
	)

	RateLimiterTokensRemaining = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rate_limiter_tokens_remaining",
			Help: "Remaining tokens in the rate limiter bucket",
		},
		[]string{"limiter_key"},
	)

	RateLimiterDeniedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limiter_denied_total",
			Help: "Total number of rate-limited (denied) requests",
		},
		[]string{"limiter_key"},
	)

	RequestQueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "request_queue_size",
			Help: "Current number of requests in the queue",
		},
		[]string{"queue_name"},
	)

	RequestQueueDroppedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "request_queue_dropped_total",
			Help: "Total number of dropped requests due to queue overflow",
		},
		[]string{"queue_name"},
	)
)

func RecordAPIKeyRequest(provider, status string) {
	APIKeyRequestTotal.WithLabelValues(provider, status).Inc()
}

func RecordAPIKeyRequestError(provider, errorType string) {
	APIKeyRequestErrors.WithLabelValues(provider, errorType).Inc()
}

func RecordAPIKeyRequestLatency(provider string, durationSeconds float64) {
	APIKeyRequestLatency.WithLabelValues(provider).Observe(durationSeconds)
}

func SetRateLimiterTokensRemaining(key string, tokens float64) {
	RateLimiterTokensRemaining.WithLabelValues(key).Set(tokens)
}

func RecordRateLimiterDenied(key string) {
	RateLimiterDeniedTotal.WithLabelValues(key).Inc()
}

func SetRequestQueueSize(queueName string, size int) {
	RequestQueueSize.WithLabelValues(queueName).Set(float64(size))
}

func RecordRequestQueueDropped(queueName string) {
	RequestQueueDroppedTotal.WithLabelValues(queueName).Inc()
}

// RecordRoutingDecision records a routing decision
func RecordRoutingDecision(provider, merchantRegion, merchantType, mode string, durationSeconds float64) {
	RoutingDecisionTotal.WithLabelValues(provider, merchantRegion, merchantType, mode).Inc()
	RoutingDecisionDuration.WithLabelValues(provider).Observe(durationSeconds)
}

// RecordRoutingFallback records a routing fallback event
func RecordRoutingFallback(provider, originalMode, fallbackMode, reason string) {
	RoutingFallbackTotal.WithLabelValues(provider, originalMode, fallbackMode, reason).Inc()
}

// SetRoutingFallbackActive sets the number of active fallbacks for a provider
func SetRoutingFallbackActive(provider string, count int) {
	RoutingFallbackActive.WithLabelValues(provider).Set(float64(count))
}

// RecordEndpointRequest records an endpoint request
func RecordEndpointRequest(provider, endpointType string, durationSeconds float64, isError bool, errorType string) {
	EndpointRequestDuration.WithLabelValues(provider, endpointType).Observe(durationSeconds)
	if isError {
		EndpointRequestErrors.WithLabelValues(provider, endpointType, errorType).Inc()
	}
}

// SetEndpointHealthStatus sets the health status of an endpoint
func SetEndpointHealthStatus(provider, endpointType string, status float64) {
	EndpointHealthStatus.WithLabelValues(provider, endpointType).Set(status)
}

// Token Estimation Metrics
var (
	TokenEstimationAccuracy = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "token_estimation_accuracy_ratio",
			Help:    "Token estimation accuracy ratio (estimated/actual)",
			Buckets: []float64{0.1, 0.25, 0.5, 0.75, 0.9, 1.0, 1.1, 1.25, 1.5, 2.0, 3.0, 5.0},
		},
		[]string{"model", "estimation_source"},
	)

	TokenEstimationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "token_estimation_total",
			Help: "Total number of token estimations",
		},
		[]string{"model", "estimation_source"},
	)

	TokenEstimationError = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "token_estimation_error_total",
			Help: "Total number of token estimation errors",
		},
		[]string{"model", "error_type"},
	)
)

// Cost Estimation Metrics
var (
	CostEstimationDeviation = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cost_estimation_deviation_ratio",
			Help:    "Cost estimation deviation ratio (estimated/actual)",
			Buckets: []float64{0.1, 0.25, 0.5, 0.75, 0.9, 1.0, 1.1, 1.25, 1.5, 2.0, 3.0, 5.0},
		},
		[]string{"model", "merchant_id"},
	)

	CostEstimationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cost_estimation_total",
			Help: "Total number of cost estimations",
		},
		[]string{"model", "strategy"},
	)

	CostSavingsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "routing_cost_savings_total_cents",
			Help: "Total cost savings from routing decisions in cents",
		},
		[]string{"model", "strategy"},
	)
)

// Five-Dimensional Scoring Metrics
var (
	RoutingScoreDistribution = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "routing_score_distribution",
			Help:    "Distribution of routing scores by dimension",
			Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		},
		[]string{"dimension", "strategy"},
	)

	RoutingWinnerScore = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "routing_winner_score",
			Help:    "Final score of winning candidate",
			Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		},
		[]string{"strategy", "provider"},
	)

	RoutingWeightDistribution = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "routing_weight_distribution",
			Help: "Current weight distribution for routing strategies",
		},
		[]string{"strategy", "dimension"},
	)

	RoutingCandidatesCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "routing_candidates_count",
			Help:    "Number of candidates evaluated per routing decision",
			Buckets: []float64{1, 2, 3, 5, 10, 20, 50},
		},
		[]string{"model", "strategy"},
	)
)

// RecordTokenEstimationAccuracy records token estimation accuracy
func RecordTokenEstimationAccuracy(model, source string, estimatedTokens, actualTokens int) {
	if actualTokens > 0 {
		ratio := float64(estimatedTokens) / float64(actualTokens)
		TokenEstimationAccuracy.WithLabelValues(model, source).Observe(ratio)
	}
	TokenEstimationTotal.WithLabelValues(model, source).Inc()
}

// RecordTokenEstimationError records a token estimation error
func RecordTokenEstimationError(model, errorType string) {
	TokenEstimationError.WithLabelValues(model, errorType).Inc()
}

// RecordCostEstimationDeviation records cost estimation deviation
func RecordCostEstimationDeviation(model, merchantID string, estimatedCost, actualCost float64) {
	if actualCost > 0 {
		ratio := estimatedCost / actualCost
		CostEstimationDeviation.WithLabelValues(model, merchantID).Observe(ratio)
	}
}

// RecordCostEstimation records a cost estimation
func RecordCostEstimation(model, strategy string) {
	CostEstimationTotal.WithLabelValues(model, strategy).Inc()
}

// RecordCostSavings records cost savings from routing
func RecordCostSavings(model, strategy string, savingsCents float64) {
	CostSavingsTotal.WithLabelValues(model, strategy).Add(savingsCents)
}

// RecordRoutingScore records a routing score by dimension
func RecordRoutingScore(dimension, strategy string, score float64) {
	RoutingScoreDistribution.WithLabelValues(dimension, strategy).Observe(score)
}

// RecordRoutingWinnerScore records the winning candidate's score
func RecordRoutingWinnerScore(strategy, provider string, score float64) {
	RoutingWinnerScore.WithLabelValues(strategy, provider).Observe(score)
}

// SetRoutingWeightDistribution sets the weight distribution for a strategy
func SetRoutingWeightDistribution(strategy, dimension string, weight float64) {
	RoutingWeightDistribution.WithLabelValues(strategy, dimension).Set(weight)
}

// RecordRoutingCandidatesCount records the number of candidates evaluated
func RecordRoutingCandidatesCount(model, strategy string, count int) {
	RoutingCandidatesCount.WithLabelValues(model, strategy).Observe(float64(count))
}
