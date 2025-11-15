# EXAMPLE: Spike 1 - HTML Parsing Library Results

**IMPORTANT**: This is a filled-in example template showing how to complete the SPIKE_1_HTML_PARSING_RESULTS.md template. Use this as reference when filling your own spike results.

---

# Spike 1: HTML Parsing Library Results

**Spike ID**: spike1
**Date Completed**: 2025-11-15
**Duration**: 4 hours
**Status**: success

---

## Executive Summary

goquery successfully parses Amion HTML with 100% accuracy and 1ms performance for 6-month schedules. Recommend proceeding with goquery implementation in Phase 1.8. Zero timeline impact—performance exceeds targets by 5000x.

**Recommended Library**: goquery (PuerkitoDio/goquery)
**Timeline Impact**: +0 weeks
**Risk Level**: low

---

## Library Evaluation

### Candidates Evaluated

#### Option 1: goquery (SELECTED)
- **Language/Package**: github.com/PuerkitoDio/goquery v1.8.1
- **Version**: 1.8.1
- **Strengths**:
  - jQuery-like API—familiar to developers with web experience
  - Pure Go implementation with no external dependencies
  - Parses HTML into proper DOM tree (resilient to whitespace/formatting changes)
  - Excellent error recovery for malformed HTML
  - 10K+ GitHub stars, actively maintained
- **Weaknesses**:
  - Cannot handle JavaScript-rendered content (Amion uses server-side HTML, not an issue)
  - Slower than regex for simple patterns (not a performance bottleneck)
- **Production Readiness**: yes
- **Maintenance Status**: active (last update 2 months ago)

#### Option 2: Chromedp
- **Language/Package**: github.com/chromedp/chromedp v0.9.1
- **Version**: 0.9.1
- **Strengths**:
  - Can handle JavaScript-rendered content
  - Mimics real browser interaction
  - Superior for complex modern web apps
- **Weaknesses**:
  - Requires running Chrome/Chromium process (infrastructure overhead)
  - Much slower: ~500ms per page vs 1ms with goquery
  - Adds 50-100 MB memory overhead per instance
  - Overkill for server-side HTML (Amion isn't a SPA)
- **Production Readiness**: yes
- **Maintenance Status**: active

#### Option 3: Regex + strings package
- **Language/Package**: Go stdlib (no external dependency)
- **Version**: built-in
- **Strengths**:
  - Zero dependencies
  - Fastest possible for known patterns
  - Minimal memory footprint
- **Weaknesses**:
  - Brittle: breaks with any Amion HTML structure change (happened 3x in v1)
  - Hard to maintain—regex patterns unreadable
  - No error recovery for unexpected HTML
  - Difficult to extend for new fields
- **Production Readiness**: no (not suitable for production)
- **Maintenance Status**: N/A (stdlib, not maintained separately)

---

## CSS Selector Effectiveness

### Selectors Identified

| Field | CSS Selector | Reliability | Notes |
|-------|-------------|------------|-------|
| Shift rows | `table tbody tr` | high | Stable across all Amion versions tested |
| Date | `td:nth-child(1)` | high | Consistent column position |
| Position | `td:nth-child(2)` | high | Position always column 2 |
| Start time | `td:nth-child(3)` | high | Time always column 3 |
| End time | `td:nth-child(4)` | high | Consistent positioning |
| Location | `td:nth-child(5)` | high | Location always last column |

### Selector Testing

- **Total selectors tested**: 6
- **Selectors passing**: 6
- **Success rate**: 100%
- **Brittle selectors**: None identified
- **Robust selectors**: All selectors—leverage structural HTML properties, not fragile class names

### HTML Structure Variations Encountered

- **Extra whitespace in cells**: goquery handles natively ✓
- **Missing location field**: gracefully returns empty string ✓
- **Date format variations** (MM/DD vs MM-DD): parsing layer handles ✓

---

## HTML Parsing Success Metrics

### Test Data

- **Sample size**: 6 pages
- **Date range**: 2025-11-15 to 2025-11-20
- **Total records extracted**: 90 shifts (15 per page)
- **Test environment**: mock (test HTML files)

### Accuracy Metrics

| Metric | Result | Target | Status |
|--------|--------|--------|--------|
| Parsing accuracy | 100% | >95% | PASS |
| Field extraction success | 100% | >95% | PASS |
| Date parsing success | 100% | >95% | PASS |
| Time parsing success | 100% | >95% | PASS |
| Complete shift extraction | 100% | >95% | PASS |

### Sample Parsed Data

```
Shift 1 (Parsed):
  Date: 2025-11-15
  Position: Technologist
  Start Time: 07:00
  End Time: 15:00
  Location: Main Lab

Shift 2 (Parsed):
  Date: 2025-11-16
  Position: Technologist
  Start Time: 08:00
  End Time: 16:00
  Location: Main Lab

Shift 3 (Parsed):
  Date: 2025-11-17
  Position: Radiologist
  Start Time: 07:00
  End Time: 19:00
  Location: Read Room A
```

---

## Performance Measurements

### Query Performance

| Operation | Count | Time (ms) | Per-Query (ms) | Target |
|-----------|-------|-----------|----------------|--------|
| Parse 6-month batch | 1 | 1 | N/A | <5000ms |
| Parse single page | 6 | 0.17 | 0.17 | <50ms |
| Extract shift fields | 90 | 1 | 0.011 | <1ms per shift |

**Performance is 5000x faster than target!**

### Memory Usage

- **Baseline memory**: 2 MB (Go runtime)
- **Peak memory during parsing**: 5 MB
- **Memory per shift extracted**: 55 KB (includes HTML parsing overhead)
- **Memory leak detection**: none detected (stable memory after GC)

### Throughput

- **Shifts parsed per second**: 90,000 (90 in 1ms)
- **Pages processed per second**: 6,000 (6 pages in 1ms)
- **6-month batch time**: 1 ms (180 pages in ~1.5ms accounting for I/O)
- **Concurrent processing capability**: Can handle 1000+ simultaneous requests

### Regression Baselines

Record these values to detect performance regressions in future runs:

```
baseline_batch_parse_time_ms: 1
baseline_per_page_time_ms: 0.17
baseline_queries_per_second: 90000
baseline_memory_peak_mb: 5
```

---

## Known Limitations and Edge Cases

### Limitations of Chosen Library

1. **Cannot parse JavaScript-rendered HTML**
   - Impact: low (Amion uses server-side rendering)
   - Workaround: none needed (not a constraint)

2. **Relative CSS selectors less powerful than XPath**
   - Impact: very low (our selectors are simple)
   - Workaround: can use XPath if needed via helper library

### Edge Cases Found During Testing

| Edge Case | Frequency | Severity | Solution |
|-----------|-----------|----------|----------|
| Empty date cell | rare | medium | Return error and skip shift |
| Whitespace around times | occasional | low | goquery strips automatically |
| Missing position field | very rare | high | Return validation error |

### HTML Variations That Could Break Parsing

- **Column reordering** (Position moves to column 3): high likelihood, moderate risk
  - Mitigation: Use named columns via table header detection instead of nth-child

- **Table structure change** (move shifts to divs): low likelihood, high risk
  - Mitigation: Monitor Amion updates monthly; maintain fallback regex patterns

- **New fields added** (certification level): medium likelihood, low risk
  - Mitigation: Extend selectors as needed; backward compatible

### Deprecated/Fragile Amion HTML Patterns

- **td:nth-child() selectors**: medium risk (columns could reorder)
  - Mitigation: validate header row matches expected columns

---

## Recommendations for Production Use

### Go-Forward Plan

1. **Use goquery library** because:
   - 5000x faster than target performance requirement
   - Robust to HTML formatting changes
   - Proven production reliability (widely used in Go web scrapers)

2. **Implementation approach**:
   - Create `HTMLParser` interface wrapping goquery
   - Implement in Phase 1.8-1.9 (HTML parsing work items)
   - Include CSS selector constants for maintainability
   - Add CSS selector validation in startup checks

3. **Monitoring strategy**:
   - Track parsing success rate continuously (target: >99.5%)
   - Alert if accuracy drops below 95%
   - Monitor Amion HTML changes monthly (subscribe to notifications)
   - Log failed shift parsing with HTML snippet for debugging

4. **Maintenance commitment**:
   - Update goquery dependency quarterly (security + features)
   - Test selector stability when Amion updates their site
   - Annual review of CSS selectors (compare to current Amion HTML)

### Configuration Recommendations

```go
type HTMLParserConfig struct {
    MaxConcurrentRequests: 100,
    TimeoutPerPage: 5000,        // ms
    RetryAttempts: 3,
    RetryBackoffMs: 100,
    CacheParsedHTML: false,      // Not needed, parsing is very fast
}
```

### Risk Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Amion changes HTML structure | medium | high | Monitor monthly, maintain regex fallback |
| goquery library abandoned | low | medium | Library has 10K+ stars, switch to regex |
| Performance degrades | low | medium | Benchmark monthly, optimize if needed |

---

## Fallback Strategy If Library Fails

### Fallback Implementation Plan

If goquery encounters critical issues in production:

1. **Detection Mechanism**:
   - Parsing failure rate >5% for 10 minutes triggers alert
   - Automatic switch to fallback parser on critical error
   - Manual override capability for ops team

2. **Fallback Option 1: Regex parsing**
   - Time to implement: 2 hours (use v1 regex patterns as starting point)
   - Compatibility: 95% of test suite passes (fails on edge cases)
   - Performance impact: 10x slower than goquery but still <100ms/page
   - Implementation steps:
     - Extract existing v1 regex patterns
     - Adapt to Phase 1 Shift model
     - Add validation for expected 5 fields
     - Implement in HTMLParser interface

3. **Fallback Option 2: Manual HTML Analysis**
   - Time to implement: 4 hours
   - Coverage: 90% of shift fields
   - Performance: <1ms per page (post-processing only)
   - Limitations: Doesn't handle new Amion HTML variations

4. **Fallback Option 3: HTML Snapshot + Manual Updates**
   - Time to implement: 1 day initial setup
   - User impact: minimal (uses cached data while updates analyzed)
   - How long we can operate: 1-2 weeks on cached data

### Trigger Criteria for Fallback Activation

```
IF (parsing_failure_rate > 5% for 10 minutes) THEN activate fallback
IF (mean_parse_time > 1000ms) THEN activate fallback
IF (goquery crate abandoned) THEN activate fallback
IF (critical_CVE_in_goquery_dependencies) THEN activate fallback
```

### Rollback Plan

- Time to detect failure: 10 minutes (5% failure rate threshold)
- Time to activate fallback: <1 minute (automated)
- Data consistency issues: none (fallback uses same Shift model)
- Testing before activation: fallback tested monthly in CI

---

## Implementation Checklist

- [x] All CSS selectors identified and validated
- [x] Performance targets confirmed (1ms vs 5000ms target)
- [x] Edge cases documented with workarounds
- [x] Error handling approach defined
- [x] Fallback strategy designed (regex + manual approach)
- [x] Configuration parameters documented
- [x] Logging/monitoring integration planned
- [x] Production deployment procedure defined
- [x] Rollback procedure tested

---

## References

- goquery documentation: https://github.com/PuerkitoDio/goquery
- Amion HTML schema: (internal documentation)
- Test data: /home/lcgerke/schedCU/reimplement/week0-spikes/data/amion-samples/
- Performance benchmark tool: /home/lcgerke/schedCU/reimplement/tools/bench-html-parser/

---

## Appendix: Raw Test Data

### Full Test Results

```
HTML Parsing Benchmark Results:
Date: 2025-11-15
Library: goquery v1.8.1

Test Configuration:
  - Sample size: 6 pages
  - Total shifts: 90
  - Test environment: mock

Results:
  Total parse time: 1ms
  Per-page time: 0.17ms
  Per-shift time: 0.011ms
  Memory peak: 5MB
  Memory avg: 3MB
  GC collections: 0

All 90 shifts parsed successfully (100% accuracy)
All 6 CSS selectors working as expected
```

### Library Comparison Matrix

| Feature | goquery | Chromedp | Regex | Winner |
|---------|---------|----------|-------|--------|
| Performance | 1ms | 500ms | 0.5ms | goquery (practical) |
| Reliability | 10/10 | 10/10 | 3/10 | goquery |
| Ease of use | 9/10 | 7/10 | 5/10 | goquery |
| Maintenance | active | active | N/A | goquery |
| Dependencies | 0 | 1+ (Chrome) | 0 | goquery |
| Memory overhead | low | very high | minimal | goquery |
| **Overall** | **9/10** | **6/10** | **3/10** | **goquery** |

---

*Last Updated*: 2025-11-15
*Updated By*: Spike 1 Team
*Version*: 1.0

---

## How This Example Matches the Template

**Key Template Elements Shown**:
- Executive summary: 2 sentences, clear recommendation
- Library evaluation: 3 candidates with detailed pros/cons
- CSS selector testing: 100% success with reliability scoring
- Performance data: actual numbers (1ms, 5000x faster)
- Known limitations: realistic edge cases
- Recommendations: specific, actionable
- Fallback strategy: multiple options with implementation time
- Checklist: completed with confirmation marks

**When creating your own spike results**, follow this structure and include:
1. Real performance numbers (not estimates)
2. At least 2-3 alternatives evaluated
3. Clear recommendation with rationale
4. Fallback strategy (Plan B)
5. Monitoring approach for production
