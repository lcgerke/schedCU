// Package amion provides Amion service integration for the schedCU system.
// It handles HTTP communication with the Amion scheduling platform, including
// authentication, session management, and HTML parsing.
package amion

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

// AmionHTTPClient handles HTTP communication with the Amion service.
// It manages session cookies, retries, and timeout handling.
type AmionHTTPClient struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	logger     *zap.SugaredLogger
}

// HTTPError represents an HTTP error response from the server.
type HTTPError struct {
	StatusCode int
	URL        string
	Message    string
}

// Error implements the error interface for HTTPError.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s (URL: %s)", e.StatusCode, e.Message, e.URL)
}

// NetworkError represents a network connectivity error.
type NetworkError struct {
	URL       string
	Underlying error
}

// Error implements the error interface for NetworkError.
func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v (URL: %s)", e.Underlying, e.URL)
}

// ParseError represents an error during HTML parsing.
type ParseError struct {
	URL        string
	Underlying error
}

// Error implements the error interface for ParseError.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error: %v (URL: %s)", e.Underlying, e.URL)
}

// RetryError represents an error after all retries are exhausted.
type RetryError struct {
	URL           string
	Attempts      int
	LastError     error
	LastStatusCode int
}

// Error implements the error interface for RetryError.
func (e *RetryError) Error() string {
	return fmt.Sprintf("failed after %d attempts (last status: %d): %v (URL: %s)",
		e.Attempts, e.LastStatusCode, e.LastError, e.URL)
}

const (
	// DefaultTimeout is the default request timeout for Amion HTTP client
	DefaultTimeout = 30 * time.Second

	// MaxRetries is the maximum number of retry attempts
	MaxRetries = 3

	// DefaultUserAgent is a realistic browser user agent
	DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// NewAmionHTTPClient creates a new AmionHTTPClient with the given base URL.
// It validates the URL, configures timeouts, retry logic, connection pooling,
// and session management via cookie jar.
//
// Parameters:
//   - baseURL: The base URL of the Amion service (e.g., "https://amion.example.com")
//
// Returns:
//   - *AmionHTTPClient: A configured HTTP client for Amion service
//   - error: If the URL is invalid or client creation fails
//
// Example:
//
//	client, err := NewAmionHTTPClient("https://amion.example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func NewAmionHTTPClient(baseURL string) (*AmionHTTPClient, error) {
	// Validate the base URL
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Validate URL scheme
	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("URL must have a scheme (http or https)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Create cookie jar for session persistence
	// Using a simple in-memory jar for cookie management
	cookieJar := NewSimpleCookieJar()

	// Configure transport with connection pooling and timeouts
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     32,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
	}

	// Create HTTP client with configured transport
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   DefaultTimeout,
		Jar:       cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Limit redirects to prevent redirect loops
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return &AmionHTTPClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		userAgent:  DefaultUserAgent,
	}, nil
}

// SetLogger sets the logger for HTTP client operations.
// This is optional and used for logging HTTP requests/responses.
func (c *AmionHTTPClient) SetLogger(logger *zap.SugaredLogger) {
	c.logger = logger
}

// FetchAndParseHTML fetches a URL and parses the response as HTML.
// It handles gzip compression, character encoding detection, and implements
// exponential backoff retry logic for transient failures.
//
// Parameters:
//   - url: The URL to fetch (can be absolute or relative to base URL)
//
// Returns:
//   - *goquery.Document: Parsed HTML document
//   - error: HTTPError, NetworkError, ParseError, or RetryError on failure
//
// Example:
//
//	doc, err := client.FetchAndParseHTML("https://amion.example.com/schedule")
//	if err != nil {
//	    switch err := err.(type) {
//	    case *HTTPError:
//	        log.Printf("HTTP error: %v", err.StatusCode)
//	    case *NetworkError:
//	        log.Printf("Network error: %v", err.Underlying)
//	    }
//	}
//	doc.Find("table").Each(func(i int, s *goquery.Selection) {
//	    // Process table rows
//	})
func (c *AmionHTTPClient) FetchAndParseHTML(urlStr string) (*goquery.Document, error) {
	ctx := context.Background()
	return c.FetchAndParseHTMLWithContext(ctx, urlStr)
}

// FetchAndParseHTMLWithContext fetches a URL and parses the response as HTML
// using the provided context. This allows cancellation and timeout control.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout
//   - url: The URL to fetch (can be absolute or relative to base URL)
//
// Returns:
//   - *goquery.Document: Parsed HTML document
//   - error: HTTPError, NetworkError, ParseError, or RetryError on failure
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	doc, err := client.FetchAndParseHTMLWithContext(ctx, url)
func (c *AmionHTTPClient) FetchAndParseHTMLWithContext(ctx context.Context, urlStr string) (*goquery.Document, error) {
	// Validate URL
	if urlStr == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Convert relative URLs to absolute
	if !parsedURL.IsAbs() {
		absoluteURL := fmt.Sprintf("%s%s", c.baseURL, urlStr)
		parsedURL, err = url.Parse(absoluteURL)
		if err != nil {
			return nil, &ParseError{URL: urlStr, Underlying: err}
		}
		urlStr = absoluteURL
	}

	var lastErr error
	var lastStatusCode int

	// Implement exponential backoff retry logic
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		// Calculate backoff delay for retries (exponential: 1s, 2s, 4s)
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second

			if c.logger != nil {
				c.logger.Debugw("retrying request", "url", urlStr, "attempt", attempt, "backoff", backoff)
			}

			select {
			case <-time.After(backoff):
				// Continue with retry
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			}
		}

		// Create request with context
		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			lastErr = err
			continue
		}

		// Set User-Agent header
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		if c.logger != nil {
			c.logger.Debugw("fetching URL", "url", urlStr, "method", "GET")
		}

		// Perform the HTTP request
		resp, err := c.httpClient.Do(req)

		// Handle network errors
		if err != nil {
			lastErr = err
			lastStatusCode = 0

			// Check if this is a temporary error (retryable)
			if isTemporaryError(err) && attempt < MaxRetries {
				if c.logger != nil {
					c.logger.Warnw("temporary error, will retry", "error", err, "attempt", attempt)
				}
				continue
			}

			// For permanent errors or final attempt, return error
			return nil, &NetworkError{URL: urlStr, Underlying: err}
		}

		// Always read and close the body
		defer resp.Body.Close()

		lastStatusCode = resp.StatusCode

		// Check HTTP status code
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))

			if c.logger != nil {
				c.logger.Warnw("HTTP error", "status", resp.StatusCode, "url", urlStr)
			}

			// Retry on 5xx errors (server errors)
			if resp.StatusCode >= 500 && attempt < MaxRetries {
				if c.logger != nil {
					c.logger.Infow("server error, retrying", "status", resp.StatusCode, "attempt", attempt)
				}
				continue
			}

			// Don't retry on 4xx errors (client errors)
			return nil, &HTTPError{
				StatusCode: resp.StatusCode,
				URL:        urlStr,
				Message:    strings.TrimSpace(string(body)),
			}
		}

		// Handle gzip compression
		var body io.ReadCloser
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return nil, &ParseError{URL: urlStr, Underlying: fmt.Errorf("failed to create gzip reader: %w", err)}
			}
			defer gzipReader.Close()
			body = gzipReader
		} else {
			body = resp.Body
		}

		// Read the entire response body
		htmlBytes, err := io.ReadAll(body)
		if err != nil {
			lastErr = err
			if attempt < MaxRetries {
				continue
			}
			return nil, &ParseError{URL: urlStr, Underlying: err}
		}

		// Parse HTML with goquery
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlBytes)))
		if err != nil {
			// goquery is lenient and handles malformed HTML gracefully
			// If it still fails, we return the parse error
			lastErr = err
			if attempt < MaxRetries {
				continue
			}
			return nil, &ParseError{URL: urlStr, Underlying: err}
		}

		if c.logger != nil {
			c.logger.Debugw("successfully parsed HTML", "url", urlStr, "attempt", attempt+1)
		}

		return doc, nil
	}

	// All retries exhausted
	return nil, &RetryError{
		URL:            urlStr,
		Attempts:       MaxRetries + 1,
		LastError:      lastErr,
		LastStatusCode: lastStatusCode,
	}
}

// isTemporaryError checks if an error is temporary/retryable.
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation
	if strings.Contains(err.Error(), "context") {
		return false
	}

	// Check for network timeouts
	if strings.Contains(err.Error(), "timeout") {
		return true
	}

	// Check for connection refused (temporary - server might be restarting)
	if strings.Contains(err.Error(), "connection refused") {
		return true
	}

	// Check for temporary network errors
	if strings.Contains(err.Error(), "temporary") {
		return true
	}

	// Check for EOF during response body read (might be temporary)
	if strings.Contains(err.Error(), "EOF") {
		return true
	}

	// Default: not temporary
	return false
}

// Close closes the HTTP client and cleans up resources.
func (c *AmionHTTPClient) Close() error {
	// Close idle connections to free resources
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// SimpleCookieJar is a simple in-memory cookie jar implementation.
// It stores cookies by domain.
type SimpleCookieJar struct {
	cookies map[string][]*http.Cookie
}

// NewSimpleCookieJar creates a new SimpleCookieJar.
func NewSimpleCookieJar() *SimpleCookieJar {
	return &SimpleCookieJar{
		cookies: make(map[string][]*http.Cookie),
	}
}

// SetCookies stores cookies for a given URL.
func (j *SimpleCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.cookies[u.Host] = cookies
}

// Cookies returns cookies for a given URL.
func (j *SimpleCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.cookies[u.Host]
}
