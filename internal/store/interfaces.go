package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)
	FindUserByLogin(ctx context.Context, user models.User) (models.User, error)
}

// ErrorClassificator defines a strategy for categorizing errors produced by persistence layers.
type ErrorClassificator interface {
	Classify(err error) ErrorClassification // maps an error into a predefined ErrorClassification enum
}
