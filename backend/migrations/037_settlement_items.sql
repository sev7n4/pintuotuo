-- Settlement Items Table
-- Version: 037
-- Date: 2026-04-06
-- Description: 创建结算明细表，存储逐笔账单记录

-- ============================================================
-- Phase 1: 创建结算明细表
-- ============================================================

CREATE TABLE IF NOT EXISTS settlement_items (
  id SERIAL PRIMARY KEY,
  settlement_id INT NOT NULL REFERENCES merchant_settlements(id) ON DELETE CASCADE,
  api_usage_log_id INT NOT NULL REFERENCES api_usage_logs(id),
  
  -- 冗余字段，提升查询性能
  user_id INT NOT NULL,
  merchant_id INT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
  provider VARCHAR(50) NOT NULL,
  model VARCHAR(100) NOT NULL,
  
  input_tokens INT NOT NULL,
  output_tokens INT NOT NULL,
  cost DECIMAL(12, 6) NOT NULL,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  UNIQUE(settlement_id, api_usage_log_id)
);

COMMENT ON TABLE settlement_items IS '结算明细表，存储逐笔账单记录';
COMMENT ON COLUMN settlement_items.settlement_id IS '结算单ID';
COMMENT ON COLUMN settlement_items.api_usage_log_id IS 'API使用日志ID';
COMMENT ON COLUMN settlement_items.user_id IS '用户ID';
COMMENT ON COLUMN settlement_items.merchant_id IS '商户ID';
COMMENT ON COLUMN settlement_items.provider IS 'Provider名称';
COMMENT ON COLUMN settlement_items.model IS '模型名称';
COMMENT ON COLUMN settlement_items.input_tokens IS '输入Token数';
COMMENT ON COLUMN settlement_items.output_tokens IS '输出Token数';
COMMENT ON COLUMN settlement_items.cost IS '成本（USD）';

-- ============================================================
-- Phase 2: 创建索引
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_settlement_items_settlement ON settlement_items(settlement_id);
CREATE INDEX IF NOT EXISTS idx_settlement_items_log ON settlement_items(api_usage_log_id);
CREATE INDEX IF NOT EXISTS idx_settlement_items_merchant ON settlement_items(merchant_id);
CREATE INDEX IF NOT EXISTS idx_settlement_items_created ON settlement_items(created_at);
CREATE INDEX IF NOT EXISTS idx_settlement_items_user ON settlement_items(user_id);

-- ============================================================
-- Phase 3: 创建更新时间触发器
-- ============================================================

CREATE OR REPLACE FUNCTION update_settlement_items_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.created_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_settlement_items_updated_at
  BEFORE UPDATE ON settlement_items
  FOR EACH ROW
  EXECUTE FUNCTION update_settlement_items_updated_at();
