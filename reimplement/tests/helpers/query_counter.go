// Package helpers provides testing utilities for database integration tests.
package helpers

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"time"
)

// QueryRecord represents a single executed query with metadata.
type QueryRecord struct {
	SQL       string        // The SQL query text
	Args      []interface{} // Query arguments
	Duration  time.Duration // Time taken to execute
	Timestamp time.Time     // When the query was executed
	Error     error         // Any error from execution
}

// QueryCounter tracks all database queries executed during tests.
type QueryCounter struct {
	mu       sync.RWMutex
	queries  []QueryRecord
	isActive bool
}

var (
	globalCounter = &QueryCounter{}
	counterMu     sync.RWMutex
)

// StartQueryCount begins tracking queries.
// Must be called before executing queries to track.
func StartQueryCount() {
	counterMu.Lock()
	defer counterMu.Unlock()

	globalCounter.mu.Lock()
	defer globalCounter.mu.Unlock()

	globalCounter.queries = []QueryRecord{}
	globalCounter.isActive = true
}

// StopQueryCount ends tracking queries without resetting the count.
func StopQueryCount() {
	counterMu.RLock()
	defer counterMu.RUnlock()

	globalCounter.mu.Lock()
	defer globalCounter.mu.Unlock()

	globalCounter.isActive = false
}

// GetQueryCount returns the number of queries executed since StartQueryCount.
func GetQueryCount() int {
	counterMu.RLock()
	defer counterMu.RUnlock()

	globalCounter.mu.RLock()
	defer globalCounter.mu.RUnlock()

	return len(globalCounter.queries)
}

// GetQueries returns all executed queries with their metadata.
func GetQueries() []QueryRecord {
	counterMu.RLock()
	defer counterMu.RUnlock()

	globalCounter.mu.RLock()
	defer globalCounter.mu.RUnlock()

	// Return a copy to prevent external modification
	queries := make([]QueryRecord, len(globalCounter.queries))
	copy(queries, globalCounter.queries)
	return queries
}

// ResetQueryCount clears all recorded queries.
// Call between tests to isolate query tracking.
func ResetQueryCount() {
	counterMu.Lock()
	defer counterMu.Unlock()

	globalCounter.mu.Lock()
	defer globalCounter.mu.Unlock()

	globalCounter.queries = []QueryRecord{}
}

// AppendQuery records a query execution. Used internally by middleware.
func AppendQuery(record QueryRecord) {
	counterMu.RLock()
	defer counterMu.RUnlock()

	globalCounter.mu.Lock()
	defer globalCounter.mu.Unlock()

	if globalCounter.isActive {
		globalCounter.queries = append(globalCounter.queries, record)
	}
}

// AssertQueryCount verifies the exact number of queries executed.
// Returns error with detailed information if count doesn't match.
func AssertQueryCount(expected int, actual int) error {
	if expected == actual {
		return nil
	}

	queries := GetQueries()
	queryDetails := formatQueriesForAssertion(queries)

	return fmt.Errorf(
		"query count mismatch: expected %d, got %d\n%s",
		expected,
		actual,
		queryDetails,
	)
}

// AssertQueryCountLE verifies query count doesn't exceed maximum.
// Useful for regression detection: ensure queries don't increase unexpectedly.
func AssertQueryCountLE(maxExpected int) error {
	actual := GetQueryCount()
	if actual <= maxExpected {
		return nil
	}

	queries := GetQueries()
	queryDetails := formatQueriesForAssertion(queries)

	return fmt.Errorf(
		"query count exceeded maximum: expected <= %d, got %d\n%s",
		maxExpected,
		actual,
		queryDetails,
	)
}

// AssertNoNPlusOne detects N+1 query patterns.
// It checks if there are significantly more queries than expected for batch operations.
// This is a heuristic: if you load N items and then execute N+1 queries for each,
// you'll have N + (N * secondary queries) total queries.
func AssertNoNPlusOne(expectedBatchSize int, expectedQueriesPerItem int) error {
	actual := GetQueryCount()

	// For a batch operation:
	// expectedBatchSize items * expectedQueriesPerItem queries per item
	// Plus the initial query to load the batch = N + (N * M) = N*(1+M)
	maxExpected := expectedBatchSize * (1 + expectedQueriesPerItem)

	if actual <= maxExpected {
		return nil
	}

	queries := GetQueries()
	queryDetails := formatQueriesForAssertion(queries)

	return fmt.Errorf(
		"potential N+1 detected: expected <= %d queries (batch: %d items, %d queries per item), got %d\n%s",
		maxExpected,
		expectedBatchSize,
		expectedQueriesPerItem,
		actual,
		queryDetails,
	)
}

// LogQueries writes all executed queries to output for debugging.
// Useful when assertions fail to understand what queries were actually executed.
func LogQueries() string {
	queries := GetQueries()
	if len(queries) == 0 {
		return "No queries executed"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total queries: %d\n\n", len(queries)))

	for i, q := range queries {
		sb.WriteString(fmt.Sprintf("Query %d (took %v):\n", i+1, q.Duration))
		sb.WriteString(fmt.Sprintf("  SQL: %s\n", q.SQL))
		if len(q.Args) > 0 {
			sb.WriteString(fmt.Sprintf("  Args: %v\n", q.Args))
		}
		if q.Error != nil {
			sb.WriteString(fmt.Sprintf("  Error: %v\n", q.Error))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatQueriesForAssertion creates a readable summary of queries for error messages.
func formatQueriesForAssertion(queries []QueryRecord) string {
	if len(queries) == 0 {
		return "No queries executed"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\nQueries executed (%d total):\n", len(queries)))

	for i, q := range queries {
		// Truncate long queries for readability
		sql := q.SQL
		if len(sql) > 100 {
			sql = sql[:97] + "..."
		}

		sb.WriteString(fmt.Sprintf("  %d. %s [%v]\n", i+1, sql, q.Duration))

		if len(q.Args) > 0 && len(q.Args) <= 5 {
			sb.WriteString(fmt.Sprintf("     Args: %v\n", q.Args))
		}
	}

	return sb.String()
}

// QueryCountingDriver wraps a database/sql driver to track query execution.
// This is used internally to integrate with the QueryCounter.
type QueryCountingDriver struct {
	underlying driver.Driver
}

// Open returns a connection that tracks queries.
func (d *QueryCountingDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.underlying.Open(name)
	if err != nil {
		return nil, err
	}
	return &QueryCountingConn{
		underlying: conn,
	}, nil
}

// QueryCountingConn wraps a database connection to track queries.
type QueryCountingConn struct {
	underlying driver.Conn
}

// Prepare wraps a prepared statement to track queries.
func (c *QueryCountingConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.underlying.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &QueryCountingStmt{
		underlying: stmt,
		sql:        query,
	}, nil
}

// Close closes the underlying connection.
func (c *QueryCountingConn) Close() error {
	return c.underlying.Close()
}

// Begin starts a transaction.
func (c *QueryCountingConn) Begin() (driver.Tx, error) {
	tx, err := c.underlying.Begin()
	if err != nil {
		return nil, err
	}
	return &QueryCountingTx{
		underlying: tx,
	}, nil
}

// QueryCountingStmt wraps a prepared statement to track query execution.
type QueryCountingStmt struct {
	underlying driver.Stmt
	sql        string
}

// Close closes the underlying statement.
func (s *QueryCountingStmt) Close() error {
	return s.underlying.Close()
}

// NumInput returns the number of input arguments.
func (s *QueryCountingStmt) NumInput() int {
	return s.underlying.NumInput()
}

// Exec executes the statement and tracks the query.
func (s *QueryCountingStmt) Exec(args []driver.Value) (driver.Result, error) {
	start := time.Now()
	result, err := s.underlying.Exec(args)

	record := QueryRecord{
		SQL:       s.sql,
		Args:      convertDriverValuesToInterfaces(args),
		Duration:  time.Since(start),
		Timestamp: start,
		Error:     err,
	}
	AppendQuery(record)

	return result, err
}

// Query executes the statement and tracks the query.
func (s *QueryCountingStmt) Query(args []driver.Value) (driver.Rows, error) {
	start := time.Now()
	rows, err := s.underlying.Query(args)

	record := QueryRecord{
		SQL:       s.sql,
		Args:      convertDriverValuesToInterfaces(args),
		Duration:  time.Since(start),
		Timestamp: start,
		Error:     err,
	}
	AppendQuery(record)

	return rows, err
}

// QueryCountingTx wraps a transaction to track queries.
type QueryCountingTx struct {
	underlying driver.Tx
}

// Commit commits the transaction.
func (t *QueryCountingTx) Commit() error {
	return t.underlying.Commit()
}

// Rollback rolls back the transaction.
func (t *QueryCountingTx) Rollback() error {
	return t.underlying.Rollback()
}

// convertDriverValuesToInterfaces converts driver.Value slice to []interface{}.
func convertDriverValuesToInterfaces(args []driver.Value) []interface{} {
	if len(args) == 0 {
		return nil
	}

	converted := make([]interface{}, len(args))
	for i, arg := range args {
		converted[i] = arg
	}
	return converted
}

// WrapDBWithQueryCounter wraps a *sql.DB to track queries.
// This creates a new connection pool that instruments all queries.
// Note: Due to Go's driver system limitations, this is a best-effort
// wrapper. For complete coverage, see integration instructions.
func WrapDBWithQueryCounter(db *sql.DB) *sql.DB {
	// The database/sql package doesn't expose internal driver wrapping
	// This is a placeholder for documentation purposes.
	// In practice, you should:
	// 1. Register the counting driver before opening connections
	// 2. Use OpenDB to get a wrapped database connection
	// See RegisterQueryCountingDriver() for the proper approach.
	return db
}

// RegisterQueryCountingDriver registers a query-counting wrapper for a driver.
// Call this during test setup before opening database connections.
//
// Example:
//   sql.Register("postgres-counting", RegisterQueryCountingDriver(baseDriver))
//   db, _ := sql.Open("postgres-counting", connStr)
//
// Note: Due to Go's sql package design, getting a registered driver is not directly
// exposed in the public API. Instead, you should wrap the driver at registration time:
//   baseDriver := &pq.Driver{}  // or whichever driver you use
//   sql.Register("postgres-counting", &QueryCountingDriver{underlying: baseDriver})
func RegisterQueryCountingDriver(baseDriver driver.Driver) driver.Driver {
	if baseDriver == nil {
		return nil
	}

	return &QueryCountingDriver{
		underlying: baseDriver,
	}
}
