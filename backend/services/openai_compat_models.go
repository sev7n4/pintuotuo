package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// OpenAIModelListItem matches the OpenAI /v1/models list entry shape used by common SDKs.
type OpenAIModelListItem struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// OpenAIModelsListResponse is the top-level envelope for GET /v1/models.
type OpenAIModelsListResponse struct {
	Object string                `json:"object"`
	Data   []OpenAIModelListItem `json:"data"`
}

// ListOpenAIModelsFromCatalog returns active SPUs joined to active model_providers.
// Model id uses "provider_code/model_name" when model_name is set; this matches ResolveOpenAICompatModel "provider/model" syntax.
func ListOpenAIModelsFromCatalog(ctx context.Context, db *sql.DB) (*OpenAIModelsListResponse, error) {
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT
			mp.code,
			COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name) AS model_id,
			COALESCE(sp.updated_at, sp.created_at, CURRENT_TIMESTAMP) AS ts
		FROM spus sp
		INNER JOIN model_providers mp ON mp.code = sp.model_provider AND mp.status = 'active'
		WHERE sp.status = 'active'
		ORDER BY mp.sort_order NULLS LAST, sp.sort_order NULLS LAST, sp.id
	`)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer rows.Close()

	var out []OpenAIModelListItem
	for rows.Next() {
		var code, modelID string
		var ts time.Time
		if err := rows.Scan(&code, &modelID, &ts); err != nil {
			continue
		}
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}
		id := code + "/" + modelID
		out = append(out, OpenAIModelListItem{
			ID:      id,
			Object:  "model",
			Created: ts.Unix(),
			OwnedBy: code,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &OpenAIModelsListResponse{Object: "list", Data: out}, nil
}
