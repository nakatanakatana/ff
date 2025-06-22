package ff_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nakatanakatana/ff"
)

const testETagKey = "test-etag-key.rss"

func TestCacheMiddleware(t *testing.T) {
	t.Parallel()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	middleware, err := ff.NewCacheMiddleware(testHandler)
	if err != nil {
		t.Fatalf("Failed to create cache middleware: %v", err)
	}

	params := url.Values{}
	params.Set("url", "https://example.com/feed")
	params.Set("title.contains", "test")

	req := httptest.NewRequest(http.MethodGet, "/?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test response" {
		t.Errorf("Expected 'test response', got %s", w.Body.String())
	}

	// Check Content-Type header includes charset
	contentType := w.Header().Get("Content-Type")
	expectedContentType := "application/rss+xml; charset=utf-8"

	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type %q, got %q", expectedContentType, contentType)
	}

	cacheKey := middleware.GetCacheKey(params)
	cachePath := filepath.Join(middleware.TmpDir, cacheKey)

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("Cache file should exist after first request")
	}

	w2 := httptest.NewRecorder()
	middleware.ServeHTTP(w2, req)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected cached response status 200, got %d", w2.Code)
	}

	if !strings.Contains(w2.Body.String(), "test response") {
		t.Errorf("Expected cached response to contain 'test response', got %s", w2.Body.String())
	}

	// Check Content-Type header for cached response too
	contentType2 := w2.Header().Get("Content-Type")
	if contentType2 != expectedContentType {
		t.Errorf("Expected cached Content-Type %q, got %q", expectedContentType, contentType2)
	}

	os.Remove(cachePath)
}

func TestGetCacheKey(t *testing.T) {
	t.Parallel()

	middleware := &ff.CacheMiddleware{}

	params1 := url.Values{}
	params1.Set("url", "https://example.com/feed")
	params1.Set("title.contains", "test")

	params2 := url.Values{}
	params2.Set("url", "https://example.com/feed")
	params2.Set("title.contains", "different")

	key1 := middleware.GetCacheKey(params1)
	key2 := middleware.GetCacheKey(params2)

	if key1 == key2 {
		t.Error("Different parameters should generate different cache keys")
	}

	if !strings.HasSuffix(key1, ".rss") {
		t.Error("Cache key should have .rss extension")
	}
}

func TestIsCacheFresh(t *testing.T) {
	t.Parallel()

	middleware := &ff.CacheMiddleware{}

	// Set up times for testing (use UTC to avoid timezone issues)
	now := time.Now().UTC()
	cacheTime := now.Add(-1 * time.Hour)      // Cache was created 1 hour ago
	oldContentTime := now.Add(-2 * time.Hour) // Content is 2 hours old (older than cache)
	newContentTime := now                     // Content is fresh (newer than cache)

	// Test case 1: Last-Modified is older than cache time (cache is fresh)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Last-Modified", oldContentTime.Format(http.TimeFormat))
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	if !middleware.IsCacheFresh(context.Background(), server.URL, "test-key", cacheTime) {
		t.Errorf("Cache should be fresh when Last-Modified (%v) is older than cache time (%v)",
			oldContentTime, cacheTime)
	}

	// Test case 2: Last-Modified is newer than cache time (cache is stale)
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Last-Modified", newContentTime.Format(http.TimeFormat))
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server2.Close()

	if middleware.IsCacheFresh(context.Background(), server2.URL, "test-key", cacheTime) {
		t.Errorf("Cache should be stale when Last-Modified (%v) is newer than cache time (%v)",
			newContentTime, cacheTime)
	}
}

func TestCacheETagStorage(t *testing.T) {
	t.Parallel()

	middleware, err := ff.NewCacheMiddleware(nil)
	if err != nil {
		t.Fatalf("Failed to create cache middleware: %v", err)
	}

	cacheKey := testETagKey

	// Test storing and retrieving ETag
	testETag := `"abc123"`
	middleware.StoreETag(cacheKey, testETag)

	retrievedETag := middleware.GetStoredETag(cacheKey)
	if retrievedETag != testETag {
		t.Errorf("Expected ETag %s, got %s", testETag, retrievedETag)
	}

	// Clean up
	middleware.RemoveETag(cacheKey)
}

func TestCacheETagValidation(t *testing.T) {
	t.Parallel()

	middleware, err := ff.NewCacheMiddleware(nil)
	if err != nil {
		t.Fatalf("Failed to create cache middleware: %v", err)
	}

	cacheKey := testETagKey
	cacheTime := time.Now().UTC().Add(-1 * time.Hour)

	// Test ETag-based cache validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("ETag", `"abc123"`)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Store the ETag first
	middleware.StoreETag(cacheKey, `"abc123"`)

	if !middleware.IsCacheFresh(context.Background(), server.URL, cacheKey, cacheTime) {
		t.Error("Cache should be fresh when ETag matches")
	}

	// Clean up
	middleware.RemoveETag(cacheKey)
}

func TestCacheETagMismatch(t *testing.T) {
	t.Parallel()

	middleware, err := ff.NewCacheMiddleware(nil)
	if err != nil {
		t.Fatalf("Failed to create cache middleware: %v", err)
	}

	cacheKey := testETagKey
	cacheTime := time.Now().UTC().Add(-1 * time.Hour)

	// Store initial ETag
	middleware.StoreETag(cacheKey, `"abc123"`)

	// Test with different ETag
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("ETag", `"xyz789"`)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	if middleware.IsCacheFresh(context.Background(), server.URL, cacheKey, cacheTime) {
		t.Error("Cache should be stale when ETag is different")
	}

	// Clean up
	middleware.RemoveETag(cacheKey)
}

func TestCacheETagNotModified(t *testing.T) {
	t.Parallel()

	middleware, err := ff.NewCacheMiddleware(nil)
	if err != nil {
		t.Fatalf("Failed to create cache middleware: %v", err)
	}

	cacheKey := testETagKey
	cacheTime := time.Now().UTC().Add(-1 * time.Hour)

	// Store the ETag first
	middleware.StoreETag(cacheKey, `"abc123"`)

	// Test 304 Not Modified response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			if r.Header.Get("If-None-Match") == `"abc123"` {
				w.WriteHeader(http.StatusNotModified)
			} else {
				w.Header().Set("ETag", `"abc123"`)
				w.WriteHeader(http.StatusOK)
			}
		}
	}))
	defer server.Close()

	if !middleware.IsCacheFresh(context.Background(), server.URL, cacheKey, cacheTime) {
		t.Error("Cache should be fresh when server returns 304 Not Modified")
	}

	// Clean up
	middleware.RemoveETag(cacheKey)
}
