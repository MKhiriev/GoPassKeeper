// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

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

// keyChainService is the private implementation of [KeyChainService].
type keyChainService struct {
	// Argon2id tuning parameters. Stored in the struct so they can be
	// adjusted per deployment target (e.g. mobile vs. desktop).
	argonTime    uint32
	argonMemory  uint32
	argonThreads uint8
	argonKeyLen  uint32
}

// NewKeyChainService constructs a [KeyChainService] with the Argon2id
// parameters recommended by OWASP (2024):
//   - time cost:   1 iteration
//   - memory cost: 64 MiB
//   - parallelism: 4 threads
//   - key length:  32 bytes (256 bits)
func NewKeyChainService() KeyChainService {
	return &keyChainService{
		argonTime:    1,
		argonMemory:  64 * 1024, // 64 MiB
		argonThreads: 4,
		argonKeyLen:  32, // 256 bits
	}
}

// GenerateEncryptionSalt implements [KeyChainService]. It reads 16 random
// bytes from the OS CSPRNG and returns them as the encryption salt. Returns
// an error if the random read fails.
func (k *keyChainService) GenerateEncryptionSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// GenerateDEK implements [KeyChainService]. It reads 32 random bytes from
// the OS CSPRNG and returns them as the data-encryption key. Returns an
// error if the random read fails.
func (k *keyChainService) GenerateDEK() ([]byte, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, err
	}
	return dek, nil
}

// GenerateKEK implements [KeyChainService]. It derives a 256-bit
// key-encryption key from masterPassword and salt using Argon2id with the
// parameters stored in the receiver. The result exists only in client memory
// and is never transmitted to the server.
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

// GetEncryptedDEK implements [KeyChainService]. It wraps DEK with KEK using
// AES-256-GCM. A random 12-byte nonce is prepended to the ciphertext so
// that the decryption side can locate it: blob = nonce ‖ ciphertext.
// Returns an error if cipher creation or the random nonce read fails.
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

	// Prepend the nonce so DecryptDEK can split it out without side-channel.
	encryptedDEK := gcm.Seal(nil, nonce, DEK, nil)
	return append(nonce, encryptedDEK...), nil
}

// GenerateAuthHash implements [KeyChainService]. It computes
// SHA-256(KEK ‖ authSalt) and returns the digest. The fixed authSalt
// domain-separates this hash from the KEK itself, ensuring the two values
// have different purposes even if derived from the same material.
func (k *keyChainService) GenerateAuthHash(KEK []byte, authSalt string) []byte {
	h := sha256.New()
	h.Write(KEK)
	h.Write([]byte(authSalt)) // authSalt domain-separates AuthHash from KEK
	return h.Sum(nil)
}

// DecryptDEK implements [KeyChainService]. It unwraps the encrypted DEK blob
// produced by [keyChainService.GetEncryptedDEK] using KEK and AES-256-GCM.
// The blob must be at least as long as the GCM nonce (12 bytes). Returns the
// plaintext DEK, or an error if the blob is too short, the KEK is wrong, or
// the ciphertext is corrupted (authentication-tag mismatch).
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

	// Split the blob into nonce and actual ciphertext.
	nonce, ciphertext := encryptedDEK[:nonceSize], encryptedDEK[nonceSize:]

	// Decrypt and verify auth tag. An error here almost always means the
	// user entered the wrong master password, producing a wrong KEK.
	dek, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return dek, nil
}

// EncryptData implements [KeyChainService]. It marshals data to JSON, then
// encrypts it with DEK using AES-256-GCM. The output is a Base64
// (standard encoding) string of the blob: nonce (12 bytes) ‖ ciphertext.
// Returns an error if marshalling, cipher creation, or nonce generation fails.
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

// DecryptData implements [KeyChainService]. It Base64-decodes encryptedB64,
// splits out the nonce, decrypts the ciphertext with DEK via AES-256-GCM,
// and unmarshals the resulting JSON into target. target must be a non-nil
// pointer, identical to the requirement of [encoding/json.Unmarshal]. Returns
// an error if any step (decoding, cipher creation, decryption, or
// unmarshalling) fails.
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
