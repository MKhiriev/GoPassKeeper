package http

import "net/http"

type responseData struct {
	status int
	size   int
	body   []byte
}

// responseWriter
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	size        int
	body        []byte
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.status = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	w.body = b
	return n, err
}
