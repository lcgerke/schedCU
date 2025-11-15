package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestPersonCreation tests person entity creation
func TestPersonCreation(t *testing.T) {
	id := uuid.New()
	person := &Person{
		ID:        id,
		Email:     "dr.smith@hospital.com",
		Name:      "Dr. Smith",
		Specialty: SpecialtyBoth,
		Active:    true,
		Aliases:   []string{"Smith, Dr.", "D. Smith"},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	assert.Equal(t, id, person.ID)
	assert.Equal(t, "dr.smith@hospital.com", person.Email)
	assert.Equal(t, SpecialtyBoth, person.Specialty)
	assert.False(t, person.IsDeleted())
}

// TestPersonSoftDelete tests soft delete functionality
func TestPersonSoftDelete(t *testing.T) {
	person := &Person{
		ID:        uuid.New(),
		Email:     "dr.smith@hospital.com",
		Name:      "Dr. Smith",
		Specialty: SpecialtyBodyOnly,
		Active:    true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	deleterID := uuid.New()
	person.SoftDelete(deleterID)

	assert.True(t, person.IsDeleted())
	assert.NotNil(t, person.DeletedAt)
}

// TestAssignmentCreation tests assignment entity creation
func TestAssignmentCreation(t *testing.T) {
	assignment := &Assignment{
		ID:              uuid.New(),
		PersonID:        uuid.New(),
		ShiftInstanceID: uuid.New(),
		ScheduleDate:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		OriginalShiftType: "ON1",
		Source:          AssignmentSourceAmion,
		CreatedAt:       time.Now().UTC(),
	}

	assert.Equal(t, AssignmentSourceAmion, assignment.Source)
	assert.Equal(t, "ON1", assignment.OriginalShiftType)
	assert.False(t, assignment.IsDeleted())
}

// TestAssignmentSoftDelete tests assignment soft delete
func TestAssignmentSoftDelete(t *testing.T) {
	assignment := &Assignment{
		ID:              uuid.New(),
		PersonID:        uuid.New(),
		ShiftInstanceID: uuid.New(),
		CreatedAt:       time.Now().UTC(),
	}

	deleterID := uuid.New()
	assignment.SoftDelete(deleterID)

	assert.True(t, assignment.IsDeleted())
	assert.NotNil(t, assignment.DeletedAt)
	assert.Equal(t, deleterID, *assignment.DeletedBy)
}

// TestScheduleVersionCreation tests schedule version creation
func TestScheduleVersionCreation(t *testing.T) {
	hospitalID := uuid.New()
	creatorID := uuid.New()

	version := &ScheduleVersion{
		ID:                 uuid.New(),
		HospitalID:         hospitalID,
		Status:             VersionStatusStaging,
		EffectiveStartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EffectiveEndDate:   time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		ShiftInstances:     []ShiftInstance{},
		CreatedAt:          time.Now().UTC(),
		CreatedBy:          creatorID,
		UpdatedAt:          time.Now().UTC(),
		UpdatedBy:          creatorID,
	}

	assert.Equal(t, VersionStatusStaging, version.Status)
	assert.False(t, version.IsDeleted())
	assert.Empty(t, version.ShiftInstances)
}

// TestScheduleVersionPromotion tests promoting version from staging to production
func TestScheduleVersionPromotion(t *testing.T) {
	version := &ScheduleVersion{
		ID:     uuid.New(),
		Status: VersionStatusStaging,
	}

	promoterID := uuid.New()
	err := version.Promote(promoterID)

	assert.NoError(t, err)
	assert.Equal(t, VersionStatusProduction, version.Status)
	assert.Equal(t, promoterID, version.UpdatedBy)
}

// TestScheduleVersionPromotionError tests promoting from invalid state
func TestScheduleVersionPromotionError(t *testing.T) {
	version := &ScheduleVersion{
		ID:     uuid.New(),
		Status: VersionStatusProduction,
	}

	err := version.Promote(uuid.New())

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidVersionStateTransition, err)
}

// TestScheduleVersionArchive tests archiving production version
func TestScheduleVersionArchive(t *testing.T) {
	version := &ScheduleVersion{
		ID:     uuid.New(),
		Status: VersionStatusProduction,
	}

	archiverID := uuid.New()
	err := version.Archive(archiverID)

	assert.NoError(t, err)
	assert.Equal(t, VersionStatusArchived, version.Status)
	assert.Equal(t, archiverID, version.UpdatedBy)
}

// TestScheduleVersionArchiveError tests archiving non-production version
func TestScheduleVersionArchiveError(t *testing.T) {
	version := &ScheduleVersion{
		ID:     uuid.New(),
		Status: VersionStatusStaging,
	}

	err := version.Archive(uuid.New())

	assert.Error(t, err)
	assert.Equal(t, ErrCannotArchiveNonProduction, err)
}

// TestScrapeBatchCreation tests batch creation
func TestScrapeBatchCreation(t *testing.T) {
	batch := &ScrapeBatch{
		ID:              uuid.New(),
		HospitalID:      uuid.New(),
		State:           BatchStatePending,
		WindowStartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		WindowEndDate:   time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		ScrapedAt:       time.Now().UTC(),
		CreatedAt:       time.Now().UTC(),
	}

	assert.Equal(t, BatchStatePending, batch.State)
	assert.False(t, batch.IsDeleted())
	assert.Nil(t, batch.CompletedAt)
}

// TestScrapeBatchCompletion tests completing a batch
func TestScrapeBatchCompletion(t *testing.T) {
	batch := &ScrapeBatch{
		ID:    uuid.New(),
		State: BatchStatePending,
	}

	completerId := uuid.New()
	batch.MarkComplete(completerId, 150)

	assert.Equal(t, BatchStateComplete, batch.State)
	assert.NotNil(t, batch.CompletedAt)
	assert.Equal(t, 150, batch.RowCount)
}

// TestScrapeBatchFailure tests failing a batch
func TestScrapeBatchFailure(t *testing.T) {
	batch := &ScrapeBatch{
		ID:    uuid.New(),
		State: BatchStatePending,
	}

	errorMsg := "Connection timeout to Amion"
	batch.MarkFailed(errorMsg)

	assert.Equal(t, BatchStateFailed, batch.State)
	assert.NotNil(t, batch.CompletedAt)
	assert.NotNil(t, batch.ErrorMessage)
	assert.Equal(t, errorMsg, *batch.ErrorMessage)
}

// TestScrapeBatchArchival tests archiving a batch
func TestScrapeBatchArchival(t *testing.T) {
	batch := &ScrapeBatch{
		ID:    uuid.New(),
		State: BatchStateComplete,
	}

	archiverID := uuid.New()
	batch.MarkArchived(archiverID)

	assert.NotNil(t, batch.ArchivedAt)
	assert.Equal(t, archiverID, *batch.ArchivedBy)
}

// TestScrapeBatchSoftDelete tests soft delete on batch
func TestScrapeBatchSoftDelete(t *testing.T) {
	batch := &ScrapeBatch{
		ID:    uuid.New(),
		State: BatchStatePending,
	}

	deleterID := uuid.New()
	batch.SoftDelete(deleterID)

	assert.True(t, batch.IsDeleted())
	assert.NotNil(t, batch.DeletedAt)
	assert.Equal(t, deleterID, *batch.DeletedBy)
}

// TestShiftInstanceCreation tests shift instance creation
func TestShiftInstanceCreation(t *testing.T) {
	shift := &ShiftInstance{
		ID:                  uuid.New(),
		ScheduleVersionID:   uuid.New(),
		ShiftType:           ShiftTypeON1,
		ScheduleDate:        time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		StartTime:           "22:00",
		EndTime:             "06:00",
		HospitalID:          uuid.New(),
		StudyType:           StudyTypeNeuroImaging,
		SpecialtyConstraint: SpecialtyNeuroOnly,
		DesiredCoverage:     2,
		IsMandatory:         true,
		CreatedAt:           time.Now().UTC(),
	}

	assert.Equal(t, ShiftTypeON1, shift.ShiftType)
	assert.Equal(t, SpecialtyNeuroOnly, shift.SpecialtyConstraint)
	assert.Equal(t, 2, shift.DesiredCoverage)
	assert.True(t, shift.IsMandatory)
}

// TestCoverageCalculationCreation tests coverage calculation creation
func TestCoverageCalculationCreation(t *testing.T) {
	coverage := &CoverageCalculation{
		ID:                         uuid.New(),
		ScheduleVersionID:          uuid.New(),
		HospitalID:                 uuid.New(),
		CalculationDate:            time.Now().UTC(),
		CalculationPeriodStartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		CalculationPeriodEndDate:   time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		CoverageByPosition: map[string]int{
			"ER_Doctor":      3,
			"ICU_Nurse":      2,
		},
		QueryCount: 5,
	}

	assert.Equal(t, 5, coverage.QueryCount)
	assert.Equal(t, 3, coverage.CoverageByPosition["ER_Doctor"])
}

// TestAuditLogCreation tests audit log creation
func TestAuditLogCreation(t *testing.T) {
	log := &AuditLog{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Action:    "PROMOTE_VERSION",
		Resource:  "ScheduleVersion#123",
		OldValues: `{"status":"STAGING"}`,
		NewValues: `{"status":"PRODUCTION"}`,
		Timestamp: time.Now().UTC(),
		IPAddress: "192.168.1.100",
	}

	assert.Equal(t, "PROMOTE_VERSION", log.Action)
	assert.Contains(t, log.OldValues, "STAGING")
	assert.Contains(t, log.NewValues, "PRODUCTION")
}

// TestValidateSpecialty tests specialty validation
func TestValidateSpecialty(t *testing.T) {
	assert.True(t, ValidateSpecialty("BODY_ONLY"))
	assert.True(t, ValidateSpecialty("NEURO_ONLY"))
	assert.True(t, ValidateSpecialty("BOTH"))
	assert.False(t, ValidateSpecialty("INVALID"))
	assert.False(t, ValidateSpecialty(""))
}

// TestValidateShiftType tests shift type validation
func TestValidateShiftType(t *testing.T) {
	assert.True(t, ValidateShiftType("ON1"))
	assert.True(t, ValidateShiftType("ON2"))
	assert.True(t, ValidateShiftType("MidC"))
	assert.True(t, ValidateShiftType("MidL"))
	assert.True(t, ValidateShiftType("DAY"))
	assert.False(t, ValidateShiftType("INVALID"))
	assert.False(t, ValidateShiftType(""))
}

// TestValidateVersionStatus tests version status validation
func TestValidateVersionStatus(t *testing.T) {
	assert.True(t, ValidateVersionStatus("STAGING"))
	assert.True(t, ValidateVersionStatus("PRODUCTION"))
	assert.True(t, ValidateVersionStatus("ARCHIVED"))
	assert.False(t, ValidateVersionStatus("INVALID"))
	assert.False(t, ValidateVersionStatus(""))
}

// TestValidateBatchState tests batch state validation
func TestValidateBatchState(t *testing.T) {
	assert.True(t, ValidateBatchState("PENDING"))
	assert.True(t, ValidateBatchState("COMPLETE"))
	assert.True(t, ValidateBatchState("FAILED"))
	assert.False(t, ValidateBatchState("INVALID"))
	assert.False(t, ValidateBatchState(""))
}
