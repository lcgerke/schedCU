package entity

import (
	"time"

	"github.com/google/uuid"
)

// ShiftInstance represents an individual shift assignment within a ScheduleVersion.
// Each shift belongs to exactly one schedule version and represents
// a specific position/role assignment with time boundaries.
type ShiftInstance struct {
	ID                    uuid.UUID
	ScheduleVersionID     uuid.UUID
	ShiftType             string // e.g., "Morning", "Afternoon", "Night"
	Position              string // e.g., "Senior Doctor", "Nurse", "Admin"
	StartTime             *time.Time
	EndTime               *time.Time
	Location              string
	StaffMember           string // Person name or identifier
	SpecialtyConstraint   *string
	StudyType             *string
	RequiredQualification string
	CreatedAt             time.Time
	CreatedBy             uuid.UUID
	UpdatedAt             time.Time
	UpdatedBy             uuid.UUID
	DeletedAt             *time.Time
}

// NewShiftInstance creates a new ShiftInstance with default values.
func NewShiftInstance(
	scheduleVersionID uuid.UUID,
	shiftType string,
	position string,
	location string,
	staffMember string,
	userID uuid.UUID,
) *ShiftInstance {
	now := time.Now()
	return &ShiftInstance{
		ID:                uuid.New(),
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         shiftType,
		Position:          position,
		Location:          location,
		StaffMember:       staffMember,
		CreatedAt:         now,
		CreatedBy:         userID,
		UpdatedAt:         now,
		UpdatedBy:         userID,
	}
}

// IsValid performs basic validation of the shift instance.
// Returns true if all required fields are present and valid.
func (si *ShiftInstance) IsValid() bool {
	return si.ID != uuid.Nil &&
		si.ScheduleVersionID != uuid.Nil &&
		si.ShiftType != "" &&
		si.Position != "" &&
		si.Location != "" &&
		si.StaffMember != ""
}
