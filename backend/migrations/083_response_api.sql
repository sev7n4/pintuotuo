CREATE TABLE IF NOT EXISTS stored_responses (
    id SERIAL PRIMARY KEY,
    response_id VARCHAR(64) NOT NULL UNIQUE,
    user_id INT NOT NULL REFERENCES users(id),
    merchant_id INT NOT NULL,
    model VARCHAR(100) NOT NULL,
    input JSONB NOT NULL,
    output JSONB,
    tool_calls JSONB,
    usage JSONB,
    status VARCHAR(20) DEFAULT 'completed',
    background_job_id VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP + INTERVAL '7 days',
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_stored_responses_user ON stored_responses(user_id);
CREATE INDEX IF NOT EXISTS idx_stored_responses_response_id ON stored_responses(response_id);
CREATE INDEX IF NOT EXISTS idx_stored_responses_expires ON stored_responses(expires_at);
CREATE INDEX IF NOT EXISTS idx_stored_responses_background_job ON stored_responses(background_job_id) WHERE background_job_id IS NOT NULL;

INSERT INTO endpoint_pricing (endpoint_type, provider_code, unit_type, unit_price) VALUES
    ('responses', 'openai', 'token', 0.001)
ON CONFLICT (endpoint_type, provider_code) DO NOTHING;

-- Down Migration
-- DROP TABLE IF EXISTS stored_responses;
-- DELETE FROM endpoint_pricing WHERE endpoint_type = 'responses';
