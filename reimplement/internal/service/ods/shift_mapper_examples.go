package ods

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Example1_BasicMapping demonstrates simple shift data mapping.
func Example1_BasicMapping() {
	// Create a mapper with UTC timezone
	mapper := NewShiftInstanceMapper(time.UTC)
	
	// Example raw data from ODS parser
	raw := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "DAY",
		RequiredStaffing: "3",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}
	
	// Create UUIDs (in real code, these would come from context)
	scheduleVersionID := uuid.New()
	userID := uuid.New()
	
	// Map to entity
	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)
	if err != nil {
		fmt.Printf("Mapping failed: %v\n", err)
		return
	}
	
	// shift is now ready for persistence
	fmt.Printf("Created shift ID: %s\n", shift.ID)
	fmt.Printf("Shift type: %s\n", shift.ShiftType)
}

// Example2_ErrorHandling demonstrates comprehensive error handling.
func Example2_ErrorHandling() {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()
	
	// Test case: invalid shift type
	raw := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "MORNING",  // Invalid - must be DAY, NIGHT, or WEEKEND
		RequiredStaffing: "3",
	}
	
	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)
	
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
		// Output: invalid shift type: "MORNING" (must be one of: [DAY NIGHT WEEKEND])
	}
	
	if shift != nil {
		fmt.Println("ERROR: shift should be nil on validation failure")
	}
}

// Example3_OptionalFields demonstrates handling of optional fields.
func Example3_OptionalFields() {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()
	
	// Data with all optional fields populated
	raw := RawShiftData{
		Date:                "2025-11-20",
		ShiftType:           "NIGHT",
		RequiredStaffing:    "2",
		SpecialtyConstraint: "Cardiology",
		StudyType:           "Echo",
	}
	
	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)
	if err != nil {
		fmt.Printf("Mapping failed: %v\n", err)
		return
	}
	
	// Optional fields are now pointers to strings
	if shift.SpecialtyConstraint != nil {
		fmt.Printf("Specialty: %s\n", *shift.SpecialtyConstraint)
	}
	
	if shift.StudyType != nil {
		fmt.Printf("Study Type: %s\n", *shift.StudyType)
	}
}

// Example4_TimezoneHandling demonstrates timezone-aware mapping.
func Example4_TimezoneHandling() {
	// Load US Eastern timezone
	estLoc, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Printf("Failed to load timezone: %v\n", err)
		return
	}
	
	// Create mapper with EST timezone
	mapper := NewShiftInstanceMapper(estLoc)
	
	scheduleVersionID := uuid.New()
	userID := uuid.New()
	
	raw := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "DAY",
		RequiredStaffing: "4",
	}
	
	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)
	if err != nil {
		fmt.Printf("Mapping failed: %v\n", err)
		return
	}
	
	// Date is parsed in EST timezone
	fmt.Printf("Shift date: %s\n", shift.StartTime.Format("2006-01-02"))
	fmt.Printf("Timezone: %s\n", shift.StartTime.Location())
}

// Example5_ConfigurablePastDates demonstrates past date handling.
func Example5_ConfigurablePastDates() {
	// Configuration 1: Allow past dates (default)
	mapperAllow := NewShiftInstanceMapper(time.UTC)
	
	scheduleVersionID := uuid.New()
	userID := uuid.New()
	
	rawPast := RawShiftData{
		Date:             "2020-01-01",
		ShiftType:        "DAY",
		RequiredStaffing: "2",
	}
	
	_, err := mapperAllow.MapToShiftInstance(rawPast, scheduleVersionID, userID)
	if err == nil {
		fmt.Println("Past date accepted (default behavior)")
	}

	// Configuration 2: Reject past dates
	mapperRejectPast := NewShiftInstanceMapperWithConfig(ShiftMapperConfig{
		AllowPastDates: false,
		Timezone:       time.UTC,
	})

	_, err = mapperRejectPast.MapToShiftInstance(rawPast, scheduleVersionID, userID)
	if err != nil {
		fmt.Printf("Past date rejected: %v\n", err)
	}
}

// Example6_IntegrationWithImporter demonstrates use in the ODS import workflow.
func Example6_IntegrationWithImporter(
	ctx context.Context,
	rawShifts []RawShiftData,
	scheduleVersionID uuid.UUID,
	userID uuid.UUID,
	repository interface{}, // In real code: ShiftInstanceRepository
) []interface{} { // In real code: []ShiftInstance
	
	// Create mapper for the import operation
	mapper := NewShiftInstanceMapper(time.UTC)
	
	var results []interface{}
	
	for idx, rawShift := range rawShifts {
		// Map raw data to entity
		shift, err := mapper.MapToShiftInstance(rawShift, scheduleVersionID, userID)
		if err != nil {
			fmt.Printf("Row %d: Mapping failed: %v\n", rawShift.RowMetadata.Row, err)
			// In real implementation, record error in error collector
			continue
		}
		
		// In a real implementation, persist to database:
		// createdShift, dbErr := repository.Create(ctx, shift)
		// if dbErr != nil {
		//     fmt.Printf("Row %d: Database error: %v\n", rawShift.RowMetadata.Row, dbErr)
		//     continue
		// }
		
		results = append(results, shift)
		fmt.Printf("Row %d: Successfully mapped shift (ID: %s)\n", idx+2, shift.ID)
	}
	
	return results
}

// Example7_BatchProcessing demonstrates processing multiple shifts with error collection.
func Example7_BatchProcessing(rawShifts []RawShiftData) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()
	
	successCount := 0
	failureCount := 0
	
	for _, rawShift := range rawShifts {
		shift, err := mapper.MapToShiftInstance(rawShift, scheduleVersionID, userID)
		if err != nil {
			fmt.Printf("Error at row %d: %v\n", rawShift.RowMetadata.Row, err)
			failureCount++
			continue
		}
		
		// Process successful shift
		fmt.Printf("Processed shift: %s (%s)\n", shift.ID, shift.ShiftType)
		successCount++
	}
	
	fmt.Printf("\nSummary: %d successful, %d failed\n", successCount, failureCount)
}

// Example8_ValidateBeforePersist demonstrates validation before persistence.
func Example8_ValidateBeforePersist(
	rawShift RawShiftData,
	scheduleVersionID uuid.UUID,
	userID uuid.UUID,
) error {
	mapper := NewShiftInstanceMapper(time.UTC)
	
	// Validate and map
	shift, err := mapper.MapToShiftInstance(rawShift, scheduleVersionID, userID)
	if err != nil {
		// Return validation error before any database operation
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// All validation has passed at this point
	// Safe to proceed with persistence
	fmt.Printf("Shift validated and ready for persistence: %s\n", shift.ID)
	
	// In real code: repository.Create(ctx, shift)
	return nil
}

// Example9_AllValidShiftTypes demonstrates all valid shift types.
func Example9_AllValidShiftTypes() {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()
	
	validTypes := []string{"DAY", "NIGHT", "WEEKEND"}
	
	for _, shiftType := range validTypes {
		raw := RawShiftData{
			Date:             "2025-11-20",
			ShiftType:        shiftType,
			RequiredStaffing: "1",
		}
		
		shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)
		if err != nil {
			fmt.Printf("Unexpected error for %s: %v\n", shiftType, err)
			continue
		}
		
		fmt.Printf("✓ Valid shift type: %s\n", shift.ShiftType)
	}
}

// Example10_EdgeCases demonstrates handling edge cases.
func Example10_EdgeCases() {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()
	
	// Edge case 1: Whitespace around values
	raw1 := RawShiftData{
		Date:                "  2025-11-20  ",
		ShiftType:           "  DAY  ",
		RequiredStaffing:    "  5  ",
		SpecialtyConstraint: "  Oncology  ",
	}
	
	_, err := mapper.MapToShiftInstance(raw1, scheduleVersionID, userID)
	if err != nil {
		fmt.Printf("Whitespace handling failed: %v\n", err)
	} else {
		fmt.Println("✓ Whitespace trimmed correctly")
	}
	
	// Edge case 2: Leap year date
	raw2 := RawShiftData{
		Date:             "2024-02-29",
		ShiftType:        "NIGHT",
		RequiredStaffing: "3",
	}
	
	shift2, err := mapper.MapToShiftInstance(raw2, scheduleVersionID, userID)
	if err != nil {
		fmt.Printf("Leap year handling failed: %v\n", err)
	} else {
		fmt.Printf("✓ Leap year date accepted: %s\n", shift2.StartTime.Format("2006-01-02"))
	}
	
	// Edge case 3: Large staffing numbers
	raw3 := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "WEEKEND",
		RequiredStaffing: "100",
	}
	
	_, err = mapper.MapToShiftInstance(raw3, scheduleVersionID, userID)
	if err != nil {
		fmt.Printf("Large number handling failed: %v\n", err)
	} else {
		fmt.Println("✓ Large staffing numbers accepted")
	}
}

// Note: These are example functions for documentation purposes.
// In actual tests, use the unit tests in shift_mapper_test.go.
