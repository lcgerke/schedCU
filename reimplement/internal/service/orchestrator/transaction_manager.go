package orchestrator

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// Warnable is an interface for errors that can be classified as warnings (non-critical).
type Warnable interface {
	error
	IsWarning() bool
}

// Transaction defines the interface for a database transaction.
type Transaction interface {
	Commit() error
	Rollback() error
}

// TxBeginExec defines the interface for database transaction execution.
type TxBeginExec interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
}

// TransactionManager manages database transactions for orchestrator phases.
type TransactionManager struct {
	db             TxBeginExec
	isolationLevel sql.IsolationLevel
	auditLogger    *log.Logger
}

// NewTransactionManager creates a new TransactionManager with Read Committed isolation.
func NewTransactionManager(db TxBeginExec) *TransactionManager {
	return &TransactionManager{
		db:             db,
		isolationLevel: sql.LevelReadCommitted,
		auditLogger:    log.Default(),
	}
}

// IsolationLevel returns the isolation level used for all transactions.
func (tm *TransactionManager) IsolationLevel() sql.IsolationLevel {
	return tm.isolationLevel
}

// Phase1Transaction executes a function within a database transaction for ODS import.
func (tm *TransactionManager) Phase1Transaction(ctx context.Context, fn func(context.Context) error) error {
	return tm.executePhaseTransaction(ctx, "phase 1 (ODS import)", fn)
}

// Phase2Transaction executes a function within a database transaction for Amion scraping.
func (tm *TransactionManager) Phase2Transaction(ctx context.Context, fn func(context.Context) error) error {
	return tm.executePhaseTransaction(ctx, "phase 2 (Amion scraping)", fn)
}

// Phase3Transaction executes a function within a database transaction for coverage calculation.
func (tm *TransactionManager) Phase3Transaction(ctx context.Context, fn func(context.Context) error) error {
	return tm.executePhaseTransaction(ctx, "phase 3 (coverage calculation)", fn)
}

// executePhaseTransaction is the core transaction execution logic shared by all phases.
func (tm *TransactionManager) executePhaseTransaction(
	ctx context.Context,
	phaseName string,
	fn func(context.Context) error,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	txOpts := &sql.TxOptions{
		Isolation: tm.isolationLevel,
		ReadOnly:  false,
	}

	tx, err := tm.db.BeginTx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("failed to begin %s transaction: %w", phaseName, err)
	}

	fnErr := fn(ctx)

	shouldRollback := false
	if fnErr != nil {
		if warnable, ok := fnErr.(Warnable); ok && warnable.IsWarning() {
			tm.auditLogger.Printf("[%s] Warning: %v (committing)", phaseName, fnErr)
			shouldRollback = false
		} else {
			shouldRollback = true
			tm.auditLogger.Printf("[%s] Error: %v (rolling back)", phaseName, fnErr)
		}
	}

	var finalErr error
	if shouldRollback {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			tm.auditLogger.Printf("[%s] Rollback error: %v (original: %v)", phaseName, rollbackErr, fnErr)
			finalErr = fmt.Errorf("%s failed and rollback failed: %v", phaseName, fnErr)
		} else {
			tm.auditLogger.Printf("[%s] Rolled back", phaseName)
			finalErr = fnErr
		}
	} else {
		if commitErr := tx.Commit(); commitErr != nil {
			tm.auditLogger.Printf("[%s] Commit error: %v", phaseName, commitErr)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				tm.auditLogger.Printf("[%s] Rollback after commit error: %v", phaseName, rollbackErr)
			}
			finalErr = fmt.Errorf("%s commit failed: %w", phaseName, commitErr)
		} else {
			tm.auditLogger.Printf("[%s] Committed", phaseName)
			finalErr = fnErr
		}
	}

	return finalErr
}
