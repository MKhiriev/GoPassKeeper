// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// buildRouter creates a minimal chi.Mux with a set of routes for tests.
// It intentionally does not use Handler.Init() to avoid service/logger setup.
func buildRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/api/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("items"))
	})
	router.Post("/api/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	router.Put("/api/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.Delete("/api/resource", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}

// ---- Table test ----

func TestCheckHTTPMethod_TableTest(t *testing.T) {
	router := buildRouter()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		// Existing route + valid method -> handler responds.
		{
			name:           "GET /api/items — registered, should pass through",
			method:         http.MethodGet,
			path:           "/api/items",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST /api/items — registered, should pass through",
			method:         http.MethodPost,
			path:           "/api/items",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "PUT /api/items — registered, should pass through",
			method:         http.MethodPut,
			path:           "/api/items",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET /api/users — registered, should pass through",
			method:         http.MethodGet,
			path:           "/api/users",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DELETE /api/resource — registered, should pass through",
			method:         http.MethodDelete,
			path:           "/api/resource",
			expectedStatus: http.StatusNoContent,
		},
		// Existing route + invalid method -> 404.
		{
			name:           "DELETE /api/items — method not registered → 404",
			method:         http.MethodDelete,
			path:           "/api/items",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "PATCH /api/items — method not registered → 404",
			method:         http.MethodPatch,
			path:           "/api/items",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "POST /api/users — method not registered → 404",
			method:         http.MethodPost,
			path:           "/api/users",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "GET /api/resource — method not registered → 404",
			method:         http.MethodGet,
			path:           "/api/resource",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "PUT /api/users — method not registered → 404",
			method:         http.MethodPut,
			path:           "/api/users",
			expectedStatus: http.StatusNotFound,
		},
		// Non-existing route: chi returns 404 before MethodNotAllowed.
		{
			name:           "GET /api/nonexistent — route does not exist",
			method:         http.MethodGet,
			path:           "/api/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

// ---- Existing route with valid method forwards response body ----

func TestCheckHTTPMethod_PassThroughBody(t *testing.T) {
	router := buildRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "items", rr.Body.String())
}

// ---- Invalid method always returns 404, not 405 ----

func TestCheckHTTPMethod_WrongMethodReturns404NotMethodNotAllowed(t *testing.T) {
	router := buildRouter()

	wrongMethods := []string{
		http.MethodDelete,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodHead,
	}

	for _, method := range wrongMethods {
		t.Run(method+" /api/items", func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/items", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusNotFound, rr.Code,
				"wrong method on existing route should return 404, not 405")
			assert.NotEqual(t, http.StatusMethodNotAllowed, rr.Code)
		})
	}
}

// ---- Route with a single method returns 404 for all others ----

func TestCheckHTTPMethod_SingleMethodRoute(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/only-get", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.MethodNotAllowed(CheckHTTPMethod(router))

	// The only registered method should pass.
	req := httptest.NewRequest(http.MethodGet, "/only-get", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// All other methods should return 404.
	for _, method := range []string{
		http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodOptions,
	} {
		t.Run("wrong: "+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/only-get", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusNotFound, rr.Code)
		})
	}
}

// ---- Route with multiple methods allows each registered one ----

func TestCheckHTTPMethod_MultiMethodRoute(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/multi", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	router.Post("/multi", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusCreated) })
	router.Delete("/multi", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	router.MethodNotAllowed(CheckHTTPMethod(router))

	registered := map[string]int{
		http.MethodGet:    http.StatusOK,
		http.MethodPost:   http.StatusCreated,
		http.MethodDelete: http.StatusNoContent,
	}
	unregistered := []string{http.MethodPut, http.MethodPatch, http.MethodOptions}

	for method, wantStatus := range registered {
		t.Run("registered: "+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/multi", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, wantStatus, rr.Code)
		})
	}

	for _, method := range unregistered {
		t.Run("unregistered: "+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/multi", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusNotFound, rr.Code)
		})
	}
}

// ---- Concurrent requests: no races ----

func TestCheckHTTPMethod_ConcurrentRequests(t *testing.T) {
	router := buildRouter()
	const n = 50
	done := make(chan int, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			var method, path string
			var wantStatus int
			if i%2 == 0 {
				method, path, wantStatus = http.MethodGet, "/api/items", http.StatusOK
			} else {
				method, path, wantStatus = http.MethodDelete, "/api/items", http.StatusNotFound
			}
			req := httptest.NewRequest(method, path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			done <- rr.Code
			_ = wantStatus
		}(i)
	}

	for i := 0; i < n; i++ {
		code := <-done
		assert.True(t, code == http.StatusOK || code == http.StatusNotFound,
			"unexpected status code: %d", code)
	}
}
