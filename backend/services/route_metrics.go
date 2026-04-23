package services

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	APIKeyLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_key_latency_seconds",
			Help:    "API Key latency in seconds",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"api_key_id", "provider"},
	)

	APIKeyErrorRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "api_key_error_rate",
			Help: "API Key error rate",
		},
		[]string{"api_key_id", "provider"},
	)

	APIKeySuccessRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "api_key_success_rate",
			Help: "API Key success rate",
		},
		[]string{"api_key_id", "provider"},
	)

	RateLimiterTokensRemaining = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rate_limiter_tokens_remaining",
			Help: "Remaining tokens in rate limiter",
		},
		[]string{"key"},
	)

	RequestQueueLength = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "request_queue_length",
			Help: "Length of request queue",
		},
		[]string{"priority"},
	)

	ConnectionPoolActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "connection_pool_active",
			Help: "Number of active connections in pool",
		},
		[]string{"api_key_id"},
	)

	ConnectionPoolSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "connection_pool_size",
			Help: "Total size of connection pool",
		},
		[]string{"api_key_id"},
	)

	RouteAwarenessUpdateTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "route_awareness_update_total",
			Help: "Total number of route awareness updates",
		},
		[]string{"api_key_id", "status"},
	)

	StatusCollectorRunsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "status_collector_runs_total",
			Help: "Total number of status collector runs",
		},
		[]string{"status"},
	)

	RoutingDecisionDurationByStrategy = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "routing_decision_duration_by_strategy_seconds",
			Help:    "Routing decision duration by strategy in seconds",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .005, .01},
		},
		[]string{"strategy", "provider"},
	)

	RoutingDecisionTotalByStrategy = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "routing_decision_total_by_strategy",
			Help: "Total number of routing decisions by strategy",
		},
		[]string{"strategy", "provider", "result"},
	)
)

type MetricsRecorder struct{}

func NewMetricsRecorder() *MetricsRecorder {
	return &MetricsRecorder{}
}

func (m *MetricsRecorder) RecordAPIKeyLatency(apiKeyID int, provider string, latencySeconds float64) {
	APIKeyLatency.WithLabelValues(
		strconv.Itoa(apiKeyID),
		provider,
	).Observe(latencySeconds)
}

func (m *MetricsRecorder) UpdateAPIKeyErrorRate(apiKeyID int, provider string, errorRate float64) {
	APIKeyErrorRate.WithLabelValues(
		strconv.Itoa(apiKeyID),
		provider,
	).Set(errorRate)
}

func (m *MetricsRecorder) UpdateAPIKeySuccessRate(apiKeyID int, provider string, successRate float64) {
	APIKeySuccessRate.WithLabelValues(
		strconv.Itoa(apiKeyID),
		provider,
	).Set(successRate)
}

func (m *MetricsRecorder) UpdateRateLimiterTokens(key string, remaining float64) {
	RateLimiterTokensRemaining.WithLabelValues(key).Set(remaining)
}

func (m *MetricsRecorder) UpdateRequestQueueLength(priority string, length float64) {
	RequestQueueLength.WithLabelValues(priority).Set(length)
}

func (m *MetricsRecorder) UpdateConnectionPool(apiKeyID int, active, size float64) {
	apiKeyIDStr := strconv.Itoa(apiKeyID)
	ConnectionPoolActive.WithLabelValues(apiKeyIDStr).Set(active)
	ConnectionPoolSize.WithLabelValues(apiKeyIDStr).Set(size)
}

func (m *MetricsRecorder) RecordRouteAwarenessUpdate(apiKeyID int, status string) {
	RouteAwarenessUpdateTotal.WithLabelValues(
		strconv.Itoa(apiKeyID),
		status,
	).Inc()
}

func (m *MetricsRecorder) RecordStatusCollectorRun(status string) {
	StatusCollectorRunsTotal.WithLabelValues(status).Inc()
}

func (m *MetricsRecorder) RecordRoutingDecisionByStrategy(strategy, provider string, durationSeconds float64, result string) {
	RoutingDecisionDurationByStrategy.WithLabelValues(strategy, provider).Observe(durationSeconds)
	RoutingDecisionTotalByStrategy.WithLabelValues(strategy, provider, result).Inc()
}
