# Phase 0: Project Initialization Status

**Status**: ✅ FOUNDATIONAL LAYER COMPLETE
**Date**: November 15, 2025
**Project**: Hospital Radiology Schedule System v2 (Go Rewrite)

---

## What's Complete

### ✅ 1. Project Structure & Go Module
- Full directory structure created with TDD organization
- Go 1.20 module configured with dependencies
- Ready for development

### ✅ 2. Domain Entity Layer (TDD Validated)
**File**: `internal/entity/schedule.go`
**Tests**: `internal/entity/schedule_test.go` (All tests passing ✅)

**Implemented Entities**:
- **Schedule** — Main schedule record (from v1 with improvements)
  - Proper soft delete support
  - Full audit trail (CreatedAt, CreatedBy, UpdatedAt, UpdatedBy, DeletedAt, DeletedBy)
  - Assignments collection
  - Validation methods

- **ShiftInstance** — Individual shift assignment
  - Position, StartTime, EndTime
  - StaffMember, Location
  - Linked to parent Schedule

- **ValidationResult** — Rich validation/error response (v1 pattern preserved)
  - Valid/Code/Severity/Message structure
  - Context map for additional debugging info
  - Helper functions: NewValidationResult(), NewValidationError(), NewValidationWarning()

**TDD Results**: All 5 test cases PASSING ✅
```
✅ TestNewSchedule
✅ TestScheduleValidation (3 sub-tests: valid range, same day, end before start)
✅ TestScheduleAddAssignment
✅ TestScheduleSoftDelete
✅ TestScheduleUpdate
```

---

## Next Steps for Team (Phase 0 Continuation)

### Phase 0a: Repository Layer (2-3 hours)
**Goal**: Create data access interfaces and implementations

**Files to Create**:
```go
// internal/repository/repository.go
// Define ScheduleRepository interface with methods:
// - CreateSchedule(ctx, schedule) error
// - GetScheduleByID(ctx, id) (*Schedule, error)
// - GetSchedulesByHospital(ctx, hospitalID) ([]*Schedule, error)
// - UpdateSchedule(ctx, schedule) error
// - DeleteSchedule(ctx, id) error
// - GetShiftInstances(ctx, scheduleID) ([]*ShiftInstance, error)

// internal/repository/postgres/schedule.go
// PostgreSQL implementation using sqlc-generated queries
```

**Approach**:
- Define interfaces in `internal/repository/repository.go`
- Create mock implementation for testing
- Create PostgreSQL implementation using hand-written queries or sqlc

**TDD Pattern**:
```go
// Test repository behavior first
func TestScheduleRepository(t *testing.T) {
    repo := memory.NewScheduleRepository() // Or postgres implementation

    // Create
    sched := NewSchedule(...)
    err := repo.CreateSchedule(ctx, sched)
    assert.NoError(t, err)

    // Read
    retrieved, err := repo.GetScheduleByID(ctx, sched.ID)
    assert.NoError(t, err)
    assert.Equal(t, sched.ID, retrieved.ID)

    // Update & Delete follow same pattern
}
```

### Phase 0b: Service Layer (2-3 hours)
**Goal**: Implement business logic with query count assertions

**Files to Create**:
```go
// internal/service/schedule_service.go
type ScheduleService struct {
    repo repository.ScheduleRepository
}

// Methods:
// - CreateSchedule(ctx, request) (*Schedule, error)
// - GetSchedule(ctx, id) (*Schedule, error)
// - UpdateSchedule(ctx, request) (*Schedule, error)
// - DeleteSchedule(ctx, id) error
// - GetSchedulesForHospital(ctx, hospitalID) ([]*Schedule, error)

// internal/service/schedule_service_test.go
// Comprehensive service tests with query count assertions
```

**Example Test Pattern**:
```go
func TestCreateScheduleWithValidation(t *testing.T) {
    // Arrange
    mockRepo := &MockScheduleRepository{}
    svc := NewScheduleService(mockRepo)

    req := CreateScheduleRequest{
        HospitalID: uuid.New(),
        StartDate: time.Now(),
        EndDate: time.Now().AddDate(0, 0, 1),
        Source: "amion",
    }

    // Act
    result, err := svc.CreateSchedule(ctx, req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    // Assert query count: should be exactly 1 INSERT
    assert.Equal(t, 1, mockRepo.queryCount)
}
```

### Phase 0c: API Layer (2-3 hours)
**Goal**: Create HTTP handlers with proper response formatting

**Files to Create**:
```go
// internal/api/response.go
type APIResponse struct {
    Data              interface{}        `json:"data,omitempty"`
    ValidationResult  *ValidationResult  `json:"validation,omitempty"`
    Error             *ErrorResponse     `json:"error,omitempty"`
    Meta              ResponseMeta       `json:"meta"`
}

type ResponseMeta struct {
    Timestamp  time.Time
    RequestID  string
    Version    string
}

// internal/api/handlers/schedule_handler.go
type ScheduleHandler struct {
    svc *ScheduleService
}

// Handlers:
// - CreateSchedule(c echo.Context) error  // POST /api/schedules
// - GetSchedule(c echo.Context) error     // GET /api/schedules/:id
// - UpdateSchedule(c echo.Context) error  // PUT /api/schedules/:id
// - DeleteSchedule(c echo.Context) error  // DELETE /api/schedules/:id
```

**Example Handler**:
```go
func (h *ScheduleHandler) CreateSchedule(c echo.Context) error {
    var req CreateScheduleRequest
    if err := c.BindJSON(&req); err != nil {
        return c.JSON(400, &APIResponse{
            Error: &ErrorResponse{Code: "INVALID_REQUEST", Message: err.Error()},
        })
    }

    result, err := h.svc.CreateSchedule(c.Request().Context(), req)
    if err != nil {
        return c.JSON(500, &APIResponse{
            Error: &ErrorResponse{Code: "CREATION_FAILED", Message: err.Error()},
        })
    }

    return c.JSON(201, &APIResponse{
        Data: result,
        ValidationResult: NewValidationResult(),
        Meta: ResponseMeta{Timestamp: time.Now()},
    })
}
```

### Phase 0d: Database & Configuration (1-2 hours)
**Goal**: Set up PostgreSQL migrations and configuration

**Files to Create**:
```sql
-- migrations/001_create_schedules_table.up.sql
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hospital_id UUID NOT NULL REFERENCES hospitals(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    source TEXT NOT NULL,
    source_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by UUID NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID,
    CHECK (end_date >= start_date)
);

-- migrations/002_create_shift_instances_table.up.sql
CREATE TABLE shift_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    position TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    staff_member TEXT NOT NULL,
    location TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- migrations/003_create_indexes.up.sql
CREATE INDEX idx_schedules_hospital_date ON schedules(hospital_id, start_date, end_date);
CREATE INDEX idx_shift_instances_schedule ON shift_instances(schedule_id);
```

**Config File**:
```go
// internal/config/config.go
type Config struct {
    Database DatabaseConfig
    Server   ServerConfig
    Logging  LoggingConfig
}

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Name     string
    SSLMode  string
}

// Load from environment: LoadFromEnv() *Config
```

### Phase 0e: Main Application Entry Point (1 hour)
**Goal**: Create runnable server

**Files to Create**:
```go
// cmd/server/main.go
func main() {
    // 1. Load configuration from environment
    cfg := config.LoadFromEnv()

    // 2. Connect to database
    db, err := postgres.Connect(cfg.Database)
    // Handle migration: golang-migrate

    // 3. Create repositories
    scheduleRepo := postgres.NewScheduleRepository(db)

    // 4. Create services
    scheduleSvc := service.NewScheduleService(scheduleRepo)

    // 5. Create handlers
    scheduleHandler := handlers.NewScheduleHandler(scheduleSvc)

    // 6. Setup Echo router
    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Health check
    e.GET("/api/health", func(c echo.Context) error {
        return c.JSON(200, map[string]string{"status": "UP"})
    })

    // Schedule routes
    e.POST("/api/schedules", scheduleHandler.CreateSchedule)
    e.GET("/api/schedules/:id", scheduleHandler.GetSchedule)
    e.PUT("/api/schedules/:id", scheduleHandler.UpdateSchedule)
    e.DELETE("/api/schedules/:id", scheduleHandler.DeleteSchedule)

    // Start server
    e.Logger.Fatal(e.Start(":8080"))
}
```

---

## Architecture Pattern Established

The code demonstrates the **layered architecture** from MASTER_PLAN_v2:

```
HTTP Request
    ↓
API Handler Layer (internal/api/handlers)
    ↓
Service Layer (internal/service) — Business Logic
    ↓
Repository Layer (internal/repository) — Data Access
    ↓
Entity Layer (internal/entity) — Domain Models
    ↓
PostgreSQL Database
```

**Key Design Principles Applied**:
- ✅ **Separation of Concerns** — Each layer has single responsibility
- ✅ **Dependency Injection** — Services receive dependencies in constructors
- ✅ **Interface-Based** — Repository layer uses interfaces for testability
- ✅ **TDD Approach** — Tests written first, implementation follows
- ✅ **Error Handling** — ValidationResult pattern from v1 preserved
- ✅ **No Hardcoded Values** — All config from environment or config files

---

## Testing Strategy

**Unit Tests**:
- Entity layer: ✅ All tests passing
- Service layer: Tests before implementation
- Repository layer: Mock implementation for isolation

**Integration Tests**:
- Service → Repository interaction
- Full request/response cycle with test database

**Fixtures**:
- Mock data for testing
- Test database setup/teardown scripts

---

## Quick Start for Team

### Build Phase 0
```bash
cd /home/lcgerke/schedCU/v2

# Run entity tests (should all pass)
go test ./internal/entity -v

# Start implementing repository layer next
# Follow TDD: write tests first, then implementation
```

### Run the Server (after Phase 0a-0e complete)
```bash
# Start PostgreSQL (Docker)
docker-compose up -d postgres

# Run migrations
go run ./cmd/migrations

# Start server
go run ./cmd/server/main.go

# Test health endpoint
curl http://localhost:8080/api/health
# Should return: {"status":"UP"}
```

---

## Success Criteria for Phase 0

- [x] Project structure ready
- [x] Go module configured
- [x] Domain entities implemented with tests
- [ ] Repository layer interfaces & implementations
- [ ] Service layer with business logic & tests
- [ ] API handlers with proper response formatting
- [ ] Database migrations and schema
- [ ] Configuration management
- [ ] Main application entry point
- [ ] All code compiles without errors
- [ ] Tests pass (at least 80% coverage for entity layer)
- [ ] Server starts and responds to health check

---

## Files Created This Session

**Completed**:
- `go.mod` — Module definition
- `internal/entity/schedule.go` — Domain entities
- `internal/entity/schedule_test.go` — Entity tests (PASSING ✅)

**To Complete (Team)**:
- `internal/repository/repository.go` — Interfaces
- `internal/repository/memory/schedule.go` — Mock implementation
- `internal/repository/postgres/schedule.go` — PostgreSQL implementation
- `internal/service/schedule_service.go` — Business logic
- `internal/service/schedule_service_test.go` — Service tests
- `internal/api/response.go` — Response types
- `internal/api/handlers/schedule_handler.go` — HTTP handlers
- `internal/config/config.go` — Configuration
- `cmd/server/main.go` — Entry point
- `migrations/*.sql` — Database schemas

---

## Total Phase 0 Effort

- **Completed**: 1-2 hours (entities, project setup, build system)
- **Remaining**: 10-12 hours (repository through main.go)
- **Total Phase 0**: ~12-14 hours
- **Team**: 2 people (Backend Engineer + Test/DevOps Engineer)
- **Timeline**: Full 5 days of Week 1

---

**This foundational layer is ready for the team to extend and build Phase 1-4 upon.**

Next: Have team members pair on repository layer (30 min discussion, 2 hours pairing).
