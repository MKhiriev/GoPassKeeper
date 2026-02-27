package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

// ClientStorages groups all client-side storage repositories into a single
// value that can be passed around the service layer. Currently it holds only
// [LocalPrivateDataRepository]; additional repositories can be added here as
// the feature set grows.
type ClientStorages struct {
	// PrivateDataRepository is the SQLite-backed repository for encrypted
	// vault items stored locally on the client device.
	PrivateDataRepository LocalPrivateDataRepository
}

// NewClientStorages initialises the client storage layer using the supplied
// configuration and logger. It performs the following steps:
//  1. Opens an SQLite connection to the file path specified in cfg.DB.DSN,
//     creating the database file if it does not yet exist.
//  2. Runs pending schema migrations via [DB.Migrate].
//  3. Constructs and returns a [ClientStorages] value wired to a fresh
//     [LocalPrivateDataRepository].
//
// Returns an error if the database connection cannot be established or if
// migration fails.
func NewClientStorages(cfg config.ClientStorage, logger *logger.Logger) (*ClientStorages, error) {
	logger.Info().Msg("creating new storages...")

	db, err := NewConnectSQLite(context.Background(), cfg.DB, logger)
	if err != nil {
		return nil, fmt.Errorf("postgres connection error: %w", err)
	}

	if err := db.Migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &ClientStorages{
		PrivateDataRepository: NewLocalPrivateDataRepository(db, logger),
	}, nil
}
