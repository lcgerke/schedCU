package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
)

// Database provides access to all repositories
type Database interface {
	// Transaction management
	BeginTx(ctx context.Context) (Transaction, error)

	// Repository accessors
	HospitalRepository() HospitalRepository
	PersonRepository() PersonRepository
	ScheduleVersionRepository() ScheduleVersionRepository
	ShiftInstanceRepository() ShiftInstanceRepository
	AssignmentRepository() AssignmentRepository
	ScrapeBatchRepository() ScrapeBatchRepository
	CoverageCalculationRepository() CoverageCalculationRepository
	AuditLogRepository() AuditLogRepository
	UserRepository() UserRepository
	JobQueueRepository() JobQueueRepository

	// Connection management
	Close() error
	Health(ctx context.Context) error
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error

	HospitalRepository() HospitalRepository
	PersonRepository() PersonRepository
	ScheduleVersionRepository() ScheduleVersionRepository
	ShiftInstanceRepository() ShiftInstanceRepository
	AssignmentRepository() AssignmentRepository
	ScrapeBatchRepository() ScrapeBatchRepository
	CoverageCalculationRepository() CoverageCalculationRepository
	AuditLogRepository() AuditLogRepository
	UserRepository() UserRepository
	JobQueueRepository() JobQueueRepository
}

// HospitalRepository defines data access operations for hospitals
type HospitalRepository interface {
	Create(ctx context.Context, hospital *entity.Hospital) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Hospital, error)
	GetAll(ctx context.Context) ([]*entity.Hospital, error)
	Update(ctx context.Context, hospital *entity.Hospital) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// PersonRepository defines data access operations for persons (staff members)
type PersonRepository interface {
	Create(ctx context.Context, person *entity.Person) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Person, error)
	GetByEmail(ctx context.Context, email string) (*entity.Person, error)
	GetByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.Person, error)
	Update(ctx context.Context, person *entity.Person) error
	Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// ScheduleVersionRepository defines data access operations for schedule versions
type ScheduleVersionRepository interface {
	Create(ctx context.Context, version *entity.ScheduleVersion) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ScheduleVersion, error)
	GetByHospitalAndStatus(ctx context.Context, hospitalID uuid.UUID, status entity.VersionStatus) ([]*entity.ScheduleVersion, error)
	GetActiveVersion(ctx context.Context, hospitalID uuid.UUID, date time.Time) (*entity.ScheduleVersion, error)
	Update(ctx context.Context, version *entity.ScheduleVersion) error
	Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error
	ListByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.ScheduleVersion, error)
	Count(ctx context.Context) (int64, error)
}

// ShiftInstanceRepository defines data access operations for shift instances
type ShiftInstanceRepository interface {
	Create(ctx context.Context, shift *entity.ShiftInstance) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ShiftInstance, error)
	GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.ShiftInstance, error)
	GetByDateRange(ctx context.Context, scheduleVersionID uuid.UUID, startDate, endDate time.Time) ([]*entity.ShiftInstance, error)
	Update(ctx context.Context, shift *entity.ShiftInstance) error
	Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) (int64, error)
}

// AssignmentRepository defines data access operations for shift assignments
type AssignmentRepository interface {
	Create(ctx context.Context, assignment *entity.Assignment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Assignment, error)
	GetByShiftInstance(ctx context.Context, shiftInstanceID uuid.UUID) ([]*entity.Assignment, error)
	GetByPerson(ctx context.Context, personID uuid.UUID) ([]*entity.Assignment, error)
	GetByPersonAndDateRange(ctx context.Context, personID uuid.UUID, startDate, endDate time.Time) ([]*entity.Assignment, error)
	GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.Assignment, error)
	Update(ctx context.Context, assignment *entity.Assignment) error
	Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error
	Count(ctx context.Context) (int64, error)

	// Batch operations (no N+1 queries)
	GetAllByShiftIDs(ctx context.Context, shiftInstanceIDs []uuid.UUID) ([]*entity.Assignment, error)
}

// ScrapeBatchRepository defines data access operations for scrape batches
type ScrapeBatchRepository interface {
	Create(ctx context.Context, batch *entity.ScrapeBatch) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ScrapeBatch, error)
	GetByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.ScrapeBatch, error)
	GetByStatus(ctx context.Context, status entity.BatchState) ([]*entity.ScrapeBatch, error)
	Update(ctx context.Context, batch *entity.ScrapeBatch) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// CoverageCalculationRepository defines data access operations for coverage calculations
type CoverageCalculationRepository interface {
	Create(ctx context.Context, calculation *entity.CoverageCalculation) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CoverageCalculation, error)
	GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.CoverageCalculation, error)
	GetLatestByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) (*entity.CoverageCalculation, error)
	GetByHospitalAndDate(ctx context.Context, hospitalID uuid.UUID, date time.Time) ([]*entity.CoverageCalculation, error)
	Update(ctx context.Context, calculation *entity.CoverageCalculation) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// AuditLogRepository defines data access operations for audit logs
type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error)
	GetByUser(ctx context.Context, userID uuid.UUID) ([]*entity.AuditLog, error)
	GetByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]*entity.AuditLog, error)
	GetByAction(ctx context.Context, action string) ([]*entity.AuditLog, error)
	ListRecent(ctx context.Context, limit int) ([]*entity.AuditLog, error)
	Count(ctx context.Context) (int64, error)
}

// UserRepository defines data access operations for users
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.User, error)
	GetByRole(ctx context.Context, role entity.UserRole) ([]*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// JobQueueRepository defines data access operations for job queue
type JobQueueRepository interface {
	Create(ctx context.Context, job *entity.JobQueue) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.JobQueue, error)
	GetByStatus(ctx context.Context, status entity.JobQueueStatus) ([]*entity.JobQueue, error)
	GetByType(ctx context.Context, jobType string) ([]*entity.JobQueue, error)
	GetPending(ctx context.Context) ([]*entity.JobQueue, error)
	Update(ctx context.Context, job *entity.JobQueue) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CleanupOldJobs(ctx context.Context, daysOld int) (int64, error)
}

// NotFoundError represents a record not found error
type NotFoundError struct {
	ResourceType string
	ResourceID   string
}

// Error implements the error interface for NotFoundError
func (e *NotFoundError) Error() string {
	return "not found: " + e.ResourceType + " " + e.ResourceID
}

// IsNotFound checks if an error is a NotFoundError
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
	Field   string
}

// Error implements the error interface for ValidationError
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}
