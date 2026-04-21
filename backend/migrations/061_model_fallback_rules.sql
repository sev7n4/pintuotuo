-- Model-level fallback chains (OpenAI-compat catalog keys: provider/model)
-- Version: 061

CREATE TABLE IF NOT EXISTS model_fallback_rules (
  id SERIAL PRIMARY KEY,
  source_model VARCHAR(512) NOT NULL,
  fallback_models TEXT[] NOT NULL DEFAULT '{}',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT uq_model_fallback_rules_source UNIQUE (source_model)
);

CREATE INDEX IF NOT EXISTS idx_model_fallback_rules_enabled ON model_fallback_rules(enabled);

COMMENT ON TABLE model_fallback_rules IS 'Per-source catalog model keys and ordered fallback list; validated against active SPU + model_providers.';
