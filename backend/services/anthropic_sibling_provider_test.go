package services

import "testing"

func TestAnthropicSiblingProviderCode(t *testing.T) {
	t.Parallel()
	if got := AnthropicSiblingProviderCode("alibaba"); got != "alibaba_anthropic" {
		t.Fatalf("got %q", got)
	}
	if got := AnthropicSiblingProviderCode("  OpenAI "); got != "openai_anthropic" {
		t.Fatalf("got %q", got)
	}
	if AnthropicSiblingProviderCode("") != "" {
		t.Fatal("expected empty")
	}
}

func TestProviderUsesAnthropicHTTP(t *testing.T) {
	t.Parallel()
	if !ProviderUsesAnthropicHTTP("alibaba_anthropic", "") {
		t.Fatal("suffix provider")
	}
	if !ProviderUsesAnthropicHTTP("alibaba", "anthropic") {
		t.Fatal("api_format anthropic")
	}
	if ProviderUsesAnthropicHTTP("alibaba", "openai") {
		t.Fatal("openai should not use anthropic auth")
	}
}
