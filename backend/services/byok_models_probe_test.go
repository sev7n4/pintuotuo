package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFilterBYOKModelsForProvider_DropsForeignAndNoise(t *testing.T) {
	in := []string{
		"*",
		"openai/*",
		"anthropic/claude-3.5-sonnet",
		"gemini-2.0-flash",
		"gemini/gemini-2.0-flash-exp",
		"gpt-4o-mini",
	}
	got := FilterBYOKModelsForProvider("google", in)
	if len(got) != 2 {
		t.Fatalf("google filtered = %v, want 2 gemini ids", got)
	}
}

func TestProbeLitellmBYOKModels_UpstreamGETWithBYOK(t *testing.T) {
	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"openrouter": {ProviderAPIBaseURL: ""},
	})

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("path %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer sk-byok") {
			t.Errorf("auth %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"anthropic/claude-3.5-sonnet"},{"id":"openai/gpt-4o-mini"}]}`))
	}))
	defer upstream.Close()

	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"openrouter": {Template: "openrouter/{model_id}", ProviderAPIBaseURL: upstream.URL},
	})

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected gateway call %s %s", r.Method, r.URL.Path)
	}))
	defer gateway.Close()

	probe, err := ProbeLitellmBYOKModels(
		context.Background(),
		&http.Client{Timeout: 5 * time.Second},
		gateway.URL+"/v1",
		"sk-master",
		"openrouter",
		"sk-byok",
		modelProviderOpenAI,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if probe == nil || !probe.Success {
		t.Fatalf("probe = %+v", probe)
	}
	if len(probe.Models) < 1 {
		t.Fatalf("models = %v", probe.Models)
	}
}

func TestProbeLitellmBYOKModels_ChatPathUsesCatalogNotPOSTModels(t *testing.T) {
	t.Setenv("LITELLM_MASTER_KEY", "sk-litellm-master-test")
	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"google": {
			Template:           "gemini/{model_id}",
			ProviderAPIBaseURL: "https://generativelanguage.googleapis.com/v1beta/openai",
		},
	})

	var postModels, postChat int
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/models":
			postModels++
			w.WriteHeader(http.StatusMethodNotAllowed)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/chat/completions":
			postChat++
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["user_config"] == nil {
				t.Error("expected user_config on chat")
			}
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"message":"invalid key"}}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer gateway.Close()

	probe, err := ProbeLitellmBYOKModels(
		context.Background(),
		&http.Client{Timeout: 5 * time.Second},
		gateway.URL+"/v1",
		"sk-litellm-master-test",
		"google",
		"sk-byok-google",
		modelProviderOpenAI,
		"",
	)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	if postModels > 0 {
		t.Fatal("should not POST /v1/models")
	}
	if postChat == 0 {
		t.Fatal("expected chat probe")
	}
	if probe == nil || !probe.Success {
		t.Fatalf("path reachable should succeed with catalog models: %+v", probe)
	}
	if len(probe.Models) == 0 {
		t.Fatal("expected catalog/predefined models after path OK")
	}
}

func TestResolveProbeUpstreamBaseURL_OpenrouterWithoutTemplate(t *testing.T) {
	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"openrouter": {ProviderAPIBaseURL: "https://openrouter.ai/api/v1"},
	})
	defer ResetLitellmCacheForTest()
	got := ResolveProbeUpstreamBaseURL("openrouter")
	if got != "https://openrouter.ai/api/v1" {
		t.Fatalf("got %q", got)
	}
}
