# Phase 1 Parallelization Plan

## Executive Summary

**YES - HIGHLY PARALLELIZABLE!** Phase 1 can be safely divided across a 3-person team with minimal blocking.

**Critical insight**: The three core import services (ODS, Amion, Coverage Calculator) are **100% independent** - they share only the ValidationResult pattern and entity types (both immutable infrastructure).

**Wall-clock timeline: ~3 days** (with 3 people working 8h/day in parallel)
- Amion service is critical path (12-16h)
- Others complete in parallel without blocking
- ScheduleOrchestrator blocked only by Amion completion

---

## Dependency Analysis

### Zero-Dependency Startup Phase
```
ValidationResult Framework (2-4h)
  → No blocking factors
  → Can start immediately
  → Unblocks everything else
```

### Three Independent Parallel Tracks
```
ODS Service          Amion Service       Coverage Calculator
├─ Depends on:       ├─ Depends on:       ├─ Depends on:
│  ✅ Entities        │  ✅ Entities        │  ✅ Entities
│  ✅ Validation      │  ✅ Validation      │  ✅ Validation
│  ✅ Repository      │  ✅ HTTP client     │  ✅ Repository
├─ Estimate: 8-12h   ├─ Estimate: 12-16h  ├─ Estimate: 8-10h
└─ INDEPENDENT       └─ INDEPENDENT       └─ INDEPENDENT
```

### Blocking Point
```
All three services must complete before ScheduleOrchestrator
  ↓
ScheduleOrchestrator orchestrates all three (4-6h)
  ↓
Integration tests (6-8h) - can start once Orchestrator complete
```

---

## Work Distribution (3-Person Team)

### Team Roles
1. **Backend Engineer** - Core services, database optimization
2. **Integration Engineer** - Scraping, external APIs, data transformation
3. **QA/Infrastructure Engineer** - Testing, assertions, observability

---

## PHASE 1 TIMELINE WITH PARALLELIZATION

### Day 1 Morning (Shared)
**All 3 people (2 hours)**

**Task**: ValidationResult Framework Implementation
- Person A: Core ValidationResult struct + marshaling
- Person B: Error severity enum + message templates  
- Person C: JSON serialization tests + fixtures

**Deliverable**: ✅ Validation framework ready, tested, documented
**Blocker cleared**: All downstream services can now proceed

---

### Day 1 Afternoon - Day 2 Evening (PARALLEL - NO BLOCKING)

**Person A (Backend) - 16-24 hours**
```
├─ 0.5h: Code review ValidationResult
├─ 8-12h: ODS Import Service
│  ├─ TDD: tests for file parsing, error collection
│  ├─ Implementation: EDSImportService with ODS library
│  ├─ Integration: Repository calls for creating ShiftInstances
│  └─ Error handling: Validation errors collected, not fail-fast
├─ 4-6h: ScheduleOrchestrator (waits for all services)
│  ├─ 3-phase workflow orchestration
│  ├─ Error propagation from all services
│  └─ Transaction management
└─ 2-4h: Integration test setup
```

**Person B (Scraping) - 14-18 hours**
```
├─ 0.5h: Code review ValidationResult
├─ 12-16h: Amion Import Service ⭐ CRITICAL PATH
│  ├─ TDD: mock HTML response tests (from Spike 1)
│  ├─ Implementation: HTTP client + goquery selectors
│  ├─ Concurrency: 5 concurrent scrapers with rate limiting
│  ├─ Error handling: HTML parsing errors collected
│  ├─ Performance: test 6-month scrape time
│  └─ Fallback: Document Chromedp approach if goquery fails
├─ 1-2h: ScheduleOrchestrator participation (integration only)
└─ Parallel to A: Integration tests for Amion
```

**Person C (QA/Infrastructure) - 18-24 hours**
```
├─ 0.5h: Code review ValidationResult
├─ 2-3h: Query Count Assertion Framework
│  ├─ Add query counting to Testcontainers
│  ├─ Assert exactly N queries per operation
│  └─ Tests that fail if N+1 regression detected
├─ 8-10h: Coverage Calculator Service
│  ├─ TDD: pure function tests for coverage resolution
│  ├─ Implementation: batch query patterns (no N+1)
│  ├─ Query count assertions in tests
│  └─ Performance benchmarks
├─ 3-4h: API Error Handling Module
│  ├─ ApiResponse struct definition
│  ├─ ValidationResult embedding
│  ├─ JSON serialization tests
│  └─ Integration with Echo handlers (Phase 2)
├─ 2-3h: Documentation preparation
│  ├─ Spike result templates
│  ├─ Architecture decision documents
│  └─ Performance benchmark results
└─ Parallel to A & B: Integration test infrastructure
```

---

### Day 3 Morning (Synchronized)

**All 3 people (4-6 hours)**

**Task**: ScheduleOrchestrator Integration + End-to-End Testing

At this point:
- ✅ ValidationResult implemented
- ✅ ODS Service complete (Person A)
- ✅ Amion Service complete (Person B)
- ✅ Coverage Calculator complete (Person C)

**Synchronized work**:
1. Person A: Write ScheduleOrchestrator orchestration logic
2. Person B & C: Assist with integration points
3. All: Write end-to-end workflow tests
   - ODS file → ShiftInstances
   - Amion scrape → Assignments
   - Coverage resolution → Coverage metrics
   - Full workflow in one transaction

**Deliverable**: ✅ Complete orchestrated workflow, tested end-to-end

---

### Day 3 Afternoon (Wrap-up)

**All 3 people (2-3 hours)**

**Task**: Final Documentation + Review
- Document all spike results (Amion HTML, ODS parsing, performance)
- Architecture decision log (why each choice)
- Integration test results
- Performance benchmarks
- Code review all components

**Deliverable**: ✅ Phase 1 complete with full documentation

---

## Critical Dependencies & Blocking Points

### ✅ Can Start Immediately (0 blockers)
- ValidationResult Framework
- Entity documentation/code review
- Test infrastructure prep

### ✅ Can Start After ValidationResult (no other blockers)
- ODS Service (independent, mocked file parsing)
- Amion Service (independent, mocked HTML)
- Coverage Calculator (independent, mocked repository)
- API Error Handling (independent, uses ValidationResult)
- Query Count Framework (independent, test infrastructure)

### ❌ BLOCKED by all three services
- ScheduleOrchestrator (depends on ODS + Amion + Coverage)
  - Unblocks: Integration tests
  - Critical path determines: Amion Service (12-16h longest)

### ❌ BLOCKED by ScheduleOrchestrator
- End-to-end workflow tests
- Performance benchmarking
- Final integration verification

---

## Safe Parallelism Guarantees

### No Shared State
```
✅ Each service has own test fixtures
✅ Each service mocks its dependencies
✅ No database locks (Testcontainers per-test isolation)
✅ No file system conflicts
✅ No configuration conflicts
```

### Deterministic Merge Point
```
Services combine at ScheduleOrchestrator layer only
  ↓
No merge conflicts in individual services
  ↓
All three services work to same interfaces (ValidationResult)
  ↓
Integration is straightforward composition
```

### Test Isolation
```
Each service's tests are fully isolated:
- ODS tests: file parsing, validation
- Amion tests: HTML parsing, rate limiting
- Coverage tests: algorithm correctness
- Orchestrator tests: integration of all three

No test cross-contamination possible.
```

---

## Wall-Clock Timeline Breakdown

### With 1 Person (Sequential)
```
ValidationResult:    2-4h
ODS Service:         8-12h
Amion Service:       12-16h (CRITICAL)
Coverage Service:    8-10h
ScheduleOrchestrator: 4-6h
Integration Tests:   6-8h
────────────────────────
TOTAL:               40-56 hours = 5-7 days (8h/day)
```

### With 3 People (Parallel)
```
Day 1:
  Morning: ValidationResult (2-4h) - ALL TOGETHER
  Afternoon: 
    Person A → ODS (starts 4-6h of work)
    Person B → Amion (starts 12-16h of work) ⭐ LONGEST
    Person C → Coverage (starts 8-10h of work)

Day 2:
  All continue in parallel
  A finishes ODS (now waiting for B)
  C finishes Coverage (now waiting for B)
  B still working on Amion (critical path)

Day 3:
  Morning: Orchestrator + E2E tests (synchronized, 4-6h)
  Afternoon: Documentation (2-3h)

CRITICAL PATH: 2-4 (Validation) + 12-16 (Amion) + 4-6 (Orchestrator) + 6-8 (Tests) + 2-3 (Docs)
             = 26-37 hours ≈ 3-4 days wall-clock
             
ACTUAL: 3-4 days (Amion is bottleneck, others wait at Orchestrator)
```

### With 2 People (Constrained Parallel)
```
Person A: ValidationResult → ODS → Orchestrator → Tests
          2-4 + 8-12 + 4-6 + 3-4 = 17-26h (3-4 days, waits for B)

Person B: ValidationResult → Amion + Coverage
          2-4 + (12-16 + 8-10) = 22-30h (3-4 days, critical path)

Both must synchronize at Orchestrator
TOTAL: Still ~3-4 days (Amion critical path unchanged)
```

---

## Spike Dependencies

**This plan assumes:**
- ✅ Spike 1 (Amion HTML parsing): Goquery works OR Chromedp fallback documented
- ✅ Spike 2 (Job library): Asynq or Machinery selected (Phase 2, not Phase 1)
- ✅ Spike 3 (ODS library): Validated library available with error collection pattern

If Spike 1 fails (JavaScript-heavy Amion):
- Add 2 weeks to Phase 1 (Chromedp implementation)
- Timeline extends to ~5-6 weeks
- Still parallelizable (Amion just takes longer)

If Spike 3 fails (ODS library unavailable):
- Add 1-2 weeks to Phase 1 (custom ODS parser)
- Amion still critical path (Amion 12-16h > ODS 10-14h)

---

## Recommended Team Assignment

### Person A: Backend Engineer
- **Strengths**: Database, algorithms, services
- **Phase 1 work**: ValidationResult, ODS Service, Orchestrator
- **Effort**: 20-30h
- **Outcome**: Core data import logic

### Person B: Integration/Scraping Engineer
- **Strengths**: Web scraping, HTTP clients, error handling
- **Phase 1 work**: ValidationResult, Amion Service
- **Effort**: 14-20h (critical path)
- **Outcome**: External data integration

### Person C: QA/Infrastructure Engineer
- **Strengths**: Testing, observability, automation
- **Phase 1 work**: ValidationResult, Query Assertions, Coverage Calculator, Testing Framework
- **Effort**: 18-24h
- **Outcome**: Quality gates and performance assertions

---

## Blockers & Mitigations

| Blocker | Mitigation | Impact |
|---------|-----------|--------|
| Amion HTML changes during development | Mock responses from Spike 1, stub CSS selectors | Low - mocks allow dev to continue |
| ODS library bugs | Build custom parser as fallback | Medium - adds 1-2 weeks if triggered |
| Spike 1 fails (JavaScript) | Switch to Chromedp (2 week cost already documented) | Medium-High - extends timeline 2 weeks |
| Database availability | Use Testcontainers for all tests | Low - already implemented |
| Network failures (Amion scraping) | Implement retry logic + rate limiting | Low - part of service design |

---

## Success Criteria (End of Phase 1)

- ✅ ValidationResult framework: Complete, tested, documented
- ✅ ODS Service: Parses files, collects errors, creates ShiftInstances
- ✅ Amion Service: Scrapes schedules, handles rate limiting, creates Assignments
- ✅ Coverage Calculator: Resolves coverage, uses batch queries only
- ✅ ScheduleOrchestrator: Orchestrates all three in proper sequence
- ✅ Query count assertions: Prevent N+1 regressions
- ✅ 80%+ coverage on all services
- ✅ End-to-end workflow tested (ODS → Amion → Coverage)
- ✅ Performance benchmarks documented
- ✅ All Spike 1/2/3 results documented

---

## Conclusion

**Phase 1 is SAFELY PARALLELIZABLE** with clear separation of concerns:

1. **Minimal shared state** - Each service independent
2. **Clear merge point** - ScheduleOrchestrator orchestrates
3. **No circular dependencies** - DAG of services
4. **Deterministic blocking** - Only Amion (critical path) holds up Orchestrator
5. **Excellent scalability** - Works with 1, 2, or 3 people

**Recommended execution**: 3-person parallel team
- **Timeline**: 3-4 days (vs 5-7 days sequential)
- **Efficiency**: 3× speedup on parallel track
- **Quality**: Each specialist focused on their domain
