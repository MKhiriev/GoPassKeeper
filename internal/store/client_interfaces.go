package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

// LocalStorage is client-side local storage.
type LocalStorage interface {
	Save(ctx context.Context, data ...*models.PrivateData) error
	GetAll(ctx context.Context, userID int64) ([]models.PrivateData, error)
	Get(ctx context.Context, clientSideID string) (models.PrivateData, error)
	GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
	Update(ctx context.Context, data models.PrivateData) error
	SoftDelete(ctx context.Context, clientSideID string, version int64) error

	SaveSession(ctx context.Context, userID int64, token string) error
	LoadSession(ctx context.Context) (userID int64, token string, err error)
	ClearSession(ctx context.Context) error
}

// LocalPrivateDataRepository is low-level local vault repository.
type LocalPrivateDataRepository interface {
	SavePrivateData(ctx context.Context, data ...*models.PrivateData) error
	GetPrivateData(ctx context.Context, clientSideID string) (models.PrivateData, error)
	GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error)
	GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
	UpdatePrivateData(ctx context.Context, data models.PrivateData) error
	DeletePrivateData(ctx context.Context, clientSideID string, version int64) error
}
