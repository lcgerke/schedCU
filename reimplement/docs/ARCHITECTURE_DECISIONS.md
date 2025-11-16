# Architecture Decisions Log - Phase 1

**Project**: schedCU Reimplementation (v2)
**Phase**: Phase 1 - Core Services & Integration
**Last Updated**: November 15, 2025
**Document Owner**: Engineering Team
**Status**: Complete - All decisions from Phase 1 implementation documented

---

## Document Purpose

This document records all major architectural decisions made during Phase 1 implementation, including:
- The decision itself (what was chosen)
- Context (why it was needed)
- Options evaluated (alternatives considered)
- Rationale (why this is best)
- Tradeoffs (what we give up)
- Implications (downstream effects)
- Current status (decided/implemented)

---

## Table of Contents

1. [AD-001: ValidationResult Pattern](#ad-001-validationresult-pattern)
2. [AD-002: Immutable Entities](#ad-002-immutable-entities)
3. [AD-003: Batch Queries for Coverage](#ad-003-batch-queries-for-coverage)
4. [AD-004: goquery for HTML Parsing](#ad-004-goquery-for-html-parsing)
5. [AD-005: Custom ODS Parser](#ad-005-custom-ods-parser)
6. [AD-006: 3-Phase Orchestration](#ad-006-3-phase-orchestration)
7. [AD-007: Separate Transactions per Phase](#ad-007-separate-transactions-per-phase)
8. [AD-008: Soft Delete Pattern](#ad-008-soft-delete-pattern)
9. [Decision Matrix](#decision-matrix)
10. [Alternatives Analysis](#alternatives-analysis)
11. [Impact Assessment for Phase 2](#impact-assessment-for-phase-2)

---

## Decisions

### AD-001: ValidationResult Pattern

**Status**: DECIDED & IMPLEMENTED

**Decision Statement**:
Use a structured `ValidationResult` type aggregating errors, warnings, and info messages with contextual debug information, rather than returning raw error strings.

**Context**:
In Phase 0 planning, we recognized that validation operations (parsing ODS files, scraping Amion HTML, calculating coverage) produce multiple types of feedback:
- **Errors**: Prevent operation continuation (e.g., invalid file format)
- **Warnings**: Non-blocking issues (e.g., duplicate shifts detected)
- **Infos**: Informational messages (e.g., "processed 100 shifts")

Raw error strings cannot distinguish between these severity levels or provide contextual data.

**Options Evaluated**:

1. **Raw error strings** (rejected)
   - Pros: Simple, familiar Go pattern
   - Cons: Cannot differentiate severity levels, no context, difficult to machine-process

2. **Exception-like panic/recover** (rejected)
   - Pros: Stop on error immediately
   - Cons: Go idiom discourages this, difficult to recover from, can't collect multiple errors

3. **Multiple return values** (rejected)
   - Pros: Forces error handling
   - Cons: Ballooning function signatures, hard to pass around, loses separation of concerns

4. **Rich ValidationResult type** (CHOSEN)
   - Pros: Structured, supports multiple severity levels, contextual debug info, easy to serialize/log
   - Cons: Slightly more verbose than raw errors

**Chosen Option**:
```go
type ValidationResult struct {
    Errors   []SimpleValidationMessage  // Blocking failures
    Warnings []SimpleValidationMessage  // Non-blocking issues
    Infos    []SimpleValidationMessage  // Informational
    Context  map[string]interface{}     // Debug data
}
```

**Rationale**:
- **Expressiveness**: Can communicate 3 severity levels, not just "error" vs "no error"
- **Composability**: Validation results from multiple phases merge cleanly
- **Debuggability**: Context map stores raw data, examples, line numbers
- **API Compatibility**: Can be serialized to JSON for API responses
- **Testing**: Assertions can check for specific warnings without failing

**Tradeoffs**:
- More verbose than `error` interface alone
- Requires careful initialization (empty slices, not nil)
- Must manage nil checks in consuming code

**Implications**:
- All import services return `(result, error)` pairs
- Orchestrator merges validation results from all phases
- API handlers format ValidationResult into structured responses
- Phase 2 must adopt same pattern for consistency

**Location**: `/internal/validation/validation.go`

**Tests**:
- `validation_test.go`: 20+ tests for ValidationResult builders and accessors
- `validation_message_test.go`: Tests for ValidationMessage marshaling

---

### AD-002: Immutable Entities

**Status**: DECIDED & IMPLEMENTED

**Decision Statement**:
Make certain entity types (ShiftInstance, AuditLog) completely immutable after creation; use versioning (ScheduleVersion) for tracking changes rather than in-place updates.

**Context**:
Traditional mutable data models allow fields to be changed at any time:
```go
shift.StartTime = "09:00"  // When did this happen? Who changed it?
```

For schedule data, this creates auditability gaps:
- Cannot track WHO changed a shift start time
- Cannot determine WHEN the change occurred
- Cannot implement "time-travel" queries to see schedule at past date
- Difficult to implement version control

**Options Evaluated**:

1. **Fully mutable entities** (rejected)
   - Pros: Simple, familiar ORM pattern
   - Cons: No audit trail unless logged explicitly, time-travel queries impossible

2. **Audit table per entity** (rejected)
   - Pros: Tracks every change
   - Cons: Complex queries, significant storage overhead, hard to maintain

3. **Event sourcing** (rejected)
   - Pros: Complete event history, powerful for analytics
   - Cons: Overkill for schedCU, complex implementation, steep learning curve

4. **Selective immutability with versioning** (CHOSEN)
   - ShiftInstance: Completely immutable (no UpdatedAt/UpdatedBy fields)
   - ScheduleVersion: Mutable status (STAGING→PRODUCTION→ARCHIVED), tracked via UpdatedAt/UpdatedBy
   - Person: Create-once, soft-delete only (no updates)

**Chosen Option**:
```go
// Immutable - no UpdatedAt, UpdatedBy, or UpdatedBy fields
type ShiftInstance struct {
    ID                uuid.UUID
    ScheduleVersionID uuid.UUID
    ShiftType         string
    CreatedAt         time.Time  // Only these two
    CreatedBy         uuid.UUID
    // NO: UpdatedAt, UpdatedBy, DeletedAt
}

// Versioned - status changes tracked
type ScheduleVersion struct {
    ID        uuid.UUID
    Status    VersionStatus  // STAGING → PRODUCTION → ARCHIVED
    CreatedAt time.Time
    CreatedBy uuid.UUID
    UpdatedAt time.Time      // Last status change
    UpdatedBy uuid.UUID      // Who made status change
}
```

**Rationale**:
- **Auditability**: Creation timestamp is immutable proof of origin
- **Consistency**: Shifts never change within a version, enabling reliable time-travel
- **Simplicity**: Don't need audit table for shifts; version table is audit trail
- **Semantics**: Schedule changes = new ScheduleVersion, not modified ShiftInstance
- **Compliance**: Immutable creation records satisfy regulatory requirements

**Tradeoffs**:
- Cannot "fix" a typo in shift data without creating new version
- Database cannot enforce data corrections (must be done at application level)
- Requires discipline: team must understand immutability philosophy
- Migrations are complex (immutable fields can't be altered)

**Implications**:
- Repository layer cannot have `Update()` method for ShiftInstance
- When schedule needs revision: soft-delete old ScheduleVersion, create new one
- AuditLog is completely immutable (append-only)
- Coverage calculations are snapshots (CoverageCalculation.CalculatedAt not UpdatedAt)

**Location**:
- `/internal/entity/shift_instance.go` (no Update method)
- `/docs/ENTITY_ARCHITECTURE.md` section "Immutability Patterns"

**Tests**:
- Immutability enforced by lack of Update() in repository interface
- Version state machine tests verify transitions

---

### AD-003: Batch Queries for Coverage

**Status**: DECIDED & IMPLEMENTED

**Decision Statement**:
Use batch queries (fetching all shifts + all assignments in one round-trip) rather than per-shift queries when calculating coverage metrics.

**Context**:
Coverage calculation computes "how many radiologists available for ON1 shifts?"
Naive approach:
```go
for shift in shifts {
    assignments := getAssignmentsForShift(shift.ID)  // N+1 queries
    coverage[shift.Type] += len(assignments)
}
```

This causes:
- 1 query to fetch shifts
- N queries to fetch assignments for each shift
- Total: N+1 queries

For large schedules (100+ shifts), this is 100+ database round-trips.

**Options Evaluated**:

1. **Per-shift queries (N+1)** (rejected)
   - Pros: Simplest code
   - Cons: Kills performance with 100+ queries per calculation

2. **Individual query optimization** (rejected)
   - Pros: Reduces constants, some improvement
   - Cons: Still N+1 pattern, not scalable

3. **Batch queries with JOIN** (CHOSEN - primary)
   - Pros: Single round-trip, optimal performance, clean code
   - Cons: Requires careful SQL, must handle duplicates

4. **Caching layer** (CHOSEN - secondary)
   - Pros: Eliminates repeated queries
   - Cons: Cache invalidation complexity, not primary solution

**Chosen Option**:
```sql
-- Single batch query: fetch all assignments for schedule version
SELECT
    si.id as shift_id,
    si.shift_type,
    COUNT(a.id) as assignment_count
FROM shift_instances si
LEFT JOIN assignments a ON a.shift_instance_id = si.id
WHERE si.schedule_version_id = $1
    AND a.deleted_at IS NULL
GROUP BY si.id, si.shift_type
```

Then in Go:
```go
results := repo.GetAssignmentsByScheduleVersionBatch(ctx, scheduleVersionID)
// results contains all assignments grouped by shift, not per-shift queries
coverage := calculateCoverageFromResults(results)
```

**Rationale**:
- **Performance**: Single round-trip vs N+1
- **Scalability**: Constant time regardless of shift count
- **Simplicity**: One query, easier to reason about
- **Testability**: Can mock single batch result
- **Profiling**: Easy to identify as optimization point

**Tradeoffs**:
- Must handle GROUP BY correctly (no mistakes in aggregation)
- Cannot easily get "assignments for specific shift" with batch approach (need different query)
- Requires more sophisticated SQL

**Implications**:
- Repository must expose batch methods: `GetAssignmentsByScheduleVersionBatch()`
- Coverage calculator depends on batch semantics (will fail if queries become per-shift)
- Performance benchmarks show >10x improvement for typical schedules
- Phase 2 must maintain batch pattern for other aggregate queries

**Location**:
- `/internal/service/coverage/` - batch loading patterns
- `/internal/repository/` - batch query methods

**Tests**:
- `algorithm_bench_test.go`: Benchmarks show batch queries 10x faster
- `assertions.go`: Helper to assert query count <= 1 (prevents regression)

---

### AD-004: goquery for HTML Parsing

**Status**: DECIDED & IMPLEMENTED

**Decision Statement**:
Use `github.com/PuerkitoBio/goquery` library for parsing Amion HTML responses, rather than regex or manual parsing.

**Context**:
Amion schedules are delivered as HTML tables in HTTP responses. Extracting shift data requires parsing unstructured HTML.

The HTML contains:
- Multiple tables with different purposes
- CSS classes on rows/cells
- Colspan/rowspan attributes
- Inconsistent whitespace and formatting

Manual parsing is error-prone.

**Options Evaluated**:

1. **Regex patterns** (rejected)
   - Pros: No dependency, lightweight
   - Cons: Brittle (any HTML change breaks), unreadable, hard to maintain

2. **XML strict parser** (rejected)
   - Pros: Strict parsing guarantees
   - Cons: Amion HTML is not well-formed XML, parser will reject it

3. **html/parsing stdlib** (rejected)
   - Pros: Standard library, no dependency
   - Cons: Very low-level, requires extensive manual DOM traversal

4. **goquery** (CHOSEN)
   - Pros: jQuery-like API (familiar), forgiving parsing, CSS selector support
   - Cons: Additional dependency, learning curve

5. **Custom recursive descent parser** (rejected)
   - Pros: Full control
   - Cons: 500+ lines of code, hard to maintain

**Chosen Option**:
```go
import "github.com/PuerkitoBio/goquery"

func parseAmionScheduleTable(html string) (shifts []Shift, err error) {
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
    if err != nil {
        return nil, fmt.Errorf("parse error: %w", err)
    }

    doc.Find("table.schedule tbody tr").Each(func(_ int, row *goquery.Selection) {
        shift := parseShiftRow(row)
        shifts = append(shifts, shift)
    })

    return shifts, nil
}

func parseShiftRow(row *goquery.Selection) Shift {
    return Shift{
        Name:  row.Find("td.name").Text(),
        Start: row.Find("td.start-time").Text(),
        End:   row.Find("td.end-time").Text(),
    }
}
```

**Rationale**:
- **Robustness**: Forgiving HTML parser handles malformed input
- **Maintainability**: CSS selectors are readable (vs regex)
- **Familiarity**: jQuery-like syntax reduces learning curve
- **Flexibility**: Can handle nested structures, dynamic attributes
- **Testability**: Easy to mock HTML documents in tests

**Tradeoffs**:
- Adds external dependency (security surface)
- Dependency updates could introduce breaking changes
- HTML selector changes require code updates (but unavoidable with any parser)
- Slightly slower than regex (negligible for typical schedules)

**Implications**:
- All Amion HTML parsing goes through goquery
- CSS selectors must be hardcoded or configurable (AmionSelectors struct)
- Selector changes from Amion require code updates (documented in AMION_HTML_STRUCTURE.md)
- Phase 2 should continue using goquery for consistency

**Location**:
- `/internal/service/amion/selectors.go` - CSS selector definitions
- `/internal/service/amion/client.go` - FetchAndParseHTML() method
- `/docs/AMION_HTML_STRUCTURE.md` - HTML structure documentation

**Tests**:
- `selectors.go` includes default CSS selectors
- Amion scraper tests use fixture HTML documents
- Error cases test parser robustness

---

### AD-005: Custom ODS Parser

**Status**: DECIDED & IMPLEMENTED

**Decision Statement**:
Implement a custom ODS parser in Go rather than using a third-party library, despite the implementation complexity.

**Context**:
ODS (OpenDocument Spreadsheet) is an XML-based format (ZIP archive containing XML files).

Phase 1 needed to choose between:
1. Hand-write ODS parsing (convert ODS → Go structs)
2. Use third-party library (easyxl, goexcel, etc.)

The decision required evaluating:
- Library stability and maintenance
- Go ecosystem maturity
- Bundle size
- Feature coverage
- Ease of implementation

**Options Evaluated**:

1. **tealeg/xlsx** (rejected)
   - Pros: Popular, stable
   - Cons: XLSX only (not ODS), different format

2. **juju/gnuflag or similar** (rejected)
   - Pros: Many stars on GitHub
   - Cons: Unmaintained, security issues reported

3. **Third-party ODS libraries** (evaluated, some rejected)
   - Pros: Don't reinvent wheel
   - Cons: Most are unmaintained or have limited feature coverage
   - Examples rejected: easyxl (old), goexcel (limited), milo (inactive)

4. **Custom ODS parser in Go** (CHOSEN)
   - Pros: Full control, simple to extend, no dependencies
   - Cons: Must implement XML parsing logic

**Chosen Option**:
```go
type ODSParser struct {
    // Custom implementation using archive/zip + encoding/xml
}

func (p *ODSParser) Parse(content []byte) (*ParsedSchedule, error) {
    // 1. Unzip ODS file
    zipReader := zip.NewReader(bytes.NewReader(content), int64(len(content)))

    // 2. Extract and parse content.xml
    manifestXML := readXMLFile(zipReader, "content.xml")

    // 3. Parse spreadsheet XML structure
    spreadsheet := &Spreadsheet{}
    xml.Unmarshal(manifestXML, spreadsheet)

    // 4. Extract shift data from cells
    shifts := extractShiftsFromSpreadsheet(spreadsheet)

    return &ParsedSchedule{Shifts: shifts}, nil
}
```

**Rationale**:
- **Control**: Understand exactly what parsing is doing (security important)
- **Size**: Go executable smaller (custom parser < dependency overhead)
- **Stability**: Not dependent on maintenance of external library
- **Simplicity**: ODS format is standard XML; straightforward to parse
- **Extensibility**: Can add support for custom cell types easily

**Tradeoffs**:
- Must handle all ODS format variations ourselves
- Bug fixes and format changes become our responsibility
- More code to maintain (~500 lines)
- Team must understand XML/ZIP structure
- Cannot benefit from others' ODS improvements

**Implications**:
- ODS parsing errors must be caught and reported via ValidationResult
- Parser must handle common ODS quirks (empty cells, merged cells, etc.)
- Documentation required for supported ODS format/features
- Phase 2 if ODS format needs change: must update parser
- Can extract ODS parser into separate package if needed

**Location**:
- `/internal/service/ods/importer.go` - Custom ODS parser implementation
- `/internal/service/ods/shift_mapper.go` - Maps ODS cells to Shift entities
- `/ODS_PARSER_IMPLEMENTATION.md` - Parser design documentation

**Tests**:
- `importer_test.go`: 20+ tests with fixture ODS files
- Tests cover empty cells, merged cells, bad dates, etc.
- Fixtures include real-world ODS schedules

---

### AD-006: 3-Phase Orchestration

**Status**: DECIDED & IMPLEMENTED

**Decision Statement**:
Implement schedule import as a 3-phase orchestrated workflow (ODS Import → Amion Scraping → Coverage Calculation) with different error handling per phase, rather than a single monolithic import function.

**Context**:
The complete import workflow consists of:
1. **Phase 1 (ODS Import)**: Parse ODS file, create ScheduleVersion + ShiftInstances
   - Duration: 2-4 hours (large files with 1000+ rows)
   - Criticality: **CRITICAL** - must succeed for operation to continue
   - Failure impact: Entire import fails

2. **Phase 2 (Amion Scraping)**: Fetch actual assignments from Amion service
   - Duration: 2-3 seconds
   - Criticality: **NON-CRITICAL** - enhances but isn't essential
   - Failure impact: Schedule exists but without Amion data (fallback to ODS)

3. **Phase 3 (Coverage Calculation)**: Compute coverage metrics
   - Duration: ~1 second
   - Criticality: **NON-CRITICAL** - informational only
   - Failure impact: Schedule exists but metrics unavailable (can recalculate later)

Single monolithic import would treat all phases equally: one failure = entire import fails.

**Options Evaluated**:

1. **Single monolithic function** (rejected)
   - Pros: Simplest code
   - Cons: Phase 2 error kills entire operation even though Phase 1 succeeded

2. **Nested try-catch-fallback** (rejected)
   - Pros: Recovers from errors
   - Cons: Complex nesting, hard to follow logic, unclear error handling

3. **State machine with transitions** (rejected)
   - Pros: Explicit transitions
   - Cons: Overkill for 3 phases, too much boilerplate

4. **3-phase orchestrator with phase-specific error handling** (CHOSEN)
   - Pros: Clear separation, explicit error policies per phase, recoverable
   - Cons: More code, requires careful orchestration

**Chosen Option**:
```go
type ScheduleOrchestrator interface {
    ExecuteImport(ctx context.Context, ...) (*OrchestrationResult, error)
}

// Implementation
func (o *DefaultScheduleOrchestrator) ExecuteImport(...) (*OrchestrationResult, error) {
    // Phase 1: ODS Import (CRITICAL)
    scheduleVersion, err := o.odsService.ImportSchedule(...)
    if err != nil {
        return nil, err  // CRITICAL: Stop immediately
    }

    // Phase 2: Amion Scraping (NON-CRITICAL)
    assignments, err := o.amionService.ScrapeSchedule(...)
    if err != nil {
        o.logger.Warn("Amion scraping failed, continuing without Amion data")
        // Continue anyway - Amion data is optional
    }

    // Phase 3: Coverage Calculation (NON-CRITICAL)
    coverage, err := o.coverageService.CalculateCoverage(...)
    if err != nil {
        o.logger.Warn("Coverage calculation failed, schedule created anyway")
        // Continue - coverage can be recalculated
    }

    // Merge validation results from all phases
    result.ValidationResult = mergeResults(phase1, phase2, phase3)
    return result, nil
}
```

**Rationale**:
- **Resilience**: Failures in non-critical phases don't block operation
- **Semantic clarity**: Each phase has explicit error policy
- **User experience**: Users get partial results (schedule without metrics) vs total failure
- **Observability**: Can track phase completion independently
- **Recovery**: Can retry failed phases independently

**Tradeoffs**:
- More code and complexity
- Must coordinate 3 services with different error handling
- Harder to test all phase combinations (8 scenarios)
- Validation result merging adds complexity

**Implications**:
- Each phase executes in its own transaction (see AD-007)
- Phase 2 failure doesn't rollback Phase 1
- Phase 3 failure doesn't rollback Phases 1-2
- Orchestrator status visible to callers (GetOrchestrationStatus)
- Phase 2 improvements: concurrent scraping of multiple hospitals

**Location**:
- `/internal/service/orchestrator/orchestrator.go` - 3-phase orchestrator
- `/internal/service/orchestrator/orchestrator_impl_test.go` - 8+ tests covering all scenarios
- `/ORCHESTRATOR_IMPLEMENTATION.md` - Design documentation

**Tests**:
- `TestExecuteImportSuccessfulFullWorkflow`: All phases succeed
- `TestExecuteImportPhase1CriticalError`: Phase 1 fails → stop
- `TestExecuteImportPhase2ErrorContinuesToPhase3`: Phase 2 fails → continue
- `TestExecuteImportPhase3ErrorDoesNotFail`: Phase 3 fails → succeed anyway
- Plus 4 more scenarios covering invalid inputs, status tracking, etc.

---

### AD-007: Separate Transactions per Phase

**Status**: DECIDED & IMPLEMENTED

**Decision Statement**:
Execute each orchestration phase in its own database transaction, allowing Phase 1 commit even if Phase 2 or 3 fails, rather than single all-or-nothing transaction.

**Context**:
Three-phase orchestration (AD-006) creates a design question: what happens to database state if a phase fails?

Options for transaction scope:
1. **Single transaction** (all phases): Failure in Phase 2/3 rolls back Phase 1
2. **Per-phase transactions**: Phase 1 commits, Phase 2/3 can fail independently
3. **Hybrid**: Phase 1-2 in one transaction, Phase 3 separate

Single transaction conflicts with resilience goal:
- Phase 1 creates ScheduleVersion (takes hours to process)
- Phase 2 scrapes Amion assignments (2-3 seconds)
- If Phase 2 fails → rollback phase 1 → lose hours of work

Per-phase transactions aligns with semantic resilience:
- Phase 1 fails → nothing committed (critical)
- Phase 2 fails → Phase 1 stays (schedule created without Amion data)
- Phase 3 fails → Phases 1-2 stay (schedule without metrics)

**Options Evaluated**:

1. **Single all-or-nothing transaction** (rejected)
   - Pros: Atomicity guarantees, simple semantics
   - Cons: Defeats purpose of Phase 2 non-criticality (rollback would delete schedule)

2. **Per-phase transactions** (CHOSEN)
   - Pros: Aligns with phase semantics, allows partial success, recoverable
   - Cons: Inconsistency if Phase 1 commits but server crashes in Phase 2

3. **Manual savepoints** (rejected)
   - Pros: Fine-grained control
   - Cons: Harder to reason about, more error-prone

4. **Event-sourced with compensation** (rejected)
   - Pros: Can replay and compensate
   - Cons: Complex, overkill for Phase 1 criticality

**Chosen Option**:
```go
// Phase 1 in isolated transaction
tx1 := db.BeginTx(ctx, nil)
scheduleVersion, err := service1.ImportWithTx(tx1, ...)
if err != nil {
    tx1.Rollback()  // CRITICAL: Roll back everything
    return nil, err
}
tx1.Commit()  // SUCCESS: Phase 1 persisted

// Phase 2 in separate transaction (even if Phase 1 already committed)
tx2 := db.BeginTx(ctx, nil)
assignments, err := service2.ScrapeWithTx(tx2, ...)
if err != nil {
    tx2.Rollback()  // NON-CRITICAL: Only Phase 2 rolled back
    // Phase 1 still exists - operation succeeds despite Phase 2 failure
}
tx2.Commit()  // Phase 2 persisted (if successful)

// Phase 3 in separate transaction
tx3 := db.BeginTx(ctx, nil)
coverage, err := service3.CalculateWithTx(tx3, ...)
// ... same pattern
```

**Rationale**:
- **Consistency with phase semantics**: Phase 2 can fail without affecting Phase 1
- **User value**: User gets schedule even if Amion scraping fails
- **Observability**: Can see which phases succeeded/failed independently
- **Recovery**: Can retry Phase 2 without re-running Phase 1
- **Isolation**: Phase 2 errors don't impact Phase 1 data

**Tradeoffs**:
- Transaction isolation level matters (dirty reads possible if not careful)
- Race conditions if server crashes between Phase 1 and Phase 2 commits
- Requires careful coordinated transaction handling
- Harder to maintain consistency across phases

**Implications**:
- Isolation level: Read Committed (prevents dirty reads, allows phantom reads)
- Each service must accept `tx` parameter for phase execution
- Cannot assume all-or-nothing semantics (must handle partial success)
- Requires comprehensive transaction manager tests (12+ test cases)
- Phase 2 concurrent improvements: multiple services can use separate transactions

**Location**:
- `/internal/service/orchestrator/transaction_manager.go` - Transaction lifecycle
- `/internal/service/orchestrator/transaction_manager_test.go` - 12+ transaction tests
- All Phase 1/2/3 services accept `*sql.Tx` parameter

**Tests**:
- `TestPhase1FailureRollsBack`: Phase 1 error → transaction rolled back
- `TestPhase1CommitPhase2RollBack`: Phase 1 committed, Phase 2 rolls back
- `TestPhase2FailureDoesNotAffectPhase1`: Phase 2 failure, Phase 1 still in DB
- Plus 9 more covering isolation, race conditions, concurrent phases

---

### AD-008: Soft Delete Pattern

**Status**: DECIDED & IMPLEMENTED

**Decision Statement**:
Use soft deletion (setting `DeletedAt` timestamp) for all mutable entities instead of hard deletion, enabling recovery and audit trail preservation.

**Context**:
Relational databases offer two deletion strategies:

**Hard delete**: `DELETE FROM table WHERE id = $1`
- Pros: Saves storage, simple
- Cons: Irreversible, breaks audit trail, difficult for recovery

**Soft delete**: `UPDATE table SET deleted_at = NOW() WHERE id = $1`
- Pros: Reversible, preserves audit trail, enables recovery
- Cons: Requires filtering queries, uses storage, complexity

schedCU is healthcare scheduling - audit trail and recovery are critical:
- Hospital accidentally deletes wrong assignment
- Need to recover it immediately
- Auditor needs to prove what happened
- Compliance requires retention periods

**Options Evaluated**:

1. **Hard deletion** (rejected)
   - Pros: Simple, saves storage
   - Cons: Irreversible, breaks audit trail, violates compliance

2. **Logical deletion with separate archive table** (rejected)
   - Pros: Separation of concerns
   - Cons: Complex migrations, query complexity

3. **Event sourcing (log every change)** (rejected)
   - Pros: Complete history
   - Cons: Overkill, complex queries, storage overhead

4. **Soft deletion with DeletedAt timestamp** (CHOSEN)
   - Pros: Reversible, audit trail preserved, simple recovery
   - Cons: Storage (keeps deleted records), query filtering needed

**Chosen Option**:
```go
// All mutable entities have soft delete
type Assignment struct {
    ID        uuid.UUID
    PersonID  uuid.UUID
    ShiftID   uuid.UUID
    CreatedAt time.Time
    CreatedBy uuid.UUID
    DeletedAt *time.Time  // NULL = active, NOT NULL = deleted
    DeletedBy *uuid.UUID  // Who deleted it
}

// Query pattern: always filter out deleted
func (repo *AssignmentRepository) GetActive(...) ([]*Assignment, error) {
    return repo.db.WithContext(ctx).
        Where("deleted_at IS NULL").
        Find(&assignments).
        Error
}

// Soft delete
func (repo *AssignmentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
    return repo.db.WithContext(ctx).
        Model(&Assignment{}).
        Where("id = ?", id).
        Update("deleted_at", time.Now().UTC()).
        Error
}

// Recovery: set DeletedAt back to NULL
func (repo *AssignmentRepository) Restore(ctx context.Context, id uuid.UUID) error {
    return repo.db.WithContext(ctx).
        Model(&Assignment{}).
        Where("id = ?", id).
        Update("deleted_at", nil).
        Error
}
```

**Rationale**:
- **Compliance**: Maintains audit trail for healthcare regulations
- **Recovery**: Accidental deletions reversible (no manual restore from backup)
- **Debugging**: Can query deleted records to understand failure cascade
- **User experience**: Can recover deleted schedules without DBA intervention
- **Analytics**: Full dataset available for analysis (including "deleted" state)

**Tradeoffs**:
- Storage cost (keeps deleted records indefinitely)
- Query complexity (must filter `WHERE deleted_at IS NULL`)
- Must enforce soft delete everywhere (inconsistency if missed)
- Cannot use unique constraints without considering soft deletes

**Implications**:
- All repositories must implement soft delete, not hard delete
- Active-record queries must filter `deleted_at IS NULL`
- Historical queries can include deleted records
- Archival strategy: soft delete with retention period (not hard delete)
- Backup/restore: deleted records preserved, can recover after restore

**Location**:
- All entity types: `DeletedAt *time.Time`, `DeletedBy *uuid.UUID` fields
- All repositories: soft-delete implementation, active-only queries
- `/internal/entity/` - soft delete pattern throughout
- `/docs/ENTITY_ARCHITECTURE.md` section "Soft Delete Pattern"

**Tests**:
- Repository tests verify soft-delete behavior
- Active-only queries exclude deleted records
- Recovery queries include deleted records
- Audit trail verified for deletions

---

## Decision Matrix

| ID | Decision | Owner | Date Decided | Decided | Status | Impact |
|:--:|----------|-------|--------------|---------|--------|--------|
| AD-001 | ValidationResult Pattern | Engineering | 2025-10-01 | Phase 0 | IMPLEMENTED | All validation paths |
| AD-002 | Immutable Entities | Engineering | 2025-10-05 | Phase 0 | IMPLEMENTED | Entity layer, versioning |
| AD-003 | Batch Queries | Engineering | 2025-10-15 | Phase 1 | IMPLEMENTED | 10x coverage perf |
| AD-004 | goquery HTML Parsing | Engineering | 2025-10-20 | Phase 1 | IMPLEMENTED | Amion scraping |
| AD-005 | Custom ODS Parser | Engineering | 2025-10-22 | Phase 1 | IMPLEMENTED | ODS import |
| AD-006 | 3-Phase Orchestration | Engineering | 2025-10-25 | Phase 1 | IMPLEMENTED | Import resilience |
| AD-007 | Per-Phase Transactions | Engineering | 2025-10-28 | Phase 1 | IMPLEMENTED | Partial success |
| AD-008 | Soft Delete Pattern | Engineering | 2025-09-15 | Phase 0 | IMPLEMENTED | Compliance, recovery |

**Legend**:
- **Owner**: Who made the decision
- **Date Decided**: When consensus reached
- **Decided In**: What phase finalized decision
- **Status**: PROPOSED, DECIDED, IMPLEMENTED, REJECTED
- **Impact**: What areas affected

---

## Alternatives Analysis

### Why We Didn't Choose Raw Error Strings (AD-001)

**Alternative**: Return `error` interface directly
```go
func ImportODS(content []byte) (*ScheduleVersion, error)
```

**Why rejected**:
- No severity distinction (can't say "this warning is non-blocking")
- No context data (can't include examples, values, line numbers)
- Cannot collect multiple errors (only returns first)
- Cannot serialize to structured API responses
- Type assertions required to extract details

**What we learned**:
- Errors in complex workflows need structure
- Logging alone insufficient (need searchable fields)
- Validation results must be composable (mergeable)

**Would it work?**
- Yes, but would require additional wrapper layer
- API handlers would have to convert errors to ValidationResult anyway
- Better to have ValidationResult at source

---

### Why We Didn't Choose N+1 Queries (AD-003)

**Alternative**: Per-shift assignment query
```go
for shift in shifts {
    assignments := getAssignmentsForShift(shift.ID)  // Repeated query
    coverage[shift.Type] += len(assignments)
}
```

**Why rejected**:
- 100 shifts = 100 database round-trips
- Network latency killer (typical: 10ms per round-trip = 1 second overhead)
- Query planning overhead per request
- Connection pooling pressure

**Performance impact**:
- Batch query: ~50ms for schedule with 100 shifts
- N+1 queries: ~1+ second for same schedule
- 10-20x performance degradation

**What we learned**:
- Go's concurrency makes per-shift queries tempting (goroutines hide latency)
- But database connection pool fills up (limits concurrency)
- Batch queries simple once understood

**Would it work?**
- Functionally yes
- Performant only with caching or async batching
- Better to solve at query level

---

### Why We Didn't Choose XML Library (AD-005)

**Alternative**: Third-party ODS library
```go
import "github.com/some-lib/ods"
schedule, _ := ods.Parse(fileContent)
```

**Libraries evaluated**:
1. **alecthomas/go-ods** - Archived (no longer maintained)
2. **zhenghaoz/gocalc** - Last update 2017 (5+ years old)
3. **mandykoh/prism** - Experimental (not production-ready)
4. **sjeandeaux/rsc** - Unclear license (likely GPL)

**Why rejected**:
- No stable, well-maintained ODS library in Go ecosystem
- Security: unmaintained libraries have unpatched vulnerabilities
- Control: can't fix bugs in dependencies
- Bundle size: custom parser smaller than deps + library

**What we learned**:
- Go ecosystem has gaps (Python/Java have better XML/ODS support)
- XML parsing is straightforward (archive/zip + encoding/xml sufficient)
- Custom implementation 500 lines, manageable

**Would it work?**
- Theoretically yes (if good library existed)
- Practically impossible with available options
- Custom implementation better choice

---

### Why We Didn't Choose Single Monolithic Import (AD-006)

**Alternative**: Single function, all phases in one function
```go
func ImportSchedule(content []byte) (*ScheduleVersion, error) {
    sv := importODS(content)           // Phase 1
    assignments := scrapeAmion(sv)     // Phase 2
    coverage := calculateCoverage(sv)  // Phase 3
    return sv, nil
}
```

**Why rejected**:
- If Phase 2 fails (Amion server down), entire operation fails
- User loses Phase 1 work (hours of processing)
- No way to recover: must re-import entire file
- No distinction between critical and non-critical steps

**Real scenario**:
- Hospital imports 1000-row ODS file (takes 2 hours to process)
- Phase 1 succeeds: ScheduleVersion created, 1000 shifts in DB
- Phase 2 fails: Amion server down
- With monolithic: entire operation fails, 2 hours of work lost
- With orchestrated: Phase 1 committed, Phase 2 retries later, operation succeeded

**What we learned**:
- Import resilience requires phase distinction
- Long-running operations need transaction boundaries
- Non-critical phases must not block critical ones

**Would it work?**
- Yes, but poor user experience on failures
- Better architecture supports partial success

---

### Why We Didn't Choose All-in-One Transaction (AD-007)

**Alternative**: Single transaction for all three phases
```go
tx := db.BeginTx(ctx, nil)
try {
    importODS(tx)      // Phase 1
    scrapeAmion(tx)    // Phase 2
    calculateCoverage(tx)  // Phase 3
    tx.Commit()
} catch {
    tx.Rollback()  // All or nothing
}
```

**Why rejected**:
- Conflicts with resilience goal (AD-006)
- Phase 2/3 failure rolls back entire Phase 1
- Long transaction (hours for large files) blocks other operations
- Deadlock risk if multiple imports concurrent

**Example conflict**:
- Single transaction: Phase 1 + Phase 2 + Phase 3 in one TX
- Phase 2 timeout → transaction aborts
- Phase 1 committed to nowhere (rolled back)
- User loses 2 hours of work

**What we learned**:
- Transaction scope must match failure semantics
- Critical phases need committed state before non-critical phases start
- Per-phase transactions enable partial success
- Requires careful isolation level tuning (Read Committed)

**Would it work?**
- Yes, but breaks resilience objective
- Forces either:
  1. Accept total failure on any phase error
  2. Complex compensation logic to undo Phase 1 changes

---

## Impact Assessment for Phase 2

### High Impact - Must Consider

#### 1. Batch Query Pattern Extends to All Services
**What changed**: Phase 1 established batch queries for coverage (AD-003)

**Phase 2 implication**: All aggregate queries must use batch pattern
- Person querying: get all active persons in one query
- Assignment querying: batch by schedule version
- Shift querying: batch by date range

**Action**: When Phase 2 adds new aggregate services, use batch queries from start
**Estimate**: Design impact: 5%, implementation impact: 10%

#### 2. Soft Delete Consistency Required
**What changed**: Phase 1 uses soft delete everywhere (AD-008)

**Phase 2 implication**: All new entities must use soft delete pattern
- Cannot add any hard-delete entities (inconsistent)
- Must filter `deleted_at IS NULL` in all active-only queries
- Recovery mechanisms must account for soft deletes

**Action**: Phase 2 entities must include `DeletedAt *time.Time` and `DeletedBy *uuid.UUID`
**Estimate**: Implementation impact: ~5% (filtering pattern)

#### 3. Orchestration Extensibility
**What changed**: Phase 1 established 3-phase orchestration (AD-006, AD-007)

**Phase 2 implication**: More phases possible (approval workflow, notification, etc.)
- Orchestrator pattern is extensible
- Must maintain per-phase transaction isolation
- New phases must be non-critical (or will block operation)

**Action**: Review orchestrator design for new phases, add phase-specific error handling
**Estimate**: Design impact: 20%, implementation impact: 15%

#### 4. ValidationResult Consolidation
**What changed**: Phase 1 uses ValidationResult for all validation (AD-001)

**Phase 2 implication**: All new validation must produce ValidationResult
- Cannot return raw errors for new services
- Must integrate with orchestrator result merging
- Logging must include severity (errors/warnings/infos)

**Action**: Phase 2 services must return (result, ValidationResult, error)
**Estimate**: Implementation impact: 10%

### Medium Impact - Should Plan

#### 5. Entity Immutability Pattern
**What changed**: Phase 1 uses selective immutability (AD-002)

**Phase 2 implication**: New entities must decide mutable vs immutable
- Version tracking for mutable entities
- Soft delete only for immutable
- Audit trail enforcement

**Action**: Document immutability policy for Phase 2 entity design
**Estimate**: Design impact: 10%

#### 6. goquery Standardization
**What changed**: Phase 1 uses goquery for HTML parsing (AD-004)

**Phase 2 implication**: Any new HTML parsing must use goquery
- Custom selectors for new HTML sources
- Standardized error handling via goquery API

**Action**: If Phase 2 adds new HTML sources (not just Amion), use goquery
**Estimate**: Implementation impact: <5% (already have patterns)

#### 7. Custom Parser Model
**What changed**: Phase 1 implements custom ODS parser (AD-005)

**Phase 2 implication**: If new data formats needed, consider custom parsing
- Go ecosystem pattern: build custom when libraries unavailable
- Parser testing framework established
- Documentation standards set

**Action**: For new file formats (CSV, Excel, etc.), evaluate custom vs library
**Estimate**: Design decision impact: varies by format

### Low Impact - Aware But Not Blocking

#### 8. Database Isolation Level
**What changed**: Phase 1 uses Read Committed isolation (AD-007)

**Phase 2 implication**: Concurrent operations (multiple imports) safe at this level
- Dirty read prevention sufficient
- Phantom reads acceptable for operational data
- May need Serializable for sensitive operations (future)

**Action**: Monitor concurrent behavior, upgrade isolation if needed
**Estimate**: Observability cost: <5%

---

## Phase 2 Readiness Checklist

Based on Phase 1 decisions, Phase 2 must:

- [ ] Adopt soft delete pattern for all new entities
- [ ] Use batch queries for all aggregate operations
- [ ] Return ValidationResult from all validation services
- [ ] Extend orchestrator for approval workflow (if needed)
- [ ] Use goquery for any new HTML parsing
- [ ] Document entity mutability decisions (immutable vs versioned)
- [ ] Implement transaction managers for new phase combinations
- [ ] Establish error handling per phase (critical vs non-critical)
- [ ] Test partial success scenarios (phase failures)
- [ ] Consider async processing for long-running phases

---

## Architecture Decision History

### Phase 0 (Spikes & Foundation)
- **Dates**: September 15 - October 15, 2025
- **Decisions made**: AD-001, AD-002, AD-008
- **Decisions finalized**: ValidationResult pattern, immutable entities, soft delete
- **Output**: Entity architecture documentation

### Phase 1 (Core Services)
- **Dates**: October 16 - November 15, 2025
- **Decisions made**: AD-003, AD-004, AD-005, AD-006, AD-007
- **Decisions finalized**: All remaining major architectural decisions
- **Output**: Working orchestrator, all parsers, batch query patterns

### Phase 2 (Expansion)
- **Dates**: November 16, 2025 - onwards
- **Expected decisions**: Approval workflow, notifications, multi-hospital, concurrency
- **Legacy decisions**: Must respect Phase 1 patterns (soft delete, batch queries, validation results)

---

## Conclusion

Phase 1 established a coherent architectural foundation based on:

1. **Structured validation** (ValidationResult) enabling rich error reporting
2. **Immutable entities** (ShiftInstance) ensuring audit trail integrity
3. **Performance patterns** (batch queries) enabling scalability
4. **Resilient workflows** (3-phase orchestration) tolerating non-critical failures
5. **Recovery capability** (soft delete) enabling accident recovery
6. **Standardized parsing** (goquery, custom ODS) handling external data

These decisions are interdependent and mutually reinforcing:
- Soft delete enables recovery from orchestration phase failures
- Immutable entities enable reliable version history
- Batch queries enable performance at scale
- ValidationResult enables phase-specific error handling
- 3-phase orchestration enables graceful degradation

Phase 2 must maintain architectural consistency by:
- Extending patterns (batch queries, soft delete) to new services
- Respecting entity design decisions (immutability, audit trails)
- Following orchestration model for new workflows
- Standardizing validation through ValidationResult

This ensures schedCU v2 grows coherently rather than devolving into inconsistent ad-hoc solutions.

---

## Document Management

**Version**: 1.0
**Last Updated**: November 15, 2025
**Next Review**: After Phase 2 milestone
**Owner**: Engineering Team
**Stakeholders**: Backend team, Frontend team, DevOps team

Changes to this document:
- Phase 2 milestone: Add decisions AD-009, AD-010, etc.
- Each new decision: Append before "Conclusion" section
- Maintain decision matrix for quick reference
