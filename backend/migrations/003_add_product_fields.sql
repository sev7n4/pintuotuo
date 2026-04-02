-- Add new fields to products table for homepage features
-- Skip when legacy products table does not exist (e.g. partial DB / migration 001 marked complete incorrectly).
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'products'
  ) THEN
    ALTER TABLE products ADD COLUMN IF NOT EXISTS original_price DECIMAL(10, 2);
    ALTER TABLE products ADD COLUMN IF NOT EXISTS sold_count INTEGER DEFAULT 0;
    ALTER TABLE products ADD COLUMN IF NOT EXISTS category VARCHAR(100);

    CREATE INDEX IF NOT EXISTS idx_products_sold_count ON products(sold_count DESC);
    CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
    CREATE INDEX IF NOT EXISTS idx_products_status_stock ON products(status, stock);

    UPDATE products SET sold_count = 0 WHERE sold_count IS NULL;
    UPDATE products SET original_price = price WHERE original_price IS NULL;
  END IF;
END $$;
