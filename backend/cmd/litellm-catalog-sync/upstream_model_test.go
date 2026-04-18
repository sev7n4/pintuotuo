package main

import "testing"

func TestProviderMapEntry_upstreamLitellmModel(t *testing.T) {
	tests := []struct {
		name   string
		ent    providerMapEntry
		model  string
		expect string
	}{
		{
			name: "template zai",
			ent: providerMapEntry{
				LitellmModelTemplate: "zai/{model_id}",
				APIKeyEnv:            "ZAI_API_KEY",
			},
			model:  "glm-5",
			expect: "zai/glm-5",
		},
		{
			name: "template openai stepfun",
			ent: providerMapEntry{
				LitellmModelTemplate: "openai/{model_id}",
				APIBase:              "https://api.stepfun.com/v1",
				APIKeyEnv:            "STEPFUN_API_KEY",
			},
			model:  "step-1-8k",
			expect: "openai/step-1-8k",
		},
		{
			name: "legacy prefix",
			ent: providerMapEntry{
				LitellmPrefix: "deepseek",
				APIKeyEnv:     "DEEPSEEK_API_KEY",
			},
			model:  "deepseek-chat",
			expect: "deepseek/deepseek-chat",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ent.upstreamLitellmModel(tt.model); got != tt.expect {
				t.Fatalf("upstreamLitellmModel(%q) = %q, want %q", tt.model, got, tt.expect)
			}
		})
	}
}
