package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type StoredResponse struct {
	ID              int             `json:"id"`
	ResponseID      string          `json:"response_id"`
	UserID          int             `json:"user_id"`
	MerchantID      int             `json:"merchant_id"`
	Model           string          `json:"model"`
	Input           json.RawMessage `json:"input"`
	Output          json.RawMessage `json:"output,omitempty"`
	ToolCalls       json.RawMessage `json:"tool_calls,omitempty"`
	Usage           json.RawMessage `json:"usage,omitempty"`
	Status          string          `json:"status"`
	BackgroundJobID sql.NullString  `json:"background_job_id,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	ExpiresAt       time.Time       `json:"expires_at"`
	DeletedAt       sql.NullTime    `json:"deleted_at,omitempty"`
}

type ResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type BackgroundJobStatus struct {
	JobID      string `json:"job_id"`
	Status     string `json:"status"`
	ResponseID string `json:"response_id,omitempty"`
	Error      string `json:"error,omitempty"`
}
