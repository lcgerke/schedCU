// Package coverage provides coverage calculation services for schedule management.
package coverage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
)

// BenchmarkDataLoader_100 benchmarks data loading with 100 assignments.
// Expected: O(n) time complexity, linear memory growth
func BenchmarkDataLoader_100(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 100 shifts
	shifts := make([]*entity.ShiftInstance, 100)
	for i := 0; i < 100; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkDataLoader_1000 benchmarks data loading with 1000 assignments.
// Expected: O(n) time complexity, linear memory growth
// Should be ~10x slower than 100 assignments
func BenchmarkDataLoader_1000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 1000 shifts
	shifts := make([]*entity.ShiftInstance, 1000)
	for i := 0; i < 1000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkDataLoader_10000 benchmarks data loading with 10000 assignments.
// Expected: O(n) time complexity, linear memory growth
// Should be ~100x slower than 100 assignments
func BenchmarkDataLoader_10000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 10000 shifts
	shifts := make([]*entity.ShiftInstance, 10000)
	for i := 0; i < 10000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkDataLoaderAlloc_100 benchmarks memory allocation for 100 assignments.
// Measures heap allocations and memory growth
func BenchmarkDataLoaderAlloc_100(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 100 shifts
	shifts := make([]*entity.ShiftInstance, 100)
	for i := 0; i < 100; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkDataLoaderAlloc_1000 benchmarks memory allocation for 1000 assignments.
// Measures heap allocations and memory growth
func BenchmarkDataLoaderAlloc_1000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 1000 shifts
	shifts := make([]*entity.ShiftInstance, 1000)
	for i := 0; i < 1000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkDataLoaderAlloc_10000 benchmarks memory allocation for 10000 assignments.
// Measures heap allocations and memory growth
func BenchmarkDataLoaderAlloc_10000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 10000 shifts
	shifts := make([]*entity.ShiftInstance, 10000)
	for i := 0; i < 10000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkDataLoaderParallel_100 benchmarks data loading with 100 assignments in parallel.
// Tests concurrent access patterns to detect race conditions or lock contention
func BenchmarkDataLoaderParallel_100(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 100 shifts
	shifts := make([]*entity.ShiftInstance, 100)
	for i := 0; i < 100; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})
}

// BenchmarkDataLoaderParallel_1000 benchmarks data loading with 1000 assignments in parallel.
// Tests concurrent access patterns to detect race conditions or lock contention
func BenchmarkDataLoaderParallel_1000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 1000 shifts
	shifts := make([]*entity.ShiftInstance, 1000)
	for i := 0; i < 1000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})
}

// BenchmarkDataLoaderParallel_10000 benchmarks data loading with 10000 assignments in parallel.
// Tests concurrent access patterns to detect race conditions or lock contention
func BenchmarkDataLoaderParallel_10000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 10000 shifts
	shifts := make([]*entity.ShiftInstance, 10000)
	for i := 0; i < 10000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})
}

// BenchmarkRepositoryMock_100 benchmarks the mock repository performance with 100 items.
// Isolates repository layer performance from loader implementation
func BenchmarkRepositoryMock_100(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 100 shifts
	shifts := make([]*entity.ShiftInstance, 100)
	for i := 0; i < 100; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkRepositoryMock_1000 benchmarks the mock repository performance with 1000 items.
// Isolates repository layer performance from loader implementation
func BenchmarkRepositoryMock_1000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 1000 shifts
	shifts := make([]*entity.ShiftInstance, 1000)
	for i := 0; i < 1000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkRepositoryMock_10000 benchmarks the mock repository performance with 10000 items.
// Isolates repository layer performance from loader implementation
func BenchmarkRepositoryMock_10000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 10000 shifts
	shifts := make([]*entity.ShiftInstance, 10000)
	for i := 0; i < 10000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetByScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkRepositoryMockAlloc_100 benchmarks repository memory allocation with 100 items.
func BenchmarkRepositoryMockAlloc_100(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 100 shifts
	shifts := make([]*entity.ShiftInstance, 100)
	for i := 0; i < 100; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetByScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkRepositoryMockAlloc_1000 benchmarks repository memory allocation with 1000 items.
func BenchmarkRepositoryMockAlloc_1000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 1000 shifts
	shifts := make([]*entity.ShiftInstance, 1000)
	for i := 0; i < 1000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetByScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkRepositoryMockAlloc_10000 benchmarks repository memory allocation with 10000 items.
func BenchmarkRepositoryMockAlloc_10000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 10000 shifts
	shifts := make([]*entity.ShiftInstance, 10000)
	for i := 0; i < 10000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			"Morning",
			"Nurse",
			"ER",
			"Staff",
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetByScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkDataLoaderWithVariedShifts_100 benchmarks with diverse shift data to test real-world scenario.
// Creates shifts with different types, positions, and staff members
func BenchmarkDataLoaderWithVariedShifts_100(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 100 shifts with varied data
	shifts := make([]*entity.ShiftInstance, 100)
	shiftTypes := []string{"Morning", "Afternoon", "Night", "Overnight"}
	positions := []string{"Doctor", "Nurse", "Admin", "Technician", "Manager"}

	for i := 0; i < 100; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			shiftTypes[i%len(shiftTypes)],
			positions[i%len(positions)],
			"ER",
			"Staff"+string(rune(65+(i%26))),
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkDataLoaderWithVariedShifts_1000 benchmarks with diverse shift data for 1000 items.
func BenchmarkDataLoaderWithVariedShifts_1000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 1000 shifts with varied data
	shifts := make([]*entity.ShiftInstance, 1000)
	shiftTypes := []string{"Morning", "Afternoon", "Night", "Overnight"}
	positions := []string{"Doctor", "Nurse", "Admin", "Technician", "Manager"}

	for i := 0; i < 1000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			shiftTypes[i%len(shiftTypes)],
			positions[i%len(positions)],
			"ER",
			"Staff"+string(rune(65+(i%26))),
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkDataLoaderWithVariedShifts_10000 benchmarks with diverse shift data for 10000 items.
func BenchmarkDataLoaderWithVariedShifts_10000(b *testing.B) {
	ctx := context.Background()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create 10000 shifts with varied data
	shifts := make([]*entity.ShiftInstance, 10000)
	shiftTypes := []string{"Morning", "Afternoon", "Night", "Overnight"}
	positions := []string{"Doctor", "Nurse", "Admin", "Technician", "Manager"}

	for i := 0; i < 10000; i++ {
		shifts[i] = entity.NewShiftInstance(
			scheduleVersionID,
			shiftTypes[i%len(shiftTypes)],
			positions[i%len(positions)],
			"ER",
			"Staff"+string(rune(65+(i%26))),
			userID,
		)
	}

	repo := &MockShiftInstanceRepository{shifts: shifts}
	loader := NewCoverageDataLoader(repo)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
