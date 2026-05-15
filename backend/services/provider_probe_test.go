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

		res, err := ProbeProviderModels(context.Background(), &http.Client{Timeout: 2 * time.Second}, srv.URL+"/models", "k", "openai")
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

		res, err := ProbeProviderModels(context.Background(), &http.Client{Timeout: 2 * time.Second}, srv.URL+"/models", "k", "openai")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if res.ErrorCategory != errorCategoryRateLimited {
			t.Fatalf("category=%s want=%s", res.ErrorCategory, errorCategoryRateLimited)
		}
	})

	t.Run("network timeout maps category", func(t *testing.T) {
		// 使用本机 httptest + 短生命周期 context，避免依赖「黑洞 IP」等环境相关行为（部分环境会偶发连上或快速返回非错误路径）。
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
		defer cancel()

		res, err := ProbeProviderModels(ctx, &http.Client{Timeout: 5 * time.Second}, srv.URL, "k", "openai")
		if err == nil {
			t.Fatal("expected timeout/context error")
		}
		if res == nil {
			t.Fatal("expected non-nil result")
		}
		if res.ErrorCategory != errorCategoryNetworkTimeout && res.ErrorCategory != errorCategoryServiceUnavailable {
			t.Fatalf("unexpected category=%s", res.ErrorCategory)
		}
	})
}

func TestSetProviderAuthHeaders(t *testing.T) {
	t.Run("openai uses Bearer token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/models", nil)
		SetProviderAuthHeaders(req, "openai", "sk-test123")
		if got := req.Header.Get("Authorization"); got != "Bearer sk-test123" {
			t.Fatalf("Authorization=%s want=Bearer sk-test123", got)
		}
		if got := req.Header.Get("x-api-key"); got != "" {
			t.Fatalf("x-api-key should be empty, got=%s", got)
		}
	})

	t.Run("anthropic uses x-api-key header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/models", nil)
		SetProviderAuthHeaders(req, "anthropic", "sk-ant-test123")
		if got := req.Header.Get("x-api-key"); got != "sk-ant-test123" {
			t.Fatalf("x-api-key=%s want=sk-ant-test123", got)
		}
		if got := req.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Fatalf("anthropic-version=%s want=2023-06-01", got)
		}
		if got := req.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization should be empty for anthropic, got=%s", got)
		}
	})

	t.Run("alibaba_anthropic uses x-api-key header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/messages", nil)
		SetProviderAuthHeaders(req, "alibaba_anthropic", "sk-dashscope")
		if got := req.Header.Get("x-api-key"); got != "sk-dashscope" {
			t.Fatalf("x-api-key=%s want=sk-dashscope", got)
		}
		if got := req.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization should be empty, got=%s", got)
		}
	})

	t.Run("anthropic case insensitive", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/models", nil)
		SetProviderAuthHeaders(req, "Anthropic", "sk-ant-test")
		if got := req.Header.Get("x-api-key"); got != "sk-ant-test" {
			t.Fatalf("x-api-key=%s want=sk-ant-test", got)
		}
	})

	t.Run("unknown provider defaults to Bearer", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/models", nil)
		SetProviderAuthHeaders(req, "google", "AIza-test")
		if got := req.Header.Get("Authorization"); got != "Bearer AIza-test" {
			t.Fatalf("Authorization=%s want=Bearer AIza-test", got)
		}
		if got := req.Header.Get("x-api-key"); got != "" {
			t.Fatalf("x-api-key should be empty, got=%s", got)
		}
	})

	t.Run("content-type always set", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/models", nil)
		SetProviderAuthHeaders(req, "openai", "key")
		if got := req.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type=%s want=application/json", got)
		}
	})
}

func TestProbeProviderModels_AnthropicAuthHeaders(t *testing.T) {
	var receivedAuth string
	var receivedAPIKey string
	var receivedVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedAPIKey = r.Header.Get("x-api-key")
		receivedVersion = r.Header.Get("anthropic-version")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"id":"claude-3-5-sonnet-20241022"}]}`))
	}))
	defer srv.Close()

	_, err := ProbeProviderModels(context.Background(), &http.Client{Timeout: 2 * time.Second}, srv.URL+"/models", "sk-ant-key", "anthropic")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if receivedAuth != "" {
		t.Fatalf("Authorization should be empty for anthropic, got=%s", receivedAuth)
	}
	if receivedAPIKey != "sk-ant-key" {
		t.Fatalf("x-api-key=%s want=sk-ant-key", receivedAPIKey)
	}
	if receivedVersion != "2023-06-01" {
		t.Fatalf("anthropic-version=%s want=2023-06-01", receivedVersion)
	}
}

func TestAnthropicMessagesProbeURL(t *testing.T) {
	// 阿里云百炼 Anthropic 兼容：base 以 /v1 结尾时只追加 /messages
	if got := AnthropicMessagesProbeURL("https://dashscope.aliyuncs.com/apps/anthropic/v1"); got != "https://dashscope.aliyuncs.com/apps/anthropic/v1/messages" {
		t.Fatalf("got %s", got)
	}
}

func TestProbeProviderConnectivity_AnthropicMessages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" && r.URL.Path != "/messages" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s want POST", r.Method)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Fatalf("x-api-key=%s", r.Header.Get("x-api-key"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"type":"message","id":"msg_1","role":"assistant","content":[{"type":"text","text":"pong"}]}`))
	}))
	defer srv.Close()

	res, err := ProbeProviderConnectivity(context.Background(), &http.Client{Timeout: 2 * time.Second}, srv.URL, "test-key", "alibaba_anthropic", "anthropic", "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res == nil || !res.Success {
		t.Fatalf("probe=%+v", res)
	}
	if len(res.Models) == 0 {
		t.Fatal("expected fallback models")
	}
}
