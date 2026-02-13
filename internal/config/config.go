package config

import (
	"time"
)

type StructuredConfig struct {
	Auth         Auth     `json:"auth"`
	Storage      Storage  `json:"storage"`
	Server       Server   `json:"server"`
	Security     Security `json:"security"`
	Adapter      Adapter  `json:"adapter"`
	Workers      Workers  `json:"workers"`
	JSONFilePath string   `env:"CONFIG" json:"json_file_path"`
}

type Storage struct {
	DB    *DB    `json:"db"`
	Files *Files `json:"files"`
}

type Auth struct {
	PasswordHashKey string        `env:"PASSWORD_HASH_KEY" json:"password_hash_key"`
	TokenSignKey    string        `env:"TOKEN_SIGN_KEY" json:"token_sign_key"`
	TokenIssuer     string        `env:"TOKEN_ISSUER" json:"token_issuer"`
	TokenDuration   time.Duration `env:"TOKEN_DURATION" json:"token_duration"`
}

type Server struct {
	HTTPAddress    string        `env:"ADDRESS" json:"http_address"`
	GRPCAddress    string        `env:"GRPC_ADDRESS" json:"grpc_address"`
	RequestTimeout time.Duration `env:"REQUEST_TIMEOUT" json:"request_timeout"`
}

type Security struct {
	HashKey string `env:"HASH_KEY" json:"hash_key"`
}

type DB struct {
	DSN string `env:"DATABASE_URI" json:"dsn"`
}

type Files struct {
	BinaryDataDir string `env:"BINARY_DATA_DIR" json:"binary_data_dir"`
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
