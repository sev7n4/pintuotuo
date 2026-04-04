package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

type HealthCheckLevel string

const (
	HealthCheckLevelHigh   HealthCheckLevel = "high"
	HealthCheckLevelMedium HealthCheckLevel = "medium"
	HealthCheckLevelLow    HealthCheckLevel = "low"
	HealthCheckLevelDaily  HealthCheckLevel = "daily"
)

const (
	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"
)

var healthCheckIntervalMap = map[HealthCheckLevel]int{
	HealthCheckLevelHigh:   60,
	HealthCheckLevelMedium: 300,
	HealthCheckLevelLow:    1800,
	HealthCheckLevelDaily:  86400,
}

type ProviderHealth struct {
	APIKeyID         int                    `json:"api_key_id"`
	Provider         string                 `json:"provider"`
	Status           string                 `json:"status"`
	LatencyMs        int                    `json:"latency_ms"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	ModelsAvailable  []string               `json:"models_available,omitempty"`
	PricingInfo      map[string]interface{} `json:"pricing_info,omitempty"`
	LastCheckedAt    time.Time              `json:"last_checked_at"`
	FailureCount     int                    `json:"failure_count"`
	ConsecutiveCount int                    `json:"consecutive_count"`
}

type HealthCheckResult struct {
	Success      bool
	Status       string
	LatencyMs    int
	ErrorMessage string
	ModelsFound  []string
	PricingInfo  map[string]interface{}
	CheckType    string
}

type HealthChecker struct {
	httpClient *http.Client
	db         *sql.DB
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		db: config.GetDB(),
	}
}

func (s *HealthChecker) GetHealthCheckInterval(level string) int {
	levelEnum := HealthCheckLevel(level)
	if interval, ok := healthCheckIntervalMap[levelEnum]; ok {
		return interval
	}
	return 300
}

func (s *HealthChecker) LightweightPing(ctx context.Context, apiKey *models.MerchantAPIKey) (*HealthCheckResult, error) {
	endpoint := apiKey.EndpointURL
	if endpoint == "" {
		endpoint = s.getDefaultEndpoint(apiKey.Provider)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
			CheckType:    "lightweight",
		}, nil
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.getDecryptedAPIKey(apiKey)))
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := s.httpClient.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("connection failed: %v", err),
			CheckType:    "lightweight",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &HealthCheckResult{
			Success:   true,
			Status:    HealthStatusHealthy,
			LatencyMs: latencyMs,
			CheckType: "lightweight",
		}, nil
	}

	return &HealthCheckResult{
		Success:      false,
		Status:       HealthStatusDegraded,
		LatencyMs:    latencyMs,
		ErrorMessage: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		CheckType:    "lightweight",
	}, nil
}

func (s *HealthChecker) FullVerification(ctx context.Context, apiKey *models.MerchantAPIKey) (*HealthCheckResult, error) {
	endpoint := apiKey.EndpointURL
	if endpoint == "" {
		endpoint = s.getDefaultEndpoint(apiKey.Provider)
	}

	modelsEndpoint := endpoint + "/v1/models"

	req, err := http.NewRequestWithContext(ctx, "GET", modelsEndpoint, nil)
	if err != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
			CheckType:    "full",
		}, nil
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.getDecryptedAPIKey(apiKey)))
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := s.httpClient.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("connection failed: %v", err),
			CheckType:    "full",
		}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var modelsResp struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &modelsResp); err != nil {
			return &HealthCheckResult{
				Success:      false,
				Status:       HealthStatusDegraded,
				LatencyMs:    latencyMs,
				ErrorMessage: fmt.Sprintf("failed to parse models: %v", err),
				CheckType:    "full",
			}, nil
		}

		models := make([]string, 0, len(modelsResp.Data))
		for _, m := range modelsResp.Data {
			models = append(models, m.ID)
		}

		return &HealthCheckResult{
			Success:     true,
			Status:      HealthStatusHealthy,
			LatencyMs:   latencyMs,
			ModelsFound: models,
			PricingInfo: s.extractPricingInfo(apiKey.Provider),
			CheckType:   "full",
		}, nil
	}

	return &HealthCheckResult{
		Success:      false,
		Status:       HealthStatusUnhealthy,
		LatencyMs:    latencyMs,
		ErrorMessage: fmt.Sprintf("verification failed with status: %d, body: %s", resp.StatusCode, string(body)),
		CheckType:    "full",
	}, nil
}

func (s *HealthChecker) RecordRequestResult(ctx context.Context, apiKeyID int, isSuccess bool, latencyMs int) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	if isSuccess {
		_, err := db.ExecContext(ctx, `
			UPDATE merchant_api_keys 
			SET consecutive_failures = 0, 
			    health_status = CASE WHEN health_status IN ('unhealthy', 'degraded') THEN 'healthy' ELSE health_status END,
			    last_health_check_at = CURRENT_TIMESTAMP
			WHERE id = $1`,
			apiKeyID,
		)
		return err
	}

	_, err := db.ExecContext(ctx, `
		UPDATE merchant_api_keys 
		SET consecutive_failures = consecutive_failures + 1,
		    last_health_check_at = CURRENT_TIMESTAMP
		WHERE id = $1`,
		apiKeyID,
	)
	if err != nil {
		return err
	}

	var consecutiveFailures int
	err = db.QueryRowContext(ctx, `SELECT consecutive_failures FROM merchant_api_keys WHERE id = $1`, apiKeyID).Scan(&consecutiveFailures)
	if err != nil {
		return err
	}

	if consecutiveFailures >= 5 {
		_, err = db.ExecContext(ctx, `
			UPDATE merchant_api_keys SET health_status = 'unhealthy' WHERE id = $1`,
			apiKeyID,
		)
		return err
	}

	if consecutiveFailures >= 2 {
		_, err = db.ExecContext(ctx, `
			UPDATE merchant_api_keys SET health_status = 'degraded' WHERE id = $1`,
			apiKeyID,
		)
	}

	return err
}

func (s *HealthChecker) CalculateFailureRate(ctx context.Context, apiKeyID int, windowMinutes int) (float64, error) {
	db := config.GetDB()
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	var totalRequests, failedRequests int
	err := db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status != 'healthy') as failed
		FROM api_key_health_history 
		WHERE api_key_id = $1 
		  AND created_at > NOW() - INTERVAL '1 minute' * $2`,
		apiKeyID, windowMinutes,
	).Scan(&totalRequests, &failedRequests)

	if err != nil {
		return 0, err
	}

	if totalRequests == 0 {
		return 0, nil
	}

	return float64(failedRequests) / float64(totalRequests) * 100, nil
}

func (s *HealthChecker) MarkAsDegraded(ctx context.Context, apiKeyID int) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.ExecContext(ctx, `
		UPDATE merchant_api_keys 
		SET health_status = 'degraded', 
		    last_health_check_at = CURRENT_TIMESTAMP
		WHERE id = $1`,
		apiKeyID,
	)
	return err
}

func (s *HealthChecker) GetProviderHealth(ctx context.Context, apiKeyID int) (*ProviderHealth, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var key models.MerchantAPIKey
	err := db.QueryRowContext(ctx, `
		SELECT id, merchant_id, provider, health_status, health_check_level, 
		       last_health_check_at, consecutive_failures, endpoint_url
		FROM merchant_api_keys WHERE id = $1`,
		apiKeyID,
	).Scan(&key.ID, &key.MerchantID, &key.Provider, &key.HealthStatus, &key.HealthCheckLevel,
		&key.LastHealthCheckAt, &key.ConsecutiveFailures, &key.EndpointURL)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("api key not found: %d", apiKeyID)
	}
	if err != nil {
		return nil, err
	}

	var lastCheckAt time.Time
	if key.LastHealthCheckAt != nil {
		lastCheckAt = *key.LastHealthCheckAt
	}

	return &ProviderHealth{
		APIKeyID:         key.ID,
		Provider:         key.Provider,
		Status:           key.HealthStatus,
		LastCheckedAt:    lastCheckAt,
		ConsecutiveCount: key.ConsecutiveFailures,
	}, nil
}

func (s *HealthChecker) GetAllProviderHealth(ctx context.Context) ([]ProviderHealth, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, merchant_id, provider, health_status, health_check_level, 
		       last_health_check_at, consecutive_failures, endpoint_url
		FROM merchant_api_keys 
		WHERE status = 'active'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ProviderHealth
	for rows.Next() {
		var key models.MerchantAPIKey
		err := rows.Scan(&key.ID, &key.MerchantID, &key.Provider, &key.HealthStatus, &key.HealthCheckLevel,
			&key.LastHealthCheckAt, &key.ConsecutiveFailures, &key.EndpointURL)
		if err != nil {
			continue
		}

		var lastCheckAt time.Time
		if key.LastHealthCheckAt != nil {
			lastCheckAt = *key.LastHealthCheckAt
		}

		results = append(results, ProviderHealth{
			APIKeyID:         key.ID,
			Provider:         key.Provider,
			Status:           key.HealthStatus,
			LastCheckedAt:    lastCheckAt,
			ConsecutiveCount: key.ConsecutiveFailures,
		})
	}

	return results, nil
}

func (s *HealthChecker) SaveHealthCheckResult(ctx context.Context, apiKeyID int, result *HealthCheckResult) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	modelsJSON, _ := json.Marshal(result.ModelsFound)
	pricingJSON, _ := json.Marshal(result.PricingInfo)

	_, err := db.ExecContext(ctx, `
		INSERT INTO api_key_health_history 
		(api_key_id, check_type, status, latency_ms, error_message, models_available, pricing_info)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		apiKeyID, result.CheckType, result.Status, result.LatencyMs, result.ErrorMessage,
		modelsJSON, pricingJSON,
	)
	if err != nil {
		return err
	}

	statusToUpdate := result.Status
	if result.Status == HealthStatusHealthy && result.CheckType == "passive" {
		statusToUpdate = "healthy"
	}

	_, err = db.ExecContext(ctx, `
		UPDATE merchant_api_keys 
		SET health_status = $1,
		    last_health_check_at = CURRENT_TIMESTAMP,
		    consecutive_failures = CASE WHEN $2 = 'healthy' THEN 0 ELSE consecutive_failures END
		WHERE id = $3`,
		statusToUpdate, result.Status, apiKeyID,
	)

	return err
}

func (s *HealthChecker) getDefaultEndpoint(provider string) string {
	endpoints := map[string]string{
		"openai":    "https://api.openai.com",
		"anthropic": "https://api.anthropic.com",
		"google":    "https://generativelanguage.googleapis.com",
		"azure":     "https://{resource}.openai.azure.com",
		"custom":    "http://localhost:8080",
	}
	if ep, ok := endpoints[provider]; ok {
		return ep
	}
	return endpoints["openai"]
}

func (s *HealthChecker) getDecryptedAPIKey(apiKey *models.MerchantAPIKey) string {
	return "dummy-key-for-health-check"
}

func (s *HealthChecker) extractPricingInfo(provider string) map[string]interface{} {
	return map[string]interface{}{
		"provider": provider,
		"updated":  time.Now().Format(time.RFC3339),
	}
}

func (s *HealthChecker) ShouldPerformCheck(apiKey *models.MerchantAPIKey) bool {
	if apiKey.LastHealthCheckAt == nil {
		return true
	}

	interval := s.GetHealthCheckInterval(apiKey.HealthCheckLevel)
	elapsed := time.Since(*apiKey.LastHealthCheckAt)

	return elapsed.Seconds() >= float64(interval)
}

func (s *HealthChecker) TriggerActiveCheck(ctx context.Context, apiKeyID int) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	var key models.MerchantAPIKey
	err := db.QueryRowContext(ctx, `
		SELECT id, merchant_id, provider, api_key_encrypted, endpoint_url, health_check_level
		FROM merchant_api_keys WHERE id = $1`,
		apiKeyID,
	).Scan(&key.ID, &key.MerchantID, &key.Provider, &key.APIKeyEncrypted, &key.EndpointURL, &key.HealthCheckLevel)

	if err != nil {
		return err
	}

	result, err := s.FullVerification(ctx, &key)
	if err != nil {
		return err
	}

	return s.SaveHealthCheckResult(ctx, apiKeyID, result)
}

func PerformHealthCheckAsync(apiKeyID int) {
	ctx := context.Background()
	checker := NewHealthChecker()

	var key models.MerchantAPIKey
	db := config.GetDB()
	if db == nil {
		return
	}

	err := db.QueryRowContext(ctx, `
		SELECT id, merchant_id, provider, api_key_encrypted, endpoint_url, health_check_level, health_status
		FROM merchant_api_keys WHERE id = $1`,
		apiKeyID,
	).Scan(&key.ID, &key.MerchantID, &key.Provider, &key.APIKeyEncrypted, &key.EndpointURL, &key.HealthCheckLevel, &key.HealthStatus)

	if err != nil {
		return
	}

	if !checker.ShouldPerformCheck(&key) {
		return
	}

	var result *HealthCheckResult
	level := HealthCheckLevel(key.HealthCheckLevel)

	if level == HealthCheckLevelHigh {
		result, _ = checker.LightweightPing(ctx, &key)
	} else {
		result, _ = checker.FullVerification(ctx, &key)
	}

	if result != nil {
		checker.SaveHealthCheckResult(ctx, apiKeyID, result)
	}
}

func IsHealthy(status string) bool {
	return status == HealthStatusHealthy
}

func IsDegraded(status string) bool {
	return status == HealthStatusDegraded
}

func IsUnhealthy(status string) bool {
	return status == HealthStatusUnhealthy
}

type TestChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	MaxTokens int `json:"max_tokens"`
}

func (s *HealthChecker) TestChatCompletion(ctx context.Context, apiKey *models.MerchantAPIKey, model string) (*HealthCheckResult, error) {
	endpoint := apiKey.EndpointURL
	if endpoint == "" {
		endpoint = s.getDefaultEndpoint(apiKey.Provider)
	}

	chatEndpoint := endpoint + "/v1/chat/completions"

	testReq := TestChatRequest{
		Model: model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 5,
	}

	body, _ := json.Marshal(testReq)
	req, err := http.NewRequestWithContext(ctx, "POST", chatEndpoint, bytes.NewReader(body))
	if err != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			ErrorMessage: fmt.Sprintf("failed to create request: %v", err),
			CheckType:    "chat_test",
		}, nil
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.getDecryptedAPIKey(apiKey)))
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := s.httpClient.Do(req)
	latencyMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return &HealthCheckResult{
			Success:      false,
			Status:       HealthStatusUnhealthy,
			LatencyMs:    latencyMs,
			ErrorMessage: fmt.Sprintf("connection failed: %v", err),
			CheckType:    "chat_test",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &HealthCheckResult{
			Success:   true,
			Status:    HealthStatusHealthy,
			LatencyMs: latencyMs,
			CheckType: "chat_test",
		}, nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return &HealthCheckResult{
		Success:      false,
		Status:       HealthStatusDegraded,
		LatencyMs:    latencyMs,
		ErrorMessage: fmt.Sprintf("status: %d, body: %s", resp.StatusCode, string(respBody)),
		CheckType:    "chat_test",
	}, nil
}
