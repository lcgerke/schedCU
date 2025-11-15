package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
)

// ShiftInstanceRepository defines the interface for shift instance data access.
type ShiftInstanceRepository interface {
	// Create inserts a new shift instance.
	// Returns an error if the shift cannot be created (e.g., schedule version doesn't exist).
	Create(ctx context.Context, shift *entity.ShiftInstance) (*entity.ShiftInstance, error)

	// GetByID retrieves a shift instance by its ID.
	// Returns ErrNotFound if the shift does not exist.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ShiftInstance, error)

	// GetByScheduleVersion retrieves all shifts in a schedule version.
	// Returns an empty slice if no shifts exist (not an error).
	GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.ShiftInstance, error)

	// CreateBatch inserts multiple shift instances in a single operation.
	// Returns the number of successfully created shifts.
	// If any shift fails, an error is returned with details of which shifts failed.
	CreateBatch(ctx context.Context, shifts []*entity.ShiftInstance) (int, error)

	// Update modifies an existing shift instance.
	// Returns an error if the shift does not exist.
	Update(ctx context.Context, shift *entity.ShiftInstance) error

	// Delete soft-deletes a shift instance.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByScheduleVersion soft-deletes all shifts in a schedule version.
	// Returns the number of shifts deleted.
	DeleteByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) (int, error)
}
