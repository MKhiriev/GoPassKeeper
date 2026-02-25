package service

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/crypto"
	"github.com/MKhiriev/go-pass-keeper/internal/mock"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// newTestAuthSvc — хелпер для создания clientAuthService с моками
func newTestAuthSvc(
	t *testing.T,
	ctrl *gomock.Controller,
) (
	*clientAuthService,
	*mock.MockServerAdapter,
	*mock.MockKeyChainService,
	*mock.MockClientCryptoService,
) {
	t.Helper()
	mockAdapter := mock.NewMockServerAdapter(ctrl)
	mockKeyChain := mock.NewMockKeyChainService(ctrl)
	mockCryptoSvc := mock.NewMockClientCryptoService(ctrl)

	storages := &store.ClientStorages{}

	svc := NewClientAuthService(storages, mockAdapter, mockKeyChain).(*clientAuthService)
	svc.clientCryptoService = mockCryptoSvc

	return svc, mockAdapter, mockKeyChain, mockCryptoSvc
}

// ── Register ─────────────────────────────────────────────────────────────────

func TestClientAuthService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, mockKeyChain, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	salt := []byte("random-salt-16bb")
	dek := []byte("random-dek-32-bytes-placeholder!!")
	kek := []byte("derived-kek-bytes")
	encryptedDek := []byte("encrypted-dek-blob")
	authHash := []byte("auth-hash-bytes")

	user := models.User{
		Login:          "testuser",
		MasterPassword: "super-secret-password",
	}

	gomock.InOrder(
		mockKeyChain.EXPECT().GenerateEncryptionSalt().Return(salt, nil),
		mockKeyChain.EXPECT().GenerateDEK().Return(dek, nil),
		mockKeyChain.EXPECT().GenerateKEK(user.MasterPassword, salt).Return(kek),
		mockKeyChain.EXPECT().GetEncryptedDEK(dek, kek).Return(encryptedDek, nil),
		mockKeyChain.EXPECT().GenerateAuthHash(kek, authSalt).Return(authHash),
		// Проверяем что адаптер вызывается с правильными base64 значениями и очищенным паролем
		mockAdapter.EXPECT().Register(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, u models.User) (models.User, error) {
				assert.Equal(t, base64.StdEncoding.EncodeToString(salt), u.EncryptionSalt)
				assert.Equal(t, base64.StdEncoding.EncodeToString(encryptedDek), u.EncryptedMasterKey)
				assert.Equal(t, base64.StdEncoding.EncodeToString(authHash), u.AuthHash)
				assert.Empty(t, u.MasterPassword, "MasterPassword должен быть очищен перед отправкой")
				return u, nil
			},
		),
	)

	err := svc.Register(ctx, user)
	require.NoError(t, err)
}

func TestClientAuthService_Register_GenerateSaltError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, mockKeyChain, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	mockKeyChain.EXPECT().GenerateEncryptionSalt().Return(nil, errors.New("entropy exhausted"))

	err := svc.Register(ctx, models.User{MasterPassword: "pass"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error generating Salt")
}

func TestClientAuthService_Register_GenerateDEKError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, mockKeyChain, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	mockKeyChain.EXPECT().GenerateEncryptionSalt().Return([]byte("salt"), nil)
	mockKeyChain.EXPECT().GenerateDEK().Return(nil, errors.New("dek generation failed"))

	err := svc.Register(ctx, models.User{MasterPassword: "pass"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error generating DEK")
}

func TestClientAuthService_Register_EncryptDEKError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, mockKeyChain, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	salt := []byte("salt")
	dek := []byte("dek")
	kek := []byte("kek")

	mockKeyChain.EXPECT().GenerateEncryptionSalt().Return(salt, nil)
	mockKeyChain.EXPECT().GenerateDEK().Return(dek, nil)
	mockKeyChain.EXPECT().GenerateKEK("pass", salt).Return(kek)
	mockKeyChain.EXPECT().GetEncryptedDEK(dek, kek).Return(nil, errors.New("aes-gcm seal failed"))

	err := svc.Register(ctx, models.User{MasterPassword: "pass"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error encription DEK")
}

func TestClientAuthService_Register_AdapterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, mockKeyChain, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	salt := []byte("salt")
	dek := []byte("dek")
	kek := []byte("kek")
	encDek := []byte("enc-dek")
	authHash := []byte("hash")

	mockKeyChain.EXPECT().GenerateEncryptionSalt().Return(salt, nil)
	mockKeyChain.EXPECT().GenerateDEK().Return(dek, nil)
	mockKeyChain.EXPECT().GenerateKEK("pass", salt).Return(kek)
	mockKeyChain.EXPECT().GetEncryptedDEK(dek, kek).Return(encDek, nil)
	mockKeyChain.EXPECT().GenerateAuthHash(kek, authSalt).Return(authHash)
	mockAdapter.EXPECT().Register(ctx, gomock.Any()).Return(models.User{}, errors.New("server unavailable"))

	err := svc.Register(ctx, models.User{MasterPassword: "pass"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRegisterOnServer)
}

// ── Login ────────────────────────────────────────────────────────────────────

func TestClientAuthService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, mockKeyChain, mockCryptoSvc := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	salt := []byte("login-salt-bytes")
	kek := []byte("derived-kek")
	authHash := []byte("auth-hash")
	encryptedDEK := []byte("encrypted-dek-blob")
	dek := []byte("decrypted-dek-32bytes!!!!!!!!!!!!")
	wantUserID := int64(42)

	user := models.User{
		Login:          "testuser",
		MasterPassword: "my-password",
	}

	gomock.InOrder(
		// L2: RequestSalt возвращает соль в base64
		mockAdapter.EXPECT().RequestSalt(ctx, user).Return(models.User{
			EncryptionSalt: base64.StdEncoding.EncodeToString(salt),
		}, nil),
		// L3: GenerateKEK из пароля + соли
		mockKeyChain.EXPECT().GenerateKEK(user.MasterPassword, salt).Return(kek),
		// L4: GenerateAuthHash
		mockKeyChain.EXPECT().GenerateAuthHash(kek, authSalt).Return(authHash),
		// L5: Login возвращает userID и encrypted_master_key
		mockAdapter.EXPECT().Login(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, u models.User) (models.User, error) {
				assert.Equal(t, base64.StdEncoding.EncodeToString(authHash), u.AuthHash)
				return models.User{
					UserID:             wantUserID,
					EncryptedMasterKey: base64.StdEncoding.EncodeToString(encryptedDEK),
				}, nil
			},
		),
		// L6: DecryptDEK
		mockKeyChain.EXPECT().DecryptDEK(encryptedDEK, kek).Return(dek, nil),
		// SetEncryptionKey на CryptoService
		mockCryptoSvc.EXPECT().SetEncryptionKey(dek),
	)

	gotUserID, gotDEK, err := svc.Login(ctx, user)
	require.NoError(t, err)
	assert.Equal(t, wantUserID, gotUserID)
	assert.Equal(t, dek, gotDEK)
}

func TestClientAuthService_Login_RequestSaltError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, _, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	user := models.User{Login: "testuser", MasterPassword: "pass"}

	mockAdapter.EXPECT().RequestSalt(ctx, user).Return(models.User{}, errors.New("user not found"))

	_, _, err := svc.Login(ctx, user)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrLoginOnServer)
}

func TestClientAuthService_Login_InvalidSaltBase64(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, _, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	user := models.User{Login: "testuser", MasterPassword: "pass"}

	mockAdapter.EXPECT().RequestSalt(ctx, user).Return(models.User{
		EncryptionSalt: "%%%not-valid-base64%%%",
	}, nil)

	_, _, err := svc.Login(ctx, user)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode encryption salt")
}

func TestClientAuthService_Login_AdapterLoginError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, mockKeyChain, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	salt := []byte("salt")
	kek := []byte("kek")
	authHash := []byte("hash")

	user := models.User{Login: "testuser", MasterPassword: "pass"}

	mockAdapter.EXPECT().RequestSalt(ctx, user).Return(models.User{
		EncryptionSalt: base64.StdEncoding.EncodeToString(salt),
	}, nil)
	mockKeyChain.EXPECT().GenerateKEK(user.MasterPassword, salt).Return(kek)
	mockKeyChain.EXPECT().GenerateAuthHash(kek, authSalt).Return(authHash)
	mockAdapter.EXPECT().Login(ctx, gomock.Any()).Return(models.User{}, errors.New("wrong credentials"))

	_, _, err := svc.Login(ctx, user)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrLoginOnServer)
}

func TestClientAuthService_Login_InvalidEncryptedMasterKeyBase64(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, mockKeyChain, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	salt := []byte("salt")
	kek := []byte("kek")
	authHash := []byte("hash")

	user := models.User{Login: "testuser", MasterPassword: "pass"}

	mockAdapter.EXPECT().RequestSalt(ctx, user).Return(models.User{
		EncryptionSalt: base64.StdEncoding.EncodeToString(salt),
	}, nil)
	mockKeyChain.EXPECT().GenerateKEK(user.MasterPassword, salt).Return(kek)
	mockKeyChain.EXPECT().GenerateAuthHash(kek, authSalt).Return(authHash)
	mockAdapter.EXPECT().Login(ctx, gomock.Any()).Return(models.User{
		UserID:             1,
		EncryptedMasterKey: "%%%invalid-base64%%%",
	}, nil)

	_, _, err := svc.Login(ctx, user)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode encrypted master key")
}

func TestClientAuthService_Login_DecryptDEKError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, mockKeyChain, _ := newTestAuthSvc(t, ctrl)
	ctx := context.Background()

	salt := []byte("salt")
	kek := []byte("kek")
	authHash := []byte("hash")
	encryptedDEK := []byte("encrypted-blob")

	user := models.User{Login: "testuser", MasterPassword: "pass"}

	mockAdapter.EXPECT().RequestSalt(ctx, user).Return(models.User{
		EncryptionSalt: base64.StdEncoding.EncodeToString(salt),
	}, nil)
	mockKeyChain.EXPECT().GenerateKEK(user.MasterPassword, salt).Return(kek)
	mockKeyChain.EXPECT().GenerateAuthHash(kek, authSalt).Return(authHash)
	mockAdapter.EXPECT().Login(ctx, gomock.Any()).Return(models.User{
		UserID:             1,
		EncryptedMasterKey: base64.StdEncoding.EncodeToString(encryptedDEK),
	}, nil)
	mockKeyChain.EXPECT().DecryptDEK(encryptedDEK, kek).Return(nil, errors.New("cipher: message authentication failed"))

	_, _, err := svc.Login(ctx, user)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt DEK")
}

// ── Integration: реальная крипта, мок только адаптер ─────────────────────────

// newIntegrationAuthSvc создаёт authService с настоящим KeyChainService и
// настоящим ClientCryptoService. Мокается только ServerAdapter — он имитирует сервер.
func newIntegrationAuthSvc(
	t *testing.T,
	ctrl *gomock.Controller,
) (
	*clientAuthService,
	*mock.MockServerAdapter,
	ClientCryptoService,
) {
	t.Helper()

	keyChain := crypto.NewKeyChainService()
	mockAdapter := mock.NewMockServerAdapter(ctrl)
	cryptoSvc := NewClientCryptoService(keyChain)

	svc := NewClientAuthService(&store.ClientStorages{}, mockAdapter, keyChain).(*clientAuthService)
	svc.clientCryptoService = cryptoSvc

	return svc, mockAdapter, cryptoSvc
}

// TestIntegration_RegisterThenLogin_Success — полный round-trip:
// Register сохраняет данные на «сервер» (мок), Login получает их обратно,
// расшифровывает DEK настоящим AES-GCM и устанавливает ключ в CryptoService.
// После этого CryptoService может шифровать/расшифровывать данные.
func TestIntegration_RegisterThenLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, cryptoSvc := newIntegrationAuthSvc(t, ctrl)
	ctx := context.Background()

	password := "my-strong-master-password"
	wantUserID := int64(77)

	// «Сервер» — хранит то, что прислал Register, и отдаёт при Login.
	var serverUser models.User

	// ── Register ──
	mockAdapter.EXPECT().Register(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, u models.User) (models.User, error) {
			// Имитируем сервер: сохраняем пользователя
			serverUser = u
			serverUser.UserID = wantUserID
			assert.NotEmpty(t, u.EncryptionSalt)
			assert.NotEmpty(t, u.EncryptedMasterKey)
			assert.NotEmpty(t, u.AuthHash)
			assert.Empty(t, u.MasterPassword)
			return serverUser, nil
		},
	)

	err := svc.Register(ctx, models.User{Login: "alice", MasterPassword: password})
	require.NoError(t, err)

	// ── Login ──
	mockAdapter.EXPECT().RequestSalt(ctx, gomock.Any()).Return(models.User{
		EncryptionSalt: serverUser.EncryptionSalt,
	}, nil)

	mockAdapter.EXPECT().Login(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, u models.User) (models.User, error) {
			// Имитируем сервер: проверяем auth_hash и возвращаем encrypted_master_key
			assert.Equal(t, serverUser.AuthHash, u.AuthHash, "AuthHash при логине должен совпадать с регистрацией")
			return models.User{
				UserID:             wantUserID,
				EncryptedMasterKey: serverUser.EncryptedMasterKey,
			}, nil
		},
	)

	gotUserID, gotDEK, err := svc.Login(ctx, models.User{Login: "alice", MasterPassword: password})
	require.NoError(t, err)
	assert.Equal(t, wantUserID, gotUserID)
	assert.Len(t, gotDEK, 32, "DEK должен быть 32 байта (AES-256)")

	// ── Проверяем что DEK реально работает: шифруем и расшифровываем данные ──
	plain := models.DecipheredPayload{
		UserID:   wantUserID,
		Type:     models.LoginPassword,
		Metadata: models.Metadata{Name: "GitHub"},
		LoginData: &models.LoginData{
			Username: "alice@example.com",
			Password: "gh-secret-token",
		},
	}

	enc, err := cryptoSvc.EncryptPayload(plain)
	require.NoError(t, err)
	assert.NotContains(t, string(enc.Data), "alice@example.com")

	got, err := cryptoSvc.DecryptPayload(enc)
	require.NoError(t, err)
	assert.Equal(t, plain.Metadata, got.Metadata)
	require.NotNil(t, got.LoginData)
	assert.Equal(t, "alice@example.com", got.LoginData.Username)
	assert.Equal(t, "gh-secret-token", got.LoginData.Password)
}

// TestIntegration_LoginWithWrongPassword — после Register пытаемся Login
// с неправильным паролем. KEK будет другой → DecryptDEK упадёт.
func TestIntegration_LoginWithWrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockAdapter, _ := newIntegrationAuthSvc(t, ctrl)
	ctx := context.Background()

	var serverUser models.User

	// ── Register ──
	mockAdapter.EXPECT().Register(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, u models.User) (models.User, error) {
			serverUser = u
			return u, nil
		},
	)

	err := svc.Register(ctx, models.User{Login: "bob", MasterPassword: "correct-password"})
	require.NoError(t, err)

	// ── Login с неправильным паролем ──
	mockAdapter.EXPECT().RequestSalt(ctx, gomock.Any()).Return(models.User{
		EncryptionSalt: serverUser.EncryptionSalt,
	}, nil)

	mockAdapter.EXPECT().Login(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, u models.User) (models.User, error) {
			// AuthHash будет другой, но сервер-мок это пропускает —
			// нас интересует именно крипто-ошибка при расшифровке DEK
			return models.User{
				UserID:             1,
				EncryptedMasterKey: serverUser.EncryptedMasterKey,
			}, nil
		},
	)

	_, _, err = svc.Login(ctx, models.User{Login: "bob", MasterPassword: "wrong-password"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt DEK")
}
