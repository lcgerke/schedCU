package helpers

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
)

// TestPersonBuilder_Default verifies PersonBuilder creates valid entities with defaults
func TestPersonBuilder_Default(t *testing.T) {
	person := NewPersonBuilder().Build()

	if person.ID == uuid.Nil {
		t.Error("expected person ID to be set")
	}
	if person.Email != "person@example.com" {
		t.Error("expected default email")
	}
	if person.Name != "Test Person" {
		t.Error("expected default name")
	}
	if person.Specialty != entity.SpecialtyBoth {
		t.Error("expected specialty to be BOTH")
	}
	if !person.Active {
		t.Error("expected person to be active")
	}
	if person.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

// TestPersonBuilder_WithMethods verifies builder methods chain and set values
func TestPersonBuilder_WithMethods(t *testing.T) {
	testID := uuid.New()
	testEmail := "custom@example.com"
	testName := "Custom Person"
	testSpecialty := entity.SpecialtyBodyOnly

	person := NewPersonBuilder().
		WithID(testID).
		WithEmail(testEmail).
		WithName(testName).
		WithSpecialty(testSpecialty).
		WithActive(false).
		Build()

	if person.ID != testID {
		t.Error("expected custom ID")
	}
	if person.Email != testEmail {
		t.Error("expected custom email")
	}
	if person.Name != testName {
		t.Error("expected custom name")
	}
	if person.Specialty != testSpecialty {
		t.Error("expected custom specialty")
	}
	if person.Active {
		t.Error("expected person to be inactive")
	}
}

// TestPersonBuilder_SoftDelete verifies soft delete tracking
func TestPersonBuilder_SoftDelete(t *testing.T) {
	now := time.Now().UTC()
	person := NewPersonBuilder().
		WithDeletedAt(&now).
		Build()

	if person.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
	if person.IsDeleted() == false {
		t.Error("expected person to be marked as deleted")
	}
}

// TestShiftInstanceBuilder_Default verifies ShiftInstanceBuilder creates valid entities
func TestShiftInstanceBuilder_Default(t *testing.T) {
	shift := NewShiftInstanceBuilder().Build()

	if shift.ID == uuid.Nil {
		t.Error("expected shift ID to be set")
	}
	if shift.ScheduleVersionID == uuid.Nil {
		t.Error("expected schedule version ID to be set")
	}
	if shift.ShiftType != entity.ShiftTypeDay {
		t.Error("expected shift type to be DAY")
	}
	if shift.StartTime != "08:00" {
		t.Error("expected default start time")
	}
	if shift.EndTime != "16:00" {
		t.Error("expected default end time")
	}
	if !shift.IsMandatory {
		t.Error("expected shift to be mandatory")
	}
}

// TestShiftInstanceBuilder_AllShiftTypes verifies all shift type options work
func TestShiftInstanceBuilder_AllShiftTypes(t *testing.T) {
	shiftTypes := []entity.ShiftType{
		entity.ShiftTypeON1,
		entity.ShiftTypeON2,
		entity.ShiftTypeMidC,
		entity.ShiftTypeMidL,
		entity.ShiftTypeDay,
	}

	for _, shiftType := range shiftTypes {
		shift := NewShiftInstanceBuilder().
			WithShiftType(shiftType).
			Build()

		if shift.ShiftType != shiftType {
			t.Errorf("expected shift type %s, got %s", shiftType, shift.ShiftType)
		}
	}
}

// TestShiftInstanceBuilder_AllStudyTypes verifies all study type options work
func TestShiftInstanceBuilder_AllStudyTypes(t *testing.T) {
	studyTypes := []entity.StudyType{
		entity.StudyTypeGeneral,
		entity.StudyTypeBodyImaging,
		entity.StudyTypeNeuroImaging,
	}

	for _, studyType := range studyTypes {
		shift := NewShiftInstanceBuilder().
			WithStudyType(studyType).
			Build()

		if shift.StudyType != studyType {
			t.Errorf("expected study type %s, got %s", studyType, shift.StudyType)
		}
	}
}

// TestAssignmentBuilder_Default verifies AssignmentBuilder creates valid entities
func TestAssignmentBuilder_Default(t *testing.T) {
	assignment := NewAssignmentBuilder().Build()

	if assignment.ID == uuid.Nil {
		t.Error("expected assignment ID to be set")
	}
	if assignment.PersonID == uuid.Nil {
		t.Error("expected person ID to be set")
	}
	if assignment.ShiftInstanceID == uuid.Nil {
		t.Error("expected shift instance ID to be set")
	}
	if assignment.Source != entity.AssignmentSourceAmion {
		t.Error("expected source to be AMION")
	}
	if !assignment.ScheduleDate.IsZero() == false {
		t.Error("expected schedule date to be set")
	}
}

// TestAssignmentBuilder_AllSources verifies all assignment source options work
func TestAssignmentBuilder_AllSources(t *testing.T) {
	sources := []entity.AssignmentSource{
		entity.AssignmentSourceAmion,
		entity.AssignmentSourceManual,
		entity.AssignmentSourceOverride,
	}

	for _, source := range sources {
		assignment := NewAssignmentBuilder().
			WithSource(source).
			Build()

		if assignment.Source != source {
			t.Errorf("expected source %s, got %s", source, assignment.Source)
		}
	}
}

// TestAssignmentBuilder_SoftDelete verifies soft delete on assignments
func TestAssignmentBuilder_SoftDelete(t *testing.T) {
	now := time.Now().UTC()
	deleterID := uuid.New()
	assignment := NewAssignmentBuilder().
		WithDeletedAt(&now).
		WithDeletedBy(&deleterID).
		Build()

	if assignment.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
	if assignment.DeletedBy == nil {
		t.Error("expected DeletedBy to be set")
	}
	if assignment.IsDeleted() == false {
		t.Error("expected assignment to be marked as deleted")
	}
}

// TestScheduleVersionBuilder_Default verifies ScheduleVersionBuilder creates valid entities
func TestScheduleVersionBuilder_Default(t *testing.T) {
	scheduleVersion := NewScheduleVersionBuilder().Build()

	if scheduleVersion.ID == uuid.Nil {
		t.Error("expected schedule version ID to be set")
	}
	if scheduleVersion.HospitalID == uuid.Nil {
		t.Error("expected hospital ID to be set")
	}
	if scheduleVersion.Status != entity.VersionStatusStaging {
		t.Error("expected status to be STAGING")
	}
	if scheduleVersion.ValidationResults == nil {
		t.Error("expected validation results to be initialized")
	}
	if scheduleVersion.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

// TestScheduleVersionBuilder_AllStatuses verifies all version status options work
func TestScheduleVersionBuilder_AllStatuses(t *testing.T) {
	statuses := []entity.VersionStatus{
		entity.VersionStatusStaging,
		entity.VersionStatusProduction,
		entity.VersionStatusArchived,
	}

	for _, status := range statuses {
		scheduleVersion := NewScheduleVersionBuilder().
			WithStatus(status).
			Build()

		if scheduleVersion.Status != status {
			t.Errorf("expected status %s, got %s", status, scheduleVersion.Status)
		}
	}
}

// TestScheduleVersionBuilder_Promotion verifies version can be promoted
func TestScheduleVersionBuilder_Promotion(t *testing.T) {
	scheduleVersion := NewScheduleVersionBuilder().Build()

	if scheduleVersion.Status != entity.VersionStatusStaging {
		t.Error("expected initial status to be STAGING")
	}

	promoterID := uuid.New()
	err := scheduleVersion.Promote(promoterID)
	if err != nil {
		t.Errorf("unexpected error promoting version: %v", err)
	}

	if scheduleVersion.Status != entity.VersionStatusProduction {
		t.Error("expected status to be PRODUCTION after promotion")
	}
}

// TestScrapeBatchBuilder_Default verifies ScrapeBatchBuilder creates valid entities
func TestScrapeBatchBuilder_Default(t *testing.T) {
	batch := NewScrapeBatchBuilder().Build()

	if batch.ID == uuid.Nil {
		t.Error("expected batch ID to be set")
	}
	if batch.HospitalID == uuid.Nil {
		t.Error("expected hospital ID to be set")
	}
	if batch.State != entity.BatchStatePending {
		t.Error("expected state to be PENDING")
	}
	if batch.WindowStartDate.IsZero() {
		t.Error("expected window start date to be set")
	}
	if batch.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

// TestScrapeBatchBuilder_StateTransitions verifies batch state methods work
func TestScrapeBatchBuilder_StateTransitions(t *testing.T) {
	batch := NewScrapeBatchBuilder().Build()

	// Test pending state
	if batch.State != entity.BatchStatePending {
		t.Error("expected initial state to be PENDING")
	}

	// Test completion
	completerID := uuid.New()
	batch.MarkComplete(completerID, 100)
	if batch.State != entity.BatchStateComplete {
		t.Error("expected state to be COMPLETE")
	}
	if batch.RowCount != 100 {
		t.Error("expected row count to be 100")
	}

	// Test archiving
	archiverID := uuid.New()
	batch.MarkArchived(archiverID)
	if batch.ArchivedAt == nil {
		t.Error("expected ArchivedAt to be set")
	}
	if batch.IsDeleted() == false {
		// Note: ArchivedAt doesn't affect IsDeleted() - that's only DeletedAt
	}
}

// TestCoverageCalculationBuilder_Default verifies CoverageCalculationBuilder creates valid entities
func TestCoverageCalculationBuilder_Default(t *testing.T) {
	coverage := NewCoverageCalculationBuilder().Build()

	if coverage.ID == uuid.Nil {
		t.Error("expected coverage ID to be set")
	}
	if coverage.ScheduleVersionID == uuid.Nil {
		t.Error("expected schedule version ID to be set")
	}
	if coverage.HospitalID == uuid.Nil {
		t.Error("expected hospital ID to be set")
	}
	if coverage.CoverageByPosition == nil {
		t.Error("expected coverage by position to be initialized")
	}
	if coverage.CoverageSummary == nil {
		t.Error("expected coverage summary to be initialized")
	}
	if coverage.ValidationErrors == nil {
		t.Error("expected validation errors to be initialized")
	}
}

// TestBuilders_Immutability verifies builder fields don't affect other builders
func TestBuilders_Immutability(t *testing.T) {
	builder1 := NewPersonBuilder().WithEmail("person1@example.com")
	person1 := builder1.Build()

	builder2 := NewPersonBuilder().WithEmail("person2@example.com")
	person2 := builder2.Build()

	if person1.Email == person2.Email {
		t.Error("expected builders to be independent")
	}

	// Verify rebuilding with same builder uses last state
	person1b := builder1.Build()
	if person1b.Email != "person1@example.com" {
		t.Error("expected builder to remember state")
	}
}

// TestBuilders_ValidEntity_Person verifies Person entities are valid
func TestBuilders_ValidEntity_Person(t *testing.T) {
	person := NewPersonBuilder().Build()

	// Verify required fields
	if person.Email == "" {
		t.Error("email is required")
	}
	if person.Name == "" {
		t.Error("name is required")
	}
	if person.ID == uuid.Nil {
		t.Error("ID is required")
	}

	// Verify constraints
	if person.Specialty != entity.SpecialtyBoth &&
		person.Specialty != entity.SpecialtyBodyOnly &&
		person.Specialty != entity.SpecialtyNeuroOnly {
		t.Error("specialty must be valid enum value")
	}
}

// TestBuilders_ValidEntity_ShiftInstance verifies ShiftInstance entities are valid
func TestBuilders_ValidEntity_ShiftInstance(t *testing.T) {
	shift := NewShiftInstanceBuilder().Build()

	// Verify required fields
	if shift.ID == uuid.Nil {
		t.Error("ID is required")
	}
	if shift.ScheduleVersionID == uuid.Nil {
		t.Error("schedule version ID is required")
	}
	if shift.StartTime == "" {
		t.Error("start time is required")
	}
	if shift.EndTime == "" {
		t.Error("end time is required")
	}

	// Verify constraints
	if shift.DesiredCoverage < 1 {
		t.Error("desired coverage must be at least 1")
	}
}

// TestBuilders_ValidEntity_Assignment verifies Assignment entities are valid
func TestBuilders_ValidEntity_Assignment(t *testing.T) {
	assignment := NewAssignmentBuilder().Build()

	// Verify required fields
	if assignment.ID == uuid.Nil {
		t.Error("ID is required")
	}
	if assignment.PersonID == uuid.Nil {
		t.Error("person ID is required")
	}
	if assignment.ShiftInstanceID == uuid.Nil {
		t.Error("shift instance ID is required")
	}

	// Verify source is valid
	if assignment.Source != entity.AssignmentSourceAmion &&
		assignment.Source != entity.AssignmentSourceManual &&
		assignment.Source != entity.AssignmentSourceOverride {
		t.Error("source must be valid enum value")
	}
}

// TestBuilders_ValidEntity_ScheduleVersion verifies ScheduleVersion entities are valid
func TestBuilders_ValidEntity_ScheduleVersion(t *testing.T) {
	scheduleVersion := NewScheduleVersionBuilder().Build()

	// Verify required fields
	if scheduleVersion.ID == uuid.Nil {
		t.Error("ID is required")
	}
	if scheduleVersion.HospitalID == uuid.Nil {
		t.Error("hospital ID is required")
	}
	if scheduleVersion.CreatedAt.IsZero() {
		t.Error("created at is required")
	}

	// Verify date constraints
	if scheduleVersion.EffectiveStartDate.After(scheduleVersion.EffectiveEndDate) {
		t.Error("start date must be before end date")
	}

	// Verify status is valid
	if scheduleVersion.Status != entity.VersionStatusStaging &&
		scheduleVersion.Status != entity.VersionStatusProduction &&
		scheduleVersion.Status != entity.VersionStatusArchived {
		t.Error("status must be valid enum value")
	}
}

// TestBuilders_ValidEntity_ScrapeBatch verifies ScrapeBatch entities are valid
func TestBuilders_ValidEntity_ScrapeBatch(t *testing.T) {
	batch := NewScrapeBatchBuilder().Build()

	// Verify required fields
	if batch.ID == uuid.Nil {
		t.Error("ID is required")
	}
	if batch.HospitalID == uuid.Nil {
		t.Error("hospital ID is required")
	}
	if batch.CreatedAt.IsZero() {
		t.Error("created at is required")
	}

	// Verify date constraints
	if batch.WindowStartDate.After(batch.WindowEndDate) {
		t.Error("window start must be before window end")
	}

	// Verify state is valid
	if batch.State != entity.BatchStatePending &&
		batch.State != entity.BatchStateComplete &&
		batch.State != entity.BatchStateFailed {
		t.Error("state must be valid enum value")
	}
}

// TestBuilders_ValidEntity_CoverageCalculation verifies CoverageCalculation entities are valid
func TestBuilders_ValidEntity_CoverageCalculation(t *testing.T) {
	coverage := NewCoverageCalculationBuilder().Build()

	// Verify required fields
	if coverage.ID == uuid.Nil {
		t.Error("ID is required")
	}
	if coverage.ScheduleVersionID == uuid.Nil {
		t.Error("schedule version ID is required")
	}
	if coverage.HospitalID == uuid.Nil {
		t.Error("hospital ID is required")
	}
	if coverage.CalculatedAt.IsZero() {
		t.Error("calculated at is required")
	}
}

// BenchmarkPersonBuilder benchmarks Person entity creation
func BenchmarkPersonBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPersonBuilder().Build()
	}
}

// BenchmarkShiftInstanceBuilder benchmarks ShiftInstance entity creation
func BenchmarkShiftInstanceBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewShiftInstanceBuilder().Build()
	}
}

// BenchmarkAssignmentBuilder benchmarks Assignment entity creation
func BenchmarkAssignmentBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewAssignmentBuilder().Build()
	}
}

// BenchmarkScheduleVersionBuilder benchmarks ScheduleVersion entity creation
func BenchmarkScheduleVersionBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewScheduleVersionBuilder().Build()
	}
}

// BenchmarkComplexBuilder benchmarks creation with multiple With* calls
func BenchmarkComplexBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewScheduleVersionBuilder().
			WithStatus(entity.VersionStatusProduction).
			WithShiftInstances(make([]entity.ShiftInstance, 5)).
			WithValidationResults(entity.NewValidationResult()).
			Build()
	}
}
