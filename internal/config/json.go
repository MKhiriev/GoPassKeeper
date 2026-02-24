package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// StructuredJSONConfig is the JSON-specific representation of the application
// configuration. It mirrors [StructuredConfig] but uses JSON struct tags and
// the custom [Duration] type so that duration values can be expressed as
// human-readable strings (e.g. "1h", "30s") in the config file.
//
// After decoding, the values are mapped into a [StructuredConfig] by
// [parseJSON].
type StructuredJSONConfig struct {
	// App holds application-level settings loaded from the JSON file.
	App struct {
		PasswordHashKey string   `json:"password_hash_key"`
		TokenSignKey    string   `json:"token_sign_key"`
		TokenIssuer     string   `json:"token_issuer"`
		TokenDuration   Duration `json:"token_duration"`
		HashKey         string   `json:"hash_key"`
		Version         string   `json:"version"`
	} `json:"app,omitempty"`

	// Storage holds database and file-storage settings loaded from the JSON file.
	Storage struct {
		DB struct {
			DSN string `json:"dsn"`
		} `json:"db,omitempty"`

		Files struct {
			BinaryDataDir string `json:"binary_data_dir"`
		} `json:"files,omitempty"`
	} `json:"storage,omitempty"`

	// Server holds HTTP and gRPC server settings loaded from the JSON file.
	Server struct {
		HTTPAddress    string   `json:"http_address"`
		GRPCAddress    string   `json:"grpc_address"`
		RequestTimeout Duration `json:"request_timeout"`
	} `json:"server,omitempty"`

	// Adapter is reserved for future external adapter configuration.
	Adapter struct {
	} `json:"adapter,omitempty"`

	// Workers is reserved for future background worker configuration.
	Workers struct {
		SyncInterval Duration `json:"sync_interval"`
	} `json:"workers,omitempty"`
}

// parseJSON opens the JSON file at jsonFilePath, decodes it into a
// [StructuredJSONConfig], and maps the result into a [StructuredConfig].
//
// JSONFilePath is intentionally left empty in the returned config so that
// the path is not re-processed during subsequent merge steps.
//
// Returns a wrapped error if the file cannot be opened or its contents
// cannot be decoded as valid JSON.
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
		App: App{
			PasswordHashKey: jsonCfg.App.PasswordHashKey,
			TokenSignKey:    jsonCfg.App.TokenSignKey,
			TokenIssuer:     jsonCfg.App.TokenIssuer,
			TokenDuration:   time.Duration(jsonCfg.App.TokenDuration),
			HashKey:         jsonCfg.App.HashKey,
			Version:         jsonCfg.App.Version,
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
		JSONFilePath: "", // intentionally cleared to prevent re-processing
	}

	return cfg, nil
}

// Duration is a thin wrapper around [time.Duration] that adds JSON
// unmarshaling support for human-readable duration strings such as "1h",
// "30m", or "15s", in addition to raw nanosecond integers.
//
// Use Duration in JSON config structs wherever a time.Duration field is
// needed. Convert back to time.Duration with a simple cast:
//
//	d := Duration(5 * time.Minute)
//	std := time.Duration(d) // → 5m0s
type Duration time.Duration

// UnmarshalJSON implements [json.Unmarshaler] for Duration.
//
// Supported JSON value types:
//   - string: parsed with [time.ParseDuration] (e.g. "1h30m", "30s").
//   - number: treated as a raw nanosecond count (same as time.Duration).
//
// Returns an error if the value is a string that cannot be parsed as a
// duration, or if the JSON value is of an unsupported type.
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

// MarshalJSON implements [json.Marshaler] for Duration.
// The value is serialized as a human-readable string using
// [time.Duration.String] (e.g. "1h0m0s", "30m0s").
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}
