package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pintuotuo/backend/models"
)

func TestHealthCheckLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected int
	}{
		{"high level", "high", 60},
		{"medium level", "medium", 300},
		{"low level", "low", 1800},
		{"daily level", "daily", 86400},
		{"unknown level defaults to medium", "unknown", 300},
		{"empty level defaults to medium", "", 300},
	}

	checker := NewHealthChecker()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.GetHealthCheckInterval(tt.level)
			if result != tt.expected {
				t.Errorf("GetHealthCheckInterval(%s) = %d, want %d", tt.level, result, tt.expected)
			}
		})
	}
}

func TestHealthStatusConstants(t *testing.T) {
	if HealthStatusHealthy != "healthy" {
		t.Errorf("HealthStatusHealthy = %s, want healthy", HealthStatusHealthy)
	}
	if HealthStatusDegraded != "degraded" {
		t.Errorf("HealthStatusDegraded = %s, want degraded", HealthStatusDegraded)
	}
	if HealthStatusUnhealthy != "unhealthy" {
		t.Errorf("HealthStatusUnhealthy = %s, want unhealthy", HealthStatusUnhealthy)
	}
	if HealthStatusUnknown != "unknown" {
		t.Errorf("HealthStatusUnknown = %s, want unknown", HealthStatusUnknown)
	}
}

func TestIsHealthy(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{HealthStatusHealthy, true},
		{HealthStatusDegraded, false},
		{HealthStatusUnhealthy, false},
		{HealthStatusUnknown, false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsHealthy(tt.status)
			if result != tt.expected {
				t.Errorf("IsHealthy(%s) = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestIsDegraded(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{HealthStatusHealthy, false},
		{HealthStatusDegraded, true},
		{HealthStatusUnhealthy, false},
		{HealthStatusUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsDegraded(tt.status)
			if result != tt.expected {
				t.Errorf("IsDegraded(%s) = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestIsUnhealthy(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{HealthStatusHealthy, false},
		{HealthStatusDegraded, false},
		{HealthStatusUnhealthy, true},
		{HealthStatusUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := IsUnhealthy(tt.status)
			if result != tt.expected {
				t.Errorf("IsUnhealthy(%s) = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestShouldPerformCheck(t *testing.T) {
	checker := NewHealthChecker()

	t.Run("should check when never checked", func(t *testing.T) {
		apiKey := &models.MerchantAPIKey{
			LastHealthCheckAt: nil,
			HealthCheckLevel:  "medium",
		}
		if !checker.ShouldPerformCheck(apiKey) {
			t.Error("ShouldPerformCheck should return true when never checked")
		}
	})

	t.Run("should check when interval elapsed", func(t *testing.T) {
		past := time.Now().Add(-10 * time.Minute)
		apiKey := &models.MerchantAPIKey{
			LastHealthCheckAt: &past,
			HealthCheckLevel:  "medium",
		}
		if !checker.ShouldPerformCheck(apiKey) {
			t.Error("ShouldPerformCheck should return true when interval elapsed")
		}
	})

	t.Run("should not check when interval not elapsed", func(t *testing.T) {
		recent := time.Now().Add(-1 * time.Minute)
		apiKey := &models.MerchantAPIKey{
			LastHealthCheckAt: &recent,
			HealthCheckLevel:  "medium",
		}
		if checker.ShouldPerformCheck(apiKey) {
			t.Error("ShouldPerformCheck should return false when interval not elapsed")
		}
	})
}

func TestResolveEndpoint(t *testing.T) {
	checker := NewHealthChecker()

	t.Run("prefer explicit endpoint_url", func(t *testing.T) {
		apiKey := &models.MerchantAPIKey{
			Provider:    "any",
			EndpointURL: "https://example.com/custom",
		}
		got, err := checker.resolveEndpoint(context.Background(), apiKey)
		if err != nil {
			t.Fatalf("resolveEndpoint returned error: %v", err)
		}
		if got != "https://example.com/custom" {
			t.Fatalf("resolveEndpoint got %s", got)
		}
	})

	t.Run("error when provider and fallback missing", func(t *testing.T) {
		apiKey := &models.MerchantAPIKey{Provider: "unknown_provider"}
		_, err := checker.resolveEndpoint(context.Background(), apiKey)
		if err == nil {
			t.Fatal("expected resolveEndpoint to return error when endpoint is not configured")
		}
	})
}

func TestHealthCheckResult(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := &HealthCheckResult{
			Success:     true,
			Status:      HealthStatusHealthy,
			LatencyMs:   100,
			ModelsFound: []string{"gpt-4", "gpt-3.5-turbo"},
			CheckType:   "full",
		}
		if !result.Success {
			t.Error("result.Success should be true")
		}
		if result.Status != HealthStatusHealthy {
			t.Errorf("result.Status = %s, want healthy", result.Status)
		}
		if len(result.ModelsFound) != 2 {
			t.Errorf("len(result.ModelsFound) = %d, want 2", len(result.ModelsFound))
		}
	})

	t.Run("failure result", func(t *testing.T) {
		result := &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			ErrorMessage: "connection refused",
			CheckType:    "lightweight",
		}
		if result.Success {
			t.Error("result.Success should be false")
		}
		if result.ErrorMessage == "" {
			t.Error("result.ErrorMessage should not be empty for failure")
		}
	})
}

func TestProviderHealth(t *testing.T) {
	now := time.Now()
	health := &ProviderHealth{
		APIKeyID:         1,
		Provider:         "openai",
		Status:           HealthStatusHealthy,
		LatencyMs:        150,
		ModelsAvailable:  []string{"gpt-4", "gpt-3.5-turbo"},
		LastCheckedAt:    now,
		FailureCount:     0,
		ConsecutiveCount: 0,
	}

	if health.APIKeyID != 1 {
		t.Errorf("health.APIKeyID = %d, want 1", health.APIKeyID)
	}
	if health.Provider != "openai" {
		t.Errorf("health.Provider = %s, want openai", health.Provider)
	}
	if health.Status != HealthStatusHealthy {
		t.Errorf("health.Status = %s, want healthy", health.Status)
	}
	if len(health.ModelsAvailable) != 2 {
		t.Errorf("len(health.ModelsAvailable) = %d, want 2", len(health.ModelsAvailable))
	}
}

func TestExtractPricingInfo(t *testing.T) {
	checker := NewHealthChecker()
	pricing := checker.extractPricingInfo("openai")

	if pricing == nil {
		t.Fatal("extractPricingInfo returned nil")
	}
	if pricing["provider"] != "openai" {
		t.Errorf("pricing[provider] = %v, want openai", pricing["provider"])
	}
	if _, ok := pricing["updated"]; !ok {
		t.Error("pricing should contain 'updated' field")
	}
}

func TestOpenAICompatProbeURLs(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		wantM    string
		wantChat string
	}{
		{
			name:     "base_without_v1",
			base:     "https://api.deepseek.com",
			wantM:    "https://api.deepseek.com/v1/models",
			wantChat: "https://api.deepseek.com/v1/chat/completions",
		},
		{
			name:     "base_with_trailing_slash_no_v1",
			base:     "https://api.siliconflow.cn/",
			wantM:    "https://api.siliconflow.cn/v1/models",
			wantChat: "https://api.siliconflow.cn/v1/chat/completions",
		},
		{
			name:     "base_already_openai_v1_root",
			base:     "https://api.stepfun.com/v1",
			wantM:    "https://api.stepfun.com/v1/models",
			wantChat: "https://api.stepfun.com/v1/chat/completions",
		},
		{
			name:     "zhipu_paas_v4_not_v1",
			base:     "https://open.bigmodel.cn/api/paas/v4",
			wantM:    "https://open.bigmodel.cn/api/paas/v4/models",
			wantChat: "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		},
		{
			name:     "base_v1_with_trailing_slash",
			base:     "https://api.openai.com/v1/",
			wantM:    "https://api.openai.com/v1/models",
			wantChat: "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "dashscope_compatible_mode_v1",
			base:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
			wantM:    "https://dashscope.aliyuncs.com/compatible-mode/v1/models",
			wantChat: "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if g := OpenAICompatModelsProbeURL(tt.base); g != tt.wantM {
				t.Errorf("models URL = %q, want %q", g, tt.wantM)
			}
			if g := OpenAICompatChatCompletionsURL(tt.base); g != tt.wantChat {
				t.Errorf("chat URL = %q, want %q", g, tt.wantChat)
			}
		})
	}
}

func TestLightweightPing_InvalidEndpoint(t *testing.T) {
	checker := NewHealthChecker()
	ctx := context.Background()

	apiKey := &models.MerchantAPIKey{
		ID:          1,
		Provider:    "custom",
		EndpointURL: "http://invalid-endpoint-that-does-not-exist.local",
	}

	result, err := checker.LightweightPing(ctx, apiKey)
	if err != nil {
		t.Logf("LightweightPing returned error (expected for invalid endpoint): %v", err)
	}

	if result != nil && result.Success {
		t.Error("LightweightPing should fail for invalid endpoint")
	}
}

func TestFullVerification_InvalidEndpoint(t *testing.T) {
	checker := NewHealthChecker()
	ctx := context.Background()

	apiKey := &models.MerchantAPIKey{
		ID:          1,
		Provider:    "custom",
		EndpointURL: "http://invalid-endpoint-that-does-not-exist.local",
	}

	result, err := checker.FullVerification(ctx, apiKey)
	if err != nil {
		t.Logf("FullVerification returned error (expected for invalid endpoint): %v", err)
	}

	if result != nil && result.Success {
		t.Error("FullVerification should fail for invalid endpoint")
	}
}

func TestTestChatCompletion_InvalidEndpoint(t *testing.T) {
	checker := NewHealthChecker()
	ctx := context.Background()

	apiKey := &models.MerchantAPIKey{
		ID:          1,
		Provider:    "custom",
		EndpointURL: "http://invalid-endpoint-that-does-not-exist.local",
	}

	result, err := checker.TestChatCompletion(ctx, apiKey, "gpt-4")
	if err != nil {
		t.Logf("TestChatCompletion returned error (expected for invalid endpoint): %v", err)
	}

	if result != nil && result.Success {
		t.Error("TestChatCompletion should fail for invalid endpoint")
	}
}

func TestTestChatCompletion_StructuredErrorFields(t *testing.T) {
	checker := NewHealthChecker()
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "chat-req-1")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":"invalid_api_key","message":"invalid api key"}}`))
	}))
	defer srv.Close()

	apiKey := &models.MerchantAPIKey{
		ID:          1,
		Provider:    "openai",
		EndpointURL: srv.URL,
	}

	result, err := checker.TestChatCompletion(ctx, apiKey, "gpt-4o-mini")
	if err != nil {
		t.Fatalf("TestChatCompletion unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Success {
		t.Fatal("expected failure")
	}
	if result.ErrorCategory != errorCategoryAuthInvalidKey {
		t.Fatalf("category=%s want=%s", result.ErrorCategory, errorCategoryAuthInvalidKey)
	}
	if result.ProviderErrorCode != "invalid_api_key" {
		t.Fatalf("provider_error_code=%s want=invalid_api_key", result.ProviderErrorCode)
	}
	if result.ProviderRequestID != "chat-req-1" {
		t.Fatalf("provider_request_id=%s want=chat-req-1", result.ProviderRequestID)
	}
	if result.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status_code=%d want=%d", result.StatusCode, http.StatusUnauthorized)
	}
	if result.EndpointUsed == "" {
		t.Fatal("endpoint_used should not be empty")
	}
}
