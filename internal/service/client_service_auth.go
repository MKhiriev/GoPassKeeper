package service

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/crypto"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
)

var authSalt = "my-super-secret-auth-salt"

type clientAuthService struct {
	localStore          *store.ClientStorages
	adapter             adapter.ServerAdapter
	clientCryptoService ClientCryptoService
	crypto              crypto.KeyChainService
}

func NewClientAuthService(localStore *store.ClientStorages, serverAdapter adapter.ServerAdapter, crypto crypto.KeyChainService, cryptoSvc ClientCryptoService) ClientAuthService {
	return &clientAuthService{localStore: localStore, adapter: serverAdapter, crypto: crypto, clientCryptoService: cryptoSvc}
}

func (a *clientAuthService) Register(ctx context.Context, user models.User) error {
	salt, err := a.crypto.GenerateEncryptionSalt()
	if err != nil {
		return fmt.Errorf("error generating Salt: %v", err)
	}

	dek, err := a.crypto.GenerateDEK()
	if err != nil {
		return fmt.Errorf("error generating DEK: %v", err)
	}

	kek := a.crypto.GenerateKEK(user.MasterPassword, salt)

	encryptedDek, err := a.crypto.GetEncryptedDEK(dek, kek)
	if err != nil {
		return fmt.Errorf("error encription DEK: %v", err)
	}

	authHashBytes := a.crypto.GenerateAuthHash(kek, authSalt)

	// All byte slices are base64-encoded for safe storage in the database.
	user.EncryptionSalt = base64.StdEncoding.EncodeToString(salt)
	user.EncryptedMasterKey = base64.StdEncoding.EncodeToString(encryptedDek)
	user.AuthHash = base64.StdEncoding.EncodeToString(authHashBytes)

	user.MasterPassword = ""

	_, err = a.adapter.Register(ctx, user)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRegisterOnServer, err)
	}

	return nil
}

func (a *clientAuthService) Login(ctx context.Context, user models.User) (int64, []byte, error) {
	// L2: получаем encryption_salt с сервера по логину
	userWithSalt, err := a.adapter.RequestSalt(ctx, user)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %v", ErrLoginOnServer, err)
	}

	// L3: декодируем соль и вычисляем KEK из пароля + соли
	saltBytes, err := base64.StdEncoding.DecodeString(userWithSalt.EncryptionSalt)
	if err != nil {
		return 0, nil, fmt.Errorf("decode encryption salt: %w", err)
	}
	kek := a.crypto.GenerateKEK(user.MasterPassword, saltBytes)

	// L4: вычисляем AuthHash и кладем в модель
	authHashBytes := a.crypto.GenerateAuthHash(kek, authSalt)
	user.AuthHash = base64.StdEncoding.EncodeToString(authHashBytes)

	// L5: отправляем login + auth_hash, получаем encrypted_master_key
	foundUser, err := a.adapter.Login(ctx, user)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %v", ErrLoginOnServer, err)
	}

	// L6: декодируем encrypted_master_key и расшифровываем DEK с помощью KEK
	encryptedBlob, err := base64.StdEncoding.DecodeString(foundUser.EncryptedMasterKey)
	if err != nil {
		return 0, nil, fmt.Errorf("decode encrypted master key: %w", err)
	}

	dek, err := a.crypto.DecryptDEK(encryptedBlob, kek)
	if err != nil {
		return 0, nil, fmt.Errorf("decrypt DEK: %w", err)
	}

	a.clientCryptoService.SetEncryptionKey(dek)

	return foundUser.UserID, dek, nil
}
