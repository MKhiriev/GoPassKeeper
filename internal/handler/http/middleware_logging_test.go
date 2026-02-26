package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// injectLogger puts zerolog.Logger into request context the same way
// withTraceID middleware does (via zerolog/log.Ctx).
func injectLogger(r *http.Request, l zerolog.Logger) *http.Request {
	ctx := l.WithContext(r.Context())
	return r.WithContext(ctx)
}

// newTestLogger creates a logger that writes to the provided buffer.
func newTestLogger(buf *bytes.Buffer) zerolog.Logger {
	return zerolog.New(buf).With().Timestamp().Logger()
}

// makeRequest creates a test request with a logger in context.
func makeRequest(method, path string, buf *bytes.Buffer) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	l := newTestLogger(buf)
	return injectLogger(req, l)
}

// ---- Table test ----

func TestWithLogging_TableTest(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		handlerStatus    int
		handlerResponse  string
		handlerDelay     time.Duration
		checkLogContains []string
	}{
		{
			name:            "GET 200",
			method:          http.MethodGet,
			path:            "/test",
			handlerStatus:   http.StatusOK,
			handlerResponse: "OK",
			checkLogContains: []string{
				`"method":"GET"`,
				`"uri":"/test"`,
				`"status":200`,
				`"duration":`,
				`"size":2`,
			},
		},
		{
			name:            "POST 201",
			method:          http.MethodPost,
			path:            "/api/data",
			handlerStatus:   http.StatusCreated,
			handlerResponse: "Created",
			checkLogContains: []string{
				`"method":"POST"`,
				`"uri":"/api/data"`,
				`"status":201`,
			},
		},
		{
			name:          "PUT 204 no body",
			method:        http.MethodPut,
			path:          "/update",
			handlerStatus: http.StatusNoContent,
			checkLogContains: []string{
				`"method":"PUT"`,
				`"uri":"/update"`,
				`"status":204`,
				`"size":0`,
			},
		},
		{
			name:            "DELETE 200",
			method:          http.MethodDelete,
			path:            "/resource/123",
			handlerStatus:   http.StatusOK,
			handlerResponse: "Deleted",
			checkLogContains: []string{
				`"method":"DELETE"`,
				`"uri":"/resource/123"`,
				`"status":200`,
			},
		},
		{
			name:            "GET 500 error",
			method:          http.MethodGet,
			path:            "/error",
			handlerStatus:   http.StatusInternalServerError,
			handlerResponse: "Internal Server Error",
			checkLogContains: []string{
				`"status":500`,
			},
		},
		{
			name:            "GET 404 not found",
			method:          http.MethodGet,
			path:            "/notfound",
			handlerStatus:   http.StatusNotFound,
			handlerResponse: "Not Found",
			checkLogContains: []string{
				`"status":404`,
				`"uri":"/notfound"`,
			},
		},
		{
			name:            "query parameters preserved in uri",
			method:          http.MethodGet,
			path:            "/search?q=test&limit=10",
			handlerStatus:   http.StatusOK,
			handlerResponse: "Results",
			checkLogContains: []string{
				`"uri":"/search?q=test&limit=10"`,
				`"status":200`,
			},
		},
		{
			name:            "PATCH request",
			method:          http.MethodPatch,
			path:            "/resource",
			handlerStatus:   http.StatusOK,
			handlerResponse: "Patched",
			checkLogContains: []string{
				`"method":"PATCH"`,
				`"status":200`,
			},
		},
		{
			name:          "HEAD request",
			method:        http.MethodHead,
			path:          "/check",
			handlerStatus: http.StatusOK,
			checkLogContains: []string{
				`"method":"HEAD"`,
				`"status":200`,
			},
		},
		{
			name:          "OPTIONS request",
			method:        http.MethodOptions,
			path:          "/api",
			handlerStatus: http.StatusOK,
			checkLogContains: []string{
				`"method":"OPTIONS"`,
				`"status":200`,
			},
		},
		{
			name:            "slow handler — duration logged",
			method:          http.MethodGet,
			path:            "/slow",
			handlerStatus:   http.StatusOK,
			handlerResponse: "Done",
			handlerDelay:    50 * time.Millisecond,
			checkLogContains: []string{
				`"duration":`,
				`"status":200`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.handlerDelay > 0 {
					time.Sleep(tt.handlerDelay)
				}
				w.WriteHeader(tt.handlerStatus)
				if tt.handlerResponse != "" {
					_, _ = w.Write([]byte(tt.handlerResponse))
				}
			})

			middleware := withLogging(next)

			req := makeRequest(tt.method, tt.path, &logBuf)
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.handlerStatus, rr.Code)

			logOutput := logBuf.String()
			assert.NotEmpty(t, logOutput, "log should not be empty")

			for _, expected := range tt.checkLogContains {
				assert.Contains(t, logOutput, expected, "log should contain: %s", expected)
			}
		})
	}
}

// ---- Response size ----

func TestWithLogging_ResponseSize(t *testing.T) {
	var logBuf bytes.Buffer

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strings.Repeat("a", 1024)))
	})

	middleware := withLogging(next)

	req := makeRequest(http.MethodGet, "/test", &logBuf)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, `"size":`, "log should contain size field")
	assert.Contains(t, logOutput, `1024`, "log should contain correct size value")
}

// ---- No explicit WriteHeader should log 200 ----

func TestWithLogging_NoStatusWritten(t *testing.T) {
	var logBuf bytes.Buffer

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("implicit 200"))
	})

	middleware := withLogging(next)

	req := makeRequest(http.MethodGet, "/test", &logBuf)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, logBuf.String(), `"status":200`)
}

// ---- Concurrent requests: no races ----

func TestWithLogging_ConcurrentRequests(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := withLogging(next)

	const n = 50
	done := make(chan struct{}, n)

	for i := 0; i < n; i++ {
		go func() {
			var buf bytes.Buffer
			req := makeRequest(http.MethodGet, "/concurrent", &buf)
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Contains(t, buf.String(), `"status":200`)
			done <- struct{}{}
		}()
	}

	for i := 0; i < n; i++ {
		<-done
	}
}

// ---- Duration accuracy ----

func TestWithLogging_DurationAccuracy(t *testing.T) {
	delay := 80 * time.Millisecond
	var logBuf bytes.Buffer

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	})
	middleware := withLogging(next)

	req := makeRequest(http.MethodGet, "/slow", &logBuf)
	rr := httptest.NewRecorder()

	start := time.Now()
	middleware.ServeHTTP(rr, req)
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, delay, "handler delay should be observed")
	assert.Contains(t, logBuf.String(), `"duration":`)
}

// ---- Panic is not suppressed ----

func TestWithLogging_PanicNotSuppressed(t *testing.T) {
	var logBuf bytes.Buffer

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	middleware := withLogging(next)

	req := makeRequest(http.MethodGet, "/panic", &logBuf)
	rr := httptest.NewRecorder()

	assert.Panics(t, func() {
		middleware.ServeHTTP(rr, req)
	}, "withLogging should not recover panics")
}

// ---- logger.Nop(): middleware works without a real logger ----

func TestWithLogging_NopLogger(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := withLogging(next)

	// Put nop logger into request context.
	nop := logger.Nop()
	req := httptest.NewRequest(http.MethodGet, "/nop", nil)
	ctx := nop.Logger.WithContext(req.Context())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	assert.NotPanics(t, func() {
		middleware.ServeHTTP(rr, req)
	})
	assert.Equal(t, http.StatusOK, rr.Code)
}
