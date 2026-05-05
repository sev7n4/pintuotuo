ALTER TABLE api_usage_logs
  DROP CONSTRAINT api_usage_logs_key_id_fkey,
  ADD CONSTRAINT api_usage_logs_key_id_fkey
    FOREIGN KEY (key_id) REFERENCES merchant_api_keys(id) ON DELETE CASCADE;

ALTER TABLE routing_decisions
  DROP CONSTRAINT routing_decisions_selected_api_key_id_fkey,
  ADD CONSTRAINT routing_decisions_selected_api_key_id_fkey
    FOREIGN KEY (selected_api_key_id) REFERENCES merchant_api_keys(id) ON DELETE SET NULL;

ALTER TABLE routing_decision_logs
  DROP CONSTRAINT routing_decision_logs_api_key_id_fkey,
  ADD CONSTRAINT routing_decision_logs_api_key_id_fkey
    FOREIGN KEY (api_key_id) REFERENCES merchant_api_keys(id) ON DELETE SET NULL;

ALTER TABLE merchant_sku_cost_history
  DROP CONSTRAINT merchant_sku_cost_history_merchant_api_key_id_fkey,
  ADD CONSTRAINT merchant_sku_cost_history_merchant_api_key_id_fkey
    FOREIGN KEY (merchant_api_key_id) REFERENCES merchant_api_keys(id) ON DELETE SET NULL;

ALTER TABLE merchant_skus
  DROP CONSTRAINT merchant_skus_api_key_id_fkey,
  ADD CONSTRAINT merchant_skus_api_key_id_fkey
    FOREIGN KEY (api_key_id) REFERENCES merchant_api_keys(id) ON DELETE SET NULL;
