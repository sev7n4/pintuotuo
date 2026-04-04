-- Merchant API Key Verification Migration
-- Version: 029
-- Date: 2026-04-04
-- Description: 添加商户API Key验证字段和验证记录表

-- ============================================================
-- Phase 1: 添加验证字段到 merchant_api_keys
-- ============================================================

ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS verification_result VARCHAR(20) DEFAULT 'pending';
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS verification_message TEXT;
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS models_supported JSONB;

COMMENT ON COLUMN merchant_api_keys.verified_at IS '验证通过时间';
COMMENT ON COLUMN merchant_api_keys.verification_result IS '验证结果: pending/verified/failed';
COMMENT ON COLUMN merchant_api_keys.verification_message IS '验证消息或错误原因';
COMMENT ON COLUMN merchant_api_keys.models_supported IS '支持的模型列表';

-- ============================================================
-- Phase 2: 创建验证记录表
-- ============================================================

CREATE TABLE IF NOT EXISTS api_key_verifications (
  id SERIAL PRIMARY KEY,
  api_key_id INT NOT NULL REFERENCES merchant_api_keys(id) ON DELETE CASCADE,
  
  verification_type VARCHAR(20) NOT NULL,
  status VARCHAR(20) NOT NULL,
  
  connection_test BOOLEAN DEFAULT FALSE,
  connection_latency_ms INT,
  
  models_found JSONB,
  models_count INT DEFAULT 0,
  
  pricing_verified BOOLEAN DEFAULT FALSE,
  pricing_info JSONB,
  
  error_code VARCHAR(50),
  error_message TEXT,
  
  started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP
);

COMMENT ON TABLE api_key_verifications IS 'API Key验证记录表';
COMMENT ON COLUMN api_key_verifications.verification_type IS '验证类型: initial/periodic/manual';
COMMENT ON COLUMN api_key_verifications.status IS '验证状态: pending/in_progress/success/failed';
COMMENT ON COLUMN api_key_verifications.connection_test IS '连接测试是否通过';
COMMENT ON COLUMN api_key_verifications.models_found IS '发现的模型列表';
COMMENT ON COLUMN api_key_verifications.pricing_verified IS '定价信息是否验证';

CREATE INDEX IF NOT EXISTS idx_verifications_api_key ON api_key_verifications(api_key_id);
CREATE INDEX IF NOT EXISTS idx_verifications_status ON api_key_verifications(status);
CREATE INDEX IF NOT EXISTS idx_verifications_started ON api_key_verifications(started_at);

-- ============================================================
-- Phase 3: 添加成本定价字段到 merchant_api_keys
-- ============================================================

ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS cost_input_rate DECIMAL(10, 6);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS cost_output_rate DECIMAL(10, 6);
ALTER TABLE merchant_api_keys ADD COLUMN IF NOT EXISTS profit_margin DECIMAL(5, 2) DEFAULT 20.00;

COMMENT ON COLUMN merchant_api_keys.cost_input_rate IS '商户成本输入Token单价（元/1K）';
COMMENT ON COLUMN merchant_api_keys.cost_output_rate IS '商户成本输出Token单价（元/1K）';
COMMENT ON COLUMN merchant_api_keys.profit_margin IS '利润率百分比';
