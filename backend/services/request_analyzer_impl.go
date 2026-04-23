package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type IRequestAnalyzer interface {
	Analyze(ctx context.Context, req *http.Request, body []byte) (*RequestAnalysis, error)
	EstimateTokens(text string) int
	DetectIntent(body []byte) RequestIntent
	AssessComplexity(analysis *RequestAnalysis) RequestComplexity
}

type RequestAnalyzer struct {
	tokenEstimator *TokenEstimator
}

func NewRequestAnalyzer() *RequestAnalyzer {
	return &RequestAnalyzer{
		tokenEstimator: NewTokenEstimator(),
	}
}

func (a *RequestAnalyzer) Analyze(ctx context.Context, req *http.Request, body []byte) (*RequestAnalysis, error) {
	analysis := &RequestAnalysis{
		Timestamp: time.Now(),
	}

	intent := a.DetectIntent(body)
	analysis.Intent = intent

	var requestBody map[string]interface{}
	if err := json.Unmarshal(body, &requestBody); err == nil {
		if model, ok := requestBody["model"].(string); ok {
			analysis.Model = model
		}

		if stream, ok := requestBody["stream"].(bool); ok {
			analysis.Stream = stream
		}

		if temp, ok := requestBody["temperature"].(float64); ok {
			analysis.Temperature = &temp
		}

		if maxTokens, ok := requestBody["max_tokens"].(float64); ok {
			maxTokensInt := int(maxTokens)
			analysis.MaxTokens = &maxTokensInt
		}

		promptText := a.extractPromptText(requestBody)
		analysis.PromptLength = len(promptText)
		analysis.EstimatedTokens = a.EstimateTokens(promptText)
	}

	analysis.Complexity = a.AssessComplexity(analysis)

	return analysis, nil
}

func (a *RequestAnalyzer) EstimateTokens(text string) int {
	return a.tokenEstimator.Estimate(text)
}

func (a *RequestAnalyzer) DetectIntent(body []byte) RequestIntent {
	var requestBody map[string]interface{}
	if err := json.Unmarshal(body, &requestBody); err != nil {
		return IntentUnknown
	}

	if _, ok := requestBody["messages"]; ok {
		return IntentChat
	}

	if _, ok := requestBody["prompt"]; ok {
		return IntentCompletion
	}

	if _, ok := requestBody["input"]; ok {
		if model, ok := requestBody["model"].(string); ok {
			if strings.Contains(model, "embed") {
				return IntentEmbedding
			}
		}
		return IntentCompletion
	}

	if _, ok := requestBody["image"]; ok {
		return IntentImage
	}

	if _, ok := requestBody["audio"]; ok {
		return IntentAudio
	}

	if _, ok := requestBody["moderation"]; ok {
		return IntentModeration
	}

	return IntentUnknown
}

func (a *RequestAnalyzer) AssessComplexity(analysis *RequestAnalysis) RequestComplexity {
	score := 0

	if analysis.EstimatedTokens > 4000 {
		score += 3
	} else if analysis.EstimatedTokens > 2000 {
		score += 2
	} else if analysis.EstimatedTokens > 1000 {
		score += 1
	}

	if analysis.Stream {
		score += 1
	}

	if analysis.MaxTokens != nil && *analysis.MaxTokens > 2000 {
		score += 1
	}

	if analysis.Intent == IntentChat {
		score += 1
	}

	if score >= 4 {
		return ComplexityComplex
	} else if score >= 2 {
		return ComplexityMedium
	}
	return ComplexitySimple
}

func (a *RequestAnalyzer) extractPromptText(requestBody map[string]interface{}) string {
	var textBuilder strings.Builder

	if messages, ok := requestBody["messages"].([]interface{}); ok {
		for _, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				if content, ok := msgMap["content"].(string); ok {
					textBuilder.WriteString(content)
					textBuilder.WriteString(" ")
				}
			}
		}
	}

	if prompt, ok := requestBody["prompt"].(string); ok {
		textBuilder.WriteString(prompt)
	}

	if input, ok := requestBody["input"].(string); ok {
		textBuilder.WriteString(input)
	}

	if inputArray, ok := requestBody["input"].([]interface{}); ok {
		for _, item := range inputArray {
			if str, ok := item.(string); ok {
				textBuilder.WriteString(str)
				textBuilder.WriteString(" ")
			}
		}
	}

	return textBuilder.String()
}

type TokenEstimator struct {
	avgCharsPerToken float64
	chineseRegex     *regexp.Regexp
}

func NewTokenEstimator() *TokenEstimator {
	return &TokenEstimator{
		avgCharsPerToken: 4.0,
		chineseRegex:     regexp.MustCompile(`[\p{Han}]`),
	}
}

func (e *TokenEstimator) Estimate(text string) int {
	if len(text) == 0 {
		return 0
	}

	chineseCount := len(e.chineseRegex.FindAllString(text, -1))
	nonChineseText := e.chineseRegex.ReplaceAllString(text, "")
	nonChineseCount := len(nonChineseText)

	chineseTokens := float64(chineseCount) / 1.5
	nonChineseTokens := float64(nonChineseCount) / e.avgCharsPerToken

	totalTokens := int(chineseTokens + nonChineseTokens)

	if totalTokens < 1 {
		return 1
	}

	return totalTokens
}

func ReadRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return []byte{}, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}
