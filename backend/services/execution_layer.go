package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	GatewayModeDirect  = "direct"
	GatewayModeLitellm = "litellm"
	GatewayModeProxy   = "proxy"
)

type ExecutionLayer struct {
	engine *ExecutionEngine
	db     *sql.DB
}

type ExecutionLayerInput struct {
	RoutingDecision *RoutingDecision `json:"routing_decision"`
	RequestBody     []byte           `json:"request_body"`
	ProviderConfig  *ExecutionProviderConfig
	DecryptedAPIKey string          `json:"decrypted_api_key"`
	Messages        []Message       `json:"messages"`
	Stream          bool            `json:"stream"`
	Options         json.RawMessage `json:"options"`
}

type ExecutionLayerOutput struct {
	Result     *ExecutionResult `json:"result"`
	Decision   *RoutingDecision `json:"decision"`
	DurationMs int              `json:"duration_ms"`
}

type ExecutionProviderConfig struct {
	Code            string                 `json:"code"`
	Name            string                 `json:"name"`
	APIBaseURL      string                 `json:"api_base_url"`
	APIFormat       string                 `json:"api_format"`
	GatewayMode     string                 `json:"gateway_mode"`
	ProviderRegion  string                 `json:"provider_region"`
	RouteStrategy   map[string]interface{} `json:"route_strategy"`
	Endpoints       map[string]interface{} `json:"endpoints"`
	BYOKEndpointURL string                 `json:"byok_endpoint_url"`
	BYOKRouteMode   string                 `json:"byok_route_mode"`
	BYOKRouteConfig map[string]interface{} `json:"byok_route_config"`
	BYOKFallbackURL string                 `json:"byok_fallback_url"`
}

func NewExecutionLayer(db *sql.DB, engine *ExecutionEngine) *ExecutionLayer {
	if engine == nil {
		engine = NewExecutionEngine()
	}
	return &ExecutionLayer{
		engine: engine,
		db:     db,
	}
}

func (l *ExecutionLayer) Execute(ctx context.Context, input *ExecutionLayerInput) (*ExecutionLayerOutput, error) {
	startTime := time.Now()

	if input.ProviderConfig == nil {
		return nil, fmt.Errorf("provider config is required")
	}

	if input.DecryptedAPIKey == "" {
		return nil, fmt.Errorf("decrypted API key is required")
	}

	gatewayMode := l.determineGatewayMode(input.ProviderConfig)
	input.ProviderConfig.GatewayMode = gatewayMode

	endpointURL := l.resolveEndpoint(input.ProviderConfig)
	if endpointURL == "" {
		return nil, fmt.Errorf("failed to resolve endpoint URL")
	}

	authToken := l.resolveAuthToken(input.ProviderConfig, input.DecryptedAPIKey)

	model := ""
	if input.RoutingDecision != nil {
		model = input.RoutingDecision.SelectedModel
	}

	if model == "" && len(input.RequestBody) > 0 {
		var req struct {
			Model string `json:"model"`
		}
		if err := json.Unmarshal(input.RequestBody, &req); err == nil {
			model = req.Model
		}
	}

	execInput := &ExecutionInput{
		Provider:        input.ProviderConfig.Code,
		Model:           model,
		APIKey:          authToken,
		EndpointURL:     endpointURL,
		RequestFormat:   input.ProviderConfig.APIFormat,
		RequestBody:     input.RequestBody,
		Messages:        input.Messages,
		Stream:          input.Stream,
		Options:         input.Options,
		OriginalAPIKey:  input.DecryptedAPIKey,
		GatewayMode:     gatewayMode,
		ProviderBaseURL: input.ProviderConfig.APIBaseURL,
	}

	result, err := l.engine.ExecuteWithRetry(ctx, execInput)
	if err != nil {
		result = &ExecutionResult{
			Success:      false,
			ErrorMessage: err.Error(),
		}
	}

	if input.RoutingDecision != nil {
		l.recordExecutionResult(input.RoutingDecision, result)
	}

	return &ExecutionLayerOutput{
		Result:     result,
		Decision:   input.RoutingDecision,
		DurationMs: int(time.Since(startTime).Milliseconds()),
	}, nil
}

func (l *ExecutionLayer) recordExecutionResult(decision *RoutingDecision, result *ExecutionResult) {
	if decision == nil || result == nil {
		return
	}

	decision.ExecutionSuccess = result.Success
	decision.ExecutionStatusCode = result.StatusCode
	decision.ExecutionLatencyMs = result.LatencyMs
	decision.ExecutionErrorMessage = result.ErrorMessage

	execResultData := map[string]interface{}{
		"success":       result.Success,
		"status_code":   result.StatusCode,
		"latency_ms":    result.LatencyMs,
		"error_message": result.ErrorMessage,
		"provider":      result.Provider,
		"actual_model":  result.ActualModel,
	}

	if result.Usage != nil {
		execResultData["usage"] = map[string]interface{}{
			"prompt_tokens":     result.Usage.PromptTokens,
			"completion_tokens": result.Usage.CompletionTokens,
			"total_tokens":      result.Usage.TotalTokens,
		}
	}

	if jsonData, err := json.Marshal(execResultData); err == nil {
		decision.ExecutionLayerResult = jsonData
	}
}

func (l *ExecutionLayer) UpdateRoutingDecisionLog(ctx context.Context, decisionID int, result *ExecutionResult) error {
	if l.db == nil {
		return nil
	}

	query := `
		UPDATE routing_decision_logs 
		SET execution_success = $1,
		    execution_status_code = $2,
		    execution_latency_ms = $3,
		    execution_layer_result = $4,
		    updated_at = NOW()
		WHERE id = $5
	`

	execResultJSON, _ := json.Marshal(map[string]interface{}{
		"success":       result.Success,
		"status_code":   result.StatusCode,
		"latency_ms":    result.LatencyMs,
		"error_message": result.ErrorMessage,
		"provider":      result.Provider,
		"actual_model":  result.ActualModel,
	})

	_, err := l.db.ExecContext(ctx, query,
		result.Success,
		result.StatusCode,
		result.LatencyMs,
		execResultJSON,
		decisionID,
	)

	return err
}

func (l *ExecutionLayer) GetProviderConfig(ctx context.Context, providerCode string) (*ExecutionProviderConfig, error) {
	if l.db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	query := `
		SELECT code, name, COALESCE(api_base_url, ''), COALESCE(api_format, 'openai')
		FROM model_providers
		WHERE code = $1 AND status = 'active'
	`

	var cfg ExecutionProviderConfig
	err := l.db.QueryRowContext(ctx, query, providerCode).Scan(
		&cfg.Code,
		&cfg.Name,
		&cfg.APIBaseURL,
		&cfg.APIFormat,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("provider %s not found or inactive", providerCode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query provider config: %w", err)
	}

	return &cfg, nil
}

func (l *ExecutionLayer) resolveEndpoint(cfg *ExecutionProviderConfig) string {
	if cfg == nil {
		return ""
	}

	if cfg.BYOKEndpointURL != "" {
		return cfg.BYOKEndpointURL
	}

	if cfg.BYOKRouteConfig != nil {
		if endpoints, ok := cfg.BYOKRouteConfig["endpoints"].(map[string]interface{}); ok {
			region := cfg.ProviderRegion
			if region == "" {
				region = "overseas"
			}
			if url, ok := endpoints[region].(string); ok && url != "" {
				return url
			}
			if defaultURL, ok := endpoints["default"].(string); ok && defaultURL != "" {
				return defaultURL
			}
		}
		if baseURL, ok := cfg.BYOKRouteConfig["base_url"].(string); ok && baseURL != "" {
			return baseURL
		}
	}

	if cfg.BYOKFallbackURL != "" {
		return cfg.BYOKFallbackURL
	}

	switch cfg.GatewayMode {
	case GatewayModeLitellm:
		if cfg.Endpoints != nil {
			if litellmEndpoints, ok := cfg.Endpoints[GatewayModeLitellm].(map[string]interface{}); ok {
				region := cfg.ProviderRegion
				if region == "" {
					region = "domestic"
				}
				if url, ok := litellmEndpoints[region].(string); ok && url != "" {
					return url
				}
			}
		}
		litellmURL := os.Getenv("LLM_GATEWAY_LITELLM_URL")
		if litellmURL != "" {
			return litellmURL + "/v1"
		}
		return cfg.APIBaseURL

	case GatewayModeProxy:
		if cfg.Endpoints != nil {
			if proxyEndpoints, ok := cfg.Endpoints[GatewayModeProxy].(map[string]interface{}); ok {
				if gaapURL, ok := proxyEndpoints["gaap"].(string); ok && gaapURL != "" {
					return gaapURL
				}
				for _, v := range proxyEndpoints {
					if url, ok := v.(string); ok && url != "" {
						return url
					}
				}
			}
		}
		proxyURL := os.Getenv("LLM_GATEWAY_PROXY_URL")
		if proxyURL != "" {
			return proxyURL
		}
		return cfg.APIBaseURL

	default:
		if cfg.Endpoints != nil {
			if directEndpoints, ok := cfg.Endpoints[GatewayModeDirect].(map[string]interface{}); ok {
				region := cfg.ProviderRegion
				if region == "" {
					region = "overseas"
				}
				if url, ok := directEndpoints[region].(string); ok && url != "" {
					return url
				}
			}
		}
		return cfg.APIBaseURL
	}
}

func (l *ExecutionLayer) resolveAuthToken(cfg *ExecutionProviderConfig, originalAPIKey string) string {
	if cfg == nil {
		return originalAPIKey
	}

	switch cfg.GatewayMode {
	case GatewayModeLitellm:
		masterKey := os.Getenv("LITELLM_MASTER_KEY")
		if masterKey != "" {
			return masterKey
		}
		return originalAPIKey

	case GatewayModeProxy:
		proxyToken := os.Getenv("LLM_GATEWAY_PROXY_TOKEN")
		if proxyToken != "" {
			return proxyToken
		}
		return originalAPIKey

	default:
		return originalAPIKey
	}
}

func (l *ExecutionLayer) determineGatewayMode(cfg *ExecutionProviderConfig) string {
	if cfg == nil {
		return GatewayModeDirect
	}

	if cfg.BYOKRouteMode != "" && cfg.BYOKRouteMode != "auto" {
		return cfg.BYOKRouteMode
	}

	if cfg.GatewayMode != "" {
		return cfg.GatewayMode
	}

	envMode := os.Getenv("LLM_GATEWAY_ACTIVE")
	if envMode != "" && envMode != "none" {
		return envMode
	}

	return GatewayModeDirect
}
