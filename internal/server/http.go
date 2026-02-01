package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
)

type httpServer struct {
	server *http.Server
}

func newHTTPServer(handler http.Handler, cfg *config.Server) *httpServer {
	return &httpServer{
		server: &http.Server{
			Addr:         cfg.HTTPAddress,
			Handler:      handler,
			ReadTimeout:  cfg.RequestTimeout,
			WriteTimeout: cfg.RequestTimeout,
		},
	}
}

func (h *httpServer) RunServer() {
	if err := h.server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server ListenAndServe: %v\n", err)
	}
}

func (h *httpServer) Shutdown() {
	if err := h.server.Shutdown(context.Background()); h.server != nil && err != nil {
		// ошибки закрытия Listener
		fmt.Printf("HTTP server Shutdown: %v\n", err)
	}
}
