package services

import (
	"context"
	"database/sql"
	"strings"

	"github.com/pintuotuo/backend/config"
)

// ListCatalogModelIDsForProvider returns active SPU model ids for a provider (same SSOT as ListOpenAIModelsFromCatalog).
func ListCatalogModelIDsForProvider(ctx context.Context, provider string) []string {
	return ListCatalogModelIDsForProviderDB(ctx, config.GetDB(), provider)
}

// ListCatalogModelIDsForProviderDB is the testable entry using an explicit database handle.
func ListCatalogModelIDsForProviderDB(ctx context.Context, db *sql.DB, provider string) []string {
	if db == nil {
		return nil
	}
	providers := catalogProviderCodes(provider)
	if len(providers) == 0 {
		return nil
	}
	placeholders := make([]string, len(providers))
	args := make([]interface{}, len(providers))
	for i, p := range providers {
		placeholders[i] = "$" + string(rune('1'+i))
		args[i] = p
	}
	// Build query with dynamic IN - use simple loop for small provider list
	var models []string
	seen := make(map[string]struct{})
	for _, pcode := range providers {
		rows, err := db.QueryContext(ctx, `
			SELECT TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name))
			FROM spus sp
			INNER JOIN model_providers mp ON mp.code = sp.model_provider AND mp.status = 'active'
			WHERE sp.status = 'active'
			  AND lower(trim(sp.model_provider)) = lower(trim($1))
			  AND TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name)) <> ''
			ORDER BY COALESCE(sp.sort_order, 0), sp.id`,
			pcode,
		)
		if err != nil {
			continue
		}
		for rows.Next() {
			var modelID string
			if err := rows.Scan(&modelID); err != nil {
				continue
			}
			modelID = strings.TrimSpace(modelID)
			if modelID == "" {
				continue
			}
			if _, ok := seen[modelID]; ok {
				continue
			}
			seen[modelID] = struct{}{}
			models = append(models, modelID)
		}
		rows.Close()
	}
	return models
}

func catalogProviderCodes(provider string) []string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return nil
	}
	out := []string{provider}
	if strings.HasSuffix(provider, AnthropicSiblingProviderSuffix) {
		primary := strings.TrimSuffix(provider, AnthropicSiblingProviderSuffix)
		if primary != "" && primary != provider {
			out = append(out, primary)
		}
	}
	return out
}

// ProbeCatalogModelsForProvider returns platform catalog model ids for probes, with code predefined as last resort.
func ProbeCatalogModelsForProvider(ctx context.Context, provider string) []string {
	if models := ListCatalogModelIDsForProvider(ctx, provider); len(models) > 0 {
		return models
	}
	return GetPredefinedModels(provider)
}

// CatalogProbeCandidates returns ordered model ids to try for litellm chat path probes (DB catalog first).
func CatalogProbeCandidates(ctx context.Context, provider string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(m string) {
		m = strings.TrimSpace(m)
		if m == "" {
			return
		}
		if _, ok := seen[m]; ok {
			return
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	for _, m := range ListCatalogModelIDsForProvider(ctx, provider) {
		add(m)
	}
	for _, m := range GetPredefinedModels(provider) {
		add(m)
	}
	return out
}
