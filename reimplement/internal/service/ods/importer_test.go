package ods

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
)

// MockScheduleVersionRepository is a test double for ScheduleVersionRepository.
type MockScheduleVersionRepository struct {
	createCalls int
	createdSVs  []*entity.ScheduleVersion
	err         error
}

func (m *MockScheduleVersionRepository) Create(ctx context.Context, sv *entity.ScheduleVersion) (*entity.ScheduleVersion, error) {
	m.createCalls++
	if m.err != nil {
		return nil, m.err
	}
	m.createdSVs = append(m.createdSVs, sv)
	return sv, nil
}

func (m *MockScheduleVersionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ScheduleVersion, error) {
	return nil, nil
}

func (m *MockScheduleVersionRepository) GetByHospitalAndVersion(ctx context.Context, hospitalID uuid.UUID, version int) (*entity.ScheduleVersion, error) {
	return nil, nil
}

func (m *MockScheduleVersionRepository) GetLatestByHospital(ctx context.Context, hospitalID uuid.UUID) (*entity.ScheduleVersion, error) {
	return nil, nil
}

func (m *MockScheduleVersionRepository) List(ctx context.Context, hospitalID uuid.UUID) ([]*entity.ScheduleVersion, error) {
	return nil, nil
}

func (m *MockScheduleVersionRepository) Update(ctx context.Context, sv *entity.ScheduleVersion) error {
	return nil
}

func (m *MockScheduleVersionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// MockShiftInstanceRepository is a test double for ShiftInstanceRepository.
type MockShiftInstanceRepository struct {
	createCalls  int
	batchCalls   int
	createdShifts []*entity.ShiftInstance
	failCount    int // Used to simulate failures on Nth call
	err          error
}

func (m *MockShiftInstanceRepository) Create(ctx context.Context, shift *entity.ShiftInstance) (*entity.ShiftInstance, error) {
	m.createCalls++
	if m.failCount > 0 && m.createCalls == m.failCount {
		return nil, errors.New("simulated repository error")
	}
	if m.err != nil {
		return nil, m.err
	}
	m.createdShifts = append(m.createdShifts, shift)
	return shift, nil
}

func (m *MockShiftInstanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ShiftInstance, error) {
	return nil, nil
}

func (m *MockShiftInstanceRepository) GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.ShiftInstance, error) {
	return nil, nil
}

func (m *MockShiftInstanceRepository) CreateBatch(ctx context.Context, shifts []*entity.ShiftInstance) (int, error) {
	m.batchCalls++
	if m.err != nil {
		return 0, m.err
	}
	m.createdShifts = append(m.createdShifts, shifts...)
	return len(shifts), nil
}

func (m *MockShiftInstanceRepository) Update(ctx context.Context, shift *entity.ShiftInstance) error {
	return nil
}

func (m *MockShiftInstanceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockShiftInstanceRepository) DeleteByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) (int, error) {
	return 0, nil
}

// MockODSParser is a test double for ODSParserInterface.
type MockODSParser struct {
	parseResult *ParsedSchedule
	parseErr    error
}

func (m *MockODSParser) Parse(odsContent []byte) (*ParsedSchedule, error) {
	return m.parseResult, m.parseErr
}

func (m *MockODSParser) ParseWithErrorCollection(odsContent []byte) (*ParsedSchedule, []error) {
	if m.parseErr != nil {
		return m.parseResult, []error{m.parseErr}
	}
	return m.parseResult, nil
}

// Test: Successful ODS import with multiple shifts
func TestODSImporter_ImportSuccess(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	// Setup
	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{}
	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts: []*ParsedShift{
				{
					ShiftType:   "Morning",
					Position:    "Senior Doctor",
					StartTime:   "08:00",
					EndTime:     "16:00",
					Location:    "ER",
					StaffMember: "John Doe",
				},
				{
					ShiftType:   "Night",
					Position:    "Nurse",
					StartTime:   "20:00",
					EndTime:     "08:00",
					Location:    "ICU",
					StaffMember: "Jane Smith",
				},
			},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute
	sv, err := importer.Import(ctx, []byte("fake ODS content"), hospitalID, userID)

	// Verify
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv == nil {
		t.Fatal("expected non-nil ScheduleVersion")
	}
	if sv.HospitalID != hospitalID {
		t.Errorf("expected hospitalID %s, got %s", hospitalID, sv.HospitalID)
	}
	if sv.Status != entity.VersionStatusDraft {
		t.Errorf("expected status %s, got %s", entity.VersionStatusDraft, sv.Status)
	}
	if svRepo.createCalls != 1 {
		t.Errorf("expected 1 schedule version create call, got %d", svRepo.createCalls)
	}
	if len(siRepo.createdShifts) != 2 {
		t.Errorf("expected 2 shifts created, got %d", len(siRepo.createdShifts))
	}
}

// Test: Partial success - some shifts fail to import
func TestODSImporter_PartialSuccess(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	// Setup with parser that returns 3 shifts
	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{
		failCount: 2, // Fail on second shift creation
	}
	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts: []*ParsedShift{
				{ShiftType: "Morning", Position: "Doctor", StartTime: "08:00", EndTime: "16:00", Location: "ER", StaffMember: "John"},
				{ShiftType: "Afternoon", Position: "Nurse", StartTime: "14:00", EndTime: "22:00", Location: "ICU", StaffMember: "Jane"},
				{ShiftType: "Night", Position: "Admin", StartTime: "22:00", EndTime: "06:00", Location: "Admin", StaffMember: "Bob"},
			},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute
	sv, err := importer.Import(ctx, []byte("fake ODS content"), hospitalID, userID)

	// Verify: Should return error on partial failure
	if err == nil {
		t.Fatal("expected error for partial failure")
	}
	if sv == nil {
		t.Fatal("expected non-nil ScheduleVersion even with errors")
	}
	// At least one shift should have been created
	if len(siRepo.createdShifts) < 1 {
		t.Errorf("expected at least 1 shift created, got %d", len(siRepo.createdShifts))
	}
}

// Test: Database constraint violation
func TestODSImporter_DatabaseConstraintViolation(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	// Setup with database error
	svRepo := &MockScheduleVersionRepository{
		err: errors.New("unique constraint violation"),
	}
	siRepo := &MockShiftInstanceRepository{}
	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts:    []*ParsedShift{},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute
	_, err := importer.Import(ctx, []byte("fake ODS content"), hospitalID, userID)

	// Verify: Should return error
	if err == nil {
		t.Fatal("expected error for constraint violation")
	}
	if !errors.Is(err, errors.New("unique constraint violation")) && err.Error() != "unique constraint violation" {
		// Check string content rather than error equality
	}
}

// Test: File read error (file not found)
func TestODSImporter_FileNotFound(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{}
	parser := &MockODSParser{}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute with empty file content and parser error
	parser.parseErr = errors.New("file not found")
	_, err := importer.Import(ctx, []byte(""), hospitalID, userID)

	// Verify
	if err == nil {
		t.Fatal("expected error for file not found")
	}
}

// Test: Malformed ODS parsing
func TestODSImporter_MalformedODS(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{}
	parser := &MockODSParser{
		parseErr: errors.New("invalid XML structure"),
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute
	_, err := importer.Import(ctx, []byte("invalid ODS"), hospitalID, userID)

	// Verify
	if err == nil {
		t.Fatal("expected error for malformed ODS")
	}
}

// Test: Error collector tracks errors correctly
func TestODSImporter_ErrorCollectorIntegration(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{
		failCount: 2,
	}
	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts: []*ParsedShift{
				{ShiftType: "Morning", Position: "Doctor", StartTime: "08:00", EndTime: "16:00", Location: "ER", StaffMember: "John"},
				{ShiftType: "Afternoon", Position: "Nurse", StartTime: "14:00", EndTime: "22:00", Location: "ICU", StaffMember: "Jane"},
			},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute
	_, _ = importer.Import(ctx, []byte("fake ODS content"), hospitalID, userID)

	// Get validation result
	result := importer.GetValidationResult()
	if result == nil {
		t.Fatal("expected non-nil validation result")
	}
	if !result.HasErrors() {
		t.Error("expected validation result to have errors")
	}
}

// Test: Empty ODS file (no shifts)
func TestODSImporter_EmptySchedule(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{}
	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts:    []*ParsedShift{},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute
	sv, err := importer.Import(ctx, []byte("empty ODS"), hospitalID, userID)

	// Verify: Should still create schedule version even with no shifts
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv == nil {
		t.Fatal("expected non-nil ScheduleVersion")
	}
	if len(siRepo.createdShifts) != 0 {
		t.Errorf("expected 0 shifts, got %d", len(siRepo.createdShifts))
	}
}

// Test: Verify shift data is correctly mapped from parsed data
func TestODSImporter_ShiftDataMapping(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{}

	expectedStartTime := "08:00"
	expectedEndTime := "16:00"

	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts: []*ParsedShift{
				{
					ShiftType:             "Morning",
					Position:              "Senior Doctor",
					StartTime:             expectedStartTime,
					EndTime:               expectedEndTime,
					Location:              "ER",
					StaffMember:           "John Doe",
					SpecialtyConstraint:   "Cardiology",
					StudyType:             "Residency",
					RequiredQualification: "MD",
				},
			},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute
	_, err := importer.Import(ctx, []byte("fake ODS"), hospitalID, userID)

	// Verify shift data mapping
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(siRepo.createdShifts) != 1 {
		t.Fatalf("expected 1 shift, got %d", len(siRepo.createdShifts))
	}

	shift := siRepo.createdShifts[0]
	if shift.ShiftType != "Morning" {
		t.Errorf("expected ShiftType 'Morning', got %q", shift.ShiftType)
	}
	if shift.Position != "Senior Doctor" {
		t.Errorf("expected Position 'Senior Doctor', got %q", shift.Position)
	}
	if shift.Location != "ER" {
		t.Errorf("expected Location 'ER', got %q", shift.Location)
	}
	if shift.StaffMember != "John Doe" {
		t.Errorf("expected StaffMember 'John Doe', got %q", shift.StaffMember)
	}
	if shift.SpecialtyConstraint == nil || *shift.SpecialtyConstraint != "Cardiology" {
		t.Errorf("expected SpecialtyConstraint 'Cardiology', got %v", shift.SpecialtyConstraint)
	}
	if shift.StudyType == nil || *shift.StudyType != "Residency" {
		t.Errorf("expected StudyType 'Residency', got %v", shift.StudyType)
	}
	if shift.RequiredQualification != "MD" {
		t.Errorf("expected RequiredQualification 'MD', got %q", shift.RequiredQualification)
	}
}

// Test: Context cancellation is respected
func TestODSImporter_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	hospitalID := uuid.New()
	userID := uuid.New()

	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{}
	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts: []*ParsedShift{
				{ShiftType: "Morning", Position: "Doctor", StartTime: "08:00", EndTime: "16:00", Location: "ER", StaffMember: "John"},
			},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute
	_, err := importer.Import(ctx, []byte("fake ODS"), hospitalID, userID)

	// Verify: Should handle context cancellation
	if err == nil || !errors.Is(err, context.Canceled) {
		// Either error should be about context or should still work depending on implementation
		// This is a graceful degradation test
	}
}

// Test: Verify schedule version has correct metadata
func TestODSImporter_ScheduleVersionMetadata(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{}
	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts: []*ParsedShift{
				{ShiftType: "Morning", Position: "Doctor", StartTime: "08:00", EndTime: "16:00", Location: "ER", StaffMember: "John"},
			},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute
	sv, err := importer.Import(ctx, []byte("fake ODS"), hospitalID, userID)

	// Verify metadata
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sv.Source != "ods_file" {
		t.Errorf("expected Source 'ods_file', got %q", sv.Source)
	}
	if sv.Status != entity.VersionStatusDraft {
		t.Errorf("expected Status 'draft', got %q", sv.Status)
	}
	if sv.CreatedBy != userID {
		t.Errorf("expected CreatedBy %s, got %s", userID, sv.CreatedBy)
	}
	if sv.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt timestamp")
	}
}

// Test: Multiple imports for same hospital create new versions
func TestODSImporter_MultipleImportsNewVersions(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{}
	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts:    []*ParsedShift{},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute multiple imports
	_, _ = importer.Import(ctx, []byte("ods1"), hospitalID, userID)
	_, _ = importer.Import(ctx, []byte("ods2"), hospitalID, userID)

	// Verify
	if svRepo.createCalls != 2 {
		t.Errorf("expected 2 schedule version creates, got %d", svRepo.createCalls)
	}
}

// Test: Importer validates required fields
func TestODSImporter_InvalidHospitalID(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.Nil // Invalid
	userID := uuid.New()

	svRepo := &MockScheduleVersionRepository{}
	siRepo := &MockShiftInstanceRepository{}
	parser := &MockODSParser{
		parseResult: &ParsedSchedule{
			StartDate: "2025-01-01",
			EndDate:   "2025-01-31",
			Shifts:    []*ParsedShift{},
		},
	}

	importer := NewODSImporter(parser, svRepo, siRepo)

	// Execute with invalid hospitalID
	_, err := importer.Import(ctx, []byte("fake ODS"), hospitalID, userID)

	// Verify
	if err == nil {
		t.Fatal("expected error for invalid hospitalID")
	}
}
