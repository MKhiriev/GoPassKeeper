// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/tui"
	"github.com/MKhiriev/go-pass-keeper/models"
)

// App is the concrete interactive client runtime.
//
// It coordinates authentication, encryption-key setup, initial synchronization,
// periodic background sync jobs, and the main terminal UI loop.
type App struct {
	services    *service.ClientServices
	tui         *tui.TUI
	syncJobTime time.Duration
	buildInfo   models.AppBuildInfo
}

// NewApp constructs an [App] using the provided services, terminal UI, worker
// configuration, and build metadata.
//
// The logger parameter is accepted for API consistency with other constructors
// in the project, but is not currently used directly by this type.
func NewApp(services *service.ClientServices, ui *tui.TUI, cfg config.ClientWorkers, buildInfo models.AppBuildInfo, logger *logger.Logger) (*App, error) {

	return &App{
		services:    services,
		tui:         ui,
		syncJobTime: cfg.SyncInterval,
		buildInfo:   buildInfo,
	}, nil
}

// Run executes the full client lifecycle.
//
// Flow:
//  1. Run login flow and obtain authenticated user ID and encryption key.
//  2. Configure encryption key in private-data service.
//  3. Perform an initial full sync (non-fatal warning on failure).
//  4. Start periodic background sync job.
//  5. Run the main TUI loop.
//  6. On logout request, restart the lifecycle from login.
func (a *App) Run() error {
	ctx := context.Background()

	userID, key, err := a.tui.LoginFlow(ctx, a.buildInfo)
	if err != nil {
		if errors.Is(err, tui.ErrUserQuit) {
			return nil
		}
		return err
	}

	a.services.PrivateDataService.SetEncryptionKey(key)

	if err = a.services.SyncService.FullSync(ctx, userID); err != nil {
		fmt.Fprintf(os.Stderr, "sync warning: %v\n", err)
	}

	a.services.SyncJob.Start(ctx, userID, a.syncJobTime)
	defer a.services.SyncJob.Stop()

	logout, err := a.tui.MainLoop(ctx, userID, a.buildInfo)
	if logout {
		return a.Run()
	}

	return err
}
