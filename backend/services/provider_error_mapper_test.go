package services

import (
	"errors"
	"net/http"
	"testing"
)

func TestMapProviderError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		code       string
		msg        string
		headers    http.Header
		netErr     error
		wantCat    string
		retryable  bool
	}{
		{
			name:       "invalid api key by code",
			statusCode: http.StatusUnauthorized,
			code:       "invalid_api_key",
			msg:        "invalid api key",
			wantCat:    errorCategoryAuthInvalidKey,
		},
		{
			name:       "rate limited",
			statusCode: http.StatusTooManyRequests,
			code:       "rate_limit_exceeded",
			msg:        "too many requests",
			wantCat:    errorCategoryRateLimited,
			retryable:  true,
		},
		{
			name:       "quota insufficient",
			statusCode: http.StatusPaymentRequired,
			code:       "insufficient_quota",
			msg:        "quota exceeded",
			wantCat:    errorCategoryQuotaInsufficient,
		},
		{
			name:       "network timeout",
			netErr:     errors.New("dial tcp timeout"),
			wantCat:    errorCategoryNetworkTimeout,
			retryable:  true,
		},
		{
			name:       "request id extraction",
			statusCode: http.StatusBadRequest,
			code:       "bad_request",
			msg:        "bad request",
			headers:    http.Header{"X-Request-Id": []string{"req-123"}},
			wantCat:    errorCategoryUpstreamBadRequest,
		},
		{
			name:       "502 bad gateway retryable",
			statusCode: http.StatusBadGateway,
			code:       "",
			msg:        "",
			wantCat:    errorCategoryServiceUnavailable,
			retryable:  true,
		},
		{
			name:       "500 internal retryable",
			statusCode: http.StatusInternalServerError,
			code:       "",
			msg:        "",
			wantCat:    errorCategoryServiceUnavailable,
			retryable:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapProviderError(tt.statusCode, tt.code, tt.msg, tt.headers, tt.netErr, `{"error":"x"}`)
			if got.Category != tt.wantCat {
				t.Fatalf("category=%s want=%s", got.Category, tt.wantCat)
			}
			if got.Retryable != tt.retryable {
				t.Fatalf("retryable=%v want=%v", got.Retryable, tt.retryable)
			}
			if tt.name == "request id extraction" && got.ProviderRequestID != "req-123" {
				t.Fatalf("request_id=%s want=req-123", got.ProviderRequestID)
			}
		})
	}
}

func TestSuggestModelFallbackAfterFailure(t *testing.T) {
	if !SuggestModelFallbackAfterFailure(MapProviderError(http.StatusNotFound, "model_not_found", "x", nil, nil, "")) {
		t.Fatal("model not found should suggest fallback")
	}
	if SuggestModelFallbackAfterFailure(MapProviderError(http.StatusUnauthorized, "invalid_api_key", "x", nil, nil, "")) {
		t.Fatal("invalid key should not suggest fallback")
	}
	if !SuggestModelFallbackAfterFailure(MapProviderError(http.StatusTooManyRequests, "rate", "limit", nil, nil, "")) {
		t.Fatal("429 should suggest fallback")
	}
}

func TestHTTPUpstreamRetryable(t *testing.T) {
	if !HTTPUpstreamRetryable(http.StatusTooManyRequests, []byte(`{}`), nil) {
		t.Fatal("429 should be retryable")
	}
	if !HTTPUpstreamRetryable(http.StatusBadGateway, []byte(`{}`), nil) {
		t.Fatal("502 should be retryable")
	}
	if HTTPUpstreamRetryable(http.StatusUnauthorized, []byte(`{"error":{"code":"invalid_api_key"}}`), nil) {
		t.Fatal("401 invalid key should not retry")
	}
}
