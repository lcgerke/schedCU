# Documentation Templates Guide

This directory contains standardized templates for Phase 1 documentation, designed to ensure consistency and completeness across all spike results, architectural decisions, and performance benchmarks.

---

## Quick Reference

### Available Templates

| Template | File | Purpose | Duration | When to Use |
|----------|------|---------|----------|-----------|
| **Spike 1 Results** | `SPIKE_1_HTML_PARSING_RESULTS.md` | Document HTML parsing library evaluation | 2-3 hours | After Spike 1 completion |
| **Spike 3 Results** | `SPIKE_3_ODS_LIBRARY_RESULTS.md` | Document ODS library evaluation | 1-2 hours | After Spike 3 completion |
| **Architecture Decision** | `ARCHITECTURE_DECISION.md` | Record major technical decisions | 1-2 hours | When making architectural choices |
| **Performance Benchmark** | `PERFORMANCE_BENCHMARK.md` | Document performance test results | 1-2 hours | After each performance measurement |

---

## Template Descriptions

### 1. Spike 1 HTML Parsing Results (`SPIKE_1_HTML_PARSING_RESULTS.md`)

**Purpose**: Document the evaluation of HTML parsing libraries for Amion schedule extraction.

**Key Sections**:
- Executive summary with recommendation
- Library evaluation comparing 2-3 candidates
- CSS selector effectiveness analysis
- HTML parsing success metrics (accuracy, coverage)
- Performance measurements (queries/second, memory usage)
- Known limitations and edge cases
- Production recommendations
- Fallback strategy if library fails

**Completion Time**: 2-3 hours

**Responsible Team**: HTML Parsing Spike Team

**Example Use Case**:
```
Scenario: Team completed evaluation of goquery vs Chromedp vs regex approach
Action: Fill template with findings
Output: Documented recommendation for goquery with performance metrics
Impact: Informs implementation decision in Phase 1.8
```

**Key Questions Answered**:
- Which library should we use for HTML parsing?
- What CSS selectors reliably extract shift data?
- Does the library meet performance targets (<5000ms for 6 months)?
- What happens if the library fails in production?

---

### 2. Spike 3 ODS Library Results (`SPIKE_3_ODS_LIBRARY_RESULTS.md`)

**Purpose**: Document the evaluation of ODS (OpenDocument Spreadsheet) parsing libraries.

**Key Sections**:
- Library choice and rationale
- Comparison to alternatives
- Error handling approach
- File size limits and performance characteristics
- Supported ODS features matrix
- Known issues and limitations
- Integration complexity assessment
- Custom parser fallback approach

**Completion Time**: 1-2 hours

**Responsible Team**: ODS Library Spike Team

**Example Use Case**:
```
Scenario: Team evaluated 3 ODS libraries and selected one
Action: Fill template with feature matrix, performance data, integration plan
Output: Documented rationale for library choice + integration roadmap
Impact: Guides Phase 1.1-1.6 (ODS service implementation)
```

**Key Questions Answered**:
- Which ODS library is production-ready?
- What features does it support? (merged cells, formulas, etc.)
- What file sizes can it handle?
- What's the custom parser fallback if library fails?

---

### 3. Architecture Decision (`ARCHITECTURE_DECISION.md`)

**Purpose**: Record major architectural choices with full context and tradeoff analysis.

**Key Sections**:
- Decision statement (what choice was made)
- Context (why the decision was needed)
- Options evaluated (2-3 alternatives with pros/cons)
- Chosen option with rationale
- Tradeoffs (what we give up)
- Implications (downstream effects)
- Implementation plan
- Monitoring and validation approach
- Rollback strategy

**Completion Time**: 1-2 hours

**Responsible Team**: Architecture/Spike Team

**Example Use Case**:
```
Scenario: Team decides to use validator/v10 for input validation
Action: Fill ADR template with options evaluated, rationale, implications
Output: Formal architecture decision record ADR-001
Impact: Creates accountability and historical record for decision
```

**Key Questions Answered**:
- What architectural choice was made and why?
- What alternatives were considered?
- What tradeoffs are we accepting?
- What systems are affected downstream?
- How will we know if the decision was correct?

---

### 4. Performance Benchmark (`PERFORMANCE_BENCHMARK.md`)

**Purpose**: Document performance test results with baselines for regression detection.

**Key Sections**:
- Operation being tested
- Test setup (system config, data config, methodology)
- Results (timing, memory, CPU, throughput)
- Analysis (bottlenecks, scaling characteristics)
- Baseline metrics for regression detection
- Optimization recommendations
- Regression detection thresholds

**Completion Time**: 1-2 hours

**Responsible Team**: Performance Testing Team

**Example Use Case**:
```
Scenario: Team benchmarks HTML parsing performance
Action: Run benchmarks, fill template with results
Output: Baseline BENCH-001 for HTML parsing (1ms for 6 months)
Impact: Future benchmarks compare against this baseline to detect regressions
```

**Key Questions Answered**:
- How fast is this operation?
- Does it meet our performance targets?
- Are we worse than the previous version?
- Where is the bottleneck?
- How should we optimize?

---

## How to Use These Templates

### Step 1: Choose the Right Template

```
Is this work...                          Then use...
├─ Spike 1 HTML parsing evaluation      → SPIKE_1_HTML_PARSING_RESULTS.md
├─ Spike 3 ODS library evaluation       → SPIKE_3_ODS_LIBRARY_RESULTS.md
├─ Major technical decision (goquery,   → ARCHITECTURE_DECISION.md
│  DB schema, caching strategy)
└─ Performance measurement (benchmark)  → PERFORMANCE_BENCHMARK.md
```

### Step 2: Create Your Document

1. Copy the template to your target location
2. Name it appropriately:
   - For results: `spike1_results_goquery.md` or `spike3_results_xlsx.md`
   - For ADRs: `ADR-001-html-parsing-library.md`
   - For benchmarks: `BENCH-001-html-parsing-6mo.md`
3. Fill in all sections, replacing `[placeholders]` with real values
4. If a section doesn't apply, explain why (e.g., "No fallback needed")

### Step 3: Complete the Template

**Required sections** (always fill):
- Executive summary
- Key findings/decision
- Rationale
- Implications

**Recommended sections** (fill unless N/A):
- Performance metrics
- Error handling
- Integration plan
- Monitoring strategy

**Optional sections** (fill if applicable):
- Appendices with raw data
- References to external documents
- Detailed implementation plans

### Step 4: Review and Finalize

Before considering your document complete:

- [ ] All `[placeholders]` replaced with real values
- [ ] Numbers justified (measurements, not guesses)
- [ ] Key decisions explained with context
- [ ] Trade-offs explicitly listed
- [ ] Implications identified for downstream teams
- [ ] Monitoring/validation approach specified
- [ ] Peer review by team lead
- [ ] Commit to git repo

### Step 5: Link and Index

1. Add link to `TEMPLATES_INDEX.md` (see below)
2. Reference from relevant phase documentation
3. Link from related ADRs and decisions

---

## Example: Filled-In Template

### Completed Example: Spike 1 HTML Parsing Results

See `EXAMPLE_SPIKE_1_RESULTS.md` for a completed example with:
- Real-world performance data from goquery benchmarking
- Actual CSS selectors identified
- Sample parsing accuracy metrics
- Fallback strategy to regex parsing
- Production recommendations

**Key takeaway from example**:
```
Recommendation: Use goquery for HTML parsing
- Achieves 1ms for 6-month batch (target: <5000ms) ✓
- 100% parsing accuracy on test data ✓
- 5 CSS selectors identified and stable ✓
- Fallback: regex patterns with 80% coverage
```

---

## Template Sections Deep Dive

### Executive Summary Section

Every template starts with an executive summary. This should be:
- **1-2 sentences maximum**
- **Answerable by busy readers in 30 seconds**
- **Include recommendation and timeline impact**

Example:
```
goquery library successfully parses Amion HTML with 1ms performance on 6-month batch.
Recommend proceeding with goquery implementation. Zero timeline impact (performance
exceeds targets by 5000x).
```

### Results vs. Context Sections

- **Context sections** answer "why did we do this?"
- **Results sections** answer "what did we find?"

Don't skip context—it's crucial for understanding decisions months later.

### Quantitative Data Guidelines

Every claim should have a number:

- "Performance is good" ❌
- "Performance is 1ms per page, 95% faster than regex" ✓

Benchmark commands and exact test conditions enable reproduction.

### Risk and Fallback Strategy

Every technical decision must have a fallback:

```
Primary: Use goquery library
Fallback 1: Regex parsing (80% coverage, 2 days to implement)
Fallback 2: Manual HTML snapshot + updates (5 days to implement)
Trigger: If goquery library abandoned or critical CVE discovered
```

---

## Common Pitfalls to Avoid

### 1. Incomplete Comparisons

❌ **Bad**: "goquery is better than regex"
✓ **Good**: "goquery is 5000x faster than regex and handles malformed HTML gracefully"

Include specific metrics for each alternative.

### 2. Missing Fallbacks

❌ **Bad**: "We're using library X" (no Plan B)
✓ **Good**: "We're using library X, with fallback to custom parser in case of failure"

Every dependency needs a fallback.

### 3. Vague Recommendations

❌ **Bad**: "Consider using this library"
✓ **Good**: "RECOMMEND using this library because [3 specific reasons with data]"

Take a clear stance backed by data.

### 4. Skipping Implementation Details

❌ **Bad**: "Integration is straightforward"
✓ **Good**: "Integration requires 3 hours: add dependency (30min), write wrapper (90min), test (60min)"

Estimate effort explicitly.

### 5. No Monitoring Plan

❌ **Bad**: "Deploy and move on"
✓ **Good**: "Monitor parsing success rate daily; alert if <99% accuracy for 1 hour"

Specify what you'll measure to validate the decision.

---

## Template Navigation

### Finding Information Quickly

```
I want to know...                    Look in...
├─ Which library to use              → Executive Summary of SPIKE template
├─ Performance numbers               → Results section of PERFORMANCE_BENCHMARK
├─ Trade-offs of decision            → Tradeoffs section of ARCHITECTURE_DECISION
├─ Fallback strategy                 → Fallback section of SPIKE template
├─ Why this choice was made          → Rationale section of ARCHITECTURE_DECISION
├─ How to reproduce tests            → Appendix of PERFORMANCE_BENCHMARK
├─ Integration complexity            → Integration Complexity section of SPIKE_3
└─ Potential risks                   → Risk Assessment of ARCHITECTURE_DECISION
```

### Cross-Linking Templates

Templates reference each other:

```
ARCHITECTURE_DECISION (ADR-001: Use goquery)
    ↓ links to
SPIKE_1_HTML_PARSING_RESULTS (recommends goquery with data)
    ↓ links to
PERFORMANCE_BENCHMARK (BENCH-001: goquery throughput)
```

When updating one, check if others need updates.

---

## Template Maintenance

### When to Update a Template

- When spike findings change (e.g., library performance improves)
- When implementation reveals new constraints
- When a decision is revisited and reversed
- Quarterly review to catch outdated information

### Version Control

Each template has a Version field. Increment:
- **1.0**: Initial completion
- **1.1**: Minor fixes (typos, clarifications)
- **2.0**: Significant change (decision reversed, new data)

### Status Tracking

Architecture Decision templates use Status field:
- **Decided**: Decision made, implementation proceeding
- **Pending**: Awaiting approval or more information
- **Rejected**: Decision made not to pursue this option
- **Superseded by ADR-XXX**: This decision was replaced by another

---

## Integration with Phase 1 Workflow

### Timing

```
Week 0 (Spikes)           Week 1 (Phase 1)
├─ Run Spike 1            ├─ Use Spike 1 results
├─ Run Spike 3            ├─ Implement based on spike recommendations
├─ Fill spike templates   ├─ Create ADRs for implementation decisions
└─ Document findings      ├─ Benchmark implementations
                          └─ Fill performance templates
```

### Artifacts Generated

| Phase | Artifact | Template |
|-------|----------|----------|
| Spike phase | Spike 1 results | SPIKE_1_HTML_PARSING_RESULTS.md |
| Spike phase | Spike 3 results | SPIKE_3_ODS_LIBRARY_RESULTS.md |
| Phase 1 | Architectural decisions | ARCHITECTURE_DECISION.md (multiple) |
| Phase 1 | Performance baselines | PERFORMANCE_BENCHMARK.md (multiple) |

---

## Getting Help

### Template Questions

**Q: I'm not sure if something applies to my situation**
A: Fill in "N/A - [reason]" instead of leaving blank. Explain why the section doesn't apply.

**Q: Can I skip a section?**
A: No. Either fill it or explain why it's not applicable. Every section exists for a reason.

**Q: How detailed should results be?**
A: Detailed enough that someone could make a decision based on your document without asking follow-up questions.

**Q: Should I include raw data?**
A: Yes, put raw benchmark output/logs in the Appendix. Readers can drill down if needed.

---

## Checklist for Template Completion

Use this before submitting:

- [ ] All `[placeholders]` replaced with actual values
- [ ] Executive summary is 1-2 sentences
- [ ] Key decision/finding stated clearly
- [ ] All metrics include units (ms, MB, %, ops/sec)
- [ ] Performance claims backed by numbers
- [ ] Fallback strategy specified for technical decisions
- [ ] Downstream implications identified
- [ ] Monitoring/validation approach specified
- [ ] References provided for external links
- [ ] Version number incremented
- [ ] Date and author filled in
- [ ] Peer reviewed by team lead
- [ ] Committed to git

---

## References

- [Phase 1 Master Plan](../MASTER_PLAN_v2.md)
- [Phase 1 Parallelization](../PHASE_1_20_AGENT_PARALLELIZATION.md)
- [Spike Results Location](../week0-spikes/results/)

---

*Template Guide Version*: 1.0
*Last Updated*: 2025-11-15
*Maintained By*: Phase 1 Documentation Team
