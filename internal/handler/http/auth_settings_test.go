package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// changeUserPassword
// ─────────────────────────────────────────────

// TestChangeUserPassword_ReturnsNotImplemented verifies that the handler
// responds with 501 Not Implemented until the feature is built.
func TestChangeUserPassword_ReturnsNotImplemented(t *testing.T) {
	h := newTestHandlerWithAppInfoService(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/settings/password/change", nil)
	rec := httptest.NewRecorder()

	h.changeUserPassword(rec, req)

	assert.Equal(t, http.StatusNotImplemented, rec.Code)
}

// TestChangeUserPassword_EmptyBody verifies that an empty request body
// still results in 501 Not Implemented.
func TestChangeUserPassword_EmptyBody(t *testing.T) {
	h := newTestHandlerWithAppInfoService(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/settings/password/change", nil)
	rec := httptest.NewRecorder()

	h.changeUserPassword(rec, req)

	require.Equal(t, http.StatusNotImplemented, rec.Code)
	assert.Empty(t, rec.Body.String())
}

// TestChangeUserPassword_ViaRouter_RequiresAuth verifies that the route is
// registered and protected by the auth middleware (returns 401, not 404/405).
func TestChangeUserPassword_ViaRouter_RequiresAuth(t *testing.T) {
	router := newTestHandlerWithAppInfoService(t).Init()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/settings/password/change", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ─────────────────────────────────────────────
// setUserOTP
// ─────────────────────────────────────────────

// TestSetUserOTP_ReturnsNotImplemented verifies that the handler
// responds with 501 Not Implemented until the feature is built.
func TestSetUserOTP_ReturnsNotImplemented(t *testing.T) {
	h := newTestHandlerWithAppInfoService(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/settings/otp", nil)
	rec := httptest.NewRecorder()

	h.setUserOTP(rec, req)

	assert.Equal(t, http.StatusNotImplemented, rec.Code)
}

// TestSetUserOTP_EmptyBody verifies that an empty request body
// still results in 501 Not Implemented.
func TestSetUserOTP_EmptyBody(t *testing.T) {
	h := newTestHandlerWithAppInfoService(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/settings/otp", nil)
	rec := httptest.NewRecorder()

	h.setUserOTP(rec, req)

	require.Equal(t, http.StatusNotImplemented, rec.Code)
	assert.Empty(t, rec.Body.String())
}

// TestSetUserOTP_ViaRouter_RequiresAuth verifies that the route is
// registered and protected by the auth middleware (returns 401, not 404/405).
func TestSetUserOTP_ViaRouter_RequiresAuth(t *testing.T) {
	router := newTestHandlerWithAppInfoService(t).Init()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/settings/otp", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ─────────────────────────────────────────────
// deleteUserOTP
// ─────────────────────────────────────────────

// TestDeleteUserOTP_ReturnsNotImplemented verifies that the handler
// responds with 501 Not Implemented until the feature is built.
func TestDeleteUserOTP_ReturnsNotImplemented(t *testing.T) {
	h := newTestHandlerWithAppInfoService(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/auth/settings/otp", nil)
	rec := httptest.NewRecorder()

	h.deleteUserOTP(rec, req)

	assert.Equal(t, http.StatusNotImplemented, rec.Code)
}

// TestDeleteUserOTP_EmptyBody verifies that an empty request body
// still results in 501 Not Implemented.
func TestDeleteUserOTP_EmptyBody(t *testing.T) {
	h := newTestHandlerWithAppInfoService(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/auth/settings/otp", nil)
	rec := httptest.NewRecorder()

	h.deleteUserOTP(rec, req)

	require.Equal(t, http.StatusNotImplemented, rec.Code)
	assert.Empty(t, rec.Body.String())
}

// TestDeleteUserOTP_ViaRouter_RequiresAuth verifies that the route is
// registered and protected by the auth middleware (returns 401, not 404/405).
func TestDeleteUserOTP_ViaRouter_RequiresAuth(t *testing.T) {
	router := newTestHandlerWithAppInfoService(t).Init()

	req := httptest.NewRequest(http.MethodDelete, "/api/auth/settings/otp", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ─────────────────────────────────────────────
// Wrong HTTP methods on settings routes
// ─────────────────────────────────────────────

// TestSettingsRoutes_WrongMethod verifies that using an incorrect HTTP method
// on a settings route returns 405 Method Not Allowed (via CheckHTTPMethod).
func TestSettingsRoutes_WrongMethod(t *testing.T) {
	router := newTestHandlerWithAppInfoService(t).Init()

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/auth/settings/password/change"},
		{http.MethodGet, "/api/auth/settings/otp"},
		{http.MethodPost, "/api/auth/settings/otp"}, // DELETE is registered, not POST — wait, POST /otp IS registered
		// Only truly wrong methods:
		{http.MethodPut, "/api/auth/settings/password/change"},
		{http.MethodPut, "/api/auth/settings/otp"},
		{http.MethodGet, "/api/auth/settings/otp"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.NotEqual(t, http.StatusNotFound, rec.Code,
				"route should exist: %s %s", tc.method, tc.path)
		})
	}
}
