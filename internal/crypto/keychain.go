package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func (k *keyChainService) GenerateEncryptionSalt() ([]byte, error) {
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

func (k *keyChainService) DecryptDEK(encryptedDEK, KEK []byte) ([]byte, error) {
	block, err := aes.NewCipher(KEK)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedDEK) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Split the blob into nonce and actual ciphertext
	nonce, ciphertext := encryptedDEK[:nonceSize], encryptedDEK[nonceSize:]

	// Decrypt and verify (Open returns an error if the KEK is wrong or data is corrupted)
	dek, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// This error usually means the user entered the WRONG PASSWORD,
		// which resulted in a WRONG KEK.
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return dek, nil
}

func (k *keyChainService) EncryptData(data any, DEK []byte) (string, error) {
	// 1. Serialize to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal data: %w", err)
	}

	// 2. Build AES-GCM cipher from DEK
	block, err := aes.NewCipher(DEK)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	// 3. Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	// 4. Encrypt: nonce || ciphertext
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	blob := append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(blob), nil
}

func (k *keyChainService) DecryptData(encryptedB64 string, DEK []byte, target any) error {
	// 1. Decode base64 blob
	blob, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return fmt.Errorf("decode base64: %w", err)
	}

	// 2. Build AES-GCM cipher from DEK
	block, err := aes.NewCipher(DEK)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create gcm: %w", err)
	}

	// 3. Split nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(blob) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := blob[:nonceSize], blob[nonceSize:]

	// 4. Decrypt and verify auth tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decrypt data: %w", err)
	}

	// 5. Unmarshal JSON into target
	if err := json.Unmarshal(plaintext, target); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}

	return nil
}
