# Quick Start: AmionErrorCollector

## Installation

The error collector is located in `internal/service/amion` and is ready to use.

## Basic Usage

```go
import "github.com/schedcu/reimplement/internal/service/amion"

// Create a new collector
collector := amion.NewAmionErrorCollector()

// Add errors as you encounter them
collector.AddError(amion.MissingCell, 5, 3, "Expected shift date")
collector.AddError(amion.InvalidValue, 5, 4, "Invalid time format")

// Check for errors
if collector.HasErrors() {
    fmt.Printf("Found %d errors\n", collector.ErrorCount())
}

// Convert to ValidationResult for API response
result := collector.ToValidationResult()
if !result.IsValid() {
    // Handle errors
}
```

## Available Error Types

```go
amion.MissingCell     // Expected cell not found
amion.InvalidValue    // Cell value wrong type/format
amion.MissingRow      // Expected row not found
amion.InvalidHTML     // HTML structure mismatch
amion.EmptyTable      // No data rows found
amion.EncodingError   // Character encoding issue
```

## Key Methods

| Method | Purpose |
|--------|---------|
| `NewAmionErrorCollector()` | Create a new collector |
| `AddError(type, row, col, details)` | Collect an error |
| `ErrorCount()` | Get total error count |
| `HasErrors()` | Check if any errors exist |
| `GetErrors()` | Get all errors (safe copy) |
| `Clear()` | Clear all errors |
| `GroupErrorsByType()` | Group errors by type |
| `ToValidationResult()` | Convert to ValidationResult |

## Integration with HTML Parsing

```go
func parseShiftTable(doc *goquery.Document) ([]Shift, *validation.ValidationResult) {
    collector := amion.NewAmionErrorCollector()
    shifts := make([]Shift, 0)

    doc.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
        rowNum := i + 1

        // Get date
        date := strings.TrimSpace(row.Find("td").Eq(0).Text())
        if date == "" {
            collector.AddError(amion.MissingCell, rowNum, 1, "Missing date")
            return // Skip row, continue collecting
        }

        // Get time
        time := strings.TrimSpace(row.Find("td").Eq(1).Text())
        if time == "" {
            collector.AddError(amion.MissingCell, rowNum, 2, "Missing time")
            return
        }

        shifts = append(shifts, Shift{Date: date, Time: time})
    })

    return shifts, collector.ToValidationResult()
}
```

## Error Message Format

Error messages follow this pattern:

```
[ErrorType@CellRef] Details

Examples:
[MissingCell@R5C3] Expected shift date
[InvalidValue@R10C4] Invalid time format: '25:00'
[InvalidHTML] Table element not found
```

Cell references use RC notation:
- `R5C3` = Row 5, Column 3
- `R10` = Row 10 (row-level error)
- `table` = Structural error

## Getting Results

```go
// Convert to ValidationResult
result := collector.ToValidationResult()

// Check if valid
if !result.IsValid() {
    // Log errors
    for _, err := range result.Errors {
        log.Printf("[%s] %s", err.Field, err.Message)
    }
}

// Get context info
totalErrors, _ := result.GetContext("total_errors")
typeBreakdown, _ := result.GetContext("errors_by_type")
```

## Thread Safety

The collector is thread-safe for concurrent use:

```go
go func() {
    collector.AddError(amion.InvalidValue, 5, 4, "Bad value")
}()

go func() {
    collector.AddError(amion.MissingCell, 6, 3, "Missing cell")
}()

// Safe to call from multiple goroutines
count := collector.ErrorCount()
```

## Testing

Run error collector tests:

```bash
# All error collector tests
go test -v ./internal/service/amion -run "ErrorCollector|AddError|GroupErrors|Validation"

# Full amion package tests
go test -v ./internal/service/amion
```

## Examples

See `error_collector_examples.go` for complete working examples:
- Basic usage
- HTML parsing integration
- Error grouping
- ValidationResult conversion
- Clear and reset
- Error retrieval

## Next Steps

1. Import the package
2. Create a collector
3. Add errors as you parse
4. Convert to ValidationResult
5. Return results to caller

For detailed documentation, see `ERROR_HANDLING.md`.
