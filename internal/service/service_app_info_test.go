package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// NewAppInfoService
// ─────────────────────────────────────────────

func TestNewAppInfoService_Success(t *testing.T) {
	cfg := config.App{Version: "1.0.0"}

	svc, err := NewAppInfoService(cfg, logger.Nop())

	require.NoError(t, err)
	require.NotNil(t, svc)
}

func TestNewAppInfoService_EmptyVersion_ReturnsError(t *testing.T) {
	cfg := config.App{Version: ""}

	svc, err := NewAppInfoService(cfg, logger.Nop())

	assert.Nil(t, svc)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrVersionIsNotSpecified))
}

func TestNewAppInfoService_ReturnsAppInfoServiceInterface(t *testing.T) {
	cfg := config.App{Version: "2.5.1"}

	svc, err := NewAppInfoService(cfg, logger.Nop())

	require.NoError(t, err)
	// compile-time check: returned value must satisfy the interface
	var _ AppInfoService = svc
}

// ─────────────────────────────────────────────
// GetAppVersion
// ─────────────────────────────────────────────

func TestGetAppVersion_ReturnsConfiguredVersion(t *testing.T) {
	cfg := config.App{Version: "3.1.4"}
	svc, err := NewAppInfoService(cfg, logger.Nop())
	require.NoError(t, err)

	got := svc.GetAppVersion(context.Background())

	assert.Equal(t, "3.1.4", got)
}

func TestGetAppVersion_VersionIsStable(t *testing.T) {
	cfg := config.App{Version: "0.0.1"}
	svc, err := NewAppInfoService(cfg, logger.Nop())
	require.NoError(t, err)

	ctx := context.Background()
	first := svc.GetAppVersion(ctx)
	second := svc.GetAppVersion(ctx)

	assert.Equal(t, first, second, "version must not change between calls")
}

func TestGetAppVersion_DifferentInstances_IndependentVersions(t *testing.T) {
	svc1, err := NewAppInfoService(config.App{Version: "1.0.0"}, logger.Nop())
	require.NoError(t, err)

	svc2, err := NewAppInfoService(config.App{Version: "2.0.0"}, logger.Nop())
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", svc1.GetAppVersion(context.Background()))
	assert.Equal(t, "2.0.0", svc2.GetAppVersion(context.Background()))
}

func TestGetAppVersion_VersionWithSpecialChars(t *testing.T) {
	version := "v1.2.3-beta+build.42"
	svc, err := NewAppInfoService(config.App{Version: version}, logger.Nop())
	require.NoError(t, err)

	assert.Equal(t, version, svc.GetAppVersion(context.Background()))
}

func TestGetAppVersion_CancelledContext_StillReturnsVersion(t *testing.T) {
	svc, err := NewAppInfoService(config.App{Version: "1.0.0"}, logger.Nop())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// GetAppVersion does not use ctx, so it must still return the version
	assert.Equal(t, "1.0.0", svc.GetAppVersion(ctx))
}
