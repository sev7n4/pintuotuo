package services

import (
	"context"
	"database/sql"
	"fmt"
)

type UnifiedRouter struct {
	db    *sql.DB
	cache map[string]*RouteDecision
}

type RouteDecision struct {
	Mode             string `json:"mode"`
	Endpoint         string `json:"endpoint"`
	FallbackMode     string `json:"fallback_mode"`
	FallbackEndpoint string `json:"fallback_endpoint"`
	Reason           string `json:"reason"`
}

type ProviderConfig struct {
	Code           string
	Name           string
	ProviderRegion string
	RouteStrategy  map[string]interface{}
	Endpoints      map[string]interface{}
}

type MerchantConfig struct {
	ID              int
	Type            string
	Region          string
	RoutePreference map[string]interface{}
}

func NewUnifiedRouter(db *sql.DB) *UnifiedRouter {
	return &UnifiedRouter{
		db:    db,
		cache: make(map[string]*RouteDecision),
	}
}

func (r *UnifiedRouter) DecideRoute(
	ctx context.Context,
	providerConfig *ProviderConfig,
	merchantConfig *MerchantConfig,
) (*RouteDecision, error) {
	decision := &RouteDecision{
		Mode:   "direct",
		Reason: "default",
	}

	userKey := fmt.Sprintf("%s_users", merchantConfig.Region)
	if merchantConfig.Type == "enterprise" {
		userKey = "enterprise_users"
	}

	strategy, ok := providerConfig.RouteStrategy[userKey].(map[string]interface{})
	if !ok {
		if defaultMode, ok := providerConfig.RouteStrategy["default_mode"].(string); ok {
			strategy = map[string]interface{}{"mode": defaultMode}
		} else {
			strategy = map[string]interface{}{"mode": "auto"}
		}
	}

	mode, _ := strategy["mode"].(string)
	if mode == "auto" {
		if providerConfig.ProviderRegion == "overseas" && merchantConfig.Region == "domestic" {
			mode = "litellm"
			decision.Reason = "auto: domestic user accessing overseas provider"
		} else {
			mode = "direct"
			decision.Reason = "auto: direct connection"
		}
	}

	decision.Mode = mode

	decision.Endpoint = r.SelectEndpoint(mode, merchantConfig.Region, providerConfig.Endpoints)

	if fallbackMode, ok := strategy["fallback_mode"].(string); ok {
		decision.FallbackMode = fallbackMode
		decision.FallbackEndpoint = r.SelectEndpoint(fallbackMode, merchantConfig.Region, providerConfig.Endpoints)
	}

	return decision, nil
}

func (r *UnifiedRouter) SelectEndpoint(
	mode string,
	region string,
	endpoints map[string]interface{},
) string {
	modeEndpoints, ok := endpoints[mode].(map[string]interface{})
	if !ok {
		return ""
	}

	endpoint, _ := modeEndpoints[region].(string)
	return endpoint
}
