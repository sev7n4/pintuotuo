-- Settlement Enhancement Migration
-- Version: 030
-- Date: 2026-04-04
-- Description: 增强结算系统，支持商户确认、财务审核、争议处理

-- ============================================================
-- Phase 1: 添加结算确认字段
-- ============================================================

ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS merchant_confirmed BOOLEAN DEFAULT FALSE;
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS merchant_confirmed_at TIMESTAMP;
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS finance_approved BOOLEAN DEFAULT FALSE;
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS finance_approved_at TIMESTAMP;
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS finance_approved_by INT REFERENCES users(id);
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS marked_paid_at TIMESTAMP;
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS marked_paid_by INT REFERENCES users(id);

COMMENT ON COLUMN merchant_settlements.merchant_confirmed IS '商户是否已确认';
COMMENT ON COLUMN merchant_settlements.merchant_confirmed_at IS '商户确认时间';
COMMENT ON COLUMN merchant_settlements.finance_approved IS '财务是否已审核';
COMMENT ON COLUMN merchant_settlements.finance_approved_at IS '财务审核时间';
COMMENT ON COLUMN merchant_settlements.finance_approved_by IS '财务审核人ID';
COMMENT ON COLUMN merchant_settlements.marked_paid_at IS '标记打款时间';
COMMENT ON COLUMN merchant_settlements.marked_paid_by IS '标记打款人ID';

-- ============================================================
-- Phase 2: 创建打款记录表
-- ============================================================

CREATE TABLE IF NOT EXISTS payment_records (
  id SERIAL PRIMARY KEY,
  settlement_id INT NOT NULL REFERENCES merchant_settlements(id) ON DELETE CASCADE,
  
  payment_method VARCHAR(50) NOT NULL,
  payment_amount DECIMAL(12, 2) NOT NULL,
  transaction_id VARCHAR(100),
  
  bank_name VARCHAR(100),
  bank_account VARCHAR(50),
  account_name VARCHAR(100),
  
  status VARCHAR(20) DEFAULT 'pending',
  paid_at TIMESTAMP,
  
  notes TEXT,
  
  created_by INT REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE payment_records IS '打款记录表';
COMMENT ON COLUMN payment_records.payment_method IS '打款方式: bank/alipay/wechat';
COMMENT ON COLUMN payment_records.transaction_id IS '交易流水号';
COMMENT ON COLUMN payment_records.status IS '状态: pending/completed/failed';

CREATE INDEX IF NOT EXISTS idx_payment_records_settlement ON payment_records(settlement_id);
CREATE INDEX IF NOT EXISTS idx_payment_records_status ON payment_records(status);
CREATE INDEX IF NOT EXISTS idx_payment_records_created ON payment_records(created_at);

-- ============================================================
-- Phase 3: 创建结算争议表
-- ============================================================

CREATE TABLE IF NOT EXISTS settlement_disputes (
  id SERIAL PRIMARY KEY,
  settlement_id INT NOT NULL REFERENCES merchant_settlements(id) ON DELETE CASCADE,
  merchant_id INT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
  
  dispute_type VARCHAR(50) NOT NULL,
  dispute_reason TEXT NOT NULL,
  evidence_urls JSONB,
  
  original_amount DECIMAL(12, 2),
  disputed_amount DECIMAL(12, 2),
  adjusted_amount DECIMAL(12, 2),
  
  status VARCHAR(20) DEFAULT 'pending',
  
  handled_by INT REFERENCES users(id),
  handled_at TIMESTAMP,
  resolution_notes TEXT,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE settlement_disputes IS '结算争议表';
COMMENT ON COLUMN settlement_disputes.dispute_type IS '争议类型: amount_error/missing_orders/pricing_dispute/other';
COMMENT ON COLUMN settlement_disputes.status IS '状态: pending/reviewing/resolved/rejected';

CREATE INDEX IF NOT EXISTS idx_disputes_settlement ON settlement_disputes(settlement_id);
CREATE INDEX IF NOT EXISTS idx_disputes_merchant ON settlement_disputes(merchant_id);
CREATE INDEX IF NOT EXISTS idx_disputes_status ON settlement_disputes(status);

-- ============================================================
-- Phase 4: 创建结算对账表
-- ============================================================

CREATE TABLE IF NOT EXISTS settlement_reconciliations (
  id SERIAL PRIMARY KEY,
  settlement_id INT NOT NULL REFERENCES merchant_settlements(id) ON DELETE CASCADE,
  
  order_count_expected INT,
  order_count_actual INT,
  order_count_diff INT,
  
  usage_expected DECIMAL(20, 2),
  usage_actual DECIMAL(20, 2),
  usage_diff DECIMAL(20, 2),
  
  amount_expected DECIMAL(12, 2),
  amount_actual DECIMAL(12, 2),
  amount_diff DECIMAL(12, 2),
  
  has_anomalies BOOLEAN DEFAULT FALSE,
  anomaly_details JSONB,
  
  reconciled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  reconciled_by INT REFERENCES users(id)
);

COMMENT ON TABLE settlement_reconciliations IS '结算对账表';
COMMENT ON COLUMN settlement_reconciliations.has_anomalies IS '是否存在异常';
COMMENT ON COLUMN settlement_reconciliations.anomaly_details IS '异常详情';

CREATE INDEX IF NOT EXISTS idx_reconciliations_settlement ON settlement_reconciliations(settlement_id);

-- ============================================================
-- Phase 5: 创建更新时间触发器
-- ============================================================

CREATE OR REPLACE FUNCTION update_payment_records_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_payment_records_updated_at
  BEFORE UPDATE ON payment_records
  FOR EACH ROW
  EXECUTE FUNCTION update_payment_records_updated_at();

CREATE OR REPLACE FUNCTION update_settlement_disputes_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_settlement_disputes_updated_at
  BEFORE UPDATE ON settlement_disputes
  FOR EACH ROW
  EXECUTE FUNCTION update_settlement_disputes_updated_at();
