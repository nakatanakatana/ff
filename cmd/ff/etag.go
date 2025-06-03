package main

import (
	"bytes"
	"crypto/md5"
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
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriterWithETag) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func etagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		wrapper := &responseWriterWithETag{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			code:           http.StatusOK,
		}

		next.ServeHTTP(wrapper, r)

		if wrapper.code == http.StatusOK {
			hash := md5.Sum(wrapper.body.Bytes())
			etag := hex.EncodeToString(hash[:])

			w.Header().Set("ETag", fmt.Sprintf("\"%s\"", etag))
		}
	})
}
