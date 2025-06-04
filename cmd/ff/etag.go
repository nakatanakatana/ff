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
	i, err := w.body.Write(b)
	if err != nil {
		return i, fmt.Errorf("body.Write Error: %w", err)
	}

	return i, nil
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

		if wrapper.code == http.StatusOK && wrapper.body.Len() > 0 {
			hash := md5.Sum(wrapper.body.Bytes()) //nolint:gosec
			etag := hex.EncodeToString(hash[:])
			w.Header().Set("ETag", fmt.Sprintf("\"%s\"", etag))
		}

		if r.Method == http.MethodGet {
			if wrapper.body.Len() > 0 {
				w.Write(wrapper.body.Bytes()) //nolint:errcheck
			}
		}
	})
}
