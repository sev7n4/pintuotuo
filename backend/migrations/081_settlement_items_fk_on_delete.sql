ALTER TABLE settlement_items
  ALTER COLUMN api_usage_log_id DROP NOT NULL;

ALTER TABLE settlement_items
  DROP CONSTRAINT settlement_items_api_usage_log_id_fkey,
  ADD CONSTRAINT settlement_items_api_usage_log_id_fkey
    FOREIGN KEY (api_usage_log_id) REFERENCES api_usage_logs(id) ON DELETE SET NULL;
