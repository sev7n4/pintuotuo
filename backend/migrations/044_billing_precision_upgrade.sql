-- Billing precision upgrade
-- Version: 044
-- 目标：统一元/1K口径下的小额高精度入账，避免四舍五入到 0

ALTER TABLE api_usage_logs
  ALTER COLUMN cost TYPE DECIMAL(20, 10);

ALTER TABLE settlement_items
  ALTER COLUMN cost TYPE DECIMAL(20, 10);

ALTER TABLE token_transactions
  ALTER COLUMN amount TYPE DECIMAL(20, 6);

ALTER TABLE tokens
  ALTER COLUMN balance TYPE DECIMAL(20, 6),
  ALTER COLUMN total_used TYPE DECIMAL(20, 6),
  ALTER COLUMN total_earned TYPE DECIMAL(20, 6);

ALTER TABLE merchant_api_keys
  ALTER COLUMN quota_limit TYPE DECIMAL(20, 6),
  ALTER COLUMN quota_used TYPE DECIMAL(20, 6);

ALTER TABLE merchant_settlements
  ALTER COLUMN total_sales TYPE DECIMAL(20, 6),
  ALTER COLUMN platform_fee TYPE DECIMAL(20, 6),
  ALTER COLUMN settlement_amount TYPE DECIMAL(20, 6);

ALTER TABLE settlement_disputes
  ALTER COLUMN original_amount TYPE DECIMAL(20, 6),
  ALTER COLUMN disputed_amount TYPE DECIMAL(20, 6),
  ALTER COLUMN adjusted_amount TYPE DECIMAL(20, 6);

COMMENT ON COLUMN api_usage_logs.cost IS '请求成本（元，高精度）';
COMMENT ON COLUMN settlement_items.cost IS '结算明细成本（元，高精度）';
COMMENT ON COLUMN token_transactions.amount IS '账户变动金额（元，高精度）';
COMMENT ON COLUMN tokens.balance IS '用户余额（元，高精度）';
COMMENT ON COLUMN merchant_api_keys.quota_limit IS '商户 API Key 配额上限（元，高精度）';
COMMENT ON COLUMN merchant_api_keys.quota_used IS '商户 API Key 已用配额（元，高精度）';
