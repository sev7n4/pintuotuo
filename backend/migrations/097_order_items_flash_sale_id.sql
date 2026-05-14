-- Backfill order_items.flash_sale_id for databases where 051_order_items.sql
-- created order_items after 012_add_flash_sales.sql ran (012 only ALTERs when
-- the table already exists), leaving the column missing while app code INSERTs it.

ALTER TABLE order_items ADD COLUMN IF NOT EXISTS flash_sale_id INTEGER REFERENCES flash_sales(id);
