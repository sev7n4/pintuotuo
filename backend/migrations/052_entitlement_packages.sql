-- Entitlement package definitions (admin-configurable bundle)
-- Version: 052

CREATE TABLE IF NOT EXISTS entitlement_packages (
  id SERIAL PRIMARY KEY,
  package_code VARCHAR(80) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS entitlement_package_items (
  id SERIAL PRIMARY KEY,
  package_id INT NOT NULL REFERENCES entitlement_packages(id) ON DELETE CASCADE,
  sku_id INT NOT NULL REFERENCES skus(id) ON DELETE RESTRICT,
  default_quantity INT NOT NULL DEFAULT 1 CHECK (default_quantity > 0),
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (package_id, sku_id)
);

CREATE INDEX IF NOT EXISTS idx_entitlement_packages_status_sort ON entitlement_packages(status, sort_order, id);
CREATE INDEX IF NOT EXISTS idx_entitlement_package_items_package_id ON entitlement_package_items(package_id);
CREATE INDEX IF NOT EXISTS idx_entitlement_package_items_sku_id ON entitlement_package_items(sku_id);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'update_entitlement_packages_updated_at'
  ) THEN
    CREATE TRIGGER update_entitlement_packages_updated_at
      BEFORE UPDATE ON entitlement_packages
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'update_entitlement_package_items_updated_at'
  ) THEN
    CREATE TRIGGER update_entitlement_package_items_updated_at
      BEFORE UPDATE ON entitlement_package_items
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;
