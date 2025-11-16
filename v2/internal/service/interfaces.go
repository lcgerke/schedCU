package service

import (
	"context"
	"io"
	"time"

	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/validation"
)

// ScheduleVersionService manages schedule versions (creation, promotion, archival)
type ScheduleVersionService interface {
	CreateVersion(ctx context.Context, hospitalID entity.HospitalID, startDate, endDate entity.Date, creatorID entity.UserID) (*entity.ScheduleVersion, error)
	GetVersion(ctx context.Context, id entity.ScheduleVersionID) (*entity.ScheduleVersion, error)
	GetActiveVersion(ctx context.Context, hospitalID entity.HospitalID, date entity.Date) (*entity.ScheduleVersion, error)
	ListVersionsByStatus(ctx context.Context, hospitalID entity.HospitalID, status entity.VersionStatus) ([]*entity.ScheduleVersion, error)
	ListAllVersions(ctx context.Context, hospitalID entity.HospitalID) ([]*entity.ScheduleVersion, error)
	PromoteToProduction(ctx context.Context, id entity.ScheduleVersionID, promoterID entity.UserID) error
	Archive(ctx context.Context, id entity.ScheduleVersionID, archiverID entity.UserID) error
	Delete(ctx context.Context, id entity.ScheduleVersionID, deleterID entity.UserID) error
	PromoteAndArchiveOthers(ctx context.Context, id entity.ScheduleVersionID, promoterID entity.UserID) error
}

// ODSImportService handles importing schedules from ODS files
type ODSImportService interface {
	ImportODSFile(
		ctx context.Context,
		hospitalID entity.HospitalID,
		version *entity.ScheduleVersion,
		filename string,
		content io.Reader,
	) (*entity.ScrapeBatch, *validation.Result, error)
}

// AmionImportService handles scraping and importing schedules from Amion
type AmionImportService interface {
	ScrapeAndImport(ctx context.Context, hospitalID entity.HospitalID, scheduleVersion *entity.ScheduleVersion, config AmionScraperConfig) (*entity.ScrapeBatch, *validation.Result, error)
}

// CoverageCalculator calculates coverage metrics for schedules
type CoverageCalculator interface {
	CalculateCoverageForSchedule(ctx context.Context, scheduleVersionID entity.ScheduleVersionID, startDate, endDate time.Time) (*entity.CoverageCalculation, error)
	CalculateCoverage(ctx context.Context, scheduleVersionID entity.ScheduleVersionID, startDate, endDate time.Time) (*entity.CoverageCalculation, *validation.Result)
}

// ScheduleOrchestrator coordinates the full scheduling workflow
type ScheduleOrchestrator interface {
	// ExecuteFullWorkflow executes the complete 3-phase workflow
	// This is a placeholder - actual signatures are in implementation
}

// ODSParser parses ODS files and extracts schedule data
type ODSParser interface {
	ParseFile(filePath string, result *validation.Result) (*ODSData, error)
}
