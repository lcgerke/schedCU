package ods

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/repository"
	"github.com/schedcu/reimplement/internal/validation"
)

// ODSImporter handles the complete workflow of importing ODS files into the database.
// It coordinates parsing, validation, and repository operations with comprehensive error handling.
//
// Example usage:
//
//	importer := ods.NewODSImporter(parser, svRepo, siRepo)
//	scheduleVersion, err := importer.Import(ctx, fileContent, hospitalID, userID)
//	if err != nil {
//	    log.Printf("Import failed: %v", err)
//	    result := importer.GetValidationResult()
//	    // handle partial import with result.Errors
//	}
type ODSImporter struct {
	parser            ODSParserInterface
	svRepository      repository.ScheduleVersionRepository
	siRepository      repository.ShiftInstanceRepository
	errorCollector    *ODSErrorCollector
	lastValidationResult *validation.ValidationResult
}

// NewODSImporter creates a new ODS importer with the given dependencies.
// All parameters are required and must be non-nil.
func NewODSImporter(
	parser ODSParserInterface,
	svRepository repository.ScheduleVersionRepository,
	siRepository repository.ShiftInstanceRepository,
) *ODSImporter {
	return &ODSImporter{
		parser:         parser,
		svRepository:   svRepository,
		siRepository:   siRepository,
		errorCollector: NewODSErrorCollector(),
	}
}

// Import executes the complete ODS import workflow.
//
// Workflow:
// 1. Validate input parameters (hospitalID, userID not nil)
// 2. Parse ODS file content
// 3. Create ScheduleVersion in database
// 4. For each shift in parsed data:
//    a. Create ShiftInstance entity
//    b. Persist to database
//    c. Collect any errors
// 5. Build and return validation result
//
// Returns:
// - On success: ScheduleVersion and nil error (all shifts imported)
// - On partial success: ScheduleVersion and non-nil error (some shifts failed)
// - On critical failure: nil ScheduleVersion and error (schedule version creation failed)
//
// The ValidationResult is always available via GetValidationResult() regardless of error.
func (oi *ODSImporter) Import(
	ctx context.Context,
	odsContent []byte,
	hospitalID uuid.UUID,
	userID uuid.UUID,
) (*entity.ScheduleVersion, error) {
	// Reset error collector for new import
	oi.errorCollector = NewODSErrorCollector()

	// Validate inputs
	if hospitalID == uuid.Nil {
		err := fmt.Errorf("invalid hospital ID: nil UUID")
		oi.errorCollector.AddCritical("schedule_version", "", "hospital_id", err.Error(), err)
		oi.lastValidationResult = oi.errorCollector.BuildValidationResult()
		return nil, err
	}

	if userID == uuid.Nil {
		err := fmt.Errorf("invalid user ID: nil UUID")
		oi.errorCollector.AddCritical("schedule_version", "", "user_id", err.Error(), err)
		oi.lastValidationResult = oi.errorCollector.BuildValidationResult()
		return nil, err
	}

	// Check context before proceeding
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Step 1: Parse ODS file
	parsedSchedule, parseErr := oi.parser.Parse(odsContent)
	if parseErr != nil {
		oi.errorCollector.AddCritical("parse", "odsfile", "content", parseErr.Error(), parseErr)
		oi.lastValidationResult = oi.errorCollector.BuildValidationResult()
		return nil, fmt.Errorf("failed to parse ODS file: %w", parseErr)
	}

	if parsedSchedule == nil {
		err := fmt.Errorf("parser returned nil schedule")
		oi.errorCollector.AddCritical("parse", "odsfile", "schedule", err.Error(), err)
		oi.lastValidationResult = oi.errorCollector.BuildValidationResult()
		return nil, err
	}

	// Step 2: Parse dates from parsed schedule
	startDate, endDate, dateErr := oi.parseDates(parsedSchedule)
	if dateErr != nil {
		oi.errorCollector.AddCritical("schedule_version", "", "dates", dateErr.Error(), dateErr)
		oi.lastValidationResult = oi.errorCollector.BuildValidationResult()
		return nil, fmt.Errorf("failed to parse schedule dates: %w", dateErr)
	}

	// Step 3: Create ScheduleVersion
	scheduleVersion := entity.NewScheduleVersion(
		hospitalID,
		1, // TODO: Get next version from repository
		startDate,
		endDate,
		"ods_file",
		userID,
	)

	createdSV, svErr := oi.svRepository.Create(ctx, scheduleVersion)
	if svErr != nil {
		oi.errorCollector.AddCritical("schedule_version", scheduleVersion.ID.String(), "database", svErr.Error(), svErr)
		oi.lastValidationResult = oi.errorCollector.BuildValidationResult()
		return nil, fmt.Errorf("failed to create schedule version: %w", svErr)
	}

	oi.errorCollector.RecordScheduleCreated()
	oi.errorCollector.SetTotalShifts(len(parsedSchedule.Shifts))

	// Step 4: Import shifts
	importHadErrors := false
	for idx, parsedShift := range parsedSchedule.Shifts {
		if ctx.Err() != nil {
			// Context cancelled, stop import
			break
		}

		if parsedShift == nil {
			oi.errorCollector.AddMinor("shift", fmt.Sprintf("shift_%d", idx), "", "parsed shift is nil")
			continue
		}

		shift := oi.createShiftInstanceFromParsed(createdSV.ID, parsedShift, userID)

		createdShift, shiftErr := oi.siRepository.Create(ctx, shift)
		if shiftErr != nil {
			importHadErrors = true
			oi.errorCollector.AddMajor(
				"shift",
				fmt.Sprintf("%s_%s", shift.ShiftType, idx),
				"database",
				fmt.Sprintf("failed to create shift: %s", shiftErr.Error()),
				shiftErr,
			)
			oi.errorCollector.RecordShiftFailed()
			continue
		}

		if createdShift != nil {
			oi.errorCollector.RecordShiftCreated()
		}
	}

	// Build validation result
	oi.lastValidationResult = oi.errorCollector.BuildValidationResult()

	// Return result
	if importHadErrors {
		return createdSV, fmt.Errorf(
			"import completed with %d errors: %d/%d shifts created",
			oi.errorCollector.MajorErrorCount(),
			oi.errorCollector.createdShifts,
			len(parsedSchedule.Shifts),
		)
	}

	return createdSV, nil
}

// GetValidationResult returns the most recent validation result from Import.
// Returns nil if Import has not been called yet.
func (oi *ODSImporter) GetValidationResult() *validation.ValidationResult {
	return oi.lastValidationResult
}

// parseDates parses start and end dates from the parsed schedule.
func (oi *ODSImporter) parseDates(ps *ParsedSchedule) (time.Time, time.Time, error) {
	startDate, err := time.Parse("2006-01-02", ps.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date %q: %w", ps.StartDate, err)
	}

	endDate, err := time.Parse("2006-01-02", ps.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date %q: %w", ps.EndDate, err)
	}

	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("end date %s is before start date %s", ps.EndDate, ps.StartDate)
	}

	return startDate, endDate, nil
}

// createShiftInstanceFromParsed converts parsed shift data into a ShiftInstance entity.
func (oi *ODSImporter) createShiftInstanceFromParsed(
	scheduleVersionID uuid.UUID,
	ps *ParsedShift,
	userID uuid.UUID,
) *entity.ShiftInstance {
	shift := entity.NewShiftInstance(
		scheduleVersionID,
		ps.ShiftType,
		ps.Position,
		ps.Location,
		ps.StaffMember,
		userID,
	)

	shift.RequiredQualification = ps.RequiredQualification

	// Optional fields
	if ps.SpecialtyConstraint != "" {
		shift.SpecialtyConstraint = &ps.SpecialtyConstraint
	}
	if ps.StudyType != "" {
		shift.StudyType = &ps.StudyType
	}

	return shift
}

// BatchImport imports multiple ODS files for the same hospital.
// Returns a map of file identifiers to import results (either ScheduleVersion or error).
// This is useful for bulk operations.
func (oi *ODSImporter) BatchImport(
	ctx context.Context,
	files map[string][]byte, // filename -> content
	hospitalID uuid.UUID,
	userID uuid.UUID,
) map[string]interface{} {
	results := make(map[string]interface{})

	for filename, content := range files {
		sv, err := oi.Import(ctx, content, hospitalID, userID)
		if err != nil {
			results[filename] = err
		} else {
			results[filename] = sv
		}
	}

	return results
}

// GetErrorMetrics returns a summary of errors from the last import.
// Useful for reporting and monitoring.
func (oi *ODSImporter) GetErrorMetrics() map[string]interface{} {
	ec := oi.errorCollector
	if ec == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"total_errors":        ec.ErrorCount(),
		"critical_errors":     countBySeverity(ec.AllErrors(), ErrorSeverityCritical),
		"major_errors":        countBySeverity(ec.AllErrors(), ErrorSeverityMajor),
		"minor_errors":        countBySeverity(ec.AllErrors(), ErrorSeverityMinor),
		"info_messages":       countBySeverity(ec.AllErrors(), ErrorSeverityInfo),
		"total_shifts":        ec.totalShifts,
		"created_shifts":      ec.createdShifts,
		"failed_shifts":       ec.failedShifts,
		"success_rate_shifts": calculateSuccessRate(ec.createdShifts, ec.totalShifts),
	}
}

// countBySeverity counts errors of a specific severity.
func countBySeverity(errors []ImportError, severity ErrorSeverity) int {
	count := 0
	for _, e := range errors {
		if e.Severity == severity {
			count++
		}
	}
	return count
}

// calculateSuccessRate returns percentage of successful shifts.
func calculateSuccessRate(created, total int) float64 {
	if total == 0 {
		return 100.0
	}
	return (float64(created) / float64(total)) * 100.0
}
