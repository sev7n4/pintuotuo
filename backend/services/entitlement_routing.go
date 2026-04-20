package services

import (
	"database/sql"
	"fmt"
	"strings"
)

// EntitlementRoutingContext holds API keys the caller may use under strict entitlement:
// order/subscription → sku → merchant_skus → merchant_api_keys (same SPU scope as ResolveChosenPricingVersion).
type EntitlementRoutingContext struct {
	// AllowedAPIKeyIDs is the union of keys reachable from the chosen entitled SKU row(s).
	AllowedAPIKeyIDs map[int]struct{}
	// AllowedMerchantSKUIDs is merchant_skus.id tied to those keys (same scope).
	AllowedMerchantSKUIDs map[int]struct{}
	// SelectedSKU is the platform sku_id whose merchant_skus produced the first non-empty allowlist.
	SelectedSKU int
	// APIKeyToMerchantSKU picks one merchant_sku per api_key for procurement / logging (stable: min ms.id).
	APIKeyToMerchantSKU map[int]int
}

// AllowsAPIKey reports whether id is in the entitlement allowlist (strict routing).
func (e *EntitlementRoutingContext) AllowsAPIKey(id int) bool {
	if e == nil || e.AllowedAPIKeyIDs == nil {
		return false
	}
	_, ok := e.AllowedAPIKeyIDs[id]
	return ok
}

// AllowsMerchantSKU reports whether merchant_sku_id is allowed for this request.
func (e *EntitlementRoutingContext) AllowsMerchantSKU(msID int) bool {
	if e == nil || e.AllowedMerchantSKUIDs == nil {
		return false
	}
	_, ok := e.AllowedMerchantSKUIDs[msID]
	return ok
}

// MerchantSKUForAPIKey returns the merchant_sku_id to log for a chosen api_key_id, if known.
func (e *EntitlementRoutingContext) MerchantSKUForAPIKey(apiKeyID int) (int, bool) {
	if e == nil || e.APIKeyToMerchantSKU == nil {
		return 0, false
	}
	v, ok := e.APIKeyToMerchantSKU[apiKeyID]
	return v, ok
}

// ResolveEntitlementRoutingContext walks entitled SKUs for (provider, model), ordered by quota score (desc),
// and returns the first SKU that has at least one operational merchant_sku → api_key edge.
// If no SKU yields keys, AllowedAPIKeyIDs is empty (caller should fail closed).
func ResolveEntitlementRoutingContext(db *sql.DB, userID int, provider, model string) (*EntitlementRoutingContext, error) {
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	pv := normalizeProvider(provider)
	mv := strings.TrimSpace(model)
	if pv == "" || mv == "" {
		return &EntitlementRoutingContext{
			AllowedAPIKeyIDs:      map[int]struct{}{},
			AllowedMerchantSKUIDs: map[int]struct{}{},
			APIKeyToMerchantSKU:   map[int]int{},
		}, nil
	}

	skus, err := listEntitledSKUsOrderedByQuota(db, userID, pv, mv)
	if err != nil {
		return nil, err
	}
	if len(skus) == 0 {
		return &EntitlementRoutingContext{
			AllowedAPIKeyIDs:      map[int]struct{}{},
			AllowedMerchantSKUIDs: map[int]struct{}{},
			APIKeyToMerchantSKU:   map[int]int{},
		}, nil
	}

	for _, row := range skus {
		edges, qerr := buildAllowlistEdgesForSKU(db, row.SkuID, pv)
		if qerr != nil {
			return nil, qerr
		}
		if len(edges) == 0 {
			continue
		}
		ctx := &EntitlementRoutingContext{
			AllowedAPIKeyIDs:      make(map[int]struct{}),
			AllowedMerchantSKUIDs: make(map[int]struct{}),
			SelectedSKU:           row.SkuID,
			APIKeyToMerchantSKU:   make(map[int]int),
		}
		for _, e := range edges {
			ctx.AllowedAPIKeyIDs[e.APIKeyID] = struct{}{}
			ctx.AllowedMerchantSKUIDs[e.MerchantSKUID] = struct{}{}
			if prev, ok := ctx.APIKeyToMerchantSKU[e.APIKeyID]; !ok || e.MerchantSKUID < prev {
				ctx.APIKeyToMerchantSKU[e.APIKeyID] = e.MerchantSKUID
			}
		}
		return ctx, nil
	}

	return &EntitlementRoutingContext{
		AllowedAPIKeyIDs:      map[int]struct{}{},
		AllowedMerchantSKUIDs: map[int]struct{}{},
		APIKeyToMerchantSKU:   map[int]int{},
	}, nil
}

type entitledSKUQuotaRow struct {
	SkuID      int
	QuotaScore float64
}

type entitlementKeyEdge struct {
	MerchantSKUID int
	MerchantID    int
	APIKeyID      int
}

func listEntitledSKUsOrderedByQuota(db *sql.DB, userID int, providerNorm, modelTrim string) ([]entitledSKUQuotaRow, error) {
	// SSOT: same SPU match as ResolveChosenPricingVersion; quota = sum(order line qty) + subscription token remainder.
	const q = `
WITH matched_lines AS (
  SELECT oi.sku_id, oi.quantity::double precision AS q
  FROM orders o
  INNER JOIN order_items oi ON oi.order_id = o.id
  INNER JOIN skus s ON s.id = oi.sku_id
  INNER JOIN spus sp ON sp.id = s.spu_id
  WHERE o.user_id = $1
    AND o.status = 'paid'
    AND o.fulfilled_at IS NOT NULL
    AND o.pricing_version_id IS NOT NULL
    AND sp.status = 'active'
    AND lower(trim(sp.model_provider)) = lower(trim($2::text))
    AND (
      lower(trim(coalesce(nullif(trim(sp.provider_model_id), ''), sp.model_name))) = lower(trim($3::text))
      OR lower(trim(sp.model_name)) = lower(trim($3::text))
    )
  UNION ALL
  SELECT us.sku_id,
    GREATEST(0.0, COALESCE(s.token_amount::double precision, 0) - COALESCE(us.used_tokens::double precision, 0))
  FROM user_subscriptions us
  INNER JOIN skus s ON s.id = us.sku_id
  INNER JOIN spus sp ON sp.id = s.spu_id
  WHERE us.user_id = $1
    AND us.status = 'active'
    AND us.end_date >= CURRENT_DATE
    AND us.pricing_version_id IS NOT NULL
    AND sp.status = 'active'
    AND lower(trim(sp.model_provider)) = lower(trim($2::text))
    AND (
      lower(trim(coalesce(nullif(trim(sp.provider_model_id), ''), sp.model_name))) = lower(trim($3::text))
      OR lower(trim(sp.model_name)) = lower(trim($3::text))
    )
)
SELECT sku_id, COALESCE(SUM(q), 0)::double precision AS quota_score
FROM matched_lines
GROUP BY sku_id
ORDER BY quota_score DESC, sku_id ASC`

	rows, err := db.Query(q, userID, providerNorm, modelTrim)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entitledSKUQuotaRow
	for rows.Next() {
		var r entitledSKUQuotaRow
		if err := rows.Scan(&r.SkuID, &r.QuotaScore); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func buildAllowlistEdgesForSKU(db *sql.DB, skuID int, providerNorm string) ([]entitlementKeyEdge, error) {
	const q = `
SELECT ms.id, ms.merchant_id, mak.id
FROM merchant_skus ms
INNER JOIN merchant_api_keys mak ON mak.id = ms.api_key_id
INNER JOIN merchants m ON m.id = ms.merchant_id
  AND m.status IN ('active', 'approved')
  AND m.lifecycle_status <> 'suspended'
WHERE ms.sku_id = $1
  AND ms.status = 'active'
  AND ms.api_key_id IS NOT NULL
  AND lower(trim(mak.provider)) = lower(trim($2::text))
  AND mak.status = 'active'
  AND (mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')
  AND mak.health_status IN ('healthy', 'degraded')
  AND (mak.quota_limit IS NULL OR mak.quota_used < mak.quota_limit)
ORDER BY ms.id ASC`

	rows, err := db.Query(q, skuID, providerNorm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []entitlementKeyEdge
	for rows.Next() {
		var e entitlementKeyEdge
		if err := rows.Scan(&e.MerchantSKUID, &e.MerchantID, &e.APIKeyID); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}
