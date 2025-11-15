package helpers

import (
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
)

// PersonBuilder builds Person entities with fluent interface
type PersonBuilder struct {
	id        uuid.UUID
	email     string
	name      string
	specialty entity.SpecialtyType
	active    bool
	aliases   []string
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

// NewPersonBuilder creates a new PersonBuilder
func NewPersonBuilder() *PersonBuilder {
	now := time.Now().UTC()
	return &PersonBuilder{
		id:        uuid.New(),
		email:     "person@example.com",
		name:      "Test Person",
		specialty: entity.SpecialtyBoth,
		active:    true,
		aliases:   []string{},
		createdAt: now,
		updatedAt: now,
	}
}

// Default creates a PersonBuilder with sensible defaults (calls NewPersonBuilder)
func PersonBuilder_Default() *PersonBuilder {
	return NewPersonBuilder()
}

func (pb *PersonBuilder) WithID(id uuid.UUID) *PersonBuilder {
	pb.id = id
	return pb
}

func (pb *PersonBuilder) WithEmail(email string) *PersonBuilder {
	pb.email = email
	return pb
}

func (pb *PersonBuilder) WithName(name string) *PersonBuilder {
	pb.name = name
	return pb
}

func (pb *PersonBuilder) WithSpecialty(specialty entity.SpecialtyType) *PersonBuilder {
	pb.specialty = specialty
	return pb
}

func (pb *PersonBuilder) WithActive(active bool) *PersonBuilder {
	pb.active = active
	return pb
}

func (pb *PersonBuilder) WithAliases(aliases []string) *PersonBuilder {
	pb.aliases = aliases
	return pb
}

func (pb *PersonBuilder) WithCreatedAt(createdAt time.Time) *PersonBuilder {
	pb.createdAt = createdAt
	return pb
}

func (pb *PersonBuilder) WithUpdatedAt(updatedAt time.Time) *PersonBuilder {
	pb.updatedAt = updatedAt
	return pb
}

func (pb *PersonBuilder) WithDeletedAt(deletedAt *time.Time) *PersonBuilder {
	pb.deletedAt = deletedAt
	return pb
}

// Build creates the Person entity
func (pb *PersonBuilder) Build() *entity.Person {
	return &entity.Person{
		ID:        pb.id,
		Email:     pb.email,
		Name:      pb.name,
		Specialty: pb.specialty,
		Active:    pb.active,
		Aliases:   pb.aliases,
		CreatedAt: pb.createdAt,
		UpdatedAt: pb.updatedAt,
		DeletedAt: pb.deletedAt,
	}
}

// ShiftInstanceBuilder builds ShiftInstance entities with fluent interface
type ShiftInstanceBuilder struct {
	id                  uuid.UUID
	scheduleVersionID   uuid.UUID
	shiftType           entity.ShiftType
	scheduleDate        time.Time
	startTime           string
	endTime             string
	hospitalID          uuid.UUID
	studyType           entity.StudyType
	specialtyConstraint entity.SpecialtyType
	desiredCoverage     int
	isMandatory         bool
	createdAt           time.Time
	createdBy           uuid.UUID
}

// NewShiftInstanceBuilder creates a new ShiftInstanceBuilder
func NewShiftInstanceBuilder() *ShiftInstanceBuilder {
	now := time.Now().UTC()
	return &ShiftInstanceBuilder{
		id:                uuid.New(),
		scheduleVersionID: uuid.New(),
		shiftType:         entity.ShiftTypeDay,
		scheduleDate:      now,
		startTime:         "08:00",
		endTime:           "16:00",
		hospitalID:        uuid.New(),
		studyType:         entity.StudyTypeGeneral,
		specialtyConstraint: entity.SpecialtyBoth,
		desiredCoverage:   1,
		isMandatory:       true,
		createdAt:         now,
		createdBy:         uuid.New(),
	}
}

// Default creates a ShiftInstanceBuilder with sensible defaults
func ShiftInstanceBuilder_Default() *ShiftInstanceBuilder {
	return NewShiftInstanceBuilder()
}

func (sib *ShiftInstanceBuilder) WithID(id uuid.UUID) *ShiftInstanceBuilder {
	sib.id = id
	return sib
}

func (sib *ShiftInstanceBuilder) WithScheduleVersionID(scheduleVersionID uuid.UUID) *ShiftInstanceBuilder {
	sib.scheduleVersionID = scheduleVersionID
	return sib
}

func (sib *ShiftInstanceBuilder) WithShiftType(shiftType entity.ShiftType) *ShiftInstanceBuilder {
	sib.shiftType = shiftType
	return sib
}

func (sib *ShiftInstanceBuilder) WithScheduleDate(scheduleDate time.Time) *ShiftInstanceBuilder {
	sib.scheduleDate = scheduleDate
	return sib
}

func (sib *ShiftInstanceBuilder) WithStartTime(startTime string) *ShiftInstanceBuilder {
	sib.startTime = startTime
	return sib
}

func (sib *ShiftInstanceBuilder) WithEndTime(endTime string) *ShiftInstanceBuilder {
	sib.endTime = endTime
	return sib
}

func (sib *ShiftInstanceBuilder) WithHospitalID(hospitalID uuid.UUID) *ShiftInstanceBuilder {
	sib.hospitalID = hospitalID
	return sib
}

func (sib *ShiftInstanceBuilder) WithStudyType(studyType entity.StudyType) *ShiftInstanceBuilder {
	sib.studyType = studyType
	return sib
}

func (sib *ShiftInstanceBuilder) WithSpecialtyConstraint(specialtyConstraint entity.SpecialtyType) *ShiftInstanceBuilder {
	sib.specialtyConstraint = specialtyConstraint
	return sib
}

func (sib *ShiftInstanceBuilder) WithDesiredCoverage(desiredCoverage int) *ShiftInstanceBuilder {
	sib.desiredCoverage = desiredCoverage
	return sib
}

func (sib *ShiftInstanceBuilder) WithIsMandatory(isMandatory bool) *ShiftInstanceBuilder {
	sib.isMandatory = isMandatory
	return sib
}

func (sib *ShiftInstanceBuilder) WithCreatedAt(createdAt time.Time) *ShiftInstanceBuilder {
	sib.createdAt = createdAt
	return sib
}

func (sib *ShiftInstanceBuilder) WithCreatedBy(createdBy uuid.UUID) *ShiftInstanceBuilder {
	sib.createdBy = createdBy
	return sib
}

// Build creates the ShiftInstance entity
func (sib *ShiftInstanceBuilder) Build() *entity.ShiftInstance {
	return &entity.ShiftInstance{
		ID:                  sib.id,
		ScheduleVersionID:   sib.scheduleVersionID,
		ShiftType:           sib.shiftType,
		ScheduleDate:        sib.scheduleDate,
		StartTime:           sib.startTime,
		EndTime:             sib.endTime,
		HospitalID:          sib.hospitalID,
		StudyType:           sib.studyType,
		SpecialtyConstraint: sib.specialtyConstraint,
		DesiredCoverage:     sib.desiredCoverage,
		IsMandatory:         sib.isMandatory,
		CreatedAt:           sib.createdAt,
		CreatedBy:           sib.createdBy,
	}
}

// AssignmentBuilder builds Assignment entities with fluent interface
type AssignmentBuilder struct {
	id                uuid.UUID
	personID          uuid.UUID
	shiftInstanceID   uuid.UUID
	scheduleDate      time.Time
	originalShiftType string
	source            entity.AssignmentSource
	createdAt         time.Time
	createdBy         uuid.UUID
	deletedAt         *time.Time
	deletedBy         *uuid.UUID
}

// NewAssignmentBuilder creates a new AssignmentBuilder
func NewAssignmentBuilder() *AssignmentBuilder {
	now := time.Now().UTC()
	return &AssignmentBuilder{
		id:                uuid.New(),
		personID:          uuid.New(),
		shiftInstanceID:   uuid.New(),
		scheduleDate:      now,
		originalShiftType: "DAY",
		source:            entity.AssignmentSourceAmion,
		createdAt:         now,
		createdBy:         uuid.New(),
	}
}

// Default creates an AssignmentBuilder with sensible defaults
func AssignmentBuilder_Default() *AssignmentBuilder {
	return NewAssignmentBuilder()
}

func (ab *AssignmentBuilder) WithID(id uuid.UUID) *AssignmentBuilder {
	ab.id = id
	return ab
}

func (ab *AssignmentBuilder) WithPersonID(personID uuid.UUID) *AssignmentBuilder {
	ab.personID = personID
	return ab
}

func (ab *AssignmentBuilder) WithShiftInstanceID(shiftInstanceID uuid.UUID) *AssignmentBuilder {
	ab.shiftInstanceID = shiftInstanceID
	return ab
}

func (ab *AssignmentBuilder) WithScheduleDate(scheduleDate time.Time) *AssignmentBuilder {
	ab.scheduleDate = scheduleDate
	return ab
}

func (ab *AssignmentBuilder) WithOriginalShiftType(originalShiftType string) *AssignmentBuilder {
	ab.originalShiftType = originalShiftType
	return ab
}

func (ab *AssignmentBuilder) WithSource(source entity.AssignmentSource) *AssignmentBuilder {
	ab.source = source
	return ab
}

func (ab *AssignmentBuilder) WithCreatedAt(createdAt time.Time) *AssignmentBuilder {
	ab.createdAt = createdAt
	return ab
}

func (ab *AssignmentBuilder) WithCreatedBy(createdBy uuid.UUID) *AssignmentBuilder {
	ab.createdBy = createdBy
	return ab
}

func (ab *AssignmentBuilder) WithDeletedAt(deletedAt *time.Time) *AssignmentBuilder {
	ab.deletedAt = deletedAt
	return ab
}

func (ab *AssignmentBuilder) WithDeletedBy(deletedBy *uuid.UUID) *AssignmentBuilder {
	ab.deletedBy = deletedBy
	return ab
}

// Build creates the Assignment entity
func (ab *AssignmentBuilder) Build() *entity.Assignment {
	return &entity.Assignment{
		ID:                ab.id,
		PersonID:          ab.personID,
		ShiftInstanceID:   ab.shiftInstanceID,
		ScheduleDate:      ab.scheduleDate,
		OriginalShiftType: ab.originalShiftType,
		Source:            ab.source,
		CreatedAt:         ab.createdAt,
		CreatedBy:         ab.createdBy,
		DeletedAt:         ab.deletedAt,
		DeletedBy:         ab.deletedBy,
	}
}

// ScheduleVersionBuilder builds ScheduleVersion entities with fluent interface
type ScheduleVersionBuilder struct {
	id                 uuid.UUID
	hospitalID         uuid.UUID
	status             entity.VersionStatus
	effectiveStartDate time.Time
	effectiveEndDate   time.Time
	scrapeBatchID      *uuid.UUID
	validationResults  *entity.ValidationResult
	shiftInstances     []entity.ShiftInstance
	createdAt          time.Time
	createdBy          uuid.UUID
	updatedAt          time.Time
	updatedBy          uuid.UUID
	deletedAt          *time.Time
	deletedBy          *uuid.UUID
}

// NewScheduleVersionBuilder creates a new ScheduleVersionBuilder
func NewScheduleVersionBuilder() *ScheduleVersionBuilder {
	now := time.Now().UTC()
	return &ScheduleVersionBuilder{
		id:                 uuid.New(),
		hospitalID:         uuid.New(),
		status:             entity.VersionStatusStaging,
		effectiveStartDate: now,
		effectiveEndDate:   now.AddDate(0, 0, 7),
		validationResults:  entity.NewValidationResult(),
		shiftInstances:     []entity.ShiftInstance{},
		createdAt:          now,
		createdBy:          uuid.New(),
		updatedAt:          now,
		updatedBy:          uuid.New(),
	}
}

// Default creates a ScheduleVersionBuilder with sensible defaults
func ScheduleVersionBuilder_Default() *ScheduleVersionBuilder {
	return NewScheduleVersionBuilder()
}

func (svb *ScheduleVersionBuilder) WithID(id uuid.UUID) *ScheduleVersionBuilder {
	svb.id = id
	return svb
}

func (svb *ScheduleVersionBuilder) WithHospitalID(hospitalID uuid.UUID) *ScheduleVersionBuilder {
	svb.hospitalID = hospitalID
	return svb
}

func (svb *ScheduleVersionBuilder) WithStatus(status entity.VersionStatus) *ScheduleVersionBuilder {
	svb.status = status
	return svb
}

func (svb *ScheduleVersionBuilder) WithEffectiveStartDate(effectiveStartDate time.Time) *ScheduleVersionBuilder {
	svb.effectiveStartDate = effectiveStartDate
	return svb
}

func (svb *ScheduleVersionBuilder) WithEffectiveEndDate(effectiveEndDate time.Time) *ScheduleVersionBuilder {
	svb.effectiveEndDate = effectiveEndDate
	return svb
}

func (svb *ScheduleVersionBuilder) WithScrapeBatchID(scrapeBatchID *uuid.UUID) *ScheduleVersionBuilder {
	svb.scrapeBatchID = scrapeBatchID
	return svb
}

func (svb *ScheduleVersionBuilder) WithValidationResults(validationResults *entity.ValidationResult) *ScheduleVersionBuilder {
	svb.validationResults = validationResults
	return svb
}

func (svb *ScheduleVersionBuilder) WithShiftInstances(shiftInstances []entity.ShiftInstance) *ScheduleVersionBuilder {
	svb.shiftInstances = shiftInstances
	return svb
}

func (svb *ScheduleVersionBuilder) WithCreatedAt(createdAt time.Time) *ScheduleVersionBuilder {
	svb.createdAt = createdAt
	return svb
}

func (svb *ScheduleVersionBuilder) WithCreatedBy(createdBy uuid.UUID) *ScheduleVersionBuilder {
	svb.createdBy = createdBy
	return svb
}

func (svb *ScheduleVersionBuilder) WithUpdatedAt(updatedAt time.Time) *ScheduleVersionBuilder {
	svb.updatedAt = updatedAt
	return svb
}

func (svb *ScheduleVersionBuilder) WithUpdatedBy(updatedBy uuid.UUID) *ScheduleVersionBuilder {
	svb.updatedBy = updatedBy
	return svb
}

func (svb *ScheduleVersionBuilder) WithDeletedAt(deletedAt *time.Time) *ScheduleVersionBuilder {
	svb.deletedAt = deletedAt
	return svb
}

func (svb *ScheduleVersionBuilder) WithDeletedBy(deletedBy *uuid.UUID) *ScheduleVersionBuilder {
	svb.deletedBy = deletedBy
	return svb
}

// Build creates the ScheduleVersion entity
func (svb *ScheduleVersionBuilder) Build() *entity.ScheduleVersion {
	return &entity.ScheduleVersion{
		ID:                 svb.id,
		HospitalID:         svb.hospitalID,
		Status:             svb.status,
		EffectiveStartDate: svb.effectiveStartDate,
		EffectiveEndDate:   svb.effectiveEndDate,
		ScrapeBatchID:      svb.scrapeBatchID,
		ValidationResults:  svb.validationResults,
		ShiftInstances:     svb.shiftInstances,
		CreatedAt:          svb.createdAt,
		CreatedBy:          svb.createdBy,
		UpdatedAt:          svb.updatedAt,
		UpdatedBy:          svb.updatedBy,
		DeletedAt:          svb.deletedAt,
		DeletedBy:          svb.deletedBy,
	}
}

// ScrapeBatchBuilder builds ScrapeBatch entities with fluent interface
type ScrapeBatchBuilder struct {
	id              uuid.UUID
	hospitalID      uuid.UUID
	state           entity.BatchState
	windowStartDate time.Time
	windowEndDate   time.Time
	scrapedAt       time.Time
	completedAt     *time.Time
	rowCount        int
	ingestChecksum  string
	errorMessage    *string
	createdAt       time.Time
	createdBy       uuid.UUID
	deletedAt       *time.Time
	deletedBy       *uuid.UUID
	archivedAt      *time.Time
	archivedBy      *uuid.UUID
}

// NewScrapeBatchBuilder creates a new ScrapeBatchBuilder
func NewScrapeBatchBuilder() *ScrapeBatchBuilder {
	now := time.Now().UTC()
	return &ScrapeBatchBuilder{
		id:              uuid.New(),
		hospitalID:      uuid.New(),
		state:           entity.BatchStatePending,
		windowStartDate: now,
		windowEndDate:   now.AddDate(0, 0, 7),
		scrapedAt:       now,
		rowCount:        0,
		ingestChecksum:  "default-checksum",
		createdAt:       now,
		createdBy:       uuid.New(),
	}
}

// Default creates a ScrapeBatchBuilder with sensible defaults
func ScrapeBatchBuilder_Default() *ScrapeBatchBuilder {
	return NewScrapeBatchBuilder()
}

func (sbb *ScrapeBatchBuilder) WithID(id uuid.UUID) *ScrapeBatchBuilder {
	sbb.id = id
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithHospitalID(hospitalID uuid.UUID) *ScrapeBatchBuilder {
	sbb.hospitalID = hospitalID
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithState(state entity.BatchState) *ScrapeBatchBuilder {
	sbb.state = state
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithWindowStartDate(windowStartDate time.Time) *ScrapeBatchBuilder {
	sbb.windowStartDate = windowStartDate
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithWindowEndDate(windowEndDate time.Time) *ScrapeBatchBuilder {
	sbb.windowEndDate = windowEndDate
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithScrapedAt(scrapedAt time.Time) *ScrapeBatchBuilder {
	sbb.scrapedAt = scrapedAt
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithCompletedAt(completedAt *time.Time) *ScrapeBatchBuilder {
	sbb.completedAt = completedAt
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithRowCount(rowCount int) *ScrapeBatchBuilder {
	sbb.rowCount = rowCount
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithIngestChecksum(ingestChecksum string) *ScrapeBatchBuilder {
	sbb.ingestChecksum = ingestChecksum
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithErrorMessage(errorMessage *string) *ScrapeBatchBuilder {
	sbb.errorMessage = errorMessage
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithCreatedAt(createdAt time.Time) *ScrapeBatchBuilder {
	sbb.createdAt = createdAt
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithCreatedBy(createdBy uuid.UUID) *ScrapeBatchBuilder {
	sbb.createdBy = createdBy
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithDeletedAt(deletedAt *time.Time) *ScrapeBatchBuilder {
	sbb.deletedAt = deletedAt
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithDeletedBy(deletedBy *uuid.UUID) *ScrapeBatchBuilder {
	sbb.deletedBy = deletedBy
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithArchivedAt(archivedAt *time.Time) *ScrapeBatchBuilder {
	sbb.archivedAt = archivedAt
	return sbb
}

func (sbb *ScrapeBatchBuilder) WithArchivedBy(archivedBy *uuid.UUID) *ScrapeBatchBuilder {
	sbb.archivedBy = archivedBy
	return sbb
}

// Build creates the ScrapeBatch entity
func (sbb *ScrapeBatchBuilder) Build() *entity.ScrapeBatch {
	return &entity.ScrapeBatch{
		ID:              sbb.id,
		HospitalID:      sbb.hospitalID,
		State:           sbb.state,
		WindowStartDate: sbb.windowStartDate,
		WindowEndDate:   sbb.windowEndDate,
		ScrapedAt:       sbb.scrapedAt,
		CompletedAt:     sbb.completedAt,
		RowCount:        sbb.rowCount,
		IngestChecksum:  sbb.ingestChecksum,
		ErrorMessage:    sbb.errorMessage,
		CreatedAt:       sbb.createdAt,
		CreatedBy:       sbb.createdBy,
		DeletedAt:       sbb.deletedAt,
		DeletedBy:       sbb.deletedBy,
		ArchivedAt:      sbb.archivedAt,
		ArchivedBy:      sbb.archivedBy,
	}
}

// CoverageCalculationBuilder builds CoverageCalculation entities
type CoverageCalculationBuilder struct {
	id                        uuid.UUID
	scheduleVersionID         uuid.UUID
	hospitalID                uuid.UUID
	calculationDate           time.Time
	calculationPeriodStartDate time.Time
	calculationPeriodEndDate  time.Time
	coverageByPosition        map[string]int
	coverageSummary           map[string]interface{}
	validationErrors          *entity.ValidationResult
	queryCount                int
	calculatedAt              time.Time
	calculatedBy              uuid.UUID
}

// NewCoverageCalculationBuilder creates a new CoverageCalculationBuilder
func NewCoverageCalculationBuilder() *CoverageCalculationBuilder {
	now := time.Now().UTC()
	return &CoverageCalculationBuilder{
		id:                        uuid.New(),
		scheduleVersionID:         uuid.New(),
		hospitalID:                uuid.New(),
		calculationDate:           now,
		calculationPeriodStartDate: now,
		calculationPeriodEndDate:  now.AddDate(0, 0, 7),
		coverageByPosition:        make(map[string]int),
		coverageSummary:           make(map[string]interface{}),
		validationErrors:          entity.NewValidationResult(),
		queryCount:                0,
		calculatedAt:              now,
		calculatedBy:              uuid.New(),
	}
}

// Default creates a CoverageCalculationBuilder with sensible defaults
func CoverageCalculationBuilder_Default() *CoverageCalculationBuilder {
	return NewCoverageCalculationBuilder()
}

func (ccb *CoverageCalculationBuilder) WithID(id uuid.UUID) *CoverageCalculationBuilder {
	ccb.id = id
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithScheduleVersionID(scheduleVersionID uuid.UUID) *CoverageCalculationBuilder {
	ccb.scheduleVersionID = scheduleVersionID
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithHospitalID(hospitalID uuid.UUID) *CoverageCalculationBuilder {
	ccb.hospitalID = hospitalID
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithCalculationDate(calculationDate time.Time) *CoverageCalculationBuilder {
	ccb.calculationDate = calculationDate
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithCalculationPeriodStartDate(calculationPeriodStartDate time.Time) *CoverageCalculationBuilder {
	ccb.calculationPeriodStartDate = calculationPeriodStartDate
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithCalculationPeriodEndDate(calculationPeriodEndDate time.Time) *CoverageCalculationBuilder {
	ccb.calculationPeriodEndDate = calculationPeriodEndDate
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithCoverageByPosition(coverageByPosition map[string]int) *CoverageCalculationBuilder {
	ccb.coverageByPosition = coverageByPosition
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithCoverageSummary(coverageSummary map[string]interface{}) *CoverageCalculationBuilder {
	ccb.coverageSummary = coverageSummary
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithValidationErrors(validationErrors *entity.ValidationResult) *CoverageCalculationBuilder {
	ccb.validationErrors = validationErrors
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithQueryCount(queryCount int) *CoverageCalculationBuilder {
	ccb.queryCount = queryCount
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithCalculatedAt(calculatedAt time.Time) *CoverageCalculationBuilder {
	ccb.calculatedAt = calculatedAt
	return ccb
}

func (ccb *CoverageCalculationBuilder) WithCalculatedBy(calculatedBy uuid.UUID) *CoverageCalculationBuilder {
	ccb.calculatedBy = calculatedBy
	return ccb
}

// Build creates the CoverageCalculation entity
func (ccb *CoverageCalculationBuilder) Build() *entity.CoverageCalculation {
	return &entity.CoverageCalculation{
		ID:                        ccb.id,
		ScheduleVersionID:         ccb.scheduleVersionID,
		HospitalID:                ccb.hospitalID,
		CalculationDate:           ccb.calculationDate,
		CalculationPeriodStartDate: ccb.calculationPeriodStartDate,
		CalculationPeriodEndDate:  ccb.calculationPeriodEndDate,
		CoverageByPosition:        ccb.coverageByPosition,
		CoverageSummary:           ccb.coverageSummary,
		ValidationErrors:          ccb.validationErrors,
		QueryCount:                ccb.queryCount,
		CalculatedAt:              ccb.calculatedAt,
		CalculatedBy:              ccb.calculatedBy,
	}
}
