package config

import (
	"time"
)

// TODO find the required configs
// TODO add env and json tags

type StructuredConfig struct {
	Auth         Auth
	DB           DBConfig
	Server       Server
	Adapter      Adapter
	Workers      Workers
	jsonFilePath string
}

type DBConfig struct {
	DSN string
}

type Auth struct {
	PasswordHashKey string
	TokenSignKey    string
	TokenIssuer     string
	TokenDuration   time.Duration
}

type Server struct {
	HTTPAddress    string
	GRPCAddress    string
	RequestTimeout time.Duration
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
