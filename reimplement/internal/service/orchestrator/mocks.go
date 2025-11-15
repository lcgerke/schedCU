// Package orchestrator provides interfaces for coordinating schedule import and calculation workflows.
package orchestrator

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/validation"
)

// MockODSImportService is a test mock for ODSImportService.
// It allows tests to control what values are returned.
type MockODSImportService struct {
	// ImportScheduleFunc allows tests to override ImportSchedule behavior
	ImportScheduleFunc func(
		ctx context.Context,
		filePath string,
		hospitalID uuid.UUID,
		userID uuid.UUID,
	) (*entity.ScheduleVersion, *validation.ValidationResult, error)

	// ImportScheduleCalls tracks all calls to ImportSchedule
	ImportScheduleCalls []MockODSImportServiceCall
}

// MockODSImportServiceCall contains details about a single ImportSchedule call
type MockODSImportServiceCall struct {
	FilePath   string
	HospitalID uuid.UUID
	UserID     uuid.UUID
}

// ImportSchedule implements ODSImportService.ImportSchedule
func (m *MockODSImportService) ImportSchedule(
	ctx context.Context,
	filePath string,
	hospitalID uuid.UUID,
	userID uuid.UUID,
) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
	m.ImportScheduleCalls = append(m.ImportScheduleCalls, MockODSImportServiceCall{
		FilePath:   filePath,
		HospitalID: hospitalID,
		UserID:     userID,
	})

	if m.ImportScheduleFunc != nil {
		return m.ImportScheduleFunc(ctx, filePath, hospitalID, userID)
	}

	// Default behavior: return a valid schedule version
	return &entity.ScheduleVersion{
		ID:         uuid.New(),
		HospitalID: hospitalID,
		Version:    1,
		Status:     entity.VersionStatusDraft,
		StartDate:  time.Now(),
		EndDate:    time.Now().AddDate(0, 1, 0),
		Source:     "ods_file",
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
		CreatedBy:  userID,
		UpdatedAt:  time.Now(),
		UpdatedBy:  userID,
	}, validation.NewValidationResult(), nil
}

// MockAmionScraperService is a test mock for AmionScraperService.
// It allows tests to control what values are returned.
type MockAmionScraperService struct {
	// ScrapeScheduleFunc allows tests to override ScrapeSchedule behavior
	ScrapeScheduleFunc func(
		ctx context.Context,
		startDate time.Time,
		monthCount int,
		hospitalID uuid.UUID,
		userID uuid.UUID,
	) ([]entity.Assignment, *validation.ValidationResult, error)

	// ScrapeScheduleCalls tracks all calls to ScrapeSchedule
	ScrapeScheduleCalls []MockAmionScraperServiceCall
}

// MockAmionScraperServiceCall contains details about a single ScrapeSchedule call
type MockAmionScraperServiceCall struct {
	StartDate  time.Time
	MonthCount int
	HospitalID uuid.UUID
	UserID     uuid.UUID
}

// ScrapeSchedule implements AmionScraperService.ScrapeSchedule
func (m *MockAmionScraperService) ScrapeSchedule(
	ctx context.Context,
	startDate time.Time,
	monthCount int,
	hospitalID uuid.UUID,
	userID uuid.UUID,
) ([]entity.Assignment, *validation.ValidationResult, error) {
	m.ScrapeScheduleCalls = append(m.ScrapeScheduleCalls, MockAmionScraperServiceCall{
		StartDate:  startDate,
		MonthCount: monthCount,
		HospitalID: hospitalID,
		UserID:     userID,
	})

	if m.ScrapeScheduleFunc != nil {
		return m.ScrapeScheduleFunc(ctx, startDate, monthCount, hospitalID, userID)
	}

	// Default behavior: return empty assignments list
	return []entity.Assignment{}, validation.NewValidationResult(), nil
}

// MockCoverageCalculatorService is a test mock for CoverageCalculatorService.
// It allows tests to control what values are returned.
type MockCoverageCalculatorService struct {
	// CalculateFunc allows tests to override Calculate behavior
	CalculateFunc func(
		ctx context.Context,
		scheduleVersionID uuid.UUID,
	) (*CoverageMetrics, error)

	// CalculateCalls tracks all calls to Calculate
	CalculateCalls []MockCoverageCalculatorServiceCall
}

// MockCoverageCalculatorServiceCall contains details about a single Calculate call
type MockCoverageCalculatorServiceCall struct {
	ScheduleVersionID uuid.UUID
}

// Calculate implements CoverageCalculatorService.Calculate
func (m *MockCoverageCalculatorService) Calculate(
	ctx context.Context,
	scheduleVersionID uuid.UUID,
) (*CoverageMetrics, error) {
	m.CalculateCalls = append(m.CalculateCalls, MockCoverageCalculatorServiceCall{
		ScheduleVersionID: scheduleVersionID,
	})

	if m.CalculateFunc != nil {
		return m.CalculateFunc(ctx, scheduleVersionID)
	}

	// Default behavior: return valid coverage metrics
	return &CoverageMetrics{
		ScheduleVersionID:   scheduleVersionID,
		CoveragePercentage:  100.0,
		AssignedPositions:   10,
		RequiredPositions:   10,
		UncoveredShifts:     []*entity.ShiftInstance{},
		OverallocatedShifts: []*entity.ShiftInstance{},
		CalculatedAt:        time.Now(),
		Details:             make(map[string]interface{}),
	}, nil
}

// MockScheduleOrchestrator is a test mock for ScheduleOrchestrator.
// It allows tests to control what values are returned.
type MockScheduleOrchestrator struct {
	// ExecuteImportFunc allows tests to override ExecuteImport behavior
	ExecuteImportFunc func(
		ctx context.Context,
		filePath string,
		hospitalID uuid.UUID,
		userID uuid.UUID,
	) (*OrchestrationResult, error)

	// ExecuteImportCalls tracks all calls to ExecuteImport
	ExecuteImportCalls []MockScheduleOrchestratorCall

	// currentStatus tracks the status of the orchestrator
	currentStatus OrchestrationStatus
	statusMutex  chan struct{} // Use channel for simple synchronization in tests
}

// MockScheduleOrchestratorCall contains details about a single ExecuteImport call
type MockScheduleOrchestratorCall struct {
	FilePath   string
	HospitalID uuid.UUID
	UserID     uuid.UUID
}

// NewMockScheduleOrchestrator creates a new mock orchestrator
func NewMockScheduleOrchestrator() *MockScheduleOrchestrator {
	return &MockScheduleOrchestrator{
		currentStatus: OrchestrationStatusIDLE,
		statusMutex:   make(chan struct{}, 1),
	}
}

// ExecuteImport implements ScheduleOrchestrator.ExecuteImport
func (m *MockScheduleOrchestrator) ExecuteImport(
	ctx context.Context,
	filePath string,
	hospitalID uuid.UUID,
	userID uuid.UUID,
) (*OrchestrationResult, error) {
	m.ExecuteImportCalls = append(m.ExecuteImportCalls, MockScheduleOrchestratorCall{
		FilePath:   filePath,
		HospitalID: hospitalID,
		UserID:     userID,
	})

	// Update status to IN_PROGRESS
	m.statusMutex <- struct{}{}
	m.currentStatus = OrchestrationStatusINPROGRESS
	<-m.statusMutex

	if m.ExecuteImportFunc != nil {
		result, err := m.ExecuteImportFunc(ctx, filePath, hospitalID, userID)

		// Update status based on result
		m.statusMutex <- struct{}{}
		if err != nil {
			m.currentStatus = OrchestrationStatusFAILED
		} else if result != nil {
			m.currentStatus = OrchestrationStatusCOMPLETED
		}
		<-m.statusMutex

		return result, err
	}

	// Default behavior: return a valid orchestration result
	now := time.Now()
	result := &OrchestrationResult{
		ScheduleVersion: &entity.ScheduleVersion{
			ID:         uuid.New(),
			HospitalID: hospitalID,
			Version:    1,
			Status:     entity.VersionStatusDraft,
			StartDate:  now,
			EndDate:    now.AddDate(0, 1, 0),
			Source:     "ods_file",
			Metadata:   make(map[string]interface{}),
			CreatedAt:  now,
			CreatedBy:  userID,
			UpdatedAt:  now,
			UpdatedBy:  userID,
		},
		Assignments: []entity.Assignment{},
		Coverage: &CoverageMetrics{
			ScheduleVersionID:   uuid.Nil,
			CoveragePercentage:  100.0,
			AssignedPositions:   0,
			RequiredPositions:   0,
			UncoveredShifts:     []*entity.ShiftInstance{},
			OverallocatedShifts: []*entity.ShiftInstance{},
			CalculatedAt:        now,
			Details:             make(map[string]interface{}),
		},
		ValidationResult: validation.NewValidationResult(),
		Duration:         10 * time.Millisecond,
		CompletedAt:      now,
		Metadata:         make(map[string]interface{}),
	}

	// Update status to COMPLETED
	m.statusMutex <- struct{}{}
	m.currentStatus = OrchestrationStatusCOMPLETED
	<-m.statusMutex

	return result, nil
}

// GetOrchestrationStatus implements ScheduleOrchestrator.GetOrchestrationStatus
func (m *MockScheduleOrchestrator) GetOrchestrationStatus() OrchestrationStatus {
	m.statusMutex <- struct{}{}
	defer func() { <-m.statusMutex }()
	return m.currentStatus
}

// SetStatus allows tests to manually set the orchestrator status
func (m *MockScheduleOrchestrator) SetStatus(status OrchestrationStatus) {
	m.statusMutex <- struct{}{}
	defer func() { <-m.statusMutex }()
	m.currentStatus = status
}
