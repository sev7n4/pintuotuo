-- Order digital fulfillment idempotency + compute_points SKU type

ALTER TABLE orders ADD COLUMN IF NOT EXISTS fulfilled_at TIMESTAMPTZ;

COMMENT ON COLUMN orders.fulfilled_at IS 'Set when digital goods were delivered; used for idempotent fulfillment on payment callbacks.';

ALTER TABLE skus DROP CONSTRAINT IF EXISTS skus_sku_type_check;

ALTER TABLE skus ADD CONSTRAINT skus_sku_type_check CHECK (sku_type IN (
  'token_pack', 'subscription', 'concurrent', 'trial', 'compute_points'
));
