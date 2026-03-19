-- Update payments table with additional fields
ALTER TABLE payments ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS pay_method VARCHAR(20);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS out_trade_no VARCHAR(100) UNIQUE;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS transaction_id VARCHAR(100);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS paid_at TIMESTAMP WITH TIME ZONE;

-- Rename 'method' column to 'pay_method' if exists
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'payments' AND column_name = 'method') THEN
    ALTER TABLE payments RENAME COLUMN method TO pay_method_temp;
    UPDATE payments SET pay_method = pay_method_temp WHERE pay_method IS NULL;
    ALTER TABLE payments DROP COLUMN pay_method_temp;
  END IF;
END $$;

-- Create indexes for payments
CREATE INDEX IF NOT EXISTS idx_payments_user ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_out_trade_no ON payments(out_trade_no);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created ON payments(created_at DESC);
