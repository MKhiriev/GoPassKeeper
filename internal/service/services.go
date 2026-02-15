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

func NewServices(storages *store.Storages, cfg config.Services, logger *logger.Logger) (*Services, error) {
	logger.Info().Msg("creating new services...")
	return &Services{
		AuthService:        NewAuthService(storages.UserRepository, cfg, logger),
		PrivateDataService: NewPrivateDataService(storages.PrivateDataStorage, cfg, logger),
	}, nil
}
