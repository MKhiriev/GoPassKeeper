package store

import "github.com/MKhiriev/go-pass-keeper/internal/config"

type Storages struct {
	UserRepository     UserRepository
	PrivateDataStorage PrivateDataStorage
}

func NewStorages(cfg config.Storage) (*Storages, error) {
	return &Storages{}, nil
}
