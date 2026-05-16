package services

import "testing"

func TestPickProbeModelFromCatalog(t *testing.T) {
	models := []string{
		"anthropic/claude-3-7-sonnet-20250219",
		"gpt-4o-mini",
		"gemini/gemini-2.0-flash",
	}

	if got := pickProbeModelFromCatalog("openai", models); got != "gpt-4o-mini" {
		t.Fatalf("openai: got %q", got)
	}
	if got := pickProbeModelFromCatalog("google", models); got != "gemini/gemini-2.0-flash" {
		t.Fatalf("google: got %q", got)
	}
	if got := pickProbeModelFromCatalog("anthropic", models); got != "anthropic/claude-3-7-sonnet-20250219" {
		t.Fatalf("anthropic: got %q", got)
	}
}

func TestSelectProbeModel_IgnoresForeignCatalogEntry(t *testing.T) {
	models := []string{"anthropic/claude-3-7-sonnet-20250219"}
	if got := selectProbeModel("openai", models, ""); got != "gpt-4o-mini" {
		t.Fatalf("got %q, want provider default gpt-4o-mini", got)
	}
}

func TestSelectQuotaProbeModel_LitellmIgnoresGatewayCatalog(t *testing.T) {
	gatewayModels := []string{"*", "openai/*", "anthropic/claude-3-7-sonnet-20250219", "openai/dall-e-2"}
	got := selectQuotaProbeModel("openai", gatewayModels, "", GatewayModeLitellm)
	if got != "gpt-4o" {
		t.Fatalf("openai litellm quota probe: got %q want gpt-4o from predefined SSOT", got)
	}
	got = selectQuotaProbeModel("google", gatewayModels, "", GatewayModeLitellm)
	if got != "gemini-2.0-flash" {
		t.Fatalf("google litellm quota probe: got %q", got)
	}
}

func TestBuildLitellmUserConfig_MatchesTemplateSSOT(t *testing.T) {
	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"google": {Template: "gemini/{model_id}", ProviderAPIBaseURL: "https://generativelanguage.googleapis.com/v1beta/openai/v1"},
	})
	uc := BuildLitellmUserConfig("google", "gemini-2.0-flash", "sk-byok", "https://generativelanguage.googleapis.com/v1beta/openai/v1")
	ml := uc["model_list"].([]map[string]interface{})[0]
	if ml["model_name"] != "gemini-2.0-flash" {
		t.Fatalf("model_name %v", ml["model_name"])
	}
	params := ml["litellm_params"].(map[string]interface{})
	if params["model"] != "gemini/gemini-2.0-flash" {
		t.Fatalf("litellm model %v", params["model"])
	}
}

func TestFormatLitellmModelForOpenRouter(t *testing.T) {
	if got := formatLitellmModelForOpenRouter("anthropic/claude-3.5-sonnet"); got != "openrouter/anthropic/claude-3.5-sonnet" {
		t.Fatalf("got %q", got)
	}
	if got := formatLitellmModelForOpenRouter("openrouter/google/gemini-2.0-flash"); got != "openrouter/google/gemini-2.0-flash" {
		t.Fatalf("got %q", got)
	}
}
