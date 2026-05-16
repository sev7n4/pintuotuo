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
	case modelProviderAnthropic, modelProviderAlibabaAnthropic:
		return strings.Contains(modelLower, "claude") ||
			strings.HasPrefix(modelLower, "anthropic/")
	case modelProviderOpenRouter:
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

// selectQuotaProbeModel picks the catalog model id used for deep quota chat probes.
// Under litellm, ignore LiteLLM gateway GET /v1/models listings (wildcards / cross-vendor noise);
// use platform catalog ids (models_supported, predefinedModels) then ResolveLitellmModelFromCache — same SSOT as api_proxy.
func selectQuotaProbeModel(provider string, models []string, userSelected, routeMode string) string {
	if strings.TrimSpace(userSelected) != "" {
		return strings.TrimSpace(userSelected)
	}
	if routeMode == GatewayModeLitellm {
		if picked := pickCatalogProbeModel(provider, models); picked != "" {
			return picked
		}
		return defaultCatalogProbeModel(provider)
	}
	return selectProbeModel(provider, models, "")
}

func pickCatalogProbeModel(provider string, models []string) string {
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m == "" || isGatewayCatalogNoise(m) {
			continue
		}
		if modelMatchesProviderProbe(provider, strings.ToLower(m)) {
			return m
		}
	}
	return ""
}

func isGatewayCatalogNoise(model string) bool {
	ml := strings.ToLower(strings.TrimSpace(model))
	if ml == "" || ml == "*" {
		return true
	}
	if strings.HasSuffix(ml, "/*") {
		return true
	}
	excluded := []string{
		"dall-e", "gpt-image", "/image", "sora", "transcribe", "whisper",
		"embedding", "moderation", "realtime", "codex-mini", "deep-research",
	}
	for _, s := range excluded {
		if strings.Contains(ml, s) {
			return true
		}
	}
	return false
}

func defaultCatalogProbeModel(provider string) string {
	if models := GetPredefinedModels(provider); len(models) > 0 {
		return models[0]
	}
	return selectProbeModel(provider, nil, "")
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
