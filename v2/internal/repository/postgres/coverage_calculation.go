package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
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
		return fmt.Errorf("failed to marshal coverage by position: %w", err)
	}

	coverageSummaryJSON, err := json.Marshal(calc.CoverageSummary)
	if err != nil {
		return fmt.Errorf("failed to marshal coverage summary: %w", err)
	}

	validationJSON, err := json.Marshal(calc.ValidationErrors)
	if err != nil {
		return fmt.Errorf("failed to marshal validation errors: %w", err)
	}

	query := `
		INSERT INTO coverage_calculations (
			id, schedule_version_id, hospital_id, calculation_date,
			calculation_period_start_date, calculation_period_end_date,
			coverage_by_position, coverage_summary, validation_errors,
			query_count, calculated_at, calculated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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
		validationJSON,
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

	var coverageByPositionJSON, coverageSummaryJSON, validationJSON []byte

	query := `
		SELECT id, schedule_version_id, hospital_id, calculation_date,
		       calculation_period_start_date, calculation_period_end_date,
		       coverage_by_position, coverage_summary, validation_errors,
		       query_count, calculated_at, calculated_by
		FROM coverage_calculations
		WHERE id = $1
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
		&validationJSON,
		&calc.QueryCount,
		&calc.CalculatedAt,
		&calc.CalculatedBy,
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
		return nil, fmt.Errorf("failed to unmarshal coverage by position: %w", err)
	}

	if err := json.Unmarshal(coverageSummaryJSON, &calc.CoverageSummary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal coverage summary: %w", err)
	}

	if len(validationJSON) > 0 {
		if err := json.Unmarshal(validationJSON, &calc.ValidationErrors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal validation errors: %w", err)
		}
	}

	return calc, nil
}

// GetByScheduleVersion retrieves all coverage calculations for a schedule version
func (r *CoverageCalculationRepository) GetByScheduleVersion(ctx context.Context, versionID uuid.UUID) ([]*entity.CoverageCalculation, error) {
	query := `
		SELECT id, schedule_version_id, hospital_id, calculation_date,
		       calculation_period_start_date, calculation_period_end_date,
		       coverage_by_position, coverage_summary, validation_errors,
		       query_count, calculated_at, calculated_by
		FROM coverage_calculations
		WHERE schedule_version_id = $1
		ORDER BY calculated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, versionID)
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
		var coverageByPositionJSON, coverageSummaryJSON, validationJSON []byte

		err := rows.Scan(
			&calc.ID,
			&calc.ScheduleVersionID,
			&calc.HospitalID,
			&calc.CalculationDate,
			&calc.CalculationPeriodStartDate,
			&calc.CalculationPeriodEndDate,
			&coverageByPositionJSON,
			&coverageSummaryJSON,
			&validationJSON,
			&calc.QueryCount,
			&calc.CalculatedAt,
			&calc.CalculatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coverage calculation: %w", err)
		}

		if err := json.Unmarshal(coverageByPositionJSON, &calc.CoverageByPosition); err != nil {
			return nil, fmt.Errorf("failed to unmarshal coverage by position: %w", err)
		}

		if err := json.Unmarshal(coverageSummaryJSON, &calc.CoverageSummary); err != nil {
			return nil, fmt.Errorf("failed to unmarshal coverage summary: %w", err)
		}

		if len(validationJSON) > 0 {
			if err := json.Unmarshal(validationJSON, &calc.ValidationErrors); err != nil {
				return nil, fmt.Errorf("failed to unmarshal validation errors: %w", err)
			}
		}

		calcs = append(calcs, calc)
	}

	return calcs, rows.Err()
}

// GetLatestByScheduleVersion retrieves the most recent coverage calculation for a schedule version
func (r *CoverageCalculationRepository) GetLatestByScheduleVersion(ctx context.Context, versionID uuid.UUID) (*entity.CoverageCalculation, error) {
	calc := &entity.CoverageCalculation{
		CoverageByPosition: make(map[string]int),
		CoverageSummary:    make(map[string]interface{}),
	}

	var coverageByPositionJSON, coverageSummaryJSON, validationJSON []byte

	query := `
		SELECT id, schedule_version_id, hospital_id, calculation_date,
		       calculation_period_start_date, calculation_period_end_date,
		       coverage_by_position, coverage_summary, validation_errors,
		       query_count, calculated_at, calculated_by
		FROM coverage_calculations
		WHERE schedule_version_id = $1
		ORDER BY calculated_at DESC
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, versionID).Scan(
		&calc.ID,
		&calc.ScheduleVersionID,
		&calc.HospitalID,
		&calc.CalculationDate,
		&calc.CalculationPeriodStartDate,
		&calc.CalculationPeriodEndDate,
		&coverageByPositionJSON,
		&coverageSummaryJSON,
		&validationJSON,
		&calc.QueryCount,
		&calc.CalculatedAt,
		&calc.CalculatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "CoverageCalculation",
			ResourceID:   versionID.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest coverage calculation: %w", err)
	}

	if err := json.Unmarshal(coverageByPositionJSON, &calc.CoverageByPosition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal coverage by position: %w", err)
	}

	if err := json.Unmarshal(coverageSummaryJSON, &calc.CoverageSummary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal coverage summary: %w", err)
	}

	if len(validationJSON) > 0 {
		if err := json.Unmarshal(validationJSON, &calc.ValidationErrors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal validation errors: %w", err)
		}
	}

	return calc, nil
}

// Count returns the total number of coverage calculations
func (r *CoverageCalculationRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM coverage_calculations`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count coverage calculations: %w", err)
	}
	return count, nil
}
