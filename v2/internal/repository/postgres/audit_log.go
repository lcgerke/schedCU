package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
)

// AuditLogRepository implements repository.AuditLogRepository for PostgreSQL
type AuditLogRepository struct {
	db *sql.DB
}

// NewAuditLogRepository creates a new AuditLogRepository
func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log (immutable)
func (r *AuditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	query := `
		INSERT INTO audit_logs (
			id, user_id, action, resource, old_values, new_values, timestamp, ip_address
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.Action,
		log.Resource,
		log.OldValues,
		log.NewValues,
		log.Timestamp,
		log.IPAddress,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByID retrieves an audit log by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
	log := &entity.AuditLog{}

	query := `
		SELECT id, user_id, action, resource, old_values, new_values, timestamp, ip_address
		FROM audit_logs
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.UserID,
		&log.Action,
		&log.Resource,
		&log.OldValues,
		&log.NewValues,
		&log.Timestamp,
		&log.IPAddress,
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

	return log, nil
}

// GetByUser retrieves audit logs for a specific user
func (r *AuditLogRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource, old_values, new_values, timestamp, ip_address
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.Resource,
			&log.OldValues,
			&log.NewValues,
			&log.Timestamp,
			&log.IPAddress,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetByAction retrieves audit logs for a specific action
func (r *AuditLogRepository) GetByAction(ctx context.Context, action string) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource, old_values, new_values, timestamp, ip_address
		FROM audit_logs
		WHERE action = $1
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, action)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by action: %w", err)
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.Resource,
			&log.OldValues,
			&log.NewValues,
			&log.Timestamp,
			&log.IPAddress,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// ListRecent retrieves the most recent audit logs
func (r *AuditLogRepository) ListRecent(ctx context.Context, limit int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource, old_values, new_values, timestamp, ip_address
		FROM audit_logs
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.Resource,
			&log.OldValues,
			&log.NewValues,
			&log.Timestamp,
			&log.IPAddress,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// Count returns the total number of audit logs
func (r *AuditLogRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM audit_logs`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}
	return count, nil
}
