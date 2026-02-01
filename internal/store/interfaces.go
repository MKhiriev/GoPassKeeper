package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) error
	FindUserByLogin(ctx context.Context, user models.User) (models.User, error)
}
