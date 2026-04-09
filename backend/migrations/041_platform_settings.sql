-- 平台级可配置项（健康调度成本等），变更通过 NOTIFY 热更新各实例。
CREATE TABLE IF NOT EXISTS platform_settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO platform_settings (key, value) VALUES
  ('health_scheduler_enabled', 'true'),
  ('health_scheduler_interval_seconds', '3600'),
  ('health_scheduler_batch', '2')
ON CONFLICT (key) DO NOTHING;

CREATE OR REPLACE FUNCTION notify_platform_settings_changed() RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('platform_settings_changed', '');
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tr_platform_settings_notify ON platform_settings;
CREATE TRIGGER tr_platform_settings_notify
  AFTER INSERT OR UPDATE OR DELETE ON platform_settings
  FOR EACH ROW EXECUTE FUNCTION notify_platform_settings_changed();

COMMENT ON TABLE platform_settings IS 'Platform-wide key-value settings; LISTEN platform_settings_changed for hot reload';
