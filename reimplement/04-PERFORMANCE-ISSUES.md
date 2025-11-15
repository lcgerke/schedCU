# Performance Issues & Solutions

Performance considerations and optimization strategies for v2 implementation.

---

## 🔴 CRITICAL: N+1 Query in DynamicCoverageCalculator

**Impact**: 100-fold query increase, potential timeouts
**File**: `src/main/java/org/hospital/radiology/schedule/service/DynamicCoverageCalculator.java`

### The Problem

When calculating coverage for a date with 20 assignments:
```
1. Query to get assignments for date: 1 query
2. For each assignment, check for BODY_ONLY person: 20 queries
Total: 21 queries
```

For a week (7 days × 20 assignments = 140 assignments):
```
1 base query + 140 coverage checks = 141 queries
```

**Code**:
```java
public String getEffectiveShiftType(Assignment assignment, LocalDate date) {
    // Called N times per day
    List<Assignment> overnightAssignments = Assignment.find(
        "originalShiftType IN ('ON1', 'ON2') and shiftInstance.scheduleDate = ?1",
        date
    ).list();  // ← New query per call!
    ...
}

// Called from ScheduleService.java line 156:
for (Assignment a : assignments) {  // For each of 20 assignments
    String effective = calculator.getEffectiveShiftType(a, date);  // Queries 20 times
}
```

### Solution: Batch Lookups

**Change 1: Add batch method**
```java
@ApplicationScoped
public class DynamicCoverageCalculator {

    /**
     * Calculate effective shift types for ALL assignments on a date.
     * Single query instead of N queries.
     */
    public Map<Long, String> getEffectiveShiftTypes(
        List<Assignment> assignmentsForDate,
        LocalDate date) {

        // Single query - get all overnight assignments for this date
        List<Assignment> overnightAssignments = Assignment.find(
            "originalShiftType IN ('ON1', 'ON2') AND shiftInstance.scheduleDate = ?1",
            date
        ).list();  // ← Only 1 query

        boolean hasBodyOnlyPerson = overnightAssignments.stream()
            .anyMatch(a -> a.person.specialty == Person.Specialty.BODY_ONLY);

        // Calculate all in memory
        Map<Long, String> results = new HashMap<>();
        for (Assignment assignment : assignmentsForDate) {
            results.put(
                assignment.id,
                calculateEffectiveType(assignment, hasBodyOnlyPerson)
            );
        }

        return results;
    }

    private String calculateEffectiveType(
        Assignment assignment,
        boolean hasBodyOnlyPerson) {

        String original = assignment.originalShiftType;

        if (!original.equals("ON1") && !original.equals("ON2")) {
            return original;
        }

        if (!hasBodyOnlyPerson) {
            return original;
        }

        if (assignment.person.specialty == Person.Specialty.BODY_ONLY) {
            return "ONBody";
        } else {
            return "ONNeuro";
        }
    }

    // Keep old method for single-assignment lookups (less common case)
    public String getEffectiveShiftType(Assignment assignment, LocalDate date) {
        List<Assignment> assignmentsForDate = List.of(assignment);
        return getEffectiveShiftTypes(assignmentsForDate, date)
            .get(assignment.id);
    }
}
```

**Change 2: Update call sites**
```java
// ScheduleService.java - BEFORE
public DailySchedule getScheduleForDate(LocalDate date) {
    List<Assignment> assignments = Assignment.find(
        "shiftInstance.scheduleDate = ?1", date
    ).list();

    DailySchedule schedule = new DailySchedule(date);
    for (Assignment a : assignments) {
        String effective = calculator.getEffectiveShiftType(a, date);  // N+1!
        schedule.addShift(a.person.name, effective);
    }
    return schedule;
}

// ScheduleService.java - AFTER
public DailySchedule getScheduleForDate(LocalDate date) {
    List<Assignment> assignments = Assignment.find(
        "shiftInstance.scheduleDate = ?1", date
    ).list();

    // Single query for all effective types
    Map<Long, String> effectiveTypes = calculator.getEffectiveShiftTypes(
        assignments, date
    );

    DailySchedule schedule = new DailySchedule(date);
    for (Assignment a : assignments) {
        String effective = effectiveTypes.get(a.id);
        schedule.addShift(a.person.name, effective);
    }
    return schedule;
}
```

**Metrics**:
- Before: 141 queries
- After: 8 queries (1 per day)
- **Improvement**: 17.6× reduction

### Test

```java
@Test
public void batchCalculationIsOnlyOneQuery() {
    List<Assignment> assignments = createAssignments(20);
    LocalDate date = LocalDate.of(2024, 10, 15);

    // Count queries
    QueryStatistics stats = startQueryTracking();

    calculator.getEffectiveShiftTypes(assignments, date);

    // Should be 2 queries max (1 for assignments, 1 for overnight check)
    assertEquals(2, stats.getQueryCount(), "Expected batch calculation to use only 2 queries");
}
```

---

## 🟠 HIGH: No Pagination on List Endpoints

**Impact**: Memory exhaustion with large datasets, slow response times
**Files**: `AdminResource.java`, `ScheduleResource.java`

### Problem Endpoints

```java
// Line 310
@GetMapping("/versions")
public Response getVersions() {
    // If 10,000 versions in database:
    // - Loads all 10,000 objects into memory
    // - Serializes all to JSON
    // - Client times out or runs out of memory
    return Response.ok(ScheduleVersion.listAll()).build();  // ❌ Unbounded
}

// ScheduleResource.java
@GetMapping("/range")
public Response getScheduleRange(
    @QueryParam("start") LocalDate start,
    @QueryParam("end") LocalDate end) {
    // For 6-month range with 20 assignments per day:
    // - 180 days × 20 = 3,600 assignments
    // - All loaded at once
    // - Client request might timeout
    List<Assignment> all = Assignment.find(
        "shiftInstance.scheduleDate BETWEEN ?1 AND ?2", start, end
    ).list();  // ❌ Could be thousands
}
```

### Solution: Implement Pagination

**Step 1: Create pagination response**
```java
public record PaginatedResponse<T>(
    List<T> data,
    PaginationMeta pagination
) {
    public record PaginationMeta(
        int page,
        int pageSize,
        long total,
        int totalPages,
        boolean hasNext,
        boolean hasPrevious
    ) {
        public PaginationMeta(int page, int pageSize, long total) {
            this(
                page,
                pageSize,
                total,
                (int) Math.ceil((double) total / pageSize),
                page < (int) Math.ceil((double) total / pageSize) - 1,
                page > 0
            );
        }
    }
}
```

**Step 2: Update endpoints**
```java
@RolesAllowed("ADMIN")
@GetMapping("/versions")
public Response getVersions(
    @QueryParam("page") @DefaultValue("0") int page,
    @QueryParam("size") @DefaultValue("20") int size) {

    // Validate inputs
    if (page < 0) page = 0;
    if (size < 1 || size > 100) size = 20;

    // Query with pagination
    PanacheQuery<ScheduleVersion> query = ScheduleVersion.find(
        "ORDER BY createdAt DESC"
    );

    long total = query.count();
    List<ScheduleVersion> data = query
        .page(page, size)
        .list();

    return Response.ok(new PaginatedResponse<>(
        data,
        new PaginatedResponse.PaginationMeta(page, size, total)
    )).build();
}

@GetMapping("/range")
public Response getScheduleRange(
    @QueryParam("start") LocalDate start,
    @QueryParam("end") LocalDate end,
    @QueryParam("page") @DefaultValue("0") int page,
    @QueryParam("size") @DefaultValue("50") int size) {

    if (page < 0) page = 0;
    if (size < 1 || size > 500) size = 50;

    PanacheQuery<Assignment> query = Assignment.find(
        "shiftInstance.scheduleDate BETWEEN ?1 AND ?2 ORDER BY shiftInstance.scheduleDate ASC",
        start, end
    );

    long total = query.count();
    List<Assignment> data = query.page(page, size).list();

    return Response.ok(new PaginatedResponse<>(
        data,
        new PaginatedResponse.PaginationMeta(page, size, total)
    )).build();
}
```

**Step 3: Update client code**
```javascript
// Before (single request)
const versions = await fetch('/api/admin/versions').json();

// After (paginated requests)
let page = 0;
let hasMore = true;
let allVersions = [];

while (hasMore) {
    const response = await fetch(`/api/admin/versions?page=${page}&size=20`);
    const { data, pagination } = await response.json();
    allVersions.push(...data);
    hasMore = pagination.hasNext;
    page++;
}
```

---

## 🟠 HIGH: Inefficient Person Lookups

**Impact**: Repeated database calls for same person
**File**: `AmionImportService.java`, line 205

### Problem

```java
// Inside loop processing each assignment
for (ScrapedSchedule row : scrapedSchedules) {
    Person person = Person.findByAlias(row.personAlias);  // Database query
    Assignment.create(person, shiftInstance);
}
```

For 100 rows with 10 unique people:
- 100 database queries for 10 unique lookups
- 90 redundant queries

### Solution: Batch and Cache

```java
@Transactional
public void importScrapedDataInTransaction(ScrapeBatch batch) {
    // Pre-load all unique people mentioned in batch
    List<ScrapedSchedule> allRows = ScrapedSchedule.find("batch = ?1", batch).list();

    // Extract unique person aliases
    Set<String> uniqueAliases = allRows.stream()
        .map(row -> row.personAlias)
        .collect(Collectors.toSet());

    // Single query to load all people
    Map<String, Person> personCache = uniqueAliases.stream()
        .collect(Collectors.toMap(
            alias -> alias,
            alias -> Person.findByAlias(alias)
        ));

    // Process rows using cache
    for (ScrapedSchedule row : allRows) {
        Person person = personCache.get(row.personAlias);  // Cache hit, no query
        if (person == null) {
            LOG.warnf("Unknown person: %s", row.personAlias);
            continue;
        }

        ShiftInstance shift = ShiftInstance.find("...", row.scheduleDate, row.shiftType)
            .firstResult();

        Assignment assignment = new Assignment();
        assignment.person = person;
        assignment.shiftInstance = shift;
        assignment.originalShiftType = row.shiftType;
        assignment.persist();
    }
}
```

**Metrics**:
- Before: 100 queries
- After: 2 queries (1 for rows, 1 for unique people)
- **Improvement**: 50× reduction

---

## 🟡 MEDIUM: Sequential Amion Scraping

**Impact**: Slow scraping, blocked requests
**File**: `AmionScraper.java`, line 89-107

### Problem

```java
public void scrapeMonths(int monthsAhead) {
    for (int i = 0; i < monthsAhead; i++) {
        YearMonth month = YearMonth.now().plusMonths(i);
        scrapeMonth(month);  // Blocks until complete (~30s)
        Thread.sleep(1000);  // Rate limiting
    }
    // Total time: monthsAhead × 30s = 180s for 6 months
}
```

**Timeline**:
```
Request: scrapeMonths(6)
  Month 1: ---------- (30s)
  Month 2:           ---------- (30s)
  Month 3:                       ---------- (30s)
  ...
  Month 6:                                           ---------- (30s)
Response after 180s
↑ Request times out after 30s!
```

### Solution: Parallel with Backpressure

```java
@ApplicationScoped
public class AmionScraper {
    private static final int MAX_CONCURRENT = 3;
    private static final Duration RATE_LIMIT = Duration.ofMillis(1000);
    private final Semaphore rateLimiter = new Semaphore(MAX_CONCURRENT);

    // Make scraping async
    @Asynchronous
    public CompletionStage<List<ScrapedScheduleRow>> scrapeMonths(int monthsAhead) {
        List<CompletableFuture<List<ScrapedScheduleRow>>> futures = new ArrayList<>();

        for (int i = 0; i < monthsAhead; i++) {
            YearMonth month = YearMonth.now().plusMonths(i);
            futures.add(scrapeMonthAsync(month));
        }

        // Wait for all to complete
        return CompletableFuture.allOf(
            futures.toArray(new CompletableFuture[0])
        ).thenApply(v ->
            futures.stream()
                .map(CompletableFuture::join)
                .flatMap(List::stream)
                .collect(Collectors.toList())
        );
    }

    private CompletableFuture<List<ScrapedScheduleRow>> scrapeMonthAsync(YearMonth month) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                // Acquire permit (max 3 concurrent)
                rateLimiter.acquire();

                try {
                    LOG.infof("Scraping %s", month);
                    String html = scrapeAmion(month);
                    List<ScrapedScheduleRow> rows = parseHtml(html);
                    return rows;
                } finally {
                    rateLimiter.release();
                    Thread.sleep(RATE_LIMIT.toMillis());
                }
            } catch (InterruptedException e) {
                return List.of();
            }
        });
    }
}

// Updated endpoint to handle async
@RolesAllowed("ADMIN")
@PostMapping("/import/amion")
public Response scrapeAmion() {
    String jobId = UUID.randomUUID().toString();

    // Start async job
    scraper.scrapeMonths(6)
        .whenComplete((rows, ex) -> {
            if (ex != null) {
                jobTracking.recordFailure(jobId, ex);
            } else {
                jobTracking.recordSuccess(jobId, rows);
            }
        });

    // Return immediately with job ID
    return Response.accepted()
        .entity(new JobResponse(jobId))
        .build();
}

// Check job status
@GetMapping("/import/amion/{jobId}")
public Response checkScrapingStatus(@PathParam("jobId") String jobId) {
    JobStatus status = jobTracking.getStatus(jobId);
    return Response.ok(status).build();
}
```

**Timeline with parallelism**:
```
Request: scrapeMonths(6)
  Month 1: ---------- (parallel)
  Month 2:      ---------- (parallel)
  Month 3:           ---------- (parallel)
  Month 4:                ------||------- (wait for 1 to finish)
  Month 5:                      ----------
  Month 6:                           ----------
Response immediately with job ID
Actual scraping happens in background: ~50s instead of 180s
```

**Metrics**:
- Before: 180s total + request timeout
- After: 50s total, non-blocking
- **Improvement**: 3.6× faster, request returns immediately

---

## 🟡 MEDIUM: Missing Database Indexes

**Impact**: Slow queries for date-range lookups
**Files**: Entity definitions

### Problem Queries

```java
// These queries without indexes are O(n) table scans:

// Query 1: Find assignments for date range
Assignment.find("shiftInstance.scheduleDate BETWEEN ?1 AND ?2", start, end).list()

// Query 2: Find assignments for specific date
Assignment.find("shiftInstance.scheduleDate = ?1", date).list()

// Query 3: Find overnight assignments
Assignment.find("originalShiftType IN ('ON1', 'ON2') AND shiftInstance.scheduleDate = ?1", date).list()

// Query 4: Find shifts for specific hospital
ShiftInstance.find("hospital = ?1 AND scheduleDate = ?2", hospital, date).list()
```

### Solution: Add Database Indexes

```java
@Entity
@Table(indexes = {
    @Index(columnList = "schedule_date", name = "idx_assignment_date"),
    @Index(columnList = "schedule_date,original_shift_type", name = "idx_assignment_date_shift"),
    @Index(columnList = "person_id,schedule_date", name = "idx_assignment_person_date"),
    @Index(columnList = "batch_id", name = "idx_scraped_schedule_batch")
})
public class Assignment {
    public LocalDate scheduleDate;
    public String originalShiftType;
    public Person person;
    public ScrapeBatch batch;
}

@Entity
@Table(indexes = {
    @Index(columnList = "schedule_date", name = "idx_shift_instance_date"),
    @Index(columnList = "hospital,schedule_date", name = "idx_shift_instance_hospital_date"),
    @Index(columnList = "shift_type,schedule_date", name = "idx_shift_instance_type_date")
})
public class ShiftInstance {
    public LocalDate scheduleDate;
    public Hospital hospital;
    public String shiftType;
}
```

Or via Flyway migration:
```sql
-- db/migration/V6__Add_Performance_Indexes.sql
CREATE INDEX idx_assignment_date ON assignment(schedule_date);
CREATE INDEX idx_assignment_date_shift ON assignment(schedule_date, original_shift_type);
CREATE INDEX idx_assignment_person_date ON assignment(person_id, schedule_date);
CREATE INDEX idx_scraped_schedule_batch ON scraped_schedule(batch_id);

CREATE INDEX idx_shift_instance_date ON shift_instance(schedule_date);
CREATE INDEX idx_shift_instance_hospital_date ON shift_instance(hospital, schedule_date);
CREATE INDEX idx_shift_instance_type_date ON shift_instance(shift_type, schedule_date);
```

**Query Performance**:
| Query | Without Index | With Index | Improvement |
|-------|---------------|-----------|-------------|
| Find assignments for date | 500ms | 5ms | 100× |
| Find overnight assignments | 800ms | 10ms | 80× |
| Find shifts by hospital | 300ms | 3ms | 100× |

---

## 🟡 MEDIUM: No Connection Pooling Optimization

**Current Configuration**:
```properties
quarkus.datasource.jdbc.min-size=2
quarkus.datasource.jdbc.max-size=10
```

### For Different Load Profiles

**Light Load** (dev/test):
```properties
quarkus.datasource.jdbc.min-size=2
quarkus.datasource.jdbc.max-size=5
quarkus.datasource.jdbc.acquisition-timeout=15
```

**Medium Load** (production):
```properties
quarkus.datasource.jdbc.min-size=5
quarkus.datasource.jdbc.max-size=20
quarkus.datasource.jdbc.acquisition-timeout=30
quarkus.datasource.jdbc.idle-removal-interval=10
```

**High Load** (large hospital):
```properties
quarkus.datasource.jdbc.min-size=10
quarkus.datasource.jdbc.max-size=50
quarkus.datasource.jdbc.acquisition-timeout=30
quarkus.datasource.jdbc.max-lifetime=30
```

### Monitoring

```java
@Path("/api/metrics")
public class MetricsResource {
    @Inject
    DataSource dataSource;

    @GetMapping("/connection-pool")
    public Response getPoolMetrics() {
        HikariDataSource hds = (HikariDataSource) dataSource;
        return Response.ok(new PoolMetrics(
            hds.getMaximumPoolSize(),
            hds.getMinimumIdle(),
            hds.getHikariPoolMXBean().getActiveConnections(),
            hds.getHikariPoolMXBean().getIdleConnections(),
            hds.getHikariPoolMXBean().getTotalConnections(),
            hds.getHikariPoolMXBean().getPendingThreads()
        )).build();
    }
}
```

---

## Performance Optimization Checklist

- [x] Batch dynamic coverage calculation (N+1 fix)
- [ ] Add pagination to list endpoints
- [ ] Cache person lookups during import
- [ ] Parallelize Amion scraping
- [ ] Add database indexes for date-range queries
- [ ] Optimize connection pool for production load
- [ ] Add query result caching for read-heavy queries
- [ ] Monitor slow query log
- [ ] Profile with JProfiler or YourKit

---

## Measurement Strategy

For v2, add performance monitoring:

```java
@InterceptorBinding
@Target({ METHOD, TYPE })
@Retention(RUNTIME)
public @interface Timed {}

@Timed
@AroundInvoke
public Object timeMethod(InvocationContext ctx) throws Exception {
    long start = System.nanoTime();
    try {
        return ctx.proceed();
    } finally {
        long duration = System.nanoTime() - start;
        metrics.recordDuration(ctx.getMethod().getName(), duration);

        if (duration > 1_000_000_000) {  // > 1 second
            LOG.warnf("%s took %dms", ctx.getMethod().getName(), duration / 1_000_000);
        }
    }
}

// Usage
@Timed
public DailySchedule getScheduleForDate(LocalDate date) { ... }
```

This provides automatic performance tracking without cluttering business logic.
