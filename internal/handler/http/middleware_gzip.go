package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		w := gzip.NewWriter(nil)
		return w
	},
}

var gzipReaderPool = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

func withGZip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		contentEncoding := req.Header.Get("Content-Encoding")
		isGzipRequest := strings.Contains(contentEncoding, "gzip")

		if isGzipRequest && req.Body != nil {
			gzipReader := gzipReaderPool.Get().(*gzip.Reader)
			if err := gzipReader.Reset(req.Body); err != nil {
				gzipReaderPool.Put(gzipReader)
				http.Error(w, "Invalid gzip data", http.StatusBadRequest)
				return
			}

			req.Body = &wrappedReadCloser{
				Reader: gzipReader,
				OnClose: func() {
					gzipReader.Close()
					gzipReaderPool.Put(gzipReader)
				},
			}
			req.Header.Del("Content-Encoding")
		}

		if !supportsGzip {
			next.ServeHTTP(w, req)
			return
		}

		gzipWriter := gzipWriterPool.Get().(*gzip.Writer)

		gzipRW := &gzipResponseWriter{
			ResponseWriter: w,
			gzipWriter:     gzipWriter,
		}

		gzipWriter.Reset(w)

		next.ServeHTTP(gzipRW, req)

		gzipWriter.Close()
		gzipWriterPool.Put(gzipWriter)
	})
}

type wrappedReadCloser struct {
	io.Reader
	OnClose func()
}

func (w *wrappedReadCloser) Close() error {
	if w.OnClose != nil {
		w.OnClose()
	}
	return nil
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.Header().Set("Content-Encoding", "gzip")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.gzipWriter.Write(data)
}

func (w *gzipResponseWriter) Close() error {
	return w.gzipWriter.Close()
}
