-- Model Token Statistics: 存储各模型的 Token 使用统计数据
-- Version: 072
-- 目标：
-- 1) 创建 model_token_statistics 表，记录各模型的 Token 使用统计
-- 2) 为智能路由提供模型 Token 消耗的预测数据

-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- 1. 创建 model_token_statistics 表
-- ============================================================================

CREATE TABLE IF NOT EXISTS model_token_statistics (
    id SERIAL PRIMARY KEY,
    model_name VARCHAR(100) NOT NULL UNIQUE,
    
    -- Token 统计（基于历史请求）
    avg_input_tokens NUMERIC(10,2) DEFAULT 0,
    avg_output_tokens NUMERIC(10,2) DEFAULT 0,
    p50_input_tokens INT DEFAULT 0,
    p50_output_tokens INT DEFAULT 0,
    p90_input_tokens INT DEFAULT 0,
    p90_output_tokens INT DEFAULT 0,
    
    -- 比例统计
    input_output_ratio NUMERIC(5,2) DEFAULT 1.0,
    
    -- 样本统计
    total_requests INT DEFAULT 0,
    sample_start_date DATE,
    sample_end_date DATE,
    
    -- 元数据
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ============================================================================
-- 2. 添加字段注释
-- ============================================================================

COMMENT ON TABLE model_token_statistics IS '模型 Token 统计表：存储各模型的 Token 使用统计数据，用于智能路由预测';
COMMENT ON COLUMN model_token_statistics.model_name IS '模型名称：唯一标识一个模型';
COMMENT ON COLUMN model_token_statistics.avg_input_tokens IS '平均输入 Token 数';
COMMENT ON COLUMN model_token_statistics.avg_output_tokens IS '平均输出 Token 数';
COMMENT ON COLUMN model_token_statistics.p50_input_tokens IS '输入 Token 的 P50 分位数';
COMMENT ON COLUMN model_token_statistics.p50_output_tokens IS '输出 Token 的 P50 分位数';
COMMENT ON COLUMN model_token_statistics.p90_input_tokens IS '输入 Token 的 P90 分位数';
COMMENT ON COLUMN model_token_statistics.p90_output_tokens IS '输出 Token 的 P90 分位数';
COMMENT ON COLUMN model_token_statistics.input_output_ratio IS '输入输出 Token 比例：avg_input_tokens / avg_output_tokens';
COMMENT ON COLUMN model_token_statistics.total_requests IS '总请求数：用于统计的样本总数';
COMMENT ON COLUMN model_token_statistics.sample_start_date IS '样本开始日期：统计数据的时间范围起点';
COMMENT ON COLUMN model_token_statistics.sample_end_date IS '样本结束日期：统计数据的时间范围终点';
COMMENT ON COLUMN model_token_statistics.created_at IS '创建时间';
COMMENT ON COLUMN model_token_statistics.updated_at IS '更新时间';

-- ============================================================================
-- 3. 创建索引
-- ============================================================================

-- 模型名称索引（唯一）
CREATE UNIQUE INDEX IF NOT EXISTS idx_model_token_statistics_model_name ON model_token_statistics(model_name);

-- 样本日期索引：用于按时间范围查询统计数据
CREATE INDEX IF NOT EXISTS idx_model_token_statistics_sample_dates ON model_token_statistics(sample_start_date, sample_end_date);

-- 更新时间索引：用于查找最近更新的统计数据
CREATE INDEX IF NOT EXISTS idx_model_token_statistics_updated_at ON model_token_statistics(updated_at DESC);

-- 总请求数索引：用于查找样本量充足的统计数据
CREATE INDEX IF NOT EXISTS idx_model_token_statistics_total_requests ON model_token_statistics(total_requests DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- ============================================================================
-- 1. 删除索引
-- ============================================================================

DROP INDEX IF EXISTS idx_model_token_statistics_total_requests;
DROP INDEX IF EXISTS idx_model_token_statistics_updated_at;
DROP INDEX IF EXISTS idx_model_token_statistics_sample_dates;
DROP INDEX IF EXISTS idx_model_token_statistics_model_name;

-- ============================================================================
-- 2. 删除表
-- ============================================================================

DROP TABLE IF EXISTS model_token_statistics;

-- +goose StatementEnd
