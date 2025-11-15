package amion

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ScrapedShifts represents the result of scraping multiple months of data
type ScrapedShifts struct {
	// Shifts contains all successfully extracted shifts
	Shifts []RawAmionShift

	// Errors contains any errors that occurred during scraping
	Errors []ScrapingError

	// Warnings contains non-fatal issues like duplicates
	Warnings []ScrapingWarning

	// DuplicateCount tracks the number of duplicate shifts detected
	DuplicateCount int

	// MonthsProcessed tracks how many months were successfully scraped
	MonthsProcessed int

	// MonthsFailed tracks how many months failed to scrape
	MonthsFailed int
}

// ScrapingError represents an error that occurred while scraping a specific URL
type ScrapingError struct {
	Month     string // YYYY-MM format
	URL       string
	Error     error
	ErrorType string // "network", "parse", "http", "retry"
}

// ScrapingWarning represents a non-fatal warning (e.g., duplicate shifts)
type ScrapingWarning struct {
	Month   string // YYYY-MM format
	Message string
}

// AmionScraper coordinates scraping of Amion schedule data.
// It uses a goroutine pool for concurrency and rate limiting to avoid overwhelming the server.
type AmionScraper struct {
	client        *AmionHTTPClient
	pool          *GoroutinePool
	limiter       *RateLimiter
	selectors     *AmionSelectors
	logger        interface{} // For compatibility - zap.SugaredLogger optional
}

// NewAmionScraper creates a new AmionScraper with the specified configuration.
//
// Parameters:
//   - client: The HTTP client to use for fetching
//   - pool: The goroutine pool for concurrent fetching (typically 5 workers)
//   - limiter: The rate limiter (typically 1 second between requests)
//   - selectors: CSS selectors for HTML parsing
//
// Returns:
//   - *AmionScraper: A new scraper instance
//
// Example:
//
//	client, _ := NewAmionHTTPClient("https://amion.example.com")
//	pool := NewGoroutinePool(5)
//	limiter := NewRateLimiter(1 * time.Second)
//	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())
func NewAmionScraper(client *AmionHTTPClient, pool *GoroutinePool, limiter *RateLimiter, selectors *AmionSelectors) *AmionScraper {
	return &AmionScraper{
		client:    client,
		pool:      pool,
		limiter:   limiter,
		selectors: selectors,
	}
}

// ScrapeSchedule scrapes the schedule data for multiple months.
// It returns all successfully extracted shifts plus any errors/warnings encountered.
// Does not fail on partial errors - returns what succeeded.
//
// Parameters:
//   - startDate: The starting month (year-month as time.Time)
//   - monthCount: Number of months to scrape (e.g., 6 for 6 months)
//
// Returns:
//   - *ScrapedShifts: Results including shifts, errors, and warnings
//   - error: Only returned if the entire operation failed (e.g., context cancelled)
//
// Example:
//
//	startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
//	results, err := scraper.ScrapeSchedule(startDate, 6)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Scraped %d shifts, %d errors, %d warnings\n",
//	    len(results.Shifts), len(results.Errors), len(results.Warnings))
func (s *AmionScraper) ScrapeSchedule(startDate time.Time, monthCount int) (*ScrapedShifts, error) {
	ctx := context.Background()
	return s.ScrapeScheduleWithContext(ctx, startDate, monthCount)
}

// ScrapeScheduleWithContext scrapes the schedule data with context support.
// Allows cancellation and timeout control.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - startDate: The starting month
//   - monthCount: Number of months to scrape
//
// Returns:
//   - *ScrapedShifts: Results including shifts, errors, and warnings
//   - error: Context errors or if the entire operation failed
func (s *AmionScraper) ScrapeScheduleWithContext(ctx context.Context, startDate time.Time, monthCount int) (*ScrapedShifts, error) {
	if monthCount < 1 {
		return nil, fmt.Errorf("monthCount must be at least 1, got %d", monthCount)
	}

	results := &ScrapedShifts{
		Shifts:   make([]RawAmionShift, 0),
		Errors:   make([]ScrapingError, 0),
		Warnings: make([]ScrapingWarning, 0),
	}

	// Generate URLs for each month
	monthURLs := s.generateMonthURLs(startDate, monthCount)

	// Use a channel to collect results from workers
	shiftsChan := make(chan []RawAmionShift, len(monthURLs))
	errorsChan := make(chan ScrapingError, len(monthURLs))
	dupCountChan := make(chan int, len(monthURLs))

	// Prepare deduplication tracking
	seenShifts := make(map[string]bool)
	seenMutex := &sync.Mutex{}

	// Submit jobs to the pool
	for _, monthURL := range monthURLs {
		job := s.createScrapingJob(monthURL, shiftsChan, errorsChan, dupCountChan, seenShifts, seenMutex)
		err := s.pool.Submit(job)
		if err == ErrQueueFull {
			// Queue is full - wait a bit and retry
			time.Sleep(100 * time.Millisecond)
			err = s.pool.Submit(job)
			if err != nil {
				results.Errors = append(results.Errors, ScrapingError{
					Month:     monthURL.Month,
					URL:       monthURL.URL,
					Error:     fmt.Errorf("failed to submit job to pool: %w", err),
					ErrorType: "queue",
				})
			}
		}
	}

	// Wait for all jobs to complete
	err := s.pool.Wait(ctx)
	if err != nil {
		return results, err
	}

	// Collect all results from channels
	close(shiftsChan)
	close(errorsChan)
	close(dupCountChan)

	for shifts := range shiftsChan {
		results.Shifts = append(results.Shifts, shifts...)
	}

	for scrapingErr := range errorsChan {
		results.Errors = append(results.Errors, scrapingErr)
		results.MonthsFailed++
	}

	for dupCount := range dupCountChan {
		results.DuplicateCount += dupCount
	}

	// Calculate successful months
	results.MonthsProcessed = len(monthURLs) - results.MonthsFailed

	return results, nil
}

// monthURL represents a URL for a specific month
type monthURL struct {
	Month string // YYYY-MM format
	URL   string // Full URL to fetch
}

// generateMonthURLs generates URLs for each month starting from startDate.
// Handles year boundaries correctly (Dec -> Jan of next year).
//
// Parameters:
//   - startDate: Starting month (uses year and month from this date)
//   - monthCount: Number of months to generate
//
// Returns:
//   - []monthURL: List of month URLs in order
func (s *AmionScraper) generateMonthURLs(startDate time.Time, monthCount int) []monthURL {
	urls := make([]monthURL, 0, monthCount)

	currentDate := startDate
	seenMonths := make(map[string]bool)

	for i := 0; i < monthCount; i++ {
		yearMonth := fmt.Sprintf("%04d-%02d", currentDate.Year(), currentDate.Month())

		// Prevent duplicate months
		if !seenMonths[yearMonth] {
			url := fmt.Sprintf("/schedule/%s", yearMonth)
			urls = append(urls, monthURL{
				Month: yearMonth,
				URL:   url,
			})
			seenMonths[yearMonth] = true
		}

		// Move to next month, handling year boundaries
		currentDate = currentDate.AddDate(0, 1, 0)
	}

	return urls
}

// createScrapingJob creates a job function for a single month's scraping.
// The job respects rate limiting, fetches the URL, extracts shifts, and handles errors.
func (s *AmionScraper) createScrapingJob(
	monthURL monthURL,
	shiftsChan chan<- []RawAmionShift,
	errorsChan chan<- ScrapingError,
	dupCountChan chan<- int,
	seenShifts map[string]bool,
	seenMutex *sync.Mutex,
) Job {
	return func(ctx context.Context) error {
		// Apply rate limiting
		s.limiter.Wait()

		// Fetch and parse the HTML
		doc, err := s.client.FetchAndParseHTMLWithContext(ctx, monthURL.URL)
		if err != nil {
			errorType := "unknown"
			switch err.(type) {
			case *HTTPError:
				errorType = "http"
			case *NetworkError:
				errorType = "network"
			case *ParseError:
				errorType = "parse"
			case *RetryError:
				errorType = "retry"
			}

			errorsChan <- ScrapingError{
				Month:     monthURL.Month,
				URL:       monthURL.URL,
				Error:     err,
				ErrorType: errorType,
			}
			return nil // Don't fail the job, just record the error
		}

		// Extract shifts from the document
		extractionResult := ExtractShiftsWithSelectors(doc, s.selectors)

		// Filter for this month's shifts and check for duplicates
		monthShifts := make([]RawAmionShift, 0)
		dupCount := 0

		for _, shift := range extractionResult.Shifts {
			// Check if shift is for this month
			if !dateStartsWith(shift.Date, monthURL.Month) {
				continue
			}

			// Check for duplicates
			shiftKey := fmt.Sprintf("%s|%s", shift.Date, shift.ShiftType)

			seenMutex.Lock()
			if seenShifts[shiftKey] {
				dupCount++
			} else {
				seenShifts[shiftKey] = true
				monthShifts = append(monthShifts, shift)
			}
			seenMutex.Unlock()
		}

		// Send results to channels
		if len(monthShifts) > 0 || len(extractionResult.Shifts) == 0 {
			shiftsChan <- monthShifts
		} else {
			// All shifts were duplicates or not for this month
			shiftsChan <- make([]RawAmionShift, 0)
		}

		dupCountChan <- dupCount

		return nil
	}
}

// dateStartsWith checks if a date string starts with a month string (YYYY-MM format).
// Both should be in YYYY-MM-DD and YYYY-MM formats respectively.
func dateStartsWith(date, month string) bool {
	return len(date) >= len(month) && date[:len(month)] == month
}

// AddWarning adds a warning to the results (for duplicate detection, etc).
// This is exposed as a utility for external callers.
func (sr *ScrapedShifts) AddWarning(month, message string) {
	sr.Warnings = append(sr.Warnings, ScrapingWarning{
		Month:   month,
		Message: message,
	})
}

// HasErrors returns true if any errors occurred during scraping.
func (sr *ScrapedShifts) HasErrors() bool {
	return len(sr.Errors) > 0
}

// ErrorCount returns the number of errors that occurred.
func (sr *ScrapedShifts) ErrorCount() int {
	return len(sr.Errors)
}

// WarningCount returns the number of warnings that occurred.
func (sr *ScrapedShifts) WarningCount() int {
	return len(sr.Warnings)
}

// TotalShifts returns the total number of shifts scraped (including duplicates detected).
func (sr *ScrapedShifts) TotalShifts() int {
	return len(sr.Shifts) + sr.DuplicateCount
}

// FormattedErrors returns a formatted string of all errors for logging.
func (sr *ScrapedShifts) FormattedErrors() string {
	if !sr.HasErrors() {
		return ""
	}

	result := fmt.Sprintf("Scraping errors (%d):\n", len(sr.Errors))
	for _, err := range sr.Errors {
		result += fmt.Sprintf("  [%s] %s: %v\n", err.Month, err.ErrorType, err.Error)
	}
	return result
}

// FormattedWarnings returns a formatted string of all warnings for logging.
func (sr *ScrapedShifts) FormattedWarnings() string {
	if len(sr.Warnings) == 0 {
		return ""
	}

	result := fmt.Sprintf("Scraping warnings (%d):\n", len(sr.Warnings))
	for _, warn := range sr.Warnings {
		result += fmt.Sprintf("  [%s] %s\n", warn.Month, warn.Message)
	}
	return result
}
