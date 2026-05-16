package services

import "strings"

// BuildLitellmUserConfig builds LiteLLM clientside_auth user_config for BYOK traffic.
// catalogModel is the platform catalog model id (SKU / client model); litellm_params.model
// is resolved via model_providers.litellm_model_template (SSOT), same for domestic and overseas gateways.
func BuildLitellmUserConfig(provider, catalogModel, decryptedKey, upstreamBaseURL string) map[string]interface{} {
	catalogModel = strings.TrimSpace(catalogModel)
	provider = strings.TrimSpace(provider)

	litellmModel, err := ResolveLitellmModelFromCache(provider, catalogModel)
	if err != nil || litellmModel == "" {
		if strings.EqualFold(provider, modelProviderOpenRouter) {
			litellmModel = formatLitellmModelForOpenRouter(catalogModel)
		} else {
			modelName := catalogModel
			if idx := strings.LastIndex(catalogModel, "/"); idx >= 0 {
				modelName = catalogModel[idx+1:]
			}
			litellmModel = modelName
		}
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
