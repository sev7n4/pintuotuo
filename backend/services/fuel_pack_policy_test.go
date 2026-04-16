package services

import "testing"

func TestIsStrictModelSKU(t *testing.T) {
	cases := []struct {
		name           string
		provider       string
		modelName      string
		providerModel  string
		expectEligible bool
	}{
		{name: "normal provider model_name", provider: "openai", modelName: "gpt-4o", expectEligible: true},
		{name: "normal provider provider_model_id", provider: "anthropic", providerModel: "claude-3-7-sonnet", expectEligible: true},
		{name: "internal provider", provider: "internal", modelName: "virtual", expectEligible: false},
		{name: "virtual goods provider", provider: "virtual_goods", modelName: "fuel", expectEligible: false},
		{name: "empty model", provider: "openai", expectEligible: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := IsStrictModelSKU(c.provider, c.modelName, c.providerModel)
			if got != c.expectEligible {
				t.Fatalf("expect=%v got=%v", c.expectEligible, got)
			}
		})
	}
}

func TestValidateFuelPackBundle(t *testing.T) {
	err := ValidateFuelPackBundle([]OrderLinePolicyInput{
		{SKUType: "token_pack", ModelProvider: "internal", ModelName: "fuel"},
	})
	if err == nil {
		t.Fatalf("expected restriction error for token_pack-only bundle")
	}

	err = ValidateFuelPackBundle([]OrderLinePolicyInput{
		{SKUType: "token_pack", ModelProvider: "internal", ModelName: "fuel"},
		{SKUType: "subscription", ModelProvider: "openai", ModelName: "gpt-4o"},
	})
	if err != nil {
		t.Fatalf("expected mixed bundle to pass, got error=%v", err)
	}
}
