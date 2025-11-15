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

// AssignmentRepository implements repository.AssignmentRepository for PostgreSQL
type AssignmentRepository struct {
	db *sql.DB
}

// NewAssignmentRepository creates a new AssignmentRepository
func NewAssignmentRepository(db *sql.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// Create creates a new assignment
func (r *AssignmentRepository) Create(ctx context.Context, assignment *entity.Assignment) error {
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}

	query := `
		INSERT INTO assignments (id, person_id, shift_instance_id, schedule_date, original_shift_type, source, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		assignment.ID,
		assignment.PersonID,
		assignment.ShiftInstanceID,
		assignment.ScheduleDate,
		assignment.OriginalShiftType,
		string(assignment.Source),
		assignment.CreatedAt,
		assignment.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	return nil
}

// GetByID retrieves an assignment by ID
func (r *AssignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Assignment, error) {
	assignment := &entity.Assignment{}

	query := `
		SELECT id, person_id, shift_instance_id, schedule_date, original_shift_type, source,
		       created_at, created_by, deleted_at, deleted_by
		FROM assignments
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&assignment.ID,
		&assignment.PersonID,
		&assignment.ShiftInstanceID,
		&assignment.ScheduleDate,
		&assignment.OriginalShiftType,
		(*string)(&assignment.Source),
		&assignment.CreatedAt,
		&assignment.CreatedBy,
		&assignment.DeletedAt,
		&assignment.DeletedBy,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "Assignment",
			ResourceID:   id.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	return assignment, nil
}

// GetByShiftInstance retrieves all assignments for a shift instance
func (r *AssignmentRepository) GetByShiftInstance(ctx context.Context, shiftInstanceID uuid.UUID) ([]*entity.Assignment, error) {
	query := `
		SELECT id, person_id, shift_instance_id, schedule_date, original_shift_type, source,
		       created_at, created_by, deleted_at, deleted_by
		FROM assignments
		WHERE shift_instance_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, shiftInstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*entity.Assignment
	for rows.Next() {
		assignment := &entity.Assignment{}
		err := rows.Scan(
			&assignment.ID,
			&assignment.PersonID,
			&assignment.ShiftInstanceID,
			&assignment.ScheduleDate,
			&assignment.OriginalShiftType,
			(*string)(&assignment.Source),
			&assignment.CreatedAt,
			&assignment.CreatedBy,
			&assignment.DeletedAt,
			&assignment.DeletedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignments: %w", err)
	}

	return assignments, nil
}

// GetByPerson retrieves all assignments for a person
func (r *AssignmentRepository) GetByPerson(ctx context.Context, personID uuid.UUID) ([]*entity.Assignment, error) {
	query := `
		SELECT id, person_id, shift_instance_id, schedule_date, original_shift_type, source,
		       created_at, created_by, deleted_at, deleted_by
		FROM assignments
		WHERE person_id = $1 AND deleted_at IS NULL
		ORDER BY schedule_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*entity.Assignment
	for rows.Next() {
		assignment := &entity.Assignment{}
		err := rows.Scan(
			&assignment.ID,
			&assignment.PersonID,
			&assignment.ShiftInstanceID,
			&assignment.ScheduleDate,
			&assignment.OriginalShiftType,
			(*string)(&assignment.Source),
			&assignment.CreatedAt,
			&assignment.CreatedBy,
			&assignment.DeletedAt,
			&assignment.DeletedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignments: %w", err)
	}

	return assignments, nil
}

// GetByPersonAndDateRange retrieves assignments for a person within a date range
func (r *AssignmentRepository) GetByPersonAndDateRange(ctx context.Context, personID uuid.UUID, startDate, endDate entity.Date) ([]*entity.Assignment, error) {
	query := `
		SELECT id, person_id, shift_instance_id, schedule_date, original_shift_type, source,
		       created_at, created_by, deleted_at, deleted_by
		FROM assignments
		WHERE person_id = $1 AND schedule_date >= $2 AND schedule_date <= $3 AND deleted_at IS NULL
		ORDER BY schedule_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, personID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*entity.Assignment
	for rows.Next() {
		assignment := &entity.Assignment{}
		err := rows.Scan(
			&assignment.ID,
			&assignment.PersonID,
			&assignment.ShiftInstanceID,
			&assignment.ScheduleDate,
			&assignment.OriginalShiftType,
			(*string)(&assignment.Source),
			&assignment.CreatedAt,
			&assignment.CreatedBy,
			&assignment.DeletedAt,
			&assignment.DeletedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignments: %w", err)
	}

	return assignments, nil
}

// GetByScheduleVersion retrieves all assignments for a schedule version
func (r *AssignmentRepository) GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.Assignment, error) {
	query := `
		SELECT a.id, a.person_id, a.shift_instance_id, a.schedule_date, a.original_shift_type, a.source,
		       a.created_at, a.created_by, a.deleted_at, a.deleted_by
		FROM assignments a
		INNER JOIN shift_instances si ON a.shift_instance_id = si.id
		WHERE si.schedule_version_id = $1 AND a.deleted_at IS NULL
		ORDER BY a.schedule_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, scheduleVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*entity.Assignment
	for rows.Next() {
		assignment := &entity.Assignment{}
		err := rows.Scan(
			&assignment.ID,
			&assignment.PersonID,
			&assignment.ShiftInstanceID,
			&assignment.ScheduleDate,
			&assignment.OriginalShiftType,
			(*string)(&assignment.Source),
			&assignment.CreatedAt,
			&assignment.CreatedBy,
			&assignment.DeletedAt,
			&assignment.DeletedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignments: %w", err)
	}

	return assignments, nil
}

// Update updates an assignment
func (r *AssignmentRepository) Update(ctx context.Context, assignment *entity.Assignment) error {
	query := `
		UPDATE assignments
		SET person_id = $2, shift_instance_id = $3, schedule_date = $4, original_shift_type = $5, source = $6
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		assignment.ID,
		assignment.PersonID,
		assignment.ShiftInstanceID,
		assignment.ScheduleDate,
		assignment.OriginalShiftType,
		string(assignment.Source),
	)

	if err != nil {
		return fmt.Errorf("failed to update assignment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &repository.NotFoundError{
			ResourceType: "Assignment",
			ResourceID:   assignment.ID.String(),
		}
	}

	return nil
}

// Delete marks an assignment as deleted
func (r *AssignmentRepository) Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error {
	query := `
		UPDATE assignments
		SET deleted_at = NOW(), deleted_by = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, deleterID)
	if err != nil {
		return fmt.Errorf("failed to delete assignment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &repository.NotFoundError{
			ResourceType: "Assignment",
			ResourceID:   id.String(),
		}
	}

	return nil
}

// Count returns the count of active assignments
func (r *AssignmentRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	query := `
		SELECT COUNT(*) FROM assignments
		WHERE deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count assignments: %w", err)
	}

	return count, nil
}

// GetAllByShiftIDs retrieves all assignments for multiple shift instance IDs (batch query for N+1 prevention)
func (r *AssignmentRepository) GetAllByShiftIDs(ctx context.Context, shiftInstanceIDs []uuid.UUID) ([]*entity.Assignment, error) {
	if len(shiftInstanceIDs) == 0 {
		return []*entity.Assignment{}, nil
	}

	query := `
		SELECT id, person_id, shift_instance_id, schedule_date, original_shift_type, source,
		       created_at, created_by, deleted_at, deleted_by
		FROM assignments
		WHERE shift_instance_id = ANY($1) AND deleted_at IS NULL
		ORDER BY shift_instance_id, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(shiftInstanceIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*entity.Assignment
	for rows.Next() {
		assignment := &entity.Assignment{}
		err := rows.Scan(
			&assignment.ID,
			&assignment.PersonID,
			&assignment.ShiftInstanceID,
			&assignment.ScheduleDate,
			&assignment.OriginalShiftType,
			(*string)(&assignment.Source),
			&assignment.CreatedAt,
			&assignment.CreatedBy,
			&assignment.DeletedAt,
			&assignment.DeletedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assignments: %w", err)
	}

	return assignments, nil
}
