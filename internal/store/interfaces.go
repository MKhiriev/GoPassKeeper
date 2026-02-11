package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type PrivateDataStorage interface {
	Save(ctx context.Context, data models.PrivateData) error
	SaveAll(ctx context.Context, data []models.PrivateData) error
	Get(ctx context.Context, data models.PrivateData) (models.PrivateData, error)
	GetAll(ctx context.Context) ([]models.PrivateData, error)
}

type PrivateDataRepository interface {
	SavePrivateData(ctx context.Context, data models.PrivateData) error
	SaveAllPrivateData(ctx context.Context, data []models.PrivateData) error
	GetPrivateData(ctx context.Context, data models.PrivateData) (models.PrivateData, error)
	GetAllPrivateData(ctx context.Context) ([]models.PrivateData, error)
}

type PrivateDataFileStorage interface {
	SaveBinaryDataToFile(ctx context.Context, data models.PrivateData) error
	LoadBinaryDataFromFile(ctx context.Context, data models.PrivateData) (models.PrivateData, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	FindUserByLogin(ctx context.Context, user models.User) (models.User, error)
}

// ErrorClassificator defines a strategy for categorizing errors produced by persistence layers.
type ErrorClassificator interface {
	Classify(err error) ErrorClassification // maps an error into a predefined ErrorClassification enum
}
