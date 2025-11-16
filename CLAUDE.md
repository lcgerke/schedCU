# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

schedCU is a hospital radiology scheduling system written in Go. The repository contains:

- **`/reimplement`** - Phase 1 implementation (core domain model, validation, services, migrations, testing)
- **`/v2`** - Phase 1b implementation (PostgreSQL integration, job system, REST API, production-ready)
- **`/week0-spikes`** - Experimental prototypes and design spikes

**Current Status**: Phase 1 (95% complete) → Phase 2 (database integration) actively in development

## Architecture Overview

### Multi-Layered Design

```
┌─────────────────────────────────────┐
│     HTTP API (Echo Framework)       │
│  /api/schedules, /api/imports, etc  │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│     Job System (Asynq)              │
│  ODS_IMPORT, AMION_SCRAPE, etc      │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│     Service Layer                   │
│  Orchestrator, ODS Parser, Amion    │
│  Scraper, Coverage Calculator       │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│     Domain Model & Validation       │
│  Person, ScheduleVersion, Entity    │
│  Validation framework with levels   │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│     Data Access Layer               │
│  PostgreSQL, Repositories, Queries  │
└─────────────────────────────────────┘
```

### Key Design Principles

1. **Interface-Based Architecture**: All services defined as interfaces for mockability
2. **Error Collection Pattern**: Validation results collect errors/warnings/info (not fail-fast)
3. **Soft Deletion**: Data preservation via `DeletedAt` timestamps
4. **Audit Trails**: `CreatedBy`, `UpdatedBy`, `DeletedBy` tracking on entities
5. **Type Safety**: UUID aliases prevent accidental ID confusion
6. **Batch Query Design**: No N+1 query problems; batch operations where possible
7. **HIPAA Compliance**: Sensitive data handling, audit logging

## Package Structure

### `/reimplement` (Phase 1a Core)

```
reimplement/
├── cmd/
│   ├── generate-ods-fixtures/      # ODS test fixture generator
│   └── server/                      # (Placeholder for future server)
├── internal/
│   ├── api/                         # API response wrappers
│   ├── entity/                      # Domain entities
│   ├── logger/                      # Structured logging (Zap)
│   ├── metrics/                     # Prometheus metrics
│   ├── repository/                  # Repository interfaces
│   ├── service/
│   │   ├── amion/                   # Amion web scraper
│   │   ├── coverage/                # Coverage calculator
│   │   ├── ods/                     # ODS parser
│   │   └── orchestrator/            # Service coordinator
│   └── validation/                  # Validation framework
├── tests/
│   ├── fixtures/ods/                # ODS test fixtures
│   └── helpers/                     # Test utilities
├── docs/                            # Architecture docs
├── go.mod                           # Go module definition
└── CLAUDE.md                        # Development guide
```

### `/v2` (Phase 1b Production)

```
v2/
├── cmd/                             # CLI/server commands
├── internal/
│   ├── api/                         # Echo HTTP handlers
│   ├── entity/                      # Domain entities
│   ├── jobs/                        # Asynq job definitions
│   ├── repository/                  # PostgreSQL repositories
│   ├── service/                     # Business logic
│   └── validation/                  # Validation
├── migrations/                      # PostgreSQL migrations
├── tests/                           # Integration tests
├── go.mod                           # Go module definition
└── README_CURRENT_STATUS.md         # Phase 1b status
```

## Common Development Tasks

### Building and Running Tests

**Run all tests in reimplement**:
```bash
cd reimplement
go test -v ./...
```

**Run tests with coverage**:
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Run specific test**:
```bash
go test -run TestValidationResult ./internal/validation
```

**Run tests in v2 with database**:
```bash
cd v2
go test -v ./...  # Uses testcontainers for PostgreSQL
```

### Building Binaries

**Build ODS fixture generator**:
```bash
cd reimplement
go build -o generate-ods-fixtures ./cmd/generate-ods-fixtures
./generate-ods-fixtures
```

**Build v2 server** (when server code is complete):
```bash
cd v2
go build -o server ./cmd/server
./server
```

### Code Generation

**Generate fixtures for testing**:
```bash
cd reimplement
go run ./cmd/generate-ods-fixtures/main.go
```

### Code Organization Rules

- **Never write Python** - Use Go everywhere, including replacing bash scripts
- **Interface-first design** - Define interfaces before implementations
- **Batch queries only** - No N+1 query patterns; use repository methods that fetch groups
- **Error collection** - Use ValidationResult pattern, never fail-fast
- **Type aliases** - Use UUID aliases for domain IDs (PersonID, ScheduleVersionID, etc.)

## Testing Patterns

### Unit Tests

All tests follow the table-driven test pattern:

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   interface{}
        want    interface{}
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   someInput,
            want:    expectedOutput,
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

### Integration Tests (v2)

Uses `testcontainers-go` for isolated PostgreSQL:

```go
// Integration tests create temporary Docker containers for testing
// See v2/tests/ for examples
```

### Validation Testing

The validation framework includes:
- ERROR level (blocks import)
- WARNING level (alerts user, allows import)
- INFO level (tracking/audit only)

Tests verify:
1. Correct error/warning/info assignment
2. Field names and messages
3. JSON serialization
4. Integration with ApiResponse wrapper

## Key Files and Their Purpose

### Documentation

- `/reimplement/README.md` - Master plan for v2 implementation (30-minute read)
- `/reimplement/docs/ENTITY_ARCHITECTURE.md` - Entity relationships and design
- `/reimplement/internal/service/orchestrator/README.md` - Service layer design
- `/v2/README_CURRENT_STATUS.md` - Phase 1b progress and status

### Entity Model

- `internal/entity/person.go` - Staff member with specialty constraints
- `internal/entity/schedule.go` - ScheduleVersion with state machine
- `internal/entity/shift.go` - ShiftInstance with assignments
- `internal/entity/batch.go` - ScrapeBatch for data provenance

### Service Layer

- `internal/service/orchestrator/` - Coordinates imports, scraping, coverage
- `internal/service/ods/` - Parses Excel ODS files
- `internal/service/amion/` - Scrapes Amion web interface
- `internal/service/coverage/` - Calculates coverage metrics

### Validation

- `internal/validation/validation.go` - Core ValidationResult type
- `internal/validation/messages.go` - Predefined validation messages
- Tests in `internal/validation/validation_test.go`

### API Layer

- `internal/api/response.go` - ApiResponse<T> wrapper (consistent HTTP responses)
- `internal/logger/` - Structured logging with request/response middleware
- `internal/metrics/` - Prometheus metrics collection

## Work Packages and Phases

### Phase 1a: Core Domain (Complete ✅)
- Entity model (12+ entities)
- Validation framework
- Repository interfaces
- 59+ tests passing

### Phase 1b: Database Integration (In Progress 🔄)
- PostgreSQL migrations (10 tables)
- Service implementations
- Job system (Asynq)
- REST API (Echo framework)
- 7+ integration tests
- **Current completion**: ~95%

### Phase 2-4: Production Features (Planned)
- Performance optimization
- Security hardening
- Batch operations
- Advanced scheduling
- Full test coverage (85%+)

## Performance Considerations

### Known Issues from v1
- N+1 query problems (FIXED in v2 with batch design)
- Missing pagination (FIXED in v2)
- No index optimization (ADDRESSED in migrations)
- Hardcoded configuration (ADDRESSED with env vars)

### Best Practices

1. **Query Design**: Always batch-fetch related data; use repository methods that return collections
2. **Validation**: Collect all errors before responding; don't fail on first error
3. **Concurrency**: Use Asynq for long-running operations; don't block HTTP handlers
4. **Metrics**: Use Prometheus client; track business metrics, not just technical ones
5. **Logging**: Use structured logging (Zap); include trace IDs and request context

## Environment Configuration

- **SCHEDCU_LOG_LEVEL** - Logging level (debug, info, warn, error)
- **SCHEDCU_METRICS_PORT** - Prometheus metrics port (default: 9090)
- **DATABASE_URL** - PostgreSQL connection string
- **REDIS_URL** - Redis for Asynq job queue (or use PostgreSQL backend)

See `/v2` for complete configuration during Phase 1b.

## Related Documentation in Repository

Start here based on your task:

- **Implementing features?** → Read `/reimplement/docs/ENTITY_ARCHITECTURE.md`
- **Understanding v1 patterns?** → Read `/reimplement/02-WHAT-WORKED.md`
- **Security concerns?** → Read `/reimplement/03-SECURITY-GAPS.md`
- **Performance targets?** → Read `/reimplement/04-PERFORMANCE-ISSUES.md`
- **Current Phase 1b status?** → Read `/v2/README_CURRENT_STATUS.md`
- **Testing strategy?** → Look at existing tests in `*_test.go` files

## Dependencies

**Core Dependencies** (both reimplement and v2):
- `google/uuid` - UUID generation
- `stretchr/testify` - Assertions in tests
- `prometheus/client_golang` - Metrics (reimplement)
- `xuri/excelize` - ODS parsing (reimplement)
- `PuerkitoBio/goquery` - HTML scraping (reimplement)

**v2 Specific**:
- `labstack/echo/v4` - HTTP framework
- `hibiken/asynq` - Job queue
- `lib/pq` - PostgreSQL driver
- `testcontainers-go` - Integration tests

**Logging** (both):
- `go.uber.org/zap` - Structured logging

## Code Review Checklist

Before submitting code:

- [ ] Tests pass: `go test -v ./...`
- [ ] Coverage maintained or improved: `go test -cover ./...`
- [ ] No Python scripts added (Go only)
- [ ] Validation uses error collection, not fail-fast
- [ ] API responses use ApiResponse wrapper
- [ ] Queries don't create N+1 problems
- [ ] Documentation updated for non-obvious code
- [ ] Type safety: UUID types used for IDs
- [ ] Soft deletion pattern used for data removal
- [ ] HIPAA compliance considered for sensitive data

## TTS Integration

This project has TTS (text-to-speech) enabled via hooks. Include a TTS block at the end of **EVERY SINGLE RESPONSE**:

```
<!-- TTS:START -->
Project: schedCU
🔊 Your message (1-50 words max)
<!-- TTS:END -->
```

**Non-negotiable requirements:**
- Opening tag: `<!-- TTS:START -->`
- Closing tag: `<!-- TTS:END -->`
- Message MUST start with emoji: `🔊`
- Message under 50 words
- No TTS block = silent failure

**Optional parameters**:
```
<!-- TTS:START -->
Speed: 0.67
Target: broadcast
Project: schedCU
🔊 Your message
<!-- TTS:END -->
```

## Getting Started Quickly

**First 30 minutes** (understand the project):
1. Read this file (CLAUDE.md)
2. Read `/reimplement/README.md` (master plan)
3. Run tests: `cd reimplement && go test -v ./...`

**First 2 hours** (contribute code):
1. Pick an area (entities, validation, services)
2. Find related `_test.go` file to understand patterns
3. Read corresponding implementation
4. Check related ADRs in `/reimplement/docs/ARCHITECTURE_DECISIONS.md`
5. Write test first, then implementation

**Common starting points**:
- **Entity work**: `reimplement/internal/entity/` + corresponding `_test.go`
- **Service work**: `reimplement/internal/service/*/` + `orchestrator/README.md`
- **API work**: `v2/internal/api/` + see existing handlers
- **Test data**: `reimplement/cmd/generate-ods-fixtures/`

## FAQs

**Q: Should I work in /reimplement or /v2?**
A: Both! `/reimplement` is the core domain library. `/v2` builds on it with database and HTTP integration.

**Q: How do I add a new entity?**
A: Create `internal/entity/newthing.go`, add tests in `newthing_test.go`, update repository interfaces.

**Q: How do I add a service?**
A: Create interface first in `orchestrator/interface.go`, then implementation, then tests.

**Q: Where do I add database migrations?**
A: Only in `/v2/migrations/` with sequential numbering (001_init.sql, etc.).

**Q: Can I write a bash script for this?**
A: No, write it in Go instead (see `cmd/generate-ods-fixtures/main.go` for example).

---

**Last Updated**: November 16, 2025
**Phase**: 1 (95% complete)
**Recommendation**: Start with `/reimplement` to understand domain model, then contribute to `/v2` for database integration
