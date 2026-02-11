package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer, h.withTraceID, h.withLogging, withGZip)

	router.Route("/api", func(api chi.Router) {

		// auth service routes
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", h.register)
			auth.Post("/login", h.login)

			auth.Route("/settings", func(settings chi.Router) {
				settings.Use(h.auth)

				settings.Post("/password/change", h.changeUserPassword)
				settings.Post("/otp", h.setUserOTP)
				settings.Delete("/otp", h.deleteUserOTP)
			})
		})

		// user data routes
		api.Route("/data", func(data chi.Router) {
			data.Use(h.auth)

			data.Post("/", h.upload)

			data.Get("/all", h.downloadAllUserData)
			data.Get("/{type}/{id}", h.download)
			data.Post("/value", h.downloadMultiple)

			data.Put("/{type}/{id}", h.update)
			data.Delete("/{type}/{id}", h.delete)
		})

		// client-server sync data routes
		api.Route("/sync", func(sync chi.Router) {
			sync.Use(h.auth)

			sync.Get("/diff", h.getClientServerDiff)

			sync.Get("/{type}/{id}", h.syncSpecificUserData)
		})

		// client-server sync routes
		api.Route("/version", func(version chi.Router) {
			version.Get("/", h.getServerVersion)
		})
	})

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
