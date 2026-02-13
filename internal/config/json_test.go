package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJSON_Success(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")

	// Durations in JSON must be valid for time.Duration's TextUnmarshal (string, e.g. "30s").
	jsonBody := `{
		"auth": {
			"password_hash_key": "hash_secret",
			"token_sign_key": "jwt_secret",
			"token_issuer": "test_issuer",
			"token_duration": "1h"
		},
		"server": {
			"http_address": "localhost:8080",
			"grpc_address": "localhost:9090",
			"request_timeout": "30s"
		},
		"security": {
			"hash_key": "security_hash"
		},
		"storage": {
			"db": { "dsn": "postgres://user:pass@localhost/db" },
			"files": { "binary_data_dir": "/var/data" }
		}
	}`

	require.NoError(t, os.WriteFile(p, []byte(jsonBody), 0o600))

	// Act
	cfg, err := parseJSON(p)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "hash_secret", cfg.Auth.PasswordHashKey)
	assert.Equal(t, "jwt_secret", cfg.Auth.TokenSignKey)
	assert.Equal(t, "test_issuer", cfg.Auth.TokenIssuer)
	assert.Equal(t, time.Hour, cfg.Auth.TokenDuration)

	assert.Equal(t, "localhost:8080", cfg.Server.HTTPAddress)
	assert.Equal(t, "localhost:9090", cfg.Server.GRPCAddress)
	assert.Equal(t, 30*time.Second, cfg.Server.RequestTimeout)

	assert.Equal(t, "security_hash", cfg.Security.HashKey)

	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.Storage.DB.DSN)
	assert.Equal(t, "/var/data", cfg.Storage.Files.BinaryDataDir)
}

func TestParseJSON_FileNotFound(t *testing.T) {
	// Act
	cfg, err := parseJSON("definitely-does-not-exist.json")

	// Assert
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "error reading a json file")
}

func TestParseJSON_InvalidJSON(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	require.NoError(t, os.WriteFile(p, []byte(`{ this is not json }`), 0o600))

	// Act
	cfg, err := parseJSON(p)

	// Assert
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "error decoding json configs")
}

func TestParseJSON_InvalidDuration(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	p := filepath.Join(dir, "bad_duration.json")

	// token_duration should be a duration string; make it invalid.
	jsonBody := `{
		"auth": { "token_duration": "not-a-duration" }
	}`
	require.NoError(t, os.WriteFile(p, []byte(jsonBody), 0o600))

	// Act
	cfg, err := parseJSON(p)

	// Assert
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "error decoding json configs")
}

func TestParseJSON_EmptyObject(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	p := filepath.Join(dir, "empty.json")
	require.NoError(t, os.WriteFile(p, []byte(`{}`), 0o600))

	// Act
	cfg, err := parseJSON(p)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// With non-pointer nested structs, all fields are zero values.
	assert.Equal(t, StructuredConfig{}, *cfg)
}

func TestParseJSON_PartialObject(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	p := filepath.Join(dir, "partial.json")

	jsonBody := `{
		"server": { "http_address": "127.0.0.1:8000" }
	}`
	require.NoError(t, os.WriteFile(p, []byte(jsonBody), 0o600))

	// Act
	cfg, err := parseJSON(p)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "127.0.0.1:8000", cfg.Server.HTTPAddress)
	assert.Empty(t, cfg.Server.GRPCAddress)
	assert.Zero(t, cfg.Server.RequestTimeout)

	// Others remain zero
	assert.Equal(t, Auth{}, cfg.Auth)
	assert.Equal(t, Security{}, cfg.Security)
	assert.Equal(t, Storage{}, cfg.Storage)
}
