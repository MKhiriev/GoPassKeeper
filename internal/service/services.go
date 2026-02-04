package service

import (
	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

type Services struct {
	AuthService AuthService
}

func NewServices(repositories store.Repositories, cfg config.StructuredConfig, logger *logger.Logger) *Services {
	return &Services{
		NewAuthService(repositories.UserRepository, cfg.Auth, logger),
	}
}
