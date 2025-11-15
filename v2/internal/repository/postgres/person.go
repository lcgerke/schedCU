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

// PersonRepository implements repository.PersonRepository for PostgreSQL
type PersonRepository struct {
	db *sql.DB
}

// NewPersonRepository creates a new PersonRepository
func NewPersonRepository(db *sql.DB) *PersonRepository {
	return &PersonRepository{db: db}
}

// Create creates a new person in the database
func (r *PersonRepository) Create(ctx context.Context, person *entity.Person) error {
	if person.ID == uuid.Nil {
		person.ID = uuid.New()
	}

	query := `
		INSERT INTO persons (id, email, name, specialty, active, aliases, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		person.ID,
		person.Email,
		person.Name,
		string(person.Specialty),
		person.Active,
		pq.Array(person.Aliases),
		person.CreatedAt,
		person.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create person: %w", err)
	}

	return nil
}

// GetByID retrieves a person by ID
func (r *PersonRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Person, error) {
	person := &entity.Person{}

	query := `
		SELECT id, email, name, specialty, active, aliases, created_at, updated_at, deleted_at
		FROM persons
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&person.ID,
		&person.Email,
		&person.Name,
		(*string)(&person.Specialty),
		&person.Active,
		pq.Array(&person.Aliases),
		&person.CreatedAt,
		&person.UpdatedAt,
		&person.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "Person",
			ResourceID:   id.String(),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get person: %w", err)
	}

	return person, nil
}

// GetByEmail retrieves a person by email
func (r *PersonRepository) GetByEmail(ctx context.Context, email string) (*entity.Person, error) {
	person := &entity.Person{}

	query := `
		SELECT id, email, name, specialty, active, aliases, created_at, updated_at, deleted_at
		FROM persons
		WHERE email = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&person.ID,
		&person.Email,
		&person.Name,
		(*string)(&person.Specialty),
		&person.Active,
		pq.Array(&person.Aliases),
		&person.CreatedAt,
		&person.UpdatedAt,
		&person.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &repository.NotFoundError{
			ResourceType: "Person",
			ResourceID:   email,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get person by email: %w", err)
	}

	return person, nil
}

// GetByHospital retrieves all persons associated with a hospital
func (r *PersonRepository) GetByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.Person, error) {
	// Note: This needs a person_hospitals bridge table or hospital_id on persons table
	// For now, returning empty slice - implement based on actual schema
	return []*entity.Person{}, nil
}

// Update updates a person's record
func (r *PersonRepository) Update(ctx context.Context, person *entity.Person) error {
	query := `
		UPDATE persons
		SET email = $2, name = $3, specialty = $4, active = $5, aliases = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		person.ID,
		person.Email,
		person.Name,
		string(person.Specialty),
		person.Active,
		pq.Array(person.Aliases),
		person.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update person: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &repository.NotFoundError{
			ResourceType: "Person",
			ResourceID:   person.ID.String(),
		}
	}

	return nil
}

// Delete marks a person as deleted
func (r *PersonRepository) Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error {
	query := `
		UPDATE persons
		SET deleted_at = NOW(), deleted_by = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, deleterID)
	if err != nil {
		return fmt.Errorf("failed to delete person: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &repository.NotFoundError{
			ResourceType: "Person",
			ResourceID:   id.String(),
		}
	}

	return nil
}

// Count returns the count of active persons
func (r *PersonRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	query := `
		SELECT COUNT(*) FROM persons
		WHERE deleted_at IS NULL AND active = true
	`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count persons: %w", err)
	}

	return count, nil
}
