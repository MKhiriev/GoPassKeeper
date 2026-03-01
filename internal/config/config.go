// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package config

import (
	"time"
)

// StructuredConfig is the top-level configuration container for the
// go-pass-keeper application. It aggregates all sub-configurations and is
// populated by merging values from environment variables, command-line flags,
// and an optional JSON file.
//
// Struct tags:
//   - envPrefix — prefix applied to all nested env tag lookups (caarlos0/env).
//   - env       — direct environment variable name for scalar fields.
type StructuredConfig struct {
	// App holds application-level settings such as cryptographic keys,
	// token parameters, and the application version.
	App App `envPrefix:"APP_"`

	// Storage holds configuration for all persistence backends, including
	// the relational database and the binary file store.
	Storage Storage `envPrefix:"STORAGE_"`

	// Server holds network address and timeout settings for the HTTP and
	// gRPC servers.
	Server Server `envPrefix:"SERVER_"`

	// Adapter holds configuration for external adapter integrations.
	// Currently empty; reserved for future use.
	Adapter Adapter `envPrefix:"ADAPTER_"`

	// Workers holds configuration for background worker processes.
	// Currently empty; reserved for future use.
	Workers Workers `envPrefix:"WORKERS_"`

	// JSONFilePath is the optional path to a JSON configuration file.
	// When non-empty, the file is parsed and merged on top of the values
	// already loaded from environment variables and flags.
	// Populated via the CONFIG environment variable or the -c / -config flag.
	JSONFilePath string `env:"CONFIG"`
}

// Storage groups the configuration for all storage backends used by the
// application.
type Storage struct {
	// DB holds the relational database connection settings.
	DB DB `envPrefix:"DB_"`

	// Files holds the file-system storage settings for binary vault data.
	Files Files `envPrefix:"FILES_"`
}

// App holds application-level configuration values that control security,
// token lifecycle, and versioning.
type App struct {
	// PasswordHashKey is the secret key used when hashing user passwords
	// with HMAC-SHA256. Must be kept confidential.
	// Env: APP_PASSWORD_HASH_KEY
	PasswordHashKey string `env:"PASSWORD_HASH_KEY"`

	// TokenSignKey is the secret key used to sign and verify JWT tokens.
	// Must be kept confidential.
	// Env: APP_TOKEN_SIGN_KEY
	TokenSignKey string `env:"TOKEN_SIGN_KEY"`

	// TokenIssuer is the "iss" claim embedded in every issued JWT token.
	// It identifies the service that issued the token and is validated on
	// every authenticated request.
	// Env: APP_TOKEN_ISSUER
	TokenIssuer string `env:"TOKEN_ISSUER"`

	// TokenDuration specifies how long a JWT token remains valid after
	// issuance (e.g. "1h", "30m").
	// Env: APP_TOKEN_DURATION
	TokenDuration time.Duration `env:"TOKEN_DURATION"`

	// HashKey is the HMAC key used for request integrity checking
	// (e.g. the HashSHA256 header). Distinct from PasswordHashKey.
	// Env: APP_HASH_KEY
	HashKey string `env:"HASH_KEY"`

	// Version is the semantic version string of the running application
	// (e.g. "1.2.3"). Exposed via the /api/version/ endpoint.
	// Env: APP_VERSION
	Version string `env:"VERSION"`
}

// Server holds network and timeout settings for the inbound transport layer.
type Server struct {
	// HTTPAddress is the TCP address on which the HTTP server listens,
	// in "host:port" format (e.g. "0.0.0.0:8080").
	// Env: SERVER_ADDRESS
	HTTPAddress string `env:"ADDRESS"`

	// GRPCAddress is the TCP address on which the gRPC server listens,
	// in "host:port" format (e.g. "0.0.0.0:9090").
	// Env: SERVER_GRPC_ADDRESS
	GRPCAddress string `env:"GRPC_ADDRESS"`

	// RequestTimeout is the maximum duration allowed for a single inbound
	// request before the server cancels it (e.g. "30s", "1m").
	// Env: SERVER_REQUEST_TIMEOUT
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT"`
}

// DB holds connection settings for the relational database backend.
type DB struct {
	// DSN is the PostgreSQL Data Source Name (connection string) used to
	// open the database connection
	// (e.g. "postgres://user:pass@localhost:5432/dbname?sslmode=disable").
	// Env: STORAGE_DB_DATABASE_URI
	DSN string `env:"DATABASE_URI"`
}

// Files holds file-system settings for the binary vault data store.
type Files struct {
	// BinaryDataDir is the absolute or relative path to the directory where
	// encrypted binary vault files are stored and served from.
	// Env: STORAGE_FILES_BINARY_DATA_DIR
	BinaryDataDir string `env:"BINARY_DATA_DIR"`
}

// Adapter holds configuration for external adapter integrations.
// The struct is currently empty and is reserved for future third-party
// service configuration (e.g. object storage, message brokers).
type Adapter struct {
	// HTTPAddress is the TCP address on which the HTTP server listens,
	// in "host:port" format (e.g. "0.0.0.0:8080").
	// Env: ADAPTER__ADDRESS
	HTTPAddress string `env:"ADDRESS"`

	// GRPCAddress is the TCP address on which the gRPC server listens,
	// in "host:port" format (e.g. "0.0.0.0:9090").
	// Env: ADAPTER__GRPC_ADDRESS
	GRPCAddress string `env:"GRPC_ADDRESS"`

	// RequestTimeout is the maximum duration allowed for a single inbound
	// request before the server cancels it (e.g. "30s", "1m").
	// Env: ADAPTER_REQUEST_TIMEOUT
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT"`
}

// Workers holds configuration for background worker processes.
// The struct is currently empty and is reserved for future worker
// configuration (e.g. concurrency limits, queue sizes).
type Workers struct {
	SyncInterval time.Duration `env:"SYNC_INTERVAL"`
}

// GetStructuredConfig loads, merges, and validates the application
// configuration from all available sources in the following priority order
// (last source wins for non-zero fields):
//  1. Environment variables
//  2. Command-line flags
//  3. JSON file (path resolved from sources 1 and 2)
//
// Returns a fully populated *StructuredConfig or an error if any source
// fails to load or the final config fails validation.
func GetStructuredConfig() (*StructuredConfig, error) {
	return newConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		build()
}
