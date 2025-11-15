# Error Propagation Quick Reference

**Location**: `internal/service/orchestrator/error_handler.go`
**Tests**: `internal/service/orchestrator/error_handler_test.go`
**Status**: Complete - 25+ unit tests passing

## Overview

The `ErrorPropagator` collects and merges validation results from all three orchestration phases (ODS Import, Amion Scraping, Coverage Calculation) with sophisticated error classification and decision logic.

## Core Components

### ErrorPropagator Struct

```go
propagator := NewErrorPropagator()
```

### Error Hierarchy

The system classifies errors into three severity levels:

#### Critical Errors (STOP)
Stop execution immediately. These should never be recovered from.

**Patterns detected:**
- File parsing/format issues: `"invalid file format"`, `"invalid zip"`, `"parse error"`, `"corrupted"`
- Database constraint violations: `"duplicate"`, `"constraint violation"`, `"unique constraint"`, `"foreign key"`
- Disk space errors: `"no space left"`, `"disk full"`, `"out of memory"`
- System failures: `"permission denied"`, `"connection refused"`, `"connection timeout"`

**Example:**
```go
vr := validation.NewValidationResult()
vr.AddError("file", "failed to parse ODS file: invalid ZIP format")
if !propagator.ShouldContinue(vr, PhaseODSImport) {
    // Stop - cannot continue without valid file
    rollback()
}
```

#### Major Errors (SKIP & CONTINUE)
Skip the current entity, but continue processing others. Critical in Phase 1, non-critical in Phase 2+.

**Patterns detected:**
- Invalid data: `"invalid date"`, `"invalid format"`, `"invalid shift"`
- Unsupported types: `"unsupported"`, `"not supported"`
- Missing fields: `"missing required"`, `"missing field"`, `"required field"`
- Type issues: `"invalid type"`, `"type mismatch"`

**Example:**
```go
vr := validation.NewValidationResult()
vr.AddError("shift", "invalid shift date: expected YYYY-MM-DD")
if propagator.IsMajorError(vr) {
    // In Phase 2+, we can continue
    // Skip this shift and process the next one
}
```

#### Minor Errors (LOG & CONTINUE)
Log and continue. These don't affect processing decisions.

**Examples:**
- `"optional field missing"`
- `"duplicate warning from Amion"`
- `"deprecated format detected"`

## API Reference

### Merging ValidationResults

**Basic merge (no phase context):**
```go
propagator := NewErrorPropagator()
result := propagator.MergeValidationResults(odsResult, amionResult, coverageResult)
```

**Merge with phase context (preserves which phase errored):**
```go
result := propagator.MergeValidationResultsWithContext(map[int]*validation.ValidationResult{
    0: odsResult,      // PhaseODSImport
    1: amionResult,    // PhaseAmionScrape
    2: coverageResult, // PhaseCoverageCalculation
})
```

Returns a new `ValidationResult` containing:
- All errors, warnings, and infos combined
- All context values merged (later values overwrite)
- `"phases_with_errors"` context field listing which phases had errors

### Decision Logic

**Should Continue?**
```go
// Returns false if should stop, true if can continue
shouldContinue := propagator.ShouldContinue(vr, phase)

// Decision rules:
// - Warnings only: Continue
// - Minor errors: Continue
// - Major errors in Phase 2+: Continue (skip entity)
// - Major errors in Phase 1: Stop (foundation is broken)
// - Critical errors (any phase): Stop
```

**Error Classification:**
```go
isCritical := propagator.IsCriticalError(vr, phase)   // Critical?
isMajor := propagator.IsMajorError(vr)                // Major?
isMinor := propagator.IsMinorError(vr)                // Minor?
```

## Usage Patterns

### In Orchestrator Workflow

```go
// Phase 1: ODS Import
odsResult, odsErr := importer.Import(ctx, content, hospitalID, userID)
if !propagator.ShouldContinue(odsResult, PhaseODSImport) {
    // Critical error - stop entire workflow
    return nil, odsResult
}

// Phase 2: Amion Scraping
amionResult, amionErr := scraper.Scrape(ctx, ...)
if !propagator.ShouldContinue(amionResult, PhaseAmionScrape) {
    // Critical error in Phase 2 - stop
    return sv, amionResult
}

// Phase 3: Coverage Calculation
coverageResult, covErr := calculator.Calculate(ctx, ...)
if !propagator.ShouldContinue(coverageResult, PhaseCoverageCalculation) {
    // Critical error in Phase 3 - still return with warnings
    return sv, coverageResult
}

// Merge all results for final response
merged := propagator.MergeValidationResultsWithContext(map[int]*validation.ValidationResult{
    0: odsResult,
    1: amionResult,
    2: coverageResult,
})

return sv, merged
```

### Error Collection from Multiple Sources

```go
// Collect errors from different validators
results := []*validation.ValidationResult{}

for i, shift := range shifts {
    vr := validateShift(shift)
    results = append(results, vr)
}

// Merge all validation results
merged := propagator.MergeValidationResults(results...)

// Report on what went wrong
if merged.HasErrors() {
    log.Printf("Validation errors in %d shifts:", merged.ErrorCount())
    for _, err := range merged.Errors {
        log.Printf("  %s: %s", err.Field, err.Message)
    }
}
```

## Test Coverage

Total tests: 25+

**Merge functionality (6 tests):**
- Empty merge
- Single result
- Multiple results
- Phase context labels
- Context preservation
- Conflict resolution

**Decision logic (6 tests):**
- Continue on warnings
- Continue on major errors in Phase 2+
- Stop on critical errors
- Stop on constraint violations
- Stop on disk full
- Major errors in Phase 1

**Error classification (3 tests):**
- Critical error detection
- Major error detection
- Minor error detection

**Error patterns (2 tests):**
- Critical patterns (10 patterns tested)
- Major patterns (5 patterns tested)

**Integration tests (8+ tests):**
- Nil handling
- Empty results
- Conflicting severities
- Info message preservation
- Phase context tracking
- And more...

## Performance

All tests run in <1ms total.
No allocations for matching (string patterns are pre-defined).

## Future Enhancements

Possible improvements:
1. Custom error classifier registration (pluggable patterns)
2. Error codes instead of string matching (more reliable)
3. Structured logging integration for error events
4. Metrics collection (errors per phase, by type, etc.)
5. Error recovery strategies (retry, fallback, etc.)
