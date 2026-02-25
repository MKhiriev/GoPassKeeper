// Package http implements the HTTP transport layer of the application.
// It provides middleware, route handlers, and request/response utilities
// for the REST API. Authentication, logging, tracing, compression, and
// integrity-checking concerns are all handled at this layer before
// requests are forwarded to the service layer.
package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Init constructs and returns a fully configured [chi.Mux] router that
// serves all API endpoints of the application.
//
// # Global middleware
//
// Every request passes through the following middleware chain in order:
//   - [middleware.Recoverer] — catches panics in handlers, logs the stack
//     trace, and returns HTTP 500 to the client so the server stays alive.
//   - [Handler.withTraceID] — resolves or generates a trace ID and stores
//     an enriched logger in the request context for structured tracing.
//   - withLogging — emits a structured access-log entry (URI, method,
//     status, duration, response size) after each request completes.
//   - withGZip — transparently decompresses gzip-encoded request bodies and
//     compresses response bodies for clients that advertise gzip support.
//
// # Route groups
//
// All routes are nested under the "/api" prefix:
//
//	/api/auth
//	  POST /register       — create a new user account (public).
//	  POST /login          — authenticate and receive a JWT (public).
//	  /settings            — account settings (requires JWT via [Handler.auth]):
//	    POST /password/change — update the master password.
//	    POST /otp             — enable or update the OTP secret.
//	    DELETE /otp           — disable OTP for the account.
//
//	/api/data              — vault item operations (requires JWT):
//	  POST /               — upload new vault items
//	                         (additionally guarded by [uploadHashing]).
//	  GET  /all            — download all vault items for the authenticated user.
//	  POST /download       — download a specific subset of vault items.
//	  PUT  /update         — update existing vault items
//	                         (additionally guarded by [updateHashing]).
//	  DELETE /delete       — soft-delete vault items.
//
//	/api/sync              — client-server synchronisation (requires JWT):
//	  GET /                — retrieve the diff between client and server state.
//	  GET /specific        — retrieve states for a specific subset of items.
//
//	/api/version           — server metadata (public):
//	  GET /                — return the current server version string.
//
// # Method-not-allowed behaviour
//
// [CheckHTTPMethod] is registered as the MethodNotAllowed handler. It
// overrides chi's default HTTP 405 response and returns HTTP 404 instead,
// preventing callers from discovering which HTTP methods are supported on
// a given route through error-code enumeration.
func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer, h.withTraceID, withLogging, withGZip)

	router.Route("/api", func(api chi.Router) {

		// Authentication and account-management routes.
		api.Route("/auth", func(auth chi.Router) {
			// Public endpoints — no JWT required.
			auth.Post("/register", h.register)
			auth.Post("/login", h.login)
			auth.Post("/params", h.params)

			// Protected settings endpoints — JWT required via h.auth.
			auth.Route("/settings", func(settings chi.Router) {
				settings.Use(h.auth)

				settings.Post("/password/change", h.changeUserPassword)
				settings.Post("/otp", h.setUserOTP)
				settings.Delete("/otp", h.deleteUserOTP)
			})
		})

		// Vault item (private data) routes — JWT required for all endpoints.
		api.Route("/data", func(data chi.Router) {
			data.Use(h.auth)

			// uploadHashing verifies the transport integrity checksum of the
			// uploaded payload before the request reaches the upload handler.
			data.With(uploadHashing).Post("/", h.upload)

			data.Get("/all", h.downloadAllUserData)
			data.Post("/download", h.downloadMultiple)

			// updateHashing verifies the transport integrity checksum of the
			// update payload before the request reaches the update handler.
			data.With(updateHashing).Put("/update", h.update)
			data.Delete("/delete", h.delete)
		})

		// Client-server synchronisation routes — JWT required for all endpoints.
		api.Route("/sync", func(sync chi.Router) {
			sync.Use(h.auth)

			sync.Get("/", h.getClientServerDiff)
			sync.Get("/specific", h.syncSpecificUserData)
		})

		// Server metadata routes — public, no authentication required.
		api.Route("/version", func(version chi.Router) {
			version.Get("/", h.getServerVersion)
		})
	})

	// Replace chi's default 405 Method Not Allowed with 404 Not Found so that
	// callers cannot enumerate supported HTTP methods through error codes.
	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
