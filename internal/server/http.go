package server

import (
	"context"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

type httpServer struct {
	server *http.Server

	logger *logger.Logger
}

func newHTTPServer(handler http.Handler, cfg config.Server, logger *logger.Logger) *httpServer {
	return &httpServer{
		server: &http.Server{
			Addr:         cfg.HTTPAddress,
			Handler:      handler,
			ReadTimeout:  cfg.RequestTimeout,
			WriteTimeout: cfg.RequestTimeout,
		},
		logger: logger,
	}
}

func (h *httpServer) RunServer() {
	if err := h.server.ListenAndServe(); err != nil {
		h.logger.Debug().Msgf("HTTP server ListenAndServe: %v\n", err)
	}
}

func (h *httpServer) Shutdown() {
	if err := h.server.Shutdown(context.Background()); h.server != nil && err != nil {
		// ошибки закрытия Listener
		h.logger.Error().Msgf("HTTP server Shutdown: %v\n", err)
	}
}
