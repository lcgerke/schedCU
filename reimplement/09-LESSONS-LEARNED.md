# Lessons Learned from v1 Implementation

Key insights and takeaways from building the Hospital Radiology Schedule System that should inform v2 development.

---

## 1. Domain-Driven Design Works for Complex Business Logic

### What We Got Right

The system correctly identified and modeled core domain concepts:
- **ScheduleVersion**: Temporal versioning, STAGING/PRODUCTION state machine
- **ShiftInstance**: Coverage rules as first-class entity
- **ScrapeBatch**: Atomic operations with traceability
- **Assignment**: Minimal link between Person and Shift

**Impact**: The system correctly represents the business problem. When requirements changed (person specialties), the model adapted without major refactoring.

### Lesson for v2

**Don't skip domain modeling.** Spend time with business stakeholders to identify:
- What are the core entities?
- What are state transitions?
- What are the invariants (rules that must always be true)?

Example invariant discovered in v1:
```
"If a BODY_ONLY person works ON1/ON2, then other people that night
must be assigned to ONNeuro, not the original shift."
```

This drove the entire dynamic coverage calculator. Identifying this early saved weeks of rework.

---

## 2. Lazy Evaluation Beats Eager Evaluation for Dynamic Data

### The Mistake

v0 approach: Calculate and store `reassignedShiftType` when creating assignments
```java
// v0 - STALE DATA PROBLEM
if (hasBodyOnlySpecialist) {
    assignment.reassignedShiftType = "ONBody";
    assignment.save();  // Stored once, never updated
}

// Later, person's specialty changes → stale data remains
```

### The Fix (v1)

Calculate at query time:
```java
// v1 - ALWAYS FRESH
String effective = calculator.getEffectiveShiftType(assignment, date);
// Uses current person.specialty, no stale data
```

**Trade-off Analysis**:

| Approach | Pros | Cons |
|----------|------|------|
| **Eager (v0)** | Fast queries, simple code | Stale data, migrations needed |
| **Lazy (v1)** | Always fresh, no stale data | More queries, needs optimization |

**Lesson**: For data that changes frequently (person specialties, preferences), prefer lazy evaluation even if it costs query performance. Fix query performance with batching, caching, indexes—not by storing stale data.

### Application to v2

When designing new features, ask:
- "How often does this data change?"
- "If it changes, do all derived values need recalculation?"
- "If yes → lazy evaluation"
- "If no → eager evaluation is fine"

---

## 3. Batch Processing Prevents Silent Data Loss

### What Happened

Early versions didn't have ScrapeBatch. Result:
```
Scrape 500 assignments
Parse 400 successfully
Store 350 (150 discarded silently due to bugs)
↓
No way to know data was lost
No way to replay
No audit trail
```

### The Fix (Scrape Batch)

Every scrape is atomic:
```
Scrape 500 assignments
  ↓
Create ScrapeBatch (PENDING)
  ↓
Parse all 500
  ↓
Validate all 500 succeed
  ↓
Ingest all 500 (transaction)
  ↓
Mark ScrapeBatch as COMPLETE
  ↓
If any error → ScrapeBatch FAILED, no data stored
```

**Benefit**: Data integrity + auditability.

### Lesson for v2

**Always use batching for external data sources.** The pattern:
```
1. Create batch header (PENDING)
2. Fetch/parse all data
3. Validate all data
4. Write all data (single transaction)
5. Mark batch as COMPLETE/FAILED
```

Apply this to:
- ODS imports
- Amion scraping
- Manual imports
- API data

---

## 4. Validation Framework Should Be Extensible

### Why v1 Nailed This

```java
validation.addError("UNKNOWN_SHIFT_TYPE", "Unknown shift: XRAY");
validation.addWarning("INSUFFICIENT_COVERAGE", "Only 1 person assigned...");

// Can promote?
if (!validation.canPromote()) {
    // Show warnings and ask user
}

// Can import?
if (!validation.canImport()) {
    // Block import if errors
}
```

This gives users **choice** while preventing **disasters**.

### Lesson for v2

Validation framework should:
1. Collect ALL errors/warnings (not fail fast)
2. Categorize by severity (ERROR, WARNING, INFO)
3. Store for audit trail (JSON in database)
4. Have clear semantics:
   - `canImport()` = no errors
   - `canPromote()` = no errors/warnings
   - `canDelete()` = different rules

This pattern prevents the "fix one error, discover 10 more" debugging cycle.

---

## 5. Security Bypass Happens When "Testing" Code Ships

### What We Found

```java
@PermitAll  // TODO: Remove after testing ← SHIPPED!
@PostMapping("/import/amion")
public Response scrapeAmion() { ... }
```

Multiple endpoints with TODO comments that made it to the codebase.

### Why This Happened

1. **Dev convenience**: Developers added `@PermitAll` to test quickly
2. **Forgotten**: Removed from local testing, but committed to git
3. **No review**: Code reviewers didn't catch it
4. **Test coverage**: No security tests verified endpoints were protected

### Lesson for v2

**Security is not optional or "nice to have":**

1. **Never ship with test bypasses**: `@PermitAll`, commented security checks, hardcoded passwords
2. **Security tests are required**: Every endpoint must have a test asserting it requires auth
3. **Code review checklist**: Reviewer must check:
   - [ ] No `@PermitAll`
   - [ ] No hardcoded credentials
   - [ ] No commented security
   - [ ] No TODO security items
4. **Pre-commit hook** should catch `@PermitAll` in production code

Example pre-commit hook:
```bash
#!/bin/bash
if git diff --cached | grep -q "@PermitAll\|@Override"; then
    echo "ERROR: Found security bypass markers in code"
    echo "Remove @PermitAll and test TODOs before committing"
    exit 1
fi
```

---

## 6. End-to-End Tests Are Valuable But Insufficient

### What v1 Did Well

```java
ScheduleCalendarE2ETest          // Tests real UI, real database
ScheduleCalendarLiveE2ETest      // Tests against actual Amion
ConsolidatedCalendarE2ETest      // Tests complex data scenarios
```

These caught bugs that unit tests missed.

### What v1 Missed

```
❌ No unit tests for ODSImportService (parse logic)
❌ No unit tests for AmionImportService (scrape logic)
❌ No unit tests for DynamicCoverageCalculator (core logic)
❌ No integration tests for APIs
❌ No security tests
```

Result: Bugs only discovered in e2e tests (slow to run, hard to debug).

### Lesson for v2

**Use the test pyramid**:

```
        /\
       /E2E\          10% - End-to-end tests (full system)
      /    \          - Real database, real HTTP calls
     /------\         - Slow but catch integration issues
    /        \
   /Integration\      30% - Integration tests
  /          \        - Actual DAOs, services, but mocked external
 /_____________\      - Medium speed
/              \
/  Unit Tests   \     60% - Unit tests
/________________\    - Mocked dependencies
                      - Fast, cheap to run

What v1 had:        What v2 should have:
    /\                    /\
   /E2E\                 /E2E\
  /    \ (lots)         /    \
 /------\              /------\
/        \            /        \
/   Minimal\         / Good      \
/          \        /Integration  \
/_____________\    /_____________\
/              \  /              \
/  Unit Tests   \/  Many Tests    \
/________________\________________\

Goal: Each layer should be self-contained,
failures should point to specific component
```

Priority for v2:
1. **Unit tests** for all services (ODS parser, Amion scraper, coverage calculator)
2. **Integration tests** for workflows (ODS import → Amion scrape → coverage resolution)
3. **API tests** for all endpoints (contract validation, security)
4. **E2E tests** for critical user workflows (schedule import, promotion, viewing)

---

## 7. Refactoring Takes Time, Plan For It

### What Happened

v1 was supposed to be "just maintenance" but required:
- Dynamic coverage refactor (1 week)
- Scrape batch refactor (1 week)
- Database migrations (4 hours each)
- Test writing (3 days)

Total: ~2.5 weeks of unexpected refactoring.

### Why It Happened

The codebase had technical debt:
- Dead code (`reassignedShiftType` field)
- Scalability issues (N+1 queries)
- Security issues (hardcoded values)

Each refactoring required:
1. Understanding existing code
2. Designing better approach
3. Implementing change
4. Writing tests
5. Database migration
6. Verification

### Lesson for v2

**Budget refactoring time upfront:**

Sprint velocity calculation:
```
Capacity per sprint = 40 hours

Feature development:    30 hours (75%)
Refactoring/debt:       10 hours (25%)

Do NOT squeeze refactoring. It prevents future speed.
```

Build in time for:
- Test writing (don't skip)
- Code review (don't speed up)
- Documentation (needed for next person)
- Performance optimization (don't defer)

**This actually makes delivery FASTER** because bugs discovered in production take 10× as long to fix.

---

## 8. Configuration Management is Security-Critical

### What We Got Wrong

```properties
quarkus.datasource.password=postgres  # ❌ Hardcoded in Git
scraping.amion.file-id=!1854...       # ❌ Hardcoded in Git
```

### What v2 Should Do

```properties
# application.properties (committed to Git)
quarkus.datasource.password=${DB_PASSWORD}

# Actual values (NOT committed)
export DB_PASSWORD=actual-password
```

Deployment:
```bash
# Never share passwords
git log --all | grep password  # Should return nothing

# Use secret management
- Docker: secrets volume
- Kubernetes: Secret objects
- Environment: .env file (local only)
```

### Lesson for v2

**Create security checklist before first deploy:**
- [ ] No hardcoded credentials in Git
- [ ] All secrets in environment variables
- [ ] `.env` added to `.gitignore`
- [ ] Secrets verification in CI (scan for patterns)
- [ ] Test `git log` contains no passwords

---

## 9. Documentation is Investment, Not Overhead

### What Worked Well

v1 has 12 markdown documents:
```
docs/
├── COMPLETE_SYSTEM_SPECIFICATION.md
├── AMION_PARSING_DECISIONS.md
├── NATURAL_LANGUAGE_PATTERNS.md
├── troubleshooting/
└── when-planning/
```

Result: New person can ramp up in 2-3 days instead of 2-3 weeks.

### Lesson for v2

**Document:**
1. **Architecture** - How components interact
2. **Decisions** - Why we chose this approach
3. **Gotchas** - What surprised us
4. **Troubleshooting** - How to debug issues
5. **Runbooks** - How to do common operations

Format:
```markdown
# Decision: Why We Use JWT Instead of Session Cookies

## Context
Multiple deployment instances needed to share auth state

## Options Considered
1. Session cookies + Redis (stateful)
2. JWT tokens (stateless)
3. OAuth2 (external dependency)

## Decision
JWT tokens (stateless, simple, scalable)

## Consequences
- ✅ Works across load balancers
- ✅ No session server needed
- ❌ Token revocation is harder
- ❌ Logout takes until token expires

## Revisit Date
2025-06-01 (when we have 100+ users)
```

---

## 10. Testing Improves Design

### What We Discovered

When writing tests for `CoverageResolutionService`, we discovered:
- The logic could be extracted into `DynamicCoverageCalculator`
- The extracted logic was cleaner and more testable
- The full service became simpler

**Before testing**: 642-line service with complex nested logic
**After extracting for testability**: Multiple focused services, each <300 lines

### Lesson for v2

**Test-driven development works.** Not just for correctness, but for design:

```java
// Start with what you want to test
@Test
public void testDynamicCoverageWhenBodyOnlyPersonWorks() {
    // If this test is hard to write, the design is probably wrong
    var calculator = new DynamicCoverageCalculator();
    var effective = calculator.getEffectiveShiftType(assignment, date);
    assertEquals("ONBody", effective);
}
```

This forces you to:
1. Keep units small (one responsibility)
2. Minimize dependencies (injectable, mockable)
3. Make behavior explicit (not hidden in complex methods)

---

## 11. Person Registry Synchronization is Elegant

### What Worked

```yaml
# persons.yaml (version controlled)
- name: John Smith
  email: john.smith@hospital.org
  specialty: BOTH

- name: Jane Doe
  email: jane.doe@hospital.org
  specialty: BODY_ONLY
```

```java
@Startup
public class PersonRegistryService {
    @Transactional
    public void reconcileRegistry() {
        // Load from YAML
        List<Person> canonical = loadFromYaml("persons.yaml");
        // Sync with database
        syncDatabase(canonical);
    }
}
```

**Benefits**:
- Version controlled (Git shows history)
- Reviewable (see who changed what)
- Rollbackable (checkout previous version)
- No manual database edits
- Reconciliation on every startup

### Lesson for v2

Use YAML/JSON for configuration that:
- Changes infrequently
- Needs audit trail
- Affects system behavior

Do NOT use for:
- Passwords (use env vars)
- API keys (use secrets)
- Frequently-changing values (use database)

---

## 12. Simple is Better Than Clever

### Example from v1

Early design tried to be too clever:
```java
// Tries to be too smart
public String getEffectiveShiftType(
    Assignment a, LocalDate date,
    List<Person> specialtyCache,
    Map<LocalDate, List<Assignment>> overnightCache,
    boolean forceRecalculate) {
    // 15 parameters, hard to understand
}
```

Better approach:
```java
// Simple, focused
public String getEffectiveShiftType(Assignment a, LocalDate date) {
    // One job: return effective shift type
}

// If performance needed, optimize at call site:
Map<Long, String> effectiveTypes = calculator.getEffectiveShiftTypes(
    assignmentsForDate, date
);
```

### Lesson for v2

**Prefer clarity over cleverness:**
- Simple code is easier to test
- Simple code is easier to debug
- Simple code is easier to modify
- Simple code is easier to review

Optimization (complexity) should come ONLY after:
1. It works (pass tests)
2. It's measured (you know it's slow)
3. It's necessary (makes measurable difference)

---

## Summary: Key Principles for v2

| Principle | Why | How |
|-----------|-----|-----|
| **Domain-Driven** | Aligns with business, survives requirement changes | Spend time modeling with stakeholders |
| **Lazy Evaluation** | Prevents stale data | Calculate on query, batch for performance |
| **Atomic Batches** | Prevents data loss | Transaction per batch, not per row |
| **Extensible Validation** | Guides users, prevents disasters | Collect all errors, categorize by severity |
| **Security First** | No after-the-fact fixes | Tests for every security requirement |
| **Test Pyramid** | Fast feedback, good design | 60% unit, 30% integration, 10% e2e |
| **Budget Refactoring** | Prevents technical debt spiral | 25% of sprint for cleaning |
| **Config as Code** | Secure, auditable, reviewable | Env vars + YAML, never hardcode |
| **Document Decisions** | Future-proofs project | Why, not just how |
| **Test-Driven Design** | Better architecture | Tests drive design |
| **YAML for Config** | Version controlled, auditable | But not for secrets |
| **Simple Over Clever** | Sustainable code | Optimize after measuring |

These principles transformed a "works but fragile" system (v0) into a solid, maintainable system (v1).
Applying them rigorously in v2 will create a system that can last 5+ years.
