package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"schedcu/v2/internal/entity"
	"schedcu/v2/internal/repository"
)

// ScheduleVersionRepository implements repository.ScheduleVersionRepository for PostgreSQL
type ScheduleVersionRepository struct {
	db *sql.DB
}

// NewScheduleVersionRepository creates a new ScheduleVersionRepository
func NewScheduleVersionRepository(db *sql.DB) *ScheduleVersionRepository {
	return &ScheduleVersionRepository{db: db}
}

// Create creates a new schedule version
func (r *ScheduleVersionRepository) Create(ctx context.Context, version *entity.ScheduleVersion) error {
	if version.ID == uuid.Nil {
		version.ID = uuid.New()
	}

	validationJSON, err := json.Marshal(version.ValidationResults)
	if err != nil {
		return fmt.Errorf("failed to marshal validation results: %w", err)
	}

	query := `
		INSERT INTO schedule_versions
		(id, hospital_id, status, effective_start_date, effective_end_date, scrape_batch_id, validation_results, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		version.ID,
		version.HospitalID,
		string(version.Status),
		version.EffectiveStartDate,
		version.EffectiveEndDate,
		version.ScrapeBatchID,
		validationJSON,
		version.CreatedAt,
		version.CreatedBy,
		version.UpdatedAt,
		version.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create schedule version: %w", err)
	}

	return nil
}

// GetByID retrieves a schedule version by ID
func (r *ScheduleVersionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ScheduleVersion, error) {
	version := &entity.ScheduleVersion{}

	query := `
		SELECT id, hospital_id, status, effective_start_date, effective_end_date, scrape_batch_id, validation_results,
		       created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		FROM schedule_versions
		WHERE id = $1 AND deleted_at IS NULL
	`

	var validationJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&version.ID,
		&version.HospitalID,
		(*string)(&version.Status),
		&version.EffectiveStartDate,
		&version.EffectiveEndDate,
		&version.ScrapeBatchID,
		&validationJSON,
		&version.CreatedAt,
		&version.CreatedBy,
		&version.UpdatedAt,
		&version.UpdatedBy,
		&version.DeletedAt,
		&version.DeletedBy,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "ScheduleVersion",
			ResourceID:   id.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule version: %w", err)
	}

	// Unmarshal validation results
	if len(validationJSON) > 0 {
		if err := json.Unmarshal(validationJSON, &version.ValidationResults); err != nil {
			return nil, fmt.Errorf("failed to unmarshal validation results: %w", err)
		}
	}

	return version, nil
}

// GetByHospitalAndStatus retrieves schedule versions for a hospital with a specific status
func (r *ScheduleVersionRepository) GetByHospitalAndStatus(ctx context.Context, hospitalID uuid.UUID, status entity.ScheduleVersionStatus) ([]*entity.ScheduleVersion, error) {
	query := `
		SELECT id, hospital_id, status, effective_start_date, effective_end_date, scrape_batch_id, validation_results,
		       created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		FROM schedule_versions
		WHERE hospital_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, hospitalID, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to query schedule versions: %w", err)
	}
	defer rows.Close()

	var versions []*entity.ScheduleVersion
	for rows.Next() {
		version := &entity.ScheduleVersion{}
		var validationJSON []byte

		err := rows.Scan(
			&version.ID,
			&version.HospitalID,
			(*string)(&version.Status),
			&version.EffectiveStartDate,
			&version.EffectiveEndDate,
			&version.ScrapeBatchID,
			&validationJSON,
			&version.CreatedAt,
			&version.CreatedBy,
			&version.UpdatedAt,
			&version.UpdatedBy,
			&version.DeletedAt,
			&version.DeletedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule version: %w", err)
		}

		if len(validationJSON) > 0 {
			if err := json.Unmarshal(validationJSON, &version.ValidationResults); err != nil {
				return nil, fmt.Errorf("failed to unmarshal validation results: %w", err)
			}
		}

		versions = append(versions, version)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule versions: %w", err)
	}

	return versions, nil
}

// GetActiveVersion retrieves the active (PRODUCTION) schedule version for a hospital on a given date
func (r *ScheduleVersionRepository) GetActiveVersion(ctx context.Context, hospitalID uuid.UUID, date entity.Date) (*entity.ScheduleVersion, error) {
	version := &entity.ScheduleVersion{}

	query := `
		SELECT id, hospital_id, status, effective_start_date, effective_end_date, scrape_batch_id, validation_results,
		       created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		FROM schedule_versions
		WHERE hospital_id = $1
		  AND status = 'PRODUCTION'
		  AND effective_start_date <= $2
		  AND effective_end_date >= $2
		  AND deleted_at IS NULL
		LIMIT 1
	`

	var validationJSON []byte

	err := r.db.QueryRowContext(ctx, query, hospitalID, date).Scan(
		&version.ID,
		&version.HospitalID,
		(*string)(&version.Status),
		&version.EffectiveStartDate,
		&version.EffectiveEndDate,
		&version.ScrapeBatchID,
		&validationJSON,
		&version.CreatedAt,
		&version.CreatedBy,
		&version.UpdatedAt,
		&version.UpdatedBy,
		&version.DeletedAt,
		&version.DeletedBy,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "ScheduleVersion",
			ResourceID:   hospitalID.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active schedule version: %w", err)
	}

	if len(validationJSON) > 0 {
		if err := json.Unmarshal(validationJSON, &version.ValidationResults); err != nil {
			return nil, fmt.Errorf("failed to unmarshal validation results: %w", err)
		}
	}

	return version, nil
}

// Update updates a schedule version
func (r *ScheduleVersionRepository) Update(ctx context.Context, version *entity.ScheduleVersion) error {
	validationJSON, err := json.Marshal(version.ValidationResults)
	if err != nil {
		return fmt.Errorf("failed to marshal validation results: %w", err)
	}

	query := `
		UPDATE schedule_versions
		SET hospital_id = $2, status = $3, effective_start_date = $4, effective_end_date = $5,
		    scrape_batch_id = $6, validation_results = $7, updated_at = $8, updated_by = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		version.ID,
		version.HospitalID,
		string(version.Status),
		version.EffectiveStartDate,
		version.EffectiveEndDate,
		version.ScrapeBatchID,
		validationJSON,
		version.UpdatedAt,
		version.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to update schedule version: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &repository.NotFoundError{
			ResourceType: "ScheduleVersion",
			ResourceID:   version.ID.String(),
		}
	}

	return nil
}

// Delete marks a schedule version as deleted
func (r *ScheduleVersionRepository) Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error {
	query := `
		UPDATE schedule_versions
		SET deleted_at = NOW(), deleted_by = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, deleterID)
	if err != nil {
		return fmt.Errorf("failed to delete schedule version: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &repository.NotFoundError{
			ResourceType: "ScheduleVersion",
			ResourceID:   id.String(),
		}
	}

	return nil
}

// ListByHospital retrieves all schedule versions for a hospital
func (r *ScheduleVersionRepository) ListByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.ScheduleVersion, error) {
	query := `
		SELECT id, hospital_id, status, effective_start_date, effective_end_date, scrape_batch_id, validation_results,
		       created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		FROM schedule_versions
		WHERE hospital_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, hospitalID)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedule versions: %w", err)
	}
	defer rows.Close()

	var versions []*entity.ScheduleVersion
	for rows.Next() {
		version := &entity.ScheduleVersion{}
		var validationJSON []byte

		err := rows.Scan(
			&version.ID,
			&version.HospitalID,
			(*string)(&version.Status),
			&version.EffectiveStartDate,
			&version.EffectiveEndDate,
			&version.ScrapeBatchID,
			&validationJSON,
			&version.CreatedAt,
			&version.CreatedBy,
			&version.UpdatedAt,
			&version.UpdatedBy,
			&version.DeletedAt,
			&version.DeletedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule version: %w", err)
		}

		if len(validationJSON) > 0 {
			if err := json.Unmarshal(validationJSON, &version.ValidationResults); err != nil {
				return nil, fmt.Errorf("failed to unmarshal validation results: %w", err)
			}
		}

		versions = append(versions, version)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule versions: %w", err)
	}

	return versions, nil
}

// Count returns the count of active schedule versions
func (r *ScheduleVersionRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	query := `
		SELECT COUNT(*) FROM schedule_versions
		WHERE deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count schedule versions: %w", err)
	}

	return count, nil
}

// Scan implements sql.Scanner for ScheduleVersionStatus
func (s *entity.ScheduleVersionStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	if v, ok := value.(string); ok {
		*s = entity.ScheduleVersionStatus(v)
		return nil
	}

	if v, ok := value.([]byte); ok {
		*s = entity.ScheduleVersionStatus(v)
		return nil
	}

	return fmt.Errorf("cannot scan %T into ScheduleVersionStatus", value)
}

// Value implements driver.Valuer for ScheduleVersionStatus
func (s entity.ScheduleVersionStatus) Value() (driver.Value, error) {
	return string(s), nil
}
