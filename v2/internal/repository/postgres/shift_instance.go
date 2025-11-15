package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
)

// ShiftInstanceRepository implements repository.ShiftInstanceRepository for PostgreSQL
type ShiftInstanceRepository struct {
	db *sql.DB
}

// NewShiftInstanceRepository creates a new ShiftInstanceRepository
func NewShiftInstanceRepository(db *sql.DB) *ShiftInstanceRepository {
	return &ShiftInstanceRepository{db: db}
}

// scanShiftInstance scans shift instance row data handling nullable string fields
func scanShiftInstance(scanner interface{ Scan(...interface{}) error }, shift *entity.ShiftInstance) error {
	var startTime sql.NullString
	var endTime sql.NullString
	var studyType sql.NullString
	var specialtyConstraint sql.NullString

	err := scanner.Scan(
		&shift.ID,
		&shift.ScheduleVersionID,
		&shift.HospitalID,
		(*string)(&shift.ShiftType),
		&shift.ScheduleDate,
		&startTime,
		&endTime,
		&studyType,
		&specialtyConstraint,
		&shift.DesiredCoverage,
		&shift.IsMandatory,
		&shift.CreatedAt,
		&shift.CreatedBy,
	)

	if err == nil {
		if startTime.Valid {
			shift.StartTime = startTime.String
		}
		if endTime.Valid {
			shift.EndTime = endTime.String
		}
		if studyType.Valid {
			shift.StudyType = entity.StudyType(studyType.String)
		}
		if specialtyConstraint.Valid {
			shift.SpecialtyConstraint = entity.SpecialtyType(specialtyConstraint.String)
		}
	}

	return err
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
			is_mandatory, created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
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
		       is_mandatory, created_at, created_by
		FROM shift_instances
		WHERE id = $1
	`

	err := scanShiftInstance(r.db.QueryRowContext(ctx, query, id), shift)

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
func (r *ShiftInstanceRepository) GetByScheduleVersion(ctx context.Context, versionID uuid.UUID) ([]*entity.ShiftInstance, error) {
	query := `
		SELECT id, schedule_version_id, hospital_id, shift_type, schedule_date,
		       start_time, end_time, study_type, specialty_constraint, desired_coverage,
		       is_mandatory, created_at, created_by
		FROM shift_instances
		WHERE schedule_version_id = $1
		ORDER BY schedule_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query shifts by schedule version: %w", err)
	}
	defer rows.Close()

	var shifts []*entity.ShiftInstance
	for rows.Next() {
		shift := &entity.ShiftInstance{}
		if err := scanShiftInstance(rows, shift); err != nil {
			return nil, fmt.Errorf("failed to scan shift instance: %w", err)
		}
		shifts = append(shifts, shift)
	}

	return shifts, rows.Err()
}

// GetByDateRange retrieves shifts within a date range for a schedule version
func (r *ShiftInstanceRepository) GetByDateRange(ctx context.Context, versionID uuid.UUID, startDate, endDate entity.Date) ([]*entity.ShiftInstance, error) {
	query := `
		SELECT id, schedule_version_id, hospital_id, shift_type, schedule_date,
		       start_time, end_time, study_type, specialty_constraint, desired_coverage,
		       is_mandatory, created_at, created_by
		FROM shift_instances
		WHERE schedule_version_id = $1 AND schedule_date >= $2 AND schedule_date <= $3
		ORDER BY schedule_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, versionID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query shifts by date range: %w", err)
	}
	defer rows.Close()

	var shifts []*entity.ShiftInstance
	for rows.Next() {
		shift := &entity.ShiftInstance{}
		if err := scanShiftInstance(rows, shift); err != nil {
			return nil, fmt.Errorf("failed to scan shift instance: %w", err)
		}
		shifts = append(shifts, shift)
	}

	return shifts, rows.Err()
}

// GetAllByShiftIDs retrieves multiple shifts by their IDs (batch operation)
func (r *ShiftInstanceRepository) GetAllByShiftIDs(ctx context.Context, shiftIDs []uuid.UUID) ([]*entity.ShiftInstance, error) {
	if len(shiftIDs) == 0 {
		return []*entity.ShiftInstance{}, nil
	}

	query := `
		SELECT id, schedule_version_id, hospital_id, shift_type, schedule_date,
		       start_time, end_time, study_type, specialty_constraint, desired_coverage,
		       is_mandatory, created_at, created_by
		FROM shift_instances
		WHERE id = ANY($1)
		ORDER BY schedule_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(shiftIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query shifts by IDs: %w", err)
	}
	defer rows.Close()

	var shifts []*entity.ShiftInstance
	for rows.Next() {
		shift := &entity.ShiftInstance{}
		if err := scanShiftInstance(rows, shift); err != nil {
			return nil, fmt.Errorf("failed to scan shift instance: %w", err)
		}
		shifts = append(shifts, shift)
	}

	return shifts, rows.Err()
}

// Count returns the total number of shifts
func (r *ShiftInstanceRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM shift_instances`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count shifts: %w", err)
	}
	return count, nil
}

// CountByScheduleVersion returns the number of shifts in a schedule version
func (r *ShiftInstanceRepository) CountByScheduleVersion(ctx context.Context, versionID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM shift_instances WHERE schedule_version_id = $1`
	err := r.db.QueryRowContext(ctx, query, versionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count shifts by version: %w", err)
	}
	return count, nil
}
