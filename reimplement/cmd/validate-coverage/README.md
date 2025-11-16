# Coverage Validation and Natural Language Summary Tools

This directory contains tools for validating coverage in radiology schedules and generating human-readable summaries from complex coverage data.

## Tools

### 1. `parse_coverage_grid.py` - Coverage Grid Validator

Validates that every study type (CT, MRI, X-Ray, etc.) has coverage for both weekdays and weekends.

**Usage:**
```bash
python3 parse_coverage_grid.py [path/to/file.ods]
```

**Output:**
- Study type coverage breakdown (23 study types)
- Category summary (CT, MRI, X-Ray, Ultrasound)
- Time period coverage (weekday vs weekend)
- Gap detection (missing coverage)
- Validation status

**Example:**
```bash
$ python3 parse_coverage_grid.py /home/user/schedCU/cuSchedNormalized.ods

✅ VALIDATION PASSED
   Total study types: 23
   Total categories: 8
   Coverage gaps: 0
```

### 2. `natural_language_summary.py` - Human-Readable Summary Generator

Transforms technical coverage data into plain English summaries suitable for administrators, schedulers, and non-technical stakeholders.

**Usage:**
```bash
python3 natural_language_summary.py [path/to/file.ods]
```

**Generates:**
- **Executive Summary** - High-level overview for administrators
- **Plain English Summary** - Non-technical explanation
- **Modality Summary** - Coverage by imaging type (CT, MRI, etc.)
- **Hospital Summary** - Coverage by hospital location
- **Time Coverage Summary** - When services are available
- **Specialty Summary** - Neuro vs Body radiology coverage
- **Key Insights** - Important observations and talking points
- **Validation Status** - Pass/fail with recommendations

**Example Output:**
```
EXECUTIVE COVERAGE SUMMARY

🏥 COVERAGE STATUS: ✅ FULLY OPERATIONAL

All 23 imaging study types have complete 24/7 coverage
across 7 weekday time periods and 5 weekend time periods.

WHAT THIS MEANS:

✓ Every type of medical imaging scan can be performed at any time
✓ No gaps in coverage - patients can be served 24 hours a day, 7 days a week
✓ Both weekday and weekend shifts are fully staffed
✓ All hospital locations have adequate radiologist coverage
```

### 3. `main.go` - Go-Based ODS Parser (Experimental)

Attempted Go implementation using excelize library. Currently has EOF issues with LibreOffice ODS files.

**Status:** Not working - use Python tools instead

### 4. `parse_ods.py` - Alternative Shift Parser

Alternative approach that tries to parse shifts row-by-row. Works better with traditional shift schedules rather than coverage grids.

**Status:** Working but less suitable for current ODS format

## ODS File Format

The tools expect ODS files organized as **coverage grids**:

- **Rows** = Study types/modalities (e.g., "CPMC CT Neuro", "Allen MRI Body")
- **Columns** = Shift positions (e.g., "Mid Body", "ON1", "ON2")
- **Sheets** = Time periods × Day types (e.g., "Mid Weekday Body 5-6 pm", "ON Weekend Neuro 1 am - 8 am")
- **'x' marker** = Coverage exists for that study type during that shift

Example:
```
Sheet: "Mid Weekday Body 5-6 pm"

                | Mid Body | Mid Neuro | Mid3 |
----------------|----------|-----------|------|
CPMC CT Body    |    x     |           |      |
CPMC CT Neuro   |          |     x     |      |
CPMC MRI Body   |    x     |           |      |
...
```

## Validation Logic

### Coverage Validation

1. **Parse all sheets** from ODS file
2. **Extract coverage markers** ('x' cells)
3. **Group by study type** and day type (weekday/weekend)
4. **Check for gaps** - every study type must have both weekday AND weekend coverage
5. **Report results** - pass/fail with detailed breakdown

### Natural Language Generation

1. **Aggregate data** by multiple dimensions (modality, hospital, specialty)
2. **Calculate statistics** (total studies, coverage percentages, time periods)
3. **Generate sections**:
   - Executive summary for decision-makers
   - Plain English for general audience
   - Technical details for schedulers
   - Recommendations based on gaps
4. **Format output** with clear hierarchy and visual indicators

## Example Workflow

**Validate a new schedule:**
```bash
# 1. Check for coverage gaps
python3 parse_coverage_grid.py new_schedule.ods

# 2. Generate summary for stakeholders
python3 natural_language_summary.py new_schedule.ods > summary.txt

# 3. Distribute summary to administrators
email summary.txt to admin@hospital.com
```

## Integration with Go Coverage Algorithm

The Go package `/v2/internal/service/coverage/natural_language.go` provides similar functionality integrated with the coverage calculation algorithm:

```go
// Calculate coverage metrics
metrics := coverage.ResolveCoverage(assignments, requirements)

// Generate natural language summary
summary := coverage.GenerateNaturalLanguageSummary(metrics, "November 2025 Schedule")

// Print formatted summary
fmt.Println(summary.FormatAsText())
```

See `/v2/internal/service/coverage/natural_language.go` for the Go implementation.

## Dependencies

### Python Tools
- Python 3.6+
- Standard library only (zipfile, xml.etree, collections)

### Go Tools
- Go 1.20+
- github.com/xuri/excelize/v2 (for ODS parsing)
- github.com/google/uuid (for UUID handling)

## Testing

Validated on `cuSchedNormalized.ods`:
- ✅ 23 study types
- ✅ 4 hospitals (CPMC, Allen, NYPLH, CHONY)
- ✅ 4 modalities (CT, MRI, Ultrasound, X-Ray)
- ✅ 7 weekday time periods
- ✅ 5 weekend time periods
- ✅ 0 coverage gaps

## Future Enhancements

- [ ] Support for additional ODS formats
- [ ] Gap severity scoring (critical vs minor gaps)
- [ ] Historical coverage trend analysis
- [ ] Export to PDF/HTML for reporting
- [ ] Integration with scheduling systems
- [ ] Real-time validation during schedule creation
- [ ] Coverage forecasting based on historical data
- [ ] Automated email reports

## License

Part of the schedCU radiology scheduling system.
