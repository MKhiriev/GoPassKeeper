package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/argon2"
)

// keyChainService — приватная реализация интерфейса KeyChainService.
type keyChainService struct {
	// Параметры Argon2id. Вынесены в структуру, чтобы легко менять
	// под разные устройства (мобильный vs десктоп).
	argonTime    uint32
	argonMemory  uint32
	argonThreads uint8
	argonKeyLen  uint32
}

// NewKeyChainService создает новый экземпляр KeyChainService
// с рекомендуемыми параметрами Argon2id (OWASP 2024).
func NewKeyChainService() KeyChainService {
	return &keyChainService{
		argonTime:    1,
		argonMemory:  64 * 1024, // 64 MB
		argonThreads: 4,
		argonKeyLen:  32, // 256 бит
	}
}

func (k *keyChainService) GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func (k *keyChainService) GenerateDEK() ([]byte, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, err
	}
	return dek, nil
}

func (k *keyChainService) GenerateKEK(masterPassword string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(masterPassword),
		salt,
		k.argonTime,
		k.argonMemory,
		k.argonThreads,
		k.argonKeyLen,
	)
}

func (k *keyChainService) GetEncryptedDEK(DEK, KEK []byte) ([]byte, error) {
	block, err := aes.NewCipher(KEK)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Nonce добавляем в начало, чтобы при расшифровке знать где он
	encryptedDEK := gcm.Seal(nil, nonce, DEK, nil)
	return append(nonce, encryptedDEK...), nil
}

func (k *keyChainService) GenerateAuthHash(KEK []byte, authSalt string) []byte {
	h := sha256.New()
	h.Write(KEK)
	h.Write([]byte(authSalt)) // authSalt отделяет AuthHash от KEK (разное назначение)
	return h.Sum(nil)
}
