package config

import "errors"

var (
	ErrInvalidAdapterConfigs = errors.New("invalid adapter configuration")
	ErrInvalidStorageConfigs = errors.New("invalid storage configuration")
	ErrInvalidAppConfigs     = errors.New("invalid app configuration")
	ErrInvalidWorkerConfigs  = errors.New("invalid worker configuration")
)
