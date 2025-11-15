package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
)

// AssignmentRepository defines the interface for assignment data access.
type AssignmentRepository interface {
	// Create inserts a new assignment.
	// Returns an error if the assignment cannot be created (e.g., person or shift doesn't exist).
	// May also return a constraint violation error if the person-shift combination already exists.
	Create(ctx context.Context, assignment *entity.Assignment) (*entity.Assignment, error)

	// GetByID retrieves an assignment by its ID.
	// Returns ErrNotFound if the assignment does not exist.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Assignment, error)

	// GetByPersonAndShift retrieves an active assignment for a given person and shift.
	// Returns ErrNotFound if no active assignment exists.
	GetByPersonAndShift(ctx context.Context, personID uuid.UUID, shiftInstanceID uuid.UUID) (*entity.Assignment, error)

	// GetByPerson retrieves all active assignments for a person.
	// Returns an empty slice if no assignments exist (not an error).
	GetByPerson(ctx context.Context, personID uuid.UUID) ([]*entity.Assignment, error)

	// GetByShiftInstance retrieves all active assignments for a shift.
	// Returns an empty slice if no assignments exist (not an error).
	GetByShiftInstance(ctx context.Context, shiftInstanceID uuid.UUID) ([]*entity.Assignment, error)

	// GetByScheduleVersion retrieves all active assignments in a schedule version.
	// Returns an empty slice if no assignments exist (not an error).
	GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.Assignment, error)

	// CreateBatch inserts multiple assignments in a single operation.
	// Returns the number of successfully created assignments.
	// If any assignment fails, an error is returned with details of which assignments failed.
	CreateBatch(ctx context.Context, assignments []*entity.Assignment) (int, error)

	// Update modifies an existing assignment.
	// Returns an error if the assignment does not exist.
	Update(ctx context.Context, assignment *entity.Assignment) error

	// Delete soft-deletes an assignment.
	Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error

	// DeleteByScheduleVersion soft-deletes all assignments in a schedule version.
	// Returns the number of assignments deleted.
	DeleteByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID, deleterID uuid.UUID) (int, error)
}
