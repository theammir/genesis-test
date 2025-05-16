-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS confirmations (
	email varchar(128) REFERENCES subscriptions(email),
	token varchar(128) NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS confirmations;
-- +goose StatementEnd
