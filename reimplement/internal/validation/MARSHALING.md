# ValidationResult JSON Marshaling Documentation

## Overview

The `ValidationResult` struct provides comprehensive validation error reporting with support for JSON marshaling and unmarshaling. This document describes the JSON serialization format and usage patterns.

## JSON Structure

### Basic Format

```json
{
  "errors": [
    {
      "field": "email",
      "message": "invalid email format"
    }
  ],
  "warnings": [
    {
      "field": "age",
      "message": "under 18"
    }
  ],
  "infos": [
    {
      "field": "source",
      "message": "from external API"
    }
  ],
  "context": {
    "user_id": 123,
    "request_id": "req-abc123"
  }
}
```

### Fields

- **errors** (array): List of validation errors that prevent the operation
- **warnings** (array): List of validation warnings that don't prevent the operation
- **infos** (array): List of informational messages about the validation process
- **context** (object): Arbitrary key-value pairs for debugging and tracing

### Message Structure

Each message (error, warning, or info) is a simple object with two fields:

```json
{
  "field": "field_name",
  "message": "descriptive message"
}
```

- **field**: Identifier for where the message originated (field name, component, source)
- **message**: Human-readable description of the issue

## Complete Example

```json
{
  "errors": [
    {
      "field": "file_format",
      "message": "expected ODS file but got XLSX"
    },
    {
      "field": "row_5",
      "message": "missing required field: employee_id"
    },
    {
      "field": "row_5",
      "message": "invalid date format in shift_date: expected YYYY-MM-DD"
    }
  ],
  "warnings": [
    {
      "field": "row_3",
      "message": "shift overlaps with previous assignment"
    },
    {
      "field": "row_8",
      "message": "employee works more than 12 hours in a single day"
    }
  ],
  "infos": [
    {
      "field": "parsing",
      "message": "processed 50 rows successfully"
    },
    {
      "field": "validation",
      "message": "validation completed in 245ms"
    }
  ],
  "context": {
    "file_name": "schedule_2024_01.ods",
    "file_size_bytes": 51200,
    "sheet_name": "January",
    "rows_processed": 50,
    "rows_skipped": 2,
    "processing_time_ms": 245,
    "execution_host": "validator-prod-1",
    "metadata": {
      "version": "1.0.0",
      "build_number": 12345,
      "environment": "production"
    },
    "flags": [
      "strict_validation",
      "check_overlaps",
      "check_max_hours"
    ],
    "source_system": "amion_api"
  }
}
```

## Usage Patterns

### Creating ValidationResult

```go
vr := validation.NewValidationResult()
vr.AddError("email", "invalid email format")
vr.AddWarning("age", "under 18")
vr.AddInfo("source", "from API")
vr.SetContext("user_id", 123)
vr.SetContext("request_id", "req-abc123")
```

### Marshaling to JSON

```go
// Compact JSON
data, err := json.Marshal(vr)

// Indented JSON (pretty-printed)
data, err := json.MarshalIndent(vr, "", "  ")
```

### Unmarshaling from JSON

```go
var vr validation.ValidationResult
err := json.Unmarshal(data, &vr)
```

### Round-Trip Serialization

```go
// Create ValidationResult
original := validation.NewValidationResult()
original.AddError("field", "message")

// Marshal to JSON
data, _ := json.Marshal(original)

// Unmarshal back
roundtripped := &validation.ValidationResult{}
json.Unmarshal(data, roundtripped)

// original and roundtripped are semantically equivalent
```

## Data Types in Context

The `context` map can contain various JSON-compatible types:

### Primitives

```go
vr.SetContext("count", 42)           // JSON number
vr.SetContext("name", "test")        // JSON string
vr.SetContext("enabled", true)       // JSON boolean
vr.SetContext("value", 3.14)         // JSON number (float)
vr.SetContext("empty", nil)          // JSON null
```

### Collections

```go
// JSON array
vr.SetContext("tags", []string{"tag1", "tag2", "tag3"})

// JSON object
vr.SetContext("metadata", map[string]interface{}{
  "version": "1.0.0",
  "build": 123,
})

// Nested structure
vr.SetContext("config", map[string]interface{}{
  "outer": map[string]interface{}{
    "inner": "value",
  },
})
```

## Edge Cases

### Empty Messages
Both `field` and `message` can be empty strings, which is valid:

```json
{
  "errors": [
    {
      "field": "",
      "message": "generic error"
    }
  ],
  "warnings": [],
  "infos": [],
  "context": {}
}
```

### Special Characters

Messages can contain:
- Unicode characters: `"message": "unicode: 你好世界 مرحبا العالم"`
- Escaped characters: `"message": "line1\nline2\ttab"`
- Quotes: `"message": "double \"quoted\" text"`

All special characters are properly escaped during JSON marshaling.

### Null Context Values

`context` can contain null values:

```go
vr.SetContext("nullable_field", nil)
```

Which marshals to:

```json
{
  "context": {
    "nullable_field": null
  }
}
```

### Large Messages

There's no limit on:
- Number of errors/warnings/infos: tested with 100+ messages
- Size of context values: complex nested structures supported
- Length of field/message strings: any valid UTF-8

## Serialization Format Details

### JSON Tags

```go
type ValidationResult struct {
  Errors   []SimpleValidationMessage `json:"errors"`
  Warnings []SimpleValidationMessage `json:"warnings"`
  Infos    []SimpleValidationMessage `json:"infos"`
  Context  map[string]interface{}    `json:"context"`
}

type SimpleValidationMessage struct {
  Field   string `json:"field"`
  Message string `json:"message"`
}
```

### Marshaling Behavior

- All fields are included in JSON (no omitempty)
- Empty arrays remain empty: `"errors": []`
- Empty objects remain empty: `"context": {}`
- null values are preserved in context

### Unmarshaling Behavior

Go's JSON unmarshaling is lenient:
- Missing fields use zero values
- Unknown JSON fields are ignored
- Type mismatches cause errors (e.g., string where number expected)

## Round-Trip Guarantees

The JSON serialization format is fully round-trippable:

1. Create ValidationResult
2. Marshal to JSON
3. Unmarshal back
4. The result is semantically identical to the original

Tested scenarios:
- Empty results
- All message types mixed
- Complex nested context structures
- Special characters and unicode
- Large numbers of messages (100+)
- Various context data types
- Duplicate messages (order preserved)

## Testing

Test fixtures are available in `tests/fixtures/`:
- `validation_empty.json` - Empty result
- `validation_simple.json` - Simple result with mixed messages
- `validation_complex.json` - Complex result with nested context

All fixtures pass:
- JSON parsing validation
- Round-trip serialization
- Structure verification

## Performance Considerations

- Marshaling is fast: `json.Marshal()` is Go's standard implementation
- Unmarshaling is efficient: direct into struct fields via reflection
- Context maps support any JSON-serializable type
- No custom serialization logic required

## Compatibility

- Go 1.20+
- Standard library `encoding/json` package
- Compatible with any JSON parser
- No external dependencies

## Examples

See `tests/fixtures/` directory for complete JSON examples:

1. **validation_empty.json** - Minimal valid result
2. **validation_simple.json** - Common validation scenario
3. **validation_complex.json** - Complex multi-error scenario with nested data

Each fixture demonstrates:
- Proper JSON structure
- Round-trip serialization
- Context usage
- Error/warning/info grouping
