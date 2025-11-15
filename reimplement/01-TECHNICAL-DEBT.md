# Technical Debt & Issues

This document catalogues all known technical debt, bugs, and design problems in v1.

---

## Critical Issues (Must Fix Before Production)

### 1. Security Bypass: Admin Endpoints Unprotected

**Severity**: 🔴 CRITICAL
**Impact**: Anyone can trigger schedule imports, promote to production
**Files**: `src/main/java/org/hospital/radiology/schedule/api/AdminResource.java`

**Problem**:
```java
// Line 92
@PermitAll  // TODO: Remove after testing
@PostMapping("/import/amion")
public Response scrapeAmion() { ... }

// Lines 118, 149, 197, 303 - similar bypasses
```

**Current State**: Multiple endpoints have TODO comments to remove `@PermitAll` before production.

**Fix**:
```java
@RolesAllowed("ADMIN")  // Enforce ADMIN role
@PostMapping("/import/amion")
public Response scrapeAmion() { ... }
```

**Test Coverage Needed**: Add security integration tests
```java
@Test
public void testAdminEndpointRequiresAuth() {
    given()
        .when().post("/api/admin/import/amion")
        .then().statusCode(401);  // Should fail without token
}
```

---

### 2. Dead Code: Deprecated Coverage Field

**Severity**: 🔴 CRITICAL
**Impact**: Database schema inconsistency, migration uncertainty
**Files**: `src/main/java/org/hospital/radiology/schedule/entity/Assignment.java`

**Problem**:
```java
public class Assignment {
    public String originalShiftType;    // ✅ Used - what Amion said
    public String reassignedShiftType;  // ❌ DEPRECATED - calculated now
}
```

The field `reassignedShiftType` is deprecated because the new `DynamicCoverageCalculator` computes coverage at query time instead of storing it.

**Current State**:
- Database still has column
- Code may check both fields
- Stale data could corrupt results

**Fix**:
1. **Database migration** to remove column:
   ```sql
   ALTER TABLE assignment DROP COLUMN reassigned_shift_type;
   ```
2. **Code cleanup**: Search for all references and remove
3. **Verify**: Ensure `DynamicCoverageCalculator` is used everywhere

**Migration Risk**: Medium (data loss possible, but field is stale)

---

### 3. N+1 Query: Dynamic Coverage Calculator

**Severity**: 🔴 CRITICAL
**Impact**: Performance degradation with 100+ assignments, potential timeout
**File**: `src/main/java/org/hospital/radiology/schedule/service/DynamicCoverageCalculator.java`

**Problem**:
```java
// Line 46-49
public String getEffectiveShiftType(Assignment assignment, LocalDate date) {
    // This is called per assignment (N times)
    List<Assignment> overnightAssignments = Assignment.find(
        "originalShiftType IN ('ON1', 'ON2') and shiftInstance.scheduleDate = ?1",
        date
    ).list();  // ← NEW QUERY PER CALL

    boolean hasBodyOnlyPerson = overnightAssignments.stream()
        .anyMatch(a -> a.person.specialty == Person.Specialty.BODY_ONLY);
    ...
}
```

**Current Usage**:
- Called in `ScheduleService.getScheduleForDate()` (line 156)
- Loops through all assignments for date
- Result: If 20 assignments on a night, 20 queries + 1 parent query = 21 total

**Fix**:
```java
public class DynamicCoverageCalculator {

    // Batch version - called ONCE per date
    public Map<String, String> getEffectiveShiftTypes(
        List<Assignment> assignmentsForDate, LocalDate date) {

        // Query once per date
        List<Assignment> overnightAssignments = Assignment.find(
            "originalShiftType IN ('ON1', 'ON2') and shiftInstance.scheduleDate = ?1",
            date
        ).list();

        boolean hasBodyOnlyPerson = overnightAssignments.stream()
            .anyMatch(a -> a.person.specialty == Person.Specialty.BODY_ONLY);

        // Apply logic to each assignment
        Map<String, String> results = new HashMap<>();
        for (Assignment a : assignmentsForDate) {
            results.put(a.id.toString(), getEffectiveType(a, hasBodyOnlyPerson));
        }
        return results;
    }

    private String getEffectiveType(Assignment a, boolean hasBodyOnly) {
        // ... existing logic
    }
}
```

**Call Site Change**:
```java
// Before (N+1)
for (Assignment a : assignments) {
    String effective = calculator.getEffectiveShiftType(a, date);
}

// After (1 query)
Map<String, String> effectiveTypes = calculator.getEffectiveShiftTypes(assignments, date);
for (Assignment a : assignments) {
    String effective = effectiveTypes.get(a.id.toString());
}
```

**Test to Add**:
```java
@Test
public void testBatchEffectiveShiftTypesIsSingleQuery() {
    // Mock and verify Assignment.find() called exactly once
    int queryCount = getQueryCount();
    calculator.getEffectiveShiftTypes(20assignments, date);
    assertTrue(getQueryCount() - queryCount == 1);
}
```

---

## High Priority Issues

### 4. Test Coverage Gaps

**Severity**: 🟠 HIGH
**Impact**: Unknown bugs in critical paths, regression risk
**Files**: Missing test files

**Missing Unit Tests**:

| Service | Status | Risk |
|---------|--------|------|
| `ODSImportService` | ❌ No tests | HIGH - File parsing is complex |
| `AmionImportService` | ❌ No tests | HIGH - Web scraping is fragile |
| `DynamicCoverageCalculator` | ❌ No tests | HIGH - Core logic, new code |
| `CoverageResolutionService` | ✅ Has tests | - |
| `ScrapeBatchValidator` | ❌ No tests | MEDIUM - Validation logic |
| `PersonRegistryService` | ❌ No tests | MEDIUM - YAML sync |
| `ScheduleService` | ❌ No tests | MEDIUM - Query logic |
| API Endpoints | ❌ No tests | MEDIUM - Contract verification |

**Fix**: Create test files:
```java
// src/test/java/.../service/ODSImportServiceTest.java
@QuarkusTest
public class ODSImportServiceTest {
    @Inject ODSImportService service;
    @Inject ScheduleVersionRepository versions;

    @Test
    public void testImportODSFile() throws IOException {
        // Test with sample ODS file
        ScheduleVersion v = service.importODSFile(testFile, startDate);
        assertEquals("MidC", v.shiftInstances[0].shiftType);
    }

    @Test
    public void testUnknownShiftTypeValidationError() {
        // Test validation catches unknown shifts
    }
}
```

**Effort**: ~3 days (80-100 tests needed)

---

### 5. Long Methods Needing Extraction

**Severity**: 🟠 HIGH
**Impact**: Hard to test, understand, maintain
**File**: `src/main/java/org/hospital/radiology/schedule/service/CoverageResolutionService.java`

**Problem Method**: `applyDynamicReassignment()` - 125 lines, 8 nested levels

```java
// Lines 46-170
public void applyDynamicReassignment(ScrapeBatch batch, ScheduleVersion version) {
    // 125 lines of nested logic for:
    // 1. Finding unassigned shifts
    // 2. Checking for BODY_ONLY specialists
    // 3. Applying reassignment rules
    // 4. Validating coverage
    // 5. Handling edge cases
}
```

**Refactor Into**:
```java
public void applyDynamicReassignment(ScrapeBatch batch, ScheduleVersion version) {
    var unassignedShifts = findUnassignedShifts(version);
    for (ShiftInstance shift : unassignedShifts) {
        var candidates = findCandidates(shift);
        var assignment = selectBestCandidate(shift, candidates);
        if (assignment != null) {
            createAssignment(shift, assignment);
        }
    }
}

private List<ShiftInstance> findUnassignedShifts(ScheduleVersion version) { ... }

private List<Person> findCandidates(ShiftInstance shift) { ... }

private Person selectBestCandidate(ShiftInstance shift, List<Person> candidates) { ... }

private void createAssignment(ShiftInstance shift, Person person) { ... }
```

**Additional Long Methods**:
- `CoverageResolutionService.resolveCoverageForDateRange()` - 98 lines
- `AmionScraper.parseDocument()` - 84 lines (tightly coupled to HTML structure)

---

### 6. Hardcoded Configuration Values

**Severity**: 🟠 HIGH
**Impact**: Secret leakage, brittle deployment, environment-specific bugs
**Files**: Multiple

**Hardcoded Database Password**:
```properties
# application.properties, line 12
quarkus.datasource.password=postgres  # ❌ Should use ${DB_PASSWORD:postgres}
```

**Hardcoded Amion File ID**:
```properties
# application.properties, line 56
scraping.amion.file-id=!1854eed1hnew_28821_6  # ❌ What is this? Magic string?
```

**Hardcoded File Paths**:
```java
// AmionScraper.java, line 152
Files.write(Paths.get("/tmp/amion-" + month + ".html"), html.getBytes());
// ❌ Hardcoded /tmp, no cleanup, potential disk issues
```

**Fix**:
```properties
# application.properties
quarkus.datasource.password=${DB_PASSWORD:postgres}
scraping.amion.file-id=${AMION_FILE_ID:}
scraping.amion.cache-dir=${AMION_CACHE_DIR:/tmp/amion-cache}
```

```java
@ConfigProperty(name = "scraping.amion.cache-dir")
String cacheDir;

public void saveHtml(String html, String month) throws IOException {
    Path dir = Paths.get(cacheDir);
    Files.createDirectories(dir);
    Path file = dir.resolve("amion-" + month + ".html");
    Files.write(file, html.getBytes());
}
```

---

### 7. No Input Validation on File Uploads

**Severity**: 🟠 HIGH
**Impact**: XXE attacks, malicious files, DoS
**File**: `src/main/java/org/hospital/radiology/schedule/api/AdminResource.java`

**Problem**:
```java
@PostMapping("/import/ods")
public Response importODS(@FormParam("file") InputStream file) {
    // ❌ No content-type check
    // ❌ No file size limit
    // ❌ No XXE protection on ODS parsing
    // ❌ File could be any format

    DirectODSParser parser = new DirectODSParser();
    parser.parse(file);  // ← Potential XXE attack vector
}
```

**Fix**:
```java
@PostMapping("/import/ods")
public Response importODS(
    @FormParam("file") InputStream file,
    @FormParam("filename") String filename) {

    // Validate content type
    if (!filename.toLowerCase().endsWith(".ods")) {
        return Response.status(400).entity("Only .ods files allowed").build();
    }

    // Validate size (max 10MB)
    long fileSize = file.available();
    if (fileSize > 10 * 1024 * 1024) {
        return Response.status(413).entity("File too large").build();
    }

    // Parse with XXE protection
    DirectODSParser parser = new DirectODSParser();
    parser.disableXXE();  // Ensure parser is safe
    ScheduleVersion version = parser.parse(file);

    return Response.ok(version).build();
}
```

---

### 8. No Pagination on List Endpoints

**Severity**: 🟠 HIGH
**Impact**: Performance issues with large datasets, memory exhaustion
**Files**: `AdminResource.java`, `ScheduleResource.java`

**Problem Endpoints**:

```java
// Line 310: Returns ALL versions
@GetMapping("/versions")
public Response getVersions() {
    return Response.ok(ScheduleVersion.listAll()).build();  // ❌ No limit
}

// ScheduleResource.java: Date range queries
@GetMapping("/range")
public Response getScheduleRange(
    @QueryParam("start") LocalDate start,
    @QueryParam("end") LocalDate end) {
    // Could return 1000+ assignments with no pagination
}
```

**Fix**:
```java
@GetMapping("/versions")
public Response getVersions(
    @QueryParam("page") @DefaultValue("0") int page,
    @QueryParam("size") @DefaultValue("20") int size) {

    if (size > 100) size = 100;  // Cap at 100

    var query = ScheduleVersion.find(
        "order by createdAt desc"
    ).page(page, size);

    return Response.ok(new PageResponse(
        query.list(),
        query.count(),
        page,
        size
    )).build();
}
```

**Helper Class**:
```java
public class PageResponse<T> {
    public List<T> data;
    public long total;
    public int page;
    public int size;
    public int pages;

    public PageResponse(List<T> data, long total, int page, int size) {
        this.data = data;
        this.total = total;
        this.page = page;
        this.size = size;
        this.pages = (int) Math.ceil((double) total / size);
    }
}
```

---

## Medium Priority Issues

### 9. Magic Strings for Shift Types

**Severity**: 🟡 MEDIUM
**Impact**: Runtime errors, typos go undetected until tests
**Locations**: Throughout codebase

**Problem**:
```java
// Instead of:
if (shift.equals("ON1"))  // ❌ String literal
if (shift.equals("MidC")) // ❌ String literal

// Should be:
if (shift.equals(ShiftType.ON_1.code()))      // ✅ Constant
if (shift.equals(ShiftType.MID_COVERAGE.code())) // ✅ Constant
```

**Fix**: Strengthen `ShiftType` enum:
```java
public enum ShiftType {
    ON_1("ON1", "Overnight 1", StudyType.NEURO, 8),
    ON_2("ON2", "Overnight 2", StudyType.NEURO, 8),
    MID_COVERAGE("MidC", "Mid Coverage", StudyType.BOTH, 8),
    MID_LATE("MidL", "Mid-Late", StudyType.BOTH, 8),
    MID_3RD("Mid3", "Mid 3rd", StudyType.BODY, 8),
    ON_BODY("ONBody", "Overnight Body", StudyType.BODY, 8),
    ON_NEURO("ONNeuro", "Overnight Neuro", StudyType.NEURO, 8);

    private final String code;
    private final String label;
    private final StudyType studyType;
    private final int hours;

    ShiftType(String code, String label, StudyType studyType, int hours) {
        this.code = code;
        this.label = label;
        this.studyType = studyType;
        this.hours = hours;
    }

    public static ShiftType fromCode(String code) {
        for (ShiftType st : values()) {
            if (st.code.equals(code)) return st;
        }
        throw new IllegalArgumentException("Unknown shift type: " + code);
    }

    public String code() { return code; }
    public String label() { return label; }
    public StudyType studyType() { return studyType; }
    public int hours() { return hours; }
}
```

---

### 10. No Rate Limiting on Login Endpoint

**Severity**: 🟡 MEDIUM
**Impact**: Brute force attacks possible
**File**: `src/main/java/org/hospital/radiology/schedule/api/AuthResource.java`

**Problem**:
```java
@PostMapping("/login")
public Response login(LoginRequest request) {
    // ❌ No rate limiting
    // ❌ No login attempt tracking
    // ❌ No CAPTCHA on repeated failures
    User user = User.findByEmail(request.email);
    ...
}
```

**Fix**: Add rate limiter (dependency: `io.quarkus:quarkus-cache`)
```java
@PostMapping("/login")
@CacheResult(cacheName = "login-attempts")
public Response login(LoginRequest request, @CacheKey String email) {
    // Quarkus Cache annotation + rate limit interceptor
    // Or use: https://github.com/dasniko/quarkus-jwt-with-rate-limit
}
```

Or implement simple solution:
```java
public class LoginAttemptService {
    private Map<String, LoginAttempt> attempts = new ConcurrentHashMap<>();

    public boolean isBlocked(String email) {
        LoginAttempt attempt = attempts.get(email);
        if (attempt == null) return false;

        if (attempt.failureCount >= 5) {
            if (System.currentTimeMillis() - attempt.lastFailureTime < 15 * 60 * 1000) {
                return true;  // Block for 15 minutes
            }
            attempt.reset();
        }
        return false;
    }

    public void recordFailure(String email) {
        attempts.computeIfAbsent(email, k -> new LoginAttempt())
            .recordFailure();
    }
}
```

---

### 11. Inconsistent Error Response Format

**Severity**: 🟡 MEDIUM
**Impact**: API clients must handle multiple formats, confusion

**Problem**:
```java
// Some endpoints return:
Response.ok(new ApiResponse<>(data, "Success")).build()

// Others return:
Response.status(400).entity("Invalid input").build()

// Others return:
Response.serverError().entity(new ErrorResponse(error)).build()

// Still others return:
throw new IllegalArgumentException("Invalid shift type")
```

**Fix**: Create unified response wrapper:
```java
public record ApiResponse<T>(
    T data,
    ApiMetadata meta,
    ApiError error
) {
    public ApiResponse(T data) {
        this(data, new ApiMetadata(), null);
    }

    public ApiResponse(ApiError error) {
        this(null, new ApiMetadata(), error);
    }
}

public record ApiError(
    String code,
    String message,
    Map<String, Object> details
) {
    public ApiError(String code, String message) {
        this(code, message, Map.of());
    }
}

public record ApiMetadata(
    long timestamp,
    String version
) {
    public ApiMetadata() {
        this(System.currentTimeMillis(), "1.0");
    }
}

// Usage:
@PostMapping("/import/ods")
public Response importODS(...) {
    try {
        ScheduleVersion v = service.import(...);
        return Response.ok(new ApiResponse<>(v)).build();
    } catch (ValidationException e) {
        return Response.status(400)
            .entity(new ApiResponse<>(new ApiError("VALIDATION_ERROR", e.getMessage())))
            .build();
    }
}
```

---

### 12. No Async Processing for Long-Running Tasks

**Severity**: 🟡 MEDIUM
**Impact**: Blocked request threads, poor UX, timeout issues
**File**: `AdminResource.java` (all scraping/import endpoints)

**Current**:
```java
@PostMapping("/workflow")
public Response executeWorkflow(@FormParam("file") InputStream file) {
    // Blocks HTTP thread for 5-10 minutes while:
    // 1. Parsing ODS
    // 2. Scraping Amion
    // 3. Resolving coverage
    // Client gets timeout (default 30s)
    return Response.ok(orchestrator.executeCompleteWorkflow(...)).build();
}
```

**Fix**: Make async with job tracking
```java
@PostMapping("/workflow")
public Response executeWorkflow(@FormParam("file") InputStream file) {
    String jobId = UUID.randomUUID().toString();

    // Start async job
    workflowService.executeWorkflowAsync(jobId, file);

    // Return immediately
    return Response.accepted()
        .entity(new JobResponse(jobId))
        .build();
}

// Check status
@GetMapping("/workflow/{jobId}")
public Response getJobStatus(@PathParam("jobId") String jobId) {
    WorkflowJob job = workflowService.getJob(jobId);
    return Response.ok(new JobStatusResponse(
        job.id,
        job.status,  // PENDING, IN_PROGRESS, COMPLETE, FAILED
        job.progress,
        job.result
    )).build();
}

// Background execution
@ApplicationScoped
public class WorkflowService {
    @Asynchronous
    public CompletionStage<ScheduleVersion> executeWorkflowAsync(
        String jobId, InputStream file) {
        try {
            var result = orchestrator.executeCompleteWorkflow(file, ...);
            workflowJobs.put(jobId, new WorkflowJob(jobId, "COMPLETE", result));
            return CompletableFuture.completedStage(result);
        } catch (Exception e) {
            workflowJobs.put(jobId, new WorkflowJob(jobId, "FAILED", e));
            return CompletableFuture.failedStage(e);
        }
    }
}
```

---

## Low Priority Issues

### 13. Inconsistent Null Handling

**Severity**: 🟡 LOW
**Impact**: Occasional NPE, defensive programming patterns

Various methods don't consistently check for null before dereferencing:
```java
// ❌ Unsafe
Person person = Person.findByAlias(name);
person.email.send(...);  // NPE if person is null

// ✅ Safe
Person person = Person.findByAlias(name);
if (person != null) {
    person.email.send(...);
}

// ✅ Better
Optional<Person> person = Person.find("alias", name).firstResultOptional();
person.ifPresent(p -> p.email.send(...));
```

**Fix**: Adopt Optional pattern throughout:
```java
public Optional<Person> findByAlias(String alias) {
    return Person.find("alias", alias).firstResultOptional();
}
```

---

### 14. Limited Logging for Troubleshooting

**Severity**: 🟡 LOW
**Impact**: Hard to debug production issues

**Missing Context**:
- Request IDs (correlation across logs)
- Performance metrics (query times)
- Amion HTML parsing failures (why did it fail?)

**Fix**: Add structured logging
```java
public void scrapeMonth(YearMonth month) {
    String requestId = MDC.get("request-id");
    long start = System.currentTimeMillis();

    try {
        LOG.infof("[%s] Scraping Amion for %s", requestId, month);
        Document doc = scrapeAmion(month);
        long duration = System.currentTimeMillis() - start;
        LOG.infof("[%s] Scraped %s in %dms", requestId, month, duration);
    } catch (IOException e) {
        LOG.errorf(e, "[%s] Failed to scrape Amion for %s: %s",
            requestId, month, e.getMessage());
    }
}
```

---

## Summary by Priority

| Issue | Severity | Effort | Impact |
|-------|----------|--------|--------|
| Security bypass | 🔴 Critical | 2h | Data corruption risk |
| Dead code | 🔴 Critical | 4h | Schema inconsistency |
| N+1 queries | 🔴 Critical | 1d | Performance |
| Test gaps | 🟠 High | 3d | Reliability |
| Long methods | 🟠 High | 2d | Maintainability |
| Hardcoded values | 🟠 High | 4h | Deployment |
| No file validation | 🟠 High | 1d | Security |
| No pagination | 🟠 High | 1d | Scalability |
| Magic strings | 🟡 Medium | 4h | Type safety |
| No rate limit | 🟡 Medium | 4h | Security |
| Inconsistent errors | 🟡 Medium | 2d | API design |
| No async | 🟡 Medium | 2d | UX |
| Null handling | 🟡 Low | 2d | Reliability |
| Limited logging | 🟡 Low | 1d | Observability |

**Total estimated effort**: 2-3 weeks for comprehensive cleanup
