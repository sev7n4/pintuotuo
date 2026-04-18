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
