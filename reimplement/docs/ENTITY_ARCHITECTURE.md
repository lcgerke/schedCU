# Entity Architecture Documentation

**Phase:** Phase 1, Work Package [0.4]
**Last Updated:** November 15, 2025
**Location:** `/schedCU/v2/internal/entity/`

## Overview

The schedCU entity architecture defines the core domain objects used throughout the system. The design emphasizes:

- **Immutability**: Entities are immutable after creation; changes tracked via versioning
- **Soft Deletion**: Data preservation through `DeletedAt` timestamps rather than hard deletes
- **Audit Trails**: All mutable operations tracked via `CreatedBy`, `UpdatedBy`, `DeletedBy` fields
- **Type Safety**: UUID aliases for domain identifiers prevent accidental ID confusion
- **State Machines**: Explicit status transitions for Schedule Versions and Batch operations

---

## Entity Relationship Diagram

```
                          Hospital (Root)
                              |
                ______________|______________
               |              |              |
               v              v              v
           Person        ScheduleVersion    ScrapeBatch
               |              |              |
               |              |________      |
               |                     |      |
               |___________          |      |
                           |         |      |
                       Assignment    |      |
                           |    _____|      |
                           |   |           |
                      ShiftInstance        |
                           |              |
                      (Immutable)         |
                                          |
                       CoverageCalculation |
                                          |
                    AuditLog (Append-only)
                    User (Admin, Scheduler, Viewer)
                    JobQueue (Async operations)
```

### Relationship Summary

| From | To | Type | Notes |
|------|-----|------|-------|
| Person | Hospital | Many-to-One | Staff member belongs to hospital |
| Assignment | Person | Many-to-One | Links person to shifts |
| Assignment | ShiftInstance | Many-to-One | Each assignment claims a shift |
| ShiftInstance | ScheduleVersion | Many-to-One | Shifts belong to schedule version |
| ScheduleVersion | Hospital | Many-to-One | Schedule is hospital-specific |
| ScheduleVersion | ScrapeBatch | One-to-Optional | Links to data source batch |
| CoverageCalculation | ScheduleVersion | Many-to-One | Calculates coverage for version |
| ScrapeBatch | Hospital | Many-to-One | Batch is hospital-specific |
| AuditLog | User | Many-to-One | User performs audit action |

---

## Core Entities

### 1. Person (Staff Member)

**Purpose:** Represents a radiologist staff member with specialty constraints.

**Key Fields:**
```go
type Person struct {
	ID        uuid.UUID          // Unique identifier
	Email     string             // Primary identifier for matching
	Name      string             // Display name
	Specialty SpecialtyType      // BODY_ONLY | NEURO_ONLY | BOTH
	Active    bool               // Soft activity flag
	Aliases   []string           // Alternative names for matching Amion data
	CreatedAt time.Time          // Immutable creation timestamp
	UpdatedAt time.Time          // Last modification (initial = CreatedAt)
	DeletedAt *time.Time         // Soft delete marker (NULL = active)
}
```

**Immutability Guarantee:**
- `CreatedAt` is set once and never changes
- `DeletedAt` is set only on deletion (soft delete pattern)
- No `DeletedBy` field (deletion tracked via timestamp only)

**State Transitions:**
```
[CREATE] → Active (DeletedAt is NULL)
  ↓
[SOFT DELETE] → Inactive (DeletedAt is set to current time)
  ↓
[QUERY] → WHERE deleted_at IS NULL (only active persons)
```

**Querying Pattern (SQL):**
```sql
-- Get all active persons for a hospital
SELECT * FROM persons
WHERE hospital_id = $1 AND deleted_at IS NULL;

-- Get all persons (including deleted)
SELECT * FROM persons
WHERE hospital_id = $1;
```

---

### 2. ShiftInstance (Required Shift)

**Purpose:** Represents a required shift within a schedule version. Immutable once created.

**Key Fields:**
```go
type ShiftInstance struct {
	ID                  uuid.UUID     // Unique identifier
	ScheduleVersionID   uuid.UUID     // Parent schedule version
	ShiftType           ShiftType     // ON1 | ON2 | MidC | MidL | DAY
	ScheduleDate        time.Time     // Date of the shift
	StartTime           string        // HH:MM format
	EndTime             string        // HH:MM format
	HospitalID          uuid.UUID     // Hospital context
	StudyType           StudyType     // GENERAL | BODY | NEURO
	SpecialtyConstraint SpecialtyType // Guides coverage resolution
	DesiredCoverage     int           // How many people needed
	IsMandatory         bool          // Must be covered
	CreatedAt           time.Time     // Immutable creation timestamp
	CreatedBy           uuid.UUID     // User who created shift
}
```

**Immutability Guarantee:**
- No `UpdatedAt`, `UpdatedBy`, `DeletedAt` fields
- Once created, shift instances CANNOT be modified or deleted
- Changes require creating a new `ScheduleVersion`
- This ensures audit trail integrity for schedule versioning

**Why Immutable?**
Schedule versions represent snapshots. Allowing shift modifications would:
1. Break audit trails
2. Make version history unreliable
3. Complicate conflict resolution

**Related Methods:**
```go
// Retrieve shifts for active assignments only
ShiftInstancesByScheduleVersion(scheduleVersionID uuid.UUID) []ShiftInstance
```

---

### 3. Assignment (Person-to-Shift Mapping)

**Purpose:** Maps a person to a shift instance, tracking the data source.

**Key Fields:**
```go
type Assignment struct {
	ID                uuid.UUID          // Unique identifier
	PersonID          uuid.UUID          // Foreign key to Person
	ShiftInstanceID   uuid.UUID          // Foreign key to ShiftInstance
	ScheduleDate      time.Time          // Date of assignment
	OriginalShiftType string             // Original name from Amion
	Source            AssignmentSource   // AMION | MANUAL | OVERRIDE
	CreatedAt         time.Time          // Creation timestamp
	CreatedBy         uuid.UUID          // User who created assignment
	DeletedAt         *time.Time         // Soft delete marker
	DeletedBy         *uuid.UUID         // User who deleted assignment
}

type AssignmentSource string
const (
	AssignmentSourceAmion    = "AMION"
	AssignmentSourceManual   = "MANUAL"
	AssignmentSourceOverride = "OVERRIDE"
)
```

**Soft Delete Pattern:**
```go
// Mark as deleted without removing data
func (a *Assignment) SoftDelete(deleterID uuid.UUID) {
	now := time.Now().UTC()
	a.DeletedAt = &now
	a.DeletedBy = &deleterID
}

// Check if deleted
func (a *Assignment) IsDeleted() bool {
	return a.DeletedAt != nil
}
```

**Querying Pattern:**
```sql
-- Get active assignments for a schedule
SELECT * FROM assignments
WHERE shift_instance_id IN (
  SELECT id FROM shift_instances
  WHERE schedule_version_id = $1
) AND deleted_at IS NULL;

-- Get assignment history (including deletions)
SELECT * FROM assignments
WHERE person_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;
```

**Audit Capability:**
- `OriginalShiftType` preserves Amion naming
- `Source` tracks whether assignment came from Amion, manual entry, or override
- `CreatedBy` / `DeletedBy` enable user accountability

---

### 4. ScheduleVersion (Temporal Schedule)

**Purpose:** Represents a schedule snapshot with version control and promotion workflow.

**Key Fields:**
```go
type ScheduleVersion struct {
	ID                 uuid.UUID          // Unique identifier
	HospitalID         uuid.UUID          // Hospital context
	Status             VersionStatus      // STAGING | PRODUCTION | ARCHIVED
	EffectiveStartDate time.Time          // Schedule validity period
	EffectiveEndDate   time.Time
	ScrapeBatchID      *uuid.UUID         // Soft-link to source batch
	ValidationResults  *ValidationResult  // Validation status
	ShiftInstances     []ShiftInstance    // Immutable shift list
	CreatedAt          time.Time          // Creation timestamp
	CreatedBy          uuid.UUID          // Creator user ID
	UpdatedAt          time.Time          // Last status change
	UpdatedBy          uuid.UUID          // User who changed status
	DeletedAt          *time.Time         // Soft delete marker
	DeletedBy          *uuid.UUID         // User who deleted
}

type VersionStatus string
const (
	VersionStatusStaging     = "STAGING"
	VersionStatusProduction  = "PRODUCTION"
	VersionStatusArchived    = "ARCHIVED"
)
```

**State Machine:**
```
┌─────────────────────────────────────────┐
│    ScheduleVersion Lifecycle            │
└─────────────────────────────────────────┘

[CREATE]
  ↓ Status = STAGING
[STAGING] ← Validation, testing phase
  ↓
[PROMOTE] → Status = PRODUCTION (only one active at a time)
  ↓
[PRODUCTION] ← Live schedule being used
  ↓
[ARCHIVE] → Status = ARCHIVED (replaced by newer version)
  ↓
[ARCHIVED] ← Historical record

Alternative: [STAGING] → [SOFT DELETE] (rejected schedules)
```

**State Transition Methods:**
```go
// Promote from STAGING to PRODUCTION
func (sv *ScheduleVersion) Promote(promoterID uuid.UUID) error {
	if sv.Status != VersionStatusStaging {
		return ErrInvalidVersionStateTransition
	}
	sv.Status = VersionStatusProduction
	sv.UpdatedAt = time.Now().UTC()
	sv.UpdatedBy = promoterID
	return nil
}

// Archive from PRODUCTION to ARCHIVED
func (sv *ScheduleVersion) Archive(archiverID uuid.UUID) error {
	if sv.Status != VersionStatusProduction {
		return ErrCannotArchiveNonProduction
	}
	sv.Status = VersionStatusArchived
	sv.UpdatedAt = time.Now().UTC()
	sv.UpdatedBy = archiverID
	return nil
}
```

**Audit Trail:**
- `CreatedAt`, `CreatedBy`: Who created this version
- `UpdatedAt`, `UpdatedBy`: Last status change (promotion/archival)
- `DeletedAt`, `DeletedBy`: Soft deletion tracking

**Important Design Notes:**
- `ScrapeBatchID` is a soft-link (no foreign key constraint) allowing batch deletion without cascade
- `ShiftInstances` are immutable by design
- Only ONE `ScheduleVersion` should be in PRODUCTION at any time
- Soft delete is separate from versioning (can delete entire versions)

---

### 5. ScrapeBatch (Atomic Import Operation)

**Purpose:** Groups data from a single scrape/import operation with full traceability.

**Key Fields:**
```go
type ScrapeBatch struct {
	ID              uuid.UUID    // Unique identifier
	HospitalID      uuid.UUID    // Hospital context
	State           BatchState   // PENDING | COMPLETE | FAILED
	WindowStartDate time.Time    // Date range of scrape
	WindowEndDate   time.Time
	ScrapedAt       time.Time    // When scrape occurred
	CompletedAt     *time.Time   // When processing completed
	RowCount        int          // Rows imported
	IngestChecksum  string       // Data integrity verification
	ErrorMessage    *string      // Error details if FAILED
	CreatedAt       time.Time    // Batch creation time
	CreatedBy       uuid.UUID    // User who created batch
	DeletedAt       *time.Time   // Soft delete marker
	DeletedBy       *uuid.UUID   // User who deleted batch
	ArchivedAt      *time.Time   // Archival timestamp
	ArchivedBy      *uuid.UUID   // User who archived
}

type BatchState string
const (
	BatchStatePending  = "PENDING"
	BatchStateComplete = "COMPLETE"
	BatchStateFailed   = "FAILED"
)
```

**Batch Lifecycle:**
```
[CREATE] State = PENDING
  ↓ (Processing)
[MarkComplete] → State = COMPLETE, CompletedAt set
  ↓
[Optional] MarkArchived → ArchivedAt set (for cleanup)

OR

[CREATE] State = PENDING
  ↓ (Error during processing)
[MarkFailed] → State = FAILED, ErrorMessage set, CompletedAt set
  ↓
[CLEANUP] → Soft delete if needed
```

**State Transition Methods:**
```go
// Mark batch as successfully completed
func (b *ScrapeBatch) MarkComplete(completerID uuid.UUID, rowCount int) {
	now := time.Now().UTC()
	b.State = BatchStateComplete
	b.CompletedAt = &now
	b.RowCount = rowCount
}

// Mark batch as failed
func (b *ScrapeBatch) MarkFailed(errorMsg string) {
	now := time.Now().UTC()
	b.State = BatchStateFailed
	b.CompletedAt = &now
	b.ErrorMessage = &errorMsg
}

// Archive for cleanup
func (b *ScrapeBatch) MarkArchived(archiverID uuid.UUID) {
	now := time.Now().UTC()
	b.ArchivedAt = &now
	b.ArchivedBy = &archiverID
}
```

**Data Integrity:**
- `IngestChecksum`: Detects corrupted imports; allows replay detection
- `RowCount`: Verification of expected vs. actual rows
- `ErrorMessage`: Clear error reporting for failed batches

**Querying Patterns:**
```sql
-- Get recent successful imports
SELECT * FROM scrape_batches
WHERE hospital_id = $1
  AND state = 'COMPLETE'
  AND deleted_at IS NULL
ORDER BY scraped_at DESC LIMIT 10;

-- Get all batches including errors
SELECT * FROM scrape_batches
WHERE hospital_id = $1
ORDER BY created_at DESC;
```

---

### 6. CoverageCalculation (Coverage Analysis)

**Purpose:** Stores pre-calculated coverage metrics for a schedule version.

**Key Fields:**
```go
type CoverageCalculation struct {
	ID                         uuid.UUID              // Unique identifier
	ScheduleVersionID          uuid.UUID              // Parent schedule version
	HospitalID                 uuid.UUID              // Hospital context
	CalculationDate            time.Time              // When calculation was performed
	CalculationPeriodStartDate time.Time              // Period being analyzed
	CalculationPeriodEndDate   time.Time
	CoverageByPosition         map[string]int         // Position → count (JSONB)
	CoverageSummary            map[string]interface{} // Summary metrics (JSONB)
	ValidationErrors           *ValidationResult      // Issues found
	QueryCount                 int                    // Performance metric
	CalculatedAt               time.Time              // Timestamp
	CalculatedBy               uuid.UUID              // User who calculated
}
```

**JSONB Column Usage:**
```json
// Example CoverageByPosition
{
  "ON1": 2,
  "ON2": 2,
  "MidC": 1,
  "MidL": 1,
  "DAY": 3
}

// Example CoverageSummary
{
  "total_shifts": 15,
  "covered_shifts": 13,
  "uncovered_shifts": 2,
  "coverage_percentage": 86.67,
  "body_imaging_count": 8,
  "neuro_imaging_count": 7
}
```

**Design Notes:**
- JSONB enables flexible metrics without schema changes
- Pre-calculated to avoid expensive recalculations
- `ValidationErrors` tracks issues (under-coverage, constraint violations)
- `QueryCount` used for performance profiling

---

### 7. AuditLog (Compliance & Debugging)

**Purpose:** Immutable append-only log of all administrative actions.

**Key Fields:**
```go
type AuditLog struct {
	ID        uuid.UUID // Unique identifier
	UserID    uuid.UUID // User performing action
	Action    string    // PROMOTE_VERSION, IMPORT_ODS, DELETE_ASSIGNMENT, etc.
	Resource  string    // e.g., "ScheduleVersion#123", "Assignment#456"
	OldValues string    // JSON snapshot
	NewValues string    // JSON snapshot
	Timestamp time.Time // UTC timestamp
	IPAddress string    // Source IP for compliance
}
```

**Append-Only Pattern:**
- No `DeletedAt` field (cannot be deleted or modified)
- Provides immutable record of all changes
- Enables debugging of cascading failures
- Satisfies compliance/regulatory requirements

**Audit Trail Examples:**
```json
{
  "action": "PROMOTE_VERSION",
  "resource": "ScheduleVersion#550e8400-e29b-41d4-a716-446655440000",
  "old_values": {"status": "STAGING", "updated_at": "2025-11-15T10:00:00Z"},
  "new_values": {"status": "PRODUCTION", "updated_at": "2025-11-15T10:15:00Z"}
}

{
  "action": "DELETE_ASSIGNMENT",
  "resource": "Assignment#f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "old_values": {"deleted_at": null},
  "new_values": {"deleted_at": "2025-11-15T11:00:00Z"}
}
```

---

### 8. Additional Entities

#### User (Authentication & Authorization)
```go
type User struct {
	ID          uuid.UUID    // Unique identifier
	Email       string       // Login identifier
	Name        string
	PasswordHash string
	Role        UserRole     // ADMIN | SCHEDULER | VIEWER
	HospitalID  *uuid.UUID   // NULL for system admin
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt *time.Time
	DeletedAt   *time.Time   // Soft delete
}

type UserRole string
const (
	UserRoleAdmin     = "ADMIN"
	UserRoleScheduler = "SCHEDULER"
	UserRoleViewer    = "VIEWER"
)
```

#### JobQueue (Async Operations)
```go
type JobQueue struct {
	ID          uuid.UUID
	JobType     string                 // ODS_IMPORT, AMION_IMPORT, COVERAGE_CALCULATION
	Payload     map[string]interface{} // Job-specific data
	Status      JobQueueStatus         // PENDING | PROCESSING | COMPLETE | FAILED | RETRY
	Result      map[string]interface{} // Results after completion
	ErrorMessage *string
	RetryCount  int
	MaxRetries  int
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

type JobQueueStatus string
const (
	JobQueueStatusPending    = "PENDING"
	JobQueueStatusProcessing = "PROCESSING"
	JobQueueStatusComplete   = "COMPLETE"
	JobQueueStatusFailed     = "FAILED"
	JobQueueStatusRetry      = "RETRY"
)
```

#### Schedule (Simplified View)
```go
type Schedule struct {
	ID          uuid.UUID
	HospitalID  uuid.UUID
	StartDate   time.Time
	EndDate     time.Time
	Source      string         // 'amion', 'ods_file', 'manual'
	SourceID    *string
	Assignments []interface{}
	CreatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedAt   time.Time
	UpdatedBy   uuid.UUID
	DeletedAt   *time.Time
	DeletedBy   *uuid.UUID
}
```

#### ValidationResult (Status & Errors)
```go
type ValidationResult struct {
	Valid    bool                   // Is valid?
	Code     string                 // VALIDATION_SUCCESS, PARSE_ERROR, etc.
	Severity string                 // INFO, WARNING, ERROR
	Message  string                 // Human-readable message
	Context  map[string]interface{} // Additional details
}
```

---

## Soft Delete Pattern

The system uses **soft deletion** throughout to preserve audit trails and enable recovery.

### Implementation

```go
// IsDeleted checks if entity is soft-deleted
func (p *Person) IsDeleted() bool {
	return p.DeletedAt != nil
}

// SoftDelete marks entity as deleted with audit trail
func (p *Person) SoftDelete(deleterID uuid.UUID) {
	now := time.Now().UTC()
	p.DeletedAt = &now
	// Note: Person doesn't track DeletedBy, but others do
}

func (a *Assignment) SoftDelete(deleterID uuid.UUID) {
	now := time.Now().UTC()
	a.DeletedAt = &now
	a.DeletedBy = &deleterID  // Full audit trail
}
```

### Query Patterns

**Active Records Only:**
```sql
SELECT * FROM assignments
WHERE deleted_at IS NULL;
```

**All Records (Including Deleted):**
```sql
SELECT * FROM assignments;
```

**Recovery/Undelete:**
```go
// To "undelete", set DeletedAt to NULL and track via audit log
UPDATE assignments
SET deleted_at = NULL, updated_by = $1, updated_at = NOW()
WHERE id = $2;
```

### Benefits

1. **Compliance**: Historical records preserved for audits
2. **Recovery**: Accidental deletions reversible
3. **Debugging**: Understand cascading effects of deletions
4. **Analytics**: Full dataset available for analysis

---

## Type Aliases (Domain IDs)

All entity IDs use UUID type aliases for type safety:

```go
type (
	HospitalID        = uuid.UUID
	PersonID          = uuid.UUID
	ScheduleVersionID = uuid.UUID
	ShiftInstanceID   = uuid.UUID
	AssignmentID      = uuid.UUID
	ScrapeBatchID     = uuid.UUID
	CoverageID        = uuid.UUID
	AuditLogID        = uuid.UUID
	UserID            = uuid.UUID
	JobQueueID        = uuid.UUID
)

// Prevents this bug:
// person := GetPerson(shiftID)  // Type error! Caught at compile time
// Instead of being caught at runtime
```

**Benefits:**
- Prevents accidental ID mixing (PersonID vs. ShiftInstanceID)
- Self-documenting code
- Compile-time safety

---

## Immutability Patterns

The system uses **selective immutability** to balance auditability with functionality.

### Completely Immutable
- **ShiftInstance**: No updates after creation (versioning via ScheduleVersion)
- **AuditLog**: Append-only, no deletes or updates

### Create-Once + Soft Delete
- **Person**: Created once, soft-deleted but not modified
- **Assignment**: Created once, can be soft-deleted
- **ScrapeBatch**: Created once, state machine (PENDING → COMPLETE/FAILED)

### Mutable with Audit Trail
- **ScheduleVersion**: Status transitions (STAGING → PRODUCTION → ARCHIVED), tracked via UpdatedAt/UpdatedBy
- **User**: Can be updated (role changes, password updates), tracked via UpdatedAt

### Why This Matters

1. **Auditability**: Immutable creation fields ensure origin tracking
2. **Consistency**: Users never surprised by untracked changes
3. **Distributed Systems**: Reduces race conditions with versioning
4. **Data Recovery**: Complete history for root cause analysis

---

## JSONB Column Usage

The system uses PostgreSQL JSONB columns for flexible data without schema changes.

### ScheduleVersion.ValidationResults
```go
ValidationResults *ValidationResult

// Stored in DB as JSONB:
{
  "valid": true,
  "code": "VALIDATION_SUCCESS",
  "severity": "INFO",
  "message": "Validation passed",
  "context": {
    "shifts_validated": 15,
    "assignments_validated": 25,
    "warnings": 2
  }
}
```

### CoverageCalculation Metrics
```go
CoverageByPosition map[string]int
CoverageSummary map[string]interface{}

// Stored as JSONB for reporting/analytics:
{
  "coverage_by_position": {
    "ON1": 2, "ON2": 2, "MidC": 1, ...
  },
  "coverage_summary": {
    "total_shifts": 15,
    "covered_shifts": 13,
    "coverage_percentage": 86.67,
    "specialty_constraints": {...},
    "under_covered_shifts": [...]
  }
}
```

### ScrapeBatch.ErrorMessage
```go
ErrorMessage *string // Single string or JSON if complex

// Can store:
"Connection timeout after 30s"
// or
{
  "error_type": "validation_error",
  "field": "assignment_date",
  "value": "2025-99-99",
  "reason": "Invalid date format"
}
```

**Benefits:**
- **Flexibility**: Add new metrics without migrations
- **Queryability**: PostgreSQL JSONB operators enable complex queries
- **Backward Compatibility**: Old records coexist with new schema

---

## Foreign Key Constraints & Soft Links

### Hard Foreign Keys (Enforced)
```
Assignment.PersonID → Person.ID
Assignment.ShiftInstanceID → ShiftInstance.ID
ShiftInstance.ScheduleVersionID → ScheduleVersion.ID
ScheduleVersion.HospitalID → Hospital.ID
```

These prevent orphaned records and maintain referential integrity.

### Soft Links (No Constraint)
```
ScheduleVersion.ScrapeBatchID → ScrapeBatch.ID (optional)
CoverageCalculation.ScheduleVersionID → ScheduleVersion.ID (optional)
```

Soft links allow:
- Deleting batches without cascading to versions
- Versions existing without source batch reference
- More flexibility in data lifecycle

---

## Error Handling

Domain-specific errors for validation and state transitions:

```go
var (
	ErrInvalidVersionStateTransition = errors.New(
		"invalid version state transition")
	ErrCannotArchiveNonProduction = errors.New(
		"cannot archive non-production version")
	ErrInvalidDateRange = errors.New(
		"invalid date range: end date must be after start date")
	ErrUnknownShiftType = errors.New(
		"unknown shift type")
	ErrUnknownSpecialty = errors.New(
		"unknown specialty type")
)
```

Validation functions for enums:

```go
// Validate status strings before state transitions
ValidateVersionStatus(status string) bool
ValidateBatchState(state string) bool
ValidateSpecialty(specialty string) bool
ValidateShiftType(shiftType string) bool

// Validate business rules
ValidateDateRange(startDate, endDate time.Time) error
```

---

## Entity Lifecycle Examples

### Schedule Version Workflow

```
1. CREATE ScheduleVersion(status=STAGING)
   - Import from ScrapeBatch
   - Create child ShiftInstances (immutable)
   - Run validation

2. VALIDATE
   - Check coverage adequacy
   - Verify specialty constraints
   - Store ValidationResults

3. TEST (in STAGING)
   - Review assignments
   - Check for conflicts
   - Optional: Run additional analysis

4. PROMOTE to PRODUCTION
   - Verify no other PRODUCTION versions active
   - SetUpdatedAt, SetUpdatedBy
   - Track via AuditLog

5. ARCHIVE old PRODUCTION
   - When new version replaces it
   - Transition to ARCHIVED
   - Track via AuditLog

6. SOFT DELETE (optional)
   - For rejected STAGING versions
   - SetDeletedAt, SetDeletedBy
   - Full recovery possible
```

### Assignment Workflow

```
1. CREATE Assignment (source=AMION or MANUAL)
   - PersonID → Person
   - ShiftInstanceID → ShiftInstance
   - Track source (Amion vs Manual vs Override)
   - AuditLog entry created

2. ACTIVE
   - Included in coverage calculations
   - Visible in schedule reports
   - DeletedAt is NULL

3. OVERRIDE (optional)
   - Create new Assignment with source=OVERRIDE
   - Soft delete original
   - Track who made change

4. SOFT DELETE
   - SetDeletedAt to current time
   - SetDeletedBy to user ID
   - No longer included in active queries
   - AuditLog entry created
   - Recovery possible

5. RECOVERY (optional)
   - Clear DeletedAt
   - Track via AuditLog
```

### Batch Operation Workflow

```
1. CREATE ScrapeBatch (state=PENDING)
   - Timestamp: ScrapedAt
   - Track: CreatedBy, CreatedAt

2. PROCESS
   - Import data from source
   - Calculate IngestChecksum
   - Count RowCount

3. SUCCESS → MarkComplete
   - State = COMPLETE
   - CompletedAt = now
   - RowCount = imported count

4. FAILURE → MarkFailed
   - State = FAILED
   - ErrorMessage = error details
   - CompletedAt = now

5. ARCHIVE (optional)
   - MarkArchived by admin
   - ArchivedAt set

6. CLEANUP (optional)
   - Soft delete if needed
   - DeletedAt set
```

---

## Entity File Structure

**Location:** `/home/lcgerke/schedCU/v2/internal/entity/`

| File | Purpose |
|------|---------|
| `entities.go` | Core entity definitions, type aliases, state machine methods |
| `schedule.go` | Schedule & ValidationResult, builder functions |
| `errors.go` | Domain-specific errors, validation functions |
| `entities_test.go` | Entity behavior tests |
| `schedule_test.go` | Schedule-specific tests |

---

## Key Design Decisions

### 1. Immutable ShiftInstances
**Decision:** ShiftInstances cannot be modified or deleted after creation.

**Rationale:**
- Ensures schedule version integrity
- Prevents accidental changes affecting assignments
- Enables consistent time-travel queries
- If changes needed: create new ScheduleVersion

### 2. Soft Deletion Everywhere
**Decision:** No hard deletes; use DeletedAt timestamps instead.

**Rationale:**
- Preserves audit trail and recovery capability
- Complies with data retention policies
- Enables debugging cascading failures
- Better for regulatory/compliance requirements

### 3. No UpdatedBy on Person
**Decision:** Person entity tracks CreatedAt but not UpdatedAt/UpdatedBy.

**Rationale:**
- Persons are primarily append-only (soft delete only)
- Immutability reduces tracking complexity
- Source of truth for person data is external (email system)

### 4. UpdatedBy on ScheduleVersion
**Decision:** Track status changes via UpdatedAt/UpdatedBy.

**Rationale:**
- State transitions (promotion, archival) are significant
- Need to know who made promotion decision
- Enables accountability for schedule changes
- Different from Person (which tracks deletion only)

### 5. Type Aliases for IDs
**Decision:** Use UUID type aliases (PersonID, ShiftInstanceID, etc.).

**Rationale:**
- Compile-time type safety
- Prevents ID confusion
- Self-documenting code
- No runtime overhead

### 6. JSONB for Flexible Metrics
**Decision:** Use JSONB columns for ValidationResults and Coverage metrics.

**Rationale:**
- New metrics don't require migrations
- Flexible structure for different validation rules
- PostgreSQL JSONB enables querying without changes
- Backward compatible with evolving requirements

---

## Recommended Improvements for Phase 2

### High Priority

1. **Add createdBy/updatedBy to Hospital**
   - Currently missing audit trail
   - Need to know who created each hospital
   - Recommendation: Add CreatedAt, CreatedBy, UpdatedAt, UpdatedBy

2. **Add Hospital Context to User**
   - UserRole.SCHEDULER needs hospital scope
   - Add HospitalID (nullable for ADMIN)
   - Enables multi-hospital access control

3. **Add UpdatedAt/UpdatedBy to CoverageCalculation**
   - Currently only CalculatedAt/CalculatedBy
   - Needed if calculations can be regenerated
   - Recommendation: Add UpdatedAt, UpdatedBy for consistency

4. **Add Status to CoverageCalculation**
   - Track calculation state (PENDING, COMPLETE, FAILED)
   - Similar to BatchState pattern
   - Enables error handling and retries

### Medium Priority

5. **Refine ValidationResult Storage**
   - Currently *ValidationResult in ScheduleVersion
   - Consider: Separate ValidationLog entity
   - Enables multiple validation runs per version

6. **Add Tags/Labels to ScheduleVersion**
   - For scheduling (e.g., "Q4 Schedule", "Holiday Coverage")
   - Metadata for reporting
   - JSONB tags field would work

7. **Add Constraint History**
   - Track when specialties changed
   - Separate entity: PersonSpecialtyHistory
   - Enables audit of coverage changes

### Lower Priority

8. **Add Notification Preferences to User**
   - JSONB notification_channels field
   - Track email, SMS, in-app preferences
   - Enables multi-channel notifications

9. **Add Role-Based Permissions**
   - Separate Permissions entity
   - Link via UserRole
   - Enables granular access control

10. **Add Activity Tracking**
    - Last accessed schedule version
    - Last viewed assignment
    - User engagement metrics

---

## Summary

The schedCU entity architecture provides:

- **Immutability**: ShiftInstances never change; enables reliable versioning
- **Auditability**: Complete CreatedBy/UpdatedBy/DeletedBy trails
- **Soft Deletion**: Data preservation and recovery capability
- **Type Safety**: UUID aliases prevent ID confusion
- **Flexibility**: JSONB for evolving metrics
- **State Machines**: Explicit ScheduleVersion and ScrapeBatch workflows

The design balances:
- **Compliance** (immutable audit logs, no hard deletes)
- **Usability** (clear state transitions, easy soft delete/restore)
- **Performance** (indexed soft-delete queries, JSONB flexibility)
- **Maintainability** (domain-specific errors, validation functions)

This foundation supports the Phase 1 requirements and provides a solid base for Phase 2 enhancements.
