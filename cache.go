package ff

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	httpMethodHead = "HEAD"
	requestTimeout = 10 * time.Second
	filePerms      = 0o600
	dirPerms       = 0o755
)

type CacheMiddleware struct {
	TmpDir    string
	next      http.Handler
	etags     map[string]string
	etagMutex sync.RWMutex
	fsys      fs.FS
}

func NewCacheMiddleware(next http.Handler) (*CacheMiddleware, error) {
	tmpDir := os.TempDir()
	cacheDir := filepath.Join(tmpDir, "ff-cache")

	if err := os.MkdirAll(cacheDir, dirPerms); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &CacheMiddleware{
		TmpDir: cacheDir,
		next:   next,
		etags:  make(map[string]string),
		fsys:   os.DirFS(cacheDir),
	}, nil
}

func (c *CacheMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	cacheKey := c.GetCacheKey(queries)
	cachePath := filepath.Join(c.TmpDir, cacheKey)

	if stat, err := os.Stat(cachePath); err == nil {
		upstreamURLs := queries["url"]
		if len(upstreamURLs) > 0 {
			if c.IsCacheFresh(r.Context(), upstreamURLs[0], cacheKey, stat.ModTime()) {
				http.ServeFileFS(w, r, c.fsys, cacheKey)

				return
			}

			os.Remove(cachePath)
			c.RemoveETag(cacheKey)
		} else {
			http.ServeFileFS(w, r, c.fsys, cacheKey)

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
		responseRecorder.captureAndStoreETag(r.Context(), upstreamURLs[0])
	}
}

func (c *CacheMiddleware) GetCacheKey(params url.Values) string {
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

func (c *CacheMiddleware) IsCacheFresh(
	ctx context.Context, upstreamURL string, cacheKey string, cacheTime time.Time,
) bool {
	req, err := http.NewRequestWithContext(ctx, httpMethodHead, upstreamURL, nil)
	if err != nil {
		return true
	}

	storedETag := c.GetStoredETag(cacheKey)
	if storedETag != "" {
		req.Header.Set("If-None-Match", storedETag)
	}

	client := &http.Client{
		Timeout: requestTimeout,
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

func (c *CacheMiddleware) GetStoredETag(cacheKey string) string {
	c.etagMutex.RLock()
	defer c.etagMutex.RUnlock()

	return c.etags[cacheKey]
}

func (c *CacheMiddleware) StoreETag(cacheKey string, etag string) {
	if etag == "" {
		return
	}

	c.etagMutex.Lock()
	defer c.etagMutex.Unlock()
	c.etags[cacheKey] = etag
}

func (c *CacheMiddleware) RemoveETag(cacheKey string) {
	c.etagMutex.Lock()
	defer c.etagMutex.Unlock()
	delete(c.etags, cacheKey)
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	if err := os.WriteFile(r.cachePath, data, filePerms); err != nil {
		return 0, fmt.Errorf("failed to write cache file: %w", err)
	}

	n, err := r.ResponseWriter.Write(data)
	if err != nil {
		return n, fmt.Errorf("failed to write response: %w", err)
	}

	return n, nil
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) captureAndStoreETag(ctx context.Context, upstreamURL string) {
	req, err := http.NewRequestWithContext(ctx, httpMethodHead, upstreamURL, nil)
	if err != nil {
		return
	}

	client := &http.Client{
		Timeout: requestTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if etag := resp.Header.Get("ETag"); etag != "" {
		r.cacheMiddleware.StoreETag(r.cacheKey, etag)
	}
}
