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
)

type ProviderModelSyncService struct{}

func NewProviderModelSyncService() *ProviderModelSyncService {
	return &ProviderModelSyncService{}
}

func (s *ProviderModelSyncService) SyncProviderModels(ctx context.Context, providerCode string) (int, error) {
	db := config.GetDB()
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	var apiBaseURL, apiFormat string
	err := db.QueryRowContext(ctx,
		`SELECT COALESCE(NULLIF(TRIM(api_base_url), ''), ''), COALESCE(api_format, 'openai')
		 FROM model_providers WHERE code = $1 AND status = 'active'`,
		providerCode,
	).Scan(&apiBaseURL, &apiFormat)
	if err != nil {
		return 0, fmt.Errorf("provider %s not found or inactive: %w", providerCode, err)
	}

	if apiBaseURL == "" {
		return 0, fmt.Errorf("provider %s has no api_base_url configured", providerCode)
	}

	modelsURL := buildModelsURL(apiBaseURL)

	client := newProxyAwareHTTPClient(15*time.Second, resolveProviderRouteMode(providerCode))

	probe, err := ProbeProviderModels(ctx, client, modelsURL, "", providerCode)
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

func buildModelsURL(apiBaseURL string) string {
	b := strings.TrimRight(apiBaseURL, "/")
	if hasOpenAICompatVersionedRootSuffix(b) {
		return b + "/models"
	}
	return b + "/v1/models"
}

func resolveProviderRouteMode(providerCode string) string {
	region := getProviderRegionStatic(providerCode)
	if region == "" || region == "domestic" {
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
		return "domestic"
	}
	var region string
	err := db.QueryRow(
		"SELECT COALESCE(provider_region, 'domestic') FROM model_providers WHERE code = $1",
		provider,
	).Scan(&region)
	if err != nil {
		return "domestic"
	}
	return region
}
