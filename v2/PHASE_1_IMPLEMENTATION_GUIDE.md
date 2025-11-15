# Phase 1 Implementation Guide - Core Services & Database

**Status**: Ready to implement | Foundation complete | Phase 0b >95% done
**Date**: November 15, 2025
**Focus**: Complete validation framework tests, implement core services with TDD

---

## What Phase 0b Delivered (COMPLETED ✅)

### Entity Model
- 12+ fully implemented entities (Person, ScheduleVersion, ShiftInstance, Assignment, ScrapeBatch, etc.)
- Type aliases preventing confusion (PersonID, HospitalID, Date, Time)
- Soft delete pattern on all entities
- Audit trail fields (CreatedBy, CreatedAt, UpdatedBy, UpdatedAt)
- State machines for version promotion (STAGING→PRODUCTION→ARCHIVED)
- Batch lifecycle (PENDING→COMPLETE/FAILED)

### Validation Framework
- ValidationResult struct with severity levels (ERROR, WARNING, INFO)
- Full method set: AddError, AddWarning, AddInfo, HasErrors, CanImport, CanPromote
- JSON marshaling/unmarshaling
- Error collection pattern (not fail-fast)
- 14 comprehensive tests
- **Status**: 100% feature-complete ✅

### PostgreSQL Infrastructure
- Connection management framework
- **9 PostgreSQL repositories** (3 existing + 6 new this session):
  - PersonRepository ✅
  - ScheduleVersionRepository ✅
  - AssignmentRepository ✅
  - ShiftInstanceRepository ✅ (NEW)
  - ScrapeBatchRepository ✅ (NEW)
  - CoverageCalculationRepository ✅ (NEW)
  - AuditLogRepository ✅ (NEW)
  - UserRepository ✅ (NEW)
  - JobQueueRepository ✅ (NEW)
- **10 database migrations** (all schema specified)
- JSONB support for complex data (validation results, coverage data, audit details)
- Indexes on critical query paths

### Service Layer (Partially Fixed)
- `ods_import_service.go` - ODS file parsing structure (TODO: real parsing logic)
- `amion_import_service.go` - Amion scraper structure (TODO: goquery implementation)
- `schedule_version_service.go` - Version management (TODO: complete testing)
- `schedule_orchestrator.go` - 3-phase workflow (TODO: integration testing)
- `coverage.go` - Dynamic coverage calculator structure (TODO: batch query optimization)

---

## Phase 1 Goals (4 weeks)

**Primary Goal**: Implement entity layer, validation framework, and core business logic

**Success Criteria**:
- All 9 repositories fully tested with integration tests
- ValidationResult working identically to v1 Java version
- DynamicCoverageCalculator with batch query optimization (no N+1)
- 80%+ unit test coverage on services
- Core services feature-complete for Phase 2 integration

---

## Week-by-Week Breakdown

### Week 1: Integration Tests & Repository Finalization

**Day 1-2: Testcontainers Setup**
- [x] Create `postgres_test.go` with test helper
- [ ] Test PostgreSQL container startup
- [ ] Verify all 10 migrations run correctly
- [ ] Create cleanup utilities for test isolation

**Day 3-5: Repository Integration Tests**
- [ ] PersonRepository integration tests (CRUD, GetByEmail, GetByHospital)
- [ ] ShiftInstanceRepository tests (date range queries, CountByScheduleVersion)
- [ ] ScrapeBatchRepository tests (GetByStatus, lifecycle tracking)
- [ ] CoverageCalculationRepository tests (JSONB storage/retrieval, GetLatestByScheduleVersion)
- [ ] AuditLogRepository tests (GetByUser, GetByAction, ListRecent)
- [ ] UserRepository tests (role-based queries)
- [ ] JobQueueRepository tests (job lifecycle, CleanupOldJobs)

**Target**: 85%+ coverage on all repositories, all tests passing

**Key Assertion Pattern**:
```go
// Query count should be 1 for single GetByID call
// Batch operations should be 2-3 queries max (no N+1)
// Assert exact query count to catch regressions
```

---

### Week 2: Validation Framework Tests & Enhancement

**Goal**: Ensure ValidationResult matches v1 Java semantics exactly

**Tasks**:
- [ ] Review v1 ValidationResult usage in Java codebase
- [ ] Document exact semantics (error collection, JSON format, severity levels)
- [ ] Add comprehensive tests for edge cases:
  - Empty result
  - Multiple errors with different severities
  - Mixed error/warning/info
  - Context data in each message
  - JSON serialization/deserialization
  - CanImport logic (true if only INFO/WARNING)
  - CanPromote logic (true if no ERRORS)
  - Summary generation
  - MessagesByCode, MessagesBySeverity filtering

- [ ] Integration tests with services:
  - ODS parser collecting validation errors
  - Amion scraper collecting errors
  - Coverage calculator identifying gaps

**Target**: 100% confidence that ValidationResult is feature-complete

---

### Week 3: Core Service Layer Implementation

#### Part A: ODS Import Service (3 days)
**File**: `internal/service/ods_import_service.go`

**Already Implemented**:
- Structure and interface
- ScrapeBatch lifecycle management
- Error handling pattern

**TODO**:
1. **Implement parseODSFile**:
   - Use validated ODS library from MASTER_PLAN Spike 3
   - Parse shifts from ODS cells
   - Extract shift type, date, time, coverage, specialty constraint
   - Collect parsing errors in ValidationResult
   - Return parsed schedules (not fail-fast)

2. **Implement validation logic**:
   - Validate shift types against known enum values
   - Validate dates within schedule version window
   - Validate specialty constraints
   - Collect all validation errors

3. **Write tests**:
   - Mock ODS file with known shifts
   - Verify parsing accuracy
   - Verify error collection
   - Test with large files (performance)

4. **Type-Safe Pattern**:
   ```go
   // After ValidationResult.Add() calls, check HasErrors()
   if result.HasErrors() && len(schedules) == 0 {
       batch.State = entity.BatchStateFailed
       return batch, result, nil
   }
   // Can import even with warnings
   if result.CanImport() {
       // proceed with import
   }
   ```

#### Part B: Dynamic Coverage Calculator (2 days)
**File**: `internal/service/coverage.go`

**Already Implemented**:
- Structure and interface
- Batch query pattern

**TODO**:
1. **Fix aggregateCoverage**:
   - Compare desired vs actual coverage by shift position
   - Group shifts by position key (shift type + study type + specialty)
   - Calculate coverage percentage
   - Identify gaps (positions with <80% coverage)

2. **Optimize getPositionKey**:
   - Generate consistent position identifier
   - Normalize shift type, study type, specialty

3. **Implement buildSummary**:
   - Calculate statistics:
     - Total shifts
     - Total desired coverage
     - Total assigned coverage
     - Average coverage percentage
     - Positions covered
   - Return as map[string]interface{}

4. **Write tests with query count assertions**:
   - Test batch query optimization
   - Verify exactly 2 queries (shifts + assignments)
   - NOT 1 + N queries per shift
   - Test with 100 shifts, 500 assignments
   - Assert query count never exceeds 3

**Critical Pattern** (prevents N+1 from v1):
```go
// Load all shifts in one query
shifts := GetByDateRange(ctx, versionID, start, end) // 1 query

// Load ALL assignments for these shifts in ONE query (not loop)
shiftIDs := extractIDs(shifts)
assignments := GetAllByShiftIDs(ctx, shiftIDs) // 1 query

// Process in memory
for shift := range shifts {
  for assignment := range assignments {
    if assignment.ShiftInstanceID == shift.ID {
      // aggregate
    }
  }
}
// Total: 2 queries regardless of number of shifts
// v1 bug: would be 1 + N queries (one per shift)
```

#### Part C: Schedule Version Service (1 day)
**File**: `internal/service/schedule_version_service.go`

**Status**: Service layer mostly complete, needs testing

**TODO**:
1. Write unit tests:
   - CreateVersion (creates in STAGING)
   - GetVersion (by ID)
   - GetActiveVersion (finds PRODUCTION for date)
   - ListVersionsByStatus
   - PromoteToProduction (STAGING→PRODUCTION)
   - Archive (PRODUCTION→ARCHIVED)
   - PromoteAndArchiveOthers (only one PRODUCTION at a time)

2. Test state transitions:
   - Can't promote non-STAGING
   - Can't archive non-PRODUCTION
   - Only one PRODUCTION per hospital

3. Test UpdatedBy/UpdatedAt tracking

**Target**: 100% coverage, all state transitions validated

---

### Week 4: Database Integration & Performance Testing

**Goal**: Verify repositories, connections, and performance

#### Part A: Repository Integration (3 days)
1. **Verify all repositories work together**:
   - Create hospital → create persons → create schedule version
   - Create shifts → create assignments
   - Query with joins across tables
   - Soft delete cascading correctly

2. **Test foreign key constraints**:
   - Deleting hospital should cascade soft delete to children
   - Foreign key constraints prevent orphan data
   - Audit trail tracks who deleted what

3. **Test transactions** (if implemented):
   - ODS import must be atomic
   - Rollback on validation failure

#### Part B: Performance Testing (2 days)
1. **Query performance**:
   - EXPLAIN ANALYZE for critical paths
   - Verify indexes are used
   - Benchmark against SLAs:
     - GetByID: <10ms
     - GetAll small result set: <50ms
     - Batch operations: <200ms

2. **Query count assertions**:
   - GetByScheduleVersion(shift) → exactly 1 query
   - GetAllByShiftIDs(assignments) → exactly 1 query
   - No N+1 regressions detected

3. **Load test**:
   - 100 hospitals
   - 500 persons per hospital
   - 1000 shifts per schedule version
   - 5000 assignments total
   - Verify no performance degradation

**Success**: All services <200ms response time under load

---

## Implementation Standards (TDD)

### For Each Service Component:

1. **Write Test First**:
   ```go
   func TestODSImportService_ParseValidFile(t *testing.T) {
       // Arrange
       mockODSData := loadTestODS("valid.ods")
       service := NewODSImportService(mockRepo, mockValidator)

       // Act
       result, err := service.parseODSFile(ctx, mockODSData)

       // Assert
       assert.NoError(t, err)
       assert.Equal(t, 10, len(result.Schedules))
       assert.False(t, result.HasErrors())
   }
   ```

2. **Implement to Pass Test**

3. **Refactor for Quality**:
   - Extract common patterns
   - Add error handling
   - Document assumptions

4. **Add Edge Case Tests**:
   - Empty files
   - Invalid data
   - Large files
   - Concurrent access

### Code Quality Requirements:
- [ ] No file > 400 lines (split if needed)
- [ ] All public functions have docstrings
- [ ] All errors wrapped with context (`fmt.Errorf("context: %w", err)`)
- [ ] No magic numbers (use constants)
- [ ] No hardcoded values
- [ ] Type-safe enums (not string constants)
- [ ] Tests document expected behavior

---

## Critical Patterns to Verify

### 1. Soft Delete Pattern (throughout)
```go
// All SELECTs must include deleted_at IS NULL check
WHERE id = $1 AND deleted_at IS NULL

// All DELETEs must be UPDATE to set deleted_at, deleted_by
UPDATE table SET deleted_at = NOW(), deleted_by = $1 WHERE id = $2
```
**Test**: Verify deleted records are not accessible

### 2. Batch Query Optimization (critical for performance)
```go
// GOOD: Single query with IN clause
SELECT * FROM assignments WHERE shift_instance_id IN ($1, $2, ..., $N)

// BAD: Loop with separate query per item
for shift := range shifts {
    SELECT * FROM assignments WHERE shift_instance_id = shift.ID
}
```
**Test**: Assert query count never exceeds 3 for any operation

### 3. Error Collection Pattern (from v1)
```go
// Collect ALL errors, don't fail fast
for _, field := range fields {
    if !isValid(field) {
        result.AddError(code, message)
    }
}
// Can still process if only warnings
if result.CanImport() {
    proceed()
}
```
**Test**: Verify errors are collected and CanImport() works correctly

### 4. Type Safety (enums, not strings)
```go
// GOOD
shift.ShiftType = entity.ShiftTypeDay
batch.State = entity.BatchStatePending

// BAD
shift.ShiftType = "day"
batch.Status = "PENDING"
```
**Test**: Type checker enforces at compile time

---

## Code Review Checklist

Before marking code as complete:
- [ ] Tests written first (TDD)
- [ ] 85%+ test coverage
- [ ] No N+1 query patterns
- [ ] All soft deletes use IS NULL filter
- [ ] All errors wrapped with context
- [ ] Docstrings explain WHY, not just WHAT
- [ ] No magic numbers
- [ ] Type-safe (enums, not strings)
- [ ] File < 400 lines (split if larger)
- [ ] No hardcoded values
- [ ] All TODOs removed
- [ ] Follows SOLID principles
- [ ] Integration tests with Testcontainers
- [ ] Query count assertions pass

---

## Deliverables by End of Phase 1

1. **Repositories**:
   - [x] 9 repositories implemented (3 + 6 new)
   - [ ] All 9 with integration tests
   - [ ] All query count assertions passing
   - [ ] Soft delete verified across all

2. **Validation Framework**:
   - [x] ValidationResult implemented
   - [ ] 100% test coverage
   - [ ] Matches v1 Java semantics exactly
   - [ ] Integration tests with services

3. **Core Services**:
   - [ ] ODS import service (parsing + validation)
   - [ ] Dynamic coverage calculator (batch queries, no N+1)
   - [ ] Schedule version service (state machine)
   - [ ] Schedule orchestrator (3-phase workflow)
   - [ ] All services with 80%+ test coverage

4. **Documentation**:
   - [ ] Schema documentation (v1 → v2 mapping)
   - [ ] Service architecture (how they work together)
   - [ ] Query count assertions framework
   - [ ] Type system guide (enums vs strings)

5. **Test Coverage**:
   - [ ] Unit tests: 80%+ on services
   - [ ] Integration tests: All repositories with Testcontainers
   - [ ] Performance tests: Query count assertions
   - [ ] Edge cases: Empty data, large files, invalid input

---

## Success Metrics

| Metric | Target | Status |
|--------|--------|--------|
| Repository integration tests | 100% passing | Pending |
| ValidationResult feature parity | 100% with v1 | In progress |
| Service unit test coverage | 80%+ | Pending |
| Core services feature complete | 100% | Pending |
| Query count assertions | All passing | Pending |
| No N+1 regressions | 0 found | TBD |
| Performance SLAs | All <200ms | TBD |
| Code quality | All checklist items | TBD |

---

## Go/No-Go Gate for Phase 2

**Cannot proceed to Phase 2 without:**
- [ ] All 9 repositories tested and working
- [ ] ValidationResult 100% feature-complete
- [ ] DynamicCoverageCalculator with batch query optimization
- [ ] Core services 80%+ tested
- [ ] No N+1 query regressions detected
- [ ] All soft delete patterns verified
- [ ] 85%+ coverage on core services
- [ ] Performance tests passing

---

## Next Steps

1. **Immediately**: Run integration test setup
   ```bash
   cd /home/lcgerke/schedCU/v2
   go test ./internal/repository/postgres -v -run TestPersonRepository_CRUD
   ```

2. **Day 1-2**: Get Testcontainers working
   - Fix any Docker/container issues
   - Verify PostgreSQL starts correctly
   - Verify migrations run

3. **Day 3-5**: Implement all repository tests
   - Test each CRUD operation
   - Test soft delete
   - Add query count assertions

4. **Week 2**: Complete validation framework tests

5. **Week 3**: Implement core services with TDD

6. **Week 4**: Performance testing + finalization

---

**Total Phase 1 Effort**: 4 weeks (full-time)
**Team Size**: 1-2 people
**Next Gate**: Phase 2 begins when all above items ✅

