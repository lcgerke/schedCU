package entity

import (
	"time"

	"github.com/google/uuid"
)

// AssignmentSource represents the origin of an assignment.
type AssignmentSource string

const (
	AssignmentSourceAmion    AssignmentSource = "AMION"
	AssignmentSourceManual   AssignmentSource = "MANUAL"
	AssignmentSourceOverride AssignmentSource = "OVERRIDE"
)

// Assignment represents a mapping between a person and a shift instance.
// It tracks which staff member is assigned to which shift and the source of the assignment.
type Assignment struct {
	ID              uuid.UUID
	PersonID        uuid.UUID
	ShiftInstanceID uuid.UUID
	ScheduleDate    time.Time
	OriginalShiftType string // Original name from Amion (e.g., "Technologist")
	Source          AssignmentSource
	CreatedAt       time.Time
	CreatedBy       uuid.UUID
	DeletedAt       *time.Time
	DeletedBy       *uuid.UUID
}

// NewAssignment creates a new Assignment with default values.
func NewAssignment(
	personID uuid.UUID,
	shiftInstanceID uuid.UUID,
	scheduleDate time.Time,
	originalShiftType string,
	source AssignmentSource,
	userID uuid.UUID,
) *Assignment {
	now := time.Now()
	return &Assignment{
		ID:                uuid.New(),
		PersonID:          personID,
		ShiftInstanceID:   shiftInstanceID,
		ScheduleDate:      scheduleDate,
		OriginalShiftType: originalShiftType,
		Source:            source,
		CreatedAt:         now,
		CreatedBy:         userID,
	}
}

// IsValid performs basic validation of the assignment.
// Returns true if all required fields are present and valid.
func (a *Assignment) IsValid() bool {
	return a.ID != uuid.Nil &&
		a.PersonID != uuid.Nil &&
		a.ShiftInstanceID != uuid.Nil &&
		!a.ScheduleDate.IsZero() &&
		a.Source != "" &&
		a.CreatedBy != uuid.Nil
}

// IsDeleted checks if this assignment has been soft-deleted.
func (a *Assignment) IsDeleted() bool {
	return a.DeletedAt != nil
}

// SoftDelete marks the assignment as deleted.
func (a *Assignment) SoftDelete(deleterID uuid.UUID) {
	now := time.Now()
	a.DeletedAt = &now
	a.DeletedBy = &deleterID
}
