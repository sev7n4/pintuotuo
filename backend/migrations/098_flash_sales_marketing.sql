-- Flash sale admin/marketing fields aligned with entitlement_packages display patterns
-- Version: 098

ALTER TABLE flash_sales
  ADD COLUMN IF NOT EXISTS is_featured BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS badge_text VARCHAR(50) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS badge_text_secondary VARCHAR(50) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS marketing_line VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS promo_label VARCHAR(80) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS promo_ends_at TIMESTAMPTZ;
