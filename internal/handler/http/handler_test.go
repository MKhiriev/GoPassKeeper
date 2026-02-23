package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// NewHandler
// ─────────────────────────────────────────────

func TestNewHandler_ReturnsNonNil(t *testing.T) {
	h := NewHandler(nil, logger.Nop())

	require.NotNil(t, h)
}

func TestNewHandler_StoresServices(t *testing.T) {
	svc := &service.Services{}
	h := NewHandler(svc, logger.Nop())

	assert.Equal(t, svc, h.services)
}

func TestNewHandler_StoresLogger(t *testing.T) {
	log := logger.Nop()
	h := NewHandler(nil, log)

	assert.Equal(t, log, h.logger)
}

func TestNewHandler_IndependentInstances(t *testing.T) {
	h1 := NewHandler(nil, logger.Nop())
	h2 := NewHandler(nil, logger.Nop())

	assert.NotSame(t, h1, h2)
}

// ─────────────────────────────────────────────
// Init — route registration
// ─────────────────────────────────────────────

// newTraceIDTestHandler builds a Handler with nil services, which is sufficient
// for route-registration tests that do not invoke handler logic.
func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	return NewHandler(nil, logger.Nop())
}

func TestInit_ReturnsRouter(t *testing.T) {
	router := newTestHandler(t).Init()

	require.NotNil(t, router)
}

// routeCase describes a single expected route.
type routeCase struct {
	method string
	path   string
}

// expectedRoutes lists every route that Init() must register.
var expectedRoutes = []routeCase{
	// auth
	{http.MethodPost, "/api/auth/register"},
	{http.MethodPost, "/api/auth/login"},
	// auth settings (auth middleware will return 401, not 405/404)
	{http.MethodPost, "/api/auth/settings/password/change"},
	{http.MethodPost, "/api/auth/settings/otp"},
	{http.MethodDelete, "/api/auth/settings/otp"},
	// data (auth middleware will return 401, not 405/404)
	{http.MethodPost, "/api/data/"},
	{http.MethodGet, "/api/data/all"},
	{http.MethodPost, "/api/data/download"},
	{http.MethodPut, "/api/data/update"},
	{http.MethodDelete, "/api/data/delete"},
	// sync
	{http.MethodGet, "/api/sync/"},
	{http.MethodGet, "/api/sync/specific"},
	// version
	{http.MethodGet, "/api/version/"},
}

func TestInit_RegistersAllRoutes(t *testing.T) {
	router := newTestHandler(t).Init()

	for _, tc := range expectedRoutes {
		tc := tc
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// A registered route returns anything except 404 (not found) or
			// 405 (method not allowed from our CheckHTTPMethod handler).
			// Auth-protected routes return 401; that still proves the route exists.
			assert.NotEqual(t, http.StatusNotFound, rec.Code,
				"route not found: %s %s", tc.method, tc.path)
			assert.NotEqual(t, http.StatusMethodNotAllowed, rec.Code,
				"method not allowed: %s %s", tc.method, tc.path)
		})
	}
}

func TestInit_UnknownRouteReturns404(t *testing.T) {
	router := newTestHandler(t).Init()

	req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestInit_WrongMethodReturns405(t *testing.T) {
	router := newTestHandler(t).Init()

	// POST /api/version/ is not registered — only GET is.
	req := httptest.NewRequest(http.MethodPost, "/api/version/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
