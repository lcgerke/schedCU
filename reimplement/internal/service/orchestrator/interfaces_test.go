// Package orchestrator provides interfaces for coordinating schedule import and calculation workflows.
package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/validation"
)

// TestODSImportServiceContractInterfaceExists verifies ODSImportService interface exists
func TestODSImportServiceContractInterfaceExists(t *testing.T) {
	var _ ODSImportService = (*MockODSImportService)(nil)
}

// TestODSImportServiceImportScheduleSignature verifies ImportSchedule method signature
func TestODSImportServiceImportScheduleSignature(t *testing.T) {
	mock := &MockODSImportService{
		ImportScheduleFunc: func(
			ctx context.Context,
			filePath string,
			hospitalID uuid.UUID,
			userID uuid.UUID,
		) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			return &entity.ScheduleVersion{}, validation.NewValidationResult(), nil
		},
	}

	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	schedVersion, valResult, err := mock.ImportSchedule(ctx, "/tmp/test.ods", hospitalID, userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if schedVersion == nil {
		t.Fatal("expected ScheduleVersion, got nil")
	}
	if valResult == nil {
		t.Fatal("expected ValidationResult, got nil")
	}
}

// TestODSImportServiceMockTracksCalls verifies mock tracks ImportSchedule calls
func TestODSImportServiceMockTracksCalls(t *testing.T) {
	mock := &MockODSImportService{}
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	mock.ImportSchedule(ctx, "/tmp/test1.ods", hospitalID, userID)
	mock.ImportSchedule(ctx, "/tmp/test2.ods", hospitalID, userID)

	if len(mock.ImportScheduleCalls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(mock.ImportScheduleCalls))
	}
	if mock.ImportScheduleCalls[0].FilePath != "/tmp/test1.ods" {
		t.Errorf("expected /tmp/test1.ods, got %s", mock.ImportScheduleCalls[0].FilePath)
	}
	if mock.ImportScheduleCalls[1].FilePath != "/tmp/test2.ods" {
		t.Errorf("expected /tmp/test2.ods, got %s", mock.ImportScheduleCalls[1].FilePath)
	}
}

// TestODSImportServiceReturnsValidDefaultValues verifies default mock values are valid
func TestODSImportServiceReturnsValidDefaultValues(t *testing.T) {
	mock := &MockODSImportService{}
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	schedVersion, valResult, err := mock.ImportSchedule(ctx, "/tmp/test.ods", hospitalID, userID)

	if err != nil {
		t.Fatalf("default should return no error, got %v", err)
	}
	if schedVersion == nil {
		t.Fatal("default should return ScheduleVersion")
	}
	if schedVersion.ID == uuid.Nil {
		t.Fatal("ScheduleVersion ID should not be nil")
	}
	if schedVersion.HospitalID != hospitalID {
		t.Errorf("HospitalID should match input, got %s", schedVersion.HospitalID)
	}
	if schedVersion.CreatedBy != userID {
		t.Errorf("CreatedBy should match input, got %s", schedVersion.CreatedBy)
	}
	if schedVersion.Status != entity.VersionStatusDraft {
		t.Errorf("Status should be DRAFT, got %s", schedVersion.Status)
	}
	if valResult == nil {
		t.Fatal("default should return ValidationResult")
	}
	if valResult.HasErrors() {
		t.Fatal("default ValidationResult should not have errors")
	}
}

// TestAmionScraperServiceContractInterfaceExists verifies AmionScraperService interface exists
func TestAmionScraperServiceContractInterfaceExists(t *testing.T) {
	var _ AmionScraperService = (*MockAmionScraperService)(nil)
}

// TestAmionScraperServiceScrapeScheduleSignature verifies ScrapeSchedule method signature
func TestAmionScraperServiceScrapeScheduleSignature(t *testing.T) {
	mock := &MockAmionScraperService{
		ScrapeScheduleFunc: func(
			ctx context.Context,
			startDate time.Time,
			monthCount int,
			hospitalID uuid.UUID,
			userID uuid.UUID,
		) ([]entity.Assignment, *validation.ValidationResult, error) {
			return []entity.Assignment{}, validation.NewValidationResult(), nil
		},
	}

	ctx := context.Background()
	startDate := time.Now()
	hospitalID := uuid.New()

	assignments, valResult, err := mock.ScrapeSchedule(ctx, startDate, 3, hospitalID, uuid.New())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if assignments == nil {
		t.Fatal("expected assignments slice, got nil")
	}
	if valResult == nil {
		t.Fatal("expected ValidationResult, got nil")
	}
}

// TestAmionScraperServiceMockTracksCalls verifies mock tracks ScrapeSchedule calls
func TestAmionScraperServiceMockTracksCalls(t *testing.T) {
	mock := &MockAmionScraperService{}
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()
	startDate1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	startDate2 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	mock.ScrapeSchedule(ctx, startDate1, 1, hospitalID, userID)
	mock.ScrapeSchedule(ctx, startDate2, 2, hospitalID, userID)

	if len(mock.ScrapeScheduleCalls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(mock.ScrapeScheduleCalls))
	}
	if mock.ScrapeScheduleCalls[0].MonthCount != 1 {
		t.Errorf("expected monthCount 1, got %d", mock.ScrapeScheduleCalls[0].MonthCount)
	}
	if mock.ScrapeScheduleCalls[1].MonthCount != 2 {
		t.Errorf("expected monthCount 2, got %d", mock.ScrapeScheduleCalls[1].MonthCount)
	}
}

// TestCoverageCalculatorServiceContractInterfaceExists verifies CoverageCalculatorService interface exists
func TestCoverageCalculatorServiceContractInterfaceExists(t *testing.T) {
	var _ CoverageCalculatorService = (*MockCoverageCalculatorService)(nil)
}

// TestCoverageCalculatorServiceCalculateSignature verifies Calculate method signature
func TestCoverageCalculatorServiceCalculateSignature(t *testing.T) {
	mock := &MockCoverageCalculatorService{
		CalculateFunc: func(
			ctx context.Context,
			scheduleVersionID uuid.UUID,
		) (*CoverageMetrics, error) {
			return &CoverageMetrics{}, nil
		},
	}

	ctx := context.Background()
	scheduleVersionID := uuid.New()

	metrics, err := mock.Calculate(ctx, scheduleVersionID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if metrics == nil {
		t.Fatal("expected CoverageMetrics, got nil")
	}
}

// TestCoverageCalculatorServiceMockTracksCalls verifies mock tracks Calculate calls
func TestCoverageCalculatorServiceMockTracksCalls(t *testing.T) {
	mock := &MockCoverageCalculatorService{}
	ctx := context.Background()
	scheduleVersionID1 := uuid.New()
	scheduleVersionID2 := uuid.New()

	mock.Calculate(ctx, scheduleVersionID1)
	mock.Calculate(ctx, scheduleVersionID2)

	if len(mock.CalculateCalls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(mock.CalculateCalls))
	}
	if mock.CalculateCalls[0].ScheduleVersionID != scheduleVersionID1 {
		t.Errorf("expected scheduleVersionID1, got %s", mock.CalculateCalls[0].ScheduleVersionID)
	}
	if mock.CalculateCalls[1].ScheduleVersionID != scheduleVersionID2 {
		t.Errorf("expected scheduleVersionID2, got %s", mock.CalculateCalls[1].ScheduleVersionID)
	}
}

// TestCoverageMetricsStructure verifies CoverageMetrics fields are accessible
func TestCoverageMetricsStructure(t *testing.T) {
	scheduleVersionID := uuid.New()
	metrics := &CoverageMetrics{
		ScheduleVersionID:   scheduleVersionID,
		CoveragePercentage:  75.5,
		AssignedPositions:   10,
		RequiredPositions:   13,
		UncoveredShifts:     make([]*entity.ShiftInstance, 0),
		OverallocatedShifts: make([]*entity.ShiftInstance, 0),
		CalculatedAt:        time.Now(),
		Details:             make(map[string]interface{}),
	}

	if metrics.ScheduleVersionID != scheduleVersionID {
		t.Errorf("ScheduleVersionID mismatch")
	}
	if metrics.CoveragePercentage != 75.5 {
		t.Errorf("CoveragePercentage mismatch")
	}
	if metrics.AssignedPositions != 10 {
		t.Errorf("AssignedPositions mismatch")
	}
	if metrics.RequiredPositions != 13 {
		t.Errorf("RequiredPositions mismatch")
	}
}

// TestOrchestrationStatusConstants verifies OrchestrationStatus values
func TestOrchestrationStatusConstants(t *testing.T) {
	tests := []struct {
		status OrchestrationStatus
		want   string
	}{
		{OrchestrationStatusIDLE, "IDLE"},
		{OrchestrationStatusINPROGRESS, "IN_PROGRESS"},
		{OrchestrationStatusCOMPLETED, "COMPLETED"},
		{OrchestrationStatusFAILED, "FAILED"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("expected %s, got %s", tt.want, string(tt.status))
		}
	}
}

// TestOrchestrationResultStructure verifies OrchestrationResult fields are accessible
func TestOrchestrationResultStructure(t *testing.T) {
	scheduleVersionID := uuid.New()
	hospitalID := uuid.New()
	now := time.Now()

	result := &OrchestrationResult{
		ScheduleVersion: &entity.ScheduleVersion{
			ID:        scheduleVersionID,
			HospitalID: hospitalID,
		},
		Assignments: []entity.Assignment{},
		Coverage: &CoverageMetrics{
			ScheduleVersionID:   scheduleVersionID,
			CoveragePercentage:  100.0,
			AssignedPositions:   5,
			RequiredPositions:   5,
			CalculatedAt:        now,
		},
		ValidationResult: validation.NewValidationResult(),
		Duration:         100 * time.Millisecond,
		CompletedAt:      now,
		Metadata:         make(map[string]interface{}),
	}

	if result.ScheduleVersion == nil {
		t.Fatal("ScheduleVersion should not be nil")
	}
	if result.Assignments == nil {
		t.Fatal("Assignments should not be nil")
	}
	if result.Coverage == nil {
		t.Fatal("Coverage should not be nil")
	}
	if result.ValidationResult == nil {
		t.Fatal("ValidationResult should not be nil")
	}
	if result.Duration != 100*time.Millisecond {
		t.Errorf("Duration mismatch")
	}
	if result.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}
}

// TestScheduleOrchestratorContractInterfaceExists verifies ScheduleOrchestrator interface exists
func TestScheduleOrchestratorContractInterfaceExists(t *testing.T) {
	var _ ScheduleOrchestrator = (*MockScheduleOrchestrator)(nil)
}

// TestScheduleOrchestratorExecuteImportSignature verifies ExecuteImport method signature
func TestScheduleOrchestratorExecuteImportSignature(t *testing.T) {
	mock := NewMockScheduleOrchestrator()
	mock.ExecuteImportFunc = func(
		ctx context.Context,
		filePath string,
		hospitalID uuid.UUID,
		userID uuid.UUID,
	) (*OrchestrationResult, error) {
		return &OrchestrationResult{
			ScheduleVersion:  &entity.ScheduleVersion{},
			Assignments:      []entity.Assignment{},
			Coverage:         &CoverageMetrics{},
			ValidationResult: validation.NewValidationResult(),
		}, nil
	}

	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	result, err := mock.ExecuteImport(ctx, "/tmp/test.ods", hospitalID, userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected OrchestrationResult, got nil")
	}
}

// TestScheduleOrchestratorGetOrchestrationStatusSignature verifies GetOrchestrationStatus method signature
func TestScheduleOrchestratorGetOrchestrationStatusSignature(t *testing.T) {
	mock := NewMockScheduleOrchestrator()

	status := mock.GetOrchestrationStatus()

	if status != OrchestrationStatusIDLE {
		t.Errorf("expected IDLE, got %s", status)
	}
}

// TestScheduleOrchestratorMockTracksCalls verifies mock tracks ExecuteImport calls
func TestScheduleOrchestratorMockTracksCalls(t *testing.T) {
	mock := NewMockScheduleOrchestrator()
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	mock.ExecuteImport(ctx, "/tmp/test1.ods", hospitalID, userID)
	mock.ExecuteImport(ctx, "/tmp/test2.ods", hospitalID, userID)

	if len(mock.ExecuteImportCalls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(mock.ExecuteImportCalls))
	}
	if mock.ExecuteImportCalls[0].FilePath != "/tmp/test1.ods" {
		t.Errorf("expected /tmp/test1.ods, got %s", mock.ExecuteImportCalls[0].FilePath)
	}
	if mock.ExecuteImportCalls[1].FilePath != "/tmp/test2.ods" {
		t.Errorf("expected /tmp/test2.ods, got %s", mock.ExecuteImportCalls[1].FilePath)
	}
}

// TestScheduleOrchestratorStatusTransitions verifies status changes correctly
func TestScheduleOrchestratorStatusTransitions(t *testing.T) {
	mock := NewMockScheduleOrchestrator()

	// Initial status should be IDLE
	if mock.GetOrchestrationStatus() != OrchestrationStatusIDLE {
		t.Errorf("expected IDLE, got %s", mock.GetOrchestrationStatus())
	}

	// Status should change to IN_PROGRESS then COMPLETED during ExecuteImport
	mock.ExecuteImport(context.Background(), "/tmp/test.ods", uuid.New(), uuid.New())

	if mock.GetOrchestrationStatus() != OrchestrationStatusCOMPLETED {
		t.Errorf("expected COMPLETED, got %s", mock.GetOrchestrationStatus())
	}
}

// TestScheduleOrchestratorStatusFailureTransition verifies status changes on error
func TestScheduleOrchestratorStatusFailureTransition(t *testing.T) {
	mock := NewMockScheduleOrchestrator()
	mock.ExecuteImportFunc = func(
		ctx context.Context,
		filePath string,
		hospitalID uuid.UUID,
		userID uuid.UUID,
	) (*OrchestrationResult, error) {
		return nil, errors.New("import failed")
	}

	mock.ExecuteImport(context.Background(), "/tmp/test.ods", uuid.New(), uuid.New())

	if mock.GetOrchestrationStatus() != OrchestrationStatusFAILED {
		t.Errorf("expected FAILED, got %s", mock.GetOrchestrationStatus())
	}
}

// TestScheduleOrchestratorDefaultBehavior verifies default mock returns valid results
func TestScheduleOrchestratorDefaultBehavior(t *testing.T) {
	mock := NewMockScheduleOrchestrator()
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	result, err := mock.ExecuteImport(ctx, "/tmp/test.ods", hospitalID, userID)

	if err != nil {
		t.Fatalf("default should return no error, got %v", err)
	}
	if result == nil {
		t.Fatal("default should return OrchestrationResult")
	}
	if result.ScheduleVersion == nil {
		t.Fatal("default result should have ScheduleVersion")
	}
	if result.ScheduleVersion.HospitalID != hospitalID {
		t.Errorf("HospitalID should match input")
	}
	if result.ScheduleVersion.CreatedBy != userID {
		t.Errorf("CreatedBy should match input")
	}
	if result.ValidationResult == nil {
		t.Fatal("default result should have ValidationResult")
	}
	if result.Coverage == nil {
		t.Fatal("default result should have Coverage")
	}
	if result.Duration == 0 {
		t.Fatal("default result should have Duration")
	}
	if result.CompletedAt.IsZero() {
		t.Fatal("default result should have CompletedAt")
	}
}

// TestInterfaceImplementationWithDifferentMocks verifies any implementation works
func TestInterfaceImplementationWithDifferentMocks(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	// Test with different mock services
	var ods ODSImportService = &MockODSImportService{}
	var amion AmionScraperService = &MockAmionScraperService{}
	var coverage CoverageCalculatorService = &MockCoverageCalculatorService{}
	var orchestrator ScheduleOrchestrator = NewMockScheduleOrchestrator()

	// All should be usable as their interface types
	_, _, err := ods.ImportSchedule(ctx, "/tmp/test.ods", hospitalID, userID)
	if err != nil {
		t.Fatalf("ODS interface call failed: %v", err)
	}

	_, _, err = amion.ScrapeSchedule(ctx, time.Now(), 1, hospitalID, userID)
	if err != nil {
		t.Fatalf("Amion interface call failed: %v", err)
	}

	_, err = coverage.Calculate(ctx, uuid.New())
	if err != nil {
		t.Fatalf("Coverage interface call failed: %v", err)
	}

	_, err = orchestrator.ExecuteImport(ctx, "/tmp/test.ods", hospitalID, userID)
	if err != nil {
		t.Fatalf("Orchestrator interface call failed: %v", err)
	}

	status := orchestrator.GetOrchestrationStatus()
	if status == "" {
		t.Fatal("Status should not be empty")
	}
}

// BenchmarkMockODSImportService benchmarks mock ODS importer
func BenchmarkMockODSImportService(b *testing.B) {
	mock := &MockODSImportService{}
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ImportSchedule(ctx, "/tmp/test.ods", hospitalID, userID)
	}
}

// BenchmarkMockAmionScraperService benchmarks mock Amion scraper
func BenchmarkMockAmionScraperService(b *testing.B) {
	mock := &MockAmionScraperService{}
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ScrapeSchedule(ctx, time.Now(), 3, hospitalID, userID)
	}
}

// BenchmarkMockCoverageCalculatorService benchmarks mock coverage calculator
func BenchmarkMockCoverageCalculatorService(b *testing.B) {
	mock := &MockCoverageCalculatorService{}
	ctx := context.Background()
	scheduleVersionID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.Calculate(ctx, scheduleVersionID)
	}
}

// BenchmarkMockScheduleOrchestrator benchmarks mock orchestrator
func BenchmarkMockScheduleOrchestrator(b *testing.B) {
	mock := NewMockScheduleOrchestrator()
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExecuteImport(ctx, "/tmp/test.ods", hospitalID, userID)
	}
}
