package helpers

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
)

// Factory functions create valid entities with sensible defaults

// CreateValidPerson creates a valid Person with all required fields
func CreateValidPerson() *entity.Person {
	return NewPersonBuilder().Build()
}

// CreateValidPersonWithEmail creates a valid Person with a specific email
func CreateValidPersonWithEmail(email string) *entity.Person {
	return NewPersonBuilder().
		WithEmail(email).
		Build()
}

// CreateValidPersonWithSpecialty creates a valid Person with a specific specialty
func CreateValidPersonWithSpecialty(specialty entity.SpecialtyType) *entity.Person {
	return NewPersonBuilder().
		WithSpecialty(specialty).
		Build()
}

// CreateValidPersonInactive creates a valid but inactive Person
func CreateValidPersonInactive() *entity.Person {
	return NewPersonBuilder().
		WithActive(false).
		Build()
}

// CreateValidPersonDeleted creates a valid but deleted Person
func CreateValidPersonDeleted() *entity.Person {
	now := time.Now().UTC()
	return NewPersonBuilder().
		WithDeletedAt(&now).
		Build()
}

// CreateValidShiftInstance creates a valid ShiftInstance with all required fields
func CreateValidShiftInstance() *entity.ShiftInstance {
	return NewShiftInstanceBuilder().Build()
}

// CreateValidShiftInstanceWithType creates a valid ShiftInstance with a specific shift type
func CreateValidShiftInstanceWithType(shiftType entity.ShiftType) *entity.ShiftInstance {
	return NewShiftInstanceBuilder().
		WithShiftType(shiftType).
		Build()
}

// CreateValidShiftInstanceWithDate creates a valid ShiftInstance on a specific date
func CreateValidShiftInstanceWithDate(date time.Time) *entity.ShiftInstance {
	return NewShiftInstanceBuilder().
		WithScheduleDate(date).
		Build()
}

// CreateValidShiftInstanceWithStudyType creates a valid ShiftInstance with a specific study type
func CreateValidShiftInstanceWithStudyType(studyType entity.StudyType) *entity.ShiftInstance {
	return NewShiftInstanceBuilder().
		WithStudyType(studyType).
		Build()
}

// CreateValidShiftInstanceOptional creates a non-mandatory ShiftInstance
func CreateValidShiftInstanceOptional() *entity.ShiftInstance {
	return NewShiftInstanceBuilder().
		WithIsMandatory(false).
		Build()
}

// CreateValidShiftInstanceWithCoverage creates a ShiftInstance requiring multiple people
func CreateValidShiftInstanceWithCoverage(coverage int) *entity.ShiftInstance {
	return NewShiftInstanceBuilder().
		WithDesiredCoverage(coverage).
		Build()
}

// CreateValidAssignment creates a valid Assignment with all required fields
func CreateValidAssignment() *entity.Assignment {
	return NewAssignmentBuilder().Build()
}

// CreateValidAssignmentWithSource creates a valid Assignment from a specific source
func CreateValidAssignmentWithSource(source entity.AssignmentSource) *entity.Assignment {
	return NewAssignmentBuilder().
		WithSource(source).
		Build()
}

// CreateValidAssignmentFromAmion creates a valid Assignment sourced from Amion
func CreateValidAssignmentFromAmion() *entity.Assignment {
	return NewAssignmentBuilder().
		WithSource(entity.AssignmentSourceAmion).
		Build()
}

// CreateValidAssignmentFromManual creates a valid Assignment sourced from manual entry
func CreateValidAssignmentFromManual() *entity.Assignment {
	return NewAssignmentBuilder().
		WithSource(entity.AssignmentSourceManual).
		Build()
}

// CreateValidAssignmentDeleted creates a valid but deleted Assignment
func CreateValidAssignmentDeleted() *entity.Assignment {
	now := time.Now().UTC()
	deleterID := uuid.New()
	return NewAssignmentBuilder().
		WithDeletedAt(&now).
		WithDeletedBy(&deleterID).
		Build()
}

// CreateValidScheduleVersion creates a valid ScheduleVersion in STAGING state
func CreateValidScheduleVersion() *entity.ScheduleVersion {
	return NewScheduleVersionBuilder().Build()
}

// CreateValidScheduleVersionProduction creates a valid ScheduleVersion in PRODUCTION state
func CreateValidScheduleVersionProduction() *entity.ScheduleVersion {
	return NewScheduleVersionBuilder().
		WithStatus(entity.VersionStatusProduction).
		Build()
}

// CreateValidScheduleVersionArchived creates a valid ScheduleVersion in ARCHIVED state
func CreateValidScheduleVersionArchived() *entity.ScheduleVersion {
	return NewScheduleVersionBuilder().
		WithStatus(entity.VersionStatusArchived).
		Build()
}

// CreateValidScheduleVersionWithShifts creates a ScheduleVersion with ShiftInstances
func CreateValidScheduleVersionWithShifts(shiftCount int) *entity.ScheduleVersion {
	shifts := make([]entity.ShiftInstance, shiftCount)
	scheduleVersionID := uuid.New()
	for i := 0; i < shiftCount; i++ {
		shifts[i] = *NewShiftInstanceBuilder().
			WithScheduleVersionID(scheduleVersionID).
			Build()
	}
	return NewScheduleVersionBuilder().
		WithID(scheduleVersionID).
		WithShiftInstances(shifts).
		Build()
}

// CreateValidScheduleVersionWithValidation creates a ScheduleVersion with validation results
func CreateValidScheduleVersionWithValidation(validResult *entity.ValidationResult) *entity.ScheduleVersion {
	return NewScheduleVersionBuilder().
		WithValidationResults(validResult).
		Build()
}

// CreateValidScrapeBatch creates a valid ScrapeBatch in PENDING state
func CreateValidScrapeBatch() *entity.ScrapeBatch {
	return NewScrapeBatchBuilder().Build()
}

// CreateValidScrapeBatchComplete creates a valid completed ScrapeBatch
func CreateValidScrapeBatchComplete() *entity.ScrapeBatch {
	now := time.Now().UTC()
	return NewScrapeBatchBuilder().
		WithState(entity.BatchStateComplete).
		WithCompletedAt(&now).
		WithRowCount(100).
		Build()
}

// CreateValidScrapeBatchFailed creates a valid failed ScrapeBatch
func CreateValidScrapeBatchFailed() *entity.ScrapeBatch {
	now := time.Now().UTC()
	errMsg := "Database connection failed"
	return NewScrapeBatchBuilder().
		WithState(entity.BatchStateFailed).
		WithCompletedAt(&now).
		WithErrorMessage(&errMsg).
		Build()
}

// CreateValidScrapeBatchArchived creates a valid archived ScrapeBatch
func CreateValidScrapeBatchArchived() *entity.ScrapeBatch {
	now := time.Now().UTC()
	return NewScrapeBatchBuilder().
		WithState(entity.BatchStateComplete).
		WithCompletedAt(&now).
		WithArchivedAt(&now).
		WithRowCount(50).
		Build()
}

// CreateValidScrapeBatchDeleted creates a valid deleted ScrapeBatch
func CreateValidScrapeBatchDeleted() *entity.ScrapeBatch {
	now := time.Now().UTC()
	deleterID := uuid.New()
	return NewScrapeBatchBuilder().
		WithDeletedAt(&now).
		WithDeletedBy(&deleterID).
		Build()
}

// CreateValidCoverageCalculation creates a valid CoverageCalculation
func CreateValidCoverageCalculation() *entity.CoverageCalculation {
	coverage := make(map[string]int)
	coverage["ER Doctor"] = 2
	coverage["Nurse"] = 4
	return NewCoverageCalculationBuilder().
		WithCoverageByPosition(coverage).
		Build()
}

// CreateValidCoverageCalculationWithMetrics creates a CoverageCalculation with summary metrics
func CreateValidCoverageCalculationWithMetrics() *entity.CoverageCalculation {
	coverage := make(map[string]int)
	coverage["ER Doctor"] = 2
	coverage["Nurse"] = 4
	coverage["Tech"] = 3

	summary := make(map[string]interface{})
	summary["total_positions"] = 3
	summary["total_required"] = 9
	summary["satisfaction_rate"] = 0.95

	return NewCoverageCalculationBuilder().
		WithCoverageByPosition(coverage).
		WithCoverageSummary(summary).
		Build()
}

// CreateValidCoverageCalculationWithValidationErrors creates a CoverageCalculation with validation errors
func CreateValidCoverageCalculationWithValidationErrors() *entity.CoverageCalculation {
	validResult := entity.NewValidationError(
		"UNKNOWN_PEOPLE",
		"Unknown radiologist: Dr. Smith",
	)

	return NewCoverageCalculationBuilder().
		WithValidationErrors(validResult).
		Build()
}

// BulkCreateValidPeople creates multiple valid Person entities
func BulkCreateValidPeople(count int) []*entity.Person {
	people := make([]*entity.Person, count)
	for i := 0; i < count; i++ {
		email := fmt.Sprintf("person%d@example.com", i+1)
		people[i] = CreateValidPersonWithEmail(email)
	}
	return people
}

// BulkCreateValidShiftInstances creates multiple valid ShiftInstance entities
func BulkCreateValidShiftInstances(count int) []*entity.ShiftInstance {
	shifts := make([]*entity.ShiftInstance, count)
	shiftTypes := []entity.ShiftType{
		entity.ShiftTypeDay,
		entity.ShiftTypeMidC,
		entity.ShiftTypeMidL,
		entity.ShiftTypeON1,
		entity.ShiftTypeON2,
	}
	for i := 0; i < count; i++ {
		shifts[i] = CreateValidShiftInstanceWithType(shiftTypes[i%len(shiftTypes)])
	}
	return shifts
}

// BulkCreateValidAssignments creates multiple valid Assignment entities
func BulkCreateValidAssignments(count int) []*entity.Assignment {
	assignments := make([]*entity.Assignment, count)
	sources := []entity.AssignmentSource{
		entity.AssignmentSourceAmion,
		entity.AssignmentSourceManual,
		entity.AssignmentSourceOverride,
	}
	for i := 0; i < count; i++ {
		assignments[i] = CreateValidAssignmentWithSource(sources[i%len(sources)])
	}
	return assignments
}

// CreateValidHospital creates a valid Hospital entity
func CreateValidHospital() *entity.Hospital {
	return &entity.Hospital{
		ID:        uuid.New(),
		Name:      "Test Hospital",
		Code:      "TESTHSP",
		Location:  "Test City, State",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// CreateValidHospitalWithCode creates a Hospital with a specific code
func CreateValidHospitalWithCode(code string) *entity.Hospital {
	return &entity.Hospital{
		ID:        uuid.New(),
		Name:      fmt.Sprintf("Hospital %s", code),
		Code:      code,
		Location:  "Test City, State",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// CreateValidUser creates a valid User with VIEWER role
func CreateValidUser() *entity.User {
	now := time.Now().UTC()
	return &entity.User{
		ID:          uuid.New(),
		Email:       "user@example.com",
		Name:        "Test User",
		PasswordHash: "hashed_password_here",
		Role:        entity.UserRoleViewer,
		HospitalID:  nil, // System admin (no hospital)
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CreateValidUserAdmin creates a valid User with ADMIN role
func CreateValidUserAdmin() *entity.User {
	now := time.Now().UTC()
	return &entity.User{
		ID:          uuid.New(),
		Email:       "admin@example.com",
		Name:        "Admin User",
		PasswordHash: "hashed_password_here",
		Role:        entity.UserRoleAdmin,
		HospitalID:  nil,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CreateValidUserScheduler creates a valid User with SCHEDULER role
func CreateValidUserScheduler() *entity.User {
	now := time.Now().UTC()
	hospitalID := uuid.New()
	return &entity.User{
		ID:          uuid.New(),
		Email:       "scheduler@example.com",
		Name:        "Scheduler User",
		PasswordHash: "hashed_password_here",
		Role:        entity.UserRoleScheduler,
		HospitalID:  &hospitalID,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CreateValidAuditLog creates a valid AuditLog entry
func CreateValidAuditLog() *entity.AuditLog {
	return &entity.AuditLog{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Action:    "PROMOTE_VERSION",
		Resource:  fmt.Sprintf("ScheduleVersion#%s", uuid.New().String()),
		OldValues: `{"status":"STAGING"}`,
		NewValues: `{"status":"PRODUCTION"}`,
		Timestamp: time.Now().UTC(),
		IPAddress: "192.168.1.1",
	}
}

// CreateValidJobQueue creates a valid JobQueue entry
func CreateValidJobQueue() *entity.JobQueue {
	now := time.Now().UTC()
	return &entity.JobQueue{
		ID:          uuid.New(),
		JobType:     "ODS_IMPORT",
		Payload:     make(map[string]interface{}),
		Status:      entity.JobQueueStatusPending,
		Result:      make(map[string]interface{}),
		RetryCount:  0,
		MaxRetries:  3,
		CreatedAt:   now,
		StartedAt:   nil,
		CompletedAt: nil,
	}
}
