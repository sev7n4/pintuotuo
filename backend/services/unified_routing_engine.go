package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/metrics"
)

type IUnifiedRoutingEngine interface {
	Route(ctx context.Context, req *RoutingRequest) (*RoutingDecision, error)
	ExecuteWithStrategy(ctx context.Context, req *RoutingRequest, strategyOutput *StrategyOutput) (*RoutingDecision, error)
}

type RoutingRequest struct {
	RequestID           string                 `json:"request_id"`
	MerchantID          int                    `json:"merchant_id"`
	Model               string                 `json:"model"`
	Provider            string                 `json:"provider,omitempty"`
	RequestBody         map[string]interface{} `json:"request_body"`
	UserPrefs           map[string]interface{} `json:"user_preferences,omitempty"`
	CostBudget          *float64               `json:"cost_budget,omitempty"`
	ComplianceReqs      []string               `json:"compliance_requirements,omitempty"`
	AllowedKeyIDs       []int                  `json:"allowed_key_ids,omitempty"`
	Stream              bool                   `json:"stream,omitempty"`
	MaxTokens           int                    `json:"max_tokens,omitempty"`
	Priority            string                 `json:"priority,omitempty"`
	ExecuteImmediately  bool                   `json:"execute_immediately"`
	DecryptedAPIKey     string                 `json:"decrypted_api_key,omitempty"`
	ProviderBaseURL     string                 `json:"provider_base_url,omitempty"`
	ProviderAPIFormat   string                 `json:"provider_api_format,omitempty"`
}

type UnifiedRoutingEngine struct {
	db              *sql.DB
	strategyEngine  *RoutingStrategyEngine
	requestAnalyzer *RequestAnalyzer
	smartRouter     *SmartRouter
}

func NewUnifiedRoutingEngine() *UnifiedRoutingEngine {
	return &UnifiedRoutingEngine{
		db:              config.GetDB(),
		strategyEngine:  NewRoutingStrategyEngine(),
		requestAnalyzer: NewRequestAnalyzer(),
		smartRouter:     GetSmartRouter(),
	}
}

func (e *UnifiedRoutingEngine) Route(ctx context.Context, req *RoutingRequest) (*RoutingDecision, error) {
	startTime := time.Now()

	decision := &RoutingDecision{
		RequestID:  req.RequestID,
		MerchantID: req.MerchantID,
		Model:      req.Model,
		Provider:   req.Provider,
		Timestamp:  startTime,
	}

	defer func() {
		decision.DecisionDurationMs = int(time.Since(startTime).Milliseconds())
	}()

	bodyBytes, err := json.Marshal(req.RequestBody)
	if err != nil {
		decision.DecisionResult = string(DecisionResultFailed)
		decision.ErrorMessage = fmt.Sprintf("failed to marshal request body: %v", err)
		return decision, fmt.Errorf(decision.ErrorMessage)
	}

	analysis, err := e.requestAnalyzer.Analyze(ctx, nil, bodyBytes)
	if err != nil {
		decision.DecisionResult = string(DecisionResultFailed)
		decision.ErrorMessage = fmt.Sprintf("failed to analyze request: %v", err)
		return decision, fmt.Errorf(decision.ErrorMessage)
	}

	reqCtx := &RequestContext{
		MerchantID:      req.MerchantID,
		RequestAnalysis: analysis,
		UserPreferences: req.UserPrefs,
		CostBudget:      req.CostBudget,
		ComplianceReqs:  req.ComplianceReqs,
		Timestamp:       startTime,
	}

	strategyOutput, err := e.strategyEngine.DefineGoal(ctx, reqCtx)
	if err != nil {
		decision.DecisionResult = string(DecisionResultFailed)
		decision.ErrorMessage = fmt.Sprintf("failed to define strategy goal: %v", err)
		return decision, fmt.Errorf(decision.ErrorMessage)
	}

	decision.StrategyLayerGoal = strategyOutput.Goal

	if strategyInputBytes, err := json.Marshal(reqCtx); err == nil {
		decision.StrategyLayerInput = strategyInputBytes
	}

	if strategyOutputBytes, err := json.Marshal(strategyOutput); err == nil {
		decision.StrategyLayerOutput = strategyOutputBytes
	}

	return e.ExecuteWithStrategy(ctx, req, strategyOutput)
}

func (e *UnifiedRoutingEngine) ExecuteWithStrategy(ctx context.Context, req *RoutingRequest, strategyOutput *StrategyOutput) (*RoutingDecision, error) {
	startTime := time.Now()

	decision := &RoutingDecision{
		RequestID:           req.RequestID,
		MerchantID:          req.MerchantID,
		Model:               req.Model,
		Provider:            req.Provider,
		StrategyLayerGoal:   strategyOutput.Goal,
		StrategyLayerReason: strategyOutput.Reason,
		Timestamp:           startTime,
	}

	tokenEstimationService := NewTokenEstimationService()
	tokenEstimation := tokenEstimationService.EstimateTokens(req)
	decision.EstimatedInputTokens = tokenEstimation.EstimatedInputTokens
	decision.EstimatedOutputTokens = tokenEstimation.EstimatedOutputTokens
	decision.TokenEstimationSource = tokenEstimation.Source

	metrics.RecordTokenEstimationAccuracy(req.Model, tokenEstimation.Source,
		int(tokenEstimation.EstimatedInputTokens+tokenEstimation.EstimatedOutputTokens), 0)

	strategyInput := map[string]interface{}{
		"request_id":   req.RequestID,
		"merchant_id":  req.MerchantID,
		"model":        req.Model,
		"provider":     req.Provider,
		"allowed_keys": req.AllowedKeyIDs,
		"cost_budget":  req.CostBudget,
	}
	if inputBytes, err := json.Marshal(strategyInput); err == nil {
		decision.StrategyLayerInput = inputBytes
	}

	strategyOutputData := map[string]interface{}{
		"goal":        strategyOutput.Goal,
		"weights":     strategyOutput.Weights,
		"constraints": strategyOutput.Constraints,
		"priority":    strategyOutput.Priority,
		"reason":      strategyOutput.Reason,
	}
	if outputBytes, err := json.Marshal(strategyOutputData); err == nil {
		decision.StrategyLayerOutput = outputBytes
	}

	defer func() {
		decision.DecisionDurationMs = int(time.Since(startTime).Milliseconds())
	}()

	candidate, allCandidates, err := e.smartRouter.SelectProviderWithStrategyOutput(
		ctx,
		req.Model,
		req.Provider,
		strategyOutput,
		req.AllowedKeyIDs,
	)

	if err != nil {
		decision.DecisionResult = string(DecisionResultFailed)
		decision.ErrorMessage = fmt.Sprintf("failed to select provider: %v", err)
		return decision, fmt.Errorf(decision.ErrorMessage)
	}

	var candidateScores []RoutingCandidateScore
	for _, c := range allCandidates {
		candidateScores = append(candidateScores, RoutingCandidateScore{
			APIKeyID:      c.APIKeyID,
			MerchantID:    c.MerchantID,
			Provider:      c.Provider,
			Model:         c.Model,
			Score:         c.Score,
			PriceScore:    c.PriceScore,
			LatencyScore:  c.LatencyScore,
			SuccessScore:  c.SuccessScore,
			Region:        c.Region,
			SecurityLevel: c.SecurityLevel,
			HealthStatus:  c.HealthStatus,
			Verified:      c.Verified,
			InputPrice:    c.InputPrice,
			OutputPrice:   c.OutputPrice,
			AvgLatencyMs:  c.AvgLatencyMs,
			SuccessRate:   c.SuccessRate,
		})
	}
	decision.DecisionLayerCandidates = candidateScores

	decision.SelectedAPIKeyID = candidate.APIKeyID
	decision.SelectedMerchantID = candidate.MerchantID
	decision.SelectedProvider = candidate.Provider
	decision.SelectedModel = candidate.Model
	decision.InputTokenCost = candidate.InputPrice
	decision.OutputTokenCost = candidate.OutputPrice
	decision.RoutingMode = determineRoutingMode(candidate)
	decision.DecisionResult = string(DecisionResultSuccess)

	estimatedCost := (float64(tokenEstimation.EstimatedInputTokens)*candidate.InputPrice +
		float64(tokenEstimation.EstimatedOutputTokens)*candidate.OutputPrice) / 1000000.0
	metrics.RecordCostEstimation(req.Model, string(strategyOutput.Goal))
	metrics.RecordCostEstimationDeviation(req.Model, strconv.Itoa(candidate.MerchantID), estimatedCost, 0)

	decisionOutput := map[string]interface{}{
		"api_key_id":        candidate.APIKeyID,
		"merchant_id":       candidate.MerchantID,
		"provider":          candidate.Provider,
		"model":             candidate.Model,
		"score":             candidate.Score,
		"price_score":       candidate.PriceScore,
		"latency_score":     candidate.LatencyScore,
		"success_score":     candidate.SuccessScore,
		"region":            candidate.Region,
		"security_level":    candidate.SecurityLevel,
		"routing_mode":      decision.RoutingMode,
		"input_token_cost":  candidate.InputPrice,
		"output_token_cost": candidate.OutputPrice,
	}

	if outputBytes, err := json.Marshal(decisionOutput); err == nil {
		decision.DecisionLayerOutput = outputBytes
	}

	return decision, nil
}

func determineRoutingMode(candidate *RoutingCandidate) string {
	if candidate == nil {
		return "unknown"
	}
	return "direct"
}

func (e *UnifiedRoutingEngine) LogDecision(ctx context.Context, decision *RoutingDecision) error {
	query := `
		INSERT INTO routing_decision_logs (
			request_id, merchant_id, api_key_id,
			strategy_layer_goal, strategy_layer_reason, strategy_layer_input, strategy_layer_output,
			decision_layer_candidates, decision_layer_output, routing_mode,
			execution_layer_input, execution_layer_result, execution_success, execution_status_code, execution_latency_ms, execution_error_message,
			decision_duration_ms, decision_result, error_message, created_at,
			selected_merchant_id, input_token_cost, output_token_cost,
			estimated_input_tokens, estimated_output_tokens, token_estimation_source
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
	`

	var apiKeyID *int
	if decision.SelectedAPIKeyID > 0 {
		apiKeyID = &decision.SelectedAPIKeyID
	}

	var merchantID *int
	if decision.MerchantID > 0 {
		merchantID = &decision.MerchantID
	}

	var selectedMerchantID *int
	if decision.SelectedMerchantID > 0 {
		selectedMerchantID = &decision.SelectedMerchantID
	}

	candidatesJSON, _ := json.Marshal(decision.DecisionLayerCandidates)

	var strategyLayerInput, strategyLayerOutput, decisionLayerOutput, executionLayerInput, executionLayerResult interface{}
	if len(decision.StrategyLayerInput) > 0 {
		strategyLayerInput = decision.StrategyLayerInput
	}
	if len(decision.StrategyLayerOutput) > 0 {
		strategyLayerOutput = decision.StrategyLayerOutput
	}
	if len(decision.DecisionLayerOutput) > 0 {
		decisionLayerOutput = decision.DecisionLayerOutput
	}
	if len(decision.ExecutionLayerInput) > 0 {
		executionLayerInput = decision.ExecutionLayerInput
	}
	if len(decision.ExecutionLayerResult) > 0 {
		executionLayerResult = decision.ExecutionLayerResult
	}

	_, err := e.db.ExecContext(ctx, query,
		decision.RequestID,
		merchantID,
		apiKeyID,
		string(decision.StrategyLayerGoal),
		decision.StrategyLayerReason,
		strategyLayerInput,
		strategyLayerOutput,
		candidatesJSON,
		decisionLayerOutput,
		decision.RoutingMode,
		executionLayerInput,
		executionLayerResult,
		decision.ExecutionSuccess,
		decision.ExecutionStatusCode,
		decision.ExecutionLatencyMs,
		decision.ExecutionErrorMessage,
		decision.DecisionDurationMs,
		decision.DecisionResult,
		decision.ErrorMessage,
		decision.Timestamp,
		selectedMerchantID,
		decision.InputTokenCost,
		decision.OutputTokenCost,
		decision.EstimatedInputTokens,
		decision.EstimatedOutputTokens,
		decision.TokenEstimationSource,
	)

	return err
}
