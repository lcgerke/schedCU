# Master Plan v2 — Changelog & Improvements

## Summary

MASTER_PLAN_v2.md incorporates critical feedback to de-risk the project and address unvalidated assumptions. The original plan was comprehensive but relied on several unproven assertions about external systems (Amion HTML parsing, job library availability, ODS library capabilities).

**Key change**: Added **Week 0 Dependency Validation Spikes** before Phase 0 begins. These 3-5 days eliminate the biggest unknowns and reduce overall risk from Medium-High to Low-Medium.

---

## Major Improvements

### 1. **Week 0: Dependency Validation Spikes** (NEW)

**Problem addressed**: The original plan assumed:
- Amion serves HTML that goquery can parse efficiently (10× speedup)
- Asynq (Redis) is available in hospital infrastructure
- ODS library supports the error collection pattern

These are unvalidated assumptions that could invalidate timeline and performance claims.

**Solution**: Three parallel spikes (3-5 days total, 1-2 people):

**Spike 1: Amion HTML Parsing Feasibility**
- Download live Amion page, test goquery CSS selectors
- Validate parsing accuracy and performance
- **Success criteria**: goquery works (2-3s for 6 months) OR clear fallback to Chromedp with cost (+2 weeks)
- **Eliminates**: The biggest performance assumption; de-risks Phase 3

**Spike 2: Job Library Evaluation**
- Verify Redis availability in hospital infrastructure
- Test Asynq vs Machinery side-by-side
- **Success criteria**: Asynq works OR Machinery (PostgreSQL) works OR custom documented with cost (+3 weeks)
- **Eliminates**: Dependency risk; ensures job system is feasible

**Spike 3: ODS Library Validation**
- Test chosen library with production file sizes
- Validate error collection pattern works
- **Success criteria**: Library meets requirements OR wrapper layer designed with cost (+1 week)
- **Eliminates**: File parsing risk; ensures error handling pattern

**Timeline impact**:
- Best case: All spikes succeed, no timeline impact (overhead absorbed)
- Moderate case: 1 spike fails, fallback adds 1-2 weeks to subsequent phase
- Worst case: Multiple failures, timeline extends to 4-5 months (but v1 is operational, no urgency)

---

### 2. **Risk Level Reassessment**

**Original plan stated**: "Risk Level: Low-Medium"

**Updated assessment**: "Risk Level: Medium-High before Week 0, Low-Medium after Week 0"

**Rationale**:
- Original plan made 3 unvalidated assertions about external systems
- Amion scraping assumption (10× speedup) lacks proof
- Job library choice depends on infrastructure (unconfirmed)
- ODS library pattern compatibility (unconfirmed)

**New approach**: De-risk these with Week 0 spikes, then proceed with confidence.

---

### 3. **Infrastructure Clarifications** (NEW)

Added explicit tasks to confirm hospital infrastructure details:

- **S3 availability**: For audit log archival (or fallback to PostgreSQL archive)
- **Redis availability**: For Asynq job queue (or fallback to Machinery)
- **Vault readiness**: For credential storage (or fallback to env vars)

These are now Week 0 deliverables, not assumptions.

---

### 4. **Data Continuity Strategy — Clarified**

**Original**: "Audit logs → S3/cold storage"

**Updated**: "Audit logs → S3 (preferred) OR PostgreSQL archive (if S3 unavailable)"

**Rationale**: Not all hospitals use S3; must support PostgreSQL-based archive as fallback.

---

### 5. **Amion Scraping — Fallback Plan Explicit**

**Original**: "Fallback: If Amion uses heavy JavaScript, switch to headless browser (Chromedp)"

**Updated**: Three layers:
1. Try goquery (Spike 1 validates this)
2. If goquery fails, fallback is Chromedp (documented in Spike 1)
3. Chromedp adds ~2 weeks to Phase 3 timeline (explicit cost)
4. Phase 3 weeks are adjusted based on which approach is chosen

**Benefit**: No surprises; team knows exactly what to do if HTML parsing fails.

---

### 6. **On-Call Runbooks** (NEW)

**Original plan**: Mentioned "monitoring" but lacked operational guidance

**Updated**: Phase 4 explicitly includes runbooks for:
- How to analyze audit logs for compliance investigations
- How to troubleshoot token-related issues (Vault down, rotation failures)
- How to handle Amion scraping failures
- How to monitor job queue health
- Alert thresholds and escalation procedures

**Deliverable**: Runbooks are tested with on-call team before production cutover.

---

### 7. **Team Skill Validation** (NEW)

Added explicit pre-Week 0 questions:

- "Has the team worked with Go on production systems?"
- "Does anyone have experience with Vault integration?"
- "Is there PostgreSQL expertise available for performance tuning?"
- "Do we have on-call staff trained in Go debugging?"

**Benefit**: Identifies skill gaps before starting, allows for training or hiring adjustments.

---

### 8. **Definition of Done — More Specific**

**Original**: "Schema translated (v1 → v2 mapping document complete)"

**Updated**: "Schema translated (v1 → v2 mapping document with **rationale for each field**)"

**Rationale**: Ensures the team understands WHY each field exists, not just what it is.

Similar improvements throughout for "core services feature-complete", "team trained on decisions", etc.

---

### 9. **Success Metrics — Conditional on Spikes**

**Original**: Fixed 10× performance improvement claim

**Updated**: "10× performance improvement (Amion scraping: 180s → 2-3s) — **pending Spike 1 validation; otherwise documented actual improvement**"

**Rationale**: Honest about unknowns. If Chromedp is required, performance will be less than 10×; we'll document actual number and adjust expectations.

---

### 10. **Timeline Adjusted**

**Original**: "14 weeks to production (standard full-time, 3 people)"

**Updated**: "15-16 weeks to production (includes Week 0 validation + Phase 0-4)"

**Breakdown**:
- Week 0: Dependency validation spikes (3-5 days in parallel)
- Week 1: v1 security fixes + Phase 0 starts
- Weeks 1-7: Phase 0-1 (schema, core services)
- Weeks 7-10: Phase 2 (API, security)
- Weeks 10-13: Phase 3 (scrapers) — adjusted if Spike 1 fails
- Weeks 13-15: Phase 4 (testing, polish)
- Weeks 15-16: Cutover

**Caveat**: If multiple spikes fail, timeline extends further, but v1 remains operational and secure (no pressure).

---

### 11. **Quality Gates Enhanced**

**New gate before Phase 0**:
```
CANNOT proceed to Phase 0 without:
✓ Week 0 spikes completed (Amion parsing, job library, ODS library validated)
✓ Infrastructure clarifications confirmed
```

**Rationale**: Ensures critical dependencies are verified before committing team effort.

---

### 12. **Fallback Costs Documented**

Added explicit cost/timeline for each failure scenario:

| Failure | Fallback Cost |
|---------|---------------|
| Amion goquery fails | +2 weeks (switch to Chromedp) |
| Asynq unavailable | Custom job queue +3 weeks OR Machinery baseline |
| ODS library insufficient | Wrapper layer +1 week |
| Multiple failures | 4-5 month timeline total |

**Benefit**: Team can make informed decisions if issues arise; no surprises.

---

### 13. **Documentation Structure**

**New directories in project**:
```
docs/
├── schema/              # v1 → v2 mapping with rationale
├── runbooks/            # On-call procedures (token issues, audit logs, etc.)
└── spikes/              # Week 0 spike results (Amion parsing, job library, ODS)
```

**Benefit**: Spike results are documented and referenced throughout the plan.

---

## Section-by-Section Changes

### Executive Summary
- **Added**: "Validate critical assumptions early before committing resources"
- **Added**: "10× faster ... — subject to Amion HTML validation"
- **Added**: "Week 0: Dependency validation spikes (de-risk major assumptions)"

### 17 Critical Decisions
- **Decision 4**: Updated to show both Asynq and Machinery as viable (post-Spike 2)
- **Decision 7**: Added "On-call playbook for token-related issues"
- **Decision 12**: Added "Test all migrations in CI using Testcontainers"
- **Decision 13**: Shows ODS library choice is validated in Spike 3
- **Decision 14**: Explicit fallback to Chromedp with timeline cost
- **Decision 15**: S3 OR PostgreSQL archive (clarified infrastructure requirement)
- **Decision 16**: Added "On-call runbook for audit log analysis"
- **Decision 17**: Added "Alert thresholds and on-call escalation procedures documented"

### Implementation Timeline
- **Added**: "Week 0: Dependency Validation (NEW — 3-5 days)" section
- **Added**: Team skill validation questions
- **Phase 3**: Timeline adjusted based on Spike 1 results
- **Phase 4**: Explicit on-call runbook testing requirement

### Success Metrics
- **New**: "Week 0 Completion (Dependency Validation)" section with specific gates
- **Updated**: Phase 0 now includes "Spike results documented in `docs/spikes/`"
- **Updated**: Phase 4 includes "On-call team has completed runbook review"

### Risk Mitigation
- **Added**: 10+ new rows covering spike-related risks with explicit mitigations
- **Updated**: All fallback timelines documented with cost

### Quality Gates
- **New**: "CANNOT proceed to Phase 0 without: Week 0 spikes completed"

---

## What Did NOT Change

The following remain unchanged from the original plan:

1. **Core architectural decisions** (17 decisions locked)
2. **Phase structure** (Phase 0-4 + Cutover)
3. **Team roles** (3 people: Backend, API/Security, Test/DevOps)
4. **v1 security fixes** (Week 1, parallel track)
5. **Key services** (ValidationResult, ScrapeBatch, DynamicCoverageCalculator)
6. **Testing strategy** (85% coverage, query count assertions)
7. **Go + PostgreSQL + Docker + Vault + Prometheus stack**

The improvements are about **de-risking and clarifying**, not redesigning the approach.

---

## Files Generated

- **MASTER_PLAN_v2.md**: The updated plan (new active file)
- **MASTER_PLAN_DEPRECATED.md**: Original plan (kept for reference)
- **MASTER_PLAN_v2_CHANGELOG.md**: This file

---

## Next Steps

1. **Review** MASTER_PLAN_v2.md with the team (30 min)
2. **Decide**: Do we commit to Week 0 spikes? (5 min)
3. **If yes**: Assign spike owners and start next week (3-5 days total)
4. **If no**: Proceed directly to Phase 0 (accept original risk level: Medium-High)

**Recommendation**: **Strongly** recommend Week 0 spikes. The cost is minimal (3-5 days) and the benefit is high (eliminate biggest unknowns). This is the difference between a Medium-High risk project and a Low-Medium risk project.
