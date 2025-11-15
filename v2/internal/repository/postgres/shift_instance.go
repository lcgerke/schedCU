package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"schedcu/v2/internal/entity"
	"schedcu/v2/internal/repository"
)

// ShiftInstanceRepository implements repository.ShiftInstanceRepository for PostgreSQL
type ShiftInstanceRepository struct {
	db *sql.DB
}

// NewShiftInstanceRepository creates a new ShiftInstanceRepository
func NewShiftInstanceRepository(db *sql.DB) *ShiftInstanceRepository {
	return &ShiftInstanceRepository{db: db}
}

// Create creates a new shift instance
func (r *ShiftInstanceRepository) Create(ctx context.Context, shift *entity.ShiftInstance) error {
	if shift.ID == uuid.Nil {
		shift.ID = uuid.New()
	}

	query := `
		INSERT INTO shift_instances (
			id, schedule_version_id, hospital_id, shift_type, schedule_date,
			start_time, end_time, study_type, specialty_constraint, desired_coverage,
			is_mandatory, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.ExecContext(ctx, query,
		shift.ID,
		shift.ScheduleVersionID,
		shift.HospitalID,
		string(shift.ShiftType),
		shift.ScheduleDate,
		shift.StartTime,
		shift.EndTime,
		string(shift.StudyType),
		string(shift.SpecialtyConstraint),
		shift.DesiredCoverage,
		shift.IsMandatory,
		shift.CreatedAt,
		shift.CreatedBy,
		shift.UpdatedAt,
		shift.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create shift instance: %w", err)
	}

	return nil
}

// GetByID retrieves a shift instance by ID
func (r *ShiftInstanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ShiftInstance, error) {
	shift := &entity.ShiftInstance{}

	query := `
		SELECT id, schedule_version_id, hospital_id, shift_type, schedule_date,
		       start_time, end_time, study_type, specialty_constraint, desired_coverage,
		       is_mandatory, created_at, created_by, updated_at, updated_by, deleted_at
		FROM shift_instances
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&shift.ID,
		&shift.ScheduleVersionID,
		&shift.HospitalID,
		(*string)(&shift.ShiftType),
		&shift.ScheduleDate,
		&shift.StartTime,
		&shift.EndTime,
		(*string)(&shift.StudyType),
		(*string)(&shift.SpecialtyConstraint),
		&shift.DesiredCoverage,
		&shift.IsMandatory,
		&shift.CreatedAt,
		&shift.CreatedBy,
		&shift.UpdatedAt,
		&shift.UpdatedBy,
		&shift.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "ShiftInstance",
			ResourceID:   id.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get shift instance: %w", err)
	}

	return shift, nil
}

// GetByScheduleVersion retrieves all shifts for a schedule version
func (r *ShiftInstanceRepository) GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.ShiftInstance, error) {
	query := `
		SELECT id, schedule_version_id, hospital_id, shift_type, schedule_date,
		       start_time, end_time, study_type, specialty_constraint, desired_coverage,
		       is_mandatory, created_at, created_by, updated_at, updated_by, deleted_at
		FROM shift_instances
		WHERE schedule_version_id = $1 AND deleted_at IS NULL
		ORDER BY schedule_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, scheduleVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query shifts: %w", err)
	}
	defer rows.Close()

	var shifts []*entity.ShiftInstance
	for rows.Next() {
		shift := &entity.ShiftInstance{}
		err := rows.Scan(
			&shift.ID,
			&shift.ScheduleVersionID,
			&shift.HospitalID,
			(*string)(&shift.ShiftType),
			&shift.ScheduleDate,
			&shift.StartTime,
			&shift.EndTime,
			(*string)(&shift.StudyType),
			(*string)(&shift.SpecialtyConstraint),
			&shift.DesiredCoverage,
			&shift.IsMandatory,
			&shift.CreatedAt,
			&shift.CreatedBy,
			&shift.UpdatedAt,
			&shift.UpdatedBy,
			&shift.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shift: %w", err)
		}
		shifts = append(shifts, shift)
	}

	return shifts, rows.Err()
}

// GetByDateRange retrieves shifts within a date range for a schedule version
func (r *ShiftInstanceRepository) GetByDateRange(ctx context.Context, scheduleVersionID uuid.UUID, startDate, endDate interface{}) ([]*entity.ShiftInstance, error) {
	query := `
		SELECT id, schedule_version_id, hospital_id, shift_type, schedule_date,
		       start_time, end_time, study_type, specialty_constraint, desired_coverage,
		       is_mandatory, created_at, created_by, updated_at, updated_by, deleted_at
		FROM shift_instances
		WHERE schedule_version_id = $1 AND schedule_date BETWEEN $2 AND $3 AND deleted_at IS NULL
		ORDER BY schedule_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, scheduleVersionID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query shifts by date range: %w", err)
	}
	defer rows.Close()

	var shifts []*entity.ShiftInstance
	for rows.Next() {
		shift := &entity.ShiftInstance{}
		err := rows.Scan(
			&shift.ID,
			&shift.ScheduleVersionID,
			&shift.HospitalID,
			(*string)(&shift.ShiftType),
			&shift.ScheduleDate,
			&shift.StartTime,
			&shift.EndTime,
			(*string)(&shift.StudyType),
			(*string)(&shift.SpecialtyConstraint),
			&shift.DesiredCoverage,
			&shift.IsMandatory,
			&shift.CreatedAt,
			&shift.CreatedBy,
			&shift.UpdatedAt,
			&shift.UpdatedBy,
			&shift.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shift: %w", err)
		}
		shifts = append(shifts, shift)
	}

	return shifts, rows.Err()
}

// Update updates a shift instance
func (r *ShiftInstanceRepository) Update(ctx context.Context, shift *entity.ShiftInstance) error {
	query := `
		UPDATE shift_instances
		SET shift_type = $1, schedule_date = $2, start_time = $3, end_time = $4,
		    study_type = $5, specialty_constraint = $6, desired_coverage = $7,
		    is_mandatory = $8, updated_at = $9, updated_by = $10
		WHERE id = $11 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		string(shift.ShiftType),
		shift.ScheduleDate,
		shift.StartTime,
		shift.EndTime,
		string(shift.StudyType),
		string(shift.SpecialtyConstraint),
		shift.DesiredCoverage,
		shift.IsMandatory,
		shift.UpdatedAt,
		shift.UpdatedBy,
		shift.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update shift instance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "ShiftInstance",
			ResourceID:   shift.ID.String(),
		}
	}

	return nil
}

// Delete soft-deletes a shift instance
func (r *ShiftInstanceRepository) Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error {
	query := `
		UPDATE shift_instances
		SET deleted_at = NOW(), deleted_by = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, deleterID, id)
	if err != nil {
		return fmt.Errorf("failed to delete shift instance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "ShiftInstance",
			ResourceID:   id.String(),
		}
	}

	return nil
}

// Count returns the total count of non-deleted shifts
func (r *ShiftInstanceRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM shift_instances WHERE deleted_at IS NULL`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count shifts: %w", err)
	}

	return count, nil
}

// CountByScheduleVersion returns the count of shifts for a schedule version
func (r *ShiftInstanceRepository) CountByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*) FROM shift_instances
		WHERE schedule_version_id = $1 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRowContext(ctx, query, scheduleVersionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count shifts for schedule version: %w", err)
	}

	return count, nil
}
