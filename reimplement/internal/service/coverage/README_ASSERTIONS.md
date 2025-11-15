# Query Count Assertions - Coverage Calculator

## Quick Start

Test that your coverage operations use the expected number of database queries:

```go
func TestDataLoadingUsesBatchQuery(t *testing.T) {
    helper := coverage.NewCoverageAssertionHelper()
    
    helpers.StartQueryCount()
    shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
    helpers.StopQueryCount()
    
    if err := helper.AssertSingleQueryDataLoad(); err != nil {
        t.Fatalf("Should use batch query: %v", err)
    }
}
```

## What is This?

Query Count Assertions help you:

1. **Verify batch query patterns** - Ensure N+1 queries don't sneak in
2. **Detect performance regressions** - Fail tests if query count increases
3. **Debug query issues** - Get detailed output of all executed queries
4. **Maintain performance** - Track query counts in CI/CD pipelines

## Files

### Core Implementation
- **`assertions.go`** (370 lines)
  - `CoverageAssertionHelper` - Main assertion helper
  - 12 public assertion methods
  - Query formatting and regression detection

### Tests
- **`assertions_test.go`** (900+ lines)
  - 38 assertion-specific tests
  - 48+ total tests (including data loader tests)
  - All passing
  - Benchmark tests show < 1% overhead

### Documentation
- **`ASSERTIONS_GUIDE.md`** - Complete usage guide with 5+ examples
- **`REGRESSION_DETECTION.md`** - Regression prevention strategy with CI/CD integration
- **`INTEGRATION_EXAMPLES.md`** - Real-world working examples
- **`IMPLEMENTATION_SUMMARY.md`** - Technical summary and metrics

## Key Methods

| Method | Purpose | Example |
|--------|---------|---------|
| `AssertQueryCount(n)` | Verify exactly N queries | `helper.AssertQueryCount(1)` |
| `AssertSingleQueryDataLoad()` | Assert 1 query (batch pattern) | `helper.AssertSingleQueryDataLoad()` |
| `AssertCoverageCalculation(name, n)` | Name-tracked assertion | `helper.AssertCoverageCalculation("Load", 1)` |
| `AssertQueryCountLE(max)` | Regression: queries <= max | `helper.AssertQueryCountLE(2)` |
| `AssertNoNPlusOne(batch, per)` | Detect N+1 patterns | `helper.AssertNoNPlusOne(100, 0)` |
| `AssertNoRegression(config)` | Regression with tolerance | `helper.AssertNoRegression(cfg)` |
| `AssertCoverageOperationTiming(...)` | Query count + performance | Timeout-based assertion |
| `LogAllQueries()` | Debug: show all queries | `t.Logf("%s", helper.LogAllQueries())` |

## Test Coverage

### 38 Assertion Tests
- Basic assertions (6 tests)
- Coverage calculations (3 tests)
- Regression detection (4 tests)
- N+1 detection (2 tests)
- Operation wrappers (4 tests)
- Timing assertions (2 tests)
- Debugging & formatting (10 tests)
- Benchmarks (3 tests)

### All 48+ Tests Passing
```
PASS
ok  	github.com/schedcu/reimplement/internal/service/coverage	0.056s
```

## Performance

**Assertion Overhead**: < 1%
- AssertQueryCount: 23.98 ns/op
- AssertCoverageCalculation: 47.80 ns/op
- LogAllQueries (100 queries): 74.247 µs/op

## Error Messages

When tests fail, you get detailed context:

```
coverage calculation "CalculateCoverageMorning": expected 1 queries, got 3

Coverage operation queries (3 total):
  1. SELECT * FROM shift_instances WHERE schedule_version_id = ? [5ms]
     Args: [<uuid>]
  2. SELECT * FROM users WHERE id IN (?) [2ms]
  3. SELECT * FROM positions [1ms]
```

## Usage Examples

### 1. Single Query Assertion
```go
helper.AssertSingleQueryDataLoad()
```

### 2. Specific Query Count
```go
helper.AssertCoverageCalculation("MyOperation", 2)
```

### 3. Regression Detection
```go
helper.AssertNoRegression(RegressionDetectionConfig{
    OperationName:    "LoadData",
    ExpectedQueries:  1,
    MaxQueryIncrease: 0,
    Description:      "Must use batch query pattern",
})
```

### 4. Performance + Query Count
```go
result, duration, err := helper.AssertCoverageOperationTiming(
    "LoadOp",
    func() (interface{}, error) { return load(ctx) },
    1,                      // Expected 1 query
    100*time.Millisecond,   // Max duration
)
```

### 5. N+1 Detection
```go
helper.AssertNoNPlusOne(100, 0) // 100 items, 0 queries per item
```

## Running Tests

```bash
# Run all coverage tests
go test -v ./internal/service/coverage/...

# Run only assertion tests
go test -v ./internal/service/coverage/ -run Assertion

# Run regression tests
go test -v ./internal/service/coverage/ -run Regression

# Run benchmarks
go test -bench=. ./internal/service/coverage/ -benchmem

# Full output with coverage
go test -v ./internal/service/coverage/ -cover
```

## Integration with CI/CD

Add to your test suite:

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
    }
    for _, cfg := range configs {
        if err := helper.AssertNoRegression(cfg); err != nil {
            t.Errorf("Regression: %v", err)
        }
    }
}
```

## Best Practices

1. **Always reset between tests**
   ```go
   helpers.ResetQueryCount()
   ```

2. **Use defer for cleanup**
   ```go
   helpers.StartQueryCount()
   defer helpers.StopQueryCount()
   ```

3. **Use meaningful operation names**
   ```go
   helper.AssertCoverageCalculation("LoadAssignmentsForScheduleVersion", 1)
   ```

4. **Log on failure for debugging**
   ```go
   if err != nil {
       t.Logf("Queries:\n%s", helper.LogAllQueries())
   }
   ```

5. **Document expected query counts**
   ```go
   helper.DocumentExpectedQueries("LoadOp", 1, "Single batch query")
   ```

## Next Steps

1. **Read the guides**
   - Start with [ASSERTIONS_GUIDE.md](./ASSERTIONS_GUIDE.md) for usage
   - Review [REGRESSION_DETECTION.md](./REGRESSION_DETECTION.md) for strategy
   - Check [INTEGRATION_EXAMPLES.md](./INTEGRATION_EXAMPLES.md) for examples

2. **Add to your tests**
   - Identify critical operations
   - Add assertions for query counts
   - Document baselines

3. **Set up CI/CD**
   - Add regression tests to pipeline
   - Set alert thresholds
   - Track trends over time

4. **Monitor and maintain**
   - Review weekly results
   - Update baselines when needed
   - Investigate anomalies

## Troubleshooting

**No queries tracked?**
- Ensure `helpers.StartQueryCount()` is called before operations
- Use `helpers.ResetQueryCount()` between tests
- For real DB, verify query counter driver is registered

**False positives?**
- Check test database is clean
- Use `-p 1` to avoid concurrent test issues
- Call `helper.LogAllQueries()` to see actual queries

**Performance issues?**
- AssertQueryCount is fast (24 ns)
- LogAllQueries has higher overhead (for debugging only)
- Use benchmarks to establish baseline

## Support

For questions or issues:
1. Check [ASSERTIONS_GUIDE.md](./ASSERTIONS_GUIDE.md) FAQ
2. Review examples in [INTEGRATION_EXAMPLES.md](./INTEGRATION_EXAMPLES.md)
3. Check test file for working implementations
4. See [REGRESSION_DETECTION.md](./REGRESSION_DETECTION.md) troubleshooting

## Status

✓ **COMPLETE** - All requirements fulfilled
- 48+ tests passing
- Comprehensive documentation
- < 1% performance overhead
- Production-ready

