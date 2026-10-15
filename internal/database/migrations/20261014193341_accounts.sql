-- +goose Up
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS accounts (
    "uuid" TEXT PRIMARY KEY,
    "discord_id" TEXT NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
SELECT 'down SQL query';

DROP TABLE IF EXISTS accounts;
