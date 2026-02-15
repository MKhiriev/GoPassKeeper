package config

import (
	"time"
)

type StructuredConfig struct {
	App          App     `envPrefix:"APP_"`
	Storage      Storage `envPrefix:"STORAGE_"`
	Server       Server  `envPrefix:"SERVER_"`
	Adapter      Adapter `envPrefix:"ADAPTER_"`
	Workers      Workers `envPrefix:"WORKERS_"`
	JSONFilePath string  `env:"CONFIG"`
}

type Storage struct {
	DB    DB    `envPrefix:"DB_"`
	Files Files `envPrefix:"FILES_"`
}

type App struct {
	PasswordHashKey string        `env:"PASSWORD_HASH_KEY"`
	TokenSignKey    string        `env:"TOKEN_SIGN_KEY"`
	TokenIssuer     string        `env:"TOKEN_ISSUER"`
	TokenDuration   time.Duration `env:"TOKEN_DURATION"`

	HashKey string `env:"HASH_KEY"`

	Version string `env:"VERSION"`
}

type Server struct {
	HTTPAddress    string        `env:"ADDRESS"`
	GRPCAddress    string        `env:"GRPC_ADDRESS"`
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT"`
}

type DB struct {
	DSN string `env:"DATABASE_URI"`
}

type Files struct {
	BinaryDataDir string `env:"BINARY_DATA_DIR"`
}

type Adapter struct {
}

type Workers struct {
}

func GetStructuredConfig() (*StructuredConfig, error) {
	return newConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		build()
}
