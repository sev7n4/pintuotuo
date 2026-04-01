package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/utils"
)

const providerAnthropic = "anthropic"

type providerRuntimeConfig struct {
	Code       string
	Name       string
	APIBaseURL string
	APIFormat  string
}

func legacyProviderList() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":         "openai",
			"display_name": "OpenAI",
			"api_format":   "openai",
		},
		{
			"name":         "anthropic",
			"display_name": "Anthropic",
			"api_format":   "anthropic",
		},
		{
			"name":         "google",
			"display_name": "Google AI",
			"api_format":   "openai",
		},
	}
}

type APIProxyRequest struct {
	Provider      string          `json:"provider" binding:"required"`
	Model         string          `json:"model" binding:"required"`
	Messages      []ChatMessage   `json:"messages" binding:"required"`
	Stream        bool            `json:"stream"`
	Options       json.RawMessage `json:"options,omitempty"`
	APIKeyID      *int            `json:"api_key_id,omitempty"`
	MerchantSKUID *int            `json:"merchant_sku_id,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIProxyResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []APIChoice `json:"choices"`
	Usage   APIUsage    `json:"usage"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIChoice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason"`
}

type APIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func ProxyAPIRequest(c *gin.Context) {
	startTime := time.Now()
	requestID := uuid.New().String()

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	var req APIProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	proxyAPIRequestCore(c, userIDInt, requestID, startTime, req, c.Request.URL.Path)
}

func proxyAPIRequestCore(c *gin.Context, userIDInt int, requestID string, startTime time.Time, req APIProxyRequest, requestPath string) {
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var tokenBalance float64
	err := db.QueryRow("SELECT balance FROM tokens WHERE user_id = $1", userIDInt).Scan(&tokenBalance)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrTokenNotFound)
		return
	}

	if tokenBalance <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInsufficientBalance)
		return
	}

	var apiKey models.MerchantAPIKey
	err = selectAPIKeyForRequest(db, req, &apiKey)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_KEY_NOT_FOUND",
			"No available API key for this provider",
			http.StatusServiceUnavailable,
			err,
		))
		return
	}

	decryptedKey, err := utils.Decrypt(apiKey.APIKeyEncrypted)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"DECRYPTION_FAILED",
			"Failed to decrypt API key",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	providerCfg, err := getProviderRuntimeConfig(db, req.Provider)
	if err == sql.ErrNoRows {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"UNSUPPORTED_PROVIDER",
			fmt.Sprintf("Provider %s is not supported", req.Provider),
			http.StatusBadRequest,
			nil,
		))
		return
	}
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	baseURL := strings.TrimRight(providerCfg.APIBaseURL, "/")
	if baseURL == "" {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"UNSUPPORTED_PROVIDER",
			fmt.Sprintf("Provider %s is missing api_base_url", req.Provider),
			http.StatusBadRequest,
			nil,
		))
		return
	}

	endpoint := fmt.Sprintf("%s/chat/completions", baseURL)
	if providerCfg.APIFormat == providerAnthropic {
		endpoint = fmt.Sprintf("%s/messages", baseURL)
	}

	requestBody := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	}

	if req.Options != nil {
		var options map[string]interface{}
		if unmarshalErr := json.Unmarshal(req.Options, &options); unmarshalErr == nil {
			for k, v := range options {
				requestBody[k] = v
			}
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"REQUEST_BUILD_FAILED",
			"Failed to build request body",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"REQUEST_CREATE_FAILED",
			"Failed to create request",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	switch providerCfg.APIFormat {
	case providerAnthropic:
		httpReq.Header.Set("x-api-key", decryptedKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	default:
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", decryptedKey))
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"API_REQUEST_FAILED",
			"Failed to send request to provider",
			http.StatusBadGateway,
			err,
		))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"RESPONSE_READ_FAILED",
			"Failed to read response",
			http.StatusInternalServerError,
			err,
		))
		return
	}

	var apiResp APIProxyResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		c.Data(resp.StatusCode, "application/json", body)
		return
	}

	latency := int(time.Since(startTime).Milliseconds())

	var inputTokens, outputTokens int
	var cost float64

	if apiResp.Usage.TotalTokens > 0 {
		inputTokens = apiResp.Usage.PromptTokens
		outputTokens = apiResp.Usage.CompletionTokens
		cost = calculateTokenCost(req.Provider, req.Model, inputTokens, outputTokens)
	}

	if cost > 0 {
		tx, err := db.Begin()
		if err == nil {
			_, err = tx.Exec(
				"UPDATE tokens SET balance = balance - $1, total_used = total_used + $1 WHERE user_id = $2",
				cost, userIDInt,
			)
			if err == nil {
				_, err = tx.Exec(
					"INSERT INTO token_transactions (user_id, type, amount, reason) VALUES ($1, $2, $3, $4)",
					userIDInt, "usage", -cost, fmt.Sprintf("API call: %s/%s", req.Provider, req.Model),
				)
			}
			if err == nil {
				_, err = tx.Exec(
					"UPDATE merchant_api_keys SET quota_used = quota_used + $1, last_used_at = $2 WHERE id = $3",
					cost, time.Now(), apiKey.ID,
				)
			}
			if err == nil {
				_, err = tx.Exec(
					"INSERT INTO api_usage_logs (user_id, key_id, request_id, provider, model, method, path, status_code, latency_ms, input_tokens, output_tokens, cost) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
					userIDInt, apiKey.ID, requestID, req.Provider, req.Model, "POST", requestPath, resp.StatusCode, latency, inputTokens, outputTokens, cost,
				)
			}

			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}

		ctx := context.Background()
		cache.Delete(ctx, cache.TokenBalanceKey(userIDInt))
	}

	c.Data(resp.StatusCode, "application/json", body)
}

func calculateTokenCost(provider, model string, inputTokens, outputTokens int) float64 {
	var inputRate, outputRate float64

	switch provider {
	case "openai":
		switch {
		case strings.Contains(model, "gpt-4-turbo"):
			inputRate = 0.01 / 1000
			outputRate = 0.03 / 1000
		case strings.Contains(model, "gpt-4"):
			inputRate = 0.03 / 1000
			outputRate = 0.06 / 1000
		case strings.Contains(model, "gpt-3.5-turbo"):
			inputRate = 0.0005 / 1000
			outputRate = 0.0015 / 1000
		default:
			inputRate = 0.001 / 1000
			outputRate = 0.002 / 1000
		}
	case providerAnthropic:
		switch {
		case strings.Contains(model, "claude-3-opus"):
			inputRate = 0.015 / 1000
			outputRate = 0.075 / 1000
		case strings.Contains(model, "claude-3-sonnet"):
			inputRate = 0.003 / 1000
			outputRate = 0.015 / 1000
		case strings.Contains(model, "claude-3-haiku"):
			inputRate = 0.00025 / 1000
			outputRate = 0.00125 / 1000
		default:
			inputRate = 0.003 / 1000
			outputRate = 0.015 / 1000
		}
	case "google":
		inputRate = 0.00025 / 1000
		outputRate = 0.0005 / 1000
	default:
		inputRate = 0.001 / 1000
		outputRate = 0.002 / 1000
	}

	return float64(inputTokens)*inputRate + float64(outputTokens)*outputRate
}

func getProviderRuntimeConfig(db *sql.DB, providerCode string) (providerRuntimeConfig, error) {
	var cfg providerRuntimeConfig
	err := db.QueryRow(
		`SELECT code, name, COALESCE(api_base_url, ''), api_format
		 FROM model_providers
		 WHERE code = $1 AND status = 'active'
		 LIMIT 1`,
		providerCode,
	).Scan(&cfg.Code, &cfg.Name, &cfg.APIBaseURL, &cfg.APIFormat)
	return cfg, err
}

func selectAPIKeyForRequest(db *sql.DB, req APIProxyRequest, apiKey *models.MerchantAPIKey) error {
	if req.APIKeyID != nil && *req.APIKeyID > 0 {
		err := db.QueryRow(
			`SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status
			 FROM merchant_api_keys
			 WHERE id = $1 AND provider = $2 AND status = 'active'`,
			*req.APIKeyID, req.Provider,
		).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.APISecretEncrypted, &apiKey.QuotaLimit, &apiKey.QuotaUsed, &apiKey.Status)
		if err == nil {
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
	}

	if req.MerchantSKUID != nil && *req.MerchantSKUID > 0 {
		err := db.QueryRow(
			`SELECT mak.id, mak.merchant_id, mak.provider, mak.api_key_encrypted, mak.api_secret_encrypted, mak.quota_limit, mak.quota_used, mak.status
			 FROM merchant_skus ms
			 JOIN merchant_api_keys mak ON mak.id = ms.api_key_id
			 WHERE ms.id = $1 AND ms.status = 'active' AND mak.provider = $2 AND mak.status = 'active'`,
			*req.MerchantSKUID, req.Provider,
		).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.APISecretEncrypted, &apiKey.QuotaLimit, &apiKey.QuotaUsed, &apiKey.Status)
		if err == nil {
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}
	}

	return db.QueryRow(
		`SELECT id, merchant_id, provider, api_key_encrypted, api_secret_encrypted, quota_limit, quota_used, status
		 FROM merchant_api_keys
		 WHERE provider = $1 AND status = 'active'
		 ORDER BY (quota_limit - quota_used) DESC
		 LIMIT 1`,
		req.Provider,
	).Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Provider, &apiKey.APIKeyEncrypted, &apiKey.APISecretEncrypted, &apiKey.QuotaLimit, &apiKey.QuotaUsed, &apiKey.Status)
}

func GetAPIProviders(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusOK, legacyProviderList())
		return
	}

	rows, err := db.Query(
		`SELECT code, name, api_format FROM model_providers WHERE status = 'active' ORDER BY sort_order ASC`,
	)
	if err != nil {
		c.JSON(http.StatusOK, legacyProviderList())
		return
	}
	defer rows.Close()

	providers := make([]map[string]interface{}, 0)
	for rows.Next() {
		var code, name, apiFormat string
		if scanErr := rows.Scan(&code, &name, &apiFormat); scanErr != nil {
			continue
		}
		providers = append(providers, map[string]interface{}{
			"name":         code,
			"display_name": name,
			"api_format":   apiFormat,
		})
	}
	if len(providers) == 0 {
		providers = legacyProviderList()
	}

	c.JSON(http.StatusOK, providers)
}

func GetAPIUsageStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var stats struct {
		TotalRequests int     `json:"total_requests"`
		TotalTokens   int     `json:"total_tokens"`
		TotalCost     float64 `json:"total_cost"`
		AvgLatencyMs  int     `json:"avg_latency_ms"`
	}

	db.QueryRow(
		"SELECT COUNT(*), COALESCE(SUM(input_tokens + output_tokens), 0), COALESCE(SUM(cost), 0), COALESCE(AVG(latency_ms), 0) FROM api_usage_logs WHERE user_id = $1",
		userIDInt,
	).Scan(&stats.TotalRequests, &stats.TotalTokens, &stats.TotalCost, &stats.AvgLatencyMs)

	rows, err := db.Query(
		"SELECT provider, COUNT(*) as count, SUM(cost) as cost FROM api_usage_logs WHERE user_id = $1 GROUP BY provider ORDER BY count DESC",
		userIDInt,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	byProvider := make([]map[string]interface{}, 0)
	for rows.Next() {
		var provider string
		var count int
		var cost float64
		if err := rows.Scan(&provider, &count, &cost); err != nil {
			continue
		}
		byProvider = append(byProvider, map[string]interface{}{
			"provider": provider,
			"count":    count,
			"cost":     cost,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":       stats,
		"by_provider": byProvider,
	})
}
