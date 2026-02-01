package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// CheckHTTPMethod returns 404 in cases when accessing existing routes with wrong HTTP method
func CheckHTTPMethod(router *chi.Mux) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestedURL := r.URL.Path
		requestedHTTPMethod := r.Method

		allRoutes := router.Routes()
		var foundRoute chi.Route
		for _, route := range allRoutes {
			if route.Pattern == requestedURL {
				foundRoute = route
				break
			}
		}

		if _, ok := foundRoute.Handlers[requestedHTTPMethod]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		router.ServeHTTP(w, r)
	}
}
