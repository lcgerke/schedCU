# Repository Optimization Checklist

**Phase:** Phase 1 Implementation
**Target Completion:** Before Phase 1 development begins
**Total Recommendations:** 12 action items
**Priority Distribution:** 4 Critical, 4 High, 4 Medium

---

## Critical Priority (Must Complete)

### [ ] 1. Add PersonRepository.GetByIDs() Batch Method

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/person.go`

**Rationale:**
- **Risk:** HIGH - Assignment processing will trigger N+1 pattern without this
- **Impact:** 100+ assignments require 100 individual person lookups without batch
- **Savings:** 100 queries → 1 query (99% reduction)

**Implementation Template:**
```go
// GetByIDs retrieves multiple persons by their IDs (batch operation)
func (r *PersonRepository) GetByIDs(ctx context.Context, personIDs []uuid.UUID) ([]*entity.Person, error) {
    if len(personIDs) == 0 {
        return []*entity.Person{}, nil
    }

    query := `
        SELECT id, email, name, specialty, active, aliases, created_at, updated_at, deleted_at
        FROM persons
        WHERE id = ANY($1) AND deleted_at IS NULL
        ORDER BY name ASC
    `

    rows, err := r.db.QueryContext(ctx, query, pq.Array(personIDs))
    if err != nil {
        return nil, fmt.Errorf("failed to query persons by IDs: %w", err)
    }
    defer rows.Close()

    var persons []*entity.Person
    for rows.Next() {
        person := &entity.Person{}
        if err := r.scanPerson(rows, person); err != nil {
            return nil, fmt.Errorf("failed to scan person: %w", err)
        }
        persons = append(persons, person)
    }

    return persons, rows.Err()
}

// Helper to reduce scan duplication
func (r *PersonRepository) scanPerson(scanner interface{ Scan(...interface{}) error }, person *entity.Person) error {
    return scanner.Scan(
        &person.ID,
        &person.Email,
        &person.Name,
        (*string)(&person.Specialty),
        &person.Active,
        pq.Array(&person.Aliases),
        &person.CreatedAt,
        &person.UpdatedAt,
        &person.DeletedAt,
    )
}
```

**Interface Update Required:** Add to `repository.PersonRepository` interface in `/internal/repository/repository.go`

**Tests Required:**
- ✓ Empty slice handling
- ✓ Single person batch
- ✓ Multiple person batch
- ✓ Non-existent persons (should return empty for missing IDs)
- ✓ Soft-delete filtering

**Estimated Effort:** 30 minutes
- [ ] Method implementation
- [ ] Tests written
- [ ] Interface updated
- [ ] Integration tested with assignment processing

---

### [ ] 2. Implement JobQueueRepository.GetPending()

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/job_queue.go`

**Rationale:**
- **Risk:** HIGH - Job processing service cannot function without this
- **Missing:** Interface defines this method (repository.go line 163) but not implemented
- **Use Case:** Background job workers need to fetch pending jobs

**Implementation Template:**
```go
// GetPending retrieves all pending jobs ordered by creation time
func (r *JobQueueRepository) GetPending(ctx context.Context) ([]*entity.JobQueue, error) {
    query := `
        SELECT id, job_type, status, payload, result,
               retry_count, max_retries, error_message,
               created_at, started_at, completed_at
        FROM job_queue
        WHERE status = $1
        ORDER BY created_at ASC
    `

    rows, err := r.db.QueryContext(ctx, query, string(entity.JobQueueStatusPending))
    if err != nil {
        return nil, fmt.Errorf("failed to query pending jobs: %w", err)
    }
    defer rows.Close()

    var jobs []*entity.JobQueue
    for rows.Next() {
        job := &entity.JobQueue{
            Payload: make(map[string]interface{}),
            Result:  make(map[string]interface{}),
        }
        var payloadJSON, resultJSON []byte

        err := rows.Scan(
            &job.ID,
            &job.JobType,
            (*string)(&job.Status),
            &payloadJSON,
            &resultJSON,
            &job.RetryCount,
            &job.MaxRetries,
            &job.ErrorMessage,
            &job.CreatedAt,
            &job.StartedAt,
            &job.CompletedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan job: %w", err)
        }

        if len(payloadJSON) > 0 {
            if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
                return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
            }
        }

        if len(resultJSON) > 0 {
            if err := json.Unmarshal(resultJSON, &job.Result); err != nil {
                return nil, fmt.Errorf("failed to unmarshal result: %w", err)
            }
        }

        jobs = append(jobs, job)
    }

    return jobs, rows.Err()
}
```

**Tests Required:**
- ✓ No pending jobs returns empty slice
- ✓ Multiple pending jobs in creation order
- ✓ Ignores completed/failed jobs
- ✓ JSON payload handling

**Estimated Effort:** 20 minutes
- [ ] Method implementation
- [ ] Tests written
- [ ] Verified against job processing service

---

### [ ] 3. Implement JobQueueRepository.CleanupOldJobs()

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/job_queue.go`

**Rationale:**
- **Risk:** MEDIUM - Job table will grow unbounded without cleanup
- **Missing:** Interface defines this method (repository.go line 167) but not implemented
- **Use Case:** Maintenance task to remove completed jobs older than N days

**Implementation Template:**
```go
// CleanupOldJobs deletes completed jobs older than the specified number of days
func (r *JobQueueRepository) CleanupOldJobs(ctx context.Context, daysOld int) (int64, error) {
    query := `
        DELETE FROM job_queue
        WHERE status IN ($1, $2, $3)
          AND completed_at < NOW() - INTERVAL '1 day' * $4
    `

    result, err := r.db.ExecContext(ctx, query,
        string(entity.JobQueueStatusCompleted),
        string(entity.JobQueueStatusFailed),
        string(entity.JobQueueStatusCancelled),
        daysOld,
    )

    if err != nil {
        return 0, fmt.Errorf("failed to cleanup old jobs: %w", err)
    }

    rowsDeleted, err := result.RowsAffected()
    if err != nil {
        return 0, fmt.Errorf("failed to get rows affected: %w", err)
    }

    return rowsDeleted, nil
}
```

**Tests Required:**
- ✓ No jobs older than threshold
- ✓ Deletes only old completed jobs
- ✓ Preserves pending/in-progress jobs
- ✓ Returns correct row count

**Estimated Effort:** 15 minutes
- [ ] Method implementation
- [ ] Tests written
- [ ] Maintenance task integration tested

---

### [ ] 4. Add Soft-Delete Filtering to CoverageCalculationRepository

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/coverage_calculation.go`

**Rationale:**
- **Risk:** HIGH - Deleted calculations will appear in queries
- **Current State:** Schema has deleted_at but queries don't filter it
- **Impact:** Stale/deleted coverage data could be used in reports
- **Pattern:** All other repositories implement this correctly

**Changes Required:**

In `GetByScheduleVersion()` (line 138):
```diff
- WHERE schedule_version_id = $1
+ WHERE schedule_version_id = $1 AND deleted_at IS NULL
```

In `GetLatestByScheduleVersion()` (line 216):
```diff
- WHERE schedule_version_id = $1
+ WHERE schedule_version_id = $1 AND deleted_at IS NULL
```

In `GetByHospitalAndDate()` - **NOTE:** This method is defined in interface but missing from implementation!
```go
func (r *CoverageCalculationRepository) GetByHospitalAndDate(ctx context.Context, hospitalID uuid.UUID, date entity.Date) ([]*entity.CoverageCalculation, error) {
    query := `
        SELECT id, schedule_version_id, hospital_id, calculation_date,
               calculation_period_start_date, calculation_period_end_date,
               coverage_by_position, coverage_summary, validation_errors,
               query_count, calculated_at, calculated_by
        FROM coverage_calculations
        WHERE hospital_id = $1 AND calculation_date = $2 AND deleted_at IS NULL
        ORDER BY calculated_at DESC
    `
    // ... rest of query implementation
}
```

**Also Add to GetByID()** (line 92):
```diff
- WHERE id = $1
+ WHERE id = $1 AND deleted_at IS NULL
```

**Tests Required:**
- ✓ Deleted calculations not returned
- ✓ Active calculations returned
- ✓ Filtering consistent across all methods

**Estimated Effort:** 10 minutes
- [ ] All queries updated
- [ ] GetByHospitalAndDate implemented
- [ ] Tests verified

---

## High Priority

### [ ] 5. Add Soft-Delete Filtering to JobQueueRepository

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/job_queue.go`

**Rationale:**
- **Current State:** No soft-delete filtering in queries
- **Inconsistency:** All other repositories implement soft delete pattern
- **Impact:** Deleted jobs could reappear in queries

**Changes Required:**

In `GetByStatus()` (line 133):
```diff
- WHERE status = $1
+ WHERE status = $1 AND deleted_at IS NULL
```

In `GetByType()` (line 193):
```diff
- WHERE job_type = $1
+ WHERE job_type = $1 AND deleted_at IS NULL
```

In `GetByID()` (line 84):
```diff
- WHERE id = $1
+ WHERE id = $1 AND deleted_at IS NULL
```

In `GetPending()` - Add when implementing:
```sql
WHERE status = 'PENDING' AND deleted_at IS NULL
```

**Tests Required:**
- ✓ Deleted jobs not returned
- ✓ Active jobs returned
- ✓ Consistency across all methods

**Estimated Effort:** 8 minutes
- [ ] All queries updated
- [ ] Tests verified

---

### [ ] 6. Add ScrapeBatchRepository.GetByHospitalIDs() Batch Method

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/scrape_batch.go`

**Rationale:**
- **Use Case:** Multi-hospital ODS import orchestration
- **Pattern:** Matches AssignmentRepository.GetAllByShiftIDs()
- **Impact:** Avoid N queries for M hospitals

**Implementation Template:**
```go
// GetByHospitalIDs retrieves scrape batches for multiple hospitals (batch operation)
func (r *ScrapeBatchRepository) GetByHospitalIDs(ctx context.Context, hospitalIDs []uuid.UUID) ([]*entity.ScrapeBatch, error) {
    if len(hospitalIDs) == 0 {
        return []*entity.ScrapeBatch{}, nil
    }

    query := `
        SELECT id, hospital_id, state, window_start_date, window_end_date,
               scraped_at, row_count, error_message, created_at, created_by, deleted_at
        FROM scrape_batches
        WHERE hospital_id = ANY($1) AND deleted_at IS NULL
        ORDER BY hospital_id, created_at DESC
    `

    rows, err := r.db.QueryContext(ctx, query, pq.Array(hospitalIDs))
    if err != nil {
        return nil, fmt.Errorf("failed to query scrape batches by hospital IDs: %w", err)
    }
    defer rows.Close()

    var batches []*entity.ScrapeBatch
    for rows.Next() {
        batch := &entity.ScrapeBatch{}
        err := rows.Scan(
            &batch.ID,
            &batch.HospitalID,
            (*string)(&batch.State),
            &batch.WindowStartDate,
            &batch.WindowEndDate,
            &batch.ScrapedAt,
            &batch.RowCount,
            &batch.ErrorMessage,
            &batch.CreatedAt,
            &batch.CreatedBy,
            &batch.DeletedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan scrape batch: %w", err)
        }
        batches = append(batches, batch)
    }

    return batches, rows.Err()
}
```

**Interface Update Required:** Add to `repository.ScrapeBatchRepository` interface

**Tests Required:**
- ✓ Empty slice handling
- ✓ Single hospital batch
- ✓ Multiple hospital batches
- ✓ Ordered correctly by hospital then creation

**Estimated Effort:** 25 minutes
- [ ] Method implementation
- [ ] Tests written
- [ ] Interface updated

---

### [ ] 7. Add ScheduleVersionRepository.GetByHospitalIDs() Batch Method

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/schedule_version.go`

**Rationale:**
- **Use Case:** Phase 1 orchestration service needs multiple versions at once
- **Pattern:** Matches existing batch methods
- **Impact:** Avoid N queries for M hospitals

**Implementation Template:**
```go
// GetByHospitalIDs retrieves schedule versions for multiple hospitals (batch operation)
func (r *ScheduleVersionRepository) GetByHospitalIDs(ctx context.Context, hospitalIDs []uuid.UUID) ([]*entity.ScheduleVersion, error) {
    if len(hospitalIDs) == 0 {
        return []*entity.ScheduleVersion{}, nil
    }

    query := `
        SELECT id, hospital_id, status, effective_start_date, effective_end_date,
               scrape_batch_id, validation_results, created_at, created_by, updated_at,
               updated_by, deleted_at, deleted_by
        FROM schedule_versions
        WHERE hospital_id = ANY($1) AND deleted_at IS NULL
        ORDER BY hospital_id, created_at DESC
    `

    rows, err := r.db.QueryContext(ctx, query, pq.Array(hospitalIDs))
    if err != nil {
        return nil, fmt.Errorf("failed to query schedule versions by hospital IDs: %w", err)
    }
    defer rows.Close()

    var versions []*entity.ScheduleVersion
    for rows.Next() {
        version := &entity.ScheduleVersion{}
        var validationJSON []byte

        err := rows.Scan(
            &version.ID,
            &version.HospitalID,
            (*string)(&version.Status),
            &version.EffectiveStartDate,
            &version.EffectiveEndDate,
            &version.ScrapeBatchID,
            &validationJSON,
            &version.CreatedAt,
            &version.CreatedBy,
            &version.UpdatedAt,
            &version.UpdatedBy,
            &version.DeletedAt,
            &version.DeletedBy,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan schedule version: %w", err)
        }

        if len(validationJSON) > 0 {
            if err := json.Unmarshal(validationJSON, &version.ValidationResults); err != nil {
                return nil, fmt.Errorf("failed to unmarshal validation results: %w", err)
            }
        }

        versions = append(versions, version)
    }

    return versions, rows.Err()
}
```

**Interface Update Required:** Add to `repository.ScheduleVersionRepository` interface

**Tests Required:**
- ✓ Empty slice handling
- ✓ Single hospital batch
- ✓ Multiple hospital batches
- ✓ JSON validation handling
- ✓ Ordered correctly

**Estimated Effort:** 30 minutes
- [ ] Method implementation
- [ ] JSON unmarshaling tested
- [ ] Tests written
- [ ] Interface updated

---

### [ ] 8. Implement AuditLogRepository.GetByResource()

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/audit_log.go`

**Rationale:**
- **Missing:** Interface defines this method (repository.go line 139) but not implemented
- **Use Case:** Retrieve audit trail for specific resource changes
- **Required:** For compliance and debugging

**Implementation Template:**
```go
// GetByResource retrieves audit logs for a specific resource
func (r *AuditLogRepository) GetByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]*entity.AuditLog, error) {
    query := `
        SELECT id, user_id, action, resource, old_values, new_values, timestamp, ip_address
        FROM audit_logs
        WHERE resource = $1 AND user_id = $2
        ORDER BY timestamp DESC
    `

    rows, err := r.db.QueryContext(ctx, query, resourceType, resourceID)
    if err != nil {
        return nil, fmt.Errorf("failed to query audit logs by resource: %w", err)
    }
    defer rows.Close()

    var logs []*entity.AuditLog
    for rows.Next() {
        log := &entity.AuditLog{}
        err := rows.Scan(
            &log.ID,
            &log.UserID,
            &log.Action,
            &log.Resource,
            &log.OldValues,
            &log.NewValues,
            &log.Timestamp,
            &log.IPAddress,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan audit log: %w", err)
        }
        logs = append(logs, log)
    }

    return logs, rows.Err()
}
```

**Tests Required:**
- ✓ Single resource audit trail
- ✓ Multiple actions on same resource
- ✓ No logs for unknown resource
- ✓ Ordered by timestamp descending

**Estimated Effort:** 15 minutes
- [ ] Method implementation
- [ ] Tests written
- [ ] Verified with audit service

---

## Medium Priority

### [ ] 9. Create JSON Marshaling Helper Functions

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/postgres.go`

**Rationale:**
- **Duplication:** 3 repositories repeat JSON marshal/unmarshal patterns
- **Consistency:** Centralize error handling for JSON operations
- **Maintainability:** Single place to update JSON handling

**Implementation Template:**
```go
package postgres

import (
    "encoding/json"
    "fmt"
)

// marshalJSON converts a value to JSON bytes with error wrapping
func marshalJSON(v interface{}, fieldName string) ([]byte, error) {
    data, err := json.Marshal(v)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal %s: %w", fieldName, err)
    }
    return data, nil
}

// unmarshalJSON converts JSON bytes to a value with error wrapping
func unmarshalJSON(data []byte, v interface{}, fieldName string) error {
    if len(data) == 0 {
        return nil  // null/empty is acceptable
    }
    err := json.Unmarshal(data, v)
    if err != nil {
        return fmt.Errorf("failed to unmarshal %s: %w", fieldName, err)
    }
    return nil
}
```

**Usage Example (Before):**
```go
validationJSON, err := json.Marshal(version.ValidationResults)
if err != nil {
    return fmt.Errorf("failed to marshal validation results: %w", err)
}

// ... later ...

if len(validationJSON) > 0 {
    if err := json.Unmarshal(validationJSON, &version.ValidationResults); err != nil {
        return nil, fmt.Errorf("failed to unmarshal validation results: %w", err)
    }
}
```

**Usage Example (After):**
```go
validationJSON, err := marshalJSON(version.ValidationResults, "validation_results")
if err != nil {
    return err
}

// ... later ...

if err := unmarshalJSON(validationJSON, &version.ValidationResults, "validation_results"); err != nil {
    return nil, err
}
```

**Refactor Locations:**
- ScheduleVersionRepository (1 field)
- CoverageCalculationRepository (3 fields)
- JobQueueRepository (2 fields)

**Estimated Effort:** 20 minutes
- [ ] Helper functions created
- [ ] ScheduleVersionRepository refactored
- [ ] CoverageCalculationRepository refactored
- [ ] JobQueueRepository refactored
- [ ] Tests verify unchanged behavior

---

### [ ] 10. Fix PersonRepository.GetByHospital() Implementation

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/person.go` line 125-128

**Current State:**
```go
func (r *PersonRepository) GetByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.Person, error) {
    // Note: This needs a person_hospitals bridge table or hospital_id on persons table
    // For now, returning empty slice - implement based on actual schema
    return []*entity.Person{}, nil
}
```

**Rationale:**
- **Blocker:** Method returns empty every time, preventing hospital person lists
- **Required:** For hospital-specific staff management
- **Decision Needed:** Confirm schema (bridge table vs direct column)

**Options:**

**Option A: Direct Column (If persons table has hospital_id)**
```go
func (r *PersonRepository) GetByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.Person, error) {
    query := `
        SELECT id, email, name, specialty, active, aliases, created_at, updated_at, deleted_at
        FROM persons
        WHERE hospital_id = $1 AND deleted_at IS NULL
        ORDER BY name ASC
    `
    // ... standard query implementation
}
```

**Option B: Bridge Table (If person_hospitals junction table exists)**
```go
func (r *PersonRepository) GetByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.Person, error) {
    query := `
        SELECT DISTINCT p.id, p.email, p.name, p.specialty, p.active, p.aliases,
                p.created_at, p.updated_at, p.deleted_at
        FROM persons p
        INNER JOIN person_hospitals ph ON p.id = ph.person_id
        WHERE ph.hospital_id = $1 AND p.deleted_at IS NULL
        ORDER BY p.name ASC
    `
    // ... standard query implementation
}
```

**Action Required:**
1. [ ] Confirm actual schema structure
2. [ ] Implement based on schema
3. [ ] Add tests for hospital staff retrieval
4. [ ] Update documentation with chosen pattern

**Estimated Effort:** 15 minutes (decision) + 15 minutes (implementation)

---

### [ ] 11. Fix CoverageCalculationRepository.GetByHospitalAndDate() Missing

**Location:** `/home/lcgerke/schedCU/v2/internal/repository/postgres/coverage_calculation.go`

**Current State:**
- Interface defines method (repository.go line 128)
- Implementation completely missing from postgres/coverage_calculation.go

**Required Implementation:**
```go
// GetByHospitalAndDate retrieves coverage calculations for a hospital on a specific date
func (r *CoverageCalculationRepository) GetByHospitalAndDate(ctx context.Context, hospitalID uuid.UUID, date entity.Date) ([]*entity.CoverageCalculation, error) {
    query := `
        SELECT id, schedule_version_id, hospital_id, calculation_date,
               calculation_period_start_date, calculation_period_end_date,
               coverage_by_position, coverage_summary, validation_errors,
               query_count, calculated_at, calculated_by
        FROM coverage_calculations
        WHERE hospital_id = $1 AND calculation_date = $2 AND deleted_at IS NULL
        ORDER BY calculated_at DESC
    `

    rows, err := r.db.QueryContext(ctx, query, hospitalID, date)
    if err != nil {
        return nil, fmt.Errorf("failed to query coverage calculations: %w", err)
    }
    defer rows.Close()

    var calcs []*entity.CoverageCalculation
    for rows.Next() {
        calc := &entity.CoverageCalculation{
            CoverageByPosition: make(map[string]int),
            CoverageSummary:    make(map[string]interface{}),
        }
        var coverageByPositionJSON, coverageSummaryJSON, validationJSON []byte

        err := rows.Scan(
            &calc.ID,
            &calc.ScheduleVersionID,
            &calc.HospitalID,
            &calc.CalculationDate,
            &calc.CalculationPeriodStartDate,
            &calc.CalculationPeriodEndDate,
            &coverageByPositionJSON,
            &coverageSummaryJSON,
            &validationJSON,
            &calc.QueryCount,
            &calc.CalculatedAt,
            &calc.CalculatedBy,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan coverage calculation: %w", err)
        }

        if err := json.Unmarshal(coverageByPositionJSON, &calc.CoverageByPosition); err != nil {
            return nil, fmt.Errorf("failed to unmarshal coverage by position: %w", err)
        }

        if err := json.Unmarshal(coverageSummaryJSON, &calc.CoverageSummary); err != nil {
            return nil, fmt.Errorf("failed to unmarshal coverage summary: %w", err)
        }

        if len(validationJSON) > 0 {
            if err := json.Unmarshal(validationJSON, &calc.ValidationErrors); err != nil {
                return nil, fmt.Errorf("failed to unmarshal validation errors: %w", err)
            }
        }

        calcs = append(calcs, calc)
    }

    return calcs, rows.Err()
}
```

**Tests Required:**
- ✓ Single date retrieval
- ✓ Multiple calculations same date
- ✓ No calculations on date
- ✓ JSON field handling
- ✓ Soft-delete filtering

**Estimated Effort:** 20 minutes
- [ ] Method implemented
- [ ] Tests written
- [ ] Integrated with coverage service

---

### [ ] 12. Create Query Count Expectations Document

**Location:** `/home/lcgerke/schedCU/v2/docs/QUERY_COUNT_EXPECTATIONS.md`

**Rationale:**
- **Planning:** Phase 1 teams need to know expected query counts
- **Performance:** Baseline to detect N+1 regressions
- **Monitoring:** Track optimization effectiveness

**Document Should Include:**

For each Phase 1 operation:
1. Operation name
2. Expected query count (optimized)
3. Naive implementation count (without batch)
4. Key queries (SELECT x, UPDATE y, INSERT z)
5. Index dependencies
6. Potential bottlenecks

**Example Format:**

```markdown
## ODS Import (100 shifts scenario)

### Expected Query Count: 15-20 queries

#### Breakdown:
- 1x Hospital validation
- 1x Schedule version check
- 1x Person lookup (batch)
- 1x Shift instances batch create
- 100x Assignment creates
- 1x Coverage calculation
- 3x Audit logs

### Naive Implementation: 208 queries
(Without batch operations)

### Index Dependencies:
- shift_instances(schedule_version_id, schedule_date)
- assignments(shift_instance_id, deleted_at)

### Potential Bottlenecks:
- 100 individual INSERT statements for assignments
- Missing PersonRepository.GetByIDs() causes N+1
```

**Operations to Document:**
1. ODS Import (100 shifts)
2. Schedule Publication
3. Coverage Calculation
4. Assignment Modification
5. Person Management
6. Job Queue Processing
7. Report Generation

**Estimated Effort:** 30 minutes
- [ ] Document created with structure
- [ ] All Phase 1 operations listed
- [ ] Query counts verified
- [ ] Indexed and published

---

## Summary

### Completion Status

| Item | Priority | Status | Effort | Owner |
|---|---|---|---|---|
| PersonRepository.GetByIDs() | Critical | Not Started | 30m | @Dev |
| JobQueueRepository.GetPending() | Critical | Not Started | 20m | @Dev |
| JobQueueRepository.CleanupOldJobs() | Critical | Not Started | 15m | @Dev |
| CoverageCalculation soft-delete | Critical | Not Started | 10m | @Dev |
| JobQueue soft-delete | High | Not Started | 8m | @Dev |
| ScrapeBatch.GetByHospitalIDs() | High | Not Started | 25m | @Dev |
| ScheduleVersion.GetByHospitalIDs() | High | Not Started | 30m | @Dev |
| AuditLog.GetByResource() | High | Not Started | 15m | @Dev |
| JSON helpers | Medium | Not Started | 20m | @Dev |
| PersonRepository.GetByHospital() | Medium | Blocked | 30m | @Design |
| CoverageCalculation.GetByHospitalAndDate() | Medium | Not Started | 20m | @Dev |
| Query Count Expectations doc | Medium | Not Started | 30m | @Tech Writer |

**Total Effort:** ~253 minutes (4.2 hours)

### Recommended Execution Order

1. **Day 1 - Critical Items (1.5 hours)**
   - PersonRepository.GetByIDs()
   - JobQueueRepository.GetPending()
   - JobQueueRepository.CleanupOldJobs()

2. **Day 2 - High Priority (1.5 hours)**
   - Add soft-delete filtering (10 minutes)
   - Add batch methods (55 minutes)
   - Implement AuditLog.GetByResource() (15 minutes)

3. **Day 3 - Medium Priority (1.2 hours)**
   - JSON helpers refactoring (20 minutes)
   - Fix implementations (50 minutes)
   - Query count documentation (30 minutes)

### Phase 1 Blockers Resolution

**Before Phase 1 can start:**
- ✗ PersonRepository.GetByIDs() - BLOCKS assignment processing
- ✗ CoverageCalculation soft-delete - BLOCKS report accuracy
- ✗ JobQueue GetPending/CleanupOldJobs - BLOCKS job processing

**Recommended parallelization:**
- 2 developers on critical items (day 1)
- 2 developers on high priority items (day 2)
- 1 person on documentation (concurrent with days 1-2)

---

**Document Created:** November 15, 2025
**Target Completion:** Before Phase 1 Development
**Review Frequency:** Weekly during implementation
