-- 手机号与第三方账号绑定（扩展认证）
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20) UNIQUE;
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone) WHERE phone IS NOT NULL AND phone <> '';

CREATE TABLE IF NOT EXISTS user_identity_links (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider VARCHAR(32) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  display_name VARCHAR(255),
  meta JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(provider, external_id)
);

CREATE INDEX IF NOT EXISTS idx_user_identity_links_user ON user_identity_links(user_id);
