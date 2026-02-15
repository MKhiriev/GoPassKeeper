package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

type Storages struct {
	UserRepository     UserRepository
	PrivateDataStorage PrivateDataStorage
}

func NewStorages(cfg config.Storage, logger *logger.Logger) (*Storages, error) {
	logger.Info().Msg("creating new storages...")
	db, err := NewConnectPostgres(context.Background(), cfg.DB, logger)
	if err != nil {
		return nil, fmt.Errorf("postgres connection error: %w", err)
	}

	return &Storages{
		UserRepository:     NewUserRepository(db, logger),
		PrivateDataStorage: NewPrivateDataStorage(db, cfg, logger),
	}, nil
}
