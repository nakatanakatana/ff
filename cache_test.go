package ff

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCacheMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware, err := NewCacheMiddleware(testHandler)
	if err != nil {
		t.Fatalf("Failed to create cache middleware: %v", err)
	}

	params := url.Values{}
	params.Set("url", "https://example.com/feed")
	params.Set("title.contains", "test")

	req := httptest.NewRequest("GET", "/?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test response" {
		t.Errorf("Expected 'test response', got %s", w.Body.String())
	}

	cacheKey := middleware.getCacheKey(params)
	cachePath := filepath.Join(middleware.tmpDir, cacheKey)

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

	os.Remove(cachePath)
}

func TestGetCacheKey(t *testing.T) {
	middleware := &CacheMiddleware{}

	params1 := url.Values{}
	params1.Set("url", "https://example.com/feed")
	params1.Set("title.contains", "test")

	params2 := url.Values{}
	params2.Set("url", "https://example.com/feed")
	params2.Set("title.contains", "different")

	key1 := middleware.getCacheKey(params1)
	key2 := middleware.getCacheKey(params2)

	if key1 == key2 {
		t.Error("Different parameters should generate different cache keys")
	}

	if !strings.HasSuffix(key1, ".rss") {
		t.Error("Cache key should have .rss extension")
	}
}

func TestIsCacheFresh(t *testing.T) {
	middleware := &CacheMiddleware{}

	// Set up times for testing (use UTC to avoid timezone issues)
	now := time.Now().UTC()
	cacheTime := now.Add(-1 * time.Hour)      // Cache was created 1 hour ago
	oldContentTime := now.Add(-2 * time.Hour) // Content is 2 hours old (older than cache)
	newContentTime := now                     // Content is fresh (newer than cache)

	// Test case 1: Last-Modified is older than cache time (cache is fresh)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Last-Modified", oldContentTime.Format(http.TimeFormat))
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	if !middleware.isCacheFresh(server.URL, "test-key", cacheTime) {
		t.Errorf("Cache should be fresh when Last-Modified (%v) is older than cache time (%v)",
			oldContentTime, cacheTime)
	}

	// Test case 2: Last-Modified is newer than cache time (cache is stale)
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Last-Modified", newContentTime.Format(http.TimeFormat))
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server2.Close()

	if middleware.isCacheFresh(server2.URL, "test-key", cacheTime) {
		t.Errorf("Cache should be stale when Last-Modified (%v) is newer than cache time (%v)",
			newContentTime, cacheTime)
	}
}

func TestCacheETag(t *testing.T) {
	middleware, err := NewCacheMiddleware(nil)
	if err != nil {
		t.Fatalf("Failed to create cache middleware: %v", err)
	}

	cacheKey := "test-etag-key.rss"

	// Test storing and retrieving ETag
	testETag := `"abc123"`
	middleware.storeETag(cacheKey, testETag)

	retrievedETag := middleware.getStoredETag(cacheKey)
	if retrievedETag != testETag {
		t.Errorf("Expected ETag %s, got %s", testETag, retrievedETag)
	}

	// Clean up
	middleware.removeETag(cacheKey)

	// Test ETag-based cache validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("ETag", `"abc123"`)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Store the ETag first
	middleware.storeETag(cacheKey, `"abc123"`)

	cacheTime := time.Now().UTC().Add(-1 * time.Hour)

	if !middleware.isCacheFresh(server.URL, cacheKey, cacheTime) {
		t.Error("Cache should be fresh when ETag matches")
	}

	// Test with different ETag
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("ETag", `"xyz789"`)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server2.Close()

	if middleware.isCacheFresh(server2.URL, cacheKey, cacheTime) {
		t.Error("Cache should be stale when ETag is different")
	}

	// Test 304 Not Modified response
	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			if r.Header.Get("If-None-Match") == `"abc123"` {
				w.WriteHeader(http.StatusNotModified)
			} else {
				w.Header().Set("ETag", `"abc123"`)
				w.WriteHeader(http.StatusOK)
			}
		}
	}))
	defer server3.Close()

	if !middleware.isCacheFresh(server3.URL, cacheKey, cacheTime) {
		t.Error("Cache should be fresh when server returns 304 Not Modified")
	}

	// Clean up
	middleware.removeETag(cacheKey)
}
