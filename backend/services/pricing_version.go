package services

import (
	"database/sql"

	"github.com/lib/pq"
)

const pricingVersionCodeBaseline = "baseline"

// BaselinePricingVersionID returns the id of the baseline retail pricing version (migration 045), or invalid if missing.
func BaselinePricingVersionID(q interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}) sql.NullInt64 {
	var id sql.NullInt64
	err := q.QueryRow(
		`SELECT id FROM pricing_versions WHERE code = $1 LIMIT 1`,
		pricingVersionCodeBaseline,
	).Scan(&id)
	if err != nil {
		return sql.NullInt64{}
	}
	return id
}

// LatestUserPricingVersionID returns the pricing_version_id from the user's most recent fulfilled paid order (IE-4).
func LatestUserPricingVersionID(db *sql.DB, userID int) sql.NullInt64 {
	var id sql.NullInt64
	err := db.QueryRow(`
		SELECT pricing_version_id FROM orders
		WHERE user_id = $1 AND status = 'paid' AND fulfilled_at IS NOT NULL
		  AND pricing_version_id IS NOT NULL
		ORDER BY fulfilled_at DESC
		LIMIT 1
	`, userID).Scan(&id)
	if err != nil {
		return sql.NullInt64{}
	}
	return id
}

// CostFromPer1KRates applies 元/1K tokens to usage (same formula as PricingService.CalculateCost).
func CostFromPer1KRates(inputPrice, outputPrice float64, inputTokens, outputTokens int) float64 {
	return float64(inputTokens)*inputPrice/1000 + float64(outputTokens)*outputPrice/1000
}

// CalculateCostFromPricingVersion loads snapshot rates for provider/model from pricing_version_spu_rates + spus.
// Model matches provider_model_id (preferred) or model_name, aligned with api-usage-guide and entitlement resolution.
func CalculateCostFromPricingVersion(db *sql.DB, versionID int, provider, model string, inputTokens, outputTokens int) (float64, bool) {
	var inRate, outRate float64
	err := db.QueryRow(`
		SELECT r.provider_input_rate, r.provider_output_rate
		FROM pricing_version_spu_rates r
		INNER JOIN spus p ON p.id = r.spu_id
		WHERE r.pricing_version_id = $1
		  AND lower(trim(p.model_provider)) = lower(trim($2::text))
		  AND (
		    lower(trim(coalesce(nullif(trim(p.provider_model_id), ''), p.model_name))) = lower(trim($3::text))
		    OR lower(trim(p.model_name)) = lower(trim($3::text))
		  )
		  AND p.status = 'active'
		LIMIT 1
	`, versionID, provider, model).Scan(&inRate, &outRate)
	if err != nil {
		return 0, false
	}
	return CostFromPer1KRates(inRate, outRate, inputTokens, outputTokens), true
}

// SPUPer1KSnapshot is per-SPU input/output rate (元/1K tokens) in a pricing_version snapshot row.
type SPUPer1KSnapshot struct {
	InputPer1K  float64
	OutputPer1K float64
}

// LockedSPUPer1KSnapshots loads snapshot rates for many SPUs in one query; SPUs without a row are omitted.
func LockedSPUPer1KSnapshots(db *sql.DB, pricingVersionID int64, spuIDs []int) (map[int]SPUPer1KSnapshot, error) {
	if len(spuIDs) == 0 {
		return map[int]SPUPer1KSnapshot{}, nil
	}
	rows, err := db.Query(`
		SELECT spu_id, COALESCE(provider_input_rate, 0), COALESCE(provider_output_rate, 0)
		FROM pricing_version_spu_rates
		WHERE pricing_version_id = $1 AND spu_id = ANY($2::int[])
	`, pricingVersionID, pq.Array(spuIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int]SPUPer1KSnapshot)
	for rows.Next() {
		var sid int
		var inR, outR float64
		if err := rows.Scan(&sid, &inR, &outR); err != nil {
			return nil, err
		}
		out[sid] = SPUPer1KSnapshot{InputPer1K: inR, OutputPer1K: outR}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
