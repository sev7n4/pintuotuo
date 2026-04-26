package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type ThreeLayerRoutingPipeline struct {
	strategyEngine  IStrategyEngine
	decisionEngine  IUnifiedRoutingEngine
	executionEngine *ExecutionEngine
}

func NewThreeLayerRoutingPipeline() *ThreeLayerRoutingPipeline {
	return &ThreeLayerRoutingPipeline{
		strategyEngine:  NewRoutingStrategyEngine(),
		decisionEngine:  NewUnifiedRoutingEngine(),
		executionEngine: NewExecutionEngine(),
	}
}

func NewThreeLayerRoutingPipelineWithEngines(strategy IStrategyEngine, decision IUnifiedRoutingEngine, execution *ExecutionEngine) *ThreeLayerRoutingPipeline {
	return &ThreeLayerRoutingPipeline{
		strategyEngine:  strategy,
		decisionEngine:  decision,
		executionEngine: execution,
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

	if req.ExecuteImmediately {
		decision, err = p.executeExecutionLayer(ctx, req, decision)
		if err != nil {
			return nil, fmt.Errorf("execution layer failed: %w", err)
		}
	}

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

func (p *ThreeLayerRoutingPipeline) executeExecutionLayer(ctx context.Context, req *RoutingRequest, decision *RoutingDecision) (*RoutingDecision, error) {
	if decision == nil {
		return nil, fmt.Errorf("decision is required for execution")
	}

	if req.DecryptedAPIKey == "" {
		return decision, nil
	}

	execInput := &ExecutionInput{
		Provider:      decision.SelectedProvider,
		Model:         decision.SelectedModel,
		APIKey:        req.DecryptedAPIKey,
		EndpointURL:   req.ProviderBaseURL,
		RequestFormat: req.ProviderAPIFormat,
		Stream:        req.Stream,
	}

	if req.RequestBody != nil {
		bodyBytes, err := json.Marshal(req.RequestBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		execInput.RequestBody = bodyBytes

		if messages, ok := req.RequestBody["messages"]; ok {
			if msgBytes, err := json.Marshal(messages); err == nil {
				var msgs []Message
				if json.Unmarshal(msgBytes, &msgs) == nil {
					execInput.Messages = msgs
				}
			}
		}

		if options, ok := req.RequestBody["options"]; ok {
			if optBytes, err := json.Marshal(options); err == nil {
				execInput.Options = optBytes
			}
		}
	}

	execLayer := NewExecutionLayer(nil, p.executionEngine)
	layerInput := &ExecutionLayerInput{
		RoutingDecision: decision,
		ProviderConfig: &ExecutionProviderConfig{
			Code:       decision.SelectedProvider,
			APIBaseURL: req.ProviderBaseURL,
			APIFormat:  req.ProviderAPIFormat,
		},
		DecryptedAPIKey: req.DecryptedAPIKey,
		Stream:          req.Stream,
	}

	output, err := execLayer.Execute(ctx, layerInput)
	if err != nil {
		p.recordExecutionError(decision, err)
		return decision, err
	}

	decision.ExecutionSuccess = output.Result.Success
	decision.ExecutionStatusCode = output.Result.StatusCode
	decision.ExecutionLatencyMs = output.Result.LatencyMs
	if output.Result.ErrorMessage != "" {
		decision.ExecutionErrorMessage = output.Result.ErrorMessage
	}

	if output.Result.Success {
		decision.DecisionResult = string(DecisionResultSuccess)
	} else {
		decision.DecisionResult = string(DecisionResultFailed)
		decision.ErrorMessage = output.Result.ErrorMessage
	}

	return decision, nil
}

func (p *ThreeLayerRoutingPipeline) recordExecutionError(decision *RoutingDecision, err error) {
	decision.ExecutionSuccess = false
	decision.ExecutionStatusCode = 500
	decision.ExecutionErrorMessage = err.Error()
	decision.DecisionResult = string(DecisionResultFailed)
	decision.ErrorMessage = err.Error()
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

func (p *ThreeLayerRoutingPipeline) RecordExecutionInput(decision *RoutingDecision, input *ExecutionLayerInputData) {
	if input == nil {
		return
	}
	if inputBytes, err := json.Marshal(input); err == nil {
		decision.ExecutionLayerInput = inputBytes
	}
}

func (p *ThreeLayerRoutingPipeline) RecordExecutionResultExtended(decision *RoutingDecision, result *ExecutionLayerResultData) {
	if result == nil {
		return
	}

	decision.ExecutionSuccess = result.Success
	decision.ExecutionStatusCode = result.StatusCode
	decision.ExecutionLatencyMs = result.LatencyMs
	decision.ExecutionErrorMessage = result.ErrorMessage

	if !result.Success {
		decision.DecisionResult = string(DecisionResultFailed)
		if result.ErrorMessage != "" {
			decision.ErrorMessage = result.ErrorMessage
		}
	} else {
		decision.DecisionResult = string(DecisionResultSuccess)
	}

	if resultBytes, err := json.Marshal(result); err == nil {
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
