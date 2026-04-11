-- Pricing versions (IE-1): immutable retail rate snapshots for order-time binding
-- Version: 045
-- Date: 2026-04-10

-- ============================================================
-- 1. Version header (one row per published price list)
-- ============================================================

CREATE TABLE IF NOT EXISTS pricing_versions (
  id SERIAL PRIMARY KEY,
  code VARCHAR(64) NOT NULL,
  description TEXT,
  effective_from TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uq_pricing_versions_code UNIQUE (code)
);

COMMENT ON TABLE pricing_versions IS '零售价目版本：用于订单绑定下单时点单价快照';
COMMENT ON COLUMN pricing_versions.code IS '唯一标识，如 baseline / v2026Q2';
COMMENT ON COLUMN pricing_versions.effective_from IS '版本对外生效时间（展示/审计用）';

CREATE INDEX IF NOT EXISTS idx_pricing_versions_effective ON pricing_versions(effective_from DESC);

-- ============================================================
-- 2. Per-SPU rates under a version (元/1K tokens，与 spus 口径一致)
-- ============================================================

CREATE TABLE IF NOT EXISTS pricing_version_spu_rates (
  id SERIAL PRIMARY KEY,
  pricing_version_id INT NOT NULL REFERENCES pricing_versions(id) ON DELETE CASCADE,
  spu_id INT NOT NULL REFERENCES spus(id) ON DELETE CASCADE,
  provider_input_rate DECIMAL(10, 6) NOT NULL,
  provider_output_rate DECIMAL(10, 6) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uq_pricing_version_spu UNIQUE (pricing_version_id, spu_id)
);

COMMENT ON TABLE pricing_version_spu_rates IS '某价目版本下各 SPU 的输入/输出单价快照';
COMMENT ON COLUMN pricing_version_spu_rates.provider_input_rate IS '元/1K input tokens';
COMMENT ON COLUMN pricing_version_spu_rates.provider_output_rate IS '元/1K output tokens';

CREATE INDEX IF NOT EXISTS idx_pv_spu_rates_version ON pricing_version_spu_rates(pricing_version_id);
CREATE INDEX IF NOT EXISTS idx_pv_spu_rates_spu ON pricing_version_spu_rates(spu_id);

-- ============================================================
-- 3. 订单绑定价目版本（NULL = 历史订单或未绑定，扣费仍走现有逻辑）
-- ============================================================

ALTER TABLE orders ADD COLUMN IF NOT EXISTS pricing_version_id INT REFERENCES pricing_versions(id);

COMMENT ON COLUMN orders.pricing_version_id IS '下单时绑定的价目版本；NULL 表示未启用版本化扣费';

CREATE INDEX IF NOT EXISTS idx_orders_pricing_version ON orders(pricing_version_id);

-- ============================================================
-- 4. Seed: baseline 版本 + 从当前 spus 复制一版快照（幂等）
-- ============================================================

INSERT INTO pricing_versions (code, description, effective_from)
VALUES (
  'baseline',
  '045 migration: snapshot copied from spus.provider_*_rate at first apply',
  CURRENT_TIMESTAMP
)
ON CONFLICT (code) DO NOTHING;

INSERT INTO pricing_version_spu_rates (pricing_version_id, spu_id, provider_input_rate, provider_output_rate)
SELECT pv.id, s.id, s.provider_input_rate, s.provider_output_rate
FROM spus s
JOIN pricing_versions pv ON pv.code = 'baseline'
WHERE s.provider_input_rate IS NOT NULL
  AND s.provider_output_rate IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM pricing_version_spu_rates r
    WHERE r.pricing_version_id = pv.id AND r.spu_id = s.id
  );
