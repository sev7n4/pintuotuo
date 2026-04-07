package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/metrics"
	"github.com/pintuotuo/backend/utils"
	"github.com/redis/go-redis/v9"
)

type VerificationResult struct {
	ID                int                    `json:"id"`
	APIKeyID          int                    `json:"api_key_id"`
	VerificationType  string                 `json:"verification_type"`
	Status            string                 `json:"status"`
	ConnectionTest    bool                   `json:"connection_test"`
	ConnectionLatency int                    `json:"connection_latency_ms"`
	ModelsFound       []string               `json:"models_found"`
	ModelsCount       int                    `json:"models_count"`
	PricingVerified   bool                   `json:"pricing_verified"`
	PricingInfo       map[string]interface{} `json:"pricing_info,omitempty"`
	ErrorCode         string                 `json:"error_code,omitempty"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
	StartedAt         time.Time              `json:"started_at"`
	CompletedAt       time.Time              `json:"completed_at,omitempty"`
	RetryCount        int                    `json:"retry_count,omitempty"`
}

type APIKeyValidator struct {
	dbMu            sync.Mutex
	db              *sql.DB
	redis           *redis.Client
	maxRetries      int
	retryDelay      time.Duration
	cacheTTL        time.Duration
	verificationTTL time.Duration
}

const (
	VerificationCacheKeyPrefix = "api_key_verification:"
	VerificationCacheTTL       = 5 * time.Minute
	MaxVerificationRetries     = 3
	RetryDelayBase             = 2 * time.Second
	VerificationInterval       = 24 * time.Hour
)

var (
	apiKeyValidator     *APIKeyValidator
	apiKeyValidatorOnce sync.Once
)

func GetAPIKeyValidator() *APIKeyValidator {
	apiKeyValidatorOnce.Do(func() {
		apiKeyValidator = &APIKeyValidator{
			db:              config.GetDB(),
			redis:           cache.GetClient(),
			maxRetries:      MaxVerificationRetries,
			retryDelay:      RetryDelayBase,
			cacheTTL:        VerificationCacheTTL,
			verificationTTL: VerificationInterval,
		}
	})
	return apiKeyValidator
}

// ensureDB lazily binds config.GetDB(); safe under concurrent tests and async validation.
func (v *APIKeyValidator) ensureDB() *sql.DB {
	v.dbMu.Lock()
	defer v.dbMu.Unlock()
	if v.db == nil {
		v.db = config.GetDB()
	}
	return v.db
}

func (v *APIKeyValidator) cacheVerificationResult(ctx context.Context, result VerificationResult) error {
	if v.redis == nil {
		return nil
	}

	key := fmt.Sprintf("%s%d", VerificationCacheKeyPrefix, result.APIKeyID)
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return v.redis.Set(ctx, key, data, v.cacheTTL).Err()
}

func (v *APIKeyValidator) performVerificationWithRetry(apiKeyID int, provider, encryptedKey, verificationType string, retryCount int) {
	ctx := context.Background()
	startTime := time.Now()

	metrics.ActiveVerifications.WithLabelValues(provider).Inc()
	defer metrics.ActiveVerifications.WithLabelValues(provider).Dec()

	result := VerificationResult{
		APIKeyID:         apiKeyID,
		VerificationType: verificationType,
		Status:           "in_progress",
		StartedAt:        startTime,
		RetryCount:       retryCount,
	}

	verificationID, err := v.createVerificationRecord(apiKeyID, verificationType)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to create verification record", err, map[string]interface{}{
			"api_key_id": apiKeyID,
		})
		return
	}
	result.ID = verificationID

	decryptedKey, err := utils.Decrypt(encryptedKey)
	if err != nil {
		v.handleVerificationError(ctx, verificationID, apiKeyID, "DECRYPTION_FAILED", "Failed to decrypt API key", startTime)
		return
	}

	providerConfig, err := v.getProviderConfig(provider)
	if err != nil {
		v.handleVerificationError(ctx, verificationID, apiKeyID, "PROVIDER_NOT_FOUND", "Provider configuration not found", startTime)
		return
	}

	connectionOK, latency, providerErrCode, providerErrMsg, err := v.testConnection(providerConfig, decryptedKey)
	if err != nil {
		if retryCount < v.maxRetries {
			metrics.VerificationRetries.WithLabelValues(provider, fmt.Sprintf("%d", retryCount+1)).Inc()
			time.Sleep(v.retryDelay * time.Duration(1<<retryCount))
			go v.performVerificationWithRetry(apiKeyID, provider, encryptedKey, verificationType, retryCount+1)
			return
		}
		msg := err.Error()
		if providerErrMsg != "" {
			msg = fmt.Sprintf("%s | provider: %s", msg, providerErrMsg)
		}
		code := "CONNECTION_FAILED"
		if providerErrCode != "" {
			code = providerErrCode
		}
		v.handleVerificationError(ctx, verificationID, apiKeyID, code, msg, startTime)
		return
	}
	result.ConnectionTest = connectionOK
	result.ConnectionLatency = latency
	metrics.VerificationConnectionLatency.WithLabelValues(provider).Observe(float64(latency))

	models, err := v.fetchModels(providerConfig, decryptedKey)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to fetch models", err, map[string]interface{}{
			"api_key_id": apiKeyID,
			"provider":   provider,
		})
	} else {
		result.ModelsFound = models
		result.ModelsCount = len(models)
		metrics.ModelsDiscovered.WithLabelValues(provider).Add(float64(len(models)))
	}

	if isDeepVerification(verificationType) {
		probeSupported := isQuotaProbeSupported(provider)
		result.PricingVerified = probeSupported
		if probeSupported {
			quotaOK, quotaCode, quotaMsg := v.probeQuota(providerConfig, provider, decryptedKey, models)
			if !quotaOK {
				v.handleVerificationError(ctx, verificationID, apiKeyID, quotaCode, quotaMsg, startTime)
				return
			}
		}
	}

	result.Status = "success"
	result.CompletedAt = time.Now()

	duration := result.CompletedAt.Sub(startTime).Seconds()
	metrics.VerificationDuration.WithLabelValues(provider, verificationType).Observe(duration)
	metrics.VerificationTotal.WithLabelValues(provider, verificationType, "success").Inc()

	err = v.updateVerificationRecord(verificationID, result)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to update verification record", err, map[string]interface{}{
			"verification_id": verificationID,
		})
		return
	}

	err = v.updateAPIKeyVerificationStatus(apiKeyID, result)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to update API key verification status", err, map[string]interface{}{
			"api_key_id": apiKeyID,
		})
		return
	}

	err = v.cacheVerificationResult(ctx, result)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to cache verification result", err, map[string]interface{}{
			"api_key_id": apiKeyID,
		})
	}

	logger.LogInfo(ctx, "api_key_validator", "API key verification completed", map[string]interface{}{
		"api_key_id":         apiKeyID,
		"verification_id":    verificationID,
		"status":             result.Status,
		"connection_test":    result.ConnectionTest,
		"models_count":       result.ModelsCount,
		"connection_latency": result.ConnectionLatency,
		"retry_count":        retryCount,
	})
}

func (v *APIKeyValidator) ValidateAsync(apiKeyID int, provider, encryptedKey, verificationType string) error {
	if apiKeyID <= 0 {
		return fmt.Errorf("invalid API key ID")
	}
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if encryptedKey == "" {
		return fmt.Errorf("encrypted key cannot be empty")
	}

	go v.performVerificationWithRetry(apiKeyID, provider, encryptedKey, verificationType, 0)

	return nil
}

func (v *APIKeyValidator) testConnection(providerConfig map[string]string, apiKey string) (bool, int, string, string, error) {
	startTime := time.Now()

	baseURL := providerConfig["api_base_url"]
	if baseURL == "" {
		return false, 0, "", "", fmt.Errorf("API base URL not configured")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return false, 0, "", "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, "", "", err
	}
	defer resp.Body.Close()

	latency := int(time.Since(startTime).Milliseconds())

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, latency, "", "", nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	pCode, pMsg := extractProviderError(body)
	if pCode == "" {
		pCode = fmt.Sprintf("HTTP_%d", resp.StatusCode)
	}
	if pMsg == "" {
		pMsg = string(body)
	}
	return false, latency, pCode, pMsg, fmt.Errorf("connection test failed with status code %d", resp.StatusCode)
}

func (v *APIKeyValidator) fetchModels(providerConfig map[string]string, apiKey string) ([]string, error) {
	baseURL := providerConfig["api_base_url"]
	if baseURL == "" {
		return nil, fmt.Errorf("API base URL not configured")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch models: status code %d", resp.StatusCode)
	}

	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, err
	}

	models := make([]string, 0, len(modelsResp.Data))
	for _, m := range modelsResp.Data {
		models = append(models, m.ID)
	}

	return models, nil
}

func (v *APIKeyValidator) IsVerified(apiKeyID int) (bool, error) {
	db := v.ensureDB()
	if db == nil {
		return false, fmt.Errorf("database not available")
	}

	var verificationResult string
	err := db.QueryRow(
		"SELECT verification_result FROM merchant_api_keys WHERE id = $1",
		apiKeyID,
	).Scan(&verificationResult)

	if err != nil {
		return false, err
	}

	return verificationResult == "verified", nil
}

func (v *APIKeyValidator) GetVerificationHistory(apiKeyID int, limit int) ([]VerificationResult, error) {
	db := v.ensureDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}

	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, api_key_id, verification_type, status,
			   connection_test, connection_latency_ms,
			   models_found, models_count,
			   pricing_verified, pricing_info,
			   error_code, error_message,
			   started_at, completed_at
		FROM api_key_verifications
		WHERE api_key_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`

	rows, err := db.Query(query, apiKeyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []VerificationResult
	for rows.Next() {
		var result VerificationResult
		var modelsJSON []byte
		var pricingJSON []byte
		var latency sql.NullInt64
		var completedAt sql.NullTime
		var errCode, errMsg sql.NullString

		err := rows.Scan(
			&result.ID, &result.APIKeyID, &result.VerificationType, &result.Status,
			&result.ConnectionTest, &latency,
			&modelsJSON, &result.ModelsCount,
			&result.PricingVerified, &pricingJSON,
			&errCode, &errMsg,
			&result.StartedAt, &completedAt,
		)
		if err != nil {
			logger.LogError(context.Background(), "api_key_validator", "GetVerificationHistory scan failed", err, map[string]interface{}{
				"api_key_id": apiKeyID,
			})
			continue
		}

		if latency.Valid {
			result.ConnectionLatency = int(latency.Int64)
		}
		if completedAt.Valid {
			result.CompletedAt = completedAt.Time
		}
		if errCode.Valid {
			result.ErrorCode = errCode.String
		}
		if errMsg.Valid {
			result.ErrorMessage = errMsg.String
		}

		if len(modelsJSON) > 0 {
			json.Unmarshal(modelsJSON, &result.ModelsFound)
		}
		if len(pricingJSON) > 0 {
			json.Unmarshal(pricingJSON, &result.PricingInfo)
		}

		history = append(history, result)
	}

	return history, nil
}

func (v *APIKeyValidator) createVerificationRecord(apiKeyID int, verificationType string) (int, error) {
	db := v.ensureDB()
	if db == nil {
		return 0, fmt.Errorf("database not available")
	}

	var verificationID int
	err := db.QueryRow(
		`INSERT INTO api_key_verifications (api_key_id, verification_type, status, started_at)
		 VALUES ($1, $2, 'in_progress', NOW())
		 RETURNING id`,
		apiKeyID, verificationType,
	).Scan(&verificationID)

	return verificationID, err
}

func (v *APIKeyValidator) updateVerificationRecord(verificationID int, result VerificationResult) error {
	db := v.ensureDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}

	modelsJSON, _ := json.Marshal(result.ModelsFound)
	pricingJSON, _ := json.Marshal(result.PricingInfo)

	_, err := db.Exec(
		`UPDATE api_key_verifications 
		 SET status = $1, connection_test = $2, connection_latency_ms = $3,
		     models_found = $4, models_count = $5, pricing_verified = $6, pricing_info = $7,
		     completed_at = $8, error_code = $9, error_message = $10
		 WHERE id = $11`,
		result.Status, result.ConnectionTest, result.ConnectionLatency,
		modelsJSON, result.ModelsCount, result.PricingVerified, pricingJSON,
		result.CompletedAt, nullStr(result.ErrorCode), nullStr(result.ErrorMessage),
		verificationID,
	)

	return err
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (v *APIKeyValidator) updateAPIKeyVerificationStatus(apiKeyID int, result VerificationResult) error {
	db := v.ensureDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}

	modelsJSON, _ := json.Marshal(result.ModelsFound)
	verifyMsg := result.ErrorMessage
	if verifyMsg == "" && result.ErrorCode != "" {
		verifyMsg = result.ErrorCode
	}
	if result.Status == "failed" && result.ErrorCode != "" && result.ErrorMessage != "" {
		verifyMsg = fmt.Sprintf("[%s] %s", result.ErrorCode, result.ErrorMessage)
	} else if result.Status == "failed" && result.ErrorCode != "" && verifyMsg != "" && verifyMsg != result.ErrorCode {
		verifyMsg = fmt.Sprintf("[%s] %s", result.ErrorCode, verifyMsg)
	}
	dbResult := normalizeVerificationDBStatus(result.Status)

	_, err := db.Exec(
		`UPDATE merchant_api_keys 
		 SET verification_result = $1, verified_at = $2, models_supported = $3, verification_message = $4, updated_at = NOW()
		 WHERE id = $5`,
		dbResult, result.CompletedAt, modelsJSON, verifyMsg, apiKeyID,
	)

	return err
}

func (v *APIKeyValidator) getProviderConfig(provider string) (map[string]string, error) {
	db := v.ensureDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var apiBaseURL string
	err := db.QueryRow(
		"SELECT COALESCE(api_base_url, '') FROM model_providers WHERE code = $1 AND status = 'active'",
		provider,
	).Scan(&apiBaseURL)

	if err != nil {
		return nil, err
	}

	return map[string]string{
		"api_base_url": apiBaseURL,
	}, nil
}

func (v *APIKeyValidator) handleVerificationError(ctx context.Context, verificationID, apiKeyID int, errorCode, errorMessage string, startTime time.Time) {
	result := VerificationResult{
		ID:           verificationID,
		Status:       "failed",
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		CompletedAt:  time.Now(),
	}

	var provider string
	if db := v.ensureDB(); db != nil {
		db.QueryRow(
			"SELECT provider FROM api_key_verifications v JOIN merchant_api_keys k ON v.api_key_id = k.id WHERE v.id = $1",
			verificationID,
		).Scan(&provider)
	}

	if provider != "" {
		duration := result.CompletedAt.Sub(startTime).Seconds()
		metrics.VerificationDuration.WithLabelValues(provider, "error").Observe(duration)
		metrics.VerificationTotal.WithLabelValues(provider, "error", "failed").Inc()
	}

	err := v.updateVerificationRecord(verificationID, result)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to update verification record with error", err, map[string]interface{}{
			"verification_id": verificationID,
		})
	}

	if apiKeyID > 0 {
		if uerr := v.updateAPIKeyVerificationStatus(apiKeyID, result); uerr != nil {
			logger.LogError(ctx, "api_key_validator", "Failed to update merchant_api_keys after verification error", uerr, map[string]interface{}{
				"api_key_id": apiKeyID,
			})
		}
	}

	logger.LogError(ctx, "api_key_validator", "API key verification failed", fmt.Errorf("%s", errorMessage), map[string]interface{}{
		"verification_id": verificationID,
		"error_code":      errorCode,
	})
}

func normalizeVerificationDBStatus(status string) string {
	if status == "success" {
		return "verified"
	}
	return status
}

func isDeepVerification(verificationType string) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(verificationType)), "deep")
}

func isQuotaProbeSupported(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai", "zhipu", "anthropic":
		return true
	default:
		return false
	}
}

func (v *APIKeyValidator) probeQuota(providerConfig map[string]string, provider, apiKey string, models []string) (bool, string, string) {
	baseURL := strings.TrimRight(providerConfig["api_base_url"], "/")
	if baseURL == "" {
		return false, "QUOTA_PROBE_CONFIG_ERROR", "provider api_base_url is empty"
	}
	model := selectProbeModel(provider, models)
	if model == "" {
		return false, "QUOTA_PROBE_MODEL_MISSING", "no model available for quota probe"
	}

	endpoint := baseURL + "/chat/completions"
	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
		"max_tokens": 1,
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, "QUOTA_PROBE_REQUEST_BUILD_FAILED", err.Error()
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return false, "QUOTA_PROBE_NETWORK_ERROR", err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, "", ""
	}

	rawBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	code, msg := extractProviderError(rawBody)
	if code == "" {
		code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
	}
	if msg == "" {
		msg = strings.TrimSpace(string(rawBody))
	}
	return false, code, msg
}

func selectProbeModel(provider string, models []string) string {
	if len(models) > 0 {
		return models[0]
	}
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "zhipu":
		return "glm-4"
	case "openai":
		return "gpt-4o-mini"
	case "anthropic":
		return "claude-3-5-sonnet-20241022"
	default:
		return ""
	}
}

func extractProviderError(body []byte) (string, string) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "", ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", trimmed
	}

	if errNode, ok := payload["error"].(map[string]interface{}); ok {
		return getString(errNode, "code"), firstNonEmpty(
			getString(errNode, "message"),
			getString(errNode, "msg"),
		)
	}
	return getString(payload, "code"), firstNonEmpty(
		getString(payload, "message"),
		getString(payload, "msg"),
		trimmed,
	)
}

func getString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprintf("%v", s)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
