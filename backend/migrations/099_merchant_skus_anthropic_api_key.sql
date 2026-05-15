-- 商户 SKU：可选 Anthropic/Messages 出站专用密钥（与主 api_key_id 同商户、同 SKU 行）
-- 产品约定：密钥的 merchant_api_keys.provider 须为 {spus.model_provider}_anthropic（如 alibaba_anthropic），见代码常量。

ALTER TABLE merchant_skus
  ADD COLUMN IF NOT EXISTS anthropic_api_key_id INTEGER REFERENCES merchant_api_keys(id) ON DELETE SET NULL;

COMMENT ON COLUMN merchant_skus.anthropic_api_key_id IS '可选：Claude/Anthropic 入口出站专用 Key；须为 {model_provider}_anthropic 且与同 SKU 主 api_key_id 同属一商户。';

CREATE UNIQUE INDEX IF NOT EXISTS ux_merchant_skus_one_active_per_anthropic_api_key
  ON merchant_skus (anthropic_api_key_id)
  WHERE status = 'active' AND anthropic_api_key_id IS NOT NULL;

-- PostgreSQL 不允许 CREATE OR REPLACE VIEW 在中间插入列（会报 cannot change name of view column）。
DROP VIEW IF EXISTS merchant_sku_details;

CREATE VIEW merchant_sku_details AS
SELECT
  ms.id,
  ms.merchant_id,
  ms.sku_id,
  ms.api_key_id,
  ms.anthropic_api_key_id,
  ms.status,
  ms.sales_count,
  ms.total_sales_amount,
  ms.created_at,
  ms.updated_at,
  s.sku_code,
  s.sku_type,
  s.token_amount,
  s.compute_points,
  s.retail_price,
  s.original_price,
  s.valid_days,
  s.group_enabled,
  s.group_discount_rate,
  sp.name AS spu_name,
  sp.model_provider,
  sp.model_name,
  sp.model_tier,
  mak.name AS api_key_name,
  mak.provider AS api_key_provider,
  mak_ant.name AS anthropic_api_key_name,
  mak_ant.provider AS anthropic_api_key_provider,
  ms.cost_input_rate,
  ms.cost_output_rate,
  ms.profit_margin,
  ms.custom_pricing_enabled,
  COALESCE(sp.provider_input_rate, 0) AS spu_input_rate,
  COALESCE(sp.provider_output_rate, 0) AS spu_output_rate
FROM merchant_skus ms
JOIN skus s ON ms.sku_id = s.id
JOIN spus sp ON s.spu_id = sp.id
LEFT JOIN merchant_api_keys mak ON ms.api_key_id = mak.id
LEFT JOIN merchant_api_keys mak_ant ON ms.anthropic_api_key_id = mak_ant.id;

COMMENT ON VIEW merchant_sku_details IS '商户SKU详情视图（含主/Anthropic 密钥展示与成本）';
