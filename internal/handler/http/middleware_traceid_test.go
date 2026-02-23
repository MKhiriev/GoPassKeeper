package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestHandler создаёт Handler с nop-логгером (без вывода в stdout).
func newTestHandler() *Handler {
	return &Handler{logger: logger.Nop()}
}

// ---- Helpers ----

func executeWithTraceID(h *Handler, traceIDHeader string) (*httptest.ResponseRecorder, *http.Request) {
	var capturedReq *http.Request
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	})

	middleware := h.withTraceID(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if traceIDHeader != "" {
		req.Header.Set("X-Trace-ID", traceIDHeader)
	}

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	return rr, capturedReq
}

// ---- Таблица: заголовок ответа X-Trace-ID ----

func TestWithTraceID_TableTest(t *testing.T) {
	tests := []struct {
		name            string
		requestTraceID  string
		wantSameTraceID bool // true — ответный header должен совпасть с requestTraceID
		wantValidUUID   bool // true — ответный header должен быть валидным UUID v4
		wantNextCalled  bool
		wantStatus      int
	}{
		{
			name:            "trace ID from request header is reused",
			requestTraceID:  "my-custom-trace-id",
			wantSameTraceID: true,
			wantNextCalled:  true,
			wantStatus:      http.StatusOK,
		},
		{
			name:           "no trace ID in request — UUID generated",
			requestTraceID: "",
			wantValidUUID:  true,
			wantNextCalled: true,
			wantStatus:     http.StatusOK,
		},
		{
			name:            "UUID v4 string as incoming trace ID",
			requestTraceID:  "550e8400-e29b-41d4-a716-446655440000",
			wantSameTraceID: true,
			wantNextCalled:  true,
			wantStatus:      http.StatusOK,
		},
		{
			name:            "long custom trace ID is preserved",
			requestTraceID:  "very-long-trace-id-that-is-still-valid-0123456789abcdef",
			wantSameTraceID: true,
			wantNextCalled:  true,
			wantStatus:      http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler()
			nextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(tt.wantStatus)
			})

			middleware := h.withTraceID(next)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.requestTraceID != "" {
				req.Header.Set(traceIDHeader, tt.requestTraceID)
			}

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			responseTraceID := rr.Header().Get(traceIDHeader)
			require.NotEmpty(t, responseTraceID, "X-Trace-ID header must be set in response")

			if tt.wantSameTraceID {
				assert.Equal(t, tt.requestTraceID, responseTraceID)
			}

			if tt.wantValidUUID {
				_, err := uuid.Parse(responseTraceID)
				assert.NoError(t, err, "generated trace ID should be a valid UUID, got: %s", responseTraceID)
			}

			assert.Equal(t, tt.wantNextCalled, nextCalled)
			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}

// ---- Генерация уникальных trace ID при отсутствии заголовка ----

func TestWithTraceID_GeneratesUniqueUUIDs(t *testing.T) {
	h := newTestHandler()
	seen := make(map[string]struct{})

	for i := 0; i < 100; i++ {
		rr, _ := executeWithTraceID(h, "")
		id := rr.Header().Get(traceIDHeader)
		require.NotEmpty(t, id)

		_, err := uuid.Parse(id)
		require.NoError(t, err, "trace ID must be valid UUID, got: %s", id)

		_, duplicate := seen[id]
		assert.False(t, duplicate, "duplicate trace ID generated: %s", id)
		seen[id] = struct{}{}
	}
}

// ---- Trace ID попадает в контекст запроса ----

func TestWithTraceID_TraceIDInContext(t *testing.T) {
	h := newTestHandler()

	t.Run("custom trace ID from header is in context logger", func(t *testing.T) {
		customID := "trace-context-test"
		var ctxLogger *logger.Logger

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxLogger = logger.FromRequest(r)
			w.WriteHeader(http.StatusOK)
		})

		middleware := h.withTraceID(next)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(traceIDHeader, customID)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		// Логгер должен быть доступен из контекста (не nil, не паникует)
		require.NotNil(t, ctxLogger)
	})

	t.Run("generated trace ID is in context logger", func(t *testing.T) {
		var ctxLogger *logger.Logger

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxLogger = logger.FromRequest(r)
			w.WriteHeader(http.StatusOK)
		})

		middleware := h.withTraceID(next)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		require.NotNil(t, ctxLogger)
	})
}

// ---- Next handler всегда вызывается ----

func TestWithTraceID_AlwaysCallsNext(t *testing.T) {
	h := newTestHandler()
	nextCalled := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
	})

	middleware := h.withTraceID(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusTeapot, rr.Code)
}

// ---- Заголовок X-Trace-ID всегда присутствует в ответе ----

func TestWithTraceID_ResponseHeaderAlwaysSet(t *testing.T) {
	h := newTestHandler()

	t.Run("without incoming trace ID", func(t *testing.T) {
		rr, _ := executeWithTraceID(h, "")
		assert.NotEmpty(t, rr.Header().Get(traceIDHeader))
	})

	t.Run("with incoming trace ID", func(t *testing.T) {
		rr, _ := executeWithTraceID(h, "existing-id")
		assert.Equal(t, "existing-id", rr.Header().Get(traceIDHeader))
	})
}

// ---- Concurrent requests — нет гонок ----

func TestWithTraceID_ConcurrentRequests(t *testing.T) {
	h := newTestHandler()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := h.withTraceID(next)

	const n = 50
	done := make(chan string, n)

	for i := 0; i < n; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)
			done <- rr.Header().Get(traceIDHeader)
		}()
	}

	seen := make(map[string]struct{})
	for i := 0; i < n; i++ {
		id := <-done
		require.NotEmpty(t, id)
		_, err := uuid.Parse(id)
		require.NoError(t, err)
		seen[id] = struct{}{}
	}

	assert.Len(t, seen, n, "all generated trace IDs should be unique")
}

// ---- Оригинальный запрос не мутируется ----

func TestWithTraceID_OriginalRequestNotMutated(t *testing.T) {
	h := newTestHandler()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := h.withTraceID(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	originalCtx := req.Context()

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Контекст оригинального запроса не должен измениться
	assert.Equal(t, originalCtx, req.Context(), "original request context should not be mutated")
}
