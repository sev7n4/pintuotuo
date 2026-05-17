package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pintuotuo/backend/models"
)

func TestResolveProviderConnectivityBase_LitellmUsesGatewayNotDirect(t *testing.T) {
	t.Setenv("LITELLM_MASTER_KEY", "sk-litellm-master-test")

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v1/chat/completions" {
			if got := r.Header.Get("Authorization"); got != "Bearer sk-litellm-master-test" {
				t.Errorf("Authorization = %q, want LiteLLM master key", got)
			}
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"message":"invalid key"}}`))
			return
		}
		t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
	}))
	defer gateway.Close()

	t.Setenv("LLM_GATEWAY_LITELLM_URL_OVERSEAS", gateway.URL)
	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"google": {ProviderAPIBaseURL: "https://generativelanguage.googleapis.com/v1beta/openai"},
	})

	apiKey := &models.MerchantAPIKey{
		ID:        1,
		Provider:  "google",
		Region:    regionOverseas,
		RouteMode: GatewayModeLitellm,
		RouteConfig: map[string]interface{}{
			"endpoints": map[string]interface{}{
				"litellm": map[string]interface{}{
					"overseas": "http://litellm-overseas:4000/v1",
				},
			},
		},
	}

	hc := NewHealthChecker()
	base, token, mode, err := hc.ResolveProviderConnectivityBase(context.Background(), apiKey)
	if err != nil {
		t.Fatalf("ResolveProviderConnectivityBase: %v", err)
	}
	if mode != GatewayModeLitellm {
		t.Fatalf("mode = %q, want litellm", mode)
	}
	if token != "sk-litellm-master-test" {
		t.Fatalf("token = %q, want master key", token)
	}
	wantBase := gateway.URL + "/v1"
	if base != wantBase {
		t.Fatalf("base = %q, want %q", base, wantBase)
	}

	probe, err := hc.probeMerchantKeyConnectivity(context.Background(), apiKey, base, token, mode, "sk-byok-test")
	if err != nil {
		t.Fatalf("probeMerchantKeyConnectivity: %v", err)
	}
	if probe == nil || !probe.Success {
		t.Fatalf("probe = %+v, want success", probe)
	}
	if len(probe.Models) == 0 {
		t.Fatalf("models = %v, want catalog/predefined gemini ids", probe.Models)
	}
	for _, m := range probe.Models {
		if !strings.Contains(strings.ToLower(m), "gemini") {
			t.Fatalf("models = %v, want provider-filtered gemini only", probe.Models)
		}
	}
}

func TestResolveLitellmEndpoint_OverseasEnvFallback(t *testing.T) {
	t.Setenv("LLM_GATEWAY_LITELLM_URL", "http://domestic-litellm:4000")
	t.Setenv("LLM_GATEWAY_LITELLM_URL_OVERSEAS", "http://overseas-litellm:4000")

	hc := NewHealthChecker()
	ep, err := hc.resolveLitellmEndpoint(context.Background(), &models.MerchantAPIKey{Region: regionOverseas})
	if err != nil {
		t.Fatal(err)
	}
	if ep != "http://overseas-litellm:4000/v1" {
		t.Fatalf("endpoint = %q", ep)
	}
}
