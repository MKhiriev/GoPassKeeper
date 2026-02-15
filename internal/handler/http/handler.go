package http

import (
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
)

type Handler struct {
	services *service.Services

	logger *logger.Logger
}

func NewHandler(services *service.Services, logger *logger.Logger) *Handler {
	logger.Debug().Msg("http handler created")
	return &Handler{
		services: services,
		logger:   logger,
	}
}
