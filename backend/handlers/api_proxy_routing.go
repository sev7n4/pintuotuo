package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/pintuotuo/backend/services"
)

type MerchantRouteInfo struct {
	Type   string
	Region string
}

//nolint:unused // Will be used in Phase 3
func getMerchantRouteInfo(db *sql.DB, merchantID int) (*MerchantRouteInfo, error) {
	if db == nil {
		return &MerchantRouteInfo{Type: "regular", Region: "domestic"}, nil
	}

	var merchantType, region string
	err := db.QueryRow(
		`SELECT COALESCE(merchant_type, 'regular'), COALESCE(region, 'domestic')
		 FROM merchants
		 WHERE id = $1
		 LIMIT 1`,
		merchantID,
	).Scan(&merchantType, &region)

	if err == sql.ErrNoRows {
		return &MerchantRouteInfo{Type: "regular", Region: "domestic"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query merchant info: %w", err)
	}

	return &MerchantRouteInfo{Type: merchantType, Region: region}, nil
}

//nolint:unused // Will be used in Phase 3
func resolveRouteDecision(
	db *sql.DB,
	providerCfg *providerRuntimeConfig,
	merchantID int,
) (*services.RouteDecision, error) {
	merchantInfo, err := getMerchantRouteInfo(db, merchantID)
	if err != nil {
		return nil, err
	}

	if len(providerCfg.RouteStrategy) > 0 {
		providerConfig := &services.ProviderConfig{
			Code:           providerCfg.Code,
			ProviderRegion: providerCfg.ProviderRegion,
			RouteStrategy:  providerCfg.RouteStrategy,
			Endpoints:      providerCfg.Endpoints,
		}

		merchantConfig := &services.MerchantConfig{
			ID:     merchantID,
			Type:   merchantInfo.Type,
			Region: merchantInfo.Region,
		}

		router := services.NewUnifiedRouter(nil)
		decision, err := router.DecideRoute(context.TODO(), providerConfig, merchantConfig)
		if err != nil {
			return fallbackToEnvDecision(providerCfg), nil
		}

		return decision, nil
	}

	return fallbackToEnvDecision(providerCfg), nil
}

//nolint:unused // Will be used in Phase 3
func fallbackToEnvDecision(cfg *providerRuntimeConfig) *services.RouteDecision {
	mode := determineGatewayModeFromEnv()
	endpoint := resolveEndpointFromEnv(mode, cfg)

	return &services.RouteDecision{
		Mode:     mode,
		Endpoint: endpoint,
		Reason:   "fallback to environment variable",
	}
}

//nolint:unused // Will be used in Phase 3
func determineGatewayModeFromEnv() string {
	envMode := os.Getenv("LLM_GATEWAY_ACTIVE")
	if envMode != "" && envMode != llmGatewayNone {
		return envMode
	}
	return services.GatewayModeDirect
}

//nolint:unused // Will be used in Phase 3
func resolveEndpointFromEnv(mode string, cfg *providerRuntimeConfig) string {
	switch mode {
	case services.GatewayModeLitellm:
		litellmURL := os.Getenv("LLM_GATEWAY_LITELLM_URL")
		if litellmURL != "" {
			return litellmURL + "/v1"
		}
		if cfg != nil {
			return cfg.APIBaseURL
		}
		return ""

	case services.GatewayModeProxy:
		proxyURL := os.Getenv("LLM_GATEWAY_PROXY_URL")
		if proxyURL != "" {
			return proxyURL
		}
		if cfg != nil {
			return cfg.APIBaseURL
		}
		return ""

	default:
		if cfg != nil {
			return cfg.APIBaseURL
		}
		return ""
	}
}
