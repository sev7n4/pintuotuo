-- SKU Migration Sprint - Phase 1: Data Layer Migration
-- Version: 016
-- Date: 2026-03-29
-- Description: 为现有关联表添加 SKU 字段，创建兼容视图

-- ============================================================
-- Phase 1: 为现有关联表添加 SKU 字段
-- ============================================================

-- 1. orders 表增加 sku_id 和 spu_id
ALTER TABLE orders ADD COLUMN IF NOT EXISTS sku_id INT REFERENCES skus(id);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS spu_id INT REFERENCES spus(id);

CREATE INDEX IF NOT EXISTS idx_orders_sku_id ON orders(sku_id);
CREATE INDEX IF NOT EXISTS idx_orders_spu_id ON orders(spu_id);

COMMENT ON COLUMN orders.sku_id IS '关联的SKU ID，新订单使用此字段';
COMMENT ON COLUMN orders.spu_id IS '关联的SPU ID，用于快速查询';

-- 2. cart_items 表增加 sku_id 和 spu_id
ALTER TABLE cart_items ADD COLUMN IF NOT EXISTS sku_id INT REFERENCES skus(id);
ALTER TABLE cart_items ADD COLUMN IF NOT EXISTS spu_id INT REFERENCES spus(id);

CREATE INDEX IF NOT EXISTS idx_cart_items_sku_id ON cart_items(sku_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_spu_id ON cart_items(spu_id);

COMMENT ON COLUMN cart_items.sku_id IS '关联的SKU ID，新购物车项使用此字段';
COMMENT ON COLUMN cart_items.spu_id IS '关联的SPU ID，用于快速查询';

-- 3. groups 表增加 sku_id 和 spu_id
ALTER TABLE groups ADD COLUMN IF NOT EXISTS sku_id INT REFERENCES skus(id);
ALTER TABLE groups ADD COLUMN IF NOT EXISTS spu_id INT REFERENCES spus(id);

CREATE INDEX IF NOT EXISTS idx_groups_sku_id ON groups(sku_id);
CREATE INDEX IF NOT EXISTS idx_groups_spu_id ON groups(spu_id);

COMMENT ON COLUMN groups.sku_id IS '关联的SKU ID，拼团配置从此SKU读取';
COMMENT ON COLUMN groups.spu_id IS '关联的SPU ID，用于快速查询';

-- 4. favorites 表增加 sku_id
ALTER TABLE favorites ADD COLUMN IF NOT EXISTS sku_id INT REFERENCES skus(id);

CREATE INDEX IF NOT EXISTS idx_favorites_sku_id ON favorites(sku_id);

COMMENT ON COLUMN favorites.sku_id IS '关联的SKU ID';

-- 5. browse_history 表增加 sku_id
ALTER TABLE browse_history ADD COLUMN IF NOT EXISTS sku_id INT REFERENCES skus(id);

CREATE INDEX IF NOT EXISTS idx_browse_history_sku_id ON browse_history(sku_id);

COMMENT ON COLUMN browse_history.sku_id IS '关联的SKU ID';

-- ============================================================
-- Phase 2: 创建兼容视图
-- ============================================================

-- 创建 products_v2 视图，使旧 API 返回 SKU 数据
CREATE OR REPLACE VIEW products_v2 AS
SELECT 
  s.id,
  s.sku_code,
  s.merchant_id,
  sp.name,
  sp.name || ' - ' || 
    CASE s.sku_type 
      WHEN 'token_pack' THEN COALESCE(s.token_amount::text, '0') || ' Tokens'
      WHEN 'subscription' THEN COALESCE(s.subscription_period, 'monthly')
      WHEN 'concurrent' THEN COALESCE(s.concurrent_requests::text, '1') || ' 并发'
      ELSE s.sku_type
    END as description,
  s.retail_price as price,
  s.original_price,
  CASE WHEN s.stock = -1 THEN 999999 ELSE s.stock END as stock,
  s.sales_count as sold_count,
  sp.model_tier as category,
  s.status,
  s.created_at,
  s.updated_at,
  s.spu_id as model_id,
  s.id as package_id,
  s.sku_type,
  s.token_amount,
  s.compute_points,
  s.group_enabled,
  s.group_discount_rate
FROM skus s
JOIN spus sp ON s.spu_id = sp.id
WHERE s.status = 'active' AND sp.status = 'active';

COMMENT ON VIEW products_v2 IS '兼容视图，使旧 API 返回 SKU 数据';

-- ============================================================
-- Phase 3: 数据迁移（可选，如有现有数据）
-- ============================================================

-- 为现有订单关联 SKU（如果存在平台自营 SKU）
-- UPDATE orders o SET sku_id = s.id, spu_id = s.spu_id
-- FROM skus s
-- WHERE s.merchant_id IS NULL AND s.status = 'active'
-- AND o.sku_id IS NULL;

-- ============================================================
-- Phase 4: 添加订单 SKU 类型字段
-- ============================================================

ALTER TABLE orders ADD COLUMN IF NOT EXISTS sku_type VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS token_amount BIGINT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS compute_points DECIMAL(15, 2);

COMMENT ON COLUMN orders.sku_type IS 'SKU类型: token_pack/subscription/concurrent/trial';
COMMENT ON COLUMN orders.token_amount IS 'Token数量（token_pack类型）';
COMMENT ON COLUMN orders.compute_points IS '算力点数量';

-- ============================================================
-- Phase 5: 添加拼团价格字段
-- ============================================================

ALTER TABLE orders ADD COLUMN IF NOT EXISTS group_price DECIMAL(10, 2);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS discount_rate DECIMAL(5, 2);

COMMENT ON COLUMN orders.group_price IS '拼团价格';
COMMENT ON COLUMN orders.discount_rate IS '折扣率(%)';

-- ============================================================
-- Phase 6: 更新触发器
-- ============================================================

-- 订单创建时自动填充 SKU 信息
CREATE OR REPLACE FUNCTION populate_order_sku_info()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.sku_id IS NOT NULL AND NEW.sku_type IS NULL THEN
    SELECT sku_type, token_amount, compute_points 
    INTO NEW.sku_type, NEW.token_amount, NEW.compute_points
    FROM skus WHERE id = NEW.sku_id;
  END IF;
  
  IF NEW.sku_id IS NOT NULL AND NEW.spu_id IS NULL THEN
    SELECT spu_id INTO NEW.spu_id FROM skus WHERE id = NEW.sku_id;
  END IF;
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_populate_order_sku_info
  BEFORE INSERT ON orders
  FOR EACH ROW
  EXECUTE FUNCTION populate_order_sku_info();
