package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nakatanakatana/ff"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

//nolint:funlen
func TestHandlerInvalidRequest(t *testing.T) {
	t.Parallel()

	filtersMap := ff.CreateFiltersMap([]string{}, []string{})
	modifiersMap := ff.CreateModifierMap()
	handler := createHandler(filtersMap, modifiersMap)

	testCases := []struct {
		name             string
		setupMockServer  func() (*httptest.Server, func())
		requestURL       string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name: "URL parameter is required",
			setupMockServer: func() (*httptest.Server, func()) {
				return nil, func() {}
			},
			requestURL:       "/",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "must set URL",
		},
		{
			name: "Multiple URL parameters are not allowed",
			setupMockServer: func() (*httptest.Server, func()) {
				return nil, func() {}
			},
			requestURL:       "/?url=http://example.com&url=http://another.com",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "cannot set multiple URL",
		},
		{
			name: "Invalid feed URL should return BadRequest",
			setupMockServer: func() (*httptest.Server, func()) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte("Not Found"))
				}))

				return server, server.Close
			},
			requestURL:       "/?url=%s",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "404",
		},
		{
			name: "Invalid XML feed should return BadRequest",
			setupMockServer: func() (*httptest.Server, func()) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("This is not a valid XML feed"))
				}))

				return server, server.Close
			},
			requestURL:       "/?url=%s",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "ParseURL Error",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockServer, cleanup := tc.setupMockServer()
			defer cleanup()

			requestURL := tc.requestURL
			if mockServer != nil {
				requestURL = strings.ReplaceAll(requestURL, "%s", mockServer.URL)
			}

			req := httptest.NewRequest(http.MethodGet, requestURL, nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assert.Assert(t, cmp.Contains(rec.Body.String(), tc.expectedResponse))
		})
	}
}

func TestHandlerWithETagMiddleware_HEADRequest(t *testing.T) {
	t.Parallel()

	feedContent := `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title>RSS Title</title>
    <link>http://example.com</link>
    <description>This is an example RSS feed</description>
    <item>
      <title>Example entry</title>
      <link>http://example.com/1</link>
      <description>Example description</description>
    </item>
    <item>
      <title>Second entry</title>
      <link>http://example.com/2</link>
      <description>Second description</description>
    </item>
  </channel>
</rss>
` // Note: Trailing newline is intentional to match Fprintln and golden file.

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// The mock server should serve the exact content that ToRss() would generate,
		// because createHandler itself uses the ToRss() output.
		// So, the mock server's content should match the 'feedContent' above.
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8") // Match what real handler would do
		// The mock server should serve the exact content that ToRss() would generate,
		// because createHandler itself uses the ToRss() output.
		// So, the mock server's content should match the 'feedContent' above.
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8") // Match what real handler would do
		w.WriteHeader(http.StatusOK)
		// Use Fprintln to mimic the main handler's behavior regarding potential newlines.
		fmt.Fprintln(w, strings.TrimSuffix(feedContent, "\n")) // Trim suffix \n because Fprintln will add one
	}))
	defer mockServer.Close()

	filtersMap := ff.CreateFiltersMap([]string{}, []string{})
	modifiersMap := ff.CreateModifierMap()
	handler := createHandler(filtersMap, modifiersMap)
	wrappedHandler := etagMiddleware(handler)

	targetURL := "/?url=" + mockServer.URL

	// 1. GET Request to establish baseline and ETag
	reqGet := httptest.NewRequest(http.MethodGet, targetURL, nil)
	recGet := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(recGet, reqGet)

	assert.Equal(t, http.StatusOK, recGet.Code, "GET request: status code mismatch")
	assert.Assert(t, recGet.Body.String() != "", "GET request: body should not be empty")
	expectedContentType := "application/rss+xml; charset=utf-8"
	assert.Equal(t, expectedContentType, recGet.Header().Get("Content-Type"), "GET request: Content-Type mismatch")

	etagGet := recGet.Header().Get("ETag")
	assert.Assert(t, etagGet != "", "GET request: ETag should not be empty")

	// Verify GET body content. Since feedContent now includes a trailing newline (like Fprintln would add),
	// and recGet.Body.String() will also have it from the handler's Fprintln,
	// a direct string comparison should work.
	assert.Equal(t, feedContent, recGet.Body.String(), "GET request: body content mismatch")

	// 2. HEAD Request
	reqHead := httptest.NewRequest(http.MethodHead, targetURL, nil)
	recHead := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(recHead, reqHead)

	assert.Equal(t, http.StatusOK, recHead.Code, "HEAD request: status code mismatch")
	assert.Equal(t, "", recHead.Body.String(), "HEAD request: body must be empty")
	assert.Equal(t, expectedContentType, recHead.Header().Get("Content-Type"), "HEAD request: Content-Type mismatch")

	etagHead := recHead.Header().Get("ETag")
	assert.Assert(t, etagHead != "", "HEAD request: ETag should not be empty")
	assert.Equal(t, etagGet, etagHead, "HEAD request: ETag should match ETag from GET request")
}

//nolint:funlen
func TestHandlerSuccess(t *testing.T) {
	t.Parallel()

	filtersMap := ff.CreateFiltersMap([]string{}, []string{})
	modifiersMap := ff.CreateModifierMap()
	handler := createHandler(filtersMap, modifiersMap)

	testCases := []struct {
		name           string
		requestURL     string
		expectedStatus int
		goldenFile     string
	}{
		{
			name:           "Valid RSS feed should return OK",
			requestURL:     "/?url=%s",
			expectedStatus: http.StatusOK,
			goldenFile:     "valid-rss-feed",
		},
		{
			name:           "Valid RSS feed with filter query should return filtered feed",
			requestURL:     "/?url=%s&title.contains=Second",
			expectedStatus: http.StatusOK,
			goldenFile:     "filtered-rss-feed",
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
  <title>RSS Title</title>
  <link>http://example.com</link>
  <description>This is an example RSS feed</description>
  <item>
    <title>Example entry</title>
    <link>http://example.com/1</link>
    <description>Example description</description>
  </item>
  <item>
    <title>Second entry</title>
    <link>http://example.com/2</link>
    <description>Second description</description>
  </item>
</channel>
</rss>`))
			}))
			defer mockServer.Close()

			requestURL := strings.ReplaceAll(tc.requestURL, "%s", mockServer.URL)

			req := httptest.NewRequest(http.MethodGet, requestURL, nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			golden.Assert(t, rec.Body.String(), tc.goldenFile)
		})
	}
}
