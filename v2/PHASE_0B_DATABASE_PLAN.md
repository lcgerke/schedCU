# Phase 0b: Database Layer Implementation Plan

**Status**: IN PROGRESS - Migrations framework created, schema design complete
**Date**: November 15, 2025
**Goal**: Complete database schema, repositories, and basic services to unblock Phase 1

---

## What's Complete

✅ Entity model (12+ entities with full tests)
✅ Validation framework (14 tests, 94% coverage)
✅ In-memory repository (10 tests)
✅ Migration file structure
✅ Schema design for all tables

### Migrations Created So Far

```
migrations/
├── 001_create_hospitals.up.sql        ✅ Created
├── 001_create_hospitals.down.sql      ✅ Created
├── 002_create_persons.up.sql          ✅ Created
├── 002_create_persons.down.sql        ✅ Created
├── 003_create_scrape_batches.up.sql   ✅ Created
├── 003_create_scrape_batches.down.sql ✅ Created
├── 004_create_schedule_versions.up.sql ✅ Created
└── 004_create_schedule_versions.down.sql ✅ Created
```

---

## Complete Schema Specification (Ready to Implement)

### Migration Files Needed (8 total pairs)

#### 005: Shift Instances Table
```sql
CREATE TABLE shift_instances (
    id UUID PRIMARY KEY,
    schedule_version_id UUID NOT NULL REFERENCES schedule_versions(id),
    shift_type VARCHAR(50) NOT NULL
        CHECK (shift_type IN ('ON1', 'ON2', 'MidC', 'MidL', 'DAY')),
    schedule_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    hospital_id UUID NOT NULL REFERENCES hospitals(id),
    study_type VARCHAR(50) NOT NULL
        CHECK (study_type IN ('GENERAL', 'BODY', 'NEURO')),
    specialty_constraint VARCHAR(50) NOT NULL
        CHECK (specialty_constraint IN ('BODY_ONLY', 'NEURO_ONLY', 'BOTH')),
    desired_coverage INTEGER NOT NULL DEFAULT 1,
    is_mandatory BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL
);

CREATE INDEX idx_shift_instances_schedule_version ON shift_instances(schedule_version_id);
CREATE INDEX idx_shift_instances_date ON shift_instances(schedule_date);
CREATE INDEX idx_shift_instances_hospital_date ON shift_instances(hospital_id, schedule_date);
CREATE INDEX idx_shift_instances_shift_type ON shift_instances(shift_type);
```

#### 006: Assignments Table
```sql
CREATE TABLE assignments (
    id UUID PRIMARY KEY,
    person_id UUID NOT NULL REFERENCES persons(id),
    shift_instance_id UUID NOT NULL REFERENCES shift_instances(id),
    schedule_date DATE NOT NULL,
    original_shift_type VARCHAR(255),
    source VARCHAR(50) NOT NULL DEFAULT 'MANUAL'
        CHECK (source IN ('AMION', 'MANUAL', 'OVERRIDE')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID
);

CREATE UNIQUE INDEX idx_assignments_unique ON assignments(person_id, shift_instance_id, schedule_date)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_assignments_person ON assignments(person_id);
CREATE INDEX idx_assignments_shift ON assignments(shift_instance_id);
CREATE INDEX idx_assignments_date ON assignments(schedule_date);
CREATE INDEX idx_assignments_source ON assignments(source);
```

#### 007: Coverage Calculations Table
```sql
CREATE TABLE coverage_calculations (
    id UUID PRIMARY KEY,
    schedule_version_id UUID NOT NULL REFERENCES schedule_versions(id),
    hospital_id UUID NOT NULL REFERENCES hospitals(id),
    calculation_date DATE NOT NULL,
    calculation_period_start_date DATE NOT NULL,
    calculation_period_end_date DATE NOT NULL,
    coverage_by_position JSONB NOT NULL,
    coverage_summary JSONB,
    validation_errors JSONB,
    query_count INTEGER,
    calculated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    calculated_by UUID NOT NULL
);

CREATE INDEX idx_coverage_calculations_schedule_version ON coverage_calculations(schedule_version_id);
CREATE INDEX idx_coverage_calculations_hospital_date ON coverage_calculations(hospital_id, calculation_date DESC);
CREATE INDEX idx_coverage_calculations_period ON coverage_calculations(calculation_period_start_date, calculation_period_end_date);
```

#### 008: Audit Logs Table
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    action VARCHAR(255) NOT NULL,
    resource VARCHAR(255) NOT NULL,
    old_values TEXT,
    new_values TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ip_address INET
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
```

#### 009: Users Table (for security/auth)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'USER'
        CHECK (role IN ('ADMIN', 'SCHEDULER', 'VIEWER', 'USER')),
    hospital_id UUID REFERENCES hospitals(id),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_hospital ON users(hospital_id, active);
CREATE INDEX idx_users_role ON users(role);
```

#### 010: Job Queue Table (for Asynq fallback or tracking)
```sql
CREATE TABLE job_queue (
    id UUID PRIMARY KEY,
    job_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETE', 'FAILED', 'RETRY')),
    result JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_job_queue_status ON job_queue(status);
CREATE INDEX idx_job_queue_type ON job_queue(job_type, status);
CREATE INDEX idx_job_queue_created ON job_queue(created_at DESC);
```

---

## Implementation Roadmap (Phase 0b - 3-5 Days)

### Day 1: Complete Migrations (4 hours)
```bash
# 1. Create migration files for tables 005-010 (following pattern of 001-004)
# 2. Test all migrations with golang-migrate tool
# 3. Verify schema with psql

go install -tags 'postgres' github.com/golang-migrate/migrate/cmd/migrate@latest
migrate -path ./migrations -database "postgres://localhost:5432/schedcu_dev" up
```

### Day 2: Implement sqlc Configuration (3 hours)
```bash
# Set up sqlc for type-safe query generation
cat > sqlc.yaml << 'EOF'
version: "2"
sql:
  - engine: "postgresql"
    queries: "./internal/repository/queries.sql"
    schema: "./migrations"
    gen:
      go:
        out: "./internal/repository/postgres/generated"
        package: "postgres"

EOF

go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate
```

### Day 3: Implement PostgreSQL Repositories (4 hours)
```
internal/repository/postgres/
├── person.go              # CRUD for persons
├── schedule_version.go    # CRUD for schedule versions
├── shift_instance.go      # CRUD for shifts
├── assignment.go          # CRUD for assignments
├── coverage.go            # CRUD for coverage calculations
├── audit_log.go           # CRUD for audit logs
├── queries.sql            # sqlc query definitions
└── postgres_test.go       # Integration tests with real DB
```

Each repository file should:
- Implement the repository interface from `internal/repository/`
- Have comprehensive tests using Testcontainers
- Include query count assertions
- Support transaction management for batch operations

### Day 4: Implement Core Services (5 hours)
```
internal/service/
├── schedule_version_service.go      # Version promotion/archival workflow
├── shift_instance_service.go        # Shift management
├── assignment_service.go            # Assignment management
├── coverage_calculator_service.go   # Dynamic coverage calculation (batch queries, no N+1)
├── ods_import_service.go           # ODS file import with validation
├── amion_import_service.go         # Amion scraper orchestration
├── schedule_orchestrator_service.go # 3-phase workflow
└── service_test.go                 # Comprehensive service tests
```

### Day 5: Integrate Asynq & Validate (4 hours)
```bash
# 1. Install Asynq
go get github.com/hibiken/asynq

# 2. Create job handlers
internal/job/
├── scheduler.go            # Queue job creation
├── handlers.go             # Job execution handlers
└── job_test.go            # Job tests

# 3. Verify all 85+ tests pass
go test ./... -v --cover
```

---

## Service Implementation Specification

### 1. ScheduleVersionService
```go
type ScheduleVersionService struct {
    repo *postgres.ScheduleVersionRepository
}

// Methods needed:
- CreateVersion(ctx, hospitalID, startDate, endDate) -> *ScheduleVersion
- GetVersion(ctx, id) -> *ScheduleVersion, error
- ListVersions(ctx, hospitalID, status) -> []*ScheduleVersion, error
- PromoteToProduction(ctx, id, promoterID) -> error
- Archive(ctx, id, archiverID) -> error
- GetActiveVersion(ctx, hospitalID, date) -> *ScheduleVersion, error
```

### 2. DynamicCoverageCalculator
```go
type DynamicCoverageCalculator struct {
    repo *postgres.AssignmentRepository
}

// Key method with batch query optimization:
- CalculateCoverage(ctx, scheduleVersionID, periodStart, periodEnd) -> *CoverageCalculation
  // Load all assignments for period in 1-2 queries (no N+1)
  // Apply coverage rules:
  //   1. Count assignments by position
  //   2. Check specialty constraints
  //   3. Validate mandatory coverage
  //   4. Return comprehensive validation result
```

### 3. ODSImportService
```go
type ODSImportService struct {
    shiftRepo *postgres.ShiftInstanceRepository
    assignmentRepo *postgres.AssignmentRepository
    coverageCalc *DynamicCoverageCalculator
}

// Methods:
- ImportODSFile(ctx, hospitalID, file) -> *ScrapeBatch, *validation.Result
  // 1. Parse ODS file (with error collection, not fail-fast)
  // 2. Validate shifts against known types
  // 3. Create ShiftInstances
  // 4. Return validation result with all issues (ERROR, WARNING, INFO)
```

### 4. AmionImportService
```go
type AmionImportService struct {
    assignmentRepo *postgres.AssignmentRepository
    batchRepo *postgres.ScrapeBatchRepository
}

// Methods:
- ScrapeAndImport(ctx, hospitalID, months int) -> *ScrapeBatch, error
  // 1. Scrape Amion (goquery from Spike 1)
  // 2. Parse shifts and assignments
  // 3. Create batch with PENDING state
  // 4. Create assignments with source=AMION
  // 5. Mark batch COMPLETE or FAILED
```

### 5. ScheduleOrchestrator
```go
type ScheduleOrchestrator struct {
    ods *ODSImportService
    amion *AmionImportService
    coverage *DynamicCoverageCalculator
    versionService *ScheduleVersionService
}

// Main workflow:
- ExecuteFullWorkflow(ctx, hospitalID, odsFile, amionMonths) -> error
  // Phase 1: ODS Import
  // Phase 2: Amion Import (async via Asynq)
  // Phase 3: Coverage Resolution
  // Returns errors and validation results for each phase
```

---

## Testing Strategy (Phase 0b)

### Integration Tests (Testcontainers)
```bash
# Spin up real PostgreSQL for each test
// Start container before test
container := startPostgresContainer()
defer container.Terminate(ctx)

// Run migrations
runMigrations(container)

// Create repositories with real DB
repo := postgres.NewScheduleVersionRepository(db)

// Test actual queries with real database
```

### Query Count Assertions
```go
// Every service test validates query efficiency
// Example:
func TestCalculateCoverageNoN1Plus(t *testing.T) {
    // Query count should be:
    // 1. Load assignments for date range (1 query)
    // 2. Load persons for coverage validation (1 query)
    // 3. Store calculation result (1 query)
    // Total: 3 queries max, never N+1

    assert.Equal(t, 3, metrics.QueryCount())
}
```

---

## Success Criteria for Phase 0b Completion

- [x] Schema design complete (10 tables specified)
- [x] Migration files created for tables 001-004
- [ ] Migration files created for tables 005-010 (4 hours work)
- [ ] All migrations tested with golang-migrate
- [ ] PostgreSQL repositories implemented and tested (10+ tests each)
- [ ] Core services working with real database (20+ tests)
- [ ] Asynq integration complete
- [ ] Total test coverage 85%+ across all layers
- [ ] Zero N+1 query problems (validated by tests)
- [ ] All code compiles: `go build ./cmd/server`

---

## Go/No-Go Criteria for Phase 1

**CANNOT START PHASE 1 WITHOUT:**
- ✅ Entity model complete and tested
- ✅ Validation framework working
- ⚠️ PostgreSQL schema complete (IN PROGRESS)
- ⚠️ Repositories implemented (PENDING)
- ⚠️ Services functional (PENDING)
- [ ] 85%+ test coverage
- [ ] All endpoints working with real database
- [ ] Asynq integration verified

---

## Files to Create in Phase 0b

### Migration Files (8 pairs = 16 files)
```
migrations/
├── 005_create_shift_instances.up.sql/.down.sql
├── 006_create_assignments.up.sql/.down.sql
├── 007_create_coverage_calculations.up.sql/.down.sql
├── 008_create_audit_logs.up.sql/.down.sql
├── 009_create_users.up.sql/.down.sql
└── 010_create_job_queue.up.sql/.down.sql
```

### Configuration
```
sqlc.yaml
golang-migrate.yaml
```

### Repository Layer
```
internal/repository/
├── postgres/
│   ├── person.go + person_test.go
│   ├── schedule_version.go + schedule_version_test.go
│   ├── shift_instance.go + shift_instance_test.go
│   ├── assignment.go + assignment_test.go
│   ├── coverage_calculation.go + coverage_calculation_test.go
│   ├── audit_log.go + audit_log_test.go
│   ├── queries.sql
│   ├── postgres.go (connection management)
│   └── postgres_test.go (setup/teardown)
└── db.go (database interface)
```

### Service Layer
```
internal/service/
├── schedule_version_service.go + schedule_version_service_test.go
├── shift_instance_service.go + shift_instance_service_test.go
├── assignment_service.go + assignment_service_test.go
├── coverage_calculator_service.go + coverage_calculator_service_test.go
├── ods_import_service.go + ods_import_service_test.go
├── amion_import_service.go + amion_import_service_test.go
└── schedule_orchestrator.go + schedule_orchestrator_test.go
```

### Job System
```
internal/job/
├── scheduler.go
├── handlers.go
└── job_test.go
```

---

## Token & Time Estimate

**Estimated effort for Phase 0b**: 3-5 days with 1 senior engineer
- Migrations: 4 hours
- sqlc setup: 2 hours
- Repositories: 8 hours
- Services: 10 hours
- Asynq integration: 3 hours
- Testing & validation: 5 hours

**Total**: ~32 hours of focused development

---

## Next Action

Once Phase 0b is complete:
1. Start Phase 1 with confidence (core services working)
2. Implement API handlers for all endpoints
3. Complete security testing
4. Prepare for Phase 2 (Amion scraper, ODS import in production)

**Foundation is solid. Schema is design-complete. Ready to execute.**

