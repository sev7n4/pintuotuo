package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/utils"
)

type ProviderModelSyncService struct{}

func NewProviderModelSyncService() *ProviderModelSyncService {
	return &ProviderModelSyncService{}
}

func (s *ProviderModelSyncService) SyncProviderModels(ctx context.Context, providerCode string, apiKeyID int) (int, error) {
	db := config.GetDB()
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	var apiBaseURL, apiFormat string
	err := db.QueryRowContext(ctx,
		`SELECT COALESCE(NULLIF(TRIM(api_base_url), ''), ''),
		        COALESCE(NULLIF(TRIM(api_format), ''), 'openai')
		 FROM model_providers WHERE code = $1 AND status = 'active'`,
		providerCode,
	).Scan(&apiBaseURL, &apiFormat)
	if err != nil {
		return 0, fmt.Errorf("provider %s not found or inactive: %w", providerCode, err)
	}

	if apiBaseURL == "" {
		return 0, fmt.Errorf("provider %s has no api_base_url configured", providerCode)
	}

	apiKey, err := s.getDecryptedAPIKey(ctx, providerCode, apiKeyID)
	if err != nil {
		return 0, fmt.Errorf("failed to get API key for provider %s: %w", providerCode, err)
	}

	client := newProxyAwareHTTPClient(15*time.Second, resolveProviderRouteMode(providerCode))

	probe, err := ProbeProviderConnectivity(ctx, client, apiBaseURL, apiKey, providerCode, apiFormat)
	if err != nil {
		return 0, fmt.Errorf("failed to probe models for provider %s: %w", providerCode, err)
	}

	if !probe.Success {
		return 0, fmt.Errorf("probe failed for provider %s: %s", providerCode, probe.ErrorMsg)
	}

	syncedCount, err := s.upsertModels(ctx, providerCode, probe.Models)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert models for provider %s: %w", providerCode, err)
	}

	logger.LogInfo(ctx, "provider_model_sync", "Synced provider models", map[string]interface{}{
		"provider_code":  providerCode,
		"models_count":   len(probe.Models),
		"upserted_count": syncedCount,
	})

	return syncedCount, nil
}

func (s *ProviderModelSyncService) getDecryptedAPIKey(ctx context.Context, providerCode string, apiKeyID int) (string, error) {
	db := config.GetDB()

	var encryptedKey string
	var query string
	var args []interface{}

	if apiKeyID > 0 {
		query = `SELECT api_key_encrypted FROM merchant_api_keys WHERE id = $1 AND provider = $2 AND status = 'active'`
		args = []interface{}{apiKeyID, providerCode}
	} else {
		query = `SELECT api_key_encrypted FROM merchant_api_keys WHERE provider = $1 AND status = 'active' ORDER BY id LIMIT 1`
		args = []interface{}{providerCode}
	}

	err := db.QueryRowContext(ctx, query, args...).Scan(&encryptedKey)
	if err != nil {
		return "", fmt.Errorf("no active API key found for provider %s", providerCode)
	}

	decrypted, err := utils.Decrypt(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return decrypted, nil
}

func (s *ProviderModelSyncService) upsertModels(ctx context.Context, providerCode string, modelIDs []string) (int, error) {
	db := config.GetDB()
	now := time.Now()
	syncedCount := 0

	for _, modelID := range modelIDs {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}

		var existingID int
		err := db.QueryRowContext(ctx,
			`SELECT id FROM provider_models WHERE provider_code = $1 AND model_id = $2`,
			providerCode, modelID,
		).Scan(&existingID)

		if err == nil {
			_, err = db.ExecContext(ctx,
				`UPDATE provider_models SET is_active = true, synced_at = $1, updated_at = $1
				 WHERE id = $2`,
				now, existingID,
			)
		} else {
			_, err = db.ExecContext(ctx,
				`INSERT INTO provider_models (provider_code, model_id, is_active, synced_at, created_at, updated_at)
				 VALUES ($1, $2, true, $3, $3, $3)`,
				providerCode, modelID, now,
			)
		}

		if err != nil {
			logger.LogError(ctx, "provider_model_sync", "Failed to upsert model", err, map[string]interface{}{
				"provider_code": providerCode,
				"model_id":      modelID,
			})
			continue
		}
		syncedCount++
	}

	_, _ = db.ExecContext(ctx,
		`UPDATE provider_models SET is_active = false, updated_at = $1
		 WHERE provider_code = $2 AND synced_at < $1`,
		now, providerCode,
	)

	return syncedCount, nil
}

func (s *ProviderModelSyncService) ListProviderModels(ctx context.Context, providerCode string, activeOnly bool) ([]models.ProviderModel, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, provider_code, model_id, COALESCE(display_name, ''), 
	                  reference_input_price, reference_output_price, COALESCE(reference_currency, 'USD'),
	                  is_active, synced_at, created_at, updated_at
	           FROM provider_models WHERE provider_code = $1`
	args := []interface{}{providerCode}

	if activeOnly {
		query += ` AND is_active = true`
	}
	query += ` ORDER BY model_id`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.ProviderModel
	for rows.Next() {
		var m models.ProviderModel
		err := rows.Scan(&m.ID, &m.ProviderCode, &m.ModelID, &m.DisplayName,
			&m.ReferenceInputPrice, &m.ReferenceOutputPrice, &m.ReferenceCurrency,
			&m.IsActive, &m.SyncedAt, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			continue
		}
		result = append(result, m)
	}

	return result, nil
}

func resolveProviderRouteMode(providerCode string) string {
	region := getProviderRegionStatic(providerCode)
	if region == "" || region == regionDomestic {
		return GatewayModeDirect
	}
	httpsProxy := os.Getenv("HTTPS_PROXY")
	if httpsProxy == "" {
		httpsProxy = os.Getenv("https_proxy")
	}
	if httpsProxy != "" {
		return GatewayModeProxy
	}
	return GatewayModeDirect
}

func getProviderRegionStatic(provider string) string {
	db := config.GetDB()
	if db == nil {
		return regionDomestic
	}
	var region string
	err := db.QueryRow(
		"SELECT COALESCE(provider_region, $1) FROM model_providers WHERE code = $2",
		regionDomestic, provider,
	).Scan(&region)
	if err != nil {
		return regionDomestic
	}
	return region
}
