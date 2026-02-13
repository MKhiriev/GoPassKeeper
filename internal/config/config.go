package config

import (
	"time"
)

type StructuredConfig struct {
	Auth         Auth    `json:"auth"`
	Storage      Storage `json:"storage"`
	Server       Server  `json:"server"`
	Adapter      Adapter `json:"adapter"`
	Workers      Workers `json:"workers"`
	JSONFilePath string  `json:"json_file_path"`
}

type Storage struct {
	DB    *DB    `json:"db"`
	Files *Files `json:"files"`
}

type Auth struct {
	PasswordHashKey string        `json:"password_hash_key"`
	TokenSignKey    string        `json:"token_sign_key"`
	TokenIssuer     string        `json:"token_issuer"`
	TokenDuration   time.Duration `json:"token_duration"`
}

type Server struct {
	HTTPAddress    string        `json:"http_address"`
	GRPCAddress    string        `json:"grpc_address"`
	RequestTimeout time.Duration `json:"request_timeout"`
}

type DB struct {
	DSN string `json:"dsn"`
}

type Files struct {
	BinaryDataDir string `json:"binary_data_dir"` // todo for now we ignore this
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
