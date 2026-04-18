package main

import "testing"

func TestExtractModelNamesFromYAML(t *testing.T) {
	y := `
model_list:
  - model_name: gpt-4
    litellm_params:
      model: openai/gpt-4
  - model_name: openai/gpt-5
    litellm_params:
      model: openai/gpt-5
`
	n := extractModelNamesFromYAML(y)
	if _, ok := n["gpt-4"]; !ok {
		t.Fatal("expected gpt-4")
	}
	if _, ok := n["openai/gpt-5"]; !ok {
		t.Fatal("expected openai/gpt-5")
	}
}

func TestNameSetContainsCI(t *testing.T) {
	names := map[string]struct{}{"GPT-4": {}}
	if !nameSetContainsCI(names, "gpt-4") {
		t.Fatal("expected ci match")
	}
}

func TestVerifyFallbackModelNamesInList(t *testing.T) {
	yaml := `
model_list:
  - model_name: gpt-5-mini
    litellm_params:
      model: openai/gpt-5-mini
  - model_name: gpt-5-nano
    litellm_params:
      model: openai/gpt-5-nano
router_settings:
  routing_strategy: simple-shuffle
  fallbacks:
    - {"gpt-5-mini": ["gpt-5-nano"]}
litellm_settings:
  drop_params: true
`
	names := extractModelNamesFromYAML(yaml)
	miss := verifyFallbackModelNamesInList(yaml, names)
	if len(miss) != 0 {
		t.Fatalf("expected no missing, got %v", miss)
	}

	yamlBad := `
model_list:
  - model_name: gpt-5-mini
    litellm_params:
      model: openai/gpt-5-mini
router_settings:
  fallbacks:
    - {"gpt-5-mini": ["unknown-model"]}
`
	names2 := extractModelNamesFromYAML(yamlBad)
	miss2 := verifyFallbackModelNamesInList(yamlBad, names2)
	if len(miss2) != 1 || miss2[0] != "unknown-model" {
		t.Fatalf("expected missing unknown-model, got %v", miss2)
	}
}
