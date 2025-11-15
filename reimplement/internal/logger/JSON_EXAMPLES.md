# JSON Log Output Examples

This document shows real JSON output from the logger in production mode. These examples are useful for:
- Understanding log schema for log aggregation system configuration
- Parsing logs in downstream systems
- Setting up alerts and dashboards
- Filtering and searching logs

## Basic Examples

### Simple Info Message

```json
{
  "level": "info",
  "timestamp": "2025-11-15T10:30:45.123Z",
  "caller": "logger/logger_test.go:39",
  "message": "test message"
}
```

### Structured Fields

```json
{
  "level": "info",
  "timestamp": "2025-11-15T10:30:45.456Z",
  "caller": "internal/api/schedule_handler.go:42",
  "message": "Processing schedule request",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user-12345",
  "schedule_id": "sch-98765"
}
```

### Error Message

```json
{
  "level": "error",
  "timestamp": "2025-11-15T10:30:46.789Z",
  "caller": "internal/logger/logger.go:145",
  "message": "Error occurred",
  "error": "database connection failed",
  "operation": "create_schedule",
  "status": 500,
  "retry_count": 3
}
```

## HTTP Request Logging

### Successful GET Request

```json
{
  "level": "info",
  "timestamp": "2025-11-15T10:30:45.123Z",
  "caller": "internal/logger/middleware.go:116",
  "message": "HTTP request processed",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/schedules",
  "status": 200,
  "duration_ms": 45
}
```

### Failed POST Request

```json
{
  "level": "error",
  "timestamp": "2025-11-15T10:30:46.234Z",
  "caller": "internal/logger/middleware.go:108",
  "message": "HTTP request processed",
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "method": "POST",
  "path": "/api/schedules",
  "status": 500,
  "duration_ms": 250
}
```

### Not Found Error

```json
{
  "level": "error",
  "timestamp": "2025-11-15T10:30:47.345Z",
  "caller": "internal/logger/middleware.go:108",
  "message": "HTTP request processed",
  "request_id": "550e8400-e29b-41d4-a716-446655440002",
  "method": "GET",
  "path": "/api/schedules/nonexistent",
  "status": 404,
  "duration_ms": 12
}
```

## Service Call Logging

### Successful Service Call

```json
{
  "level": "info",
  "timestamp": "2025-11-15T10:30:48.456Z",
  "caller": "internal/logger/logger.go:166",
  "message": "Service call succeeded",
  "service": "user-service",
  "operation": "GetUser",
  "duration_ms": 120
}
```

### Failed Service Call

```json
{
  "level": "error",
  "timestamp": "2025-11-15T10:30:49.567Z",
  "caller": "internal/logger/logger.go:157",
  "message": "Service call failed",
  "service": "notification-service",
  "operation": "SendEmail",
  "duration_ms": 5000,
  "error": "service timeout"
}
```

## Complex Scenarios

### Multi-Field Request Processing

```json
{
  "level": "info",
  "timestamp": "2025-11-15T10:30:50.678Z",
  "caller": "internal/service/schedule_service.go:123",
  "message": "Creating schedule with assignments",
  "request_id": "550e8400-e29b-41d4-a716-446655440003",
  "correlation_id": "corr-12345-67890",
  "user_id": "user-12345",
  "hospital_id": "hosp-54321",
  "schedule_name": "January 2024 Schedule",
  "num_assignments": 45,
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
```

### Database Operation Logging

```json
{
  "level": "info",
  "timestamp": "2025-11-15T10:30:51.789Z",
  "caller": "internal/repository/schedule_repository.go:87",
  "message": "Schedule saved to database",
  "request_id": "550e8400-e29b-41d4-a716-446655440003",
  "schedule_id": "sch-uuid-12345",
  "table": "schedules",
  "operation": "INSERT",
  "duration_ms": 23,
  "rows_affected": 1
}
```

### Cascading Operations with Context

```json
{
  "level": "info",
  "timestamp": "2025-11-15T10:30:52.890Z",
  "caller": "internal/service/schedule_service.go:156",
  "message": "Assigning staff to shifts",
  "request_id": "550e8400-e29b-41d4-a716-446655440003",
  "correlation_id": "corr-12345-67890",
  "schedule_id": "sch-uuid-12345",
  "assignments_count": 45,
  "current_index": 12,
  "staff_member": "Jane Doe",
  "shift_id": "shift-uuid-67890",
  "position": "ER Doctor"
}
```

### Performance Warning

```json
{
  "level": "warn",
  "timestamp": "2025-11-15T10:30:53.901Z",
  "caller": "internal/service/schedule_service.go:234",
  "message": "Slow database query detected",
  "request_id": "550e8400-e29b-41d4-a716-446655440003",
  "query": "SELECT * FROM shift_instances WHERE schedule_id = ?",
  "duration_ms": 2500,
  "threshold_ms": 1000,
  "overbudget_ms": 1500,
  "query_hash": "abc123def456"
}
```

## Log Aggregation Patterns

### Pattern: Request Tracing

All logs related to a single request share the same `request_id`:

```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Request started"
}

{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Database query executed"
}

{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Service call completed"
}

{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Response sent"
}
```

### Pattern: Distributed Tracing

Related requests across services share `correlation_id`:

**Service A**:
```json
{
  "request_id": "req-111",
  "correlation_id": "corr-123",
  "message": "Calling user-service",
  "service": "schedule-service"
}
```

**Service B (user-service)**:
```json
{
  "request_id": "req-222",
  "correlation_id": "corr-123",
  "message": "Processing request from schedule-service",
  "service": "user-service"
}
```

**Service A (continued)**:
```json
{
  "request_id": "req-111",
  "correlation_id": "corr-123",
  "message": "user-service call completed",
  "service": "schedule-service"
}
```

### Pattern: Error Context

```json
{
  "level": "error",
  "message": "Schedule creation failed",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "error": "duplicate schedule for date range",
  "error_code": "SCHEDULE_CONFLICT",
  "error_details": {
    "existing_schedule_id": "sch-existing-123",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "conflict_type": "date_overlap"
  },
  "operation": "create_schedule",
  "user_id": "user-12345",
  "hospital_id": "hosp-54321"
}
```

## Field Reference

### Always Present Fields

| Field | Type | Description |
|-------|------|-------------|
| `level` | string | Log level: debug, info, warn, error |
| `timestamp` | string | ISO8601 timestamp |
| `caller` | string | File and line: package/file.go:123 |
| `message` | string | Main log message |

### Common Custom Fields

| Field | Type | Description |
|-------|------|-------------|
| `request_id` | string | UUID for request tracing |
| `correlation_id` | string | UUID for distributed tracing |
| `user_id` | string | User identifier |
| `operation` | string | Operation name |
| `error` | string | Error message |
| `duration_ms` | integer | Duration in milliseconds |

### HTTP Middleware Fields

| Field | Type | Description |
|-------|------|-------------|
| `method` | string | HTTP method (GET, POST, etc.) |
| `path` | string | Request path |
| `status` | integer | HTTP status code |
| `duration_ms` | integer | Request duration in milliseconds |

## Querying Examples

### Datadog

```
# All errors for a request
service:schedcu request_id:550e8400 level:error

# Slow requests
service:schedcu duration_ms:[1000 TO 999999]

# Specific operation failures
service:schedcu operation:create_schedule level:error

# Service call errors
service:schedcu message:"Service call failed" service:user-service
```

### Elasticsearch/Kibana

```
# All errors
level:error AND service:schedcu

# Request timeline
request_id:550e8400-e29b-41d4-a716-446655440000

# Errors by operation
level:error AND operation:*

# Slow database queries
duration_ms:[2000 TO *]
```

### Splunk

```
service=schedcu level=error
| stats count by operation

service=schedcu request_id=550e8400-e29b-41d4-a716-446655440000
| timechart count

service=schedcu duration_ms > 1000
| top limit=20 path
```

## JSON Schema

This schema describes the structure of log entries:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["level", "timestamp", "caller", "message"],
  "properties": {
    "level": {
      "type": "string",
      "enum": ["debug", "info", "warn", "error"]
    },
    "timestamp": {
      "type": "string",
      "format": "date-time"
    },
    "caller": {
      "type": "string",
      "pattern": "^.+\\.go:\\d+$"
    },
    "message": {
      "type": "string"
    },
    "request_id": {
      "type": "string",
      "format": "uuid"
    },
    "correlation_id": {
      "type": "string",
      "format": "uuid"
    },
    "error": {
      "type": "string"
    },
    "duration_ms": {
      "type": "integer",
      "minimum": 0
    }
  },
  "additionalProperties": true
}
```

## Performance Impact

These JSON logs are optimized for:
- **Log aggregation tools** - Direct JSON parsing, no additional processing
- **Search performance** - Pre-structured fields for efficient indexing
- **Storage efficiency** - Compact binary JSON representation
- **Parsing speed** - No escaping or transformation needed

## Log Sampling

For high-traffic applications, consider sampling:

```go
// Sample 10% of INFO logs in production
if logger.InfoLevel && rand.Float64() > 0.1 {
    return
}
```

This reduces storage costs while maintaining full error logs.

## Retention Recommendations

| Log Level | Retention | Reason |
|-----------|-----------|--------|
| DEBUG | 1 week | High volume, development focused |
| INFO | 30 days | Standard operational logs |
| WARN | 90 days | Important events requiring review |
| ERROR | 180 days+ | Critical for debugging production issues |

## Integration with Grafana

```json
{
  "datasource": "Datadog",
  "targets": [
    {
      "expr": "service:schedcu level:error",
      "refId": "A"
    }
  ],
  "dashboard": "Schedule Application Monitoring"
}
```

Create dashboards tracking:
- Error rate by operation
- Request duration percentiles
- Service call failures
- Database query performance
- Request tracing for specific request_ids
