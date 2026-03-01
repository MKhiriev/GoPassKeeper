// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

// withLogging is an HTTP middleware that records structured access-log entries
// for every request processed by the handler chain.
//
// For each incoming request the middleware captures:
//   - uri — the raw request URI as sent by the client ([http.Request.RequestURI]),
//     including the path, query string, and any encoded characters.
//   - method — the HTTP method (GET, POST, PUT, …).
//   - status — the HTTP status code written by the downstream handler.
//     If the handler never calls [http.ResponseWriter.WriteHeader], the status
//     defaults to 0 and is recorded as such; the actual wire response will be
//     200 OK because the standard library sends 200 on the first [Write].
//   - duration — wall-clock time elapsed from the moment withLogging wraps the
//     request until the downstream handler returns.
//   - size — the total number of bytes written to the response body by the
//     downstream handler, as tracked by the intercepting [responseWriter].
//
// The log entry is emitted at INFO level via the context-scoped logger obtained
// from [logger.FromRequest]. The logger must have been placed in the request
// context by an earlier middleware (e.g. withTraceID) before withLogging is
// invoked; otherwise the global zerolog no-op logger is used and the entry is
// silently discarded.
//
// withLogging does not recover from panics. If the downstream handler panics,
// the log entry is never written and the panic propagates to the caller.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromRequest(r)

		start := time.Now()

		// Capture URI and method before delegating, in case a downstream
		// handler mutates the request (e.g. via r.WithContext).
		uri := r.RequestURI
		method := r.Method

		// Read and restore the request body so downstream handlers can still read it.
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				log.Debug().RawJSON("incoming data", bodyBytes).Msg("incoming request")
				// Restore the body for downstream handlers.
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		// Wrap the ResponseWriter so that the status code and response body
		// size written by downstream handlers can be observed after the call.
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
