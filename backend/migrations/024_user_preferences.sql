-- User Preferences Migration
-- Version: 024
-- Date: 2026-04-04
-- Description: 添加用户偏好设置字段，支持场景筛选和预算级别

-- ============================================================
-- Phase 1: 添加用户偏好设置字段
-- ============================================================

ALTER TABLE users ADD COLUMN IF NOT EXISTS preferred_scenarios JSONB DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS budget_level VARCHAR(20) DEFAULT 'medium';

COMMENT ON COLUMN users.preferred_scenarios IS '用户偏好的使用场景列表，如 ["coding", "writing", "analysis"]';
COMMENT ON COLUMN users.budget_level IS '用户预算级别: low/medium/high';

-- ============================================================
-- Phase 2: 创建用户偏好设置历史表
-- ============================================================

CREATE TABLE IF NOT EXISTS user_preference_history (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  preference_type VARCHAR(50) NOT NULL,
  old_value JSONB,
  new_value JSONB,
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE user_preference_history IS '用户偏好设置变更历史';
COMMENT ON COLUMN user_preference_history.preference_type IS '偏好类型: preferred_scenarios/budget_level';

CREATE INDEX IF NOT EXISTS idx_user_pref_history_user ON user_preference_history(user_id);
CREATE INDEX IF NOT EXISTS idx_user_pref_history_type ON user_preference_history(preference_type);
CREATE INDEX IF NOT EXISTS idx_user_pref_history_created ON user_preference_history(created_at);

-- ============================================================
-- Phase 3: 创建更新时间触发器
-- ============================================================

CREATE OR REPLACE FUNCTION update_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION update_users_updated_at();
