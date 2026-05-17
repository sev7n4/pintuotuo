package services

import (
	"context"
	"encoding/json"
	"io"
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
	if got[0] != "gemini-2.0-flash" || got[1] != "gemini/gemini-2.0-flash-exp" {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestProbeLitellmBYOKModels_PostUserConfigNotMasterGET(t *testing.T) {
	t.Setenv("LITELLM_MASTER_KEY", "sk-litellm-master-test")
	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"google": {
			Template:           "gemini/{model_id}",
			ProviderAPIBaseURL: "https://generativelanguage.googleapis.com/v1beta/openai",
		},
	})

	var postModels, getModels int
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer sk-litellm-master-test" {
			t.Errorf("Authorization = %q, want master key", auth)
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/models":
			postModels++
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("body: %v", err)
			}
			if payload["user_config"] == nil {
				t.Fatal("expected user_config in POST /v1/models")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"id":"gemini-2.0-flash"},{"id":"gpt-4o-mini"}]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/models":
			getModels++
			w.WriteHeader(http.StatusForbidden)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer gateway.Close()

	t.Setenv("LLM_GATEWAY_LITELLM_URL_OVERSEAS", gateway.URL)

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
		t.Fatalf("ProbeLitellmBYOKModels: %v", err)
	}
	if probe == nil || !probe.Success {
		t.Fatalf("probe = %+v", probe)
	}
	if postModels == 0 {
		t.Fatal("expected POST /v1/models with user_config")
	}
	if getModels > 0 {
		t.Fatal("should not use master-key GET /v1/models when POST user_config succeeds")
	}
	if len(probe.Models) != 1 || probe.Models[0] != "gemini-2.0-flash" {
		t.Fatalf("models = %v, want provider-filtered gemini only", probe.Models)
	}
}

func TestProbeLitellmBYOKModels_FallbackUpstreamGETWithBYOK(t *testing.T) {
	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"openai": {ProviderAPIBaseURL: ""},
	})

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("path %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer sk-byok-openai") {
			t.Errorf("auth %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o-mini"}]}`))
	}))
	defer upstream.Close()

	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v1/models" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer gateway.Close()

	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"openai": {ProviderAPIBaseURL: upstream.URL},
	})

	probe, err := ProbeLitellmBYOKModels(
		context.Background(),
		&http.Client{Timeout: 5 * time.Second},
		gateway.URL+"/v1",
		"sk-master",
		"openai",
		"sk-byok-openai",
		modelProviderOpenAI,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if probe == nil || !probe.Success || len(probe.Models) != 1 || probe.Models[0] != "gpt-4o-mini" {
		t.Fatalf("probe = %+v", probe)
	}
}
