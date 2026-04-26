package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
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
	Code        string `json:"code"`
	Name        string `json:"name"`
	APIBaseURL  string `json:"api_base_url"`
	APIFormat   string `json:"api_format"`
	GatewayMode string `json:"gateway_mode"`
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

	execInput, err := l.prepareExecutionInput(input)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare execution input: %w", err)
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

func (l *ExecutionLayer) prepareExecutionInput(input *ExecutionLayerInput) (*ExecutionInput, error) {
	if input.ProviderConfig == nil {
		return nil, fmt.Errorf("provider config is required")
	}

	if input.DecryptedAPIKey == "" {
		return nil, fmt.Errorf("decrypted API key is required")
	}

	endpointURL := input.ProviderConfig.APIBaseURL
	if endpointURL == "" {
		return nil, fmt.Errorf("API base URL is required")
	}

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

	return &ExecutionInput{
		Provider:      input.ProviderConfig.Code,
		Model:         model,
		APIKey:        input.DecryptedAPIKey,
		EndpointURL:   endpointURL,
		RequestFormat: input.ProviderConfig.APIFormat,
		RequestBody:   input.RequestBody,
		Messages:      input.Messages,
		Stream:        input.Stream,
		Options:       input.Options,
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
