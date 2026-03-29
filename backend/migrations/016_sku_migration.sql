-- SKU Migration - Phase 2
-- Version: 016
-- Date: 2026-03-29
-- Description: 关联表迁移到 SKU 架构

-- ============================================================
-- Phase 1: orders 表迁移
-- ============================================================

-- 添加 sku_id 和 spu_id 字段
ALTER TABLE orders ADD COLUMN sku_id INT REFERENCES skus(id);
ALTER TABLE orders ADD COLUMN spu_id INT REFERENCES spus(id);

-- 添加注释
COMMENT ON COLUMN orders.sku_id IS '关联的SKU ID';
COMMENT ON COLUMN orders.spu_id IS '关联的SPU ID';

-- ============================================================
-- Phase 2: cart_items 表迁移
-- ============================================================

-- 添加 sku_id 和 spu_id 字段
ALTER TABLE cart_items ADD COLUMN sku_id INT REFERENCES skus(id);
ALTER TABLE cart_items ADD COLUMN spu_id INT REFERENCES spus(id);

-- 添加注释
COMMENT ON COLUMN cart_items.sku_id IS '关联的SKU ID';
COMMENT ON COLUMN cart_items.spu_id IS '关联的SPU ID';

-- ============================================================
-- Phase 3: groups 表迁移
-- ============================================================

-- 添加 sku_id 和 spu_id 字段
ALTER TABLE groups ADD COLUMN sku_id INT REFERENCES skus(id);
ALTER TABLE groups ADD COLUMN spu_id INT REFERENCES spus(id);

-- 添加注释
COMMENT ON COLUMN groups.sku_id IS '关联的SKU ID';
COMMENT ON COLUMN groups.spu_id IS '关联的SPU ID';

-- ============================================================
-- Phase 4: favorites 表迁移
-- ============================================================

-- 添加 sku_id 字段
ALTER TABLE favorites ADD COLUMN sku_id INT REFERENCES skus(id);

-- 添加注释
COMMENT ON COLUMN favorites.sku_id IS '关联的SKU ID';

-- ============================================================
-- Phase 5: browse_history 表迁移
-- ============================================================

-- 添加 sku_id 字段
ALTER TABLE browse_history ADD COLUMN sku_id INT REFERENCES skus(id);

-- 添加注释
COMMENT ON COLUMN browse_history.sku_id IS '关联的SKU ID';

-- ============================================================
-- Phase 6: 创建兼容视图
-- ============================================================

-- 先删除旧视图（如果存在）
DROP VIEW IF EXISTS products_v2 CASCADE;

-- 创建 products_v2 兼容视图
CREATE VIEW products_v2 AS
SELECT 
  s.id,
  s.sku_code as sku_code,
  s.merchant_id,
  sp.name as name,
  sp.name || ' - ' || 
    CASE s.sku_type 
      WHEN 'token_pack' THEN s.token_amount::text || ' Tokens'
      WHEN 'subscription' THEN COALESCE(s.subscription_period, 'monthly')
      ELSE s.sku_type
    END as description,
  s.retail_price as price,
  s.original_price,
  CASE WHEN s.stock = -1 THEN 999999 ELSE s.stock END as stock,
  s.sales_count as sold_count,
  sp.model_tier as category,
  s.status,
  s.created_at,
  s.updated_at
FROM skus s
JOIN spus sp ON s.spu_id = sp.id
WHERE s.status = 'active' AND sp.status = 'active';

COMMENT ON VIEW products_v2 IS '商品兼容视图，用于旧API兼容';

-- ============================================================
-- Phase 7: 数据迁移
-- ============================================================

-- 将现有 products 数据映射到 spus 和 skus
-- 注意：这里需要根据实际业务逻辑进行映射
-- 示例：将平台自营商品迁移到 SKU
INSERT INTO skus (spu_id, sku_code, sku_type, token_amount, compute_points, retail_price, original_price, stock, valid_days, status, group_enabled, min_group_size, max_group_size)
SELECT 
  sp.id,
  'SKU-' || UPPER(s.spu_code) || '-DEFAULT',
  'token_pack',
  NULL,
  NULL,
  p.base_compute_points * 1000,
  1.0,
  365,
  'active',
  true,
  2,
  10
FROM spus sp
WHERE sp.status = 'active'
ON CONFLICT (sku_code) DO NOTHING;

-- 更新 orders 表关联
UPDATE orders o SET sku_id = s.id, spu_id = s.spu_id
FROM skus s
WHERE s.sku_code = 'SKU-' || (SELECT spu_code FROM spus sp WHERE sp.id = s.spu_id)
  AND s.status = 'active';

-- 更新 cart_items 表关联
UPDATE cart_items c SET sku_id = s.id, spu_id = s.spu_id
FROM skus s
WHERE s.sku_code = 'SKU-' || (SELECT spu_code FROM spus sp WHERE sp.id = s.spu_id)
  AND s.status = 'active';

-- 更新 groups 表关联
UPDATE groups g SET sku_id = s.id, spu_id = s.spu_id
FROM skus s
WHERE s.sku_code = 'SKU-' || (SELECT spu_code FROM spus sp WHERE sp.id = s.spu_id)
  AND s.status = 'active';
