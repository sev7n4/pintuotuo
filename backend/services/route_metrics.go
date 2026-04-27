package services

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RouteDecisionCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "route_decision_total",
			Help: "Total number of route decisions",
		},
		[]string{"provider", "mode", "merchant_type", "region"},
	)

	ExecutionLayerLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "execution_layer_latency_ms",
			Help:    "Execution layer latency in milliseconds",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		},
		[]string{"provider", "mode", "success"},
	)

	RouteDecisionSourceCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "route_decision_source_total",
			Help: "Total number of route decisions by source",
		},
		[]string{"source"},
	)

	RouteCacheResultCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "route_cache_result_total",
			Help: "Total number of route cache results",
		},
		[]string{"result"},
	)

	ExecutionLayerErrorsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "execution_layer_errors_total",
			Help: "Total number of execution layer errors",
		},
		[]string{"provider", "mode", "error_type"},
	)
)

//nolint:unused // Will be used in Phase 3 Part 2
func categorizeError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()

	switch {
	case stringsContains(errStr, "timeout"):
		return "timeout"
	case stringsContains(errStr, "connection refused"):
		return "connection_refused"
	case stringsContains(errStr, "ECONNREFUSED"):
		return "connection_refused"
	case stringsContains(errStr, "rate limit"):
		return "rate_limit"
	case stringsContains(errStr, "429"):
		return "rate_limit"
	case stringsContains(errStr, "context canceled"):
		return "context_canceled"
	case stringsContains(errStr, "context deadline"):
		return "timeout"
	default:
		return "unknown"
	}
}

//nolint:unused // Will be used in Phase 3 Part 2
func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringsContainsHelper(s, substr))
}

//nolint:unused // Will be used in Phase 3 Part 2
func stringsContainsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
