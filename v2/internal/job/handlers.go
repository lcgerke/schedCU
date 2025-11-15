package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/service"
)

// JobHandlers manages job execution handlers
type JobHandlers struct {
	odsImporter    *service.ODSImportService
	amionImporter  *service.AmionImportService
	coverageCalc   *service.DynamicCoverageCalculator
	versionService *service.ScheduleVersionService
}

// NewJobHandlers creates a new job handlers instance
func NewJobHandlers(
	odsImporter *service.ODSImportService,
	amionImporter *service.AmionImportService,
	coverageCalc *service.DynamicCoverageCalculator,
	versionService *service.ScheduleVersionService,
) *JobHandlers {
	return &JobHandlers{
		odsImporter:    odsImporter,
		amionImporter:  amionImporter,
		coverageCalc:   coverageCalc,
		versionService: versionService,
	}
}

// RegisterHandlers registers all job handlers with the Asynq mux
func (h *JobHandlers) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeODSImport, h.HandleODSImport)
	mux.HandleFunc(TypeAmionScrape, h.HandleAmionScrape)
	mux.HandleFunc(TypeCoverageCalc, h.HandleCoverageCalculation)
}

// HandleODSImport handles ODS import jobs
func (h *JobHandlers) HandleODSImport(ctx context.Context, t *asynq.Task) error {
	var payload ODSImportPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	log.Printf("Executing ODS import job: hospital=%s, filename=%s", payload.HospitalID, payload.Filename)

	// Get the schedule version
	_, err := h.versionService.GetVersion(ctx, payload.VersionID)
	if err != nil {
		log.Printf("Failed to get schedule version: %v", err)
		return fmt.Errorf("schedule version not found: %w", err)
	}

	// NOTE: In production, we would read the ODS file from storage (S3, local disk, etc.)
	// For Phase 1b, this is a placeholder. Real implementation in Phase 3.

	// Execute import
	// TODO: Read file from storage and pass to ImportODSFile

	log.Printf("ODS import completed for hospital=%s", payload.HospitalID)

	return nil
}

// HandleAmionScrape handles Amion scraping jobs
func (h *JobHandlers) HandleAmionScrape(ctx context.Context, t *asynq.Task) error {
	var payload AmionScrapePayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	log.Printf("Executing Amion scrape job: hospital=%s, months=%d", payload.HospitalID, payload.MonthsBack)

	// Get the schedule version
	version, err := h.versionService.GetVersion(ctx, payload.VersionID)
	if err != nil {
		log.Printf("Failed to get schedule version: %v", err)
		return fmt.Errorf("schedule version not found: %w", err)
	}

	// Configure Amion scraper
	// NOTE: Password would come from Vault in production
	config := service.AmionScraperConfig{
		Username:          payload.Username,
		Password:          "", // TODO: Load from Vault
		MonthsToScrape:    payload.MonthsBack,
		ConcurrentWorkers: 5, // From Spike 1: optimal concurrent scrapers
	}

	// Execute scrape
	batch, result, err := h.amionImporter.ScrapeAndImport(ctx, entity.HospitalID(payload.HospitalID), version, config)
	if err != nil {
		log.Printf("Amion scrape failed: %v", err)
		return fmt.Errorf("amion scrape error: %w", err)
	}

	if batch.State == entity.BatchStateFailed {
		log.Printf("Amion scrape produced no valid data: %s", *batch.ErrorMessage)
		return fmt.Errorf("amion scrape failed: %s", *batch.ErrorMessage)
	}

	log.Printf("Amion scrape completed: hospital=%s, records=%d, errors=%d",
		payload.HospitalID, batch.RowCount, len(result.Messages))

	return nil
}

// HandleCoverageCalculation handles coverage calculation jobs
func (h *JobHandlers) HandleCoverageCalculation(ctx context.Context, t *asynq.Task) error {
	var payload CoverageCalcPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	log.Printf("Executing coverage calculation job: version=%s, period=%s to %s",
		payload.ScheduleVersionID, payload.StartDate, payload.EndDate)

	// Calculate coverage
	_, err := h.coverageCalc.CalculateCoverageForSchedule(
		ctx,
		payload.ScheduleVersionID,
		payload.StartDate,
		payload.EndDate,
	)

	if err != nil {
		log.Printf("Coverage calculation failed: %v", err)
		return fmt.Errorf("coverage calculation failed: %w", err)
	}

	log.Printf("Coverage calculation completed: version=%s",
		payload.ScheduleVersionID)

	return nil
}
