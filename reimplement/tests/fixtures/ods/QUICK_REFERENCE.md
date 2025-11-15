# ODS Fixtures - Quick Reference

## At a Glance

| Fixture | Shifts | Purpose | Status | Errors |
|---------|--------|---------|--------|--------|
| `valid_schedule.ods` | 150 | Happy path testing | ✓ All valid | 0 |
| `partial_schedule.ods` | 50 | Optional columns | ✓ All valid | 0 |
| `invalid_schedule.ods` | 30 | Error handling | Mixed | 4 |
| `large_schedule.ods` | 1200+ | Performance testing | ✓ All valid | 0 |

## File Locations

```
/home/lcgerke/schedCU/reimplement/tests/fixtures/ods/
├── valid_schedule.ods       (3.8 KB)
├── partial_schedule.ods     (2.7 KB)
├── invalid_schedule.ods     (2.5 KB)
├── large_schedule.ods       (14 KB)
├── fixtures.json            (metadata)
├── README.md                (full documentation)
├── USAGE_EXAMPLES.md        (code examples)
└── ods_fixtures_test.go     (verification tests)
```

## Column Mapping

| Column | Type | Required | Valid Values | Invalid Example |
|--------|------|----------|--------------|-----------------|
| Date | String | Yes | YYYY-MM-DD | "invalid-date" |
| ShiftType | String | Yes | Morning, Afternoon, Night | "INVALID_TYPE" |
| RequiredStaffing | Numeric | Yes | 1-10 | "twenty" or "" |
| SpecialtyConstraint | String | No | Emergency, ICU, Surgery, etc. | (any invalid) |
| StudyType | String | No | Type-A, Type-B, Type-C | (any invalid) |

## Quick Start

### Load and Parse
```go
data, _ := os.ReadFile("tests/fixtures/ods/valid_schedule.ods")
sheets, err := parser.Parse(data)
```

### Handle Errors
```go
sheets, errs := parser.ParseWithErrorCollection(data)
if len(errs) > 0 {
    log.Printf("Collected %d errors", len(errs))
}
```

### Test Performance
```go
data, _ := os.ReadFile("tests/fixtures/ods/large_schedule.ods")
// Should parse 1200+ rows in < 100ms
```

## Metadata File Format

`fixtures.json` contains:
```json
{
  "fixtures": [
    {
      "name": "valid_schedule.ods",
      "type": "valid",
      "shift_count": 150,
      "columns": ["Date", "ShiftType", "RequiredStaffing", "SpecialtyConstraint", "StudyType"],
      "expected_errors": 0
    }
  ]
}
```

## Verification Status

All fixtures verified as:
- Valid ODS 1.2 format
- Valid ZIP archives with correct structure
- Proper XML content
- Correct MIME types
- Reasonable file sizes

**Test Results**: 11/11 tests passing

## Error Patterns in invalid_schedule.ods

The invalid fixture intentionally includes:
1. Missing required field (RequiredStaffing = "")
2. Invalid enumeration (ShiftType = "INVALID_TYPE")
3. Wrong data type (RequiredStaffing = "twenty")
4. Empty rows (skipped cells)

## Regenerating Fixtures

```bash
go run cmd/generate-ods-fixtures/main.go
```

This is deterministic and creates identical files each time.

## Data Characteristics

### Date Generation
- Start: 2025-01-01
- Pattern: Sequential (one per day)
- Range: 6 months (valid), 1+ years (large)

### ShiftType Cycle
- Pattern: Morning → Afternoon → Night → Morning...
- Repeat interval: 3 shifts

### RequiredStaffing
- Pattern: (index % 10) + 1
- Range: 1-10
- Distribution: Evenly distributed

### SpecialtyConstraint
- Pattern: Cycle through 5 specialties
- Values: Emergency, ICU, Surgery, Pediatrics, Cardiology
- Repeat interval: 5 rows

### StudyType
- Pattern: Cycle through 3 types
- Values: Type-A, Type-B, Type-C
- Repeat interval: 3 rows

## Common Test Patterns

### Basic Parsing
```go
// Load, parse, verify no errors
data, _ := os.ReadFile("tests/fixtures/ods/valid_schedule.ods")
sheets, err := parser.Parse(data)
assert.NoError(t, err)
```

### Error Collection
```go
// Load, parse with error collection, verify error count
data, _ := os.ReadFile("tests/fixtures/ods/invalid_schedule.ods")
sheets, errs := parser.ParseWithErrorCollection(data)
assert.Equal(t, 4, len(errs))
```

### Performance Benchmark
```go
// Load large fixture, measure parse time
data, _ := os.ReadFile("tests/fixtures/ods/large_schedule.ods")
// Expected: < 100ms for 1200+ shifts
```

### Column Flexibility
```go
// Load partial fixture, verify missing columns handled
data, _ := os.ReadFile("tests/fixtures/ods/partial_schedule.ods")
sheets, _ := parser.Parse(data)
// Should have 4 columns instead of 5
```

## File Format

All ODS files contain:

```
mimetype                    (46 bytes, uncompressed)
├── application/vnd.oasis.opendocument.spreadsheet
META-INF/manifest.xml       (531 bytes)
├── File declarations
content.xml                 (20KB - 800KB)
├── Table data with rows and cells
├── Namespaces properly declared
└── Valid XML structure
styles.xml                  (1.6 KB)
├── Empty but valid style declarations
settings.xml                (318 bytes)
└── Document settings
```

## Expected Parse Behavior

| Fixture | Behavior | Notes |
|---------|----------|-------|
| valid_schedule.ods | Parse without error | All data valid |
| partial_schedule.ods | Parse without error | Handles missing StudyType |
| invalid_schedule.ods | Parse with 4 errors | Returns valid + errors |
| large_schedule.ods | Parse all rows | Tests performance |

## Integration Points

- No dependencies on other components
- Works with any ODS parser
- Compatible with LibreOffice, OpenOffice, Google Sheets
- Can be imported into Excel with ODS plugin

## Size Expectations

| Fixture | Compressed | Uncompressed (content.xml) | Ratio |
|---------|-----------|--------------------------|-------|
| valid_schedule.ods | 3.8 KB | 104 KB | 3.7% |
| partial_schedule.ods | 2.7 KB | 30 KB | 9% |
| invalid_schedule.ods | 2.5 KB | 20 KB | 12.5% |
| large_schedule.ods | 14 KB | 820 KB | 1.7% |

## Validation Rules

Parser should enforce:
1. **Required Fields**: Date, ShiftType, RequiredStaffing must exist
2. **Data Types**: RequiredStaffing must be numeric
3. **Enumerations**: ShiftType must be Morning, Afternoon, or Night
4. **Format**: Dates must be valid YYYY-MM-DD

## Documentation Files

- **README.md** - Complete reference documentation
- **USAGE_EXAMPLES.md** - 25+ practical code examples
- **QUICK_REFERENCE.md** - This file
- **ods_fixtures_test.go** - Verification test suite

## Troubleshooting

**"Invalid ZIP file" error**
→ Regenerate fixtures: `go run cmd/generate-ods-fixtures/main.go`

**"No sheets extracted"**
→ Verify content.xml exists and is valid XML

**"Parser crashes on invalid data"**
→ Use ParseWithErrorCollection instead of Parse

**"Performance is slow"**
→ Check parsing algorithm, benchmark with large_schedule.ods

## Performance Baseline

Expected parse times on modern hardware:
- valid_schedule (150 shifts): < 10ms
- partial_schedule (50 shifts): < 5ms
- invalid_schedule (30 shifts): < 5ms
- large_schedule (1200+ shifts): < 100ms

## Best Practices

1. Always test with all 4 fixtures
2. Use invalid_schedule for error handling tests
3. Use large_schedule for performance benchmarks
4. Verify metadata matches actual data
5. Check file can be opened with spreadsheet application

## See Also

- README.md - Full documentation
- USAGE_EXAMPLES.md - Code examples
- cmd/generate-ods-fixtures/main.go - Generator source
