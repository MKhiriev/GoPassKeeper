package store

import (
	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

type Storages struct {
	UserRepository     UserRepository
	PrivateDataStorage PrivateDataStorage
}

func NewStorages(cfg config.Storage, logger *logger.Logger) (*Storages, error) {
	logger.Info().Msg("creating new storages...")
	return &Storages{}, nil
}
