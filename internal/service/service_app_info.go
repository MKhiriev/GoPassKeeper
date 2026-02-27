package service

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

// appInfoService is the concrete implementation of AppInfoService.
// It holds the application version string read from configuration at startup
// and exposes it through the AppInfoService interface.
type appInfoService struct {
	// appVersion is the semantic version string of the running application
	// (e.g. "1.2.3" or "dev"), sourced from config.App.Version.
	appVersion string

	// logger is the structured logger used for diagnostic output.
	logger *logger.Logger
}

// NewAppInfoService constructs a new AppInfoService from the provided
// application configuration and logger.
//
// It validates that cfg.Version is non-empty; if the version is missing,
// ErrVersionIsNotSpecified is returned so that the application fails fast
// at startup rather than serving an empty version string at runtime.
//
// Returns the initialised AppInfoService or an error if validation fails.
func NewAppInfoService(cfg config.App, logger *logger.Logger) (AppInfoService, error) {
	if cfg.Version == "" {
		return nil, ErrVersionIsNotSpecified
	}

	return &appInfoService{
		appVersion: cfg.Version,
		logger:     logger,
	}, nil
}

// GetAppVersion returns the semantic version string of the running application.
// The value is set once at construction time and is safe for concurrent use.
func (s *appInfoService) GetAppVersion(ctx context.Context) string {
	return s.appVersion
}
