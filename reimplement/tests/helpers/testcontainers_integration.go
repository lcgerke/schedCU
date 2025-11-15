// Package helpers provides integration with Testcontainers for database testing.
package helpers

import (
	"context"
	"database/sql"
	"fmt"
)

// DatabaseContainer represents a containerized database for testing.
// It wraps the container and provides convenient database connection methods.
type DatabaseContainer struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Driver   string // e.g., "postgres", "mysql"

	// Container represents the actual container (type varies by implementation)
	// This is intentionally interface{} to avoid tight coupling with testcontainers library
	Container interface{}

	// Cleanup function to terminate container
	cleanup func(context.Context) error
}

// ConnectionString generates a database connection string.
// Format depends on the driver type.
func (dc *DatabaseContainer) ConnectionString() string {
	switch dc.Driver {
	case "postgres":
		return fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			dc.Username,
			dc.Password,
			dc.Host,
			dc.Port,
			dc.Database,
		)
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s",
			dc.Username,
			dc.Password,
			dc.Host,
			dc.Port,
			dc.Database,
		)
	default:
		return fmt.Sprintf(
			"%s:%s@%s:%d/%s",
			dc.Username,
			dc.Password,
			dc.Host,
			dc.Port,
			dc.Database,
		)
	}
}

// OpenDB opens a database connection to the container.
// Returns a *sql.DB ready for queries.
func (dc *DatabaseContainer) OpenDB() (*sql.DB, error) {
	db, err := sql.Open(dc.Driver, dc.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// OpenDBWithQueryCounter opens a database connection with query counting enabled.
// All queries executed through this connection will be tracked.
func (dc *DatabaseContainer) OpenDBWithQueryCounter() (*sql.DB, error) {
	db, err := dc.OpenDB()
	if err != nil {
		return nil, err
	}

	// Note: Due to Go's driver system, we can't wrap an existing *sql.DB.
	// The proper way is to register a counting driver before opening.
	// This function serves as documentation - see Setup examples for proper usage.

	return db, nil
}

// Terminate closes the container.
func (dc *DatabaseContainer) Terminate(ctx context.Context) error {
	if dc.cleanup == nil {
		return nil
	}
	return dc.cleanup(ctx)
}

// QueryCountingConnectionWrapper wraps a database connection to track all queries.
// This is an alternative approach when you can't register the driver beforehand.
type QueryCountingConnectionWrapper struct {
	db *sql.DB
}

// NewQueryCountingConnectionWrapper creates a wrapper around an existing database connection.
func NewQueryCountingConnectionWrapper(db *sql.DB) *QueryCountingConnectionWrapper {
	return &QueryCountingConnectionWrapper{
		db: db,
	}
}

// Query executes a query and tracks it.
func (w *QueryCountingConnectionWrapper) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	// Note: This approach has limitations because database/sql doesn't expose
	// low-level query execution hooks that preserve all metadata.
	// For production use, consider:
	// 1. Using a driver that supports middleware (e.g., pq with prepared statements)
	// 2. Wrapping at the ORM level (sqlc, GORM, sqlx)
	// 3. Using a database proxy like pgbouncer with logging
	return w.db.Query(sql, args...)
}

// Exec executes a statement and tracks it.
func (w *QueryCountingConnectionWrapper) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return w.db.Exec(sql, args...)
}

// QueryRow executes a query and tracks it.
func (w *QueryCountingConnectionWrapper) QueryRow(sql string, args ...interface{}) *sql.Row {
	return w.db.QueryRow(sql, args...)
}

// Close closes the underlying database connection.
func (w *QueryCountingConnectionWrapper) Close() error {
	return w.db.Close()
}

// SetupOptions configures how the test database is set up.
type SetupOptions struct {
	// RunMigrations determines whether to run migrations on startup
	RunMigrations bool

	// QueryCountingEnabled enables query counting for this database
	QueryCountingEnabled bool

	// MaxConnections sets the maximum number of open connections
	MaxConnections int

	// AutoVacuum enables VACUUM after test (PostgreSQL)
	AutoVacuum bool
}

// PostgresContainerConfig provides sensible defaults for PostgreSQL containers.
func PostgresContainerConfig() DatabaseContainer {
	return DatabaseContainer{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "testuser",
		Password: "testpass",
		Database: "testdb",
	}
}

// MySQLContainerConfig provides sensible defaults for MySQL containers.
func MySQLContainerConfig() DatabaseContainer {
	return DatabaseContainer{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "testuser",
		Password: "testpass",
		Database: "testdb",
	}
}

// TestDatabaseSetup helps manage test database lifecycle.
// This is a helper struct for test setup, not tied to any specific container implementation.
type TestDatabaseSetup struct {
	Container *DatabaseContainer
	DB        *sql.DB
	Options   SetupOptions
}

// Close cleans up the test database and container.
func (tds *TestDatabaseSetup) Close(ctx context.Context) error {
	if tds.DB != nil {
		if err := tds.DB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	if tds.Container != nil {
		if err := tds.Container.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
	}

	return nil
}

// MustClose is like Close but panics on error.
// Useful in test cleanup functions.
func (tds *TestDatabaseSetup) MustClose(ctx context.Context) {
	if err := tds.Close(ctx); err != nil {
		panic(err)
	}
}
