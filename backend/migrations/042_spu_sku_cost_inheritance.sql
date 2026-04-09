-- SPU/SKU/Key cost inheritance hardening
-- Version: 042
-- Description:
-- 1) SPU provider_input_rate / provider_output_rate become non-null defaults (catalog truth source)
-- 2) merchant_skus cost fields backfilled from SPU and become non-null
-- 3) merchant_sku_details view exposes cost + reference source fields

-- Phase 1: SPU catalog reference cost should not stay NULL.
UPDATE spus
SET provider_input_rate = COALESCE(provider_input_rate, 0),
    provider_output_rate = COALESCE(provider_output_rate, 0)
WHERE provider_input_rate IS NULL OR provider_output_rate IS NULL;

ALTER TABLE spus ALTER COLUMN provider_input_rate SET DEFAULT 0;
ALTER TABLE spus ALTER COLUMN provider_output_rate SET DEFAULT 0;
ALTER TABLE spus ALTER COLUMN provider_input_rate SET NOT NULL;
ALTER TABLE spus ALTER COLUMN provider_output_rate SET NOT NULL;

-- Phase 2: merchant_skus inherit default cost from SPU when empty.
UPDATE merchant_skus ms
SET cost_input_rate = COALESCE(ms.cost_input_rate, sp.provider_input_rate, 0),
    cost_output_rate = COALESCE(ms.cost_output_rate, sp.provider_output_rate, 0),
    profit_margin = COALESCE(ms.profit_margin, 20.00),
    custom_pricing_enabled = COALESCE(ms.custom_pricing_enabled, FALSE)
FROM skus s
JOIN spus sp ON sp.id = s.spu_id
WHERE ms.sku_id = s.id
  AND (
    ms.cost_input_rate IS NULL
    OR ms.cost_output_rate IS NULL
    OR ms.profit_margin IS NULL
    OR ms.custom_pricing_enabled IS NULL
  );

ALTER TABLE merchant_skus ALTER COLUMN cost_input_rate SET DEFAULT 0;
ALTER TABLE merchant_skus ALTER COLUMN cost_output_rate SET DEFAULT 0;
ALTER TABLE merchant_skus ALTER COLUMN profit_margin SET DEFAULT 20.00;
ALTER TABLE merchant_skus ALTER COLUMN custom_pricing_enabled SET DEFAULT FALSE;
ALTER TABLE merchant_skus ALTER COLUMN cost_input_rate SET NOT NULL;
ALTER TABLE merchant_skus ALTER COLUMN cost_output_rate SET NOT NULL;
ALTER TABLE merchant_skus ALTER COLUMN profit_margin SET NOT NULL;
ALTER TABLE merchant_skus ALTER COLUMN custom_pricing_enabled SET NOT NULL;

COMMENT ON COLUMN merchant_skus.cost_input_rate IS '商户SKU成本输入Token单价（元/1K），默认继承SPU参考价';
COMMENT ON COLUMN merchant_skus.cost_output_rate IS '商户SKU成本输出Token单价（元/1K），默认继承SPU参考价';
COMMENT ON COLUMN merchant_skus.custom_pricing_enabled IS '是否为自部署自定义成本；false 表示继承官方目录默认';

-- Phase 3: refresh merchant detail view with cost fields and reference values.
CREATE OR REPLACE VIEW merchant_sku_details AS
SELECT
  ms.id,
  ms.merchant_id,
  ms.sku_id,
  ms.api_key_id,
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
  sp.name as spu_name,
  sp.model_provider,
  sp.model_name,
  sp.model_tier,
  mak.name as api_key_name,
  mak.provider as api_key_provider,
  ms.cost_input_rate,
  ms.cost_output_rate,
  ms.profit_margin,
  ms.custom_pricing_enabled,
  COALESCE(sp.provider_input_rate, 0) as spu_input_rate,
  COALESCE(sp.provider_output_rate, 0) as spu_output_rate
FROM merchant_skus ms
JOIN skus s ON ms.sku_id = s.id
JOIN spus sp ON s.spu_id = sp.id
LEFT JOIN merchant_api_keys mak ON ms.api_key_id = mak.id;

COMMENT ON VIEW merchant_sku_details IS '商户SKU详情视图（含商户成本与SPU参考成本）';
