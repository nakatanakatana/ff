package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

func TestETagMiddleware(t *testing.T) {
	t.Run("Should add ETag header for successful GET requests", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test content"))
		})

		middleware := etagMiddleware(handler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test content", rec.Body.String())

		etag := rec.Header().Get("ETag")
		assert.Assert(t, etag != "", "ETag header should not be empty")
		expectedETag := "\"9473fdd0d880a43c21b7778d34872157\""
		assert.Equal(t, expectedETag, etag)
	})

	t.Run("Should not add ETag header for non-GET requests", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test content"))
		})

		middleware := etagMiddleware(handler)

		req := httptest.NewRequest("POST", "/", nil)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test content", rec.Body.String())

		etag := rec.Header().Get("ETag")
		assert.Equal(t, "", etag, "ETag header should be empty for non-GET requests")
	})

	t.Run("Should not add ETag header for non-200 responses", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		})

		middleware := etagMiddleware(handler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "not found", rec.Body.String())

		etag := rec.Header().Get("ETag")
		assert.Equal(t, "", etag, "ETag header should be empty for non-200 responses")
	})
}
