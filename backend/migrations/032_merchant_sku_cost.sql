-- Merchant SKU Cost Migration
-- Version: 032
-- Date: 2026-04-04
-- Description: 添加商户SKU成本定价字段

-- ============================================================
-- Phase 1: 添加成本定价字段到 merchant_skus
-- ============================================================

ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS cost_input_rate DECIMAL(10, 6);
ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS cost_output_rate DECIMAL(10, 6);
ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS profit_margin DECIMAL(5, 2) DEFAULT 20.00;
ALTER TABLE merchant_skus ADD COLUMN IF NOT EXISTS custom_pricing_enabled BOOLEAN DEFAULT FALSE;

COMMENT ON COLUMN merchant_skus.cost_input_rate IS '商户成本输入Token单价（元/1K）';
COMMENT ON COLUMN merchant_skus.cost_output_rate IS '商户成本输出Token单价（元/1K）';
COMMENT ON COLUMN merchant_skus.profit_margin IS '利润率百分比';
COMMENT ON COLUMN merchant_skus.custom_pricing_enabled IS '是否启用自定义定价';

-- ============================================================
-- Phase 2: 添加成本定价字段到 merchant_api_keys（如果不存在）
-- ============================================================

ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS cost_input_rate DECIMAL(10, 6);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS cost_output_rate DECIMAL(10, 6);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS profit_margin DECIMAL(5, 2) DEFAULT 20.00;

-- ============================================================
-- Phase 3: 创建成本定价历史表
-- ============================================================

CREATE TABLE IF NOT EXISTS merchant_sku_cost_history (
  id SERIAL PRIMARY KEY,
  merchant_sku_id INT REFERENCES merchant_skus(id) ON DELETE SET NULL,
  merchant_api_key_id INT REFERENCES merchant_api_keys(id) ON DELETE SET NULL,
  
  old_input_rate DECIMAL(10, 6),
  old_output_rate DECIMAL(10, 6),
  new_input_rate DECIMAL(10, 6),
  new_output_rate DECIMAL(10, 6),
  
  reason VARCHAR(255),
  
  changed_by INT REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE merchant_sku_cost_history IS '商户SKU成本定价变更历史';

CREATE INDEX IF NOT EXISTS idx_sku_cost_history_sku ON merchant_sku_cost_history(merchant_sku_id);
CREATE INDEX IF NOT EXISTS idx_sku_cost_history_api_key ON merchant_sku_cost_history(merchant_api_key_id);
CREATE INDEX IF NOT EXISTS idx_sku_cost_history_created ON merchant_sku_cost_history(created_at);
