package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"schedcu/v2/internal/entity"
	"schedcu/v2/internal/repository"
)

// CoverageCalculationRepository implements repository.CoverageCalculationRepository for PostgreSQL
type CoverageCalculationRepository struct {
	db *sql.DB
}

// NewCoverageCalculationRepository creates a new CoverageCalculationRepository
func NewCoverageCalculationRepository(db *sql.DB) *CoverageCalculationRepository {
	return &CoverageCalculationRepository{db: db}
}

// Create creates a new coverage calculation
func (r *CoverageCalculationRepository) Create(ctx context.Context, calc *entity.CoverageCalculation) error {
	if calc.ID == uuid.Nil {
		calc.ID = uuid.New()
	}

	coverageByPositionJSON, err := json.Marshal(calc.CoverageByPosition)
	if err != nil {
		return fmt.Errorf("failed to marshal coverage_by_position: %w", err)
	}

	coverageSummaryJSON, err := json.Marshal(calc.CoverageSummary)
	if err != nil {
		return fmt.Errorf("failed to marshal coverage_summary: %w", err)
	}

	query := `
		INSERT INTO coverage_calculations (
			id, schedule_version_id, hospital_id, calculation_date,
			calculation_period_start_date, calculation_period_end_date,
			coverage_by_position, coverage_summary, query_count,
			calculated_at, calculated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		calc.ID,
		calc.ScheduleVersionID,
		calc.HospitalID,
		calc.CalculationDate,
		calc.CalculationPeriodStartDate,
		calc.CalculationPeriodEndDate,
		coverageByPositionJSON,
		coverageSummaryJSON,
		calc.QueryCount,
		calc.CalculatedAt,
		calc.CalculatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create coverage calculation: %w", err)
	}

	return nil
}

// GetByID retrieves a coverage calculation by ID
func (r *CoverageCalculationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CoverageCalculation, error) {
	calc := &entity.CoverageCalculation{
		CoverageByPosition: make(map[string]int),
		CoverageSummary:    make(map[string]interface{}),
	}

	var coverageByPositionJSON, coverageSummaryJSON []byte

	query := `
		SELECT id, schedule_version_id, hospital_id, calculation_date,
		       calculation_period_start_date, calculation_period_end_date,
		       coverage_by_position, coverage_summary, query_count,
		       calculated_at, calculated_by, deleted_at
		FROM coverage_calculations
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&calc.ID,
		&calc.ScheduleVersionID,
		&calc.HospitalID,
		&calc.CalculationDate,
		&calc.CalculationPeriodStartDate,
		&calc.CalculationPeriodEndDate,
		&coverageByPositionJSON,
		&coverageSummaryJSON,
		&calc.QueryCount,
		&calc.CalculatedAt,
		&calc.CalculatedBy,
		&calc.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "CoverageCalculation",
			ResourceID:   id.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coverage calculation: %w", err)
	}

	if err := json.Unmarshal(coverageByPositionJSON, &calc.CoverageByPosition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal coverage_by_position: %w", err)
	}
	if err := json.Unmarshal(coverageSummaryJSON, &calc.CoverageSummary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal coverage_summary: %w", err)
	}

	return calc, nil
}

// GetByScheduleVersion retrieves all coverage calculations for a schedule version
func (r *CoverageCalculationRepository) GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.CoverageCalculation, error) {
	query := `
		SELECT id, schedule_version_id, hospital_id, calculation_date,
		       calculation_period_start_date, calculation_period_end_date,
		       coverage_by_position, coverage_summary, query_count,
		       calculated_at, calculated_by, deleted_at
		FROM coverage_calculations
		WHERE schedule_version_id = $1 AND deleted_at IS NULL
		ORDER BY calculated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, scheduleVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query coverage calculations: %w", err)
	}
	defer rows.Close()

	var calcs []*entity.CoverageCalculation
	for rows.Next() {
		calc := &entity.CoverageCalculation{
			CoverageByPosition: make(map[string]int),
			CoverageSummary:    make(map[string]interface{}),
		}
		var coverageByPositionJSON, coverageSummaryJSON []byte

		err := rows.Scan(
			&calc.ID,
			&calc.ScheduleVersionID,
			&calc.HospitalID,
			&calc.CalculationDate,
			&calc.CalculationPeriodStartDate,
			&calc.CalculationPeriodEndDate,
			&coverageByPositionJSON,
			&coverageSummaryJSON,
			&calc.QueryCount,
			&calc.CalculatedAt,
			&calc.CalculatedBy,
			&calc.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coverage calculation: %w", err)
		}

		if err := json.Unmarshal(coverageByPositionJSON, &calc.CoverageByPosition); err != nil {
			return nil, fmt.Errorf("failed to unmarshal coverage_by_position: %w", err)
		}
		if err := json.Unmarshal(coverageSummaryJSON, &calc.CoverageSummary); err != nil {
			return nil, fmt.Errorf("failed to unmarshal coverage_summary: %w", err)
		}

		calcs = append(calcs, calc)
	}

	return calcs, rows.Err()
}

// GetLatestByScheduleVersion retrieves the most recent coverage calculation for a schedule version
func (r *CoverageCalculationRepository) GetLatestByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) (*entity.CoverageCalculation, error) {
	calc := &entity.CoverageCalculation{
		CoverageByPosition: make(map[string]int),
		CoverageSummary:    make(map[string]interface{}),
	}

	var coverageByPositionJSON, coverageSummaryJSON []byte

	query := `
		SELECT id, schedule_version_id, hospital_id, calculation_date,
		       calculation_period_start_date, calculation_period_end_date,
		       coverage_by_position, coverage_summary, query_count,
		       calculated_at, calculated_by, deleted_at
		FROM coverage_calculations
		WHERE schedule_version_id = $1 AND deleted_at IS NULL
		ORDER BY calculated_at DESC
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, scheduleVersionID).Scan(
		&calc.ID,
		&calc.ScheduleVersionID,
		&calc.HospitalID,
		&calc.CalculationDate,
		&calc.CalculationPeriodStartDate,
		&calc.CalculationPeriodEndDate,
		&coverageByPositionJSON,
		&coverageSummaryJSON,
		&calc.QueryCount,
		&calc.CalculatedAt,
		&calc.CalculatedBy,
		&calc.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "CoverageCalculation",
			ResourceID:   scheduleVersionID.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest coverage calculation: %w", err)
	}

	if err := json.Unmarshal(coverageByPositionJSON, &calc.CoverageByPosition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal coverage_by_position: %w", err)
	}
	if err := json.Unmarshal(coverageSummaryJSON, &calc.CoverageSummary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal coverage_summary: %w", err)
	}

	return calc, nil
}

// GetByHospitalAndDate retrieves coverage calculations for a hospital on a specific date
func (r *CoverageCalculationRepository) GetByHospitalAndDate(ctx context.Context, hospitalID uuid.UUID, date interface{}) ([]*entity.CoverageCalculation, error) {
	query := `
		SELECT id, schedule_version_id, hospital_id, calculation_date,
		       calculation_period_start_date, calculation_period_end_date,
		       coverage_by_position, coverage_summary, query_count,
		       calculated_at, calculated_by, deleted_at
		FROM coverage_calculations
		WHERE hospital_id = $1 AND DATE(calculation_period_start_date) <= $2 AND DATE(calculation_period_end_date) >= $2
		  AND deleted_at IS NULL
		ORDER BY calculated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, hospitalID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to query coverage calculations: %w", err)
	}
	defer rows.Close()

	var calcs []*entity.CoverageCalculation
	for rows.Next() {
		calc := &entity.CoverageCalculation{
			CoverageByPosition: make(map[string]int),
			CoverageSummary:    make(map[string]interface{}),
		}
		var coverageByPositionJSON, coverageSummaryJSON []byte

		err := rows.Scan(
			&calc.ID,
			&calc.ScheduleVersionID,
			&calc.HospitalID,
			&calc.CalculationDate,
			&calc.CalculationPeriodStartDate,
			&calc.CalculationPeriodEndDate,
			&coverageByPositionJSON,
			&coverageSummaryJSON,
			&calc.QueryCount,
			&calc.CalculatedAt,
			&calc.CalculatedBy,
			&calc.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coverage calculation: %w", err)
		}

		if err := json.Unmarshal(coverageByPositionJSON, &calc.CoverageByPosition); err != nil {
			return nil, fmt.Errorf("failed to unmarshal coverage_by_position: %w", err)
		}
		if err := json.Unmarshal(coverageSummaryJSON, &calc.CoverageSummary); err != nil {
			return nil, fmt.Errorf("failed to unmarshal coverage_summary: %w", err)
		}

		calcs = append(calcs, calc)
	}

	return calcs, rows.Err()
}

// Update updates a coverage calculation
func (r *CoverageCalculationRepository) Update(ctx context.Context, calc *entity.CoverageCalculation) error {
	coverageByPositionJSON, err := json.Marshal(calc.CoverageByPosition)
	if err != nil {
		return fmt.Errorf("failed to marshal coverage_by_position: %w", err)
	}

	coverageSummaryJSON, err := json.Marshal(calc.CoverageSummary)
	if err != nil {
		return fmt.Errorf("failed to marshal coverage_summary: %w", err)
	}

	query := `
		UPDATE coverage_calculations
		SET coverage_by_position = $1, coverage_summary = $2, query_count = $3,
		    calculated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		coverageByPositionJSON,
		coverageSummaryJSON,
		calc.QueryCount,
		calc.CalculatedAt,
		calc.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update coverage calculation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "CoverageCalculation",
			ResourceID:   calc.ID.String(),
		}
	}

	return nil
}

// Delete soft-deletes a coverage calculation
func (r *CoverageCalculationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE coverage_calculations
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete coverage calculation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "CoverageCalculation",
			ResourceID:   id.String(),
		}
	}

	return nil
}

// Count returns the total count of non-deleted coverage calculations
func (r *CoverageCalculationRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM coverage_calculations WHERE deleted_at IS NULL`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count coverage calculations: %w", err)
	}

	return count, nil
}
