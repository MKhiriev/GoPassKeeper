-- +goose Up
-- +goose StatementBegin

ALTER TABLE ciphers
    ADD COLUMN IF NOT EXISTS version BIGINT;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

ALTER TABLE ciphers
    DROP COLUMN IF EXISTS version;

-- +goose StatementEnd