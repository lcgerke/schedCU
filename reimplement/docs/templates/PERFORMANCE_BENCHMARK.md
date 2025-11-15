# Performance Benchmark Report

**Benchmark ID**: [BENCH-001 | BENCH-002 | etc.]
**Date**: [YYYY-MM-DD]
**Benchmark Duration**: [X hours]
**Environment**: [mock | staging | production]
**Version**: [git commit hash or release version]

---

## Executive Summary

[1-2 sentence summary of benchmark findings and key performance metrics]

**Overall Performance**: [Excellent | Good | Acceptable | Needs Optimization]
**Meets Target**: [Yes | No | Partially]
**Regression Detected**: [Yes | No]

---

## Operation

### What Was Tested?

**Operation Name**: [e.g., "HTML Parsing of 6-Month Amion Schedules"]

**Description**:
[Detailed description of what operation is being benchmarked]

Example: "Parsing HTML schedules from Amion covering 6 months (approximately 180 pages) with extraction of shift details including date, position, start time, end time, and location."

### Purpose

[Why was this operation chosen for benchmarking?]

Example: "This operation is critical to the system's performance. It runs on every sync cycle and must complete within 5 seconds to avoid timeout and user-visible delays."

### Related Code/Components

- [Component 1]: `/path/to/file.go` [function name]
- [Component 2]: `/path/to/file.go` [function name]
- [Component 3]: `/path/to/file.go` [function name]

---

## Setup

### Test Conditions

#### System Configuration

- **CPU**: [number of cores, model]
- **RAM**: [amount]
- **Storage**: [type: SSD/HDD], [speed characteristics]
- **OS**: [Linux version, kernel]
- **Go Version**: [X.Y.Z]

#### Data Configuration

| Parameter | Value | Notes |
|-----------|-------|-------|
| Sample size | [N records] | [description] |
| Date range | [YYYY-MM-DD to YYYY-MM-DD] | [months covered] |
| Data complexity | [simple/medium/complex] | [explanation] |
| Cache state | [cold/warm] | [if applicable] |
| Network conditions | [local/simulated/real] | [latency/bandwidth] |

#### Test Methodology

1. **Warmup Runs**: [N iterations]
   - Purpose: [JIT compilation, cache warming]
   - Data discarded: [yes/no]

2. **Measurement Runs**: [N iterations]
   - Confidence level: [95%/99%]
   - Outlier removal: [none/drop top/bottom N%]

3. **Isolation**:
   - [ ] CPU isolated (set affinity)
   - [ ] Memory isolation (cgroups)
   - [ ] I/O isolation (separate disk)
   - [ ] Network isolation (no background traffic)

#### Profiling Tools Used

- [Tool 1]: [version], [what measured]
- [Tool 2]: [version], [what measured]
- [Tool 3]: [version], [what measured]

---

## Results

### Summary Table

| Metric | Value | Target | Status | Notes |
|--------|-------|--------|--------|-------|
| Total runtime | [X ms] | [Y ms] | [PASS/FAIL] | [notes] |
| Per-unit time | [X ms] | [Y ms] | [PASS/FAIL] | [notes] |
| Throughput | [X ops/sec] | [Y ops/sec] | [PASS/FAIL] | [notes] |
| Memory peak | [X MB] | [Y MB] | [PASS/FAIL] | [notes] |
| Memory average | [X MB] | [Y MB] | [PASS/FAIL] | [notes] |
| CPU usage | [X%] | [Y%] | [PASS/FAIL] | [notes] |
| GC pause time | [X ms] | [Y ms] | [PASS/FAIL] | [notes] |
| Allocation rate | [X MB/s] | [Y MB/s] | [PASS/FAIL] | [notes] |

### Detailed Timing Results

#### Execution Time Distribution

```
Min:        [X ms]
P25:        [X ms]
P50/Median: [X ms]
P75:        [X ms]
P95:        [X ms]
P99:        [X ms]
Max:        [X ms]

Mean:       [X ms]
StdDev:     [X ms]
CoeffVar:   [X%]
```

#### Time Breakdown (if applicable)

| Phase | Time (ms) | % of Total | Notes |
|-------|-----------|-----------|-------|
| [phase 1] | [X] | [X%] | [description] |
| [phase 2] | [X] | [X%] | [description] |
| [phase 3] | [X] | [X%] | [description] |
| [phase 4] | [X] | [X%] | [description] |
| **Total** | **[X]** | **100%** | |

### Memory Analysis

#### Memory Usage Profile

```
Baseline:     [X MB]
Peak:         [X MB]
Average:      [X MB]
Final:        [X MB]
Peak offset:  [X ms] (when in execution)
```

#### Memory Breakdown (if available)

| Component | Allocation | % of Peak | Notes |
|-----------|------------|-----------|-------|
| [component 1] | [X MB] | [X%] | [description] |
| [component 2] | [X MB] | [X%] | [description] |
| [component 3] | [X MB] | [X%] | [description] |

#### Garbage Collection

- **GC runs**: [N]
- **Total GC time**: [X ms]
- **Average pause**: [X ms]
- **Max pause**: [X ms]
- **GC efficiency**: [X%] (heap freed / GC time)

### CPU Analysis

#### CPU Profile

```
Total CPU time: [X ms]
User time:     [X ms]
System time:   [X ms]
CPU utilization: [X%] (wall clock)
```

#### Hot Functions (Top 10)

| Rank | Function | Time (ms) | % of Total | Allocations |
|------|----------|-----------|-----------|------------|
| 1 | [func name] | [X] | [X%] | [N] |
| 2 | [func name] | [X] | [X%] | [N] |
| 3 | [func name] | [X] | [X%] | [N] |
| 4 | [func name] | [X] | [X%] | [N] |
| 5 | [func name] | [X] | [X%] | [N] |

### Throughput Results

| Metric | Value |
|--------|-------|
| Operations per second | [X] |
| Records per second | [X] |
| Queries per second | [X] |
| MB/s processed | [X] |

---

## Analysis

### Performance Assessment

#### Target Compliance

**Target**: [X ms for operation]
**Actual**: [X ms]
**Status**: [PASS | FAIL]
**Margin**: [+/- X% | X ms]

[If fails target]: Explain why target was not met and implications.

#### Scaling Characteristics

How does performance scale with input size?

| Input Size | Time | Per-Unit | Scaling |
|-----------|------|---------|---------|
| [size 1] | [time] | [time/unit] | - |
| [size 2] | [time] | [time/unit] | [O(n) / O(n log n) / O(n²) / etc.] |
| [size 3] | [time] | [time/unit] | [O(n) / O(n log n) / O(n²) / etc.] |

### Bottleneck Analysis

#### Identified Bottlenecks

1. **[Bottleneck 1]**: [X% of execution time]
   - Root cause: [explanation]
   - Impact: [severity]
   - Optimization potential: [high/medium/low]

2. **[Bottleneck 2]**: [X% of execution time]
   - Root cause: [explanation]
   - Impact: [severity]
   - Optimization potential: [high/medium/low]

3. **[Bottleneck 3]**: [X% of execution time]
   - Root cause: [explanation]
   - Impact: [severity]
   - Optimization potential: [high/medium/low]

#### Critical Path

[Describe the sequence of operations that determine overall execution time]

### Comparison to Baseline

#### Regression Detection

| Metric | Previous | Current | Change | Status |
|--------|----------|---------|--------|--------|
| Total time | [X ms] | [X ms] | [+/- X%] | [PASS/FAIL] |
| Memory peak | [X MB] | [X MB] | [+/- X%] | [PASS/FAIL] |
| Throughput | [X ops/s] | [X ops/s] | [+/- X%] | [PASS/FAIL] |

**Regression threshold**: +10% from baseline
**Status**: [No regression detected | Minor regression | Major regression]

#### Improvement Tracking

Compared to [previous version/implementation]:

- **Runtime improvement**: [+/- X%]
- **Memory improvement**: [+/- X%]
- **Throughput improvement**: [+/- X%]

### Performance Profile

```
[Insert flame graph, timeline, or other visualization here]
[Or describe profile in text form if visual not available]
```

---

## Recommendations

### Optimization Opportunities

#### High Impact (Quick Wins)

1. **[Optimization 1]**
   - Current cost: [X ms / X MB]
   - Estimated improvement: [X% faster | X MB saved]
   - Effort: [X hours]
   - Risk: [low | medium | high]
   - Implementation: [brief description]

2. **[Optimization 2]**
   - Current cost: [X ms / X MB]
   - Estimated improvement: [X% faster | X MB saved]
   - Effort: [X hours]
   - Risk: [low | medium | high]
   - Implementation: [brief description]

#### Medium Impact (Planned)

1. **[Optimization 1]**
   - Current cost: [X ms / X MB]
   - Estimated improvement: [X% faster | X MB saved]
   - Effort: [X hours]
   - Risk: [low | medium | high]
   - Implementation: [brief description]

#### Low Priority (Investigate Later)

1. **[Optimization 1]**
   - Current cost: [X ms / X MB]
   - Estimated improvement: [X% faster | X MB saved]
   - Effort: [X hours]
   - Risk: [low | medium | high]
   - Notes: [why this is low priority]

### Configuration Tuning

#### Recommended Settings

```go
type PerformanceConfig struct {
    MaxConcurrency: [N],
    BatchSize: [N],
    CacheSize: [N MB],
    Timeout: [N ms],
    RetryAttempts: [N],
}
```

#### Environment Variables

```bash
export SCHED_BATCH_SIZE=[N]
export SCHED_MAX_GOROUTINES=[N]
export SCHED_CACHE_MB=[N]
export SCHED_TIMEOUT_MS=[N]
```

### Next Steps

1. [action 1]: [owner], [target date]
2. [action 2]: [owner], [target date]
3. [action 3]: [owner], [target date]

---

## Regression Detection

### Baseline Metrics

Store these as regression baselines for future benchmarks:

```yaml
benchmark_id: BENCH-001
operation: "[operation name]"
date: 2025-11-15
baseline_metrics:
  runtime_ms: [value]
  runtime_p95_ms: [value]
  memory_peak_mb: [value]
  throughput_ops_per_sec: [value]
  gc_time_ms: [value]
```

### Regression Thresholds

These are the limits for automated alerts:

| Metric | Threshold | Alert Action |
|--------|-----------|--------------|
| Runtime increase | +10% | Investigate |
| Memory increase | +15% | Investigate |
| Throughput decrease | -10% | Investigate |
| GC time increase | +20% | Investigate |

### Monitoring Integration

This benchmark should be integrated into:
- [ ] CI/CD pipeline (on each commit)
- [ ] Nightly benchmark runs
- [ ] Release verification testing
- [ ] Performance dashboard

---

## Test Artifacts

### Reproducibility

To reproduce this benchmark:

```bash
cd /path/to/schedCU
git checkout [commit hash or tag]
go test -bench=BenchmarkName -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof -benchtime=10s
```

### Generated Profiles

- **CPU profile**: `cpu.prof`
- **Memory profile**: `mem.prof`
- **Trace**: `trace.out`
- **Benchmark output**: `benchmark.txt`

### Analysis Tools

To view profiles:

```bash
go tool pprof cpu.prof
go tool pprof mem.prof
go tool trace trace.out
```

---

## Appendix: Raw Data

### Full Benchmark Output

```
[Paste complete output from benchmark run]
```

### Detailed Timing Data

```
[CSV or table with raw timing data for each iteration]
```

### Memory Profile Data

```
[Allocation details from memory profiler]
```

### Environment Details

```
uname -a: [output]
go version: [output]
go env: [relevant variables]
```

---

## References

- [Link to benchmark code]
- [Link to previous benchmark reports]
- [Link to optimization work items]
- [Link to target specification]

---

*Last Updated*: [YYYY-MM-DD]
*Benchmark Operator*: [name/agent]
*Version*: 1.0
*Next Review*: [YYYY-MM-DD]
