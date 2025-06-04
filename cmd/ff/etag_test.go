package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

//nolint:funlen
func TestETagMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("Should add ETag header for successful GET requests", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test content"))
		})

		middleware := etagMiddleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test content", rec.Body.String())

		etag := rec.Header().Get("ETag")
		assert.Assert(t, etag != "", "ETag header should not be empty")

		expectedETag := "\"9473fdd0d880a43c21b7778d34872157\""
		assert.Equal(t, expectedETag, etag)
	})

	t.Run("Should not add ETag for specific non-GET/HEAD requests (e.g., POST)", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test content"))
		})

		middleware := etagMiddleware(handler)

		reqPost := httptest.NewRequest(http.MethodPost, "/", nil)
		recPost := httptest.NewRecorder()
		middleware.ServeHTTP(recPost, reqPost)

		assert.Equal(t, http.StatusOK, recPost.Code)
		assert.Equal(t, "test content", recPost.Body.String())
		assert.Equal(t, "", recPost.Header().Get("ETag"), "ETag header should be empty for POST requests")

		reqPut := httptest.NewRequest(http.MethodPut, "/", nil)
		recPut := httptest.NewRecorder()
		middleware.ServeHTTP(recPut, reqPut)

		assert.Equal(t, http.StatusOK, recPut.Code)
		assert.Equal(t, "test content", recPut.Body.String())
		assert.Equal(t, "", recPut.Header().Get("ETag"), "ETag header should be empty for PUT requests")
	})

	t.Run("Should not add ETag header for GET requests with non-200 responses", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		})

		middleware := etagMiddleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "not found", rec.Body.String())

		etag := rec.Header().Get("ETag")
		assert.Equal(t, "", etag, "ETag header should be empty for non-200 GET responses")
	})

	t.Run("Should handle HEAD requests correctly", func(t *testing.T) {
		t.Parallel()

		testContent := "test content for HEAD"
		expectedETagForHead := "\"359e4a2ceb40364ccde9fc7b56c93948\""

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(testContent))
		})
		middleware := etagMiddleware(handler)

		reqGet := httptest.NewRequest(http.MethodGet, "/", nil)
		recGet := httptest.NewRecorder()
		middleware.ServeHTTP(recGet, reqGet)

		assert.Equal(t, http.StatusOK, recGet.Code)
		assert.Equal(t, testContent, recGet.Body.String(), "Body for GET request did not match")
		etagGet := recGet.Header().Get("ETag")
		assert.Equal(t, expectedETagForHead, etagGet, "ETag for GET request was not as expected")
		assert.Equal(t, "text/plain", recGet.Header().Get("Content-Type"))

		reqHead := httptest.NewRequest(http.MethodHead, "/", nil)
		recHead := httptest.NewRecorder()
		middleware.ServeHTTP(recHead, reqHead)

		assert.Equal(t, http.StatusOK, recHead.Code, "Status code for HEAD should be OK")
		assert.Equal(t, "", recHead.Body.String(), "Body for HEAD request must be empty")
		etagHead := recHead.Header().Get("ETag")
		assert.Assert(t, etagHead != "", "ETag header should not be empty for HEAD requests")
		assert.Equal(t, etagGet, etagHead, "ETag for HEAD should match ETag for GET")
		assert.Equal(t, "text/plain", recHead.Header().Get("Content-Type"), "Content-Type for HEAD was not as expected")
	})

	t.Run("Should not add ETag for HEAD requests with non-200 responses", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		middleware := etagMiddleware(handler)

		req := httptest.NewRequest(http.MethodHead, "/", nil)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "", rec.Body.String(), "Body for HEAD request with error must be empty")
		assert.Equal(t, "", rec.Header().Get("ETag"), "ETag header should be empty for non-200 HEAD responses")
	})

	t.Run("Should not add ETag for HEAD requests with empty body response", func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		middleware := etagMiddleware(handler)

		req := httptest.NewRequest(http.MethodHead, "/", nil)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "", rec.Body.String(), "Body for HEAD request with empty handler response must be empty")
		assert.Equal(t, "", rec.Header().Get("ETag"), "ETag header should be empty for HEAD responses with no body content from handler")
	})

	t.Run("Should not add ETag for GET requests with empty body response", func(t *testing.T) {
		t.Parallel()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		middleware := etagMiddleware(handler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "", rec.Body.String())
		assert.Equal(t, "", rec.Header().Get("ETag"), "ETag should be empty for GET with no body")
	})
}
