package http

import (
	"net/http"
	"time"
)

func (h *Handler) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		lw := &responseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		h.logger.Info().
			Str("uri", uri).
			Str("method", method).
			Int("status", lw.status).
			Dur("duration", duration).
			Int("size", lw.size).
			Send()
	})
}
