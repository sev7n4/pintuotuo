-- Idempotent subscription expiry reminders (7d / 1d before end_date)

CREATE TABLE IF NOT EXISTS subscription_reminders (
  id SERIAL PRIMARY KEY,
  subscription_id INT NOT NULL REFERENCES user_subscriptions(id) ON DELETE CASCADE,
  kind VARCHAR(10) NOT NULL CHECK (kind IN ('7d', '1d')),
  sent_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  channel VARCHAR(32),
  UNIQUE (subscription_id, kind)
);

CREATE INDEX IF NOT EXISTS idx_subscription_reminders_sent ON subscription_reminders(sent_at);

COMMENT ON TABLE subscription_reminders IS 'One row per (subscription, kind) to dedupe scheduled expiry reminders.';
