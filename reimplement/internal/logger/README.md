# Logger Package

The `logger` package provides structured logging capabilities for the SchedCU v2 application using Uber's `zap` library. It supports both development and production environments with environment-specific configurations.

## Features

- **Structured Logging**: JSON output in production, human-readable console output in development
- **Log Levels**: Debug, Info, Warn, Error with environment-specific filtering
- **Request ID Injection**: Automatic RequestID generation and context injection for request tracing
- **Correlation ID Tracking**: Service-to-service correlation for distributed tracing
- **HTTP Middleware**: RequestID and request logging middleware for HTTP handlers
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Performance Optimized**: Minimal overhead in production mode

## Installation

The package is part of the internal modules and doesn't need to be separately installed. The main dependency is:

```bash
go get go.uber.org/zap
go get github.com/google/uuid
```

## Quick Start

### Basic Logger Initialization

```go
import (
    "github.com/schedcu/v2/internal/logger"
)

func main() {
    // Initialize logger - reads from APP_ENV environment variable
    log, err := logger.NewLogger("production")
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    // Use logger
    log.Info("Application started")
    log.Errorw("An error occurred", "operation", "startup", "code", 500)
}
```

### With HTTP Server

```go
import (
    "net/http"
    "github.com/schedcu/v2/internal/logger"
)

func main() {
    log, _ := logger.NewLogger("production")
    defer log.Sync()

    // Create router and add middleware
    mux := http.NewServeMux()

    // Apply middleware in order
    var handler http.Handler = mux
    handler = logger.LoggingMiddleware(log)(handler)
    handler = logger.RequestIDMiddleware(log)(handler)

    // Register handlers
    mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })

    http.ListenAndServe(":8080", handler)
}
```

## API Reference

### Logger Initialization

#### `NewLogger(env string) (*zap.SugaredLogger, error)`

Creates and returns a new SugaredLogger configured for the given environment.

**Parameters:**
- `env`: Environment name ("development", "production", or empty to read from APP_ENV variable)

**Returns:**
- `*zap.SugaredLogger`: Configured logger instance
- `error`: Error if logger creation failed

**Behavior:**

| Environment | Output Format | Level | Features |
|---|---|---|---|
| development/dev | Console (colored) | Debug+ | Stack traces, verbose output |
| production (default) | JSON | Info+ | Optimized for log aggregation |

**Example:**
```go
log, err := logger.NewLogger("development")
if err != nil {
    log.Fatal("Failed to create logger:", err)
}
```

### Context Helpers

#### `WithRequestID(ctx context.Context, requestID string) context.Context`

Injects a RequestID into the given context. Used to track individual requests through the system.

**Parameters:**
- `ctx`: Base context
- `requestID`: Unique identifier for this request

**Returns:**
- `context.Context`: New context with RequestID injected

**Example:**
```go
ctx := context.Background()
ctx = logger.WithRequestID(ctx, "req-12345")
```

#### `ExtractRequestID(ctx context.Context) string`

Retrieves the RequestID from the given context.

**Parameters:**
- `ctx`: Context to search

**Returns:**
- `string`: The RequestID (empty string if not found)

**Example:**
```go
requestID := logger.ExtractRequestID(ctx)
if requestID != "" {
    log.Infow("Processing request", "request_id", requestID)
}
```

#### `WithCorrelationID(ctx context.Context, correlationID string) context.Context`

Injects a CorrelationID into the given context for tracking related requests across services.

**Parameters:**
- `ctx`: Base context
- `correlationID`: Identifier for correlated operations

**Returns:**
- `context.Context`: New context with CorrelationID injected

#### `ExtractCorrelationID(ctx context.Context) string`

Retrieves the CorrelationID from the given context.

### Convenience Functions

#### `LogRequest(logger, method, path, statusCode, durationMS)`

Logs an HTTP request with standard fields.

**Example:**
```go
logger.LogRequest(log, "GET", "/api/schedules", 200, 45)
```

**Output (production JSON):**
```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:45.123Z",
  "message": "HTTP request processed",
  "method": "GET",
  "path": "/api/schedules",
  "status": 200,
  "duration_ms": 45
}
```

#### `LogError(logger, error, context)`

Logs an error with contextual metadata.

**Example:**
```go
logger.LogError(log, err, map[string]interface{}{
    "operation": "create_schedule",
    "user_id": "user-123",
})
```

#### `LogServiceCall(logger, service, operation, durationMS, error)`

Logs a service-to-service call with duration and error information.

**Example:**
```go
// Successful call
logger.LogServiceCall(log, "user-service", "GetUserByID", 120, nil)

// Failed call
logger.LogServiceCall(log, "user-service", "UpdateUser", 5000, err)
```

### Middleware

#### `RequestIDMiddleware(logger) func(http.Handler) http.Handler`

HTTP middleware that injects a RequestID into the request context.

**Behavior:**
- Checks for `X-Request-ID` header in incoming request
- If present, uses that value
- If absent, generates a new UUID
- Injects RequestID into request context

**Usage:**
```go
mux := http.NewServeMux()
var handler http.Handler = mux
handler = logger.RequestIDMiddleware(log)(handler)
```

**Header Examples:**
- With existing ID: `X-Request-ID: my-request-id-123`
- Generated ID: UUID like `550e8400-e29b-41d4-a716-446655440000`

#### `LoggingMiddleware(logger) func(http.Handler) http.Handler`

HTTP middleware that logs request details including method, path, status, and duration.

**Behavior:**
- Wraps response writer to capture status code
- Measures request duration
- Logs at INFO level for successful responses (< 400)
- Logs at ERROR level for client/server errors (>= 400)
- Extracts and includes RequestID from context

**Usage (after RequestIDMiddleware):**
```go
var handler http.Handler = mux
handler = logger.LoggingMiddleware(log)(handler)
handler = logger.RequestIDMiddleware(log)(handler)
```

**Log Output (successful request, production JSON):**
```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:45.123Z",
  "message": "HTTP request processed",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/schedules",
  "status": 200,
  "duration_ms": 45
}
```

**Log Output (error, production JSON):**
```json
{
  "level": "error",
  "timestamp": "2024-01-15T10:30:45.123Z",
  "message": "HTTP request processed",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/api/schedules",
  "status": 500,
  "duration_ms": 250
}
```

#### `CorrelationIDMiddleware(logger) func(http.Handler) http.Handler`

HTTP middleware that injects a CorrelationID into the request context.

**Behavior:**
- Similar to RequestIDMiddleware but for CorrelationID
- Used to track related requests across multiple services
- Checks for `X-Correlation-ID` header

**Usage:**
```go
var handler http.Handler = mux
handler = logger.CorrelationIDMiddleware(log)(handler)
handler = logger.RequestIDMiddleware(log)(handler)
```

## Configuration

### Environment Variables

| Variable | Description | Default | Values |
|---|---|---|---|
| `APP_ENV` | Application environment | production | development, production |

### Logger Configuration by Environment

#### Development Mode

```go
logger.NewLogger("development")
```

Features:
- Console output with ANSI color codes
- Debug level and above (verbose)
- Stack traces included for errors
- Human-readable format

Example Output:
```
2025-11-15T16:52:14.792-0500	[34mINFO[0m	logger/logger.go:120	HTTP request processed	{"method": "GET", "path": "/api/schedules", "status": 200, "duration_ms": 45}
```

#### Production Mode

```go
logger.NewLogger("production")
```

Features:
- JSON output optimized for log aggregation (Datadog, ELK, etc.)
- Info level and above
- Minimal overhead
- ISO8601 timestamps
- Caller information included

Example Output:
```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:45.123Z",
  "caller": "logger/logger.go:120",
  "message": "HTTP request processed",
  "method": "GET",
  "path": "/api/schedules",
  "status": 200,
  "duration_ms": 45
}
```

## Complete Example

```go
package main

import (
    "context"
    "net/http"
    "github.com/schedcu/v2/internal/logger"
)

func main() {
    // Initialize logger
    log, err := logger.NewLogger("production")
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    log.Info("Starting application")

    // Create HTTP server with middleware
    mux := http.NewServeMux()

    // Register handlers
    mux.HandleFunc("/api/schedules", scheduleHandler(log))

    // Apply middleware in order (last applied = first executed)
    var handler http.Handler = mux
    handler = logger.LoggingMiddleware(log)(handler)
    handler = logger.CorrelationIDMiddleware(log)(handler)
    handler = logger.RequestIDMiddleware(log)(handler)

    // Start server
    log.Info("Starting HTTP server", "port", 8080)
    http.ListenAndServe(":8080", handler)
}

func scheduleHandler(log *zap.SugaredLogger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract IDs from context
        requestID := logger.ExtractRequestID(r.Context())
        correlationID := logger.ExtractCorrelationID(r.Context())

        // Use them for logging
        log.Infow("Processing schedule request",
            "request_id", requestID,
            "correlation_id", correlationID,
            "user_id", r.Header.Get("X-User-ID"),
        )

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }
}
```

## Best Practices

1. **Always sync the logger before exiting**: `defer log.Sync()`
2. **Use structured logging**: `log.Infow("message", "key", value)` not string concatenation
3. **Include context in errors**: Use `LogError` to include operation context
4. **Apply middleware in correct order**: RequestIDMiddleware should be innermost (applied last)
5. **Use RequestID for tracing**: Log RequestID with important operations for debugging
6. **Use CorrelationID for service calls**: Include CorrelationID in outbound requests to maintain correlation
7. **Pass context through call chain**: Always propagate context with RequestID/CorrelationID
8. **Environment-specific configuration**: Use APP_ENV to match your deployment environment

## Testing

Run tests with coverage:
```bash
go test ./internal/logger -v -cover
```

Current coverage: 86.1% of statements

Test categories:
- Logger initialization (development/production modes)
- JSON output validation
- Log level functionality
- Context injection and extraction
- RequestID generation and preservation
- CorrelationID handling
- HTTP middleware integration
- Concurrent logging
- Error conditions

## Dependencies

- `go.uber.org/zap` (v1.27.0): Core logging library
- `github.com/google/uuid` (v1.6.0): UUID generation for RequestID/CorrelationID

## Performance Considerations

- **Logger Creation**: O(1) - minimal overhead
- **Logging Operations**: O(1) amortized - zap uses buffering
- **Middleware Overhead**: < 1ms per request for typical use cases
- **Memory**: ~1KB per logger instance
- **Thread Safety**: Fully concurrent-safe across goroutines

Production mode is optimized for performance and should be used in deployment environments.
