package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// Mock
// ─────────────────────────────────────────────

// mockAppInfoService implements service.AppInfoService for testing.
type mockAppInfoService struct {
	version string
}

func (m *mockAppInfoService) GetAppVersion(_ context.Context) string {
	return m.version
}

// newHandlerWithAppInfo builds a Handler whose AppInfoService is replaced
// with the provided mock. All other service fields are left nil because
// getServerVersion does not use them.
func newHandlerWithAppInfo(t *testing.T, svc service.AppInfoService) *Handler {
	t.Helper()
	return NewHandler(
		&service.Services{AppInfoService: svc},
		logger.Nop(),
	)
}

// ─────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────

func TestGetServerVersion_WritesVersion(t *testing.T) {
	const want = "1.2.3"

	h := newHandlerWithAppInfo(t, &mockAppInfoService{version: want})

	req := httptest.NewRequest(http.MethodGet, "/api/version/", nil)
	rec := httptest.NewRecorder()

	h.getServerVersion(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, want, rec.Body.String())
}

func TestGetServerVersion_EmptyVersion(t *testing.T) {
	h := newHandlerWithAppInfo(t, &mockAppInfoService{version: ""})

	req := httptest.NewRequest(http.MethodGet, "/api/version/", nil)
	rec := httptest.NewRecorder()

	h.getServerVersion(rec, req)

	assert.Equal(t, "", rec.Body.String())
}

func TestGetServerVersion_VersionWithSpecialChars(t *testing.T) {
	const want = "v2.0.0-beta+build.42"

	h := newHandlerWithAppInfo(t, &mockAppInfoService{version: want})

	req := httptest.NewRequest(http.MethodGet, "/api/version/", nil)
	rec := httptest.NewRecorder()

	h.getServerVersion(rec, req)

	assert.Equal(t, want, rec.Body.String())
}

func TestGetServerVersion_ViaRouter(t *testing.T) {
	const want = "3.0.0"

	h := newHandlerWithAppInfo(t, &mockAppInfoService{version: want})
	router := h.Init()

	req := httptest.NewRequest(http.MethodGet, "/api/version/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, want, rec.Body.String())
}

func TestGetServerVersion_ContentTypeNotJSON(t *testing.T) {
	h := newHandlerWithAppInfo(t, &mockAppInfoService{version: "1.0.0"})

	req := httptest.NewRequest(http.MethodGet, "/api/version/", nil)
	rec := httptest.NewRecorder()

	h.getServerVersion(rec, req)

	// Handler writes plain text — Content-Type must NOT be application/json.
	assert.NotEqual(t, "application/json", rec.Header().Get("Content-Type"))
}
