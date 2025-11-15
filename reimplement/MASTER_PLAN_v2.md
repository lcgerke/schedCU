# Hospital Radiology Schedule System v2: Master Plan (v2)

**Status**: Ready for implementation with dependency validation
**Date**: 2025-11-15 (Updated with Risk Validation Strategy)
**Approach**: Go rewrite preserving proven v1 patterns + early validation spikes
**Timeline**: 3 months to production (+ 1 week v1 security fixes in parallel + 1 week dependency validation)
**Team**: 3 people (backend, API/security, testing/DevOps)

---

## Executive Summary

You are rewriting the Hospital Radiology Schedule System from Java/Quarkus to Go. This is a **translation of proven patterns to a more maintainable stack**, incorporating all learnings from v1 (documented in reimplement/).

**Key principles:**
- **Preserve** proven patterns from v1 (ValidationResult, ScrapeBatch, DynamicCoverageCalculator, entity relationships)
- **Translate** domain model to Go idioms (not redesign from scratch)
- **Fix** architectural issues (N+1 queries, security bypass, explicit over implicit)
- **Leverage** Go's strengths (concurrency, simplicity, explicit code, fast startup)
- **Use** existing hospital infrastructure (PostgreSQL, Vault, Docker)
- **Maintain** historical data for compliance (HIPAA, audit trail)
- **Validate** critical assumptions early before committing resources

**Expected improvements over v1:**
- 10× faster Amion scraping (sequential 180s → parallel 2-3s via goroutines) — **subject to Amion HTML validation**
- 50% smaller memory footprint (Go binary vs JVM)
- Better security (refresh tokens, Vault integration, rate limiting)
- **Easier to maintain** (explicit queries, no ORM magic, simpler codebase)
- Production-ready monitoring from day one
- Zero technical debt (no deprecated fields, clean patterns)

**Critical path difference from typical rewrite:**
- **Week 0**: Dependency validation spikes (de-risk major assumptions)
- Week 1: **Parallel track** fixes v1 security issues → v1 production-ready
- Weeks 2-14: v2 development proceeds without urgency
- No pressure if v2 takes longer; v1 is secure and operational

---

## NEW: Week 0 — Dependency Validation Spikes

**Goal**: Eliminate critical unknowns before Phase 0 begins. Reduce risk from Medium-High to Low-Medium.

### Three Critical Spikes (3-5 days total)

**Spike 1: Amion HTML Parsing Feasibility** (2 days)
- **Risk**: Amion may use JavaScript-heavy rendering, invalidating the 10× performance claim
- **Task**:
  - Download current Amion schedule page with HTTP client (no browser automation)
  - Test goquery CSS selectors against actual HTML
  - Measure parsing time with 6-month historical data simulation
  - Document CSS selector paths and stability
- **Success criteria**:
  - Goquery can extract shift data with >95% accuracy OR
  - Clear documentation of fallback to Chromedp (with implementation timeline)
- **Failure path**: If HTML is JavaScript-heavy, pivot to Chromedp (adds 2 weeks to Phase 3, but known cost)

**Spike 2: Asynq + Job Library Evaluation** (2 days)
- **Risk**: Redis may not be available in hospital infrastructure; job result storage complexity
- **Task**:
  - Verify Redis availability/status in hospital infrastructure
  - Evaluate both Asynq (Redis-based) and Machinery (PostgreSQL-based)
  - Test sample job execution with both libraries
  - Document production readiness of each
- **Success criteria**:
  - Redis is available + Asynq works OR
  - PostgreSQL-based Machinery verified + documented as fallback
- **Failure path**: If neither works, timeline extends (custom job queue adds 3 weeks)

**Spike 3: ODS Library & Performance** (1-2 days)
- **Risk**: ODS library may not support error collection pattern; performance with large files unknown
- **Task**:
  - Evaluate `github.com/knieriem/odf` (or alternative) with production-size ODS files
  - Test error collection pattern (collect all errors, not fail-fast)
  - Measure parse time and memory usage
  - Document file size limits
- **Success criteria**:
  - Library supports error collection OR
  - Custom wrapper layer designed to add this capability
- **Failure path**: If no viable library, build lightweight ODS parser (adds 1-2 weeks)

### Deliverables from Week 0
- [ ] Amion HTML parsing validated (goquery works OR Chromedp fallback documented with cost)
- [ ] Job library selected with rationale (Asynq OR Machinery OR custom)
- [ ] ODS library validated with error collection pattern
- [ ] **Risk level downgraded from Medium-High to Low-Medium**
- [ ] Any fallback timelines documented in revised Phase estimates

### Timeline Impact
- **Best case**: All spikes succeed, no timeline impact (Week 0 overhead absorbed)
- **Moderate case**: 1 spike fails, fallback adds 1-2 weeks to subsequent phase
- **Worst case**: Multiple failures → timeline extends to 4-5 months (v1 still operational, no urgency)

---

## 17 Critical Decisions (Locked In, Post-Validation)

### Architecture & Data Layer

**Decision 1: Data Access Strategy**
- **Choice**: SQL-first with sqlc
- **Why**: Type-safe queries, explicit SQL, prevents N+1 from v1
- **Implication**: All queries visible, auditable, no hidden joins
- **Key**: Query count assertions in tests prevent regression

**Decision 2: Entity Model Strategy** ✅ UPDATED
- **Choice**: Evolve v1 schema to Go idioms (not redesign from scratch)
- **Why**: v1 schema encodes proven domain knowledge; preserve this, fix issues
- **Implementation**:
  - **Port** v1 entities to Go structs (ScheduleVersion, ShiftInstance, ScrapeBatch, Assignment, Person, AuditLog)
  - **Remove** deprecated fields (reassigned_shift_type)
  - **Add** missing constraints (foreign keys, check constraints)
  - **Keep** proven patterns (soft delete, temporal validity, audit trail, batch lifecycle)
  - **Design for sqlc** (explicit queries, no ORM magic)
- **Phase 0 critical task**: Study v1 schema deeply, document WHY each field exists before porting
- **Schema includes**: Same 20-30 tables as v1, cleaned up and constrained properly
- **Key learnings applied**: Foreign key constraints, indexes on date/shift queries, proper soft delete pattern

**Decision 3: Service Layer Patterns**
- **Choice**: Replicate v1 patterns in Go idioms
- **Why**: ValidationResult, ScrapeBatch, DynamicCoverageCalculator are proven and elegant
- **Services to replicate**:
  - ValidationResult (struct with methods, JSON marshaling, same severity levels)
  - ScrapeBatch (state machine: PENDING → COMPLETE/FAILED, exact same lifecycle)
  - DynamicCoverageCalculator (pure functions, batch queries, no N+1)
  - ScheduleOrchestrator (3-phase workflow: ODS → Amion → coverage)
  - ODSImportService (with validation, collect all errors)
  - AmionImportService (with batch lifecycle)
  - CoverageResolutionService (explicit, non-legacy version)

**Decision 4: Concurrency & Long-Running Tasks** ✅ UPDATED
- **Choice**: Asynq job library with Redis backend (or Machinery if Redis unavailable)
- **Why**: Battle-tested, built-in retry/monitoring, don't build custom job infrastructure
- **Implementation** (post-Spike 2):
  - If Redis available: Use `github.com/hibiken/asynq` for async job processing
  - If Redis unavailable: Use `github.com/RichardKnop/machinery` with PostgreSQL broker
  - Job types: ODS import, Amion scraping, coverage resolution
  - Long-running tasks return job ID immediately
  - Job monitoring dashboard (Asynq web UI or equivalent)
  - Job results stored in PostgreSQL, queued in Redis (or Machinery's PostgreSQL queue)
  - Built-in retry, timeout, priority, scheduled tasks
- **Performance gain**: 6 months of Amion data: 180s → 2-3s (parallel goroutines) — **pending Spike 1 validation**
- **Why not custom**: Less code to maintain, proven patterns, monitoring included

### API & Security

**Decision 5: REST Framework**
- **Choice**: Echo (Go HTTP framework)
- **Why**: Lightweight, Go-idiomatic, excellent middleware support
- **Middleware**: Auth, logging, request ID tracking, CORS, rate limiting

**Decision 6: Error Response Format**
- **Choice**: Unified ApiResponse with embedded ValidationResult
- **Why**: Keeps v1's rich error semantics, clean HTTP contract
- **Structure**:
  ```json
  {
    "data": {...},
    "validation": {ValidationResult},
    "error": {...},
    "meta": {timestamp, requestId, version}
  }
  ```
- **Key**: Validation errors return full ValidationResult (code, severity, context), not just HTTP status

**Decision 7: Authentication & Authorization** ✅ UPDATED
- **Choice**: JWT (access + refresh tokens) + Vault for secrets
- **Security improvements over v1**:
  - Access token: 15 min (short-lived, theft window small)
  - Refresh token: 7 days (long-lived, HTTP-only cookie)
  - Rate limiting: 5 login attempts, 15 min lockout
  - Vault stores JWT signing keys (no keys in code)
  - Immediate logout via refresh token revocation
- **Implementation**: Manual refresh token rotation, Vault client at startup
- **NEW**: On-call playbook for token-related issues (what to do if Vault is down, how to handle token rotation failures)

**Decision 8: Deployment**
- **Choice**: Docker containers
- **Why**: Go binary ~50MB, fast startup, easy Vault integration, can scale later
- **Docker Compose** for local dev, Docker for production
- **Startup**: golang-migrate runs DB migrations, then app starts

### Code Organization & Testing

**Decision 9: Dependency Injection**
- **Choice**: Manual constructor injection (idiomatic Go)
- **Why**: Explicit (no magic), testable, Go-standard, works with Vault injection
- **Implementation**: Services accept dependencies in constructors, wired in main.go
- **Testing**: Mock specific dependencies needed

**Decision 10: Project Structure**
- **Choice**: Layered by responsibility
- **Layout**:
  ```
  schedjas/
  ├── internal/
  │   ├── entity/          # Domain types (translated from v1)
  │   ├── repository/      # sqlc-generated DB queries
  │   ├── service/         # Business logic (v1 patterns in Go)
  │   ├── api/             # Echo handlers
  │   ├── job/             # Asynq/Machinery job handlers
  │   └── validation/      # ValidationResult, error types
  ├── cmd/
  │   └── schedjas/        # Main entrypoint
  ├── migrations/
  │   └── *.sql            # golang-migrate database migrations
  ├── docs/
  │   ├── schema/          # v1 → v2 schema mapping documentation
  │   ├── runbooks/        # On-call playbooks
  │   └── spikes/          # Week 0 spike results
  └── tests/
      └── fixtures/        # Test data, sample ODS files
  ```
- **Why**: Clear separation, prevents circular imports, easy navigation

**Decision 11: Testing Strategy**
- **Choice**: Layered tests (unit + integration + E2E, mirroring structure)
- **Target coverage**: 85% (60% unit, 30% integration, 10% E2E)
- **Test locations**: `*_test.go` next to code (Go standard)
- **Critical tests**:
  - Repository tests: **Query count assertions** (prevent N+1 regressions)
  - Service tests: Business logic with mocked DB
  - API tests: HTTP contracts, auth, error formats
  - Job tests: Job execution (Asynq or Machinery)
  - **Amion scraper tests**: Mock HTML responses from Spike 1 results
- **Integration**: Testcontainers for real PostgreSQL in tests
- **Regression prevention**: Tests enforce patterns (no N+1, all endpoints protected, ValidationResult usage)

**Decision 12: Database Migrations**
- **Choice**: golang-migrate
- **Why**: Go-native, lightweight, industry standard, versioned in Git
- **Format**: `migrations/001_*.up.sql` and `.down.sql`
- **Execution**: Automatic on Docker startup
- **Source**: Based on v1 schema (evolved, not redesigned)
- **NEW**: Test all migrations in CI using Testcontainers (catch failures before production)

### Data & Performance

**Decision 13: ODS File Parsing** ✅ UPDATED
- **Choice**: Go ODS library (validated in Spike 3)
- **Why**: Focus on business logic, not XML parsing
- **Error handling**: Validation errors collected (not fail-fast), same as v1
- **Performance**: Spike 3 determines file size limits and parse time

**Decision 14: Amion Web Scraping** ✅ UPDATED
- **Choice**: Simple HTTP client + goquery (CSS selectors) — **pending Spike 1 validation**
- **Why**: If Amion serves HTML directly, 10× faster than browser automation
- **Performance**: 6 months parallel (5 concurrent goroutines) = 2-3 seconds vs v1's 180s — **subject to Spike 1 success**
- **Fallback**: If Amion uses heavy JavaScript, switch to headless browser (Chromedp) with documented timeline and cost
- **Goroutines**: 5 concurrent scrapers with rate limiting (1 sec between requests)
- **NEW**: Fallback plan explicitly documented with 2-week timeline cost if goquery approach fails

**Decision 15: Data Continuity Strategy** ✅ UPDATED
- **Choice**: Fresh operational data + historical audit preservation
- **Why**: HIPAA compliance, dispute resolution, no migration complexity for operational data
- **Infrastructure clarification needed** (from hospital stakeholders):
  - Do we have S3/cold storage for audit log archival?
  - If not, use PostgreSQL-based archive (same machine, separate schema)
  - Clarify retention: 6-year HIPAA compliance for audit logs
- **Implementation**:
  - **v2 operational data**: Import current ODS file + last 30 days of Amion schedules (clean start)
  - **v1 historical data**: Keep v1 running in read-only mode for 1 year
  - **Audit logs**: Export v1 audit logs to archive (S3 OR PostgreSQL archive) before v1 shutdown
  - **Historical queries**: Optional "historical schedule" API endpoint (if needed)
  - **Retention**: After 1 year, export v1 database to cold storage, decommission v1
- **Cutover**: Parallel run for 1 week (v1 read-only, v2 operational, monitor both)
- **Compliance**: 6-year HIPAA audit trail maintained via v1 archive + v2 ongoing logs
- **Decommission timeline**: v1 kept for 1 year, then archived to cold storage

**Decision 16: Audit Logging**
- **Choice**: Keep with retention policy (same as v1)
- **Why**: Hospital compliance (HIPAA), debugging, accountability
- **Implementation**:
  - AuditLog table: user, action, resource, timestamp, old/new values
  - All admin actions logged
  - 1-year retention policy in v2
  - Structured JSON logs to stdout
  - v1 audit logs archived for 6-year HIPAA compliance
- **Compliance**: Queryable history for audits
- **NEW**: On-call runbook for audit log analysis (how to query for compliance investigations)

**Decision 17: Observability** ✅ UPDATED
- **Choice**: Logs + Metrics (Prometheus-compatible)
- **Why**: Industry standard, Go ecosystem excellent, cost-effective
- **Logs**: Structured (JSON), go to stdout, Docker/K8s aggregates
- **Metrics**: Prometheus format on `/metrics` endpoint
- **Key metrics**:
  - Schedule import duration
  - Amion scrape duration
  - Coverage resolution duration
  - Validation error counts
  - **Query counts per request** (alert if >10, catches N+1 regressions)
  - **Method execution time by name** (catches long-running operations)
  - **Endpoint auth failures** (catches security bypass attempts)
  - **Job retry counts** (catches scraping/parsing failures)
  - API response times by endpoint
- **Regression prevention**: CI fails if query count increases for same endpoint
- **NEW**: Alert thresholds and on-call escalation procedures documented
- **Phase 2**: Add Prometheus scraper + Grafana dashboards if needed

---

## CRITICAL: v1 Security Fixes (Parallel Track)

### Why This Matters

v1's security issues **cannot wait 14 weeks** for v2 to ship. This parallel track ensures v1 is production-ready while v2 develops.

### Week 1 Tasks (Parallel to v2 Phase 0)

**Team allocation**:
- **1 person**: v1 security fixes (Week 1 only)
- **2 people**: v2 Phase 0 (schema study, setup) — OR Week 0 Dependency Validation if proceeding directly

**Security fixes** (6-8 hours total):

1. **Remove @PermitAll bypass** (2 hours)
   - Remove `@PermitAll` from all admin endpoints
   - Add `@RolesAllowed("ADMIN")` to AdminResource methods
   - Test all admin endpoints require authentication

2. **Move credentials to Vault/env vars** (2 hours)
   - Database password → `POSTGRES_PASSWORD` env var
   - SMTP credentials → Vault or env vars
   - JWT signing keys → Vault

3. **Add file upload validation** (2 hours)
   - Max file size: 10MB
   - Allowed MIME types: ODS only
   - XXE protection: Disable external entities in XML parser

4. **Test and deploy** (2 hours)
   - Security test suite (all endpoints require auth)
   - Deploy to production
   - Monitor for issues

**Result**: After Week 1, v1 is secure and production-ready. v2 can proceed at normal pace without urgency.

---

## Implementation Timeline (Updated)

### Week 0: Dependency Validation (NEW — 3-5 days)

**Goal**: De-risk critical assumptions before committing full team to Phase 0

**Spike Team** (can be 1-2 people):
- **Spike 1**: Amion HTML parsing (Backend Engineer or Test Engineer)
- **Spike 2**: Job library evaluation (Test/DevOps Engineer)
- **Spike 3**: ODS library (Backend Engineer)

**Parallel**: Management/PM confirms infrastructure details (S3 availability, Vault readiness, Redis status)

**Deliverables**:
- [ ] Amion HTML parsing validation (goquery feasibility OR Chromedp fallback with cost)
- [ ] Job library selection (Asynq OR Machinery, documented)
- [ ] ODS library validated
- [ ] Infrastructure clarifications (S3, Redis, Vault status)
- [ ] Revised timeline (if any spikes failed, adjust Phases 1-4 accordingly)

---

### Phase 0: Pre-Development (1 week + Week 0 spikes)

**Goal**: Translate v1 schema to Go, prepare infrastructure, fix v1 security

**Day 1: v1 Security Fixes (Parallel Track)**
- **Person**: API & Security Engineer
- Remove `@PermitAll` from admin endpoints (2 hours)
- Move credentials to Vault/env vars (2 hours)
- Add file upload validation (2 hours)
- Security test suite (2 hours)
- **Result**: v1 production-ready by end of Day 1

**Day 1-2: Schema Translation (Not Redesign)**
- **People**: Backend Engineer + Test/DevOps Engineer
- Export v1 schema to SQL (analyze with `pg_dump --schema-only`)
- Create v1 → v2 entity mapping document:
  - Each v1 table → Go struct
  - Document **WHY** each field exists (reference v1 code/docs)
  - Identify deprecated fields to remove (`reassigned_shift_type`)
  - Identify missing constraints to add
- Design indexes for query patterns (from v1 analysis)
- **Critical**: Don't redesign, translate + fix issues

**Day 3: Audit Log Export & Data Continuity**
- **Person**: Backend Engineer
- Write script: v1 `audit_log` table → JSON export (or PostgreSQL archive if S3 unavailable)
- Test export process with production data
- Document HIPAA retention policy with stakeholders
- Plan v1 read-only mode (connection string change + UI banner)

**Day 4: Job Library Setup (Post-Spike 2)**
- **Person**: Test/DevOps Engineer
- Integrate selected job library (Asynq OR Machinery) in sample Go project
- Test sample scraping job with selected backend
- Verify compatibility with hospital infrastructure
- Document job patterns for team

**Day 5-7: Project Setup & Team Preparation**
- Create Go module structure (`internal/`, `cmd/`, `migrations/`)
- Set up sqlc configuration
- Create golang-migrate migration files (from schema translation)
- Create Dockerfile and docker-compose.yml
- Set up Git hooks (prevent security bypasses, e.g., check for unprotected endpoints)
- Read `reimplement/` documentation (all team members)
- Review all 17 decisions
- Set up development environment
- Create project charter (signed)
- **NEW**: Set up spike results documentation in `docs/spikes/`

**Deliverables**:
- [x] **v1 security patched and deployed to production** (Day 1)
- [ ] PostgreSQL schema (evolved from v1, not redesigned)
- [ ] v1 entity → Go struct mapping document (with rationale for each field)
- [ ] Audit log export script tested
- [ ] Job library integrated and tested (Asynq OR Machinery)
- [ ] golang-migrate migrations (20-30 tables from v1)
- [ ] Go project structure
- [ ] Dockerfile and docker-compose.yml
- [ ] Query count assertion framework in test infrastructure
- [ ] Development environment working locally
- [ ] Pre-commit hooks preventing security bypasses
- [ ] **Spike results documented** (Amion HTML parsing, job library, ODS library)

---

### Phase 1: Core Services & Database (4 weeks)

**Goal**: Implement entity layer, database access, core business logic

**Week 1: Entities & Repositories**
- Create entity types in `internal/entity/` (translated from v1):
  - ScheduleVersion, ShiftInstance, Assignment, Person
  - ScrapeBatch, ScrapedSchedule
  - User, AuditLog
- Write SQL queries in `internal/repository/*.sql`
- Generate sqlc code from SQL
- Create repository interfaces (optional but helpful for testing)
- Begin unit tests for entity validation

**Week 2: Validation Framework**
- Implement ValidationResult struct (port from v1 Java)
- Create ValidationMessage type with severity levels (ERROR, WARNING, INFO)
- JSON marshaling/unmarshaling
- Integration tests for validation logic
- Tests verify v1 patterns work in Go (same semantics)

**Week 3: Core Service Layer**
- Implement ODS import service:
  - Parse ODS files using validated library (from Spike 3)
  - Validate shifts against known types
  - Create ShiftInstances
  - Return ValidationResult (collect all errors, don't fail fast)
- Implement dynamic coverage calculator:
  - Batch query optimization (load all assignments for date range in 1-2 queries)
  - Test with query count assertions (assert exactly N queries, no more)
- Implement schedule orchestrator (3-phase workflow: ODS → Amion → coverage)

**Week 4: Database Integration**
- Implement all repository methods using sqlc
- Integration tests with Testcontainers (spin up PostgreSQL)
- Query performance tests:
  - Verify indexes work (EXPLAIN ANALYZE)
  - Assert query counts (no N+1)
  - Assert response time SLAs
- Data integrity tests (foreign keys enforce constraints, soft delete works)

**Deliverables**:
- 80% of core services working
- ValidationResult passing all tests (same behavior as v1 Java version)
- DynamicCoverageCalculator with performance tests (query count assertions)
- Query count assertions preventing N+1 regressions
- Integration tests running with real PostgreSQL

---

### Phase 2: API Layer & Security (3 weeks)

**Goal**: Implement REST endpoints, security, job system

**Week 1: Authentication & Job System**
- Implement JWT token generation/validation
- Set up Vault client (fetch secrets at startup)
- Implement refresh token rotation
- Integrate selected job library (Asynq OR Machinery) for job processing:
  - Job handlers for ODS import, Amion scraping
  - Job status monitoring/dashboard
  - Job results stored in PostgreSQL
- Rate limiting middleware (5 login attempts, 15 min lockout)

**Week 2: REST API Endpoints**
- Implement all endpoints with Echo
- Error handler for unified ApiResponse format
- Auth middleware on protected endpoints (NO unprotected admin endpoints)
- Request ID middleware (correlation logging)
- Implement async endpoints (ODS import, Amion scraping):
  - Return job ID immediately
  - Job runs in background
  - Poll `/api/jobs/{id}` for status
- Tests for every endpoint:
  - Happy path
  - Auth failures (401)
  - Authorization failures (403)
  - Validation failures (400 with ValidationResult)
  - Error responses

**Week 3: Security Hardening**
- File upload validation (size, type, XXE protection)
- Input sanitization
- CORS configuration
- Rate limiting on sensitive endpoints
- Security tests for all admin endpoints (100% coverage)
- Audit logging on sensitive operations
- Pre-commit hook enforcement (CI fails if unprotected endpoints detected)
- CI/CD security gates (no bypasses, no hardcoded secrets)

**Deliverables**:
- All endpoints working with clean error handling
- Security tests passing (100% endpoint coverage, NO unprotected admin endpoints)
- Job system operational (no custom job code)
- Vault integration proven
- No security bypass patterns possible (enforced by tests + pre-commit hooks)

---

### Phase 3: Scrapers & Integrations (2.5 weeks, adjusted per Spike 1)

**Goal**: Implement Amion scraper, ODS file handling, data flow

**Week 1: Amion Scraper** (timeline adjusted based on Spike 1)
- **If goquery works** (Spike 1 success):
  - HTTP client + goquery for HTML parsing
  - Parallel scraping (5 concurrent months via goroutines)
  - Rate limiting (1 sec between requests)
  - HTML parsing error handling (ValidationResult)
  - Tests with mock HTML responses from Spike 1 results
  - Performance: verify 6 months = 2-3 seconds (vs v1's 180s)
- **If goquery fails** (Spike 1 reveals JavaScript requirement):
  - Add 2 weeks to timeline
  - Implement headless browser (Chromedp)
  - Same rate limiting and error handling
  - Performance target same: 2-3 seconds for 6 months

**Week 2: ODS File Handling**
- File upload endpoint (with validation: size, type, XXE protection)
- ODS parsing with validated library (from Spike 3)
- Shift type validation against known types
- Error collection (ValidationResult with all issues, not fail-fast)
- Tests with sample ODS files from production
- XXE attack prevention verified (disable external entities)

**Week 3 (partial): Integration & Async Jobs**
- ODS import as job (Asynq or Machinery)
- Amion scraping as job
- Job status tracking via dashboard
- Job result storage and retrieval (PostgreSQL)
- Coverage resolution triggered after Amion import completes
- Full workflow tests (ODS → Amion → coverage, end-to-end)
- Performance benchmarks (document 10× improvement — or actual improvement if different)

**Deliverables**:
- Working Amion scraper with documented performance
- ODS import with file validation (XXE protected)
- Job system fully operational
- Complete data import workflow tested
- Performance improvements documented and measured

---

### Phase 4: Testing, Monitoring, Polish (2 weeks)

**Goal**: 85% test coverage, observability, production readiness

**Week 1: Test Coverage & Monitoring**
- Add missing unit tests (target 85% coverage)
- Integration tests for workflows
- E2E tests for critical paths (ODS → Amion → coverage)
- Structured logging implementation (JSON to stdout)
- Prometheus metrics instrumentation:
  - Query count per request (alert if >10)
  - Method execution time (alert if >1s)
  - Job retry counts
  - Endpoint auth failure counts
- Tests for monitoring (verify metrics exported correctly)

**Week 2: Documentation & Production Readiness**
- Code documentation (why, not just what)
- API documentation (OpenAPI/Swagger spec)
- **NEW**: On-call runbooks:
  - How to analyze audit logs for compliance
  - How to handle token-related issues
  - How to troubleshoot Amion scraping failures
  - How to monitor job queue health
  - Escalation procedures for monitoring alerts
- Troubleshooting guide
- Schema documentation (v1 → v2 mapping, why each field exists)
- Security hardening verification (penetration test checklist)
- Load testing (does it scale to 100 concurrent users?)
- Cutover plan refinement (step-by-step procedure)

**Deliverables**:
- 85%+ test coverage
- Production monitoring configured (Prometheus metrics exported)
- Complete documentation (runbooks, API docs, schema docs)
- On-call runbooks (audit analysis, token issues, troubleshooting)
- Cutover plan finalized (dry-run tested)
- v2 ready for staging/UAT

---

### Cutover Week (1 week)

**Timeline**:
- **Monday**: Deploy v2 to staging, run validation (import test ODS, scrape test Amion)
- **Tuesday-Wednesday**: User acceptance testing (radiologists test workflows)
- **Thursday**: Final data import (current ODS + last 30 days Amion schedules)
- **Friday**: Cutover during off-hours
  - Set v1 to read-only mode (change DB connection to read replica, add UI banner)
  - Switch traffic to v2 (update DNS/load balancer)
  - Monitor for errors (logs, metrics, user feedback)
- **Following week**: Parallel operation
  - v2 handles all operations
  - v1 kept running for historical data queries
  - Monitor v2 metrics, ready to rollback if needed

**Rollback plan**:
- Switch traffic back to v1 (DNS change)
- v1 exits read-only mode
- Investigate v2 issues
- Re-attempt cutover when fixed

**Decommission timeline**:
- **Week 2-52**: v1 runs in read-only mode for historical queries
- **Week 52**: Export v1 audit logs to cold storage (S3 or PostgreSQL archive)
- **Week 53**: Shut down v1, archive database to cold storage

---

## Team & Roles

**3-person team, 3-4 months** (14-16 weeks including Week 0 validation):

1. **Backend Engineer** (40% services, 40% database, 20% integration)
   - Core services (ODS, Amion, coverage resolution)
   - Database schema translation and migrations
   - Job integration
   - Performance optimization
   - **Week 0**: Lead Spike 1 (Amion HTML parsing) + contribute to Spike 3
   - **Week 1**: Schema translation (v1 → v2 mapping)

2. **API & Security Engineer** (50% API, 30% security, 20% integration)
   - REST endpoints (Echo handlers)
   - Authentication/authorization (JWT, Vault)
   - File upload validation
   - Security testing
   - Error handling
   - **Week 0**: Contribute to infrastructure clarification
   - **Week 1**: v1 security fixes (parallel track)

3. **Test & DevOps Engineer** (70% testing, 30% DevOps)
   - Unit/integration/E2E test suites
   - Test infrastructure (Testcontainers)
   - Docker/deployment setup
   - Monitoring and observability (Prometheus)
   - CI/CD pipeline
   - **Week 0**: Lead Spike 2 (Job library evaluation)
   - **Week 1**: Asynq/Machinery integration and testing

**Skill requirements**:
- All: Go programming (intermediate+)
- All: PostgreSQL basics
- Backend: ODS parsing, data migration, algorithms
- API: REST design, security (auth/encryption)
- Test: Testing strategies, Docker, CI/CD, Prometheus

**Team skill validation** (pre-Week 0):
- Has the team worked with Go on production systems?
- Does anyone have experience with Vault integration?
- Is there PostgreSQL expertise available for performance tuning?
- Do we have on-call staff trained in Go debugging?

---

## Success Metrics & Go/No-Go Gates

### Week 0 Completion (Dependency Validation)
- [ ] Amion HTML parsing validated (goquery works OR Chromedp fallback documented with timeline cost)
- [ ] Job library selected (Asynq OR Machinery documented and tested)
- [ ] ODS library validated with error collection capability
- [ ] Infrastructure clarifications documented (S3, Redis, Vault status)
- [ ] Team confidence in approach increased (risk level downgraded to Low-Medium)
- **Go/No-Go**: Proceed if all spikes succeed or fallbacks are acceptable. If >2 spikes fail, pause and reassess (v1 still operational, no rush).

### Phase 0 Completion
- [x] **v1 security patched and deployed** (CRITICAL)
- [ ] Schema translated (v1 → v2 mapping document complete)
- [ ] golang-migrate structure working
- [ ] Job library integrated
- [ ] Docker environment runs locally
- [ ] Team trained on decisions
- [ ] Spike results documented
- **Go/No-Go**: Proceed if v1 is secure AND schema translation complete AND spike results incorporated

### Phase 1 Completion
- [ ] ValidationResult tests passing (same semantics as v1 Java)
- [ ] DynamicCoverageCalculator passing performance tests (query count assertions)
- [ ] Integration tests with Testcontainers working
- [ ] Core services 80% feature-complete
- **Go/No-Go**: Proceed if services functional, no performance regressions

### Phase 2 Completion
- [ ] All endpoints tested (security tests passing)
- [ ] JWT + refresh tokens working
- [ ] Vault integration proven
- [ ] Job system operational
- [ ] **No unprotected admin endpoints** (enforced by tests)
- **Go/No-Go**: Proceed if security audit passes

### Phase 3 Completion
- [ ] Amion scraper working, performance documented (2-3s for 6 months OR actual performance from implementation)
- [ ] ODS import with validation (XXE protected)
- [ ] Full workflow tested (ODS → Amion → coverage)
- [ ] Performance benchmarks documented
- **Go/No-Go**: Proceed if performance goals met or fallback performance is acceptable

### Phase 4 Completion
- [ ] Test coverage 85%+
- [ ] Monitoring/observability working (Prometheus metrics exported)
- [ ] Documentation complete (runbooks including on-call procedures)
- [ ] Cutover plan finalized and dry-run tested
- [ ] Load testing passed (100 concurrent users)
- **Go/No-Go**: Production ready

### Production Cutover
- [ ] Data validation passed (ODS/Amion import verified)
- [ ] v2 in staging for 48 hours, no errors
- [ ] Rollback procedure tested
- [ ] v1 in read-only mode (not shut down)
- **Go/No-Go**: Switch users to v2

---

## Risk Mitigation (Updated)

### High-Risk Items & Mitigation

| Risk | Mitigation |
|------|-----------|
| **Amion HTML parsing not feasible** | **MITIGATED**: Spike 1 validates goquery OR documents Chromedp fallback (+2 weeks) |
| **Job library not available** | **MITIGATED**: Spike 2 validates Asynq OR Machinery OR documents custom solution (+3 weeks) |
| **ODS library doesn't support error collection** | **MITIGATED**: Spike 3 validates OR documents wrapper layer (+1 week) |
| **v2 takes longer than 14 weeks** | v1 is secure and operational (Week 1 fixes), no pressure |
| **Business requirements change mid-rewrite** | v1 can accept changes, v2 incorporates when ready |
| **Schema redesign loses domain knowledge** | **MITIGATED**: Schema translated, not redesigned (preserve v1 patterns) |
| **Historical data needed for compliance** | **MITIGATED**: v1 kept for 1 year, audit logs archived 6 years |
| **Infrastructure doesn't support approach** | **MITIGATED**: Week 0 confirms Redis, Vault, S3 (or PostgreSQL archive) availability |
| Amion HTML format changes | Keep Chromedp approach ready (fallback from Spike 1) |
| Performance doesn't meet goals | Parallel scraping fallback, database optimization phase |
| Security breach during development | Code review checklist, pre-commit hooks, security tests |
| Data corruption during cutover | Validation scripts, dry-run migration, rollback to v1 |
| Team learning curve (Go) | Pair programming on critical paths, clear code reviews |
| Vault integration issues | Fallback to env vars (Phase 1), migrate to Vault (Phase 2) |
| Database migration failures | Testcontainers test all migrations before production |
| On-call team unprepared | **NEW**: Runbooks written during Phase 4 with team walkthrough |

### Quality Gates (Non-Negotiable)

```
CANNOT proceed to Phase 0 without:
✓ Week 0 spikes completed (Amion parsing, job library, ODS library validated)
✓ Infrastructure clarifications confirmed

CANNOT deploy v2 without:
✓ v1 security fixes deployed (Week 1)
✓ All security tests passing
✓ No unprotected admin endpoints (enforced by tests + pre-commit hooks)
✓ All credentials in Vault/env vars (no hardcoded)
✓ File upload validation implemented (XXE protection)
✓ Rate limiting on login endpoint
✓ Audit logging functional
✓ Test coverage 85%+
✓ No N+1 query patterns (query count assertions pass)
✓ Monitoring/observability working (Prometheus metrics)
✓ Documentation complete (runbooks, API docs, schema mapping)
✓ On-call runbooks tested with team
✓ v1 historical data preservation plan executed
```

---

## Key Learnings Applied from v1

See `reimplement/` for full analysis:

**From 02-WHAT-WORKED.md** (patterns to keep):
- ValidationResult with severity levels → **Exact same in v2 (translated to Go)**
- ScrapeBatch lifecycle → **Exact same in v2 (same states, same transitions)**
- ScheduleVersion temporal versioning → **Exact same in v2**
- Person registry YAML sync → **Same pattern in v2**
- Soft delete strategy → **Same approach in v2**
- E2E testing → **Keep for critical workflows**
- **Entity relationships** → **Preserve in v2 (proven domain model)**

**From 01-TECHNICAL-DEBT.md** (mistakes to prevent):
- N+1 queries → **Query count assertions in tests (CI fails if regression)**
- Admin endpoint bypass → **Security tests for all endpoints + pre-commit hooks**
- Hardcoded credentials → **Vault from day 1**
- No file upload validation → **Implemented Phase 2 (XXE protection)**
- Long methods → **Code review enforces <30 LOC/method**
- Magic strings → **Go constants (enums), no string literals**

**From 03-SECURITY-GAPS.md** (security improvements):
- JWT now has refresh tokens (not 8-hour static)
- All credentials in Vault (not in config files)
- File uploads validated (size, type, XXE)
- Rate limiting on login (prevent brute force)
- All endpoints protected (security tests + pre-commit hooks verify)

**From 04-PERFORMANCE-ISSUES.md** (performance fixes):
- DynamicCoverageCalculator batched (no N+1, query count assertions)
- Amion scraping parallelized (180s → 2-3s via goroutines) — **pending Spike 1 validation**
- Pagination on all list endpoints
- Database indexes designed upfront (from v1 analysis)
- Query performance tests prevent regression

---

## Definition of Done (Phase-by-Phase)

### Week 0 Done
- All three spikes completed with documented results
- Infrastructure clarifications obtained from hospital stakeholders
- Team confidence level assessed
- Revised timeline (if any spikes failed) communicated

### Phase 0 Done
- **v1 security patched and deployed to production**
- Schema translated (v1 → v2 mapping document with rationale)
- Job library integrated (Asynq OR Machinery, based on Spike 2)
- Migrations created (from schema translation)
- Team understands all decisions
- Development environment working
- Project charter signed
- Spike results documented in `docs/spikes/`

### Phase 1 Done
- Core services feature-complete
- ValidationResult working correctly (same semantics as v1)
- DynamicCoverageCalculator optimized (no N+1, query count assertions pass)
- 80% unit test coverage on services
- Query count assertions passing

### Phase 2 Done
- All REST endpoints working
- Security tests passing (100% of endpoints protected, NO bypasses)
- JWT + refresh tokens operational
- Vault integration proven
- Error responses consistent (unified ApiResponse format)
- Job system working (Asynq or Machinery)
- Rate limiting preventing brute force

### Phase 3 Done
- Amion scraper working with documented performance
- ODS import with validation (XXE protected)
- Full workflow tested (ODS → Amion → coverage, end-to-end)
- File upload validation preventing XXE
- Audit logging functional

### Phase 4 Done
- Test coverage 85%+
- Structured logging in place (JSON to stdout)
- Prometheus metrics exported (query counts, job retries, etc.)
- Documentation complete (runbooks, API docs, schema mapping)
- On-call runbooks tested with team
- Monitoring configured (Prometheus scraper)
- Load testing passed (100 concurrent users)
- Cutover plan finalized and dry-run tested

### Production Done
- v2 receiving 100% of traffic
- Monitoring shows no errors
- v1 in read-only mode (not decommissioned, for historical queries)
- Team trained on operations
- Runbooks updated
- On-call team has completed runbook review

---

## What This Plan Delivers

**On v1's timeline** (3-4 months, 3 people, with Week 0 validation):
- Complete Go rewrite with **clean, maintainable architecture**
- All v1 features + security improvements
- **Preserved v1 domain knowledge** (schema evolved, not redesigned)
- 10× performance improvement (Amion scraping: 180s → 2-3s) — **pending Spike 1 validation; otherwise documented actual improvement**
- 85%+ test coverage (vs v1's 60%)
- Production-grade monitoring from day 1
- Zero technical debt (no deprecated fields, explicit code, no ORM magic)
- **De-risked approach** (Week 0 validates critical assumptions)

**Safe cutover**:
- Fresh operational data (no migration complexity)
- Historical data preserved (v1 read-only for 1 year, audit logs archived 6 years)
- v1 kept as fallback during parallel operation
- Validation before switching
- Rollback possible if needed

**Long-term benefits**:
- **Go's simplicity makes code easier to maintain** (primary motivation)
- Explicit patterns prevent future N+1, security bypasses
- Monitoring catches issues before production problems
- Clean schema supports 5+ years of operation
- Less code to maintain (Asynq/Machinery vs custom job system, sqlc vs ORM magic)
- **On-call team has clear runbooks for production issues**

---

## Next Steps

1. **This week**: Share master plan (v2) with team, emphasize Week 0 validation importance
2. **Week 0**: Run three dependency validation spikes (3-5 days parallel)
3. **Week 1**: v1 security fixes (parallel track) → production-ready + Phase 0 begins
4. **Week 1-7**: Phase 0-1 (schema translation, core services)
5. **Week 7-10**: Phase 2 (API, security)
6. **Week 10-13**: Phase 3 (scrapers, integration)
7. **Week 13-15**: Phase 4 (testing, polish)
8. **Week 15-16**: Cutover week

**Total: 15-16 weeks to production** (standard full-time, 3 people, includes Week 0 validation)
**Critical: Week 0 de-risks the entire approach; Week 1 delivers secure v1** (removes urgency from v2 timeline)

---

## Document References

- **reimplement/00-OVERVIEW.md** — v1 assessment (B+ grade)
- **reimplement/01-TECHNICAL-DEBT.md** — Issues to prevent (14 issues detailed)
- **reimplement/02-WHAT-WORKED.md** — Patterns to replicate (12+ patterns)
- **reimplement/03-SECURITY-GAPS.md** — Security improvements (5 areas)
- **reimplement/04-PERFORMANCE-ISSUES.md** — Performance targets (10-100× improvements)
- **reimplement/09-LESSONS-LEARNED.md** — Design principles (12 insights)
- **schedJas/CLAUDE.md** — v1 development guide (reference for context)
- **schedJas/ARCHITECTURE_REFACTORED.md** — Dynamic coverage architecture (preserve in v2)

---

## Approval & Sign-Off

Project Charter for v2 Go Rewrite:

```
Approved by: _________________________ Date: _________
Architect:   _________________________ Date: _________
PM:          _________________________ Date: _________
Lead Dev:    _________________________ Date: _________
```

---

**Status**: Ready to Begin Week 0 Dependency Validation
**Confidence Level**: Medium-High (17 decisions locked, Week 0 validation incoming, clear roadmap)
**Risk Level**: Low-Medium (v1 security fixed Week 1, schema preserved not redesigned, Week 0 de-risks critical assumptions)
**Success Probability**: 90% on time (post-Week 0), 98% eventual delivery
**Key Success Factor**: Maintainability over novelty + early validation of critical assumptions
