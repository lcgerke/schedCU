# Query Counter Framework - Usage Guide

The Query Counter Framework provides comprehensive query tracking for database integration tests. It helps detect performance regressions, N+1 query patterns, and validates query counts for critical code paths.

## Overview

The framework tracks all SQL queries executed during tests and provides assertions to verify query counts and patterns.

**Key Features:**
- Query count tracking
- Query metadata (SQL text, arguments, duration, timestamp)
- Error messages with helpful context
- N+1 pattern detection
- Regression detection
- Thread-safe operation
- Simple test integration

## Basic Usage

### 1. Initialize Query Counting

In your test setup, call `StartQueryCount()`:

```go
func TestUserCreation(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
    defer helpers.StopQueryCount()

    // Your test code here
    user, err := service.CreateUser(ctx, "john@example.com")
    require.NoError(t, err)

    // Assert query count
    count := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCount(1, count))
}
```

### 2. Track Query Execution

Queries are automatically tracked when integrated with your database driver (see Integration section).

To manually track a query (useful for testing):

```go
helpers.AppendQuery(helpers.QueryRecord{
    SQL:   "SELECT * FROM users WHERE id = ?",
    Args:  []interface{}{42},
})
```

### 3. Assert Query Counts

Use `AssertQueryCount()` for exact count validation:

```go
func TestScheduleRetrieval(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    schedule, err := repo.GetSchedule(ctx, scheduleID)
    require.NoError(t, err)

    // Should execute exactly 1 query
    count := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCount(1, count),
        "Expected 1 query, got %d\n%s", count, helpers.LogQueries())
}
```

### 4. Detect Regressions

Use `AssertQueryCountLE()` to ensure query count doesn't increase:

```go
func TestCoverageCalculation(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    coverage, err := calculator.ResolveCoverage(ctx, assignments)
    require.NoError(t, err)

    // Ensure we don't regress to more queries
    require.NoError(t, helpers.AssertQueryCountLE(5),
        "Query count increased: %s", helpers.LogQueries())
}
```

### 5. Detect N+1 Patterns

Use `AssertNoNPlusOne()` to catch N+1 query patterns:

```go
func TestBatchUserAssignment(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    const batchSize = 100
    assignments, err := service.AssignUsersToShifts(ctx, userIDs, shiftID)
    require.NoError(t, err)

    // Should not execute more than 1 query per user
    // Formula: batchSize * (1 + queriesPerItem) = 100 * (1 + 1) = 200
    require.NoError(t, helpers.AssertNoNPlusOne(batchSize, 1),
        "N+1 detected:\n%s", helpers.LogQueries())
}
```

### 6. Debug Query Execution

When tests fail, use `LogQueries()` to see what was executed:

```go
if err := helpers.AssertQueryCount(1, helpers.GetQueryCount()); err != nil {
    t.Logf("Query log:\n%s", helpers.LogQueries())
    t.Fail()
}
```

## Integration with Test Database

### With Testcontainers (PostgreSQL)

```go
import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "github.com/schedcu/reimplement/tests/helpers"
)

func setupTestDB(ctx context.Context, t *testing.T) *sql.DB {
    // Create container request
    req := testcontainers.ContainerRequest{
        Image:        "postgres:15",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "testuser",
            "POSTGRES_PASSWORD": "testpass",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }

    // Start container
    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    require.NoError(t, err)
    t.Cleanup(func() {
        _ = container.Terminate(ctx)
    })

    // Get connection details
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")

    // Create database connection
    dsn := fmt.Sprintf(
        "postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
        host,
        port.Port(),
    )

    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)

    // Run migrations
    // ... your migration code ...

    return db
}

func TestScheduleRepository(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(ctx, t)
    defer db.Close()

    // Query counting setup
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Your test code
    repo := repository.NewScheduleRepository(db)
    schedule, err := repo.GetByID(ctx, scheduleID)
    require.NoError(t, err)

    // Assert queries
    require.NoError(t, helpers.AssertQueryCount(1, helpers.GetQueryCount()))
}
```

### Wrapping an Existing Database

For applications that don't use driver-level hooking, you can manually track queries:

```go
// Create a wrapper around database operations
func (repo *ScheduleRepository) GetByID(ctx context.Context, id uuid.UUID) (*Schedule, error) {
    sql := "SELECT * FROM schedules WHERE id = $1"

    // Track the query
    start := time.Now()
    rows, err := repo.db.QueryContext(ctx, sql, id)

    helpers.AppendQuery(helpers.QueryRecord{
        SQL:       sql,
        Args:      []interface{}{id},
        Duration:  time.Since(start),
        Timestamp: start,
        Error:     err,
    })

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // ... rest of implementation ...
}
```

## Assertion Patterns

### Pattern 1: Exact Count (Most Strict)

```go
// Verify exactly 5 queries executed
count := helpers.GetQueryCount()
require.NoError(t, helpers.AssertQueryCount(5, count),
    "Expected exactly 5 queries, got %d", count)
```

Use when you know exactly how many queries should execute.

### Pattern 2: Maximum Count (Regression Detection)

```go
// Ensure we don't execute more than 10 queries
require.NoError(t, helpers.AssertQueryCountLE(10),
    "Query count regressed: %s", helpers.LogQueries())
```

Use for regression detection - ensures queries don't increase over time.

### Pattern 3: N+1 Detection

```go
// Load 100 items, expect 1 query per item max (includes initial load)
require.NoError(t, helpers.AssertNoNPlusOne(100, 1),
    "Potential N+1 detected: %s", helpers.LogQueries())
```

Use when batch-loading related items to detect N+1 patterns.

### Pattern 4: No Queries (Cached/No-Op)

```go
helpers.ResetQueryCount()
helpers.StartQueryCount()

// This should not hit database
result := cache.Get(key)

require.Equal(t, 0, helpers.GetQueryCount(), "Should use cache, not database")
```

Use when testing cache behavior or no-op functions.

## Query Record Structure

Each tracked query is recorded with full metadata:

```go
type QueryRecord struct {
    SQL       string        // The SQL query text
    Args      []interface{} // Query arguments
    Duration  time.Duration // Time taken to execute
    Timestamp time.Time     // When the query was executed
    Error     error         // Any error from execution
}
```

Access individual queries:

```go
queries := helpers.GetQueries()
for i, q := range queries {
    fmt.Printf("Query %d:\n", i+1)
    fmt.Printf("  SQL: %s\n", q.SQL)
    fmt.Printf("  Args: %v\n", q.Args)
    fmt.Printf("  Duration: %v\n", q.Duration)
    fmt.Printf("  Error: %v\n", q.Error)
}
```

## Common Patterns & Anti-Patterns

### N+1 Anti-Pattern Detection

```go
func TestLoadSchedulesWithStaffing(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Simulating: Load 10 schedules, then for each load staffing
    // Bad: SELECT * FROM schedules (1 query)
    //      SELECT * FROM staffing WHERE schedule_id = $1 (10 queries)
    // Total: 11 queries (N+1)

    for i := 0; i < 10; i++ {
        helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT..."})
    }

    // This catches the N+1
    err := helpers.AssertNoNPlusOne(10, 0)
    require.Error(t, err) // Should fail because we have N+1
}
```

### Proper Batch Loading

```go
func TestLoadSchedulesWithStaffingOptimized(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Optimized: Load 10 schedules with staffing in 2 queries
    // SELECT * FROM schedules (1 query)
    // SELECT * FROM staffing WHERE schedule_id = ANY($1) (1 query)
    // Total: 2 queries

    helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM schedules"})
    helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM staffing WHERE schedule_id = ANY($1)"})

    // This should pass
    require.NoError(t, helpers.AssertQueryCount(2, helpers.GetQueryCount()))
}
```

## Error Message Examples

### Query Count Mismatch

```
query count mismatch: expected 1, got 3

Queries executed (3 total):
  1. SELECT * FROM users WHERE id = $1 [1.234ms]
  2. SELECT * FROM orders WHERE user_id = $1 [2.456ms]
  3. SELECT COUNT(*) FROM orders WHERE user_id = $1 [0.567ms]
```

### N+1 Detection

```
potential N+1 detected: expected <= 20 queries (batch: 10 items, 1 queries per item), got 21

Queries executed (21 total):
  1. SELECT * FROM users LIMIT 10 [1.234ms]
  2. SELECT * FROM orders WHERE user_id = $1 [0.234ms]
  3. SELECT * FROM orders WHERE user_id = $1 [0.245ms]
  ...
```

## Testing Guidelines

### 1. Always Reset Before Test

```go
func TestSomething(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()
    // ... test code ...
}
```

### 2. Test in Transaction Isolation

```go
func TestRepositoryWithIsolation(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(ctx, t)

    // Wrap each test in a transaction
    tx, err := db.BeginTx(ctx, nil)
    require.NoError(t, err)
    defer tx.Rollback()

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    repo := repository.NewScheduleRepository(tx)
    // ... test code ...
}
```

### 3. Separate Test Concerns

```go
// Good: One assertion per concept
func TestLoadSingleSchedule(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    schedule, err := repo.GetByID(ctx, id)
    require.NoError(t, err)
    require.NoError(t, helpers.AssertQueryCount(1, helpers.GetQueryCount()))
}

func TestLoadMultipleSchedules(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    schedules, err := repo.ListByDateRange(ctx, start, end)
    require.NoError(t, err)
    // Different assertion - may be different query count
    require.NoError(t, helpers.AssertQueryCountLE(5, helpers.GetQueryCount()))
}
```

## Troubleshooting

### Queries Not Being Tracked

Check:
1. Is `StartQueryCount()` called before executing queries?
2. Is the database driver properly integrated?
3. Are queries using the tracked connection?

### Assertion Fails Unexpectedly

Debug using:
```go
if err := helpers.AssertQueryCount(expected, helpers.GetQueryCount()); err != nil {
    t.Logf("Full query log:\n%s", helpers.LogQueries())
    t.Error(err)
}
```

### Performance Test Intermittency

Some queries may vary in count due to:
- Caching behavior
- Prepared statement creation
- Connection pool initialization

Use `AssertQueryCountLE()` for tolerance in performance-sensitive areas.

## Advanced Usage

### Custom Query Tracking

For frameworks that don't expose low-level hooks:

```go
type TrackedRepository struct {
    db *sql.DB
}

func (tr *TrackedRepository) execWithTracking(sql string, args ...interface{}) {
    start := time.Now()
    rows, err := tr.db.Query(sql, args...)

    helpers.AppendQuery(helpers.QueryRecord{
        SQL:       sql,
        Args:      args,
        Duration:  time.Since(start),
        Timestamp: start,
        Error:     err,
    })

    // ... rest of implementation ...
}
```

### Query Performance Analysis

```go
func analyzeQueryPerformance() {
    queries := helpers.GetQueries()

    slowestQuery := QueryRecord{}
    for _, q := range queries {
        if q.Duration > slowestQuery.Duration {
            slowestQuery = q
        }
    }

    fmt.Printf("Slowest query: %s (%v)", slowestQuery.SQL, slowestQuery.Duration)
}
```

### Batch Testing

```go
func TestMultipleOperations(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Operation 1
    op1Count := helpers.GetQueryCount()
    service.Operation1(ctx)
    op1Actual := helpers.GetQueryCount() - op1Count
    require.NoError(t, helpers.AssertQueryCount(2, op1Actual))

    // Operation 2
    op2Count := helpers.GetQueryCount()
    service.Operation2(ctx)
    op2Actual := helpers.GetQueryCount() - op2Count
    require.NoError(t, helpers.AssertQueryCount(1, op2Actual))
}
```

## Thread Safety

The Query Counter is thread-safe and can be used in concurrent tests:

```go
func TestConcurrentQueryExecution(t *testing.T) {
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    wg := sync.WaitGroup{}
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            service.ExecuteQuery(ctx)
        }()
    }

    wg.Wait()
    require.Equal(t, 10, helpers.GetQueryCount())
}
```

## See Also

- `query_counter.go` - Core implementation
- `testcontainers_integration.go` - Database container helpers
- `query_counter_test.go` - Comprehensive test examples
- `integration_test.go` - Integration patterns
