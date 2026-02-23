package handler

import (
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestLogger returns a no-op logger suitable for use in tests.
func newTestLogger() *logger.Logger {
	return logger.Nop()
}

// newTestServices returns a nil *service.Services. Both http.NewHandler and
// grpc.NewHandler only store the pointer without dereferencing it, so nil is
// safe for construction-time tests.
func newTestServices() *service.Services {
	return nil
}

// TestNewHandlers_BothAddresses verifies that when both HTTPAddress and
// GRPCAddress are configured, both handlers are initialised and no error is
// returned.
func TestNewHandlers_BothAddresses(t *testing.T) {
	cfg := config.Server{
		HTTPAddress: ":8080",
		GRPCAddress: ":9090",
	}

	h, err := NewHandlers(newTestServices(), cfg, newTestLogger())

	require.NoError(t, err)
	require.NotNil(t, h)
	assert.NotNil(t, h.HTTP, "expected HTTP handler to be initialised")
	assert.NotNil(t, h.GRPC, "expected gRPC handler to be initialised")
}

// TestNewHandlers_OnlyHTTP verifies that when only HTTPAddress is configured,
// the HTTP handler is initialised and the gRPC handler remains nil.
func TestNewHandlers_OnlyHTTP(t *testing.T) {
	cfg := config.Server{
		HTTPAddress: ":8080",
	}

	h, err := NewHandlers(newTestServices(), cfg, newTestLogger())

	require.NoError(t, err)
	require.NotNil(t, h)
	assert.NotNil(t, h.HTTP, "expected HTTP handler to be initialised")
	assert.Nil(t, h.GRPC, "expected gRPC handler to be nil")
}

// TestNewHandlers_OnlyGRPC verifies that when only GRPCAddress is configured,
// the gRPC handler is initialised and the HTTP handler remains nil.
func TestNewHandlers_OnlyGRPC(t *testing.T) {
	cfg := config.Server{
		GRPCAddress: ":9090",
	}

	h, err := NewHandlers(newTestServices(), cfg, newTestLogger())

	require.NoError(t, err)
	require.NotNil(t, h)
	assert.Nil(t, h.HTTP, "expected HTTP handler to be nil")
	assert.NotNil(t, h.GRPC, "expected gRPC handler to be initialised")
}

// TestNewHandlers_NoAddresses verifies that when neither HTTPAddress nor
// GRPCAddress is configured, NewHandlers returns errNoHandlersAreCreated and
// a nil *Handlers.
func TestNewHandlers_NoAddresses(t *testing.T) {
	cfg := config.Server{}

	h, err := NewHandlers(newTestServices(), cfg, newTestLogger())

	require.ErrorIs(t, err, errNoHandlersAreCreated)
	assert.Nil(t, h)
}

// TestNewHandlers_ReturnType verifies that the returned value is of type
// *Handlers.
func TestNewHandlers_ReturnType(t *testing.T) {
	cfg := config.Server{HTTPAddress: ":8080"}

	h, err := NewHandlers(newTestServices(), cfg, newTestLogger())

	require.NoError(t, err)
	assert.IsType(t, &Handlers{}, h)
}

// TestNewHandlers_IndependentInstances verifies that two calls to NewHandlers
// produce independent *Handlers instances.
func TestNewHandlers_IndependentInstances(t *testing.T) {
	cfg := config.Server{HTTPAddress: ":8080", GRPCAddress: ":9090"}

	h1, err1 := NewHandlers(newTestServices(), cfg, newTestLogger())
	h2, err2 := NewHandlers(newTestServices(), cfg, newTestLogger())

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotSame(t, h1, h2)
	assert.NotSame(t, h1.HTTP, h2.HTTP)
	assert.NotSame(t, h1.GRPC, h2.GRPC)
}
