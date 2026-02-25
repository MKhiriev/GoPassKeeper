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
)

type App struct {
	services    *service.ClientServices
	tui         *tui.TUI
	syncJobTime time.Duration
}

func NewApp(services *service.ClientServices, ui *tui.TUI, cfg config.ClientWorkers, logger *logger.Logger) (*App, error) {

	return &App{services: services, tui: ui, syncJobTime: cfg.SyncInterval}, nil
}

func (a *App) Run() error {
	ctx := context.Background()

	userID, key, err := a.tui.LoginFlow(ctx)
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

	logout, err := a.tui.MainLoop(ctx, userID)
	if logout {
		return a.Run()
	}

	return err
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
