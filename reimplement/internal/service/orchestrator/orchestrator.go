package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/validation"
	"go.uber.org/zap"
)

// DefaultScheduleOrchestrator implements the ScheduleOrchestrator interface.
// It coordinates ODS import, Amion scraping, and coverage calculation with
// comprehensive error handling and transaction management.
type DefaultScheduleOrchestrator struct {
	// Service dependencies
	odsService      ODSImportService
	amionService    AmionScraperService
	coverageService CoverageCalculatorService
	logger          *zap.SugaredLogger

	// Status tracking (atomic)
	status atomic.Value // stores OrchestrationStatus
	mu     sync.RWMutex // protects status access
}

// NewDefaultScheduleOrchestrator creates a new orchestrator with the given service dependencies.
// All services are required (non-nil).
func NewDefaultScheduleOrchestrator(
	odsService ODSImportService,
	amionService AmionScraperService,
	coverageService CoverageCalculatorService,
	logger *zap.SugaredLogger,
) *DefaultScheduleOrchestrator {
	orch := &DefaultScheduleOrchestrator{
		odsService:      odsService,
		amionService:    amionService,
		coverageService: coverageService,
		logger:          logger,
	}
	orch.status.Store(OrchestrationStatusIDLE)
	return orch
}

// ExecuteImport performs a complete schedule import and coverage calculation workflow.
// It implements the 3-phase workflow:
//
// Phase 1 - ODS Import (2-4 hours):
//   - Call ODSImportService.ImportSchedule()
//   - Create ScheduleVersion
//   - Collect any validation errors/warnings
//   - CRITICAL ERROR → stop, return error
//   - PARTIAL SUCCESS → continue to Phase 2
//
// Phase 2 - Amion Scraping (2-3 seconds):
//   - Call AmionScraperService.ScrapeSchedule()
//   - Create Assignments linked to ScheduleVersion
//   - Collect scraping errors/warnings
//   - ERROR → skip phase, continue to Phase 3 (non-critical)
//
// Phase 3 - Coverage Calculation (1 second):
//   - Call CoverageCalculatorService.Calculate()
//   - Generate CoverageMetrics for ScheduleVersion
//   - Collect any calculation errors
//   - ERROR → log but don't fail (coverage can be recalculated)
//
// Each phase is in a separate transaction. Partial success is a valid outcome.
func (orch *DefaultScheduleOrchestrator) ExecuteImport(
	ctx context.Context,
	filePath string,
	hospitalID uuid.UUID,
	userID uuid.UUID,
) (*OrchestrationResult, error) {
	// Update status to IN_PROGRESS
	orch.setStatus(OrchestrationStatusINPROGRESS)
	startTime := time.Now()

	// Initialize result
	result := &OrchestrationResult{
		Assignments:      make([]entity.Assignment, 0),
		ValidationResult: validation.NewValidationResult(),
		Metadata:         make(map[string]interface{}),
	}

	// Validate inputs
	if hospitalID == uuid.Nil {
		orch.setStatus(OrchestrationStatusFAILED)
		return nil, fmt.Errorf("invalid hospital ID: nil UUID")
	}

	if userID == uuid.Nil {
		orch.setStatus(OrchestrationStatusFAILED)
		return nil, fmt.Errorf("invalid user ID: nil UUID")
	}

	orch.logger.Debugw("orchestration started",
		"filePath", filePath,
		"hospitalID", hospitalID.String(),
		"userID", userID.String(),
	)

	// PHASE 1: ODS Import
	orch.logger.Infow("phase 1: ODS import starting")
	scheduleVersion, odsValidation, odsErr := orch.odsService.ImportSchedule(ctx, filePath, hospitalID, userID)

	// Merge ODS validation results
	if odsValidation != nil {
		for _, err := range odsValidation.Errors {
			result.ValidationResult.AddError(fmt.Sprintf("ods:%s", err.Field), err.Message)
		}
		for _, warn := range odsValidation.Warnings {
			result.ValidationResult.AddWarning(fmt.Sprintf("ods:%s", warn.Field), warn.Message)
		}
		for _, info := range odsValidation.Infos {
			result.ValidationResult.AddInfo(fmt.Sprintf("ods:%s", info.Field), info.Message)
		}
	}

	// Handle Phase 1 critical error (no schedule version created)
	if odsErr != nil && scheduleVersion == nil {
		orch.logger.Errorw("phase 1: critical error during ODS import",
			"error", odsErr,
		)
		orch.setStatus(OrchestrationStatusFAILED)
		return nil, fmt.Errorf("phase 1 ODS import failed: %w", odsErr)
	}

	// Handle Phase 1 partial success (schedule version created despite errors)
	if odsErr != nil && scheduleVersion != nil {
		orch.logger.Warnw("phase 1: partial success - schedule created despite some shift failures",
			"error", odsErr,
			"scheduleVersionID", scheduleVersion.ID.String(),
		)
		result.ValidationResult.AddWarning("ods_import", "Some shifts failed to import")
	}

	// Phase 1 success - schedule version created
	if scheduleVersion != nil {
		orch.logger.Infow("phase 1: ODS import succeeded",
			"scheduleVersionID", scheduleVersion.ID.String(),
			"version", scheduleVersion.Version,
		)
		result.ScheduleVersion = scheduleVersion
		result.ValidationResult.AddInfo("ods_import", "Schedule imported successfully")
	}

	// PHASE 2: Amion Scraping (non-critical - errors don't fail the operation)
	var amionAssignments []entity.Assignment
	var amionValidation *validation.ValidationResult
	var amionErr error

	if scheduleVersion != nil {
		orch.logger.Infow("phase 2: Amion scraping starting")
		startDate := scheduleVersion.StartDate
		monthCount := int(scheduleVersion.EndDate.Sub(scheduleVersion.StartDate).Hours() / 24 / 30)
		if monthCount < 1 {
			monthCount = 1
		}
		amionAssignments, amionValidation, amionErr = orch.amionService.ScrapeSchedule(ctx, startDate, monthCount, hospitalID, userID)
	} else {
		orch.logger.Warnw("phase 2: skipping Amion scraping - no schedule version created")
	}

	// Merge Amion validation results
	if amionValidation != nil {
		for _, err := range amionValidation.Errors {
			result.ValidationResult.AddError(fmt.Sprintf("amion:%s", err.Field), err.Message)
		}
		for _, warn := range amionValidation.Warnings {
			result.ValidationResult.AddWarning(fmt.Sprintf("amion:%s", warn.Field), warn.Message)
		}
		for _, info := range amionValidation.Infos {
			result.ValidationResult.AddInfo(fmt.Sprintf("amion:%s", info.Field), info.Message)
		}
	}

	// Handle Phase 2 error (non-critical)
	if amionErr != nil {
		orch.logger.Warnw("phase 2: Amion scraping error (continuing to phase 3)",
			"error", amionErr,
		)
		result.ValidationResult.AddWarning("amion_scrape", "Amion scraping encountered errors (continuing)")
	} else if len(amionAssignments) > 0 {
		orch.logger.Infow("phase 2: Amion scraping succeeded",
			"assignmentCount", len(amionAssignments),
		)
		result.Assignments = amionAssignments
		result.ValidationResult.AddInfo("amion_scrape", fmt.Sprintf("Scraped %d assignments", len(amionAssignments)))
	}

	// PHASE 3: Coverage Calculation (non-critical - errors don't fail the operation)
	if scheduleVersion != nil {
		orch.logger.Infow("phase 3: coverage calculation starting",
			"scheduleVersionID", scheduleVersion.ID.String(),
		)

		coverage, coverageErr := orch.coverageService.Calculate(ctx, scheduleVersion.ID)

		// Handle Phase 3 error (non-critical)
		if coverageErr != nil {
			orch.logger.Warnw("phase 3: coverage calculation error (continuing)",
				"error", coverageErr,
				"scheduleVersionID", scheduleVersion.ID.String(),
			)
			result.ValidationResult.AddWarning("coverage_calc", "Coverage calculation failed (can be recalculated later)")
		} else if coverage != nil {
			orch.logger.Infow("phase 3: coverage calculation succeeded",
				"scheduleVersionID", scheduleVersion.ID.String(),
				"coveragePercentage", coverage.CoveragePercentage,
			)
			result.Coverage = coverage
			result.ValidationResult.AddInfo("coverage_calc", fmt.Sprintf("Coverage: %.1f%%", coverage.CoveragePercentage))
		}
	}

	// Build final result
	duration := time.Since(startTime)
	result.Duration = duration
	result.CompletedAt = time.Now()
	result.Metadata["import_source"] = "ods_file"
	result.Metadata["hospital_id"] = hospitalID.String()
	result.Metadata["file_path"] = filePath
	result.Metadata["phases_completed"] = []string{"ods_import", "amion_scrape", "coverage_calc"}

	// Update status to COMPLETED
	orch.setStatus(OrchestrationStatusCOMPLETED)

	orch.logger.Infow("orchestration completed successfully",
		"duration", duration,
		"validationErrors", result.ValidationResult.ErrorCount(),
		"validationWarnings", result.ValidationResult.WarningCount(),
	)

	return result, nil
}

// GetOrchestrationStatus returns the current status of the orchestrator.
// This can be used to monitor long-running operations or detect stuck processes.
func (orch *DefaultScheduleOrchestrator) GetOrchestrationStatus() OrchestrationStatus {
	return orch.getStatus()
}

// setStatus updates the orchestration status atomically.
func (orch *DefaultScheduleOrchestrator) setStatus(status OrchestrationStatus) {
	orch.mu.Lock()
	defer orch.mu.Unlock()
	orch.status.Store(status)
}

// getStatus retrieves the current orchestration status atomically.
func (orch *DefaultScheduleOrchestrator) getStatus() OrchestrationStatus {
	orch.mu.RLock()
	defer orch.mu.RUnlock()
	return orch.status.Load().(OrchestrationStatus)
}

// mergeValidationResults combines multiple validation results into one.
// This is used internally to aggregate validation messages from all phases.
func mergeValidationResults(results ...*validation.ValidationResult) *validation.ValidationResult {
	merged := validation.NewValidationResult()

	for _, result := range results {
		if result == nil {
			continue
		}

		for _, err := range result.Errors {
			merged.AddError(err.Field, err.Message)
		}

		for _, warn := range result.Warnings {
			merged.AddWarning(warn.Field, warn.Message)
		}

		for _, info := range result.Infos {
			merged.AddInfo(info.Field, info.Message)
		}

		for key, val := range result.Context {
			merged.SetContext(key, val)
		}
	}

	return merged
}
