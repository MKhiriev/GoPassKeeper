package service

import (
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

type Services struct {
	AppInfoService     AppInfoService
	AuthService        AuthService
	PrivateDataService PrivateDataService
}

func NewServices(storages *store.Storages, cfg config.App, logger *logger.Logger) (*Services, error) {
	logger.Info().Msg("creating new services...")

	appService, err := NewAppInfoService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("error creating app info service: %w", err)
	}

	return &Services{
		AppInfoService:     appService,
		AuthService:        NewAuthService(storages.UserRepository, cfg, logger),
		PrivateDataService: NewPrivateDataService(storages.PrivateDataStorage, cfg, logger),
	}, nil
}
