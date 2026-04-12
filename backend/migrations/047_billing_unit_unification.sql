-- 047_billing_unit_unification.sql
-- 计费单位口径统一：预扣费机制与多层级配置

-- 1. api_usage_logs 新增 token_usage 字段
ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS token_usage BIGINT;
COMMENT ON COLUMN api_usage_logs.token_usage IS '模型 Token 使用量 (input + output)';

-- 历史数据回填
UPDATE api_usage_logs 
SET token_usage = input_tokens + output_tokens 
WHERE token_usage IS NULL;

-- 2. merchant_settlements 新增字段
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS total_sales_cny DECIMAL(20,6);
ALTER TABLE merchant_settlements ADD COLUMN IF NOT EXISTS total_tokens BIGINT;
COMMENT ON COLUMN merchant_settlements.total_sales_cny IS '人民币销售总额';
COMMENT ON COLUMN merchant_settlements.total_tokens IS '模型 Token 使用总量';

-- 3. 新增预扣费表
CREATE TABLE IF NOT EXISTS pre_deductions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    request_id VARCHAR(255) NOT NULL UNIQUE,
    pre_deduct_amount BIGINT NOT NULL,
    actual_amount BIGINT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    settled_at TIMESTAMP,
    CONSTRAINT fk_pre_deductions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pre_deductions_user_id ON pre_deductions(user_id);
CREATE INDEX IF NOT EXISTS idx_pre_deductions_request_id ON pre_deductions(request_id);
CREATE INDEX IF NOT EXISTS idx_pre_deductions_status ON pre_deductions(status);

COMMENT ON TABLE pre_deductions IS '预扣费记录表';
COMMENT ON COLUMN pre_deductions.pre_deduct_amount IS '预扣 Token 数量';
COMMENT ON COLUMN pre_deductions.actual_amount IS '实际消耗 Token 数量';
COMMENT ON COLUMN pre_deductions.status IS '状态: pending(待结算), settled(已结算), cancelled(已取消)';

-- 4. SPU 级别预扣费配置
ALTER TABLE spus ADD COLUMN IF NOT EXISTS pre_deduct_multiplier INT;
ALTER TABLE spus ADD COLUMN IF NOT EXISTS pre_deduct_max_multiplier INT;
COMMENT ON COLUMN spus.pre_deduct_multiplier IS '预扣费倍率（继承优先级：SKU > SPU > Provider）';
COMMENT ON COLUMN spus.pre_deduct_max_multiplier IS '预扣费最大倍率';

-- 5. SKU 级别预扣费配置
ALTER TABLE skus ADD COLUMN IF NOT EXISTS pre_deduct_multiplier INT;
ALTER TABLE skus ADD COLUMN IF NOT EXISTS pre_deduct_max_multiplier INT;
COMMENT ON COLUMN skus.pre_deduct_multiplier IS '预扣费倍率（最高优先级）';
COMMENT ON COLUMN skus.pre_deduct_max_multiplier IS '预扣费最大倍率';

-- 6. Provider 级别预扣费配置使用现有 segment_config 字段
-- segment_config 示例: {"pre_deduct_multiplier": 2, "pre_deduct_max_multiplier": 10}
-- 无需新增字段，使用现有 model_providers.segment_config
