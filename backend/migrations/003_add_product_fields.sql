-- Add new fields to products table for homepage features
ALTER TABLE products ADD COLUMN IF NOT EXISTS original_price DECIMAL(10, 2);
ALTER TABLE products ADD COLUMN IF NOT EXISTS sold_count INTEGER DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS category VARCHAR(100);

-- Create indexes for homepage queries
CREATE INDEX IF NOT EXISTS idx_products_sold_count ON products(sold_count DESC);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_status_stock ON products(status, stock);

-- Update existing products with default values
UPDATE products SET sold_count = 0 WHERE sold_count IS NULL;
UPDATE products SET original_price = price WHERE original_price IS NULL;
