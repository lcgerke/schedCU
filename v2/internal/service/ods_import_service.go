package service

import (
	"context"
	"fmt"
	"io"

	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
	"github.com/schedcu/v2/internal/validation"
)

// odsImportService is the concrete implementation of ODSImportService
type odsImportService struct {
	shiftRepo      repository.ShiftInstanceRepository
	assignmentRepo repository.AssignmentRepository
	versionRepo    repository.ScheduleVersionRepository
	coverageCalc   CoverageCalculator
}

// NewODSImportService creates a new ODS import service
func NewODSImportService(
	shiftRepo repository.ShiftInstanceRepository,
	assignmentRepo repository.AssignmentRepository,
	versionRepo repository.ScheduleVersionRepository,
	coverageCalc CoverageCalculator,
) ODSImportService {
	return &odsImportService{
		shiftRepo:      shiftRepo,
		assignmentRepo: assignmentRepo,
		versionRepo:    versionRepo,
		coverageCalc:   coverageCalc,
	}
}

// ImportODSFile imports a schedule from an ODS file
// Returns a ScrapeBatch, a validation result (with all issues collected), and any fatal errors
func (s *odsImportService) ImportODSFile(
	ctx context.Context,
	hospitalID entity.HospitalID,
	version *entity.ScheduleVersion,
	filename string,
	content io.Reader,
) (*entity.ScrapeBatch, *validation.Result, error) {

	// Create a batch to track the import
	batch := &entity.ScrapeBatch{
		HospitalID:      hospitalID,
		State:           entity.BatchStatePending,
		WindowStartDate: version.EffectiveStartDate,
		WindowEndDate:   version.EffectiveEndDate,
		ScrapedAt:       entity.Now(),
		CreatedAt:       entity.Now(),
		CreatedBy:       version.CreatedBy,
	}

	// Initialize validation result (collect all errors, don't fail fast)
	result := validation.NewResult()

	// Parse ODS file and extract schedules
	// NOTE: Using mock implementation for Phase 1b; real ODS parsing in Phase 3
	schedules, parseErrors := s.parseODSFile(ctx, content)
	for _, msg := range parseErrors {
		result.Add(msg.Severity, msg.Code, msg.Text, msg.Context)
	}

	// If we have critical parse errors, mark batch as failed
	if result.HasErrors() && len(schedules) == 0 {
		batch.State = entity.BatchStateFailed
		errMsg := "Failed to parse ODS file: no schedules extracted"
		batch.ErrorMessage = &errMsg
		return batch, result, nil
	}

	// Import each schedule into the database
	for _, sched := range schedules {
		if err := s.importSchedule(ctx, version, sched, result); err != nil {
			result.AddError("SCHEDULE_IMPORT_FAILED", fmt.Sprintf("Failed to import schedule: %v", err))
			batch.State = entity.BatchStateFailed
			return batch, result, nil
		}
	}

	// Mark batch as complete
	batch.State = entity.BatchStateComplete
	batch.RowCount = len(schedules)

	return batch, result, nil
}

// importSchedule imports a single schedule into the database
func (s *odsImportService) importSchedule(
	ctx context.Context,
	version *entity.ScheduleVersion,
	sched *parsedSchedule,
	result *validation.Result,
) error {

	// Validate and import shifts
	for _, shift := range sched.Shifts {
		// Create shift instance
		shiftInstance := &entity.ShiftInstance{
			ScheduleVersionID:  version.ID,
			HospitalID:         version.HospitalID,
			ShiftType:          shift.Type,
			ScheduleDate:       shift.Date,
			StartTime:          "00:00", // Format: HH:MM (will be set from parsed data)
			EndTime:            "23:59",
			StudyType:          shift.StudyType,
			SpecialtyConstraint: shift.SpecialtyConstraint,
			DesiredCoverage:    shift.DesiredCoverage,
			IsMandatory:        shift.IsMandatory,
			CreatedAt:          entity.Now(),
			CreatedBy:          version.CreatedBy,
		}

		// Save shift instance
		if err := s.shiftRepo.Create(ctx, shiftInstance); err != nil {
			result.AddError("SHIFT_CREATION_FAILED", fmt.Sprintf("Failed to create shift: %v", err))
			continue
		}

		// Create assignments from parsed data
		for _, assignment := range shift.Assignments {
			assign := &entity.Assignment{
				PersonID:          assignment.PersonID,
				ShiftInstanceID:   shiftInstance.ID,
				ScheduleDate:      shift.Date,
				OriginalShiftType: string(shift.Type),
				Source:            entity.AssignmentSourceManual,
				CreatedAt:         entity.Now(),
				CreatedBy:         version.CreatedBy,
			}

			if err := s.assignmentRepo.Create(ctx, assign); err != nil {
				result.AddError("ASSIGNMENT_CREATION_FAILED", fmt.Sprintf("Failed to assign person: %v", err))
				continue
			}
		}
	}

	return nil
}

// parseODSFile parses an ODS file and returns schedules
// This is a placeholder for Phase 1b; real implementation in Phase 3
func (s *odsImportService) parseODSFile(
	ctx context.Context,
	content io.Reader,
) ([]*parsedSchedule, []*validation.Message) {

	var schedules []*parsedSchedule
	var errors []*validation.Message

	// TODO: Implement real ODS parsing using ODS library from Spike 3
	// For now, return empty to unblock service layer testing

	return schedules, errors
}

// parsedSchedule represents a schedule parsed from ODS file
type parsedSchedule struct {
	Name   string
	Shifts []*parsedShift
}

// parsedShift represents a shift parsed from ODS
type parsedShift struct {
	Date                 entity.Date
	StartTime            entity.Time
	EndTime              entity.Time
	Type                 entity.ShiftType
	StudyType            entity.StudyType
	SpecialtyConstraint  entity.SpecialtyType
	DesiredCoverage      int
	IsMandatory          bool
	Assignments          []*parsedAssignment
}

// parsedAssignment represents an assignment parsed from ODS
type parsedAssignment struct {
	PersonID entity.PersonID
	Role     string
}
