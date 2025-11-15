package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
)

// ScrapeBatchRepository implements repository.ScrapeBatchRepository for PostgreSQL
type ScrapeBatchRepository struct {
	db *sql.DB
}

// NewScrapeBatchRepository creates a new ScrapeBatchRepository
func NewScrapeBatchRepository(db *sql.DB) *ScrapeBatchRepository {
	return &ScrapeBatchRepository{db: db}
}

// Create creates a new scrape batch
func (r *ScrapeBatchRepository) Create(ctx context.Context, batch *entity.ScrapeBatch) error {
	if batch.ID == uuid.Nil {
		batch.ID = uuid.New()
	}

	query := `
		INSERT INTO scrape_batches (
			id, hospital_id, state, window_start_date, window_end_date,
			scraped_at, row_count, error_message, created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		batch.ID,
		batch.HospitalID,
		string(batch.State),
		batch.WindowStartDate,
		batch.WindowEndDate,
		batch.ScrapedAt,
		batch.RowCount,
		batch.ErrorMessage,
		batch.CreatedAt,
		batch.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create scrape batch: %w", err)
	}

	return nil
}

// GetByID retrieves a scrape batch by ID
func (r *ScrapeBatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ScrapeBatch, error) {
	batch := &entity.ScrapeBatch{}

	query := `
		SELECT id, hospital_id, state, window_start_date, window_end_date,
		       scraped_at, row_count, error_message, created_at, created_by, deleted_at
		FROM scrape_batches
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&batch.ID,
		&batch.HospitalID,
		(*string)(&batch.State),
		&batch.WindowStartDate,
		&batch.WindowEndDate,
		&batch.ScrapedAt,
		&batch.RowCount,
		&batch.ErrorMessage,
		&batch.CreatedAt,
		&batch.CreatedBy,
		&batch.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "ScrapeBatch",
			ResourceID:   id.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scrape batch: %w", err)
	}

	return batch, nil
}

// GetByHospital retrieves all scrape batches for a hospital
func (r *ScrapeBatchRepository) GetByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.ScrapeBatch, error) {
	query := `
		SELECT id, hospital_id, state, window_start_date, window_end_date,
		       scraped_at, row_count, error_message, created_at, created_by, deleted_at
		FROM scrape_batches
		WHERE hospital_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, hospitalID)
	if err != nil {
		return nil, fmt.Errorf("failed to query scrape batches: %w", err)
	}
	defer rows.Close()

	var batches []*entity.ScrapeBatch
	for rows.Next() {
		batch := &entity.ScrapeBatch{}
		err := rows.Scan(
			&batch.ID,
			&batch.HospitalID,
			(*string)(&batch.State),
			&batch.WindowStartDate,
			&batch.WindowEndDate,
			&batch.ScrapedAt,
			&batch.RowCount,
			&batch.ErrorMessage,
			&batch.CreatedAt,
			&batch.CreatedBy,
			&batch.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scrape batch: %w", err)
		}
		batches = append(batches, batch)
	}

	return batches, rows.Err()
}

// GetByStatus retrieves all scrape batches with a specific status
func (r *ScrapeBatchRepository) GetByStatus(ctx context.Context, status entity.BatchState) ([]*entity.ScrapeBatch, error) {
	query := `
		SELECT id, hospital_id, state, window_start_date, window_end_date,
		       scraped_at, row_count, error_message, created_at, created_by, deleted_at
		FROM scrape_batches
		WHERE state = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to query scrape batches by status: %w", err)
	}
	defer rows.Close()

	var batches []*entity.ScrapeBatch
	for rows.Next() {
		batch := &entity.ScrapeBatch{}
		err := rows.Scan(
			&batch.ID,
			&batch.HospitalID,
			(*string)(&batch.State),
			&batch.WindowStartDate,
			&batch.WindowEndDate,
			&batch.ScrapedAt,
			&batch.RowCount,
			&batch.ErrorMessage,
			&batch.CreatedAt,
			&batch.CreatedBy,
			&batch.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scrape batch: %w", err)
		}
		batches = append(batches, batch)
	}

	return batches, rows.Err()
}

// Update updates a scrape batch
func (r *ScrapeBatchRepository) Update(ctx context.Context, batch *entity.ScrapeBatch) error {
	query := `
		UPDATE scrape_batches
		SET state = $1, row_count = $2, error_message = $3, scraped_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		string(batch.State),
		batch.RowCount,
		batch.ErrorMessage,
		batch.ScrapedAt,
		batch.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update scrape batch: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "ScrapeBatch",
			ResourceID:   batch.ID.String(),
		}
	}

	return nil
}

// Delete soft-deletes a scrape batch
func (r *ScrapeBatchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE scrape_batches
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete scrape batch: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "ScrapeBatch",
			ResourceID:   id.String(),
		}
	}

	return nil
}

// Count returns the total count of non-deleted scrape batches
func (r *ScrapeBatchRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM scrape_batches WHERE deleted_at IS NULL`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count scrape batches: %w", err)
	}

	return count, nil
}
