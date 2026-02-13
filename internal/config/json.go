package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type StructuredJSONConfig struct {
	Services struct {
		PasswordHashKey string   `json:"password_hash_key"`
		TokenSignKey    string   `json:"token_sign_key"`
		TokenIssuer     string   `json:"token_issuer"`
		TokenDuration   Duration `json:"token_duration"`
		HashKey         string   `json:"hash_key"`
	} `json:"services,omitempty"`

	Storage struct {
		DB struct {
			DSN string `json:"dsn"`
		} `json:"db,omitempty"`

		Files struct {
			BinaryDataDir string `json:"binary_data_dir"`
		} `json:"files,omitempty"`
	} `json:"storage,omitempty"`

	Server struct {
		HTTPAddress    string   `json:"http_address"`
		GRPCAddress    string   `json:"grpc_address"`
		RequestTimeout Duration `json:"request_timeout"`
	} `json:"server,omitempty"`

	Adapter struct{} `json:"adapter,omitempty"`

	Workers struct{} `json:"workers,omitempty"`
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
		Services: Services{
			PasswordHashKey: jsonCfg.Services.PasswordHashKey,
			TokenSignKey:    jsonCfg.Services.TokenSignKey,
			TokenIssuer:     jsonCfg.Services.TokenIssuer,
			TokenDuration:   time.Duration(jsonCfg.Services.TokenDuration),
			HashKey:         jsonCfg.Services.HashKey,
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
