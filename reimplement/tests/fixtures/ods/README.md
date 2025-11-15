# ODS Test Fixtures

This directory contains test fixtures for the ODS (OpenDocument Spreadsheet) parsing functionality. All fixtures are real binary ODS files (ZIP archives containing XML) that can be parsed by any ODS-compatible application.

## Fixture Overview

### 1. `valid_schedule.ods`
- **Type**: Valid/Complete fixture
- **Shifts**: 150 shifts spanning 6 months
- **Columns**: Date, ShiftType, RequiredStaffing, SpecialtyConstraint, StudyType
- **Expected Errors**: 0
- **Use Case**: Happy path testing - parser should extract all data without errors
- **Characteristics**:
  - All required columns present
  - All optional columns present
  - All cells properly formatted
  - No empty cells in required fields
  - Dates span from 2025-01-01 to 2025-05-30

### 2. `partial_schedule.ods`
- **Type**: Partial/Minimal fixture
- **Shifts**: 50 shifts
- **Columns**: Date, ShiftType, RequiredStaffing, SpecialtyConstraint (StudyType omitted)
- **Expected Errors**: 0
- **Use Case**: Testing optional column handling
- **Characteristics**:
  - Missing optional StudyType column
  - Some empty cells in non-critical fields
  - All required columns present and valid
  - Proper formatting maintained

### 3. `invalid_schedule.ods`
- **Type**: Invalid/Error fixture
- **Shifts**: 30 shifts with intentional errors
- **Columns**: Date, ShiftType, RequiredStaffing, SpecialtyConstraint, StudyType
- **Expected Errors**: 4
- **Use Case**: Testing error collection and resilience
- **Characteristics**:
  - Row with missing RequiredStaffing (required field)
  - Row with invalid ShiftType ("INVALID_TYPE")
  - Row with non-numeric RequiredStaffing ("twenty")
  - Empty rows scattered throughout
  - Parser should collect errors without crashing
  - Valid rows should still be extracted

### 4. `large_schedule.ods`
- **Type**: Performance/Scale fixture
- **Shifts**: 1,200+ shifts spanning 1+ years
- **Columns**: Date, ShiftType, RequiredStaffing, SpecialtyConstraint, StudyType
- **Expected Errors**: 0
- **Use Case**: Performance testing and stress testing
- **Characteristics**:
  - All data is valid
  - Large file size (14+ KB compressed)
  - Tests parser performance with realistic data volumes
  - Can be used to verify memory efficiency

## Fixture Metadata

Fixture metadata is stored in `fixtures.json`:
```json
{
  "fixtures": [
    {
      "name": "valid_schedule.ods",
      "type": "valid",
      "shift_count": 150,
      "columns": [...],
      "expected_errors": 0
    },
    ...
  ]
}
```

Each fixture entry contains:
- `name`: Filename of the ODS fixture
- `type`: Category (valid, partial, invalid, large)
- `shift_count`: Number of shift rows in the fixture
- `columns`: Array of column names in the spreadsheet
- `expected_errors`: Number of validation errors expected during parsing

## ODS File Format

All files are valid OpenDocument Spreadsheet (ODS) files conforming to the ODS 1.2 specification. Each ODS file is a ZIP archive containing:

- **mimetype**: Text file declaring MIME type
- **META-INF/manifest.xml**: File manifest with MIME types
- **content.xml**: Main spreadsheet data in XML format
- **styles.xml**: Style definitions
- **settings.xml**: Document settings

The spreadsheet content uses standard ODS namespaces and can be opened in:
- LibreOffice Calc
- OpenOffice Calc
- Google Sheets
- Microsoft Excel (with ODS support plugin)
- Any ODS-compliant spreadsheet application

## Column Definitions

### Date
- **Type**: Text/String in YYYY-MM-DD format
- **Example**: "2025-01-15"
- **Required**: Yes
- **Validation**: Must be a valid date

### ShiftType
- **Type**: Text/String
- **Valid Values**: "Morning", "Afternoon", "Night"
- **Invalid Values**: "INVALID_TYPE"
- **Required**: Yes
- **Validation**: Must be one of the valid values

### RequiredStaffing
- **Type**: Numeric (Integer)
- **Valid Values**: 1-10
- **Invalid Values**: Empty string, "twenty"
- **Required**: Yes
- **Validation**: Must be a positive integer

### SpecialtyConstraint
- **Type**: Text/String
- **Valid Values**: "Emergency", "ICU", "Surgery", "Pediatrics", "Cardiology"
- **Required**: No (optional)
- **Validation**: Free-form string, any value acceptable

### StudyType
- **Type**: Text/String
- **Valid Values**: "Type-A", "Type-B", "Type-C"
- **Required**: No (optional)
- **Validation**: Free-form string, any value acceptable

## Using Fixtures in Tests

### Parsing Valid Fixture
```go
func TestParseValidSchedule(t *testing.T) {
    data, err := os.ReadFile("tests/fixtures/ods/valid_schedule.ods")
    require.NoError(t, err)

    parser := NewODSParser()
    sheets, err := parser.Parse(data)

    require.NoError(t, err)
    assert.Equal(t, 150, len(sheets[0].Rows))
}
```

### Parsing Invalid Fixture (with Error Collection)
```go
func TestParseInvalidSchedule(t *testing.T) {
    data, err := os.ReadFile("tests/fixtures/ods/invalid_schedule.ods")
    require.NoError(t, err)

    parser := NewODSParser()
    sheets, errs := parser.ParseWithErrorCollection(data)

    // Should return sheets despite errors
    assert.NotNil(t, sheets)

    // Should collect expected errors
    assert.Equal(t, 4, len(errs))
}
```

### Performance Testing
```go
func BenchmarkParselargeSchedule(b *testing.B) {
    data, _ := os.ReadFile("tests/fixtures/ods/large_schedule.ods")
    parser := NewODSParser()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        parser.Parse(data)
    }
}
```

## Regenerating Fixtures

Fixtures are generated using the `cmd/generate-ods-fixtures` utility. To regenerate:

```bash
go run cmd/generate-ods-fixtures/main.go
```

This will:
1. Create all ODS files with current data
2. Update fixture counts and metadata
3. Preserve all error injection patterns

### Modifying Fixtures

To customize fixture generation:
1. Edit `cmd/generate-ods-fixtures/main.go`
2. Modify the `generateAllFixtures()` function
3. Adjust `generateValidShifts()`, `generatePartialShifts()`, etc. as needed
4. Run generator again

### Adding New Fixtures

To add a new fixture type:
1. Add new `FixtureConfig` in `generateAllFixtures()`
2. Create corresponding `generate*Shifts()` function
3. Inject test patterns as needed
4. Run generator

## Validation Rules

The parser implements the following validation:

### Data Type Validation
- Numeric fields must contain only digits
- Date fields must be valid ISO 8601 dates
- Text fields accept any value

### Required Field Validation
- Date, ShiftType, RequiredStaffing are required
- Missing required fields trigger validation errors
- Parser collects all errors, doesn't fail-fast

### Business Rule Validation
- ShiftType must be one of: Morning, Afternoon, Night
- RequiredStaffing must be positive integer >= 1
- Invalid values trigger descriptive error messages

## Performance Characteristics

Current performance with large fixture:
- File size: ~14 KB compressed
- Uncompressed XML: ~800+ KB
- Shift count: 1,200+
- Expected parse time: <100ms on modern hardware
- Memory usage: <50MB during parsing

## Troubleshooting

### "Invalid ZIP file" Error
- Ensure ODS file exists and is readable
- Verify file wasn't corrupted during download
- Try regenerating with `go run cmd/generate-ods-fixtures/main.go`

### Parser Returns No Sheets
- Verify `content.xml` exists in the ODS file
- Check that `content.xml` contains valid XML
- Ensure table elements have proper namespaces

### Missing Columns
- Verify column names in fixture metadata match actual ODS columns
- Check that all rows have cells for declared columns
- Regenerate if fixture data became inconsistent

## Testing Strategy

### Unit Tests
- Test each fixture type individually
- Verify error counts match expected values
- Check column extraction accuracy

### Integration Tests
- Test fixture loading with actual parser
- Verify end-to-end parsing pipeline
- Test error handling and recovery

### Performance Tests
- Benchmark parsing of large fixture
- Monitor memory usage during parsing
- Profile to identify bottlenecks

## Dependencies

- Go 1.20+
- Standard library only for fixture generation
- Optional: LibreOffice/OpenOffice for manual verification

## Notes

- All fixtures use consistent datetime format (YYYY-MM-DD)
- Empty cells are represented as `<table:table-cell/>`
- Type information is preserved in value-type attributes
- Fixtures are deterministic and reproducible
