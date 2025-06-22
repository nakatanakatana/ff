package ff

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
