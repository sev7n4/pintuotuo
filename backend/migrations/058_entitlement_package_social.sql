-- Entitlement package: favorites, likes, reviews + order linkage for sales stats
-- Version: 058

CREATE TABLE IF NOT EXISTS entitlement_package_favorites (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  package_id INT NOT NULL REFERENCES entitlement_packages(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (user_id, package_id)
);

CREATE INDEX IF NOT EXISTS idx_ep_fav_package ON entitlement_package_favorites(package_id);

CREATE TABLE IF NOT EXISTS entitlement_package_likes (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  package_id INT NOT NULL REFERENCES entitlement_packages(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (user_id, package_id)
);

CREATE INDEX IF NOT EXISTS idx_ep_like_package ON entitlement_package_likes(package_id);

CREATE TABLE IF NOT EXISTS entitlement_package_reviews (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  package_id INT NOT NULL REFERENCES entitlement_packages(id) ON DELETE CASCADE,
  rating SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
  comment TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (user_id, package_id)
);

CREATE INDEX IF NOT EXISTS idx_ep_rev_package ON entitlement_package_reviews(package_id);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'update_entitlement_package_reviews_updated_at'
  ) THEN
    CREATE TRIGGER update_entitlement_package_reviews_updated_at
      BEFORE UPDATE ON entitlement_package_reviews
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

ALTER TABLE orders ADD COLUMN IF NOT EXISTS entitlement_package_id INT REFERENCES entitlement_packages(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_orders_entitlement_package_id ON orders(entitlement_package_id)
  WHERE entitlement_package_id IS NOT NULL;
