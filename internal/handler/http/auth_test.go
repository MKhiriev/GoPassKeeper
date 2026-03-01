// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// Mock AuthService
// ─────────────────────────────────────────────

// mockAuthService implements service.AuthService for unit tests.
// Each method field can be overridden per test case.
type mockAuthService struct {
	registerUserFn func(ctx context.Context, user models.User) (models.User, error)
	loginFn        func(ctx context.Context, user models.User) (models.User, error)
	createTokenFn  func(ctx context.Context, user models.User) (models.Token, error)
	parseTokenFn   func(ctx context.Context, tokenString string) (models.Token, error)
	paramsFn       func(ctx context.Context, user models.User) (models.User, error)
}

func (m *mockAuthService) RegisterUser(ctx context.Context, user models.User) (models.User, error) {
	return m.registerUserFn(ctx, user)
}

func (m *mockAuthService) Login(ctx context.Context, user models.User) (models.User, error) {
	return m.loginFn(ctx, user)
}

func (m *mockAuthService) CreateToken(ctx context.Context, user models.User) (models.Token, error) {
	return m.createTokenFn(ctx, user)
}

func (m *mockAuthService) ParseToken(ctx context.Context, tokenString string) (models.Token, error) {
	return m.parseTokenFn(ctx, tokenString)
}

func (m *mockAuthService) Params(ctx context.Context, user models.User) (models.User, error) {
	return m.paramsFn(ctx, user)
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

// newHandlerWithAuth builds a Handler with the given AuthService mock.
func newHandlerWithAuth(t *testing.T, auth service.AuthService) *Handler {
	t.Helper()
	svcs := &service.Services{
		AppInfoService: &mockAppInfoService{version: "test"},
		AuthService:    auth,
	}
	return NewHandler(svcs, logger.Nop())
}

// userBody serialises a models.User to a JSON request body string.
func userBody(t *testing.T, u models.User) string {
	t.Helper()
	b, err := json.Marshal(u)
	require.NoError(t, err)
	return string(b)
}

// stubToken returns a models.Token with the given signed string.
func stubToken(signed string) models.Token {
	return models.Token{SignedString: signed}
}

// validUser is a convenience fixture used across multiple tests.
var validUser = models.User{
	Login:          "alice",
	MasterPassword: "hashed-password",
}

// ─────────────────────────────────────────────
// register — success
// ─────────────────────────────────────────────

// TestRegister_Success verifies that a valid registration request results in
// 200 OK and an Authorization header containing the issued Bearer token.
func TestRegister_Success(t *testing.T) {
	const signedToken = "signed.jwt.token"

	auth := &mockAuthService{
		registerUserFn: func(_ context.Context, u models.User) (models.User, error) {
			return u, nil
		},
		createTokenFn: func(_ context.Context, _ models.User) (models.Token, error) {
			return stubToken(signedToken), nil
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Bearer "+signedToken, rec.Header().Get("Authorization"))
}

// ─────────────────────────────────────────────
// register — invalid JSON
// ─────────────────────────────────────────────

// TestRegister_InvalidJSON verifies that a malformed request body results in
// 400 Bad Request.
func TestRegister_InvalidJSON(t *testing.T) {
	h := newHandlerWithAuth(t, &mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader("{invalid json}"))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid JSON was passed")
}

// TestRegister_EmptyBody verifies that an empty request body results in
// 400 Bad Request.
func TestRegister_EmptyBody(t *testing.T) {
	h := newHandlerWithAuth(t, &mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(""))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ─────────────────────────────────────────────
// register — RegisterUser errors
// ─────────────────────────────────────────────

// TestRegister_InvalidDataProvided verifies that service.ErrInvalidDataProvided
// maps to 400 Bad Request.
func TestRegister_InvalidDataProvided(t *testing.T) {
	auth := &mockAuthService{
		registerUserFn: func(_ context.Context, _ models.User) (models.User, error) {
			return models.User{}, service.ErrInvalidDataProvided
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid data provided")
}

// TestRegister_LoginAlreadyExists verifies that store.ErrLoginAlreadyExists
// maps to 409 Conflict.
func TestRegister_LoginAlreadyExists(t *testing.T) {
	auth := &mockAuthService{
		registerUserFn: func(_ context.Context, _ models.User) (models.User, error) {
			return models.User{}, store.ErrLoginAlreadyExists
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "login already exists")
}

// TestRegister_UnexpectedError verifies that an unknown error from RegisterUser
// maps to 500 Internal Server Error.
func TestRegister_UnexpectedError(t *testing.T) {
	auth := &mockAuthService{
		registerUserFn: func(_ context.Context, _ models.User) (models.User, error) {
			return models.User{}, errors.New("db connection lost")
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ─────────────────────────────────────────────
// register — CreateToken error
// ─────────────────────────────────────────────

// TestRegister_CreateTokenFails verifies that a token creation failure after
// successful registration maps to 500 Internal Server Error.
func TestRegister_CreateTokenFails(t *testing.T) {
	auth := &mockAuthService{
		registerUserFn: func(_ context.Context, u models.User) (models.User, error) {
			return u, nil
		},
		createTokenFn: func(_ context.Context, _ models.User) (models.Token, error) {
			return models.Token{}, errors.New("signing key unavailable")
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ─────────────────────────────────────────────
// register — wrapped errors
// ─────────────────────────────────────────────

// TestRegister_WrappedLoginAlreadyExists verifies that a wrapped
// store.ErrLoginAlreadyExists is still matched via errors.Is.
func TestRegister_WrappedLoginAlreadyExists(t *testing.T) {
	auth := &mockAuthService{
		registerUserFn: func(_ context.Context, _ models.User) (models.User, error) {
			return models.User{}, errors.Join(errors.New("outer"), store.ErrLoginAlreadyExists)
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

// ─────────────────────────────────────────────
// login — success
// ─────────────────────────────────────────────

// TestLogin_Success verifies that a valid login request results in
// 200 OK and an Authorization header containing the issued Bearer token.
func TestLogin_Success(t *testing.T) {
	const signedToken = "login.jwt.token"

	auth := &mockAuthService{
		loginFn: func(_ context.Context, u models.User) (models.User, error) {
			return u, nil
		},
		createTokenFn: func(_ context.Context, _ models.User) (models.Token, error) {
			return stubToken(signedToken), nil
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Bearer "+signedToken, rec.Header().Get("Authorization"))
}

// ─────────────────────────────────────────────
// login — invalid JSON
// ─────────────────────────────────────────────

// TestLogin_InvalidJSON verifies that a malformed request body results in
// 400 Bad Request.
func TestLogin_InvalidJSON(t *testing.T) {
	h := newHandlerWithAuth(t, &mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{bad json"))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid JSON was passed")
}

// TestLogin_EmptyBody verifies that an empty request body results in
// 400 Bad Request.
func TestLogin_EmptyBody(t *testing.T) {
	h := newHandlerWithAuth(t, &mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(""))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ─────────────────────────────────────────────
// login — Login errors
// ─────────────────────────────────────────────

// TestLogin_InvalidDataProvided verifies that service.ErrInvalidDataProvided
// maps to 400 Bad Request.
func TestLogin_InvalidDataProvided(t *testing.T) {
	auth := &mockAuthService{
		loginFn: func(_ context.Context, _ models.User) (models.User, error) {
			return models.User{}, service.ErrInvalidDataProvided
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid data provided")
}

// TestLogin_UserNotFound verifies that store.ErrNoUserWasFound
// maps to 401 Unauthorized.
func TestLogin_UserNotFound(t *testing.T) {
	auth := &mockAuthService{
		loginFn: func(_ context.Context, _ models.User) (models.User, error) {
			return models.User{}, store.ErrNoUserWasFound
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid login/password")
}

// TestLogin_WrongPassword verifies that service.ErrWrongPassword
// maps to 401 Unauthorized.
func TestLogin_WrongPassword(t *testing.T) {
	auth := &mockAuthService{
		loginFn: func(_ context.Context, _ models.User) (models.User, error) {
			return models.User{}, service.ErrWrongPassword
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid login/password")
}

// TestLogin_UnexpectedError verifies that an unknown error from Login
// maps to 500 Internal Server Error.
func TestLogin_UnexpectedError(t *testing.T) {
	auth := &mockAuthService{
		loginFn: func(_ context.Context, _ models.User) (models.User, error) {
			return models.User{}, errors.New("unexpected db error")
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ─────────────────────────────────────────────
// login — CreateToken error
// ─────────────────────────────────────────────

// TestLogin_CreateTokenFails verifies that a token creation failure after
// successful login maps to 500 Internal Server Error.
func TestLogin_CreateTokenFails(t *testing.T) {
	auth := &mockAuthService{
		loginFn: func(_ context.Context, u models.User) (models.User, error) {
			return u, nil
		},
		createTokenFn: func(_ context.Context, _ models.User) (models.Token, error) {
			return models.Token{}, errors.New("signing key unavailable")
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ─────────────────────────────────────────────
// login — wrapped errors
// ─────────────────────────────────────────────

// TestLogin_WrappedWrongPassword verifies that a wrapped service.ErrWrongPassword
// is still matched via errors.Is.
func TestLogin_WrappedWrongPassword(t *testing.T) {
	auth := &mockAuthService{
		loginFn: func(_ context.Context, _ models.User) (models.User, error) {
			return models.User{}, errors.Join(errors.New("outer"), service.ErrWrongPassword)
		},
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ─────────────────────────────────────────────
// Authorization header format
// ─────────────────────────────────────────────

// TestRegister_AuthorizationHeaderFormat verifies the exact format of the
// Authorization header: "Bearer <token>".
func TestRegister_AuthorizationHeaderFormat(t *testing.T) {
	const signed = "abc.def.ghi"

	auth := &mockAuthService{
		registerUserFn: func(_ context.Context, u models.User) (models.User, error) { return u, nil },
		createTokenFn:  func(_ context.Context, _ models.User) (models.Token, error) { return stubToken(signed), nil },
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.register(rec, req)

	assert.Equal(t, "Bearer abc.def.ghi", rec.Header().Get("Authorization"))
}

// TestLogin_AuthorizationHeaderFormat verifies the exact format of the
// Authorization header: "Bearer <token>".
func TestLogin_AuthorizationHeaderFormat(t *testing.T) {
	const signed = "x.y.z"

	auth := &mockAuthService{
		loginFn:       func(_ context.Context, u models.User) (models.User, error) { return u, nil },
		createTokenFn: func(_ context.Context, _ models.User) (models.Token, error) { return stubToken(signed), nil },
	}

	h := newHandlerWithAuth(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(userBody(t, validUser)))
	rec := httptest.NewRecorder()

	h.login(rec, req)

	assert.Equal(t, "Bearer x.y.z", rec.Header().Get("Authorization"))
}
