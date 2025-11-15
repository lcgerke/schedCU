# Logger Configuration Guide

This document describes configuration options and best practices for the logging framework.

## Environment Variables

### APP_ENV

Controls the logging mode and configuration.

```bash
# Development mode (human-readable console output)
export APP_ENV=development

# Production mode (JSON output)
export APP_ENV=production
```

**Default**: `production` (if not set)

**Valid Values**:
- `development` or `dev`: Development configuration
- Any other value defaults to production configuration

## Initialization

### From Environment Variable

When no argument is provided, `NewLogger()` reads from the `APP_ENV` environment variable:

```go
logger, err := logger.NewLogger("")
// Uses APP_ENV environment variable
// Falls back to production if not set
```

### Explicit Environment

Override environment by passing it explicitly:

```go
logger, err := logger.NewLogger("development")
// Always use development configuration
```

## Configuration Profiles

### Development Profile

**When to use**: Local development, debugging, testing

```go
logger, _ := logger.NewLogger("development")
```

**Characteristics**:

| Setting | Value |
|---------|-------|
| Output | Console (stdout) |
| Format | Human-readable with ANSI colors |
| Level | Debug and above |
| Caller | Included with file:line |
| Stack traces | Included for errors and warnings |
| Timestamps | Human-readable format |
| Performance | Not optimized (acceptable for dev) |

**Sample Output**:

```
2025-11-15T16:52:14.792-0500	[35mDEBUG[0m	logger/logger.go:120	Fetching schedule	{"schedule_id": "sch-123"}
2025-11-15T16:52:14.812-0500	[34mINFO[0m	logger/logger.go:150	Schedule found	{"schedule_id": "sch-123"}
```

**Use Cases**:
- `LOCAL_DEV`: Running locally with `make run`
- `TESTING`: Running tests with `go test`
- `DEBUGGING`: Investigating production issues locally

### Production Profile

**When to use**: Staging, production, performance-critical environments

```go
logger, _ := logger.NewLogger("production")
```

**Characteristics**:

| Setting | Value |
|---------|-------|
| Output | stdout (JSON) |
| Format | JSON for log aggregation systems |
| Level | Info and above |
| Caller | Included with filename:line |
| Stack traces | Only for panics/severe errors |
| Timestamps | ISO8601 format |
| Performance | Optimized for throughput |

**Sample Output**:

```json
{
  "level": "info",
  "timestamp": "2025-11-15T16:52:14.792-0500",
  "caller": "logger/logger.go:150",
  "message": "Schedule found",
  "schedule_id": "sch-123"
}
```

**Use Cases**:
- `STAGING`: Pre-production environment
- `PRODUCTION`: Live environment
- `CI_CD`: Continuous integration/deployment pipelines

## Log Levels

### Development Mode

**Available Levels** (in order of verbosity):
1. **DEBUG** - Detailed diagnostic information
2. **INFO** - General informational messages
3. **WARN** - Warning messages for potentially problematic situations
4. **ERROR** - Error messages for failures

**All levels are captured** in development mode.

### Production Mode

**Available Levels** (in order of verbosity):
1. **INFO** - General informational messages
2. **WARN** - Warning messages for potentially problematic situations
3. **ERROR** - Error messages for failures

**DEBUG messages are suppressed** in production mode for performance.

## Output Destinations

### Standard Output (stdout)

All log messages go to stdout:

```bash
# Capture logs to file
go run main.go > logs.txt 2>&1

# Real-time log monitoring
go run main.go | jq '.'  # Pretty-print JSON in production
```

### Standard Error (stderr)

Error messages specifically are also written to stderr:

```bash
# Separate error logs
go run main.go 2> errors.txt
```

## JSON Output Structure (Production)

The JSON format is optimized for log aggregation systems like Datadog, ELK Stack, and Splunk.

### Standard Fields

Every log entry includes:

```json
{
  "level": "info|warn|error|debug",
  "timestamp": "2025-11-15T16:52:14.792-0500",
  "caller": "package/file.go:123",
  "message": "Log message here"
}
```

### Custom Fields

Additional fields are appended as JSON properties:

```json
{
  "level": "info",
  "timestamp": "2025-11-15T16:52:14.792-0500",
  "caller": "logger/logger.go:150",
  "message": "Processing request",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/schedules",
  "status": 200
}
```

### Structured Logging Format

When using `Infow`, `Warnw`, `Errorw`:

```go
logger.Infow("message here",
    "key1", value1,
    "key2", value2,
)

// Produces JSON:
// {
//   "level": "info",
//   "timestamp": "...",
//   "caller": "...",
//   "message": "message here",
//   "key1": value1,
//   "key2": value2
// }
```

## Performance Tuning

### Development Mode

Development mode is not optimized for performance, which is acceptable since it's used locally.

**Performance characteristics**:
- ~5-10ms per log call (acceptable for debugging)
- Console I/O is synchronous (human-readable)
- Stack traces are always collected

### Production Mode

Production mode is heavily optimized:

**Performance characteristics**:
- ~100 microseconds per log call
- Asynchronous buffering reduces I/O overhead
- Minimal allocations using zap's object pools
- Suitable for high-throughput systems (10K+ req/sec)

**Tips for optimal performance**:

1. Always call `defer log.Sync()` in main:
   ```go
   func main() {
       log, _ := logger.NewLogger("production")
       defer log.Sync()  // Flush buffered logs
   }
   ```

2. Avoid expensive operations in log fields:
   ```go
   // Bad - expensive function called every time
   logger.Infow("message", "result", expensiveFunction())

   // Good - only called when needed
   if log.Check(zapcore.InfoLevel, "") != nil {
       logger.Infow("message", "result", expensiveFunction())
   }
   ```

3. Prefer structured logging:
   ```go
   // Bad - string concatenation
   logger.Info("User " + userID + " created schedule " + scheduleID)

   // Good - structured fields
   logger.Infow("User created schedule", "user_id", userID, "schedule_id", scheduleID)
   ```

## Integration with Log Aggregation

### Datadog

Configure Datadog agent to ingest JSON logs:

```yaml
# datadog.yaml
logs:
  enabled: true
  config:
    - type: file
      path: /var/log/app/*.json
      service: schedcu
      source: go
      parser: json
```

Query in Datadog:
```
service:schedcu status:error
service:schedcu request_id:550e8400-e29b-41d4-a716-446655440000
```

### ELK Stack (Elasticsearch, Logstash, Kibana)

Filebeat configuration:

```yaml
# filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /var/log/app/*.json
    json.message_key: message
    json.keys_under_root: true

output.elasticsearch:
  hosts: ["localhost:9200"]
```

### Splunk

Forward JSON logs:

```bash
# Add to Splunk inputs.conf
[default]
SHOULD_LINEMERGE = false
DATETIME_CONFIG = CURRENT

[http://schedcu-logs]
source = http://localhost:8088
```

### AWS CloudWatch

Use CloudWatch Logs agent:

```json
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/app/*.json",
            "log_group_name": "/aws/schedcu/logs",
            "log_stream_name": "app-logs",
            "timestamp_format": "%Y-%m-%dT%H:%M:%S"
          }
        ]
      }
    }
  }
}
```

## Docker Configuration

### Development Dockerfile

```dockerfile
FROM golang:1.20-alpine

WORKDIR /app
COPY . .

# Run in development mode
ENV APP_ENV=development
CMD ["go", "run", "cmd/server/main.go"]
```

### Production Dockerfile

```dockerfile
FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .

# Run in production mode
ENV APP_ENV=production
CMD ["./server"]
```

### Docker Compose

```yaml
version: '3'

services:
  app:
    build: .
    environment:
      - APP_ENV=production
      - LOG_LEVEL=info
    ports:
      - "8080:8080"
    volumes:
      - ./logs:/var/log/app
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Kubernetes Configuration

### Deployment with Environment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: schedcu-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: schedcu
  template:
    metadata:
      labels:
        app: schedcu
    spec:
      containers:
      - name: app
        image: schedcu:latest
        env:
        - name: APP_ENV
          value: "production"
        - name: LOG_LEVEL
          value: "info"
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

### Fluentd Integration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/app/*.json
      pos_file /var/log/app.log.pos
      tag schedcu.app
      format json
    </source>

    <match schedcu.**>
      @type datadog
      @id output_datadog
      api_key "#{ENV['DD_API_KEY']}"
      provider gcp
      service schedcu
    </match>
```

## Testing

### Test Environment Setup

```go
func setupTestLogger(t *testing.T) *zap.SugaredLogger {
    log, err := logger.NewLogger("development")
    if err != nil {
        t.Fatalf("Failed to create logger: %v", err)
    }
    t.Cleanup(func() {
        log.Sync()
    })
    return log
}

func TestWithLogger(t *testing.T) {
    log := setupTestLogger(t)
    // Test code here
}
```

### Capturing Log Output

```go
// Capture JSON output
var buf bytes.Buffer
w := os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
// Redirect stdout and capture
// (Complex in practice, consider using log mocking libraries)
```

## Best Practices

### 1. Consistent Environment Configuration

Always explicitly set `APP_ENV` in your deployment:

```bash
# Good
export APP_ENV=production
go run main.go

# Avoid (defaults to production, less explicit)
go run main.go
```

### 2. Use Structured Logging

```go
// Bad
logger.Info("User " + userID + " logged in from " + ipAddress)

// Good
logger.Infow("User logged in",
    "user_id", userID,
    "ip_address", ipAddress,
)
```

### 3. Include Context IDs

```go
// Always include RequestID for tracing
logger.Infow("Processing request",
    "request_id", logger.ExtractRequestID(ctx),
    "user_id", userID,
)
```

### 4. Error Logging with Context

```go
// Provide context when logging errors
if err != nil {
    logger.LogError(log, err, map[string]interface{}{
        "operation": "create_schedule",
        "schedule_name": req.Name,
        "user_id": req.UserID,
    })
}
```

### 5. Graceful Shutdown

```go
func main() {
    log, _ := logger.NewLogger("production")
    defer log.Sync()  // Always sync before exit

    // Application code
}
```

## Troubleshooting

### Logs Not Appearing

1. **Check environment**: Verify `APP_ENV` is set correctly
2. **Check level**: Development mode shows DEBUG, production shows INFO+
3. **Call Sync()**: In main, ensure `defer log.Sync()` is called
4. **Check output**: Logs go to stdout, not stderr (except errors)

### Performance Issues

1. **Avoid expensive operations**: Don't call heavy functions in log fields
2. **Use buffering**: Production mode buffers, development mode doesn't
3. **Reduce log level**: Production mode suppresses DEBUG messages

### JSON Parsing Issues

1. **Check format**: Production mode outputs one JSON object per line
2. **Line endings**: Each log entry is a complete JSON object on one line
3. **Tool compatibility**: Use `jq` or JSON-aware log tools for parsing

## Reference

### Log Output Formats

**Development Format**:
```
TIMESTAMP[LEVEL]	CALLER	MESSAGE	{FIELDS}
```

**Production Format**:
```json
{"level":"...","timestamp":"...","caller":"...","message":"...","field":"value"}
```

### Performance Metrics

| Operation | Development | Production |
|-----------|-------------|------------|
| Logger creation | <1ms | <1ms |
| Simple log call | 5-10ms | 100-200μs |
| Structured log (5 fields) | 10-15ms | 200-400μs |
| Middleware overhead per request | 1-2ms | 100-300μs |

### File Locations

| File | Purpose |
|------|---------|
| `/internal/logger/logger.go` | Core logger implementation |
| `/internal/logger/middleware.go` | HTTP middleware |
| `/internal/logger/logger_test.go` | Logger tests |
| `/internal/logger/middleware_test.go` | Middleware tests |
| `/internal/logger/README.md` | API documentation |
| `/internal/logger/EXAMPLES.md` | Usage examples |
| `/internal/logger/CONFIG.md` | This file |
