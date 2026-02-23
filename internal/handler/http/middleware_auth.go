// Package http implements the HTTP transport layer of the application.
// It provides middleware, route handlers, and request/response utilities
// for the REST API. Authentication, logging, tracing, compression, and
// integrity-checking concerns are all handled at this layer before
// requests are forwarded to the service layer.
package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
)

// auth is an HTTP middleware that enforces JWT-based authentication.
//
// It inspects the incoming "Authorization" header, extracts the bearer token,
// validates it via [service.AuthService.ParseToken], and — on success — stores
// the authenticated user's ID in the request context under [utils.UserIDCtxKey]
// before delegating to the next handler.
//
// The middleware rejects requests with HTTP 401 Unauthorized in the following cases:
//   - The "Authorization" header is absent ([ErrEmptyAuthorizationHeader]).
//   - The header value cannot be parsed as a bearer token
//     ([ErrInvalidAuthorizationHeader] or [ErrEmptyToken]).
//   - The token has expired ([service.ErrTokenIsExpired]).
//   - The token is otherwise invalid or cannot be parsed.
//
// All rejection events are logged using the context-scoped logger obtained
// via [logger.FromRequest].
func (h *Handler) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromRequest(r)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Err(ErrEmptyAuthorizationHeader).Send()
			http.Error(w, ErrEmptyAuthorizationHeader.Error(), http.StatusUnauthorized)
			return
		}

		tokenString, err := getTokenFromAuthHeader(authHeader)
		if err != nil {
			log.Err(err).Send()
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		token, err := h.services.AuthService.ParseToken(ctx, tokenString)

		if err != nil {
			switch {
			case errors.Is(err, service.ErrTokenIsExpired):
				log.Err(err).Msg("token expired")
				http.Error(w, service.ErrTokenIsExpired.Error(), http.StatusUnauthorized)
				return
			default:
				log.Err(err).Msg("error occurred during parsing token")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}

		// Store the authenticated user's ID in the context so that downstream
		// handlers can retrieve it without re-parsing the token.
		ctx = context.WithValue(ctx, utils.UserIDCtxKey, token.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getTokenFromAuthHeader extracts the bearer token string from a raw
// "Authorization" HTTP header value.
//
// The header is expected to follow the standard format:
//
//	Authorization: <scheme> <token>
//
// For example:
//
//	Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
//
// It returns the following sentinel errors:
//   - [ErrInvalidAuthorizationHeader] — if the header contains fewer than
//     two space-separated parts (i.e. the token is missing entirely).
//   - [ErrEmptyToken] — if the second part exists but is an empty string.
func getTokenFromAuthHeader(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) < 2 {
		return "", ErrInvalidAuthorizationHeader
	}

	tokenString := parts[1]
	if tokenString == "" {
		return "", ErrEmptyToken
	}

	return tokenString, nil
}
