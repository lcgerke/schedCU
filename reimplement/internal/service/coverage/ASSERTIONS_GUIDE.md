# Query Count Assertions for Coverage Calculator

## Overview

The `CoverageAssertionHelper` provides test utilities for asserting database query execution counts in coverage operations. This helps detect N+1 query patterns, performance regressions, and ensures batch query implementations work correctly.

## Key Features

- **Query Count Assertions**: Verify exact query counts for operations
- **Regression Detection**: Track expected query counts and fail if they increase
- **Error Messages**: Detailed output includes all executed queries for debugging
- **Performance**: Assertion overhead < 1% (23.98 ns/op for basic assertion)
- **N+1 Detection**: Heuristic detection for common query patterns
- **Timing Assertions**: Combined query count and performance assertions

## Basic Usage

### 1. Single Query Assertion

Assert that data loading uses exactly 1 query (batch query pattern):

```go
func TestDataLoadingBatchQuery(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    // Your operation here
    shifts, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

    if err := helper.AssertSingleQueryDataLoad(); err != nil {
        t.Fatalf("data load should use 1 query: %v", err)
    }
}
```

### 2. Specific Query Count Assertion

Assert a specific number of queries for coverage calculations:

```go
func TestCoverageCalculation(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    // Execute coverage calculation
    shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
    metrics := CalculateCoverage(shifts, requirements)

    if err := helper.AssertCoverageCalculation("CalculateCoverageMorning", 1); err != nil {
        t.Fatalf("unexpected query count: %v", err)
    }
}
```

### 3. Regression Detection

Document expected query counts and fail if they increase:

```go
func TestNoRegressionInQueryCount(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

    err := helper.AssertNoRegression(RegressionDetectionConfig{
        OperationName:    "LoadAssignmentsForScheduleVersion",
        ExpectedQueries:  1,
        MaxQueryIncrease: 0, // No increase allowed
        Description:      "Data loading should always be 1 batch query",
    })

    if err != nil {
        t.Fatalf("query count regression detected: %v", err)
    }
}
```

### 4. N+1 Query Detection

Detect patterns where queries increase unexpectedly:

```go
func TestNoNPlusOne(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
    metrics := CalculateCoverage(shifts, requirements)

    // With 100 shifts and 0 queries per shift, expect 100 * (1 + 0) = 100
    if err := helper.AssertNoNPlusOne(100, 0); err != nil {
        t.Fatalf("N+1 pattern detected: %v", err)
    }
}
```

### 5. Performance + Query Count Assertions

Assert both query count and execution time:

```go
func TestPerformanceAndQueryCount(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    result, duration, err := helper.AssertCoverageOperationTiming(
        "LoadAssignmentsForScheduleVersion",
        func() (interface{}, error) {
            return loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
        },
        1,                      // Expected 1 query
        100 * time.Millisecond, // Max duration
    )

    if err != nil {
        t.Fatalf("assertion failed: %v", err)
    }

    t.Logf("Operation completed: %d results in %v", len(result), duration)
}
```

## Advanced Usage

### Operation Wrappers

Use convenience wrappers for complete operation tracking:

```go
func TestWithOperationWrapper(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    result, err := helper.AssertDataLoaderOperation(
        "LoadAssignments",
        func() (interface{}, error) {
            return loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
        },
        1, // Expected 1 query
    )

    if err != nil {
        t.Fatalf("operation failed: %v", err)
    }

    shifts := result.([]*entity.ShiftInstance)
    // Process shifts...
}
```

### Context-Aware Operations

Supports context for cancellation and timeouts:

```go
func TestWithContext(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    result, err := helper.AssertDataLoaderOperationWithContext(
        ctx,
        "LoadAssignmentsWithContext",
        func(c context.Context) (interface{}, error) {
            return loader.LoadAssignmentsForScheduleVersion(c, scheduleVersionID)
        },
        1,
    )

    if err != nil {
        t.Fatalf("context-aware operation failed: %v", err)
    }
}
```

### Maximum Query Count (Regression)

Allow query count increase up to a tolerance:

```go
func TestMaxQueryCountWithTolerance(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

    // Expected: 1, but allow up to 1 additional query due to caching overhead
    if err := helper.AssertQueryCountLE(2); err != nil {
        t.Fatalf("query count regression: %v", err)
    }
}
```

### Debug Query Output

Log all executed queries for debugging:

```go
func TestWithQueryLogging(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

    if err := helper.AssertCoverageCalculation("LoadOp", 1); err != nil {
        t.Logf("Test failed. Executed queries:\n%s", helper.LogAllQueries())
        t.Fatalf("assertion failed: %v", err)
    }
}
```

## Error Message Examples

### Query Count Mismatch

When expected count doesn't match actual count:

```
coverage calculation "CalculateCoverageMorning": expected 1 queries, got 3

Coverage operation queries (3 total):
  1. SELECT * FROM shift_instances WHERE schedule_version_id = ? [5ms]
     Args: [<uuid>]
  2. SELECT * FROM users WHERE id IN (?) [2ms]
  3. SELECT * FROM positions [1ms]
```

### Regression Detection

When query count exceeds expected threshold:

```
REGRESSION DETECTED in "LoadAssignmentsForScheduleVersion":
  Description: Data loading should always be 1 batch query
  Expected: <= 1 queries (base: 1, tolerance: 0)
  Actual: 5
  Change: +4 queries

Coverage operation queries (5 total):
  1. SELECT * FROM shifts [1ms]
  2. SELECT * FROM users [2ms]
  3. SELECT * FROM positions [1ms]
  4. SELECT * FROM departments [1ms]
  5. SELECT * FROM roles [1ms]
```

### N+1 Pattern Detection

When queries indicate N+1 pattern:

```
potential N+1 detected: expected <= 20 queries (batch: 10 items, 1 queries per item), got 21

Queries executed (21 total):
  1. SELECT * FROM shifts [1ms]
  2. SELECT * FROM users WHERE id = ? [2ms]
  3. SELECT * FROM positions [1ms]
  ... (repeated pattern for all 10 items)
```

### Performance Violation

When operation exceeds maximum duration:

```
LoadAssignmentsForScheduleVersion exceeded max duration: expected <= 100ms, got 250ms
```

## Best Practices

### 1. Document Expected Query Counts

Use the documentation feature to maintain a record of expected counts:

```go
func init() {
    helper := NewCoverageAssertionHelper()

    helper.DocumentExpectedQueries(
        "LoadAssignmentsForScheduleVersion",
        1,
        "Single batch query using IN clause or WHERE id IN (...)",
    )

    helper.DocumentExpectedQueries(
        "CalculateCoverageMorning",
        1,
        "No additional queries after data load",
    )
}
```

### 2. Always Reset Query Count Between Tests

Ensure query tracking is isolated per test:

```go
func TestSomething(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    // Your test here
}
```

### 3. Use Meaningful Operation Names

Operation names should clearly describe what's being tested:

```go
// Good
err := helper.AssertCoverageCalculation("LoadAssignmentsForScheduleVersion", 1)

// Less clear
err := helper.AssertCoverageCalculation("Load", 1)
```

### 4. Combine Multiple Assertions

Use multiple assertions to verify different aspects:

```go
func TestComprehensive(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    helpers.StartQueryCount()
    shifts, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
    helpers.StopQueryCount()

    // Check operation succeeded
    if err != nil {
        t.Fatalf("operation failed: %v", err)
    }

    // Check query count
    if err := helper.AssertSingleQueryDataLoad(); err != nil {
        t.Fatalf("query assertion failed: %v", err)
    }

    // Check result data
    if len(shifts) == 0 {
        t.Fatal("expected shifts")
    }
}
```

### 5. Use With Integration Tests

Track real database queries:

```go
func TestWithRealDatabase(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    db, _ := setupTestDatabase()
    defer db.Close()

    helper := NewCoverageAssertionHelper()
    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

    // Real database queries are tracked
    if err := helper.AssertSingleQueryDataLoad(); err != nil {
        t.Fatalf("regression in batch query pattern: %v", err)
    }
}
```

## Performance Characteristics

Assertion overhead is minimal:

| Operation | Latency | Memory |
|-----------|---------|--------|
| AssertQueryCount | 23.98 ns | 0 B |
| AssertCoverageCalculation | 47.80 ns | 0 B |
| LogAllQueries (100 queries) | 74.247 µs | 42 KB |
| Data Load (10 items) | 345.1 ns | 248 B |
| Data Load (1000 items) | 8.312 µs | 17.5 KB |

## Regression Detection Strategy

### 1. Establish Baseline

Document expected query counts for all key operations:

```
LoadAssignmentsForScheduleVersion: 1 query (batch pattern)
CalculateCoverageMorning: 0 additional queries
SaveCoverageResults: 1 insert query
```

### 2. CI/CD Integration

Run regression detection in CI/CD:

```go
func TestNoRegressions(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    configs := []RegressionDetectionConfig{
        {
            OperationName:    "LoadAssignmentsForScheduleVersion",
            ExpectedQueries:  1,
            MaxQueryIncrease: 0,
            Description:      "Batch query pattern",
        },
        {
            OperationName:    "CalculateCoverageMorning",
            ExpectedQueries:  0,
            MaxQueryIncrease: 0,
            Description:      "No queries after data load",
        },
    }

    for _, config := range configs {
        // Test each operation and verify no regression
        if err := helper.AssertNoRegression(config); err != nil {
            t.Errorf("Regression detected: %v", err)
        }
    }
}
```

### 3. Gradual Tolerance

Allow for gradual optimization:

```go
// v1.0: Allow up to 5 queries
MaxQueryIncrease: 5

// v1.1: Reduce to 3 queries
MaxQueryIncrease: 3

// v1.2: Reduce to 1 query
MaxQueryIncrease: 1

// v2.0: No increase allowed
MaxQueryIncrease: 0
```

## Integration Examples

### With Coverage Calculator

```go
func TestCoverageCalculatorQueryCounts(t *testing.T) {
    helper := NewCoverageAssertionHelper()

    // Load data
    helpers.StartQueryCount()
    shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
    if err := helper.AssertSingleQueryDataLoad(); err != nil {
        t.Fatalf("load failed: %v", err)
    }

    // Calculate coverage (no additional queries expected)
    metrics := CalculateCoverage(shifts, requirements)
    if err := helper.AssertQueryCount(1); err != nil {
        t.Fatalf("calculation used unexpected queries: %v", err)
    }

    helpers.StopQueryCount()

    // Verify results
    if metrics == nil {
        t.Fatal("expected metrics")
    }
}
```

## Troubleshooting

### No Queries Tracked

If queries aren't being tracked:

1. Ensure `helpers.StartQueryCount()` is called before operations
2. Ensure `helpers.StopQueryCount()` is called after operations (or use defer)
3. For real database, verify database connection uses registered query counter driver
4. For mocks, manually track using `helpers.AppendQuery()`

### False Positives

If assertions fail unexpectedly:

1. Check if test database state is clean (use reset between tests)
2. Verify no concurrent test execution (use `-p 1` flag)
3. Use `helper.LogAllQueries()` to see actual queries
4. Check for caching or connection pooling side effects

### Performance Issues

If assertions are slow:

1. `LogAllQueries()` has higher overhead (for debugging)
2. Consider using only `AssertQueryCount()` for fast assertions
3. Use benchmark tests to establish baseline performance
4. Profile with `go test -cpuprofile=cpu.prof`

## API Reference

### CoverageAssertionHelper Methods

| Method | Purpose |
|--------|---------|
| `AssertQueryCount(expected)` | Verify exact query count |
| `AssertSingleQueryDataLoad()` | Verify operation used 1 query |
| `AssertCoverageCalculation(name, expected)` | Verify calculation query count |
| `AssertQueryCountLE(max)` | Verify queries <= max (regression) |
| `AssertNoNPlusOne(batchSize, qPerItem)` | Detect N+1 patterns |
| `AssertDataLoaderOperation(...)` | Wrapped operation with assertions |
| `AssertDataLoaderOperationWithContext(...)` | Context-aware operation wrapper |
| `AssertCoverageOperationTiming(...)` | Combined query count + timing |
| `LogAllQueries()` | Format executed queries for debugging |
| `GetExpectedQueryCount(name)` | Retrieve documented query count |
| `AssertNoRegression(config)` | Regression detection |
| `DocumentExpectedQueries(...)` | Document expected query counts |

---

For questions or updates, see the test file: `assertions_test.go`
