package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type PrivateDataStorage interface {
	Save(ctx context.Context, data ...*models.PrivateData) error

	Get(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error)
	GetAll(ctx context.Context, userID int64) ([]models.PrivateData, error)

	GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
	GetStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error)

	Update(ctx context.Context, updateRequests models.UpdateRequest) error
	Delete(ctx context.Context, deleteRequests models.DeleteRequest) error
}

type PrivateDataRepository interface {
	SavePrivateData(ctx context.Context, data ...*models.PrivateData) error

	GetPrivateData(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error)
	GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error)

	GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
	GetStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error)

	UpdatePrivateData(ctx context.Context, updateRequests models.UpdateRequest) error
	DeletePrivateData(ctx context.Context, deleteRequests models.DeleteRequest) error
}

type PrivateDataFileStorage interface {
	SaveBinaryDataToFile(ctx context.Context, fileName string, data ...models.PrivateData) error
	LoadBinaryDataFromFile(ctx context.Context, fileName string) ([]models.PrivateData, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	FindUserByLogin(ctx context.Context, user models.User) (models.User, error)
}

// ErrorClassificator defines a strategy for categorizing errors produced by persistence layers.
type ErrorClassificator interface {
	Classify(err error) ErrorClassification // maps an error into a predefined ErrorClassification enum
}
