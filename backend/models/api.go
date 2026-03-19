package models

import "time"

type APIKeyPool struct {
	ID                 int       `json:"id"`
	MerchantID         int       `json:"merchant_id"`
	Name               string    `json:"name"`
	Provider           string    `json:"provider"`
	APIKeyEncrypted    string    `json:"-"`
	APISecretEncrypted string    `json:"-"`
	QuotaLimit         float64   `json:"quota_limit"`
	QuotaUsed          float64   `json:"quota_used"`
	Status             string    `json:"status"`
	LastUsedAt         time.Time `json:"last_used_at,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type APIRequest struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	KeyID        int       `json:"key_id"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Cost         float64   `json:"cost"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type APIUsageLog struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	KeyID        int       `json:"key_id"`
	RequestID    string    `json:"request_id"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	LatencyMs    int       `json:"latency_ms"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Cost         float64   `json:"cost"`
	CreatedAt    time.Time `json:"created_at"`
}
