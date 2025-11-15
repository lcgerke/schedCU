# Testcontainers Integration Setup Guide

This guide explains how to integrate the Query Counter Framework with Testcontainers for full end-to-end testing with containerized databases.

## Prerequisites

```bash
# Install Testcontainers Go module
go get github.com/testcontainers/testcontainers-go

# Install the database driver you'll use
go get github.com/lib/pq               # PostgreSQL
# OR
go get github.com/go-sql-driver/mysql  # MySQL
```

## Step 1: Create Database Container Helper

Create a new file `tests/helpers/postgres_container.go`:

```go
package helpers

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    _ "github.com/lib/pq"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

// StartPostgresContainer starts a PostgreSQL container for testing.
// Returns a DatabaseContainer and cleanup function.
func StartPostgresContainer(ctx context.Context) (*DatabaseContainer, error) {
    req := testcontainers.ContainerRequest{
        Image:        "postgres:15",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "testuser",
            "POSTGRES_PASSWORD": "testpass",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections").
            WithStartupTimeout(30 * time.Second),
    }

    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    if err != nil {
        return nil, fmt.Errorf("failed to start container: %w", err)
    }

    host, err := container.Host(ctx)
    if err != nil {
        _ = container.Terminate(ctx)
        return nil, fmt.Errorf("failed to get container host: %w", err)
    }

    port, err := container.MappedPort(ctx, "5432")
    if err != nil {
        _ = container.Terminate(ctx)
        return nil, fmt.Errorf("failed to get container port: %w", err)
    }

    dbContainer := &DatabaseContainer{
        Host:     host,
        Port:     port.Int(),
        Username: "testuser",
        Password: "testpass",
        Database: "testdb",
        Driver:   "postgres",
        Container: container,
        cleanup: func(ctx context.Context) error {
            return container.Terminate(ctx)
        },
    }

    return dbContainer, nil
}

// StartMySQLContainer starts a MySQL container for testing.
func StartMySQLContainer(ctx context.Context) (*DatabaseContainer, error) {
    req := testcontainers.ContainerRequest{
        Image:        "mysql:8.0",
        ExposedPorts: []string{"3306/tcp"},
        Env: map[string]string{
            "MYSQL_ROOT_PASSWORD": "rootpass",
            "MYSQL_USER":          "testuser",
            "MYSQL_PASSWORD":      "testpass",
            "MYSQL_DATABASE":      "testdb",
        },
        WaitingFor: wait.ForLog("ready for connections").
            WithStartupTimeout(30 * time.Second),
    }

    container, err := testcontainers.GenericContainer(ctx,
        testcontainers.GenericContainerRequest{
            ContainerRequest: req,
            Started:          true,
        })
    if err != nil {
        return nil, fmt.Errorf("failed to start container: %w", err)
    }

    host, err := container.Host(ctx)
    if err != nil {
        _ = container.Terminate(ctx)
        return nil, fmt.Errorf("failed to get container host: %w", err)
    }

    port, err := container.MappedPort(ctx, "3306")
    if err != nil {
        _ = container.Terminate(ctx)
        return nil, fmt.Errorf("failed to get container port: %w", err)
    }

    dbContainer := &DatabaseContainer{
        Host:     host,
        Port:     port.Int(),
        Username: "testuser",
        Password: "testpass",
        Database: "testdb",
        Driver:   "mysql",
        Container: container,
        cleanup: func(ctx context.Context) error {
            return container.Terminate(ctx)
        },
    }

    return dbContainer, nil
}
```

## Step 2: Create Test Setup Utility

Create `tests/integration/setup.go`:

```go
package integration

import (
    "context"
    "database/sql"
    "fmt"
    "testing"

    "github.com/schedcu/reimplement/tests/helpers"
)

// SetupPostgresDB starts a PostgreSQL container and returns a test database setup.
func SetupPostgresDB(t *testing.T) *helpers.TestDatabaseSetup {
    ctx := context.Background()

    // Start container
    container, err := helpers.StartPostgresContainer(ctx)
    if err != nil {
        t.Fatalf("failed to start container: %v", err)
    }

    // Open database
    db, err := container.OpenDB()
    if err != nil {
        _ = container.Terminate(ctx)
        t.Fatalf("failed to open database: %v", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(2)

    // Setup cleanup
    t.Cleanup(func() {
        _ = db.Close()
        _ = container.Terminate(context.Background())
    })

    return &helpers.TestDatabaseSetup{
        Container: container,
        DB:        db,
        Options: helpers.SetupOptions{
            RunMigrations:        true,
            QueryCountingEnabled: true,
            MaxConnections:       10,
        },
    }
}

// SetupMySQLDB starts a MySQL container and returns a test database setup.
func SetupMySQLDB(t *testing.T) *helpers.TestDatabaseSetup {
    ctx := context.Background()

    // Start container
    container, err := helpers.StartMySQLContainer(ctx)
    if err != nil {
        t.Fatalf("failed to start container: %v", err)
    }

    // Open database
    db, err := container.OpenDB()
    if err != nil {
        _ = container.Terminate(ctx)
        t.Fatalf("failed to open database: %v", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(2)

    // Setup cleanup
    t.Cleanup(func() {
        _ = db.Close()
        _ = container.Terminate(context.Background())
    })

    return &helpers.TestDatabaseSetup{
        Container: container,
        DB:        db,
        Options: helpers.SetupOptions{
            RunMigrations:        true,
            QueryCountingEnabled: true,
            MaxConnections:       10,
        },
    }
}

// RunMigrations applies database migrations.
// Customize this for your actual migration system.
func RunMigrations(t *testing.T, db *sql.DB) {
    // TODO: Implement your actual migrations
    // Example with golang-migrate:
    // m, err := migrate.New("file://migrations", dbURL)
    // if err != nil {
    //     t.Fatalf("migration error: %v", err)
    // }
    // if err := m.Up(); err != nil && err != migrate.ErrNoChange {
    //     t.Fatalf("migration up error: %v", err)
    // }
}
```

## Step 3: Use in Integration Tests

Create `tests/integration/repository_test.go`:

```go
package integration

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
    "github.com/schedcu/reimplement/tests/helpers"
)

func TestScheduleRepositoryCreate(t *testing.T) {
    // Setup
    setup := SetupPostgresDB(t)
    db := setup.DB

    // Run migrations
    RunMigrations(t, db)

    // Reset and start query counting
    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    // Execute test
    ctx := context.Background()
    // TODO: Implement test logic
    // schedule, err := repo.Create(ctx, &Schedule{...})
    // require.NoError(t, err)

    // Assert query count
    count := helpers.GetQueryCount()
    require.NoError(t, helpers.AssertQueryCount(1, count),
        "Create should execute exactly 1 query, got %d\n%s",
        count, helpers.LogQueries())
}

func TestScheduleRepositoryGetByID(t *testing.T) {
    setup := SetupPostgresDB(t)
    db := setup.DB
    RunMigrations(t, db)

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    ctx := context.Background()
    // TODO: Implement test logic
    // schedule, err := repo.GetByID(ctx, scheduleID)
    // require.NoError(t, err)

    require.NoError(t, helpers.AssertQueryCount(1, helpers.GetQueryCount()))
}

func TestScheduleRepositoryList(t *testing.T) {
    setup := SetupPostgresDB(t)
    db := setup.DB
    RunMigrations(t, db)

    // Insert test data
    // ...

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    ctx := context.Background()
    // TODO: Implement test logic
    // schedules, err := repo.List(ctx)
    // require.NoError(t, err)

    // Should use batch loading, not N+1
    require.NoError(t, helpers.AssertQueryCountLE(2))
}

func TestScheduleRepositoryBatchLoad(t *testing.T) {
    setup := SetupPostgresDB(t)
    db := setup.DB
    RunMigrations(t, db)

    const scheduleCount = 50

    // Insert test data
    // for i := 0; i < scheduleCount; i++ { ... }

    helpers.ResetQueryCount()
    helpers.StartQueryCount()

    ctx := context.Background()
    // TODO: Implement batch load test
    // schedules, err := repo.ListWithAssignments(ctx)
    // require.NoError(t, err)

    // Verify no N+1: should be max 2 queries (list + batch fetch assignments)
    require.NoError(t, helpers.AssertNoNPlusOne(scheduleCount, 1),
        "Detected N+1 pattern:\n%s", helpers.LogQueries())
}
```

## Step 4: Configure Test Fixtures

Create `tests/fixtures/schedules.sql`:

```sql
-- Test data for schedule tests
INSERT INTO schedules (id, hospital_id, start_date, end_date, created_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000100',
     '2024-01-01', '2024-01-31', NOW()),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000100',
     '2024-02-01', '2024-02-29', NOW());

INSERT INTO shift_instances (id, schedule_id, position, start_time, end_time, staff_member)
VALUES
    ('00000000-0000-0000-0000-000000001001', '00000000-0000-0000-0000-000000000001',
     'ER Doctor', '08:00', '16:00', 'John Doe'),
    ('00000000-0000-0000-0000-000000001002', '00000000-0000-0000-0000-000000000001',
     'Nurse', '16:00', '00:00', 'Jane Smith');
```

Create a fixture loader:

```go
// In tests/integration/fixtures.go
package integration

import (
    "database/sql"
    "testing"
    "io/ioutil"
)

func LoadFixtures(t *testing.T, db *sql.DB, fixtureFile string) {
    data, err := ioutil.ReadFile(fixtureFile)
    if err != nil {
        t.Fatalf("failed to read fixture file: %v", err)
    }

    if _, err := db.Exec(string(data)); err != nil {
        t.Fatalf("failed to load fixtures: %v", err)
    }
}

// Usage in test:
// LoadFixtures(t, db, "tests/fixtures/schedules.sql")
```

## Step 5: Run Tests

```bash
# Run all integration tests with query counting
go test -v ./tests/integration/

# Run with verbose logging
go test -v ./tests/integration/ -run TestScheduleRepository

# Run with race detection
go test -race ./tests/integration/
```

## Expected Output

Successful test run:

```
=== RUN   TestScheduleRepositoryCreate
--- PASS: TestScheduleRepositoryCreate (0.34s)
=== RUN   TestScheduleRepositoryGetByID
--- PASS: TestScheduleRepositoryGetByID (0.28s)
=== RUN   TestScheduleRepositoryList
--- PASS: TestScheduleRepositoryList (0.45s)
=== RUN   TestScheduleRepositoryBatchLoad
--- PASS: TestScheduleRepositoryBatchLoad (0.52s)
PASS
ok  	github.com/schedcu/reimplement/tests/integration	2.156s
```

Failed test with query logging:

```
=== RUN   TestScheduleRepositoryBatchLoad
--- FAIL: TestScheduleRepositoryBatchLoad (0.52s)
    repository_test.go:120: Detected N+1 pattern:

    Queries executed (102 total):
      1. SELECT * FROM schedules LIMIT 50 [1.234ms]
      2. SELECT * FROM assignments WHERE schedule_id = $1 [0.456ms]
      3. SELECT * FROM assignments WHERE schedule_id = $1 [0.467ms]
      ...
      101. SELECT * FROM assignments WHERE schedule_id = $1 [0.489ms]
      102. SELECT * FROM assignments WHERE schedule_id = $1 [0.478ms]
FAIL
```

## Debugging Container Issues

If containers fail to start:

```bash
# Check Docker is running
docker ps

# View container logs
docker logs <container_id>

# Check available images
docker images

# Pull image manually
docker pull postgres:15
docker pull mysql:8.0
```

## Performance Tips

1. **Reuse containers** across tests when possible
2. **Set appropriate timeouts** for slow environments
3. **Use connection pooling** - `SetMaxOpenConns(10)`
4. **Clean up properly** - use `t.Cleanup()` or `defer`
5. **Run tests in parallel** - use `t.Parallel()` for independent tests

## Advanced Patterns

### Parallel Tests with Separate Containers

```go
func TestMultipleDBs(t *testing.T) {
    t.Run("PostgreSQL", func(t *testing.T) {
        t.Parallel()
        setup := SetupPostgresDB(t)
        // Your test logic
    })

    t.Run("MySQL", func(t *testing.T) {
        t.Parallel()
        setup := SetupMySQLDB(t)
        // Your test logic
    })
}
```

### Suite-Level Setup

```go
var pgContainer *helpers.DatabaseContainer

func TestMain(m *testing.M) {
    // One-time setup
    ctx := context.Background()
    var err error
    pgContainer, err = helpers.StartPostgresContainer(ctx)
    if err != nil {
        fmt.Printf("failed to start container: %v\n", err)
        os.Exit(1)
    }

    code := m.Run()

    // One-time teardown
    pgContainer.Terminate(context.Background())
    os.Exit(code)
}
```

## Troubleshooting

### Container Startup Timeout

```go
wait.ForLog("ready for connections").
    WithStartupTimeout(60 * time.Second)  // Increase timeout
```

### Port Already in Use

```go
// Testcontainers randomizes ports, but if you see conflicts:
docker ps | grep postgres  # Find conflicting container
docker stop <container_id> # Stop it
```

### Connection Pool Exhaustion

```go
db.SetMaxOpenConns(25)   // Increase connection limit
db.SetMaxIdleConns(5)    // Increase idle connections
db.SetConnMaxLifetime(5 * time.Minute)
```

## Next Steps

1. Configure migrations in `RunMigrations()`
2. Create repository tests using the patterns shown
3. Add service layer tests with query counting
4. Set up CI/CD to run integration tests
5. Monitor for query regressions over time

## See Also

- `QUERY_COUNTER_USAGE.md` - Query counting usage guide
- `EXAMPLES.md` - More test examples
- Official Testcontainers docs: https://testcontainers.com/
