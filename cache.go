package ff

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CacheMiddleware struct {
	tmpDir    string
	next      http.Handler
	etags     map[string]string
	etagMutex sync.RWMutex
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
		etags:  make(map[string]string),
	}, nil
}

func (c *CacheMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	cacheKey := c.getCacheKey(queries)
	cachePath := filepath.Join(c.tmpDir, cacheKey)

	if stat, err := os.Stat(cachePath); err == nil {
		upstreamURLs := queries["url"]
		if len(upstreamURLs) > 0 {
			if c.isCacheFresh(upstreamURLs[0], cacheKey, stat.ModTime()) {
				http.ServeFile(w, r, cachePath)
				return
			}
			os.Remove(cachePath)
			c.removeETag(cacheKey)
		} else {
			http.ServeFile(w, r, cachePath)
			return
		}
	}

	responseRecorder := &ResponseRecorder{
		ResponseWriter:  w,
		cachePath:       cachePath,
		cacheMiddleware: c,
		cacheKey:        cacheKey,
	}

	c.next.ServeHTTP(responseRecorder, r)

	upstreamURLs := queries["url"]
	if len(upstreamURLs) > 0 {
		responseRecorder.captureAndStoreETag(upstreamURLs[0])
	}
}

func (c *CacheMiddleware) getCacheKey(params url.Values) string {
	h := sha256.New()
	h.Write([]byte(params.Encode()))
	return fmt.Sprintf("%x.rss", h.Sum(nil))
}

type ResponseRecorder struct {
	http.ResponseWriter
	cachePath       string
	cacheMiddleware *CacheMiddleware
	cacheKey        string
}

func (c *CacheMiddleware) isCacheFresh(upstreamURL string, cacheKey string, cacheTime time.Time) bool {
	req, err := http.NewRequest("HEAD", upstreamURL, nil)
	if err != nil {
		return true
	}

	storedETag := c.getStoredETag(cacheKey)
	if storedETag != "" {
		req.Header.Set("If-None-Match", storedETag)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return true
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return true
	}

	currentETag := resp.Header.Get("ETag")
	if currentETag != "" && storedETag != "" {
		return currentETag == storedETag
	}

	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		if lastModTime, err := http.ParseTime(lastModified); err == nil {
			cacheTimeUTC := cacheTime.UTC()
			return !lastModTime.After(cacheTimeUTC)
		}
	}

	return false
}

func (c *CacheMiddleware) getStoredETag(cacheKey string) string {
	c.etagMutex.RLock()
	defer c.etagMutex.RUnlock()
	return c.etags[cacheKey]
}

func (c *CacheMiddleware) storeETag(cacheKey string, etag string) {
	if etag == "" {
		return
	}
	c.etagMutex.Lock()
	defer c.etagMutex.Unlock()
	c.etags[cacheKey] = etag
}

func (c *CacheMiddleware) removeETag(cacheKey string) {
	c.etagMutex.Lock()
	defer c.etagMutex.Unlock()
	delete(c.etags, cacheKey)
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	if err := os.WriteFile(r.cachePath, data, 0644); err != nil {
		return 0, fmt.Errorf("failed to write cache file: %w", err)
	}

	return r.ResponseWriter.Write(data)
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) captureAndStoreETag(upstreamURL string) {
	req, err := http.NewRequest("HEAD", upstreamURL, nil)
	if err != nil {
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if etag := resp.Header.Get("ETag"); etag != "" {
		r.cacheMiddleware.storeETag(r.cacheKey, etag)
	}
}
