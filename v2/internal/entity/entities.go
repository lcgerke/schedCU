package entity

import (
	"time"

	"github.com/google/uuid"
)

// Type aliases for domain IDs and temporal types
type (
	HospitalID         = uuid.UUID
	PersonID           = uuid.UUID
	ScheduleVersionID  = uuid.UUID
	ShiftInstanceID    = uuid.UUID
	AssignmentID       = uuid.UUID
	ScrapeBatchID      = uuid.UUID
	CoverageID         = uuid.UUID
	AuditLogID         = uuid.UUID
	UserID             = uuid.UUID
	JobQueueID         = uuid.UUID
	Date               = time.Time
	Time               = time.Time
)

// Helper functions for creating instances
func Now() time.Time {
	return time.Now().UTC()
}

func NowPtr() *time.Time {
	now := time.Now().UTC()
	return &now
}

// Hospital represents a hospital facility
type Hospital struct {
	ID        uuid.UUID
	Name      string
	Code      string
	Location  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Person represents a hospital staff member with specialty constraints
type Person struct {
	ID        uuid.UUID
	Email     string // Primary identifier
	Name      string
	Specialty SpecialtyType // BODY_ONLY | NEURO_ONLY | BOTH
	Active    bool
	Aliases   []string // For matching Amion names
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// SpecialtyType defines radiologist specialties
type SpecialtyType string

const (
	SpecialtyBodyOnly  SpecialtyType = "BODY_ONLY"
	SpecialtyNeuroOnly SpecialtyType = "NEURO_ONLY"
	SpecialtyBoth      SpecialtyType = "BOTH"
)

// StudyType defines types of radiological studies
type StudyType string

const (
	StudyTypeGeneral   StudyType = "GENERAL"
	StudyTypeBodyImaging StudyType = "BODY"
	StudyTypeNeuroImaging StudyType = "NEURO"
)

// ShiftType defines shift naming conventions
type ShiftType string

const (
	ShiftTypeON1   ShiftType = "ON1"   // Overnight 1
	ShiftTypeON2   ShiftType = "ON2"   // Overnight 2
	ShiftTypeMidC  ShiftType = "MidC"  // Middle day first call
	ShiftTypeMidL  ShiftType = "MidL"  // Middle day last call
	ShiftTypeDay   ShiftType = "DAY"   // Day shift
)

// ScheduleVersion represents a temporal version of a schedule
// Enables time-travel queries, version management, and promotion workflows
type ScheduleVersion struct {
	ID                  uuid.UUID
	HospitalID          uuid.UUID
	Status              VersionStatus // STAGING | PRODUCTION | ARCHIVED
	EffectiveStartDate  time.Time
	EffectiveEndDate    time.Time
	ScrapeBatchID       *uuid.UUID // Soft-link (no hard FK)
	ValidationResults   *ValidationResult
	ShiftInstances      []ShiftInstance
	CreatedAt           time.Time
	CreatedBy           uuid.UUID
	UpdatedAt           time.Time
	UpdatedBy           uuid.UUID
	DeletedAt           *time.Time
	DeletedBy           *uuid.UUID
}

// VersionStatus represents the lifecycle state of a schedule
type VersionStatus string

const (
	VersionStatusStaging      VersionStatus = "STAGING"      // Not yet live
	VersionStatusProduction   VersionStatus = "PRODUCTION"   // Currently active
	VersionStatusArchived     VersionStatus = "ARCHIVED"     // Historical
)

// ShiftInstance represents a required shift with metadata
// Immutable once created as part of ScheduleVersion
type ShiftInstance struct {
	ID                  uuid.UUID
	ScheduleVersionID   uuid.UUID
	ShiftType           ShiftType
	ScheduleDate        time.Time
	StartTime           string // HH:MM format
	EndTime             string // HH:MM format
	HospitalID          uuid.UUID
	StudyType           StudyType
	SpecialtyConstraint SpecialtyType // Guides coverage resolution
	DesiredCoverage     int           // How many people needed
	IsMandatory         bool
	CreatedAt           time.Time
	CreatedBy           uuid.UUID
}

// Assignment maps a person to a shift
// Source tracking enables audit trail
type Assignment struct {
	ID                uuid.UUID
	PersonID          uuid.UUID
	ShiftInstanceID   uuid.UUID
	ScheduleDate      time.Time
	OriginalShiftType string // What Amion said
	Source            AssignmentSource
	CreatedAt         time.Time
	CreatedBy         uuid.UUID
	DeletedAt         *time.Time
	DeletedBy         *uuid.UUID
}

// AssignmentSource tracks where an assignment came from
type AssignmentSource string

const (
	AssignmentSourceAmion    AssignmentSource = "AMION"
	AssignmentSourceManual   AssignmentSource = "MANUAL"
	AssignmentSourceOverride AssignmentSource = "OVERRIDE"
)

// ScrapeBatch groups data from one scrape operation
// Provides atomic batch operations with full traceability
type ScrapeBatch struct {
	ID               uuid.UUID
	HospitalID       uuid.UUID
	State            BatchState // PENDING | COMPLETE | FAILED
	WindowStartDate  time.Time
	WindowEndDate    time.Time
	ScrapedAt        time.Time
	CompletedAt      *time.Time
	RowCount         int
	IngestChecksum   string    // Detects corrupted imports
	ErrorMessage     *string
	CreatedAt        time.Time
	CreatedBy        uuid.UUID
	DeletedAt        *time.Time // Soft delete
	DeletedBy        *uuid.UUID
	ArchivedAt       *time.Time // Archival support
	ArchivedBy       *uuid.UUID
}

// BatchState represents the lifecycle of a batch operation
type BatchState string

const (
	BatchStatePending   BatchState = "PENDING"
	BatchStateComplete  BatchState = "COMPLETE"
	BatchStateFailed    BatchState = "FAILED"
)

// AuditLog tracks all admin actions for compliance and debugging
type AuditLog struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Action     string    // e.g., "PROMOTE_VERSION", "IMPORT_ODS"
	Resource   string    // e.g., "ScheduleVersion#123"
	OldValues  string    // JSON
	NewValues  string    // JSON
	Timestamp  time.Time
	IPAddress  string
}

// CoverageCalculation represents calculated coverage for a schedule
type CoverageCalculation struct {
	ID                          uuid.UUID
	ScheduleVersionID           uuid.UUID
	HospitalID                  uuid.UUID
	CalculationDate             time.Time
	CalculationPeriodStartDate  time.Time
	CalculationPeriodEndDate    time.Time
	CoverageByPosition          map[string]int    // Position → count
	CoverageSummary             map[string]interface{}
	ValidationErrors            *ValidationResult
	QueryCount                  int // For performance testing
	CalculatedAt                time.Time
	CalculatedBy                uuid.UUID
}

// IsDeleted checks if an entity is soft-deleted
func (p *Person) IsDeleted() bool {
	return p.DeletedAt != nil
}

// IsDeleted checks if an assignment is soft-deleted
func (a *Assignment) IsDeleted() bool {
	return a.DeletedAt != nil
}

// IsDeleted checks if a batch is soft-deleted
func (b *ScrapeBatch) IsDeleted() bool {
	return b.DeletedAt != nil
}

// IsDeleted checks if a schedule version is soft-deleted
func (sv *ScheduleVersion) IsDeleted() bool {
	return sv.DeletedAt != nil
}

// SoftDelete marks a person as deleted
func (p *Person) SoftDelete(deleterID uuid.UUID) {
	now := time.Now().UTC()
	p.DeletedAt = &now
}

// SoftDelete marks an assignment as deleted
func (a *Assignment) SoftDelete(deleterID uuid.UUID) {
	now := time.Now().UTC()
	a.DeletedAt = &now
	a.DeletedBy = &deleterID
}

// SoftDelete marks a batch as deleted
func (b *ScrapeBatch) SoftDelete(deleterID uuid.UUID) {
	now := time.Now().UTC()
	b.DeletedAt = &now
	b.DeletedBy = &deleterID
}

// SoftDelete marks a schedule version as deleted
func (sv *ScheduleVersion) SoftDelete(deleterID uuid.UUID) {
	now := time.Now().UTC()
	sv.DeletedAt = &now
	sv.DeletedBy = &deleterID
}

// MarkArchived marks a batch as archived
func (b *ScrapeBatch) MarkArchived(archiverID uuid.UUID) {
	now := time.Now().UTC()
	b.ArchivedAt = &now
	b.ArchivedBy = &archiverID
}

// MarkComplete transitions a batch to complete state
func (b *ScrapeBatch) MarkComplete(completerID uuid.UUID, rowCount int) {
	now := time.Now().UTC()
	b.State = BatchStateComplete
	b.CompletedAt = &now
	b.RowCount = rowCount
}

// MarkFailed transitions a batch to failed state
func (b *ScrapeBatch) MarkFailed(errorMsg string) {
	now := time.Now().UTC()
	b.State = BatchStateFailed
	b.CompletedAt = &now
	b.ErrorMessage = &errorMsg
}

// Promote transitions a version from STAGING to PRODUCTION
func (sv *ScheduleVersion) Promote(promoterID uuid.UUID) error {
	if sv.Status != VersionStatusStaging {
		return ErrInvalidVersionStateTransition
	}
	sv.Status = VersionStatusProduction
	sv.UpdatedAt = time.Now().UTC()
	sv.UpdatedBy = promoterID
	return nil
}

// Archive transitions a version to ARCHIVED
func (sv *ScheduleVersion) Archive(archiverID uuid.UUID) error {
	if sv.Status != VersionStatusProduction {
		return ErrCannotArchiveNonProduction
	}
	sv.Status = VersionStatusArchived
	sv.UpdatedAt = time.Now().UTC()
	sv.UpdatedBy = archiverID
	return nil
}

// User represents a system user with authentication and authorization
type User struct {
	ID         uuid.UUID
	Email      string    // Unique identifier
	Name       string
	PasswordHash string
	Role       UserRole  // ADMIN | SCHEDULER | VIEWER
	HospitalID *uuid.UUID // NULL for system admin
	Active     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	LastLoginAt *time.Time
	DeletedAt  *time.Time
}

// UserRole defines user authorization levels
type UserRole string

const (
	UserRoleAdmin     UserRole = "ADMIN"
	UserRoleScheduler UserRole = "SCHEDULER"
	UserRoleViewer    UserRole = "VIEWER"
)

// JobQueue represents an async job for processing
type JobQueue struct {
	ID          uuid.UUID
	JobType     string // ODS_IMPORT | AMION_IMPORT | COVERAGE_CALCULATION
	Payload     map[string]interface{} // Job-specific data
	Status      JobQueueStatus // PENDING | PROCESSING | COMPLETE | FAILED | RETRY
	Result      map[string]interface{}
	ErrorMessage *string
	RetryCount  int
	MaxRetries  int
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// JobQueueStatus represents the status of a job in the queue
type JobQueueStatus string

const (
	JobQueueStatusPending    JobQueueStatus = "PENDING"
	JobQueueStatusProcessing JobQueueStatus = "PROCESSING"
	JobQueueStatusComplete   JobQueueStatus = "COMPLETE"
	JobQueueStatusFailed     JobQueueStatus = "FAILED"
	JobQueueStatusRetry      JobQueueStatus = "RETRY"
)
