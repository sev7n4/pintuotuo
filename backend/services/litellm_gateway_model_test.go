package services

import "testing"

func TestLitellmGatewayRequestModel_FromTemplate(t *testing.T) {
	SetLitellmCacheForTest(map[string]LitellmTemplateEntry{
		"google": {Template: "gemini/{model_id}"},
		"openrouter": {Template: "openrouter/{model_id}"},
	})
	defer ResetLitellmCacheForTest()

	if got := LitellmGatewayRequestModel("google", "gemini-2.0-flash"); got != "gemini/gemini-2.0-flash" {
		t.Fatalf("google: got %q", got)
	}
	if got := LitellmGatewayRequestModel("openrouter", "anthropic/claude-3.5-sonnet"); got != "openrouter/anthropic/claude-3.5-sonnet" {
		t.Fatalf("openrouter: got %q", got)
	}
}

func TestLitellmGatewayRequestModel_AnthropicPrefix(t *testing.T) {
	ResetLitellmCacheForTest()
	got := LitellmGatewayRequestModel("anthropic", "claude-3-5-sonnet-20241022")
	if got != "anthropic/claude-3-5-sonnet-20241022" {
		t.Fatalf("got %q", got)
	}
}
