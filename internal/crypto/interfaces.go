package crypto

//go:generate mockgen -source=interfaces.go -destination=../mock/keychain_service_mock.go -package=mock

// KeyChainService отвечает за всю клиентскую криптографию в схеме Zero-Knowledge.
// Он не знает ничего о сети, базе данных или пользователях.
// Его единственная задача — генерировать и защищать ключи.
//
// Схема работы:
//
//	Salt, DEK = GenerateEncryptionSalt() + GenerateDEK()   (Шаг 1)
//	KEK       = GenerateKEK(password, salt)       (Шаг 2)
//	EncDEK    = GetEncryptedDEK(DEK, KEK)         (Шаг 3)
//	AuthHash  = GenerateAuthHash(KEK, authSalt)   (Шаг 4)
type KeyChainService interface {
	// GenerateEncryptionSalt генерирует случайную соль (16 байт / 128 бит).
	// Соль не является секретом — она хранится на сервере открыто.
	// Нужна для того, чтобы одинаковые пароли давали разные KEK.
	// Шаг 1.
	GenerateEncryptionSalt() ([]byte, error)

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

	// DecryptDEK unwraps the encrypted DEK using the KEK.
	// It expects the input blob to be in the format: nonce || ciphertext.
	// Returns the original DEK or an error if authentication fails (e.g., wrong password/KEK).
	DecryptDEK(encryptedDEK, KEK []byte) ([]byte, error)

	// EncryptData serializes the given value to JSON and encrypts it with the DEK.
	// Returns a base64-encoded blob (nonce || ciphertext) safe to store on the server.
	EncryptData(data any, DEK []byte) (string, error)

	// DecryptData decrypts a base64-encoded blob with the DEK and unmarshals
	// the result into the target pointer (same as json.Unmarshal).
	DecryptData(encryptedB64 string, DEK []byte, target any) error
}
