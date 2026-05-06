package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
	ConnectionError   string                 `json:"connection_error,omitempty"`
	NetworkStatus     string                 `json:"network_status,omitempty"`
	ModelsFound       []string               `json:"models_found"`
	ModelsCount       int                    `json:"models_count"`
	PricingVerified   bool                   `json:"pricing_verified"`
	PricingInfo       map[string]interface{} `json:"pricing_info,omitempty"`
	ErrorCode         string                 `json:"error_code,omitempty"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
	ErrorCategory     string                 `json:"error_category,omitempty"`
	RouteMode         string                 `json:"route_mode,omitempty"`
	EndpointUsed      string                 `json:"endpoint_used,omitempty"`
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
	verificationStatusSuccess  = "success"
	verificationStatusFailed   = "failed"
	verificationResultVerified = "verified"

	ErrCodeQuotaProbeRequestBuildFailed = "QUOTA_PROBE_REQUEST_BUILD_FAILED"
	ErrCodeQuotaProbeNetworkError       = "QUOTA_PROBE_NETWORK_ERROR"
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

	routeMode, routeConfig, region := v.getAPIKeyRouteConfig(apiKeyID, provider)
	if routeMode != "" && routeMode != GatewayModeDirect {
		v.performVerificationWithRouteMode(apiKeyID, provider, encryptedKey, verificationType, routeMode, routeConfig, region, retryCount)
		return
	}

	baseURL := strings.TrimRight(providerConfig["api_base_url"], "/")
	modelsURL := baseURL + "/models"
	probeClient := newProxyAwareHTTPClient(10*time.Second, routeMode)
	var probe *ProbeModelsResult
	var probeErr error
	for attempt := retryCount; ; attempt++ {
		probe, probeErr = ProbeProviderModels(ctx, probeClient, modelsURL, decryptedKey, provider)
		if probeErr == nil && probe != nil && probe.Success {
			result.RetryCount = attempt
			break
		}
		if attempt >= v.maxRetries || !shouldRetryVerificationAttempt(probe, probeErr) {
			msg := "connection test failed"
			if probeErr != nil {
				msg = probeErr.Error()
			}
			code := "CONNECTION_FAILED"
			if probe != nil && strings.TrimSpace(probe.ErrorCode) != "" {
				code = probe.ErrorCode
			}
			if probe != nil && strings.TrimSpace(probe.ErrorMsg) != "" {
				msg = firstNonEmpty(msg, probe.ErrorMsg)
			}
			v.handleVerificationError(ctx, verificationID, apiKeyID, code, msg, startTime)
			return
		}
		metrics.VerificationRetries.WithLabelValues(provider, fmt.Sprintf("%d", attempt+1)).Inc()
		time.Sleep(v.retryDelay * time.Duration(1<<attempt))
	}
	result.ConnectionTest = true
	result.ConnectionLatency = probe.LatencyMs
	metrics.VerificationConnectionLatency.WithLabelValues(provider).Observe(float64(probe.LatencyMs))
	result.ModelsFound = probe.Models
	result.ModelsCount = len(probe.Models)
	metrics.ModelsDiscovered.WithLabelValues(provider).Add(float64(len(probe.Models)))

	if isDeepVerification(verificationType) {
		apiFmt := strings.ToLower(strings.TrimSpace(providerConfig["api_format"]))
		probeSupported := apiFmt == modelProviderOpenAI
		result.PricingVerified = probeSupported
		if probeSupported {
			quotaOK, quotaCode, quotaMsg := v.probeQuota(providerConfig, provider, decryptedKey, result.ModelsFound)
			if !quotaOK {
				v.handleVerificationError(ctx, verificationID, apiKeyID, quotaCode, quotaMsg, startTime)
				return
			}
		}
	}

	result.Status = verificationStatusSuccess
	result.CompletedAt = time.Now()

	duration := result.CompletedAt.Sub(startTime).Seconds()
	metrics.VerificationDuration.WithLabelValues(provider, verificationType).Observe(duration)
	metrics.VerificationTotal.WithLabelValues(provider, verificationType, verificationStatusSuccess).Inc()

	err = v.updateVerificationRecord(verificationID, result)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to update verification record", err, map[string]interface{}{
			"verification_id": verificationID,
		})
		return
	}

	if isDeepVerification(verificationType) {
		err = v.updateAPIKeyVerificationStatus(apiKeyID, result)
		if err != nil {
			logger.LogError(ctx, "api_key_validator", "Failed to update API key verification status", err, map[string]interface{}{
				"api_key_id": apiKeyID,
			})
			return
		}
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
		"retry_count":        result.RetryCount,
		"is_deep":            isDeepVerification(verificationType),
	})
}

func shouldRetryVerificationAttempt(probe *ProbeModelsResult, probeErr error) bool {
	if probeErr != nil {
		errInfo := MapProviderError(0, "", probeErr.Error(), nil, probeErr, "")
		return errInfo.Retryable
	}
	if probe == nil {
		return false
	}
	errInfo := MapProviderError(probe.StatusCode, probe.ErrorCode, probe.ErrorMsg, nil, nil, probe.RawErrorExcerpt)
	return errInfo.Retryable
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

	return verificationResult == verificationResultVerified, nil
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
			   route_mode, endpoint_used, error_category,
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
		var routeMode, endpointUsed, errorCategory sql.NullString

		err := rows.Scan(
			&result.ID, &result.APIKeyID, &result.VerificationType, &result.Status,
			&result.ConnectionTest, &latency,
			&modelsJSON, &result.ModelsCount,
			&result.PricingVerified, &pricingJSON,
			&errCode, &errMsg,
			&routeMode, &endpointUsed, &errorCategory,
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
		if routeMode.Valid {
			result.RouteMode = routeMode.String
		}
		if endpointUsed.Valid {
			result.EndpointUsed = endpointUsed.String
		}
		if errorCategory.Valid {
			result.ErrorCategory = errorCategory.String
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
		     completed_at = $8, error_code = $9, error_message = $10,
		     route_mode = $11, endpoint_used = $12, error_category = $13
		 WHERE id = $14`,
		result.Status, result.ConnectionTest, result.ConnectionLatency,
		modelsJSON, result.ModelsCount, result.PricingVerified, pricingJSON,
		result.CompletedAt, nullStr(result.ErrorCode), nullStr(result.ErrorMessage),
		nullStr(result.RouteMode), nullStr(result.EndpointUsed), nullStr(result.ErrorCategory),
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
	dbResult := normalizeVerificationDBStatus(result.Status)
	return v.updateAPIKeyVerificationStatusWithStatus(apiKeyID, result, dbResult)
}

func (v *APIKeyValidator) updateAPIKeyVerificationStatusWithStatus(apiKeyID int, result VerificationResult, dbResult string) error {
	db := v.ensureDB()
	if db == nil {
		return fmt.Errorf("database not available")
	}

	modelsJSON, _ := json.Marshal(result.ModelsFound)
	verifyMsg := result.ErrorMessage
	if verifyMsg == "" && result.ErrorCode != "" {
		verifyMsg = result.ErrorCode
	}
	if result.Status == verificationStatusFailed && result.ErrorCode != "" && result.ErrorMessage != "" {
		verifyMsg = fmt.Sprintf("[%s] %s", result.ErrorCode, result.ErrorMessage)
	} else if result.Status == verificationStatusFailed && result.ErrorCode != "" && verifyMsg != "" && verifyMsg != result.ErrorCode {
		verifyMsg = fmt.Sprintf("[%s] %s", result.ErrorCode, verifyMsg)
	}

	var verifiedAt interface{}
	if dbResult == VerificationStatusVerified {
		verifiedAt = result.CompletedAt
	}

	var merchantID int
	err := db.QueryRow(
		`UPDATE merchant_api_keys 
		 SET verification_result = $1, verified_at = $2, models_supported = $3, verification_message = $4, updated_at = NOW()
		 WHERE id = $5
		 RETURNING merchant_id`,
		dbResult, verifiedAt, modelsJSON, verifyMsg, apiKeyID,
	).Scan(&merchantID)
	if err != nil {
		return err
	}

	cache.Delete(context.Background(), cache.MerchantAPIKeysKey(merchantID))

	return nil
}

func (v *APIKeyValidator) getProviderConfig(provider string) (map[string]string, error) {
	db := v.ensureDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var apiBaseURL, apiFormat string
	err := db.QueryRow(
		`SELECT COALESCE(api_base_url, ''), COALESCE(NULLIF(trim(api_format), ''), $1)
		 FROM model_providers WHERE code = $2 AND status = 'active'`,
		modelProviderOpenAI, provider,
	).Scan(&apiBaseURL, &apiFormat)

	if err != nil {
		return nil, err
	}

	return map[string]string{
		"api_base_url": apiBaseURL,
		"api_format":   apiFormat,
	}, nil
}

func (v *APIKeyValidator) getAPIKeyRouteConfig(apiKeyID int, provider string) (string, map[string]interface{}, string) {
	db := v.ensureDB()
	if db == nil {
		return "", nil, ""
	}

	var routeMode string
	var routeConfigJSON []byte
	var region string
	var providerRegion string

	err := db.QueryRow(
		`SELECT COALESCE(mak.route_mode, 'auto'),
		        mak.route_config,
		        COALESCE(mak.region, 'domestic'),
		        COALESCE(mp.provider_region, 'domestic')
		 FROM merchant_api_keys mak
		 LEFT JOIN model_providers mp ON mak.provider = mp.code
		 WHERE mak.id = $1`,
		apiKeyID,
	).Scan(&routeMode, &routeConfigJSON, &region, &providerRegion)

	if err != nil {
		return "", nil, ""
	}

	var routeConfig map[string]interface{}
	if len(routeConfigJSON) > 0 {
		_ = json.Unmarshal(routeConfigJSON, &routeConfig)
	}

	if routeMode == RouteModeAuto || routeMode == "" {
		resolved := resolveAutoRouteMode(providerRegion)
		return resolved, routeConfig, region
	}

	return routeMode, routeConfig, region
}

func (v *APIKeyValidator) handleVerificationError(ctx context.Context, verificationID, apiKeyID int, errorCode, errorMessage string, startTime time.Time) {
	errInfo := MapProviderError(0, errorCode, errorMessage, nil, nil, "")
	result := VerificationResult{
		ID:            verificationID,
		Status:        verificationStatusFailed,
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
		ErrorCategory: errInfo.Category,
		CompletedAt:   time.Now(),
	}

	var provider string
	var currentStatus string
	if db := v.ensureDB(); db != nil {
		db.QueryRow(
			"SELECT k.provider, k.verification_result FROM api_key_verifications v JOIN merchant_api_keys k ON v.api_key_id = k.id WHERE v.id = $1",
			verificationID,
		).Scan(&provider, &currentStatus)
	}

	if provider != "" {
		duration := result.CompletedAt.Sub(startTime).Seconds()
		metrics.VerificationDuration.WithLabelValues(provider, "error").Observe(duration)
		metrics.VerificationTotal.WithLabelValues(provider, "error", verificationStatusFailed).Inc()
	}

	err := v.updateVerificationRecord(verificationID, result)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to update verification record with error", err, map[string]interface{}{
			"verification_id": verificationID,
		})
	}

	if apiKeyID > 0 {
		newStatus := MapErrorCategoryToVerificationStatus(errInfo.Category, currentStatus)
		if uerr := v.updateAPIKeyVerificationStatusWithStatus(apiKeyID, result, newStatus); uerr != nil {
			logger.LogError(ctx, "api_key_validator", "Failed to update merchant_api_keys after verification error", uerr, map[string]interface{}{
				"api_key_id": apiKeyID,
			})
		}
	}

	logger.LogError(ctx, "api_key_validator", "API key verification failed", fmt.Errorf("%s", errorMessage), map[string]interface{}{
		"verification_id": verificationID,
		"error_code":      errorCode,
		"error_category":  errInfo.Category,
	})
}

func normalizeVerificationDBStatus(status string) string {
	if status == verificationStatusSuccess {
		return verificationResultVerified
	}
	return status
}

func isDeepVerification(verificationType string) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(verificationType)), "deep")
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
		return false, ErrCodeQuotaProbeRequestBuildFailed, err.Error()
	}
	SetProviderAuthHeaders(req, provider, apiKey)

	resp, err := newProxyAwareHTTPClient(15*time.Second, "").Do(req)
	if err != nil {
		return false, ErrCodeQuotaProbeNetworkError, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, "", ""
	}

	rawBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	code, msg := ExtractProviderError(rawBody)
	if code == "" {
		code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
	}
	if msg == "" {
		msg = strings.TrimSpace(string(rawBody))
	}
	if resp.StatusCode == http.StatusPaymentRequired || resp.StatusCode == http.StatusTooManyRequests {
		code = errorCategoryQuotaInsufficient
	}
	return false, code, msg
}

func newProxyAwareHTTPClient(timeout time.Duration, routeMode string) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if routeMode == GatewayModeProxy {
		httpsProxy := os.Getenv("HTTPS_PROXY")
		if httpsProxy == "" {
			httpsProxy = os.Getenv("https_proxy")
		}
		if httpsProxy != "" {
			proxyURL, err := url.Parse(httpsProxy)
			if err == nil {
				transport.Proxy = http.ProxyURL(proxyURL)
			}
		}
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func (v *APIKeyValidator) getProviderRegion(provider string) string {
	db := v.ensureDB()
	if db == nil {
		return regionDomestic
	}
	var providerRegion string
	err := db.QueryRow(
		"SELECT COALESCE(provider_region, 'domestic') FROM model_providers WHERE code = $1",
		provider,
	).Scan(&providerRegion)
	if err != nil {
		return regionDomestic
	}
	return providerRegion
}

func selectProbeModel(provider string, models []string) string {
	if len(models) > 0 {
		return models[0]
	}
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case modelProviderZhipu:
		return "glm-4"
	case modelProviderOpenAI:
		return "gpt-4o-mini"
	case modelProviderAnthropic:
		return "claude-3-5-sonnet-20241022"
	case modelProviderDeepseek:
		return "deepseek-chat"
	default:
		return ""
	}
}

func (v *APIKeyValidator) ValidateAsyncWithRouteMode(
	apiKeyID int,
	provider, encryptedKey, verificationType, routeMode string,
	routeConfig map[string]interface{},
	region string,
) error {
	if apiKeyID <= 0 {
		return fmt.Errorf("invalid API key ID")
	}
	if provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if encryptedKey == "" {
		return fmt.Errorf("encrypted key cannot be empty")
	}

	go v.performVerificationWithRouteMode(apiKeyID, provider, encryptedKey, verificationType, routeMode, routeConfig, region, 0)

	return nil
}

func (v *APIKeyValidator) performVerificationWithRouteMode(
	apiKeyID int,
	provider, encryptedKey, verificationType, routeMode string,
	routeConfig map[string]interface{},
	region string,
	retryCount int,
) {
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

	if routeMode == RouteModeAuto || routeMode == string(GoalAuto) {
		providerRegion := v.getProviderRegion(provider)
		resolved := resolveAutoRouteMode(providerRegion)
		logger.LogInfo(ctx, "api_key_validator", "Resolved auto route mode for verification", map[string]interface{}{
			"api_key_id":      apiKeyID,
			"original_mode":   routeMode,
			"resolved_mode":   resolved,
			"provider_region": providerRegion,
		})
		routeMode = resolved
	}

	endpoint, err := v.resolveEndpointByRouteMode(ctx, provider, routeMode, routeConfig, region)
	if err != nil {
		v.handleVerificationError(ctx, verificationID, apiKeyID, "ENDPOINT_RESOLVE_FAILED", err.Error(), startTime)
		return
	}
	result.EndpointUsed = endpoint
	result.RouteMode = routeMode

	authToken := v.resolveAuthToken(routeMode, decryptedKey)

	modelsEndpoint := endpoint
	if routeMode == GatewayModeLitellm {
		upstreamEndpoint, upstreamErr := v.resolveDirectEndpoint(ctx, provider, routeConfig, region)
		if upstreamErr == nil && upstreamEndpoint != "" {
			modelsEndpoint = upstreamEndpoint
		}
	}

	baseURL := strings.TrimRight(modelsEndpoint, "/")
	modelsURL := baseURL + "/models"

	modelsAuthToken := decryptedKey
	if routeMode == GatewayModeLitellm {
		modelsAuthToken = decryptedKey
	}

	probeClient := newProxyAwareHTTPClient(10*time.Second, routeMode)

	var probe *ProbeModelsResult
	var probeErr error
	var connectionFailed bool
	for attempt := retryCount; ; attempt++ {
		probe, probeErr = ProbeProviderModels(ctx, probeClient, modelsURL, modelsAuthToken, provider)
		if probeErr == nil && probe != nil && probe.Success {
			result.RetryCount = attempt
			break
		}
		if attempt >= v.maxRetries || !shouldRetryVerificationAttempt(probe, probeErr) {
			connectionFailed = true
			break
		}
		metrics.VerificationRetries.WithLabelValues(provider, fmt.Sprintf("%d", attempt+1)).Inc()
		time.Sleep(v.retryDelay * time.Duration(1<<attempt))
	}

	if connectionFailed {
		msg := "connection test failed"
		if probeErr != nil {
			msg = probeErr.Error()
		}
		code := "CONNECTION_FAILED"
		if probe != nil && strings.TrimSpace(probe.ErrorCode) != "" {
			code = probe.ErrorCode
		}
		if probe != nil && strings.TrimSpace(probe.ErrorMsg) != "" {
			msg = firstNonEmpty(msg, probe.ErrorMsg)
		}

		result.ConnectionTest = false
		result.ConnectionError = msg
		result.NetworkStatus = "network_error"
		result.ModelsFound = GetPredefinedModels(provider)
		result.ModelsCount = len(result.ModelsFound)

		logger.LogWarn(ctx, "api_key_validator", "Light verification failed, using predefined models", map[string]interface{}{
			"api_key_id":      apiKeyID,
			"provider":        provider,
			"error_code":      code,
			"error_message":   msg,
			"models_fallback": result.ModelsFound,
		})

		if !isDeepVerification(verificationType) {
			result.Status = verificationStatusSuccess
			result.CompletedAt = time.Now()
			duration := result.CompletedAt.Sub(startTime).Seconds()
			metrics.VerificationDuration.WithLabelValues(provider, verificationType).Observe(duration)
			metrics.VerificationTotal.WithLabelValues(provider, verificationType, verificationStatusSuccess).Inc()

			err = v.updateVerificationRecord(verificationID, result)
			if err != nil {
				logger.LogError(ctx, "api_key_validator", "Failed to update verification record", err, map[string]interface{}{
					"verification_id": verificationID,
				})
				return
			}

			err = v.cacheVerificationResult(ctx, result)
			if err != nil {
				logger.LogError(ctx, "api_key_validator", "Failed to cache verification result", err, map[string]interface{}{
					"api_key_id": apiKeyID,
				})
			}

			logger.LogInfo(ctx, "api_key_validator", "Light verification completed with fallback models", map[string]interface{}{
				"api_key_id":      apiKeyID,
				"verification_id": verificationID,
				"status":          result.Status,
				"connection_test": result.ConnectionTest,
				"models_count":    result.ModelsCount,
				"network_status":  result.NetworkStatus,
			})
			return
		}

		v.handleVerificationError(ctx, verificationID, apiKeyID, code, msg, startTime)
		return
	} else {
		result.ConnectionTest = true
		result.NetworkStatus = "ok"
		result.ConnectionLatency = probe.LatencyMs
		metrics.VerificationConnectionLatency.WithLabelValues(provider).Observe(float64(probe.LatencyMs))
		result.ModelsFound = probe.Models
		result.ModelsCount = len(probe.Models)
		metrics.ModelsDiscovered.WithLabelValues(provider).Add(float64(len(probe.Models)))
	}

	if isDeepVerification(verificationType) {
		providerConfig, providerErr := v.getProviderConfig(provider)
		if providerErr == nil {
			apiFmt := strings.ToLower(strings.TrimSpace(providerConfig["api_format"]))
			probeSupported := apiFmt == modelProviderOpenAI
			result.PricingVerified = probeSupported
			if probeSupported {
				quotaOK, quotaCode, quotaMsg := v.probeQuotaWithEndpoint(endpoint, provider, authToken, result.ModelsFound, routeMode, decryptedKey)
				if !quotaOK {
					v.handleVerificationError(ctx, verificationID, apiKeyID, quotaCode, quotaMsg, startTime)
					return
				}
			}
		}
	}

	result.Status = verificationStatusSuccess
	result.CompletedAt = time.Now()

	duration := result.CompletedAt.Sub(startTime).Seconds()
	metrics.VerificationDuration.WithLabelValues(provider, verificationType).Observe(duration)
	metrics.VerificationTotal.WithLabelValues(provider, verificationType, verificationStatusSuccess).Inc()

	err = v.updateVerificationRecord(verificationID, result)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to update verification record", err, map[string]interface{}{
			"verification_id": verificationID,
		})
		return
	}

	if isDeepVerification(verificationType) {
		err = v.updateAPIKeyVerificationStatus(apiKeyID, result)
		if err != nil {
			logger.LogError(ctx, "api_key_validator", "Failed to update API key verification status", err, map[string]interface{}{
				"api_key_id": apiKeyID,
			})
			return
		}
	}

	err = v.cacheVerificationResult(ctx, result)
	if err != nil {
		logger.LogError(ctx, "api_key_validator", "Failed to cache verification result", err, map[string]interface{}{
			"api_key_id": apiKeyID,
		})
	}

	logger.LogInfo(ctx, "api_key_validator", "API key verification with route mode completed", map[string]interface{}{
		"api_key_id":         apiKeyID,
		"verification_id":    verificationID,
		"status":             result.Status,
		"connection_test":    result.ConnectionTest,
		"models_count":       result.ModelsCount,
		"connection_latency": result.ConnectionLatency,
		"retry_count":        result.RetryCount,
		"route_mode":         routeMode,
		"endpoint_used":      endpoint,
		"is_deep":            isDeepVerification(verificationType),
	})
}

func (v *APIKeyValidator) resolveEndpointByRouteMode(ctx context.Context, provider, routeMode string, routeConfig map[string]interface{}, region string) (string, error) {
	switch routeMode {
	case GatewayModeDirect:
		return v.resolveDirectEndpoint(ctx, provider, routeConfig, region)
	case GatewayModeLitellm:
		return v.resolveLitellmEndpoint(ctx, routeConfig, region)
	case GatewayModeProxy:
		return v.resolveProxyEndpoint(ctx, provider, routeConfig)
	case string(GoalAuto):
		return v.resolveAutoEndpoint(ctx, provider, routeConfig, region)
	default:
		return v.resolveDirectEndpoint(ctx, provider, routeConfig, region)
	}
}

func (v *APIKeyValidator) resolveDirectEndpoint(ctx context.Context, provider string, routeConfig map[string]interface{}, region string) (string, error) {
	if routeConfig != nil {
		if endpoint, ok := routeConfig["endpoint_url"].(string); ok && endpoint != "" {
			return endpoint, nil
		}

		if endpoints, ok := routeConfig["endpoints"].(map[string]interface{}); ok {
			if directEndpoints, ok := endpoints[GatewayModeDirect].(map[string]interface{}); ok {
				if region == "" {
					region = regionOverseas
				}
				if url, ok := directEndpoints[region].(string); ok && url != "" {
					return url, nil
				}
			}
		}
	}

	providerConfig, err := v.getProviderConfig(provider)
	if err != nil {
		return "", fmt.Errorf("provider configuration not found: %w", err)
	}

	baseURL := strings.TrimRight(providerConfig["api_base_url"], "/")
	if baseURL == "" {
		return "", fmt.Errorf("provider api_base_url is empty")
	}

	return baseURL, nil
}

func (v *APIKeyValidator) resolveLitellmEndpoint(ctx context.Context, routeConfig map[string]interface{}, region string) (string, error) {
	if routeConfig != nil {
		if endpoints, ok := routeConfig["endpoints"].(map[string]interface{}); ok {
			if litellmEndpoints, ok := endpoints[GatewayModeLitellm].(map[string]interface{}); ok {
				if region == "" {
					region = regionDomestic
				}
				if url, ok := litellmEndpoints[region].(string); ok && url != "" {
					return url, nil
				}
			}
		}

		if baseURL, ok := routeConfig["base_url"].(string); ok && baseURL != "" {
			return baseURL, nil
		}
	}

	litellmURL := os.Getenv("LLM_GATEWAY_LITELLM_URL")
	if litellmURL != "" {
		return litellmURL + "/v1", nil
	}

	return "", fmt.Errorf("LiteLLM endpoint not configured")
}

func (v *APIKeyValidator) resolveProxyEndpoint(ctx context.Context, provider string, routeConfig map[string]interface{}) (string, error) {
	if routeConfig != nil {
		if endpoints, ok := routeConfig["endpoints"].(map[string]interface{}); ok {
			if proxyEndpoints, ok := endpoints[GatewayModeProxy].(map[string]interface{}); ok {
				if gaapURL, ok := proxyEndpoints["gaap"].(string); ok && gaapURL != "" {
					return gaapURL, nil
				}
				for _, v := range proxyEndpoints {
					if url, ok := v.(string); ok && url != "" {
						return url, nil
					}
				}
			}
		}

		if proxyURL, ok := routeConfig["proxy_url"].(string); ok && proxyURL != "" {
			return proxyURL, nil
		}
	}

	providerConfig, err := v.getProviderConfig(provider)
	if err != nil {
		return "", fmt.Errorf("Proxy endpoint not configured and provider not found: %w", err)
	}

	if baseURL := strings.TrimRight(providerConfig["api_base_url"], "/"); baseURL != "" {
		return baseURL, nil
	}

	return "", fmt.Errorf("Proxy endpoint not configured")
}

func (v *APIKeyValidator) resolveAutoEndpoint(ctx context.Context, provider string, routeConfig map[string]interface{}, region string) (string, error) {
	priority := []string{GatewayModeDirect, GatewayModeLitellm, GatewayModeProxy}

	for _, mode := range priority {
		var endpoint string
		var err error

		switch mode {
		case GatewayModeDirect:
			endpoint, err = v.resolveDirectEndpoint(ctx, provider, routeConfig, region)
		case GatewayModeLitellm:
			endpoint, err = v.resolveLitellmEndpoint(ctx, routeConfig, region)
		case GatewayModeProxy:
			endpoint, err = v.resolveProxyEndpoint(ctx, provider, routeConfig)
		}

		if err == nil && endpoint != "" {
			logger.LogInfo(ctx, "api_key_validator", "Auto mode selected endpoint", map[string]interface{}{
				"selected_mode": mode,
				"endpoint":      endpoint,
			})
			return endpoint, nil
		}
	}

	return "", fmt.Errorf("no available endpoint for auto mode")
}

func (v *APIKeyValidator) resolveAuthToken(routeMode string, originalAPIKey string) string {
	switch routeMode {
	case "litellm":
		masterKey := os.Getenv("LITELLM_MASTER_KEY")
		if masterKey != "" {
			return masterKey
		}
		return originalAPIKey
	default:
		return originalAPIKey
	}
}

func (v *APIKeyValidator) probeQuotaWithEndpoint(endpoint, provider, apiKey string, models []string, routeMode, originalAPIKey string) (bool, string, string) {
	baseURL := strings.TrimRight(endpoint, "/")
	if baseURL == "" {
		return false, "QUOTA_PROBE_CONFIG_ERROR", "endpoint is empty"
	}
	model := selectProbeModel(provider, models)
	if model == "" {
		return false, "QUOTA_PROBE_MODEL_MISSING", "no model available for quota probe"
	}

	chatEndpoint := baseURL + "/chat/completions"

	if routeMode == GatewayModeLitellm && originalAPIKey != "" {
		return v.probeQuotaViaLitellmUserConfig(chatEndpoint, provider, model, originalAPIKey, apiKey)
	}

	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
		"max_tokens": 1,
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", chatEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, ErrCodeQuotaProbeRequestBuildFailed, err.Error()
	}
	SetProviderAuthHeaders(req, provider, apiKey)

	resp, err := newProxyAwareHTTPClient(15*time.Second, routeMode).Do(req)
	if err != nil {
		return false, ErrCodeQuotaProbeNetworkError, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, "", ""
	}

	rawBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	code, msg := ExtractProviderError(rawBody)
	if code == "" {
		code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
	}
	if msg == "" {
		msg = strings.TrimSpace(string(rawBody))
	}
	if resp.StatusCode == http.StatusPaymentRequired || resp.StatusCode == http.StatusTooManyRequests {
		code = errorCategoryQuotaInsufficient
	}
	return false, code, msg
}

func litellmProviderPrefix(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case modelProviderOpenAI:
		return modelProviderOpenAI
	case modelProviderAnthropic:
		return modelProviderAnthropic
	case modelProviderDeepseek:
		return modelProviderDeepseek
	case modelProviderAlibaba:
		return "dashscope"
	case modelProviderZhipu:
		return modelProviderZhipu
	case modelProviderMoonshot:
		return modelProviderMoonshot
	case modelProviderMinimax:
		return modelProviderMinimax
	case modelProviderGoogle:
		return "gemini"
	case modelProviderStepfun:
		return modelProviderOpenAI
	case modelProviderBytedance:
		return ""
	default:
		return provider
	}
}

func resolveLitellmModelName(provider, model string) (string, error) {
	prefix := litellmProviderPrefix(provider)
	if prefix == "" {
		return "", fmt.Errorf("provider %s is not supported by litellm (requires endpoint_id for bytedance)", provider)
	}

	modelName := model
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		modelName = model[idx+1:]
	}
	return prefix + "/" + modelName, nil
}

func (v *APIKeyValidator) probeQuotaViaLitellmUserConfig(chatEndpoint, provider, model, originalAPIKey, litellmMasterKey string) (bool, string, string) {
	providerConfig, err := v.getProviderConfig(provider)
	if err != nil {
		return false, "QUOTA_PROBE_PROVIDER_CONFIG_ERROR", fmt.Sprintf("failed to get provider config: %v", err)
	}

	upstreamBaseURL := strings.TrimRight(providerConfig["api_base_url"], "/")
	if upstreamBaseURL == "" {
		return false, "QUOTA_PROBE_UPSTREAM_URL_MISSING", "upstream base URL not configured"
	}

	litellmModel, err := resolveLitellmModelName(provider, model)
	if err != nil {
		return false, "LITELLM_UNSUPPORTED_PROVIDER", err.Error()
	}

	userConfig := map[string]interface{}{
		"model_list": []map[string]interface{}{
			{
				"model_name": "probe-model",
				"litellm_params": map[string]interface{}{
					"model":    litellmModel,
					"api_key":  originalAPIKey,
					"api_base": upstreamBaseURL,
				},
			},
		},
	}

	body := map[string]interface{}{
		"model":       "probe-model",
		"messages":    []map[string]string{{"role": "user", "content": "ping"}},
		"max_tokens":  1,
		"user_config": userConfig,
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", chatEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, ErrCodeQuotaProbeRequestBuildFailed, err.Error()
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", litellmMasterKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := newProxyAwareHTTPClient(30*time.Second, GatewayModeLitellm).Do(req)
	if err != nil {
		return false, ErrCodeQuotaProbeNetworkError, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, "", ""
	}

	rawBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	code, msg := ExtractProviderError(rawBody)
	if code == "" {
		code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
	}
	if msg == "" {
		msg = strings.TrimSpace(string(rawBody))
	}
	if resp.StatusCode == http.StatusPaymentRequired || resp.StatusCode == http.StatusTooManyRequests {
		code = errorCategoryQuotaInsufficient
	}
	return false, code, msg
}
