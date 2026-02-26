package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// gzipWriterPool is a pool of reusable [gzip.Writer] instances.
//
// Pooling writers avoids repeated heap allocations for each compressed
// response. Each writer is reset to a new underlying [io.Writer] via
// [gzip.Writer.Reset] before use and returned to the pool after the
// response body has been flushed and closed.
var gzipWriterPool = sync.Pool{
	New: func() any {
		w := gzip.NewWriter(nil)
		return w
	},
}

// gzipReaderPool is a pool of reusable [gzip.Reader] instances.
//
// Pooling readers avoids repeated heap allocations for each compressed
// request body. Each reader is reset to the incoming request body via
// [gzip.Reader.Reset] before use and returned to the pool once the
// request body has been fully consumed and closed.
var gzipReaderPool = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

// withGZip is an HTTP middleware that provides transparent gzip
// compression and decompression for both request bodies and response bodies.
//
// Request decompression: if the incoming request carries a
// "Content-Encoding: gzip" header and a non-nil body, the body is
// transparently decompressed using a pooled [gzip.Reader]. The
// "Content-Encoding" header is removed from the request before it is
// forwarded so that downstream handlers see plain data. If the body
// contains invalid gzip data, the middleware responds with
// HTTP 400 Bad Request and does not call next.
//
// Response compression: if the client advertises gzip support via the
// "Accept-Encoding: gzip" header, the response body is compressed on
// the fly using a pooled [gzip.Writer] wrapped in a [gzipResponseWriter].
// The "Content-Encoding: gzip" header is set automatically when the
// first WriteHeader call is made. If the client does not advertise gzip
// support, the response is passed through to next unchanged.
//
// Both the writer and reader pools are used to minimise allocations under
// concurrent load.
func withGZip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		contentEncoding := req.Header.Get("Content-Encoding")
		isGzipRequest := strings.Contains(contentEncoding, "gzip")

		// Decompress the request body if the client sent it gzip-encoded.
		if isGzipRequest && req.Body != nil {
			gzipReader := gzipReaderPool.Get().(*gzip.Reader)
			if err := gzipReader.Reset(req.Body); err != nil {
				gzipReaderPool.Put(gzipReader)
				http.Error(w, "Invalid gzip data", http.StatusBadRequest)
				return
			}

			// Wrap the gzip reader so that closing the body also closes the
			// underlying gzip stream and returns the reader to the pool.
			req.Body = &wrappedReadCloser{
				Reader: gzipReader,
				OnClose: func() {
					gzipReader.Close()
					gzipReaderPool.Put(gzipReader)
				},
			}
			// Remove the header so downstream handlers treat the body as plain data.
			req.Header.Del("Content-Encoding")
		}

		// If the client does not support gzip, skip response compression.
		if !supportsGzip {
			next.ServeHTTP(w, req)
			return
		}

		// Acquire a pooled gzip writer, reset it to the real ResponseWriter,
		// and wrap the ResponseWriter so that all writes go through compression.
		gzipWriter := gzipWriterPool.Get().(*gzip.Writer)

		gzipRW := &gzipResponseWriter{
			ResponseWriter: w,
			gzipWriter:     gzipWriter,
		}

		gzipWriter.Reset(w)

		next.ServeHTTP(gzipRW, req)

		// Flush and close the gzip stream, then return the writer to the pool.
		gzipWriter.Close()
		gzipWriterPool.Put(gzipWriter)
	})
}

// wrappedReadCloser combines an [io.Reader] with a custom close callback.
//
// It is used to wrap a pooled [gzip.Reader] so that calling Close on the
// request body both closes the underlying gzip stream and returns the
// reader to [gzipReaderPool], preventing resource leaks.
type wrappedReadCloser struct {
	io.Reader

	// OnClose is called once when Close is invoked. It is responsible for
	// closing the underlying gzip reader and returning it to the pool.
	OnClose func()
}

// Close invokes the OnClose callback if one is set, and always returns nil.
// It satisfies the [io.ReadCloser] interface required by [http.Request.Body].
func (w *wrappedReadCloser) Close() error {
	if w.OnClose != nil {
		w.OnClose()
	}
	return nil
}

// gzipResponseWriter is an [http.ResponseWriter] decorator that compresses
// the response body using the provided [gzip.Writer].
//
// It intercepts WriteHeader to inject the "Content-Encoding: gzip" response
// header, and redirects Write calls to the gzip writer so that all response
// data is compressed transparently. The underlying [http.ResponseWriter]
// is still used for header management and status code propagation.
type gzipResponseWriter struct {
	http.ResponseWriter

	// gzipWriter is the pooled compressor to which response bytes are written.
	gzipWriter *gzip.Writer
}

// WriteHeader sets the "Content-Encoding: gzip" header on the response
// and then delegates to the underlying [http.ResponseWriter.WriteHeader].
//
// It must be called before any call to Write, or the header will be
// sent implicitly with HTTP 200 OK on the first Write.
func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.Header().Set("Content-Encoding", "gzip")
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write compresses data and writes it to the underlying gzip stream.
// It satisfies the [io.Writer] interface and is called by the HTTP server
// when the handler writes its response body.
//
// If WriteHeader has not been called before Write, the status code defaults
// to HTTP 200 OK and the "Content-Encoding: gzip" header is set at that point.
func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.gzipWriter.Write(data)
}

// Close flushes any buffered compressed data and closes the gzip stream.
// It must be called after the handler returns to ensure all data is flushed
// to the underlying [http.ResponseWriter] before the connection is finalised.
func (w *gzipResponseWriter) Close() error {
	return w.gzipWriter.Close()
}
