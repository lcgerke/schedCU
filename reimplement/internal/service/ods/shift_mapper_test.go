package ods

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
)

// TestMapToShiftInstanceValidInput tests successful mapping with all valid inputs.
func TestMapToShiftInstanceValidInput(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:                "2025-11-20",
		ShiftType:           "DAY",
		RequiredStaffing:    "3",
		SpecialtyConstraint: "Radiology",
		StudyType:           "CT",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if shift == nil {
		t.Fatalf("Expected shift to be non-nil")
	}

	if shift.ScheduleVersionID != scheduleVersionID {
		t.Errorf("Expected ScheduleVersionID %s, got %s", scheduleVersionID, shift.ScheduleVersionID)
	}

	if shift.ShiftType != "DAY" {
		t.Errorf("Expected ShiftType 'DAY', got %s", shift.ShiftType)
	}

	if shift.CreatedBy != userID {
		t.Errorf("Expected CreatedBy %s, got %s", userID, shift.CreatedBy)
	}

	if shift.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be set, got zero time")
	}

	if shift.SpecialtyConstraint == nil || *shift.SpecialtyConstraint != "Radiology" {
		t.Errorf("Expected SpecialtyConstraint 'Radiology', got %v", shift.SpecialtyConstraint)
	}

	if shift.StudyType == nil || *shift.StudyType != "CT" {
		t.Errorf("Expected StudyType 'CT', got %v", shift.StudyType)
	}
}

// TestMapToShiftInstanceInvalidDateFormat tests handling of invalid date strings.
func TestMapToShiftInstanceInvalidDateFormat(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name     string
		dateStr  string
		errMatch string
	}{
		{
			name:     "garbage date",
			dateStr:  "not-a-date",
			errMatch: "invalid date format",
		},
		{
			name:     "partial date",
			dateStr:  "2025-11",
			errMatch: "invalid date format",
		},
		{
			name:     "empty date",
			dateStr:  "",
			errMatch: "date cannot be empty",
		},
		{
			name:     "wrong date format",
			dateStr:  "11/20/2025",
			errMatch: "invalid date format",
		},
		{
			name:     "extra spaces",
			dateStr:  "2025-11-20 extra",
			errMatch: "invalid date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := RawShiftData{
				Date:             tt.dateStr,
				ShiftType:        "DAY",
				RequiredStaffing: "3",
				RowMetadata: RowMetadata{
					Row: 2,
				},
			}

			shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			if shift != nil {
				t.Errorf("Expected shift to be nil on error, got %v", shift)
			}
		})
	}
}

// TestMapToShiftInstanceInvalidShiftType tests validation of shift type enum.
func TestMapToShiftInstanceInvalidShiftType(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name      string
		shiftType string
	}{
		{
			name:      "lowercase day",
			shiftType: "day",
		},
		{
			name:      "mixed case",
			shiftType: "Day",
		},
		{
			name:      "unknown type",
			shiftType: "AFTERNOON",
		},
		{
			name:      "empty shift type",
			shiftType: "",
		},
		{
			name:      "whitespace only",
			shiftType: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := RawShiftData{
				Date:             "2025-11-20",
				ShiftType:        tt.shiftType,
				RequiredStaffing: "3",
				RowMetadata: RowMetadata{
					Row: 2,
				},
			}

			shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

			if err == nil {
				t.Errorf("Expected error for shift type %q, got nil", tt.shiftType)
			}

			if shift != nil {
				t.Errorf("Expected shift to be nil on error, got %v", shift)
			}
		})
	}
}

// TestMapToShiftInstanceValidShiftTypes tests all valid shift types.
func TestMapToShiftInstanceValidShiftTypes(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	validTypes := []string{"DAY", "NIGHT", "WEEKEND"}

	for _, shiftType := range validTypes {
		t.Run(shiftType, func(t *testing.T) {
			raw := RawShiftData{
				Date:             "2025-11-20",
				ShiftType:        shiftType,
				RequiredStaffing: "3",
				RowMetadata: RowMetadata{
					Row: 2,
				},
			}

			shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

			if err != nil {
				t.Fatalf("Expected no error for valid shift type %q: %v", shiftType, err)
			}

			if shift.ShiftType != shiftType {
				t.Errorf("Expected ShiftType %q, got %q", shiftType, shift.ShiftType)
			}
		})
	}
}

// TestMapToShiftInstanceNegativeRequiredStaffing tests rejection of negative staffing.
func TestMapToShiftInstanceNegativeRequiredStaffing(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "DAY",
		RequiredStaffing: "-5",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err == nil {
		t.Errorf("Expected error for negative required staffing, got nil")
	}

	if shift != nil {
		t.Errorf("Expected shift to be nil on error, got %v", shift)
	}
}

// TestMapToShiftInstanceZeroRequiredStaffing tests rejection of zero staffing.
func TestMapToShiftInstanceZeroRequiredStaffing(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "DAY",
		RequiredStaffing: "0",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err == nil {
		t.Errorf("Expected error for zero required staffing, got nil")
	}

	if shift != nil {
		t.Errorf("Expected shift to be nil on error, got %v", shift)
	}
}

// TestMapToShiftInstanceInvalidStaffingFormat tests handling of non-numeric staffing.
func TestMapToShiftInstanceInvalidStaffingFormat(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name             string
		requiredStaffing string
	}{
		{
			name:             "text value",
			requiredStaffing: "abc",
		},
		{
			name:             "decimal value",
			requiredStaffing: "3.5",
		},
		{
			name:             "empty string",
			requiredStaffing: "",
		},
		{
			name:             "whitespace only",
			requiredStaffing: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := RawShiftData{
				Date:             "2025-11-20",
				ShiftType:        "DAY",
				RequiredStaffing: tt.requiredStaffing,
				RowMetadata: RowMetadata{
					Row: 2,
				},
			}

			shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

			if err == nil {
				t.Errorf("Expected error for staffing value %q, got nil", tt.requiredStaffing)
			}

			if shift != nil {
				t.Errorf("Expected shift to be nil on error, got %v", shift)
			}
		})
	}
}

// TestMapToShiftInstanceDateInPast tests optional past date rejection.
func TestMapToShiftInstanceDateInPast(t *testing.T) {
	// Initialize mapper with AllowPastDates = false
	config := ShiftMapperConfig{
		AllowPastDates: false,
		Timezone:       time.UTC,
	}
	mapper := NewShiftInstanceMapperWithConfig(config)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Use a date in the past
	raw := RawShiftData{
		Date:             "2020-01-01",
		ShiftType:        "DAY",
		RequiredStaffing: "3",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err == nil {
		t.Errorf("Expected error for past date, got nil")
	}

	if shift != nil {
		t.Errorf("Expected shift to be nil on error, got %v", shift)
	}
}

// TestMapToShiftInstanceAllowPastDatesConfig tests configuration for allowing past dates.
func TestMapToShiftInstanceAllowPastDatesConfig(t *testing.T) {
	// Initialize mapper with AllowPastDates = true
	config := ShiftMapperConfig{
		AllowPastDates: true,
		Timezone:       time.UTC,
	}
	mapper := NewShiftInstanceMapperWithConfig(config)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Use a date in the past
	raw := RawShiftData{
		Date:             "2020-01-01",
		ShiftType:        "DAY",
		RequiredStaffing: "3",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err != nil {
		t.Fatalf("Expected no error when past dates allowed, got: %v", err)
	}

	if shift == nil {
		t.Fatalf("Expected shift to be non-nil")
	}
}

// TestMapToShiftInstanceLeapYearDate tests handling of leap year date.
func TestMapToShiftInstanceLeapYearDate(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:             "2024-02-29",
		ShiftType:        "DAY",
		RequiredStaffing: "3",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err != nil {
		t.Fatalf("Expected no error for leap year date, got: %v", err)
	}

	if shift == nil {
		t.Fatalf("Expected shift to be non-nil")
	}

	// Verify the date was parsed correctly
	expectedDate := time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC)
	if shift.StartTime == nil || !shift.StartTime.Equal(expectedDate) {
		t.Errorf("Expected StartTime %v, got %v", expectedDate, shift.StartTime)
	}
}

// TestMapToShiftInstanceTimezoneHandling tests timezone conversion.
func TestMapToShiftInstanceTimezoneHandling(t *testing.T) {
	// Load different timezones
	utc := time.UTC
	estLoc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load EST timezone: %v", err)
	}

	// Test with UTC
	mapperUTC := NewShiftInstanceMapper(utc)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "DAY",
		RequiredStaffing: "3",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shiftUTC, err := mapperUTC.MapToShiftInstance(raw, scheduleVersionID, userID)
	if err != nil {
		t.Fatalf("Expected no error with UTC timezone, got: %v", err)
	}

	if shiftUTC.StartTime.Location() != utc {
		t.Errorf("Expected timezone UTC, got %s", shiftUTC.StartTime.Location())
	}

	// Test with EST
	mapperEST := NewShiftInstanceMapper(estLoc)
	shiftEST, err := mapperEST.MapToShiftInstance(raw, scheduleVersionID, userID)
	if err != nil {
		t.Fatalf("Expected no error with EST timezone, got: %v", err)
	}

	if shiftEST.StartTime.Location() != estLoc {
		t.Errorf("Expected timezone EST, got %s", shiftEST.StartTime.Location())
	}
}

// TestMapToShiftInstanceCreatedAtAndCreatedBy tests timestamp fields.
func TestMapToShiftInstanceCreatedAtAndCreatedBy(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	beforeTime := time.Now().UTC()

	raw := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "DAY",
		RequiredStaffing: "3",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	afterTime := time.Now().UTC()

	// Verify CreatedAt is within reasonable time range
	if shift.CreatedAt.Before(beforeTime) || shift.CreatedAt.After(afterTime.Add(1*time.Second)) {
		t.Errorf("CreatedAt %v is not within expected range [%v, %v]", shift.CreatedAt, beforeTime, afterTime)
	}

	// Verify CreatedBy matches user ID
	if shift.CreatedBy != userID {
		t.Errorf("Expected CreatedBy %s, got %s", userID, shift.CreatedBy)
	}
}

// TestMapToShiftInstanceOptionalFieldsNil tests handling of empty optional fields.
func TestMapToShiftInstanceOptionalFieldsNil(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:                "2025-11-20",
		ShiftType:           "DAY",
		RequiredStaffing:    "3",
		SpecialtyConstraint: "",
		StudyType:           "",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if shift.SpecialtyConstraint != nil {
		t.Errorf("Expected SpecialtyConstraint to be nil, got %v", shift.SpecialtyConstraint)
	}

	if shift.StudyType != nil {
		t.Errorf("Expected StudyType to be nil, got %v", shift.StudyType)
	}
}

// TestMapToShiftInstanceOptionalFieldsPreserved tests that non-empty optional fields are preserved.
func TestMapToShiftInstanceOptionalFieldsPreserved(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:                "2025-11-20",
		ShiftType:           "DAY",
		RequiredStaffing:    "3",
		SpecialtyConstraint: "Oncology",
		StudyType:           "PET",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if shift.SpecialtyConstraint == nil || *shift.SpecialtyConstraint != "Oncology" {
		t.Errorf("Expected SpecialtyConstraint 'Oncology', got %v", shift.SpecialtyConstraint)
	}

	if shift.StudyType == nil || *shift.StudyType != "PET" {
		t.Errorf("Expected StudyType 'PET', got %v", shift.StudyType)
	}
}

// TestMapToShiftInstanceYearBoundary tests dates at year boundaries.
func TestMapToShiftInstanceYearBoundary(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name    string
		dateStr string
	}{
		{
			name:    "year start",
			dateStr: "2025-01-01",
		},
		{
			name:    "year end",
			dateStr: "2025-12-31",
		},
		{
			name:    "decade boundary",
			dateStr: "2030-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := RawShiftData{
				Date:             tt.dateStr,
				ShiftType:        "DAY",
				RequiredStaffing: "3",
				RowMetadata: RowMetadata{
					Row: 2,
				},
			}

			shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

			if err != nil {
				t.Fatalf("Expected no error for date %s: %v", tt.dateStr, err)
			}

			if shift == nil {
				t.Fatalf("Expected shift to be non-nil for date %s", tt.dateStr)
			}
		})
	}
}

// TestMapToShiftInstanceRequiredStaffingConversion tests proper int conversion.
func TestMapToShiftInstanceRequiredStaffingConversion(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name             string
		requiredStaffing string
		expectedValue    int
	}{
		{
			name:             "single digit",
			requiredStaffing: "5",
			expectedValue:    5,
		},
		{
			name:             "large number",
			requiredStaffing: "100",
			expectedValue:    100,
		},
		{
			name:             "one staff",
			requiredStaffing: "1",
			expectedValue:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := RawShiftData{
				Date:             "2025-11-20",
				ShiftType:        "DAY",
				RequiredStaffing: tt.requiredStaffing,
				RowMetadata: RowMetadata{
					Row: 2,
				},
			}

			shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

			if err != nil {
				t.Fatalf("Expected no error for staffing %s: %v", tt.requiredStaffing, err)
			}

			if shift == nil {
				t.Fatalf("Expected shift to be non-nil for staffing %s", tt.requiredStaffing)
			}
		})
	}
}

// TestMapToShiftInstanceReturnsEntity tests that mapper returns proper ShiftInstance.
func TestMapToShiftInstanceReturnsEntity(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "NIGHT",
		RequiredStaffing: "5",
		RowMetadata: RowMetadata{
			Row: 10,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify it's a proper ShiftInstance entity
	if shift == nil {
		t.Fatalf("Expected shift to be non-nil")
	}

	if _, ok := interface{}(shift).(*entity.ShiftInstance); !ok {
		t.Errorf("Expected *entity.ShiftInstance, got %T", shift)
	}

	// Verify ID is generated
	if shift.ID == uuid.Nil {
		t.Errorf("Expected ID to be generated, got nil UUID")
	}
}

// TestMapToShiftInstanceConsistency tests that the same input always produces consistent results.
func TestMapToShiftInstanceConsistency(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:             "2025-11-20",
		ShiftType:        "DAY",
		RequiredStaffing: "3",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift1, err1 := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)
	shift2, err2 := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err1 != nil || err2 != nil {
		t.Fatalf("Expected no error, got err1=%v, err2=%v", err1, err2)
	}

	// Fields should be consistent (except ID and timestamps which are generated)
	if shift1.ShiftType != shift2.ShiftType {
		t.Errorf("ShiftType mismatch: %s vs %s", shift1.ShiftType, shift2.ShiftType)
	}

	if shift1.ScheduleVersionID != shift2.ScheduleVersionID {
		t.Errorf("ScheduleVersionID mismatch")
	}

	if shift1.CreatedBy != shift2.CreatedBy {
		t.Errorf("CreatedBy mismatch")
	}
}

// TestMapToShiftInstanceWhitespaceHandling tests proper trimming of whitespace.
func TestMapToShiftInstanceWhitespaceHandling(t *testing.T) {
	mapper := NewShiftInstanceMapper(time.UTC)
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawShiftData{
		Date:                "  2025-11-20  ",
		ShiftType:           "  DAY  ",
		RequiredStaffing:    "  3  ",
		SpecialtyConstraint: "  Cardiology  ",
		RowMetadata: RowMetadata{
			Row: 2,
		},
	}

	shift, err := mapper.MapToShiftInstance(raw, scheduleVersionID, userID)

	if err != nil {
		t.Fatalf("Expected no error with whitespace, got: %v", err)
	}

	if shift == nil {
		t.Fatalf("Expected shift to be non-nil")
	}

	if shift.ShiftType != "DAY" {
		t.Errorf("Expected ShiftType 'DAY' (trimmed), got %s", shift.ShiftType)
	}

	if shift.SpecialtyConstraint == nil || *shift.SpecialtyConstraint != "Cardiology" {
		t.Errorf("Expected SpecialtyConstraint 'Cardiology' (trimmed), got %v", shift.SpecialtyConstraint)
	}
}
