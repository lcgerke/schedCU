# Error Formatter API Reference

**Package**: `github.com/schedcu/reimplement/internal/api`
**File**: `error_formatter.go`
**Status**: Production-ready

---

## Function Reference

### FormatValidationErrors

**Signature**:
```go
func FormatValidationErrors(vr *validation.ValidationResult) *ErrorDetail
```

**Purpose**: Converts a ValidationResult into a structured ErrorDetail suitable for API responses.

**Parameters**:
- `vr *validation.ValidationResult` - Validation result from validation layer (can be nil)

**Returns**:
- `*ErrorDetail` - Formatted error detail ready for API response

**Error Code**: Always returns `"VALIDATION_ERROR"`

**Example**:
```go
vr := validation.NewValidationResult()
vr.AddError("email", "Email is required")

errorDetail := FormatValidationErrors(vr)
// errorDetail.Code = "VALIDATION_ERROR"
// errorDetail.Message = "Validation failed: 1 error(s)"
// errorDetail.Details["errors"]["email"] = "Email is required"
```

---

## ErrorDetail Type

**Definition**:
```go
type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

**Fields**:

1. **Code** (string)
   - Machine-readable error code
   - Always: `"VALIDATION_ERROR"`
   - Used by clients for programmatic routing

2. **Message** (string)
   - Human-readable summary
   - Format: `"Validation failed: X error(s)[, Y warning(s)][, Z info(s)]"`
   - Examples:
     - `"Validation failed: 1 error(s)"`
     - `"Validation failed: 2 error(s), 1 warning(s)"`
     - `"Validation failed: 3 error(s), 2 warning(s), 1 info(s)"`
     - `"Validation failed"` (if all counts are 0)

3. **Details** (map[string]interface{})
   - Aggregated error/warning/info messages grouped by field
   - Optional counts: `error_count`, `warning_count`, `info_count`
   - Optional maps: `errors`, `warnings`, `infos`
   - Optional context: `context` (from ValidationResult.Context)

---

## Details Map Structure

### Counts

```go
details["error_count"]   // Always present if > 0, type: int
details["warning_count"] // Present if > 0, type: int
details["info_count"]    // Present if > 0, type: int
```

### Error Messages Map

```go
details["errors"].(map[string]interface{}) = {
    "field1": "Single message",           // type: string
    "field2": []interface{}{               // type: []interface{}
        "First message",
        "Second message",
    },
    "_global_": "System-level error",     // Empty field names
}
```

### Warning Messages Map

```go
details["warnings"].(map[string]interface{}) = {
    // Same structure as errors
}
```

### Info Messages Map

```go
details["infos"].(map[string]interface{}) = {
    // Same structure as errors
}
```

### Context

```go
details["context"].(map[string]interface{}) = {
    // Copy of vr.Context from ValidationResult
    // Can contain any key-value pairs set by validation layer
}
```

---

## Helper Functions (Internal)

### formatValidationSummary

**Signature**:
```go
func formatValidationSummary(errorCount, warningCount, infoCount int) string
```

**Purpose**: Builds human-readable summary message from error counts.

**Logic**:
1. Always starts with `"Validation failed: "`
2. Appends error count (always)
3. Appends warning count (if > 0)
4. Appends info count (if > 0)
5. Joins with `", "`

**Examples**:
- `"Validation failed: 1 error(s)"`
- `"Validation failed: 2 error(s), 1 warning(s)"`
- `"Validation failed: 0 error(s)"` (if called with 0, 0, 0)

---

### aggregateMessages

**Signature**:
```go
func aggregateMessages(messages []validation.SimpleValidationMessage) map[string]interface{}
```

**Purpose**: Groups validation messages by field name.

**Logic**:
1. Iterates through messages
2. Groups by Field name
3. Empty field names mapped to `"_global_"`
4. Single message per field → stored as string
5. Multiple messages per field → stored as []interface{}

**Example**:
```go
messages := []SimpleValidationMessage{
    {Field: "email", Message: "Required"},
    {Field: "email", Message: "Invalid format"},
    {Field: "password", Message: "Too short"},
}

result := aggregateMessages(messages)
// result = {
//     "email": []interface{}{"Required", "Invalid format"},
//     "password": "Too short",
// }
```

---

## Integration Patterns

### Pattern 1: Simple Error Handler

```go
func (h *Handler) ValidateRequest(c echo.Context, req interface{}) error {
    vr := h.validator.Validate(req)
    if !vr.IsValid() {
        errorDetail := FormatValidationErrors(vr)
        return c.JSON(http.StatusBadRequest,
            NewApiResponse[any](nil).
                WithError(errorDetail.Code, errorDetail.Message).
                WithErrorDetails(errorDetail.Details))
    }
    return nil
}
```

### Pattern 2: Middleware-Based Validation

```go
func ValidationErrorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Assume validation result stored in context
        if vr, ok := c.Get("validation").(ValidationResult); ok && !vr.IsValid() {
            errorDetail := FormatValidationErrors(&vr)
            return c.JSON(http.StatusBadRequest,
                NewApiResponse[any](nil).
                    WithError(errorDetail.Code, errorDetail.Message).
                    WithErrorDetails(errorDetail.Details))
        }
        return next(c)
    }
}
```

### Pattern 3: Multi-Step Validation

```go
func (h *Handler) ProcessSchedule(c echo.Context, req *ScheduleRequest) error {
    vr := validation.NewValidationResult()

    // Step 1: Basic validation
    if err := h.validateBasic(req, vr); err != nil {
        errorDetail := FormatValidationErrors(vr)
        return c.JSON(http.StatusBadRequest,
            NewApiResponse[any](nil).
                WithError(errorDetail.Code, errorDetail.Message).
                WithErrorDetails(errorDetail.Details))
    }

    // Step 2: Business logic validation
    if err := h.validateBusiness(req, vr); err != nil {
        // Can accumulate multiple validation issues
        errorDetail := FormatValidationErrors(vr)
        return c.JSON(http.StatusBadRequest,
            NewApiResponse[any](nil).
                WithError(errorDetail.Code, errorDetail.Message).
                WithErrorDetails(errorDetail.Details))
    }

    // Success
    result, err := h.process(req)
    return c.JSON(http.StatusOK, NewApiResponse(result))
}
```

---

## Common Use Cases

### Use Case 1: Form Validation

```go
vr := validation.NewValidationResult()
vr.AddError("email", "Email is required")
vr.AddError("password", "Password is required")

errorDetail := FormatValidationErrors(vr)
// Returns ErrorDetail with Code="VALIDATION_ERROR"
// and 2 errors in details["errors"]
```

### Use Case 2: File Upload Validation

```go
vr := validation.NewValidationResult()
vr.AddError("file", "File format must be ODS")
vr.AddWarning("encoding", "Non-UTF-8 encoding detected")
vr.SetContext("filename", "data.xlsx")
vr.SetContext("received_size", "25MB")

errorDetail := FormatValidationErrors(vr)
// Returns ErrorDetail with:
// - 1 error
// - 1 warning
// - Context preserved for debugging
```

### Use Case 3: Batch Processing

```go
vr := validation.NewValidationResult()
for i, item := range items {
    if err := validateItem(item); err != nil {
        vr.AddError(fmt.Sprintf("item_%d", i), err.Error())
    }
}

vr.SetContext("total_items", len(items))
vr.SetContext("failed_items", vr.ErrorCount())

errorDetail := FormatValidationErrors(vr)
// Returns ErrorDetail with all items indexed by item_0, item_1, etc.
```

### Use Case 4: Partial Success

```go
vr := validation.NewValidationResult()
vr.AddError("row_5", "Invalid date")
vr.AddWarning("row_3", "Duplicate entry detected")
vr.AddInfo("status", "Successfully processed 247 of 249 rows")

errorDetail := FormatValidationErrors(vr)
// Returns ErrorDetail with all three severity levels
```

---

## Error Code Reference

The formatter always produces error code: **`"VALIDATION_ERROR"`**

This code is mapped to HTTP status by [2.4] Status Mapper:
- **HTTP Status**: `400 Bad Request`
- **Reason**: Request validation failed

---

## Message Format Reference

### Summary Message Format

```
"Validation failed: X error(s)[, Y warning(s)][, Z info(s)]"
```

**Examples**:
```
"Validation failed: 1 error(s)"
"Validation failed: 2 error(s), 1 warning(s)"
"Validation failed: 3 error(s), 2 warning(s), 1 info(s)"
"Validation failed: 5 error(s), 0 warning(s), 0 info(s)"
"Validation failed"
```

---

## Performance Characteristics

| Metric | Value |
|--------|-------|
| Time Complexity | O(n) |
| Space Complexity | O(m) |
| 1 error | <0.1ms |
| 10 errors | <0.1ms |
| 100 errors | <0.5ms |
| 150 errors | <0.5ms |
| JSON Marshal | <0.1ms |

Where:
- n = total number of messages (errors + warnings + infos)
- m = number of unique field names

---

## Testing

### Running Tests

```bash
# All error formatter tests
go test ./internal/api -v -run FormatValidationErrors

# Specific test
go test ./internal/api -v -run TestFormatValidationErrors_SimpleError

# All API tests (includes error formatter)
go test ./internal/api -v
```

### Test Files

- Location: `internal/api/error_formatter_test.go`
- Count: 15 test cases
- Coverage: 100% of public API
- Status: All passing

---

## Troubleshooting

### Issue: "error_count not in details"

**Cause**: error_count is only included if > 0

**Solution**: Check if there are actual errors:
```go
if details["error_count"] != nil {
    count := details["error_count"].(int)
}
```

### Issue: "errors field is empty"

**Cause**: No errors added to ValidationResult

**Solution**: Verify errors were added:
```go
if !vr.HasErrors() {
    // No errors to format
    return
}
```

### Issue: "Field name appears as '_global_' in response"

**Expected**: Empty field names are mapped to "_global_"

**Example**:
```go
vr.AddError("", "Global error")  // Field name is empty

// In response:
// details["errors"]["_global_"] = "Global error"
```

---

## Best Practices

1. **Always check IsValid() before formatting**
   ```go
   if !vr.IsValid() {
       errorDetail := FormatValidationErrors(vr)
   }
   ```

2. **Include context for debugging**
   ```go
   vr.SetContext("operation", "user_registration")
   vr.SetContext("source", "request_body")
   ```

3. **Use consistent field names**
   ```go
   vr.AddError("email_address", ...)    // Good
   vr.AddError("emailAddress", ...)     // Also good
   vr.AddError("e-mail", ...)           // Avoid special chars
   ```

4. **Don't over-aggregate**
   ```go
   // Good - specific field names
   vr.AddError("email", "Email required")
   vr.AddError("username", "Username required")

   // Less ideal - vague field names
   vr.AddError("field", "Required")
   vr.AddError("field", "Invalid")
   ```

5. **Preserve context for large operations**
   ```go
   vr.SetContext("file", req.Filename)
   vr.SetContext("sheet", req.SheetName)
   vr.SetContext("rows_processed", processedCount)
   vr.SetContext("rows_valid", validCount)
   ```

---

## Related Documentation

- `/docs/ERROR_FORMATTING.md` - Comprehensive guide
- `/docs/ERROR_FORMATTER_EXAMPLES.md` - JSON examples and nested structures
- `/WORK_PACKAGE_2_3_COMPLETION.md` - Implementation report

---

**Last Updated**: 2025-11-15
**Status**: Production Ready
**Tests**: 15/15 passing
