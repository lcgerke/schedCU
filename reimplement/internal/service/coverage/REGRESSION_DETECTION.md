# Regression Detection Strategy for Coverage Calculator

## Overview

The Query Count Assertions system helps prevent query count regressions through systematic tracking and CI/CD integration. This document outlines the strategy for detecting and preventing performance regressions.

## Baseline Query Counts

Established baseline for coverage operations:

### Data Loading Operations

| Operation | Expected Queries | Reason |
|-----------|------------------|--------|
| `LoadAssignmentsForScheduleVersion` | 1 | Batch query using IN clause or JOIN |
| `LoadShiftRequirements` | 1 | Single fetch of requirements |
| `LoadStaffAvailability` | 1 | Batch fetch of availability |

### Calculation Operations

| Operation | Expected Queries | Reason |
|-----------|------------------|--------|
| `CalculateCoverageMorning` | 0 | In-memory calculation only |
| `CalculateCoverageNight` | 0 | In-memory calculation only |
| `CalculateCoverageAfternoon` | 0 | In-memory calculation only |

### Write Operations

| Operation | Expected Queries | Reason |
|-----------|------------------|--------|
| `SaveCoverageResults` | 1 | Single INSERT or UPDATE |
| `BatchSaveResults` | 1 | INSERT with multiple rows |

## Regression Detection Tests

### Per-Operation Tests

```go
func TestLoadAssignmentsRegressionDetection(t *testing.T) {
    helper := NewCoverageAssertionHelper()
    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    _, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
    if err != nil {
        t.Fatalf("operation failed: %v", err)
    }

    err = helper.AssertNoRegression(RegressionDetectionConfig{
        OperationName:    "LoadAssignmentsForScheduleVersion",
        ExpectedQueries:  1,
        MaxQueryIncrease: 0,
        Description:      "Must use batch query pattern - no N+1 allowed",
    })

    if err != nil {
        t.Fatalf("regression detected: %v", err)
    }
}
```

### Integration Test Suite

```go
func TestCoverageCalculatorRegressions(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    testCases := []struct {
        name              string
        operation         func() error
        expectedQueries   int
        maxQueryIncrease  int
        description       string
    }{
        {
            name: "LoadAssignments",
            operation: func() error {
                helpers.ResetQueryCount()
                helpers.StartQueryCount()
                defer helpers.StopQueryCount()

                _, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
                return err
            },
            expectedQueries:  1,
            maxQueryIncrease: 0,
            description:      "Batch query pattern required",
        },
        {
            name: "CalculateCoverage",
            operation: func() error {
                helpers.ResetQueryCount()
                helpers.StartQueryCount()
                defer helpers.StopQueryCount()

                shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
                _ = CalculateCoverage(shifts, requirements)
                return nil
            },
            expectedQueries:  1,
            maxQueryIncrease: 0,
            description:      "Only data loading should execute queries",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            if err := tc.operation(); err != nil {
                t.Fatalf("operation failed: %v", err)
            }

            if err := helper.AssertNoRegression(RegressionDetectionConfig{
                OperationName:    tc.name,
                ExpectedQueries:  tc.expectedQueries,
                MaxQueryIncrease: tc.maxQueryIncrease,
                Description:      tc.description,
            }); err != nil {
                t.Fatalf("regression test failed: %v", err)
            }
        })
    }
}
```

## Example Error Messages

### Scenario 1: N+1 Query Pattern Introduced

**What happened:** Someone modified the coverage calculation to load additional data for each shift instance without using a batch query.

**Error message:**

```
REGRESSION DETECTED in "LoadAssignmentsForScheduleVersion":
  Description: Must use batch query pattern - no N+1 allowed
  Expected: <= 1 queries (base: 1, tolerance: 0)
  Actual: 101
  Change: +100 queries

Coverage operation queries (101 total):
  1. SELECT * FROM shift_instances WHERE schedule_version_id = ? [2ms]
     Args: [uuid-here]
  2. SELECT * FROM users WHERE id = ? [1ms]
     Args: [user-id-1]
  3. SELECT * FROM positions WHERE id = ? [1ms]
     Args: [position-id-1]
  4. SELECT * FROM users WHERE id = ? [1ms]
     Args: [user-id-2]
  5. SELECT * FROM positions WHERE id = ? [1ms]
     Args: [position-id-2]
  ... (repeated 100 more times)
```

**Root cause:** Code similar to:
```go
// BAD - N+1 pattern
shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
for _, shift := range shifts {
    user, _ := repo.GetUserByID(ctx, shift.UserID)  // N+1!
    position, _ := repo.GetPositionByID(ctx, shift.PositionID)  // N+1!
    shift.User = user
    shift.Position = position
}
```

**Fix:** Use batch queries:
```go
// GOOD - Batch query pattern
shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
users, _ := repo.GetUsersByIDs(ctx, extractUserIDs(shifts))  // Single query
positions, _ := repo.GetPositionsByIDs(ctx, extractPositionIDs(shifts))  // Single query

// Map results to shifts (in-memory operation)
for _, shift := range shifts {
    shift.User = users[shift.UserID]
    shift.Position = positions[shift.PositionID]
}
```

### Scenario 2: Missing Index in Database

**What happened:** Database query performance degraded, causing query to execute multiple times internally.

**Error message:**

```
REGRESSION DETECTED in "LoadAssignmentsForScheduleVersion":
  Description: Batch query pattern required
  Expected: <= 1 queries (base: 1, tolerance: 0)
  Actual: 3
  Change: +2 queries

Coverage operation queries (3 total):
  1. SELECT * FROM shift_instances WHERE schedule_version_id = ? [45ms]
     Args: [uuid-here]
  2. SELECT * FROM shift_instances_idx WHERE schedule_version_id = ? [12ms]
  3. SELECT * FROM shift_instances WHERE status = 'active' AND schedule_version_id = ? [8ms]
```

**Root cause:** Database query optimizer rewrote the query due to missing index.

**Fix:** Add database index:
```sql
CREATE INDEX idx_shift_instances_schedule_version_id
ON shift_instances(schedule_version_id);
```

### Scenario 3: Caching Layer Issue

**What happened:** Caching implementation has a bug causing cache misses.

**Error message:**

```
REGRESSION DETECTED in "LoadAssignmentsForScheduleVersion":
  Description: Batch query pattern required
  Expected: <= 1 queries (base: 1, tolerance: 0)
  Actual: 11
  Change: +10 queries

Coverage operation queries (11 total):
  1. SELECT * FROM cache_keys WHERE key LIKE 'schedule_version:%' [5ms]
  2. SELECT * FROM shift_instances WHERE schedule_version_id = ? [3ms]
  3. SELECT * FROM shift_instances WHERE schedule_version_id = ? [3ms]
  4. SELECT * FROM shift_instances WHERE schedule_version_id = ? [3ms]
  ... (repeated similar queries)
```

**Root cause:** Cache key format changed but cleanup wasn't updated.

**Fix:** Update cache implementation:
```go
const cacheKeyFormat = "schedule_version:%s"

func getCacheKey(id uuid.UUID) string {
    return fmt.Sprintf(cacheKeyFormat, id.String())
}
```

## Prevention Strategies

### 1. Automated Regression Detection in CI/CD

Add to your GitHub Actions or CI workflow:

```yaml
- name: Run Regression Tests
  run: |
    go test ./internal/service/coverage/... \
      -run "TestNoRegressions|TestCoverageCalculatorRegressions" \
      -v
  env:
    # Set stricter thresholds in CI
    QUERY_COUNT_STRICT_MODE: true
```

### 2. Query Count Budget System

Implement a budget system for operations:

```go
type QueryCountBudget struct {
    Operation    string
    Budget       int
    Current      int
    LastModified time.Time
}

func CheckQueryCountBudget(budget QueryCountBudget) error {
    if budget.Current > budget.Budget {
        return fmt.Errorf(
            "query budget exceeded for %s: %d/%d (over by %d)",
            budget.Operation,
            budget.Current,
            budget.Budget,
            budget.Current-budget.Budget,
        )
    }
    return nil
}
```

### 3. Gradual Optimization Plan

Establish improvement targets:

```
Phase 1 (v1.0): Baseline established
  LoadAssignmentsForScheduleVersion: 1 query (current)
  CalculateCoverageMorning: 0 additional queries
  Total: 1 query per calculation

Phase 2 (v1.1): Minor optimizations
  Target: Keep at 1 query (prevent regressions)
  Tolerance: ±0 queries

Phase 3 (v1.2): Caching improvements
  Target: Reduce average to 0.5 queries (with caching)
  Tolerance: ±0 queries for non-cached case

Phase 4 (v2.0): Batch processing
  Target: 1 query for 1000 shifts
  Tolerance: ±0 queries
```

### 4. Code Review Checklist

Add to pull request review template:

```markdown
## Performance Checklist
- [ ] No N+1 query patterns detected
- [ ] Query count assertions still pass
- [ ] No new database calls in loops
- [ ] Batch operations use IN clauses or JOINs
- [ ] Database indexes exist for filtered queries
- [ ] Caching strategy is documented
- [ ] Performance impact < 5% (if applicable)
```

## Monitoring and Alerting

### Query Count Metrics

Track query counts over time:

```go
type QueryCountMetric struct {
    Operation     string
    Timestamp     time.Time
    QueryCount    int
    Duration      time.Duration
    P50, P95, P99 int // Percentiles
}

func RecordMetric(operation string, count int, duration time.Duration) {
    metric := QueryCountMetric{
        Operation:  operation,
        Timestamp:  time.Now(),
        QueryCount: count,
        Duration:   duration,
    }
    // Send to monitoring system (Prometheus, DataDog, etc.)
}
```

### Alert Thresholds

Set up alerts for:

1. **Query count increase**: Alert if any operation exceeds baseline by > 10%
2. **Performance degradation**: Alert if duration increases by > 25%
3. **Test failures**: Alert if regression detection tests fail in CI
4. **Trend analysis**: Alert if rolling average increases over 7 days

## Maintenance Schedule

### Weekly
- Review test results
- Check for patterns in query count changes
- Verify all regression tests pass

### Monthly
- Analyze trends
- Update baseline if intentional improvements made
- Update documentation
- Review and update tolerance levels

### Quarterly
- Full regression audit
- Performance benchmarking
- Index effectiveness review
- Caching strategy evaluation

## Integration with CICD Pipeline

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running query count regression tests..."
go test ./internal/service/coverage/... \
    -run "TestNoRegressions" \
    -v

if [ $? -ne 0 ]; then
    echo "ERROR: Regression tests failed. Query counts may have changed."
    exit 1
fi
```

### GitHub Actions

```yaml
name: Query Count Regression Tests

on: [pull_request, push]

jobs:
  regression-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.19

      - name: Run Regression Tests
        run: |
          go test ./internal/service/coverage/... \
            -run "Regression|NoNPlusOne" \
            -v \
            -timeout 10m

      - name: Report Results
        if: always()
        uses: actions/github-script@v6
        with:
          script: |
            const result = ${{ job.status }};
            console.log(`Regression tests: ${result}`);
```

## Troubleshooting Regressions

### Step 1: Confirm Regression

```bash
# Run the specific failing test
go test -v ./internal/service/coverage/ \
    -run TestLoadAssignmentsRegressionDetection
```

### Step 2: Get Query Details

```go
// In your test
helper.LogAllQueries()

// Output shows exact queries being executed
```

### Step 3: Analyze Changes

Compare current behavior to baseline:

1. Did a recent commit add new database calls?
2. Did a library update change query behavior?
3. Were indexes added/removed?
4. Has the database schema changed?

### Step 4: Fix and Verify

Make necessary changes and re-run tests:

```bash
go test ./internal/service/coverage/... -run Regression -count=5
```

## Documentation

Update when baselines change:

```go
// Document why baseline changed
// Example: v1.1 migration added caching layer
helper.DocumentExpectedQueries(
    "LoadAssignmentsForScheduleVersion",
    1,
    "Single batch query - unchanged by v1.1 caching (cache is transparent)",
)
```

---

For more information, see:
- [ASSERTIONS_GUIDE.md](./ASSERTIONS_GUIDE.md) - Usage guide
- [assertions.go](./assertions.go) - Implementation
- [assertions_test.go](./assertions_test.go) - Examples
