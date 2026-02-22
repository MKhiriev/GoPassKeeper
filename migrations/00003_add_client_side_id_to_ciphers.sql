-- +goose Up
-- +goose StatementBegin

ALTER TABLE ciphers
    ADD COLUMN IF NOT EXISTS client_side_id VARCHAR(40);

ALTER TABLE ciphers
    DROP CONSTRAINT IF EXISTS ciphers_id_user_id_type_key;

ALTER TABLE ciphers
    ADD CONSTRAINT ciphers_id_user_id_client_side_id_key
        UNIQUE (user_id, client_side_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

ALTER TABLE ciphers
    DROP CONSTRAINT IF EXISTS ciphers_id_user_id_client_side_id_key;

ALTER TABLE ciphers
    ADD CONSTRAINT ciphers_id_user_id_type_key
        UNIQUE (id, user_id, type);

ALTER TABLE ciphers
    DROP COLUMN IF EXISTS client_side_id;

-- +goose StatementEnd