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

// JobQueueRepository implements repository.JobQueueRepository for PostgreSQL
type JobQueueRepository struct {
	db *sql.DB
}

// NewJobQueueRepository creates a new JobQueueRepository
func NewJobQueueRepository(db *sql.DB) *JobQueueRepository {
	return &JobQueueRepository{db: db}
}

// Create creates a new job in the queue
func (r *JobQueueRepository) Create(ctx context.Context, job *entity.JobQueue) error {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}

	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resultJSON, err := json.Marshal(job.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	query := `
		INSERT INTO job_queue (
			id, job_type, status, payload, result,
			retry_count, max_retries, error_message,
			created_at, started_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		job.ID,
		job.JobType,
		string(job.Status),
		payloadJSON,
		resultJSON,
		job.RetryCount,
		job.MaxRetries,
		job.ErrorMessage,
		job.CreatedAt,
		job.StartedAt,
		job.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// GetByID retrieves a job by ID
func (r *JobQueueRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.JobQueue, error) {
	job := &entity.JobQueue{
		Payload: make(map[string]interface{}),
		Result:  make(map[string]interface{}),
	}

	var payloadJSON, resultJSON []byte

	query := `
		SELECT id, job_type, status, payload, result,
		       retry_count, max_retries, error_message,
		       created_at, started_at, completed_at
		FROM job_queue
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.JobType,
		(*string)(&job.Status),
		&payloadJSON,
		&resultJSON,
		&job.RetryCount,
		&job.MaxRetries,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "JobQueue",
			ResourceID:   id.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	if len(payloadJSON) > 0 {
		if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	if len(resultJSON) > 0 {
		if err := json.Unmarshal(resultJSON, &job.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return job, nil
}

// GetByStatus retrieves jobs with a specific status
func (r *JobQueueRepository) GetByStatus(ctx context.Context, status entity.JobQueueStatus) ([]*entity.JobQueue, error) {
	query := `
		SELECT id, job_type, status, payload, result,
		       retry_count, max_retries, error_message,
		       created_at, started_at, completed_at
		FROM job_queue
		WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs by status: %w", err)
	}
	defer rows.Close()

	var jobs []*entity.JobQueue
	for rows.Next() {
		job := &entity.JobQueue{
			Payload: make(map[string]interface{}),
			Result:  make(map[string]interface{}),
		}
		var payloadJSON, resultJSON []byte

		err := rows.Scan(
			&job.ID,
			&job.JobType,
			(*string)(&job.Status),
			&payloadJSON,
			&resultJSON,
			&job.RetryCount,
			&job.MaxRetries,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if len(payloadJSON) > 0 {
			if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
				return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
			}
		}

		if len(resultJSON) > 0 {
			if err := json.Unmarshal(resultJSON, &job.Result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// GetByType retrieves jobs of a specific type
func (r *JobQueueRepository) GetByType(ctx context.Context, jobType string) ([]*entity.JobQueue, error) {
	query := `
		SELECT id, job_type, status, payload, result,
		       retry_count, max_retries, error_message,
		       created_at, started_at, completed_at
		FROM job_queue
		WHERE job_type = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, jobType)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs by type: %w", err)
	}
	defer rows.Close()

	var jobs []*entity.JobQueue
	for rows.Next() {
		job := &entity.JobQueue{
			Payload: make(map[string]interface{}),
			Result:  make(map[string]interface{}),
		}
		var payloadJSON, resultJSON []byte

		err := rows.Scan(
			&job.ID,
			&job.JobType,
			(*string)(&job.Status),
			&payloadJSON,
			&resultJSON,
			&job.RetryCount,
			&job.MaxRetries,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if len(payloadJSON) > 0 {
			if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
				return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
			}
		}

		if len(resultJSON) > 0 {
			if err := json.Unmarshal(resultJSON, &job.Result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// Update updates a job
func (r *JobQueueRepository) Update(ctx context.Context, job *entity.JobQueue) error {
	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resultJSON, err := json.Marshal(job.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	query := `
		UPDATE job_queue
		SET status = $1, payload = $2, result = $3,
		    retry_count = $4, max_retries = $5, error_message = $6,
		    started_at = $7, completed_at = $8
		WHERE id = $9
	`

	_, err = r.db.ExecContext(ctx, query,
		string(job.Status),
		payloadJSON,
		resultJSON,
		job.RetryCount,
		job.MaxRetries,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
		job.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

// Count returns the total number of jobs
func (r *JobQueueRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM job_queue`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count jobs: %w", err)
	}
	return count, nil
}
