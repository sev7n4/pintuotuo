-- Token lots: FIFO by expiry (earliest first); NULL expires_at = never expires (legacy / recharge pool).

CREATE TABLE IF NOT EXISTS token_lots (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  remaining_amount DECIMAL(15, 2) NOT NULL CHECK (remaining_amount >= 0),
  expires_at TIMESTAMPTZ,
  order_item_id INTEGER REFERENCES order_items(id) ON DELETE SET NULL,
  lot_type VARCHAR(50) NOT NULL DEFAULT 'token_pack',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_token_lots_user_expires ON token_lots (user_id, expires_at ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_token_lots_user_id ON token_lots (user_id);

-- One-time backfill: users with balance but no lots yet get a single legacy lot (never expires).
INSERT INTO token_lots (user_id, remaining_amount, expires_at, lot_type)
SELECT t.user_id, t.balance, NULL, 'legacy_migrated'
FROM tokens t
WHERE t.balance > 0
  AND NOT EXISTS (SELECT 1 FROM token_lots tl WHERE tl.user_id = t.user_id);
