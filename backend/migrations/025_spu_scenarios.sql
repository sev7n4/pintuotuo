-- SPU Scenarios Migration
-- Version: 025
-- Date: 2026-04-04
-- Description: 创建SPU场景分类表，支持场景筛选功能

-- ============================================================
-- Phase 1: 创建使用场景表
-- ============================================================

CREATE TABLE IF NOT EXISTS usage_scenarios (
  id SERIAL PRIMARY KEY,
  code VARCHAR(50) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  icon_url VARCHAR(500),
  sort_order INT DEFAULT 0,
  status VARCHAR(20) DEFAULT 'active',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE usage_scenarios IS '使用场景分类表';
COMMENT ON COLUMN usage_scenarios.code IS '场景代码，如 coding, writing, analysis';
COMMENT ON COLUMN usage_scenarios.name IS '场景名称';
COMMENT ON COLUMN usage_scenarios.description IS '场景描述';
COMMENT ON COLUMN usage_scenarios.icon_url IS '场景图标URL';
COMMENT ON COLUMN usage_scenarios.sort_order IS '排序顺序';
COMMENT ON COLUMN usage_scenarios.status IS '状态: active/inactive';

CREATE INDEX IF NOT EXISTS idx_usage_scenarios_code ON usage_scenarios(code);
CREATE INDEX IF NOT EXISTS idx_usage_scenarios_status ON usage_scenarios(status);

-- ============================================================
-- Phase 2: 创建SPU场景关联表
-- ============================================================

CREATE TABLE IF NOT EXISTS spu_scenarios (
  id SERIAL PRIMARY KEY,
  spu_id INT NOT NULL REFERENCES spus(id) ON DELETE CASCADE,
  scenario_id INT NOT NULL REFERENCES usage_scenarios(id) ON DELETE CASCADE,
  is_primary BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  UNIQUE(spu_id, scenario_id)
);

COMMENT ON TABLE spu_scenarios IS 'SPU与场景关联表';
COMMENT ON COLUMN spu_scenarios.is_primary IS '是否为主要场景';

CREATE INDEX IF NOT EXISTS idx_spu_scenarios_spu ON spu_scenarios(spu_id);
CREATE INDEX IF NOT EXISTS idx_spu_scenarios_scenario ON spu_scenarios(scenario_id);

-- ============================================================
-- Phase 3: 添加SPU性能字段
-- ============================================================

ALTER TABLE spus ADD COLUMN IF NOT EXISTS avg_latency_ms INT;
ALTER TABLE spus ADD COLUMN IF NOT EXISTS availability_rate DECIMAL(5, 2) DEFAULT 99.90;
ALTER TABLE spus ADD COLUMN IF NOT EXISTS last_health_check_at TIMESTAMP;

COMMENT ON COLUMN spus.avg_latency_ms IS '平均响应延迟（毫秒）';
COMMENT ON COLUMN spus.availability_rate IS '可用率百分比，如 99.90';
COMMENT ON COLUMN spus.last_health_check_at IS '最后健康检查时间';

-- ============================================================
-- Phase 4: 插入场景种子数据
-- ============================================================

INSERT INTO usage_scenarios (code, name, description, sort_order) VALUES
  ('coding', '代码开发', '代码编写、调试、重构等开发场景', 1),
  ('writing', '内容创作', '文章写作、文案创作、翻译等', 2),
  ('analysis', '数据分析', '数据分析、报表生成、洞察提取', 3),
  ('chat', '智能对话', '日常对话、问答咨询、客服场景', 4),
  ('vision', '图像理解', '图像识别、图像描述、视觉问答', 5),
  ('embedding', '向量嵌入', '文本嵌入、语义搜索、相似度计算', 6),
  ('audio', '语音处理', '语音识别、语音合成、音频处理', 7),
  ('reasoning', '复杂推理', '逻辑推理、数学计算、复杂问题求解', 8)
ON CONFLICT (code) DO NOTHING;

-- ============================================================
-- Phase 5: 创建更新时间触发器
-- ============================================================

CREATE OR REPLACE FUNCTION update_usage_scenarios_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_usage_scenarios_updated_at
  BEFORE UPDATE ON usage_scenarios
  FOR EACH ROW
  EXECUTE FUNCTION update_usage_scenarios_updated_at();
