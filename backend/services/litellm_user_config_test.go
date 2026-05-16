package services

import "testing"

func TestBuildLitellmUserConfig_OpenRouter(t *testing.T) {
	uc := BuildLitellmUserConfig("openrouter", "anthropic/claude-3.5-haiku", "sk-or", "https://openrouter.ai/api/v1")
	ml := uc["model_list"].([]map[string]interface{})[0]
	if ml["model_name"] != "anthropic/claude-3.5-haiku" {
		t.Fatalf("model_name %v", ml["model_name"])
	}
	params := ml["litellm_params"].(map[string]interface{})
	if params["model"] != "openrouter/anthropic/claude-3.5-haiku" {
		t.Fatalf("litellm model %v", params["model"])
	}
}
