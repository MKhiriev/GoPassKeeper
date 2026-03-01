// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEnv_AllFields(t *testing.T) {
	// Arrange
	envVars := map[string]string{
		"CONFIG": "/path/to/config.json",

		"APP_PASSWORD_HASH_KEY": "hash_secret",
		"APP_TOKEN_SIGN_KEY":    "jwt_secret",
		"APP_TOKEN_ISSUER":      "test_issuer",
		"APP_TOKEN_DURATION":    "1h",
		"APP_HASH_KEY":          "security_hash",

		"SERVER_ADDRESS":         "localhost:8080",
		"SERVER_GRPC_ADDRESS":    "localhost:9090",
		"SERVER_REQUEST_TIMEOUT": "30s",

		// Storage has nested prefixes: STORAGE_ + DB_ / FILES_
		"STORAGE_DB_DATABASE_URI":       "postgres://user:pass@localhost/db",
		"STORAGE_FILES_BINARY_DATA_DIR": "/var/data",
	}
	setEnvVars(t, envVars)

	// Act
	cfg := &StructuredConfig{}
	err := parseEnv(cfg)

	// Assert
	require.NoError(t, err)

	assert.Equal(t, "/path/to/config.json", cfg.JSONFilePath)

	assert.Equal(t, "hash_secret", cfg.App.PasswordHashKey)
	assert.Equal(t, "jwt_secret", cfg.App.TokenSignKey)
	assert.Equal(t, "test_issuer", cfg.App.TokenIssuer)
	assert.Equal(t, time.Hour, cfg.App.TokenDuration)

	assert.Equal(t, "localhost:8080", cfg.Server.HTTPAddress)
	assert.Equal(t, "localhost:9090", cfg.Server.GRPCAddress)
	assert.Equal(t, 30*time.Second, cfg.Server.RequestTimeout)

	assert.Equal(t, "security_hash", cfg.App.HashKey)

	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.Storage.DB.DSN)
	assert.Equal(t, "/var/data", cfg.Storage.Files.BinaryDataDir)
}

func TestParseEnv_PartialFields(t *testing.T) {
	// Arrange
	envVars := map[string]string{
		"APP_TOKEN_SIGN_KEY": "jwt_secret",
		"SERVER_ADDRESS":     "localhost:8080",
	}
	setEnvVars(t, envVars)

	// Act
	cfg := &StructuredConfig{}
	err := parseEnv(cfg)

	// Assert
	require.NoError(t, err)

	// App partially filled
	assert.Empty(t, cfg.App.PasswordHashKey)
	assert.Equal(t, "jwt_secret", cfg.App.TokenSignKey)
	assert.Empty(t, cfg.App.TokenIssuer)
	assert.Zero(t, cfg.App.TokenDuration)

	// Server partially filled
	assert.Equal(t, "localhost:8080", cfg.Server.HTTPAddress)
	assert.Empty(t, cfg.Server.GRPCAddress)
	assert.Zero(t, cfg.Server.RequestTimeout)

	// Others untouched
	assert.Empty(t, cfg.App.HashKey)
	assert.Empty(t, cfg.Storage.DB.DSN)
	assert.Empty(t, cfg.Storage.Files.BinaryDataDir)
	assert.Empty(t, cfg.JSONFilePath)
}

func TestParseEnv_EmptyEnv(t *testing.T) {
	// Arrange
	clearEnvVars(t)

	// Act
	cfg := &StructuredConfig{}
	err := parseEnv(cfg)

	// Assert
	require.NoError(t, err)

	// In this version all nested fields are non-pointer values,
	// so "empty" state is represented by zero values.
	assert.Equal(t, "", cfg.JSONFilePath)

	assert.Equal(t, App{}, cfg.App)
	assert.Equal(t, Server{}, cfg.Server)
	assert.Equal(t, Storage{}, cfg.Storage)
}

func TestParseEnv_OnlyStorageDB(t *testing.T) {
	// Arrange
	envVars := map[string]string{
		"STORAGE_DB_DATABASE_URI": "postgres://localhost/testdb",
	}
	setEnvVars(t, envVars)

	// Act
	cfg := &StructuredConfig{}
	err := parseEnv(cfg)

	// Assert
	require.NoError(t, err)

	assert.Equal(t, "postgres://localhost/testdb", cfg.Storage.DB.DSN)
	assert.Empty(t, cfg.Storage.Files.BinaryDataDir)
}

func TestParseEnv_OnlyStorageFiles(t *testing.T) {
	// Arrange
	envVars := map[string]string{
		"STORAGE_FILES_BINARY_DATA_DIR": "/tmp/files",
	}
	setEnvVars(t, envVars)

	// Act
	cfg := &StructuredConfig{}
	err := parseEnv(cfg)

	// Assert
	require.NoError(t, err)

	assert.Empty(t, cfg.Storage.DB.DSN)
	assert.Equal(t, "/tmp/files", cfg.Storage.Files.BinaryDataDir)
}

func TestParseEnv_InvalidDuration(t *testing.T) {
	// Arrange
	envVars := map[string]string{
		"APP_TOKEN_DURATION": "invalid_duration",
	}
	setEnvVars(t, envVars)

	// Act
	cfg := &StructuredConfig{}
	err := parseEnv(cfg)

	// Assert
	require.Error(t, err)
	// Error wording may vary depending on parseEnv internals; assert loosely.
	assert.Contains(t, err.Error(), "env")
}

func TestParseEnv_DurationFormats(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{"hours", "2h", 2 * time.Hour},
		{"minutes", "45m", 45 * time.Minute},
		{"seconds", "30s", 30 * time.Second},
		{"combined", "1h30m", 90 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			envVars := map[string]string{
				"SERVER_REQUEST_TIMEOUT": tt.envValue,
			}
			setEnvVars(t, envVars)

			// Act
			cfg := &StructuredConfig{}
			err := parseEnv(cfg)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg.Server.RequestTimeout)
		})
	}
}

// Helpers

func setEnvVars(t *testing.T, vars map[string]string) {
	t.Helper()
	clearEnvVars(t)
	for k, v := range vars {
		require.NoError(t, os.Setenv(k, v))
		t.Cleanup(func() { _ = os.Unsetenv(k) })
	}
}

func clearEnvVars(t *testing.T) {
	t.Helper()
	keys := []string{
		"CONFIG",

		"APP_PASSWORD_HASH_KEY",
		"APP_TOKEN_SIGN_KEY",
		"APP_TOKEN_ISSUER",
		"APP_TOKEN_DURATION",

		"SERVER_ADDRESS",
		"SERVER_GRPC_ADDRESS",
		"SERVER_REQUEST_TIMEOUT",

		"SECURITY_HASH_KEY",

		"STORAGE_DB_DATABASE_URI",
		"STORAGE_FILES_BINARY_DATA_DIR",
	}
	for _, k := range keys {
		_ = os.Unsetenv(k)
	}
}
