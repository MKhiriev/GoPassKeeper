package service

import (
	"context"
	"time"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type ClientCryptoService interface {
	DeriveKey(masterPassword string, userID int64) []byte
	EncryptPayload(plain models.PrivateDataPayload, key []byte) (models.PrivateDataPayload, error)
	DecryptPayload(cipher models.PrivateDataPayload, key []byte) (models.PrivateDataPayload, error)
	ComputeHash(payload models.PrivateDataPayload) string
}

type ClientAuthService interface {
	Register(ctx context.Context, user models.User) error
	Login(ctx context.Context, user models.User) (encryptionKey []byte, err error)
	RestoreSession(ctx context.Context) (userID int64, token string, err error)
	Logout(ctx context.Context) error
}

type ClientPrivateDataService interface {
	SetEncryptionKey(key []byte)

	Create(ctx context.Context, userID int64, plain models.PrivateDataPayload) error
	GetAll(ctx context.Context, userID int64) ([]models.PrivateData, error)
	Get(ctx context.Context, clientSideID string) (models.PrivateData, error)
	Update(ctx context.Context, data models.PrivateData) error
	Delete(ctx context.Context, clientSideID string, version int64) error
}

type ClientSyncService interface {
	FullSync(ctx context.Context, userID int64) error
	ExecutePlan(ctx context.Context, plan models.SyncPlan) error
}

type ClientSyncJob interface {
	Start(ctx context.Context, userID int64, interval time.Duration)
	Stop()
}
