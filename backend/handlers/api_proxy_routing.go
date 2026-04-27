package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
)

type smartRoutingPick struct {
	APIKeyID        *int
	CandidatesJSON  []byte
	RoutingDecision *services.RoutingDecision
}

type strategyRuntimeSnapshot struct {
	StrategyCode      string `json:"strategy_code"`
	MaxRetries        int    `json:"max_retries"`
	InitialDelayMs    int    `json:"initial_delay_ms"`
	CircuitThreshold  int    `json:"circuit_breaker_threshold"`
	CircuitTimeoutSec int    `json:"circuit_breaker_timeout_sec"`
}

type routingDecisionPayload struct {
	Candidates            json.RawMessage         `json:"candidates"`
	StrategyRuntime       strategyRuntimeSnapshot `json:"strategy_runtime"`
	EffectivePolicySource string                  `json:"effective_policy_source,omitempty"`
}

type traceTopCandidate struct {
	APIKeyID int     `json:"api_key_id"`
	Provider string  `json:"provider"`
	Model    string  `json:"model,omitempty"`
	Score    float64 `json:"score"`
}

func entitlementKeyFilterForRouter(strict bool, ent *services.EntitlementRoutingContext) []int {
	if !strict {
		return nil
	}
	if ent == nil || len(ent.AllowedAPIKeyIDs) == 0 {
		return []int{}
	}
	out := make([]int, 0, len(ent.AllowedAPIKeyIDs))
	for id := range ent.AllowedAPIKeyIDs {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}

func pickDeterministicEntitledKey(ent *services.EntitlementRoutingContext) (apiKeyID int, merchantSKUID int) {
	if ent == nil || len(ent.AllowedAPIKeyIDs) == 0 {
		return 0, 0
	}
	minK := 0
	for k := range ent.AllowedAPIKeyIDs {
		if minK == 0 || k < minK {
			minK = k
		}
	}
	msid, _ := ent.MerchantSKUForAPIKey(minK)
	return minK, msid
}

func scanMerchantAPIKeyQuotaRow(row *sql.Row, apiKey *models.MerchantAPIKey) error {
	var qLim sql.NullFloat64
	if err := row.Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.APISecretEncrypted, &qLim, &apiKey.QuotaUsed, &apiKey.Status); err != nil {
		return err
	}
	apiKey.QuotaLimit = utils.NullFloat64Ptr(qLim)
	return nil
}

func resolveMerchantIDByUser(db *sql.DB, userID int) (int, error) {
	var merchantID int
	err := db.QueryRow("SELECT id FROM merchants WHERE user_id = $1 AND "+sqlMerchantOperational+" LIMIT 1", userID).Scan(&merchantID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return merchantID, nil
}

func trySelectAPIKeyWithSmartRouter(req APIProxyRequest, strategyCode string, keyFilter []int, requestID string, merchantID int) smartRoutingPick {
	if strings.TrimSpace(req.Provider) == "" {
		return smartRoutingPick{}
	}
	if keyFilter != nil && len(keyFilter) == 0 {
		return smartRoutingPick{}
	}

	pipeline := services.NewThreeLayerRoutingPipeline()
	routingReq := &services.RoutingRequest{
		RequestID:     requestID,
		MerchantID:    merchantID,
		Model:         req.Model,
		Provider:      req.Provider,
		AllowedKeyIDs: keyFilter,
		RequestBody:   map[string]interface{}{"messages": req.Messages},
	}

	decision, err := pipeline.Execute(context.Background(), routingReq)
	if err != nil || decision == nil || decision.SelectedAPIKeyID == 0 {
		return smartRoutingPick{}
	}

	var candidatesJSON []byte
	if len(decision.DecisionLayerCandidates) > 0 {
		candidatesJSON, _ = json.Marshal(map[string]interface{}{
			"candidates": decision.DecisionLayerCandidates,
		})
	}

	return smartRoutingPick{
		APIKeyID:        &decision.SelectedAPIKeyID,
		CandidatesJSON:  candidatesJSON,
		RoutingDecision: decision,
	}
}

func routingStrategyWithSource() (code string, source string) {
	code = strings.TrimSpace(os.Getenv("SMART_ROUTING_STRATEGY"))
	if code != "" {
		return code, policySourceEnv
	}
	code = strings.TrimSpace(services.GetSmartRouter().GetDefaultStrategyCode())
	if code != "" {
		return code, policySourceDB
	}
	return string(services.RoutingStrategyBalanced), policySourceDefault
}

func shouldUseSmartRouting(userID int, requestID string) bool {
	enabled := strings.TrimSpace(strings.ToLower(os.Getenv("SMART_ROUTING_ENABLE")))
	if enabled == "false" || enabled == "0" || enabled == "off" {
		return false
	}
	percent := 100
	if raw := strings.TrimSpace(os.Getenv("SMART_ROUTING_GRAY_PERCENT")); raw != "" {
		if p, err := strconv.Atoi(raw); err == nil {
			if p < 0 {
				p = 0
			}
			if p > 100 {
				p = 100
			}
			percent = p
		}
	}
	if percent == 0 {
		return false
	}
	if percent == 100 {
		return true
	}
	seed := userID*31 + len(requestID)*17
	for _, ch := range requestID {
		seed += int(ch)
	}
	slot := seed % 100
	if slot < 0 {
		slot = -slot
	}
	return slot < percent
}

func selectAPIKeyForRequest(db *sql.DB, userID, merchantID int, req APIProxyRequest, apiKey *models.MerchantAPIKey, ent *services.EntitlementRoutingContext) error {
	if req.APIKeyID != nil && *req.APIKeyID > 0 {
		keyPick := `SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status
			 FROM merchant_api_keys mak
			 INNER JOIN merchants m ON m.id = mak.merchant_id
			 WHERE mak.id = $1 AND mak.provider = $2 AND mak.status = 'active'
			   AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')
			   AND (mak.quota_limit IS NULL OR mak.quota_used < mak.quota_limit)
			   AND m.status IN ('active', 'approved')
			   AND m.lifecycle_status <> 'suspended'`
		if merchantID <= 0 {
			if ent != nil && ent.AllowsAPIKey(*req.APIKeyID) {
				return scanMerchantAPIKeyQuotaRow(
					db.QueryRow(keyPick+` LIMIT 1`, *req.APIKeyID, req.Provider),
					apiKey,
				)
			}
			keyPick += ` AND m.user_id = $3`
			return scanMerchantAPIKeyQuotaRow(
				db.QueryRow(keyPick+` LIMIT 1`, *req.APIKeyID, req.Provider, userID),
				apiKey,
			)
		}
		keyPick += ` AND mak.merchant_id = $3 LIMIT 1`
		err := scanMerchantAPIKeyQuotaRow(
			db.QueryRow(keyPick, *req.APIKeyID, req.Provider, merchantID),
			apiKey,
		)
		if err == nil {
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
	}

	if req.MerchantSKUID != nil && *req.MerchantSKUID > 0 {
		if merchantID <= 0 {
			if ent != nil && ent.AllowsMerchantSKU(*req.MerchantSKUID) {
				err := scanMerchantAPIKeyQuotaRow(
					db.QueryRow(
						`SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status
						 FROM merchant_skus ms
						 JOIN merchant_api_keys mak ON mak.id = ms.api_key_id
						 JOIN merchants m ON m.id = ms.merchant_id
						 WHERE ms.id = $1 AND ms.status = 'active'
						   AND mak.provider = $2 AND mak.status = 'active'
						   AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')
						   AND m.status IN ('active', 'approved')
						   AND m.lifecycle_status <> 'suspended'
						   AND (mak.quota_limit IS NULL OR mak.quota_used < mak.quota_limit)
						 LIMIT 1`,
						*req.MerchantSKUID, req.Provider,
					),
					apiKey,
				)
				if err == nil {
					return nil
				}
				if err != sql.ErrNoRows {
					return err
				}
			}
			return sql.ErrNoRows
		}
		err := scanMerchantAPIKeyQuotaRow(
			db.QueryRow(
				`SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status
				 FROM merchant_skus ms
				 JOIN merchant_api_keys mak ON mak.id = ms.api_key_id
				 JOIN merchants m ON m.id = ms.merchant_id
				 WHERE ms.id = $1 AND ms.status = 'active'
				   AND ms.merchant_id = $2
				   AND m.user_id = $3
				   AND mak.provider = $4 AND mak.status = 'active'
				   AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')
				   AND m.lifecycle_status <> 'suspended'`,
				*req.MerchantSKUID, merchantID, userID, req.Provider,
			),
			apiKey,
		)
		if err == nil {
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
	}

	if merchantID > 0 {
		return scanMerchantAPIKeyQuotaRow(
			db.QueryRow(
				`SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status
				 FROM merchant_api_keys
				 WHERE provider = $1 AND status = 'active'
				   AND merchant_id = $2
				   AND (verified_at IS NOT NULL OR verification_result = 'verified')
				   AND (quota_limit IS NULL OR quota_used < quota_limit)
				 ORDER BY COALESCE((quota_limit - quota_used)::double precision, 1e30::double precision) DESC
				 LIMIT 1`,
				req.Provider, merchantID,
			),
			apiKey,
		)
	}

	return scanMerchantAPIKeyQuotaRow(
		db.QueryRow(
			`SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status
			 FROM merchant_api_keys mak
			 INNER JOIN merchants m ON m.id = mak.merchant_id
			 WHERE mak.provider = $1 AND mak.status = 'active'
			   AND m.user_id = $2
			   AND m.status IN ('active', 'approved')
			   AND m.lifecycle_status <> 'suspended'
			   AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')
			   AND (mak.quota_limit IS NULL OR mak.quota_used < mak.quota_limit)
			 ORDER BY COALESCE((mak.quota_limit - mak.quota_used)::double precision, 1e30::double precision) DESC
			 LIMIT 1`,
			req.Provider, userID,
		),
		apiKey,
	)
}

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
