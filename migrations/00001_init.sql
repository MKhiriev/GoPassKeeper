-- +goose Up
-- +goose StatementBegin

-- =====================================================
-- USERS
-- =====================================================
CREATE TABLE IF NOT EXISTS users (
    user_id BIGSERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    master_password TEXT NOT NULL,
    master_password_hint TEXT,
    name TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
);

-- =====================================================
-- DATA TYPES
-- =====================================================
CREATE TABLE IF NOT EXISTS data_types (
    id INTEGER PRIMARY KEY,
    description TEXT NOT NULL
);
INSERT INTO data_types (id, description)
VALUES
    (1, 'login_password'),
    (2, 'text'),
    (3, 'binary'),
    (4, 'bank_card')
ON CONFLICT (id) DO NOTHING;

-- =====================================================
-- CIPHERS (PrivateData)
-- =====================================================
CREATE TABLE IF NOT EXISTS ciphers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type INTEGER NOT NULL REFERENCES data_types(id),
    metadata TEXT NOT NULL,
    data TEXT NOT NULL,
    notes TEXT,
    additional_fields TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,

    UNIQUE (id, user_id, type)
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP
    TABLE IF EXISTS ciphers;
DROP
    TABLE IF EXISTS data_types;
DROP
    TABLE IF EXISTS users;
-- +goose StatementEnd
