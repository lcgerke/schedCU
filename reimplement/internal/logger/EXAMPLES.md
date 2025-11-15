# Logger Usage Examples

This document provides practical examples of using the logging framework in various scenarios.

## Table of Contents

1. [Basic Logging](#basic-logging)
2. [HTTP Server Setup](#http-server-setup)
3. [Request Tracing](#request-tracing)
4. [Service Integration](#service-integration)
5. [Error Handling](#error-handling)
6. [Context Propagation](#context-propagation)
7. [Production Configuration](#production-configuration)

## Basic Logging

### Simple Logger Creation

```go
package main

import (
    "log"
    "github.com/schedcu/v2/internal/logger"
)

func main() {
    // Create logger (uses APP_ENV environment variable if provided)
    log, err := logger.NewLogger("development")
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    // Simple logging
    log.Info("Application started")
    log.Warn("Warning message")
    log.Error("Error message")
}
```

### Structured Logging

```go
func createSchedule(log *zap.SugaredLogger, scheduleID string, userID string) {
    // Include structured fields with every log
    log.Infow("Creating schedule",
        "schedule_id", scheduleID,
        "user_id", userID,
        "action", "create",
    )

    // Log with multiple fields
    log.Debugw("Schedule details",
        "start_date", "2024-01-15",
        "end_date", "2024-02-15",
        "facility", "St. Mary's Hospital",
        "facility_id", "fac-12345",
    )
}
```

## HTTP Server Setup

### Basic HTTP Server with Logging Middleware

```go
package main

import (
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

    // Create router
    mux := http.NewServeMux()

    // Register handlers
    mux.HandleFunc("/api/health", healthHandler)
    mux.HandleFunc("/api/schedules", getSchedulesHandler)

    // Apply middleware (order matters - last applied is first executed)
    var handler http.Handler = mux
    handler = logger.LoggingMiddleware(log)(handler)
    handler = logger.RequestIDMiddleware(log)(handler)

    // Start server
    log.Info("Starting server", "port", "8080")
    http.ListenAndServe(":8080", handler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"healthy"}`))
}

func getSchedulesHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`[{"id":"sch-1","name":"January Schedule"}]`))
}
```

### HTTP Server with Multiple Middleware

```go
func setupRouter(log *zap.SugaredLogger) http.Handler {
    mux := http.NewServeMux()

    // API routes
    mux.HandleFunc("/api/schedules", func(w http.ResponseWriter, r *http.Request) {
        requestID := logger.ExtractRequestID(r.Context())
        log.Infow("Retrieving schedules", "request_id", requestID)

        w.Header().Set("X-Request-ID", requestID)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("[]"))
    })

    // Health check (can be outside main middleware)
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    // Apply middleware in reverse order (last applied = first executed)
    var handler http.Handler = mux
    handler = logger.LoggingMiddleware(log)(handler)
    handler = logger.CorrelationIDMiddleware(log)(handler)
    handler = logger.RequestIDMiddleware(log)(handler)

    return handler
}

func main() {
    log, _ := logger.NewLogger("production")
    defer log.Sync()

    handler := setupRouter(log)
    http.ListenAndServe(":8080", handler)
}
```

## Request Tracing

### Using RequestID for Distributed Tracing

```go
func processScheduleRequest(w http.ResponseWriter, r *http.Request, log *zap.SugaredLogger) {
    // Extract RequestID from context (injected by middleware)
    requestID := logger.ExtractRequestID(r.Context())

    // Include RequestID in all logs related to this request
    log.Infow("Processing schedule request",
        "request_id", requestID,
        "method", r.Method,
        "path", r.URL.Path,
    )

    // When calling other services, pass RequestID in headers
    serviceReq, _ := http.NewRequest("GET", "http://schedule-service/api/schedules", nil)
    serviceReq.Header.Set("X-Request-ID", requestID)

    resp, err := http.DefaultClient.Do(serviceReq)
    if err != nil {
        logger.LogError(log, err, map[string]interface{}{
            "request_id": requestID,
            "operation": "fetch_schedules",
            "service": "schedule-service",
        })
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    log.Infow("Schedule request completed",
        "request_id", requestID,
        "status", resp.StatusCode,
    )

    w.WriteHeader(http.StatusOK)
}
```

### Request ID in Headers

```go
func addRequestIDToResponse(w http.ResponseWriter, r *http.Request) {
    requestID := logger.ExtractRequestID(r.Context())

    // Include RequestID in response header for client reference
    w.Header().Set("X-Request-ID", requestID)
    w.Header().Set("X-Correlation-ID", logger.ExtractCorrelationID(r.Context()))
}
```

## Service Integration

### Logging Service Calls

```go
import (
    "time"
    "github.com/schedcu/v2/internal/logger"
)

func getUserFromService(ctx context.Context, log *zap.SugaredLogger, userID string) (User, error) {
    start := time.Now()

    // Fetch user from external service
    user, err := callUserService(ctx, userID)

    duration := time.Since(start).Milliseconds()

    // Log the service call
    logger.LogServiceCall(log, "user-service", "GetUser", duration, err)

    if err != nil {
        // Also log with context
        logger.LogError(log, err, map[string]interface{}{
            "service": "user-service",
            "operation": "GetUser",
            "user_id": userID,
            "duration_ms": duration,
        })
        return nil, err
    }

    return user, nil
}

func callUserService(ctx context.Context, userID string) (User, error) {
    // Implementation
    return User{}, nil
}
```

### Multi-Service Orchestration

```go
func createScheduleWithAssignments(ctx context.Context, log *zap.SugaredLogger, req ScheduleRequest) (Schedule, error) {
    requestID := logger.ExtractRequestID(ctx)

    log.Infow("Creating schedule with assignments",
        "request_id", requestID,
        "num_assignments", len(req.Assignments),
    )

    // Create schedule
    schedule := Schedule{}
    {
        start := time.Now()
        var err error
        schedule, err = createScheduleInDB(ctx, req)
        duration := time.Since(start).Milliseconds()

        logger.LogServiceCall(log, "schedule-db", "Insert", duration, err)
        if err != nil {
            return nil, err
        }
    }

    // Assign staff for each slot
    for _, assignment := range req.Assignments {
        start := time.Now()
        err := assignStaff(ctx, schedule.ID, assignment)
        duration := time.Since(start).Milliseconds()

        logger.LogServiceCall(log, "assignment-service", "Create", duration, err)
        if err != nil {
            logger.LogError(log, err, map[string]interface{}{
                "request_id": requestID,
                "schedule_id": schedule.ID,
                "assignment": assignment,
            })
            // Decide: fail entire operation or continue?
            return nil, err
        }
    }

    log.Infow("Schedule creation completed",
        "request_id", requestID,
        "schedule_id", schedule.ID,
        "assignments", len(req.Assignments),
    )

    return schedule, nil
}
```

## Error Handling

### Logging Different Error Types

```go
func handleDatabaseError(log *zap.SugaredLogger, err error, operation string, context map[string]interface{}) {
    context["error_type"] = "database"
    context["operation"] = operation
    logger.LogError(log, err, context)
}

func handleValidationError(log *zap.SugaredLogger, err error, fields map[string]interface{}) {
    context := map[string]interface{}{
        "error_type": "validation",
        "invalid_fields": fields,
    }
    logger.LogError(log, err, context)
}

func processSchedule(log *zap.SugaredLogger, req ScheduleRequest) error {
    // Validate input
    if req.StartDate.After(req.EndDate) {
        err := fmt.Errorf("invalid date range")
        handleValidationError(log, err, map[string]interface{}{
            "start_date": req.StartDate,
            "end_date": req.EndDate,
        })
        return err
    }

    // Save to database
    err := saveSchedule(req)
    if err != nil {
        handleDatabaseError(log, err, "save_schedule", map[string]interface{}{
            "schedule_name": req.Name,
            "facility_id": req.FacilityID,
        })
        return err
    }

    return nil
}
```

### Panic Recovery with Logging

```go
func recoverMiddleware(log *zap.SugaredLogger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    requestID := logger.ExtractRequestID(r.Context())
                    logger.LogError(log, fmt.Errorf("panic recovered"), map[string]interface{}{
                        "request_id": requestID,
                        "panic": err,
                        "method": r.Method,
                        "path": r.URL.Path,
                    })
                    http.Error(w, "Internal server error", http.StatusInternalServerError)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

## Context Propagation

### Passing Context Through Function Calls

```go
func handleScheduleRequest(w http.ResponseWriter, r *http.Request, log *zap.SugaredLogger) {
    ctx := r.Context()

    // Process with context propagation
    schedule, err := getScheduleWithContext(ctx, log, "sch-123")
    if err != nil {
        requestID := logger.ExtractRequestID(ctx)
        logger.LogError(log, err, map[string]interface{}{
            "request_id": requestID,
            "operation": "get_schedule",
        })
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(schedule)
}

func getScheduleWithContext(ctx context.Context, log *zap.SugaredLogger, scheduleID string) (Schedule, error) {
    requestID := logger.ExtractRequestID(ctx)

    log.Debugw("Fetching schedule",
        "request_id", requestID,
        "schedule_id", scheduleID,
    )

    // Pass context to database call
    return queryScheduleFromDB(ctx, scheduleID)
}

func queryScheduleFromDB(ctx context.Context, scheduleID string) (Schedule, error) {
    // Context can be used for query timeout, cancellation, etc.
    // RequestID is available via logger.ExtractRequestID(ctx)
    return Schedule{}, nil
}
```

### Creating Child Contexts with Additional IDs

```go
func processWithCorrelation(w http.ResponseWriter, r *http.Request, log *zap.SugaredLogger) {
    ctx := r.Context()
    requestID := logger.ExtractRequestID(ctx)

    // Create correlation ID for related operations
    correlationID := "corr-" + uuid.New().String()
    ctx = logger.WithCorrelationID(ctx, correlationID)

    log.Infow("Starting correlated operations",
        "request_id", requestID,
        "correlation_id", correlationID,
    )

    // All child operations can extract both IDs
    processChild1(ctx, log)
    processChild2(ctx, log)

    w.WriteHeader(http.StatusOK)
}

func processChild1(ctx context.Context, log *zap.SugaredLogger) {
    requestID := logger.ExtractRequestID(ctx)
    correlationID := logger.ExtractCorrelationID(ctx)

    log.Infow("Child operation 1",
        "request_id", requestID,
        "correlation_id", correlationID,
    )
}

func processChild2(ctx context.Context, log *zap.SugaredLogger) {
    requestID := logger.ExtractRequestID(ctx)
    correlationID := logger.ExtractCorrelationID(ctx)

    log.Infow("Child operation 2",
        "request_id", requestID,
        "correlation_id", correlationID,
    )
}
```

## Production Configuration

### Environment-Based Setup

```go
package main

import (
    "os"
    "github.com/schedcu/v2/internal/logger"
)

func initLogger() (*zap.SugaredLogger, error) {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "production" // Safe default
    }

    log, err := logger.NewLogger(env)
    if err != nil {
        return nil, err
    }

    // Log initialization
    log.Infow("Logger initialized",
        "environment", env,
        "version", "1.0.0",
    )

    return log, nil
}
```

### Graceful Shutdown with Logging

```go
import (
    "os"
    "os/signal"
    "syscall"
)

func main() {
    log, _ := logger.NewLogger("production")
    defer log.Sync()

    log.Info("Application starting")

    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Run server...

    // Wait for signal
    sig := <-sigChan
    log.Infow("Shutdown signal received", "signal", sig.String())

    // Graceful shutdown
    log.Info("Shutting down...")

    log.Info("Shutdown complete")
}
```

### Structured Logging for Monitoring

```go
func monitorApplicationHealth(log *zap.SugaredLogger) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        metrics := collectMetrics()

        log.Infow("Application health check",
            "memory_mb", metrics.MemoryMB,
            "goroutines", metrics.NumGoroutines,
            "open_connections", metrics.OpenDBConnections,
            "requests_per_sec", metrics.RequestsPerSec,
            "avg_response_time_ms", metrics.AvgResponseTimeMS,
            "error_rate_percent", metrics.ErrorRatePercent,
        )

        // Alert on anomalies
        if metrics.ErrorRatePercent > 5.0 {
            log.Warnw("High error rate detected",
                "error_rate_percent", metrics.ErrorRatePercent,
                "threshold", 5.0,
            )
        }
    }
}

func collectMetrics() Metrics {
    // Implementation
    return Metrics{}
}
```

## Log Output Examples

### Development Mode Output

```
2025-11-15T16:52:14.792-0500	[35mDEBUG[0m	logger/logger.go:120	Processing schedule request	{"request_id": "550e8400-e29b-41d4-a716-446655440000", "schedule_id": "sch-123"}
2025-11-15T16:52:14.812-0500	[34mINFO[0m	logger/logger.go:150	Schedule processing completed	{"request_id": "550e8400-e29b-41d4-a716-446655440000", "assignments": 12}
2025-11-15T16:52:15.120-0500	[33mWARN[0m	logger/logger.go:180	Service call slow	{"service": "user-service", "duration_ms": 2500}
2025-11-15T16:52:15.145-0500	[34mINFO[0m	logger/middleware.go:116	HTTP request processed	{"request_id": "550e8400-e29b-41d4-a716-446655440000", "method": "POST", "path": "/api/schedules", "status": 200, "duration_ms": 351}
```

### Production Mode Output

```json
{"level":"debug","timestamp":"2025-11-15T16:52:14.792-0500","caller":"logger/logger.go:120","message":"Processing schedule request","request_id":"550e8400-e29b-41d4-a716-446655440000","schedule_id":"sch-123"}
{"level":"info","timestamp":"2025-11-15T16:52:14.812-0500","caller":"logger/logger.go:150","message":"Schedule processing completed","request_id":"550e8400-e29b-41d4-a716-446655440000","assignments":12}
{"level":"warn","timestamp":"2025-11-15T16:52:15.120-0500","caller":"logger/logger.go:180","message":"Service call slow","service":"user-service","duration_ms":2500}
{"level":"info","timestamp":"2025-11-15T16:52:15.145-0500","caller":"logger/middleware.go:116","message":"HTTP request processed","request_id":"550e8400-e29b-41d4-a716-446655440000","method":"POST","path":"/api/schedules","status":200,"duration_ms":351}
```
