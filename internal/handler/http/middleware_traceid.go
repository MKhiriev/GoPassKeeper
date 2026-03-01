// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// traceIDHeader is the name of the HTTP header used to propagate the
// distributed trace identifier between the client and the server.
//
// On inbound requests the middleware reads this header to reuse a
// trace ID supplied by the caller (e.g. an upstream service or the client
// application). On outbound responses the same header is set so that the
// caller can correlate its own logs with server-side log entries.
const traceIDHeader = "X-Trace-ID"

// withTraceID is an HTTP middleware that attaches a trace ID to every
// request for distributed tracing and structured logging purposes.
//
// Trace ID resolution follows this precedence:
//  1. If the incoming request contains a non-empty "X-Trace-ID" header,
//     its value is reused as the trace ID. This allows an upstream service
//     or the client to propagate an existing trace across service boundaries.
//  2. If the header is absent or empty, a new random UUID v4 is generated
//     via [uuid.NewString] and used as the trace ID for this request.
//
// Once the trace ID is determined, the middleware:
//   - Creates a child logger from [Handler.logger] via [logger.Logger.GetChildLogger]
//     and permanently attaches the "trace_id" field to it, so that every
//     subsequent log entry emitted within the request lifetime carries the
//     trace ID without requiring callers to add it manually.
//   - Stores the enriched logger in the request context using
//     [zerolog.Logger.WithContext], making it retrievable by downstream
//     middleware and handlers via [logger.FromRequest].
//   - Sets the "X-Trace-ID" response header to the resolved trace ID so
//     that clients can correlate their requests with server log entries.
//
// withTraceID must be placed early in the middleware chain — before any
// middleware that uses [logger.FromRequest] — so that the context-scoped
// logger is available to all subsequent handlers.
func (h *Handler) withTraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Resolve the trace ID: prefer the value supplied by the caller so
		// that an existing distributed trace can be continued seamlessly.
		// Fall back to a freshly generated UUID when no trace ID is provided.
		var traceID string
		if traceIDFromRequestHeader := r.Header.Get(traceIDHeader); traceIDFromRequestHeader != "" {
			traceID = traceIDFromRequestHeader
		} else {
			traceID = uuid.NewString()
		}

		// Derive a child logger and permanently embed the trace ID so that
		// every log entry written during this request is correlated by trace.
		l := h.logger.GetChildLogger()
		l.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("trace_id", traceID)
		})

		// Store the enriched logger in the request context and update the
		// request so downstream middleware and handlers can retrieve it via
		// logger.FromRequest.
		r = r.WithContext(l.WithContext(ctx))

		// Echo the trace ID back to the caller in the response header.
		w.Header().Set(traceIDHeader, traceID)

		next.ServeHTTP(w, r)
	})
}
