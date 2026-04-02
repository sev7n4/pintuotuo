-- Deprecate legacy products table as source of truth: allow SKU-only orders/groups/cart rows.
-- Safe when no production data depends on products(id) FK integrity.
-- Version: 019
-- Date: 2026-04-02

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_product_id_fkey;
ALTER TABLE groups DROP CONSTRAINT IF EXISTS groups_product_id_fkey;
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_product_id_fkey;

ALTER TABLE orders ALTER COLUMN product_id DROP NOT NULL;
ALTER TABLE groups ALTER COLUMN product_id DROP NOT NULL;
ALTER TABLE cart_items ALTER COLUMN product_id DROP NOT NULL;

COMMENT ON COLUMN orders.product_id IS 'Deprecated: use sku_id. NULL for SKU-only orders.';
COMMENT ON COLUMN groups.product_id IS 'Deprecated: use sku_id. NULL for SKU-only groups.';
COMMENT ON COLUMN cart_items.product_id IS 'Deprecated: use sku_id. NULL when line is SKU-based.';
