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

// AuditLogRepository implements repository.AuditLogRepository for PostgreSQL
type AuditLogRepository struct {
	db *sql.DB
}

// NewAuditLogRepository creates a new AuditLogRepository
func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log
func (r *AuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	query := `
		INSERT INTO audit_logs (
			id, user_id, action, resource_type, resource_id,
			details, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		detailsJSON,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByID retrieves an audit log by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
	log := &entity.AuditLog{
		Details: make(map[string]interface{}),
	}

	var detailsJSON []byte

	query := `
		SELECT id, user_id, action, resource_type, resource_id, details, created_at
		FROM audit_logs
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.UserID,
		&log.Action,
		&log.ResourceType,
		&log.ResourceID,
		&detailsJSON,
		&log.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "AuditLog",
			ResourceID:   id.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
		return nil, fmt.Errorf("failed to unmarshal details: %w", err)
	}

	return log, nil
}

// GetByUser retrieves audit logs for a specific user
func (r *AuditLogRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, details, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{
			Details: make(map[string]interface{}),
		}
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&detailsJSON,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetByResource retrieves audit logs for a specific resource
func (r *AuditLogRepository) GetByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, details, created_at
		FROM audit_logs
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{
			Details: make(map[string]interface{}),
		}
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&detailsJSON,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetByAction retrieves audit logs for a specific action
func (r *AuditLogRepository) GetByAction(ctx context.Context, action string) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, details, created_at
		FROM audit_logs
		WHERE action = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, action)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{
			Details: make(map[string]interface{}),
		}
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&detailsJSON,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// ListRecent retrieves the most recent audit logs
func (r *AuditLogRepository) ListRecent(ctx context.Context, limit int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, details, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{
			Details: make(map[string]interface{}),
		}
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&detailsJSON,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// Count returns the total count of audit logs
func (r *AuditLogRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM audit_logs`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}
