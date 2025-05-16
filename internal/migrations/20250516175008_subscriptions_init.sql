-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
	CREATE TYPE frequency AS ENUM ('hourly', 'daily');
EXCEPTION
	WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS subscriptions (
		email varchar(128) NOT NULL PRIMARY KEY,
		city varchar(128) NOT NULL,
		frequency frequency NOT NULL,
		confirmed bool NOT NULL DEFAULT false
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS frequency;
DROP TABLE IF EXISTS subscriptions;
-- +goose StatementEnd
