-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE confirmations
DROP CONSTRAINT IF EXISTS confirmations_token_key;

ALTER TABLE confirmations
ALTER COLUMN token SET DATA TYPE uuid
USING gen_random_uuid();

ALTER TABLE confirmations
ALTER COLUMN token SET DEFAULT gen_random_uuid();

ALTER TABLE confirmations
ADD CONSTRAINT confirmations_token_key UNIQUE(token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE confirmations
DROP CONSTRAINT IF EXISTS confirmations_token_key;

ALTER TABLE confirmations
ALTER COLUMN token DROP DEFAULT;

ALTER TABLE confirmations
ALTER COLUMN token SET DATA TYPE varchar(128)
USING token::text;

ALTER TABLE confirmations
ALTER COLUMN token SET NOT NULL;

ALTER TABLE confirmations
ADD CONSTRAINT confirmations_token_key UNIQUE(token);
-- +goose StatementEnd
