# Query Counter Framework - Practical Examples

This document provides real-world examples of using the Query Counter Framework in your test suite.

## Example 1: Simple Repository Test

Test a single query operation:

```go
package repository_test

import (
    "context"
    "testing"
    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
    "github.com/schedcu/reimplement/tests/helpers"
)

func TestScheduleRepositoryGetByID(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    // Setup
    scheduleID := uuid.New()
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Execute
    repo := repository.NewScheduleRepository(db)
    schedule, err := repo.GetByID(ctx, scheduleID)

    // Assert
    require.NoError(t, err)
    require.NotNil(t, schedule)

    // Verify query count
    count := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCount(1, count),
        "GetByID should execute exactly 1 query")
}
```

## Example 2: Batch Operation with N+1 Detection

Test a batch operation while preventing N+1:

```go
func TestScheduleRepositoryListByDateRange(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    // Setup: Insert multiple schedules
    const scheduleCount = 50
    for i := 0; i < scheduleCount; i++ {
        // ... insert schedules ...
    }

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Execute: Batch fetch
    repo := repository.NewScheduleRepository(db)
    schedules, err := repo.ListByDateRange(ctx, startDate, endDate)

    // Assert: Verify no N+1 pattern
    require.NoError(t, err)
    require.Len(t, schedules, scheduleCount)

    // Should be 2 queries: one for schedules, one for assignments
    count := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertNoNPlusOne(scheduleCount, 1),
        "ListByDateRange should use batch loading, not N+1. Log:\n%s",
        helpers.LogQueries())
}
```

## Example 3: Service Layer Testing

Test business logic with query count assertions:

```go
func TestCoverageCalculatorResolveCoverage(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    // Setup: Create test data
    shifts := createTestShifts(t, db)
    assignments := createTestAssignments(t, db, shifts)

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Execute
    calculator := service.NewCoverageCalculator(db)
    coverage, err := calculator.ResolveCoverage(ctx, assignments)

    // Assert
    require.NoError(t, err)
    require.NotNil(t, coverage)

    // Verify query efficiency
    count := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCountLE(3),
        "ResolveCoverage should be efficient (max 3 queries)")
}
```

## Example 4: Multi-Step Operation Testing

Test a complex operation that involves multiple steps:

```go
func TestScheduleServiceCreateWithValidation(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    req := &CreateScheduleRequest{
        Name:        "Winter 2024",
        HospitalID:  hospitalID,
        StartDate:   time.Now(),
        EndDate:     time.Now().AddDate(0, 0, 30),
    }

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Execute
    svc := service.NewScheduleService(db)
    schedule, validationErrors, err := svc.CreateWithValidation(ctx, req)

    // Assert
    require.NoError(t, err)
    require.Empty(t, validationErrors)
    require.NotNil(t, schedule)

    // Verify: Create (1) + Validate (1-2) = 2-3 queries
    count := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCountLE(3),
        "CreateWithValidation should use at most 3 queries")
}
```

## Example 5: Regression Detection in CI

Use query assertions to detect performance regressions:

```go
// Run this in your CI pipeline with -count=10 to ensure consistency
func TestPerformanceRegression_CoverageCalculation(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    // Load test data
    assignments := loadTestAssignments(t, db)

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Execute the operation multiple times
    calculator := service.NewCoverageCalculator(db)
    for i := 0; i < 10; i++ {
        _, err := calculator.ResolveCoverage(ctx, assignments)
        require.NoError(t, err)
    }

    // Assert: Should not increase query count
    // Maximum allowed: 10 iterations * 2 queries per iteration = 20
    require.NoError(t, helpers.AssertQueryCountLE(20),
        "Query count regression detected. Expected <= 20, got %d\n%s",
        helpers.GetQueryCount(),
        helpers.LogQueries())
}
```

## Example 6: Concurrent Operation Testing

Test thread-safe database operations:

```go
func TestConcurrentScheduleAssignment(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    const workerCount = 10
    const assignmentsPerWorker = 5

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Execute: Concurrent assignments
    var wg sync.WaitGroup
    errors := make(chan error, workerCount)

    svc := service.NewScheduleService(db)

    for w := 0; w < workerCount; w++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for i := 0; i < assignmentsPerWorker; i++ {
                _, err := svc.AssignStaffMember(ctx, shiftID, staffID)
                if err != nil {
                    errors <- err
                }
            }
        }(w)
    }

    wg.Wait()
    close(errors)

    // Verify no errors
    for err := range errors {
        require.NoError(t, err)
    }

    // Verify query count
    // Each assignment: 1 query
    // Total: workerCount * assignmentsPerWorker = 50
    totalQueries := workerCount * assignmentsPerWorker
    count := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCountLE(totalQueries + 5),
        "Expected around %d queries (with tolerance), got %d",
        totalQueries, count)
}
```

## Example 7: Caching Validation

Verify that caching prevents database queries:

```go
func TestCacheEffectiveness(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    // Create a cached repository
    cache := newTestCache()
    repo := repository.NewScheduleRepository(db)
    cached := repository.NewCachedRepository(repo, cache)

    scheduleID := uuid.New()
    insertTestSchedule(t, db, scheduleID)

    // First access: hits database
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
    schedule1, err := cached.GetByID(ctx, scheduleID)
    require.NoError(t, err)
    firstAccessQueries := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCount(1, firstAccessQueries))

    // Second access: hits cache
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
    schedule2, err := cached.GetByID(ctx, scheduleID)
    require.NoError(t, err)
    cachedAccessQueries := helpers.GetQueryCount()
    require.Equal(t, 0, cachedAccessQueries, "Cached access should not hit database")

    // Verify data is identical
    require.Equal(t, schedule1.ID, schedule2.ID)
}
```

## Example 8: Error Case with Query Logging

Debug test failures with detailed query logs:

```go
func TestScheduleRepositoryGetByIDWithErrorLogging(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    scheduleID := uuid.New()
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    repo := repository.NewScheduleRepository(db)
    schedule, err := repo.GetByID(ctx, scheduleID)

    // Log queries when assertion fails
    if err := helpers.AssertQueryCount(1, helpers.GetQueryCount()); err != nil {
        t.Logf("Assertion failed: %v\n", err)
        t.Logf("Query log:\n%s\n", helpers.LogQueries())
        t.Fail()
    }

    require.NoError(t, err)
    require.NotNil(t, schedule)
}
```

## Example 9: Transaction Testing

Test transactional behavior with query counting:

```go
func TestTransactionalScheduleCreation(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    // Begin transaction
    tx, err := db.BeginTx(ctx, nil)
    require.NoError(t, err)
    defer tx.Rollback()

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Execute within transaction
    repo := repository.NewScheduleRepository(tx)
    schedule := &entity.Schedule{
        ID:        uuid.New(),
        StartDate: time.Now(),
        EndDate:   time.Now().AddDate(0, 0, 30),
    }

    err = repo.Create(ctx, schedule)
    require.NoError(t, err)

    // Verify query count
    count := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCount(1, count),
        "Create within transaction should be 1 query")

    // Verify data exists in transaction
    retrieved, err := repo.GetByID(ctx, schedule.ID)
    require.NoError(t, err)
    require.NotNil(t, retrieved)

    // Transaction queries don't include COMMIT yet
    finalCount := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCount(2, finalCount),
        "Should have Create + GetByID = 2 queries")
}
```

## Example 10: Performance Baseline Establishment

Establish performance baselines for critical paths:

```go
func TestPerformanceBaseline_BulkScheduleImport(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    const scheduleCount = 100
    const shiftsPerSchedule = 50

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Execute bulk import
    svc := service.NewImportService(db)
    err := svc.ImportSchedules(ctx, createTestScheduleData(scheduleCount, shiftsPerSchedule))
    require.NoError(t, err)

    count := helpers.GetQueryCount()
    expectedMax := scheduleCount + (scheduleCount * shiftsPerSchedule)

    // Log baseline for future reference
    t.Logf("PERFORMANCE BASELINE\n")
    t.Logf("  Schedules: %d\n", scheduleCount)
    t.Logf("  Shifts per schedule: %d\n", shiftsPerSchedule)
    t.Logf("  Total queries: %d\n", count)
    t.Logf("  Theoretical minimum: %d (one per entity)\n", expectedMax)
    t.Logf("  Actual vs theoretical: %.2f%%\n", float64(count)/float64(expectedMax)*100)

    // Ensure we're not doing significantly worse than theoretical minimum
    // Allow 10% overhead for transaction boundaries
    require.NoError(t, helpers.AssertQueryCountLE(int(float64(expectedMax)*1.1)),
        "Bulk import query count within acceptable range")
}
```

## Example 11: Comparing Two Implementations

Compare query efficiency between two implementations:

```go
func TestQueryEfficiency_OldVsNewImplementation(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    defer db.Close()

    assignments := createTestAssignments(t, db, 100)

    // Test old implementation
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
    oldCoverage, err := service.OldCoverageCalculator{DB: db}.Resolve(ctx, assignments)
    require.NoError(t, err)
    oldQueryCount := helpers.GetQueryCount()

    // Test new implementation
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
    newCoverage, err := service.NewCoverageCalculator{DB: db}.Resolve(ctx, assignments)
    require.NoError(t, err)
    newQueryCount := helpers.GetQueryCount()

    // Results should be identical
    require.Equal(t, oldCoverage, newCoverage)

    // New implementation should be more efficient
    t.Logf("OLD implementation: %d queries\n", oldQueryCount)
    t.Logf("NEW implementation: %d queries\n", newQueryCount)
    t.Logf("Improvement: %.1f%%\n", float64(oldQueryCount-newQueryCount)/float64(oldQueryCount)*100)

    require.Less(t, newQueryCount, oldQueryCount, "New implementation should be more efficient")
}
```

## Example 12: Suite-Level Setup and Teardown

Manage query counting across test suites:

```go
type ScheduleRepositoryTestSuite struct {
    suite.Suite
    db   *sql.DB
}

func (suite *ScheduleRepositoryTestSuite) SetupSuite() {
    // One-time setup for all tests
    suite.db = setupTestDB(suite.T())
}

func (suite *ScheduleRepositoryTestSuite) TearDownSuite() {
    suite.db.Close()
}

func (suite *ScheduleRepositoryTestSuite) SetupTest() {
    // Per-test setup
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
}

func (suite *ScheduleRepositoryTestSuite) TestGetByID() {
    repo := repository.NewScheduleRepository(suite.db)
    schedule, err := repo.GetByID(context.Background(), uuid.New())

    suite.NoError(err)
    suite.NoError(helpers.AssertQueryCount(1, helpers.GetQueryCount()))
}

func (suite *ScheduleRepositoryTestSuite) TestList() {
    repo := repository.NewScheduleRepository(suite.db)
    schedules, err := repo.List(context.Background())

    suite.NoError(err)
    // Allow for pagination - should be at most 2 queries
    suite.NoError(helpers.AssertQueryCountLE(2))
}

func TestScheduleRepositoryTestSuite(t *testing.T) {
    suite.Run(t, new(ScheduleRepositoryTestSuite))
}
```

## Example 13: Custom Assertion Helpers

Create custom helpers for your project:

```go
// In your test helpers package
func AssertSingleQuery(t *testing.T, expectedSQL string) {
    count := helpers.GetQueryCount()
    if err := helpers.AssertQueryCount(1, count); err != nil {
        t.Logf("Query count mismatch: %v", err)
        t.Fail()
    }

    queries := helpers.GetQueries()
    if len(queries) != 1 {
        t.Fatalf("Expected 1 query, got %d", len(queries))
    }

    if !strings.Contains(queries[0].SQL, expectedSQL) {
        t.Errorf("Expected query containing '%s', got '%s'",
            expectedSQL, queries[0].SQL)
    }
}

func AssertNoAdditionalQueries(t *testing.T) error {
    return helpers.AssertQueryCount(0, helpers.GetQueryCount())
}

// Usage in test
func TestWithCustomAssertions(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    repo.GetByID(ctx, id)

    AssertSingleQuery(t, "SELECT")
}
```

## Key Takeaways

1. **Always reset** before each test: `helpers.ResetQueryCount()`
2. **Start counting** after setup: `helpers.StartQueryCount()`
3. **Use exact counts** for critical paths: `AssertQueryCount(expected, actual)`
4. **Use max counts** for regression detection: `AssertQueryCountLE(max)`
5. **Detect N+1** patterns: `AssertNoNPlusOne(batchSize, queriesPerItem)`
6. **Log queries** when debugging: `helpers.LogQueries()`

For more details, see `QUERY_COUNTER_USAGE.md`.
