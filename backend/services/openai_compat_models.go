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

// ListOpenAIModelsFromCatalog returns active SPUs joined to active model_providers（全局上架目录，非按用户过滤）。
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

// ListOpenAIModelsEntitledForUser returns OpenAI-style model ids the user may call under strict entitlement:
// union of SPUs reachable via paid+fulfilled orders and active subscriptions, matching ResolveChosenPricingVersion scope.
// Id format matches ListOpenAIModelsFromCatalog (provider_code + "/" + display model id).
func ListOpenAIModelsEntitledForUser(ctx context.Context, db *sql.DB, userID int) (*OpenAIModelsListResponse, error) {
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	// DISTINCT ON keeps one row per (code, model_id); prefer latest ts when duplicates exist.
	rows, err := db.QueryContext(ctx, `
SELECT DISTINCT ON (code, model_id)
  code,
  model_id,
  ts,
  provider_input_rate,
  provider_output_rate
FROM (
  SELECT DISTINCT
    mp.code,
    TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name)) AS model_id,
    COALESCE(sp.updated_at, sp.created_at, CURRENT_TIMESTAMP) AS ts,
    sp.provider_input_rate,
    sp.provider_output_rate,
    COALESCE(mp.sort_order, 0) AS mp_sort,
    COALESCE(sp.sort_order, 0) AS sp_sort,
    sp.id AS spu_id
  FROM orders o
  INNER JOIN order_items oi ON oi.order_id = o.id
  INNER JOIN skus s ON s.id = oi.sku_id
  INNER JOIN spus sp ON sp.id = s.spu_id
  INNER JOIN model_providers mp ON mp.code = sp.model_provider AND mp.status = 'active'
  WHERE o.user_id = $1
    AND o.status = 'paid'
    AND o.fulfilled_at IS NOT NULL
    AND o.pricing_version_id IS NOT NULL
    AND sp.status = 'active'
    AND TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name)) <> ''
  UNION ALL
  SELECT DISTINCT
    mp.code,
    TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name)) AS model_id,
    COALESCE(sp.updated_at, sp.created_at, CURRENT_TIMESTAMP) AS ts,
    sp.provider_input_rate,
    sp.provider_output_rate,
    COALESCE(mp.sort_order, 0) AS mp_sort,
    COALESCE(sp.sort_order, 0) AS sp_sort,
    sp.id AS spu_id
  FROM user_subscriptions us
  INNER JOIN skus s ON s.id = us.sku_id
  INNER JOIN spus sp ON sp.id = s.spu_id
  INNER JOIN model_providers mp ON mp.code = sp.model_provider AND mp.status = 'active'
  WHERE us.user_id = $1
    AND us.status = 'active'
    AND us.end_date >= CURRENT_DATE
    AND us.pricing_version_id IS NOT NULL
    AND sp.status = 'active'
    AND TRIM(COALESCE(NULLIF(TRIM(sp.provider_model_id), ''), sp.model_name)) <> ''
) ent
ORDER BY code, model_id, ts DESC, mp_sort, sp_sort, spu_id DESC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list entitled models: %w", err)
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
