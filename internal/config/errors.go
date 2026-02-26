package config

import "errors"

// Validation errors returned by [ClientConfig.validate] when required
// configuration groups are incomplete or invalid.
var (
	// ErrInvalidAdapterConfigs indicates invalid client adapter settings
	// (for example, missing HTTP address or request timeout).
	ErrInvalidAdapterConfigs = errors.New("invalid adapter configuration")
	// ErrInvalidStorageConfigs indicates invalid client storage settings
	// (for example, empty DSN or unsupported in-memory DSN).
	ErrInvalidStorageConfigs = errors.New("invalid storage configuration")
	// ErrInvalidAppConfigs indicates invalid application-level settings
	// required by the client (for example, missing hash key).
	ErrInvalidAppConfigs = errors.New("invalid app configuration")
	// ErrInvalidWorkerConfigs indicates invalid background worker settings
	// (for example, zero sync interval).
	ErrInvalidWorkerConfigs = errors.New("invalid worker configuration")
)
