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
	// Pricing 来自 SPU 目录（元/1K tokens）；与商户侧计费展示一致，便于「透明定价」发现。
	Pricing *OpenAIModelPricing `json:"pricing,omitempty"`
}

// OpenAIModelPricing 目录级参考单价（非单次请求扣费结果）。
type OpenAIModelPricing struct {
	InputPer1KTokens  float64 `json:"input_per_1k_tokens"`
	OutputPer1KTokens float64 `json:"output_per_1k_tokens"`
	Currency          string  `json:"currency"`
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
			COALESCE(sp.updated_at, sp.created_at, CURRENT_TIMESTAMP) AS ts,
			sp.provider_input_rate,
			sp.provider_output_rate
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
		var inRate, outRate float64
		if err := rows.Scan(&code, &modelID, &ts, &inRate, &outRate); err != nil {
			continue
		}
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}
		id := code + "/" + modelID
		item := OpenAIModelListItem{
			ID:      id,
			Object:  "model",
			Created: ts.Unix(),
			OwnedBy: code,
		}
		item.Pricing = &OpenAIModelPricing{
			InputPer1KTokens:  inRate,
			OutputPer1KTokens: outRate,
			Currency:          "CNY",
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &OpenAIModelsListResponse{Object: "list", Data: out}, nil
}
