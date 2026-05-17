package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/logger"
)

// CatalogGapItem is one model id in a provider catalog diff.
type CatalogGapItem struct {
	ModelID     string `json:"model_id"`
	DisplayName string `json:"display_name,omitempty"`
}

// CatalogGapStaleSPU is an active SPU whose model id no longer appears in the latest provider_models sync.
type CatalogGapStaleSPU struct {
	SPUID          int    `json:"spu_id"`
	SPUCode        string `json:"spu_code"`
	Name           string `json:"name"`
	ModelID        string `json:"model_id"`
	ActiveSKUCount int    `json:"active_sku_count"`
}

// ProviderCatalogGap compares provider_models (upstream snapshot) with spus (sellable catalog).
type ProviderCatalogGap struct {
	ProviderCode       string               `json:"provider_code"`
	LastSyncedAt       *time.Time           `json:"last_synced_at,omitempty"`
	ProviderModelCount int                  `json:"provider_model_count"`
	SPUModelCount      int                  `json:"spu_model_count"`
	PendingOnboard     []CatalogGapItem     `json:"pending_onboard"`
	StaleSPUs          []CatalogGapStaleSPU `json:"stale_spus"`
}

// SPUDraftFromProviderModel is the result of creating an inactive SPU draft from provider_models.
type SPUDraftFromProviderModel struct {
	ModelID string `json:"model_id"`
	SPUID   int    `json:"spu_id"`
	SPUCode string `json:"spu_code"`
	Skipped bool   `json:"skipped,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type ProviderCatalogGapService struct {
	sync *ProviderModelSyncService
}

func NewProviderCatalogGapService() *ProviderCatalogGapService {
	return &ProviderCatalogGapService{sync: NewProviderModelSyncService()}
}

func normalizeCatalogModelID(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func spuCatalogModelID(providerModelID, modelName string) string {
	if v := strings.TrimSpace(providerModelID); v != "" {
		return v
	}
	return strings.TrimSpace(modelName)
}

// CompareProviderCatalog lists models in provider_models not covered by any SPU, and active SPUs missing from latest sync.
func (s *ProviderCatalogGapService) CompareProviderCatalog(ctx context.Context, providerCode string) (*ProviderCatalogGap, error) {
	return CompareProviderCatalogDB(ctx, config.GetDB(), providerCode)
}

// CompareProviderCatalogDB is the testable entry using an explicit database handle.
func CompareProviderCatalogDB(ctx context.Context, db *sql.DB, providerCode string) (*ProviderCatalogGap, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	providerCode = strings.TrimSpace(providerCode)
	if providerCode == "" {
		return nil, fmt.Errorf("provider code is required")
	}

	gap := &ProviderCatalogGap{ProviderCode: providerCode}

	var lastSync sql.NullTime
	_ = db.QueryRowContext(ctx,
		`SELECT MAX(synced_at) FROM provider_models WHERE provider_code = $1`,
		providerCode,
	).Scan(&lastSync)
	if lastSync.Valid {
		gap.LastSyncedAt = &lastSync.Time
	}

	providerSet, err := loadProviderModelIDSet(ctx, db, providerCode, true)
	if err != nil {
		return nil, err
	}
	gap.ProviderModelCount = len(providerSet)

	spuByModel, err := loadSPUModelIndex(ctx, db, providerCode)
	if err != nil {
		return nil, err
	}
	gap.SPUModelCount = len(spuByModel)

	for _, pm := range providerSet {
		if _, covered := spuByModel[normalizeCatalogModelID(pm.modelID)]; covered {
			continue
		}
		gap.PendingOnboard = append(gap.PendingOnboard, CatalogGapItem{
			ModelID:     pm.modelID,
			DisplayName: pm.displayName,
		})
	}

	// Only flag stale SPUs when we have a provider_models snapshot (avoid noise before first sync).
	if len(providerSet) > 0 {
		for normID, spu := range spuByModel {
			if spu.status != "active" {
				continue
			}
			if _, ok := providerSet[normID]; ok {
				continue
			}
			gap.StaleSPUs = append(gap.StaleSPUs, CatalogGapStaleSPU{
				SPUID:          spu.id,
				SPUCode:        spu.spuCode,
				Name:           spu.name,
				ModelID:        spu.modelID,
				ActiveSKUCount: spu.activeSKUCount,
			})
		}
	}

	return gap, nil
}

type providerModelEntry struct {
	modelID     string
	displayName string
}

type spuModelEntry struct {
	id             int
	spuCode        string
	name           string
	modelID        string
	status         string
	activeSKUCount int
}

func loadProviderModelIDSet(ctx context.Context, db *sql.DB, providerCode string, activeOnly bool) (map[string]providerModelEntry, error) {
	q := `SELECT model_id, COALESCE(display_name, '') FROM provider_models WHERE provider_code = $1`
	if activeOnly {
		q += ` AND is_active = true`
	}
	rows, err := db.QueryContext(ctx, q, providerCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]providerModelEntry)
	for rows.Next() {
		var modelID, displayName string
		if err := rows.Scan(&modelID, &displayName); err != nil {
			continue
		}
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}
		out[normalizeCatalogModelID(modelID)] = providerModelEntry{modelID: modelID, displayName: displayName}
	}
	return out, rows.Err()
}

func loadSPUModelIndex(ctx context.Context, db *sql.DB, providerCode string) (map[string]spuModelEntry, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT sp.id, sp.spu_code, sp.name,
		       TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name)) AS model_id,
		       sp.status,
		       (SELECT COUNT(*)::int FROM skus s WHERE s.spu_id = sp.id AND s.status = 'active')
		FROM spus sp
		WHERE lower(trim(sp.model_provider)) = lower(trim($1))
		  AND TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name)) <> ''`,
		providerCode,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]spuModelEntry)
	for rows.Next() {
		var e spuModelEntry
		if err := rows.Scan(&e.id, &e.spuCode, &e.name, &e.modelID, &e.status, &e.activeSKUCount); err != nil {
			continue
		}
		e.modelID = strings.TrimSpace(e.modelID)
		if e.modelID == "" {
			continue
		}
		out[normalizeCatalogModelID(e.modelID)] = e
	}
	return out, rows.Err()
}

// CreateSPUDraftsFromProviderModels creates inactive SPUs for models not yet in catalog (idempotent per model_id).
func (s *ProviderCatalogGapService) CreateSPUDraftsFromProviderModels(ctx context.Context, providerCode string, modelIDs []string) ([]SPUDraftFromProviderModel, error) {
	return CreateSPUDraftsFromProviderModelsDB(ctx, config.GetDB(), providerCode, modelIDs)
}

// CreateSPUDraftsFromProviderModelsDB is the testable entry using an explicit database handle.
func CreateSPUDraftsFromProviderModelsDB(ctx context.Context, db *sql.DB, providerCode string, modelIDs []string) ([]SPUDraftFromProviderModel, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	providerCode = strings.TrimSpace(providerCode)
	if providerCode == "" {
		return nil, fmt.Errorf("provider code is required")
	}
	if len(modelIDs) == 0 {
		return nil, fmt.Errorf("model_ids is required")
	}

	spuIndex, err := loadSPUModelIndex(ctx, db, providerCode)
	if err != nil {
		return nil, err
	}

	var providerName string
	_ = db.QueryRowContext(ctx,
		`SELECT COALESCE(name, code) FROM model_providers WHERE code = $1`,
		providerCode,
	).Scan(&providerName)

	var results []SPUDraftFromProviderModel
	for _, rawID := range modelIDs {
		modelID := strings.TrimSpace(rawID)
		if modelID == "" {
			continue
		}
		if _, exists := spuIndex[normalizeCatalogModelID(modelID)]; exists {
			results = append(results, SPUDraftFromProviderModel{
				ModelID: modelID,
				Skipped: true,
				Reason:  "spu_already_exists",
			})
			continue
		}

		spuCode, err := uniqueDraftSPUCode(ctx, db, providerCode, modelID)
		if err != nil {
			return results, err
		}
		name := fmt.Sprintf("%s %s（草稿）", strings.TrimSpace(providerName), modelID)
		if len(name) > 200 {
			name = name[:200]
		}

		var spuID int
		err = db.QueryRowContext(ctx, `
			INSERT INTO spus (
				spu_code, name, model_provider, model_name, provider_model_id,
				model_tier, base_compute_points, status, sort_order
			) VALUES ($1, $2, $3, $4, $5, 'lite', 1.0, 'inactive', 0)
			RETURNING id`,
			spuCode, name, providerCode, modelID, modelID,
		).Scan(&spuID)
		if err != nil {
			return results, fmt.Errorf("create draft spu for %s: %w", modelID, err)
		}

		spuIndex[normalizeCatalogModelID(modelID)] = spuModelEntry{id: spuID, modelID: modelID}
		results = append(results, SPUDraftFromProviderModel{
			ModelID: modelID,
			SPUID:   spuID,
			SPUCode: spuCode,
		})
	}

	return results, nil
}

func uniqueDraftSPUCode(ctx context.Context, db *sql.DB, providerCode, modelID string) (string, error) {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '-', r == '_', r == '.':
			return r
		default:
			return '-'
		}
	}, modelID)
	safe = strings.Trim(safe, "-")
	if safe == "" {
		safe = "model"
	}
	if len(safe) > 40 {
		safe = safe[:40]
	}
	base := fmt.Sprintf("SPU-%s-%s", strings.ToUpper(providerCode), strings.ToUpper(safe))
	code := base
	for i := 0; i < 5; i++ {
		var exists int
		err := db.QueryRowContext(ctx, `SELECT 1 FROM spus WHERE spu_code = $1`, code).Scan(&exists)
		if err == sql.ErrNoRows {
			return code, nil
		}
		if err != nil {
			return "", err
		}
		code = fmt.Sprintf("%s-%d", base, i+2)
	}
	return "", fmt.Errorf("could not allocate unique spu_code for %s", modelID)
}

// SyncAllActiveProviderModels runs upstream sync for each active provider with api_base_url (scheduler / manual).
func (s *ProviderCatalogGapService) SyncAllActiveProviderModels(ctx context.Context) (map[string]int, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT code FROM model_providers
		WHERE status = 'active'
		  AND code <> '__default__'
		  AND COALESCE(NULLIF(TRIM(api_base_url), ''), '') <> ''
		ORDER BY sort_order NULLS LAST, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]int)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			continue
		}
		count, syncErr := s.sync.SyncProviderModels(ctx, code, 0)
		if syncErr != nil {
			logger.LogError(ctx, "provider_catalog_sync", "scheduled sync failed", syncErr, map[string]interface{}{
				"provider_code": code,
			})
			results[code] = -1
			continue
		}
		results[code] = count
		logger.LogInfo(ctx, "provider_catalog_sync", "scheduled sync ok", map[string]interface{}{
			"provider_code": code,
			"synced_count":  count,
		})
	}
	return results, rows.Err()
}
