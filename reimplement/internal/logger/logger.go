package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// contextKeys are the keys used for storing values in context
type contextKey string

const (
	requestIDKey     contextKey = "request-id"
	correlationIDKey contextKey = "correlation-id"
)

// NewLogger creates and returns a new SugaredLogger configured for the given environment.
// If env is empty, it reads from the APP_ENV environment variable.
// Defaults to production mode if not specified or unrecognized.
//
// Development mode:
//   - Console output with colorized text
//   - Verbose logging (Debug level and above)
//   - Stack traces included
//   - JSON is not used for better readability
//
// Production mode:
//   - JSON output to stdout
//   - Info level and above
//   - No stack traces by default
//   - Optimized for log aggregation systems
func NewLogger(env string) (*zap.SugaredLogger, error) {
	// If env is empty, read from environment variable
	if env == "" {
		env = os.Getenv("APP_ENV")
	}

	var config zap.Config

	switch env {
	case "development", "dev":
		// Development configuration: human-readable, verbose output
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}

	default:
		// Production configuration: JSON output, optimized
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
		// Add caller information for debugging
		config.EncoderConfig.CallerKey = "caller"
		config.EncoderConfig.LevelKey = "level"
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger.Sugar(), nil
}

// WithRequestID injects a RequestID into the given context.
// This ID should be unique per request and used for tracing a single request
// through the system.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// ExtractRequestID retrieves the RequestID from the given context.
// Returns an empty string if no RequestID is found.
func ExtractRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithCorrelationID injects a CorrelationID into the given context.
// This ID is used to track related requests across multiple services.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// ExtractCorrelationID retrieves the CorrelationID from the given context.
// Returns an empty string if no CorrelationID is found.
func ExtractCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

// LogRequest logs an HTTP request with method, path, status code, and duration.
// Used by HTTP middleware to log request details.
//
// Example output (production JSON):
//
//	{
//	  "level": "info",
//	  "timestamp": "2024-01-15T10:30:45.123Z",
//	  "message": "HTTP request processed",
//	  "method": "GET",
//	  "path": "/api/schedules",
//	  "status": 200,
//	  "duration_ms": 45
//	}
func LogRequest(logger *zap.SugaredLogger, method, path string, statusCode, durationMS int64) {
	logger.Infow("HTTP request processed",
		"method", method,
		"path", path,
		"status", statusCode,
		"duration_ms", durationMS,
	)
}

// LogError logs an error with additional context information.
// Used to log application errors with contextual metadata.
//
// Example:
//
//	LogError(logger, err, map[string]interface{}{
//	  "operation": "create_schedule",
//	  "user_id": "user-123",
//	})
func LogError(logger *zap.SugaredLogger, err error, context map[string]interface{}) {
	fields := []interface{}{"error", err}

	// Add context fields to the log
	for key, value := range context {
		fields = append(fields, key, value)
	}

	logger.Errorw("Error occurred", fields...)
}

// LogServiceCall logs a call to an external or internal service.
// Used to track service-to-service communication and performance.
//
// Example:
//
//	LogServiceCall(logger, "user-service", "GetUserByID", 150, nil)
//	LogServiceCall(logger, "notification-service", "SendEmail", 2000, err)
func LogServiceCall(logger *zap.SugaredLogger, service, operation string, durationMS int64, err error) {
	if err != nil {
		logger.Errorw("Service call failed",
			"service", service,
			"operation", operation,
			"duration_ms", durationMS,
			"error", err,
		)
		return
	}

	logger.Infow("Service call succeeded",
		"service", service,
		"operation", operation,
		"duration_ms", durationMS,
	)
}
