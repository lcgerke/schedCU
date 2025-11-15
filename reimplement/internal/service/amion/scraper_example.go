package amion

// ExampleAmionScraper_Basic demonstrates basic usage of the AmionScraper
// to fetch 6 months of schedule data.
//
// Example usage:
//
//	func main() {
//	    client, _ := NewAmionHTTPClient("https://amion.example.com")
//	    defer client.Close()
//
//	    pool := NewGoroutinePool(5)
//	    defer pool.Close()
//
//	    limiter := NewRateLimiter(1 * time.Second)
//	    scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())
//
//	    startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
//	    results, err := scraper.ScrapeSchedule(startDate, 6)
//
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    fmt.Printf("Scraped %d shifts\n", len(results.Shifts))
//	    fmt.Printf("Errors: %d, Warnings: %d\n", len(results.Errors), len(results.Warnings))
//	}
func ExampleAmionScraper_Basic() {
	// This is an example function - it's not executed but demonstrates usage
}

// ExampleAmionScraper_WithErrorHandling demonstrates error handling
// for partial failures during scraping.
//
// Example usage:
//
//	func main() {
//	    client, _ := NewAmionHTTPClient("https://amion.example.com")
//	    defer client.Close()
//
//	    pool := NewGoroutinePool(5)
//	    defer pool.Close()
//
//	    limiter := NewRateLimiter(1 * time.Second)
//	    scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())
//
//	    startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
//	    results, err := scraper.ScrapeSchedule(startDate, 6)
//
//	    if err != nil {
//	        // Complete failure - context cancelled or similar
//	        log.Fatalf("Scraping failed completely: %v", err)
//	    }
//
//	    // Partial success is OK - some months may have failed
//	    fmt.Printf("Months processed: %d\n", results.MonthsProcessed)
//	    fmt.Printf("Months failed: %d\n", results.MonthsFailed)
//	    fmt.Printf("Shifts extracted: %d\n", len(results.Shifts))
//
//	    if results.HasErrors() {
//	        fmt.Print(results.FormattedErrors())
//	    }
//
//	    if len(results.Warnings) > 0 {
//	        fmt.Print(results.FormattedWarnings())
//	    }
//
//	    // Process shifts even with errors
//	    for _, shift := range results.Shifts {
//	        fmt.Printf("Shift: %s %s %s-%s\n",
//	            shift.Date, shift.ShiftType, shift.StartTime, shift.EndTime)
//	    }
//	}
func ExampleAmionScraper_WithErrorHandling() {
	// This is an example function - it's not executed but demonstrates usage
}

// ExampleAmionScraper_DuplicateHandling demonstrates duplicate detection.
//
// Example usage:
//
//	func main() {
//	    client, _ := NewAmionHTTPClient("https://amion.example.com")
//	    defer client.Close()
//
//	    pool := NewGoroutinePool(5)
//	    defer pool.Close()
//
//	    limiter := NewRateLimiter(1 * time.Second)
//	    scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())
//
//	    startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
//	    results, _ := scraper.ScrapeSchedule(startDate, 6)
//
//	    // Duplicates are automatically detected and removed
//	    fmt.Printf("Unique shifts: %d\n", len(results.Shifts))
//	    fmt.Printf("Duplicates detected: %d\n", results.DuplicateCount)
//	    fmt.Printf("Total before dedup: %d\n", results.TotalShifts())
//
//	    // Duplicates are based on: date + shift_type
//	    // So a shift on 2025-11-15 with type "Technologist" that appears
//	    // in multiple months would only be returned once
//	}
func ExampleAmionScraper_DuplicateHandling() {
	// This is an example function - it's not executed but demonstrates usage
}

// ExampleAmionScraper_PerformanceMetrics demonstrates monitoring performance
// of the scraping operation.
//
// Example usage:
//
//	func main() {
//	    client, _ := NewAmionHTTPClient("https://amion.example.com")
//	    defer client.Close()
//
//	    pool := NewGoroutinePool(5)
//	    defer pool.Close()
//
//	    limiter := NewRateLimiter(1 * time.Second)
//	    scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())
//
//	    startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
//
//	    startTime := time.Now()
//	    results, _ := scraper.ScrapeSchedule(startDate, 6)
//	    duration := time.Since(startTime)
//
//	    // Performance metrics
//	    fmt.Printf("Total time: %v\n", duration)
//	    fmt.Printf("Shifts per second: %.2f\n",
//	        float64(len(results.Shifts))/duration.Seconds())
//	    fmt.Printf("Months per second: %.2f\n",
//	        float64(results.MonthsProcessed)/duration.Seconds())
//	    fmt.Printf("Success rate: %.1f%%\n",
//	        float64(results.MonthsProcessed)*100/6)
//
//	    // Target: < 3 seconds for 6 months with 5 workers and 1 second rate limiting
//	    if duration < 3*time.Second {
//	        fmt.Println("Performance target met!")
//	    }
//	}
func ExampleAmionScraper_PerformanceMetrics() {
	// This is an example function - it's not executed but demonstrates usage
}

// ExampleAmionScraper_ConfigurationTuning demonstrates how to tune the scraper
// for different performance requirements.
//
// Example usage:
//
//	func main() {
//	    client, _ := NewAmionHTTPClient("https://amion.example.com")
//	    defer client.Close()
//
//	    // Configuration 1: Conservative (fewer workers, longer rate limit)
//	    // Use when targeting a server with strict rate limiting
//	    pool1 := NewGoroutinePool(2)
//	    limiter1 := NewRateLimiter(2 * time.Second)
//	    scraper1 := NewAmionScraper(client, pool1, limiter1, DefaultSelectors())
//
//	    // Configuration 2: Aggressive (more workers, shorter rate limit)
//	    // Use when targeting a robust server and performance is critical
//	    pool2 := NewGoroutinePool(10)
//	    limiter2 := NewRateLimiter(500 * time.Millisecond)
//	    scraper2 := NewAmionScraper(client, pool2, limiter2, DefaultSelectors())
//
//	    // Configuration 3: Balanced (default)
//	    // 5 workers, 1 second rate limit - good for most scenarios
//	    pool3 := NewGoroutinePool(5)
//	    limiter3 := NewRateLimiter(1 * time.Second)
//	    scraper3 := NewAmionScraper(client, pool3, limiter3, DefaultSelectors())
//
//	    startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
//
//	    fmt.Println("Conservative approach:")
//	    results1, _ := scraper1.ScrapeSchedule(startDate, 6)
//	    fmt.Printf("  Completed: %d/%d months, Shifts: %d\n",
//	        results1.MonthsProcessed, 6, len(results1.Shifts))
//
//	    fmt.Println("Balanced approach:")
//	    results3, _ := scraper3.ScrapeSchedule(startDate, 6)
//	    fmt.Printf("  Completed: %d/%d months, Shifts: %d\n",
//	        results3.MonthsProcessed, 6, len(results3.Shifts))
//	}
func ExampleAmionScraper_ConfigurationTuning() {
	// This is an example function - it's not executed but demonstrates usage
}

// ExampleAmionScraper_IntegrationWithDatabase demonstrates how to integrate
// scraper results with database storage.
//
// Example usage:
//
//	type ShiftRepository interface {
//	    Insert(ctx context.Context, shift *RawAmionShift) error
//	    InsertBatch(ctx context.Context, shifts []RawAmionShift) error
//	}
//
//	func storeScrapedShifts(repo ShiftRepository, results *ScrapedShifts) error {
//	    // Check for errors/warnings
//	    if results.MonthsFailed > 0 {
//	        fmt.Printf("Warning: %d months failed to scrape\n", results.MonthsFailed)
//	    }
//
//	    if results.DuplicateCount > 0 {
//	        fmt.Printf("Info: %d duplicate shifts detected and removed\n", results.DuplicateCount)
//	    }
//
//	    // Store all shifts in database
//	    ctx := context.Background()
//	    if err := repo.InsertBatch(ctx, results.Shifts); err != nil {
//	        return fmt.Errorf("failed to store shifts: %w", err)
//	    }
//
//	    fmt.Printf("Successfully stored %d shifts from %d months\n",
//	        len(results.Shifts), results.MonthsProcessed)
//	    return nil
//	}
func ExampleAmionScraper_IntegrationWithDatabase() {
	// This is an example function - it's not executed but demonstrates usage
}

// ExampleAmionScraper_FullWorkflow demonstrates a complete workflow including
// configuration, scraping, error handling, and results processing.
//
// Example usage:
//
//	func main() {
//	    // Step 1: Setup HTTP client
//	    client, err := NewAmionHTTPClient("https://amion.example.com")
//	    if err != nil {
//	        log.Fatalf("Failed to create HTTP client: %v", err)
//	    }
//	    defer client.Close()
//
//	    // Step 2: Setup worker pool and rate limiter
//	    // 5 workers, 1 second between requests = target of ~3 seconds for 6 months
//	    pool := NewGoroutinePool(5)
//	    defer pool.Close()
//
//	    limiter := NewRateLimiter(1 * time.Second)
//
//	    // Step 3: Create scraper with custom selectors if needed
//	    selectors := DefaultSelectors()
//	    scraper := NewAmionScraper(client, pool, limiter, selectors)
//
//	    // Step 4: Scrape schedule data
//	    startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
//	    fmt.Println("Scraping 6 months of schedule data...")
//
//	    startTime := time.Now()
//	    results, err := scraper.ScrapeSchedule(startDate, 6)
//	    duration := time.Since(startTime)
//
//	    if err != nil {
//	        log.Fatalf("Scraping failed: %v", err)
//	    }
//
//	    // Step 5: Report results
//	    fmt.Printf("Scraping completed in %v\n", duration)
//	    fmt.Printf("Months processed: %d, Failed: %d\n",
//	        results.MonthsProcessed, results.MonthsFailed)
//	    fmt.Printf("Shifts extracted: %d\n", len(results.Shifts))
//	    fmt.Printf("Duplicates detected: %d\n", results.DuplicateCount)
//
//	    if results.HasErrors() {
//	        fmt.Printf("Errors occurred:\n%s\n", results.FormattedErrors())
//	    }
//
//	    if len(results.Warnings) > 0 {
//	        fmt.Printf("Warnings:\n%s\n", results.FormattedWarnings())
//	    }
//
//	    // Step 6: Process shifts
//	    for _, shift := range results.Shifts {
//	        fmt.Printf("  %s: %s from %s to %s (%s)\n",
//	            shift.Date, shift.ShiftType, shift.StartTime, shift.EndTime, shift.Location)
//	    }
//	}
func ExampleAmionScraper_FullWorkflow() {
	// This is an example function - it's not executed but demonstrates usage
}
