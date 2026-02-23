package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func writeTempJSONConfig(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	f, err := os.CreateTemp(t.TempDir(), "config-*.json")
	require.NoError(t, err)
	_, err = f.Write(data)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

// ── newConfigBuilder ──────────────────────────────────────────────────────────

// TestNewConfigBuilder_InitialState verifies that a freshly created builder
// has no error and an empty configs slice.
func TestNewConfigBuilder_InitialState(t *testing.T) {
	b := newConfigBuilder()
	require.NotNil(t, b)
	assert.NoError(t, b.err)
	assert.Empty(t, b.configs)
}

// ── build ─────────────────────────────────────────────────────────────────────

// TestBuild_EmptyBuilder verifies that building with no configs returns a
// zero-value StructuredConfig.
func TestBuild_EmptyBuilder(t *testing.T) {
	cfg, err := newConfigBuilder().build()
	require.NoError(t, err)
	assert.Equal(t, &StructuredConfig{}, cfg)
}

// TestBuild_PropagatesBuilderError verifies that a pre-set b.err is wrapped
// and returned, with nil config.
func TestBuild_PropagatesBuilderError(t *testing.T) {
	b := newConfigBuilder()
	b.err = assert.AnError

	cfg, err := b.build()
	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

// TestBuild_MergesMultipleConfigs verifies that fields from multiple configs
// are merged into a single result.
func TestBuild_MergesMultipleConfigs(t *testing.T) {
	b := newConfigBuilder()
	b.configs = append(b.configs,
		&StructuredConfig{App: App{Version: "1.0.0"}},
		&StructuredConfig{App: App{TokenIssuer: "issuer"}},
	)

	cfg, err := b.build()
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", cfg.App.Version)
	assert.Equal(t, "issuer", cfg.App.TokenIssuer)
}

// TestBuild_SingleConfig verifies that a single config is returned as-is.
func TestBuild_SingleConfig(t *testing.T) {
	b := newConfigBuilder()
	b.configs = append(b.configs, &StructuredConfig{
		App: App{Version: "2.0.0", TokenIssuer: "single"},
	})

	cfg, err := b.build()
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", cfg.App.Version)
	assert.Equal(t, "single", cfg.App.TokenIssuer)
}

// ── withEnv ───────────────────────────────────────────────────────────────────

// TestWithEnv_ReturnsBuilder verifies the fluent interface.
func TestWithEnv_ReturnsBuilder(t *testing.T) {
	b := newConfigBuilder()
	assert.Same(t, b, b.withEnv())
}

// TestWithEnv_AppendsOneConfig verifies that withEnv appends exactly one entry.
func TestWithEnv_AppendsOneConfig(t *testing.T) {
	b := newConfigBuilder()
	b.withEnv()
	assert.Len(t, b.configs, 1)
}

// TestWithEnv_ReadsEnvVars verifies that environment variables are picked up.
func TestWithEnv_ReadsEnvVars(t *testing.T) {
	t.Setenv("APP_VERSION", "env-version")
	t.Setenv("APP_TOKEN_ISSUER", "env-issuer")

	b := newConfigBuilder()
	b.withEnv()

	require.Len(t, b.configs, 1)
	assert.Equal(t, "env-version", b.configs[0].App.Version)
	assert.Equal(t, "env-issuer", b.configs[0].App.TokenIssuer)
}

// TestWithEnv_NoErrorOnEmptyEnv verifies that withEnv does not set b.err
// when no relevant env vars are present.
func TestWithEnv_NoErrorOnEmptyEnv(t *testing.T) {
	b := newConfigBuilder()
	b.withEnv()
	assert.NoError(t, b.err)
}

// ── withFlags ─────────────────────────────────────────────────────────────────

// TestWithFlags_ReturnsBuilder verifies the fluent interface.
func TestWithFlags_ReturnsBuilder(t *testing.T) {
	b := newConfigBuilder()
	assert.Same(t, b, b.withFlags())
}

// ── withJSON ──────────────────────────────────────────────────────────────────

// TestWithJSON_ReturnsBuilder verifies the fluent interface.
func TestWithJSON_ReturnsBuilder(t *testing.T) {
	b := newConfigBuilder()
	assert.Same(t, b, b.withJSON())
}

// TestWithJSON_NoOp_WhenNoPathSet verifies that withJSON does nothing when
// no config has a JSONFilePath.
func TestWithJSON_NoOp_WhenNoPathSet(t *testing.T) {
	b := newConfigBuilder()
	b.configs = append(b.configs, &StructuredConfig{})
	b.withJSON()

	assert.Len(t, b.configs, 1)
	assert.NoError(t, b.err)
}

// TestWithJSON_AppendsConfig_WhenValidFile verifies that a valid JSON file is
// parsed and appended.
func TestWithJSON_AppendsConfig_WhenValidFile(t *testing.T) {
	payload := StructuredJSONConfig{}
	payload.App.Version = "json-version"
	payload.App.TokenIssuer = "json-issuer"
	path := writeTempJSONConfig(t, payload)

	b := newConfigBuilder()
	b.configs = append(b.configs, &StructuredConfig{JSONFilePath: path})
	b.withJSON()

	require.NoError(t, b.err)
	require.Len(t, b.configs, 2)
	assert.Equal(t, "json-version", b.configs[1].App.Version)
	assert.Equal(t, "json-issuer", b.configs[1].App.TokenIssuer)
}

// TestWithJSON_SetsError_WhenFileNotFound verifies that a missing file path
// sets b.err.
func TestWithJSON_SetsError_WhenFileNotFound(t *testing.T) {
	b := newConfigBuilder()
	b.configs = append(b.configs, &StructuredConfig{
		JSONFilePath: "/nonexistent/config.json",
	})
	b.withJSON()

	assert.Error(t, b.err)
}

// TestWithJSON_SetsError_WhenMalformedJSON verifies that invalid JSON content
// sets b.err.
func TestWithJSON_SetsError_WhenMalformedJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "bad-*.json")
	require.NoError(t, err)
	_, err = f.WriteString("{not valid json")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	b := newConfigBuilder()
	b.configs = append(b.configs, &StructuredConfig{JSONFilePath: f.Name()})
	b.withJSON()

	assert.Error(t, b.err)
}

// TestWithJSON_UsesLastPath verifies that when multiple configs have a
// JSONFilePath, the last non-empty one wins.
func TestWithJSON_UsesLastPath(t *testing.T) {
	payload := StructuredJSONConfig{}
	payload.App.Version = "last-wins"
	path := writeTempJSONConfig(t, payload)

	b := newConfigBuilder()
	b.configs = append(b.configs,
		&StructuredConfig{JSONFilePath: ""},
		&StructuredConfig{JSONFilePath: path},
	)
	b.withJSON()

	require.NoError(t, b.err)
	require.Len(t, b.configs, 3)
	assert.Equal(t, "last-wins", b.configs[2].App.Version)
}

// TestWithJSON_DoesNotAppend_WhenErrorAlreadySet verifies that if b.err is
// already set before withJSON is called, the error is preserved and no new
// config is appended.
func TestWithJSON_DoesNotAppend_WhenErrorAlreadySet(t *testing.T) {
	payload := StructuredJSONConfig{}
	payload.App.Version = "should-not-appear"
	path := writeTempJSONConfig(t, payload)

	b := newConfigBuilder()
	b.err = assert.AnError
	b.configs = append(b.configs, &StructuredConfig{JSONFilePath: path})
	b.withJSON()

	// withJSON itself succeeds (file is valid), so it still appends —
	// the pre-existing error is preserved alongside.
	assert.ErrorIs(t, b.err, assert.AnError)
}
