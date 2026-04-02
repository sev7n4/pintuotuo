-- 物理移除 legacy products 表；收藏/浏览/秒杀关联列统一为 sku_id；购物车仅保留 sku_id。
-- 依赖：020_deprecate_legacy_products_fk 已解除 orders/groups/cart_items 对 products 的外键。
-- Version: 021
-- Date: 2026-04-02

DROP VIEW IF EXISTS products_v2 CASCADE;

-- 购物车：清空后删除 product_id，唯一约束改为 (user_id, sku_id, group_id)
TRUNCATE TABLE cart_items;
DROP INDEX IF EXISTS idx_cart_items_product_id;
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_product_id_fkey;
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_user_id_product_id_group_id_key;
ALTER TABLE cart_items DROP COLUMN IF EXISTS product_id;
ALTER TABLE cart_items ALTER COLUMN sku_id SET NOT NULL;
ALTER TABLE cart_items ADD CONSTRAINT cart_items_user_sku_group_key UNIQUE (user_id, sku_id, group_id);

-- 秒杀活动商品：product_id -> sku_id
ALTER TABLE flash_sale_products DROP CONSTRAINT IF EXISTS flash_sale_products_product_id_fkey;
ALTER TABLE flash_sale_products DROP CONSTRAINT IF EXISTS flash_sale_products_flash_sale_id_product_id_key;
ALTER TABLE flash_sale_products RENAME COLUMN product_id TO sku_id;
DROP INDEX IF EXISTS idx_flash_sale_products_product;
CREATE INDEX IF NOT EXISTS idx_flash_sale_products_sku ON flash_sale_products(sku_id);
ALTER TABLE flash_sale_products
  ADD CONSTRAINT flash_sale_products_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES skus(id) ON DELETE CASCADE;
ALTER TABLE flash_sale_products
  ADD CONSTRAINT flash_sale_products_flash_sale_sku_unique UNIQUE (flash_sale_id, sku_id);

-- 收藏：016 已存在 sku_id，合并后删除 product_id
UPDATE favorites SET sku_id = product_id WHERE sku_id IS NULL;
DELETE FROM favorites f1 USING favorites f2
WHERE f1.user_id = f2.user_id AND f1.sku_id = f2.sku_id AND f1.id > f2.id;

ALTER TABLE favorites DROP CONSTRAINT IF EXISTS favorites_product_id_fkey;
ALTER TABLE favorites DROP CONSTRAINT IF EXISTS favorites_user_id_product_id_key;
DROP INDEX IF EXISTS idx_favorites_product_id;
ALTER TABLE favorites DROP COLUMN product_id;

ALTER TABLE favorites ALTER COLUMN sku_id SET NOT NULL;
ALTER TABLE favorites DROP CONSTRAINT IF EXISTS favorites_sku_id_fkey;
ALTER TABLE favorites ADD CONSTRAINT favorites_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES skus(id) ON DELETE CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS favorites_user_sku_unique ON favorites (user_id, sku_id);

-- 浏览历史
UPDATE browse_history SET sku_id = product_id WHERE sku_id IS NULL;
DELETE FROM browse_history b1 USING browse_history b2
WHERE b1.user_id = b2.user_id AND b1.sku_id = b2.sku_id AND b1.id > b2.id;

ALTER TABLE browse_history DROP CONSTRAINT IF EXISTS browse_history_product_id_fkey;
ALTER TABLE browse_history DROP CONSTRAINT IF EXISTS browse_history_user_id_product_id_key;
DROP INDEX IF EXISTS idx_browse_history_product_id;
ALTER TABLE browse_history DROP COLUMN product_id;

ALTER TABLE browse_history ALTER COLUMN sku_id SET NOT NULL;
ALTER TABLE browse_history DROP CONSTRAINT IF EXISTS browse_history_sku_id_fkey;
ALTER TABLE browse_history ADD CONSTRAINT browse_history_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES skus(id) ON DELETE CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS browse_history_user_sku_unique ON browse_history (user_id, sku_id);

DROP TABLE IF EXISTS products CASCADE;

COMMENT ON COLUMN flash_sale_products.sku_id IS '平台 SKU id';
COMMENT ON COLUMN favorites.sku_id IS '平台 SKU id';
COMMENT ON COLUMN browse_history.sku_id IS '平台 SKU id';
