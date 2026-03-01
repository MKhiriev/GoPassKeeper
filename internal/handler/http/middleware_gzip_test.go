// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGZip(t *testing.T) {
	tests := []struct {
		name                 string
		acceptEncoding       string
		contentEncoding      string
		requestBody          []byte
		compressRequestBody  bool
		expectedStatus       int
		expectedResponseBody string
		checkResponseGzipped bool
		checkRequestDecoded  bool
	}{
		{
			name:                 "compress response when client accepts gzip",
			acceptEncoding:       "gzip",
			requestBody:          nil,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "Hello, World!",
			checkResponseGzipped: true,
		},
		{
			name:                 "no compression when client doesn't accept gzip",
			acceptEncoding:       "",
			requestBody:          nil,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "Hello, World!",
			checkResponseGzipped: false,
		},
		{
			name:                 "accept-encoding with multiple values including gzip",
			acceptEncoding:       "deflate, gzip, br",
			requestBody:          nil,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "Hello, World!",
			checkResponseGzipped: true,
		},
		{
			name:                 "accept-encoding with gzip and quality values",
			acceptEncoding:       "gzip;q=1.0, identity;q=0.5",
			requestBody:          nil,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "Hello, World!",
			checkResponseGzipped: true,
		},
		{
			name:                "decompress gzipped request body",
			acceptEncoding:      "",
			contentEncoding:     "gzip",
			requestBody:         []byte("Request data"),
			compressRequestBody: true,
			expectedStatus:      http.StatusOK,
			checkRequestDecoded: true,
		},
		{
			name:                 "decompress request and compress response",
			acceptEncoding:       "gzip",
			contentEncoding:      "gzip",
			requestBody:          []byte("Request data"),
			compressRequestBody:  true,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "Processed: Request data",
			checkResponseGzipped: true,
			checkRequestDecoded:  true,
		},
		{
			name:                "invalid gzip request body",
			acceptEncoding:      "",
			contentEncoding:     "gzip",
			requestBody:         []byte("not gzipped data"),
			compressRequestBody: false,
			expectedStatus:      http.StatusBadRequest,
		},
		{
			name:                 "large response body compression",
			acceptEncoding:       "gzip",
			requestBody:          nil,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: strings.Repeat("Large data ", 1000),
			checkResponseGzipped: true,
		},
		{
			name:                 "compress JSON response",
			acceptEncoding:       "gzip",
			requestBody:          nil,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: `{"message":"Hello","data":[1,2,3,4,5]}`,
			checkResponseGzipped: true,
		},
		{
			name:                "content-encoding with multiple values including gzip",
			acceptEncoding:      "",
			contentEncoding:     "gzip, deflate",
			requestBody:         []byte("Request data"),
			compressRequestBody: true,
			expectedStatus:      http.StatusOK,
			checkRequestDecoded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create next handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkRequestDecoded && r.Body != nil {
					// Read and verify decompressed request body
					body, err := io.ReadAll(r.Body)
					require.NoError(t, err, "failed to read request body")
					assert.Equal(t, string(tt.requestBody), string(body), "request body should be decompressed")

					// Verify Content-Encoding header is removed
					assert.Empty(t, r.Header.Get("Content-Encoding"), "Content-Encoding should be removed")
				}

				// Write response
				w.WriteHeader(tt.expectedStatus)
				if tt.expectedResponseBody != "" {
					if tt.checkRequestDecoded {
						w.Write([]byte("Processed: " + string(tt.requestBody)))
					} else {
						w.Write([]byte(tt.expectedResponseBody))
					}
				}
			})

			// Create middleware
			middleware := withGZip(nextHandler)

			// Prepare request body
			var requestBody io.Reader
			if tt.requestBody != nil {
				if tt.compressRequestBody {
					var buf bytes.Buffer
					gzipWriter := gzip.NewWriter(&buf)
					_, err := gzipWriter.Write(tt.requestBody)
					require.NoError(t, err)
					err = gzipWriter.Close()
					require.NoError(t, err)
					requestBody = &buf
				} else {
					requestBody = bytes.NewReader(tt.requestBody)
				}
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/test", requestBody)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			rr := httptest.NewRecorder()

			// Execute middleware
			middleware.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "unexpected status code")

			// Check response compression
			if tt.checkResponseGzipped {
				// Verify Content-Encoding header
				assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"), "Content-Encoding should be gzip")

				// Decompress and verify response body
				gzipReader, err := gzip.NewReader(rr.Body)
				require.NoError(t, err, "failed to create gzip reader")
				defer gzipReader.Close()

				decompressed, err := io.ReadAll(gzipReader)
				require.NoError(t, err, "failed to decompress response")

				assert.Equal(t, tt.expectedResponseBody, string(decompressed), "decompressed response should match")
			} else if tt.expectedResponseBody != "" && tt.expectedStatus == http.StatusOK {
				// Response should not be compressed
				assert.NotEqual(t, "gzip", rr.Header().Get("Content-Encoding"), "Content-Encoding should not be gzip")
				assert.Equal(t, tt.expectedResponseBody, rr.Body.String(), "response body should not be compressed")
			}
		})
	}
}

func TestGZip_CompressionRatio(t *testing.T) {
	// Test that compression actually reduces size for repetitive data
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write highly compressible data
		data := strings.Repeat("This is repetitive data. ", 1000)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(data))
	})

	middleware := withGZip(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	originalSize := len(strings.Repeat("This is repetitive data. ", 1000))
	compressedSize := rr.Body.Len()

	// Compressed size should be significantly smaller
	assert.Less(t, compressedSize, originalSize/10, "compressed size should be much smaller than original")
}

func TestGZip_MultipleRequests(t *testing.T) {
	// Test that pool reuse works correctly
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response"))
	})

	middleware := withGZip(nextHandler)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "request %d failed", i)
		assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"), "request %d missing gzip encoding", i)

		// Verify response can be decompressed
		gzipReader, err := gzip.NewReader(rr.Body)
		require.NoError(t, err, "request %d: failed to create gzip reader", i)

		decompressed, err := io.ReadAll(gzipReader)
		require.NoError(t, err, "request %d: failed to decompress", i)
		gzipReader.Close()

		assert.Equal(t, "Response", string(decompressed), "request %d: wrong response", i)
	}
}

func TestGZip_ConcurrentRequests(t *testing.T) {
	// Test thread safety of pool usage
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Concurrent response"))
	})

	middleware := withGZip(nextHandler)

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Accept-Encoding", "gzip")

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			gzipReader, err := gzip.NewReader(rr.Body)
			if err == nil {
				io.ReadAll(gzipReader)
				gzipReader.Close()
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestGZip_RequestBodyPoolReuse(t *testing.T) {
	// Test that request body decompression pool works correctly
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	middleware := withGZip(nextHandler)

	for i := 0; i < 5; i++ {
		// Compress request body
		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		testData := []byte("Test data " + string(rune('0'+i)))
		gzipWriter.Write(testData)
		gzipWriter.Close()

		req := httptest.NewRequest(http.MethodPost, "/test", &buf)
		req.Header.Set("Content-Encoding", "gzip")

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "request %d failed", i)
		assert.Equal(t, string(testData), rr.Body.String(), "request %d: wrong body", i)
	}
}

func TestGZipResponseWriter_WriteHeader(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	})

	middleware := withGZip(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))
}

func TestGZip_EmptyResponse(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	middleware := withGZip(nextHandler)

	req := httptest.NewRequest(http.MethodDelete, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))
}

func TestWrappedReadCloser_Close(t *testing.T) {
	closeCalled := false
	onClose := func() {
		closeCalled = true
	}

	wrapped := &wrappedReadCloser{
		Reader:  strings.NewReader("test"),
		OnClose: onClose,
	}

	err := wrapped.Close()
	assert.NoError(t, err)
	assert.True(t, closeCalled, "OnClose should be called")
}

func TestWrappedReadCloser_CloseWithoutCallback(t *testing.T) {
	wrapped := &wrappedReadCloser{
		Reader:  strings.NewReader("test"),
		OnClose: nil,
	}

	err := wrapped.Close()
	assert.NoError(t, err, "Close should not fail when OnClose is nil")
}
