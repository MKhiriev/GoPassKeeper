package crypto

// KeyChainService отвечает за всю клиентскую криптографию в схеме Zero-Knowledge.
// Он не знает ничего о сети, базе данных или пользователях.
// Его единственная задача — генерировать и защищать ключи.
//
// Схема работы:
//
//	Salt, DEK = GenerateSalt() + GenerateDEK()   (Шаг 1)
//	KEK       = GenerateKEK(password, salt)       (Шаг 2)
//	EncDEK    = GetEncryptedDEK(DEK, KEK)         (Шаг 3)
//	AuthHash  = GenerateAuthHash(KEK, authSalt)   (Шаг 4)
type KeyChainService interface {
	// GenerateSalt генерирует случайную соль (16 байт / 128 бит).
	// Соль не является секретом — она хранится на сервере открыто.
	// Нужна для того, чтобы одинаковые пароли давали разные KEK.
	// Шаг 1.
	GenerateSalt() ([]byte, error)

	// GenerateDEK генерирует случайный мастер-ключ данных (32 байта / 256 бит).
	// DEK шифрует все данные пользователя и никогда не покидает клиента в открытом виде.
	// Шаг 1.
	GenerateDEK() ([]byte, error)

	// GenerateKEK выводит ключ шифрования из пароля и соли через Argon2id.
	// KEK существует только в памяти клиента и никогда не отправляется на сервер.
	// Шаг 2.
	GenerateKEK(masterPassword string, salt []byte) []byte

	// GetEncryptedDEK шифрует DEK с помощью KEK через AES-GCM.
	// Результат (Nonce + Ciphertext) безопасно хранить на сервере —
	// без KEK это просто случайный шум.
	// Шаг 3.
	GetEncryptedDEK(DEK, KEK []byte) ([]byte, error)

	// GenerateAuthHash создает "ключ-пропуск" для сервера.
	// Это SHA-256 от KEK + authSalt. Сервер сравнивает его при логине,
	// но не может вычислить KEK обратно (хеш необратим).
	// Шаг 4.
	GenerateAuthHash(KEK []byte, authSalt string) []byte
}
