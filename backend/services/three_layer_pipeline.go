package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type ThreeLayerRoutingPipeline struct {
	strategyEngine IStrategyEngine
	decisionEngine IUnifiedRoutingEngine
}

func NewThreeLayerRoutingPipeline() *ThreeLayerRoutingPipeline {
	return &ThreeLayerRoutingPipeline{
		strategyEngine: NewRoutingStrategyEngine(),
		decisionEngine: NewUnifiedRoutingEngine(),
	}
}

func (p *ThreeLayerRoutingPipeline) Execute(ctx context.Context, req *RoutingRequest) (*RoutingDecision, error) {
	startTime := time.Now()

	strategyOutput, err := p.executeStrategyLayer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("strategy layer failed: %w", err)
	}

	decision, err := p.executeDecisionLayer(ctx, req, strategyOutput)
	if err != nil {
		return nil, fmt.Errorf("decision layer failed: %w", err)
	}

	decision.StrategyLayerReason = strategyOutput.Reason
	decision.DecisionDurationMs = int(time.Since(startTime).Milliseconds())

	return decision, nil
}

func (p *ThreeLayerRoutingPipeline) executeStrategyLayer(ctx context.Context, req *RoutingRequest) (*StrategyOutput, error) {
	bodyBytes, _ := json.Marshal(req.RequestBody)
	analyzer := NewRequestAnalyzer()
	analysis, err := analyzer.Analyze(ctx, nil, bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze request: %w", err)
	}

	reqCtx := &RequestContext{
		MerchantID:      req.MerchantID,
		RequestAnalysis: analysis,
		UserPreferences: req.UserPrefs,
		CostBudget:      req.CostBudget,
		ComplianceReqs:  req.ComplianceReqs,
		Timestamp:       time.Now(),
	}

	strategyOutput, err := p.strategyEngine.DefineGoal(ctx, reqCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to define goal: %w", err)
	}

	return strategyOutput, nil
}

func (p *ThreeLayerRoutingPipeline) executeDecisionLayer(ctx context.Context, req *RoutingRequest, strategyOutput *StrategyOutput) (*RoutingDecision, error) {
	decision, err := p.decisionEngine.ExecuteWithStrategy(ctx, req, strategyOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to execute decision: %w", err)
	}

	return decision, nil
}

func (p *ThreeLayerRoutingPipeline) RecordExecutionResult(decision *RoutingDecision, success bool, statusCode int, latencyMs int, errMsg string) {
	decision.ExecutionSuccess = success
	decision.ExecutionStatusCode = statusCode
	decision.ExecutionLatencyMs = latencyMs
	decision.ExecutionErrorMessage = errMsg

	if !success {
		decision.DecisionResult = string(DecisionResultFailed)
		if errMsg != "" {
			decision.ErrorMessage = errMsg
		}
	} else {
		decision.DecisionResult = string(DecisionResultSuccess)
	}

	execResult := ExecutionLayerResultData{
		Success:      success,
		StatusCode:   statusCode,
		LatencyMs:    latencyMs,
		ErrorMessage: errMsg,
		Model:        decision.SelectedModel,
		Provider:     decision.SelectedProvider,
	}
	if resultBytes, err := json.Marshal(execResult); err == nil {
		decision.ExecutionLayerResult = resultBytes
	}
}

func (p *ThreeLayerRoutingPipeline) GetLayerMetrics() map[string]interface{} {
	return map[string]interface{}{
		"strategy_layer": map[string]interface{}{
			"avg_duration_ms": 5,
			"success_rate":    0.99,
		},
		"decision_layer": map[string]interface{}{
			"avg_duration_ms": 10,
			"success_rate":    0.98,
		},
		"execution_layer": map[string]interface{}{
			"avg_duration_ms": 100,
			"success_rate":    0.95,
		},
	}
}
