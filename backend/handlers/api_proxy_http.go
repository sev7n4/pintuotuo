package handlers

import (
	"database/sql"
	"encoding/json"
)

const (
	providerAnthropic   = "anthropic"
	apiFormatOpenAI     = "openai"
	llmGatewayLitellm   = "litellm"
	llmGatewayProxy     = "proxy"
	llmGatewayNone      = "none"
	policySourceEnv     = "env"
	policySourceDB      = "db"
	policySourceDefault = "default"
)

func getProviderRuntimeConfig(db *sql.DB, providerCode string) (providerRuntimeConfig, error) {
	var cfg providerRuntimeConfig
	var routeStrategyJSON, endpointsJSON []byte

	err := db.QueryRow(
		`SELECT code, name, COALESCE(api_base_url, ''), api_format,
		        COALESCE(provider_region, ''), 
		        route_strategy, endpoints
		 FROM model_providers
		 WHERE code = $1 AND status = 'active'
		 LIMIT 1`,
		providerCode,
	).Scan(&cfg.Code, &cfg.Name, &cfg.APIBaseURL, &cfg.APIFormat,
		&cfg.ProviderRegion, &routeStrategyJSON, &endpointsJSON)

	if err != nil {
		return cfg, err
	}

	if len(routeStrategyJSON) > 0 {
		if err := json.Unmarshal(routeStrategyJSON, &cfg.RouteStrategy); err != nil {
			cfg.RouteStrategy = nil
		}
	}

	if len(endpointsJSON) > 0 {
		if err := json.Unmarshal(endpointsJSON, &cfg.Endpoints); err != nil {
			cfg.Endpoints = nil
		}
	}

	return cfg, nil
}
