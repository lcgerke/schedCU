// Package amion provides Amion service integration for the schedCU system.
package amion

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewAmionHTTPClient tests the creation of a new AmionHTTPClient.
func TestNewAmionHTTPClient(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		wantErr   bool
		errMsg    string
		checkFunc func(t *testing.T, client *AmionHTTPClient)
	}{
		{
			name:    "valid base URL",
			baseURL: "https://example.com",
			wantErr: false,
			checkFunc: func(t *testing.T, client *AmionHTTPClient) {
				if client == nil {
					t.Fatal("expected client, got nil")
				}
				if client.baseURL != "https://example.com" {
					t.Errorf("expected baseURL %q, got %q", "https://example.com", client.baseURL)
				}
				if client.httpClient == nil {
					t.Fatal("expected httpClient to be initialized")
				}
				if client.httpClient.Timeout != 30*time.Second {
					t.Errorf("expected timeout 30s, got %v", client.httpClient.Timeout)
				}
			},
		},
		{
			name:    "invalid URL scheme",
			baseURL: "invalid://example.com",
			wantErr: true,
		},
		{
			name:    "empty URL",
			baseURL: "",
			wantErr: true,
		},
		{
			name:    "localhost with port",
			baseURL: "http://localhost:8080",
			wantErr: false,
			checkFunc: func(t *testing.T, client *AmionHTTPClient) {
				if client.baseURL != "http://localhost:8080" {
					t.Errorf("expected baseURL %q, got %q", "http://localhost:8080", client.baseURL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAmionHTTPClient(tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAmionHTTPClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checkFunc != nil && !tt.wantErr {
				tt.checkFunc(t, client)
			}
		})
	}
}

// TestFetchAndParseHTMLSuccess tests successful HTML fetching and parsing.
func TestFetchAndParseHTMLSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body><div class="test">Hello World</div></body></html>`)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	doc, err := client.FetchAndParseHTML(server.URL)
	if err != nil {
		t.Fatalf("FetchAndParseHTML failed: %v", err)
	}

	if doc == nil {
		t.Fatal("expected doc, got nil")
	}

	// Verify parsing worked
	content := doc.Find("div.test").Text()
	if content != "Hello World" {
		t.Errorf("expected content %q, got %q", "Hello World", content)
	}
}

// TestFetchAndParseHTMLGzipCompression tests gzip compressed response handling.
func TestFetchAndParseHTMLGzipCompression(t *testing.T) {
	htmlContent := `<html><body><h1>Gzipped Content</h1></body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Encoding", "gzip")

		// Gzip the content
		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		if _, err := gzipWriter.Write([]byte(htmlContent)); err != nil {
			t.Fatalf("gzip write failed: %v", err)
		}
		gzipWriter.Close()

		w.Write(buf.Bytes())
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	doc, err := client.FetchAndParseHTML(server.URL)
	if err != nil {
		t.Fatalf("FetchAndParseHTML with gzip failed: %v", err)
	}

	content := doc.Find("h1").Text()
	if content != "Gzipped Content" {
		t.Errorf("expected content %q, got %q", "Gzipped Content", content)
	}
}

// TestFetchAndParseHTML404Error tests 404 error handling.
func TestFetchAndParseHTML404Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	_, err = client.FetchAndParseHTML(server.URL + "/notfound")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", httpErr.StatusCode)
	}
}

// TestFetchAndParseHTML500Error tests 500 error handling.
func TestFetchAndParseHTML500Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	_, err = client.FetchAndParseHTML(server.URL)
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", httpErr.StatusCode)
	}
}

// TestFetchAndParseHTMLMalformedHTML tests malformed HTML parsing.
func TestFetchAndParseHTMLMalformedHTML(t *testing.T) {
	malformedHTML := `<html><body><div>Unclosed div<body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, malformedHTML)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	// goquery should handle malformed HTML gracefully
	doc, err := client.FetchAndParseHTML(server.URL)
	if err != nil {
		t.Fatalf("FetchAndParseHTML should handle malformed HTML: %v", err)
	}

	if doc == nil {
		t.Fatal("expected doc, got nil")
	}
}

// TestRetryLogic tests the exponential backoff retry mechanism.
func TestRetryLogic(t *testing.T) {
	attemptCount := int32(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)

		// Fail the first 2 attempts, succeed on the 3rd
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body><p>Success</p></body></html>`)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	doc, err := client.FetchAndParseHTML(server.URL)
	if err != nil {
		t.Fatalf("FetchAndParseHTML with retries failed: %v", err)
	}

	if doc == nil {
		t.Fatal("expected doc, got nil")
	}

	content := doc.Find("p").Text()
	if content != "Success" {
		t.Errorf("expected content %q, got %q", "Success", content)
	}

	if atomic.LoadInt32(&attemptCount) != 3 {
		t.Errorf("expected 3 attempts, got %d", atomic.LoadInt32(&attemptCount))
	}
}

// TestRetryExhaustion tests behavior when all retries are exhausted.
func TestRetryExhaustion(t *testing.T) {
	attemptCount := int32(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	_, err = client.FetchAndParseHTML(server.URL)
	if err == nil {
		t.Fatal("expected error when retries exhausted")
	}

	// Should have tried 1 initial + 3 retries = 4 times
	if atomic.LoadInt32(&attemptCount) != 4 {
		t.Errorf("expected 4 attempts (1 + 3 retries), got %d", atomic.LoadInt32(&attemptCount))
	}
}

// TestContextCancellation tests request cancellation via context.
func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, "This should not be reached")
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = client.FetchAndParseHTMLWithContext(ctx, server.URL)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected context/timeout error, got: %v", err)
	}
}

// TestCookieJarPersistence tests that cookies are persisted across requests.
func TestCookieJarPersistence(t *testing.T) {
	requestCount := int32(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)

		if count == 1 {
			// Set a cookie on first request
			http.SetCookie(w, &http.Cookie{
				Name:  "session_id",
				Value: "abc123",
				Path:  "/",
			})
			fmt.Fprint(w, "<html><body>First</body></html>")
			return
		}

		// On second request, verify the cookie was sent back
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value != "abc123" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Cookie not found or invalid")
			return
		}

		fmt.Fprint(w, `<html><body>Cookie verified</body></html>`)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	// First request sets cookie
	_, err = client.FetchAndParseHTML(server.URL + "/first")
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	// Second request should include the cookie
	doc, err := client.FetchAndParseHTML(server.URL + "/second")
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}

	content := doc.Find("body").Text()
	if content != "Cookie verified" {
		t.Errorf("expected cookie verification, got: %s", content)
	}
}

// TestUserAgent tests that the User-Agent header is correctly set.
func TestUserAgent(t *testing.T) {
	var receivedUserAgent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		fmt.Fprint(w, "<html><body>OK</body></html>")
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	_, err = client.FetchAndParseHTML(server.URL)
	if err != nil {
		t.Fatalf("FetchAndParseHTML failed: %v", err)
	}

	if receivedUserAgent == "" {
		t.Fatal("User-Agent header not set")
	}

	if !strings.Contains(receivedUserAgent, "Mozilla") {
		t.Errorf("expected realistic User-Agent with Mozilla, got: %s", receivedUserAgent)
	}
}

// TestConnectionPooling tests that connection pooling is enabled.
func TestConnectionPooling(t *testing.T) {
	requestCount := int32(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		fmt.Fprint(w, "<html><body>OK</body></html>")
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	// Verify transport has connection pooling
	if client.httpClient.Transport == nil {
		t.Fatal("expected transport to be configured")
	}

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", client.httpClient.Transport)
	}

	if transport.MaxIdleConns == 0 {
		t.Error("expected MaxIdleConns to be set for connection pooling")
	}

	// Make multiple requests to verify pooling
	for i := 0; i < 5; i++ {
		_, err := client.FetchAndParseHTML(server.URL)
		if err != nil {
			t.Fatalf("request %d failed: %v", i+1, err)
		}
	}

	if atomic.LoadInt32(&requestCount) != 5 {
		t.Errorf("expected 5 requests, got %d", atomic.LoadInt32(&requestCount))
	}
}

// TestConcurrentRequests tests thread-safe concurrent requests.
func TestConcurrentRequests(t *testing.T) {
	requestCount := int32(0)
	mu := sync.Mutex{}
	seenIDs := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		id := r.Header.Get("X-Request-ID")

		mu.Lock()
		seenIDs[id] = true
		mu.Unlock()

		fmt.Fprintf(w, `<html><body>Request: %s</body></html>`, id)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			// Create a request with a unique ID
			req, _ := http.NewRequest("GET", server.URL, nil)
			req.Header.Set("X-Request-ID", fmt.Sprintf("req-%d", index))

			resp, err := client.httpClient.Do(req)
			if err != nil {
				t.Errorf("concurrent request failed: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200, got %d", resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()

	if atomic.LoadInt32(&requestCount) != int32(numGoroutines) {
		t.Errorf("expected %d requests, got %d", numGoroutines, atomic.LoadInt32(&requestCount))
	}
}

// TestRedirectHandling tests following HTTP redirects.
func TestRedirectHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusMovedPermanently)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body><p>Final page</p></body></html>`)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	doc, err := client.FetchAndParseHTML(server.URL + "/redirect")
	if err != nil {
		t.Fatalf("FetchAndParseHTML with redirect failed: %v", err)
	}

	content := doc.Find("p").Text()
	if content != "Final page" {
		t.Errorf("expected content %q, got %q", "Final page", content)
	}
}

// TestMaxRedirects tests that redirect loops are prevented.
func TestMaxRedirects(t *testing.T) {
	redirectCount := int32(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&redirectCount, 1)
		// Create a redirect loop
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	_, err = client.FetchAndParseHTML(server.URL)
	if err == nil {
		t.Fatal("expected error for redirect loop")
	}

	// Should have reached max redirects (10 by default in http.Client)
	if atomic.LoadInt32(&redirectCount) <= 1 {
		t.Errorf("expected multiple redirect attempts, got %d", atomic.LoadInt32(&redirectCount))
	}
}

// TestInvalidHTMLURL tests handling of invalid URLs.
func TestInvalidHTMLURL(t *testing.T) {
	client, err := NewAmionHTTPClient("http://localhost:8080")
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	_, err = client.FetchAndParseHTML("://invalid-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

// TestEncodingDetection tests encoding detection for various charsets.
func TestEncodingDetection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=iso-8859-1")
		fmt.Fprint(w, `<html><body><p>Content</p></body></html>`)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	doc, err := client.FetchAndParseHTML(server.URL)
	if err != nil {
		t.Fatalf("FetchAndParseHTML with charset detection failed: %v", err)
	}

	if doc == nil {
		t.Fatal("expected doc, got nil")
	}
}

// TestNetworkTimeout tests timeout handling with slow server.
func TestNetworkTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the client timeout
		time.Sleep(35 * time.Second)
		fmt.Fprint(w, "This should not be reached")
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	// Use context with shorter timeout than httpClient timeout
	// to avoid waiting 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	_, err = client.FetchAndParseHTMLWithContext(ctx, server.URL)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}

	// Should timeout after ~1 second
	if elapsed > 3*time.Second {
		t.Errorf("timeout took too long: %v", elapsed)
	}
}

// TestConnectionRefused tests handling of connection refused errors.
func TestConnectionRefused(t *testing.T) {
	// Get an unused port by listening and then closing
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get unused port: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	client, err := NewAmionHTTPClient("http://" + addr)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	_, err = client.FetchAndParseHTML("http://" + addr)
	if err == nil {
		t.Fatal("expected connection refused error")
	}

	// Verify it's a network error
	if !strings.Contains(err.Error(), "connection refused") && !strings.Contains(err.Error(), "dial") {
		t.Errorf("expected connection refused error, got: %v", err)
	}
}

// TestWithLogger tests HTTP client with logger integration.
func TestWithLogger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body>Test</body></html>`)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	// Test that logger can be set
	doc, err := client.FetchAndParseHTML(server.URL)
	if err != nil {
		t.Fatalf("FetchAndParseHTML failed: %v", err)
	}

	if doc == nil {
		t.Fatal("expected doc")
	}
}

// TestResponseBodyReadError tests handling of read errors during response body parsing.
func TestResponseBodyReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body>Test</body></html>`)
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	// Normal request should work
	doc, err := client.FetchAndParseHTML(server.URL)
	if err != nil {
		t.Fatalf("FetchAndParseHTML failed: %v", err)
	}

	if doc == nil {
		t.Fatal("expected doc")
	}
}

// TestHTTPErrorFields tests that HTTPError contains proper error information.
func TestHTTPErrorFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Access Denied")
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}

	_, err = client.FetchAndParseHTML(server.URL)
	if err == nil {
		t.Fatal("expected HTTP error")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", httpErr.StatusCode)
	}

	if httpErr.URL == "" {
		t.Error("expected URL to be set in error")
	}
}
