-- SPDX-License-Identifier: Apache-2.0
-- Copyright 2026 Rasul Khiriev

-- +goose Up
-- +goose StatementBegin

ALTER TABLE ciphers
    ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

ALTER TABLE ciphers
    DROP COLUMN IF EXISTS version;

-- +goose StatementEnd