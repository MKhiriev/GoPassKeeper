package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
