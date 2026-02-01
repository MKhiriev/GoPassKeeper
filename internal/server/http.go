package server

import (
	"context"
	"fmt"
	"net/http"
)

type httpServer struct {
	server *http.Server
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
