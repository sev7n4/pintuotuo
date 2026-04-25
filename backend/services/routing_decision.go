package services

import (
	"encoding/json"
	"time"
)

type RoutingDecision struct {
	RequestID  string `json:"request_id"`
	MerchantID int    `json:"merchant_id"`
	Model      string `json:"model"`
	Provider   string `json:"provider,omitempty"`

	StrategyLayerGoal   StrategyGoal    `json:"strategy_layer_goal"`
	StrategyLayerReason string          `json:"strategy_layer_reason,omitempty"`
	StrategyLayerInput  json.RawMessage `json:"strategy_layer_input,omitempty"`
	StrategyLayerOutput json.RawMessage `json:"strategy_layer_output,omitempty"`

	DecisionLayerCandidates []RoutingCandidateScore `json:"decision_layer_candidates,omitempty"`
	DecisionLayerOutput     json.RawMessage         `json:"decision_layer_output,omitempty"`
	RoutingMode             string                  `json:"routing_mode,omitempty"`

	ExecutionLayerInput   json.RawMessage `json:"execution_layer_input,omitempty"`
	ExecutionLayerResult  json.RawMessage `json:"execution_layer_result,omitempty"`
	ExecutionSuccess      bool            `json:"execution_success"`
	ExecutionStatusCode   int             `json:"execution_status_code,omitempty"`
	ExecutionLatencyMs    int             `json:"execution_latency_ms,omitempty"`
	ExecutionErrorMessage string          `json:"execution_error_message,omitempty"`

	SelectedAPIKeyID   int     `json:"selected_api_key_id"`
	SelectedMerchantID int     `json:"selected_merchant_id"`
	SelectedProvider   string  `json:"selected_provider"`
	SelectedModel      string  `json:"selected_model"`
	InputTokenCost     float64 `json:"input_token_cost,omitempty"`
	OutputTokenCost    float64 `json:"output_token_cost,omitempty"`

	DecisionDurationMs int    `json:"decision_duration_ms"`
	DecisionResult     string `json:"decision_result"`
	ErrorMessage       string `json:"error_message,omitempty"`

	Timestamp time.Time `json:"timestamp"`
}

type RoutingCandidateScore struct {
	APIKeyID      int     `json:"api_key_id"`
	MerchantID    int     `json:"merchant_id"`
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	Score         float64 `json:"score"`
	PriceScore    float64 `json:"price_score"`
	LatencyScore  float64 `json:"latency_score"`
	SuccessScore  float64 `json:"success_score"`
	Region        string  `json:"region"`
	SecurityLevel string  `json:"security_level"`
	HealthStatus  string  `json:"health_status"`
	Verified      bool    `json:"verified"`
	InputPrice    float64 `json:"input_price"`
	OutputPrice   float64 `json:"output_price"`
	AvgLatencyMs  int     `json:"avg_latency_ms"`
	SuccessRate   float64 `json:"success_rate"`
}

type DecisionResult string

const (
	DecisionResultSuccess DecisionResult = "success"
	DecisionResultFailed  DecisionResult = "failed"
	DecisionResultTimeout DecisionResult = "timeout"
)

type ExecutionLayerInputData struct {
	GatewayMode   string `json:"gateway_mode"`
	EndpointURL   string `json:"endpoint_url"`
	AuthMethod    string `json:"auth_method"`
	ResolvedModel string `json:"resolved_model"`
	RequestFormat string `json:"request_format"`
}

type ExecutionLayerResultData struct {
	Success           bool   `json:"success"`
	StatusCode        int    `json:"status_code,omitempty"`
	LatencyMs         int    `json:"latency_ms,omitempty"`
	ErrorMessage      string `json:"error_message,omitempty"`
	Model             string `json:"model,omitempty"`
	Provider          string `json:"provider,omitempty"`
	ActualModel       string `json:"actual_model,omitempty"`
	InputTokens       int    `json:"input_tokens,omitempty"`
	OutputTokens      int    `json:"output_tokens,omitempty"`
	FinishReason      string `json:"finish_reason,omitempty"`
	ErrorType         string `json:"error_type,omitempty"`
	UpstreamRequestID string `json:"upstream_request_id,omitempty"`
	CacheHit          bool   `json:"cache_hit,omitempty"`
}
