// Package coverage provides coverage calculation services for schedule management.
package coverage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/tests/helpers"
)

// CoverageAssertionHelper provides assertion utilities for coverage operations.
// It wraps coverage calculation operations and asserts query execution counts.
// This helps detect N+1 query patterns and performance regressions.
type CoverageAssertionHelper struct {
	mu sync.RWMutex
	// expectedQueryCounts stores the last expected count for regression detection
	expectedQueryCounts map[string]int
}

// NewCoverageAssertionHelper creates a new coverage assertion helper.
func NewCoverageAssertionHelper() *CoverageAssertionHelper {
	return &CoverageAssertionHelper{
		expectedQueryCounts: make(map[string]int),
	}
}

// AssertQueryCount verifies that an operation executed exactly the expected number of queries.
// This is the primary assertion method for coverage operations.
//
// Returns:
//   - nil if assertion passes
//   - error with detailed query information if assertion fails
//
// Example:
//
//	helper := NewCoverageAssertionHelper()
//	helpers.StartQueryCount()
//	defer helpers.StopQueryCount()
//
//	shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
//
//	if err := helper.AssertQueryCount(1); err != nil {
//		t.Fatalf("unexpected query count: %v", err)
//	}
func (h *CoverageAssertionHelper) AssertQueryCount(expected int) error {
	actual := helpers.GetQueryCount()
	return helpers.AssertQueryCount(expected, actual)
}

// AssertSingleQueryDataLoad asserts that data loading uses exactly 1 query.
// This is the expected behavior for batch query pattern implementations.
// Fails the test if query count differs from 1.
//
// This helper makes the intention clear in tests: "I expect data loading to be a single batch query"
//
// Example:
//
//	helpers.StartQueryCount()
//	shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
//	helpers.StopQueryCount()
//
//	if err := helper.AssertSingleQueryDataLoad(); err != nil {
//		t.Fatalf("data load should use 1 query: %v", err)
//	}
func (h *CoverageAssertionHelper) AssertSingleQueryDataLoad() error {
	return h.AssertQueryCount(1)
}

// AssertCoverageCalculation asserts that a coverage calculation uses the expected number of queries.
// Useful for regression detection: documents what queries are expected for a specific calculation.
//
// operationName: Descriptive name for the operation (e.g., "CalculateCoverageMorning")
// expectedQueries: Expected query count for this operation
//
// Returns:
//   - nil if assertion passes
//   - error with detailed query information if assertion fails
//
// Example:
//
//	helpers.StartQueryCount()
//	defer helpers.StopQueryCount()
//
//	// Perform coverage calculation
//	metrics := CalculateCoverage(shifts, requirements)
//
//	if err := helper.AssertCoverageCalculation("CalculateCoverageMorning", 1); err != nil {
//		t.Fatalf("unexpected query count in coverage calculation: %v", err)
//	}
func (h *CoverageAssertionHelper) AssertCoverageCalculation(operationName string, expectedQueries int) error {
	h.mu.Lock()
	h.expectedQueryCounts[operationName] = expectedQueries
	h.mu.Unlock()

	actual := helpers.GetQueryCount()
	if expectedQueries == actual {
		return nil
	}

	queries := helpers.GetQueries()
	queryDetails := h.formatCoverageQueries(queries)

	return fmt.Errorf(
		"coverage calculation %q: expected %d queries, got %d\n%s",
		operationName,
		expectedQueries,
		actual,
		queryDetails,
	)
}

// AssertQueryCountLE asserts that query count doesn't exceed maximum.
// Useful for regression detection: ensures queries don't increase unexpectedly.
//
// maxExpected: Maximum allowed query count
//
// Returns:
//   - nil if actual <= maxExpected
//   - error with detailed query information if assertion fails
//
// Example:
//
//	if err := helper.AssertQueryCountLE(5); err != nil {
//		t.Fatalf("query count regression: %v", err)
//	}
func (h *CoverageAssertionHelper) AssertQueryCountLE(maxExpected int) error {
	return helpers.AssertQueryCountLE(maxExpected)
}

// AssertNoNPlusOne detects N+1 query patterns in coverage operations.
// Useful for batch operation validations.
//
// expectedBatchSize: Expected number of items being processed
// expectedQueriesPerItem: Expected additional queries per item
//
// Returns:
//   - nil if no N+1 pattern detected
//   - error if actual queries exceed the expected threshold
//
// Example:
//
//	helpers.StartQueryCount()
//	defer helpers.StopQueryCount()
//
//	shifts, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
//	metrics := CalculateCoverage(shifts, requirements)
//
//	// With 100 shifts and 0 queries per shift (full batch), expect 1+100 queries
//	if err := helper.AssertNoNPlusOne(100, 0); err != nil {
//		t.Fatalf("N+1 pattern detected: %v", err)
//	}
func (h *CoverageAssertionHelper) AssertNoNPlusOne(expectedBatchSize int, expectedQueriesPerItem int) error {
	return helpers.AssertNoNPlusOne(expectedBatchSize, expectedQueriesPerItem)
}

// AssertDataLoaderOperation wraps a data loading operation and asserts query counts.
// This is a convenience method for the common pattern: start tracking, load data, stop tracking, assert.
//
// operationName: Descriptive name for logging/error messages
// operation: The function to execute (should execute queries)
// expectedQueries: Expected number of queries the operation should execute
//
// Returns:
//   - The operation result (as interface{} for flexibility)
//   - An error if the operation fails or query count assertion fails
//
// Example:
//
//	result, err := helper.AssertDataLoaderOperation(
//		"LoadAssignmentsForScheduleVersion",
//		func() (interface{}, error) {
//			return loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
//		},
//		1, // Expected 1 query
//	)
//	if err != nil {
//		t.Fatalf("assertion failed: %v", err)
//	}
func (h *CoverageAssertionHelper) AssertDataLoaderOperation(
	operationName string,
	operation func() (interface{}, error),
	expectedQueries int,
) (interface{}, error) {
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	result, err := operation()
	if err != nil {
		return nil, fmt.Errorf("%s operation failed: %w", operationName, err)
	}

	if assertErr := h.AssertCoverageCalculation(operationName, expectedQueries); assertErr != nil {
		return nil, assertErr
	}

	return result, nil
}

// AssertDataLoaderOperationWithContext wraps a data loading operation with context support.
// Same as AssertDataLoaderOperation but accepts a context parameter.
//
// Example:
//
//	result, err := helper.AssertDataLoaderOperationWithContext(
//		ctx,
//		"LoadAssignmentsForScheduleVersion",
//		func(ctx context.Context) (interface{}, error) {
//			return loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
//		},
//		1, // Expected 1 query
//	)
func (h *CoverageAssertionHelper) AssertDataLoaderOperationWithContext(
	ctx context.Context,
	operationName string,
	operation func(context.Context) (interface{}, error),
	expectedQueries int,
) (interface{}, error) {
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	result, err := operation(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s operation failed: %w", operationName, err)
	}

	if assertErr := h.AssertCoverageCalculation(operationName, expectedQueries); assertErr != nil {
		return nil, assertErr
	}

	return result, nil
}

// AssertCoverageOperationTiming verifies both query count and execution time.
// Useful for performance regression detection.
//
// operationName: Descriptive name for the operation
// operation: The function to execute
// expectedQueries: Expected number of queries
// maxDuration: Maximum acceptable duration
//
// Returns:
//   - The operation result (as interface{} for flexibility)
//   - The actual duration
//   - An error if operation fails or either assertion fails
//
// Example:
//
//	result, duration, err := helper.AssertCoverageOperationTiming(
//		"LoadAssignmentsForScheduleVersion",
//		func() (interface{}, error) {
//			return loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
//		},
//		1,                      // Expected 1 query
//		100 * time.Millisecond, // Max duration
//	)
func (h *CoverageAssertionHelper) AssertCoverageOperationTiming(
	operationName string,
	operation func() (interface{}, error),
	expectedQueries int,
	maxDuration time.Duration,
) (interface{}, time.Duration, error) {
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	start := time.Now()
	result, err := operation()
	duration := time.Since(start)

	if err != nil {
		return nil, duration, fmt.Errorf("%s operation failed: %w", operationName, err)
	}

	if assertErr := h.AssertCoverageCalculation(operationName, expectedQueries); assertErr != nil {
		return nil, duration, assertErr
	}

	if duration > maxDuration {
		return nil, duration, fmt.Errorf(
			"%s exceeded max duration: expected <= %v, got %v",
			operationName,
			maxDuration,
			duration,
		)
	}

	return result, duration, nil
}

// LogAllQueries returns a formatted string of all executed queries.
// Useful for debugging test failures.
//
// Example:
//
//	if err := helper.AssertQueryCount(1); err != nil {
//		t.Logf("Executed queries:\n%s", helper.LogAllQueries())
//	}
func (h *CoverageAssertionHelper) LogAllQueries() string {
	return helpers.LogQueries()
}

// GetExpectedQueryCount returns the last expected count for an operation.
// Useful for implementing custom assertions.
func (h *CoverageAssertionHelper) GetExpectedQueryCount(operationName string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count, exists := h.expectedQueryCounts[operationName]
	if !exists {
		return -1
	}
	return count
}

// formatCoverageQueries creates a readable summary of queries for coverage operation errors.
func (h *CoverageAssertionHelper) formatCoverageQueries(queries []helpers.QueryRecord) string {
	if len(queries) == 0 {
		return "No queries executed"
	}

	var result string
	result += fmt.Sprintf("\nCoverage operation queries (%d total):\n", len(queries))

	for i, q := range queries {
		// Truncate long queries for readability
		sql := q.SQL
		if len(sql) > 120 {
			sql = sql[:117] + "..."
		}

		result += fmt.Sprintf("  %d. %s\n", i+1, sql)

		if len(q.Args) > 0 && len(q.Args) <= 5 {
			result += fmt.Sprintf("     Args: %v\n", q.Args)
		}

		if q.Duration > 0 {
			result += fmt.Sprintf("     Duration: %v\n", q.Duration)
		}

		if q.Error != nil {
			result += fmt.Sprintf("     Error: %v\n", q.Error)
		}
	}

	return result
}

// AssertCoverageCalculationWithScheduleVersion is a specialized assertion for schedule version operations.
// Combines schedule version validation with query count assertions.
//
// Example:
//
//	if err := helper.AssertCoverageCalculationWithScheduleVersion(
//		ctx,
//		loader,
//		scheduleVersionID,
//		1, // Expected 1 query
//	); err != nil {
//		t.Fatalf("assertion failed: %v", err)
//	}
func (h *CoverageAssertionHelper) AssertCoverageCalculationWithScheduleVersion(
	ctx context.Context,
	loader *CoverageDataLoader,
	scheduleVersionID uuid.UUID,
	expectedQueries int,
) error {
	if scheduleVersionID == uuid.Nil {
		return fmt.Errorf("invalid schedule version ID for assertion")
	}

	return h.AssertCoverageCalculation(
		fmt.Sprintf("LoadAssignments[%s]", scheduleVersionID.String()[:8]),
		expectedQueries,
	)
}

// RegressionDetectionConfig holds configuration for regression detection.
type RegressionDetectionConfig struct {
	OperationName     string
	ExpectedQueries   int
	MaxQueryIncrease  int // Allow up to N additional queries (e.g., 1 for 1-query tolerance)
	Description       string
}

// AssertNoRegression checks that query count hasn't increased unexpectedly.
// Useful for continuous regression detection in CI/CD pipelines.
//
// config: Regression detection configuration
//
// Example:
//
//	err := helper.AssertNoRegression(RegressionDetectionConfig{
//		OperationName:    "LoadAssignmentsForScheduleVersion",
//		ExpectedQueries:  1,
//		MaxQueryIncrease: 0, // No increase allowed
//		Description:      "Data loading should always be 1 batch query",
//	})
func (h *CoverageAssertionHelper) AssertNoRegression(config RegressionDetectionConfig) error {
	actual := helpers.GetQueryCount()
	maxAllowed := config.ExpectedQueries + config.MaxQueryIncrease

	if actual <= maxAllowed {
		return nil
	}

	queries := helpers.GetQueries()
	queryDetails := h.formatCoverageQueries(queries)

	return fmt.Errorf(
		"REGRESSION DETECTED in %q:\n"+
			"  Description: %s\n"+
			"  Expected: <= %d queries (base: %d, tolerance: %d)\n"+
			"  Actual: %d\n"+
			"  Change: +%d queries\n%s",
		config.OperationName,
		config.Description,
		maxAllowed,
		config.ExpectedQueries,
		config.MaxQueryIncrease,
		actual,
		actual-config.ExpectedQueries,
		queryDetails,
	)
}

// DocumentExpectedQueries stores documentation about expected query counts for an operation.
// Used for regression detection and documentation purposes.
//
// operationName: Name of the operation
// expectedQueries: Expected query count
// reason: Why this count is expected (for documentation)
//
// Example:
//
//	helper.DocumentExpectedQueries(
//		"LoadAssignmentsForScheduleVersion",
//		1,
//		"Single batch query using IN clause or similar",
//	)
func (h *CoverageAssertionHelper) DocumentExpectedQueries(operationName string, expectedQueries int, reason string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.expectedQueryCounts[operationName] = expectedQueries

	// In a real implementation, this might log to a file or metrics system
	// For now, we just store the count
	_ = reason // Use reason for documentation/logging if needed
}
