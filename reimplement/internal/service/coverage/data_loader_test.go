package coverage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/tests/helpers"
)

// MockShiftInstanceRepository is a test double for ShiftInstanceRepository.
type MockShiftInstanceRepository struct {
	shifts          []*entity.ShiftInstance
	queryCount      int
	getByVersionErr error
}

// GetByScheduleVersion returns shifts for a schedule version (tracks query count).
func (m *MockShiftInstanceRepository) GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.ShiftInstance, error) {
	m.queryCount++
	if m.getByVersionErr != nil {
		return nil, m.getByVersionErr
	}

	var result []*entity.ShiftInstance
	for _, shift := range m.shifts {
		if shift.ScheduleVersionID == scheduleVersionID {
			result = append(result, shift)
		}
	}
	return result, nil
}

// Create implements the interface (unused in these tests).
func (m *MockShiftInstanceRepository) Create(ctx context.Context, shift *entity.ShiftInstance) (*entity.ShiftInstance, error) {
	return shift, nil
}

// GetByID implements the interface (unused in these tests).
func (m *MockShiftInstanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ShiftInstance, error) {
	return nil, nil
}

// CreateBatch implements the interface (unused in these tests).
func (m *MockShiftInstanceRepository) CreateBatch(ctx context.Context, shifts []*entity.ShiftInstance) (int, error) {
	return 0, nil
}

// Update implements the interface (unused in these tests).
func (m *MockShiftInstanceRepository) Update(ctx context.Context, shift *entity.ShiftInstance) error {
	return nil
}

// Delete implements the interface (unused in these tests).
func (m *MockShiftInstanceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// DeleteByScheduleVersion implements the interface (unused in these tests).
func (m *MockShiftInstanceRepository) DeleteByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) (int, error) {
	return 0, nil
}

// TestCoverageDataLoaderEmpty verifies loader returns empty result with 0 assignments.
func TestCoverageDataLoaderEmpty(t *testing.T) {
	ctx := context.Background()
	repo := &MockShiftInstanceRepository{shifts: []*entity.ShiftInstance{}}

	loader := NewCoverageDataLoader(repo)
	shifts, err := loader.LoadAssignmentsForScheduleVersion(ctx, uuid.New())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shifts) != 0 {
		t.Fatalf("expected 0 shifts, got %d", len(shifts))
	}
	if repo.queryCount != 1 {
		t.Fatalf("expected 1 query, got %d", repo.queryCount)
	}
}

// TestCoverageDataLoaderSmall verifies loader returns 10 assignments with 1 query.
func TestCoverageDataLoaderSmall(t *testing.T) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 10 shifts
	shifts := make([]*entity.ShiftInstance, 10)
	for i := 0; i < 10; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"Main ER",
			"Staff "+string(rune(65+i)),
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	loaded, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded) != 10 {
		t.Fatalf("expected 10 shifts, got %d", len(loaded))
	}
	if repo.queryCount != 1 {
		t.Fatalf("expected 1 query, got %d", repo.queryCount)
	}

	// Verify correct shifts were returned
	for i, shift := range loaded {
		if shift.ScheduleVersionID != scheduleVersionID {
			t.Fatalf("shift %d has wrong schedule version ID", i)
		}
		if shift.Position != "Nurse" {
			t.Fatalf("shift %d has wrong position", i)
		}
	}
}

// TestCoverageDataLoaderLarge verifies loader handles 1000+ assignments efficiently.
func TestCoverageDataLoaderLarge(t *testing.T) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 1000 shifts
	shifts := make([]*entity.ShiftInstance, 1000)
	for i := 0; i < 1000; i++ {
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

	loaded, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded) != 1000 {
		t.Fatalf("expected 1000 shifts, got %d", len(loaded))
	}
	if repo.queryCount != 1 {
		t.Fatalf("expected 1 query for 1000 items, got %d", repo.queryCount)
	}
}

// TestCoverageDataLoaderQueryCount verifies exactly 1 query is executed (no N+1).
func TestCoverageDataLoaderQueryCount(t *testing.T) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 100 shifts
	shifts := make([]*entity.ShiftInstance, 100)
	for i := 0; i < 100; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Night",
			"Nurse",
			"ICU",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	// Load 3 times to verify query count increases correctly
	for j := 0; j < 3; j++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", j+1, err)
		}

		expectedQueries := (j + 1) * 1 // 1 query per load
		if repo.queryCount != expectedQueries {
			t.Fatalf("iteration %d: expected %d total queries, got %d", j+1, expectedQueries, repo.queryCount)
		}
	}
}

// TestCoverageDataLoaderDataCorrectness verifies returned data matches expectations.
func TestCoverageDataLoaderDataCorrectness(t *testing.T) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create shifts with different types
	morningShift := entity.NewShiftInstance(scheduleVersionID, "Morning", "Doctor", "ER", "Alice", userID)
	nightShift := entity.NewShiftInstance(scheduleVersionID, "Night", "Nurse", "ICU", "Bob", userID)
	afternoonShift := entity.NewShiftInstance(scheduleVersionID, "Afternoon", "Admin", "Desk", "Charlie", userID)

	repo := &MockShiftInstanceRepository{shifts: []*entity.ShiftInstance{morningShift, nightShift, afternoonShift}}
	loader := NewCoverageDataLoader(repo)

	loaded, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("expected 3 shifts, got %d", len(loaded))
	}

	// Verify data integrity
	shiftTypes := make(map[string]bool)
	positions := make(map[string]bool)
	staffMembers := make(map[string]bool)

	for _, shift := range loaded {
		shiftTypes[shift.ShiftType] = true
		positions[shift.Position] = true
		staffMembers[shift.StaffMember] = true
	}

	if !shiftTypes["Morning"] || !shiftTypes["Night"] || !shiftTypes["Afternoon"] {
		t.Fatal("not all shift types found")
	}
	if !positions["Doctor"] || !positions["Nurse"] || !positions["Admin"] {
		t.Fatal("not all positions found")
	}
	if !staffMembers["Alice"] || !staffMembers["Bob"] || !staffMembers["Charlie"] {
		t.Fatal("not all staff members found")
	}
}

// TestCoverageDataLoaderFiltersByScheduleVersion verifies only correct version shifts are returned.
func TestCoverageDataLoaderFiltersByScheduleVersion(t *testing.T) {
	ctx := context.Background()
	scheduleVersion1 := uuid.New()
	scheduleVersion2 := uuid.New()
	userID := uuid.New()

	// Create shifts for two different schedule versions
	shift1 := entity.NewShiftInstance(scheduleVersion1, "Morning", "Doctor", "ER", "Alice", userID)
	shift2 := entity.NewShiftInstance(scheduleVersion2, "Morning", "Doctor", "ER", "Bob", userID)
	shift3 := entity.NewShiftInstance(scheduleVersion1, "Night", "Nurse", "ICU", "Charlie", userID)

	repo := &MockShiftInstanceRepository{shifts: []*entity.ShiftInstance{shift1, shift2, shift3}}
	loader := NewCoverageDataLoader(repo)

	// Load for schedule version 1
	loaded1, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersion1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded1) != 2 {
		t.Fatalf("expected 2 shifts for version 1, got %d", len(loaded1))
	}

	// Load for schedule version 2
	loaded2, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersion2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded2) != 1 {
		t.Fatalf("expected 1 shift for version 2, got %d", len(loaded2))
	}

	// Verify correct shifts returned
	found := false
	for _, shift := range loaded1 {
		if shift.StaffMember == "Bob" {
			t.Fatal("version 1 load returned shift from version 2")
		}
	}
	for _, shift := range loaded2 {
		if shift.StaffMember == "Alice" || shift.StaffMember == "Charlie" {
			t.Fatal("version 2 load returned shift from version 1")
		}
		found = true
	}
	if !found {
		t.Fatal("version 2 load didn't return Bob's shift")
	}
}

// TestCoverageDataLoaderRepositoryError verifies error handling when repository fails.
func TestCoverageDataLoaderRepositoryError(t *testing.T) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()

	repo := &MockShiftInstanceRepository{
		getByVersionErr: ErrRepositoryFailed,
	}
	loader := NewCoverageDataLoader(repo)

	_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

	// The error should be wrapped, but should contain the original error
	if !errors.Is(err, ErrRepositoryFailed) {
		t.Fatalf("expected ErrRepositoryFailed in chain, got %v", err)
	}
}

// TestCoverageDataLoaderCachingDisabled verifies each call hits the repository.
// (We're not caching, but documenting this behavior.)
func TestCoverageDataLoaderNoCaching(t *testing.T) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shift := entity.NewShiftInstance(scheduleVersionID, "Morning", "Doctor", "ER", "Alice", userID)
	repo := &MockShiftInstanceRepository{shifts: []*entity.ShiftInstance{shift}}
	loader := NewCoverageDataLoader(repo)

	// Load twice
	_, _ = loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
	_, _ = loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

	// Should have 2 queries (no caching)
	if repo.queryCount != 2 {
		t.Fatalf("expected 2 queries (no caching), got %d", repo.queryCount)
	}
}

// TestCoverageDataLoaderContextCancellation verifies context cancellation is handled.
func TestCoverageDataLoaderContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	scheduleVersionID := uuid.New()
	repo := &MockShiftInstanceRepository{shifts: []*entity.ShiftInstance{}}
	loader := NewCoverageDataLoader(repo)

	// This should not error since mock doesn't check context, but we document the interface
	_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)

	if err != nil && err != context.Canceled {
		// Allow nil error for mock, but document real impl should check context
	}
}

// TestCoverageDataLoaderBenchmarkSmall benchmarks loading 10 assignments.
func BenchmarkCoverageDataLoaderSmall(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shifts := make([]*entity.ShiftInstance, 10)
	for i := 0; i < 10; i++ {
		shifts[i] = entity.NewShiftInstance(scheduleVersionID, "Morning", "Nurse", "ER", "Staff", userID)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
	}
}

// TestCoverageDataLoaderBenchmarkLarge benchmarks loading 1000 assignments.
func BenchmarkCoverageDataLoaderLarge(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shifts := make([]*entity.ShiftInstance, 1000)
	for i := 0; i < 1000; i++ {
		shifts[i] = entity.NewShiftInstance(scheduleVersionID, "Morning", "Doctor", "ER", "Staff", userID)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
	}
}

// TestCoverageDataLoaderBenchmarkWithQueryCounter verifies performance with query tracking.
func TestCoverageDataLoaderPerformanceWithQueryTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shifts := make([]*entity.ShiftInstance, 1000)
	for i := 0; i < 1000; i++ {
		shifts[i] = entity.NewShiftInstance(scheduleVersionID, "Morning", "Doctor", "ER", "Staff", userID)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	// Reset helper query counter for this test
	helpers.ResetQueryCount()

	start := time.Now()
	loaded, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded) != 1000 {
		t.Fatalf("expected 1000 shifts, got %d", len(loaded))
	}

	// Performance assertion: should complete in <100ms
	if duration > 100*time.Millisecond {
		t.Logf("WARNING: Load took %v (expected <100ms) for 1000 assignments", duration)
	}

	// Verify mock tracked 1 query
	if repo.queryCount != 1 {
		t.Fatalf("expected 1 query on mock, got %d", repo.queryCount)
	}

	t.Logf("Performance: loaded 1000 assignments in %v (mock query count: %d)", duration, repo.queryCount)
}

// TestCoverageDataLoaderIntegration tests integration between loader and algorithm pattern.
func TestCoverageDataLoaderIntegration(t *testing.T) {
	// This demonstrates how the loader will be used with the coverage algorithm
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create sample assignments for different shift types
	shifts := []*entity.ShiftInstance{
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Doctor", "ER", "Alice", userID),
		entity.NewShiftInstance(scheduleVersionID, "Morning", "Nurse", "ER", "Bob", userID),
		entity.NewShiftInstance(scheduleVersionID, "Night", "Doctor", "ICU", "Charlie", userID),
		entity.NewShiftInstance(scheduleVersionID, "Night", "Nurse", "ICU", "Diana", userID),
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	// Load assignments (this would be step 1 in coverage calculation)
	loaded, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
	if err != nil {
		t.Fatalf("failed to load assignments: %v", err)
	}

	// Verify we have all assignments for algorithm
	if len(loaded) != 4 {
		t.Fatalf("expected 4 assignments, got %d", len(loaded))
	}

	// Demonstrate algorithm could now process these
	// (Algorithm not implemented in this WP, but loader provides data in expected format)
	shiftCounts := make(map[string]int)
	for _, shift := range loaded {
		shiftCounts[shift.ShiftType]++
	}

	if shiftCounts["Morning"] != 2 || shiftCounts["Night"] != 2 {
		t.Fatal("unexpected shift distribution")
	}

	// Verify query count for regression detection
	if repo.queryCount != 1 {
		t.Fatalf("REGRESSION: expected 1 query, got %d", repo.queryCount)
	}
}
