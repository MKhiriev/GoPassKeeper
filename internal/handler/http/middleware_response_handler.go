// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import "net/http"

// responseData is a value-type snapshot of a completed HTTP response.
// It is used to pass response metadata (status code, body size, and raw body)
// between components that need to inspect the response after it has been written,
// without retaining a reference to the live [responseWriter].
type responseData struct {
	// status is the HTTP status code that was written to the response.
	status int

	// size is the total number of bytes written to the response body.
	size int

	// body holds the raw bytes of the most recent Write call.
	// Note: if Write was called multiple times, body contains only the
	// slice passed in the last call, not the concatenation of all writes.
	body []byte
}

// responseWriter is a thin decorator around [http.ResponseWriter] that
// intercepts WriteHeader and Write calls to capture response metadata.
//
// It is used by middleware (e.g. withLogging) to observe the HTTP status code
// and the total number of bytes written to the response body after the
// downstream handler has returned, without buffering the entire response.
//
// responseWriter ensures that WriteHeader is forwarded to the underlying
// writer exactly once: subsequent calls are silently ignored, mirroring the
// behaviour documented by the [http.ResponseWriter] interface.
type responseWriter struct {
	http.ResponseWriter

	// status is the HTTP status code recorded on the first WriteHeader call.
	// It is zero until WriteHeader (or an implicit WriteHeader via Write) is called.
	status int

	// wroteHeader reports whether WriteHeader has already been called.
	// It guards against forwarding a second WriteHeader to the underlying writer.
	wroteHeader bool

	// size is the running total of bytes successfully written to the response body
	// across all Write calls.
	size int

	// body holds the byte slice passed to the most recent Write call.
	// It is NOT a concatenation of all writes; it is overwritten on each call.
	body []byte
}

// WriteHeader records the status code and forwards it to the underlying
// [http.ResponseWriter] exactly once.
//
// If WriteHeader has already been called for this response, the call is a
// no-op and statusCode is ignored. This matches the contract of the standard
// library's [http.ResponseWriter], which states that WriteHeader may only be
// called once per response.
func (w *responseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.status = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write writes b to the underlying [http.ResponseWriter] and accumulates
// the number of bytes written in the size field.
//
// If WriteHeader has not been called before Write, it implicitly calls
// WriteHeader with [http.StatusOK], matching the behaviour of the standard
// library's response writer.
//
// After a successful write the body field is updated to reference b (the
// slice passed to this call). Note that body is replaced on every call, not
// appended; it therefore holds only the payload of the most recent Write.
//
// The method returns the number of bytes written and any error from the
// underlying writer.
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	w.body = b
	return n, err
}
