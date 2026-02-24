package service

import (
	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

type ClientServices struct {
	CryptoService      ClientCryptoService
	AuthService        ClientAuthService
	PrivateDataService ClientPrivateDataService
	SyncService        ClientSyncService
	SyncJob            ClientSyncJob
}

func NewClientServices(localStore store.LocalStorage, serverAdapter adapter.ServerAdapter) *ClientServices {
	cryptoSvc := NewClientCryptoService()
	authSvc := NewClientAuthService(localStore, serverAdapter, cryptoSvc)
	privateSvc := NewClientPrivateDataService(localStore, serverAdapter, cryptoSvc)
	syncSvc := NewClientSyncService(localStore, serverAdapter)

	return &ClientServices{
		CryptoService:      cryptoSvc,
		AuthService:        authSvc,
		PrivateDataService: privateSvc,
		SyncService:        syncSvc,
		SyncJob:            NewClientSyncJob(syncSvc),
	}
}
