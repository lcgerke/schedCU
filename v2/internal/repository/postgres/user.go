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

// UserRepository implements repository.UserRepository for PostgreSQL
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	query := `
		INSERT INTO users (
			id, email, password_hash, hospital_id, role, full_name,
			is_active, last_login, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.HospitalID,
		string(user.Role),
		user.FullName,
		user.IsActive,
		user.LastLogin,
		user.CreatedAt,
		user.CreatedBy,
		user.UpdatedAt,
		user.UpdatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user := &entity.User{}

	query := `
		SELECT id, email, password_hash, hospital_id, role, full_name,
		       is_active, last_login, created_at, created_by, updated_at, updated_by, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.HospitalID,
		(*string)(&user.Role),
		&user.FullName,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.CreatedBy,
		&user.UpdatedAt,
		&user.UpdatedBy,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "User",
			ResourceID:   id.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}

	query := `
		SELECT id, email, password_hash, hospital_id, role, full_name,
		       is_active, last_login, created_at, created_by, updated_at, updated_by, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.HospitalID,
		(*string)(&user.Role),
		&user.FullName,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.CreatedBy,
		&user.UpdatedAt,
		&user.UpdatedBy,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "User",
			ResourceID:   email,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByHospital retrieves all users for a hospital
func (r *UserRepository) GetByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.User, error) {
	query := `
		SELECT id, email, password_hash, hospital_id, role, full_name,
		       is_active, last_login, created_at, created_by, updated_at, updated_by, deleted_at
		FROM users
		WHERE hospital_id = $1 AND deleted_at IS NULL
		ORDER BY email ASC
	`

	rows, err := r.db.QueryContext(ctx, query, hospitalID)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.HospitalID,
			(*string)(&user.Role),
			&user.FullName,
			&user.IsActive,
			&user.LastLogin,
			&user.CreatedAt,
			&user.CreatedBy,
			&user.UpdatedAt,
			&user.UpdatedBy,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetByRole retrieves all users with a specific role
func (r *UserRepository) GetByRole(ctx context.Context, role entity.UserRole) ([]*entity.User, error) {
	query := `
		SELECT id, email, password_hash, hospital_id, role, full_name,
		       is_active, last_login, created_at, created_by, updated_at, updated_by, deleted_at
		FROM users
		WHERE role = $1 AND deleted_at IS NULL
		ORDER BY email ASC
	`

	rows, err := r.db.QueryContext(ctx, query, string(role))
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.HospitalID,
			(*string)(&user.Role),
			&user.FullName,
			&user.IsActive,
			&user.LastLogin,
			&user.CreatedAt,
			&user.CreatedBy,
			&user.UpdatedAt,
			&user.UpdatedBy,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, hospital_id = $3, role = $4,
		    full_name = $5, is_active = $6, last_login = $7,
		    updated_at = $8, updated_by = $9
		WHERE id = $10 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.HospitalID,
		string(user.Role),
		user.FullName,
		user.IsActive,
		user.LastLogin,
		user.UpdatedAt,
		user.UpdatedBy,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "User",
			ResourceID:   user.ID.String(),
		}
	}

	return nil
}

// Delete soft-deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = NOW(), deleted_by = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, deleterID, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return &repository.NotFoundError{
			ResourceType: "User",
			ResourceID:   id.String(),
		}
	}

	return nil
}

// Count returns the total count of non-deleted users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}
