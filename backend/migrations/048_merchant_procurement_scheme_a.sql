-- 048_merchant_procurement_scheme_a.sql
-- 商户采购成本归因：api_usage_logs 扩展 + 一 Key 仅一条 active merchant_sku

-- 0) 治理存量：同一 api_key_id 下多条 active 时，仅保留 id 最小的一条，其余置为 inactive
UPDATE merchant_skus ms
SET status = 'inactive', updated_at = CURRENT_TIMESTAMP
WHERE ms.id IN (
  SELECT ms2.id
  FROM merchant_skus ms2
  INNER JOIN (
    SELECT api_key_id, MIN(id) AS keep_id
    FROM merchant_skus
    WHERE status = 'active' AND api_key_id IS NOT NULL
    GROUP BY api_key_id
    HAVING COUNT(*) > 1
  ) d ON d.api_key_id = ms2.api_key_id
  WHERE ms2.status = 'active'
    AND ms2.id <> d.keep_id
);

-- 1) api_usage_logs：商户 SKU 归因 + 采购成本（人民币，元）
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS merchant_sku_id INT REFERENCES merchant_skus(id) ON DELETE SET NULL;
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS procurement_cost_cny DECIMAL(20, 6);
COMMENT ON COLUMN api_usage_logs.merchant_sku_id IS '本次请求对应的在售商户 SKU（可空：非商户路由或未绑定）';
COMMENT ON COLUMN api_usage_logs.procurement_cost_cny IS '按 merchant_skus 成本单价计算的采购成本（人民币元），与 cost（内部可消费单位）独立';

CREATE INDEX IF NOT EXISTS idx_api_usage_logs_merchant_sku_id ON api_usage_logs(merchant_sku_id);

-- 2) merchant_skus：同一 api_key 在 active 状态下全局唯一（含 NULL api_key_id 不受限）
CREATE UNIQUE INDEX IF NOT EXISTS ux_merchant_skus_one_active_per_api_key
  ON merchant_skus (api_key_id)
  WHERE status = 'active' AND api_key_id IS NOT NULL;

-- 3) 结算单：周期内采购成本汇总（人民币）
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS total_procurement_cny DECIMAL(20, 6);
COMMENT ON COLUMN merchant_settlements.total_procurement_cny IS '周期内 SUM(api_usage_logs.procurement_cost_cny)，可空表示历史或无采购数据';
