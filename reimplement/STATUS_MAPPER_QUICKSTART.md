# StatusMapper Quick Start Guide

## Installation

The StatusMapper is located at:
```
internal/api/status_mapper.go
```

## Basic Usage

### 1. Create a Mapper Instance
```go
package main

import (
    "net/http"
    "github.com/schedcu/reimplement/internal/api"
    "github.com/schedcu/reimplement/internal/validation"
)

func main() {
    mapper := api.NewStatusMapper()
    // Ready to use
}
```

### 2. Map Validation Results to HTTP Status
```go
// Create a validation result with errors
vr := validation.NewValidationResult()
vr.AddError("email", "Invalid email format")
vr.AddError("password", "Password too short")

// Get HTTP status code
status := mapper.MapValidationToStatus(vr)
// Returns: 400 (Bad Request)

// Use in HTTP response
http.Error(w, "Validation failed", status)
```

### 3. Map Specific Error Codes to HTTP Status
```go
// Get HTTP status for a specific error code
status := mapper.ErrorCodeToHTTPStatus(validation.INVALID_FILE_FORMAT)
// Returns: 400 (Bad Request)

status = mapper.ErrorCodeToHTTPStatus(validation.DATABASE_ERROR)
// Returns: 500 (Internal Server Error)
```

### 4. Get Human-Readable Descriptions
```go
// Severity description
desc := mapper.SeverityToDescription(validation.ERROR)
// Returns: "Error - validation failed, action cannot proceed"

// Error code description
desc = mapper.MessageCodeToDescription(validation.MISSING_REQUIRED_FIELD)
// Returns: "A required field is missing from the request"
```

### 5. Classify HTTP Status Codes
```go
status := 400

if mapper.IsClientError(status) {
    log.Println("Client error occurred")  // true
}

if mapper.IsServerError(status) {
    log.Println("Server error occurred")  // false
}

if mapper.IsSuccess(status) {
    log.Println("Request succeeded")      // false
}
```

## Complete Example: HTTP Handler

```go
func handleUploadFile(w http.ResponseWriter, r *http.Request) {
    mapper := api.NewStatusMapper()

    // Parse and validate file
    file, _, err := r.FormFile("schedule")
    if err != nil {
        status := mapper.ErrorCodeToHTTPStatus(validation.PARSE_ERROR)
        http.Error(w, "Failed to parse file", status)
        return
    }
    defer file.Close()

    // Perform validation
    vr := validation.NewValidationResult()
    vr.AddError("file_format", "Expected ODS format")

    // Get appropriate HTTP status
    status := mapper.MapValidationToStatus(vr)

    // Create API response with formatted errors
    errorDetail := api.FormatValidationErrors(vr)
    response := api.NewApiResponse(nil).
        WithError(errorDetail.Code, errorDetail.Message).
        WithErrorDetails(errorDetail.Details)

    // Write response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(response)
}
```

## HTTP Status Code Reference

| Status | Code | Meaning | When to Use |
|--------|------|---------|------------|
| 200 | OK | Success | Valid request, validation passed or only has warnings/infos |
| 400 | Bad Request | Validation failed | Client data validation errors |
| 500 | Internal Error | Server error | Database or external service failures |

## Validation Error Codes and Mappings

```go
// Client Errors (400)
INVALID_FILE_FORMAT       // File format validation failure
MISSING_REQUIRED_FIELD    // Required field missing
DUPLICATE_ENTRY          // Duplicate data detected
PARSE_ERROR              // Data parsing failure

// Server Errors (500)
DATABASE_ERROR           // Database access failure
EXTERNAL_SERVICE_ERROR   // External API failure
UNKNOWN_ERROR            // Unknown/unhandled error
```

## Precedence Rules

When ValidationResult has multiple message types:

```
Errors     → 400 Bad Request (block action)
    ↓
Warnings   → 200 OK (with warning notification)
    ↓
Infos      → 200 OK (with info notification)
    ↓
Empty      → 200 OK (success)
```

Example:
```go
vr := validation.NewValidationResult()
vr.AddError("field1", "error")
vr.AddWarning("field2", "warning")
vr.AddInfo("source", "info")

status := mapper.MapValidationToStatus(vr)
// Returns: 400 (error takes precedence)
```

## Integration with ApiResponse

```go
response := api.NewApiResponse(userData).
    WithValidation(vr).
    WithError("VALIDATION_ERROR", "Validation failed")

// Check if successful
if !response.IsSuccess() {
    status := mapper.MapValidationToStatus(vr)
    // Send response with appropriate status code
}
```

## Common Patterns

### Pattern 1: Validate and Return Status
```go
status := mapper.MapValidationToStatus(vr)
if !vr.IsValid() {
    w.WriteHeader(status)
    // Send error response
}
```

### Pattern 2: Format and Return Errors
```go
errorDetail := api.FormatValidationErrors(vr)
response := api.NewApiResponse(nil).
    WithError(errorDetail.Code, errorDetail.Message).
    WithErrorDetails(errorDetail.Details)

w.WriteHeader(mapper.MapValidationToStatus(vr))
json.NewEncoder(w).Encode(response)
```

### Pattern 3: Handle Specific Error Codes
```go
switch code {
case validation.INVALID_FILE_FORMAT:
    status := mapper.ErrorCodeToHTTPStatus(code)
    // Handle specific error
case validation.DATABASE_ERROR:
    status := mapper.ErrorCodeToHTTPStatus(code)
    // Handle different error
}
```

## Testing

Run tests:
```bash
go test ./internal/api -v -run StatusMapper
```

Test coverage:
- 47 tests covering all scenarios
- Validation mapping (8 tests)
- Precedence rules (4 tests)
- Error code mapping (8 tests)
- Descriptions (12 tests)
- Classification (3 tests)
- Integration (2 tests)
- Edge cases (3 tests)

## API Method Reference

### MapValidationToStatus
```go
func (sm *StatusMapper) MapValidationToStatus(vr *ValidationResult) int
```
Maps ValidationResult to HTTP status code (400 or 200).

### ErrorCodeToHTTPStatus
```go
func (sm *StatusMapper) ErrorCodeToHTTPStatus(code MessageCode) int
```
Maps specific error code to HTTP status (4xx or 5xx).

### SeverityToDescription
```go
func (sm *StatusMapper) SeverityToDescription(severity Severity) string
```
Returns human-readable severity description.

### MessageCodeToDescription
```go
func (sm *StatusMapper) MessageCodeToDescription(code MessageCode) string
```
Returns human-readable error description.

### IsClientError
```go
func (sm *StatusMapper) IsClientError(statusCode int) bool
```
Returns true if status is 4xx.

### IsServerError
```go
func (sm *StatusMapper) IsServerError(statusCode int) bool
```
Returns true if status is 5xx.

### IsSuccess
```go
func (sm *StatusMapper) IsSuccess(statusCode int) bool
```
Returns true if status is 2xx.

## Notes

- StatusMapper is stateless and thread-safe
- Safe to create single instance and reuse
- All error codes default to 500 if unmapped
- Works with ValidationResult from internal/validation package
- Integrates with ApiResponse for HTTP responses
