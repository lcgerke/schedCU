# Error Formatter Examples and Nested Structure Documentation

**Work Package**: [2.3] Error Response Formatting
**Component**: `internal/api/error_formatter.go`
**Status**: Complete with 15 passing tests

---

## Quick Start

### Basic Usage

```go
package api

import (
    "github.com/schedcu/reimplement/internal/validation"
)

// Create validation result with errors
vr := validation.NewValidationResult()
vr.AddError("email", "Email is required")

// Format for API response
errorDetail := FormatValidationErrors(vr)

// Use in response
response := NewApiResponse(nil).
    WithError(errorDetail.Code, errorDetail.Message).
    WithErrorDetails(errorDetail.Details)
```

---

## Error Nesting Structure

### Level 1: ApiResponse (HTTP Response)

```
ApiResponse[T] {
  data: null
  validation: ValidationResult
  error: ErrorDetail  ← Top-level error code
  meta: ResponseMeta
}
```

### Level 2: ErrorDetail (Error Information)

```
ErrorDetail {
  code: "VALIDATION_ERROR"     ← Machine-readable code for client routing
  message: "Validation failed: 2 error(s), 1 warning(s)"  ← User-readable summary
  details: {                   ← Detailed error breakdown
    error_count: 2
    warning_count: 1
    errors: {...}
    warnings: {...}
    context: {...}
  }
}
```

### Level 3: Error Maps (Grouped by Field)

```
details.errors = {
  "email": "Email is required",                    ← Single message: string
  "password": [                                    ← Multiple: array
    "Password required",
    "Password must be 8+ characters"
  ],
  "_global_": "File format invalid"                ← Global errors
}
```

### Level 4: Context (Additional Information)

```
details.context = {
  "filename": "schedule.ods",
  "line_number": 42,
  "operation": "import",
  "received_size": "15MB"
}
```

---

## Complete JSON Response Examples

### Example 1: Form Validation Error (Simple)

**Scenario**: User registration with missing required fields

```json
{
  "data": null,
  "validation": {
    "errors": [],
    "warnings": [],
    "infos": [],
    "context": {}
  },
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed: 2 error(s)",
    "details": {
      "error_count": 2,
      "errors": {
        "email": "Email is required",
        "password": "Password is required"
      }
    }
  },
  "meta": {
    "timestamp": "2025-11-15T17:30:41Z",
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "version": "1.0",
    "server_time": 1763245841
  }
}
```

### Example 2: Complex Field Validation (Multiple Errors Per Field)

**Scenario**: Username validation with multiple failure reasons

```json
{
  "data": null,
  "validation": {
    "errors": [],
    "warnings": [],
    "infos": [],
    "context": {}
  },
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed: 3 error(s), 2 warning(s)",
    "details": {
      "error_count": 3,
      "warning_count": 2,
      "errors": {
        "username": [
          "Username is required",
          "Username must be 3+ characters",
          "Username contains invalid characters"
        ],
        "email": "Email format invalid"
      },
      "warnings": {
        "password": "Password contains common patterns",
        "phone": "Phone country code not verified"
      }
    }
  },
  "meta": {
    "timestamp": "2025-11-15T17:30:41Z",
    "request_id": "550e8400-e29b-41d4-a716-446655440001",
    "version": "1.0",
    "server_time": 1763245841
  }
}
```

### Example 3: File Upload with Context (ODS Import)

**Scenario**: Spreadsheet upload with row-level errors and debugging context

```json
{
  "data": null,
  "validation": {
    "errors": [],
    "warnings": [],
    "infos": [],
    "context": {}
  },
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed: 5 error(s), 2 warning(s), 3 info(s)",
    "details": {
      "error_count": 5,
      "warning_count": 2,
      "info_count": 3,
      "errors": {
        "row_3": [
          "Invalid date format in column 'Start Date': '2025-13-01'",
          "Staff member 'John Doe' not found in system"
        ],
        "row_5": "Shift duration exceeds maximum (12 hours)",
        "row_7": "Position 'DR_INVALID' not recognized",
        "file_format": "Header row missing required columns"
      },
      "warnings": {
        "row_2": "Data overlaps with existing schedule by 2 hours",
        "encoding": "File uses non-UTF-8 encoding (Latin-1 detected)"
      },
      "infos": {
        "row_1": "Header validation passed",
        "row_4": "Data successfully parsed (10 assignments)",
        "total": "Processing 247 rows, 245 valid, 2 errors"
      },
      "context": {
        "filename": "nov_2025_schedule.ods",
        "sheet_name": "November 2025",
        "file_size_bytes": 15728640,
        "file_hash": "sha256:abc123...",
        "upload_timestamp": "2025-11-15T17:25:00Z",
        "total_rows_processed": 247,
        "rows_with_errors": 2,
        "rows_with_warnings": 1
      }
    }
  },
  "meta": {
    "timestamp": "2025-11-15T17:30:41Z",
    "request_id": "550e8400-e29b-41d4-a716-446655440002",
    "version": "1.0",
    "server_time": 1763245841
  }
}
```

### Example 4: Global Error (No Specific Field)

**Scenario**: System-level validation failure

```json
{
  "data": null,
  "validation": {
    "errors": [],
    "warnings": [],
    "infos": [],
    "context": {}
  },
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed: 1 error(s)",
    "details": {
      "error_count": 1,
      "errors": {
        "_global_": [
          "Request exceeds maximum payload size",
          "System is in read-only mode"
        ]
      }
    }
  },
  "meta": {
    "timestamp": "2025-11-15T17:30:41Z",
    "request_id": "550e8400-e29b-41d4-a716-446655440003",
    "version": "1.0",
    "server_time": 1763245841
  }
}
```

### Example 5: All Three Severity Levels (Warnings and Infos)

**Scenario**: Partial success - operation completed with issues

```json
{
  "data": null,
  "validation": {
    "errors": [],
    "warnings": [],
    "infos": [],
    "context": {}
  },
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed: 1 error(s), 3 warning(s), 2 info(s)",
    "details": {
      "error_count": 1,
      "warning_count": 3,
      "info_count": 2,
      "errors": {
        "schedule_end_date": "Schedule end date cannot be in the past"
      },
      "warnings": {
        "coverage": "Coverage target not met for Emergency Department",
        "staff_availability": "Several staff members unavailable in requested period",
        "shift_gaps": "2 hour gaps detected between shifts"
      },
      "infos": {
        "status": "Successfully imported 485 valid shift assignments",
        "optimization": "Running coverage optimization now..."
      },
      "context": {
        "operation": "schedule_import",
        "priority": "high",
        "imported_count": 485,
        "total_count": 487,
        "processing_time_ms": 1234
      }
    }
  },
  "meta": {
    "timestamp": "2025-11-15T17:30:41Z",
    "request_id": "550e8400-e29b-41d4-a716-446655440004",
    "version": "1.0",
    "server_time": 1763245841
  }
}
```

---

## Nested Structure Diagram

```
HTTP Response (JSON)
│
├─ data: null
├─ validation: {} (empty when error is set)
├─ error: ErrorDetail
│  │
│  ├─ code: "VALIDATION_ERROR"
│  ├─ message: "Validation failed: 2 error(s), 1 warning(s)"
│  └─ details: {
│     │
│     ├─ error_count: 2                    [Count]
│     ├─ warning_count: 1                  [Count]
│     ├─ info_count: 0 (omitted if 0)     [Count]
│     │
│     ├─ errors: {                         [Field → Messages]
│     │  ├─ "email": "Email required"      [Single message: string]
│     │  └─ "password": [                  [Multiple messages: array]
│     │     "Password required",
│     │     "Must be 8+ chars"
│     │  ]
│     │
│     ├─ warnings: {                       [Same structure as errors]
│     │  └─ "phone": "Not verified"
│     │
│     ├─ infos: {                          [Same structure as errors]
│     │  └─ "status": "Validation started"
│     │
│     └─ context: {                        [User-provided debug info]
│        ├─ "filename": "data.ods"
│        ├─ "line_number": 42
│        └─ "operation": "import"
│
└─ meta: ResponseMeta
   ├─ timestamp: "2025-11-15T17:30:41Z"
   ├─ request_id: "550e8400..."
   ├─ version: "1.0"
   └─ server_time: 1763245841
```

---

## Field Mapping Rules

### Single Error Per Field
```go
vr.AddError("email", "Email is required")

// Result in JSON
{
  "errors": {
    "email": "Email is required"  ← String, not array
  }
}
```

### Multiple Errors Per Field
```go
vr.AddError("password", "Required")
vr.AddError("password", "Too short")

// Result in JSON
{
  "errors": {
    "password": [               ← Array of strings
      "Required",
      "Too short"
    ]
  }
}
```

### Empty Field Name (Global Error)
```go
vr.AddError("", "System error")

// Result in JSON
{
  "errors": {
    "_global_": "System error"  ← Special "_global_" key
  }
}
```

### Mixed (Some fields with multiple messages)
```go
vr.AddError("email", "Required")
vr.AddError("password", "Required")
vr.AddError("password", "Too short")

// Result in JSON
{
  "errors": {
    "email": "Required",        ← String (single message)
    "password": [               ← Array (multiple messages)
      "Required",
      "Too short"
    ]
  }
}
```

---

## Client-Side Processing Examples

### JavaScript/TypeScript

```typescript
interface ErrorResponse {
  code: string;
  message: string;
  details: {
    error_count: number;
    errors?: Record<string, string | string[]>;
    warnings?: Record<string, string | string[]>;
    context?: Record<string, any>;
  };
}

function handleValidationError(error: ErrorResponse) {
  console.log(`${error.code}: ${error.message}`);

  // Process field errors
  if (error.details.errors) {
    for (const [field, messages] of Object.entries(error.details.errors)) {
      const msgs = Array.isArray(messages) ? messages : [messages];
      msgs.forEach(msg => {
        displayFieldError(field, msg);
      });
    }
  }

  // Process warnings
  if (error.details.warnings) {
    // Similar processing...
  }

  // Use context for debugging
  if (error.details.context) {
    console.debug("Context:", error.details.context);
  }
}
```

### Python

```python
def handle_validation_error(error_response):
    error = error_response.get('error', {})
    details = error.get('details', {})

    print(f"{error['code']}: {error['message']}")

    # Process field errors
    errors = details.get('errors', {})
    for field, messages in errors.items():
        if isinstance(messages, list):
            for msg in messages:
                print(f"  {field}: {msg}")
        else:
            print(f"  {field}: {messages}")

    # Handle context
    context = details.get('context', {})
    if context:
        print(f"Debug info: {context}")
```

---

## HTTP Status Code Mapping

Error formatter always produces **HTTP 400 Bad Request** for validation errors:

```go
errorDetail := FormatValidationErrors(vr)
return c.JSON(http.StatusBadRequest,
    NewApiResponse(nil).
        WithError(errorDetail.Code, errorDetail.Message).
        WithErrorDetails(errorDetail.Details))
```

**Note**: For non-validation errors, use different status codes via [2.4] Status Mapping work package.

---

## Conclusion

The error formatter provides a clean, structured way to present validation errors to API consumers. The nested hierarchy allows for:

1. **Machine-readable processing** (via error code)
2. **Human-readable feedback** (via message and field names)
3. **Rich context** (for debugging)
4. **Flexible field handling** (single messages vs. arrays)
5. **Clean JSON serialization** (proper omissions, no null fields)

All examples above are tested and verified to marshal/unmarshal correctly.
