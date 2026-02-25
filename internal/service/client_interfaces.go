package service

import (
	"context"
	"time"

	"github.com/MKhiriev/go-pass-keeper/models"
)

// TODO проверить и переопределить при необходимости
type ClientCryptoService interface {
	SetEncryptionKey(key []byte)

	EncryptPayload(plain models.DecipheredPayload) (models.PrivateDataPayload, error)
	DecryptPayload(cipher models.PrivateDataPayload) (models.DecipheredPayload, error)
	ComputeHash(payload any) (string, error)
}

type ClientAuthService interface {
	Register(ctx context.Context, user models.User) error
	Login(ctx context.Context, user models.User) (encryptionKey []byte, err error)
}

type ClientPrivateDataService interface {
	Create(ctx context.Context, userID int64, plain models.DecipheredPayload) error
	GetAll(ctx context.Context, userID int64) ([]models.DecipheredPayload, error)
	Get(ctx context.Context, clientSideID string, userID int64) (models.DecipheredPayload, error)
	Update(ctx context.Context, data models.DecipheredPayload) error
	Delete(ctx context.Context, clientSideID string, userID int64) error
}

type ClientSyncService interface {
	FullSync(ctx context.Context, userID int64) error
	ExecutePlan(ctx context.Context, plan models.SyncPlan, userID int64) error
}

type ClientSyncJob interface {
	Start(ctx context.Context, userID int64, interval time.Duration)
	Stop()
}
