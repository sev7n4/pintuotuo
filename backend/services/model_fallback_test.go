package services

import "testing"

func TestDedupeFallbackChainUnique(t *testing.T) {
	in := []string{"a/b", " c/d ", "a/b", "e/f"}
	out := DedupeFallbackChainUnique(in)
	if len(out) != 3 || out[0] != "a/b" || out[1] != "c/d" || out[2] != "e/f" {
		t.Fatalf("got %#v", out)
	}
}

func TestSplitCatalogModelKey(t *testing.T) {
	p, m, err := SplitCatalogModelKey(" OpenAI/gpt-4o-mini ")
	if err != nil || p != "openai" || m != "gpt-4o-mini" {
		t.Fatalf("got %q %q err=%v", p, m, err)
	}
	if _, _, err := SplitCatalogModelKey("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFallbackGraphHasCycle(t *testing.T) {
	if !FallbackGraphHasCycle(map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}) {
		t.Fatal("expected cycle a<->b")
	}
	if FallbackGraphHasCycle(map[string][]string{
		"a": {"b"},
		"b": {"c"},
	}) {
		t.Fatal("unexpected cycle")
	}
	if !FallbackGraphHasCycle(map[string][]string{
		"a": {"a"},
	}) {
		t.Fatal("expected self-loop cycle")
	}
}
