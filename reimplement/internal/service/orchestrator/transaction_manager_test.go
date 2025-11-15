package orchestrator

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTxDB is a mock transaction for testing
type MockTxDB struct {
	mock.Mock
}

func (m *MockTxDB) Commit() error {
	return m.Called().Error(0)
}

func (m *MockTxDB) Rollback() error {
	return m.Called().Error(0)
}

// MockDBExec is a mock database executor for testing
type MockDBExec struct {
	mock.Mock
}

func (m *MockDBExec) BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Transaction), args.Error(1)
}

func TestPhase1SuccessfulCommit(t *testing.T) {
	mockTx := new(MockTxDB)
	mockTx.On("Commit").Return(nil)

	mockDB := new(MockDBExec)
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)

	tm := NewTransactionManager(mockDB)

	executed := false
	fn := func(ctx context.Context) error {
		executed = true
		return nil
	}

	err := tm.Phase1Transaction(context.Background(), fn)

	require.NoError(t, err)
	assert.True(t, executed)
	mockTx.AssertCalled(t, "Commit")
	mockTx.AssertNotCalled(t, "Rollback")
}

func TestPhase2SuccessfulCommit(t *testing.T) {
	mockTx := new(MockTxDB)
	mockTx.On("Commit").Return(nil)

	mockDB := new(MockDBExec)
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)

	tm := NewTransactionManager(mockDB)

	executed := false
	fn := func(ctx context.Context) error {
		executed = true
		return nil
	}

	err := tm.Phase2Transaction(context.Background(), fn)

	require.NoError(t, err)
	assert.True(t, executed)
	mockTx.AssertCalled(t, "Commit")
	mockTx.AssertNotCalled(t, "Rollback")
}

func TestPhase3SuccessfulCommit(t *testing.T) {
	mockTx := new(MockTxDB)
	mockTx.On("Commit").Return(nil)

	mockDB := new(MockDBExec)
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)

	tm := NewTransactionManager(mockDB)

	executed := false
	fn := func(ctx context.Context) error {
		executed = true
		return nil
	}

	err := tm.Phase3Transaction(context.Background(), fn)

	require.NoError(t, err)
	assert.True(t, executed)
	mockTx.AssertCalled(t, "Commit")
	mockTx.AssertNotCalled(t, "Rollback")
}

func TestPhase1RollbackOnError(t *testing.T) {
	mockTx := new(MockTxDB)
	mockTx.On("Rollback").Return(nil)

	mockDB := new(MockDBExec)
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)

	tm := NewTransactionManager(mockDB)

	expectedErr := errors.New("error")
	fn := func(ctx context.Context) error {
		return expectedErr
	}

	err := tm.Phase1Transaction(context.Background(), fn)

	require.Error(t, err)
	mockTx.AssertNotCalled(t, "Commit")
	mockTx.AssertCalled(t, "Rollback")
}

func TestPhase2RollbackOnError(t *testing.T) {
	mockTx := new(MockTxDB)
	mockTx.On("Rollback").Return(nil)

	mockDB := new(MockDBExec)
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)

	tm := NewTransactionManager(mockDB)

	fn := func(ctx context.Context) error {
		return errors.New("error")
	}

	err := tm.Phase2Transaction(context.Background(), fn)

	require.Error(t, err)
	mockTx.AssertNotCalled(t, "Commit")
	mockTx.AssertCalled(t, "Rollback")
}

func TestPhase3RollbackOnError(t *testing.T) {
	mockTx := new(MockTxDB)
	mockTx.On("Rollback").Return(nil)

	mockDB := new(MockDBExec)
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)

	tm := NewTransactionManager(mockDB)

	fn := func(ctx context.Context) error {
		return errors.New("error")
	}

	err := tm.Phase3Transaction(context.Background(), fn)

	require.Error(t, err)
	mockTx.AssertNotCalled(t, "Commit")
	mockTx.AssertCalled(t, "Rollback")
}

func TestWarningError(t *testing.T) {
	mockTx := new(MockTxDB)
	mockTx.On("Commit").Return(nil)

	mockDB := new(MockDBExec)
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)

	tm := NewTransactionManager(mockDB)

	warningErr := &TestWarnErr{msg: "warning"}
	fn := func(ctx context.Context) error {
		return warningErr
	}

	err := tm.Phase1Transaction(context.Background(), fn)

	require.Error(t, err)
	mockTx.AssertCalled(t, "Commit")
	mockTx.AssertNotCalled(t, "Rollback")
}

func TestBeginFailure(t *testing.T) {
	mockDB := new(MockDBExec)
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(nil, errors.New("begin error"))

	tm := NewTransactionManager(mockDB)

	executed := false
	fn := func(ctx context.Context) error {
		executed = true
		return nil
	}

	err := tm.Phase1Transaction(context.Background(), fn)

	require.Error(t, err)
	assert.False(t, executed)
}

func TestCommitFailure(t *testing.T) {
	mockTx := new(MockTxDB)
	mockTx.On("Commit").Return(errors.New("commit error"))
	mockTx.On("Rollback").Return(nil)

	mockDB := new(MockDBExec)
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)

	tm := NewTransactionManager(mockDB)

	fn := func(ctx context.Context) error {
		return nil
	}

	err := tm.Phase1Transaction(context.Background(), fn)

	require.Error(t, err)
	mockTx.AssertCalled(t, "Commit")
	mockTx.AssertCalled(t, "Rollback")
}

func TestContextCancellation(t *testing.T) {
	mockDB := new(MockDBExec)

	tm := NewTransactionManager(mockDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fn := func(ctx context.Context) error {
		return nil
	}

	err := tm.Phase1Transaction(ctx, fn)

	require.Error(t, err)
	mockDB.AssertNotCalled(t, "BeginTx")
}

func TestPartialSuccess(t *testing.T) {
	mockTx1 := new(MockTxDB)
	mockTx1.On("Commit").Return(nil)
	mockDB1 := new(MockDBExec)
	mockDB1.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx1, nil)
	tm1 := NewTransactionManager(mockDB1)

	mockTx2 := new(MockTxDB)
	mockTx2.On("Rollback").Return(nil)
	mockDB2 := new(MockDBExec)
	mockDB2.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx2, nil)
	tm2 := NewTransactionManager(mockDB2)

	fn1 := func(ctx context.Context) error { return nil }
	fn2 := func(ctx context.Context) error { return errors.New("phase 2 error") }

	err1 := tm1.Phase1Transaction(context.Background(), fn1)
	err2 := tm2.Phase2Transaction(context.Background(), fn2)

	require.NoError(t, err1)
	require.Error(t, err2)

	mockTx1.AssertCalled(t, "Commit")
	mockTx1.AssertNotCalled(t, "Rollback")

	mockTx2.AssertCalled(t, "Rollback")
	mockTx2.AssertNotCalled(t, "Commit")
}

func TestIsolationLevel(t *testing.T) {
	mockDB := new(MockDBExec)
	tm := NewTransactionManager(mockDB)

	assert.Equal(t, sql.LevelReadCommitted, tm.IsolationLevel())
}

type TestWarnErr struct {
	msg string
}

func (w *TestWarnErr) Error() string {
	return w.msg
}

func (w *TestWarnErr) IsWarning() bool {
	return true
}
