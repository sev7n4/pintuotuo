-- 邀请返利：提现申请与按订单去重（避免重复发放）

CREATE TABLE IF NOT EXISTS referral_withdrawals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    method VARCHAR(20) NOT NULL,
    account_info TEXT NOT NULL,
    request_note TEXT,
    reject_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_referral_withdrawals_user ON referral_withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_referral_withdrawals_status ON referral_withdrawals(status);

-- 同一订单仅一条返利记录（幂等）；若历史存在重复需先清理再执行
CREATE UNIQUE INDEX IF NOT EXISTS idx_referral_rewards_order_unique
    ON referral_rewards(order_id)
    WHERE order_id IS NOT NULL;
