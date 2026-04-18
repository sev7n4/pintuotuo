package services

import (
	"database/sql"
	"os"
	"strings"
	"time"
)

// EntitlementEnforcementStrict is true when API calls must match an active order/subscription SPU for (provider, model).
func EntitlementEnforcementStrict() bool {
	return strings.TrimSpace(strings.ToLower(entitlementEnforcementEnv())) == "strict"
}

func entitlementEnforcementEnv() string {
	// Lazy read for tests (can override via env in integration tests).
	if v := strings.TrimSpace(entitlementEnforcementOverride); v != "" {
		return v
	}
	// Avoid import cycle: read env here; config can wire later.
	return getEntitlementEnforcementFromEnv()
}

func getEntitlementEnforcementFromEnv() string {
	return os.Getenv("ENTITLEMENT_ENFORCEMENT")
}

// entitlementEnforcementOverride is set by tests only.
var entitlementEnforcementOverride string

// SetEntitlementEnforcementForTest sets ENTITLEMENT_ENFORCEMENT for unit tests (empty restores default env read).
func SetEntitlementEnforcementForTest(v string) {
	entitlementEnforcementOverride = v
}

// ResolveChosenPricingVersion returns the pricing_version_id to use for this request under strict entitlement:
// union of fulfilled paid orders and active subscriptions; pick row with max anchor_time.
// ok is false when no matching entitlement exists (caller should 403).
//
// SSOT：与 ListOpenAIModelsEntitledForUser 使用同一套订单/订阅→SKU→SPU 范围；GET /v1/models 在 strict 下列出的模型须能在此命中。
func ResolveChosenPricingVersion(db *sql.DB, userID int, provider, model string) (versionID int, anchor time.Time, ok bool, err error) {
	pv := normalizeProvider(provider)
	mv := strings.TrimSpace(model)
	if pv == "" || mv == "" {
		return 0, time.Time{}, false, nil
	}

	const q = `
WITH candidates AS (
  SELECT o.pricing_version_id AS pvid,
         o.fulfilled_at::timestamptz AS anchor_t
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
  SELECT us.pricing_version_id,
         coalesce(us.entitlement_anchor_at, us.updated_at)
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
SELECT pvid, anchor_t FROM candidates
ORDER BY anchor_t DESC
LIMIT 1`

	var pvid sql.NullInt64
	var at sql.NullTime
	err = db.QueryRow(q, userID, pv, mv).Scan(&pvid, &at)
	if err == sql.ErrNoRows {
		return 0, time.Time{}, false, nil
	}
	if err != nil {
		return 0, time.Time{}, false, err
	}
	if !pvid.Valid {
		return 0, time.Time{}, false, nil
	}
	if !at.Valid {
		return int(pvid.Int64), time.Now().UTC(), true, nil
	}
	return int(pvid.Int64), at.Time.UTC(), true, nil
}

func normalizeProvider(p string) string {
	return strings.ToLower(strings.TrimSpace(p))
}
