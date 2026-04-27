package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

const upstreamErrorBodyPeek = 4 * 1024

var (
	proxyHTTPOnce         sync.Once
	proxyHTTPRoundTripper http.RoundTripper
)

func proxyHTTPClient(timeout time.Duration) *http.Client {
	proxyHTTPOnce.Do(func() {
		proxyHTTPRoundTripper = otelhttp.NewTransport(http.DefaultTransport)
	})
	return &http.Client{Transport: proxyHTTPRoundTripper, Timeout: timeout}
}

func applyLitellmGatewayRetryCap(policy *services.RetryPolicy) *services.RetryPolicy {
	if policy == nil {
		return policy
	}
	active := strings.TrimSpace(strings.ToLower(os.Getenv("LLM_GATEWAY_ACTIVE")))
	if active != llmGatewayLitellm {
		return policy
	}
	cap := 1
	if v := strings.TrimSpace(os.Getenv("API_PROXY_LITELLM_MAX_RETRIES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cap = n
		}
	}
	if policy.MaxRetries > cap {
		p := *policy
		p.MaxRetries = cap
		return &p
	}
	return policy
}

func applyProxyUpstreamHeaders(c *gin.Context, httpReq *http.Request, requestID string) {
	if strings.TrimSpace(requestID) != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}
	if tracing.Enabled() {
		for _, h := range []string{"tracestate", "baggage"} {
			if v := strings.TrimSpace(c.GetHeader(h)); v != "" {
				httpReq.Header.Set(h, v)
			}
		}
		return
	}
	for _, h := range []string{"traceparent", "tracestate", "baggage"} {
		if v := strings.TrimSpace(c.GetHeader(h)); v != "" {
			httpReq.Header.Set(h, v)
		}
	}
}

func calculateTokenCost(db *sql.DB, userID int, provider, model string, inputTokens, outputTokens int, strictPricingVID *int) (float64, error) {
	if strictPricingVID != nil {
		cost, ok := services.CalculateCostFromPricingVersion(db, *strictPricingVID, provider, model, inputTokens, outputTokens)
		if !ok {
			return 0, fmt.Errorf("strict pricing snapshot miss for version %d", *strictPricingVID)
		}
		logger.LogDebug(context.Background(), "api_proxy", "Token cost from entitlement pricing_version", map[string]interface{}{
			"pricing_version_id": *strictPricingVID,
			"pricing_source":     "entitlement_strict",
			"provider":           provider,
			"model":              model,
			"input_tokens":       inputTokens,
			"output_tokens":      outputTokens,
			"cost":               cost,
		})
		return cost, nil
	}

	vid := services.LatestUserPricingVersionID(db, userID)
	if vid.Valid {
		if cost, ok := services.CalculateCostFromPricingVersion(db, int(vid.Int64), provider, model, inputTokens, outputTokens); ok {
			logger.LogDebug(context.Background(), "api_proxy", "Token cost from pricing_version snapshot", map[string]interface{}{
				"pricing_version_id": vid.Int64,
				"pricing_source":     "pricing_version_spu_rates",
				"provider":           provider,
				"model":              model,
				"input_tokens":       inputTokens,
				"output_tokens":      outputTokens,
				"cost":               cost,
			})
			return cost, nil
		}
	}

	pricingService := services.GetPricingService()
	cost := pricingService.CalculateCost(provider, model, inputTokens, outputTokens)

	logger.LogDebug(context.Background(), "api_proxy", "Token cost calculated (live SPU)", map[string]interface{}{
		"provider":       provider,
		"model":          model,
		"input_tokens":   inputTokens,
		"output_tokens":  outputTokens,
		"cost":           cost,
		"pricing_source": "live_spu",
	})

	return cost, nil
}

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
