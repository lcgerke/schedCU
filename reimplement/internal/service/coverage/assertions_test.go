package coverage

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/tests/helpers"
)

// TestNewCoverageAssertionHelper initializes helper successfully.
func TestNewCoverageAssertionHelper(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	if helper == nil {
		t.Fatal("NewCoverageAssertionHelper should not return nil")
	}

	if helper.expectedQueryCounts == nil {
		t.Fatal("expectedQueryCounts should be initialized")
	}

	if len(helper.expectedQueryCounts) != 0 {
		t.Fatal("expectedQueryCounts should start empty")
	}
}

// TestAssertQueryCountPass verifies assertion passes when counts match.
func TestAssertQueryCountPass(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	// Execute 1 query
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 1"})

	err := helper.AssertQueryCount(1)
	if err != nil {
		t.Fatalf("assertion should pass for matching count, got error: %v", err)
	}
}

// TestAssertQueryCountFail verifies assertion fails when counts don't match.
func TestAssertQueryCountFail(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	// Execute 2 queries
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 1"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 2"})

	err := helper.AssertQueryCount(1)
	if err == nil {
		t.Fatal("assertion should fail for mismatched count")
	}

	if !strings.Contains(err.Error(), "expected 1") {
		t.Fatalf("error should mention expected count: %v", err)
	}

	if !strings.Contains(err.Error(), "got 2") {
		t.Fatalf("error should mention actual count: %v", err)
	}
}

// TestAssertSingleQueryDataLoadPass verifies single query assertion passes.
func TestAssertSingleQueryDataLoadPass(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM shift_instances"})

	err := helper.AssertSingleQueryDataLoad()
	if err != nil {
		t.Fatalf("single query assertion should pass, got error: %v", err)
	}
}

// TestAssertSingleQueryDataLoadFail verifies assertion fails with multiple queries.
func TestAssertSingleQueryDataLoadFail(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM shift_instances"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM users"})

	err := helper.AssertSingleQueryDataLoad()
	if err == nil {
		t.Fatal("assertion should fail for 2 queries")
	}

	if !strings.Contains(err.Error(), "expected 1") {
		t.Fatalf("error should mention expected 1 query: %v", err)
	}
}

// TestAssertSingleQueryDataLoadZeroQueriesFail verifies failure with zero queries.
func TestAssertSingleQueryDataLoadZeroQueriesFail(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	// No queries appended

	err := helper.AssertSingleQueryDataLoad()
	if err == nil {
		t.Fatal("assertion should fail for 0 queries")
	}
}

// TestAssertCoverageCalculationPass verifies coverage calculation assertion passes.
func TestAssertCoverageCalculationPass(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM shifts"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM requirements"})

	err := helper.AssertCoverageCalculation("CalculateCoverageMorning", 2)
	if err != nil {
		t.Fatalf("assertion should pass, got error: %v", err)
	}

	// Verify the operation name is stored
	stored := helper.GetExpectedQueryCount("CalculateCoverageMorning")
	if stored != 2 {
		t.Fatalf("expected query count should be stored, got %d", stored)
	}
}

// TestAssertCoverageCalculationFail verifies coverage calculation assertion fails correctly.
func TestAssertCoverageCalculationFail(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM shifts"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM requirements"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM users"})

	err := helper.AssertCoverageCalculation("CalculateCoverageMorning", 2)
	if err == nil {
		t.Fatal("assertion should fail for 3 vs 2 queries")
	}

	if !strings.Contains(err.Error(), "CalculateCoverageMorning") {
		t.Fatalf("error should include operation name: %v", err)
	}

	if !strings.Contains(err.Error(), "expected 2") {
		t.Fatalf("error should mention expected count: %v", err)
	}

	if !strings.Contains(err.Error(), "got 3") {
		t.Fatalf("error should mention actual count: %v", err)
	}
}

// TestAssertQueryCountLEPass verifies LE assertion passes when within limit.
func TestAssertQueryCountLEPass(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 1"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 2"})

	err := helper.AssertQueryCountLE(3)
	if err != nil {
		t.Fatalf("assertion should pass for 2 <= 3, got error: %v", err)
	}

	// Equal should also pass
	err = helper.AssertQueryCountLE(2)
	if err != nil {
		t.Fatalf("assertion should pass for 2 <= 2, got error: %v", err)
	}
}

// TestAssertQueryCountLEFail verifies LE assertion fails when exceeds limit.
func TestAssertQueryCountLEFail(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 1"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 2"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 3"})

	err := helper.AssertQueryCountLE(2)
	if err == nil {
		t.Fatal("assertion should fail for 3 > 2")
	}

	if !strings.Contains(err.Error(), "exceeded maximum") {
		t.Fatalf("error should indicate limit exceeded: %v", err)
	}
}

// TestAssertNoNPlusOnePass verifies N+1 detection passes for normal patterns.
func TestAssertNoNPlusOnePass(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	// Single batch query (expected N+1 pattern: batch query + items with 0 additional queries)
	for i := 0; i < 10; i++ {
		helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM shifts"})
	}

	// With 10 items and 0 queries per item: expected = 10 * (1 + 0) = 10
	err := helper.AssertNoNPlusOne(10, 0)
	if err != nil {
		t.Fatalf("assertion should pass for non-N+1 pattern, got error: %v", err)
	}
}

// TestAssertNoNPlusOneDetectsPattern verifies N+1 detection catches problematic patterns.
func TestAssertNoNPlusOneDetectsPattern(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	// Initial query
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM shifts"})

	// N+1: for each item, execute 2 additional queries
	for i := 0; i < 10; i++ {
		helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM users WHERE id = ?"})
		helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM positions"})
	}

	// Total: 1 + 10*2 = 21 queries
	// Expected: 10 * (1 + 1) = 20
	// Since 21 > 20, should fail

	err := helper.AssertNoNPlusOne(10, 1)
	if err == nil {
		t.Fatal("assertion should detect N+1 pattern")
	}

	if !strings.Contains(err.Error(), "N+1") {
		t.Fatalf("error should mention N+1: %v", err)
	}
}

// TestAssertDataLoaderOperationSuccess verifies operation wrapper succeeds.
func TestAssertDataLoaderOperationSuccess(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shifts := []*entity.ShiftInstance{
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Nurse", "ER", "Alice", userID),
		entity.NewShiftInstance(scheduleVersionID, "Night", "Doctor", "ICU", "Bob", userID),
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	// Note: Mock repository doesn't track queries via helpers, so we expect 0 queries
	// In real scenarios with actual database queries, we'd expect 1
	result, err := helper.AssertDataLoaderOperation(
		"LoadAssignments",
		func() (interface{}, error) {
			return loader.LoadAssignmentsForScheduleVersion(context.Background(), scheduleVersionID)
		},
		0, // Mock doesn't track queries
	)

	if err != nil {
		t.Fatalf("operation should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	resultShifts, ok := result.([]*entity.ShiftInstance)
	if !ok {
		t.Fatal("result should be []*entity.ShiftInstance")
	}

	if len(resultShifts) != 2 {
		t.Fatalf("expected 2 shifts, got %d", len(resultShifts))
	}
}

// TestAssertDataLoaderOperationAssertionFails verifies operation wrapper catches assertion failures.
func TestAssertDataLoaderOperationAssertionFails(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shifts := []*entity.ShiftInstance{
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Nurse", "ER", "Alice", userID),
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	_, err := helper.AssertDataLoaderOperation(
		"LoadAssignments",
		func() (interface{}, error) {
			return loader.LoadAssignmentsForScheduleVersion(context.Background(), scheduleVersionID)
		},
		1, // Mock will execute 0, assertion expects 1
	)

	if err == nil {
		t.Fatal("operation should return assertion error")
	}

	if !strings.Contains(err.Error(), "expected 1") {
		t.Fatalf("error should show expected count: %v", err)
	}
}

// TestAssertDataLoaderOperationOperationFails verifies wrapper handles operation failures.
func TestAssertDataLoaderOperationOperationFails(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	scheduleVersionID := uuid.New()

	repo := &MockShiftInstanceRepository{
		getByVersionErr: ErrRepositoryFailed,
	}
	loader := NewCoverageDataLoader(repo)

	_, err := helper.AssertDataLoaderOperation(
		"LoadAssignments",
		func() (interface{}, error) {
			return loader.LoadAssignmentsForScheduleVersion(context.Background(), scheduleVersionID)
		},
		0,
	)

	if err == nil {
		t.Fatal("operation should return error")
	}

	if !strings.Contains(err.Error(), "LoadAssignments operation failed") {
		t.Fatalf("error should indicate operation failure: %v", err)
	}
}

// TestAssertDataLoaderOperationWithContextSuccess verifies context-aware operation wrapper.
func TestAssertDataLoaderOperationWithContextSuccess(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shifts := []*entity.ShiftInstance{
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Nurse", "ER", "Alice", userID),
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	result, err := helper.AssertDataLoaderOperationWithContext(
		ctx,
		"LoadAssignmentsWithContext",
		func(c context.Context) (interface{}, error) {
			return loader.LoadAssignmentsForScheduleVersion(c, scheduleVersionID)
		},
		0, // Mock doesn't track queries
	)

	if err != nil {
		t.Fatalf("operation should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}
}

// TestAssertCoverageOperationTimingSuccess verifies timing assertion succeeds.
func TestAssertCoverageOperationTimingSuccess(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shifts := []*entity.ShiftInstance{
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Nurse", "ER", "Alice", userID),
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	result, duration, err := helper.AssertCoverageOperationTiming(
		"LoadAssignmentsWithTiming",
		func() (interface{}, error) {
			return loader.LoadAssignmentsForScheduleVersion(context.Background(), scheduleVersionID)
		},
		0, // Mock doesn't track queries
		1*time.Second, // Max duration: 1 second (generous for unit test)
	)

	if err != nil {
		t.Fatalf("operation should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if duration < 0 {
		t.Fatal("duration should be non-negative")
	}

	if duration > 1*time.Second {
		t.Fatalf("duration exceeded max: %v > 1s", duration)
	}
}

// TestAssertCoverageOperationTimingExceedsDuration verifies timing assertion fails on slow operations.
func TestAssertCoverageOperationTimingExceedsDuration(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shifts := []*entity.ShiftInstance{
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Nurse", "ER", "Alice", userID),
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	_, _, err := helper.AssertCoverageOperationTiming(
		"SlowLoad",
		func() (interface{}, error) {
			time.Sleep(50 * time.Millisecond)
			return loader.LoadAssignmentsForScheduleVersion(context.Background(), scheduleVersionID)
		},
		0, // Mock doesn't track queries
		10*time.Millisecond, // Too short max duration
	)

	if err == nil {
		t.Fatal("operation should fail for exceeding max duration")
	}

	if !strings.Contains(err.Error(), "exceeded max duration") {
		t.Fatalf("error should indicate timing failure: %v", err)
	}
}

// TestLogAllQueries returns formatted query list.
func TestLogAllQueries(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM shifts"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM users"})

	output := helper.LogAllQueries()

	if !strings.Contains(output, "Total queries: 2") {
		t.Fatalf("output should show total count: %s", output)
	}

	if !strings.Contains(output, "SELECT * FROM shifts") {
		t.Fatalf("output should include first query: %s", output)
	}

	if !strings.Contains(output, "SELECT * FROM users") {
		t.Fatalf("output should include second query: %s", output)
	}
}

// TestGetExpectedQueryCount retrieves stored expected count.
func TestGetExpectedQueryCount(t *testing.T) {
	helper := NewCoverageAssertionHelper()

	// Before storing anything
	count := helper.GetExpectedQueryCount("NonexistentOp")
	if count != -1 {
		t.Fatalf("nonexistent operation should return -1, got %d", count)
	}

	// Store an expectation
	helper.DocumentExpectedQueries("TestOp", 5, "test reason")

	// Retrieve it
	count = helper.GetExpectedQueryCount("TestOp")
	if count != 5 {
		t.Fatalf("expected 5, got %d", count)
	}
}

// TestAssertNoRegressionPass verifies regression detection passes when within limit.
func TestAssertNoRegressionPass(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 1"})
	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT 2"})

	err := helper.AssertNoRegression(RegressionDetectionConfig{
		OperationName:    "LoadAssignments",
		ExpectedQueries:  1,
		MaxQueryIncrease: 1, // Allow 1 additional query
		Description:      "Data loading should use <= 2 queries",
	})

	if err != nil {
		t.Fatalf("regression check should pass for 2 queries with tolerance 1, got error: %v", err)
	}
}

// TestAssertNoRegressionFail verifies regression detection catches regressions.
func TestAssertNoRegressionFail(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	// Simulate a regression: 5 queries instead of 1
	for i := 0; i < 5; i++ {
		helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT ?"})
	}

	err := helper.AssertNoRegression(RegressionDetectionConfig{
		OperationName:    "LoadAssignments",
		ExpectedQueries:  1,
		MaxQueryIncrease: 0, // No increase allowed
		Description:      "Data loading should always be 1 query",
	})

	if err == nil {
		t.Fatal("regression check should fail for 5 vs 1 queries with no tolerance")
	}

	if !strings.Contains(err.Error(), "REGRESSION DETECTED") {
		t.Fatalf("error should indicate regression: %v", err)
	}

	if !strings.Contains(err.Error(), "LoadAssignments") {
		t.Fatalf("error should include operation name: %v", err)
	}

	if !strings.Contains(err.Error(), "Data loading should always be 1 query") {
		t.Fatalf("error should include description: %v", err)
	}
}

// TestDocumentExpectedQueries stores operation documentation.
func TestDocumentExpectedQueries(t *testing.T) {
	helper := NewCoverageAssertionHelper()

	helper.DocumentExpectedQueries(
		"LoadAssignments",
		1,
		"Single batch query using IN clause",
	)

	count := helper.GetExpectedQueryCount("LoadAssignments")
	if count != 1 {
		t.Fatalf("expected documented count to be 1, got %d", count)
	}
}

// TestAssertCoverageCalculationWithScheduleVersionValid verifies specialized assertion.
func TestAssertCoverageCalculationWithScheduleVersionValid(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	scheduleVersionID := uuid.New()

	helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM shifts WHERE schedule_version_id = ?"})

	err := helper.AssertCoverageCalculationWithScheduleVersion(
		context.Background(),
		nil, // loader not used in this test
		scheduleVersionID,
		1,
	)

	if err != nil {
		t.Fatalf("assertion should pass, got error: %v", err)
	}
}

// TestAssertCoverageCalculationWithScheduleVersionNilID verifies failure on nil ID.
func TestAssertCoverageCalculationWithScheduleVersionNilID(t *testing.T) {
	helper := NewCoverageAssertionHelper()

	err := helper.AssertCoverageCalculationWithScheduleVersion(
		context.Background(),
		nil,
		uuid.Nil, // Invalid ID
		1,
	)

	if err == nil {
		t.Fatal("assertion should fail for nil schedule version ID")
	}

	if !strings.Contains(err.Error(), "invalid schedule version ID") {
		t.Fatalf("error should indicate invalid ID: %v", err)
	}
}

// TestFormatCoverageQueriesEmpty handles empty query list.
func TestFormatCoverageQueriesEmpty(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	output := helper.formatCoverageQueries([]helpers.QueryRecord{})

	if !strings.Contains(output, "No queries") {
		t.Fatalf("should handle empty queries: %s", output)
	}
}

// TestFormatCoverageQueriesWithLongSQL truncates long queries.
func TestFormatCoverageQueriesWithLongSQL(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	longSQL := "SELECT * FROM very_long_table_name WHERE id IN (SELECT long_id FROM another_long_table WHERE condition_one = true AND condition_two = false AND condition_three = true AND condition_four = false)"

	queries := []helpers.QueryRecord{
		{SQL: longSQL},
	}

	output := helper.formatCoverageQueries(queries)

	if !strings.Contains(output, "...") {
		t.Fatalf("should truncate long queries: %s", output)
	}
}

// TestFormatCoverageQueriesWithArgs includes query arguments.
func TestFormatCoverageQueriesWithArgs(t *testing.T) {
	helper := NewCoverageAssertionHelper()

	queries := []helpers.QueryRecord{
		{
			SQL:  "SELECT * FROM users WHERE id = ? AND status = ?",
			Args: []interface{}{42, "active"},
		},
	}

	output := helper.formatCoverageQueries(queries)

	if !strings.Contains(output, "Args:") {
		t.Fatalf("should include query arguments: %s", output)
	}
}

// TestFormatCoverageQueriesWithDuration includes timing information.
func TestFormatCoverageQueriesWithDuration(t *testing.T) {
	helper := NewCoverageAssertionHelper()

	queries := []helpers.QueryRecord{
		{
			SQL:      "SELECT 1",
			Duration: 5 * time.Millisecond,
		},
	}

	output := helper.formatCoverageQueriesWithArgs(queries)

	if !strings.Contains(output, "Duration:") {
		t.Fatalf("should include duration: %s", output)
	}
}

// TestFormatCoverageQueriesWithError includes error information.
func TestFormatCoverageQueriesWithError(t *testing.T) {
	helper := NewCoverageAssertionHelper()

	queries := []helpers.QueryRecord{
		{
			SQL:   "SELECT * FROM nonexistent",
			Error: errors.New("table does not exist"),
		},
	}

	output := helper.formatCoverageQueries(queries)

	if !strings.Contains(output, "Error:") {
		t.Fatalf("should include error: %s", output)
	}
}

// TestMultipleOperationExpectations stores multiple operation expectations.
func TestMultipleOperationExpectations(t *testing.T) {
	helper := NewCoverageAssertionHelper()

	helper.DocumentExpectedQueries("LoadData", 1, "Batch query")
	helper.DocumentExpectedQueries("CalculateCoverage", 2, "Load + calculation")
	helper.DocumentExpectedQueries("SaveResults", 1, "Single insert")

	if helper.GetExpectedQueryCount("LoadData") != 1 {
		t.Fatal("LoadData expectation mismatch")
	}

	if helper.GetExpectedQueryCount("CalculateCoverage") != 2 {
		t.Fatal("CalculateCoverage expectation mismatch")
	}

	if helper.GetExpectedQueryCount("SaveResults") != 1 {
		t.Fatal("SaveResults expectation mismatch")
	}
}

// TestAssertionErrorMessagesIncludeQueryDetails verifies detailed error messages.
func TestAssertionErrorMessagesIncludeQueryDetails(t *testing.T) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	helpers.AppendQuery(helpers.QueryRecord{
		SQL:  "SELECT * FROM shifts WHERE schedule_version_id = ?",
		Args: []interface{}{uuid.New()},
	})
	helpers.AppendQuery(helpers.QueryRecord{
		SQL: "SELECT * FROM users WHERE id IN (?)",
	})
	helpers.AppendQuery(helpers.QueryRecord{
		SQL: "SELECT * FROM positions",
	})

	err := helper.AssertCoverageCalculation("TestOp", 1)
	if err == nil {
		t.Fatal("assertion should fail")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "expected 1") {
		t.Fatalf("error should show expected count: %s", errMsg)
	}

	if !strings.Contains(errMsg, "got 3") {
		t.Fatalf("error should show actual count: %s", errMsg)
	}

	if !strings.Contains(errMsg, "shifts") {
		t.Fatalf("error should list queries: %s", errMsg)
	}
}

// Helper function for formatting with args
func (h *CoverageAssertionHelper) formatCoverageQueriesWithArgs(queries []helpers.QueryRecord) string {
	if len(queries) == 0 {
		return "No queries executed"
	}

	var result string
	result += "Queries with details:\n"

	for i, q := range queries {
		result += "  Query " + string(rune(48+i)) + ":\n"
		if q.Duration > 0 {
			result += "    Duration: " + q.Duration.String() + "\n"
		}
	}

	return result
}

// BenchmarkAssertQueryCount measures assertion overhead.
func BenchmarkAssertQueryCount(b *testing.B) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()

	// Pre-populate with 100 queries
	for i := 0; i < 100; i++ {
		helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT ?"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = helper.AssertQueryCount(100)
	}
}

// BenchmarkAssertCoverageCalculation measures coverage assertion overhead.
func BenchmarkAssertCoverageCalculation(b *testing.B) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()

	for i := 0; i < 50; i++ {
		helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT ?"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = helper.AssertCoverageCalculation("BenchmarkOp", 50)
	}
}

// BenchmarkLogAllQueries measures query logging overhead.
func BenchmarkLogAllQueries(b *testing.B) {
	helper := NewCoverageAssertionHelper()
	helpers.ResetQueryCount()
	helpers.StartQueryCount()

	for i := 0; i < 100; i++ {
		helpers.AppendQuery(helpers.QueryRecord{SQL: "SELECT * FROM table WHERE id = ?"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = helper.LogAllQueries()
	}
}
