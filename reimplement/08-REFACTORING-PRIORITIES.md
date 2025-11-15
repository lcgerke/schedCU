# Refactoring Priorities: Execution Plan

Prioritized action plan for fixing technical debt and preparing v2.

---

## Phase 1: Critical Fixes (Before Production) - 1 Week

These MUST be fixed before any production deployment.

### Week 1 Sprint

**Team**: 2 people
**Sprint Goal**: Production-safe security and data integrity

#### Day 1: Security Bypass Fix

**Task 1.1: Remove @PermitAll from admin endpoints** (2 hours)

```java
// Find all instances
grep -r "@PermitAll" src/main/java/org/hospital/radiology/schedule/api/

// Replace with @RolesAllowed("ADMIN")
AdminResource.java lines: 92, 118, 149, 197, 303
```

**Checklist**:
- [ ] All 5 instances replaced
- [ ] Code compiles
- [ ] Server starts

**Task 1.2: Add security test for each endpoint** (2 hours)

```java
// Create AdminEndpointSecurityTest.java
@Test public void testAdminEndpointsRequireAuth() { ... }
@Test public void testAdminEndpointsRequireAdminRole() { ... }

// Run test
mvn test -Dtest=AdminEndpointSecurityTest
```

**Checklist**:
- [ ] 5 endpoints tested
- [ ] All tests pass
- [ ] Added to CI pipeline

#### Day 2: Environment Configuration

**Task 2.1: Move credentials to environment variables** (2 hours)

```properties
# Before
quarkus.datasource.password=postgres

# After
quarkus.datasource.password=${DB_PASSWORD:postgres}
```

Files to update:
- `application.properties` (remove hardcoded values)
- Create `.env.example` with placeholders
- Update deployment docs

**Checklist**:
- [ ] All credentials use env vars
- [ ] `.env` added to `.gitignore`
- [ ] `.env.example` created
- [ ] Docs updated
- [ ] Tests pass with env vars

**Task 2.2: Validate configuration on startup** (2 hours)

```java
@ApplicationScoped
public class ConfigValidator {
    @PostConstruct
    void validate() {
        if (required && !configured) {
            throw new IllegalStateException("Required config missing");
        }
    }
}
```

**Checklist**:
- [ ] App fails cleanly if required config missing
- [ ] Error message points to which config is missing
- [ ] Docs explain all required variables

#### Day 3-4: File Upload Validation

**Task 3.1: Add file upload validation** (4 hours)

Implement from `03-SECURITY-GAPS.md`:
- Content-type validation
- File size limit (50MB)
- ODS format validation
- XXE protection

**Checklist**:
- [ ] Accepts valid .ods files
- [ ] Rejects wrong file types
- [ ] Rejects files > 50MB
- [ ] XXE attack blocked
- [ ] Tests verify all above

**Task 3.2: Test with malicious files** (1 hour)

```bash
# Test 1: Wrong file type
curl -F "file=@image.jpg" http://localhost:8081/api/admin/import/ods
# Should get 400

# Test 2: Too large (create 100MB file)
dd if=/dev/zero of=large.ods bs=1M count=100
curl -F "file=@large.ods" http://localhost:8081/api/admin/import/ods
# Should get 413

# Test 3: XXE attack (in test, not manual)
@Test public void testXXEIsBlocked() { ... }
```

**Checklist**:
- [ ] Wrong types rejected
- [ ] Large files rejected
- [ ] XXE blocked
- [ ] Temp files cleaned up

#### Day 5: Verification & Testing

**Task 4.1: Verify no hardcoded secrets in Git** (1 hour)

```bash
# Check entire history
git log --all --full-history -S "password=" | wc -l
# Should return 0

git log --all --full-history -S "apikey=" | wc -l
# Should return 0

# Check for TODO security markers
grep -r "TODO.*security\|TODO.*password\|@PermitAll" src/main/java/
# Should return nothing
```

**Checklist**:
- [ ] No hardcoded credentials in current code
- [ ] No credentials in Git history
- [ ] No security TODOs in production code
- [ ] No @PermitAll markers

**Task 4.2: Security regression test** (2 hours)

```java
@QuarkusTest
public class ProductionSecurityTest {
    @Test public void allAdminEndpointsRequireAuth() { ... }
    @Test public void noPublicAdminEndpoints() { ... }
    @Test public void fileUploadValidation() { ... }
    @Test public void databaseCredentialsEnvironment() { ... }
    @Test public void noHardcodedSecrets() { ... }
}

mvn test -Dtest=ProductionSecurityTest
```

**Checklist**:
- [ ] All tests pass
- [ ] Added to CI
- [ ] Will run on every commit

#### Day 6: Database Cleanup

**Task 5.1: Remove deprecated reassignedShiftType column** (3 hours)

```sql
-- db/migration/V6__Remove_Deprecated_ReassignedShiftType.sql
ALTER TABLE assignment DROP COLUMN reassigned_shift_type;
```

**Steps**:
1. Create migration file
2. Update Assignment.java entity
3. Search codebase for references:
   ```bash
   grep -r "reassignedShiftType" src/
   # Should be empty after cleanup
   ```
4. Test migration: `mvn test`

**Checklist**:
- [ ] Migration file created
- [ ] Entity updated
- [ ] No references in code
- [ ] Tests pass
- [ ] Can rollback if needed

**Task 5.2: Remove dead code references** (1 hour)

```bash
# Search for any remaining dead code
grep -r "reassigned\|deprecated" src/main/java/ --include="*.java"
```

**Checklist**:
- [ ] No dead code references
- [ ] Code compiles
- [ ] Tests pass

#### Day 7: Testing & Documentation

**Task 6.1: Final verification** (2 hours)

```bash
# Full test suite
mvn clean test

# Build jar
mvn clean package

# Security scanning (if available)
mvn org.owasp:dependency-check-maven:check

# Check for security warnings
# Review any vulnerability reports
```

**Checklist**:
- [ ] All tests pass
- [ ] Build succeeds
- [ ] No security warnings
- [ ] Can start application
- [ ] Swagger docs available

**Task 6.2: Update deployment docs** (1 hour)

Create/update `DEPLOYMENT.md`:
```markdown
# Deployment Guide

## Required Environment Variables
- DB_PASSWORD
- AMION_FILE_ID
- SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS
- JWT_SECRET

## Security Checklist
- [ ] No hardcoded credentials
- [ ] File uploads validated
- [ ] Admin endpoints protected
- [ ] SSL/TLS enabled
- [ ] Logs don't contain passwords

## First-Time Setup
1. Set environment variables
2. Run migrations
3. Create admin user
4. Test endpoints
```

**Checklist**:
- [ ] Docs complete
- [ ] Includes security requirements
- [ ] Team has read and understood

**Release**:
- Tag: `v1.0.0-production-ready`
- Message: "Security fixes, removed deprecated code, production-safe"
- Branch to `main`, create release

---

## Phase 2: High-Priority Fixes (Next 2 Weeks)

These prevent production incidents but are not strictly critical.

### Week 2-3: Core Functionality Tests & Performance

**Goal**: Add test coverage and fix performance issues

#### Week 2: Test Coverage

**Task 2.1: Add unit tests for ODSImportService** (3 days)

```java
@QuarkusTest
public class ODSImportServiceTest {
    @Test public void importValidODSFile() { ... }
    @Test public void rejectUnknownShiftType() { ... }
    @Test public void validateCoverageRequirements() { ... }
    @Test public void handleMissingDates() { ... }
}
```

Expected: 30-50 tests covering:
- Valid file parsing
- Error cases
- Edge cases (weekends, holidays)
- Validation rules

**Task 2.2: Add unit tests for AmionImportService** (3 days)

```java
@QuarkusTest
public class AmionImportServiceTest {
    @Test public void scrapeValidAmionHTML() { ... }
    @Test public void handleHTMLParseErrors() { ... }
    @Test public void detectDuplicateBatches() { ... }
    @Test public void handleUnknownPeople() { ... }
}
```

Expected: 30-40 tests

**Task 2.3: Add unit tests for DynamicCoverageCalculator** (2 days)

```java
@QuarkusTest
public class DynamicCoverageCal culatorTest {
    @Test public void bodyOnlyPersonGetsCoverageAssignment() { ... }
    @Test public void otherPersonGetsNeuoroCoverageWhenBodyOnlyPresent() { ... }
    @Test public void noReassignmentWhenNoBodyOnlyPerson() { ... }
}
```

Expected: 20-30 tests

#### Week 3: Performance Fixes

**Task 3.1: Fix N+1 query in DynamicCoverageCalculator** (2 days)

From `04-PERFORMANCE-ISSUES.md`:
```java
// Add batch method
public Map<Long, String> getEffectiveShiftTypes(
    List<Assignment> assignments, LocalDate date) { ... }

// Update call sites
Map<Long, String> effectiveTypes = calculator.getEffectiveShiftTypes(
    assignments, date
);
```

**Task 3.2: Add pagination to list endpoints** (2 days)

```java
@GetMapping("/versions?page=0&size=20")
public Response getVersions(
    @QueryParam("page") @DefaultValue("0") int page,
    @QueryParam("size") @DefaultValue("20") int size) { ... }
```

**Task 3.3: Add database indexes** (1 day)

```sql
-- db/migration/V7__Add_Performance_Indexes.sql
CREATE INDEX idx_assignment_date ON assignment(schedule_date);
CREATE INDEX idx_shift_instance_date ON shift_instance(schedule_date);
```

**Verification**:
```bash
# Before: 21 queries for 1 day
# After: 2 queries for 1 day
# Improvement: 10.5× faster

# Add test to prevent regression:
@Test public void batchCalculationIsOneQuery() {
    // Verify only 2 queries (1 for assignments, 1 for overnight check)
}
```

---

## Phase 3: Medium-Priority Improvements (Weeks 4-6)

### Week 4: Code Refactoring

**Task 4.1: Extract long methods** (3 days)

From `01-TECHNICAL-DEBT.md`:
- `CoverageResolutionService.applyDynamicReassignment()` (125 lines → 5 methods)
- `AmionScraper.parseDocument()` (84 lines → 3 methods)
- `CoverageResolutionService.resolveCoverageForDateRange()` (98 lines → 4 methods)

Extract pattern:
```java
// Before: 125 lines in one method
public void applyDynamicReassignment(ScrapeBatch batch) { ... }

// After: Clear, testable methods
private List<ShiftInstance> findUnassignedShifts(ScheduleVersion v) { ... }
private List<Person> findCandidates(ShiftInstance shift) { ... }
private Person selectBestCandidate(ShiftInstance s, List<Person> c) { ... }
private void createAssignment(ShiftInstance s, Person p) { ... }
```

**Task 4.2: Replace magic strings with constants** (2 days)

```java
// Before
if (shift.equals("ON1"))  // ❌ Magic string

// After
if (shift.equals(ShiftType.ON_1.code()))  // ✅ Constant

// Strengthen ShiftType enum
public enum ShiftType {
    ON_1("ON1", "Overnight 1", StudyType.NEURO, 8),
    ON_2("ON2", "Overnight 2", StudyType.NEURO, 8),
    // ... etc
}
```

### Week 5: Configuration & Documentation

**Task 5.1: Move Amion file ID to environment** (1 day)

```properties
scraping.amion.file-id=${AMION_FILE_ID:}
```

Validate on startup:
```java
@PostConstruct
void validate() {
    if (fileId.isEmpty()) {
        throw new IllegalStateException("AMION_FILE_ID required");
    }
}
```

**Task 5.2: Add comprehensive documentation** (3 days)

- Architecture decision records (ADRs)
- Data flow diagrams
- Troubleshooting guide
- Runbook for common operations
- API documentation

### Week 6: Testing Infrastructure

**Task 6.1: Add rate limiting** (2 days)

```java
@PostMapping("/login")
public Response login(LoginRequest request) {
    if (isAccountLocked(request.email)) {
        return Response.status(429).entity("Too many attempts").build();
    }
    // ... login logic
}
```

**Task 6.2: Add async processing for long tasks** (3 days)

```java
@PostMapping("/workflow")
public Response executeWorkflow(...) {
    String jobId = UUID.randomUUID().toString();
    workflowService.executeAsync(jobId, file);
    return Response.accepted().entity(new JobResponse(jobId)).build();
}

@GetMapping("/workflow/{jobId}")
public Response getStatus(@PathParam("jobId") String jobId) {
    return Response.ok(workflowService.getStatus(jobId)).build();
}
```

---

## Phase 4: Nice-to-Have Improvements (Weeks 7+)

Lower priority, can be deferred:

- [ ] Circuit breaker for Amion scraping (resilience4j)
- [ ] Query result caching (redis)
- [ ] GraphQL endpoint (complex queries)
- [ ] Event sourcing (audit trail)
- [ ] Property-based testing (jqwik)
- [ ] Distributed locking (multi-instance)
- [ ] Performance monitoring dashboard

---

## Implementation Checklist

### Before Starting
- [ ] Team trained on technical debt overview
- [ ] Everyone read `02-WHAT-WORKED.md` (understand good patterns)
- [ ] Everyone read `01-TECHNICAL-DEBT.md` (understand problems)
- [ ] Everyone read `03-SECURITY-GAPS.md` (understand critical fixes)

### During Each Phase
- [ ] Create branch for each task
- [ ] Write tests FIRST (TDD)
- [ ] Code review before merge
- [ ] Verify tests pass
- [ ] Update documentation
- [ ] Merge to main

### Verification After Each Phase
- [ ] Full test suite passes: `mvn clean test`
- [ ] Build succeeds: `mvn clean package`
- [ ] No security warnings: `mvn owasp:check` (if integrated)
- [ ] Application starts: `java -jar target/quarkus-app/quarkus-run.jar`
- [ ] Swagger docs available: `http://localhost:8081/q/swagger-ui`

### Release Checklist
- [ ] Changelog updated
- [ ] Version bumped (semantic versioning)
- [ ] Git tag created
- [ ] Release notes written
- [ ] Deployment docs updated
- [ ] Team notified

---

## Risk Assessment

| Phase | Risk | Mitigation |
|-------|------|-----------|
| 1 (Security) | High | Extensive testing, code review, security audit |
| 2 (Tests) | Medium | Run full suite, no merge without tests |
| 3 (Refactoring) | Low | Tests provide coverage, refactor one method at a time |
| 4 (Improvements) | Low | No impact on core functionality |

---

## Success Metrics

After complete refactoring, the system should have:

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Test coverage | 60% | 85%+ | Measured by IDE/tool |
| Critical security issues | 5 | 0 | Security audit |
| Long methods (>100 LOC) | 3 | 0 | Code review |
| Hardcoded credentials | 3 | 0 | Git search |
| N+1 queries | 1 (known) | 0 | Performance tests |
| API endpoints protected | 4/5 | 5/5 | Security tests |
| Documentation pages | 12 | 18+ | Docs directory |
| Code review feedback | High | Low | Measured by PRs |

---

## Timeline Summary

```
Week 1: Phase 1 (Critical fixes)
  │
  ├─ Security bypass ──────┐
  ├─ Environment config ───┼─→ Production Safe
  ├─ File validation ──────┤   (Tag v1.0.0)
  └─ DB cleanup ──────────┘

Week 2-3: Phase 2 (Tests + Performance)
  │
  ├─ ODS import tests ──────┐
  ├─ Amion import tests ────┼─→ High Reliability
  ├─ Coverage calculator ───┤   (Tag v1.1.0)
  ├─ N+1 fix ──────────────┤
  ├─ Pagination ───────────┤
  └─ Indexes ──────────────┘

Week 4-6: Phase 3 (Refactoring + Docs)
  │
  ├─ Long method extraction ┐
  ├─ Magic string constants ├─→ Maintainability
  ├─ Config improvements ───┤   (Tag v1.2.0)
  ├─ Documentation ────────┤
  └─ Async processing ─────┘

Week 7+: Phase 4 (Nice-to-have)
        (Defer if needed)

Total: 6 weeks for comprehensive improvement
```

---

## Resource Allocation

**Recommended team**: 2 people (could do in 3 weeks with 3 people)

**Effort breakdown**:
- Phase 1: 40 hours (1 person-week)
- Phase 2: 60 hours (1.5 person-weeks)
- Phase 3: 80 hours (2 person-weeks)
- Phase 4: Open-ended

**Without dedicated team**:
- Spread over 12 weeks (5-10 hours/week)
- Assign 1 task per team member
- Review in pairs

---

## Next Steps

1. **This week**: Team review of all reimplement/ documents
2. **Next week**: Plan Phase 1 in detail (story points, subtasks)
3. **Following week**: Begin Phase 1 implementation
4. **Weekly**: Check-in on progress, adjust if needed
5. **End of Phase**: Retrospective and lessons learned

Good luck! 🚀
