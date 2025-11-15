# ApiResponse Design Documentation

## Overview

The `ApiResponse[T any]` is a generic type that provides a unified response format for all API endpoints. It combines data, validation results, error handling, and metadata in a single, type-safe structure.

## Core Concepts

### Type Safety
Uses Go generics to ensure type-safe responses at compile time:
```go
response := NewApiResponse[User](user)
// Type: *ApiResponse[User]
```

### Success Definition
A response is successful when:
1. Error is nil, AND
2. Validation has no errors (warnings are allowed)

### Metadata Auto-Generation
Every response automatically includes:
- Timestamp of response creation
- RequestID (UUID v4 for tracing)
- API Version ("1.0")
- ServerTime (unix timestamp)

## Structure

### ApiResponse[T]
```
Data: T                              // Response payload
Validation: *ValidationResult        // Validation details
Error: *ErrorDetail                  // Error information
Meta: *ResponseMeta                  // Response metadata
```

### ErrorDetail
```
Code: string                         // Error code (e.g., "NOT_FOUND")
Message: string                      // Human-readable message
Details: map[string]interface{}      // Contextual information
```

### ResponseMeta
```
Timestamp: time.Time                 // Response time
RequestID: string                    // UUID for tracing
Version: string                      // API version
ServerTime: int64                    // Unix timestamp
```

## Methods

### NewApiResponse[T](data T) -> *ApiResponse[T]
Creates a success response.

### WithValidation(vr) -> *ApiResponse[T]
Adds validation results.

### WithError(code, message) -> *ApiResponse[T]
Sets error details.

### WithErrorDetails(details) -> *ApiResponse[T]
Adds contextual error information.

### IsSuccess() -> bool
Checks if response represents success.

### MarshalJSON() -> ([]byte, error)
Custom JSON serialization.

## JSON Output Format

### Success Response
```json
{
  "data": {...},
  "validation": {...},
  "meta": {...}
}
```

### Error Response
```json
{
  "data": null,
  "error": {...},
  "validation": {...},
  "meta": {...}
}
```

## Usage Patterns

### Pattern 1: Simple Success
```go
response := NewApiResponse(result)
```

### Pattern 2: With Validation
```go
response := NewApiResponse(data).WithValidation(validationResult)
```

### Pattern 3: Error with Context
```go
response := NewApiResponse(EmptyData{}).
    WithError("NOT_FOUND", "Resource not found").
    WithErrorDetails(map[string]interface{}{
        "resource": "user",
        "id": "123",
    })
```

### Pattern 4: Chained Operations
```go
response := NewApiResponse(item).
    WithValidation(vr).
    WithError("PARTIAL_SUCCESS", "Completed with warnings")
```

## JSON Serialization

### Marshaling
```go
jsonBytes, err := json.Marshal(response)
```

### Unmarshaling
```go
var unmarshaledResp ApiResponse[User]
err := json.Unmarshal(jsonBytes, &unmarshaledResp)
```

### Field Omission
- Error field is omitted when nil
- Validation field is always included
- Details maps are omitted if empty

## Type Flexibility

Works with:
- Primitives: `int`, `string`, `bool`, `float64`
- Collections: `[]T`, `map[K]V`
- Structs: custom user types
- Pointers: `*T`
- Interfaces: `interface{}`
- Nil: `NewApiResponse[interface{}](nil)`

## Integration Points

### With HTTP Handlers
```go
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    response := NewApiResponse(user)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### With Validation
```go
vr := validation.NewValidationResult()
vr.AddError("field", "error message")
response := NewApiResponse(data).WithValidation(vr)
if !response.IsSuccess() {
    // Handle validation errors
}
```

### With Error Handling
```go
user, err := service.GetUser(id)
if err != nil {
    return NewApiResponse(User{}).
        WithError("SERVICE_ERROR", err.Error())
}
return NewApiResponse(user)
```

## Testing

All methods are thoroughly tested:
- 36 comprehensive test cases
- JSON serialization round-trips
- Type flexibility validation
- Performance benchmarks
- Edge case handling

Run tests with:
```bash
go test ./internal/api/ -v
```

## Performance

Benchmarks show efficient serialization:
- Marshal: ~2-3 microseconds
- Unmarshal: ~2-3 microseconds
- Round-trip: ~4-6 microseconds

See `response_test.go` for benchmark details.

## Dependencies

- `ValidationResult` from `internal/validation`
- `google/uuid` for RequestID generation
- Standard library: `encoding/json`, `time`

## Files

- `response.go` - Core implementation
- `response_test.go` - Comprehensive test suite
- `status_mapper.go` - HTTP status mapping utility
- `example_usage.go` - Usage examples
