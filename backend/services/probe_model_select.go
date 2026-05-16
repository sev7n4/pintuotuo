package services

import "strings"

// pickProbeModelFromCatalog selects a model id from a LiteLLM /v1/models list for the given BYOK provider.
// Gateway model lists are often cross-vendor; avoid using models[0] when it belongs to another provider.
func pickProbeModelFromCatalog(provider string, models []string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	var matched []string
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		if modelMatchesProviderProbe(provider, strings.ToLower(m)) {
			matched = append(matched, m)
		}
	}
	if len(matched) == 0 {
		return ""
	}
	return matched[0]
}

func modelMatchesProviderProbe(provider, modelLower string) bool {
	switch provider {
	case modelProviderOpenAI, modelProviderStepfun:
		return strings.HasPrefix(modelLower, "gpt-") ||
			strings.HasPrefix(modelLower, "o1") ||
			strings.HasPrefix(modelLower, "o3") ||
			strings.HasPrefix(modelLower, "openai/") ||
			strings.Contains(modelLower, "/gpt-")
	case modelProviderGoogle:
		return strings.HasPrefix(modelLower, "gemini") ||
			strings.HasPrefix(modelLower, "gemini/") ||
			strings.HasPrefix(modelLower, "google/")
	case modelProviderAnthropic, "alibaba_anthropic":
		return strings.Contains(modelLower, "claude") ||
			strings.HasPrefix(modelLower, "anthropic/")
	case "openrouter":
		if strings.HasPrefix(modelLower, "openrouter/") {
			return true
		}
		// OpenRouter catalog entries are often vendor/model without the openrouter/ prefix.
		return strings.HasPrefix(modelLower, "anthropic/") ||
			strings.HasPrefix(modelLower, "openai/") ||
			strings.HasPrefix(modelLower, "google/") ||
			strings.HasPrefix(modelLower, "meta-llama/") ||
			strings.HasPrefix(modelLower, "deepseek/")
	case modelProviderDeepseek:
		return strings.Contains(modelLower, "deepseek")
	case modelProviderZhipu:
		return strings.Contains(modelLower, "glm") || strings.HasPrefix(modelLower, "zai/")
	case modelProviderAlibaba:
		return strings.Contains(modelLower, "qwen") || strings.HasPrefix(modelLower, "dashscope/")
	case modelProviderMoonshot:
		return strings.Contains(modelLower, "moonshot") || strings.HasPrefix(modelLower, "kimi")
	case modelProviderMinimax:
		return strings.Contains(modelLower, "minimax")
	default:
		if provider == "" {
			return false
		}
		return strings.HasPrefix(modelLower, provider+"/")
	}
}

// formatLitellmModelForOpenRouter normalizes catalog model ids for LiteLLM openrouter/* routes.
func formatLitellmModelForOpenRouter(model string) string {
	m := strings.TrimSpace(model)
	if m == "" {
		return m
	}
	if strings.HasPrefix(strings.ToLower(m), "openrouter/") {
		return m
	}
	return "openrouter/" + strings.TrimPrefix(m, "/")
}
