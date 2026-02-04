package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	// routes without authorization
	router.Group(func(r chi.Router) {
		r.Post("/api/user/register", h.register)
		r.Post("/api/user/login", h.login)
	})

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
