-- +goose Up
-- +goose StatementBegin

ALTER TABLE ciphers
    ADD COLUMN IF NOT EXISTS client_side_id VARCHAR(40);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

ALTER TABLE ciphers
    DROP COLUMN IF EXISTS client_side_id;

-- +goose StatementEnd