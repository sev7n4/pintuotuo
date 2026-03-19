-- Performance Optimization Migration
-- Add composite indexes for common query patterns

-- Orders table composite indexes
CREATE INDEX IF NOT EXISTS idx_orders_user_status ON orders(user_id, status);
CREATE INDEX IF NOT EXISTS idx_orders_user_created ON orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status_created ON orders(status, created_at DESC);

-- Products table composite indexes
CREATE INDEX IF NOT EXISTS idx_products_status_created ON products(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_merchant_status ON products(merchant_id, status);

-- Payments table composite indexes
CREATE INDEX IF NOT EXISTS idx_payments_user_status ON payments(user_id, status);
CREATE INDEX IF NOT EXISTS idx_payments_status_created ON payments(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_order_status ON payments(order_id, status);

-- Token transactions composite indexes
CREATE INDEX IF NOT EXISTS idx_token_trans_user_created ON token_transactions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_token_trans_user_type ON token_transactions(user_id, type);

-- API usage logs composite indexes (critical for consumption queries)
CREATE INDEX IF NOT EXISTS idx_api_usage_logs_user_created ON api_usage_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_usage_logs_user_provider ON api_usage_logs(user_id, provider);
CREATE INDEX IF NOT EXISTS idx_api_usage_logs_created_status ON api_usage_logs(created_at DESC, status_code);

-- Merchant API keys composite indexes
CREATE INDEX IF NOT EXISTS idx_merchant_apikeys_provider_status ON merchant_api_keys(provider, status);
CREATE INDEX IF NOT EXISTS idx_merchant_apikeys_merchant_status ON merchant_api_keys(merchant_id, status);

-- Merchant settlements composite indexes
CREATE INDEX IF NOT EXISTS idx_merchant_settlements_merchant_status ON merchant_settlements(merchant_id, status);
CREATE INDEX IF NOT EXISTS idx_merchant_settlements_merchant_period ON merchant_settlements(merchant_id, period_start, period_end);

-- Notifications composite indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications(user_id, created_at DESC);

-- Groups composite indexes
CREATE INDEX IF NOT EXISTS idx_groups_status_deadline ON groups(status, deadline);
CREATE INDEX IF NOT EXISTS idx_groups_product_status ON groups(product_id, status);

-- Referral codes index
CREATE INDEX IF NOT EXISTS idx_referral_codes_code ON referral_codes(code);

-- Referral rewards composite indexes
CREATE INDEX IF NOT EXISTS idx_referral_rewards_referrer_status ON referral_rewards(referrer_id, status);
CREATE INDEX IF NOT EXISTS idx_referral_rewards_referrer_created ON referral_rewards(referrer_id, created_at DESC);

-- Referrals composite indexes
CREATE INDEX IF NOT EXISTS idx_referrals_referrer_created ON referrals(referrer_id, created_at DESC);

-- Add partial indexes for common status filters
CREATE INDEX IF NOT EXISTS idx_orders_pending ON orders(created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_products_active ON products(created_at DESC) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_groups_active ON groups(deadline) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(created_at DESC) WHERE is_read = FALSE;

-- Optimize text search on products
CREATE INDEX IF NOT EXISTS idx_products_name_gin ON products USING gin(to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS idx_products_description_gin ON products USING gin(to_tsvector('simple', description));

-- Add covering indexes for frequent queries
CREATE INDEX IF NOT EXISTS idx_orders_covering ON orders(user_id, status, created_at DESC) INCLUDE (product_id, total_price);

-- Analyze tables to update statistics
ANALYZE users;
ANALYZE products;
ANALYZE orders;
ANALYZE payments;
ANALYZE tokens;
ANALYZE token_transactions;
ANALYZE api_usage_logs;
ANALYZE merchants;
ANALYZE merchant_api_keys;
ANALYZE merchant_settlements;
ANALYZE notifications;
ANALYZE groups;
ANALYZE referral_codes;
ANALYZE referral_rewards;
ANALYZE referrals;
