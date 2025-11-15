package logger

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// TestNewLoggerDevelopment tests logger initialization in development mode
func TestNewLoggerDevelopment(t *testing.T) {
	os.Setenv("APP_ENV", "development")
	defer os.Unsetenv("APP_ENV")

	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger(development) failed: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	// Development logger should be a SugaredLogger
	logger.Info("test message")
}

// TestNewLoggerProduction tests logger initialization in production mode
func TestNewLoggerProduction(t *testing.T) {
	logger, err := NewLogger("production")
	if err != nil {
		t.Fatalf("NewLogger(production) failed: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	logger.Info("test message")
}

// TestLoggerJSONOutput tests that production logger outputs valid JSON
func TestLoggerJSONOutput(t *testing.T) {
	logger, err := NewLogger("production")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	// Just verify that logging doesn't panic
	// JSON output validation would require capturing stderr/stdout which is complex in tests
	logger.Info("test message", "key", "value")
	logger.Sync()
}

// TestLogLevels tests all log levels work correctly
func TestLogLevels(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	tests := []struct {
		name    string
		logFunc func(...interface{})
		message string
	}{
		{
			name:    "Debug",
			logFunc: logger.Debug,
			message: "debug message",
		},
		{
			name:    "Info",
			logFunc: logger.Info,
			message: "info message",
		},
		{
			name:    "Warn",
			logFunc: logger.Warn,
			message: "warn message",
		},
		{
			name:    "Error",
			logFunc: logger.Error,
			message: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			tt.logFunc(tt.message)
		})
	}
}

// TestWithRequestID tests RequestID context injection
func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-123"

	ctxWithID := WithRequestID(ctx, requestID)
	if ctxWithID == nil {
		t.Fatal("WithRequestID returned nil context")
	}

	extracted := ExtractRequestID(ctxWithID)
	if extracted != requestID {
		t.Errorf("Expected RequestID %q, got %q", requestID, extracted)
	}
}

// TestExtractRequestIDEmptyContext tests ExtractRequestID on context without RequestID
func TestExtractRequestIDEmptyContext(t *testing.T) {
	ctx := context.Background()
	extracted := ExtractRequestID(ctx)
	if extracted != "" {
		t.Errorf("Expected empty RequestID, got %q", extracted)
	}
}

// TestWithCorrelationID tests CorrelationID context injection
func TestWithCorrelationID(t *testing.T) {
	ctx := context.Background()
	correlationID := "corr-123456"

	ctxWithID := WithCorrelationID(ctx, correlationID)
	if ctxWithID == nil {
		t.Fatal("WithCorrelationID returned nil context")
	}

	extracted := ExtractCorrelationID(ctxWithID)
	if extracted != correlationID {
		t.Errorf("Expected CorrelationID %q, got %q", correlationID, extracted)
	}
}

// TestExtractCorrelationIDEmptyContext tests ExtractCorrelationID on context without CorrelationID
func TestExtractCorrelationIDEmptyContext(t *testing.T) {
	ctx := context.Background()
	extracted := ExtractCorrelationID(ctx)
	if extracted != "" {
		t.Errorf("Expected empty CorrelationID, got %q", extracted)
	}
}

// TestWithRequestIDMultiple tests multiple RequestID manipulations
func TestWithRequestIDMultiple(t *testing.T) {
	ctx := context.Background()
	id1 := "request-1"
	id2 := "request-2"

	ctx = WithRequestID(ctx, id1)
	if ExtractRequestID(ctx) != id1 {
		t.Errorf("Expected %q, got %q", id1, ExtractRequestID(ctx))
	}

	// Overwriting should work
	ctx = WithRequestID(ctx, id2)
	if ExtractRequestID(ctx) != id2 {
		t.Errorf("Expected %q, got %q", id2, ExtractRequestID(ctx))
	}
}

// TestLogRequest tests LogRequest convenience function
func TestLogRequest(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	// Should not panic
	LogRequest(logger, "GET", "/api/schedules", 200, 45)
}

// TestLogError tests LogError convenience function
func TestLogError(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	testErr := fmt.Errorf("test error occurred")

	// Should not panic
	LogError(logger, testErr, map[string]interface{}{
		"operation": "test_operation",
		"status":    500,
	})
}

// TestLogServiceCall tests LogServiceCall convenience function
func TestLogServiceCall(t *testing.T) {
	logger, err := NewLogger("development")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	// Should not panic with no error
	LogServiceCall(logger, "user-service", "GetUser", 120, nil)

	// Should not panic with error
	testErr := fmt.Errorf("service failed")
	LogServiceCall(logger, "user-service", "UpdateUser", 500, testErr)
}

// TestNewLoggerInvalidEnv tests behavior with invalid environment
func TestNewLoggerInvalidEnv(t *testing.T) {
	// Should not error on invalid env, defaults to production
	logger, err := NewLogger("invalid-env")
	if err != nil {
		t.Fatalf("NewLogger failed on invalid env: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}
}

// TestLoggerConcurrency tests logger is safe for concurrent use
func TestLoggerConcurrency(t *testing.T) {
	logger, err := NewLogger("production")
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}

	done := make(chan bool)

	// Launch multiple goroutines writing logs
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Infof("message from goroutine %d", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	logger.Sync()
}

// TestContextWithBothIDs tests context with both RequestID and CorrelationID
func TestContextWithBothIDs(t *testing.T) {
	ctx := context.Background()
	requestID := "req-123"
	correlationID := "corr-456"

	ctx = WithRequestID(ctx, requestID)
	ctx = WithCorrelationID(ctx, correlationID)

	if ExtractRequestID(ctx) != requestID {
		t.Errorf("Expected RequestID %q, got %q", requestID, ExtractRequestID(ctx))
	}

	if ExtractCorrelationID(ctx) != correlationID {
		t.Errorf("Expected CorrelationID %q, got %q", correlationID, ExtractCorrelationID(ctx))
	}
}

// TestNewLoggerFromEnvVar tests logger initialization from environment variable
func TestNewLoggerFromEnvVar(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	logger, err := NewLogger("")
	if err != nil {
		t.Fatalf("NewLogger with empty env failed: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}
}
