package service

import (
	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

type Services struct {
	AuthService        AuthService
	PrivateDataService PrivateDataService
}

func NewServices(repositories store.Storages, cfg config.StructuredConfig, logger *logger.Logger) *Services {
	return &Services{
		AuthService:        NewAuthService(repositories.UserRepository, cfg.Auth, logger),
		PrivateDataService: NewPrivateDataService(repositories.PrivateDataStorage, cfg.Storage, logger),
	}
}
