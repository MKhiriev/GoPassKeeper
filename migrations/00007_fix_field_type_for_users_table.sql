-- SPDX-License-Identifier: Apache-2.0
-- Copyright 2026 Rasul Khiriev

-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ALTER COLUMN login TYPE VARCHAR(20),
    ALTER COLUMN user_id DROP DEFAULT,
    ALTER COLUMN user_id ADD GENERATED ALWAYS AS IDENTITY;

DROP SEQUENCE IF EXISTS users_user_id_seq;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    ALTER COLUMN login TYPE TEXT,
    ALTER COLUMN user_id DROP IDENTITY,
    ALTER COLUMN user_id SET DEFAULT nextval('users_user_id_seq');

CREATE SEQUENCE users_user_id_seq OWNED BY users.user_id;
-- +goose StatementEnd