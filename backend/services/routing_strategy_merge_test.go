package services

import (
	"testing"
)

func TestMergeRoutingStrategyUpdate(t *testing.T) {
	existing := &RoutingStrategyConfig{
		ID:                      3,
		Name:                    "均衡策略",
		Code:                    "balanced",
		PriceWeight:             0.33,
		LatencyWeight:           0.34,
		ReliabilityWeight:       0.33,
		MaxRetryCount:           3,
		RetryBackoffBase:        1000,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   60,
		IsDefault:               true,
		Status:                  "active",
	}

	t.Run("name_only_partial_keeps_retry_fields", func(t *testing.T) {
		patch := &RoutingStrategyConfig{Name: "Updated Name"}
		got := mergeRoutingStrategyUpdate(existing, patch)
		if got.Name != "Updated Name" {
			t.Fatalf("Name = %q", got.Name)
		}
		if got.MaxRetryCount != 3 || got.RetryBackoffBase != 1000 {
			t.Fatalf("expected retry fields preserved, got max=%d backoff=%d", got.MaxRetryCount, got.RetryBackoffBase)
		}
	})

	t.Run("full_weight_patch_updates_retry_block", func(t *testing.T) {
		patch := &RoutingStrategyConfig{
			PriceWeight:             0.2,
			LatencyWeight:           0.2,
			ReliabilityWeight:       0.6,
			MaxRetryCount:           5,
			RetryBackoffBase:        2000,
			CircuitBreakerThreshold: 3,
			CircuitBreakerTimeout:   120,
		}
		got := mergeRoutingStrategyUpdate(existing, patch)
		if got.MaxRetryCount != 5 || got.RetryBackoffBase != 2000 {
			t.Fatalf("retry not applied")
		}
		if got.CircuitBreakerThreshold != 3 || got.CircuitBreakerTimeout != 120 {
			t.Fatalf("circuit not applied")
		}
	})

	t.Run("status_patch", func(t *testing.T) {
		patch := &RoutingStrategyConfig{Status: "inactive", IsDefault: false}
		got := mergeRoutingStrategyUpdate(existing, patch)
		if got.Status != "inactive" {
			t.Fatalf("Status = %q", got.Status)
		}
	})
}
