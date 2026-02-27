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

// NewClientAuthService constructs a clientAuthService wired to the provided local
// store, server adapter, key-chain service, and crypto service.
// The returned service is safe for concurrent use.
func NewClientAuthService(localStore *store.ClientStorages, serverAdapter adapter.ServerAdapter, crypto crypto.KeyChainService, cryptoSvc ClientCryptoService) ClientAuthService {
	return &clientAuthService{localStore: localStore, adapter: serverAdapter, crypto: crypto, clientCryptoService: cryptoSvc}
}

// Register implements ClientAuthService.
//
// Key-derivation steps:
//  1. Generate a random encryption salt.
//  2. Generate a random data-encryption key (DEK).
//  3. Derive a key-encryption key (KEK) from the master password and the salt.
//  4. Encrypt the DEK with the KEK to produce the encrypted master key.
//  5. Compute the auth hash from the KEK and the fixed auth salt.
//  6. Base64-encode the salt, encrypted DEK, and auth hash for safe storage.
//  7. Send the user record (without the plaintext password) to the server.
//
// Returns an error if any key-generation, encryption, or server call fails.
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

// Login implements ClientAuthService.
//
// Authentication steps:
//  1. Fetch the user's encryption salt from the server by login.
//  2. Decode the base64 salt and derive the KEK from the master password.
//  3. Compute the auth hash from the KEK and the fixed auth salt.
//  4. Send the login + auth hash to the server and receive the encrypted master key.
//  5. Decode the base64 encrypted master key and decrypt it with the KEK to get the DEK.
//  6. Store the DEK in the crypto service via SetEncryptionKey.
//
// Returns the server-assigned user ID and the plaintext DEK, or an error if
// any step fails.
func (a *clientAuthService) Login(ctx context.Context, user models.User) (int64, []byte, error) {
	// Fetch encryption_salt from the server by login.
	userWithSalt, err := a.adapter.RequestSalt(ctx, user)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %v", ErrLoginOnServer, err)
	}

	// Decode the salt and derive the KEK from the master password + salt.
	saltBytes, err := base64.StdEncoding.DecodeString(userWithSalt.EncryptionSalt)
	if err != nil {
		return 0, nil, fmt.Errorf("decode encryption salt: %w", err)
	}
	kek := a.crypto.GenerateKEK(user.MasterPassword, saltBytes)

	// Compute the auth hash and attach it to the user model.
	authHashBytes := a.crypto.GenerateAuthHash(kek, authSalt)
	user.AuthHash = base64.StdEncoding.EncodeToString(authHashBytes)

	// Send login + auth_hash to the server; receive the encrypted master key.
	foundUser, err := a.adapter.Login(ctx, user)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %v", ErrLoginOnServer, err)
	}

	// Decode the encrypted master key and decrypt the DEK using the KEK.
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
