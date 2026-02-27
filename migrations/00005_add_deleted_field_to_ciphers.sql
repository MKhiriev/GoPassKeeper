-- +goose Up
-- +goose StatementBegin

ALTER TABLE ciphers
    ADD COLUMN IF NOT EXISTS deleted BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

ALTER TABLE ciphers
    DROP COLUMN IF EXISTS deleted;

-- +goose StatementEnd