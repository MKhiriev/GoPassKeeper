package http

import (
	"net/http"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

func (h *Handler) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromRequest(r)

		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		lw := &responseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		log.Info().
			Str("uri", uri).
			Str("method", method).
			Int("status", lw.status).
			Dur("duration", duration).
			Int("size", lw.size).
			Send()
	})
}
