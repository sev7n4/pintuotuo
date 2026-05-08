-- Migration: 093_api_usage_logs_endpoint_type
-- Description: Add endpoint_type column to api_usage_logs for endpoint filtering
-- Created: 2026-05-08

ALTER TABLE api_usage_logs ADD COLUMN IF NOT EXISTS endpoint_type VARCHAR(50) DEFAULT 'chat_completions';

CREATE INDEX IF NOT EXISTS idx_api_usage_logs_endpoint_type ON api_usage_logs(endpoint_type);

-- Down Migration
-- ALTER TABLE api_usage_logs DROP COLUMN IF EXISTS endpoint_type;
-- DROP INDEX IF EXISTS idx_api_usage_logs_endpoint_type;
