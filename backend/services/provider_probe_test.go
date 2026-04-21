package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeProviderModels_ErrorMapping(t *testing.T) {
	t.Run("401 invalid key", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Request-Id", "req-401")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"invalid_api_key","message":"invalid api key"}}`))
		}))
		defer srv.Close()

		res, err := ProbeProviderModels(context.Background(), &http.Client{Timeout: 2 * time.Second}, srv.URL+"/models", "k")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if res.Success {
			t.Fatal("expected failure")
		}
		if res.ErrorCategory != errorCategoryAuthInvalidKey {
			t.Fatalf("category=%s want=%s", res.ErrorCategory, errorCategoryAuthInvalidKey)
		}
		if res.ProviderRequestID != "req-401" {
			t.Fatalf("request_id=%s want=req-401", res.ProviderRequestID)
		}
	})

	t.Run("429 rate limited", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"code":"rate_limit_exceeded","message":"too many requests"}}`))
		}))
		defer srv.Close()

		res, err := ProbeProviderModels(context.Background(), &http.Client{Timeout: 2 * time.Second}, srv.URL+"/models", "k")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if res.ErrorCategory != errorCategoryRateLimited {
			t.Fatalf("category=%s want=%s", res.ErrorCategory, errorCategoryRateLimited)
		}
	})

	t.Run("network timeout maps category", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		// 10.255.255.1 常用于不可达地址，触发超时错误
		res, err := ProbeProviderModels(ctx, &http.Client{Timeout: 50 * time.Millisecond}, "http://10.255.255.1/models", "k")
		if err == nil {
			t.Fatal("expected timeout/network error")
		}
		if res == nil {
			t.Fatal("expected non-nil result")
		}
		if res.ErrorCategory != errorCategoryNetworkTimeout && res.ErrorCategory != errorCategoryServiceUnavailable {
			t.Fatalf("unexpected category=%s", res.ErrorCategory)
		}
	})
}
