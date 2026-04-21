-- Version: 060
-- Description: add structured provider error fields for BYOK health history

ALTER TABLE api_key_health_history
  ADD COLUMN IF NOT EXISTS status_code INT,
  ADD COLUMN IF NOT EXISTS provider_error_code VARCHAR(128),
  ADD COLUMN IF NOT EXISTS provider_request_id VARCHAR(128),
  ADD COLUMN IF NOT EXISTS endpoint_used VARCHAR(512),
  ADD COLUMN IF NOT EXISTS error_category VARCHAR(64),
  ADD COLUMN IF NOT EXISTS raw_error_excerpt TEXT;

COMMENT ON COLUMN api_key_health_history.status_code IS '上游响应状态码';
COMMENT ON COLUMN api_key_health_history.provider_error_code IS '上游错误码（如 invalid_api_key）';
COMMENT ON COLUMN api_key_health_history.provider_request_id IS '上游请求ID（用于工单排查）';
COMMENT ON COLUMN api_key_health_history.endpoint_used IS '探测实际使用的 endpoint';
COMMENT ON COLUMN api_key_health_history.error_category IS '平台标准错误分类';
COMMENT ON COLUMN api_key_health_history.raw_error_excerpt IS '上游错误体摘要（脱敏/截断）';

CREATE INDEX IF NOT EXISTS idx_api_key_health_history_api_key_created
  ON api_key_health_history(api_key_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_api_key_health_history_error_category
  ON api_key_health_history(error_category);

CREATE INDEX IF NOT EXISTS idx_api_key_health_history_provider_error_code
  ON api_key_health_history(provider_error_code);
