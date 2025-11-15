# Query Count Expectations for Phase 1 Operations

**Phase:** Phase 1 Implementation
**Purpose:** Establish baseline query counts for performance monitoring and regression detection
**Date:** November 15, 2025
**Version:** 1.0

---

## Overview

This document establishes expected query counts for all major Phase 1 operations. These numbers assume:
- ✓ All batch operations are properly implemented (PersonRepository.GetByIDs, etc.)
- ✓ All soft-delete filtering is in place
- ✓ Proper indices are created on critical columns
- ✓ No application-level N+1 queries

**Query Count Formula:**
- **Optimized:** Count with all recommendations implemented
- **Current:** Count with current codebase (missing batch operations)
- **Naive:** Count with anti-patterns (per-record queries)

---

## 1. ODS Import Operations

### 1.1 ODS File Import (100 Shifts)

**Operation:** Upload ODS file, parse, validate, create schedule version and shifts

**Scenario:** Hospital uploads ODS with 100 shifts for 50 people

**Expected Query Count - OPTIMIZED: 107 queries**

**Breakdown:**

```
Initialization Phase:
  1. Get Hospital by ID (validation)                           1 query
  2. Check existing schedule versions for hospital            1 query
  3. Batch fetch all persons (50 people)                     1 query
                                            Subtotal: 3 queries

Schedule Creation Phase:
  4. Create ScheduleVersion record                            1 query
  5. Create 100 ShiftInstance records (batch)               100 queries
                                            Subtotal: 101 queries

Validation & Audit Phase:
  6. Create 1 CoverageCalculation record                      1 query
  7. Create audit log entries (import event)                  1 query
  8. Create audit log entries (per shift, sampled)           2 queries
                                            Subtotal: 4 queries

TOTAL OPTIMIZED: 108 queries
```

**Current Implementation (Missing PersonRepository.GetByIDs): 108 queries**
- Same as optimized, but person lookup pattern is different
- Still valid since batch isn't critical at 50 persons

**Naive Implementation (Without Optimization): 207 queries**

```
Without batch operations:
- 50 individual person lookups (instead of 1 batch):  +49 queries
- Shift creation loops with lookups:                   +0 queries
- Full audit trail logging:                           +50 additional logs

Total: ~207 queries
```

**Index Dependencies:**
- ✓ hospitals(id) - for validation
- ✓ persons(email, specialty) - for ODS matching
- ✓ schedule_versions(hospital_id, deleted_at) - for duplication check

**Query Breakdown by Type:**

| Type | Count | Examples |
|---|---|---|
| SELECT (single) | 3 | GetByID, GetByStatus |
| SELECT (multiple) | 2 | GetByHospitalAndStatus, batch fetch |
| INSERT | 101 | ScheduleVersion(1) + ShiftInstances(100) |
| UPDATE | 1 | Create audit log |

**Performance Expectations:**

| Phase | Expected Time | Constraint |
|---|---|---|
| Parse ODS | 100-500ms | File I/O, parsing |
| Validate Data | 50-200ms | In-memory validation |
| Database Queries | 200-500ms | Network + query execution |
| Total Operation | 500-1200ms | User-acceptable response |

**Bottlenecks:**
1. 100 individual INSERT statements for shifts - could batch with multi-row INSERT
2. Validation JSON serialization - minimal impact
3. Network latency to database - depends on infrastructure

---

### 1.2 ODS Import - Multi-Hospital (3 Hospitals × 100 Shifts each)

**Operation:** Orchestrated import for 3 hospitals in sequence

**Scenario:** System processes 3 hospital ODS files back-to-back

**Expected Query Count - OPTIMIZED: 325 queries**

**Breakdown (per hospital × 3):**
- Validation & setup: 3 queries × 3 = 9 queries
- Shift creation: 101 queries × 3 = 303 queries
- Validation & audit: 4 queries × 3 = 12 queries
- Overlap/coordination: 1 query (batch version check)

**Total: 325 queries**

**Current Implementation: 325 queries**

**Naive Implementation: 621 queries**
- 3 × 207 + 3 coordination queries

**Query Optimization with Multi-Row INSERT (Future Enhancement):**

If shift creation is optimized to use multi-row INSERT:
```sql
INSERT INTO shift_instances (...) VALUES
  ($1, $2, $3, ...),
  ($4, $5, $6, ...),
  ...
  ($298, $299, $300, ...)
```

**Result:** ~110 queries (additional 10-20% improvement possible)

**Execution Model:**
- Sequential: 1200ms × 3 = 3600ms (60 seconds)
- Concurrent: 1200ms (limited to 1 at a time by job queue)

---

### 1.3 ODS Validation (Pre-Import)

**Operation:** Validate ODS file without creating records

**Scenario:** User clicks "Preview/Validate" before import

**Expected Query Count - OPTIMIZED: 6 queries**

**Breakdown:**
```
1. Get Hospital                                    1 query
2. Get existing persons for matching             1 query
3. Get existing schedule versions                1 query
4. Get shift type definitions (reference data)   1 query
5. Get specialty definitions (reference data)    1 query
6. Create validation audit log                   1 query

TOTAL: 6 queries
```

**Performance Expectation:**
- Validation logic: 100-300ms (in-memory)
- Database queries: 20-50ms
- Total: 150-400ms

**Note:** Validation doesn't create ShiftInstance or Assignment records, so query count is dramatically lower.

---

## 2. Schedule Operations

### 2.1 Schedule Publication (Draft to Production)

**Operation:** Publish a draft schedule version to production, making it active

**Scenario:** Admin publishes draft schedule with 100 shifts/50 assignments to PRODUCTION status

**Expected Query Count - OPTIMIZED: 8 queries**

**Breakdown:**
```
1. Get ScheduleVersion by ID (validate draft status)         1 query
2. Get all ShiftInstances for version                        1 query
3. Get all Assignments (batch by shift IDs)                 1 query
4. Get CoverageCalculation for validation check             1 query
5. Update ScheduleVersion status to PRODUCTION               1 query
6. Create audit log (status change)                         1 query
7. Create audit log (publication summary)                   1 query
8. Get new active version for confirmation                  1 query

TOTAL: 8 queries
```

**Current Implementation: 8 queries** (all batch methods exist)

**Naive Implementation: 158 queries**
- Without GetAllByShiftIDs: 100 individual assignment fetches
- Without batch version check: 2 extra lookups
- Full audit trail: +50 entries

**Performance Expectations:**
- Validation: 50-100ms
- Database queries: 30-80ms
- Total response: 100-200ms

**Index Dependencies:**
- ✓ schedule_versions(hospital_id, status) - for active version lookup
- ✓ assignments(shift_instance_id) - for batch fetch
- ✓ coverage_calculations(schedule_version_id) - for validation

---

### 2.2 Schedule Modification (Shift Update)

**Operation:** Modify a single shift in a DRAFT schedule

**Scenario:** Change shift time or specialty requirement

**Expected Query Count - OPTIMIZED: 6 queries**

**Breakdown:**
```
1. Get ScheduleVersion (verify DRAFT status)              1 query
2. Get ShiftInstance (current state)                      1 query
3. Get Assignments for shift                             1 query
4. Update ShiftInstance                                  1 query
5. Update Assignments (if time changed)                  1 query
6. Create audit log (modification)                       1 query

TOTAL: 6 queries
```

**Current Implementation: 6 queries**

**Naive Implementation: 156 queries**
- Individual assignment updates: 50 queries
- Individual person updates for time conflict checks: +100 queries

**Performance Expectations:**
- Validation: 20-50ms
- Database queries: 15-40ms
- Total response: 50-100ms

**Impact on Dependent Calculations:**
- Coverage calculation invalidated → requires recalculation
- May trigger background job (see Job Queue section)

---

### 2.3 Schedule Deletion (Soft Delete)

**Operation:** Delete/archive a draft schedule version

**Scenario:** User cancels draft, removes from system

**Expected Query Count - OPTIMIZED: 5 queries**

**Breakdown:**
```
1. Get ScheduleVersion (verify exists and DRAFT)        1 query
2. Delete all ShiftInstances (soft-delete in batch)     1 query
3. Delete all Assignments (soft-delete in batch)        1 query
4. Delete ScheduleVersion itself                        1 query
5. Create audit log (deletion)                          1 query

TOTAL: 5 queries
```

**Current Implementation: 5 queries**

**Naive Implementation: ~404 queries**
- Per-shift soft deletes: 100 queries
- Per-assignment soft deletes: 100 queries
- Per-assignment audit logs: 100 entries
- Additional cascade checks: 100+ queries

**Performance Expectations:**
- Validation: 10-20ms
- Database operations: 25-60ms
- Total response: 50-100ms

**Index Dependencies:**
- ✓ shift_instances(schedule_version_id) - for batch delete
- ✓ assignments(shift_instance_id) - for batch delete

---

## 3. Assignment Operations

### 3.1 Bulk Assignment Creation (100 Shifts, 50 People)

**Operation:** Create assignments for all shifts during import/publication

**Scenario:** New schedule version gets all assignments created at once

**Expected Query Count - OPTIMIZED: 105 queries**

**Breakdown:**
```
1. Get all shifts (batch or from cache)                  1 query
2. Get all persons for capacity checks                   1 query
3. For each shift, validate availability               0 queries (in-memory)
4. Batch create 100 assignments                       100 queries
5. Create audit log entries (summary)                   1 query
6. Create coverage calculation                          1 query

TOTAL: 104 queries
```

**Current Implementation: 104 queries**

**Naive Implementation: 307 queries**
- Individual person availability checks: 100 queries
- Individual assignment creates: 100 queries
- Individual audit logs: 100+ entries

**Performance Expectations:**
- Availability validation: 100-300ms
- Database operations: 150-300ms
- Total operation: 300-600ms

**Bottleneck:** Individual INSERT statements for assignments
- Could optimize with multi-row INSERT: 2-3 large inserts instead of 100

---

### 3.2 Single Assignment Modification

**Operation:** Change a person's assignment for a single shift

**Scenario:** Manager reassigns a shift to a different person

**Expected Query Count - OPTIMIZED: 7 queries**

**Breakdown:**
```
1. Get ShiftInstance (validate exists)                   1 query
2. Get current Assignment                               1 query
3. Get Person (new assignee, validate availability)     1 query
4. Delete previous assignment                           1 query
5. Create new assignment                                1 query
6. Update coverage calculation flag                      1 query
7. Create audit log (reassignment)                       1 query

TOTAL: 7 queries
```

**Current Implementation: 7 queries**

**Naive Implementation: 157 queries**
- Availability check across all person assignments: 50+ queries
- Conflict detection: 100+ queries

**Performance Expectations:**
- Validation: 30-80ms
- Database operations: 20-50ms
- Total response: 60-150ms

**Side Effects:**
- Invalidates coverage calculation (triggers background job)
- May trigger conflict detection job

---

### 3.3 Batch Assignment Import (CSV/Excel)

**Operation:** Import assignments from external file

**Scenario:** Upload CSV with 100 person-shift mappings

**Expected Query Count - OPTIMIZED: 106 queries**

**Breakdown:**
```
1. Parse file (0 queries, in-memory)
2. Batch fetch all persons (50 unique)                    1 query
3. Batch fetch all shifts (100)                          1 query
4. Validate all rows (0 queries, in-memory)
5. Create/update 100 assignments                       100 queries
6. Update coverage calculation                          1 query
7. Create audit log entries                             3 queries

TOTAL: 106 queries
```

**Current Implementation: 106 queries**

**Naive Implementation: 257 queries**
- Per-row person lookup: 100 queries
- Per-row shift lookup: 100 queries
- Validation lookups: +57 queries

**Performance Expectations:**
- Parse CSV: 50-200ms
- Validation: 50-150ms
- Database operations: 150-300ms
- Total operation: 300-700ms

---

## 4. Coverage Calculation Operations

### 4.1 Coverage Calculation (Single Schedule Version)

**Operation:** Calculate coverage metrics for schedule version

**Scenario:** Run coverage analysis on 100-shift schedule with 50 staff

**Expected Query Count - OPTIMIZED: 8 queries**

**Breakdown:**
```
1. Get ScheduleVersion (for metadata)                      1 query
2. Get all ShiftInstances                                 1 query
3. Get all Assignments (batch by shift IDs)               1 query
4. Get all Persons (batch by person IDs)                  1 query
5. Get previous coverage calculation (for comparison)      1 query
6. Create new CoverageCalculation                         1 query
7. Create audit log (calculation summary)                 1 query
8. Get validation errors (if any)                         1 query

TOTAL: 8 queries
```

**Current Implementation: 8 queries**

**Naive Implementation: 258 queries**
- Individual assignment fetches: 100 queries
- Individual person fetches: 50 queries
- Per-shift validation: 100+ queries

**Performance Expectations:**
- Coverage algorithm: 500-1000ms (depends on complexity)
- Database operations: 40-100ms
- Total operation: 600-1200ms

**Timeout Considerations:**
- Calculation may take >30 seconds for large hospitals
- Should run as background job (see Job Queue section)

---

### 4.2 Coverage Report Generation (Multi-Version)

**Operation:** Generate coverage report comparing multiple versions

**Scenario:** Show coverage trends across 5 recent schedule versions

**Expected Query Count - OPTIMIZED: 12 queries**

**Breakdown:**
```
1. Get ScheduleVersion (metadata for report)               1 query
2. Get last 5 CoverageCalculations (batch by version)      1 query
3. Get shift instances for each version (batch)            5 queries
4. Get assignments for all versions (batch)                1 query
5. Get persons involved in any version (batch)             1 query
6. Create audit log (report generation)                    1 query
7. Cache result in audit log                               1 query

TOTAL: 12 queries
```

**Current Implementation: 12 queries**

**Naive Implementation: 365 queries**
- Per-version fetch: 5 × 3 = 15 queries
- All assignments per-shift: 500 queries

**Performance Expectations:**
- Report generation: 200-500ms
- Database operations: 50-150ms
- Total report time: 300-700ms

**Index Dependencies:**
- ✓ coverage_calculations(schedule_version_id, created_at) - for version lookups
- ✓ assignments(shift_instance_id) - for batch assignment fetch

---

## 5. Job Queue Operations

### 5.1 Job Processing (Single Job)

**Operation:** Worker picks up a pending job and executes it

**Scenario:** Background worker processes 1 coverage calculation job

**Expected Query Count - OPTIMIZED: 6 queries**

**Breakdown:**
```
1. Get pending job (GetPending batch, pick first)          1 query
2. Update job status to RUNNING                           1 query
3. Execute job (may include DB operations)              1-5 queries
4. Update job status to COMPLETED                        1 query
5. Create audit log (job completion)                      1 query
6. Log job metrics                                        0 queries

TOTAL: 5-9 queries (depending on job type)
```

**Current Implementation: Missing**
- GetPending() not implemented yet

**Naive Implementation: 107+ queries**
- Per-job status check: 1 query
- Job execution: varies
- Full logging of every step: 100+ audit entries

**Performance Expectations:**
- Job queue lookup: 5-10ms
- Job execution: 500-5000ms (depends on job)
- Database updates: 20-50ms
- Total: 600-5100ms

**Variations by Job Type:**

| Job Type | Expected Additional Queries | Typical Duration |
|---|---|---|
| Coverage Calculation | 7-10 | 500-1000ms |
| Validation Job | 3-5 | 100-300ms |
| Import Job | 100-150 | 1000-3000ms |
| Cleanup Job | 1 | 10-50ms |
| Report Generation | 12-20 | 500-1500ms |

---

### 5.2 Job Queue Cleanup

**Operation:** Remove completed jobs older than N days

**Scenario:** Maintenance task removes jobs older than 30 days

**Expected Query Count - OPTIMIZED: 1 query**

**Breakdown:**
```
1. Delete completed jobs older than 30 days              1 query

TOTAL: 1 query
```

**Current Implementation: Missing**
- CleanupOldJobs() not implemented yet

**Performance Expectations:**
- Database cleanup: 100-500ms (depends on job volume)
- Typically runs at off-hours

**Frequency:**
- Daily maintenance task
- Can be scheduled via cron or external job scheduler

---

### 5.3 Batch Job Processing (Multiple Jobs)

**Operation:** Worker processes batch of N pending jobs

**Scenario:** Run 50 pending coverage calculation jobs

**Expected Query Count - OPTIMIZED: 100-150 queries**

**Breakdown:**
```
1. Get all pending jobs                                    1 query
2. For each of 50 jobs:
   - Update status to RUNNING                          50 queries
   - Execute job (7 queries each)                      350 queries
   - Update status to COMPLETED                         50 queries
   - Create audit log                                   50 queries

TOTAL: 501 queries (over 50 jobs)
Average per job: 10 queries
```

**Expected Duration:**
- Batch fetch: 10-20ms
- 50 jobs × 700ms average = 35 seconds
- Status updates: 100-200ms
- Total: ~36 seconds

**Optimization Opportunity:**
- Parallel job processing (2-4 workers)
- Reduces total duration from 36s to 10-15s

---

## 6. User & Authentication Operations

### 6.1 User Login

**Operation:** Authenticate user and establish session

**Scenario:** User logs in with email/password

**Expected Query Count - OPTIMIZED: 3 queries**

**Breakdown:**
```
1. Get user by email                                      1 query
2. Verify password hash (in-memory)                    0 queries
3. Create audit log (login event)                         1 query
4. Get user hospital/role (for session)                   1 query

TOTAL: 3 queries
```

**Current Implementation: 3 queries**

**Performance Expectations:**
- Password verification: 50-100ms (bcrypt operations)
- Database operations: 10-20ms
- Total login: 80-150ms

---

### 6.2 User Management (List Hospital Users)

**Operation:** Get all users for a hospital

**Scenario:** Load user list for admin dashboard

**Expected Query Count - OPTIMIZED: 1 query**

**Breakdown:**
```
1. Get all users by hospital                              1 query

TOTAL: 1 query
```

**Performance Expectations (50 users):**
- Database query: 5-15ms
- Data rendering: 20-50ms
- Total: 30-70ms

---

## 7. Summary Table - All Operations

| Operation | Optimized | Current | Naive | Performance |
|---|---|---|---|---|
| ODS Import (100 shifts) | 108 | 108 | 207 | 500-1200ms |
| ODS Multi-Hospital (3×100) | 325 | 325 | 621 | 3-5s |
| ODS Validation | 6 | 6 | 56 | 150-400ms |
| Schedule Publication | 8 | 8 | 158 | 100-200ms |
| Schedule Modification | 6 | 6 | 156 | 50-100ms |
| Schedule Deletion | 5 | 5 | 404 | 50-100ms |
| Bulk Assignment Creation | 104 | 104 | 307 | 300-600ms |
| Single Assignment Update | 7 | 7 | 157 | 60-150ms |
| Assignment CSV Import | 106 | 106 | 257 | 300-700ms |
| Coverage Calculation | 8 | 8 | 258 | 600-1200ms |
| Coverage Report (5 versions) | 12 | 12 | 365 | 300-700ms |
| Single Job Processing | 5-9 | N/A | 107+ | 600-5100ms |
| Job Cleanup | 1 | N/A | 1 | 100-500ms |
| Batch Job Processing (50) | 501 | N/A | 1250+ | 36s (or 10-15s parallel) |
| User Login | 3 | 3 | 3 | 80-150ms |
| List Hospital Users | 1 | 1 | 1 | 30-70ms |

---

## 8. Performance Targets

### Query Count SLAs

| Operation | Target | Acceptable | Flag |
|---|---|---|---|
| Single entity operations | <10 queries | <20 queries | >30 queries |
| Batch operations | <120 queries | <200 queries | >250 queries |
| Report operations | <20 queries | <50 queries | >100 queries |
| Background jobs | N/A | <500 total queries | >1000 queries |

### Response Time SLAs

| Operation | Target | Acceptable | Flag |
|---|---|---|---|
| Single entity (CRUD) | <100ms | <200ms | >500ms |
| List/search results | <200ms | <500ms | >1000ms |
| Import operations | <1000ms | <2000ms | >5000ms |
| Calculations | <2000ms | <5000ms | >10000ms |
| Reports | <500ms | <2000ms | >5000ms |
| Background jobs | N/A | <10s per job | >30s per job |

---

## 9. Monitoring Recommendations

### Key Metrics to Track

1. **Query Count per Operation**
   - Track actual vs expected
   - Alert if > 150% of expected

2. **Query Execution Time**
   - Slow query log (queries > 100ms)
   - Query plan analysis for slow queries

3. **N+1 Pattern Detection**
   - Flag operations with > 50 sequential queries
   - Analyze query logs for repeated patterns

4. **Index Effectiveness**
   - Query plans using table scans
   - Missing index detection

5. **Connection Pool**
   - Active connections during peak ops
   - Queue depth/wait times

### Monitoring Implementation

```go
// Example query logger
type QueryLogger struct {
    operation string
    startTime time.Time
    queryCount int
}

// Alert if query count exceeds expected
if q.queryCount > expectedCount * 1.5 {
    log.Warnf("N+1 pattern detected: %s executed %d queries (expected %d)",
        q.operation, q.queryCount, expectedCount)
}

// Alert if operation time exceeds SLA
duration := time.Since(q.startTime)
if duration > sla {
    log.Warnf("SLA violation: %s took %v (SLA: %v)",
        q.operation, duration, sla)
}
```

---

## 10. Implementation Checkpoints

### Checkpoint 1: Before Phase 1 Development Begins
- [ ] All query count expectations documented
- [ ] Team briefed on expected query patterns
- [ ] Monitoring infrastructure ready
- [ ] Baseline query counts captured

### Checkpoint 2: Spike Phase (First 2 Weeks)
- [ ] Verify actual query counts match expectations
- [ ] Identify N+1 patterns in early code
- [ ] Adjust batch operation usage if needed
- [ ] Document actual vs expected discrepancies

### Checkpoint 3: Feature Complete (End of Phase 1)
- [ ] All operations tested against query budgets
- [ ] Performance tuning complete
- [ ] Index creation verified
- [ ] Final metrics documented

### Checkpoint 4: Production Ready
- [ ] Monitoring alerts configured
- [ ] Query logs analyzed for patterns
- [ ] Slow query log reviewed
- [ ] Performance baselines established for CI/CD

---

## 11. Query Budget Tracking Template

For each operation, track actual performance:

```
Operation: ODS Import (100 shifts)
Date: YYYY-MM-DD
Team Member: @Name

Expected Queries: 108
Actual Queries: 109
Variance: +1 (0.9%)
Status: ✓ PASS

Expected Duration: 500-1200ms
Actual Duration: 650ms
Status: ✓ PASS

Notes:
- One extra audit log created
- Performance good
- Batch operations working correctly
```

---

## 12. Future Optimization Opportunities

### Short-term (Phase 1)
1. Multi-row INSERT optimization for bulk operations
2. Connection pooling tuning
3. Query plan analysis and index optimization

### Medium-term (Phase 1b)
1. Caching layer (Redis) for frequently accessed data
2. Read replicas for reporting operations
3. Query result pagination for large datasets

### Long-term (Phase 2+)
1. Elasticsearch integration for complex searches
2. Event sourcing for audit trail optimization
3. Data warehouse/OLAP for reporting

---

## Conclusion

These query count expectations provide a baseline for Phase 1 operations. They assume:
- Proper implementation of batch methods
- Consistent soft-delete filtering
- Appropriate database indices
- No N+1 query patterns in application code

Teams should track actual query counts against these expectations and investigate any discrepancies > 20%. This enables early detection of performance regressions and N+1 patterns.

---

**Document Version:** 1.0
**Last Updated:** November 15, 2025
**Next Review:** After Phase 1 Spike Phase (week 2)
