package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pintuotuo/backend/config"
)

type IUnifiedRoutingEngine interface {
	Route(ctx context.Context, req *RoutingRequest) (*RoutingDecision, error)
	ExecuteWithStrategy(ctx context.Context, req *RoutingRequest, strategyOutput *StrategyOutput) (*RoutingDecision, error)
}

type RoutingRequest struct {
	RequestID      string                 `json:"request_id"`
	MerchantID     int                    `json:"merchant_id"`
	Model          string                 `json:"model"`
	Provider       string                 `json:"provider,omitempty"`
	RequestBody    map[string]interface{} `json:"request_body"`
	UserPrefs      map[string]interface{} `json:"user_preferences,omitempty"`
	CostBudget     *float64               `json:"cost_budget,omitempty"`
	ComplianceReqs []string               `json:"compliance_requirements,omitempty"`
	AllowedKeyIDs  []int                  `json:"allowed_key_ids,omitempty"`
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
		RequestID:         req.RequestID,
		MerchantID:        req.MerchantID,
		Model:             req.Model,
		Provider:          req.Provider,
		StrategyLayerGoal: strategyOutput.Goal,
		Timestamp:         startTime,
	}

	defer func() {
		decision.DecisionDurationMs = int(time.Since(startTime).Milliseconds())
	}()

	candidate, err := e.smartRouter.SelectProviderWithStrategyOutput(
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

	decision.SelectedAPIKeyID = candidate.APIKeyID
	decision.SelectedProvider = candidate.Provider
	decision.SelectedModel = candidate.Model
	decision.RoutingMode = determineRoutingMode(candidate)
	decision.DecisionResult = string(DecisionResultSuccess)

	decisionOutput := map[string]interface{}{
		"api_key_id":     candidate.APIKeyID,
		"provider":       candidate.Provider,
		"model":          candidate.Model,
		"score":          candidate.Score,
		"price_score":    candidate.PriceScore,
		"latency_score":  candidate.LatencyScore,
		"success_score":  candidate.SuccessScore,
		"region":         candidate.Region,
		"security_level": candidate.SecurityLevel,
		"routing_mode":   decision.RoutingMode,
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
			execution_layer_result, execution_success, execution_status_code, execution_latency_ms, execution_error_message,
			decision_duration_ms, decision_result, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	var apiKeyID *int
	if decision.SelectedAPIKeyID > 0 {
		apiKeyID = &decision.SelectedAPIKeyID
	}

	candidatesJSON, _ := json.Marshal(decision.DecisionLayerCandidates)

	_, err := e.db.ExecContext(ctx, query,
		decision.RequestID,
		decision.MerchantID,
		apiKeyID,
		string(decision.StrategyLayerGoal),
		decision.StrategyLayerReason,
		decision.StrategyLayerInput,
		decision.StrategyLayerOutput,
		candidatesJSON,
		decision.DecisionLayerOutput,
		decision.RoutingMode,
		decision.ExecutionLayerResult,
		decision.ExecutionSuccess,
		decision.ExecutionStatusCode,
		decision.ExecutionLatencyMs,
		decision.ExecutionErrorMessage,
		decision.DecisionDurationMs,
		decision.DecisionResult,
		decision.ErrorMessage,
		decision.Timestamp,
	)

	return err
}
