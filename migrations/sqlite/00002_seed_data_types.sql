-- SPDX-License-Identifier: Apache-2.0
-- Copyright 2026 Rasul Khiriev

-- +goose Up
-- +goose StatementBegin
INSERT INTO data_types (id, description)
VALUES
    (1, 'login_password'),
    (2, 'text'),
    (3, 'binary'),
    (4, 'bank_card')
ON CONFLICT (id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM data_types WHERE id IN (1, 2, 3, 4);
-- +goose StatementEnd