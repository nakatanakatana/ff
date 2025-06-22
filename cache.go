package ff

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type CacheMiddleware struct {
	tmpDir string
	next   http.Handler
}

func NewCacheMiddleware(next http.Handler) (*CacheMiddleware, error) {
	tmpDir := os.TempDir()
	cacheDir := filepath.Join(tmpDir, "ff-cache")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &CacheMiddleware{
		tmpDir: cacheDir,
		next:   next,
	}, nil
}

func (c *CacheMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cacheKey := c.getCacheKey(r.URL.Query())
	cachePath := filepath.Join(c.tmpDir, cacheKey)

	if _, err := os.Stat(cachePath); err == nil {
		http.ServeFile(w, r, cachePath)
		return
	}

	responseRecorder := &ResponseRecorder{
		ResponseWriter: w,
		cachePath:      cachePath,
	}

	c.next.ServeHTTP(responseRecorder, r)
}

func (c *CacheMiddleware) getCacheKey(params url.Values) string {
	h := sha256.New()
	h.Write([]byte(params.Encode()))
	return fmt.Sprintf("%x.rss", h.Sum(nil))
}

type ResponseRecorder struct {
	http.ResponseWriter
	cachePath string
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	if err := os.WriteFile(r.cachePath, data, 0644); err != nil {
		return 0, fmt.Errorf("failed to write cache file: %w", err)
	}

	return r.ResponseWriter.Write(data)
}
