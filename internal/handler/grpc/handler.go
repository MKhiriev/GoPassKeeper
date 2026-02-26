package grpc

import (
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
)

// Handler is the root gRPC transport handler.
//
// It stores references to the service layer and structured logger so that
// gRPC method handlers can delegate business logic and emit consistent logs.
// A handler instance is created once at startup and shared by the gRPC server.
type Handler struct {
	// services provides access to all application business operations.
	services *service.Services

	// logger is used for request-scoped and diagnostic log output.
	logger *logger.Logger
}

// NewHandler constructs a [Handler] with the provided service container and
// logger, and returns the initialized instance.
//
// Parameters:
//   - services: application service layer used by gRPC method handlers.
//   - logger: structured logger used for transport diagnostics.
func NewHandler(services *service.Services, logger *logger.Logger) *Handler {
	logger.Debug().Msg("gRPC handler created")
	return &Handler{
		services: services,
		logger:   logger,
	}
}
