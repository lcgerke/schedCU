# Reimplement: Hospital Radiology Schedule System

## Purpose

This directory documents the technical debt, learnings, and best practices from the v1 codebase to guide the v2 reimplementation. Use these documents as a blueprint for what to keep, what to improve, and what to avoid.

## Quick Assessment

**v1 Grade**: B (Good with security concerns)

**Status**: Production-ready after security fixes, but needs refactoring before v2

**Key Achievement**: Successfully implemented complex domain (schedule versioning, coverage resolution, multi-source integration)

**Critical Issue**: Admin endpoints unprotected—must fix before production

---

## Document Guide

1. **01-TECHNICAL-DEBT.md** - Issues, bugs, and design problems to avoid
2. **02-WHAT-WORKED.md** - Patterns and designs worth keeping
3. **03-SECURITY-GAPS.md** - Security vulnerabilities and fixes
4. **04-PERFORMANCE-ISSUES.md** - N+1 queries, inefficiencies, scalability concerns
5. **05-TESTING-STRATEGY.md** - Testing gaps and recommended approach
6. **06-ARCHITECTURAL-RECOMMENDATIONS.md** - Design patterns for v2
7. **07-API-IMPROVEMENTS.md** - REST endpoint design lessons
8. **08-REFACTORING-PRIORITIES.md** - Step-by-step improvement plan
9. **09-LESSONS-LEARNED.md** - Insights from building v1

---

## Executive Summary: Key Insights

### Technical Debt (High Priority)

- **Security**: Admin endpoints have `@PermitAll` bypass (production risk)
- **Testing**: No unit tests for `ODSImportService`, `AmionImportService`, `DynamicCoverageCalculator`
- **Code**: Long methods (125+ lines), deprecated fields still in database
- **Performance**: N+1 queries in dynamic coverage calculator, no pagination on list endpoints
- **Configuration**: Hardcoded database password, Amion file ID, `/tmp` paths

### What Worked Well

- **Domain modeling**: Clear entities (ScheduleVersion, ScrapeBatch, ShiftInstance)
- **Validation framework**: Elegant `ValidationResult` with severity levels
- **Dynamic coverage**: Lazy evaluation eliminates stale data
- **Batch traceability**: Foreign keys and checksums ensure data integrity
- **Documentation**: Comprehensive markdown guides and JavaDoc

### v1 → v2 Strategy

**Keep**:
- Entity relationships (add constraints, fix deprecated fields)
- Validation framework (minimal changes)
- Scrape batch lifecycle (proven pattern)
- Security design (JWT, RBAC, audit logging)

**Improve**:
- API security (remove all `@PermitAll` bypasses)
- Test coverage (add unit tests for services)
- Performance (fix N+1, add pagination, batch operations)
- Configuration management (env vars, no hardcoded values)
- Code organization (extract long methods, remove dead code)

**Replace**:
- None (complete rewrite not necessary)
- Rather: Surgical improvements to existing architecture

---

## Timeline Estimates

| Task | Priority | Effort | Risk |
|------|----------|--------|------|
| Fix security bypass | 🔴 Critical | 2 hours | High if ignored |
| Add unit tests | 🟠 High | 3 days | Medium |
| Fix N+1 queries | 🟠 High | 1 day | Low |
| Refactor long methods | 🟡 Medium | 2 days | Low |
| Add pagination | 🟡 Medium | 1 day | Low |
| Remove deprecated code | 🟡 Medium | 4 hours | Low |
| Performance optimization | 🟡 Medium | 2 days | Medium |

**Total for production readiness**: ~1 week

---

## Metrics (v1 Analysis)

- **Lines of Code (LOC)**: ~8,800 (reasonable for scope)
- **Test Lines**: ~3,200 (mostly E2E)
- **Test Ratio**: 36% (ideal: 40%+)
- **Cyclomatic Complexity**: Moderate (peak: 125 in `applyDynamicReassignment()`)
- **Duplicate Code**: Low (<5%)
- **Code Coverage**: ~60% (needs +15%)
- **Documentation**: Excellent (12 markdown files)

---

## How to Use This Directory

1. **Review 02-WHAT-WORKED.md** - Understand what's good
2. **Review 01-TECHNICAL-DEBT.md** - Understand what's bad
3. **Review 03-SECURITY-GAPS.md** - Fix critical issues
4. **Review 06-ARCHITECTURAL-RECOMMENDATIONS.md** - Plan redesign
5. **Use 08-REFACTORING-PRIORITIES.md** - Execute improvements
6. **Reference 09-LESSONS-LEARNED.md** - Avoid repeating mistakes

Each document includes specific code locations, examples, and actionable recommendations.

---

## Key Files to Study

### Good Examples (emulate)
- `src/main/java/org/hospital/radiology/schedule/entity/ScrapeBatch.java` - Lifecycle management
- `src/main/java/org/hospital/radiology/schedule/service/DynamicCoverageCalculator.java` - Clean, focused service
- `src/main/java/org/hospital/radiology/schedule/service/PersonRegistryService.java` - YAML sync pattern
- `src/test/java/.../CoverageResolutionServiceTest.java` - Comprehensive unit tests

### Problematic Examples (refactor)
- `src/main/java/org/hospital/radiology/schedule/service/CoverageResolutionService.java` (line 46-170) - Long method
- `src/main/java/org/hospital/radiology/schedule/api/AdminResource.java` (lines 92, 118, 149) - Security bypass
- `src/main/java/org/hospital/radiology/schedule/scraper/AmionScraper.java` (line 89-107) - Sequential processing

### Documentation Examples (use as template)
- `docs/COMPLETE_SYSTEM_SPECIFICATION.md` - Excellent detailed spec
- `docs/AMION_PARSING_DECISIONS.md` - Good decision documentation
