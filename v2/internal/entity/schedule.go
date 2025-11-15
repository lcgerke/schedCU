package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Schedule represents a hospital schedule covering a date range.
// Translated from v1 ScheduleVersion with improvements: proper soft delete, audit trail, no deprecated fields.
// This is a simplified representation for the basic Phase 0 API.
// For full schema, use ScheduleVersion with ShiftInstance array.
type Schedule struct {
	ID          uuid.UUID      `json:"id"`
	HospitalID  uuid.UUID      `json:"hospital_id"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	Source      string         `json:"source"` // 'amion', 'ods_file', 'manual'
	SourceID    *string        `json:"source_id,omitempty"`
	Assignments []interface{}  `json:"assignments"`
	CreatedAt   time.Time      `json:"created_at"`
	CreatedBy   uuid.UUID      `json:"created_by"`
	UpdatedAt   time.Time      `json:"updated_at"`
	UpdatedBy   uuid.UUID      `json:"updated_by"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty"`
	DeletedBy   *uuid.UUID     `json:"deleted_by,omitempty"`
}

// ValidationResult provides structured validation/error response (from v1 pattern).
type ValidationResult struct {
	Valid      bool                   `json:"valid"`
	Code       string                 `json:"code"`                 // VALIDATION_SUCCESS, PARSE_ERROR, etc.
	Severity   string                 `json:"severity"`             // INFO, WARNING, ERROR
	Message    string                 `json:"message"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// NewSchedule creates a new schedule with generated ID and timestamps.
func NewSchedule(hospitalID, userID uuid.UUID, startDate, endDate time.Time, source string) *Schedule {
	now := time.Now().UTC()
	return &Schedule{
		ID:         uuid.New(),
		HospitalID: hospitalID,
		StartDate:  startDate,
		EndDate:    endDate,
		Source:     source,
		CreatedAt:  now,
		CreatedBy:  userID,
		UpdatedAt:  now,
		UpdatedBy:  userID,
	}
}

// AddAssignment adds a shift assignment to the schedule.
// Accepts any interface{} for flexibility (can be ShiftInstance or other assignment types).
func (s *Schedule) AddAssignment(assignment interface{}) {
	if assignment != nil {
		s.Assignments = append(s.Assignments, assignment)
	}
}

// IsDeleted returns true if the schedule is soft-deleted.
func (s *Schedule) IsDeleted() bool {
	return s.DeletedAt != nil
}

// SoftDelete marks the schedule as deleted without removing data.
func (s *Schedule) SoftDelete(deleterID uuid.UUID) {
	now := time.Now().UTC()
	s.DeletedAt = &now
	s.DeletedBy = &deleterID
}

// ValidateDateRange ensures end date is after or equal to start date.
func ValidateDateRange(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return errors.New("end_date must be after or equal to start_date")
	}
	return nil
}

// NewValidationResult creates a successful validation result.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:    true,
		Code:     "VALIDATION_SUCCESS",
		Severity: "INFO",
		Message:  "Validation passed",
		Context:  make(map[string]interface{}),
	}
}

// NewValidationError creates a validation error result.
func NewValidationError(code, message string) *ValidationResult {
	return &ValidationResult{
		Valid:    false,
		Code:     code,
		Severity: "ERROR",
		Message:  message,
		Context:  make(map[string]interface{}),
	}
}

// NewValidationWarning creates a validation warning result.
func NewValidationWarning(code, message string) *ValidationResult {
	return &ValidationResult{
		Valid:    true,
		Code:     code,
		Severity: "WARNING",
		Message:  message,
		Context:  make(map[string]interface{}),
	}
}

// AddContext adds contextual information to the validation result.
func (vr *ValidationResult) AddContext(key string, value interface{}) {
	if vr.Context == nil {
		vr.Context = make(map[string]interface{})
	}
	vr.Context[key] = value
}
