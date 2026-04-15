-- Merchant invite-only registration, tenant id, lifecycle, MFA, admin IP / audit support

-- 1) 商户邀请码（平台生成，用于二维码链接中的 code）
CREATE TABLE IF NOT EXISTS merchant_invites (
    id SERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    max_uses INT NOT NULL DEFAULT 1 CHECK (max_uses > 0),
    used_count INT NOT NULL DEFAULT 0 CHECK (used_count >= 0),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    note TEXT,
    created_by_user_id INT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_merchant_invites_code ON merchant_invites(code) WHERE revoked_at IS NULL;

-- 2) 商户：对外租户 ID + 运营生命周期（与审核流 status 并存）
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS tenant_public_id UUID UNIQUE DEFAULT gen_random_uuid();
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS lifecycle_status VARCHAR(20) NOT NULL DEFAULT 'trial';
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS merchant_invite_id INT REFERENCES merchant_invites(id) ON DELETE SET NULL;

-- 生命周期：试用 / 正式 / 暂停（与 pending/reviewing 等审核状态独立）
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'merchants_lifecycle_status_check'
  ) THEN
    ALTER TABLE merchants ADD CONSTRAINT merchants_lifecycle_status_check
      CHECK (lifecycle_status IN ('trial', 'active', 'suspended'));
  END IF;
END $$;

UPDATE merchants SET lifecycle_status = 'active'
WHERE lifecycle_status = 'trial'
  AND status IN ('approved', 'active');

COMMENT ON COLUMN merchants.tenant_public_id IS '对外稳定租户标识（UUID），可写入合同/对账';
COMMENT ON COLUMN merchants.lifecycle_status IS '运营生命周期: trial | active | suspended';
COMMENT ON COLUMN merchants.merchant_invite_id IS '注册时使用的邀请记录';

-- 每租户功能开关（预留，避免后期大改）
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS feature_flags JSONB NOT NULL DEFAULT '{}';
COMMENT ON COLUMN merchants.feature_flags IS '租户级功能开关 JSON，如 {"billing_v2": true}';

-- 3) 用户 MFA（TOTP 密钥加密存储）
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_totp_secret_enc TEXT;

COMMENT ON COLUMN users.mfa_totp_secret_enc IS 'AES 加密后的 TOTP Base32 密钥';
