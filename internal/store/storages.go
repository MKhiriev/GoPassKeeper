package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

// Storages is a top-level container that aggregates all storage abstractions
// required by the application's service layer.
//
// It is constructed once at application startup via [NewStorages] and then
// injected into service constructors. Each field exposes a domain-specific
// interface so that the service layer remains decoupled from concrete
// PostgreSQL implementations.
type Storages struct {
	// UserRepository provides CRUD operations for user accounts.
	// See [UserRepository] for the full method contract.
	UserRepository UserRepository

	// PrivateDataStorage provides high-level operations for vault items,
	// coordinating between relational storage and (in the future) file storage.
	// See [PrivateDataStorage] for the full method contract.
	PrivateDataStorage PrivateDataStorage
}

// NewStorages initialises all storage dependencies and returns a ready-to-use
// [Storages] container.
//
// The function performs the following steps in order:
//  1. Opens and verifies a PostgreSQL connection using [NewConnectPostgres].
//  2. Runs pending database migrations via [DB.Migrate].
//  3. Constructs [UserRepository] and [PrivateDataStorage] backed by the
//     established connection.
//
// If any step fails, a descriptive wrapped error is returned and the caller
// should treat the application as unable to start.
//
// Parameters:
//   - cfg: storage configuration including database DSN and file storage paths.
//   - logger: structured logger used for startup diagnostics and passed down
//     to all repository implementations.
func NewStorages(cfg config.Storage, logger *logger.Logger) (*Storages, error) {
	logger.Info().Msg("creating new storages...")

	db, err := NewConnectPostgres(context.Background(), cfg.DB, logger)
	if err != nil {
		return nil, fmt.Errorf("postgres connection error: %w", err)
	}

	if err := db.Migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &Storages{
		UserRepository:     NewUserRepository(db, logger),
		PrivateDataStorage: NewPrivateDataStorage(db, cfg, logger),
	}, nil
}
