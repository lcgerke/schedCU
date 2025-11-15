package logger

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ResponseWriter is a wrapper around http.ResponseWriter that captures the status code.
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader writes the status code and captures it.
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	if !rw.written {
		rw.statusCode = statusCode
		rw.written = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write writes data to the response and ensures WriteHeader is called.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// RequestIDMiddleware is an HTTP middleware that injects a RequestID into the request context.
// It checks for an existing X-Request-ID header and uses that if present,
// otherwise generates a new UUID.
//
// Example usage:
//
//	router.Use(RequestIDMiddleware(logger))
func RequestIDMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for existing Request ID in header
			requestID := r.Header.Get("X-Request-ID")

			// If not present, generate a new one
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Inject into context
			ctx := WithRequestID(r.Context(), requestID)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggingMiddleware is an HTTP middleware that logs request details including
// method, path, status code, and duration. Must be used after RequestIDMiddleware
// to properly log the RequestID.
//
// Logs are produced at INFO level for successful requests and ERROR level for
// responses with status >= 400.
//
// Example usage:
//
//	router.Use(RequestIDMiddleware(logger))
//	router.Use(LoggingMiddleware(logger))
//
// Example log output (production JSON):
//
//	{
//	  "level": "info",
//	  "timestamp": "2024-01-15T10:30:45.123Z",
//	  "message": "HTTP request processed",
//	  "request_id": "550e8400-e29b-41d4-a716-446655440000",
//	  "method": "GET",
//	  "path": "/api/schedules",
//	  "status": 200,
//	  "duration_ms": 45
//	}
func LoggingMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap response writer to capture status code
			wrapped := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Record start time
			startTime := time.Now()

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(startTime)
			durationMS := duration.Milliseconds()

			// Extract request ID from context
			requestID := ExtractRequestID(r.Context())

			// Choose log level based on status code
			if wrapped.statusCode >= 400 {
				logger.Errorw("HTTP request processed",
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
					"status", wrapped.statusCode,
					"duration_ms", durationMS,
				)
			} else {
				logger.Infow("HTTP request processed",
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
					"status", wrapped.statusCode,
					"duration_ms", durationMS,
				)
			}
		})
	}
}

// CorrelationIDMiddleware is an HTTP middleware that injects a CorrelationID into the request context.
// It checks for an existing X-Correlation-ID header and uses that if present,
// otherwise generates a new UUID. The CorrelationID is used to track related requests
// across multiple services.
//
// Example usage:
//
//	router.Use(RequestIDMiddleware(logger))
//	router.Use(CorrelationIDMiddleware(logger))
//	router.Use(LoggingMiddleware(logger))
func CorrelationIDMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for existing Correlation ID in header
			correlationID := r.Header.Get("X-Correlation-ID")

			// If not present, generate a new one
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			// Inject into context
			ctx := WithCorrelationID(r.Context(), correlationID)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
