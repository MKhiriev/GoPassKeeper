package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type clientCryptoService struct{}

func NewClientCryptoService() ClientCryptoService {
	return &clientCryptoService{}
}

func (c *clientCryptoService) DeriveKey(masterPassword string, userID int64) []byte {
	seed := []byte(masterPassword + ":" + strconv.FormatInt(userID, 10))
	sum := sha256.Sum256(seed)
	buf := sum[:]
	for i := 0; i < 4096; i++ {
		tmp := sha256.Sum256(buf)
		buf = tmp[:]
	}
	out := make([]byte, 32)
	copy(out, buf)
	return out
}

func (c *clientCryptoService) EncryptPayload(plain models.PrivateDataPayload, key []byte) (models.PrivateDataPayload, error) {
	encMeta, err := encryptString(string(plain.Metadata), key)
	if err != nil {
		return models.PrivateDataPayload{}, fmt.Errorf("encrypt metadata: %w", err)
	}
	encData, err := encryptString(string(plain.Data), key)
	if err != nil {
		return models.PrivateDataPayload{}, fmt.Errorf("encrypt data: %w", err)
	}

	out := models.PrivateDataPayload{Metadata: models.CipheredMetadata(encMeta), Type: plain.Type, Data: models.CipheredData(encData)}
	if plain.Notes != nil {
		encNotes, err := encryptString(string(*plain.Notes), key)
		if err != nil {
			return models.PrivateDataPayload{}, fmt.Errorf("encrypt notes: %w", err)
		}
		n := models.CipheredNotes(encNotes)
		out.Notes = &n
	}
	if plain.AdditionalFields != nil {
		encFields, err := encryptString(string(*plain.AdditionalFields), key)
		if err != nil {
			return models.PrivateDataPayload{}, fmt.Errorf("encrypt additional fields: %w", err)
		}
		f := models.CipheredCustomFields(encFields)
		out.AdditionalFields = &f
	}
	return out, nil
}

func (c *clientCryptoService) DecryptPayload(cipherPayload models.PrivateDataPayload, key []byte) (models.PrivateDataPayload, error) {
	plainMeta, err := decryptString(string(cipherPayload.Metadata), key)
	if err != nil {
		return models.PrivateDataPayload{}, fmt.Errorf("decrypt metadata: %w", err)
	}
	plainData, err := decryptString(string(cipherPayload.Data), key)
	if err != nil {
		return models.PrivateDataPayload{}, fmt.Errorf("decrypt data: %w", err)
	}

	out := models.PrivateDataPayload{Metadata: models.CipheredMetadata(plainMeta), Type: cipherPayload.Type, Data: models.CipheredData(plainData)}
	if cipherPayload.Notes != nil {
		plainNotes, err := decryptString(string(*cipherPayload.Notes), key)
		if err != nil {
			return models.PrivateDataPayload{}, fmt.Errorf("decrypt notes: %w", err)
		}
		n := models.CipheredNotes(plainNotes)
		out.Notes = &n
	}
	if cipherPayload.AdditionalFields != nil {
		plainFields, err := decryptString(string(*cipherPayload.AdditionalFields), key)
		if err != nil {
			return models.PrivateDataPayload{}, fmt.Errorf("decrypt additional fields: %w", err)
		}
		f := models.CipheredCustomFields(plainFields)
		out.AdditionalFields = &f
	}

	return out, nil
}

func (c *clientCryptoService) ComputeHash(payload models.PrivateDataPayload) string {
	h := sha256.New()
	h.Write([]byte(payload.Metadata))
	h.Write([]byte(strconv.FormatInt(int64(payload.Type), 10)))
	h.Write([]byte(payload.Data))
	if payload.Notes != nil {
		h.Write([]byte(*payload.Notes))
	}
	if payload.AdditionalFields != nil {
		h.Write([]byte(*payload.AdditionalFields))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func encryptString(value string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("invalid key length: %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}

	ct := gcm.Seal(nil, nonce, []byte(value), nil)
	blob := append(nonce, ct...)
	return base64.StdEncoding.EncodeToString(blob), nil
}

func decryptString(value string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("invalid key length: %d", len(key))
	}
	blob, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(blob) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ct := blob[:nonceSize], blob[nonceSize:]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}
