-- Flash Sales Tables
-- Stores flash sale events
CREATE TABLE IF NOT EXISTS flash_sales (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) DEFAULT 'upcoming' CHECK (status IN ('upcoming', 'active', 'ended', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Stores products in flash sales
CREATE TABLE IF NOT EXISTS flash_sale_products (
    id SERIAL PRIMARY KEY,
    flash_sale_id INTEGER NOT NULL REFERENCES flash_sales(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    flash_price DECIMAL(10, 2) NOT NULL,
    original_price DECIMAL(10, 2) NOT NULL,
    stock_limit INTEGER NOT NULL DEFAULT 0,
    stock_sold INTEGER DEFAULT 0,
    per_user_limit INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(flash_sale_id, product_id)
);

-- Index for active flash sales
CREATE INDEX IF NOT EXISTS idx_flash_sales_status ON flash_sales(status);
CREATE INDEX IF NOT EXISTS idx_flash_sales_time ON flash_sales(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_flash_sale_products_sale ON flash_sale_products(flash_sale_id);
CREATE INDEX IF NOT EXISTS idx_flash_sale_products_product ON flash_sale_products(product_id);

-- Add flash_sale_id to order_items for tracking flash sale purchases
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS flash_sale_id INTEGER REFERENCES flash_sales(id);
