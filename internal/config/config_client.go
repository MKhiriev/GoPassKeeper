package config

import (
	"fmt"
	"time"
)

// ClientApp holds client-side application settings derived from the shared
// structured config.
type ClientApp struct {
	// HashKey is the HMAC key used by the client for payload integrity checks.
	HashKey string
}

// ClientAdapter holds network settings used by the client transport layer.
type ClientAdapter struct {
	// HTTPAddress is the HTTP endpoint address used by the client.
	HTTPAddress string
	// GRPCAddress is the gRPC endpoint address used by the client.
	GRPCAddress string
	// RequestTimeout is the default timeout for outbound client requests.
	RequestTimeout time.Duration
}

// ClientDB contains local database connection settings for the client.
type ClientDB struct {
	// DSN is the SQLite/PostgreSQL connection string used by the client.
	DSN string
}

// ClientStorage groups client storage backend settings.
type ClientStorage struct {
	// DB holds local database settings.
	DB ClientDB
}

// ClientWorkers contains client background worker settings.
type ClientWorkers struct {
	// SyncInterval defines how often client sync workers should run.
	SyncInterval time.Duration
}

// ClientConfig is the top-level client configuration assembled from
// [StructuredConfig].
type ClientConfig struct {
	// App contains application-level client settings.
	App ClientApp
	// Adapter contains client transport addresses and timeouts.
	Adapter ClientAdapter
	// Storage contains client storage settings.
	Storage ClientStorage
	// Workers contains background job settings.
	Workers ClientWorkers
}

// GetClientConfig builds and validates a client-specific config view from the
// merged structured configuration.
//
// It loads the base config via [GetStructuredConfig], maps only the fields
// relevant to the client runtime, and validates the resulting [ClientConfig].
func GetClientConfig() (*ClientConfig, error) {
	cfg, err := GetStructuredConfig()
	if err != nil {
		return nil, fmt.Errorf("error get structured config: %w", err)
	}

	clientCfg := &ClientConfig{
		App: ClientApp{
			HashKey: cfg.App.HashKey,
		},
		Adapter: ClientAdapter{
			HTTPAddress:    cfg.Adapter.HTTPAddress,
			GRPCAddress:    cfg.Adapter.GRPCAddress,
			RequestTimeout: cfg.Adapter.RequestTimeout,
		},
		Storage: ClientStorage{
			DB: ClientDB{
				DSN: cfg.Storage.DB.DSN,
			},
		},
		Workers: ClientWorkers{SyncInterval: cfg.Workers.SyncInterval},
	}

	return clientCfg, clientCfg.validate()
}
