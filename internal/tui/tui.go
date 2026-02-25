package tui

import (
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
)

type TUI struct {
	services *service.ClientServices
}

func New(services *service.ClientServices, logger *logger.Logger) (*TUI, error) {
	return &TUI{
		services: services,
	}, nil
}
