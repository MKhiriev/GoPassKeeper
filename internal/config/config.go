package config

import "time"

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

type StructuredConfig struct {
	Auth    Auth
	DB      DBConfig
	Server  Server
	Adapter Adapter
	Workers Workers
}

func GetStructuredConfig() (*StructuredConfig, error) {
	return &StructuredConfig{
		Auth:    Auth{},
		Server:  Server{},
		DB:      DBConfig{},
		Adapter: Adapter{},
		Workers: Workers{},
	}, nil
}
