package handler

import (
	"github.com/MKhiriev/go-pass-keeper/internal/handler/grpc"
	"github.com/MKhiriev/go-pass-keeper/internal/handler/http"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
)

type Handlers struct {
	HTTP *http.Handler
	GRPC *grpc.Handler
}

func NewHandlers(services *service.Services, logger *logger.Logger) (*Handlers, error) {
	logger.Info().Msg("creating new handlers...")

	httpHandler := http.NewHandler(services, logger)
	gRPCHandler := grpc.NewHandler(services, logger)

	if httpHandler == nil && gRPCHandler == nil {
		return nil, errNoHandlersAreCreated
	}

	return &Handlers{
		HTTP: httpHandler,
		GRPC: gRPCHandler,
	}, nil
}
