# ODS Fixture Usage Examples

This document provides practical examples of how to use the ODS test fixtures in your tests.

## Basic File Loading

```go
package yourpackage

import (
    "os"
    "testing"
)

func TestLoadFixture(t *testing.T) {
    // Load fixture file
    fixtureDir := "tests/fixtures/ods"
    data, err := os.ReadFile(os.path.Join(fixtureDir, "valid_schedule.ods"))
    if err != nil {
        t.Fatalf("Failed to load fixture: %v", err)
    }

    // Verify it's not empty
    if len(data) == 0 {
        t.Fatal("Fixture is empty")
    }
}
```

## Testing Valid Data Parsing

```go
func TestParseValidSchedule(t *testing.T) {
    data, _ := os.ReadFile("tests/fixtures/ods/valid_schedule.ods")

    parser := NewODSParser()
    sheets, err := parser.Parse(data)

    // Should not error
    if err != nil {
        t.Fatalf("Unexpected error parsing valid fixture: %v", err)
    }

    // Should have sheets
    if len(sheets) == 0 {
        t.Fatal("No sheets extracted from valid fixture")
    }

    // Verify shift count
    sheet := sheets[0]
    if len(sheet.Rows) < 150 {
        t.Errorf("Expected ~150 rows, got %d", len(sheet.Rows))
    }
}
```

## Testing Partial Data (Missing Columns)

```go
func TestParsePartialSchedule(t *testing.T) {
    data, _ := os.ReadFile("tests/fixtures/ods/partial_schedule.ods")

    parser := NewODSParser()
    sheets, err := parser.Parse(data)

    // Should handle missing optional columns gracefully
    if err != nil {
        t.Fatalf("Parser should handle missing optional columns: %v", err)
    }

    sheet := sheets[0]

    // Verify it has fewer columns than valid fixture
    if len(sheet.Rows[0].Cells) >= 5 {
        t.Error("partial_schedule should have fewer columns than valid fixture")
    }

    // Verify it still has required columns
    headers := []string{"Date", "ShiftType", "RequiredStaffing", "SpecialtyConstraint"}
    for _, header := range headers {
        found := false
        for _, cell := range sheet.Rows[0].Cells {
            if cell.Value == header {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("Missing required column: %s", header)
        }
    }
}
```

## Testing Error Collection Without Fail-Fast

```go
func TestParseInvalidScheduleCollectsErrors(t *testing.T) {
    data, _ := os.ReadFile("tests/fixtures/ods/invalid_schedule.ods")

    parser := NewODSParser()
    sheets, errs := parser.ParseWithErrorCollection(data)

    // Should return sheets despite errors
    if sheets == nil {
        t.Fatal("Should return sheets even with errors")
    }

    if len(sheets) == 0 {
        t.Fatal("Should extract some valid data")
    }

    // Should collect exactly 4 errors (as documented)
    if len(errs) != 4 {
        t.Errorf("Expected 4 errors, got %d", len(errs))
        for i, err := range errs {
            t.Logf("  Error %d: %v", i+1, err)
        }
    }
}
```

## Testing Specific Error Types

```go
func TestInvalidScheduleErrorDetails(t *testing.T) {
    data, _ := os.ReadFile("tests/fixtures/ods/invalid_schedule.ods")

    parser := NewODSParser()
    sheets, errs := parser.ParseWithErrorCollection(data)

    if len(errs) == 0 {
        t.Fatal("Expected errors in invalid fixture")
    }

    // Check error messages contain useful info
    for _, err := range errs {
        errMsg := err.Error()

        // Should indicate what failed
        if errMsg == "" {
            t.Error("Error message is empty")
        }

        // Examples of expected error patterns:
        // - "missing RequiredStaffing at row 5"
        // - "invalid ShiftType 'INVALID_TYPE' at row 8"
        // - "non-numeric RequiredStaffing 'twenty' at row 12"
    }
}
```

## Testing Performance with Large Fixture

```go
func BenchmarkParseLargeSchedule(b *testing.B) {
    data, _ := os.ReadFile("tests/fixtures/ods/large_schedule.ods")
    parser := NewODSParser()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        parser.Parse(data)
    }
}

func TestLargeFixturePerformance(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }

    data, _ := os.ReadFile("tests/fixtures/ods/large_schedule.ods")
    parser := NewODSParser()

    start := time.Now()
    sheets, err := parser.Parse(data)
    duration := time.Since(start)

    if err != nil {
        t.Fatalf("Failed to parse large fixture: %v", err)
    }

    // Should parse 1200+ shifts in reasonable time
    totalShifts := 0
    for _, sheet := range sheets {
        totalShifts += len(sheet.Rows)
    }

    if totalShifts < 1200 {
        t.Errorf("Expected 1200+ rows, got %d", totalShifts)
    }

    // Performance assertion: should complete in < 500ms
    if duration > 500*time.Millisecond {
        t.Logf("Warning: Parsing took %v (expected < 500ms)", duration)
    } else {
        t.Logf("Performance OK: Parsed %d rows in %v", totalShifts, duration)
    }
}
```

## Testing Column Data Extraction

```go
func TestExtractScheduleData(t *testing.T) {
    data, _ := os.ReadFile("tests/fixtures/ods/valid_schedule.ods")

    parser := NewODSParser()
    sheets, _ := parser.Parse(data)
    sheet := sheets[0]

    // Extract header row
    headers := make(map[string]int)
    headerRow := sheet.Rows[0]
    for colIdx, cell := range headerRow.Cells {
        headers[cell.Value] = colIdx
    }

    // Verify expected columns
    expectedColumns := []string{"Date", "ShiftType", "RequiredStaffing", "SpecialtyConstraint", "StudyType"}
    for _, col := range expectedColumns {
        if _, exists := headers[col]; !exists {
            t.Errorf("Missing column: %s", col)
        }
    }

    // Process data rows
    dateIdx := headers["Date"]
    shiftTypeIdx := headers["ShiftType"]
    staffingIdx := headers["RequiredStaffing"]

    for rowIdx := 1; rowIdx < len(sheet.Rows); rowIdx++ {
        row := sheet.Rows[rowIdx]

        date := row.Cells[dateIdx].Value
        shiftType := row.Cells[shiftTypeIdx].Value
        staffing := row.Cells[staffingIdx].Value

        // Validate extracted data
        if date == "" {
            t.Errorf("Empty date at row %d", rowIdx)
        }

        if !contains([]string{"Morning", "Afternoon", "Night"}, shiftType) {
            t.Errorf("Invalid shift type '%s' at row %d", shiftType, rowIdx)
        }

        if staffing == "" {
            t.Errorf("Missing staffing at row %d", rowIdx)
        }
    }
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

## Testing Data Type Handling

```go
func TestDataTypePreservation(t *testing.T) {
    data, _ := os.ReadFile("tests/fixtures/ods/valid_schedule.ods")

    parser := NewODSParser()
    sheets, _ := parser.Parse(data)
    sheet := sheets[0]

    // Check first data row
    firstDataRow := sheet.Rows[1]

    for colIdx, cell := range firstDataRow.Cells {
        // Each cell should have type information
        if cell.Type == "" {
            t.Errorf("Cell at column %d missing type information", colIdx)
        }

        // Numeric cells should have numeric type
        if cell.Type == "float" {
            // Validate numeric format
            _, err := strconv.ParseFloat(cell.Value, 64)
            if err != nil {
                t.Errorf("Cell marked as float has non-numeric value: %q", cell.Value)
            }
        }
    }
}
```

## Table-Driven Tests Using Fixtures

```go
func TestFixtureParsing(t *testing.T) {
    tests := []struct {
        name           string
        fixture        string
        shouldError    bool
        minRowCount    int
        maxRowCount    int
        expectedErrors int
    }{
        {
            name:           "valid schedule",
            fixture:        "valid_schedule.ods",
            shouldError:    false,
            minRowCount:    150,
            maxRowCount:    200,
            expectedErrors: 0,
        },
        {
            name:           "partial schedule",
            fixture:        "partial_schedule.ods",
            shouldError:    false,
            minRowCount:    50,
            maxRowCount:    60,
            expectedErrors: 0,
        },
        {
            name:           "invalid schedule",
            fixture:        "invalid_schedule.ods",
            shouldError:    false, // Should not error, but collect errors
            minRowCount:    20,
            maxRowCount:    40,
            expectedErrors: 4,
        },
        {
            name:           "large schedule",
            fixture:        "large_schedule.ods",
            shouldError:    false,
            minRowCount:    1200,
            maxRowCount:    1400,
            expectedErrors: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            data, err := os.ReadFile(fmt.Sprintf("tests/fixtures/ods/%s", tt.fixture))
            if err != nil {
                t.Fatalf("Failed to load fixture: %v", err)
            }

            parser := NewODSParser()
            sheets, errs := parser.ParseWithErrorCollection(data)

            if tt.shouldError && len(errs) == 0 {
                t.Error("Expected errors but got none")
            }

            if !tt.shouldError && len(errs) > tt.expectedErrors {
                t.Errorf("Unexpected errors: %v", errs)
            }

            if len(sheets) == 0 {
                t.Fatal("No sheets extracted")
            }

            rowCount := len(sheets[0].Rows)
            if rowCount < tt.minRowCount || rowCount > tt.maxRowCount {
                t.Errorf("Row count %d outside range [%d, %d]",
                    rowCount, tt.minRowCount, tt.maxRowCount)
            }

            if len(errs) != tt.expectedErrors {
                t.Errorf("Expected %d errors, got %d", tt.expectedErrors, len(errs))
            }
        })
    }
}
```

## Integration Test with Business Logic

```go
func TestScheduleProcessing(t *testing.T) {
    // Load fixture
    data, _ := os.ReadFile("tests/fixtures/ods/valid_schedule.ods")

    // Parse
    parser := NewODSParser()
    sheets, _ := parser.Parse(data)

    // Process with business logic
    processor := NewScheduleProcessor()
    schedule, err := processor.ProcessSheet(sheets[0])

    // Validate business rules
    if err != nil {
        t.Fatalf("Failed to process valid schedule: %v", err)
    }

    if schedule.ShiftCount < 150 {
        t.Errorf("Expected 150+ shifts, got %d", schedule.ShiftCount)
    }

    // Check coverage
    if schedule.CoveragePercentage < 95 {
        t.Errorf("Coverage %f%% is below 95%% threshold", schedule.CoveragePercentage)
    }
}
```

## Subtests for Fixture Variants

```go
func TestScheduleParserWithFixtures(t *testing.T) {
    fixtures := []string{
        "valid_schedule.ods",
        "partial_schedule.ods",
        "invalid_schedule.ods",
        "large_schedule.ods",
    }

    for _, fixture := range fixtures {
        t.Run(fixture, func(t *testing.T) {
            data, err := os.ReadFile(fmt.Sprintf("tests/fixtures/ods/%s", fixture))
            if err != nil {
                t.Skipf("Fixture not found: %v", err)
            }

            parser := NewODSParser()
            sheets, _ := parser.ParseWithErrorCollection(data)

            if len(sheets) == 0 {
                t.Fatal("No sheets extracted")
            }

            t.Logf("Successfully parsed %s with %d rows",
                fixture, len(sheets[0].Rows))
        })
    }
}
```

## Loading Fixture Metadata Programmatically

```go
func TestWithFixtureMetadata(t *testing.T) {
    // Load metadata
    metaData, _ := os.ReadFile("tests/fixtures/ods/fixtures.json")

    var meta struct {
        Fixtures []struct {
            Name          string   `json:"name"`
            Type          string   `json:"type"`
            ShiftCount    int      `json:"shift_count"`
            Columns       []string `json:"columns"`
            ExpectedError int      `json:"expected_errors"`
        } `json:"fixtures"`
    }

    json.Unmarshal(metaData, &meta)

    // Test each fixture from metadata
    for _, fixture := range meta.Fixtures {
        t.Run(fixture.Name, func(t *testing.T) {
            data, _ := os.ReadFile(fmt.Sprintf("tests/fixtures/ods/%s", fixture.Name))

            parser := NewODSParser()
            sheets, errs := parser.ParseWithErrorCollection(data)

            // Verify error count matches metadata
            if len(errs) != fixture.ExpectedError {
                t.Errorf("Expected %d errors, got %d", fixture.ExpectedError, len(errs))
            }

            // Verify expected columns
            if len(sheets) > 0 && len(sheets[0].Rows) > 0 {
                actualCols := len(sheets[0].Rows[0].Cells)
                expectedCols := len(fixture.Columns)

                if actualCols != expectedCols {
                    t.Logf("Note: actual columns (%d) differ from metadata (%d)",
                        actualCols, expectedCols)
                }
            }
        })
    }
}
```

## Common Testing Patterns

### Testing Resilience

```go
func TestParserResilienceWithInvalidData(t *testing.T) {
    data, _ := os.ReadFile("tests/fixtures/ods/invalid_schedule.ods")

    parser := NewODSParser()

    // Should not panic
    defer func() {
        if r := recover(); r != nil {
            t.Errorf("Parser panicked: %v", r)
        }
    }()

    sheets, errs := parser.ParseWithErrorCollection(data)

    // Should still return results
    if sheets == nil {
        t.Fatal("Parser returned nil sheets")
    }

    // Should collect errors for debugging
    t.Logf("Parser collected %d errors and %d valid sheets", len(errs), len(sheets))
}
```

### Testing Determinism

```go
func TestFixtureParsing_Deterministic(t *testing.T) {
    data, _ := os.ReadFile("tests/fixtures/ods/valid_schedule.ods")

    parser := NewODSParser()

    // Parse twice
    sheets1, _ := parser.Parse(data)
    sheets2, _ := parser.Parse(data)

    // Results should be identical
    if len(sheets1) != len(sheets2) {
        t.Error("Different number of sheets between parses")
    }

    if len(sheets1[0].Rows) != len(sheets2[0].Rows) {
        t.Error("Different number of rows between parses")
    }
}
```

## Summary

These examples demonstrate:
- Basic fixture loading
- Testing happy path scenarios
- Testing error handling
- Performance testing
- Table-driven tests
- Integration with business logic
- Metadata-driven testing
- Resilience testing

Choose the patterns that fit your testing needs and adapt them to your specific parser implementation.
