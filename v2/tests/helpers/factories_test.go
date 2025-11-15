package helpers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
)

// TestCreateValidPerson verifies factory creates valid Person
func TestCreateValidPerson(t *testing.T) {
	person := CreateValidPerson()

	if person.ID == uuid.Nil {
		t.Error("expected person ID to be set")
	}
	if person.Email == "" {
		t.Error("expected email to be set")
	}
	if person.Name == "" {
		t.Error("expected name to be set")
	}
	if !person.Active {
		t.Error("expected person to be active by default")
	}
}

// TestCreateValidPersonWithEmail verifies factory sets custom email
func TestCreateValidPersonWithEmail(t *testing.T) {
	email := "custom@hospital.com"
	person := CreateValidPersonWithEmail(email)

	if person.Email != email {
		t.Error("expected custom email")
	}
}

// TestCreateValidPersonWithSpecialty verifies factory sets specialty
func TestCreateValidPersonWithSpecialty(t *testing.T) {
	specialty := entity.SpecialtyNeuroOnly
	person := CreateValidPersonWithSpecialty(specialty)

	if person.Specialty != specialty {
		t.Error("expected specialty to be set")
	}
}

// TestCreateValidPersonInactive verifies factory creates inactive person
func TestCreateValidPersonInactive(t *testing.T) {
	person := CreateValidPersonInactive()

	if person.Active {
		t.Error("expected person to be inactive")
	}
}

// TestCreateValidPersonDeleted verifies factory creates deleted person
func TestCreateValidPersonDeleted(t *testing.T) {
	person := CreateValidPersonDeleted()

	if person.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
	if !person.IsDeleted() {
		t.Error("expected person to be marked as deleted")
	}
}

// TestCreateValidShiftInstance verifies factory creates valid ShiftInstance
func TestCreateValidShiftInstance(t *testing.T) {
	shift := CreateValidShiftInstance()

	if shift.ID == uuid.Nil {
		t.Error("expected shift ID to be set")
	}
	if shift.ScheduleVersionID == uuid.Nil {
		t.Error("expected schedule version ID to be set")
	}
	if shift.StartTime == "" {
		t.Error("expected start time to be set")
	}
	if shift.EndTime == "" {
		t.Error("expected end time to be set")
	}
}

// TestCreateValidShiftInstanceWithType verifies factory sets shift type
func TestCreateValidShiftInstanceWithType(t *testing.T) {
	shiftType := entity.ShiftTypeMidC
	shift := CreateValidShiftInstanceWithType(shiftType)

	if shift.ShiftType != shiftType {
		t.Error("expected shift type to be set")
	}
}

// TestCreateValidShiftInstanceWithDate verifies factory sets schedule date
func TestCreateValidShiftInstanceWithDate(t *testing.T) {
	// Get current date
	shift := CreateValidShiftInstance()
	originalDate := shift.ScheduleDate

	// Create with specific date
	shift2 := CreateValidShiftInstanceWithDate(originalDate)
	if shift2.ScheduleDate != originalDate {
		t.Error("expected schedule date to match")
	}
}

// TestCreateValidShiftInstanceWithStudyType verifies factory sets study type
func TestCreateValidShiftInstanceWithStudyType(t *testing.T) {
	studyType := entity.StudyTypeBodyImaging
	shift := CreateValidShiftInstanceWithStudyType(studyType)

	if shift.StudyType != studyType {
		t.Error("expected study type to be set")
	}
}

// TestCreateValidShiftInstanceOptional verifies factory creates optional shift
func TestCreateValidShiftInstanceOptional(t *testing.T) {
	shift := CreateValidShiftInstanceOptional()

	if shift.IsMandatory {
		t.Error("expected shift to be optional")
	}
}

// TestCreateValidShiftInstanceWithCoverage verifies factory sets coverage
func TestCreateValidShiftInstanceWithCoverage(t *testing.T) {
	coverage := 5
	shift := CreateValidShiftInstanceWithCoverage(coverage)

	if shift.DesiredCoverage != coverage {
		t.Error("expected coverage to be set")
	}
}

// TestCreateValidAssignment verifies factory creates valid Assignment
func TestCreateValidAssignment(t *testing.T) {
	assignment := CreateValidAssignment()

	if assignment.ID == uuid.Nil {
		t.Error("expected assignment ID to be set")
	}
	if assignment.PersonID == uuid.Nil {
		t.Error("expected person ID to be set")
	}
	if assignment.ShiftInstanceID == uuid.Nil {
		t.Error("expected shift instance ID to be set")
	}
}

// TestCreateValidAssignmentWithSource verifies factory sets source
func TestCreateValidAssignmentWithSource(t *testing.T) {
	source := entity.AssignmentSourceManual
	assignment := CreateValidAssignmentWithSource(source)

	if assignment.Source != source {
		t.Error("expected source to be set")
	}
}

// TestCreateValidAssignmentFromAmion verifies factory creates Amion assignment
func TestCreateValidAssignmentFromAmion(t *testing.T) {
	assignment := CreateValidAssignmentFromAmion()

	if assignment.Source != entity.AssignmentSourceAmion {
		t.Error("expected source to be AMION")
	}
}

// TestCreateValidAssignmentFromManual verifies factory creates manual assignment
func TestCreateValidAssignmentFromManual(t *testing.T) {
	assignment := CreateValidAssignmentFromManual()

	if assignment.Source != entity.AssignmentSourceManual {
		t.Error("expected source to be MANUAL")
	}
}

// TestCreateValidAssignmentDeleted verifies factory creates deleted assignment
func TestCreateValidAssignmentDeleted(t *testing.T) {
	assignment := CreateValidAssignmentDeleted()

	if assignment.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
	if assignment.DeletedBy == nil {
		t.Error("expected DeletedBy to be set")
	}
	if !assignment.IsDeleted() {
		t.Error("expected assignment to be marked as deleted")
	}
}

// TestCreateValidScheduleVersion verifies factory creates valid ScheduleVersion
func TestCreateValidScheduleVersion(t *testing.T) {
	scheduleVersion := CreateValidScheduleVersion()

	if scheduleVersion.ID == uuid.Nil {
		t.Error("expected schedule version ID to be set")
	}
	if scheduleVersion.HospitalID == uuid.Nil {
		t.Error("expected hospital ID to be set")
	}
	if scheduleVersion.Status != entity.VersionStatusStaging {
		t.Error("expected status to be STAGING by default")
	}
}

// TestCreateValidScheduleVersionProduction verifies factory creates production version
func TestCreateValidScheduleVersionProduction(t *testing.T) {
	scheduleVersion := CreateValidScheduleVersionProduction()

	if scheduleVersion.Status != entity.VersionStatusProduction {
		t.Error("expected status to be PRODUCTION")
	}
}

// TestCreateValidScheduleVersionArchived verifies factory creates archived version
func TestCreateValidScheduleVersionArchived(t *testing.T) {
	scheduleVersion := CreateValidScheduleVersionArchived()

	if scheduleVersion.Status != entity.VersionStatusArchived {
		t.Error("expected status to be ARCHIVED")
	}
}

// TestCreateValidScheduleVersionWithShifts verifies factory creates version with shifts
func TestCreateValidScheduleVersionWithShifts(t *testing.T) {
	shiftCount := 5
	scheduleVersion := CreateValidScheduleVersionWithShifts(shiftCount)

	if len(scheduleVersion.ShiftInstances) != shiftCount {
		t.Errorf("expected %d shifts, got %d", shiftCount, len(scheduleVersion.ShiftInstances))
	}

	// Verify all shifts belong to this version
	for _, shift := range scheduleVersion.ShiftInstances {
		if shift.ScheduleVersionID != scheduleVersion.ID {
			t.Error("expected all shifts to reference this schedule version")
		}
	}
}

// TestCreateValidScrapeBatch verifies factory creates valid ScrapeBatch
func TestCreateValidScrapeBatch(t *testing.T) {
	batch := CreateValidScrapeBatch()

	if batch.ID == uuid.Nil {
		t.Error("expected batch ID to be set")
	}
	if batch.HospitalID == uuid.Nil {
		t.Error("expected hospital ID to be set")
	}
	if batch.State != entity.BatchStatePending {
		t.Error("expected state to be PENDING by default")
	}
}

// TestCreateValidScrapeBatchComplete verifies factory creates completed batch
func TestCreateValidScrapeBatchComplete(t *testing.T) {
	batch := CreateValidScrapeBatchComplete()

	if batch.State != entity.BatchStateComplete {
		t.Error("expected state to be COMPLETE")
	}
	if batch.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
	if batch.RowCount != 100 {
		t.Error("expected row count to be 100")
	}
}

// TestCreateValidScrapeBatchFailed verifies factory creates failed batch
func TestCreateValidScrapeBatchFailed(t *testing.T) {
	batch := CreateValidScrapeBatchFailed()

	if batch.State != entity.BatchStateFailed {
		t.Error("expected state to be FAILED")
	}
	if batch.ErrorMessage == nil {
		t.Error("expected error message to be set")
	}
}

// TestCreateValidScrapeBatchArchived verifies factory creates archived batch
func TestCreateValidScrapeBatchArchived(t *testing.T) {
	batch := CreateValidScrapeBatchArchived()

	if batch.State != entity.BatchStateComplete {
		t.Error("expected state to be COMPLETE for archival")
	}
	if batch.ArchivedAt == nil {
		t.Error("expected ArchivedAt to be set")
	}
}

// TestCreateValidScrapeBatchDeleted verifies factory creates deleted batch
func TestCreateValidScrapeBatchDeleted(t *testing.T) {
	batch := CreateValidScrapeBatchDeleted()

	if batch.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
	if batch.DeletedBy == nil {
		t.Error("expected DeletedBy to be set")
	}
	if !batch.IsDeleted() {
		t.Error("expected batch to be marked as deleted")
	}
}

// TestCreateValidCoverageCalculation verifies factory creates valid CoverageCalculation
func TestCreateValidCoverageCalculation(t *testing.T) {
	coverage := CreateValidCoverageCalculation()

	if coverage.ID == uuid.Nil {
		t.Error("expected coverage ID to be set")
	}
	if coverage.ScheduleVersionID == uuid.Nil {
		t.Error("expected schedule version ID to be set")
	}
	if len(coverage.CoverageByPosition) == 0 {
		t.Error("expected coverage by position to be populated")
	}
}

// TestCreateValidCoverageCalculationWithMetrics verifies factory creates coverage with metrics
func TestCreateValidCoverageCalculationWithMetrics(t *testing.T) {
	coverage := CreateValidCoverageCalculationWithMetrics()

	if len(coverage.CoverageByPosition) == 0 {
		t.Error("expected coverage by position to be populated")
	}
	if len(coverage.CoverageSummary) == 0 {
		t.Error("expected coverage summary to be populated")
	}

	// Verify specific metrics
	if _, ok := coverage.CoverageSummary["total_positions"]; !ok {
		t.Error("expected total_positions in summary")
	}
	if _, ok := coverage.CoverageSummary["total_required"]; !ok {
		t.Error("expected total_required in summary")
	}
}

// TestCreateValidCoverageCalculationWithValidationErrors verifies factory creates coverage with errors
func TestCreateValidCoverageCalculationWithValidationErrors(t *testing.T) {
	coverage := CreateValidCoverageCalculationWithValidationErrors()

	if coverage.ValidationErrors == nil {
		t.Error("expected validation errors to be set")
	}
	if coverage.ValidationErrors.Valid {
		t.Error("expected validation errors to be invalid")
	}
}

// TestBulkCreateValidPeople verifies bulk factory creates multiple valid entities
func TestBulkCreateValidPeople(t *testing.T) {
	count := 10
	people := BulkCreateValidPeople(count)

	if len(people) != count {
		t.Errorf("expected %d people, got %d", count, len(people))
	}

	// Verify all are valid
	for i, person := range people {
		if person.ID == uuid.Nil {
			t.Errorf("person %d: expected ID to be set", i)
		}
		if person.Email == "" {
			t.Errorf("person %d: expected email to be set", i)
		}
	}

	// Verify emails are unique
	emailMap := make(map[string]bool)
	for _, person := range people {
		if emailMap[person.Email] {
			t.Error("expected all emails to be unique")
		}
		emailMap[person.Email] = true
	}
}

// TestBulkCreateValidShiftInstances verifies bulk factory creates multiple valid entities
func TestBulkCreateValidShiftInstances(t *testing.T) {
	count := 10
	shifts := BulkCreateValidShiftInstances(count)

	if len(shifts) != count {
		t.Errorf("expected %d shifts, got %d", count, len(shifts))
	}

	// Verify all are valid
	for i, shift := range shifts {
		if shift.ID == uuid.Nil {
			t.Errorf("shift %d: expected ID to be set", i)
		}
		if shift.ScheduleVersionID == uuid.Nil {
			t.Errorf("shift %d: expected schedule version ID to be set", i)
		}
	}

	// Verify shift types are distributed
	typeMap := make(map[entity.ShiftType]int)
	for _, shift := range shifts {
		typeMap[shift.ShiftType]++
	}
	if len(typeMap) == 0 {
		t.Error("expected shift types to be distributed")
	}
}

// TestBulkCreateValidAssignments verifies bulk factory creates multiple valid entities
func TestBulkCreateValidAssignments(t *testing.T) {
	count := 10
	assignments := BulkCreateValidAssignments(count)

	if len(assignments) != count {
		t.Errorf("expected %d assignments, got %d", count, len(assignments))
	}

	// Verify all are valid
	for i, assignment := range assignments {
		if assignment.ID == uuid.Nil {
			t.Errorf("assignment %d: expected ID to be set", i)
		}
		if assignment.PersonID == uuid.Nil {
			t.Errorf("assignment %d: expected person ID to be set", i)
		}
	}

	// Verify sources are distributed
	sourceMap := make(map[entity.AssignmentSource]int)
	for _, assignment := range assignments {
		sourceMap[assignment.Source]++
	}
	if len(sourceMap) == 0 {
		t.Error("expected sources to be distributed")
	}
}

// TestCreateValidHospital verifies factory creates valid Hospital
func TestCreateValidHospital(t *testing.T) {
	hospital := CreateValidHospital()

	if hospital.ID == uuid.Nil {
		t.Error("expected hospital ID to be set")
	}
	if hospital.Name == "" {
		t.Error("expected hospital name to be set")
	}
	if hospital.Code == "" {
		t.Error("expected hospital code to be set")
	}
}

// TestCreateValidHospitalWithCode verifies factory creates hospital with specific code
func TestCreateValidHospitalWithCode(t *testing.T) {
	code := "CUSTOM_CODE"
	hospital := CreateValidHospitalWithCode(code)

	if hospital.Code != code {
		t.Error("expected hospital code to match")
	}
}

// TestCreateValidUser verifies factory creates valid User
func TestCreateValidUser(t *testing.T) {
	user := CreateValidUser()

	if user.ID == uuid.Nil {
		t.Error("expected user ID to be set")
	}
	if user.Email == "" {
		t.Error("expected user email to be set")
	}
	if user.Role != entity.UserRoleViewer {
		t.Error("expected user to have VIEWER role by default")
	}
	if !user.Active {
		t.Error("expected user to be active")
	}
}

// TestCreateValidUserAdmin verifies factory creates admin user
func TestCreateValidUserAdmin(t *testing.T) {
	user := CreateValidUserAdmin()

	if user.Role != entity.UserRoleAdmin {
		t.Error("expected user to have ADMIN role")
	}
	if user.HospitalID != nil {
		t.Error("expected admin user to have no hospital affiliation")
	}
}

// TestCreateValidUserScheduler verifies factory creates scheduler user
func TestCreateValidUserScheduler(t *testing.T) {
	user := CreateValidUserScheduler()

	if user.Role != entity.UserRoleScheduler {
		t.Error("expected user to have SCHEDULER role")
	}
	if user.HospitalID == nil {
		t.Error("expected scheduler user to have hospital affiliation")
	}
}

// TestCreateValidAuditLog verifies factory creates valid AuditLog
func TestCreateValidAuditLog(t *testing.T) {
	log := CreateValidAuditLog()

	if log.ID == uuid.Nil {
		t.Error("expected audit log ID to be set")
	}
	if log.UserID == uuid.Nil {
		t.Error("expected user ID to be set")
	}
	if log.Action == "" {
		t.Error("expected action to be set")
	}
	if log.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

// TestCreateValidJobQueue verifies factory creates valid JobQueue
func TestCreateValidJobQueue(t *testing.T) {
	job := CreateValidJobQueue()

	if job.ID == uuid.Nil {
		t.Error("expected job ID to be set")
	}
	if job.JobType == "" {
		t.Error("expected job type to be set")
	}
	if job.Status != entity.JobQueueStatusPending {
		t.Error("expected job status to be PENDING by default")
	}
	if job.MaxRetries != 3 {
		t.Error("expected max retries to be 3")
	}
}

// BenchmarkFactory_Person benchmarks Person factory
func BenchmarkFactory_Person(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CreateValidPerson()
	}
}

// BenchmarkFactory_ShiftInstance benchmarks ShiftInstance factory
func BenchmarkFactory_ShiftInstance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CreateValidShiftInstance()
	}
}

// BenchmarkFactory_BulkPeople benchmarks bulk Person creation
func BenchmarkFactory_BulkPeople(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BulkCreateValidPeople(10)
	}
}
