package service

import (
	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/crypto"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

// ClientServices is the client-side service container. It groups all client
// service implementations and is constructed once at application startup via
// NewClientServices.
type ClientServices struct {
	// CryptoService handles client-side AES encryption and decryption of vault
	// payloads using the DEK derived from the user's master password.
	CryptoService ClientCryptoService

	// AuthService handles client-side user registration and authentication,
	// including key derivation (KEK/DEK) and server communication.
	AuthService ClientAuthService

	// PrivateDataService manages vault items on the client: CRUD operations
	// against the local SQLite store with automatic server propagation.
	PrivateDataService ClientPrivateDataService

	// SyncService performs bidirectional synchronisation between the local store
	// and the remote server, resolving conflicts via the sync plan algorithm.
	SyncService ClientSyncService

	// SyncJob is the background worker that periodically calls SyncService.FullSync
	// at a configurable interval while the user is logged in.
	SyncJob ClientSyncJob
}

// NewClientServices constructs and wires all client-side services.
//
// Initialisation order:
//  1. KeyChainService — provides low-level key-derivation and AES primitives.
//  2. ClientCryptoService — wraps KeyChainService for payload encryption.
//  3. ClientAuthService — handles registration/login using KeyChainService and
//     ClientCryptoService.
//  4. ClientPrivateDataService — CRUD service backed by the local store and
//     server adapter.
//  5. ClientSyncService — orchestrates full bidirectional sync.
//  6. ClientSyncJob — background ticker that calls FullSync periodically.
//
// Returns a fully initialised *ClientServices. The logger parameter is
// reserved for future structured logging and is currently unused.
func NewClientServices(localStore *store.ClientStorages, serverAdapter adapter.ServerAdapter, logger *logger.Logger) (*ClientServices, error) {
	keyChainService := crypto.NewKeyChainService()

	cryptoSvc := NewClientCryptoService(keyChainService)
	authSvc := NewClientAuthService(localStore, serverAdapter, keyChainService, cryptoSvc)
	privateSvc := NewClientPrivateDataService(localStore, serverAdapter, cryptoSvc)
	syncSvc := NewClientSyncService(localStore, serverAdapter)

	return &ClientServices{
		CryptoService:      cryptoSvc,
		AuthService:        authSvc,
		PrivateDataService: privateSvc,
		SyncService:        syncSvc,
		SyncJob:            NewClientSyncJob(syncSvc),
	}, nil
}
