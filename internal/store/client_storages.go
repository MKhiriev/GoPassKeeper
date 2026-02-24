package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

type ClientStorages struct {
	PrivateDataRepository LocalPrivateDataRepository
}

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
