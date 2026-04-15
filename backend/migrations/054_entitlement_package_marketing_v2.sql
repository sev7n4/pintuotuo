-- Entitlement package: category, secondary badge, marketing line, lightweight promo/share hints
-- Version: 054

ALTER TABLE entitlement_packages
  ADD COLUMN IF NOT EXISTS category_code VARCHAR(40) NOT NULL DEFAULT 'general',
  ADD COLUMN IF NOT EXISTS badge_text_secondary VARCHAR(50),
  ADD COLUMN IF NOT EXISTS marketing_line VARCHAR(255),
  ADD COLUMN IF NOT EXISTS promo_label VARCHAR(80),
  ADD COLUMN IF NOT EXISTS promo_ends_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_entitlement_packages_category
  ON entitlement_packages(category_code, sort_order, id);
