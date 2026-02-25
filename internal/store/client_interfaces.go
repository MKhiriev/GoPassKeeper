package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

//go:generate mockgen -source=client_interfaces.go -destination=../mock/client_store_mock.go -package=mock

// LocalPrivateDataRepository is low-level local vault repository.
type LocalPrivateDataRepository interface {
	SavePrivateData(ctx context.Context, userID int64, data ...models.PrivateData) error
	GetPrivateData(ctx context.Context, clientSideID string, userID int64) (models.PrivateData, error)
	GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error)
	GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
	UpdatePrivateData(ctx context.Context, data models.PrivateData) error
	DeletePrivateData(ctx context.Context, clientSideID string, userID int64) error
	IncrementVersion(ctx context.Context, clientSideID string, userID int64) error
}
