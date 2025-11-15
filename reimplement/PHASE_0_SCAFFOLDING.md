# Phase 0: Project Scaffolding & Infrastructure Setup

**Duration**: Week 1, 5 days (parallel with v1 security fixes)
**Team**: Backend Engineer + Test/DevOps Engineer (2 people)
**Goal**: Create production-ready project structure, database migrations, and local development environment

---

## Overview

Phase 0 sets up the foundation for v2 development:
1. Create project structure from template
2. Set up database (migrations)
3. Configure Docker & Kubernetes
4. Integrate job library (Asynq/Machinery from Spike 2)
5. Set up CI/CD pipeline
6. Team onboarding on 17 key decisions

**Success Criteria**:
- ✅ Project builds without errors
- ✅ Database migrations run successfully
- ✅ Docker containers start
- ✅ Local development environment works
- ✅ Team understands decision document
- ✅ CI/CD pipeline runs and passes

---

## Project Structure

Create the following Go project structure:

```
schedcu-v2/
├── cmd/
│   ├── server/
│   │   └── main.go                 # HTTP server entry point
│   ├── worker/
│   │   └── main.go                 # Job worker entry point
│   └── cli/
│       └── main.go                 # CLI utilities entry point
├── internal/
│   ├── app/                        # Application logic
│   │   ├── schedule_service.go
│   │   ├── coverage_calculator.go
│   │   └── scraper_service.go
│   ├── domain/                     # Domain entities & interfaces
│   │   ├── schedule.go
│   │   ├── coverage.go
│   │   └── scraper.go
│   ├── infrastructure/             # Technical implementations
│   │   ├── database/
│   │   │   ├── repository.go
│   │   │   └── migrations/
│   │   │       └── 001_schema.sql
│   │   ├── vault/
│   │   │   └── client.go
│   │   ├── jobs/
│   │   │   ├── asynq_client.go
│   │   │   └── handlers.go
│   │   └── http/
│   │       ├── router.go
│   │       └── handlers/
│   │           ├── schedule_handler.go
│   │           └── admin_handler.go
│   ├── config/                     # Configuration & initialization
│   │   └── config.go
│   └── middleware/                 # HTTP middleware
│       ├── auth.go
│       ├── logging.go
│       └── metrics.go
├── migrations/                     # Database migrations (golang-migrate)
│   ├── 001_create_schedules_table.up.sql
│   ├── 001_create_schedules_table.down.sql
│   ├── 002_create_coverage_table.up.sql
│   └── 002_create_coverage_table.down.sql
├── test/
│   ├── integration/
│   │   ├── schedule_integration_test.go
│   │   └── coverage_integration_test.go
│   ├── fixtures/
│   │   ├── sample_ods_files/
│   │   └── test_data.sql
│   └── testcontainers/
│       ├── postgres_container.go
│       └── redis_container.go
├── docs/
│   ├── DECISIONS.md                # 17 key architectural decisions
│   ├── API.md                      # API documentation
│   ├── DEPLOYMENT.md               # Deployment procedures
│   └── RUNBOOKS.md                 # On-call runbooks
├── k8s/
│   ├── base/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── configmap.yaml
│   │   └── secret.yaml
│   ├── overlays/
│   │   ├── dev/
│   │   ├── staging/
│   │   └── production/
│   └── monitoring/
│       ├── prometheus.yaml
│       └── alerts.yaml
├── .github/
│   └── workflows/
│       ├── build.yml
│       ├── test.yml
│       ├── security-scan.yml
│       └── deploy.yml
├── docker-compose.yml              # Local development
├── Dockerfile                      # Container image
├── .env.example                    # Environment template
├── go.mod                          # Go module definition
├── go.sum                          # Dependency lock file
├── Makefile                        # Common tasks
├── .pre-commit-config.yaml         # Pre-commit hooks
├── CODE_OF_CONDUCT.md              # Team standards
└── README.md                       # Project overview

Key directories:
- cmd/: Entry points for server, worker, CLI
- internal/: All application code (not exported)
- migrations/: SQL schema changes
- test/: Integration tests & fixtures
- docs/: Architecture & operations
- k8s/: Kubernetes manifests (dev/staging/production)
```

---

## Step 1: Create Go Project (1 hour)

### 1.1 Initialize Go Module

```bash
# Create project directory
mkdir -p ~/projects/schedcu-v2
cd ~/projects/schedcu-v2

# Initialize Go module
go mod init github.com/schedcu/v2

# Create directory structure
mkdir -p cmd/{server,worker,cli}
mkdir -p internal/{app,domain,infrastructure,config,middleware}
mkdir -p internal/infrastructure/{database,vault,jobs,http/handlers}
mkdir -p migrations
mkdir -p test/{integration,fixtures,testcontainers}
mkdir -p docs k8s/.github/workflows

# Create empty files for structure
touch cmd/server/main.go cmd/worker/main.go cmd/cli/main.go
touch internal/app/{schedule_service,coverage_calculator,scraper_service}.go
touch internal/domain/{schedule,coverage,scraper}.go
touch internal/infrastructure/database/repository.go
touch internal/infrastructure/vault/client.go
touch internal/infrastructure/jobs/{asynq_client,handlers}.go
touch internal/infrastructure/http/{router,handlers/schedule_handler,handlers/admin_handler}.go
touch internal/config/config.go
touch internal/middleware/{auth,logging,metrics}.go
touch migrations/001_create_schedules_table.{up,down}.sql
touch migrations/002_create_coverage_table.{up,down}.sql
```

### 1.2 Add Go Dependencies

```bash
# Database
go get github.com/lib/pq
go get github.com/jmoiron/sqlc
go get -u github.com/golang-migrate/migrate/v4

# HTTP framework
go get github.com/labstack/echo/v4

# Job queue (from Spike 2 results)
go get github.com/hibiken/asynq
go get github.com/redis/go-redis/v9

# Configuration
go get github.com/kelseyhightower/envconfig

# Logging
go get go.uber.org/zap

# Metrics
go get github.com/prometheus/client_golang

# Testing
go get github.com/stretchr/testify
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/wait

# Vault
go get github.com/hashicorp/vault/api

# Security
go get golang.org/x/crypto

# Utilities
go get github.com/google/uuid

go mod tidy
```

### 1.3 Create go.mod template

```bash
cat > go.mod <<'EOF'
module github.com/schedcu/v2

go 1.20

require (
	github.com/lib/pq v1.10.9
	github.com/labstack/echo/v4 v4.11.1
	github.com/hibiken/asynq v0.24.1
	github.com/redis/go-redis/v9 v9.0.5
	github.com/kelseyhightower/envconfig v1.4.0
	go.uber.org/zap v1.26.0
	github.com/prometheus/client_golang v1.17.0
	github.com/stretchr/testify v1.8.4
	github.com/testcontainers/testcontainers-go v0.25.0
	github.com/hashicorp/vault/api v1.10.0
	golang.org/x/crypto v0.14.0
	github.com/google/uuid v1.3.0
	github.com/golang-migrate/migrate/v4 v4.16.2
)
EOF

go mod tidy
```

---

## Step 2: Set Up Database & Migrations (1.5 hours)

### 2.1 Create Schema Migration Files

Create `migrations/001_create_schedules_table.up.sql`:
```sql
-- Schedules table (core entity from v1)
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Core schedule data
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    hospital_id UUID NOT NULL,

    -- Parsed from source (Amion, ODS, etc.)
    source TEXT NOT NULL,  -- 'amion', 'ods_file', 'manual'
    source_id TEXT,        -- External reference ID

    -- Schedule assignments
    assignments JSONB NOT NULL DEFAULT '[]'::jsonb,  -- Array of shift assignments

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by UUID NOT NULL,

    -- Soft delete support
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    deleted_by UUID,

    CONSTRAINT schedules_hospital_id_fk FOREIGN KEY (hospital_id)
        REFERENCES hospitals(id) ON DELETE CASCADE
);

-- Indexes for common queries
CREATE INDEX idx_schedules_hospital_date
    ON schedules(hospital_id, start_date, end_date);
CREATE INDEX idx_schedules_source
    ON schedules(source, source_id);
CREATE INDEX idx_schedules_created_at
    ON schedules(created_at DESC);

-- Comments for documentation
COMMENT ON TABLE schedules IS
    'Hospital staff schedules. Each record covers a date range for one hospital.';
COMMENT ON COLUMN schedules.assignments IS
    'JSONB array of shift assignments: [{position: "ER Doc", start_time: "08:00", end_time: "16:00", staff_member: "John Doe", location: "Main ER"}, ...]';
COMMENT ON COLUMN schedules.source IS
    'Schedule source system: amion (web scrape), ods_file (uploaded file), manual (admin entry)';
```

Create `migrations/001_create_schedules_table.down.sql`:
```sql
DROP INDEX IF EXISTS idx_schedules_created_at;
DROP INDEX IF EXISTS idx_schedules_source;
DROP INDEX IF EXISTS idx_schedules_hospital_date;
DROP TABLE IF EXISTS schedules;
```

Create `migrations/002_create_coverage_table.up.sql`:
```sql
-- Coverage calculations (results of DynamicCoverageCalculator)
CREATE TABLE coverage_calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Reference to source schedule
    schedule_id UUID NOT NULL,
    hospital_id UUID NOT NULL,

    -- Calculation inputs
    calculation_date DATE NOT NULL,
    calculation_period_start_date DATE NOT NULL,
    calculation_period_end_date DATE NOT NULL,

    -- Results
    coverage_by_position JSONB NOT NULL,  -- {"position": {"required": 10, "assigned": 8, "coverage": 0.8}}
    coverage_summary JSONB NOT NULL,      -- {"average_coverage": 0.85, "critical_gaps": [...]}

    -- Validation
    validation_errors JSONB DEFAULT '[]'::jsonb,  -- Error collection pattern from Spike 3
    query_count INT DEFAULT 0,            -- For performance assertion testing

    -- Audit
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    calculated_by UUID NOT NULL,

    CONSTRAINT coverage_calculations_schedule_fk FOREIGN KEY (schedule_id)
        REFERENCES schedules(id) ON DELETE CASCADE,
    CONSTRAINT coverage_calculations_hospital_fk FOREIGN KEY (hospital_id)
        REFERENCES hospitals(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_coverage_schedule
    ON coverage_calculations(schedule_id);
CREATE INDEX idx_coverage_hospital_date
    ON coverage_calculations(hospital_id, calculation_date DESC);

COMMENT ON TABLE coverage_calculations IS
    'Results of DynamicCoverageCalculator service. Shows coverage percentage by position.';
COMMENT ON COLUMN coverage_calculations.coverage_by_position IS
    'For each position (e.g., "ER Doctor"): required count, assigned count, coverage percentage.';
```

Create `migrations/002_create_coverage_table.down.sql`:
```sql
DROP INDEX IF EXISTS idx_coverage_hospital_date;
DROP INDEX IF EXISTS idx_coverage_schedule;
DROP TABLE IF EXISTS coverage_calculations;
```

### 2.2 Create Migration Runner

Create `internal/infrastructure/database/migrations.go`:
```go
package database

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

// RunMigrations executes pending database migrations
func RunMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Run pending migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	version, dirty, _ := m.Version()
	log.Printf("Database migration complete. Version: %d, Dirty: %v", version, dirty)

	return nil
}
```

---

## Step 3: Docker & Local Development (1.5 hours)

### 3.1 Create Dockerfile

Create `Dockerfile`:
```dockerfile
# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations
COPY migrations/ migrations/

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/api/health"]

EXPOSE 8080

CMD ["./server"]
```

### 3.2 Create docker-compose.yml

Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: schedcu-postgres
    environment:
      POSTGRES_DB: radiology
      POSTGRES_USER: schedcu_app
      POSTGRES_PASSWORD: dev_password_change_in_production
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U schedcu_app -d radiology"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: schedcu-redis
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes --requirepass redis_password_dev
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Development server (runs locally)
  # Start with: go run ./cmd/server

volumes:
  postgres_data:
  redis_data:
```

### 3.3 Create .env.example

Create `.env.example`:
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=radiology
DB_USER=schedcu_app
DB_PASSWORD=dev_password_change_in_production
DB_SSLMODE=disable

# Redis (for job queue)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password_dev

# Vault (if available)
VAULT_ADDR=https://vault.hospital.local:8200
VAULT_TOKEN=dev-token
VAULT_NAMESPACE=

# Server
SERVER_PORT=8080
SERVER_ENV=development
LOG_LEVEL=debug

# Amion scraper
AMION_USERNAME=your_username
AMION_PASSWORD=your_password

# File uploads
MAX_UPLOAD_SIZE_MB=10
UPLOAD_DIR=./uploads
```

---

## Step 4: Configure CI/CD Pipeline (1 hour)

### 4.1 Create GitHub Actions Workflow

Create `.github/workflows/build.yml`:
```yaml
name: Build and Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  build:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_DB: radiology_test
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_password
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.20'

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
      env:
        DB_HOST: localhost
        DB_USER: test_user
        DB_PASSWORD: test_password
        DB_NAME: radiology_test
        REDIS_HOST: localhost
        REDIS_PORT: 6379

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out

    - name: Build
      run: go build -o server ./cmd/server
```

---

## Step 5: Create DECISIONS.md (30 minutes)

Create `docs/DECISIONS.md` documenting the 17 key architectural decisions:

```markdown
# v2 Architecture Decisions

## Decision 1: Go 1.20+ as implementation language
**Chosen**: Go 1.20+
**Rationale**: Simpler codebase, excellent concurrency, fast startup
**Consequence**: Team upskills on Go; different patterns than Java

## Decision 2: PostgreSQL as primary database
**Chosen**: PostgreSQL 14+
**Rationale**: Proven, JSONB support, PostGIS for location data
**Consequence**: sqlc-generated queries (type-safe)

[... document remaining 15 decisions ...]
```

---

## Step 6: Team Onboarding (1 hour)

### 6.1 Review DECISIONS.md

Team meeting to review and confirm 17 key decisions from `docs/DECISIONS.md`

### 6.2 Local Environment Setup

Each team member:
```bash
# Clone repo
git clone https://github.com/schedcu/v2.git
cd v2

# Copy .env.example
cp .env.example .env

# Start local infrastructure
docker-compose up -d

# Run migrations
go run ./cmd/server migrate

# Build & test
go build ./cmd/server
go test ./...

# Start server locally
go run ./cmd/server
# Server should start on :8080
```

---

## Checklist: Phase 0 Complete

- [ ] Go project structure created
- [ ] All Go dependencies installed
- [ ] Database migrations run successfully
- [ ] Docker images build without errors
- [ ] docker-compose starts all services
- [ ] Local server starts and health check passes
- [ ] All tests pass (go test ./...)
- [ ] CI/CD pipeline configured
- [ ] DECISIONS.md reviewed and confirmed by team
- [ ] Team can run local environment
- [ ] Pre-commit hooks configured

---

## Next: Phase 1 Begins

Once Phase 0 complete:
1. Backend Engineer starts entities & repositories (sqlc)
2. Test/DevOps Engineer sets up integration test infrastructure
3. Team begins Phase 1: Core Services implementation

---

## Quick Start Checklist

After Phase 0 setup, new team members should:

```bash
# 1. Clone and setup
git clone https://github.com/schedcu/v2.git && cd v2
cp .env.example .env

# 2. Start local infrastructure
docker-compose up -d

# 3. Wait for services to be healthy
docker ps  # all services should be "healthy"

# 4. Run server
go run ./cmd/server

# 5. Verify
curl http://localhost:8080/api/health

# Expected response:
# {"status":"UP","components":{"database":"UP","redis":"UP"}}
```
