-- Admin merchant audit trail and optional business classification

ALTER TABLE merchants ADD COLUMN IF NOT EXISTS business_category VARCHAR(64);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS admin_notes TEXT;

CREATE INDEX IF NOT EXISTS idx_merchants_business_category ON merchants(business_category);

CREATE TABLE IF NOT EXISTS merchant_audit_logs (
    id SERIAL PRIMARY KEY,
    merchant_id INTEGER NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    admin_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(32) NOT NULL,
    company_name_snapshot VARCHAR(200),
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_merchant ON merchant_audit_logs(merchant_id);
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_created ON merchant_audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_merchant_audit_logs_admin ON merchant_audit_logs(admin_user_id);

COMMENT ON TABLE merchant_audit_logs IS '管理员对商户审核与关键操作的审计记录';
COMMENT ON COLUMN merchants.business_category IS '经营类目（管理员维护或入驻时填写）';
COMMENT ON COLUMN merchants.admin_notes IS '管理员内部备注，不对商户端展示';
