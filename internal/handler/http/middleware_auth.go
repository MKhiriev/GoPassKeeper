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

func (h *Handler) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromRequest(r)

		// token is expired case
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

		ctx = context.WithValue(ctx, utils.UserIDCtxKey, token.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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
