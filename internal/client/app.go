package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/internal/tui"
)

type App struct {
	services *service.ClientServices
	tui      *tui.TUI
}

func NewApp() (*App, error) {
	serverURL := getenv("GOPASS_SERVER_URL", "http://localhost:8080")
	dbPath := getenv("GOPASS_CLIENT_DB", ":memory:")
	hashKey := os.Getenv("GOPASS_HASH_KEY")

	localStore, err := store.NewLocalStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("create local storage: %w", err)
	}

	serverAdapter := adapter.NewHTTPServerAdapter(adapter.HTTPClientConfig{
		BaseURL: serverURL,
		HashKey: hashKey,
	})

	svcs := service.NewClientServices(localStore, serverAdapter)
	ui := tui.New(svcs.AuthService, svcs.PrivateDataService, svcs.SyncService)

	return &App{services: svcs, tui: ui}, nil
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
		if !errors.Is(err, store.ErrLocalSessionNotFound) {
			return fmt.Errorf("restore session: %w", err)
		}
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
