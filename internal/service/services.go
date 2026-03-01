// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

// Package service defines the core business logic interfaces and service
// implementations for the go-pass-keeper application.
package service

import (
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
)

// Services is the top-level container that groups all application service
// implementations. It is constructed once at startup and injected into the
// HTTP handler layer.
type Services struct {
	// AppInfoService exposes application metadata such as the current version.
	AppInfoService AppInfoService

	// AuthService handles user registration, login, and JWT token lifecycle.
	AuthService AuthService

	// PrivateDataService manages encrypted vault items on behalf of
	// authenticated users, including upload, download, sync, update, and delete
	// operations. The service is pre-wrapped with validation middleware.
	PrivateDataService PrivateDataService
}

// NewServices constructs and wires all application services from the provided
// storage layer, configuration, and logger.
//
// Initialization order:
//  1. AppInfoService — validated first; returns an error immediately if
//     cfg.Version is empty (fail-fast at startup).
//  2. HMAC hasher pool — initialised with cfg.HashKey so that AuthService can
//     hash passwords without allocating a new hasher on every request.
//  3. AuthService and PrivateDataService — constructed after the hasher pool
//     is ready.
//
// Returns a fully initialised *Services or an error if any service fails to
// initialise.
func NewServices(storages *store.Storages, cfg config.App, logger *logger.Logger) (*Services, error) {
	logger.Info().Msg("creating new services...")

	appService, err := NewAppInfoService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("error creating app info service: %w", err)
	}

	utils.InitHasherPool(cfg.HashKey)

	return &Services{
		AppInfoService:     appService,
		AuthService:        NewAuthService(storages.UserRepository, cfg, logger),
		PrivateDataService: NewPrivateDataService(storages.PrivateDataStorage, cfg, logger),
	}, nil
}
