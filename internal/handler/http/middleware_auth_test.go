package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Helpers ----

func newHandlerWithAuthService(authSvc service.AuthService) *Handler {
	return &Handler{
		logger: logger.Nop(),
		services: &service.Services{
			AuthService: authSvc,
		},
	}
}

// injectNopLogger кладёт nop-логгер в контекст запроса.
func injectNopLogger(r *http.Request) *http.Request {
	nop := logger.Nop()
	ctx := nop.Logger.WithContext(r.Context())
	return r.WithContext(ctx)
}

func executeAuth(h *Handler, authHeader string, next http.Handler) *httptest.ResponseRecorder {
	middleware := h.auth(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = injectNopLogger(req)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	return rr
}

// ---- getTokenFromAuthHeader unit tests ----

func TestGetTokenFromAuthHeader_TableTest(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantErr   error
	}{
		{
			name:      "valid Bearer token",
			header:    "Bearer my-jwt-token",
			wantToken: "my-jwt-token",
		},
		{
			name:    "missing token part",
			header:  "Bearer",
			wantErr: ErrInvalidAuthorizationHeader,
		},
		{
			name:    "empty header",
			header:  "",
			wantErr: ErrInvalidAuthorizationHeader,
		},
		{
			name:      "non-Bearer scheme still parses second part",
			header:    "Basic dXNlcjpwYXNz",
			wantToken: "dXNlcjpwYXNz",
		},
		{
			name:    "only spaces",
			header:  " ",
			wantErr: ErrEmptyToken,
		},
		{
			name:      "extra parts — second part is used",
			header:    "Bearer token extra-part",
			wantToken: "token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := getTokenFromAuthHeader(tt.header)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

// ---- auth middleware table test ----

func TestAuth_Middleware_TableTest(t *testing.T) {
	validToken := models.Token{UserID: 42}

	tests := []struct {
		name           string
		authHeader     string
		parseTokenFn   func(ctx context.Context, s string) (models.Token, error)
		expectedStatus int
		nextCalled     bool
		wantUserID     int64
	}{
		{
			name:           "empty Authorization header → 401",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			nextCalled:     false,
		},
		{
			name:           "invalid header format (no space) → 401",
			authHeader:     "BearerTokenWithoutSpace",
			expectedStatus: http.StatusUnauthorized,
			nextCalled:     false,
		},
		{
			name:       "valid token → next called, userID in context",
			authHeader: "Bearer valid-token",
			parseTokenFn: func(_ context.Context, _ string) (models.Token, error) {
				return validToken, nil
			},
			expectedStatus: http.StatusOK,
			nextCalled:     true,
			wantUserID:     42,
		},
		{
			name:       "expired token → 401 with specific error",
			authHeader: "Bearer expired-token",
			parseTokenFn: func(_ context.Context, _ string) (models.Token, error) {
				return models.Token{}, service.ErrTokenIsExpired
			},
			expectedStatus: http.StatusUnauthorized,
			nextCalled:     false,
		},
		{
			name:       "other parse error → 401",
			authHeader: "Bearer bad-token",
			parseTokenFn: func(_ context.Context, _ string) (models.Token, error) {
				return models.Token{}, service.ErrTokenIsExpiredOrInvalid
			},
			expectedStatus: http.StatusUnauthorized,
			nextCalled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var authSvc service.AuthService
			if tt.parseTokenFn != nil {
				authSvc = &mockAuthService{parseTokenFn: tt.parseTokenFn}
			} else {
				// parseTokenFn не должна вызваться — header пустой или невалидный
				authSvc = &mockAuthService{parseTokenFn: func(_ context.Context, _ string) (models.Token, error) {
					t.Fatal("ParseToken should not be called")
					return models.Token{}, nil
				}}
			}

			h := newHandlerWithAuthService(authSvc)

			nextCalled := false
			var capturedUserID any
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				capturedUserID = r.Context().Value(utils.UserIDCtxKey)
				w.WriteHeader(http.StatusOK)
			})

			rr := executeAuth(h, tt.authHeader, next)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.nextCalled, nextCalled)

			if tt.nextCalled && tt.wantUserID != 0 {
				assert.Equal(t, tt.wantUserID, capturedUserID)
			}
		})
	}
}

// ---- Тело ответа при ошибках ----

func TestAuth_ErrorResponseBodies(t *testing.T) {
	h := newHandlerWithAuthService(&mockAuthService{
		parseTokenFn: func(_ context.Context, _ string) (models.Token, error) {
			return models.Token{}, service.ErrTokenIsExpired
		},
	})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("empty header error body", func(t *testing.T) {
		rr := executeAuth(h, "", next)
		assert.Contains(t, rr.Body.String(), ErrEmptyAuthorizationHeader.Error())
	})

	t.Run("expired token error body", func(t *testing.T) {
		rr := executeAuth(h, "Bearer expired", next)
		assert.Contains(t, rr.Body.String(), service.ErrTokenIsExpired.Error())
	})
}

// ---- UserID корректно кладётся в контекст ----

func TestAuth_UserIDInContext(t *testing.T) {
	const expectedUserID int64 = 99

	h := newHandlerWithAuthService(&mockAuthService{
		parseTokenFn: func(_ context.Context, _ string) (models.Token, error) {
			return models.Token{UserID: expectedUserID}, nil
		},
	})

	var gotUserID any
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = r.Context().Value(utils.UserIDCtxKey)
		w.WriteHeader(http.StatusOK)
	})

	rr := executeAuth(h, "Bearer some-token", next)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, expectedUserID, gotUserID)
}

// ---- Оригинальный контекст не мутируется ----

func TestAuth_OriginalRequestNotMutated(t *testing.T) {
	h := newHandlerWithAuthService(&mockAuthService{
		parseTokenFn: func(_ context.Context, _ string) (models.Token, error) {
			return models.Token{UserID: 1}, nil
		},
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := h.auth(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = injectNopLogger(req)
	req.Header.Set("Authorization", "Bearer token")
	originalCtx := req.Context()

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, originalCtx, req.Context(), "original request context must not be mutated")
}

// ---- Concurrent requests — нет гонок ----

func TestAuth_ConcurrentRequests(t *testing.T) {
	h := newHandlerWithAuthService(&mockAuthService{
		parseTokenFn: func(_ context.Context, _ string) (models.Token, error) {
			return models.Token{UserID: 7}, nil
		},
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := h.auth(next)

	const n = 50
	done := make(chan int, n)

	for i := 0; i < n; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = injectNopLogger(req)
			req.Header.Set("Authorization", "Bearer concurrent-token")
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)
			done <- rr.Code
		}()
	}

	for i := 0; i < n; i++ {
		code := <-done
		assert.Equal(t, http.StatusOK, code)
	}
}
