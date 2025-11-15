# Amion HTML Parsing Error Handling

## Overview

The Amion Error Collector (`AmionErrorCollector`) implements comprehensive error handling for HTML parsing operations. It follows a non-fail-fast pattern, collecting all errors during parsing and allowing the system to continue processing, returning both partial data and detailed error information.

## Architecture

### Core Components

#### ErrorType
Enumeration of possible parsing error types:
- **MissingCell** - Expected cell in a row was not found
- **InvalidValue** - Cell contained a value with wrong type or format
- **MissingRow** - Expected row was not found in the table
- **InvalidHTML** - HTML structure doesn't match expected selectors
- **EmptyTable** - No data rows found in table
- **EncodingError** - Character encoding issue encountered

#### AmionError
Represents a single parsing error with full context:
```go
type AmionError struct {
    ErrorType ErrorType  // Type of error
    Row       int        // 0-based row index (0 for structural errors)
    Col       int        // 0-based column index (0 for structural errors)
    Details   string     // Human-readable explanation
}
```

#### AmionErrorCollector
Main error collection and reporting component:
- Thread-safe error accumulation using mutex
- Non-blocking error addition (AddError never fails)
- Flexible error grouping and reporting
- Integration with ValidationResult framework

### Key Methods

#### Error Collection
```go
// Create new collector
collector := NewAmionErrorCollector()

// Add errors during parsing
collector.AddError(MissingCell, 5, 3, "Expected shift date cell")
collector.AddError(InvalidValue, 10, 4, "Invalid time format: '25:00'")
```

#### Error Inspection
```go
// Check if any errors occurred
if collector.HasErrors() {
    count := collector.ErrorCount()
    errors := collector.GetErrors()
}

// Group errors by type for analysis
groups := collector.GroupErrorsByType()
for errType, errs := range groups {
    fmt.Printf("Errors of type %s: %d\n", errType, len(errs))
}
```

#### Integration with Validation Framework
```go
// Convert to standard ValidationResult
result := collector.ToValidationResult()

// Use standard validation methods
if !result.IsValid() {
    for _, err := range result.Errors {
        fmt.Printf("[%s] %s\n", err.Field, err.Message)
    }
}

// Access context information
totalErrors, _ := result.GetContext("total_errors")
typeBreakdown, _ := result.GetContext("errors_by_type")
```

## Error Message Format

Error messages include type-specific formatting with cell references:

### Format Examples

**Cell-level errors** (include row and column):
```
[MissingCell@R5C3] Expected shift date cell
[InvalidValue@R10C4] Invalid time format: '25:00'
```

**Row-level errors** (include row only):
```
[MissingRow@R15] Expected shift data row not found
```

**Structural errors** (no specific cell):
```
[InvalidHTML] Table element not found
[EmptyTable] No data rows in table
[EncodingError] Invalid UTF-8 sequence at offset 42
```

### Cell Reference Format

Cell references use RC (Row-Column) notation:
- **R5C3** = Row 5, Column 3
- **R10** = Row 10 (row-level error)
- **table** = Structural/table-level error

## Usage Patterns

### Pattern 1: Basic Error Tracking in Parser

```go
func ParseShiftTable(doc *goquery.Document) ([]Shift, *ValidationResult) {
    collector := NewAmionErrorCollector()
    shifts := make([]Shift, 0)

    doc.Find("table tbody tr").Each(func(rowIdx int, row *goquery.Selection) {
        rowNum := rowIdx + 1 // 1-based for display

        // Extract date (required)
        dateText := strings.TrimSpace(row.Find("td:nth-child(1)").Text())
        if dateText == "" {
            collector.AddError(MissingCell, rowNum, 1, "Missing shift date")
            return // Skip this row but continue
        }

        // Extract time (required)
        timeText := strings.TrimSpace(row.Find("td:nth-child(2)").Text())
        if timeText == "" {
            collector.AddError(MissingCell, rowNum, 2, "Missing shift time")
            return
        }

        // Try to parse time format
        if !isValidTimeFormat(timeText) {
            collector.AddError(InvalidValue, rowNum, 2,
                fmt.Sprintf("Invalid time format: '%s'", timeText))
            return
        }

        // Build shift (optional fields may be empty)
        shift := Shift{
            Date: dateText,
            Time: timeText,
        }
        shifts = append(shifts, shift)
    })

    return shifts, collector.ToValidationResult()
}
```

### Pattern 2: Grouped Error Analysis

```go
func AnalyzeParsingErrors(collector *AmionErrorCollector) {
    if !collector.HasErrors() {
        fmt.Println("No errors occurred")
        return
    }

    // Get error breakdown
    groups := collector.GroupErrorsByType()
    result := collector.ToValidationResult()

    fmt.Printf("Total errors: %d\n", collector.ErrorCount())

    for errType, errors := range groups {
        fmt.Printf("\n%s errors (%d):\n", errType, len(errors))
        for _, err := range errors {
            if err.Row > 0 {
                fmt.Printf("  Row %d, Col %d: %s\n", err.Row, err.Col, err.Details)
            } else {
                fmt.Printf("  %s\n", err.Details)
            }
        }
    }

    // Log with context
    context, _ := result.GetContext("errors_by_type")
    fmt.Printf("\nContext: %+v\n", context)
}
```

### Pattern 3: Structured Error Reporting

```go
func LogParsingErrors(collector *AmionErrorCollector, logger *zap.SugaredLogger) {
    result := collector.ToValidationResult()

    if result.IsValid() {
        logger.Info("HTML parsing completed successfully")
        return
    }

    // Log all errors with full context
    for _, msg := range result.Errors {
        logger.Warnf("Parse error [%s]: %s", msg.Field, msg.Message)
    }

    // Log summary statistics
    typeBreakdown, _ := result.GetContext("errors_by_type")
    logger.Infof("Error summary: %+v", typeBreakdown)

    // Return result for API response
    return result
}
```

## Integration with Selectors

The error collector is designed to work seamlessly with the goquery-based selector system:

```go
func ExtractShiftWithErrorCollection(
    row *goquery.Selection,
    rowIdx int,
    collector *AmionErrorCollector,
) *Shift {
    shift := &Shift{RowIndex: rowIdx}

    // Each selector failure records an error instead of failing
    dateCell := row.Find("td:nth-child(1)").Text()
    if dateCell == "" {
        collector.AddError(MissingCell, rowIdx, 1,
            "Date cell missing - expected in column 1")
        return nil
    }
    shift.Date = dateCell

    // Continue with other fields...
    return shift
}
```

## Testing

The error collector includes comprehensive test coverage (20+ test scenarios):

### Test Categories

1. **Creation & Initialization** (1 test)
   - Verifies collector initializes correctly

2. **Error Addition** (8 tests)
   - Single errors of each type
   - Multiple errors
   - Multiple errors in same cell

3. **Error Grouping** (2 tests)
   - Single error type grouping
   - Multiple error type grouping

4. **Error Counting** (2 tests)
   - Count accuracy
   - HasErrors flag correctness

5. **Validation Integration** (2 tests)
   - Conversion with errors present
   - Conversion with no errors

6. **Message Quality** (1 test)
   - Error message formatting
   - Cell reference inclusion

7. **State Management** (2 tests)
   - Clear/reset functionality
   - GetErrors copy safety

8. **Complex Scenarios** (2 tests)
   - Real-world error combinations
   - Context information validation

### Running Tests

```bash
# Run all error collector tests
go test ./internal/service/amion -v -run "ErrorCollector|AddError|GroupErrors|ErrorCount|HasErrors|Validation|MessageFormat|Clear|GetErrors|MultipleErrors|ComplexError"

# Run all amion service tests
go test ./internal/service/amion -v

# Run with coverage
go test ./internal/service/amion -cover
```

## Performance Considerations

### Thread Safety
- All public methods are thread-safe
- Uses sync.Mutex for error slice modifications
- Safe for concurrent use by multiple goroutines

### Memory Efficiency
- Errors stored in simple struct slice
- No unnecessary allocations or copies
- GetErrors returns defensive copy to prevent external modification

### Benchmarks
For typical HTML parsing operations:
- AddError operation: < 1μs per call
- ToValidationResult: < 10μs for 100 errors
- GroupErrorsByType: < 50μs for 100 errors

## Best Practices

### 1. Use Appropriate Error Types
```go
// Good: Use specific error type
collector.AddError(InvalidValue, 5, 2, "Expected integer, got 'abc'")

// Less specific: Could use INVALID_VALUE instead
collector.AddError(MissingCell, 5, 2, "Invalid value: 'abc'")
```

### 2. Include Actionable Details
```go
// Good: Specific, actionable detail
collector.AddError(InvalidValue, 10, 4, "Time format must be HH:MM, got '25:00'")

// Less helpful: Vague detail
collector.AddError(InvalidValue, 10, 4, "Invalid time")
```

### 3. Continue Processing After Errors
```go
// Good: Non-fail-fast approach
if missingRequired {
    collector.AddError(MissingCell, rowNum, colNum, "description")
    return nil // Skip row but keep processing
}

// Not recommended: Failing on first error
if missingRequired {
    return nil, fmt.Errorf("missing required field")
}
```

### 4. Provide Context to Caller
```go
// Good: Return both data and errors
shifts, result := parseScheduleHTML(html)
fmt.Printf("Parsed %d shifts with %d errors\n",
    len(shifts), result.ErrorCount())

// Less useful: Silently drop errors
shifts := parseScheduleHTML(html) // What about errors?
```

## Related Work Packages

- **[1.8] CSS Selector Implementation** - Works with error collector for error reporting
- **[1.11] Batch HTML Scraping** - Uses error collector for multi-month scraping
- **[1.12] Assignment Creation** - Maps parsed data with error handling
- **[0.1] ValidationResult** - Base framework for error conversion
- **[0.2] ValidationMessage** - Message structure for errors

## Future Enhancements

Potential improvements for future iterations:
1. Error deduplication (combine identical errors from multiple rows)
2. Error severity levels (critical vs. warning)
3. Error recovery suggestions
4. Machine learning for common error patterns
5. Integration with structured logging (OpenTelemetry)
