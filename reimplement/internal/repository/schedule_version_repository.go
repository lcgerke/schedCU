package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
)

// ScheduleVersionRepository defines the interface for schedule version data access.
type ScheduleVersionRepository interface {
	// Create inserts a new schedule version and returns the created entity.
	// Returns an error if the version already exists or other database constraints are violated.
	Create(ctx context.Context, sv *entity.ScheduleVersion) (*entity.ScheduleVersion, error)

	// GetByID retrieves a schedule version by its ID.
	// Returns ErrNotFound if the version does not exist.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ScheduleVersion, error)

	// GetByHospitalAndVersion retrieves a specific version for a hospital.
	// Returns ErrNotFound if not found.
	GetByHospitalAndVersion(ctx context.Context, hospitalID uuid.UUID, version int) (*entity.ScheduleVersion, error)

	// GetLatestByHospital retrieves the most recent version for a hospital.
	// Returns ErrNotFound if no versions exist for the hospital.
	GetLatestByHospital(ctx context.Context, hospitalID uuid.UUID) (*entity.ScheduleVersion, error)

	// List retrieves all non-deleted schedule versions for a hospital.
	List(ctx context.Context, hospitalID uuid.UUID) ([]*entity.ScheduleVersion, error)

	// Update modifies an existing schedule version.
	// Returns an error if the version does not exist.
	Update(ctx context.Context, sv *entity.ScheduleVersion) error

	// Delete soft-deletes a schedule version.
	Delete(ctx context.Context, id uuid.UUID) error
}
