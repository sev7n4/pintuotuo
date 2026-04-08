package services

import (
	"database/sql"
	"strings"

	"github.com/lib/pq"
)

// ResolveOpenAICompatModel maps OpenAI-SDK style "model" to (provider_code, upstream_model_id).
// 1) "provider/model" → lowercase provider code + remainder.
// 2) Else load model_providers.compat_prefixes (active rows); longest prefix wins (case-insensitive).
// 3) Else default provider "openai".
func ResolveOpenAICompatModel(db *sql.DB, model string) (provider string, modelName string) {
	model = strings.TrimSpace(model)
	if model == "" {
		return "", ""
	}
	if idx := strings.Index(model, "/"); idx > 0 {
		return strings.ToLower(strings.TrimSpace(model[:idx])), strings.TrimSpace(model[idx+1:])
	}

	modelLower := strings.ToLower(model)
	if db == nil {
		return modelProviderOpenAI, model
	}

	rows, err := db.Query(`
		SELECT code, compat_prefixes
		FROM model_providers
		WHERE status = 'active' AND cardinality(compat_prefixes) > 0`)
	if err != nil {
		return modelProviderOpenAI, model
	}
	defer rows.Close()

	bestLen := 0
	bestCode := ""
	for rows.Next() {
		var code string
		var prefixes pq.StringArray
		if scanErr := rows.Scan(&code, &prefixes); scanErr != nil {
			continue
		}
		for _, pfx := range prefixes {
			p := strings.ToLower(strings.TrimSpace(pfx))
			if p == "" {
				continue
			}
			if strings.HasPrefix(modelLower, p) && len(p) > bestLen {
				bestLen = len(p)
				bestCode = code
			}
		}
	}
	if bestCode != "" {
		return bestCode, model
	}
	return modelProviderOpenAI, model
}
