package services

import (
	"testing"
)

func TestResolveLitellmModelFromCache(t *testing.T) {
	tests := []struct {
		name     string
		cache    map[string]LitellmTemplateEntry
		provider string
		model    string
		want     string
		wantErr  bool
	}{
		{
			name: "zhipu with zai template",
			cache: map[string]LitellmTemplateEntry{
				"zhipu": {Template: "zai/{model_id}", APIKeyEnv: "ZAI_API_KEY"},
			},
			provider: "zhipu",
			model:    "glm-4.5",
			want:     "zai/glm-4.5",
			wantErr:  false,
		},
		{
			name: "alibaba with dashscope template",
			cache: map[string]LitellmTemplateEntry{
				"alibaba": {Template: "dashscope/{model_id}", APIKeyEnv: "DASHSCOPE_API_KEY"},
			},
			provider: "alibaba",
			model:    "qwen-turbo",
			want:     "dashscope/qwen-turbo",
			wantErr:  false,
		},
		{
			name: "google with gemini template",
			cache: map[string]LitellmTemplateEntry{
				"google": {Template: "gemini/{model_id}", APIKeyEnv: "GOOGLE_API_KEY"},
			},
			provider: "google",
			model:    "gemini-2.0-flash",
			want:     "gemini/gemini-2.0-flash",
			wantErr:  false,
		},
		{
			name: "stepfun with openai template and api_base",
			cache: map[string]LitellmTemplateEntry{
				"stepfun": {Template: "openai/{model_id}", APIKeyEnv: "STEPFUN_API_KEY", APIBase: "https://api.stepfun.com/v1"},
			},
			provider: "stepfun",
			model:    "step-1v-8k",
			want:     "openai/step-1v-8k",
			wantErr:  false,
		},
		{
			name: "model with existing slash strips prefix",
			cache: map[string]LitellmTemplateEntry{
				"zhipu": {Template: "zai/{model_id}", APIKeyEnv: "ZAI_API_KEY"},
			},
			provider: "zhipu",
			model:    "openai/glm-4",
			want:     "zai/glm-4",
			wantErr:  false,
		},
		{
			name: "template without placeholder appends slash",
			cache: map[string]LitellmTemplateEntry{
				"custom": {Template: "custom_prefix", APIKeyEnv: "CUSTOM_API_KEY"},
			},
			provider: "custom",
			model:    "model-a",
			want:     "custom_prefix/model-a",
			wantErr:  false,
		},
		{
			name:     "provider not in cache",
			cache:    map[string]LitellmTemplateEntry{},
			provider: "unknown",
			model:    "model-x",
			want:     "",
			wantErr:  true,
		},
		{
			name: "empty template in cache skipped",
			cache: map[string]LitellmTemplateEntry{
				"empty": {Template: "", APIKeyEnv: "EMPTY_API_KEY"},
			},
			provider: "empty",
			model:    "model-x",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLitellmCacheForTest(tt.cache)
			defer ResetLitellmCacheForTest()

			got, err := ResolveLitellmModelFromCache(tt.provider, tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveLitellmModelFromCache(%q, %q) error = %v, wantErr %v", tt.provider, tt.model, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveLitellmModelFromCache(%q, %q) = %q, want %q", tt.provider, tt.model, got, tt.want)
			}
		})
	}
}

func TestResolveLitellmModelFromCacheNotLoaded(t *testing.T) {
	ResetLitellmCacheForTest()
	_, err := ResolveLitellmModelFromCache("zhipu", "glm-4")
	if err == nil {
		t.Error("expected error when cache not loaded, got nil")
	}
}

func TestGetLitellmTemplateCache(t *testing.T) {
	cache := map[string]LitellmTemplateEntry{
		"zhipu":   {Template: "zai/{model_id}", APIKeyEnv: "ZAI_API_KEY"},
		"alibaba": {Template: "dashscope/{model_id}", APIKeyEnv: "DASHSCOPE_API_KEY"},
	}
	SetLitellmCacheForTest(cache)
	defer ResetLitellmCacheForTest()

	result := GetLitellmTemplateCache()
	if len(result) != 2 {
		t.Errorf("GetLitellmTemplateCache() returned %d entries, want 2", len(result))
	}
	if result["zhipu"].Template != "zai/{model_id}" {
		t.Errorf("GetLitellmTemplateCache()[\"zhipu\"].Template = %q, want %q", result["zhipu"].Template, "zai/{model_id}")
	}
}

func TestResolveLitellmModelFromCacheFallback(t *testing.T) {
	cache := map[string]LitellmTemplateEntry{
		"zhipu": {Template: "zai/{model_id}", APIKeyEnv: "ZAI_API_KEY"},
	}
	SetLitellmCacheForTest(cache)
	defer ResetLitellmCacheForTest()

	got, err := ResolveLitellmModelFromCache("zhipu", "glm-4.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "zai/glm-4.5" {
		t.Errorf("ResolveLitellmModelFromCache(\"zhipu\", \"glm-4.5\") = %q, want %q", got, "zai/glm-4.5")
	}

	fallbackGot, fallbackErr := resolveLitellmModelName("zhipu", "glm-4.5")
	if fallbackErr != nil {
		t.Fatalf("unexpected fallback error: %v", fallbackErr)
	}
	if fallbackGot != got {
		t.Errorf("cache result %q != fallback result %q, should be consistent", got, fallbackGot)
	}
}

func TestResolveLitellmUpstreamBaseURL(t *testing.T) {
	cache := map[string]LitellmTemplateEntry{
		"stepfun":  {Template: "openai/{model_id}", APIBase: "https://api.stepfun.com/v1"},
		"deepseek": {Template: "openai/{model_id}", APIBase: "https://api.deepseek.com"},
		"zhipu":    {Template: "zai/{model_id}", APIBase: ""},
	}
	SetLitellmCacheForTest(cache)
	defer ResetLitellmCacheForTest()

	t.Run("returns api_base for known provider", func(t *testing.T) {
		got := ResolveLitellmUpstreamBaseURL("stepfun")
		if got != "https://api.stepfun.com/v1" {
			t.Errorf("ResolveLitellmUpstreamBaseURL(\"stepfun\") = %q, want %q", got, "https://api.stepfun.com/v1")
		}
	})

	t.Run("returns api_base for another provider", func(t *testing.T) {
		got := ResolveLitellmUpstreamBaseURL("deepseek")
		if got != "https://api.deepseek.com" {
			t.Errorf("ResolveLitellmUpstreamBaseURL(\"deepseek\") = %q, want %q", got, "https://api.deepseek.com")
		}
	})

	t.Run("returns empty for provider with empty api_base", func(t *testing.T) {
		got := ResolveLitellmUpstreamBaseURL("zhipu")
		if got != "" {
			t.Errorf("ResolveLitellmUpstreamBaseURL(\"zhipu\") = %q, want empty", got)
		}
	})

	t.Run("returns empty for unknown provider", func(t *testing.T) {
		got := ResolveLitellmUpstreamBaseURL("unknown")
		if got != "" {
			t.Errorf("ResolveLitellmUpstreamBaseURL(\"unknown\") = %q, want empty", got)
		}
	})

	t.Run("prefers ProviderAPIBaseURL over gateway APIBase", func(t *testing.T) {
		cache := map[string]LitellmTemplateEntry{
			"moonshot": {
				Template:           "moonshot/{model_id}",
				APIBase:            "https://override-gateway.example/v1",
				ProviderAPIBaseURL: "https://api.moonshot.cn/v1",
			},
		}
		SetLitellmCacheForTest(cache)
		defer ResetLitellmCacheForTest()
		got := ResolveLitellmUpstreamBaseURL("moonshot")
		if got != "https://api.moonshot.cn/v1" {
			t.Errorf("ResolveLitellmUpstreamBaseURL(\"moonshot\") = %q, want provider api_base_url", got)
		}
	})
}
