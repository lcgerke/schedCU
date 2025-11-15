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

	query := `
		INSERT INTO job_queue (
			id, job_type, status, scheduled_for, payload,
			attempts, max_attempts, error_message,
			created_at, started_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		job.ID,
		job.JobType,
		string(job.Status),
		job.ScheduledFor,
		payloadJSON,
		job.Attempts,
		job.MaxAttempts,
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
	}

	var payloadJSON []byte

	query := `
		SELECT id, job_type, status, scheduled_for, payload,
		       attempts, max_attempts, error_message,
		       created_at, started_at, completed_at
		FROM job_queue
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.JobType,
		(*string)(&job.Status),
		&job.ScheduledFor,
		&payloadJSON,
		&job.Attempts,
		&job.MaxAttempts,
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

	if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return job, nil
}

// GetByStatus retrieves all jobs with a specific status
func (r *JobQueueRepository) GetByStatus(ctx context.Context, status entity.JobQueueStatus) ([]*entity.JobQueue, error) {
	query := `
		SELECT id, job_type, status, scheduled_for, payload,
		       attempts, max_attempts, error_message,
		       created_at, started_at, completed_at
		FROM job_queue
		WHERE status = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*entity.JobQueue
	for rows.Next() {
		job := &entity.JobQueue{
			Payload: make(map[string]interface{}),
		}
		var payloadJSON []byte

		err := rows.Scan(
			&job.ID,
			&job.JobType,
			(*string)(&job.Status),
			&job.ScheduledFor,
			&payloadJSON,
			&job.Attempts,
			&job.MaxAttempts,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// GetByType retrieves all jobs of a specific type
func (r *JobQueueRepository) GetByType(ctx context.Context, jobType string) ([]*entity.JobQueue, error) {
	query := `
		SELECT id, job_type, status, scheduled_for, payload,
		       attempts, max_attempts, error_message,
		       created_at, started_at, completed_at
		FROM job_queue
		WHERE job_type = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, jobType)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*entity.JobQueue
	for rows.Next() {
		job := &entity.JobQueue{
			Payload: make(map[string]interface{}),
		}
		var payloadJSON []byte

		err := rows.Scan(
			&job.ID,
			&job.JobType,
			(*string)(&job.Status),
			&job.ScheduledFor,
			&payloadJSON,
			&job.Attempts,
			&job.MaxAttempts,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// GetPending retrieves all pending jobs
func (r *JobQueueRepository) GetPending(ctx context.Context) ([]*entity.JobQueue, error) {
	query := `
		SELECT id, job_type, status, scheduled_for, payload,
		       attempts, max_attempts, error_message,
		       created_at, started_at, completed_at
		FROM job_queue
		WHERE status = $1 AND scheduled_for <= NOW()
		ORDER BY scheduled_for ASC
	`

	rows, err := r.db.QueryContext(ctx, query, string(entity.JobQueueStatusPending))
	if err != nil {
		return nil, fmt.Errorf("failed to query pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*entity.JobQueue
	for rows.Next() {
		job := &entity.JobQueue{
			Payload: make(map[string]interface{}),
		}
		var payloadJSON []byte

		err := rows.Scan(
			&job.ID,
			&job.JobType,
			(*string)(&job.Status),
			&job.ScheduledFor,
			&payloadJSON,
			&job.Attempts,
			&job.MaxAttempts,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if err := json.Unmarshal(payloadJSON, &job.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
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

	query := `
		UPDATE job_queue
		SET status = $1, attempts = $2, error_message = $3,
		    started_at = $4, completed_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		string(job.Status),
		job.Attempts,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
		job.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "JobQueue",
			ResourceID:   job.ID.String(),
		}
	}

	return nil
}

// Delete deletes a job from the queue
func (r *JobQueueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM job_queue WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "JobQueue",
			ResourceID:   id.String(),
		}
	}

	return nil
}

// Count returns the total count of jobs in the queue
func (r *JobQueueRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM job_queue`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	return count, nil
}

// CleanupOldJobs removes old completed jobs (retention policy)
func (r *JobQueueRepository) CleanupOldJobs(ctx context.Context, daysOld int) (int64, error) {
	query := `
		DELETE FROM job_queue
		WHERE status = $1 AND completed_at < NOW() - INTERVAL '1 day' * $2
	`

	result, err := r.db.ExecContext(ctx, query, string(entity.JobQueueStatusCompleted), daysOld)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old jobs: %w", err)
	}

	return result.RowsAffected()
}
