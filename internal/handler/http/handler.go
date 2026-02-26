package http

import (
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
)

// Handler is the root HTTP handler that wires together all route groups
// and middleware chains for the REST API.
//
// It holds references to the application's service layer and a structured
// logger so that every sub-handler and middleware can access business logic
// and emit consistent, context-enriched log entries.
//
// Handler is constructed once at application startup via [NewHandler] and
// its routes are registered by the setup methods defined in routes.go.
// It is not safe to copy a Handler after construction.
type Handler struct {
	// services provides access to all application business-logic operations.
	// Sub-handlers delegate domain work (authentication, vault management,
	// synchronization, etc.) exclusively through this interface.
	services *service.Services

	// logger is the structured logger used by the handler and all middleware
	// for request-scoped and diagnostic log output.
	logger *logger.Logger
}

// NewHandler constructs a [Handler] with the provided service container and
// logger, and returns a pointer to the initialised instance.
//
// The logger is used immediately to emit a debug-level startup message and
// is stored for use by all route handlers and middleware registered on this
// Handler.
//
// Parameters:
//   - services: the application service layer; must not be nil.
//   - logger: structured logger for request tracing and diagnostics; must not be nil.
func NewHandler(services *service.Services, logger *logger.Logger) *Handler {
	logger.Debug().Msg("http handler created")
	return &Handler{
		services: services,
		logger:   logger,
	}
}
