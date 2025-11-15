// Package orchestrator provides interfaces for coordinating schedule import and calculation workflows.
package orchestrator

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/validation"
)

// ODSImportService defines the interface for importing ODS schedule files.
// Implementations must handle file parsing, validation, and persistence with
// comprehensive error reporting and validation result collection.
//
// Concurrency: Implementations must be thread-safe. Multiple imports can be
// executed concurrently as long as they target different hospitals or schedule versions.
//
// Transaction requirements: The entire import operation should be atomic at the
// ScheduleVersion level - either all shifts are imported successfully or the
// operation fails and the database state remains consistent.
//
// Example usage:
//
//	svc, _ := /* obtain ODS importer service */
//	schedVersion, valResult, err := svc.ImportSchedule(ctx, "/path/to/file.ods", hospitalID, userID)
//	if err != nil {
//	    log.Printf("Import failed: %v", err)
//	    if valResult != nil && valResult.HasErrors() {
//	        log.Printf("Validation errors: %+v", valResult.Errors)
//	    }
//	}
type ODSImportService interface {
	// ImportSchedule imports a schedule from an ODS file and persists it to the database.
	//
	// Parameters:
	//   - ctx: context.Context for cancellation and timeout support
	//   - filePath: absolute path to the ODS file to import
	//   - hospitalID: UUID of the hospital this schedule belongs to (required, non-nil)
	//   - userID: UUID of the user performing the import for audit trails (required, non-nil)
	//
	// Returns:
	//   - *entity.ScheduleVersion: the created schedule version with draft status
	//   - *validation.ValidationResult: validation errors, warnings, and info messages
	//   - error: if file cannot be read, parsing fails, validation shows critical errors,
	//            or database operations fail
	//
	// Error handling:
	//   - File not found: returns os.IsNotExist() compatible error
	//   - Invalid file format: returns validation error with INVALID_FILE_FORMAT code
	//   - Parse errors: collects all parse errors and returns PARSE_ERROR code
	//   - Validation failures: returns the ValidationResult with HasErrors() = true
	//   - Database errors: returns DATABASE_ERROR code in validation result
	//
	// Guarantees:
	//   - If error is returned, ScheduleVersion is nil
	//   - If error is nil, ScheduleVersion is always populated (status = draft)
	//   - ValidationResult is always populated (non-nil)
	//   - Soft deletes are handled correctly (no conflicts with previous versions)
	//
	// Example error handling:
	//
	//	schedVersion, valResult, err := svc.ImportSchedule(ctx, filePath, hospitalID, userID)
	//	if err != nil {
	//	    if valResult != nil && valResult.HasErrors() {
	//	        for _, errMsg := range valResult.Errors {
	//	            log.Printf("Field %s: %s", errMsg.Field, errMsg.Message)
	//	        }
	//	    }
	//	    return err
	//	}
	ImportSchedule(
		ctx context.Context,
		filePath string,
		hospitalID uuid.UUID,
		userID uuid.UUID,
	) (*entity.ScheduleVersion, *validation.ValidationResult, error)
}

// AmionScraperService defines the interface for scraping schedule data from Amion.
// Implementations must handle concurrent scraping with rate limiting, validation,
// and comprehensive error reporting.
//
// Concurrency: Implementations should support concurrent scraping of multiple months
// with configurable concurrency limits to avoid overwhelming external services.
//
// Rate limiting: Must implement intelligent rate limiting to respect Amion's
// service limits and avoid IP blocking.
//
// Example usage:
//
//	svc, _ := /* obtain Amion scraper service */
//	assignments, valResult, err := svc.ScrapeSchedule(ctx, startDate, 3, hospitalID, userID)
//	if err != nil {
//	    log.Printf("Scrape failed: %v", err)
//	    log.Printf("Warnings: %+v", valResult.Warnings)
//	}
type AmionScraperService interface {
	// ScrapeSchedule scrapes schedule assignments from Amion for the given date range.
	//
	// Parameters:
	//   - ctx: context.Context for cancellation and timeout support
	//   - startDate: the start date for scraping (typically first day of month)
	//   - monthCount: number of months to scrape (e.g., 3 for 3 months)
	//   - hospitalID: UUID of the hospital to scrape schedules for (optional, may be used for routing)
	//   - userID: UUID of the user performing the scrape for audit trails
	//
	// Returns:
	//   - []entity.Assignment: scraped assignments with source = AMION
	//   - *validation.ValidationResult: validation errors, warnings, and info messages
	//   - error: if scraping fails critically (network error, auth failure, etc.)
	//
	// Error handling:
	//   - Network errors: returns error, ValidationResult may contain partial results
	//   - Authentication failures: returns EXTERNAL_SERVICE_ERROR
	//   - Parse errors: returns PARSE_ERROR code, may include partial results
	//   - Rate limiting: may retry automatically before returning error
	//   - Context cancellation: returns context.Err() wrapped with context
	//
	// Guarantees:
	//   - Each assignment has non-nil PersonID and ShiftInstanceID
	//   - Each assignment has source = AMION
	//   - Duplicate assignments are detected and reported in warnings
	//   - All returned assignments are created in a single batch operation
	//   - If error is returned, some or all assignments may be persisted
	//
	// Example:
	//
	//	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	//	assignments, valResult, err := svc.ScrapeSchedule(
	//	    ctx, startDate, 3, hospitalID, userID,
	//	)
	//	if err != nil {
	//	    log.Printf("Scrape error: %v", err)
	//	}
	//	log.Printf("Scraped %d assignments", len(assignments))
	//	if valResult.HasWarnings() {
	//	    for _, warn := range valResult.Warnings {
	//	        log.Printf("Warning: %s", warn.Message)
	//	    }
	//	}
	ScrapeSchedule(
		ctx context.Context,
		startDate time.Time,
		monthCount int,
		hospitalID uuid.UUID,
		userID uuid.UUID,
	) ([]entity.Assignment, *validation.ValidationResult, error)
}

// CoverageMetrics contains calculated coverage metrics for a schedule.
// This is the output of coverage calculation operations.
type CoverageMetrics struct {
	// ScheduleVersionID identifies which schedule version these metrics apply to
	ScheduleVersionID uuid.UUID

	// CoveragePercentage is the overall coverage as a percentage (0-100)
	// Calculated as: (assigned_positions / required_positions) * 100
	CoveragePercentage float64

	// AssignedPositions is the total number of positions that have been assigned
	AssignedPositions int

	// RequiredPositions is the total number of positions that need to be filled
	RequiredPositions int

	// UncoveredShifts contains shift instances that are not covered
	UncoveredShifts []*entity.ShiftInstance

	// OverallocatedShifts contains shift instances with more than one assignment
	OverallocatedShifts []*entity.ShiftInstance

	// CalculatedAt is when these metrics were computed
	CalculatedAt time.Time

	// Details contains arbitrary metrics for different shift types, locations, etc.
	// Keys might be "morning_coverage", "afternoon_coverage", "night_coverage", etc.
	Details map[string]interface{}
}

// CoverageCalculatorService defines the interface for calculating schedule coverage metrics.
// Implementations must load assignment data efficiently (single query pattern) and
// compute coverage metrics deterministically.
//
// Concurrency: Implementations must be thread-safe. Multiple calculations can be
// executed concurrently for different schedules.
//
// Query efficiency: Must use single batch query pattern to load assignments.
// Must not use N+1 query patterns.
//
// Example usage:
//
//	svc, _ := /* obtain coverage calculator service */
//	metrics, err := svc.Calculate(ctx, scheduleVersionID)
//	if err != nil {
//	    log.Printf("Coverage calculation failed: %v", err)
//	    return
//	}
//	log.Printf("Coverage: %.1f%%", metrics.CoveragePercentage)
type CoverageCalculatorService interface {
	// Calculate computes coverage metrics for a given schedule version.
	//
	// Parameters:
	//   - ctx: context.Context for cancellation and timeout support
	//   - scheduleVersionID: the schedule version to calculate coverage for (required, non-nil)
	//
	// Returns:
	//   - *CoverageMetrics: computed coverage metrics (never nil on success)
	//   - error: if schedule version is invalid, data loading fails, or calculation fails
	//
	// Error handling:
	//   - Invalid schedule version ID: returns ErrInvalidScheduleVersion
	//   - Repository fails: returns DATABASE_ERROR
	//   - Context cancellation: returns context.Err()
	//
	// Guarantees:
	//   - If error is returned, CoverageMetrics is nil
	//   - If error is nil, all metrics are populated correctly
	//   - UncoveredShifts contains shifts with 0 assignments
	//   - OverallocatedShifts contains shifts with >1 assignments
	//   - CoveragePercentage is between 0 and 100
	//   - Uses exactly 1 database query to load assignment data
	//   - No N+1 query patterns
	//
	// Example:
	//
	//	metrics, err := svc.Calculate(ctx, scheduleVersionID)
	//	if err != nil {
	//	    return fmt.Errorf("coverage calculation failed: %w", err)
	//	}
	//	fmt.Printf("Coverage: %.1f%% (%d/%d)\n",
	//	    metrics.CoveragePercentage,
	//	    metrics.AssignedPositions,
	//	    metrics.RequiredPositions,
	//	)
	Calculate(
		ctx context.Context,
		scheduleVersionID uuid.UUID,
	) (*CoverageMetrics, error)
}

// OrchestrationStatus represents the current state of an orchestration operation.
// Phase represents a stage in the orchestration workflow.
type Phase int

const (
	// PhaseODSImport is the first phase: parsing and importing ODS schedule file
	PhaseODSImport Phase = iota

	// PhaseAmionScrape is the second phase: scraping Amion for assignment data
	PhaseAmionScrape

	// PhaseCoverageCalculation is the third phase: calculating coverage metrics
	PhaseCoverageCalculation
)

// String returns the human-readable name of the phase.
func (p Phase) String() string {
	switch p {
	case PhaseODSImport:
		return "ODS_IMPORT"
	case PhaseAmionScrape:
		return "AMION_SCRAPE"
	case PhaseCoverageCalculation:
		return "COVERAGE_CALCULATION"
	default:
		return "UNKNOWN"
	}
}

// OrchestrationStatus is the status of the orchestration workflow.
type OrchestrationStatus string

const (
	// OrchestrationStatusIDLE indicates no operation is currently running
	OrchestrationStatusIDLE OrchestrationStatus = "IDLE"

	// OrchestrationStatusINPROGRESS indicates an operation is currently executing
	OrchestrationStatusINPROGRESS OrchestrationStatus = "IN_PROGRESS"

	// OrchestrationStatusCOMPLETED indicates the last operation completed successfully
	OrchestrationStatusCOMPLETED OrchestrationStatus = "COMPLETED"

	// OrchestrationStatusFAILED indicates the last operation failed
	OrchestrationStatusFAILED OrchestrationStatus = "FAILED"
)

// OrchestrationResult contains the complete outcome of an orchestration operation.
// It aggregates results from all services involved in the workflow.
type OrchestrationResult struct {
	// ScheduleVersion is the created or updated schedule version
	// Set by ODSImportService
	ScheduleVersion *entity.ScheduleVersion

	// Assignments contains all imported or scraped assignments
	// Set by ODSImportService or AmionScraperService
	Assignments []entity.Assignment

	// Coverage contains calculated coverage metrics for the schedule
	// Set by CoverageCalculatorService
	Coverage *CoverageMetrics

	// ValidationResult aggregates all validation messages from all services
	ValidationResult *validation.ValidationResult

	// Duration tracks how long the entire orchestration operation took
	Duration time.Duration

	// CompletedAt is when the orchestration operation completed
	CompletedAt time.Time

	// Metadata stores arbitrary key-value pairs for operational metadata
	// Examples: "import_source" = "ods_file", "file_size" = "2048000", etc.
	Metadata map[string]interface{}
}

// ScheduleOrchestrator coordinates the complete workflow of importing and
// processing schedule data. It orchestrates multiple services to provide
// a unified import-and-calculate experience.
//
// Responsibilities:
// 1. Coordinate ODS file import with validation
// 2. Calculate coverage metrics for the imported schedule
// 3. Track operation status and timing
// 4. Provide unified error handling and validation reporting
// 5. Support concurrent operations with status tracking
//
// Concurrency: Implementations must be thread-safe. Multiple imports can run
// concurrently, but status is per-operation (not shared across concurrent runs).
//
// Transaction guarantees: All operations within a single ExecuteImport call should
// be coordinated atomically where possible. If one operation fails, the result
// includes partial state.
//
// Example usage:
//
//	orchestrator, _ := /* obtain orchestrator */
//	result, err := orchestrator.ExecuteImport(ctx, filePath, hospitalID, userID)
//	if err != nil {
//	    log.Printf("Import failed: %v", err)
//	    return
//	}
//	log.Printf("Imported version %d with coverage %.1f%%",
//	    result.ScheduleVersion.Version,
//	    result.Coverage.CoveragePercentage,
//	)
type ScheduleOrchestrator interface {
	// ExecuteImport performs a complete schedule import and coverage calculation workflow.
	// This is the primary entry point for importing schedules.
	//
	// Workflow:
	// 1. Import schedule from ODS file via ODSImportService
	// 2. Calculate coverage metrics via CoverageCalculatorService
	// 3. Aggregate results and validation messages
	// 4. Return unified OrchestrationResult
	//
	// Parameters:
	//   - ctx: context.Context for cancellation and timeout (applies to entire workflow)
	//   - filePath: absolute path to the ODS file to import
	//   - hospitalID: UUID of the hospital (required, non-nil)
	//   - userID: UUID of the user performing the import (required, non-nil)
	//
	// Returns:
	//   - *OrchestrationResult: complete result with schedule, assignments, coverage, and validation
	//   - error: if critical failures occur; validation errors are included in result, not as error
	//
	// Error handling:
	//   - File not found or unreadable: returns error, result may be nil
	//   - Validation failures: included in result.ValidationResult, not as error
	//   - Database failures: returns error, result may have partial state
	//   - Coverage calculation failures: included in result, not as error (schedule still valid)
	//   - Context cancellation: returns context.Err()
	//
	// Guarantees:
	//   - OrchestrationResult is always populated (non-nil) on success
	//   - ScheduleVersion status is DRAFT
	//   - All assignments have correct source (ODS_FILE)
	//   - Validation errors are collected from all services
	//   - Duration measures total execution time
	//   - Atomic at ScheduleVersion level (all or nothing)
	//
	// Example with full error handling:
	//
	//	result, err := orchestrator.ExecuteImport(ctx, filePath, hospitalID, userID)
	//	if err != nil {
	//	    log.Printf("Critical failure: %v", err)
	//	    return
	//	}
	//	if result.ValidationResult.HasErrors() {
	//	    log.Printf("Validation errors: %+v", result.ValidationResult.Errors)
	//	    return
	//	}
	//	log.Printf("Import succeeded: version %d, coverage %.1f%%",
	//	    result.ScheduleVersion.Version,
	//	    result.Coverage.CoveragePercentage,
	//	)
	ExecuteImport(
		ctx context.Context,
		filePath string,
		hospitalID uuid.UUID,
		userID uuid.UUID,
	) (*OrchestrationResult, error)

	// GetOrchestrationStatus returns the current status of the orchestrator.
	// This can be used to monitor long-running operations or detect stuck processes.
	//
	// Returns:
	//   - OrchestrationStatus: the current status
	//
	// Status meanings:
	//   - IDLE: no operation is currently running
	//   - IN_PROGRESS: an operation is currently executing (ExecuteImport call is active)
	//   - COMPLETED: the last operation completed successfully
	//   - FAILED: the last operation failed
	//
	// Note: Status reflects the most recent operation only. Concurrent calls to
	// ExecuteImport are not distinguished by status.
	//
	// Example:
	//
	//	status := orchestrator.GetOrchestrationStatus()
	//	if status == OrchestrationStatusINPROGRESS {
	//	    log.Printf("Import is currently running...")
	//	} else if status == OrchestrationStatusCOMPLETED {
	//	    log.Printf("Last import succeeded")
	//	}
	GetOrchestrationStatus() OrchestrationStatus
}
