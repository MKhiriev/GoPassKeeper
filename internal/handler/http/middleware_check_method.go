// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// CheckHTTPMethod returns an [http.HandlerFunc] that is intended to be
// registered as the router's MethodNotAllowed handler via
// [chi.Mux.MethodNotAllowed].
//
// Chi's default behaviour is to respond with HTTP 405 Method Not Allowed
// whenever a request path matches a registered route but the HTTP method
// is not handled. This function overrides that behaviour: if the requested
// method is not registered for the matched route, it responds with
// HTTP 404 Not Found instead, effectively hiding the existence of the route
// from callers that use an unsupported method.
//
// If the requested method IS registered for the matched route, the request
// is forwarded to the router's normal ServeHTTP pipeline so that the
// appropriate handler executes as usual.
//
// The lookup is performed by iterating over all routes registered on router
// and comparing each route's pattern against the raw request path
// ([http.Request.URL.Path]). Only exact pattern matches are considered;
// parameterised or wildcard segments are not expanded during this check.
//
// Usage:
//
//	router := chi.NewRouter()
//	// ... register routes ...
//	router.MethodNotAllowed(CheckHTTPMethod(router))
func CheckHTTPMethod(router *chi.Mux) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestedURL := r.URL.Path
		requestedHTTPMethod := r.Method

		// Search for a route whose pattern exactly matches the requested path.
		allRoutes := router.Routes()
		var foundRoute chi.Route
		for _, route := range allRoutes {
			if route.Pattern == requestedURL {
				foundRoute = route
				break
			}
		}

		// If the matched route does not handle the requested HTTP method,
		// return 404 instead of the default 405 to avoid leaking route existence.
		if _, ok := foundRoute.Handlers[requestedHTTPMethod]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// The method is registered — delegate to the router's normal pipeline.
		router.ServeHTTP(w, r)
	}
}
