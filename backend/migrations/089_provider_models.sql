CREATE TABLE IF NOT EXISTS provider_models (
    id                      SERIAL PRIMARY KEY,
    provider_code           VARCHAR(50) NOT NULL REFERENCES model_providers(code) ON DELETE CASCADE,
    model_id                VARCHAR(200) NOT NULL,
    display_name            VARCHAR(200),
    reference_input_price   DECIMAL(10,6),
    reference_output_price  DECIMAL(10,6),
    reference_currency      VARCHAR(10) DEFAULT 'USD',
    is_active               BOOLEAN DEFAULT true,
    synced_at               TIMESTAMP WITH TIME ZONE,
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_code, model_id)
);

CREATE INDEX idx_provider_models_provider_code ON provider_models(provider_code);
CREATE INDEX idx_provider_models_active ON provider_models(provider_code, is_active);
