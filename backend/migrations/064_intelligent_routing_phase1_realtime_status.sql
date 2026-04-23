-- 智能路由系统 Phase 1 - API Key 实时状态表
-- Version: 064
-- 目标：
-- 1) 创建 api_key_realtime_status 表，存储 API Key 的实时状态信息
-- 2) 为智能路由决策层提供实时状态数据支持

-- ============================================================================
-- 1. 创建 api_key_realtime_status 表
-- ============================================================================

CREATE TABLE IF NOT EXISTS api_key_realtime_status (
    api_key_id INT PRIMARY KEY REFERENCES merchant_api_keys(id) ON DELETE CASCADE,
    
    -- 延迟指标（毫秒）
    latency_p50 INT DEFAULT 0,
    latency_p95 INT DEFAULT 0,
    latency_p99 INT DEFAULT 0,
    
    -- 成功率和错误率
    error_rate DECIMAL(5,4) DEFAULT 0.0000,
    success_rate DECIMAL(5,4) DEFAULT 1.0000,
    
    -- 连接池状态
    connection_pool_size INT DEFAULT 10,
    connection_pool_active INT DEFAULT 0,
    
    -- 限流信息
    rate_limit_remaining INT DEFAULT 0,
    rate_limit_reset_at TIMESTAMP WITH TIME ZONE,
    
    -- 负载均衡
    load_balance_weight DECIMAL(3,2) DEFAULT 1.00,
    
    -- 最后请求时间
    last_request_at TIMESTAMP WITH TIME ZONE,
    
    -- 时间戳
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 2. 添加字段注释
-- ============================================================================

COMMENT ON TABLE api_key_realtime_status IS 'API Key 实时状态表：存储 API Key 的实时性能指标和状态信息';
COMMENT ON COLUMN api_key_realtime_status.api_key_id IS 'API Key ID（主键，外键关联 merchant_api_keys）';
COMMENT ON COLUMN api_key_realtime_status.latency_p50 IS 'P50 延迟（毫秒）：50% 的请求延迟低于此值';
COMMENT ON COLUMN api_key_realtime_status.latency_p95 IS 'P95 延迟（毫秒）：95% 的请求延迟低于此值';
COMMENT ON COLUMN api_key_realtime_status.latency_p99 IS 'P99 延迟（毫秒）：99% 的请求延迟低于此值';
COMMENT ON COLUMN api_key_realtime_status.error_rate IS '错误率（0.0000-1.0000）：最近时间窗口内的错误率';
COMMENT ON COLUMN api_key_realtime_status.success_rate IS '成功率（0.0000-1.0000）：最近时间窗口内的成功率';
COMMENT ON COLUMN api_key_realtime_status.connection_pool_size IS '连接池大小：最大连接数';
COMMENT ON COLUMN api_key_realtime_status.connection_pool_active IS '活跃连接数：当前正在使用的连接数';
COMMENT ON COLUMN api_key_realtime_status.rate_limit_remaining IS '剩余限流配额：当前剩余的请求配额';
COMMENT ON COLUMN api_key_realtime_status.rate_limit_reset_at IS '限流重置时间：配额重置的时间点';
COMMENT ON COLUMN api_key_realtime_status.load_balance_weight IS '负载均衡权重（0.00-1.00）：用于加权轮询算法';
COMMENT ON COLUMN api_key_realtime_status.last_request_at IS '最后请求时间：最后一次使用此 API Key 的时间';
COMMENT ON COLUMN api_key_realtime_status.updated_at IS '更新时间：状态数据的最后更新时间';

-- ============================================================================
-- 3. 创建索引
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_api_key_realtime_status_updated_at ON api_key_realtime_status(updated_at);
CREATE INDEX IF NOT EXISTS idx_api_key_realtime_status_error_rate ON api_key_realtime_status(error_rate);
CREATE INDEX IF NOT EXISTS idx_api_key_realtime_status_success_rate ON api_key_realtime_status(success_rate);
CREATE INDEX IF NOT EXISTS idx_api_key_realtime_status_last_request_at ON api_key_realtime_status(last_request_at);

-- ============================================================================
-- 4. 创建触发器：自动更新 updated_at
-- ============================================================================

CREATE OR REPLACE FUNCTION update_api_key_realtime_status_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_api_key_realtime_status_updated_at ON api_key_realtime_status;

CREATE TRIGGER trigger_update_api_key_realtime_status_updated_at
    BEFORE UPDATE ON api_key_realtime_status
    FOR EACH ROW
    EXECUTE FUNCTION update_api_key_realtime_status_updated_at();
