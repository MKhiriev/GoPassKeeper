package handler

import (
	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/handler/grpc"
	"github.com/MKhiriev/go-pass-keeper/internal/handler/http"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
)

type Handlers struct {
	HTTP *http.Handler
	GRPC *grpc.Handler
}

func NewHandlers(services *service.Services, cfg config.Server, logger *logger.Logger) (*Handlers, error) {
	logger.Info().Msg("creating new handlers...")

	handlers := &Handlers{}

	if cfg.HTTPAddress != "" {
		handlers.HTTP = http.NewHandler(services, logger)
	}
	if cfg.GRPCAddress != "" {
		handlers.GRPC = grpc.NewHandler(services, logger)
	}

	if handlers.HTTP == nil && handlers.GRPC == nil {
		return nil, errNoHandlersAreCreated
	}

	return handlers, nil
}
