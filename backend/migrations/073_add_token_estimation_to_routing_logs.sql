-- +goose Up
-- +goose StatementBegin

ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS estimated_input_tokens NUMERIC(10,2);
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS estimated_output_tokens NUMERIC(10,2);
ALTER TABLE routing_decision_logs ADD COLUMN IF NOT EXISTS token_estimation_source VARCHAR(50);

COMMENT ON COLUMN routing_decision_logs.estimated_input_tokens IS '预估输入 Token 数';
COMMENT ON COLUMN routing_decision_logs.estimated_output_tokens IS '预估输出 Token 数';
COMMENT ON COLUMN routing_decision_logs.token_estimation_source IS 'Token 预估来源：request/statistics/fallback';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE routing_decision_logs DROP COLUMN IF EXISTS token_estimation_source;
ALTER TABLE routing_decision_logs DROP COLUMN IF EXISTS estimated_output_tokens;
ALTER TABLE routing_decision_logs DROP COLUMN IF EXISTS estimated_input_tokens;

-- +goose StatementEnd
