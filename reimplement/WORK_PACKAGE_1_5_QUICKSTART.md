# Work Package [1.5] - Quick Start Guide

## Overview

**ODSImporter** orchestrates the complete workflow of importing ODS files into the database:
1. Parse ODS file → 2. Create ScheduleVersion → 3. Import shifts → 4. Collect errors → 5. Return results

## Key Components

### ODSImporter
Main service that coordinates the import workflow.

```go
importer := ods.NewODSImporter(parser, svRepo, siRepo)
sv, err := importer.Import(ctx, fileContent, hospitalID, userID)
```

### ODSErrorCollector
Thread-safe error aggregation with 4 severity levels:
- **Critical** - Stops import immediately (first only)
- **Major** - Skips entity, continues import
- **Minor** - Logs warning, doesn't affect import
- **Info** - Informational messages

### Domain Entities

**ScheduleVersion** - Represents a versioned schedule snapshot
```go
type ScheduleVersion struct {
    ID          uuid.UUID
    HospitalID  uuid.UUID
    Version     int
    Status      VersionStatus  // draft, published, etc.
    StartDate   time.Time
    EndDate     time.Time
    // ... timestamps, audit fields
}
```

**ShiftInstance** - Individual shift within a schedule version
```go
type ShiftInstance struct {
    ID                    uuid.UUID
    ScheduleVersionID     uuid.UUID
    ShiftType             string  // "Morning", "Night", etc.
    Position              string  // "Doctor", "Nurse", etc.
    StartTime             *time.Time
    EndTime               *time.Time
    Location              string
    StaffMember           string
    // ... optional fields, timestamps
}
```

## Usage Pattern

### 1. Create Importer
```go
parser := &MyODSParser{}  // implements ODSParserInterface
importer := ods.NewODSImporter(parser, svRepo, siRepo)
```

### 2. Execute Import
```go
scheduleVersion, err := importer.Import(
    ctx,
    odsFileContent,
    hospitalID,
    userID,
)
```

### 3. Handle Results
```go
if err != nil {
    // Partial or complete failure
    result := importer.GetValidationResult()

    if result.HasErrors() {
        for _, e := range result.Errors {
            log.Printf("ERROR: %s.%s = %s", e.Field, e.Message)
        }
    }

    metrics := importer.GetErrorMetrics()
    log.Printf("Created %d/%d shifts",
        metrics["created_shifts"],
        metrics["total_shifts"],
    )
}

// scheduleVersion is always created if no critical error
if scheduleVersion != nil {
    log.Printf("Schedule version: %s (status=%s)",
        scheduleVersion.ID,
        scheduleVersion.Status,
    )
}
```

## Error Handling

### Critical Errors (Stop Import)
- File parsing fails
- Schedule version creation fails
- Invalid input parameters (nil IDs)

### Major Errors (Skip Entity)
- Individual shift creation fails
- Database constraint violations
- Foreign key violations

### Recovery Pattern
```go
sv, err := importer.Import(ctx, content, hID, uID)

if err != nil {
    result := importer.GetValidationResult()

    switch {
    case result.ErrorCount() > 0:
        // Major errors occurred (partial success)
        log.Printf("Partial failure: %s", err)
        // sv is still valid, use it

    default:
        // Critical error (complete failure)
        log.Printf("Import failed: %s", err)
        // sv is nil
    }
}
```

## Validation Result

```go
type ValidationResult struct {
    Errors   []SimpleValidationMessage  // Field-level errors
    Warnings []SimpleValidationMessage  // Field-level warnings
    Infos    []SimpleValidationMessage  // Informational
    Context  map[string]interface{}     // Metrics
}

// Methods
result.HasErrors()       // true if any ERROR
result.HasWarnings()     // true if any WARNING
result.ErrorCount()      // Total errors
result.WarningCount()    // Total warnings
result.IsValid()         // len(Errors) == 0
```

## Common Scenarios

### Scenario 1: Full Success
```
Input: Valid ODS with 10 shifts
Expected: ScheduleVersion created, 10 shifts imported, no error
```

### Scenario 2: Partial Success
```
Input: Valid ODS with 10 shifts, shift #5 has invalid data
Expected: ScheduleVersion created, 9 shifts imported, error returned
Result: err != nil, but sv is valid
        ValidationResult shows shift #5 failed
```

### Scenario 3: Critical Failure
```
Input: Malformed ODS file
Expected: ScheduleVersion NOT created, no shifts imported, error returned
Result: err != nil, sv is nil
        ValidationResult shows parse error
```

## Testing Patterns

### Mock Parser
```go
parser := &MockODSParser{
    parseResult: &ParsedSchedule{
        StartDate: "2025-01-01",
        EndDate:   "2025-01-31",
        Shifts: []*ParsedShift{...},
    },
}
```

### Mock Repository
```go
svRepo := &MockScheduleVersionRepository{}
siRepo := &MockShiftInstanceRepository{
    failCount: 2,  // Fail on 2nd call
}
```

### Execute Test
```go
importer := ods.NewODSImporter(parser, svRepo, siRepo)
sv, err := importer.Import(ctx, content, hID, uID)

// Assertions
if err == nil {
    t.Fatal("expected error for partial failure")
}
if sv == nil {
    t.Fatal("expected schedule version")
}
if len(siRepo.createdShifts) != 1 {
    t.Errorf("expected 1 shift, got %d", len(siRepo.createdShifts))
}
```

## Files to Know

| File | Purpose |
|------|---------|
| `internal/service/ods/importer.go` | Main service |
| `internal/service/ods/error_collector.go` | Error aggregation |
| `internal/service/ods/parser_interface.go` | Parser contract |
| `internal/entity/schedule_version.go` | Domain entity |
| `internal/entity/shift_instance.go` | Domain entity |
| `internal/repository/schedule_version_repository.go` | Interface |
| `internal/repository/shift_instance_repository.go` | Interface |
| `internal/service/ods/importer_test.go` | Tests |

## Next Steps

1. **Implement Repositories** - Create PostgreSQL repository implementations
2. **Implement ODS Parser** - Create actual ODS file parser
3. **Create API Handler** - Expose import via HTTP endpoint
4. **Integration Testing** - Test with real database (Testcontainers)
5. **Monitoring** - Add metrics and logging

## Support Interfaces

### ODSParserInterface
```go
type ODSParserInterface interface {
    Parse(odsContent []byte) (*ParsedSchedule, error)
    ParseWithErrorCollection(odsContent []byte) (*ParsedSchedule, []error)
}
```

### ScheduleVersionRepository
```go
type ScheduleVersionRepository interface {
    Create(ctx context.Context, sv *ScheduleVersion) (*ScheduleVersion, error)
    GetByID(ctx context.Context, id uuid.UUID) (*ScheduleVersion, error)
    // ... other methods
}
```

### ShiftInstanceRepository
```go
type ShiftInstanceRepository interface {
    Create(ctx context.Context, shift *ShiftInstance) (*ShiftInstance, error)
    CreateBatch(ctx context.Context, shifts []*ShiftInstance) (int, error)
    // ... other methods
}
```
