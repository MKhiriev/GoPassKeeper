package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type StructuredJSONConfig struct {
	Auth struct {
		PasswordHashKey string   `env:"PASSWORD_HASH_KEY" json:"password_hash_key"`
		TokenSignKey    string   `env:"TOKEN_SIGN_KEY" json:"token_sign_key"`
		TokenIssuer     string   `env:"TOKEN_ISSUER" json:"token_issuer"`
		TokenDuration   Duration `env:"TOKEN_DURATION" json:"token_duration"`
	} `envPrefix:"AUTH_" json:"auth,omitempty"`

	Storage struct {
		DB struct {
			DSN string `env:"DATABASE_URI" json:"dsn"`
		} `envPrefix:"DB_" json:"db,omitempty"`

		Files struct {
			BinaryDataDir string `env:"BINARY_DATA_DIR" json:"binary_data_dir"`
		} `envPrefix:"FILES_" json:"files,omitempty"`
	} `envPrefix:"STORAGE_" json:"storage,omitempty"`

	Server struct {
		HTTPAddress    string   `env:"ADDRESS" json:"http_address"`
		GRPCAddress    string   `env:"GRPC_ADDRESS" json:"grpc_address"`
		RequestTimeout Duration `env:"REQUEST_TIMEOUT" json:"request_timeout"`
	} `envPrefix:"SERVER_" json:"server,omitempty"`

	Security struct {
		HashKey string `env:"HASH_KEY" json:"hash_key"`
	} `envPrefix:"SECURITY_" json:"security,omitempty"`

	Adapter struct{} `envPrefix:"ADAPTER_" json:"adapter,omitempty"`

	Workers struct{} `envPrefix:"WORKERS_" json:"workers,omitempty"`
}

func parseJSON(jsonFilePath string) (*StructuredConfig, error) {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading a json file: %w", err)
	}
	defer jsonFile.Close()

	var jsonCfg StructuredJSONConfig
	if err := json.NewDecoder(jsonFile).Decode(&jsonCfg); err != nil {
		return nil, fmt.Errorf("error decoding json configs: %w", err)
	}

	cfg := &StructuredConfig{
		Auth: Auth{
			PasswordHashKey: jsonCfg.Auth.PasswordHashKey,
			TokenSignKey:    jsonCfg.Auth.TokenSignKey,
			TokenIssuer:     jsonCfg.Auth.TokenIssuer,
			TokenDuration:   time.Duration(jsonCfg.Auth.TokenDuration),
		},
		Storage: Storage{
			DB: DB{
				DSN: jsonCfg.Storage.DB.DSN,
			},
			Files: Files{
				BinaryDataDir: jsonCfg.Storage.Files.BinaryDataDir,
			},
		},
		Server: Server{
			HTTPAddress:    jsonCfg.Server.HTTPAddress,
			GRPCAddress:    jsonCfg.Server.GRPCAddress,
			RequestTimeout: time.Duration(jsonCfg.Server.RequestTimeout),
		},
		Security: Security{
			HashKey: jsonCfg.Security.HashKey,
		},
		Adapter:      Adapter{},
		Workers:      Workers{},
		JSONFilePath: "",
	}

	return cfg, nil
}

// Duration is a wrapper around time.Duration that supports JSON unmarshaling from strings like "1h", "30s"
type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return json.Unmarshal(b, (*time.Duration)(d))
	}
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}
