-- 添加智能路由所需的性能指标列
-- avg_latency_ms: 平均延迟（毫秒）
-- success_rate: 成功率（0.0-1.0）

ALTER TABLE merchant_api_keys 
ADD COLUMN IF NOT EXISTS avg_latency_ms INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS success_rate NUMERIC(5,4) DEFAULT 1.0;

-- 添加索引以优化查询
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_avg_latency_ms ON merchant_api_keys(avg_latency_ms);
CREATE INDEX IF NOT EXISTS idx_merchant_api_keys_success_rate ON merchant_api_keys(success_rate);
