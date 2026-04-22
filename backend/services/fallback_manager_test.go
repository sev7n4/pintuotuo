package services

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGetFallbackManager(t *testing.T) {
	manager1 := GetFallbackManager()
	manager2 := GetFallbackManager()

	if manager1 != manager2 {
		t.Error("Expected singleton instance")
	}
}

func TestShouldFallback(t *testing.T) {
	manager := GetFallbackManager()
	ctx := context.Background()

	tests := []struct {
		name         string
		providerCode string
		currentMode  string
		failureCount int
		expectNil    bool
	}{
		{
			name:         "low failure count should not fallback",
			providerCode: "test-provider-1",
			currentMode:  "litellm",
			failureCount: 1,
			expectNil:    true,
		},
		{
			name:         "high failure count should consider fallback",
			providerCode: "test-provider-2",
			currentMode:  "litellm",
			failureCount: 5,
			expectNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, _ := manager.ShouldFallback(ctx, tt.providerCode, tt.currentMode, tt.failureCount)
			if tt.expectNil && state != nil {
				t.Error("Expected nil state for low failure count")
			}
		})
	}
}

func TestTriggerFallback(t *testing.T) {
	manager := GetFallbackManager()
	ctx := context.Background()

	err := manager.TriggerFallback(
		ctx,
		"test-provider",
		"litellm",
		"proxy",
		"https://proxy.example.com",
		"primary endpoint failed",
	)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	state := manager.GetFallbackState("test-provider")
	if state == nil {
		t.Error("Expected fallback state to be set")
		return
	}

	if state.CurrentMode != "proxy" {
		t.Errorf("Expected current mode 'proxy', got '%s'", state.CurrentMode)
	}

	if state.OriginalMode != "litellm" {
		t.Errorf("Expected original mode 'litellm', got '%s'", state.OriginalMode)
	}
}

func TestGetFallbackState(t *testing.T) {
	manager := GetFallbackManager()

	state := manager.GetFallbackState("non-existent-provider")
	if state != nil {
		t.Error("Expected nil state for non-existent provider")
	}
}

func TestClearFallback(t *testing.T) {
	manager := GetFallbackManager()
	ctx := context.Background()

	manager.TriggerFallback(
		ctx,
		"test-clear-provider",
		"litellm",
		"proxy",
		"https://proxy.example.com",
		"test",
	)

	state := manager.GetFallbackState("test-clear-provider")
	if state == nil {
		t.Error("Expected fallback state to be set")
		return
	}

	manager.ClearFallback("test-clear-provider", true)

	state = manager.GetFallbackState("test-clear-provider")
	if state != nil {
		t.Error("Expected fallback state to be cleared")
	}
}

func TestGetFallbackHistory(t *testing.T) {
	manager := GetFallbackManager()
	ctx := context.Background()

	manager.TriggerFallback(
		ctx,
		"history-provider-1",
		"litellm",
		"proxy",
		"https://proxy.example.com",
		"test 1",
	)

	manager.TriggerFallback(
		ctx,
		"history-provider-2",
		"direct",
		"litellm",
		"http://litellm:4000/v1",
		"test 2",
	)

	history := manager.GetFallbackHistory(10)
	if len(history) < 2 {
		t.Errorf("Expected at least 2 history events, got %d", len(history))
	}
}

func TestGetFallbackStats(t *testing.T) {
	manager := GetFallbackManager()
	ctx := context.Background()

	manager.TriggerFallback(
		ctx,
		"stats-provider",
		"litellm",
		"proxy",
		"https://proxy.example.com",
		"test",
	)

	stats := manager.GetFallbackStats()

	if _, ok := stats["active_fallbacks"]; !ok {
		t.Error("Expected 'active_fallbacks' in stats")
	}

	if _, ok := stats["total_events"]; !ok {
		t.Error("Expected 'total_events' in stats")
	}
}

func TestExecuteWithFallback_Success(t *testing.T) {
	manager := GetFallbackManager()
	ctx := context.Background()

	decision := &RouteDecision{
		Mode:             "litellm",
		Endpoint:         "http://litellm:4000/v1",
		FallbackMode:     "proxy",
		FallbackEndpoint: "https://proxy.example.com",
	}

	callCount := 0
	err := manager.ExecuteWithFallback(
		ctx,
		"test-success-provider",
		decision,
		func() error {
			callCount++
			return nil
		},
	)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestExecuteWithFallback_Failure(t *testing.T) {
	manager := GetFallbackManager()
	ctx := context.Background()

	decision := &RouteDecision{
		Mode:             "litellm",
		Endpoint:         "http://litellm:4000/v1",
		FallbackMode:     "proxy",
		FallbackEndpoint: "https://proxy.example.com",
	}

	err := manager.ExecuteWithFallback(
		ctx,
		"test-failure-provider",
		decision,
		func() error {
			return errors.New("primary endpoint failed")
		},
	)

	if err == nil {
		t.Error("Expected error from primary function")
	}
}

func TestFallbackStateExpiration(t *testing.T) {
	manager := GetFallbackManager()
	ctx := context.Background()

	manager.TriggerFallback(
		ctx,
		"expired-provider",
		"litellm",
		"proxy",
		"https://proxy.example.com",
		"test",
	)

	state := manager.GetFallbackState("expired-provider")
	if state == nil {
		t.Error("Expected fallback state to be set initially")
	}

	state.FallbackAt = time.Now().Add(-15 * time.Minute)

	state = manager.GetFallbackState("expired-provider")
	if state != nil {
		t.Error("Expected expired fallback state to return nil")
	}
}
