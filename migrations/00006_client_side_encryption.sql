-- +goose Up
-- Миграция 000006: Переход на схему Zero-Knowledge (Client-side Encryption)

-- 1. Добавляем поля для шифрования.
-- Делаем их NOT NULL, так как все новые пользователи обязаны их иметь.
ALTER TABLE users
    ADD COLUMN encryption_salt text NOT NULL,
    ADD COLUMN encrypted_master_key text NOT NULL;

-- 2. Переименовываем master_password в auth_hash.
-- Теперь в этой колонке будет храниться не пароль, а "клиентский отпечаток" (Auth Hash),
-- который сервер использует только для проверки входа, но не для шифрования.
ALTER TABLE users
    RENAME COLUMN master_password TO auth_hash;

-- Добавляем комментарии для документации БД
COMMENT ON COLUMN users.encryption_salt IS
    'Случайная соль (base64/hex), генерируемая клиентом. Нужна для вычисления KEK и AuthHash из пароля.';

COMMENT ON COLUMN users.encrypted_master_key IS
    'Зашифрованный мастер-ключ (DEK). Это "сейф", который открывается только через KEK пользователя.';

COMMENT ON COLUMN users.auth_hash IS
    'Клиентский хэш пароля. Сервер сравнивает его при логине, но не знает исходный пароль.';

-- +goose Down
-- Откат изменений (возвращаем как было)
ALTER TABLE users
    RENAME COLUMN auth_hash TO master_password;

ALTER TABLE users
    DROP COLUMN encryption_salt,
    DROP COLUMN encrypted_master_key;