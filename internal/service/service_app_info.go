package service

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

type appInfoService struct {
	appVersion string

	logger *logger.Logger
}

func NewAppInfoService(cfg config.App, logger *logger.Logger) (AppInfoService, error) {
	if cfg.Version == "" {
		return nil, ErrVersionIsNotSpecified
	}

	return &appInfoService{
		appVersion: cfg.Version,
		logger:     logger,
	}, nil
}

func (s *appInfoService) GetAppVersion(ctx context.Context) string {
	return s.appVersion
}
