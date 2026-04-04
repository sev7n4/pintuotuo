-- Merchant Profile Enhancement Migration
-- Version: 031
-- Date: 2026-04-04
-- Description: 增强商户资料，添加银行账户和结算配置

-- ============================================================
-- Phase 1: 添加银行账户字段
-- ============================================================

ALTER TABLE merchants ADD COLUMN IF NOT EXISTS bank_name VARCHAR(100);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS bank_account VARCHAR(50);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS account_name VARCHAR(100);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS bank_branch VARCHAR(200);

COMMENT ON COLUMN merchants.bank_name IS '开户银行';
COMMENT ON COLUMN merchants.bank_account IS '银行账号';
COMMENT ON COLUMN merchants.account_name IS '账户名称';
COMMENT ON COLUMN merchants.bank_branch IS '开户支行';

-- ============================================================
-- Phase 2: 添加结算配置字段
-- ============================================================

ALTER TABLE merchants ADD COLUMN IF NOT EXISTS settlement_cycle VARCHAR(20) DEFAULT 'monthly';
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS monthly_threshold DECIMAL(12, 2) DEFAULT 100.00;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS auto_settle BOOLEAN DEFAULT TRUE;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS settlement_day INT DEFAULT 1;

COMMENT ON COLUMN merchants.settlement_cycle IS '结算周期: weekly/monthly/quarterly';
COMMENT ON COLUMN merchants.monthly_threshold IS '月度结算门槛（元）';
COMMENT ON COLUMN merchants.auto_settle IS '是否自动结算';
COMMENT ON COLUMN merchants.settlement_day IS '结算日（每月几号）';

-- ============================================================
-- Phase 3: 添加商户评分字段
-- ============================================================

ALTER TABLE merchants ADD COLUMN IF NOT EXISTS rating DECIMAL(3, 2) DEFAULT 5.00;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS total_orders BIGINT DEFAULT 0;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS total_sales DECIMAL(15, 2) DEFAULT 0;

COMMENT ON COLUMN merchants.rating IS '商户评分（1-5）';
COMMENT ON COLUMN merchants.total_orders IS '总订单数';
COMMENT ON COLUMN merchants.total_sales IS '总销售额';

-- ============================================================
-- Phase 4: 创建商户结算配置历史表
-- ============================================================

CREATE TABLE IF NOT EXISTS merchant_settlement_config_history (
  id SERIAL PRIMARY KEY,
  merchant_id INT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
  
  config_type VARCHAR(50) NOT NULL,
  old_value JSONB,
  new_value JSONB,
  
  changed_by INT REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE merchant_settlement_config_history IS '商户结算配置变更历史';

CREATE INDEX IF NOT EXISTS idx_settlement_config_merchant ON merchant_settlement_config_history(merchant_id);
CREATE INDEX IF NOT EXISTS idx_settlement_config_type ON merchant_settlement_config_history(config_type);
