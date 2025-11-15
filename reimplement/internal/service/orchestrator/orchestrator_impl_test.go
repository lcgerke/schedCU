package orchestrator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestExecuteImportSuccessfulFullWorkflow tests successful execution of all three phases.
func TestExecuteImportSuccessfulFullWorkflow(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	mockODS := &MockODSImportService{
		ImportScheduleFunc: func(ctx context.Context, filePath string, hospitalID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			sv := &entity.ScheduleVersion{
				ID:         uuid.New(),
				HospitalID: hospitalID,
				Version:    1,
				Status:     entity.VersionStatusDraft,
				StartDate:  time.Now(),
				EndDate:    time.Now().AddDate(0, 1, 0),
				Source:     "ods_file",
				CreatedBy:  userID,
				CreatedAt:  time.Now(),
			}
			return sv, validation.NewValidationResult(), nil
		},
	}

	mockAmion := &MockAmionScraperService{
		ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hospitalID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
			return []entity.Assignment{}, validation.NewValidationResult(), nil
		},
	}

	mockCoverage := &MockCoverageCalculatorService{
		CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
			return &CoverageMetrics{
				ScheduleVersionID:  scheduleVersionID,
				CoveragePercentage: 100.0,
				AssignedPositions:  10,
				RequiredPositions:  10,
				CalculatedAt:       time.Now(),
			}, nil
		},
	}

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
	result, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.ScheduleVersion)
	assert.Equal(t, hospitalID, result.ScheduleVersion.HospitalID)
	assert.NotNil(t, result.Coverage)
	assert.Equal(t, 100.0, result.Coverage.CoveragePercentage)
	assert.NotNil(t, result.ValidationResult)
	assert.False(t, result.ValidationResult.HasErrors())
	assert.Equal(t, OrchestrationStatusCOMPLETED, orchestrator.GetOrchestrationStatus())
}

// TestExecuteImportPhase1CriticalError tests that Phase 1 critical errors stop execution.
func TestExecuteImportPhase1CriticalError(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	mockODS := &MockODSImportService{
		ImportScheduleFunc: func(ctx context.Context, filePath string, hospitalID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			return nil, validation.NewValidationResult(), fmt.Errorf("parse error")
		},
	}

	mockAmion := &MockAmionScraperService{}
	mockCoverage := &MockCoverageCalculatorService{}

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
	result, err := orchestrator.ExecuteImport(ctx, "/path/to/invalid.ods", hospitalID, userID)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 0, len(mockAmion.ScrapeScheduleCalls))
	assert.Equal(t, 0, len(mockCoverage.CalculateCalls))
	assert.Equal(t, OrchestrationStatusFAILED, orchestrator.GetOrchestrationStatus())
}

// TestExecuteImportPhase2ErrorContinuesToPhase3 tests that Phase 2 errors don't stop Phase 3.
func TestExecuteImportPhase2ErrorContinuesToPhase3(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	mockODS := &MockODSImportService{
		ImportScheduleFunc: func(ctx context.Context, filePath string, hospitalID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			sv := &entity.ScheduleVersion{
				ID:         uuid.New(),
				HospitalID: hospitalID,
				Version:    1,
				Status:     entity.VersionStatusDraft,
				StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:    time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
				Source:     "ods_file",
				CreatedBy:  userID,
				CreatedAt:  time.Now(),
			}
			return sv, validation.NewValidationResult(), nil
		},
	}

	mockAmion := &MockAmionScraperService{
		ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hospitalID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
			return nil, validation.NewValidationResult(), fmt.Errorf("network error")
		},
	}

	phase3Called := false
	mockCoverage := &MockCoverageCalculatorService{
		CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
			phase3Called = true
			return &CoverageMetrics{
				ScheduleVersionID:  scheduleVersionID,
				CoveragePercentage: 50.0,
				CalculatedAt:       time.Now(),
			}, nil
		},
	}

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
	result, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, phase3Called)
	assert.NotNil(t, result.Coverage)
	assert.True(t, result.ValidationResult.HasWarnings())
}

// TestExecuteImportPhase3ErrorDoesNotFail tests that Phase 3 errors don't fail the operation.
func TestExecuteImportPhase3ErrorDoesNotFail(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	mockODS := &MockODSImportService{
		ImportScheduleFunc: func(ctx context.Context, filePath string, hospitalID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			sv := &entity.ScheduleVersion{
				ID:         uuid.New(),
				HospitalID: hospitalID,
				Version:    1,
				Status:     entity.VersionStatusDraft,
				Source:     "ods_file",
				CreatedBy:  userID,
				CreatedAt:  time.Now(),
			}
			return sv, validation.NewValidationResult(), nil
		},
	}

	mockAmion := &MockAmionScraperService{
		ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hospitalID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
			return []entity.Assignment{}, validation.NewValidationResult(), nil
		},
	}

	mockCoverage := &MockCoverageCalculatorService{
		CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
			return nil, fmt.Errorf("database connection lost")
		},
	}

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
	result, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.ScheduleVersion)
	assert.Nil(t, result.Coverage)
	assert.Equal(t, OrchestrationStatusCOMPLETED, orchestrator.GetOrchestrationStatus())
}

// TestGetOrchestrationStatusInitiallyIDLE tests status starts as IDLE.
func TestGetOrchestrationStatusInitiallyIDLE(t *testing.T) {
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	orchestrator := NewDefaultScheduleOrchestrator(&MockODSImportService{}, &MockAmionScraperService{}, &MockCoverageCalculatorService{}, logger)
	assert.Equal(t, OrchestrationStatusIDLE, orchestrator.GetOrchestrationStatus())
}

// TestExecuteImportInvalidInputs tests invalid inputs are rejected.
func TestExecuteImportInvalidInputs(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	orchestrator := NewDefaultScheduleOrchestrator(&MockODSImportService{}, &MockAmionScraperService{}, &MockCoverageCalculatorService{}, logger)

	result, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", uuid.Nil, userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, OrchestrationStatusFAILED, orchestrator.GetOrchestrationStatus())

	result, err = orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, uuid.Nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestExecuteImportOrchestrationResult tests result metadata and timing.
func TestExecuteImportOrchestrationResult(t *testing.T) {
	ctx := context.Background()
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	mockODS := &MockODSImportService{
		ImportScheduleFunc: func(ctx context.Context, filePath string, hospitalID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			sv := &entity.ScheduleVersion{
				ID:         uuid.New(),
				HospitalID: hospitalID,
				Version:    1,
				Status:     entity.VersionStatusDraft,
				Source:     "ods_file",
				CreatedBy:  userID,
				CreatedAt:  time.Now(),
			}
			return sv, validation.NewValidationResult(), nil
		},
	}

	mockAmion := &MockAmionScraperService{
		ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hospitalID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
			return []entity.Assignment{}, validation.NewValidationResult(), nil
		},
	}

	mockCoverage := &MockCoverageCalculatorService{
		CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
			return &CoverageMetrics{
				ScheduleVersionID:  scheduleVersionID,
				CoveragePercentage: 100.0,
				CalculatedAt:       time.Now(),
			}, nil
		},
	}

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
	startTime := time.Now()
	result, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)
	endTime := time.Now()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotNil(t, result.Metadata)
	assert.Contains(t, result.Metadata, "import_source")
	assert.Contains(t, result.Metadata, "hospital_id")
	assert.GreaterOrEqual(t, result.Duration, time.Duration(0))
	assert.Greater(t, result.CompletedAt, startTime)
	assert.Less(t, result.CompletedAt, endTime)
}
