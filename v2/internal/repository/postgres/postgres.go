package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps a SQL database connection for all PostgreSQL operations
type DB struct {
	*sql.DB
}

// New creates a new PostgreSQL database connection
func New(connString string) (*DB, error) {
	sqldb, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{sqldb}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Health checks database connectivity
func (db *DB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

// BeginTx starts a new database transaction
func (db *DB) BeginTx(ctx context.Context) (*Tx, error) {
	sqlTx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Tx{
		tx: sqlTx,
	}, nil
}

// Tx wraps a SQL transaction and provides access to repositories
type Tx struct {
	tx *sql.Tx
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	if err := tx.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	if err := tx.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

// Repository accessors for transaction
// TODO: Implement transaction-based repositories
// This requires refactoring repositories to accept a common interface
// that both *sql.DB and *sql.Tx implement (e.g., QueryExecutor interface)
