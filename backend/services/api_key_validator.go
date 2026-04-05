package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/utils"
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
}

type APIKeyValidator struct {
	db *sql.DB
}

var (
	apiKeyValidator     *APIKeyValidator
	apiKeyValidatorOnce sync.Once
)

func GetAPIKeyValidator() *APIKeyValidator {
	apiKeyValidatorOnce.Do(func() {
		apiKeyValidator = &APIKeyValidator{
			db: config.GetDB(),
		}
	})
	return apiKeyValidator
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

	go v.performVerification(apiKeyID, provider, encryptedKey, verificationType)

	return nil
}

func (v *APIKeyValidator) performVerification(apiKeyID int, provider, encryptedKey, verificationType string) {
	ctx := context.Background()
	startTime := time.Now()

	result := VerificationResult{
		APIKeyID:         apiKeyID,
		VerificationType: verificationType,
		Status:           "in_progress",
		StartedAt:        startTime,
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
		v.handleVerificationError(ctx, verificationID, "DECRYPTION_FAILED", "Failed to decrypt API key", startTime)
		return
	}

	providerConfig, err := v.getProviderConfig(provider)
	if err != nil {
		v.handleVerificationError(ctx, verificationID, "PROVIDER_NOT_FOUND", "Provider configuration not found", startTime)
		return
	}

	connectionOK, latency, err := v.testConnection(providerConfig, decryptedKey)
	if err != nil {
		v.handleVerificationError(ctx, verificationID, "CONNECTION_FAILED", err.Error(), startTime)
		return
	}
	result.ConnectionTest = connectionOK
	result.ConnectionLatency = latency

	models, err := v.fetchModels(providerConfig, decryptedKey)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to fetch models", err, map[string]interface{}{
			"api_key_id": apiKeyID,
			"provider":   provider,
		})
	} else {
		result.ModelsFound = models
		result.ModelsCount = len(models)
	}

	result.Status = "success"
	result.CompletedAt = time.Now()

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

	logger.LogInfo(ctx, "api_key_validator", "API key verification completed", map[string]interface{}{
		"api_key_id":         apiKeyID,
		"verification_id":    verificationID,
		"status":             result.Status,
		"connection_test":    result.ConnectionTest,
		"models_count":       result.ModelsCount,
		"connection_latency": result.ConnectionLatency,
	})
}

func (v *APIKeyValidator) testConnection(providerConfig map[string]string, apiKey string) (bool, int, error) {
	startTime := time.Now()

	baseURL := providerConfig["api_base_url"]
	if baseURL == "" {
		return false, 0, fmt.Errorf("API base URL not configured")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return false, 0, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	latency := int(time.Since(startTime).Milliseconds())

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, latency, nil
	}

	return false, latency, fmt.Errorf("connection test failed with status code %d", resp.StatusCode)
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
	if v.db == nil {
		v.db = config.GetDB()
	}
	if v.db == nil {
		return false, fmt.Errorf("database not available")
	}

	var verificationResult string
	err := v.db.QueryRow(
		"SELECT verification_result FROM merchant_api_keys WHERE id = $1",
		apiKeyID,
	).Scan(&verificationResult)

	if err != nil {
		return false, err
	}

	return verificationResult == "verified", nil
}

func (v *APIKeyValidator) GetVerificationHistory(apiKeyID int, limit int) ([]VerificationResult, error) {
	if v.db == nil {
		v.db = config.GetDB()
	}
	if v.db == nil {
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

	rows, err := v.db.Query(query, apiKeyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []VerificationResult
	for rows.Next() {
		var result VerificationResult
		var modelsJSON []byte
		var pricingJSON []byte

		err := rows.Scan(
			&result.ID, &result.APIKeyID, &result.VerificationType, &result.Status,
			&result.ConnectionTest, &result.ConnectionLatency,
			&modelsJSON, &result.ModelsCount,
			&result.PricingVerified, &pricingJSON,
			&result.ErrorCode, &result.ErrorMessage,
			&result.StartedAt, &result.CompletedAt,
		)
		if err != nil {
			continue
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
	if v.db == nil {
		v.db = config.GetDB()
	}
	if v.db == nil {
		return 0, fmt.Errorf("database not available")
	}

	var verificationID int
	err := v.db.QueryRow(
		`INSERT INTO api_key_verifications (api_key_id, verification_type, status, started_at)
		 VALUES ($1, $2, 'in_progress', NOW())
		 RETURNING id`,
		apiKeyID, verificationType,
	).Scan(&verificationID)

	return verificationID, err
}

func (v *APIKeyValidator) updateVerificationRecord(verificationID int, result VerificationResult) error {
	if v.db == nil {
		v.db = config.GetDB()
	}
	if v.db == nil {
		return fmt.Errorf("database not available")
	}

	modelsJSON, _ := json.Marshal(result.ModelsFound)
	pricingJSON, _ := json.Marshal(result.PricingInfo)

	_, err := v.db.Exec(
		`UPDATE api_key_verifications 
		 SET status = $1, connection_test = $2, connection_latency_ms = $3,
		     models_found = $4, models_count = $5, pricing_verified = $6, pricing_info = $7,
		     completed_at = $8
		 WHERE id = $9`,
		result.Status, result.ConnectionTest, result.ConnectionLatency,
		modelsJSON, result.ModelsCount, result.PricingVerified, pricingJSON,
		result.CompletedAt, verificationID,
	)

	return err
}

func (v *APIKeyValidator) updateAPIKeyVerificationStatus(apiKeyID int, result VerificationResult) error {
	if v.db == nil {
		v.db = config.GetDB()
	}
	if v.db == nil {
		return fmt.Errorf("database not available")
	}

	modelsJSON, _ := json.Marshal(result.ModelsFound)

	_, err := v.db.Exec(
		`UPDATE merchant_api_keys 
		 SET verification_result = $1, verified_at = $2, models_supported = $3, verification_message = $4, updated_at = NOW()
		 WHERE id = $5`,
		result.Status, result.CompletedAt, modelsJSON, "", apiKeyID,
	)

	return err
}

func (v *APIKeyValidator) getProviderConfig(provider string) (map[string]string, error) {
	if v.db == nil {
		v.db = config.GetDB()
	}
	if v.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var apiBaseURL string
	err := v.db.QueryRow(
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

func (v *APIKeyValidator) handleVerificationError(ctx context.Context, verificationID int, errorCode, errorMessage string, startTime time.Time) {
	result := VerificationResult{
		ID:           verificationID,
		Status:       "failed",
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		CompletedAt:  time.Now(),
	}

	err := v.updateVerificationRecord(verificationID, result)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to update verification record with error", err, map[string]interface{}{
			"verification_id": verificationID,
		})
	}

	logger.LogError(ctx, "api_key_validator", "API key verification failed", fmt.Errorf("%s", errorMessage), map[string]interface{}{
		"verification_id": verificationID,
		"error_code":      errorCode,
	})
}
