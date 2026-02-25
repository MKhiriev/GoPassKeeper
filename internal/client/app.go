package client

import (
	"context"
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

	var (
		userID int64
		key    []byte
		err    error
	)

	userID, _, err = a.services.AuthService.RestoreSession(ctx)
	if err != nil {
		userID, key, err = a.tui.LoginFlow(ctx)
		if err != nil {
			return err
		}
	} else {
		key, err = a.tui.UnlockWithMasterPassword(userID, a.services.CryptoService.DeriveKey)
		if err != nil {
			return err
		}
	}

	a.services.PrivateDataService.SetEncryptionKey(key)

	if err = a.services.SyncService.FullSync(ctx, userID); err != nil {
		fmt.Fprintf(os.Stderr, "sync warning: %v\n", err)
	}

	a.services.SyncJob.Start(ctx, userID, 5*time.Minute)
	defer a.services.SyncJob.Stop()

	logout, err := a.tui.MainLoop(ctx, userID)
	if err != nil {
		return err
	}
	if logout {
		return a.Run()
	}

	return nil
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
