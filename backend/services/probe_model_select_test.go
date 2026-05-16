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

func TestFormatLitellmModelForOpenRouter(t *testing.T) {
	if got := formatLitellmModelForOpenRouter("anthropic/claude-3.5-sonnet"); got != "openrouter/anthropic/claude-3.5-sonnet" {
		t.Fatalf("got %q", got)
	}
	if got := formatLitellmModelForOpenRouter("openrouter/google/gemini-2.0-flash"); got != "openrouter/google/gemini-2.0-flash" {
		t.Fatalf("got %q", got)
	}
}
