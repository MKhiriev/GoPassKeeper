package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newResponseWriter(rr *httptest.ResponseRecorder) *responseWriter {
	return &responseWriter{ResponseWriter: rr}
}

// ---- WriteHeader ----

func TestResponseWriter_WriteHeader_SetsStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	w := newResponseWriter(rr)

	w.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, w.status)
	assert.True(t, w.wroteHeader)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestResponseWriter_WriteHeader_CalledTwice_IgnoresSecond(t *testing.T) {
	rr := httptest.NewRecorder()
	w := newResponseWriter(rr)

	w.WriteHeader(http.StatusCreated)
	w.WriteHeader(http.StatusInternalServerError) // should be ignored

	assert.Equal(t, http.StatusCreated, w.status)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestResponseWriter_WriteHeader_TableTest(t *testing.T) {
	tests := []struct {
		name           string
		statusCodes    []int // multiple WriteHeader calls
		expectedStatus int
	}{
		{
			name:           "200 OK",
			statusCodes:    []int{http.StatusOK},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "201 Created",
			statusCodes:    []int{http.StatusCreated},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "400 Bad Request",
			statusCodes:    []int{http.StatusBadRequest},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "404 Not Found",
			statusCodes:    []int{http.StatusNotFound},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "500 Internal Server Error",
			statusCodes:    []int{http.StatusInternalServerError},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "double call — first wins",
			statusCodes:    []int{http.StatusAccepted, http.StatusBadRequest},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "triple call — first wins",
			statusCodes:    []int{http.StatusOK, http.StatusCreated, http.StatusNotFound},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			w := newResponseWriter(rr)

			for _, code := range tt.statusCodes {
				w.WriteHeader(code)
			}

			assert.Equal(t, tt.expectedStatus, w.status)
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.True(t, w.wroteHeader)
		})
	}
}

// ---- Write ----

func TestResponseWriter_Write_SetsImplicit200(t *testing.T) {
	rr := httptest.NewRecorder()
	w := newResponseWriter(rr)

	n, err := w.Write([]byte("hello"))

	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, http.StatusOK, w.status)
	assert.True(t, w.wroteHeader)
}

func TestResponseWriter_Write_AccumulatesSize(t *testing.T) {
	rr := httptest.NewRecorder()
	w := newResponseWriter(rr)

	_, err := w.Write([]byte("hello"))
	require.NoError(t, err)
	_, err = w.Write([]byte(" world"))
	require.NoError(t, err)

	assert.Equal(t, 11, w.size) // 5 + 6
}

func TestResponseWriter_Write_StoresLastBody(t *testing.T) {
	rr := httptest.NewRecorder()
	w := newResponseWriter(rr)

	_, _ = w.Write([]byte("first"))
	_, _ = w.Write([]byte("second"))

	// body stores the most recently written byte slice.
	assert.Equal(t, []byte("second"), w.body)
}

func TestResponseWriter_Write_AfterExplicitWriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	w := newResponseWriter(rr)

	w.WriteHeader(http.StatusAccepted)
	n, err := w.Write([]byte("data"))

	require.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, http.StatusAccepted, w.status) // status must not change to 200
	assert.Equal(t, 4, w.size)
}

func TestResponseWriter_Write_EmptyBody(t *testing.T) {
	rr := httptest.NewRecorder()
	w := newResponseWriter(rr)

	n, err := w.Write([]byte{})

	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, w.size)
	assert.Equal(t, http.StatusOK, w.status) // WriteHeader is still called
}

func TestResponseWriter_Write_TableTest(t *testing.T) {
	tests := []struct {
		name         string
		writes       [][]byte
		explicitCode int // 0 means do not call WriteHeader explicitly
		wantStatus   int
		wantSize     int
		wantBody     []byte // the last write
	}{
		{
			name:       "single write, implicit 200",
			writes:     [][]byte{[]byte("OK")},
			wantStatus: http.StatusOK,
			wantSize:   2,
			wantBody:   []byte("OK"),
		},
		{
			name:       "multiple writes accumulate size",
			writes:     [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")},
			wantStatus: http.StatusOK,
			wantSize:   9,
			wantBody:   []byte("baz"),
		},
		{
			name:         "explicit 201, then write",
			writes:       [][]byte{[]byte("created")},
			explicitCode: http.StatusCreated,
			wantStatus:   http.StatusCreated,
			wantSize:     7,
			wantBody:     []byte("created"),
		},
		{
			name:         "explicit 404, then write",
			writes:       [][]byte{[]byte("not found")},
			explicitCode: http.StatusNotFound,
			wantStatus:   http.StatusNotFound,
			wantSize:     9,
			wantBody:     []byte("not found"),
		},
		{
			name:       "empty write",
			writes:     [][]byte{{}},
			wantStatus: http.StatusOK,
			wantSize:   0,
			wantBody:   []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			w := newResponseWriter(rr)

			if tt.explicitCode != 0 {
				w.WriteHeader(tt.explicitCode)
			}

			for _, data := range tt.writes {
				_, err := w.Write(data)
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantStatus, w.status)
			assert.Equal(t, tt.wantSize, w.size)
			assert.Equal(t, tt.wantBody, w.body)
			assert.Equal(t, tt.wantSize, rr.Body.Len())
		})
	}
}

// ---- Initial state ----

func TestResponseWriter_InitialState(t *testing.T) {
	rr := httptest.NewRecorder()
	w := newResponseWriter(rr)

	assert.Equal(t, 0, w.status)
	assert.Equal(t, 0, w.size)
	assert.False(t, w.wroteHeader)
	assert.Nil(t, w.body)
}

// ---- Proxying to underlying ResponseWriter ----

func TestResponseWriter_ProxiesHeadersToUnderlying(t *testing.T) {
	rr := httptest.NewRecorder()
	w := newResponseWriter(rr)

	w.Header().Set("X-Custom", "value")
	w.WriteHeader(http.StatusTeapot)

	assert.Equal(t, "value", rr.Header().Get("X-Custom"))
	assert.Equal(t, http.StatusTeapot, rr.Code)
}
