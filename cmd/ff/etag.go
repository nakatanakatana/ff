package main

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"fmt"
	"net/http"
)

type responseWriterWithETag struct {
	http.ResponseWriter
	body *bytes.Buffer
	code int
}

func (w *responseWriterWithETag) Write(b []byte) (int, error) {
	// Only write to the buffer, not to the underlying ResponseWriter yet.
	// The actual response will be written in etagMiddleware after ETag calculation.
	return w.body.Write(b)
}

func (w *responseWriterWithETag) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func etagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			next.ServeHTTP(w, r)

			return
		}

		wrapper := &responseWriterWithETag{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			code:           http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		// Calculate and set ETag only if status is OK and there's a body
		if wrapper.code == http.StatusOK && wrapper.body.Len() > 0 {
			hash := md5.Sum(wrapper.body.Bytes()) //nolint:gosec
			etag := hex.EncodeToString(hash[:])
			w.Header().Set("ETag", fmt.Sprintf("\"%s\"", etag))
		}

		// For GET requests, write the buffered body to the actual ResponseWriter.
		// The WriteHeader method on the wrapper has already passed the status code.
		// For HEAD requests, do not write the body. Content-Length should be handled
		// correctly by the HTTP server/handler when no body is written.
		if r.Method == http.MethodGet {
			// Only write the body if it has content.
			// This ensures that if a handler correctly sets Content-Length=0 for a GET,
			// we don't accidentally write an empty buffer if body.Len() was 0.
			// However, wrapper.body will contain what was written by next.ServeHTTP.
			// If next.ServeHTTP wrote nothing, body.Len() will be 0.
			if wrapper.body.Len() > 0 {
				w.Write(wrapper.body.Bytes()) //nolint:errcheck
			}
		}
		// For HEAD requests, the response headers (including Content-Length)
		// should have been set by the handler called by next.ServeHTTP.
		// The Go HTTP server typically sets Content-Length for HEAD requests
		// based on what a GET request would have produced, if the handler
		// doesn't explicitly set it. Our wrapper.WriteHeader ensures the
		// status code is propagated.
	})
}
