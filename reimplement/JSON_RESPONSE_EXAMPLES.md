# JSON Response Examples for ApiResponse[T]

**Reference**: Work Package [2.2] - ApiResponse Tests
**Location**: `internal/api/response.go` and `internal/api/response_test.go`
**Date**: November 15, 2025

---

## Example 1: Success Response with Data

### Go Code
```go
type Schedule struct {
    ID        string `json:"id"`
    StartDate string `json:"start_date"`
    Duration  int    `json:"duration"`
}

schedule := Schedule{
    ID:        "sched-123",
    StartDate: "2025-11-15",
    Duration:  480,
}

resp := NewApiResponse(schedule)
c.JSON(200, resp)
```

### JSON Output
```json
{
  "data": {
    "id": "sched-123",
    "start_date": "2025-11-15",
    "duration": 480
  },
  "validation": {
    "errors": [],
    "warnings": [],
    "infos": [],
    "context": {}
  },
  "meta": {
    "timestamp": "2025-11-15T17:30:37.360613303-05:00",
    "request_id": "930883da-3446-4489-ae1c-143cec9a1b8a",
    "version": "1.0",
    "server_time": 1763245837
  }
}
```

### Key Characteristics
- ✅ Data contains the actual response payload
- ✅ Validation included (empty arrays indicate no issues)
- ✅ Error field omitted (null omission)
- ✅ Meta contains automatic request tracking
- ✅ Timestamp in RFC3339 format
- ✅ Unique request ID for tracing

---

## Example 2: Success Response with Validation Warnings

### Go Code
```go
schedule := Schedule{...}

vr := validation.NewValidationResult()
vr.AddWarning("duration", "Duration exceeds typical shift (>8 hours)")
vr.SetContext("typical_duration", 480)
vr.SetContext("max_duration", 600)

resp := NewApiResponse(schedule).WithValidation(vr)
c.JSON(200, resp)
```

### JSON Output
```json
{
  "data": {
    "id": "sched-123",
    "start_date": "2025-11-15",
    "duration": 600
  },
  "validation": {
    "errors": [],
    "warnings": [
      {
        "field": "duration",
        "message": "Duration exceeds typical shift (>8 hours)"
      }
    ],
    "infos": [],
    "context": {
      "typical_duration": 480,
      "max_duration": 600
    }
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ Data returned successfully (200 status)
- ✅ Warnings included (non-blocking issues)
- ✅ Context data for debugging included
- ✅ IsSuccess() returns true (no errors)
- ✅ Client can proceed with operation

---

## Example 3: Validation Error Response

### Go Code
```go
vr := validation.NewValidationResult()
vr.AddError("email", "Email address is invalid")
vr.AddError("age", "Age must be between 18 and 120")
vr.AddWarning("phone", "Phone format is unusual for region")

resp := NewApiResponse("").WithValidation(vr)
c.JSON(400, resp)
```

### JSON Output
```json
{
  "data": "",
  "validation": {
    "errors": [
      {
        "field": "email",
        "message": "Email address is invalid"
      },
      {
        "field": "age",
        "message": "Age must be between 18 and 120"
      }
    ],
    "warnings": [
      {
        "field": "phone",
        "message": "Phone format is unusual for region"
      }
    ],
    "infos": [],
    "context": {}
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 400 Bad Request status
- ✅ Multiple validation errors included
- ✅ Field-level error messages
- ✅ Warnings included for non-blocking issues
- ✅ IsSuccess() returns false
- ✅ Client can identify specific field issues

---

## Example 4: Error Response with Details

### Go Code
```go
err := fmt.Errorf("connection timeout after 30 seconds")

details := map[string]interface{}{
    "operation":       "create_schedule",
    "resource_type":   "Schedule",
    "retry_after_sec": 5,
    "error_detail":    err.Error(),
}

resp := NewApiResponse("").
    WithError("EXTERNAL_SERVICE_ERROR", "Failed to contact scheduling service").
    WithErrorDetails(details)

c.JSON(503, resp)
```

### JSON Output
```json
{
  "data": "",
  "validation": {
    "errors": [],
    "warnings": [],
    "infos": [],
    "context": {}
  },
  "error": {
    "code": "EXTERNAL_SERVICE_ERROR",
    "message": "Failed to contact scheduling service",
    "details": {
      "operation": "create_schedule",
      "resource_type": "Schedule",
      "retry_after_sec": 5,
      "error_detail": "connection timeout after 30 seconds"
    }
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 503 Service Unavailable status
- ✅ Machine-readable error code
- ✅ Human-readable error message
- ✅ Details provide debugging context
- ✅ Retry guidance included
- ✅ IsSuccess() returns false

---

## Example 5: Authentication Error

### Go Code
```go
resp := NewApiResponse("").
    WithError("UNAUTHORIZED", "Authentication token missing or invalid")

c.JSON(401, resp)
```

### JSON Output
```json
{
  "data": "",
  "validation": {
    "errors": [],
    "warnings": [],
    "infos": [],
    "context": {}
  },
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication token missing or invalid"
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 401 Unauthorized status
- ✅ No error details (sensitive information protection)
- ✅ Clear error code for client handling
- ✅ IsSuccess() returns false

---

## Example 6: Authorization Denied

### Go Code
```go
resp := NewApiResponse("").
    WithError("FORBIDDEN", "User does not have permission to access this resource")

c.JSON(403, resp)
```

### JSON Output
```json
{
  "data": "",
  "validation": {...},
  "error": {
    "code": "FORBIDDEN",
    "message": "User does not have permission to access this resource"
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 403 Forbidden status
- ✅ Distinguished from 401 (authentication vs authorization)
- ✅ Clear error message
- ✅ IsSuccess() returns false

---

## Example 7: Resource Not Found

### Go Code
```go
resp := NewApiResponse("").
    WithError("NOT_FOUND", "Schedule with ID 'sched-999' not found")

c.JSON(404, resp)
```

### JSON Output
```json
{
  "data": "",
  "validation": {...},
  "error": {
    "code": "NOT_FOUND",
    "message": "Schedule with ID 'sched-999' not found"
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 404 Not Found status
- ✅ Clear resource identification
- ✅ Helpful error message
- ✅ IsSuccess() returns false

---

## Example 8: Conflict/Duplicate

### Go Code
```go
details := map[string]interface{}{
    "conflicting_id": "sched-123",
    "existing_date":  "2025-11-15",
}

resp := NewApiResponse("").
    WithError("CONFLICT", "A schedule already exists for this date").
    WithErrorDetails(details)

c.JSON(409, resp)
```

### JSON Output
```json
{
  "data": "",
  "validation": {...},
  "error": {
    "code": "CONFLICT",
    "message": "A schedule already exists for this date",
    "details": {
      "conflicting_id": "sched-123",
      "existing_date": "2025-11-15"
    }
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 409 Conflict status
- ✅ Details help client understand the conflict
- ✅ Clear error code
- ✅ IsSuccess() returns false

---

## Example 9: Rate Limited

### Go Code
```go
details := map[string]interface{}{
    "limit":        100,
    "current":      101,
    "reset_after":  60,
    "retry_after":  60,
}

resp := NewApiResponse("").
    WithError("RATE_LIMITED", "Request rate limit exceeded").
    WithErrorDetails(details)

c.JSON(429, resp)
```

### JSON Output
```json
{
  "data": "",
  "validation": {...},
  "error": {
    "code": "RATE_LIMITED",
    "message": "Request rate limit exceeded",
    "details": {
      "limit": 100,
      "current": 101,
      "reset_after": 60,
      "retry_after": 60
    }
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 429 Too Many Requests status
- ✅ Rate limit details for client backoff
- ✅ Retry guidance
- ✅ IsSuccess() returns false

---

## Example 10: Server Error with Details

### Go Code
```go
err := database.ErrConnectionFailed
details := map[string]interface{}{
    "operation":    "INSERT",
    "table":        "schedules",
    "error_type":   "DatabaseError",
    "error_msg":    err.Error(),
}

resp := NewApiResponse("").
    WithError("DATABASE_ERROR", "Failed to save schedule").
    WithErrorDetails(details)

c.JSON(500, resp)
```

### JSON Output
```json
{
  "data": "",
  "validation": {...},
  "error": {
    "code": "DATABASE_ERROR",
    "message": "Failed to save schedule",
    "details": {
      "operation": "INSERT",
      "table": "schedules",
      "error_type": "DatabaseError",
      "error_msg": "connection failed: timeout"
    }
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 500 Internal Server Error status
- ✅ Machine-readable error code
- ✅ Safe error message (no sensitive details)
- ✅ Debug details logged separately
- ✅ IsSuccess() returns false

---

## Example 11: Partial Success (Multi-Status)

### Go Code
```go
successCount := 7
failureCount := 3
items := []interface{}{...}

details := map[string]interface{}{
    "total":    10,
    "success":  7,
    "failed":   3,
    "failed_ids": []string{"item-2", "item-5", "item-8"},
}

resp := NewApiResponse(items).
    WithError("PARTIAL_FAILURE", "Some items failed to process").
    WithErrorDetails(details)

c.JSON(207, resp)
```

### JSON Output
```json
{
  "data": [
    {...},
    {...}
  ],
  "validation": {...},
  "error": {
    "code": "PARTIAL_FAILURE",
    "message": "Some items failed to process",
    "details": {
      "total": 10,
      "success": 7,
      "failed": 3,
      "failed_ids": ["item-2", "item-5", "item-8"]
    }
  },
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 207 Multi-Status
- ✅ Data returned even with partial failure
- ✅ Clear success/failure counts
- ✅ Failed item identification
- ✅ IsSuccess() returns false (has error)

---

## Example 12: Array/Slice Response

### Go Code
```go
type ShiftSummary struct {
    Date     string `json:"date"`
    Assigned int    `json:"assigned"`
    Required int    `json:"required"`
}

shifts := []ShiftSummary{
    {Date: "2025-11-15", Assigned: 8, Required: 8},
    {Date: "2025-11-16", Assigned: 6, Required: 8},
    {Date: "2025-11-17", Assigned: 8, Required: 8},
}

resp := NewApiResponse(shifts)
c.JSON(200, resp)
```

### JSON Output
```json
{
  "data": [
    {
      "date": "2025-11-15",
      "assigned": 8,
      "required": 8
    },
    {
      "date": "2025-11-16",
      "assigned": 6,
      "required": 8
    },
    {
      "date": "2025-11-17",
      "assigned": 8,
      "required": 8
    }
  ],
  "validation": {...},
  "meta": {...}
}
```

### Key Characteristics
- ✅ Array of objects as data
- ✅ Type-safe with generics
- ✅ Clean JSON structure
- ✅ Metadata included for tracing

---

## Example 13: Map/Dictionary Response

### Go Code
```go
coverageByPosition := map[string]interface{}{
    "ER Doctor": map[string]interface{}{
        "required": 4,
        "assigned": 3,
        "coverage": 0.75,
    },
    "Nurse": map[string]interface{}{
        "required": 10,
        "assigned": 9,
        "coverage": 0.90,
    },
}

resp := NewApiResponse(coverageByPosition)
c.JSON(200, resp)
```

### JSON Output
```json
{
  "data": {
    "ER Doctor": {
      "required": 4,
      "assigned": 3,
      "coverage": 0.75
    },
    "Nurse": {
      "required": 10,
      "assigned": 9,
      "coverage": 0.90
    }
  },
  "validation": {...},
  "meta": {...}
}
```

### Key Characteristics
- ✅ Map/dictionary as data
- ✅ Nested objects supported
- ✅ Numeric values preserved
- ✅ Type-safe with T=map[string]interface{}

---

## Example 14: Empty/No Data Response

### Go Code
```go
// Delete successful, no data to return
resp := NewApiResponse(nil)
c.JSON(204, resp)
```

### JSON Output
```json
{
  "data": null,
  "validation": {...},
  "meta": {...}
}
```

### Key Characteristics
- ✅ HTTP 204 No Content status
- ✅ Null data field acceptable
- ✅ Metadata still included for tracing
- ✅ IsSuccess() returns true (no error)

---

## Example 15: Complex Nested Structure

### Go Code
```go
type ScheduleDetails struct {
    ID    string `json:"id"`
    Dates []struct {
        Date   string `json:"date"`
        Shifts []struct {
            Position string `json:"position"`
            Staff    string `json:"staff"`
            Hours    int    `json:"hours"`
        } `json:"shifts"`
    } `json:"dates"`
}

details := ScheduleDetails{...}
resp := NewApiResponse(details)
c.JSON(200, resp)
```

### JSON Output
```json
{
  "data": {
    "id": "sched-123",
    "dates": [
      {
        "date": "2025-11-15",
        "shifts": [
          {
            "position": "ER Doctor",
            "staff": "Dr. Smith",
            "hours": 8
          },
          {
            "position": "Nurse",
            "staff": "Jane Doe",
            "hours": 8
          }
        ]
      }
    ]
  },
  "validation": {...},
  "meta": {...}
}
```

### Key Characteristics
- ✅ Deep nesting supported
- ✅ Arrays of objects supported
- ✅ All levels properly serialized
- ✅ Type-safe with custom struct

---

## Key Field Reference

### Data Field
- **Type**: Generic T (inferred from constructor argument)
- **Required**: Yes, but can be nil or zero value
- **Omitted**: No (always included)
- **Examples**: Object, array, string, number, null

### Validation Field
- **Type**: `*validation.ValidationResult`
- **Required**: Always initialized (never nil)
- **Omitted**: No (always included)
- **Contents**: Errors array, warnings array, infos array, context map

### Error Field
- **Type**: `*ErrorDetail` (includes Code, Message, Details)
- **Required**: No (only when error occurred)
- **Omitted**: Yes (when nil, via `omitempty` tag)
- **Examples**: INVALID_REQUEST, NOT_FOUND, DATABASE_ERROR

### Meta Field
- **Type**: `*ResponseMeta` (includes Timestamp, RequestID, Version, ServerTime)
- **Required**: Yes (always initialized)
- **Omitted**: No (always included)
- **Auto-populated**: RequestID is UUID, timestamp is current time

---

## Field Omission Rules Summary

```
Field Name   | Omitted When     | Always Included?
-------------|------------------|------------------
data         | Never            | Yes
validation   | Never            | Yes
error        | null/nil         | Only when set
meta         | Never            | Yes
error.code   | -                | When error set
error.details| empty/nil        | Only when set
```

---

## HTTP Status Code Mapping

```
Response Type              | Status Code | Typical Use
--------------------------|-------------|------------------------
Success with data         | 200         | GET, POST, PUT
Created                   | 201         | POST (new resource)
Accepted                  | 202         | Long-running operations
No Content                | 204         | DELETE, empty responses
Multi-Status              | 207         | Batch with partial success
Bad Request               | 400         | Validation errors
Unauthorized              | 401         | Auth required
Forbidden                 | 403         | Auth failed
Not Found                 | 404         | Resource missing
Conflict                  | 409         | Duplicate/conflict
Rate Limited              | 429         | Too many requests
Server Error              | 500         | Unhandled exceptions
Service Unavailable       | 503         | External service down
```

---

**Reference Implementation**: See `internal/api/response_test.go` for working examples
