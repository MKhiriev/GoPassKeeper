package service

import (
	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/crypto"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

type ClientServices struct {
	CryptoService      ClientCryptoService
	AuthService        ClientAuthService
	PrivateDataService ClientPrivateDataService
	SyncService        ClientSyncService
	SyncJob            ClientSyncJob
}

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
