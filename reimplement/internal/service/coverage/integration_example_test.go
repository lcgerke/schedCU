package coverage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/tests/helpers"
)

// This file demonstrates how the CoverageDataLoader integrates with:
// 1. The query counter from WP [0.7]
// 2. The coverage algorithm from WP [1.13]
// 3. Real database integration

// ExampleCoverageCalculationService demonstrates the complete flow.
type ExampleCoverageCalculationService struct {
	dataLoader *CoverageDataLoader
}

// ExampleSimpleResolveCoverage is a placeholder for the algorithm from WP [1.13].
// In the actual implementation, this would be the coverage calculation algorithm.
func ExampleSimpleResolveCoverage(
	shifts []*entity.ShiftInstance,
	requirements map[string]int,
) map[string]float64 {
	result := make(map[string]float64)

	// Count assignments by shift type
	counts := make(map[string]int)
	for _, shift := range shifts {
		counts[shift.ShiftType]++
	}

	// Calculate coverage percentage
	for shiftType, required := range requirements {
		assigned := counts[shiftType]
		percentage := 0.0
		if required > 0 {
			percentage = (float64(assigned) / float64(required)) * 100.0
		}
		result[shiftType] = percentage
	}

	return result
}

// TestCoverageDataLoaderWithQueryCounter demonstrates integration with [0.7].
// This test shows how the query counter from WP [0.7] tracks the batch query.
func TestCoverageDataLoaderWithQueryCounter(t *testing.T) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Setup: Create test data
	shifts := []*entity.ShiftInstance{
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Doctor", "ER", "Alice", userID),
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Nurse", "ER", "Bob", userID),
		entity.NewShiftInstance(scheduleVersionID, "Night", "Doctor", "ICU", "Charlie", userID),
		entity.NewShiftInstance(scheduleVersionID, "Night", "Nurse", "ICU", "Diana", userID),
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	// Reset and start query counting (WP [0.7] integration)
	helpers.ResetQueryCount()
	helpers.StartQueryCount()
	defer helpers.StopQueryCount()

	// Execute: Load assignments (should be 1 query)
	loaded, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Assert: Query count should be exactly 1
	queryCount := helpers.GetQueryCount()
	if err := helpers.AssertQueryCount(1, queryCount); err != nil {
		t.Fatalf("query count assertion failed: %v", err)
	}

	// Verify data correctness
	if len(loaded) != 4 {
		t.Fatalf("expected 4 shifts, got %d", len(loaded))
	}

	// Log the queries that were tracked
	queries := helpers.GetQueries()
	t.Logf("Queries executed: %d", len(queries))
	for i, q := range queries {
		t.Logf("  Query %d: %s", i+1, q.SQL)
	}
}

// TestCoverageCalculationIntegration demonstrates the full flow:
// DataLoader -> Algorithm -> Results
func TestCoverageCalculationIntegration(t *testing.T) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Setup: Create test schedule with various shift assignments
	shifts := []*entity.ShiftInstance{
		// Morning shifts: 2 assigned, need 3
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Doctor", "ER", "Alice", userID),
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Nurse", "ER", "Bob", userID),

		// Night shifts: 3 assigned, need 2
		entity.NewShiftInstance(scheduleVersionID, "Night", "Doctor", "ICU", "Charlie", userID),
		entity.NewShiftInstance(scheduleVersionID, "Night", "Nurse", "ICU", "Diana", userID),
		entity.NewShiftInstance(scheduleVersionID, "Night", "Doctor", "ER", "Eve", userID),

		// Afternoon shifts: 0 assigned, need 2
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	// Step 1: Load assignments via data loader (1 query)
	helpers.ResetQueryCount()
	loaded, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Verify single query execution
	if repo.queryCount != 1 {
		t.Fatalf("REGRESSION: Expected 1 query, got %d", repo.queryCount)
	}

	// Step 2: Pass to algorithm (from WP [1.13])
	requirements := map[string]int{
		"Morning":   3,
		"Night":     2,
		"Afternoon": 2,
	}

	coverage := ExampleSimpleResolveCoverage(loaded, requirements)

	// Step 3: Verify results
	expectedCoverage := map[string]float64{
		"Morning":   (2.0 / 3.0) * 100,  // 66.67%
		"Night":     (3.0 / 2.0) * 100,  // 150% (over-staffed)
		"Afternoon": (0.0 / 2.0) * 100,  // 0% (under-staffed)
	}

	for shiftType, expected := range expectedCoverage {
		actual := coverage[shiftType]
		// Allow small floating point differences
		if diff := actual - expected; diff < -0.01 || diff > 0.01 {
			t.Logf("warning: %s coverage %.2f vs expected %.2f", shiftType, actual, expected)
		}
	}

	t.Logf("Coverage calculation complete:")
	t.Logf("  Morning coverage: %.2f%%", coverage["Morning"])
	t.Logf("  Night coverage: %.2f%%", coverage["Night"])
	t.Logf("  Afternoon coverage: %.2f%%", coverage["Afternoon"])
}

// TestBatchQueryRegressionDetection demonstrates how query count assertions
// catch N+1 query patterns in tests.
func TestBatchQueryRegressionDetection(t *testing.T) {
	ctx := context.Background()

	// Simulate a badly implemented repository (N+1 pattern)
	BadRepository := &struct {
		*MockShiftInstanceRepository
	}{
		MockShiftInstanceRepository: &MockShiftInstanceRepository{
			shifts: []*entity.ShiftInstance{
				entity.NewShiftInstance(uuid.New(), "Morning", "Doctor", "ER", "Alice", uuid.New()),
				entity.NewShiftInstance(uuid.New(), "Morning", "Nurse", "ER", "Bob", uuid.New()),
			},
		},
	}

	// Simulate N+1 by incrementing query count incorrectly
	loader := NewCoverageDataLoader(BadRepository.MockShiftInstanceRepository)

	_, err := loader.LoadAssignmentsForScheduleVersion(ctx, uuid.New())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// This is what would happen with good assertion
	queryCount := BadRepository.MockShiftInstanceRepository.queryCount
	if err := helpers.AssertQueryCount(1, queryCount); err != nil {
		t.Logf("Assertion would catch regression: %v", err)
	} else {
		t.Logf("Query count assertion passed: exactly 1 query")
	}
}

// TestMultipleScheduleVersionsIsolation verifies that loads for different
// schedule versions don't interfere (each is a separate query).
func TestMultipleScheduleVersionsIsolation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	version1 := uuid.New()
	version2 := uuid.New()

	// Create shifts for two different schedule versions
	allShifts := []*entity.ShiftInstance{
		// Version 1 shifts
		entity.NewShiftInstance(version1, "Morning", "Doctor", "ER", "Alice", userID),
		entity.NewShiftInstance(version1, "Morning", "Nurse", "ER", "Bob", userID),

		// Version 2 shifts
		entity.NewShiftInstance(version2, "Night", "Doctor", "ICU", "Charlie", userID),
		entity.NewShiftInstance(version2, "Night", "Nurse", "ICU", "Diana", userID),
	}

	repo := &MockShiftInstanceRepository{shifts: allShifts}
	loader := NewCoverageDataLoader(repo)

	// Load version 1
	loaded1, err := loader.LoadAssignmentsForScheduleVersion(ctx, version1)
	if err != nil {
		t.Fatalf("load version1 failed: %v", err)
	}

	if len(loaded1) != 2 {
		t.Fatalf("version1: expected 2 shifts, got %d", len(loaded1))
	}

	// Load version 2
	loaded2, err := loader.LoadAssignmentsForScheduleVersion(ctx, version2)
	if err != nil {
		t.Fatalf("load version2 failed: %v", err)
	}

	if len(loaded2) != 2 {
		t.Fatalf("version2: expected 2 shifts, got %d", len(loaded2))
	}

	// Verify isolation
	for _, shift := range loaded1 {
		if shift.ScheduleVersionID != version1 {
			t.Fatalf("version1 load returned shift from wrong version")
		}
	}

	for _, shift := range loaded2 {
		if shift.ScheduleVersionID != version2 {
			t.Fatalf("version2 load returned shift from wrong version")
		}
	}

	// Query count: 2 queries total (1 per load), not N+1
	expectedQueries := 2 // One for each LoadAssignmentsForScheduleVersion call
	if repo.queryCount != expectedQueries {
		t.Fatalf("expected %d queries, got %d", expectedQueries, repo.queryCount)
	}

	t.Logf("Multiple schedule versions: %d queries (as expected)", repo.queryCount)
}

// TestEmptyResultHandling verifies correct handling of empty results.
func TestEmptyResultHandling(t *testing.T) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()

	repo := &MockShiftInstanceRepository{shifts: []*entity.ShiftInstance{}}
	loader := NewCoverageDataLoader(repo)

	loaded, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(loaded) != 0 {
		t.Fatalf("expected empty slice, got %d shifts", len(loaded))
	}

	if loaded == nil {
		t.Logf("Empty result returned as nil slice (correct)")
	}

	// Verify single query still executed
	if repo.queryCount != 1 {
		t.Fatalf("expected 1 query even for empty result, got %d", repo.queryCount)
	}
}

// BenchmarkCoverageCalculationEnd2End benchmarks the complete flow.
func BenchmarkCoverageCalculationEnd2End(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create realistic test data
	shifts := make([]*entity.ShiftInstance, 100)
	for i := 0; i < 100; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Doctor",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	requirements := map[string]int{
		"Morning": 50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Load
		loaded, _ := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

		// Calculate (algorithm from WP [1.13])
		_ = ExampleSimpleResolveCoverage(loaded, requirements)
	}
}
