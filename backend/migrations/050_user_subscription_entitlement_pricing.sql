-- Subscription entitlement anchors for strict API pricing (plan: 权益白名单)
-- Version: 050

ALTER TABLE user_subscriptions
  ADD COLUMN IF NOT EXISTS pricing_version_id INT REFERENCES pricing_versions(id),
  ADD COLUMN IF NOT EXISTS entitlement_anchor_at TIMESTAMPTZ;

COMMENT ON COLUMN user_subscriptions.pricing_version_id IS 'API 计价快照：首期履约或自动续费时写入，与 orders.pricing_version_id 语义一致';
COMMENT ON COLUMN user_subscriptions.entitlement_anchor_at IS '权益时间锚：重叠 (provider,model) 时取最新锚点对应价目';

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_pricing_version ON user_subscriptions(pricing_version_id);

-- Backfill pricing_version_id from latest fulfilled paid order per (user_id, sku_id)
UPDATE user_subscriptions us
SET pricing_version_id = o.pricing_version_id
FROM (
  SELECT DISTINCT ON (user_id, sku_id)
    user_id, sku_id, pricing_version_id
  FROM orders
  WHERE status = 'paid'
    AND fulfilled_at IS NOT NULL
    AND pricing_version_id IS NOT NULL
    AND sku_id IS NOT NULL
  ORDER BY user_id, sku_id, fulfilled_at DESC
) o
WHERE us.user_id = o.user_id
  AND us.sku_id = o.sku_id
  AND us.pricing_version_id IS NULL;

-- Remaining rows: baseline snapshot
UPDATE user_subscriptions us
SET pricing_version_id = pv.id
FROM pricing_versions pv
WHERE pv.code = 'baseline'
  AND us.pricing_version_id IS NULL;

-- Anchor defaults
UPDATE user_subscriptions
SET entitlement_anchor_at = updated_at
WHERE entitlement_anchor_at IS NULL;
