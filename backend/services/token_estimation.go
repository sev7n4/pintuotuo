package services

import (
	"database/sql"
	"fmt"

	"github.com/pintuotuo/backend/config"
)

type TokenEstimation struct {
	EstimatedInputTokens  float64 `json:"estimated_input_tokens"`
	EstimatedOutputTokens float64 `json:"estimated_output_tokens"`
	Source                string  `json:"source"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ModelTokenStats struct {
	ModelName        string
	AvgInputTokens   float64
	AvgOutputTokens  float64
	P50InputTokens   int
	P50OutputTokens  int
	P90InputTokens   int
	P90OutputTokens  int
	InputOutputRatio float64
	TotalRequests    int
}

type TokenEstimationService struct {
	db *sql.DB
}

func NewTokenEstimationService() *TokenEstimationService {
	return &TokenEstimationService{
		db: config.GetDB(),
	}
}

func (s *TokenEstimationService) EstimateTokens(req *RoutingRequest) *TokenEstimation {
	if req == nil {
		return s.getFallbackEstimation()
	}

	if maxTokens := s.extractMaxTokens(req); maxTokens > 0 {
		inputTokens := s.estimateInputFromRequestBody(req.RequestBody)
		return &TokenEstimation{
			EstimatedInputTokens:  float64(inputTokens),
			EstimatedOutputTokens: float64(maxTokens),
			Source:                "request",
		}
	}

	if stats, err := s.getModelStatistics(req.Model); err == nil && stats != nil {
		return &TokenEstimation{
			EstimatedInputTokens:  stats.AvgInputTokens,
			EstimatedOutputTokens: stats.AvgOutputTokens,
			Source:                "statistics",
		}
	}

	return s.getFallbackEstimation()
}

func (s *TokenEstimationService) extractMaxTokens(req *RoutingRequest) int {
	if req.RequestBody == nil {
		return 0
	}

	if maxTokens, ok := req.RequestBody["max_tokens"].(float64); ok {
		return int(maxTokens)
	}

	return 0
}

func (s *TokenEstimationService) estimateInputFromRequestBody(requestBody map[string]interface{}) int {
	if requestBody == nil {
		return 0
	}

	var totalChars int

	if messages, ok := requestBody["messages"].([]interface{}); ok {
		for _, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				if content, ok := msgMap["content"].(string); ok {
					totalChars += len(content)
				}
			}
		}
	}

	if prompt, ok := requestBody["prompt"].(string); ok {
		totalChars += len(prompt)
	}

	if input, ok := requestBody["input"].(string); ok {
		totalChars += len(input)
	}

	if inputArray, ok := requestBody["input"].([]interface{}); ok {
		for _, item := range inputArray {
			if str, ok := item.(string); ok {
				totalChars += len(str)
			}
		}
	}

	return totalChars / 4
}

//nolint:unused // Reserved for future use with parsed Message slices
func (s *TokenEstimationService) estimateInputFromMessages(messages []Message) int {
	var totalChars int
	for _, msg := range messages {
		totalChars += len(msg.Content)
	}
	return totalChars / 4
}

func (s *TokenEstimationService) getModelStatistics(model string) (*ModelTokenStats, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT model_name, avg_input_tokens, avg_output_tokens,
		       p50_input_tokens, p50_output_tokens,
		       p90_input_tokens, p90_output_tokens,
		       input_output_ratio, total_requests
		FROM model_token_statistics
		WHERE model_name = $1
	`

	var stats ModelTokenStats
	err := s.db.QueryRow(query, model).Scan(
		&stats.ModelName,
		&stats.AvgInputTokens,
		&stats.AvgOutputTokens,
		&stats.P50InputTokens,
		&stats.P50OutputTokens,
		&stats.P90InputTokens,
		&stats.P90OutputTokens,
		&stats.InputOutputRatio,
		&stats.TotalRequests,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no statistics found for model: %s", model)
		}
		return nil, fmt.Errorf("failed to query model statistics: %w", err)
	}

	return &stats, nil
}

func (s *TokenEstimationService) getFallbackEstimation() *TokenEstimation {
	return &TokenEstimation{
		EstimatedInputTokens:  500,
		EstimatedOutputTokens: 500,
		Source:                "fallback",
	}
}
