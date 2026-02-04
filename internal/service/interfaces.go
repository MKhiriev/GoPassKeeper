package service

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type AuthService interface {
	RegisterUser(ctx context.Context, user models.User) (models.User, error)
	Login(ctx context.Context, user models.User) (models.User, error)
	CreateToken(ctx context.Context, user models.User) (models.Token, error)
	ParseToken(ctx context.Context, tokenString string) (models.Token, error)
}
