-- SPDX-License-Identifier: Apache-2.0
-- Copyright 2026 Rasul Khiriev

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    user_id              INTEGER  PRIMARY KEY AUTOINCREMENT,
    login                TEXT     NOT NULL UNIQUE,
    auth_hash            TEXT     NOT NULL,
    master_password_hint TEXT,
    name                 TEXT,
    created_at           DATETIME DEFAULT (datetime('now')),
    updated_at           DATETIME,
    encryption_salt      TEXT     NOT NULL,
    encrypted_master_key TEXT     NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS data_types
(
    id          INTEGER NOT NULL PRIMARY KEY,
    description TEXT    NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ciphers
(
    id                INTEGER  PRIMARY KEY AUTOINCREMENT,
    user_id           INTEGER  NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    type              INTEGER  NOT NULL REFERENCES data_types (id),
    metadata          TEXT     NOT NULL,
    data              TEXT     NOT NULL,
    notes             TEXT,
    additional_fields TEXT,
    created_at        DATETIME DEFAULT (datetime('now')),
    updated_at        DATETIME,
    version           INTEGER  NOT NULL,
    client_side_id    TEXT     NOT NULL,
    hash              TEXT     NOT NULL DEFAULT '',
    deleted           INTEGER  NOT NULL DEFAULT 0,
    CONSTRAINT ciphers_id_user_id_client_side_id_key UNIQUE (user_id, client_side_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ciphers;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS data_types;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd