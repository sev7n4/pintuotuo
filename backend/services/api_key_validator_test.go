package services

import (
	"errors"
	"testing"
	"time"
)

func TestAPIKeyValidatorStruct(t *testing.T) {
	validator := &APIKeyValidator{
		db: nil,
	}

	if validator == nil {
		t.Fatal("APIKeyValidator should not be nil")
	}
}

func TestVerificationResultStruct(t *testing.T) {
	result := VerificationResult{
		APIKeyID:          1,
		Status:            "success",
		ConnectionTest:    true,
		ConnectionLatency: 150,
		ModelsFound:       []string{"gpt-4", "gpt-3.5-turbo"},
		ModelsCount:       2,
		PricingVerified:   true,
		VerificationType:  "initial",
		StartedAt:         time.Now(),
		CompletedAt:       time.Now(),
	}

	if result.APIKeyID != 1 {
		t.Errorf("VerificationResult APIKeyID = %v, want 1", result.APIKeyID)
	}
	if result.Status != "success" {
		t.Errorf("VerificationResult Status = %v, want success", result.Status)
	}
	if !result.ConnectionTest {
		t.Errorf("VerificationResult ConnectionTest should be true")
	}
	if len(result.ModelsFound) != 2 {
		t.Errorf("VerificationResult ModelsFound length = %v, want 2", len(result.ModelsFound))
	}
}

func TestGetAPIKeyValidator(t *testing.T) {
	validator1 := GetAPIKeyValidator()
	validator2 := GetAPIKeyValidator()

	if validator1 == nil {
		t.Fatal("GetAPIKeyValidator() returned nil")
	}

	if validator1 != validator2 {
		t.Fatal("GetAPIKeyValidator() should return singleton instance")
	}
}

func TestNormalizeVerificationDBStatus(t *testing.T) {
	if got := normalizeVerificationDBStatus("success"); got != "verified" {
		t.Fatalf("normalizeVerificationDBStatus(success) = %s, want verified", got)
	}
	if got := normalizeVerificationDBStatus("failed"); got != "failed" {
		t.Fatalf("normalizeVerificationDBStatus(failed) = %s, want failed", got)
	}
}

func TestExtractProviderError(t *testing.T) {
	code, msg := ExtractProviderError([]byte(`{"error":{"code":"insufficient_quota","message":"余额不足"}}`))
	if code != "insufficient_quota" {
		t.Fatalf("code = %s, want insufficient_quota", code)
	}
	if msg != "余额不足" {
		t.Fatalf("msg = %s, want 余额不足", msg)
	}
}

func TestValidateAsync(t *testing.T) {
	validator := GetAPIKeyValidator()

	tests := []struct {
		name         string
		apiKeyID     int
		provider     string
		encryptedKey string
		wantErr      bool
	}{
		{
			name:         "Valid API Key ID",
			apiKeyID:     1,
			provider:     "openai",
			encryptedKey: "encrypted-key-data",
			wantErr:      false,
		},
		{
			name:         "Invalid API Key ID",
			apiKeyID:     0,
			provider:     "openai",
			encryptedKey: "encrypted-key-data",
			wantErr:      true,
		},
		{
			name:         "Empty provider",
			apiKeyID:     1,
			provider:     "",
			encryptedKey: "encrypted-key-data",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAsync(tt.apiKeyID, tt.provider, tt.encryptedKey, "initial")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAsync() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsVerified(t *testing.T) {
	validator := GetAPIKeyValidator()

	tests := []struct {
		name       string
		apiKeyID   int
		wantStatus string
	}{
		{
			name:       "Check verification status",
			apiKeyID:   1,
			wantStatus: "pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verified, err := validator.IsVerified(tt.apiKeyID)
			if err != nil {
				t.Logf("IsVerified() error = %v (expected for non-existent key)", err)
			}
			t.Logf("IsVerified() returned verified=%v", verified)
		})
	}
}

func TestGetVerificationHistory(t *testing.T) {
	validator := GetAPIKeyValidator()

	tests := []struct {
		name     string
		apiKeyID int
		limit    int
	}{
		{
			name:     "Get verification history",
			apiKeyID: 1,
			limit:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history, err := validator.GetVerificationHistory(tt.apiKeyID, tt.limit)
			if err != nil {
				t.Logf("GetVerificationHistory() error = %v (expected for non-existent key)", err)
			}
			t.Logf("GetVerificationHistory() returned %d records", len(history))
		})
	}
}

func TestShouldRetryVerificationAttempt(t *testing.T) {
	tests := []struct {
		name     string
		probe    *ProbeModelsResult
		probeErr error
		want     bool
	}{
		{
			name:     "retry on network timeout",
			probeErr: errors.New("dial tcp timeout"),
			want:     true,
		},
		{
			name: "retry on 429",
			probe: &ProbeModelsResult{
				StatusCode: 429,
				ErrorCode:  "rate_limit_exceeded",
				ErrorMsg:   "too many requests",
			},
			want: true,
		},
		{
			name: "do not retry on invalid api key",
			probe: &ProbeModelsResult{
				StatusCode: 401,
				ErrorCode:  "invalid_api_key",
				ErrorMsg:   "invalid api key",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRetryVerificationAttempt(tt.probe, tt.probeErr)
			if got != tt.want {
				t.Fatalf("shouldRetryVerificationAttempt()=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestLitellmProviderPrefix(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"openai", "openai"},
		{"anthropic", "anthropic"},
		{"deepseek", "deepseek"},
		{"alibaba", "dashscope"},
		{"zhipu", "zai"},
		{"moonshot", "moonshot"},
		{"minimax", "minimax"},
		{"google", "gemini"},
		{"stepfun", "openai"},
		{"bytedance", ""},
		{"unknown_provider", "unknown_provider"},
		{"ALIBABA", "dashscope"},
		{"  alibaba  ", "dashscope"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := litellmProviderPrefix(tt.provider)
			if got != tt.want {
				t.Errorf("litellmProviderPrefix(%q) = %q, want %q", tt.provider, got, tt.want)
			}
		})
	}
}

func TestResolveLitellmModelName(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		want     string
		wantErr  bool
	}{
		{
			name:     "model without slash - alibaba",
			provider: "alibaba",
			model:    "qwen-turbo",
			want:     "dashscope/qwen-turbo",
			wantErr:  false,
		},
		{
			name:     "model with slash - alibaba kimi prefix",
			provider: "alibaba",
			model:    "kimi/kimi-k2.6",
			want:     "dashscope/kimi-k2.6",
			wantErr:  false,
		},
		{
			name:     "model with slash - openai prefix",
			provider: "openai",
			model:    "openai/gpt-4",
			want:     "openai/gpt-4",
			wantErr:  false,
		},
		{
			name:     "model with slash - unknown prefix",
			provider: "alibaba",
			model:    "unknown/model-name",
			want:     "dashscope/model-name",
			wantErr:  false,
		},
		{
			name:     "bytedance unsupported",
			provider: "bytedance",
			model:    "doubao-lite",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "deepseek model without slash",
			provider: "deepseek",
			model:    "deepseek-chat",
			want:     "deepseek/deepseek-chat",
			wantErr:  false,
		},
		{
			name:     "google model with gemini prefix",
			provider: "google",
			model:    "gemini/gemini-pro",
			want:     "gemini/gemini-pro",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveLitellmModelName(tt.provider, tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveLitellmModelName(%q, %q) error = %v, wantErr %v", tt.provider, tt.model, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveLitellmModelName(%q, %q) = %q, want %q", tt.provider, tt.model, got, tt.want)
			}
		})
	}
}
