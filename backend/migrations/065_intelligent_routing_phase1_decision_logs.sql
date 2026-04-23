-- 智能路由系统 Phase 1 - 路由决策日志表
-- Version: 065
-- 目标：
-- 1) 创建 routing_decision_logs 表，记录路由决策的完整过程
-- 2) 为智能路由系统提供可观测性和审计能力

-- ============================================================================
-- 1. 创建 routing_decision_logs 表
-- ============================================================================

CREATE TABLE IF NOT EXISTS routing_decision_logs (
    id SERIAL PRIMARY KEY,
    
    -- 请求标识
    request_id VARCHAR(64) NOT NULL,
    merchant_id INT NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    
    -- 选中的 API Key
    api_key_id INT REFERENCES merchant_api_keys(id) ON DELETE SET NULL,
    
    -- 策略层输出
    strategy_layer_goal VARCHAR(50) NOT NULL,
    strategy_layer_input JSONB DEFAULT '{}'::jsonb,
    strategy_layer_output JSONB DEFAULT '{}'::jsonb,
    
    -- 决策层输出
    decision_layer_candidates JSONB DEFAULT '[]'::jsonb,
    decision_layer_output JSONB DEFAULT '{}'::jsonb,
    
    -- 执行层结果
    execution_layer_result JSONB DEFAULT '{}'::jsonb,
    
    -- 决策耗时（毫秒）
    decision_duration_ms INT DEFAULT 0,
    
    -- 决策结果
    decision_result VARCHAR(20) DEFAULT 'success',
    error_message TEXT,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 2. 添加字段注释
-- ============================================================================

COMMENT ON TABLE routing_decision_logs IS '路由决策日志表：记录智能路由系统的三层决策过程';
COMMENT ON COLUMN routing_decision_logs.request_id IS '请求 ID：唯一标识一次请求';
COMMENT ON COLUMN routing_decision_logs.merchant_id IS '商户 ID：发起请求的商户';
COMMENT ON COLUMN routing_decision_logs.api_key_id IS '选中的 API Key ID：最终选择使用的 API Key';
COMMENT ON COLUMN routing_decision_logs.strategy_layer_goal IS '策略层目标：performance_first, price_first, reliability_first, balanced, security_first, auto';
COMMENT ON COLUMN routing_decision_logs.strategy_layer_input IS '策略层输入（JSONB）：请求内容分析、用户偏好、成本预算等';
COMMENT ON COLUMN routing_decision_logs.strategy_layer_output IS '策略层输出（JSONB）：策略权重、约束条件等';
COMMENT ON COLUMN routing_decision_logs.decision_layer_candidates IS '决策层候选列表（JSONB）：所有候选的 API Key 及其评分';
COMMENT ON COLUMN routing_decision_logs.decision_layer_output IS '决策层输出（JSONB）：最终选择的 API Key 及其原因';
COMMENT ON COLUMN routing_decision_logs.execution_layer_result IS '执行层结果（JSONB）：请求执行的详细结果';
COMMENT ON COLUMN routing_decision_logs.decision_duration_ms IS '决策耗时（毫秒）：从策略层到执行层的总耗时';
COMMENT ON COLUMN routing_decision_logs.decision_result IS '决策结果：success, failed, timeout';
COMMENT ON COLUMN routing_decision_logs.error_message IS '错误信息：如果决策失败，记录错误原因';
COMMENT ON COLUMN routing_decision_logs.created_at IS '创建时间：决策日志的记录时间';

-- ============================================================================
-- 3. 创建索引
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_request_id ON routing_decision_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_merchant_id ON routing_decision_logs(merchant_id);
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_api_key_id ON routing_decision_logs(api_key_id);
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_strategy_layer_goal ON routing_decision_logs(strategy_layer_goal);
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_decision_result ON routing_decision_logs(decision_result);
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_created_at ON routing_decision_logs(created_at);

-- 复合索引：用于按商户和时间查询
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_merchant_created ON routing_decision_logs(merchant_id, created_at DESC);

-- GIN 索引：用于 JSONB 字段查询
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_strategy_input ON routing_decision_logs USING GIN(strategy_layer_input);
CREATE INDEX IF NOT EXISTS idx_routing_decision_logs_decision_output ON routing_decision_logs USING GIN(decision_layer_output);

-- ============================================================================
-- 4. 创建分区（可选，用于大数据量场景）
-- ============================================================================

-- 注意：分区表需要在生产环境中根据实际数据量决定是否启用
-- 这里提供分区表的创建脚本作为参考

-- CREATE TABLE routing_decision_logs_partitioned (
--     LIKE routing_decision_logs INCLUDING DEFAULTS INCLUDING CONSTRAINTS
-- ) PARTITION BY RANGE (created_at);

-- -- 创建月度分区
-- CREATE TABLE routing_decision_logs_2026_04 PARTITION OF routing_decision_logs_partitioned
--     FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

-- CREATE TABLE routing_decision_logs_2026_05 PARTITION OF routing_decision_logs_partitioned
--     FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
