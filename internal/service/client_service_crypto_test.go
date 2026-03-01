// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package service_test

import (
	"os"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/crypto"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// ComputeHash использует utils.HashJSONToString → utils.Hash → hasherPool.
	// Пул должен быть инициализирован до запуска любых тестов.
	utils.InitHasherPool("test-hash-key")
	os.Exit(m.Run())
}

func newRealCryptoSvc(t *testing.T) (service.ClientCryptoService, []byte) {
	t.Helper()
	keyChain := crypto.NewKeyChainService()
	svc := service.NewClientCryptoService(keyChain)

	dek, err := keyChain.GenerateDEK()
	require.NoError(t, err)

	svc.SetEncryptionKey(dek)
	return svc, dek
}

// --- EncryptPayload / DecryptPayload ---

func TestClientCryptoService_EncryptDecrypt_RoundTrip(t *testing.T) {
	svc, _ := newRealCryptoSvc(t)

	login := "user@example.com"
	pass := "s3cr3t"
	plain := models.DecipheredPayload{
		UserID:    1,
		Type:      models.LoginPassword,
		Metadata:  models.Metadata{Name: "My Bank"},
		LoginData: &models.LoginData{Username: login, Password: pass},
	}

	enc, err := svc.EncryptPayload(plain)
	require.NoError(t, err)

	// Зашифрованные поля не содержат данные в открытом виде
	assert.NotContains(t, string(enc.Metadata), "My Bank")
	assert.NotContains(t, string(enc.Data), login)
	assert.NotContains(t, string(enc.Data), pass)

	// Type не шифруется
	assert.Equal(t, plain.Type, enc.Type)

	got, err := svc.DecryptPayload(enc)
	require.NoError(t, err)

	assert.Equal(t, plain.Metadata, got.Metadata)
	require.NotNil(t, got.LoginData)
	assert.Equal(t, login, got.LoginData.Username)
	assert.Equal(t, pass, got.LoginData.Password)
	assert.Nil(t, got.Notes)
	assert.Nil(t, got.AdditionalFields)
}

func TestClientCryptoService_EncryptDecrypt_AllTypes(t *testing.T) {
	totp := "JBSWY3DPEHPK3PXP"

	tests := []struct {
		name    string
		payload models.DecipheredPayload
	}{
		{
			name: "LoginPassword",
			payload: models.DecipheredPayload{
				Type:     models.LoginPassword,
				Metadata: models.Metadata{Name: "GitHub"},
				LoginData: &models.LoginData{
					Username: "user@example.com",
					Password: "s3cr3t",
					URIs:     []models.LoginURI{{URI: "https://github.com", Match: 1}},
					TOTP:     &totp,
				},
			},
		},
		{
			name: "Text",
			payload: models.DecipheredPayload{
				Type:     models.Text,
				Metadata: models.Metadata{Name: "My Note"},
				TextData: &models.TextData{Text: "секретная заметка"},
				Notes:    &models.Notes{Notes: "доп. заметка", IsEncrypted: false},
			},
		},
		{
			name: "BankCard",
			payload: models.DecipheredPayload{
				Type:     models.BankCard,
				Metadata: models.Metadata{Name: "Visa"},
				BankCardData: &models.BankCardData{
					CardholderName: "JOHN DOE",
					Number:         "4111111111111111",
					Brand:          "Visa",
					ExpMonth:       "12",
					ExpYear:        "2027",
					Code:           "123",
				},
			},
		},
		{
			name: "Binary",
			payload: models.DecipheredPayload{
				Type:     models.Binary,
				Metadata: models.Metadata{Name: "passport.pdf"},
				BinaryData: &models.BinaryData{
					ID:       "blob-001",
					FileName: "passport.pdf",
					Size:     204800,
					Key:      "enc-key-ref",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newRealCryptoSvc(t)

			enc, err := svc.EncryptPayload(tt.payload)
			require.NoError(t, err)

			got, err := svc.DecryptPayload(enc)
			require.NoError(t, err)

			// UserID и ClientSideID не шифруются — проставляем из оригинала
			got.UserID = tt.payload.UserID
			got.ClientSideID = tt.payload.ClientSideID

			assert.Equal(t, tt.payload, got)
		})
	}
}

func TestClientCryptoService_EncryptDecrypt_WithOptionalFields(t *testing.T) {
	svc, _ := newRealCryptoSvc(t)

	notes := models.Notes{Notes: "важная заметка"}
	fields := []models.CustomField{{Type: models.Text, Data: "1234"}}

	plain := models.DecipheredPayload{
		UserID:           1,
		Type:             models.Text,
		Metadata:         models.Metadata{Name: "Note"},
		TextData:         &models.TextData{Text: "секретный текст"},
		Notes:            &notes,
		AdditionalFields: &fields,
	}

	enc, err := svc.EncryptPayload(plain)
	require.NoError(t, err)
	require.NotNil(t, enc.Notes)
	require.NotNil(t, enc.AdditionalFields)

	got, err := svc.DecryptPayload(enc)
	require.NoError(t, err)

	require.NotNil(t, got.Notes)
	assert.Equal(t, notes.Notes, got.Notes.Notes)
	require.NotNil(t, got.AdditionalFields)
	assert.Equal(t, fields, *got.AdditionalFields)
}

func TestClientCryptoService_EncryptDecrypt_NilOptionalFields(t *testing.T) {
	svc, _ := newRealCryptoSvc(t)

	plain := models.DecipheredPayload{
		UserID:       1,
		Type:         models.BankCard,
		Metadata:     models.Metadata{Name: "Card"},
		BankCardData: &models.BankCardData{Number: "4111111111111111"},
	}

	enc, err := svc.EncryptPayload(plain)
	require.NoError(t, err)
	assert.Nil(t, enc.Notes)
	assert.Nil(t, enc.AdditionalFields)

	got, err := svc.DecryptPayload(enc)
	require.NoError(t, err)
	assert.Nil(t, got.Notes)
	assert.Nil(t, got.AdditionalFields)
}

func TestClientCryptoService_Decrypt_WrongKey(t *testing.T) {
	svc, _ := newRealCryptoSvc(t)

	enc, err := svc.EncryptPayload(models.DecipheredPayload{
		UserID:   1,
		Metadata: models.Metadata{Name: "Secret"},
	})
	require.NoError(t, err)

	// Подменяем ключ
	wrongDEK, err := crypto.NewKeyChainService().GenerateDEK()
	require.NoError(t, err)
	svc.SetEncryptionKey(wrongDEK)

	_, err = svc.DecryptPayload(enc)
	require.Error(t, err)
}

// --- Свойства шифрования ---

func TestClientCryptoService_Encrypt_IsDifferentEachTime(t *testing.T) {
	svc, _ := newRealCryptoSvc(t)

	plain := models.DecipheredPayload{UserID: 1, Metadata: models.Metadata{Name: "Test"}}

	enc1, err := svc.EncryptPayload(plain)
	require.NoError(t, err)
	enc2, err := svc.EncryptPayload(plain)
	require.NoError(t, err)

	// AES-GCM с random nonce — каждый раз разный шифртекст
	assert.NotEqual(t, enc1.Metadata, enc2.Metadata)
	assert.NotEqual(t, enc1.Data, enc2.Data)
}

// --- ComputeHash ---

func TestClientCryptoService_ComputeHash_Deterministic(t *testing.T) {
	svc, _ := newRealCryptoSvc(t)

	payload := models.PrivateDataPayload{Type: models.LoginPassword, Data: "somedata"}

	h1, err := svc.ComputeHash(payload)
	require.NoError(t, err)
	h2, err := svc.ComputeHash(payload)
	require.NoError(t, err)

	assert.Equal(t, h1, h2)
	assert.NotEmpty(t, h1)
}

func TestClientCryptoService_ComputeHash_DifferentForDifferentData(t *testing.T) {
	svc, _ := newRealCryptoSvc(t)

	h1, err := svc.ComputeHash(models.PrivateDataPayload{Data: "aaa"})
	require.NoError(t, err)
	h2, err := svc.ComputeHash(models.PrivateDataPayload{Data: "bbb"})
	require.NoError(t, err)

	assert.NotEqual(t, h1, h2)
}

// --- SetEncryptionKey ---

func TestClientCryptoService_SetEncryptionKey_ChangesKey(t *testing.T) {
	svc, _ := newRealCryptoSvc(t)

	enc, err := svc.EncryptPayload(models.DecipheredPayload{
		UserID:   1,
		Metadata: models.Metadata{Name: "Test"},
	})
	require.NoError(t, err)

	newDEK, err := crypto.NewKeyChainService().GenerateDEK()
	require.NoError(t, err)
	svc.SetEncryptionKey(newDEK)

	_, err = svc.DecryptPayload(enc)
	require.Error(t, err) // Новый ключ — расшифровка должна упасть
}
