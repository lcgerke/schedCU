package coverage

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/schedcu/v2/internal/entity"
)

// ============================================================================
// Test Suite 1: Empty and Edge Cases
// ============================================================================

// TestResolveCoverage_EmptyAssignments validates that all shifts show as uncovered
// when no assignments exist
func TestResolveCoverage_EmptyAssignments(t *testing.T) {
	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 2,
		entity.ShiftTypeON2: 2,
		entity.ShiftTypeDay: 3,
	}

	metrics := ResolveCoverage([]entity.Assignment{}, requirements)

	// Should have metrics for each shift type
	assert.Equal(t, 3, len(metrics.CoverageByShiftType))

	// All should show zero assigned
	for shiftType, detail := range metrics.CoverageByShiftType {
		assert.Equal(t, 0, detail.Assigned, "ShiftType %s should have 0 assigned", shiftType)
		assert.Equal(t, 0.0, detail.CoveragePercentage)
		assert.Equal(t, StatusUncovered, detail.Status)
	}

	// All shifts should be under-staffed (no coverage)
	assert.Equal(t, 3, len(metrics.UnderStaffedShifts))
	assert.Equal(t, 0, len(metrics.OverStaffedShifts))
	assert.Equal(t, 0.0, metrics.OverallCoveragePercentage)
}

// TestResolveCoverage_EmptyRequirements validates that empty requirements produce valid metrics
func TestResolveCoverage_EmptyRequirements(t *testing.T) {
	metrics := ResolveCoverage([]entity.Assignment{}, map[entity.ShiftType]int{})

	// Should have no shift types
	assert.Equal(t, 0, len(metrics.CoverageByShiftType))
	assert.Equal(t, 0, len(metrics.UnderStaffedShifts))
	assert.Equal(t, 0, len(metrics.OverStaffedShifts))
	assert.Equal(t, 0.0, metrics.OverallCoveragePercentage)
}

// TestResolveCoverage_ZeroRequirement validates shifts with zero requirement
func TestResolveCoverage_ZeroRequirement(t *testing.T) {
	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 0,
	}

	metrics := ResolveCoverage([]entity.Assignment{}, requirements)

	// Should show as zero required, zero assigned
	detail := metrics.CoverageByShiftType[entity.ShiftTypeON1]
	assert.Equal(t, 0, detail.Required)
	assert.Equal(t, 0, detail.Assigned)
	assert.Equal(t, 0.0, detail.CoveragePercentage)
}

// TestResolveCoverage_DeletedAssignmentsIgnored validates that soft-deleted assignments don't count
func TestResolveCoverage_DeletedAssignmentsIgnored(t *testing.T) {
	now := time.Now().UTC()
	deletedTime := now.Add(-1 * time.Hour)

	assignments := []entity.Assignment{
		{
			ID:                  uuid.New(),
			PersonID:            uuid.New(),
			ShiftInstanceID:     uuid.New(),
			ScheduleDate:        now,
			OriginalShiftType:   string(entity.ShiftTypeON1),
			Source:              entity.AssignmentSourceAmion,
			CreatedAt:           now,
			CreatedBy:           uuid.New(),
			DeletedAt:           &deletedTime, // This is deleted
		},
	}

	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 1,
	}

	metrics := ResolveCoverage(assignments, requirements)

	// Deleted assignment should not be counted
	detail := metrics.CoverageByShiftType[entity.ShiftTypeON1]
	assert.Equal(t, 0, detail.Assigned)
	assert.Equal(t, StatusUncovered, detail.Status)
}

// ============================================================================
// Test Suite 2: Full Coverage Scenarios
// ============================================================================

// TestResolveCoverage_FullyCovered validates fully-staffed shifts
func TestResolveCoverage_FullyCovered(t *testing.T) {
	now := time.Now().UTC()
	person1 := uuid.New()
	person2 := uuid.New()

	assignments := []entity.Assignment{
		{
			ID:                uuid.New(),
			PersonID:          person1,
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			PersonID:          person2,
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
	}

	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 2,
	}

	metrics := ResolveCoverage(assignments, requirements)

	detail := metrics.CoverageByShiftType[entity.ShiftTypeON1]
	assert.Equal(t, 2, detail.Required)
	assert.Equal(t, 2, detail.Assigned)
	assert.Equal(t, 100.0, detail.CoveragePercentage)
	assert.Equal(t, StatusFull, detail.Status)

	// Overall should be full
	assert.Equal(t, 100.0, metrics.OverallCoveragePercentage)
	assert.Equal(t, 0, len(metrics.UnderStaffedShifts))
	assert.Equal(t, 0, len(metrics.OverStaffedShifts))
}

// TestResolveCoverage_MultipleShiftsFullCovered validates multiple shifts all fully covered
func TestResolveCoverage_MultipleShiftsFullCovered(t *testing.T) {
	now := time.Now().UTC()

	assignments := []entity.Assignment{
		// ON1: 2 people
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		// ON2: 2 people
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON2),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON2),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		// DAY: 3 people
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeDay),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeDay),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeDay),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
	}

	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 2,
		entity.ShiftTypeON2: 2,
		entity.ShiftTypeDay: 3,
	}

	metrics := ResolveCoverage(assignments, requirements)

	// Verify each shift type
	for shiftType, requirement := range requirements {
		detail, exists := metrics.CoverageByShiftType[shiftType]
		require.True(t, exists)
		assert.Equal(t, requirement, detail.Assigned)
		assert.Equal(t, 100.0, detail.CoveragePercentage)
		assert.Equal(t, StatusFull, detail.Status)
	}

	// Overall should be 100%
	assert.Equal(t, 100.0, metrics.OverallCoveragePercentage)
	assert.Equal(t, 0, len(metrics.UnderStaffedShifts))
	assert.Equal(t, 0, len(metrics.OverStaffedShifts))
}

// ============================================================================
// Test Suite 3: Partial Coverage Scenarios
// ============================================================================

// TestResolveCoverage_PartialCovered validates under-staffed shifts (PARTIAL status)
func TestResolveCoverage_PartialCovered(t *testing.T) {
	now := time.Now().UTC()

	assignments := []entity.Assignment{
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
	}

	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 2,
	}

	metrics := ResolveCoverage(assignments, requirements)

	detail := metrics.CoverageByShiftType[entity.ShiftTypeON1]
	assert.Equal(t, 2, detail.Required)
	assert.Equal(t, 1, detail.Assigned)
	assert.Equal(t, 50.0, detail.CoveragePercentage)
	assert.Equal(t, StatusPartial, detail.Status)

	// Overall should be 50%
	assert.Equal(t, 50.0, metrics.OverallCoveragePercentage)
	assert.Len(t, metrics.UnderStaffedShifts, 1)
	assert.Equal(t, entity.ShiftTypeON1, metrics.UnderStaffedShifts[0])
}

// TestResolveCoverage_MixedCoverage validates mix of full, partial, and uncovered shifts
func TestResolveCoverage_MixedCoverage(t *testing.T) {
	now := time.Now().UTC()

	assignments := []entity.Assignment{
		// ON1: 2 assigned (full)
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		// ON2: 1 assigned (partial, requires 3)
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON2),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		// DAY: 0 assigned (uncovered, requires 2)
	}

	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 2,
		entity.ShiftTypeON2: 3,
		entity.ShiftTypeDay: 2,
	}

	metrics := ResolveCoverage(assignments, requirements)

	// Verify ON1 (full)
	on1 := metrics.CoverageByShiftType[entity.ShiftTypeON1]
	assert.Equal(t, 2, on1.Assigned)
	assert.Equal(t, 100.0, on1.CoveragePercentage)
	assert.Equal(t, StatusFull, on1.Status)

	// Verify ON2 (partial)
	on2 := metrics.CoverageByShiftType[entity.ShiftTypeON2]
	assert.Equal(t, 1, on2.Assigned)
	assert.Equal(t, 33.33, math.Round(on2.CoveragePercentage*100)/100) // ~33%
	assert.Equal(t, StatusPartial, on2.Status)

	// Verify DAY (uncovered)
	day := metrics.CoverageByShiftType[entity.ShiftTypeDay]
	assert.Equal(t, 0, day.Assigned)
	assert.Equal(t, 0.0, day.CoveragePercentage)
	assert.Equal(t, StatusUncovered, day.Status)

	// Overall: (2+1+0)/(2+3+2) = 3/7 = 42.86%
	expected := math.Round((3.0/7.0)*100*100) / 100
	assert.Equal(t, expected, metrics.OverallCoveragePercentage)

	// Check categorization
	assert.Equal(t, 2, len(metrics.UnderStaffedShifts)) // ON2 and DAY
	assert.Equal(t, 0, len(metrics.OverStaffedShifts))
}

// ============================================================================
// Test Suite 4: Over-Staffing Scenarios
// ============================================================================

// TestResolveCoverage_OverStaffed validates shifts with more staff than required
func TestResolveCoverage_OverStaffed(t *testing.T) {
	now := time.Now().UTC()

	assignments := []entity.Assignment{
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
	}

	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 2,
	}

	metrics := ResolveCoverage(assignments, requirements)

	detail := metrics.CoverageByShiftType[entity.ShiftTypeON1]
	assert.Equal(t, 3, detail.Assigned)
	assert.Equal(t, 100.0, detail.CoveragePercentage) // Capped at 100%
	assert.Equal(t, StatusFull, detail.Status)

	// Overall should be capped at 100%
	assert.Equal(t, 100.0, metrics.OverallCoveragePercentage)

	// Should be marked as over-staffed
	assert.Equal(t, 1, len(metrics.OverStaffedShifts))
	assert.Equal(t, entity.ShiftTypeON1, metrics.OverStaffedShifts[0])
}

// ============================================================================
// Test Suite 5: Percentage Calculation Accuracy
// ============================================================================

// TestResolveCoverage_PercentageAccuracy validates percentage calculations
func TestResolveCoverage_PercentageAccuracy(t *testing.T) {
	testCases := []struct {
		name           string
		assigned       int
		required       int
		expectedPercent float64
	}{
		{"0 of 1", 0, 1, 0.0},
		{"1 of 2", 1, 2, 50.0},
		{"1 of 3", 1, 3, 33.33},
		{"2 of 3", 2, 3, 66.67},
		{"3 of 3", 3, 3, 100.0},
		{"4 of 3 (over)", 4, 3, 100.0}, // Capped
		{"1 of 4", 1, 4, 25.0},
		{"3 of 4", 3, 4, 75.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now().UTC()
			assignments := []entity.Assignment{}

			// Create the required number of assignments
			for i := 0; i < tc.assigned; i++ {
				assignments = append(assignments, entity.Assignment{
					ID:                uuid.New(),
					PersonID:          uuid.New(),
					ShiftInstanceID:   uuid.New(),
					ScheduleDate:      now,
					OriginalShiftType: string(entity.ShiftTypeON1),
					Source:            entity.AssignmentSourceAmion,
					CreatedAt:         now,
					CreatedBy:         uuid.New(),
				})
			}

			requirements := map[entity.ShiftType]int{
				entity.ShiftTypeON1: tc.required,
			}

			metrics := ResolveCoverage(assignments, requirements)
			detail := metrics.CoverageByShiftType[entity.ShiftTypeON1]

			assert.Equal(t, tc.expectedPercent, detail.CoveragePercentage,
				"Expected %s: %.2f%, got %.2f%%", tc.name, tc.expectedPercent, detail.CoveragePercentage)
		})
	}
}

// ============================================================================
// Test Suite 6: Duplicate Assignment Handling
// ============================================================================

// TestResolveCoverage_DuplicatesCountOnce validates that same person/shift counts once
func TestResolveCoverage_DuplicatesCountOnce(t *testing.T) {
	now := time.Now().UTC()
	person := uuid.New()
	shift := uuid.New()

	// Same person assigned to same shift twice (duplicate entry)
	assignments := []entity.Assignment{
		{
			ID:                uuid.New(),
			PersonID:          person,
			ShiftInstanceID:   shift,
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
		{
			ID:                uuid.New(),
			PersonID:          person, // Same person
			ShiftInstanceID:   shift,  // Same shift
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceManual,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		},
	}

	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 1,
	}

	metrics := ResolveCoverage(assignments, requirements)

	detail := metrics.CoverageByShiftType[entity.ShiftTypeON1]
	assert.Equal(t, 1, detail.Assigned, "Duplicate assignment should count as 1")
	assert.Equal(t, 100.0, detail.CoveragePercentage)
	assert.Equal(t, StatusFull, detail.Status)
}

// ============================================================================
// Test Suite 7: Complex Scenarios (Large Scale)
// ============================================================================

// TestResolveCoverage_LargeScale validates performance with 100+ assignments
func TestResolveCoverage_LargeScale(t *testing.T) {
	now := time.Now().UTC()

	// Create 100 assignments across 5 shift types
	var assignments []entity.Assignment
	shiftTypes := []entity.ShiftType{
		entity.ShiftTypeON1,
		entity.ShiftTypeON2,
		entity.ShiftTypeMidC,
		entity.ShiftTypeMidL,
		entity.ShiftTypeDay,
	}

	for i := 0; i < 100; i++ {
		shiftType := shiftTypes[i%len(shiftTypes)]
		assignments = append(assignments, entity.Assignment{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(shiftType),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		})
	}

	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1:  25,
		entity.ShiftTypeON2:  25,
		entity.ShiftTypeMidC: 20,
		entity.ShiftTypeMidL: 15,
		entity.ShiftTypeDay:  30,
	}

	// Should complete without errors
	metrics := ResolveCoverage(assignments, requirements)

	// Verify all shift types are processed
	assert.Equal(t, len(shiftTypes), len(metrics.CoverageByShiftType))

	// Overall should be valid
	assert.Greater(t, metrics.OverallCoveragePercentage, 0.0)
	assert.LessOrEqual(t, metrics.OverallCoveragePercentage, 100.0)
}

// TestResolveCoverage_ExtremeValues validates handling of extreme values
func TestResolveCoverage_ExtremeValues(t *testing.T) {
	now := time.Now().UTC()

	// Create many assignments
	var assignments []entity.Assignment
	for i := 0; i < 1000; i++ {
		assignments = append(assignments, entity.Assignment{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(entity.ShiftTypeON1),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		})
	}

	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 1, // Only need 1, have 1000
	}

	metrics := ResolveCoverage(assignments, requirements)

	// Should cap at 100%, not exceed
	detail := metrics.CoverageByShiftType[entity.ShiftTypeON1]
	assert.Equal(t, 100.0, detail.CoveragePercentage)
	assert.Equal(t, StatusFull, detail.Status)
}

// ============================================================================
// Test Suite 8: Summary Generation
// ============================================================================

// TestResolveCoverage_SummaryGeneration validates human-readable summary
func TestResolveCoverage_SummaryGeneration(t *testing.T) {
	testCases := []struct {
		name          string
		assignments   []entity.Assignment
		requirements  map[entity.ShiftType]int
		shouldContain string
	}{
		{
			"Full coverage",
			createAssignmentsForShiftType(entity.ShiftTypeON1, 2),
			map[entity.ShiftType]int{entity.ShiftTypeON1: 2},
			"Full coverage",
		},
		{
			"Partial coverage",
			createAssignmentsForShiftType(entity.ShiftTypeON1, 1),
			map[entity.ShiftType]int{entity.ShiftTypeON1: 2},
			"partial",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := ResolveCoverage(tc.assignments, tc.requirements)
			assert.Contains(t, metrics.Summary, tc.shouldContain)
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// createAssignmentsForShiftType creates N assignments for a given shift type
func createAssignmentsForShiftType(shiftType entity.ShiftType, count int) []entity.Assignment {
	now := time.Now().UTC()
	var assignments []entity.Assignment

	for i := 0; i < count; i++ {
		assignments = append(assignments, entity.Assignment{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(shiftType),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		})
	}

	return assignments
}

// ============================================================================
// Test Suite 9: Property-Based / Invariant Tests
// ============================================================================

// TestResolveCoverage_Invariants validates mathematical invariants that must always hold
func TestResolveCoverage_Invariants(t *testing.T) {
	testCases := []struct {
		name         string
		assignments  []entity.Assignment
		requirements map[entity.ShiftType]int
	}{
		{
			"Empty",
			[]entity.Assignment{},
			map[entity.ShiftType]int{entity.ShiftTypeON1: 2},
		},
		{
			"Full coverage",
			createAssignmentsForShiftType(entity.ShiftTypeON1, 2),
			map[entity.ShiftType]int{entity.ShiftTypeON1: 2},
		},
		{
			"Partial coverage",
			createAssignmentsForShiftType(entity.ShiftTypeON1, 1),
			map[entity.ShiftType]int{entity.ShiftTypeON1: 2},
		},
		{
			"Over coverage",
			createAssignmentsForShiftType(entity.ShiftTypeON1, 5),
			map[entity.ShiftType]int{entity.ShiftTypeON1: 2},
		},
		{
			"Multiple shift types",
			append(
				createAssignmentsForShiftType(entity.ShiftTypeON1, 2),
				createAssignmentsForShiftType(entity.ShiftTypeON2, 1)...,
			),
			map[entity.ShiftType]int{
				entity.ShiftTypeON1: 2,
				entity.ShiftTypeON2: 3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := ResolveCoverage(tc.assignments, tc.requirements)

			// INVARIANT 1: Coverage percentage must be in [0, 100]
			assert.GreaterOrEqual(t, metrics.OverallCoveragePercentage, 0.0, "Coverage >= 0%")
			assert.LessOrEqual(t, metrics.OverallCoveragePercentage, 100.0, "Coverage <= 100%")

			for _, detail := range metrics.CoverageByShiftType {
				assert.GreaterOrEqual(t, detail.CoveragePercentage, 0.0)
				assert.LessOrEqual(t, detail.CoveragePercentage, 100.0)
			}

			// INVARIANT 2: All shift types in requirements should have metrics
			for shiftType := range tc.requirements {
				_, exists := metrics.CoverageByShiftType[shiftType]
				assert.True(t, exists, "ShiftType %s should have metrics", shiftType)
			}

			// INVARIANT 3: UnderStaffedShifts should only contain shift types with insufficient staff
			for _, underStaffed := range metrics.UnderStaffedShifts {
				detail := metrics.CoverageByShiftType[underStaffed]
				assert.True(t, detail.Assigned < detail.Required,
					"UnderStaffed shift %s should have assigned < required", underStaffed)
			}

			// INVARIANT 4: OverStaffedShifts should only contain shift types with excess staff
			for _, overStaffed := range metrics.OverStaffedShifts {
				detail := metrics.CoverageByShiftType[overStaffed]
				assert.True(t, detail.Assigned > detail.Required,
					"OverStaffed shift %s should have assigned > required", overStaffed)
			}

			// INVARIANT 5: Status should match coverage percentage
			for _, detail := range metrics.CoverageByShiftType {
				switch detail.Status {
				case StatusFull:
					assert.GreaterOrEqual(t, detail.Assigned, detail.Required)
				case StatusPartial:
					assert.Greater(t, detail.Assigned, 0)
					assert.Less(t, detail.Assigned, detail.Required)
				case StatusUncovered:
					assert.Equal(t, 0, detail.Assigned)
				}
			}
		})
	}
}

// TestResolveCoverage_ThreadSafety validates that ResolveCoverage is thread-safe
func TestResolveCoverage_ThreadSafety(t *testing.T) {
	now := time.Now().UTC()
	requirements := map[entity.ShiftType]int{
		entity.ShiftTypeON1: 5,
		entity.ShiftTypeON2: 5,
	}

	// Create assignments once
	var assignments []entity.Assignment
	for i := 0; i < 10; i++ {
		shiftType := entity.ShiftTypeON1
		if i%2 == 0 {
			shiftType = entity.ShiftTypeON2
		}
		assignments = append(assignments, entity.Assignment{
			ID:                uuid.New(),
			PersonID:          uuid.New(),
			ShiftInstanceID:   uuid.New(),
			ScheduleDate:      now,
			OriginalShiftType: string(shiftType),
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         now,
			CreatedBy:         uuid.New(),
		})
	}

	// Call ResolveCoverage multiple times concurrently
	// If there's any shared state issue, this will likely fail
	results := make([]CoverageMetrics, 10)
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			results[idx] = ResolveCoverage(assignments, requirements)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all results are identical
	first := results[0]
	for i := 1; i < 10; i++ {
		assert.Equal(t, first.OverallCoveragePercentage, results[i].OverallCoveragePercentage)
		assert.Equal(t, len(first.CoverageByShiftType), len(results[i].CoverageByShiftType))
	}
}
