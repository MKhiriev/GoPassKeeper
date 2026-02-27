// Package handler provides initialization logic for all inbound transport
// adapters used by the go-pass-keeper application, including HTTP and gRPC
// handlers. The package exposes a unified Handlers struct, which bundles
// transport-specific handler implementations so they can be started uniformly
// by the application’s main entrypoint.
package handler

import (
	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/handler/http"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
)

// Handlers groups all initialized inbound transport handlers, such as
// the HTTP handler and the gRPC handler. The main application uses this
// structure to start the appropriate servers based on configuration.
type Handlers struct {
	// HTTP contains the initialized HTTP handler if HTTP is enabled in the
	// configuration. If HTTP is disabled, this field remains nil.
	HTTP *http.Handler

	// GRPC contains the initialized gRPC handler if gRPC is enabled in the
	// configuration. If gRPC is disabled, this field remains nil.
	//GRPC *grpc.Handler
}

// NewHandlers constructs the Handlers bundle from the provided service layer,
// server configuration, and logger.
//
// Behavior:
//   - If cfg.HTTPAddress is non-empty, an HTTP handler is created.
//   - If cfg.GRPCAddress is non-empty, a gRPC handler is created.
//   - If both addresses are empty, the function returns errNoHandlersAreCreated.
//
// This ensures the application fails fast if misconfigured such that no inbound
// transport is enabled.
//
// Returns:
//   - (*Handlers, nil) if one or more handlers are successfully created;
//   - (nil, error) if neither HTTP nor gRPC handlers could be initialized.
func NewHandlers(services *service.Services, cfg config.Server, logger *logger.Logger) (*Handlers, error) {
	logger.Info().Msg("creating new handlers...")

	handlers := &Handlers{}

	if cfg.HTTPAddress != "" {
		handlers.HTTP = http.NewHandler(services, logger)
	}
	//if cfg.GRPCAddress != "" {
	//	handlers.GRPC = grpc.NewHandler(services, logger)
	//}

	if handlers.HTTP == nil /*&& handlers.GRPC == nil*/ {
		return nil, errNoHandlersAreCreated
	}

	return handlers, nil
}
