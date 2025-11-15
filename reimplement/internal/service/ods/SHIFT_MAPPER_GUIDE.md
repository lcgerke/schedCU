# Shift Instance Mapper - Implementation Guide

## Overview

The `ShiftInstanceMapper` is responsible for converting raw shift data from ODS files (represented as `RawShiftData` structs) into validated `ShiftInstance` entities ready for persistence.

## Validation Rules

### 1. Date Validation

**Format**: Strictly `YYYY-MM-DD` (e.g., `2025-11-20`)

**Rules**:
- Date must be provided (cannot be empty or whitespace-only)
- Date must be valid and follow the YYYY-MM-DD format
- Invalid dates (e.g., February 30) are rejected
- Past dates can be allowed or rejected via configuration (default: allow)

**Examples**:
```
✓ Valid: "2025-11-20"
✗ Invalid: "11/20/2025" (wrong format)
✗ Invalid: "2025-13-01" (invalid month)
✗ Invalid: "2025-02-30" (invalid day)
✗ Invalid: "2025-11-20 extra" (extra content)
```

### 2. Shift Type Validation

**Allowed Values** (case-sensitive): `DAY`, `NIGHT`, `WEEKEND`

**Rules**:
- Shift type must be provided (cannot be empty or whitespace-only)
- Must match exactly one of the allowed values
- Case-sensitive: "day" and "Day" are invalid
- Input is NOT auto-corrected

**Examples**:
```
✓ Valid: "DAY", "NIGHT", "WEEKEND"
✗ Invalid: "day", "Day", "MORNING", "AFTERNOON"
✗ Invalid: "" (empty string)
```

### 3. Required Staffing Validation

**Type**: Positive integer (> 0)

**Rules**:
- Must be provided (cannot be empty)
- Must be parseable as an integer (no decimals)
- Must be positive (> 0), not zero or negative
- Whitespace is trimmed before parsing

**Examples**:
```
✓ Valid: "1", "5", "100"
✗ Invalid: "0" (zero not allowed)
✗ Invalid: "-5" (negative)
✗ Invalid: "3.5" (decimal)
✗ Invalid: "abc" (non-numeric)
```

### 4. Optional Fields

**SpecialtyConstraint** and **StudyType**:
- Optional fields
- Empty strings result in nil pointers
- Non-empty values are trimmed and preserved

## Timezone Strategy

### Design Decisions

1. **Per-Mapper Timezone**: Each mapper instance uses a single timezone for all conversions
2. **Midnight Interpretation**: All dates are interpreted as midnight (00:00:00) in the configured timezone
3. **Flexible Timezone Support**: Supports any `time.Location` (UTC, EST, PST, etc.)
4. **Consistent Representation**: All StartTime/EndTime are set to the same date at midnight

### How It Works

```go
// Create mapper with UTC timezone
mapperUTC := NewShiftInstanceMapper(time.UTC)

// Create mapper with US Eastern timezone
estLoc, _ := time.LoadLocation("America/New_York")
mapperEST := NewShiftInstanceMapper(estLoc)

// Input: "2025-11-20"
// Parsed as: 2025-11-20 00:00:00 in specified timezone
// StartTime and EndTime are both set to this value
```

### Future Enhancements

Currently, both StartTime and EndTime are set to the same date (midnight). Future versions may:
- Add shift hour offsets based on shift type (e.g., DAY = 06:00-18:00)
- Support hospital-specific time mappings
- Add explicit StartTime/EndTime in RawShiftData

## Configuration

### Default Configuration

```go
mapper := NewShiftInstanceMapper(time.UTC)
// AllowPastDates: true (past dates are accepted)
// Timezone: time.UTC
```

### Custom Configuration

```go
config := ShiftMapperConfig{
    AllowPastDates: false,  // Reject past dates
    Timezone: estLoc,        // Use EST timezone
}
mapper := NewShiftInstanceMapperWithConfig(config)
```

## Entity Generation

For each successfully validated `RawShiftData`:

1. **ID**: Generated as new UUID (uuid.New())
2. **ScheduleVersionID**: Set from parameter
3. **ShiftType**: Trimmed from input
4. **StartTime**: Parsed date at midnight in configured timezone
5. **EndTime**: Set to same value as StartTime
6. **Location**: Empty string (not provided in raw data)
7. **StaffMember**: Empty string (not provided in raw data)
8. **CreatedAt**: time.Now() at mapping time
9. **CreatedBy**: Set from parameter
10. **UpdatedAt**: time.Now() at mapping time
11. **UpdatedBy**: Set from parameter

## Integration with ODS Importer

The mapper is designed to be used in the ODS import workflow:

```go
// In ODSImporter.Import()
mapper := NewShiftInstanceMapper(time.UTC)

for _, rawShift := range parsedShifts {
    shiftEntity, err := mapper.MapToShiftInstance(
        rawShift,
        scheduleVersionID,
        userID,
    )
    if err != nil {
        // Handle validation error
        errorCollector.AddMajor("shift", ..., err.Error(), err)
        continue
    }
    
    // Persist to database
    createdShift, dbErr := siRepository.Create(ctx, shiftEntity)
    if dbErr != nil {
        // Handle database error
        continue
    }
}
```

## Error Handling

All validation errors include helpful messages:

```
// Date errors
"date cannot be empty"
"invalid date format: expected YYYY-MM-DD, got \"11/20/2025\": ..."

// Shift type errors
"shift type cannot be empty"
"invalid shift type: \"MORNING\" (must be one of: [DAY NIGHT WEEKEND])"

// Staffing errors
"required staffing cannot be empty"
"required staffing must be a valid integer, got \"abc\": ..."
"required staffing must be positive (> 0), got 0"

// UUID errors
"schedule version ID cannot be nil"
"user ID cannot be nil"
```

## Test Coverage

18 comprehensive test scenarios covering:

1. ✓ Valid input mapping
2. ✓ Invalid date formats (5 subtests)
3. ✓ Invalid shift types (5 subtests)
4. ✓ Valid shift types (3 subtests: DAY, NIGHT, WEEKEND)
5. ✓ Negative required staffing
6. ✓ Zero required staffing
7. ✓ Invalid staffing formats (4 subtests)
8. ✓ Past date rejection
9. ✓ Past dates allowed configuration
10. ✓ Leap year dates
11. ✓ Timezone handling (UTC vs EST)
12. ✓ CreatedAt/CreatedBy timestamps
13. ✓ Optional fields nil handling
14. ✓ Optional fields preservation
15. ✓ Year boundary dates (3 subtests)
16. ✓ Required staffing conversion (3 subtests)
17. ✓ Proper entity return type
18. ✓ Consistency across multiple calls
19. ✓ Whitespace handling

## Performance Characteristics

- **Time Complexity**: O(1) per shift (all operations are constant time)
- **Space Complexity**: O(1) (single entity created)
- **No External Calls**: All operations are pure validation and conversion

## Dependencies

- `github.com/google/uuid` - for UUID generation
- `github.com/schedcu/reimplement/internal/entity` - for ShiftInstance type
- Standard library: `fmt`, `strconv`, `strings`, `time`

## Future Enhancements

1. **Required Staffing Storage**: Add field to ShiftInstance to store validated staffing count
2. **Position Mapping**: Map position from shift type or other fields
3. **Location Mapping**: Extract or map location information
4. **Staff Assignment**: Handle staff member assignment from additional data
5. **Shift Hours**: Add configurable start/end times based on shift type
6. **Timezone DST Handling**: Document daylight saving time behavior
7. **Bulk Mapping**: Add batch mapping method for efficiency

## Usage Summary

```go
// 1. Create mapper instance
mapper := NewShiftInstanceMapper(time.UTC)

// 2. Prepare raw data from parser
raw := RawShiftData{
    Date:             "2025-11-20",
    ShiftType:        "DAY",
    RequiredStaffing: "3",
}

// 3. Map to entity
shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)
if err != nil {
    log.Printf("Validation failed: %v", err)
    return
}

// 4. Persist to database
createdShift, err := repository.Create(ctx, shift)
```
