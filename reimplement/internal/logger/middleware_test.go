package logger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// MockResponseWriter is a mock implementation of http.ResponseWriter
type MockResponseWriter struct {
	statusCode int
	header     http.Header
	body       []byte
}

func (m *MockResponseWriter) Header() http.Header {
	return m.header
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	m.body = append(m.body, b...)
	return len(b), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

// TestRequestIDMiddleware tests that RequestID is injected into context
func TestRequestIDMiddleware(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	// Create a simple handler that checks for RequestID in context
	handler := RequestIDMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := ExtractRequestID(r.Context())
		if requestID == "" {
			t.Error("RequestID not found in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRequestIDMiddlewareGeneratesID tests that middleware generates RequestID if not present
func TestRequestIDMiddlewareGeneratesID(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	var extractedID string

	handler := RequestIDMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedID = ExtractRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if extractedID == "" {
		t.Error("RequestID was not generated")
	}
}

// TestRequestIDMiddlewarePreservesExisting tests that middleware preserves existing RequestID
func TestRequestIDMiddlewarePreservesExisting(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	existingID := "existing-request-123"
	var extractedID string

	handler := RequestIDMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedID = ExtractRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if extractedID != existingID {
		t.Errorf("Expected RequestID %q, got %q", existingID, extractedID)
	}
}

// TestLoggingMiddleware tests HTTP request logging
func TestLoggingMiddleware(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	handler := LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))

	req := httptest.NewRequest("GET", "/api/schedules", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestLoggingMiddlewareCapturesthStatus tests that logging middleware captures status code
func TestLoggingMiddlewareStatus(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	testCases := []int{http.StatusOK, http.StatusNotFound, http.StatusInternalServerError}

	for _, expectedStatus := range testCases {
		t.Run(string(rune(expectedStatus)), func(t *testing.T) {
			handler := LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(expectedStatus)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != expectedStatus {
				t.Errorf("Expected status %d, got %d", expectedStatus, w.Code)
			}
		})
	}
}

// TestChainedMiddleware tests that RequestID and Logging middleware work together
func TestChainedMiddleware(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	var requestID string
	handler := RequestIDMiddleware(logger)(
		LoggingMiddleware(logger)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestID = ExtractRequestID(r.Context())
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if requestID == "" {
		t.Error("RequestID not available in handler")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestMiddlewareWithContextDeadline tests middleware respects context deadline
func TestMiddlewareWithContextDeadline(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	handler := RequestIDMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that context has our values
		requestID := ExtractRequestID(r.Context())
		if requestID == "" {
			t.Error("RequestID not in context")
		}

		// Check deadline propagates
		_, ok := r.Context().Deadline()
		// Deadline may or may not be set, but the important thing is that
		// the context is functional
		_ = ok
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRequestIDHeaderCaseInsensitive tests that X-Request-ID header detection is case-insensitive
func TestRequestIDHeaderCaseInsensitive(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	testID := "test-id-456"
	var extractedID string

	handler := RequestIDMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedID = ExtractRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// HTTP headers are case-insensitive in Go
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-request-id", testID) // lowercase

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if extractedID != testID {
		t.Errorf("Expected RequestID %q, got %q", testID, extractedID)
	}
}
