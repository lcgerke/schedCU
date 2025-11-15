package helpers

import (
	"testing"
)

// TestPostgresContainerConfig validates default PostgreSQL config.
func TestPostgresContainerConfig(t *testing.T) {
	config := PostgresContainerConfig()

	if config.Driver != "postgres" {
		t.Errorf("driver should be 'postgres', got '%s'", config.Driver)
	}

	if config.Username != "testuser" {
		t.Errorf("username should be 'testuser', got '%s'", config.Username)
	}

	if config.Port != 5432 {
		t.Errorf("port should be 5432, got %d", config.Port)
	}
}

// TestMySQLContainerConfig validates default MySQL config.
func TestMySQLContainerConfig(t *testing.T) {
	config := MySQLContainerConfig()

	if config.Driver != "mysql" {
		t.Errorf("driver should be 'mysql', got '%s'", config.Driver)
	}

	if config.Port != 3306 {
		t.Errorf("port should be 3306, got %d", config.Port)
	}
}

// TestConnectionStringPostgres generates correct PostgreSQL connection strings.
func TestConnectionStringPostgres(t *testing.T) {
	config := PostgresContainerConfig()
	config.Host = "db.example.com"
	config.Port = 5433

	connStr := config.ConnectionString()

	expected := "postgres://testuser:testpass@db.example.com:5433/testdb?sslmode=disable"
	if connStr != expected {
		t.Errorf("connection string mismatch:\nexpected: %s\ngot:      %s", expected, connStr)
	}
}

// TestConnectionStringMySQL generates correct MySQL connection strings.
func TestConnectionStringMySQL(t *testing.T) {
	config := MySQLContainerConfig()
	config.Host = "db.example.com"
	config.Port = 3307

	connStr := config.ConnectionString()

	expected := "testuser:testpass@tcp(db.example.com:3307)/testdb"
	if connStr != expected {
		t.Errorf("connection string mismatch:\nexpected: %s\ngot:      %s", expected, connStr)
	}
}

// TestDatabaseContainerConfig allows custom values.
func TestDatabaseContainerConfig(t *testing.T) {
	config := DatabaseContainer{
		Driver:   "custom",
		Host:     "custom-host",
		Port:     9999,
		Username: "custom-user",
		Password: "custom-pass",
		Database: "custom-db",
	}

	if config.Driver != "custom" {
		t.Error("should allow custom driver")
	}

	if config.Host != "custom-host" {
		t.Error("should allow custom host")
	}
}

// TestSetupOptionsDefaults validates setup options.
func TestSetupOptionsDefaults(t *testing.T) {
	opts := SetupOptions{
		RunMigrations:        true,
		QueryCountingEnabled: true,
		MaxConnections:       10,
		AutoVacuum:           false,
	}

	if !opts.RunMigrations {
		t.Error("RunMigrations should be true")
	}

	if !opts.QueryCountingEnabled {
		t.Error("QueryCountingEnabled should be true")
	}

	if opts.MaxConnections != 10 {
		t.Error("MaxConnections should be 10")
	}
}

// TestTestDatabaseSetupClose validates cleanup.
func TestTestDatabaseSetupClose(t *testing.T) {
	// This test validates the structure, not actual container operations
	// since we don't have actual containers in unit tests

	setup := &TestDatabaseSetup{
		Container: &DatabaseContainer{
			Driver: "postgres",
		},
		DB: nil, // In real tests, this would be a *sql.DB
		Options: SetupOptions{
			RunMigrations:        true,
			QueryCountingEnabled: true,
		},
	}

	// Verify fields are accessible
	if setup.Container.Driver != "postgres" {
		t.Error("container driver should be accessible")
	}

	if !setup.Options.RunMigrations {
		t.Error("options should be accessible")
	}
}

// TestMockQueryCounterWithDatabase simulates database testing with query counting.
// This is a mock test showing how the system would work with a real database.
func TestMockQueryCounterWithDatabase(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	// Simulate initial connection query
	AppendQuery(QueryRecord{
		SQL: "SELECT version()",
	})

	// Simulate a user creation query
	AppendQuery(QueryRecord{
		SQL:   "INSERT INTO users (name, email) VALUES (?, ?)",
		Args:  []interface{}{"John Doe", "john@example.com"},
	})

	// Simulate a user lookup query
	AppendQuery(QueryRecord{
		SQL:   "SELECT * FROM users WHERE id = ?",
		Args:  []interface{}{1},
	})

	// Verify we have 3 queries
	count := GetQueryCount()
	if count != 3 {
		t.Fatalf("expected 3 queries, got %d", count)
	}

	// Assert exact count
	if err := AssertQueryCount(3, count); err != nil {
		t.Errorf("assertion failed: %v", err)
	}

	// Verify we can get all queries
	queries := GetQueries()
	if len(queries) != 3 {
		t.Errorf("expected 3 query records, got %d", len(queries))
	}

	// Verify query details
	if !contains(queries[1].SQL, "INSERT") {
		t.Error("second query should be INSERT")
	}

	if len(queries[2].Args) != 1 || queries[2].Args[0] != 1 {
		t.Error("SELECT query should have id=1 arg")
	}
}

// TestQueryCounterWithNPlusOneDetection demonstrates N+1 detection.
func TestQueryCounterWithNPlusOneDetection(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	// Load 5 users
	AppendQuery(QueryRecord{
		SQL: "SELECT id, name FROM users LIMIT 5",
	})

	// N+1 pattern: for each user, fetch their orders
	// This is the anti-pattern we want to detect
	for i := 1; i <= 5; i++ {
		AppendQuery(QueryRecord{
			SQL:   "SELECT * FROM orders WHERE user_id = ?",
			Args:  []interface{}{i},
		})
	}

	// Total: 1 + 5 = 6 queries
	count := GetQueryCount()
	if count != 6 {
		t.Fatalf("expected 6 queries (1 + 5), got %d", count)
	}

	// Try to assert this doesn't exceed a limit (it does)
	err := AssertQueryCountLE(4)
	if err == nil {
		t.Error("should fail when query count exceeds max")
	}

	// Demonstrate N+1 detection with expected batch operation
	// If we expect 5 items with 0 additional queries per item, max should be 5
	// But we have 6, so it should fail
	err = AssertNoNPlusOne(5, 0)
	if err == nil {
		t.Error("should detect N+1 pattern")
	}

	// However, if we expect 5 items with 1 query per item, max is 5*(1+1)=10
	// We have 6, so it should pass
	err = AssertNoNPlusOne(5, 1)
	if err != nil {
		t.Errorf("should pass when within expected threshold: %v", err)
	}
}

// TestQueryCounterErrorMessageQuality validates helpful error messages.
func TestQueryCounterErrorMessageQuality(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{
		SQL:   "SELECT * FROM users WHERE id = ?",
		Args:  []interface{}{42},
	})

	AppendQuery(QueryRecord{
		SQL: "SELECT * FROM orders",
	})

	AppendQuery(QueryRecord{
		SQL: "SELECT COUNT(*) FROM logs",
	})

	// Assert with wrong expected count
	err := AssertQueryCount(1, 3)
	if err == nil {
		t.Fatal("should fail")
	}

	errMsg := err.Error()

	// Check that error message includes helpful information
	checks := []string{
		"expected 1",
		"got 3",
		"SELECT",
	}

	for _, check := range checks {
		if !contains(errMsg, check) {
			t.Errorf("error message should include '%s', got: %s", check, errMsg)
		}
	}
}

// TestLogQueriesOutput validates query logging format.
func TestLogQueriesOutput(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{
		SQL:   "SELECT * FROM users WHERE name = ?",
		Args:  []interface{}{"Alice"},
	})

	AppendQuery(QueryRecord{
		SQL: "DELETE FROM temp_data",
	})

	output := LogQueries()

	// Should contain total count
	if !contains(output, "Total queries: 2") {
		t.Errorf("output should show total count: %s", output)
	}

	// Should contain query numbers
	if !contains(output, "Query 1") {
		t.Errorf("output should number queries: %s", output)
	}

	// Should contain SQL
	if !contains(output, "SELECT") && !contains(output, "DELETE") {
		t.Errorf("output should contain query SQL: %s", output)
	}

	// Should contain args if provided
	if !contains(output, "Alice") {
		t.Errorf("output should contain query args: %s", output)
	}
}

// TestQueryCounterMultipleTestCycles demonstrates per-test isolation.
func TestQueryCounterMultipleTestCycles(t *testing.T) {
	// Test 1: Simple SELECT
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	if GetQueryCount() != 1 {
		t.Fatal("test 1: should have 1 query")
	}

	// Test 2: Fresh start
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "INSERT INTO table VALUES (1)"})
	AppendQuery(QueryRecord{SQL: "UPDATE table SET x = 2"})
	if GetQueryCount() != 2 {
		t.Fatal("test 2: should have 2 queries")
	}

	// Test 3: Another fresh start
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "DELETE FROM table"})
	if GetQueryCount() != 1 {
		t.Fatal("test 3: should have 1 query")
	}

	// Verify each test cycle is isolated
	if GetQueryCount() != 1 {
		t.Error("should be at test 3 state")
	}
}

// TestQueryCounterWithTimeout simulates query with duration.
func TestQueryCounterWithTimeout(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	// Fast query
	AppendQuery(QueryRecord{
		SQL:      "SELECT 1",
		Duration: 1 * 1000, // 1 microsecond in nanoseconds
	})

	// Slow query
	AppendQuery(QueryRecord{
		SQL:      "SELECT * FROM huge_table",
		Duration: 5000 * 1000 * 1000, // 5 seconds
	})

	queries := GetQueries()
	if len(queries) != 2 {
		t.Fatal("should have 2 queries")
	}

	if queries[0].Duration > queries[1].Duration {
		t.Error("first query should be faster than second")
	}

	output := LogQueries()
	if !contains(output, "5s") {
		t.Errorf("output should show duration: %s", output)
	}
}
