-- Notify other backend pods / listeners when routing_strategies change (including direct SQL).
CREATE OR REPLACE FUNCTION notify_routing_strategies_changed() RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('routing_strategies_changed', '');
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tr_routing_strategies_notify ON routing_strategies;
CREATE TRIGGER tr_routing_strategies_notify
  AFTER INSERT OR UPDATE OR DELETE ON routing_strategies
  FOR EACH ROW EXECUTE FUNCTION notify_routing_strategies_changed();

COMMENT ON FUNCTION notify_routing_strategies_changed() IS 'Broadcast routing_strategies changes to LISTEN routing_strategies_changed';
