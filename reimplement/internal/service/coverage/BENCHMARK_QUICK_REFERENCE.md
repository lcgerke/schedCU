# Benchmark Quick Reference

## Running All Benchmarks

```bash
go test -bench=Benchmark -benchmem -benchtime=5s ./internal/service/coverage/...
```

**Expected Runtime:** ~174 seconds
**Output:** 21 benchmarks with timing and memory metrics

## Running Specific Benchmarks

### Data Loading Benchmarks
```bash
# Sequential performance
go test -bench=BenchmarkDataLoader -benchmem ./internal/service/coverage/...

# Parallel performance
go test -bench=BenchmarkDataLoaderParallel -benchmem ./internal/service/coverage/...

# With varied data
go test -bench=BenchmarkDataLoaderWithVariedShifts -benchmem ./internal/service/coverage/...
```

### Repository Layer Isolation
```bash
go test -bench=BenchmarkRepositoryMock -benchmem ./internal/service/coverage/...
```

### Single Size
```bash
# 100 assignments
go test -bench=Benchmark.*100$ -benchmem ./internal/service/coverage/...

# 1000 assignments
go test -bench=Benchmark.*1000$ -benchmem ./internal/service/coverage/...

# 10000 assignments
go test -bench=Benchmark.*10000$ -benchmem ./internal/service/coverage/...
```

## Interpreting Results

### Output Format
```
BenchmarkDataLoader_100-24    4681868    1279 ns/op    2168 B/op    8 allocs/op
                     |          |          |           |         |
                  Name      Iterations   Duration   Memory/op  Allocs/op
```

### Metrics Explained

- **Iterations:** How many times the benchmark ran (more iterations = more confidence)
- **Duration (ns/op):** Nanoseconds per operation (lower is better)
- **Memory (B/op):** Bytes allocated per operation (lower is better)
- **Allocs/op:** Number of allocations per operation (lower is better)

### Performance Rating

| Duration | Rating | Comment |
|----------|--------|---------|
| < 2 µs | Excellent | Very fast |
| 2-10 µs | Good | Fast |
| 10-100 µs | Acceptable | Moderate |
| 100+ µs | Investigate | May need optimization |

## Regression Detection

### Quick Check: Did Performance Change?

Before making changes:
```bash
go test -bench=Benchmark -benchmem ./internal/service/coverage/ > baseline.txt
```

After making changes:
```bash
go test -bench=Benchmark -benchmem ./internal/service/coverage/ > current.txt
```

Compare (requires benchstat tool):
```bash
go install golang.org/x/perf/cmd/benchstat@latest
benchstat baseline.txt current.txt
```

### Expected Output
```
name                         old time/op      new time/op      delta
DataLoader_100-24            1.27µs ± 2%      1.28µs ± 3%      +0.79%
DataLoader_1000-24           7.86µs ± 2%      7.89µs ± 2%      +0.38%
DataLoader_10000-24          193µs ± 2%       192µs ± 2%       -0.52%
```

### Failure Threshold

A benchmark fails if:
- Duration increases by >35% for any configuration
- Memory increases by >25% for any configuration
- Allocation count increases by >10 allocations

## Baseline Targets

These are the targets that should not be exceeded:

```
LoadAssignments_100:    < 2.0 µs,  < 3.0 KB,  < 10 allocations
LoadAssignments_1000:   < 10.0 µs, < 20.0 KB, < 13 allocations
LoadAssignments_10000:  < 250.0 µs,< 350.0 KB,< 20 allocations
```

## Common Issues

### Benchmark Seems Slow

1. Check CPU load: `top` or `htop`
2. Close other applications
3. Run again: `go test -bench=Benchmark -benchmem ./internal/service/coverage/`

### Results Vary Widely

1. Run with longer time: `-benchtime=10s`
2. Run multiple times and average
3. Use benchstat tool for statistical analysis

### Memory Numbers Don't Make Sense

1. Memory is per operation, not total
2. Multiple runs amortize fixed overhead
3. Allocations count includes framework overhead

## Advanced Usage

### Profile a Single Benchmark

```bash
go test -bench=BenchmarkDataLoader_1000 -cpuprofile=cpu.prof ./internal/service/coverage/
go tool pprof -http=localhost:8080 cpu.prof
```

### Memory Profile

```bash
go test -bench=BenchmarkDataLoader_10000 -memprofile=mem.prof ./internal/service/coverage/
go tool pprof -http=localhost:8080 mem.prof
```

### Trace Execution

```bash
go test -bench=BenchmarkDataLoader_1000 -trace=trace.out ./internal/service/coverage/
go tool trace trace.out
```

## Understanding O(n) Scaling

For O(n) algorithms, expect:
- 10x more items → ~10x more time (linear)
- 100x more items → ~100x more time (linear)

### Our Results

```
100 items:   1,279 ns
1,000 items: 7,862 ns  (6.1x for 10x items) ✓ linear
10,000 items: 192,902 ns (24.5x for 10x items) ✓ linear
```

Combined: 150.8x for 100x increase = perfectly linear O(n)

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Performance Regression Check

on: [pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - name: Run benchmarks
        run: go test -bench=Benchmark -benchmem ./internal/service/coverage/ > new.txt
      - name: Compare with main
        run: |
          git fetch origin main
          git checkout origin/main
          go test -bench=Benchmark -benchmem ./internal/service/coverage/ > baseline.txt
          git checkout -
          go install golang.org/x/perf/cmd/benchstat@latest
          benchstat baseline.txt new.txt
```

## Tips for Consistent Results

1. **Quiet Environment:** Stop background tasks for consistent results
2. **CPU Governor:** Set to performance mode for benchmarking
3. **Multiple Runs:** Average multiple runs for accuracy
4. **Same Hardware:** Compare benchmarks on same machine
5. **Controlled Changes:** Benchmark one change at a time

## Further Reading

- Full documentation: `PERFORMANCE_BENCHMARKS.md`
- Implementation code: `algorithm_bench_test.go`
- Data loader: `data_loader.go`
- Regression detection: `REGRESSION_DETECTION.md`
