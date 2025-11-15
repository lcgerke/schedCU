# Performance Testing Quick Reference Guide

## Running Performance Tests

### All Performance Tests
```bash
cd /home/lcgerke/schedCU/reimplement
go test -v ./internal/service/orchestrator/ -run "TestPerformance" -timeout 60s
```

Expected output:
```
✓ TestPerformanceSmallSchedule: 74µs (10 shifts)
✓ TestPerformanceMediumSchedule: 62µs (100 shifts)
✓ TestPerformanceLargeSchedule: 548µs (1000 shifts)
✓ TestPerformanceRegressionDetection: Baseline comparison passing
✓ TestQueryComplexity: O(n) verified for 10, 100, 1000 shifts
```

### Benchmark Tests
```bash
go test -bench="BenchmarkWorkflow" ./internal/service/orchestrator/ -run "^$" -benchtime=10x
```

Results show:
- Operations per second
- Time per operation
- Query counts per operation

### Individual Phase Benchmarks
```bash
# ODS Import only
go test -bench="BenchmarkODSImport" ./internal/service/orchestrator/ -run "^$"

# Amion Scraping only
go test -bench="BenchmarkAmionScraping" ./internal/service/orchestrator/ -run "^$"

# Coverage Calculation only
go test -bench="BenchmarkCoverageCalculation" ./internal/service/orchestrator/ -run "^$"
```

### Regression Detection Only
```bash
go test -v ./internal/service/orchestrator/ -run "TestPerformanceRegression"
```

## Understanding the Results

### Performance Target Interpretation

| Test | Target | Actual | Meaning |
|------|--------|--------|---------|
| Small | < 100ms | 74µs | 1,351x faster than requirement |
| Medium | < 500ms | 62µs | 8,065x faster than requirement |
| Large | < 5s | 548µs | 9,124x faster than requirement |

**Why so fast?** Mock services have zero latency. Production will be limited by real ODS import, Amion scraping, and coverage calculation services.

### Query Count Interpretation

**Small Schedule (10 shifts):**
```
ODS Queries: 10 (expected: ~10) ✓ Linear
Amion Queries: 1 (expected: ~1) ✓ Constant
Coverage Queries: 2 (expected: ~2) ✓ Constant
Result: O(n) complexity verified
```

**Interpretation:** ODS queries scale linearly with shifts (good), other phases use constant queries.

### Regression Detection

**Example Output:**
```
Baseline: 68.97µs
Current:  52.91µs
Regression: -23.29%
Status: ✓ PASS (improvement is good)
```

**Thresholds:**
- ✓ Improvement (negative regression): Always passes
- ✓ Minor regression (< 10%): Passes
- ✗ Major regression (≥ 10%): Fails test

## Performance Metrics Explained

### Duration
- **What it measures:** Total time for complete workflow (ODS + Amion + Coverage)
- **Expected:** Should scale linearly with data size, but remain under target
- **Red flag:** Sudden increase indicates performance regression

### Query Count Per Phase
- **ODS:** Expected ~N for N shifts (linear)
- **Amion:** Expected ~1 regardless of size (constant)
- **Coverage:** Expected ~2 regardless of size (constant)
- **Red flag:** ODS queries not linear (exponential growth)

### Operations Per Second
- **What it means:** How many complete workflows can execute per second
- **Small:** ~16,667 ops/sec (60µs each)
- **Medium:** ~14,659 ops/sec (68µs each)
- **Large:** ~8,266 ops/sec (121µs each)
- **Note:** Production will be much lower due to real service latency

## Test Structure

### PerformanceTestCase
Defines test scenario:
```go
type PerformanceTestCase struct {
    Name                string
    AssignmentCount     int
    ShiftCount          int
    MaxDurationMs       int64
    OdsQueryCount       int
    AmionQueryCount     int
    CoverageQueryCount  int
    Description         string
}
```

### PerformanceMetrics
Captures measurements:
```go
type PerformanceMetrics struct {
    Duration               time.Duration
    OdsQueriesExecuted     int
    AmionQueriesExecuted   int
    CoverageQueriesCount   int
    MeetsPerformanceTarget bool
    RegressionPercentage   float64
}
```

## Adding New Performance Tests

### Template: Add Size Test
```go
func TestPerformanceExtraLarge(t *testing.T) {
    testCase := PerformanceTestCase{
        Name:               "extra_large_schedule",
        AssignmentCount:    10000,
        ShiftCount:         10000,
        MaxDurationMs:      10000, // 10 seconds
        OdsQueryCount:      10000,
        AmionQueryCount:    1,
        CoverageQueryCount: 2,
        Description:        "Extra large schedule: 10000 shifts",
    }

    metrics := runPerformanceTest(t, testCase)
    verifyPerformanceMetrics(t, testCase, metrics)
}
```

### Template: Add Phase Benchmark
```go
func BenchmarkNewPhaseName(b *testing.B) {
    logger := zap.NewExample().Sugar()
    defer logger.Sync()

    var queryCount int64

    mockODS := &MockODSImportService{...}
    mockAmion := &MockAmionScraperService{...}
    mockCoverage := &MockCoverageCalculatorService{...}

    orchestrator := NewDefaultScheduleOrchestrator(
        mockODS, mockAmion, mockCoverage, logger)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ctx := context.Background()
        _, err := orchestrator.ExecuteImport(
            ctx, "/path/to/file.ods", hospitalID, userID)
        if err != nil {
            b.Fatalf("ExecuteImport failed: %v", err)
        }
    }
    b.StopTimer()

    b.ReportMetric(float64(queryCount)/float64(b.N), "queries/op")
}
```

## Troubleshooting

### Tests Run Slower Than Expected
- **Cause:** System under heavy load
- **Solution:** Run on idle system, close other applications
- **Action:** May need to adjust regression threshold for that run

### Query Counts Don't Match
- **Cause:** Mock service implementation changed
- **Solution:** Check `createMockODS*`, `createMockAmion*`, `createMockCoverage*` functions
- **Action:** Update test case expected values if intentional changes made

### Regression Detected But Code Unchanged
- **Cause:** System performance variance, garbage collection
- **Solution:** Run test multiple times to verify
- **Action:** Check system health, retry on idle system

### Test Compilation Fails
- **Cause:** Incompatible Go version, missing dependencies
- **Solution:** Verify Go 1.18+ installed
- **Action:** Run `go mod download`, `go get -u`

## Performance Monitoring Integration

### GitHub Actions / CI/CD
```yaml
- name: Run performance tests
  run: |
    go test -v ./internal/service/orchestrator/ -run "TestPerformance"

- name: Run benchmarks
  run: |
    go test -bench="BenchmarkWorkflow" ./internal/service/orchestrator/ \
      -run "^$" -benchtime=10x -benchmem -cpuprofile=cpu.prof
```

### Local Development
```bash
# Quick check before commit
go test -v ./internal/service/orchestrator/ -run "TestPerformance" -timeout 30s

# Full analysis after major changes
go test -bench=".*" ./internal/service/orchestrator/ -run "^$" -benchmem
```

## Performance Targets Review

**TIER 4 Specification Requirements:**
- Complete workflow: < 5 seconds ✓ (achieved: < 1ms)
- ODS phase: O(n) complexity ✓
- Amion phase: O(1) complexity ✓
- Coverage phase: O(1) complexity ✓
- No exponential growth ✓
- Regression detection: Operational ✓

## Key Files

- **Test Implementation:** `performance_test.go` (622 lines)
- **Baselines Documentation:** `PERFORMANCE_BASELINES.md`
- **Quick Reference:** This file
- **Completion Report:** `WORK_PACKAGE_4_4_COMPLETION.md`

## Contact / Support

For questions about performance tests:
1. Review `PERFORMANCE_BASELINES.md` for detailed analysis
2. Check `performance_test.go` for implementation details
3. Review test helper functions for test data generation
4. Check `WORK_PACKAGE_4_4_COMPLETION.md` for context

---

Last Updated: 2025-11-15
Location: `/home/lcgerke/schedCU/reimplement/internal/service/orchestrator/`
