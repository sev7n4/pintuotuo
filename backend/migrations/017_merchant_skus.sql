-- Merchant SKUs Association Table
-- Version: 017
-- Date: 2026-03-29
-- Description: 创建商户-SKU关联表，支持商户选择平台SKU进行销售

-- ============================================================
-- Phase 1: 创建商户-SKU关联表
-- ============================================================

CREATE TABLE IF NOT EXISTS merchant_skus (
  id SERIAL PRIMARY KEY,
  merchant_id INT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
  sku_id INT NOT NULL REFERENCES skus(id) ON DELETE CASCADE,
  
  -- API Key 关联
  api_key_id INT REFERENCES merchant_api_keys(id) ON DELETE SET NULL,
  
  -- 状态
  status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  
  -- 销量统计
  sales_count BIGINT DEFAULT 0,
  total_sales_amount DECIMAL(15, 2) DEFAULT 0,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  UNIQUE(merchant_id, sku_id)
);

COMMENT ON TABLE merchant_skus IS '商户-SKU关联表，记录商户选择的平台SKU';
COMMENT ON COLUMN merchant_skus.merchant_id IS '商户ID';
COMMENT ON COLUMN merchant_skus.sku_id IS '平台SKU ID';
COMMENT ON COLUMN merchant_skus.api_key_id IS '关联的商户API Key';
COMMENT ON COLUMN merchant_skus.status IS '状态: active-在售, inactive-下架';
COMMENT ON COLUMN merchant_skus.sales_count IS '销量统计';
COMMENT ON COLUMN merchant_skus.total_sales_amount IS '销售总额';

-- ============================================================
-- Phase 2: 创建索引
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_merchant_skus_merchant ON merchant_skus(merchant_id);
CREATE INDEX IF NOT EXISTS idx_merchant_skus_sku ON merchant_skus(sku_id);
CREATE INDEX IF NOT EXISTS idx_merchant_skus_status ON merchant_skus(status);
CREATE INDEX IF NOT EXISTS idx_merchant_skus_api_key ON merchant_skus(api_key_id);

-- ============================================================
-- Phase 3: 创建更新时间触发器
-- ============================================================

CREATE OR REPLACE FUNCTION update_merchant_skus_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_merchant_skus_updated_at
  BEFORE UPDATE ON merchant_skus
  FOR EACH ROW
  EXECUTE FUNCTION update_merchant_skus_updated_at();

-- ============================================================
-- Phase 4: 创建视图 - 商户SKU详情视图
-- ============================================================

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
  mak.provider as api_key_provider
FROM merchant_skus ms
JOIN skus s ON ms.sku_id = s.id
JOIN spus sp ON s.spu_id = sp.id
LEFT JOIN merchant_api_keys mak ON ms.api_key_id = mak.id;

COMMENT ON VIEW merchant_sku_details IS '商户SKU详情视图，包含SKU和SPU信息';
