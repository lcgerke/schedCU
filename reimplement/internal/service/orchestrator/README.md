# Orchestrator Interfaces

This package defines the service interfaces for the orchestration layer of schedCU Phase 1. It provides a clean abstraction for coordinating schedule imports, Amion scraping, and coverage calculations.

## Overview

The orchestrator package provides four main interfaces:

1. **ODSImportService** - Imports ODS schedule files
2. **AmionScraperService** - Scrapes schedule data from Amion
3. **CoverageCalculatorService** - Calculates coverage metrics
4. **ScheduleOrchestrator** - Coordinates the complete workflow

## Key Design Principles

### Interface-based Architecture

All interfaces are designed to be mockable for testing. Each interface focuses on a single responsibility:

- **ODSImportService**: Parse and import ODS files into the database
- **AmionScraperService**: Fetch and scrape schedule data from Amion
- **CoverageCalculatorService**: Calculate coverage metrics efficiently
- **ScheduleOrchestrator**: Coordinate all services in a unified workflow

### Error Handling

All services return:
- An error for critical failures
- A ValidationResult with collected warnings/info
- Partial results when possible (some data might be valid even if operation fails)

### Concurrency

All interfaces are designed to be thread-safe. Multiple concurrent operations are supported as long as they don't conflict (e.g., different hospitals or schedule versions).

### Query Efficiency

The CoverageCalculatorService enforces single batch query pattern to prevent N+1 problems. All services minimize database round-trips.

## Interface Documentation

### ODSImportService

Imports a schedule from an ODS file.

```go
type ODSImportService interface {
    ImportSchedule(
        ctx context.Context,
        filePath string,
        hospitalID uuid.UUID,
        userID uuid.UUID,
    ) (*entity.ScheduleVersion, *validation.ValidationResult, error)
}
```

**Returns:**
- ScheduleVersion with status = DRAFT
- ValidationResult with any errors, warnings, or info
- Error for critical failures (file not found, etc.)

**Error Codes in ValidationResult:**
- INVALID_FILE_FORMAT - File is not a valid ODS file
- PARSE_ERROR - Could not parse file contents
- MISSING_REQUIRED_FIELD - Required field missing from data
- DUPLICATE_ENTRY - Duplicate entries found
- DATABASE_ERROR - Database operation failed
- UNKNOWN_ERROR - Unexpected error

### AmionScraperService

Scrapes schedule data from Amion.

```go
type AmionScraperService interface {
    ScrapeSchedule(
        ctx context.Context,
        startDate time.Time,
        monthCount int,
        hospitalID uuid.UUID,
        userID uuid.UUID,
    ) ([]entity.Assignment, *validation.ValidationResult, error)
}
```

**Parameters:**
- startDate: Start date for scraping (typically first day of month)
- monthCount: Number of months to scrape (e.g., 3)
- hospitalID: Hospital identifier (for routing/audit)
- userID: User performing scrape (for audit trail)

**Returns:**
- Assignments with source = AMION
- ValidationResult with warnings about duplicates, parse errors, etc.
- Error for critical failures (auth, network, etc.)

### CoverageCalculatorService

Calculates coverage metrics for a schedule.

```go
type CoverageCalculatorService interface {
    Calculate(
        ctx context.Context,
        scheduleVersionID uuid.UUID,
    ) (*CoverageMetrics, error)
}
```

**Guarantees:**
- Uses exactly 1 database query (batch query pattern)
- No N+1 query patterns
- Thread-safe for concurrent calculations

**Returns CoverageMetrics:**
- CoveragePercentage (0-100)
- AssignedPositions and RequiredPositions counts
- UncoveredShifts and OverallocatedShifts lists
- Arbitrary Details map for custom metrics

### ScheduleOrchestrator

Coordinates the complete import workflow.

```go
type ScheduleOrchestrator interface {
    ExecuteImport(
        ctx context.Context,
        filePath string,
        hospitalID uuid.UUID,
        userID uuid.UUID,
    ) (*OrchestrationResult, error)

    GetOrchestrationStatus() OrchestrationStatus
}
```

**ExecuteImport Workflow:**
1. Import schedule from ODS file
2. Calculate coverage metrics
3. Return unified OrchestrationResult

**Status Values:**
- IDLE - No operation running
- IN_PROGRESS - Operation currently executing
- COMPLETED - Last operation completed successfully
- FAILED - Last operation failed

## Supporting Types

### OrchestrationResult

Contains the complete output of an import operation:

```go
type OrchestrationResult struct {
    ScheduleVersion    *entity.ScheduleVersion
    Assignments        []entity.Assignment
    Coverage           *CoverageMetrics
    ValidationResult   *validation.ValidationResult
    Duration           time.Duration
    CompletedAt        time.Time
    Metadata           map[string]interface{}
}
```

### CoverageMetrics

Calculated coverage information:

```go
type CoverageMetrics struct {
    ScheduleVersionID      uuid.UUID
    CoveragePercentage     float64  // 0-100
    AssignedPositions      int
    RequiredPositions      int
    UncoveredShifts        []*entity.ShiftInstance
    OverallocatedShifts    []*entity.ShiftInstance
    CalculatedAt           time.Time
    Details                map[string]interface{}
}
```

## Mock Implementations

The package provides mock implementations for testing:

- **MockODSImportService** - For testing services that depend on ODS importing
- **MockAmionScraperService** - For testing services that depend on Amion scraping
- **MockCoverageCalculatorService** - For testing services that depend on coverage calculation
- **MockScheduleOrchestrator** - For testing services that depend on orchestration

### Using Mocks

```go
// Create a mock orchestrator
orchestrator := NewMockScheduleOrchestrator()

// Override default behavior
orchestrator.ExecuteImportFunc = func(
    ctx context.Context,
    filePath string,
    hospitalID uuid.UUID,
    userID uuid.UUID,
) (*OrchestrationResult, error) {
    // Custom logic
    return &OrchestrationResult{...}, nil
}

// Execute and verify
result, err := orchestrator.ExecuteImport(ctx, "/tmp/test.ods", hospitalID, userID)

// Check calls
if len(orchestrator.ExecuteImportCalls) != 1 {
    t.Fatalf("expected 1 call, got %d", len(orchestrator.ExecuteImportCalls))
}
```

## Testing Interface Contracts

The package includes comprehensive interface contract tests:

```bash
go test -v ./internal/service/orchestrator/...
```

Tests verify:
- Interface existence and signature correctness
- Mock implementations return valid default values
- Status transitions work correctly
- Call tracking works properly
- All interfaces are properly implemented

## Transaction Requirements

### ODSImportService

The entire import operation should be atomic at the ScheduleVersion level:
- Either all shifts are imported successfully
- Or the operation fails and database remains consistent
- Uses soft deletes for non-destructive updates

### AmionScraperService

- Creates assignments in a single batch operation
- Handles partial failures gracefully (some months may fail)
- Reports duplicates as warnings, not errors

### CoverageCalculatorService

- Read-only operation (no database writes)
- Single batch query for efficiency
- Thread-safe for concurrent calculations

### ScheduleOrchestrator

- Coordinates operations atomically where possible
- Returns partial state on intermediate failures
- All operations included in single ValidationResult

## Error Handling Strategy

1. **Critical Errors** → Return error, operation fails
2. **Validation Errors** → Include in ValidationResult, not as error
3. **Warnings** → Include in ValidationResult, operation continues
4. **Info** → Include in ValidationResult for diagnostics

Example:

```go
result, err := orchestrator.ExecuteImport(ctx, filePath, hospitalID, userID)
if err != nil {
    // Critical failure (file not found, auth failed, etc.)
    log.Printf("Critical error: %v", err)
    return
}

if result.ValidationResult.HasErrors() {
    // Validation errors (format issues, missing fields, etc.)
    for _, e := range result.ValidationResult.Errors {
        log.Printf("Validation error: %s: %s", e.Field, e.Message)
    }
}

if result.ValidationResult.HasWarnings() {
    // Non-critical warnings (duplicates, partial failures, etc.)
    for _, w := range result.ValidationResult.Warnings {
        log.Printf("Warning: %s: %s", w.Field, w.Message)
    }
}

// Proceed with results
log.Printf("Coverage: %.1f%%", result.Coverage.CoveragePercentage)
```

## Implementation Checklist

When implementing these interfaces:

1. **ODSImportService**
   - [ ] Read and validate ODS file
   - [ ] Parse spreadsheet content
   - [ ] Create ScheduleVersion entity
   - [ ] Create ShiftInstance entities
   - [ ] Save to database atomically
   - [ ] Collect validation errors/warnings
   - [ ] Return complete ValidationResult

2. **AmionScraperService**
   - [ ] Scrape Amion URLs with rate limiting
   - [ ] Extract shift data
   - [ ] Create Assignment entities
   - [ ] Detect and report duplicates
   - [ ] Handle partial failures
   - [ ] Batch create assignments
   - [ ] Return ValidationResult with warnings

3. **CoverageCalculatorService**
   - [ ] Load all assignments (single query)
   - [ ] Calculate coverage percentages
   - [ ] Identify uncovered shifts
   - [ ] Identify overallocated shifts
   - [ ] Thread-safe execution
   - [ ] Return complete CoverageMetrics

4. **ScheduleOrchestrator**
   - [ ] Coordinate ODS import
   - [ ] Call coverage calculator
   - [ ] Aggregate results
   - [ ] Track status transitions
   - [ ] Measure duration
   - [ ] Return OrchestrationResult

## Files

- **interfaces.go** - Interface definitions with comprehensive documentation
- **mocks.go** - Mock implementations for testing
- **interfaces_test.go** - Interface contract tests

## Testing Strategy

All interfaces are tested with:
- Contract tests verifying interface signatures
- Mock implementations with call tracking
- Status transition tests
- Default behavior verification
- Benchmark tests for performance validation

Run tests with:
```bash
go test -v ./internal/service/orchestrator/...
go test -bench=. ./internal/service/orchestrator/...
```
