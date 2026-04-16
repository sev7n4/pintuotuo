package services

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/pintuotuo/backend/logger"
)

type LLMTraceSpan struct {
	start      time.Time
	requestID  string
	userID     int
	provider   string
	model      string
	statusCode int
	errCode    string
}

func StartLLMTrace(requestID string, userID int) *LLMTraceSpan {
	return &LLMTraceSpan{
		start:     time.Now(),
		requestID: requestID,
		userID:    userID,
	}
}

func (s *LLMTraceSpan) SetRoute(provider, model string) {
	s.provider = provider
	s.model = model
}

func (s *LLMTraceSpan) SetStatusCode(statusCode int) {
	s.statusCode = statusCode
}

func (s *LLMTraceSpan) SetErrorCode(errCode string) {
	s.errCode = errCode
}

func (s *LLMTraceSpan) Finish(ctx context.Context) {
	if !langfuseEnabled() {
		return
	}
	fields := map[string]interface{}{
		"request_id": s.requestID,
		"user_id":    s.userID,
		"provider":   s.provider,
		"model":      s.model,
		"latency_ms": time.Since(s.start).Milliseconds(),
		"status":     s.statusCode,
		"error_code": s.errCode,
		"sink":       "langfuse-compatible-log",
	}
	logger.LogInfo(ctx, "llm_trace", "llm trace span finished", fields)
}

func langfuseEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("LANGFUSE_ENABLED")))
	return v == "1" || v == "true" || v == "on" || v == "yes"
}
