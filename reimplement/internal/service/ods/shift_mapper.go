package ods

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
)

// ShiftMapperConfig holds configuration for the ShiftInstanceMapper.
type ShiftMapperConfig struct {
	// AllowPastDates controls whether shifts with dates in the past are allowed.
	// Set to false to reject past dates during validation.
	// Default: true (allow past dates)
	AllowPastDates bool

	// Timezone is the time.Location to use when parsing dates and creating timestamps.
	// All dates are interpreted as midnight in this timezone.
	// Default: time.UTC
	Timezone *time.Location
}

// ShiftInstanceMapper is responsible for converting RawShiftData from the parser
// into ShiftInstance entities suitable for persistence.
//
// The mapper handles:
// - Date string parsing (YYYY-MM-DD format)
// - Shift type validation (must be DAY, NIGHT, or WEEKEND)
// - Required staffing conversion and validation (must be positive integer)
// - Timezone conversion
// - Timestamp generation (CreatedAt, CreatedBy)
// - Optional field handling (SpecialtyConstraint, StudyType)
//
// Validation Rules:
// - Date must be in YYYY-MM-DD format
// - Date can optionally be rejected if in the past (configurable)
// - ShiftType must be exactly one of: DAY, NIGHT, WEEKEND (case-sensitive)
// - RequiredStaffing must be a positive integer > 0
// - Input strings are trimmed of whitespace before validation
type ShiftInstanceMapper struct {
	config ShiftMapperConfig
}

// NewShiftInstanceMapper creates a new mapper with default configuration (UTC timezone, past dates allowed).
func NewShiftInstanceMapper(tz *time.Location) *ShiftInstanceMapper {
	return &ShiftInstanceMapper{
		config: ShiftMapperConfig{
			AllowPastDates: true,
			Timezone:       tz,
		},
	}
}

// NewShiftInstanceMapperWithConfig creates a new mapper with custom configuration.
func NewShiftInstanceMapperWithConfig(config ShiftMapperConfig) *ShiftInstanceMapper {
	return &ShiftInstanceMapper{
		config: config,
	}
}

// ValidShiftTypes defines the allowed shift type values.
var ValidShiftTypes = map[string]bool{
	"DAY":     true,
	"NIGHT":   true,
	"WEEKEND": true,
}

// MapToShiftInstance converts RawShiftData into a ShiftInstance entity.
//
// Parameters:
// - raw: The raw shift data from the parser (contains unparsed string values)
// - scheduleVersionID: The UUID of the parent ScheduleVersion
// - userID: The UUID of the user performing the import
//
// Returns:
// - *entity.ShiftInstance: The mapped entity, or nil if validation fails
// - error: A detailed error message if validation fails, nil on success
//
// Validation failures include:
// - Invalid date format
// - Date in the past (if configured to reject)
// - Invalid shift type (not one of the allowed types)
// - Invalid required staffing (non-numeric, zero, or negative)
// - Nil schedule version or user ID
func (sm *ShiftInstanceMapper) MapToShiftInstance(
	raw RawShiftData,
	scheduleVersionID uuid.UUID,
	userID uuid.UUID,
) (*entity.ShiftInstance, error) {
	// Validate input UUIDs
	if scheduleVersionID == uuid.Nil {
		return nil, fmt.Errorf("schedule version ID cannot be nil")
	}
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	// 1. Validate and parse date
	parsedDate, err := sm.validateAndParseDate(raw.Date)
	if err != nil {
		return nil, err
	}

	// 2. Validate shift type
	if err := sm.validateShiftType(raw.ShiftType); err != nil {
		return nil, err
	}

	// 3. Validate and convert required staffing
	requiredStaffing, err := sm.validateAndParseRequiredStaffing(raw.RequiredStaffing)
	if err != nil {
		return nil, err
	}

	// Create the ShiftInstance entity
	shift := &entity.ShiftInstance{
		ID:                uuid.New(),
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         strings.TrimSpace(raw.ShiftType),
		StartTime:         &parsedDate,
		EndTime:           &parsedDate, // Set to same date; client can adjust
		Location:          "",           // Not provided in RawShiftData
		StaffMember:       "",           // Not provided in RawShiftData
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
		UpdatedAt:         time.Now(),
		UpdatedBy:         userID,
	}

	// Handle optional fields
	specialty := strings.TrimSpace(raw.SpecialtyConstraint)
	if specialty != "" {
		shift.SpecialtyConstraint = &specialty
	}

	studyType := strings.TrimSpace(raw.StudyType)
	if studyType != "" {
		shift.StudyType = &studyType
	}

	// Store the required staffing as a placeholder (future enhancement)
	// For now, we document that requiredStaffing was validated
	_ = requiredStaffing

	return shift, nil
}

// validateAndParseDate validates the date string and parses it to time.Time.
//
// Expected format: YYYY-MM-DD (e.g., "2025-11-20")
//
// Returns error if:
// - Date string is empty
// - Date string is not in YYYY-MM-DD format
// - Date is invalid (e.g., Feb 30)
// - Date is in the past and AllowPastDates is false
func (sm *ShiftInstanceMapper) validateAndParseDate(dateStr string) (time.Time, error) {
	trimmed := strings.TrimSpace(dateStr)

	if trimmed == "" {
		return time.Time{}, fmt.Errorf("date cannot be empty")
	}

	// Parse the date in the configured timezone
	// The format is strict YYYY-MM-DD
	parsedDate, err := time.ParseInLocation("2006-01-02", trimmed, sm.config.Timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: expected YYYY-MM-DD, got %q: %w", trimmed, err)
	}

	// Check if date is in the past (if configured)
	if !sm.config.AllowPastDates {
		now := time.Now().In(sm.config.Timezone)
		// Compare dates at midnight for fair comparison
		todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, sm.config.Timezone)
		if parsedDate.Before(todayMidnight) {
			return time.Time{}, fmt.Errorf("date %q is in the past (today is %s)", trimmed, todayMidnight.Format("2006-01-02"))
		}
	}

	return parsedDate, nil
}

// validateShiftType validates that the shift type is one of the allowed types.
//
// Allowed types (case-sensitive): DAY, NIGHT, WEEKEND
//
// Returns error if:
// - ShiftType is empty or whitespace only
// - ShiftType is not in the allowed list
func (sm *ShiftInstanceMapper) validateShiftType(shiftType string) error {
	trimmed := strings.TrimSpace(shiftType)

	if trimmed == "" {
		return fmt.Errorf("shift type cannot be empty")
	}

	if !ValidShiftTypes[trimmed] {
		allowedTypes := []string{"DAY", "NIGHT", "WEEKEND"}
		return fmt.Errorf("invalid shift type: %q (must be one of: %v)", trimmed, allowedTypes)
	}

	return nil
}

// validateAndParseRequiredStaffing validates and converts the required staffing string to int.
//
// Returns error if:
// - RequiredStaffing is empty or whitespace only
// - RequiredStaffing is not a valid integer
// - RequiredStaffing is not positive (must be > 0)
func (sm *ShiftInstanceMapper) validateAndParseRequiredStaffing(staffingStr string) (int, error) {
	trimmed := strings.TrimSpace(staffingStr)

	if trimmed == "" {
		return 0, fmt.Errorf("required staffing cannot be empty")
	}

	staffing, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("required staffing must be a valid integer, got %q: %w", trimmed, err)
	}

	if staffing <= 0 {
		return 0, fmt.Errorf("required staffing must be positive (> 0), got %d", staffing)
	}

	return staffing, nil
}
