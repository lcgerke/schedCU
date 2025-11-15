# Integration Examples for Query Count Assertions

## Complete Test Suite Example

Here's a complete example of integrating query count assertions into your coverage tests:

### Basic Setup

```go
package coverage_test

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/schedcu/reimplement/internal/entity"
    "github.com/schedcu/reimplement/internal/service/coverage"
    "github.com/schedcu/reimplement/tests/helpers"
)

// TestFixture holds common test dependencies
type TestFixture struct {
    helper          *coverage.CoverageAssertionHelper
    loader          *coverage.CoverageDataLoader
    repository      *MockRepository
    scheduleVersion uuid.UUID
    context         context.Context
}

func setupTestFixture(t *testing.T) *TestFixture {
    return &TestFixture{
        helper:          coverage.NewCoverageAssertionHelper(),
        repository:      NewMockRepository(),
        scheduleVersion: uuid.New(),
        context:         context.Background(),
    }
}

func (f *TestFixture) setupLoader() {
    f.loader = coverage.NewCoverageDataLoader(f.repository)
}

func (f *TestFixture) startTracking() {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
}

func (f *TestFixture) stopTracking() {
    helpers.StopQueryCount()
}

func (f *TestFixture) cleanup() {
    helpers.ResetQueryCount()
}
```

### Data Loading Tests

```go
// Test single query assertion
func TestDataLoadingSingleQuery(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()
    fixture.setupLoader()

    fixture.startTracking()
    shifts, err := fixture.loader.LoadAssignmentsForScheduleVersion(
        fixture.context,
        fixture.scheduleVersion,
    )
    fixture.stopTracking()

    if err != nil {
        t.Fatalf("load failed: %v", err)
    }

    if err := fixture.helper.AssertSingleQueryDataLoad(); err != nil {
        t.Fatalf("single query assertion failed: %v", err)
        t.Logf("Executed queries:\n%s", fixture.helper.LogAllQueries())
    }

    if shifts == nil {
        t.Fatal("expected shifts, got nil")
    }
}

// Test with mock repository
func TestDataLoadingWithMockRepository(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()

    // Setup mock data
    shifts := make([]*entity.ShiftInstance, 10)
    for i := 0; i < 10; i++ {
        shifts[i] = entity.NewShiftInstance(
            fixture.scheduleVersion,
            "Morning",
            "Nurse",
            "ER",
            "Staff",
            uuid.New(),
        )
    }
    fixture.repository.shifts = shifts
    fixture.setupLoader()

    fixture.startTracking()
    loaded, err := fixture.loader.LoadAssignmentsForScheduleVersion(
        fixture.context,
        fixture.scheduleVersion,
    )
    fixture.stopTracking()

    if err != nil {
        t.Fatalf("load failed: %v", err)
    }

    // Assertions
    if err := fixture.helper.AssertCoverageCalculation("LoadMockData", 0); err != nil {
        t.Fatalf("assertion failed: %v", err)
    }

    if len(loaded) != 10 {
        t.Fatalf("expected 10 shifts, got %d", len(loaded))
    }
}

// Test regression detection
func TestDataLoadingNoRegression(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()
    fixture.setupLoader()

    fixture.startTracking()
    _, err := fixture.loader.LoadAssignmentsForScheduleVersion(
        fixture.context,
        fixture.scheduleVersion,
    )
    fixture.stopTracking()

    if err != nil {
        t.Fatalf("load failed: %v", err)
    }

    if err := fixture.helper.AssertNoRegression(
        coverage.RegressionDetectionConfig{
            OperationName:    "LoadAssignmentsForScheduleVersion",
            ExpectedQueries:  0, // Mock doesn't track
            MaxQueryIncrease: 0,
            Description:      "Data loading should use batch query pattern",
        },
    ); err != nil {
        t.Fatalf("regression detected: %v", err)
    }
}
```

### Coverage Calculation Tests

```go
// Test coverage calculation query count
func TestCoverageCalculationQueries(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()
    fixture.setupLoader()

    // Setup test data
    shifts := []*entity.ShiftInstance{
        entity.NewShiftInstance(fixture.scheduleVersion, "Morning", "Doctor", "ER", "Alice", uuid.New()),
        entity.NewShiftInstance(fixture.scheduleVersion, "Morning", "Nurse", "ER", "Bob", uuid.New()),
        entity.NewShiftInstance(fixture.scheduleVersion, "Night", "Doctor", "ICU", "Charlie", uuid.New()),
    }
    fixture.repository.shifts = shifts
    fixture.setupLoader()

    fixture.startTracking()

    // Load data
    loaded, err := fixture.loader.LoadAssignmentsForScheduleVersion(
        fixture.context,
        fixture.scheduleVersion,
    )
    if err != nil {
        t.Fatalf("load failed: %v", err)
    }

    // Should have 0 additional queries after load
    if err := fixture.helper.AssertCoverageCalculation("AfterLoad", 0); err != nil {
        t.Fatalf("query count after load: %v", err)
    }

    // Calculate coverage (in-memory operation, no queries)
    metrics := CalculateCoverage(loaded, getRequirements())

    // Total query count should still be 0 (mock)
    if err := fixture.helper.AssertCoverageCalculation("CalculateCoverageMorning", 0); err != nil {
        t.Fatalf("coverage calculation queries: %v", err)
    }

    fixture.stopTracking()

    if metrics == nil {
        t.Fatal("expected metrics")
    }
}

// Test performance and query count together
func TestCoverageCalculationPerformance(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()

    shifts := make([]*entity.ShiftInstance, 100)
    for i := 0; i < 100; i++ {
        shifts[i] = entity.NewShiftInstance(
            fixture.scheduleVersion,
            "Morning",
            "Nurse",
            "ER",
            "Staff",
            uuid.New(),
        )
    }

    fixture.startTracking()

    result, duration, err := fixture.helper.AssertCoverageOperationTiming(
        "CalculateCoverageLarge",
        func() (interface{}, error) {
            metrics := CalculateCoverage(shifts, getRequirements())
            return metrics, nil
        },
        0, // No queries (in-memory)
        100*time.Millisecond, // Should complete in < 100ms
    )

    fixture.stopTracking()

    if err != nil {
        t.Fatalf("timing assertion failed: %v", err)
    }

    if result == nil {
        t.Fatal("expected metrics")
    }

    t.Logf("Calculated coverage for 100 shifts in %v", duration)
}
```

### N+1 Detection Tests

```go
// Test that data loading doesn't have N+1 pattern
func TestNoNPlusOneInDataLoading(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()

    // Setup 50 shifts
    shifts := make([]*entity.ShiftInstance, 50)
    for i := 0; i < 50; i++ {
        shifts[i] = entity.NewShiftInstance(
            fixture.scheduleVersion,
            "Morning",
            "Nurse",
            "ER",
            "Staff",
            uuid.New(),
        )
    }
    fixture.repository.shifts = shifts
    fixture.setupLoader()

    fixture.startTracking()
    loaded, err := fixture.loader.LoadAssignmentsForScheduleVersion(
        fixture.context,
        fixture.scheduleVersion,
    )
    fixture.stopTracking()

    if err != nil {
        t.Fatalf("load failed: %v", err)
    }

    // With batch query pattern: 1 query for all 50 shifts
    // N+1 would be: 1 initial + 50 per-item queries = 51 total
    if err := fixture.helper.AssertNoNPlusOne(50, 0); err != nil {
        t.Fatalf("N+1 pattern detected: %v", err)
    }

    if len(loaded) != 50 {
        t.Fatalf("expected 50 shifts, got %d", len(loaded))
    }
}

// Test detection of actual N+1 pattern
func TestDetectsNPlusOnePattern(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()

    fixture.startTracking()

    // Simulate N+1 pattern: 1 query + 50 per-item queries
    helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM shifts"})
    for i := 0; i < 50; i++ {
        helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM users WHERE id = ?"})
    }

    fixture.stopTracking()

    // Should detect that 51 queries > expected 50 (batch + 0 per item)
    if err := fixture.helper.AssertNoNPlusOne(50, 0); err == nil {
        t.Fatal("should detect N+1 pattern")
    }
}
```

### Integration with Real Database

```go
// Integration test with test database
func TestWithTestDatabase(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping database integration test")
    }

    // Setup test database
    db, cleanup := setupTestDatabase(t)
    defer cleanup()

    fixture := setupTestFixture(t)
    defer fixture.cleanup()

    // Use real repository with database
    realRepo := NewDatabaseRepository(db)
    fixture.loader = coverage.NewCoverageDataLoader(realRepo)

    // Seed test data
    seedTestData(t, db, fixture.scheduleVersion)

    fixture.startTracking()
    shifts, err := fixture.loader.LoadAssignmentsForScheduleVersion(
        fixture.context,
        fixture.scheduleVersion,
    )
    fixture.stopTracking()

    if err != nil {
        t.Fatalf("load failed: %v", err)
    }

    // With real database, should execute exactly 1 query
    if err := fixture.helper.AssertSingleQueryDataLoad(); err != nil {
        t.Logf("Queries executed:\n%s", fixture.helper.LogAllQueries())
        t.Fatalf("single query assertion failed: %v", err)
    }

    if len(shifts) == 0 {
        t.Fatal("expected shifts from database")
    }
}

// Test with caching layer
func TestWithCachingLayer(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()

    cache := NewMemoryCache()
    cachedRepo := NewCachedRepository(
        NewMockRepository(),
        cache,
    )
    fixture.loader = coverage.NewCoverageDataLoader(cachedRepo)

    // First call: no cache hit, 1 query
    fixture.startTracking()
    shifts1, _ := fixture.loader.LoadAssignmentsForScheduleVersion(
        fixture.context,
        fixture.scheduleVersion,
    )
    fixture.stopTracking()

    firstCallQueries := helpers.GetQueryCount()

    // Second call: cache hit, 0 queries
    helpers.ResetQueryCount()
    fixture.startTracking()
    shifts2, _ := fixture.loader.LoadAssignmentsForScheduleVersion(
        fixture.context,
        fixture.scheduleVersion,
    )
    fixture.stopTracking()

    secondCallQueries := helpers.GetQueryCount()

    // Verify caching effectiveness
    if firstCallQueries <= secondCallQueries {
        t.Fatalf("caching not working: first=%d, second=%d", firstCallQueries, secondCallQueries)
    }

    // Verify data consistency
    if len(shifts1) != len(shifts2) {
        t.Fatalf("cache returned different data: %d vs %d", len(shifts1), len(shifts2))
    }
}
```

### Batch Processing Tests

```go
// Test batch processing efficiency
func TestBatchProcessingEfficiency(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()

    // Test with multiple schedule versions
    versions := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
    fixture.setupLoader()

    fixture.startTracking()

    for _, version := range versions {
        shifts, err := fixture.loader.LoadAssignmentsForScheduleVersion(
            fixture.context,
            version,
        )
        if err != nil {
            t.Fatalf("load failed: %v", err)
        }

        if shifts == nil {
            t.Fatal("expected shifts")
        }
    }

    fixture.stopTracking()

    totalQueries := helpers.GetQueryCount()

    // Should be 3 queries (1 per version) not more
    if err := fixture.helper.AssertQueryCountLE(3); err != nil {
        t.Fatalf("batch processing regression: %v", err)
    }

    t.Logf("Batch processed 3 schedule versions with %d queries", totalQueries)
}

// Test streaming/pagination
func TestPaginatedDataLoading(t *testing.T) {
    fixture := setupTestFixture(t)
    defer fixture.cleanup()

    // Create large dataset
    shifts := make([]*entity.ShiftInstance, 1000)
    for i := 0; i < 1000; i++ {
        shifts[i] = entity.NewShiftInstance(
            fixture.scheduleVersion,
            "Morning",
            "Nurse",
            "ER",
            "Staff",
            uuid.New(),
        )
    }
    fixture.repository.shifts = shifts
    fixture.setupLoader()

    // Load all at once (should be 1 query)
    fixture.startTracking()
    allShifts, _ := fixture.loader.LoadAssignmentsForScheduleVersion(
        fixture.context,
        fixture.scheduleVersion,
    )
    fixture.stopTracking()

    if err := fixture.helper.AssertSingleQueryDataLoad(); err != nil {
        t.Fatalf("batch query assertion failed: %v", err)
    }

    if len(allShifts) != 1000 {
        t.Fatalf("expected 1000 shifts, got %d", len(allShifts))
    }

    t.Logf("Efficiently loaded 1000 shifts in 1 query")
}
```

### Regression Detection CI Integration

```go
// Comprehensive regression test suite for CI/CD
func TestNoRegressionsSuite(t *testing.T) {
    testCases := []struct {
        name         string
        operation    func(*TestFixture) error
        expectedQS   int
        maxIncrease  int
        description  string
    }{
        {
            name: "LoadAssignmentsForScheduleVersion",
            operation: func(f *TestFixture) error {
                helpers.ResetQueryCount()
                f.startTracking()
                defer f.stopTracking()

                _, err := f.loader.LoadAssignmentsForScheduleVersion(f.context, f.scheduleVersion)
                return err
            },
            expectedQS:  0, // Mock
            maxIncrease: 0,
            description: "Batch query pattern must be maintained",
        },
        {
            name: "CalculateCoverageMorning",
            operation: func(f *TestFixture) error {
                helpers.ResetQueryCount()
                f.startTracking()
                defer f.stopTracking()

                shifts, _ := f.loader.LoadAssignmentsForScheduleVersion(f.context, f.scheduleVersion)
                _ = CalculateCoverage(shifts, getRequirements())
                return nil
            },
            expectedQS:  0, // Mock + in-memory
            maxIncrease: 0,
            description: "Coverage calculation should not add queries",
        },
    }

    fixture := setupTestFixture(t)
    defer fixture.cleanup()
    fixture.setupLoader()

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            if err := tc.operation(fixture); err != nil {
                t.Fatalf("operation failed: %v", err)
            }

            if err := fixture.helper.AssertNoRegression(
                coverage.RegressionDetectionConfig{
                    OperationName:    tc.name,
                    ExpectedQueries:  tc.expectedQS,
                    MaxQueryIncrease: tc.maxIncrease,
                    Description:      tc.description,
                },
            ); err != nil {
                t.Fatalf("regression test failed: %v", err)
            }
        })
    }
}

// Performance baseline test
func TestPerformanceBaseline(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping performance baseline test")
    }

    fixture := setupTestFixture(t)
    defer fixture.cleanup()
    fixture.setupLoader()

    results := make([]time.Duration, 100)

    for i := 0; i < 100; i++ {
        fixture.startTracking()

        start := time.Now()
        fixture.loader.LoadAssignmentsForScheduleVersion(
            fixture.context,
            fixture.scheduleVersion,
        )
        results[i] = time.Since(start)

        fixture.stopTracking()
    }

    // Calculate percentiles
    avg := calculateAverage(results)
    p95 := calculatePercentile(results, 95)
    p99 := calculatePercentile(results, 99)

    t.Logf("Performance baseline: avg=%v, p95=%v, p99=%v", avg, p95, p99)

    // Alert if performance degrades
    if avg > 100*time.Millisecond {
        t.Logf("WARNING: Performance degradation detected (avg=%v)", avg)
    }
}
```

## Running the Tests

```bash
# Run all coverage tests
go test -v ./internal/service/coverage/...

# Run only assertion tests
go test -v ./internal/service/coverage/ -run Assertion

# Run only regression tests
go test -v ./internal/service/coverage/ -run Regression

# Run with coverage
go test -v ./internal/service/coverage/ -cover

# Run benchmarks
go test -bench=. ./internal/service/coverage/ -benchmem

# Run integration tests only
go test -v ./internal/service/coverage/ -run Integration

# Run with verbose query logging
go test -v ./internal/service/coverage/ -run TestDataLoading
```

## Expected Output

```
=== RUN   TestDataLoadingSingleQuery
--- PASS: TestDataLoadingSingleQuery (0.00s)
=== RUN   TestCoverageCalculationQueries
--- PASS: TestCoverageCalculationQueries (0.01s)
=== RUN   TestNoNPlusOneInDataLoading
--- PASS: TestNoNPlusOneInDataLoading (0.00s)
=== RUN   TestNoRegressionsSuite
--- PASS: TestNoRegressionsSuite/LoadAssignmentsForScheduleVersion (0.00s)
--- PASS: TestNoRegressionsSuite/CalculateCoverageMorning (0.00s)

PASS
ok  	github.com/schedcu/reimplement/internal/service/coverage	0.057s
coverage: 89.3% of statements
```

---

For more information, see:
- [ASSERTIONS_GUIDE.md](./ASSERTIONS_GUIDE.md) - API reference
- [REGRESSION_DETECTION.md](./REGRESSION_DETECTION.md) - Regression strategy
- [assertions.go](./assertions.go) - Implementation
