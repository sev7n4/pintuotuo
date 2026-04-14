-- Entitlement package ops fields: schedule + recommendation + badge
-- Version: 053

ALTER TABLE entitlement_packages
  ADD COLUMN IF NOT EXISTS start_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS end_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS is_featured BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS badge_text VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_entitlement_packages_featured_sort
  ON entitlement_packages(is_featured, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_entitlement_packages_schedule
  ON entitlement_packages(start_at, end_at);
