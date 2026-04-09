-- SKU default cost fields for admin managed inheritance chain
-- Version: 043

ALTER TABLE skus ADD COLUMN IF NOT EXISTS cost_input_rate DECIMAL(10, 6);
ALTER TABLE skus ADD COLUMN IF NOT EXISTS cost_output_rate DECIMAL(10, 6);
ALTER TABLE skus ADD COLUMN IF NOT EXISTS inherit_spu_cost BOOLEAN DEFAULT TRUE;

-- Backfill: existing SKUs inherit SPU reference by default
UPDATE skus s
SET cost_input_rate = COALESCE(s.cost_input_rate, sp.provider_input_rate, 0),
    cost_output_rate = COALESCE(s.cost_output_rate, sp.provider_output_rate, 0),
    inherit_spu_cost = COALESCE(s.inherit_spu_cost, TRUE)
FROM spus sp
WHERE sp.id = s.spu_id;

ALTER TABLE skus ALTER COLUMN cost_input_rate SET DEFAULT 0;
ALTER TABLE skus ALTER COLUMN cost_output_rate SET DEFAULT 0;
ALTER TABLE skus ALTER COLUMN inherit_spu_cost SET DEFAULT TRUE;
ALTER TABLE skus ALTER COLUMN cost_input_rate SET NOT NULL;
ALTER TABLE skus ALTER COLUMN cost_output_rate SET NOT NULL;
ALTER TABLE skus ALTER COLUMN inherit_spu_cost SET NOT NULL;

COMMENT ON COLUMN skus.cost_input_rate IS '运营侧上架 SKU 默认输入成本（元/1K tokens）';
COMMENT ON COLUMN skus.cost_output_rate IS '运营侧上架 SKU 默认输出成本（元/1K tokens）';
COMMENT ON COLUMN skus.inherit_spu_cost IS '是否继承 SPU 参考成本；true 时由 SPU 自动覆盖';
