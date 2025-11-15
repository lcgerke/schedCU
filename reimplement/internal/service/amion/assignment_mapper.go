package amion

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/repository"
)

// AssignmentMapper converts RawAmionShift data into Assignment entities.
// It handles the mapping between Amion scraped data and domain model assignments,
// validating that shift instances exist and are not soft-deleted before creating assignments.
//
// The mapper is responsible for:
// - Validating all input parameters (person ID, schedule version ID, user ID)
// - Verifying that the shift instance exists and is not soft-deleted
// - Parsing the date from the RawAmionShift
// - Creating an Assignment entity with proper timestamps and audit information
// - Setting the source field to "AMION"
// - Preserving the original shift type from Amion for audit purposes
//
// Error handling:
// - Returns error if person ID is nil
// - Returns error if schedule version ID is nil
// - Returns error if user ID is nil
// - Returns error if shift instance is nil
// - Returns error if shift instance is soft-deleted
// - Returns error if date cannot be parsed
type AssignmentMapper struct {
	// Add any configuration fields if needed in the future
}

// NewAssignmentMapper creates a new AssignmentMapper instance.
func NewAssignmentMapper() *AssignmentMapper {
	return &AssignmentMapper{}
}

// MapToAssignment converts a RawAmionShift into an Assignment entity.
//
// Parameters:
// - ctx: Context for the operation
// - raw: The raw shift data from Amion scraper
// - personID: The UUID of the person being assigned
// - shiftInstance: The shift instance entity to assign the person to
// - scheduleVersionID: The UUID of the parent schedule version
// - userID: The UUID of the user creating the assignment
// - shiftRepo: The shift instance repository (for future use in constraint checking)
//
// Returns:
// - *entity.Assignment: The mapped assignment entity, or nil if validation fails
// - error: A detailed error message if validation fails, nil on success
//
// Validation failures include:
// - Nil person ID
// - Nil schedule version ID
// - Nil user ID
// - Nil shift instance
// - Soft-deleted shift instance
// - Invalid date format in raw shift
//
// The returned Assignment will have:
// - Source set to "AMION"
// - CreatedAt set to current time
// - CreatedBy set to the provided userID
// - OriginalShiftType set to the shift type from RawAmionShift
// - ScheduleDate parsed from the date string in YYYY-MM-DD format
// - DeletedAt and DeletedBy set to nil (new assignment)
func (am *AssignmentMapper) MapToAssignment(
	ctx context.Context,
	raw RawAmionShift,
	personID uuid.UUID,
	shiftInstance *entity.ShiftInstance,
	scheduleVersionID uuid.UUID,
	userID uuid.UUID,
	shiftRepo repository.ShiftInstanceRepository,
) (*entity.Assignment, error) {
	// Validate all input UUIDs
	if personID == uuid.Nil {
		return nil, fmt.Errorf("person ID cannot be nil")
	}

	if scheduleVersionID == uuid.Nil {
		return nil, fmt.Errorf("schedule version ID cannot be nil")
	}

	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	// Validate shift instance exists
	if shiftInstance == nil {
		return nil, fmt.Errorf("shift instance cannot be nil: shift not found for assignment")
	}

	// Validate shift instance is not soft-deleted
	if shiftInstance.DeletedAt != nil {
		return nil, fmt.Errorf("shift instance has been deleted: cannot create assignment to deleted shift")
	}

	// Validate shift instance belongs to the correct schedule version
	if shiftInstance.ScheduleVersionID != scheduleVersionID {
		return nil, fmt.Errorf("shift instance belongs to different schedule version: expected %s, got %s", scheduleVersionID, shiftInstance.ScheduleVersionID)
	}

	// Parse the date from RawAmionShift
	scheduleDate, err := am.parseDate(raw.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse assignment date: %w", err)
	}

	// Create the Assignment entity
	assignment := &entity.Assignment{
		ID:                uuid.New(),
		PersonID:          personID,
		ShiftInstanceID:   shiftInstance.ID,
		ScheduleDate:      scheduleDate,
		OriginalShiftType: raw.ShiftType,
		Source:            entity.AssignmentSourceAmion,
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	return assignment, nil
}

// parseDate parses a date string in YYYY-MM-DD format.
//
// Returns error if:
// - Date string is empty
// - Date string is not in YYYY-MM-DD format
// - Date is invalid (e.g., Feb 30)
func (am *AssignmentMapper) parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date cannot be empty")
	}

	// Parse the date in UTC timezone
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: expected YYYY-MM-DD, got %q: %w", dateStr, err)
	}

	return parsedDate, nil
}
