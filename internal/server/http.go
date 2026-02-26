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

// RunServer starts the HTTP listener and serves incoming requests.
func (h *httpServer) RunServer() {
	if err := h.server.ListenAndServe(); err != nil {
		h.logger.Debug().Msgf("HTTP server ListenAndServe: %v\n", err)
	}
}

// Shutdown gracefully stops the HTTP server.
func (h *httpServer) Shutdown() {
	if err := h.server.Shutdown(context.Background()); h.server != nil && err != nil {
		// Listener shutdown errors.
		h.logger.Error().Msgf("HTTP server Shutdown: %v\n", err)
	}
}
