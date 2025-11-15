# Phase 0 Implementation Plan: Core v2 Application Structure

**Duration**: 1-2 days of focused development
**Approach**: TDD, test-first implementation
**Deliverables**: Working project scaffold with core entities, repositories, services, and API handlers

---

## Part 1: Design Phase (30 minutes)

### 1.1 Core Architecture

The v2 application will follow layered architecture:

```
cmd/server/main.go                    # Entry point
internal/
├── entity/                           # Domain types
│   ├── schedule.go                   # Schedule entity
│   ├── shift_instance.go             # ShiftInstance entity
│   ├── scrape_batch.go               # ScrapeBatch entity
│   ├── assignment.go                 # Assignment entity
│   └── validation_result.go          # ValidationResult (from v1 pattern)
├── repository/                       # Data access layer (sqlc-generated)
│   ├── schedule.go                   # ScheduleRepository interface + impl
│   ├── shift_instance.go             # ShiftInstanceRepository interface + impl
│   └── queries.sql                   # sqlc query definitions
├── service/                          # Business logic layer
│   ├── schedule_service.go           # ScheduleService
│   ├── coverage_calculator.go        # DynamicCoverageCalculator (from v1)
│   └── validation_service.go         # ValidationService
├── api/                              # HTTP handlers
│   ├── schedule_handler.go           # Schedule endpoints
│   ├── response.go                   # API response types
│   └── middleware.go                 # Auth, logging, etc.
└── database/                         # Database setup
    └── migrations.go                 # Migration runner
```

### 1.2 Core Entities (from v1 schema)

**Schedule** — Main schedule record covering a date range
```go
type Schedule struct {
    ID          uuid.UUID
    StartDate   time.Time
    EndDate     time.Time
    HospitalID  uuid.UUID
    Source      string              // 'amion', 'ods_file', 'manual'
    Assignments []ShiftInstance      // Embedded assignments
    CreatedAt   time.Time
    CreatedBy   uuid.UUID
    UpdatedAt   time.Time
    UpdatedBy   uuid.UUID
    DeletedAt   *time.Time           // Soft delete
}
```

**ShiftInstance** — Individual shift assignment
```go
type ShiftInstance struct {
    ID          uuid.UUID
    ScheduleID  uuid.UUID
    Position    string               // "ER Doctor", "Nurse", etc.
    StartTime   string               // "08:00"
    EndTime     string               // "16:00"
    StaffMember string               // "John Doe"
    Location    string               // "Main ER", "ICU", etc.
    CreatedAt   time.Time
}
```

**ValidationResult** — Rich validation/error response (from v1)
```go
type ValidationResult struct {
    Valid       bool
    Code        string               // "VALIDATION_SUCCESS", "PARSE_ERROR", etc.
    Severity    string               // "INFO", "WARNING", "ERROR"
    Message     string
    Context     map[string]interface{}
}
```

### 1.3 Test Strategy

**Unit Tests**:
- Entity constructors & validation
- Service business logic
- ValidationResult behavior

**Integration Tests**:
- Repository operations with test database
- Service → Repository integration
- API handler → Service integration

**Database Tests**:
- Migration execution
- Repository CRUD operations
- Query optimization (query count assertions)

---

## Part 2: Project Initialization (30 minutes)

### 2.1 Create Go Project Structure
### 2.2 Set up Go module with dependencies
### 2.3 Configure database with migrations
### 2.4 Create entity types (TDD approach)

---

## Part 3: Implementation (90 minutes)

### Phase 3.1: Entity Layer (20 minutes)
- Schedule entity with validation
- ShiftInstance entity
- ValidationResult struct
- Unit tests for each

### Phase 3.2: Repository Layer (30 minutes)
- Define repository interfaces
- Basic CRUD operations
- sqlc query generation
- Integration tests

### Phase 3.3: Service Layer (20 minutes)
- ScheduleService with business logic
- ValidationService
- Integration tests

### Phase 3.4: API Layer (15 minutes)
- Echo HTTP handlers
- Response formatting
- API tests

### Phase 3.5: Integration & Testing (5 minutes)
- Full-stack integration tests
- Error handling verification

---

## Part 4: Success Criteria

✅ Project builds without errors: `go build ./cmd/server`
✅ All tests pass: `go test ./...`
✅ Database migrations run: `go run ./cmd/server migrate`
✅ Server starts: `go run ./cmd/server`
✅ Health endpoint works: `GET /api/health`
✅ Create schedule endpoint works: `POST /api/schedules`
✅ Test coverage >80% for core modules
✅ No hardcoded values (all configurable)
✅ Proper error handling & validation
✅ Documentation/comments explain WHY

---

## Implementation Strategy

We will use **Test-Driven Development (TDD)**:

1. Write test first (RED)
2. Implement minimal code to pass test (GREEN)
3. Refactor for clarity (REFACTOR)

This ensures:
- Tests serve as documentation
- All code is covered by tests
- Design emerges from testability requirements
- Changes don't break existing functionality

---

## Expected Outcome

After Part 4, you will have:
- ✅ Working v2 project scaffold
- ✅ Example entities (Schedule, ShiftInstance)
- ✅ Example repository layer
- ✅ Example service layer with business logic
- ✅ Example API handler with proper response formatting
- ✅ Comprehensive test suite demonstrating patterns
- ✅ Database migrations and setup
- ✅ Ready for Phase 1: Core Services expansion

This provides a **template** for the team to follow when implementing remaining entities and services.

---

## Let's Begin Implementation

Proceeding with Part 2...
