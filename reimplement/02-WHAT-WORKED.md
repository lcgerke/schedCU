# What Worked Well: Patterns to Keep

This document highlights the design patterns, architectural decisions, and implementation strategies that proved effective in v1. These should be preserved or enhanced in v2.

---

## Domain Modeling: Excellent

### Why It Works

The entity model is rich, expressive, and aligns with business concepts. Each entity has a clear responsibility and relationships are well-defined.

### Key Entities

#### 1. ScheduleVersion (Temporal Versioning)

```java
public class ScheduleVersion extends PanacheEntity {
    public ScheduleStatus status;  // STAGING, PRODUCTION, ARCHIVED
    public LocalDateTime effectiveStartTs;
    public LocalDateTime effectiveEndTs;
    public Long scrapeBatchId;  // Which batch was used
    public String validationResults;  // JSON
    public LocalDateTime createdAt;
}
```

**What Makes It Good**:
- Temporal validity (`effectiveStartTs`/`effectiveEndTs`) enables time-travel queries
- Explicit status prevents accidental promotion
- Audit trail of when each version was created
- Soft-link to `ScrapeBatch` enables traceability without hard FK constraint
- Supports multiple concurrent versions (STAGING while PRODUCTION is active)

**Use Cases Enabled**:
- Query schedule for any date in the past
- Preview new schedule before activation
- Rollback to previous version
- Understand which scrape batch created which schedule

**Recommendation for v2**: Keep this pattern. Consider adding:
- `promotedBy` (user who promoted)
- `promotedAt` (timestamp)
- `notes` (why this version was promoted)

---

#### 2. ShiftInstance (Coverage Template)

```java
public class ShiftInstance extends PanacheEntity {
    public ScheduleVersion version;
    public String shiftType;  // ON1, ON2, MidC, MidL, etc.
    public LocalDate scheduleDate;
    public LocalTime startTime;
    public LocalTime endTime;
    public Hospital hospital;
    public StudyType studyType;
    public StudySubtype studySubtype;
    public Person.Specialty specialtyConstraint;  // BODY, NEURO, BOTH
    public Integer desiredCoverage;  // How many people needed
    public Boolean isMandatory;
    public LocalDateTime createdAt;
}
```

**What Makes It Good**:
- Rich metadata (hospital, study type, time, specialty) all in one place
- Cleaner than storing these scattered across tables
- Supports complex coverage rules (e.g., MidC requires BOTH specialty)
- `specialtyConstraint` guides coverage resolution
- Immutable once created (part of versioned ScheduleVersion)

**Query Examples Enabled**:
```java
// Find all neuro overnight shifts in July
ShiftInstance.find("shiftType IN ('ON1', 'ON2') AND specialtyConstraint = ?1 AND scheduleDate >= ? AND scheduleDate <= ?",
    Specialty.NEURO, startDate, endDate)

// Find shifts at specific hospital
ShiftInstance.find("hospital = ?1 AND scheduleDate = ?2", Hospital.MAIN, date)
```

**Recommendation for v2**: Excellent as-is. No changes needed.

---

#### 3. ScrapeBatch (Traceability Header)

```java
public class ScrapeBatch extends PanacheEntity {
    public String state;  // PENDING, COMPLETE, FAILED
    public LocalDate windowStartDate;
    public LocalDate windowEndDate;
    public LocalDateTime scrapedAtTs;
    public LocalDateTime completedAtTs;
    public Integer rowCount;
    public String ingestChecksum;
    public LocalDateTime deletedAt;
    public LocalDateTime archivedAt;
    public String errorMessage;
}
```

**What Makes It Good**:
- **Atomic batch operations**: All data from one scrape is grouped
- **Traceability**: Can replay any batch, audit which batch created which schedule
- **Integrity validation**: Checksum detects corrupted imports
- **Soft delete**: Can "undo" a batch without losing data
- **Archival support**: Old batches can be moved to S3/archive
- **Lifecycle management**: Clear PENDING → COMPLETE/FAILED flow

**Prevent Scenarios**:
```
// Without batches:
Did assignment X come from the 2024-10-15 scrape or 2024-10-14?
→ Unknown, might reuse stale data

// With batches:
Assignment.batch_id points to ScrapeBatch with exact timestamp
→ Always know the source
```

**Recommendation for v2**: Excellent pattern. Enhance with:
- `sourceSystem` enum (AMION, MANUAL, IMPORT) for multi-source support
- `batchMetrics` (success rate, parse errors, time taken)
- Foreign key constraint from ScheduleVersion

---

#### 4. Assignment (Person → Shift Mapping)

```java
public class Assignment extends PanacheEntity {
    public Person person;
    public ShiftInstance shiftInstance;
    public LocalDate scheduleDate;
    public String originalShiftType;  // What Amion said
    public String source;  // AMION, MANUAL, OVERRIDE
    public LocalDateTime createdAt;
}
```

**What Makes It Good**:
- Minimal required fields (avoids data redundancy)
- `originalShiftType` preserved separately from calculated coverage
- `source` field enables audit trail (which system created this?)
- Composite key on (person, shiftInstance, date) prevents duplicates

**Why Not Store Effective Shift Type**:
✅ Original approach (v0): Store `reassignedShiftType` in database
❌ Problem: Stale when person specialty changes
✅ New approach (v1): Calculate at query time via `DynamicCoverageCalculator`
✅ Benefit: Specialty changes immediately affect coverage

**Recommendation for v2**: Keep current design. Remove the deprecated `reassignedShiftType` field.

---

## Validation Framework: Elegant & Extensible

### ValidationResult Pattern

```java
public class ValidationResult {
    private List<ValidationMessage> messages = new ArrayList<>();

    public void addError(String code, String message) {
        messages.add(new ValidationMessage(SEVERITY.ERROR, code, message));
    }

    public void addWarning(String code, String message) {
        messages.add(new ValidationMessage(SEVERITY.WARNING, code, message));
    }

    public void addInfo(String code, String message) {
        messages.add(new ValidationMessage(SEVERITY.INFO, code, message));
    }

    public boolean canImport() {
        return !messages.stream().anyMatch(m -> m.severity == SEVERITY.ERROR);
    }

    public boolean canPromote() {
        return !messages.stream().anyMatch(m ->
            m.severity == SEVERITY.ERROR || m.severity == SEVERITY.WARNING);
    }

    public static ValidationResult fromJson(String json) {
        return new ObjectMapper().readValue(json, ValidationResult.class);
    }
}
```

**What Makes It Good**:

1. **Severity Levels**: Distinguishes between must-fix (ERROR) and should-check (WARNING)
2. **Structured Codes**: `"UNKNOWN_PEOPLE"` instead of unstructured messages
3. **Serializability**: Stores as JSON in database for audit trail
4. **Multiple Use Cases**:
   - `canImport()`: Are errors present?
   - `canPromote()`: Are errors OR warnings present?
5. **User Friendly**: Clear messages explain what failed and why

**Real Examples from v1**:

```java
// ODS import validation
validation.addError("UNKNOWN_SHIFT_TYPE", "Unknown shift type: XRAY_NIGHT on 2024-10-15");
validation.addWarning("MISSING_MIDC", "No MidC assignment on weekday 2024-10-16");
validation.addWarning("INSUFFICIENT_COVERAGE", "Only 1 person for ON1 on 2024-10-20 (need 2)");

// Amion import validation
validation.addError("FAILED_TO_SCRAPE", "Amion scrape failed: timeout after 30s");
validation.addError("UNKNOWN_PEOPLE", "Unknown people in Amion: John D, Dr. Smith");
validation.addWarning("DUPLICATE_SHIFT", "Person JANE_DOE assigned to ON1 twice on 2024-10-15");

// Promotion validation
validation.canPromote()  // false if any warnings/errors
validation.canImport()   // false if any errors only
```

**Recommendation for v2**: Enhance with:
```java
public class ValidationMessage {
    public SEVERITY severity;
    public String code;
    public String message;
    public Map<String, Object> context;  // Additional data for debugging
    public String affectedDate;  // Which date caused the issue
    public String affectedEntity;  // Which person/shift

    public ValidationMessage(SEVERITY severity, String code, String message,
                           LocalDate affectedDate, String affectedEntity) {
        this.severity = severity;
        this.code = code;
        this.message = message;
        this.affectedDate = affectedDate.toString();
        this.affectedEntity = affectedEntity;
    }
}
```

This enables better error reporting to users.

---

## Dynamic Coverage Calculation: Elegant Solution

### The Problem It Solves

Before v1, coverage was calculated once and stored in database:
```java
// v0 approach - STALE DATA PROBLEM
if (hasBodyOnlySpecialist) {
    assignment.reassignedShiftType = "ONBody";  // Store once
    assignment.save();
}

// But what if person's specialty changes?
// → Stale data in database forever
```

### The Solution (v1)

```java
@ApplicationScoped
public class DynamicCoverageCalculator {
    public String getEffectiveShiftType(Assignment assignment, LocalDate date) {
        String original = assignment.originalShiftType;

        // Only ON1/ON2 subject to reassignment
        if (!original.equals("ON1") && !original.equals("ON2")) {
            return original;
        }

        // Check if there's a BODY_ONLY person tonight
        List<Assignment> overnightAssignments = Assignment.find(
            "originalShiftType IN ('ON1', 'ON2') and shiftInstance.scheduleDate = ?1",
            date
        ).list();

        boolean hasBodyOnly = overnightAssignments.stream()
            .anyMatch(a -> a.person.specialty == Person.Specialty.BODY_ONLY);

        if (!hasBodyOnly) {
            return original;  // No BODY_ONLY → return original
        }

        // BODY_ONLY exists → apply split
        if (assignment.person.specialty == Person.Specialty.BODY_ONLY) {
            return "ONBody";
        } else {
            return "ONNeuro";
        }
    }
}
```

**What Makes It Great**:

1. **Self-Healing**: When specialty changes, coverage automatically updates
2. **No Stale Data**: Calculated at query time from current state
3. **Minimal Storage**: Database only stores what Amion said
4. **Composable**: Can apply multiple rules in sequence

**Query Usage**:
```java
@ApplicationScoped
public class ScheduleService {
    public DailySchedule getScheduleForDate(LocalDate date) {
        List<Assignment> assignments = Assignment.find(
            "shiftInstance.scheduleDate = ?1", date
        ).list();

        DailySchedule schedule = new DailySchedule(date);
        for (Assignment a : assignments) {
            String effectiveType = calculator.getEffectiveShiftType(a, date);
            schedule.addShift(a.person.name, effectiveType);
        }
        return schedule;
    }
}
```

**Recommendation for v2**:
- Keep this pattern (with N+1 fix noted in Technical Debt #3)
- Consider rule engine if coverage rules get more complex
- Could add rule priority/precedence if needed

---

## Scrape Batch Workflow: Atomic & Traceable

### The Complete Flow

```java
@Transactional
public ScrapeBatch startScraping(YearMonth month) {
    ScrapeBatch batch = new ScrapeBatch();
    batch.state = "PENDING";
    batch.windowStartDate = month.atDay(1);
    batch.windowEndDate = month.atEndOfMonth();
    batch.scrapedAtTs = LocalDateTime.now();
    batch.persist();
    return batch;
}

@Transactional
public void scrapeAndIngest(ScrapeBatch batch) {
    try {
        // Scrape Amion
        String html = scraper.scrapeAmion(batch.windowStartDate);

        // Parse into rows
        List<ScrapedScheduleRow> rows = parser.parse(html);

        // Validate
        String checksum = calculateChecksum(rows);
        if (isDuplicate(checksum)) {
            batch.errorMessage = "Duplicate batch detected";
            batch.state = "FAILED";
            batch.persist();
            return;
        }

        // Ingest (all or nothing)
        for (ScrapedScheduleRow row : rows) {
            ScrapedSchedule schedule = new ScrapedSchedule();
            schedule.batch = batch;
            schedule.personAlias = row.personAlias;
            schedule.shiftType = row.shiftType;
            schedule.scheduleDate = row.scheduleDate;
            schedule.persist();
        }

        // Mark complete
        batch.rowCount = rows.size();
        batch.ingestChecksum = checksum;
        batch.state = "COMPLETE";
        batch.completedAtTs = LocalDateTime.now();
        batch.persist();

    } catch (Exception e) {
        batch.state = "FAILED";
        batch.errorMessage = e.getMessage();
        batch.persist();
        throw e;
    }
}
```

**What Makes It Great**:

1. **Atomic**: All rows from a batch are ingested together or not at all
2. **Traceable**: Every assignment links back to exact batch
3. **Idempotent**: Checksum prevents duplicate processing
4. **Auditable**: Timestamps, state transitions, error messages
5. **Safe Deletion**: Soft delete prevents orphaning schedules

**Prevents**:
```
❌ Without batches:
- 500 assignments scraped, 400 ingested, 100 lost → Silent data loss
- Can't replay a scrape
- Can't audit which scrape created which data

✅ With batches:
- All 500 ingested or none (atomic)
- Can replay by looking up batch ID
- Know exact timestamp and source
```

**Recommendation for v2**: Excellent pattern. No changes needed.

---

## REST API Design: Consistent & Well-Structured

### Response Wrapper Pattern

```java
public record ApiResponse<T>(
    T data,
    Meta meta,
    Error error
) {
    public record Meta(long timestamp, String version) {}
    public record Error(String code, String message, Object details) {}
}
```

**All endpoints follow this**:
```json
{
    "data": { ... },
    "meta": {
        "timestamp": 1697400000000,
        "version": "1.0"
    },
    "error": null
}
```

Error response:
```json
{
    "data": null,
    "meta": { ... },
    "error": {
        "code": "VALIDATION_FAILED",
        "message": "ODS import validation failed",
        "details": {
            "errors": [
                "Unknown shift type: XRAY_NIGHT",
                "Insufficient coverage on 2024-10-20"
            ]
        }
    }
}
```

**What Makes It Good**:
- Predictable structure (client doesn't need to handle 10 different formats)
- Timestamp for debugging time-based issues
- Version field for API evolution
- Metadata separate from data
- Error details included when present

**Recommendation for v2**: Enhance with:
```java
public record ApiError(
    String code,
    String message,
    List<ErrorDetail> details,
    String requestId,  // Correlation ID for logs
    String documentationUrl  // Link to docs
) {
    public record ErrorDetail(
        String field,
        String code,
        String message
    ) {}
}
```

---

## Security Design: JWT + RBAC

### Implementation

```java
@ApplicationScoped
public class JWTService {
    @ConfigProperty(name = "smallrye.jwt.sign.key.location")
    String signingKeyLocation;

    public String generateToken(User user, Duration expiry) {
        Instant now = Instant.now();
        return Jwt.claims()
            .issuer("hospital-radiology-schedule")
            .subject(user.email)
            .expiresAt(now.plus(expiry))
            .claim("oid", user.id)
            .claim("roles", user.roles)
            .sign(signingKey);
    }
}

// Endpoint protection
@RolesAllowed("ADMIN")
@PostMapping("/import/ods")
public Response importODS(...) { ... }

// Audit logging
@Transactional
public void promoteVersion(Long versionId, User user) {
    ScheduleVersion v = ScheduleVersion.findById(versionId);
    v.status = "PRODUCTION";
    v.persist();

    AuditLog log = new AuditLog();
    log.user = user;
    log.action = "PROMOTE_VERSION";
    log.oldValues = "status: STAGING";
    log.newValues = "status: PRODUCTION";
    log.timestamp = LocalDateTime.now();
    log.persist();
}
```

**What Makes It Good**:
1. **Stateless**: JWT doesn't require session storage
2. **Scalable**: Works with load balancers, multiple instances
3. **Decentralized**: No auth server needed (key-based verification)
4. **Audit Trail**: Every action logged with user
5. **Role-Based**: Simple `@RolesAllowed` annotation

**Recommendation for v2**: Current approach is solid. Just ensure:
- All admin endpoints have `@RolesAllowed("ADMIN")` (remove `@PermitAll`)
- Token expiry is reasonable (8 hours ✓)
- Keys are not hardcoded (use env vars)

---

## Documentation: Comprehensive & Clear

### What's There

- **12 markdown files** in `docs/` folder
  - `COMPLETE_SYSTEM_SPECIFICATION.md` - 500+ line detailed spec
  - `AMION_PARSING_DECISIONS.md` - Decision log for parsing rules
  - Architecture diagrams
  - Troubleshooting guides
  - Data flow documentation

### Why It's Valuable

```markdown
# Lesson: Without documentation
- New developer takes 2 weeks to understand architecture
- Decisions from 6 months ago are forgotten
- Same mistakes repeated in different places
- Why did we do this? Unknown

# Lesson: With documentation (v1 approach)
- New developer ramps up in 2-3 days
- Decisions are recorded (why we chose this approach)
- Troubleshooting guides help debug issues
- History is preserved
```

**Recommendation for v2**: Maintain and enhance:
- Keep decision logs for architectural choices
- Add performance analysis docs
- Document trade-offs explicitly
- Include runbook for common operations
- Add troubleshooting guide for common issues

---

## Testing Approach: Strong E2E Coverage

### What's Tested

```java
// Comprehensive E2E tests
ScheduleCalendarE2ETest          // UI calendar rendering
ConsolidatedCalendarE2ETest      // Multi-shift consolidated view
CalendarDataDistributionE2ETest  // Coverage distribution logic
HomepageLinksE2ETest             // Navigation

// Unit tests for critical logic
CoverageResolutionServiceTest    // Dynamic reassignment scenarios
ValidationResultTest             // Validation serialization
```

**What Makes It Good**:
- Tests actual user workflows (not just unit tests)
- Uses Selenium + HtmlUnit for browser automation
- Verifies UI rendering, not just API
- Tests data distribution (coverage is correct)

**Recommendation for v2**: Keep E2E tests AND add unit tests for:
- Import services (parsing logic)
- Validation rules
- Coverage calculator (batch version)

---

## Person Registry Sync: YAML-Based Reconciliation

### Implementation

```java
@ApplicationScoped
@Startup
public class PersonRegistryService {
    @Transactional
    public void reconcileRegistry() {
        // Load canonical person list from YAML
        List<Person> canonical = loadFromYaml("classpath:persons.yaml");

        // Remove inactive people from DB
        Person.find("active = true").list().forEach(p -> {
            if (!canonical.stream().anyMatch(c -> c.email.equals(p.email))) {
                p.active = false;
                p.persist();
            }
        });

        // Add/update active people
        canonical.forEach(person -> {
            Optional<Person> existing = Person.find("email", person.email)
                .firstResultOptional();

            if (existing.isEmpty()) {
                person.persist();
            } else {
                existing.get().specialty = person.specialty;
                existing.get().persist();
            }
        });
    }
}
```

**What Makes It Good**:
1. **Version Controlled**: Person list is in Git, can see history
2. **Reconciliable**: Startup sync ensures DB matches config
3. **Audit Trail**: Git log shows who changed what and when
4. **No Manual DB Edits**: All person changes go through code review
5. **Safe**: Soft-deletes people instead of removing

**Recommendation for v2**: Enhance with:
```yaml
# persons.yaml
- name: John Smith
  email: john.smith@hospital.org
  specialty: BOTH
  aliases:
    - "John S"
    - "J. Smith"
    - "Dr. Smith"
  phone: "+1-555-0123"
  activeFrom: "2024-01-01"
  activeTo: null  # null = still active
```

---

## What to Preserve in v2

| Pattern | File | Recommendation |
|---------|------|-----------------|
| Temporal versioning | `ScheduleVersion` | Keep as-is |
| Rich shift metadata | `ShiftInstance` | Keep as-is |
| Batch traceability | `ScrapeBatch` | Keep & enhance |
| Assignment minimalism | `Assignment` | Keep & clean up |
| Validation framework | `ValidationResult` | Keep & enhance |
| Dynamic coverage | `DynamicCoverageCalculator` | Keep & optimize (fix N+1) |
| Scrape workflow | `AmionImportService` | Keep & test |
| REST wrapper | `ApiResponse` | Keep & standardize |
| JWT security | `JWTService` | Keep & verify all endpoints |
| Documentation | `docs/` | Keep & expand |
| E2E testing | E2E tests | Keep & add units |
| YAML registry | `PersonRegistryService` | Keep & enhance |

**Summary**: The v1 architecture is fundamentally sound. v2 should focus on:
1. ✅ Keeping all these patterns
2. 🔧 Fixing the technical debt listed in 01-TECHNICAL-DEBT.md
3. ✨ Adding the enhancements suggested above

The hard part (domain modeling, architecture) is done. The work is in polish and testing.
