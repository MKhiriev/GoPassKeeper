package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const traceIDHeader = "X-Trace-ID"

func (h *Handler) withTraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var traceID string
		if traceIDFromRequestHeader := r.Header.Get(traceIDHeader); traceIDFromRequestHeader != "" {
			traceID = traceIDFromRequestHeader
		} else {
			traceID = uuid.NewString()
		}

		l := h.logger.GetChildLogger()
		l.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("trace_id", traceID)
		})
		r = r.WithContext(l.WithContext(ctx))

		w.Header().Set(traceIDHeader, traceID)
		next.ServeHTTP(w, r)
	})
}
