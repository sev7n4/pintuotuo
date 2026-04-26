package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
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

func applyGatewayOverride(cfg providerRuntimeConfig) providerRuntimeConfig {
	active := strings.TrimSpace(strings.ToLower(os.Getenv("LLM_GATEWAY_ACTIVE")))
	if cfg.APIFormat != apiFormatOpenAI || active == "" || active == "none" {
		return cfg
	}
	switch active {
	case llmGatewayLitellm:
		if base := strings.TrimSpace(os.Getenv("LLM_GATEWAY_LITELLM_URL")); base != "" {
			cfg.APIBaseURL = strings.TrimRight(base, "/") + "/v1"
		}
	}
	return cfg
}

func resolveGatewayAuthToken(cfg providerRuntimeConfig, fallbackToken string) string {
	active := strings.TrimSpace(strings.ToLower(os.Getenv("LLM_GATEWAY_ACTIVE")))
	if cfg.APIFormat != apiFormatOpenAI || active == "" || active == "none" {
		return fallbackToken
	}
	switch active {
	case llmGatewayLitellm:
		if token := strings.TrimSpace(os.Getenv("LITELLM_MASTER_KEY")); token != "" {
			return token
		}
	}
	return fallbackToken
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

func executeProviderRequestWithRetry(client *http.Client, baseReq *http.Request, policy *services.RetryPolicy) (*http.Response, int, error) {
	if policy == nil {
		policy = services.DefaultRetryPolicy
	}
	var (
		resp       *http.Response
		err        error
		retryCount int
	)
	ctx := baseReq.Context()
	for i := 0; i <= policy.MaxRetries; i++ {
		req := baseReq.Clone(ctx)
		if baseReq.GetBody != nil {
			body, bodyErr := baseReq.GetBody()
			if bodyErr == nil {
				req.Body = body
			}
		}

		resp, err = client.Do(req) // #nosec G704 -- upstream URL from admin-configured model_providers.api_base_url, not user-supplied host
		if err != nil {
			info := services.MapProviderError(0, "", err.Error(), nil, err, "")
			if !info.Retryable || i >= policy.MaxRetries {
				return nil, retryCount, err
			}
			retryCount++
			time.Sleep(policy.DelayForAttempt(i))
			continue
		}

		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < http.StatusInternalServerError {
			return resp, retryCount, nil
		}

		b, _ := io.ReadAll(io.LimitReader(resp.Body, upstreamErrorBodyPeek))
		_ = resp.Body.Close()
		status := resp.StatusCode
		headers := resp.Header
		retryable := services.HTTPUpstreamRetryable(status, b, headers)
		if !retryable || i >= policy.MaxRetries {
			resp.Body = io.NopCloser(bytes.NewReader(b))
			return resp, retryCount, nil
		}
		retryCount++
		time.Sleep(policy.DelayForAttempt(i))
	}
	return nil, retryCount, err
}
