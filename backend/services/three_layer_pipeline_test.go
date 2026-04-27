package services

import (
	"testing"
)

func TestThreeLayerRoutingPipeline_New(t *testing.T) {
	p := NewThreeLayerRoutingPipeline()
	if p == nil {
		t.Fatal("expected pipeline, got nil")
	}
	if p.strategyEngine == nil {
		t.Error("expected strategy engine, got nil")
	}
	if p.decisionEngine == nil {
		t.Error("expected decision engine, got nil")
	}
	if p.executionEngine == nil {
		t.Error("expected execution engine, got nil")
	}
}

func TestThreeLayerRoutingPipeline_NewWithEngines(t *testing.T) {
	strategy := NewRoutingStrategyEngine()
	decision := NewUnifiedRoutingEngine()
	execution := NewExecutionEngine()

	p := NewThreeLayerRoutingPipelineWithEngines(strategy, decision, execution)
	if p == nil {
		t.Fatal("expected pipeline, got nil")
	}
}

func TestThreeLayerRoutingPipeline_Execute_StrategyLayerError(t *testing.T) {
	t.Skip("requires database connection - covered by integration tests")
}

func TestThreeLayerRoutingPipeline_Execute_DecisionLayerSuccess(t *testing.T) {
	t.Skip("requires database connection - covered by integration tests")
}

func TestThreeLayerRoutingPipeline_RecordExecutionInput(t *testing.T) {
	p := NewThreeLayerRoutingPipeline()

	decision := &RoutingDecision{
		RequestID: "test-request",
	}

	input := &ExecutionLayerInputData{
		GatewayMode:   "direct",
		EndpointURL:   "https://api.openai.com/v1/chat/completions",
		AuthMethod:    "bearer",
		ResolvedModel: "gpt-4",
		RequestFormat: "openai",
	}

	p.RecordExecutionInput(decision, input)

	if decision.ExecutionLayerInput == nil {
		t.Error("expected ExecutionLayerInput to be set")
	}
}

func TestThreeLayerRoutingPipeline_RecordExecutionResult(t *testing.T) {
	p := NewThreeLayerRoutingPipeline()

	decision := &RoutingDecision{
		RequestID: "test-request",
	}

	p.RecordExecutionResult(decision, true, 200, 100, "")

	if !decision.ExecutionSuccess {
		t.Error("expected ExecutionSuccess to be true")
	}
	if decision.ExecutionStatusCode != 200 {
		t.Errorf("expected ExecutionStatusCode 200, got %d", decision.ExecutionStatusCode)
	}
	if decision.ExecutionLatencyMs != 100 {
		t.Errorf("expected ExecutionLatencyMs 100, got %d", decision.ExecutionLatencyMs)
	}
}

func TestThreeLayerRoutingPipeline_RecordExecutionResult_Failure(t *testing.T) {
	p := NewThreeLayerRoutingPipeline()

	decision := &RoutingDecision{
		RequestID: "test-request",
	}

	p.RecordExecutionResult(decision, false, 500, 50, "internal error")

	if decision.ExecutionSuccess {
		t.Error("expected ExecutionSuccess to be false")
	}
	if decision.ExecutionStatusCode != 500 {
		t.Errorf("expected ExecutionStatusCode 500, got %d", decision.ExecutionStatusCode)
	}
	if decision.ExecutionErrorMessage != "internal error" {
		t.Errorf("expected ExecutionErrorMessage 'internal error', got %s", decision.ExecutionErrorMessage)
	}
}

func TestThreeLayerRoutingPipeline_RecordExecutionResultExtended(t *testing.T) {
	p := NewThreeLayerRoutingPipeline()

	decision := &RoutingDecision{
		RequestID: "test-request",
	}

	result := &ExecutionLayerResultData{
		Success:      false,
		StatusCode:   500,
		LatencyMs:    50,
		ErrorMessage: "internal error",
		Model:        "gpt-4",
		Provider:     "openai",
		ActualModel:  "gpt-4-turbo",
		InputTokens:  20,
		OutputTokens: 10,
		FinishReason: "stop",
	}

	p.RecordExecutionResultExtended(decision, result)

	if decision.ExecutionLayerResult == nil {
		t.Error("expected ExecutionLayerResult to be set")
	}
	if decision.ExecutionSuccess {
		t.Error("expected ExecutionSuccess to be false")
	}
	if decision.ExecutionStatusCode != 500 {
		t.Errorf("expected ExecutionStatusCode 500, got %d", decision.ExecutionStatusCode)
	}
}

func TestThreeLayerRoutingPipeline_GetLayerMetrics(t *testing.T) {
	p := NewThreeLayerRoutingPipeline()

	metrics := p.GetLayerMetrics()

	if metrics == nil {
		t.Fatal("expected metrics, got nil")
	}
	if metrics["strategy_layer"] == nil {
		t.Error("expected strategy_layer metrics")
	}
	if metrics["decision_layer"] == nil {
		t.Error("expected decision_layer metrics")
	}
	if metrics["execution_layer"] == nil {
		t.Error("expected execution_layer metrics")
	}
}
