package helpers

import (
	"testing"
	"time"
)

// TestStartQueryCount initializes query counter.
func TestStartQueryCount(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	if GetQueryCount() != 0 {
		t.Error("StartQueryCount should initialize with 0 queries")
	}
}

// TestGetQueryCount returns correct query count.
func TestGetQueryCount(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	if GetQueryCount() != 0 {
		t.Fatal("initial count should be 0")
	}

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	if GetQueryCount() != 1 {
		t.Error("count should be 1 after one query")
	}

	AppendQuery(QueryRecord{SQL: "SELECT 2"})
	if GetQueryCount() != 2 {
		t.Error("count should be 2 after two queries")
	}
}

// TestGetQueries returns all query records.
func TestGetQueries(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	AppendQuery(QueryRecord{SQL: "SELECT 2"})

	queries := GetQueries()
	if len(queries) != 2 {
		t.Fatalf("expected 2 queries, got %d", len(queries))
	}

	if queries[0].SQL != "SELECT 1" {
		t.Errorf("first query should be 'SELECT 1', got '%s'", queries[0].SQL)
	}

	if queries[1].SQL != "SELECT 2" {
		t.Errorf("second query should be 'SELECT 2', got '%s'", queries[1].SQL)
	}
}

// TestResetQueryCount clears tracked queries.
func TestResetQueryCount(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	if GetQueryCount() != 1 {
		t.Fatal("count should be 1 after append")
	}

	ResetQueryCount()
	if GetQueryCount() != 0 {
		t.Error("count should be 0 after reset")
	}
}

// TestAppendQueryRecordsMetadata ensures query metadata is captured.
func TestAppendQueryRecordsMetadata(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	expected := QueryRecord{
		SQL:   "SELECT * FROM users WHERE id = ?",
		Args:  []interface{}{42},
		Error: nil,
	}

	AppendQuery(expected)

	queries := GetQueries()
	if len(queries) != 1 {
		t.Fatal("should have 1 query")
	}

	actual := queries[0]
	if actual.SQL != expected.SQL {
		t.Errorf("SQL mismatch: expected '%s', got '%s'", expected.SQL, actual.SQL)
	}

	if len(actual.Args) != 1 || actual.Args[0] != 42 {
		t.Errorf("Args mismatch: expected [42], got %v", actual.Args)
	}
}

// TestAssertQueryCount passes when count matches.
func TestAssertQueryCount(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	AppendQuery(QueryRecord{SQL: "SELECT 2"})

	err := AssertQueryCount(2, GetQueryCount())
	if err != nil {
		t.Errorf("assertion should pass for matching count, got error: %v", err)
	}
}

// TestAssertQueryCountMismatch returns error when count doesn't match.
func TestAssertQueryCountMismatch(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	AppendQuery(QueryRecord{SQL: "SELECT 2"})

	err := AssertQueryCount(1, GetQueryCount())
	if err == nil {
		t.Error("assertion should fail for mismatched count")
	}

	if !contains(err.Error(), "expected 1") {
		t.Errorf("error should mention expected count: %v", err)
	}

	if !contains(err.Error(), "got 2") {
		t.Errorf("error should mention actual count: %v", err)
	}
}

// TestAssertQueryCountLEPass passes when actual <= max.
func TestAssertQueryCountLEPass(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	AppendQuery(QueryRecord{SQL: "SELECT 2"})

	// 2 queries, max is 3 - should pass
	err := AssertQueryCountLE(3)
	if err != nil {
		t.Errorf("assertion should pass, got error: %v", err)
	}

	// 2 queries, max is 2 - should pass
	err = AssertQueryCountLE(2)
	if err != nil {
		t.Errorf("assertion should pass for equal counts, got error: %v", err)
	}
}

// TestAssertQueryCountLEFail fails when actual > max.
func TestAssertQueryCountLEFail(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	AppendQuery(QueryRecord{SQL: "SELECT 2"})
	AppendQuery(QueryRecord{SQL: "SELECT 3"})

	// 3 queries, max is 2 - should fail
	err := AssertQueryCountLE(2)
	if err == nil {
		t.Error("assertion should fail when actual > max")
	}

	if !contains(err.Error(), "exceeded maximum") {
		t.Errorf("error message should indicate exceeded limit: %v", err)
	}
}

// TestAssertNoNPlusOnePass passes for non-N+1 patterns.
func TestAssertNoNPlusOnePass(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	// Scenario: Load 10 items with 1 initial query
	// Expected: 10 items * (1 + 0 queries per item) = 10 queries
	for i := 0; i < 10; i++ {
		AppendQuery(QueryRecord{SQL: "SELECT * FROM items"})
	}

	// With batchSize=10, queriesPerItem=0: max = 10 * (1+0) = 10
	err := AssertNoNPlusOne(10, 0)
	if err != nil {
		t.Errorf("assertion should pass for non-N+1 pattern, got error: %v", err)
	}
}

// TestAssertNoNPlusOneDetectsPattern detects N+1 queries.
func TestAssertNoNPlusOneDetectsPattern(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	// Scenario: Load 10 items, then execute 2 queries per item (N+1 pattern)
	// Initial query
	AppendQuery(QueryRecord{SQL: "SELECT * FROM items"})

	// N+1: for each item, execute 2 additional queries
	for i := 0; i < 10; i++ {
		AppendQuery(QueryRecord{SQL: "SELECT * FROM related_table_1"})
		AppendQuery(QueryRecord{SQL: "SELECT * FROM related_table_2"})
	}

	// Total: 1 + 10*2 = 21 queries
	// Expected: 10 * (1 + 1) = 20 (if queriesPerItem=1)
	// Since 21 > 20, should fail

	err := AssertNoNPlusOne(10, 1)
	if err == nil {
		t.Error("assertion should detect N+1 pattern")
	}

	if !contains(err.Error(), "N+1") {
		t.Errorf("error should mention N+1: %v", err)
	}
}

// TestLogQueries formats queries for debugging.
func TestLogQueries(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	AppendQuery(QueryRecord{SQL: "SELECT 2"})

	output := LogQueries()

	if !contains(output, "Total queries: 2") {
		t.Errorf("output should show total count: %s", output)
	}

	if !contains(output, "SELECT 1") {
		t.Errorf("output should include first query: %s", output)
	}

	if !contains(output, "SELECT 2") {
		t.Errorf("output should include second query: %s", output)
	}
}

// TestQueryCounterInactive doesn't track when inactive.
func TestQueryCounterInactive(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	if GetQueryCount() != 1 {
		t.Fatal("should track when active")
	}

	StopQueryCount()
	AppendQuery(QueryRecord{SQL: "SELECT 2"})

	if GetQueryCount() != 1 {
		t.Error("should not track after StopQueryCount")
	}
}

// TestQueryCounterConcurrency ensures thread safety.
func TestQueryCounterConcurrency(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	done := make(chan bool, 100)

	// Launch 100 concurrent appends
	for i := 0; i < 100; i++ {
		go func(id int) {
			AppendQuery(QueryRecord{
				SQL: "SELECT " + string(rune(id)),
			})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	count := GetQueryCount()
	if count != 100 {
		t.Errorf("expected 100 queries, got %d", count)
	}
}

// TestQueryRecordTimestamp captures execution time.
func TestQueryRecordTimestamp(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	before := time.Now()
	time.Sleep(10 * time.Millisecond)

	AppendQuery(QueryRecord{
		SQL:       "SELECT 1",
		Duration:  5 * time.Millisecond,
		Timestamp: time.Now(),
	})

	after := time.Now()

	queries := GetQueries()
	if len(queries) != 1 {
		t.Fatal("should have 1 query")
	}

	ts := queries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp should be between before and after, got %v", ts)
	}
}

// TestQueryRecordError captures query errors.
func TestQueryRecordError(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	testErr := NewTestError("query failed")

	AppendQuery(QueryRecord{
		SQL:   "SELECT * FROM nonexistent",
		Error: testErr,
	})

	queries := GetQueries()
	if len(queries) != 1 {
		t.Fatal("should have 1 query")
	}

	if queries[0].Error == nil {
		t.Error("query should have error recorded")
	}

	if queries[0].Error != testErr {
		t.Errorf("error mismatch: expected %v, got %v", testErr, queries[0].Error)
	}
}

// TestGetQueriesCopyPreventsModification ensures external modifications don't affect counter.
func TestGetQueriesCopyPreventsModification(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	AppendQuery(QueryRecord{SQL: "SELECT 2"})

	queries := GetQueries()
	if len(queries) != 2 {
		t.Fatal("should have 2 queries")
	}

	// Modify the returned slice
	queries[0].SQL = "MODIFIED"
	queries = append(queries, QueryRecord{SQL: "SELECT 3"})

	// Counter should be unchanged
	if GetQueryCount() != 2 {
		t.Error("external modifications should not affect counter")
	}

	newQueries := GetQueries()
	if newQueries[0].SQL != "SELECT 1" {
		t.Error("modifications to returned slice should not affect stored queries")
	}
}

// TestAssertQueryCountDetailedError provides helpful debugging info.
func TestAssertQueryCountDetailedError(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	AppendQuery(QueryRecord{SQL: "SELECT * FROM users"})
	AppendQuery(QueryRecord{SQL: "SELECT * FROM orders"})
	AppendQuery(QueryRecord{SQL: "SELECT * FROM items"})

	err := AssertQueryCount(1, GetQueryCount())
	if err == nil {
		t.Fatal("assertion should fail")
	}

	errMsg := err.Error()
	if !contains(errMsg, "expected 1") || !contains(errMsg, "got 3") {
		t.Errorf("error should show count mismatch: %s", errMsg)
	}

	// Should list queries
	if !contains(errMsg, "SELECT") {
		t.Errorf("error should list queries: %s", errMsg)
	}
}

// TestEmptyQueryLog handles no queries gracefully.
func TestEmptyQueryLog(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	output := LogQueries()
	if !contains(output, "No queries") {
		t.Errorf("should handle empty queries gracefully: %s", output)
	}
}

// TestQueryCounterMultipleStartStopCycles allows reuse.
func TestQueryCounterMultipleStartStopCycles(t *testing.T) {
	// Cycle 1
	ResetQueryCount()
	StartQueryCount()
	AppendQuery(QueryRecord{SQL: "SELECT 1"})
	if GetQueryCount() != 1 {
		t.Fatal("cycle 1 count should be 1")
	}
	StopQueryCount()

	// Cycle 2
	ResetQueryCount()
	StartQueryCount()
	AppendQuery(QueryRecord{SQL: "SELECT 2"})
	if GetQueryCount() != 1 {
		t.Fatal("cycle 2 count should be 1")
	}

	// Cycle 3
	ResetQueryCount()
	StartQueryCount()
	AppendQuery(QueryRecord{SQL: "SELECT 3"})
	AppendQuery(QueryRecord{SQL: "SELECT 4"})
	if GetQueryCount() != 2 {
		t.Fatal("cycle 3 count should be 2")
	}
}

// TestQueryArgsConversion handles various argument types.
func TestQueryArgsConversion(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	args := []interface{}{
		42,
		"string",
		3.14,
		true,
		nil,
	}

	AppendQuery(QueryRecord{
		SQL:  "SELECT * FROM table WHERE id=? AND name=? AND price=? AND active=? AND optional=?",
		Args: args,
	})

	queries := GetQueries()
	if len(queries) != 1 {
		t.Fatal("should have 1 query")
	}

	if len(queries[0].Args) != 5 {
		t.Fatalf("should have 5 args, got %d", len(queries[0].Args))
	}

	if queries[0].Args[0] != 42 {
		t.Errorf("arg 0 should be 42, got %v", queries[0].Args[0])
	}

	if queries[0].Args[1] != "string" {
		t.Errorf("arg 1 should be 'string', got %v", queries[0].Args[1])
	}
}

// TestLongQueryTruncation truncates long queries in error messages.
func TestLongQueryTruncation(t *testing.T) {
	ResetQueryCount()
	StartQueryCount()

	longQuery := "SELECT * FROM very_long_table_name_that_goes_on_and_on WHERE id IN (SELECT long_id FROM another_very_long_table WHERE condition = true AND another_condition = false)"
	AppendQuery(QueryRecord{SQL: longQuery})

	err := AssertQueryCount(0, GetQueryCount())
	if err == nil {
		t.Fatal("assertion should fail")
	}

	errMsg := err.Error()
	// Check that query is truncated
	if !contains(errMsg, "...") {
		t.Errorf("long queries should be truncated in error message: %s", errMsg)
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr ||
		 len(s) >= len(substr) &&
		 (s[:len(substr)] == substr ||
		  (len(s) > len(substr) &&
		   findIndex(s, substr) >= 0)))
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// TestError is a simple error for testing.
type TestError struct {
	msg string
}

func NewTestError(msg string) *TestError {
	return &TestError{msg: msg}
}

func (e *TestError) Error() string {
	return e.msg
}
