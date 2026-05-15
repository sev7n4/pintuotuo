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
