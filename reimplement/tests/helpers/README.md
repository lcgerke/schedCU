# Test Helpers Package

Comprehensive testing utilities for database integration tests, including query counting, performance monitoring, and Testcontainers integration.

## Overview

This package provides:

1. **Query Counter Framework** - Track and assert database query execution
2. **Testcontainers Integration** - Helpers for containerized database testing
3. **Performance Monitoring** - Detect N+1 patterns and query regressions
4. **Comprehensive Documentation** - Examples and usage patterns

## Quick Start

### 1. Basic Query Counting

```go
func TestScheduleRepository(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Your test code
    schedule, err := repo.GetByID(ctx, id)
    require.NoError(t, err)

    // Assert query count
    require.NoError(t, helpers.AssertQueryCount(1, helpers.GetQueryCount()))
}
```

### 2. N+1 Detection

```go
func TestBatchLoad(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Load 100 schedules with assignments
    schedules, err := repo.ListWithAssignments(ctx)
    require.NoError(t, err)

    // Should not have N+1: expect 100 items, 1 query per item max
    require.NoError(t, helpers.AssertNoNPlusOne(len(schedules), 1))
}
```

### 3. Regression Detection

```go
func TestPerformance(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    coverage, err := calculator.ResolveCoverage(ctx, assignments)
    require.NoError(t, err)

    // Ensure queries don't exceed threshold
    require.NoError(t, helpers.AssertQueryCountLE(5))
}
```

## Files

### Core Implementation

- **`query_counter.go`** (390 lines)
  - `QueryCounter` - Main tracking structure
  - `QueryRecord` - Individual query metadata
  - `StartQueryCount()`, `GetQueryCount()`, `ResetQueryCount()`
  - `AssertQueryCount()`, `AssertQueryCountLE()`, `AssertNoNPlusOne()`
  - `QueryCountingDriver` - Database driver wrapper
  - Thread-safe operation with mutex protection

### Tests

- **`query_counter_test.go`** (35 tests, 600+ lines)
  - Core functionality tests
  - Assertion validation
  - Error handling
  - Concurrency testing
  - Edge cases

- **`integration_test.go`** (500+ lines)
  - Testcontainers integration tests
  - Connection string generation
  - Database container configuration
  - Multi-test cycle isolation
  - Query performance simulation

### Testcontainers Integration

- **`testcontainers_integration.go`** (210 lines)
  - `DatabaseContainer` - Container wrapper
  - `PostgresContainerConfig()`, `MySQLContainerConfig()`
  - `TestDatabaseSetup` - Lifecycle management
  - Connection string generation

### Documentation

- **`QUERY_COUNTER_USAGE.md`** (400+ lines)
  - Comprehensive usage guide
  - Integration patterns
  - Assertion strategies
  - Troubleshooting
  - Advanced patterns

- **`EXAMPLES.md`** (500+ lines)
  - 13+ practical examples
  - Real-world test scenarios
  - Performance testing patterns
  - Debugging techniques

## Test Statistics

```
Total Tests: 35
├── Query Counter Core Tests: 25
│   ├── Basic Operations: 6
│   ├── Assertions: 8
│   ├── Concurrency: 2
│   ├── Metadata: 3
│   └── Error Handling: 6
│
├── Integration Tests: 10
│   ├── Container Config: 5
│   ├── Connection Strings: 3
│   └── Multi-cycle Testing: 2

Success Rate: 100% (35/35)
Coverage: Core functionality + edge cases + concurrent operations
```

## Key Features

### 1. Query Tracking
- Captures all SQL queries
- Records arguments and duration
- Tracks execution timestamps
- Captures query errors
- Thread-safe with mutex protection

### 2. Assertions
- **Exact count**: `AssertQueryCount(expected, actual)`
- **Maximum count**: `AssertQueryCountLE(max)` for regression detection
- **N+1 detection**: `AssertNoNPlusOne(batchSize, queriesPerItem)`
- **Helpful error messages**: Includes query list and metadata

### 3. Integration
- Testcontainers support
- Connection string generation
- Database container lifecycle management
- Setup/teardown helpers

### 4. Debugging
- `LogQueries()` - Format all queries for output
- `GetQueries()` - Access all query metadata
- Detailed error messages with query context
- Query truncation for readability

## API Reference

### Core Functions

```go
// Lifecycle
StartQueryCount()          // Begin tracking
StopQueryCount()          // Stop without resetting
ResetQueryCount()         // Clear tracked queries

// Query Access
GetQueryCount() int       // Total count
GetQueries() []QueryRecord // All records with metadata
AppendQuery(record)        // Manual tracking

// Assertions
AssertQueryCount(expected, actual) error
AssertQueryCountLE(max) error
AssertNoNPlusOne(batchSize, queriesPerItem) error

// Debugging
LogQueries() string        // Format for output
```

### Structures

```go
type QueryRecord struct {
    SQL       string
    Args      []interface{}
    Duration  time.Duration
    Timestamp time.Time
    Error     error
}

type DatabaseContainer struct {
    Host, Port, Username, Password, Database, Driver string
}
```

## Integration with Testcontainers

```go
// Start PostgreSQL container
req := testcontainers.ContainerRequest{
    Image: "postgres:15",
    // ... configuration ...
}

container, _ := testcontainers.GenericContainer(ctx, ...)

// Create database connection
host, _ := container.Host(ctx)
port, _ := container.MappedPort(ctx, "5432")

db, _ := sql.Open("postgres", fmt.Sprintf(
    "postgres://user:pass@%s:%s/db?sslmode=disable",
    host, port.Port(),
))

// Enable query counting
helpers.ResetQueryCount()
helpers.StartQueryCount()

// Your tests here
```

## Common Patterns

### Per-Test Isolation

```go
func TestA(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
    // Test A code
}

func TestB(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
    // Test B code - independent of Test A
}
```

### Performance Regression Detection

```go
helpers.ResetQueryCount()
helpers.StartQueryCount()

result := service.CriticalOperation(ctx)

// Ensure no regression
require.NoError(t, helpers.AssertQueryCountLE(5),
    "Query count increased. Log:\n%s", helpers.LogQueries())
```

### N+1 Prevention

```go
helpers.ResetQueryCount()
helpers.StartQueryCount()

items, _ := repo.ListWithRelated(ctx)

// Ensure batch loading, not N+1
require.NoError(t, helpers.AssertNoNPlusOne(len(items), 1),
    "Detected N+1 pattern:\n%s", helpers.LogQueries())
```

## Design Decisions

1. **Global Counter**: Single global counter for simplicity - tests are sequential by default
2. **Thread-Safe**: Uses RWMutex for concurrent test support
3. **Copy-on-Read**: `GetQueries()` returns copy to prevent external modification
4. **No Panic**: Assertions return errors, don't panic (compatible with testify)
5. **Query Wrapper**: Provides `QueryCountingDriver` for low-level integration

## Limitations and Workarounds

### Driver-Level Wrapping
Go's `sql` package doesn't expose internal driver access. Two approaches:

1. **Registration-Time Wrapping** (Recommended):
   ```go
   sql.Register("postgres-counting", &QueryCountingDriver{
       underlying: &pq.Driver{},
   })
   ```

2. **Repository-Level Wrapping** (Fallback):
   ```go
   func (r *Repo) execWithTracking(sql string, args ...interface{}) {
       start := time.Now()
       rows, err := r.db.Query(sql, args...)
       helpers.AppendQuery(helpers.QueryRecord{...})
   }
   ```

### ORM Integration
For GORM, sqlc, or other ORMs:
- Hook at ORM's callback/middleware level
- Or wrap database.DB at initialization
- See `EXAMPLES.md` for patterns

## Next Steps

1. **Read Documentation**: Start with `QUERY_COUNTER_USAGE.md`
2. **Try Examples**: Review `EXAMPLES.md` for your use case
3. **Run Tests**: `go test -v ./tests/helpers/`
4. **Integrate**: Wrap your database driver or ORM
5. **Add Assertions**: Use in critical test paths

## File Sizes

```
query_counter.go               390 lines
query_counter_test.go          600 lines (35 tests)
testcontainers_integration.go  210 lines
integration_test.go            500 lines (10 tests)
QUERY_COUNTER_USAGE.md        400+ lines
EXAMPLES.md                   500+ lines
README.md                     This file
```

## Support

For issues or questions:

1. Check `QUERY_COUNTER_USAGE.md` - Comprehensive guide
2. Review `EXAMPLES.md` - Real-world patterns
3. See tests in `query_counter_test.go` - Implementation examples
4. Check test logs with `helpers.LogQueries()` - Debug aid

## License

Part of the SchedCU project. See project LICENSE.

## See Also

- `QUERY_COUNTER_USAGE.md` - Detailed usage guide
- `EXAMPLES.md` - 13+ practical examples
- `query_counter_test.go` - Test examples
- `integration_test.go` - Integration patterns
