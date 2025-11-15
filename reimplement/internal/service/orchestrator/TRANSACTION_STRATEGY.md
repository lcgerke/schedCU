# Transaction Handling for Orchestration

## Overview

The TransactionManager provides per-phase database transaction management for the orchestration workflow. Each of the three phases (ODS import, Amion scraping, and coverage calculation) executes within its own independent transaction, allowing for partial success when some phases fail while others succeed.

## Design Philosophy

### Separate Transactions Per Phase

Each orchestration phase gets its own database transaction:
- **Phase 1**: ODS file import (ScheduleVersion + ShiftInstances)
- **Phase 2**: Amion scraping (Assignments)
- **Phase 3**: Coverage calculation (CoverageMetrics)

This design enables:
1. **Partial Success**: Phase 1 succeeds and commits, Phase 2 fails and rolls back, leaving Phase 1 data intact
2. **Independent Isolation**: No blocking between phases due to lock contention
3. **Clear Semantics**: Each phase failure is isolated and audit-traceable

### Isolation Level: Read Committed

Transactions use **SQL Isolation Level: Read Committed** by default.

**Read Committed guarantees:**
- ✅ No dirty reads: Other transactions cannot see uncommitted data
- ⚠️ Non-repeatable reads allowed: Another transaction may modify rows between your reads
- ⚠️ Phantom reads allowed: Another transaction may insert rows between your queries

**Why Read Committed?**
- Good balance between safety and performance for operational data
- Prevents dirty reads (most important guarantee)
- Allows concurrent operations with minimal locking
- Standard for transactional databases

### Critical Error vs Warning Handling

Errors are classified into two categories:

**Critical Errors**: Trigger rollback
- Database constraint violations
- Parsing failures
- Missing required data
- Network errors
- Validation failures

Example:
```go
fnErr := errors.New("failed to create schedule version")
tm.Phase1Transaction(ctx, func(ctx context.Context) error {
    return fnErr  // Triggers ROLLBACK
})
```

**Warnings**: Commit despite error
- Non-fatal validation issues
- Missing optional data
- Data quality warnings

Example:
```go
type WarningError struct{ msg string }
func (w *WarningError) IsWarning() bool { return true }
func (w *WarningError) Error() string { return w.msg }

warningErr := &WarningError{msg: "some assignments missing"}
tm.Phase1Transaction(ctx, func(ctx context.Context) error {
    return warningErr  // Still COMMITS
})
```

## API

### Phase1Transaction (ODS Import)

```go
func (tm *TransactionManager) Phase1Transaction(
    ctx context.Context,
    fn func(context.Context) error,
) error
```

Executes ODS file import within a transaction. The function `fn` receives a context that will be cancelled if the transaction is aborted.

**Usage:**
```go
err := tm.Phase1Transaction(ctx, func(ctx context.Context) error {
    sv, err := odsService.ImportSchedule(ctx, filePath, hospitalID, userID)
    if err != nil {
        return fmt.Errorf("ODS import failed: %w", err)
    }
    return nil
})
if err != nil {
    log.Printf("Phase 1 failed: %v", err)
}
```

### Phase2Transaction (Amion Scraping)

```go
func (tm *TransactionManager) Phase2Transaction(
    ctx context.Context,
    fn func(context.Context) error,
) error
```

Executes Amion scraping within a transaction.

### Phase3Transaction (Coverage Calculation)

```go
func (tm *TransactionManager) Phase3Transaction(
    ctx context.Context,
    fn func(context.Context) error,
) error
```

Executes coverage calculation within a transaction.

### IsolationLevel()

```go
func (tm *TransactionManager) IsolationLevel() sql.IsolationLevel
```

Returns the isolation level (always `sql.LevelReadCommitted`).

## Rollback Strategy

### When Rollback Occurs

A transaction rolls back when:
1. Function `fn` returns a non-nil error
2. The error does NOT implement `Warnable` interface OR `Warnable.IsWarning()` returns false
3. Transaction is still active

### Rollback Guarantees

- **All-or-nothing semantics**: Either all changes in a phase commit or all are rolled back
- **No cascading rollbacks**: Previous phases are never rolled back if a later phase fails
- **Audit trail**: All rollbacks are logged with phase name and original error

### Rollback Failure Handling

If `Rollback()` fails after a critical error:
```
Original error: "failed to create assignment"
Rollback error: "lost connection to database"
Result: Error returned includes both messages
```

## Concurrency & Isolation

### Multiple Concurrent Orchestrations

Different orchestration instances can run concurrently for different schedules:

```go
// Schedule 1
go tm.Phase1Transaction(ctx, fn1)

// Schedule 2 (different hospital)
go tm.Phase2Transaction(ctx, fn2)
```

Each transaction is independent with Read Committed isolation.

### Within Single Orchestration

Phases are sequential, not concurrent. Phase 2 waits for Phase 1 to complete:

```
Phase 1 Transaction [--------COMMIT]
                              |
                              v
Phase 2 Transaction        [---ROLLBACK]
                                      |
                                      v
Phase 3 Transaction (skipped if error)
```

## Error Handling Examples

### Example 1: Phase 1 Success, Phase 2 Failure

```go
// Phase 1 succeeds
err1 := tm.Phase1Transaction(ctx, func(ctx context.Context) error {
    return odsService.ImportSchedule(ctx, filePath, hospitalID, userID)  // returns nil
})
// Phase 1 COMMITTED, err1 = nil

// Phase 2 fails
err2 := tm.Phase2Transaction(ctx, func(ctx context.Context) error {
    return errors.New("amion server unavailable")
})
// Phase 2 ROLLED BACK, err2 = non-nil

// Result: Schedule version from Phase 1 exists in database
//         No assignments from Phase 2 were created
```

### Example 2: Warning in Phase 1

```go
err := tm.Phase1Transaction(ctx, func(ctx context.Context) error {
    // Create shifts...
    // Some validation warnings detected
    return &WarningError{msg: "some staff members missing"}
})
// Phase 1 COMMITTED (despite warning), err = WarningError
// Data is saved, warning is returned to caller
```

### Example 3: Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
cancel()

err := tm.Phase1Transaction(ctx, fn)
// Returns: context.Canceled
// BeginTx was never called
```

## Testing

### Mock Transaction Interface

Tests use `mock.Mock` from testify to simulate transaction behavior:

```go
mockTx := new(MockTxDB)
mockTx.On("Commit").Return(nil)
mockTx.On("Rollback").Return(nil)

mockDB := new(MockDBExec)
mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)

tm := NewTransactionManager(mockDB)
```

### Test Coverage

All critical paths are tested:
- ✅ Successful commit (all 3 phases)
- ✅ Rollback on error (all 3 phases)
- ✅ Warning error (commit despite error)
- ✅ BeginTx failure
- ✅ Commit failure with rollback attempt
- ✅ Context cancellation
- ✅ Partial success (Phase 1 commits, Phase 2 rolls back)
- ✅ Isolation level verification

## Audit Logging

All transaction events are logged using the standard Go `log` package:

```
[phase 1 (ODS import)] Transaction rolled back successfully
[phase 2 (Amion scraping)] Warning: (committing)
[phase 3 (coverage calculation)] Committed
```

For production, replace `log.Default()` with a structured logger (e.g., `zap`).

## Performance Considerations

### Transaction Overhead

Transaction overhead is typically < 1ms per phase:
- BeginTx: ~0.1ms
- Function execution: Variable (dominates)
- Commit: ~0.3ms
- Rollback: ~0.3ms

### Lock Contention

Read Committed isolation minimizes lock duration:
- Locks released at commit (not end of session)
- Row-level locking (not table-level)
- No predicate locks (allows phantom reads)

### Concurrency

Multiple orchestrations don't block each other:
- Each uses separate transactions
- Separate database connections
- Independent isolation scopes

## Future Enhancements

1. **Savepoints**: Add intra-phase savepoints for fine-grained rollback
2. **Transaction Pooling**: Reuse transactions to reduce BeginTx overhead
3. **Metrics**: Add prometheus metrics for transaction duration and failures
4. **Distributed Transactions**: Add support for multi-database transactions
5. **Optimistic Locking**: Add optimistic concurrency control via version columns

## References

- Go `database/sql` documentation: https://golang.org/pkg/database/sql/
- SQL Isolation Levels: https://en.wikipedia.org/wiki/Isolation_(database_systems)
- PostgreSQL Transaction Isolation: https://www.postgresql.org/docs/current/transaction-iso.html
