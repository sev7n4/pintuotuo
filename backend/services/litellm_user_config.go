package services

import "strings"

// LitellmGatewayRequestModel is the top-level "model" sent to LiteLLM /v1/chat/completions.
// It must match gateway model_list wildcards (e.g. gemini/*, openrouter/*), not bare catalog ids.
func LitellmGatewayRequestModel(provider, catalogModel string) string {
	catalogModel = strings.TrimSpace(catalogModel)
	provider = strings.TrimSpace(provider)
	if catalogModel == "" {
		return ""
	}
	if strings.EqualFold(provider, modelProviderOpenRouter) {
		return formatLitellmModelForOpenRouter(catalogModel)
	}
	if litellmModel, err := ResolveLitellmModelFromCache(provider, catalogModel); err == nil && litellmModel != "" {
		return litellmModel
	}
	if canResolveLitellmModelWithoutCache(provider) {
		if resolved, resolveErr := resolveLitellmModelName(provider, catalogModel); resolveErr == nil && resolved != "" {
			return resolved
		}
	}
	return catalogModel
}

// BuildLitellmUserConfig builds LiteLLM clientside_auth user_config for BYOK traffic.
// catalogModel is the platform catalog model id (SKU / client model); litellm_params.model
// is resolved via model_providers.litellm_model_template (SSOT), same for domestic and overseas gateways.
func BuildLitellmUserConfig(provider, catalogModel, decryptedKey, upstreamBaseURL string) map[string]interface{} {
	catalogModel = strings.TrimSpace(catalogModel)
	provider = strings.TrimSpace(provider)

	litellmModel := LitellmGatewayRequestModel(provider, catalogModel)
	if litellmModel == "" {
		modelName := catalogModel
		if idx := strings.LastIndex(catalogModel, "/"); idx >= 0 {
			modelName = catalogModel[idx+1:]
		}
		litellmModel = modelName
	}

	params := map[string]interface{}{
		"model": litellmModel,
	}
	if decryptedKey != "" {
		params["api_key"] = decryptedKey
	}
	if strings.TrimSpace(upstreamBaseURL) != "" {
		params["api_base"] = strings.TrimRight(strings.TrimSpace(upstreamBaseURL), "/")
	}

	return map[string]interface{}{
		"model_list": []map[string]interface{}{
			{
				"model_name":     catalogModel,
				"litellm_params": params,
			},
		},
	}
}

func canResolveLitellmModelWithoutCache(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case modelProviderOpenAI, modelProviderAnthropic, modelProviderDeepseek,
		modelProviderAlibaba, modelProviderAlibabaAnthropic,
		modelProviderZhipu, modelProviderMoonshot, modelProviderMinimax,
		modelProviderGoogle, modelProviderStepfun, modelProviderOpenRouter:
		return true
	default:
		return false
	}
}
