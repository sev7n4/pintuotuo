-- Per-line marketing copy on entitlement package items (optional display name + value note)
-- Version: 055

ALTER TABLE entitlement_package_items
  ADD COLUMN IF NOT EXISTS display_name VARCHAR(120),
  ADD COLUMN IF NOT EXISTS value_note VARCHAR(500);
