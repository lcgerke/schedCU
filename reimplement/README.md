# Reimplement: Hospital Radiology Schedule System v2

Comprehensive analysis and refactoring guide for the Hospital Radiology Schedule System.

This directory contains the learnings from v1 implementation and detailed guidance for v2 development.

---

## Quick Start: Where to Begin

**You're implementing v2 in Go?** (START HERE)
→ Read **`10-MASTER-PLAN-v2.md`** (30 min)
- All 17 decisions locked in
- Complete 14-week timeline
- Team roles and responsibilities
- Success metrics and go/no-go gates

**Then read these in order:**
1. `02-WHAT-WORKED.md` - Patterns to replicate
2. `09-LESSONS-LEARNED.md` - Design principles
3. Reference others as needed during development

**You have 30 minutes?**
→ Read `10-MASTER-PLAN-v2.md` Executive Summary + Timeline

**You have 2 hours?**
→ Read `10-MASTER-PLAN-v2.md` completely

**You're fixing a specific issue in v1?**
→ Jump to the relevant document using the index below

---

## Document Index

| Document | Purpose | Length | For Whom |
|----------|---------|--------|----------|
| **`10-MASTER-PLAN-v2.md`** | **MASTER PLAN: All 17 decisions, timeline, roles, metrics** | **30 min** | **Everyone implementing v2** |
| `00-OVERVIEW.md` | v1 assessment and high-level overview | 10 min | v1 context |
| `02-WHAT-WORKED.md` | Patterns to replicate in v2 | 25 min | v2 Architects/Developers |
| `09-LESSONS-LEARNED.md` | Design principles from v1 | 25 min | v2 Team |
| `01-TECHNICAL-DEBT.md` | v1 issues to prevent in v2 | 30 min | Reference (specific problems) |
| `03-SECURITY-GAPS.md` | v1 security gaps (context only) | 20 min | Reference (security decisions made) |
| `04-PERFORMANCE-ISSUES.md` | v1 performance issues (context only) | 15 min | Reference (performance targets set) |
| `08-REFACTORING-PRIORITIES.md` | v1 refactoring plan (deprecated) | 30 min | Reference (v1 only, superseded by v2 plan) |

---

## v1 Assessment

**Grade**: B+ (Good with security concerns)

**What Works Well**:
- ✅ Excellent domain modeling (entities, relationships)
- ✅ Elegant validation framework
- ✅ Dynamic coverage calculation (self-healing)
- ✅ Batch traceability system
- ✅ Comprehensive documentation
- ✅ E2E test coverage

**What Needs Work**:
- 🔴 Security: Admin endpoints publicly accessible
- 🔴 Testing: No unit tests for core services
- 🔴 Performance: N+1 queries, no pagination
- 🟠 Debt: Long methods, magic strings, dead code
- 🟡 Config: Hardcoded credentials and values

---

## Critical Path: Minimum for Production

If you only have 1 week to fix v1 before production:

1. **Day 1**: Fix security bypass (remove `@PermitAll`)
2. **Day 2-3**: Move credentials to environment variables
3. **Day 4**: Add file upload validation
4. **Day 5-6**: Security testing and verification
5. **Day 7**: Cleanup and final testing

**Estimated effort**: 40 hours for 1 person

→ See `08-REFACTORING-PRIORITIES.md` Phase 1 for detailed plan

---

## Full Refactoring: For Production Quality

**Complete v1 → v2 transformation**: 6-10 weeks

**What's included**:
- Security hardening
- Test coverage improvements (60% → 85%)
- Performance optimizations (10-100× faster queries)
- Code refactoring (cleaner, maintainable)
- Documentation completion
- Configuration best practices

→ See `08-REFACTORING-PRIORITIES.md` for full plan

---

## Key Metrics

### Code Quality

| Metric | v1 | v2 Goal |
|--------|----|---------|
| Test coverage | 60% | 85%+ |
| Cyclomatic complexity | 125 (peak) | <15 per method |
| Code duplication | Low | None |
| Dead code | Some | None |
| Magic strings | Many | None (use constants) |

### Performance

| Metric | v1 | v2 |
|--------|----|----|
| N+1 queries | Yes | No |
| Query per date | 21 | 2 |
| Pagination | No | Yes |
| Async tasks | No | Yes |
| Index coverage | Partial | Full |

### Security

| Issue | v1 | v2 |
|-------|----|----|
| Admin endpoints protected | ❌ | ✅ |
| File upload validated | ❌ | ✅ |
| Credentials in env vars | ❌ | ✅ |
| Rate limiting | ❌ | ✅ |
| Security tests | ❌ | ✅ |

---

## Architecture: What to Keep

```
┌─────────────────────────────────────┐
│     REST API (Quarkus JAX-RS)      │
│  AdminResource, ScheduleResource    │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│   Service Layer (Business Logic)    │
│  Orchestrator, CoverageCalculator   │
│  ImportServices, ResolutionService  │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│    Domain Layer (Entities)          │
│  ScheduleVersion, ScrapeBatch       │
│  ShiftInstance, Assignment, Person  │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Infrastructure (DB, Parsers)       │
│  PostgreSQL, ODS Parser, Scraper    │
└─────────────────────────────────────┘
```

This architecture is **good** and should be kept for v2. Improvements:
- Strengthen dependency injection
- Reduce coupling in long methods
- Add circuit breaker for external APIs
- Implement async for long-running tasks

---

## Technology Stack

**Keep in v2**:
- Quarkus 3.5+ (lightweight, native support)
- PostgreSQL (reliable, feature-rich)
- Hibernate Panache (less boilerplate)
- JWT authentication (stateless, scalable)
- Flyway migrations (version controlled)

**Add to v2** (if needed):
- Resilience4j (circuit breaker)
- Micrometer (metrics)
- Testcontainers (integration tests)
- jqwik (property-based testing)

**Remove from v2**:
- Selenium E2E tests (if UI changes significantly)
- Temporary debug code
- Deprecated entities/fields

---

## Decision Framework for v2

When making architectural decisions:

**Ask**:
1. Does it align with v1 domain model? (Yes → Use existing pattern)
2. How often does this data change? (Frequently → Lazy evaluation)
3. What's the scale? (Large → Use batching, pagination)
4. Is it tested? (No → Add tests before shipping)
5. Is it documented? (No → Add before merge)

---

## Common Patterns to Reuse

### Domain Modeling
```java
// ✅ Use entity lifecycle (STAGING → PRODUCTION)
// ✅ Use temporal validity (effectiveStart/End)
// ✅ Use soft deletes (deletedAt)
// ✅ Use audit trail (AuditLog)
```

### Validation
```java
// ✅ Use ValidationResult with severity levels
// ✅ Collect all errors (don't fail fast)
// ✅ Store as JSON for audit trail
// ✅ Separate canImport() from canPromote()
```

### Batch Processing
```java
// ✅ Use batch headers (ScrapeBatch pattern)
// ✅ Atomic operations (all or nothing)
// ✅ Checksums for integrity
// ✅ Soft delete for retention
```

### API Design
```java
// ✅ Use response wrapper (ApiResponse<T>)
// ✅ Consistent error format
// ✅ Proper HTTP status codes
// ✅ Pagination for large lists
```

---

## What's Already Done in v1

Don't redo these:

- ✅ Core entity relationships (correct, stable)
- ✅ Validation framework (well-designed, reuse)
- ✅ Scrape batch system (proven, just needs testing)
- ✅ Dynamic coverage logic (elegant solution)
- ✅ Person registry sync (clean YAML pattern)
- ✅ JWT authentication (standard, good)

---

## What Needs Work in v1

Fix these before v2:

- 🔴 **Security**: Remove `@PermitAll` (CRITICAL)
- 🔴 **Database**: Remove deprecated `reassignedShiftType`
- 🔴 **Performance**: Fix N+1 query in DynamicCoverageCalculator
- 🟠 **Tests**: Add unit tests for services
- 🟠 **Code**: Extract long methods
- 🟡 **Config**: Move credentials to env vars

---

## Related Files

**In this repository**:
- `CLAUDE.md` - Claude Code development guide
- `README.md` - Project overview
- `pom.xml` - Maven configuration
- `docs/` - Architecture and specification documents

**External references**:
- [Quarkus Guide](https://quarkus.io/guides/)
- [Hibernate Panache](https://quarkus.io/guides/hibernate-orm-panache)
- [JWT in Quarkus](https://quarkus.io/guides/security-jwt)

---

## Contributing to v2

When implementing improvements:

1. **Read**: Start with relevant documents
2. **Understand**: Why does v1 do it this way?
3. **Plan**: What's the improvement? Why is it better?
4. **Test**: Write tests BEFORE code
5. **Review**: Get peer review
6. **Document**: Update docs and CLAUDE.md

---

## FAQ

**Q: Should we rewrite v2 from scratch?**
A: No. v1 architecture is solid. Improve it surgically:
- Keep domain model
- Keep validation framework
- Fix specific issues (security, performance, tests)

**Q: What's the biggest risk in v2?**
A: The security bypass in v1. Fix it immediately before production.

**Q: How long should v2 take?**
A: 6-10 weeks for comprehensive improvement (2 people)
Or 4 weeks minimum for production-critical fixes

**Q: What's the best place to start?**
A: Phase 1 in `08-REFACTORING-PRIORITIES.md` (security fixes)

**Q: Do we need to rewrite tests?**
A: No, keep E2E tests. Add unit tests for services.

**Q: Can we ship v1 as-is?**
A: Only after fixing Phase 1 (security). Then plan Phase 2-3.

---

## Document Generation

These documents were created through:
1. Deep code analysis (exploring codebase structure)
2. Architecture review (component interactions)
3. Technical debt assessment (known issues)
4. Security audit (vulnerability analysis)
5. Performance profiling (query patterns)
6. Best practices synthesis (industry standards)

All recommendations are backed by:
- Code locations and line numbers
- Specific examples
- Risk assessment
- Effort estimation
- Success metrics

---

## Feedback & Updates

As v2 implementation proceeds:

1. **Update** this directory with findings
2. **Document** decisions made (architecture decision records)
3. **Share** learnings with team
4. **Iterate** on recommendations based on actual work

The goal is for each subsequent reimplementation to be easier and faster.

---

## Next Actions

**This week**:
- [ ] Read `00-OVERVIEW.md` + `09-LESSONS-LEARNED.md`
- [ ] Team discussion on key findings
- [ ] Plan Phase 1 tasks

**Next week**:
- [ ] Detail Phase 1 sprint (story points, subtasks)
- [ ] Assign owners to each task
- [ ] Set up code review checklist

**Following week**:
- [ ] Begin Phase 1 implementation
- [ ] Daily standup on progress
- [ ] Security audit when complete

---

## Questions?

Refer to specific documents:
- Architecture question? → `02-WHAT-WORKED.md`
- Security issue? → `03-SECURITY-GAPS.md`
- Performance problem? → `04-PERFORMANCE-ISSUES.md`
- Testing question? → `05-TESTING-STRATEGY.md`
- Design question? → `09-LESSONS-LEARNED.md`
- Implementation plan? → `08-REFACTORING-PRIORITIES.md`

---

**Last updated**: 2024-11-15
**Status**: Ready for v2 planning
**Grade**: v1 is B+ (good with security concerns)
**Recommendation**: Fix Phase 1 before production, then plan Phase 2-3
