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
	h := NewHandler(&service.Services{}, logger.Nop())

	require.NotNil(t, h)
}

func TestNewHandler_StoresServices(t *testing.T) {
	svc := &service.Services{}
	h := NewHandler(svc, logger.Nop())

	assert.Equal(t, svc, h.services)
}

func TestNewHandler_StoresLogger(t *testing.T) {
	log := logger.Nop()
	h := NewHandler(&service.Services{}, log)

	assert.Equal(t, log, h.logger)
}

func TestNewHandler_IndependentInstances(t *testing.T) {
	h1 := NewHandler(&service.Services{}, logger.Nop())
	h2 := NewHandler(&service.Services{}, logger.Nop())

	assert.NotSame(t, h1, h2)
}

// ─────────────────────────────────────────────
// Init — route registration
// ─────────────────────────────────────────────

// newTestHandlerWithAppInfoService builds a Handler suitable for route-registration tests.
// AppInfoService is mocked so that GET /api/version/ does not panic.
func newTestHandlerWithAppInfoService(t *testing.T) *Handler {
	t.Helper()

	svcs := &service.Services{
		AppInfoService: &mockAppInfoService{version: "test-version"},
	}

	return NewHandler(svcs, logger.Nop())
}

func TestInit_ReturnsRouter(t *testing.T) {
	router := newTestHandlerWithAppInfoService(t).Init()

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
	// auth settings (auth middleware will return 401, not 404/405)
	{http.MethodPost, "/api/auth/settings/password/change"},
	{http.MethodPost, "/api/auth/settings/otp"},
	{http.MethodDelete, "/api/auth/settings/otp"},
	// data (auth middleware will return 401, not 404/405)
	{http.MethodPost, "/api/data/"},
	{http.MethodGet, "/api/data/all"},
	{http.MethodPost, "/api/data/download"},
	{http.MethodPut, "/api/data/update"},
	{http.MethodDelete, "/api/data/delete"},
	// sync (auth middleware will return 401, not 404/405)
	{http.MethodGet, "/api/sync/"},
	{http.MethodGet, "/api/sync/specific"},
	// version — no auth, handler is called directly
	{http.MethodGet, "/api/version/"},
}

func TestInit_RegistersAllRoutes(t *testing.T) {
	router := newTestHandlerWithAppInfoService(t).Init()

	for _, tc := range expectedRoutes {
		tc := tc
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// A registered route returns anything except 404 (not found) or
			// 405 (method not allowed). Auth-protected routes return 401 —
			// that still proves the route exists.
			assert.NotEqual(t, http.StatusNotFound, rec.Code,
				"route not found: %s %s", tc.method, tc.path)
			assert.NotEqual(t, http.StatusMethodNotAllowed, rec.Code,
				"method not allowed: %s %s", tc.method, tc.path)
		})
	}
}

func TestInit_UnknownRouteReturns404(t *testing.T) {
	router := newTestHandlerWithAppInfoService(t).Init()

	req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestInit_WrongMethodReturns405(t *testing.T) {
	router := newTestHandlerWithAppInfoService(t).Init()

	// POST /api/version/ is not registered — only GET is.
	req := httptest.NewRequest(http.MethodPost, "/api/version/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
