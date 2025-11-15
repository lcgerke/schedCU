package service

import (
	"context"
	"fmt"
	"time"

	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
	"github.com/schedcu/v2/internal/validation"
)

// AmionImportService handles scraping and importing schedules from Amion
type AmionImportService struct {
	assignmentRepo repository.AssignmentRepository
	batchRepo      repository.ScrapeBatchRepository
	versionRepo    repository.ScheduleVersionRepository
}

// NewAmionImportService creates a new Amion import service
func NewAmionImportService(
	assignmentRepo repository.AssignmentRepository,
	batchRepo repository.ScrapeBatchRepository,
	versionRepo repository.ScheduleVersionRepository,
) *AmionImportService {
	return &AmionImportService{
		assignmentRepo: assignmentRepo,
		batchRepo:      batchRepo,
		versionRepo:    versionRepo,
	}
}

// AmionScraperConfig contains configuration for Amion scraping
type AmionScraperConfig struct {
	Username       string
	Password       string
	MonthsToScrape int       // Number of months to look back (e.g., 6 months)
	StartDate      time.Time // Start date for scraping
	ConcurrentWorkers int    // Number of concurrent goroutines for scraping
}

// ScrapeAndImport scrapes Amion and imports the data as a batch
// Returns a ScrapeBatch and validation result (with all issues collected)
func (s *AmionImportService) ScrapeAndImport(
	ctx context.Context,
	hospitalID entity.HospitalID,
	version *entity.ScheduleVersion,
	config AmionScraperConfig,
) (*entity.ScrapeBatch, *validation.Result, error) {

	// Create a batch to track the scrape
	batch := &entity.ScrapeBatch{
		HospitalID:      hospitalID,
		State:           entity.BatchStatePending,
		WindowStartDate: version.EffectiveStartDate,
		WindowEndDate:   version.EffectiveEndDate,
		ScrapedAt:       entity.Now(),
		CreatedAt:       entity.Now(),
		CreatedBy:       version.CreatedBy,
	}

	// Initialize validation result
	result := validation.NewResult()

	// Scrape Amion (with concurrency from Spike 1)
	// NOTE: Using mock implementation for Phase 1b; real scraping in Phase 3 with goquery/Chromedp
	scraped, scrapeErrors := s.scrapeAmion(ctx, config)
	for _, msg := range scrapeErrors {
		result.Add(msg.Severity, msg.Code, msg.Text, msg.Context)
	}

	// If we have critical scrape errors, mark batch as failed
	if result.HasErrors() && len(scraped) == 0 {
		batch.State = entity.BatchStateFailed
		errMsg := "Failed to scrape Amion: no schedules extracted"
		batch.ErrorMessage = &errMsg
		return batch, result, nil
	}

	// Import each scraped schedule
	for _, scrapedSchedule := range scraped {
		if err := s.importScrapedSchedule(ctx, version, scrapedSchedule, result); err != nil {
			result.AddError("AMION_IMPORT_FAILED", fmt.Sprintf("Failed to import Amion schedule: %v", err))
			batch.State = entity.BatchStateFailed
			return batch, result, nil
		}
	}

	// Mark batch as complete
	batch.State = entity.BatchStateComplete
	batch.RowCount = len(scraped)

	return batch, result, nil
}

// importScrapedSchedule imports assignments from scraped Amion data
func (s *AmionImportService) importScrapedSchedule(
	ctx context.Context,
	version *entity.ScheduleVersion,
	scraped *scrapedAmionSchedule,
	result *validation.Result,
) error {

	// For each assignment in the scraped data
	for _, assignment := range scraped.Assignments {
		// Find the corresponding shift instance
		// NOTE: In production, this would query the database
		// For Phase 1b, we use the in-memory repo which will find it

		assign := &entity.Assignment{
			PersonID:          assignment.PersonID,
			ShiftInstanceID:   assignment.ShiftInstanceID,
			ScheduleDate:      assignment.ScheduleDate,
			OriginalShiftType: assignment.OriginalShiftType,
			Source:            entity.AssignmentSourceAmion,
			CreatedAt:         entity.Now(),
			CreatedBy:         version.CreatedBy,
		}

		// Save assignment
		if err := s.assignmentRepo.Create(ctx, assign); err != nil {
			result.AddWarning("AMION_ASSIGNMENT_FAILED", fmt.Sprintf("Failed to create assignment: %v", err))
			continue
		}
	}

	return nil
}

// scrapeAmion scrapes Amion for schedule data
// This is a placeholder for Phase 1b; real implementation in Phase 3
func (s *AmionImportService) scrapeAmion(
	ctx context.Context,
	config AmionScraperConfig,
) ([]*scrapedAmionSchedule, []*validation.Message) {

	var schedules []*scrapedAmionSchedule
	var errors []*validation.Message

	// TODO: Implement real Amion scraping using goquery (Spike 1 results)
	// Expected performance from Spike 1: 6 months of schedules in 2-3 seconds with 5 concurrent goroutines
	//
	// Placeholder implementation for Phase 1b to unblock service layer testing:
	// - Would use HTTP client + goquery CSS selectors (Spike 1 approach)
	// - Would support 5 concurrent month scrapers with rate limiting (1 sec between requests)
	// - Would collect parsing errors in validation result (not fail-fast)
	// - Would handle Chromedp fallback if goquery fails (Spike 1 fallback plan)

	return schedules, errors
}

// scrapedAmionSchedule represents a schedule scraped from Amion
type scrapedAmionSchedule struct {
	Month       time.Time
	Assignments []*scrapedAmionAssignment
}

// scrapedAmionAssignment represents an assignment scraped from Amion
type scrapedAmionAssignment struct {
	PersonID          entity.PersonID
	ShiftInstanceID   entity.ShiftInstanceID
	ScheduleDate      entity.Date
	OriginalShiftType string
	Source            string
}

// VerifyAmionConnection verifies that Amion is accessible
// Used as a health check before scraping
func (s *AmionImportService) VerifyAmionConnection(ctx context.Context, config AmionScraperConfig) error {
	// TODO: Implement health check against Amion
	// Would verify authentication and basic connectivity
	return nil
}
