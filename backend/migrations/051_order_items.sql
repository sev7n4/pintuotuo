-- Order items for multi-SKU order model (clean-break rollout)
-- Version: 051

CREATE TABLE IF NOT EXISTS order_items (
  id SERIAL PRIMARY KEY,
  order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  sku_id INT NOT NULL REFERENCES skus(id) ON DELETE RESTRICT,
  spu_id INT NOT NULL REFERENCES spus(id) ON DELETE RESTRICT,
  quantity INT NOT NULL CHECK (quantity > 0),
  unit_price DECIMAL(20, 6) NOT NULL,
  total_price DECIMAL(20, 6) NOT NULL,
  pricing_version_id INT REFERENCES pricing_versions(id) ON DELETE SET NULL,
  sku_type VARCHAR(50),
  token_amount BIGINT,
  compute_points DECIMAL(20, 6),
  fulfilled_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_sku_id ON order_items(sku_id);
CREATE INDEX IF NOT EXISTS idx_order_items_spu_id ON order_items(spu_id);
CREATE INDEX IF NOT EXISTS idx_order_items_fulfilled_at ON order_items(fulfilled_at);
CREATE INDEX IF NOT EXISTS idx_order_items_order_sku ON order_items(order_id, sku_id);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'update_order_items_updated_at'
  ) THEN
    CREATE TRIGGER update_order_items_updated_at
      BEFORE UPDATE ON order_items
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;
