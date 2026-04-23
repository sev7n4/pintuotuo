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
	StrategyLayerInput  json.RawMessage `json:"strategy_layer_input,omitempty"`
	StrategyLayerOutput json.RawMessage `json:"strategy_layer_output,omitempty"`

	DecisionLayerCandidates []RoutingCandidateScore `json:"decision_layer_candidates,omitempty"`
	DecisionLayerOutput     json.RawMessage         `json:"decision_layer_output,omitempty"`

	ExecutionLayerResult json.RawMessage `json:"execution_layer_result,omitempty"`

	SelectedAPIKeyID int    `json:"selected_api_key_id"`
	SelectedProvider string `json:"selected_provider"`
	SelectedModel    string `json:"selected_model"`

	DecisionDurationMs int    `json:"decision_duration_ms"`
	DecisionResult     string `json:"decision_result"`
	ErrorMessage       string `json:"error_message,omitempty"`

	Timestamp time.Time `json:"timestamp"`
}

type RoutingCandidateScore struct {
	APIKeyID      int     `json:"api_key_id"`
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	Score         float64 `json:"score"`
	PriceScore    float64 `json:"price_score"`
	LatencyScore  float64 `json:"latency_score"`
	SuccessScore  float64 `json:"success_score"`
	Region        string  `json:"region"`
	SecurityLevel string  `json:"security_level"`
}

type DecisionResult string

const (
	DecisionResultSuccess DecisionResult = "success"
	DecisionResultFailed  DecisionResult = "failed"
	DecisionResultTimeout DecisionResult = "timeout"
)
