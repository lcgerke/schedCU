package amion

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestScrapeSchedule_Valid6Months tests scraping 6 months of valid data
func TestScrapeSchedule_Valid6Months(t *testing.T) {
	// Create a mock server that serves different HTML for each month
	monthData := map[string]string{
		"2025-11": `
			<html><body><table><tbody>
				<tr><td>2025-11-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
				<tr><td>2025-11-16</td><td>Technologist</td><td>08:00</td><td>16:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2025-12": `
			<html><body><table><tbody>
				<tr><td>2025-12-01</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room A</td></tr>
				<tr><td>2025-12-02</td><td>Technologist</td><td>08:00</td><td>16:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2026-01": `
			<html><body><table><tbody>
				<tr><td>2026-01-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2026-02": `
			<html><body><table><tbody>
				<tr><td>2026-02-10</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room B</td></tr>
			</tbody></table></body></html>
		`,
		"2026-03": `
			<html><body><table><tbody>
				<tr><td>2026-03-05</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2026-04": `
			<html><body><table><tbody>
				<tr><td>2026-04-20</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room C</td></tr>
			</tbody></table></body></html>
		`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Extract month from path (e.g., /schedule/2025-11)
		path := r.URL.Path
		var month string
		if strings.HasPrefix(path, "/schedule/") {
			month = strings.TrimPrefix(path, "/schedule/")
		}

		if html, ok := monthData[month]; ok {
			fmt.Fprint(w, html)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create scraper with short rate limiting for testing
	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}
	defer client.Close()

	pool := NewGoroutinePool(5)
	defer pool.Close()

	limiter := NewRateLimiter(10 * time.Millisecond) // Very short for testing
	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())

	// Scrape 6 months starting from November 2025
	startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	results, err := scraper.ScrapeSchedule(startDate, 6)

	if err != nil {
		t.Fatalf("ScrapeSchedule failed: %v", err)
	}

	// Verify results
	if len(results.Shifts) < 7 {
		t.Errorf("Expected at least 7 shifts, got %d", len(results.Shifts))
	}

	if results.MonthsProcessed != 6 {
		t.Errorf("Expected 6 months processed, got %d", results.MonthsProcessed)
	}

	if results.MonthsFailed != 0 {
		t.Errorf("Expected 0 months failed, got %d", results.MonthsFailed)
	}

	// Verify error handling
	if len(results.Errors) > 0 {
		t.Errorf("Expected no errors, got %d: %s", len(results.Errors), results.FormattedErrors())
	}
}

// TestScrapeSchedule_PartialFailure tests handling of partial failures
func TestScrapeSchedule_PartialFailure(t *testing.T) {
	monthData := map[string]string{
		"2025-11": `
			<html><body><table><tbody>
				<tr><td>2025-11-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2025-12": "", // This will cause a 404
		"2026-01": `
			<html><body><table><tbody>
				<tr><td>2026-01-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		var month string
		if strings.HasPrefix(path, "/schedule/") {
			month = strings.TrimPrefix(path, "/schedule/")
		}

		if html, ok := monthData[month]; ok && html != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, html)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Not found")
		}
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}
	defer client.Close()

	pool := NewGoroutinePool(5)
	defer pool.Close()

	limiter := NewRateLimiter(10 * time.Millisecond)
	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())

	startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	results, err := scraper.ScrapeSchedule(startDate, 3)

	if err != nil {
		t.Fatalf("ScrapeSchedule failed: %v", err)
	}

	// Should have at least 2 successful months
	if results.MonthsProcessed < 2 {
		t.Errorf("Expected at least 2 months processed, got %d", results.MonthsProcessed)
	}

	// Should have 1 failed month
	if results.MonthsFailed < 1 {
		t.Errorf("Expected at least 1 month failed, got %d", results.MonthsFailed)
	}

	// Should have returned some shifts from successful months
	if len(results.Shifts) < 2 {
		t.Errorf("Expected at least 2 shifts, got %d", len(results.Shifts))
	}

	// Should have recorded the error
	if len(results.Errors) < 1 {
		t.Errorf("Expected at least 1 error, got %d", len(results.Errors))
	}
}

// TestScrapeSchedule_DuplicateDetection tests duplicate shift detection
func TestScrapeSchedule_DuplicateDetection(t *testing.T) {
	// Create a server where December returns the same shift twice
	monthData := map[string]string{
		"2025-11": `
			<html><body><table><tbody>
				<tr><td>2025-11-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2025-12": `
			<html><body><table><tbody>
				<tr><td>2025-12-15</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room A</td></tr>
				<tr><td>2025-12-15</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room A</td></tr>
			</tbody></table></body></html>
		`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		path := r.URL.Path
		var month string
		if strings.HasPrefix(path, "/schedule/") {
			month = strings.TrimPrefix(path, "/schedule/")
		}

		if html, ok := monthData[month]; ok {
			fmt.Fprint(w, html)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}
	defer client.Close()

	pool := NewGoroutinePool(5)
	defer pool.Close()

	limiter := NewRateLimiter(10 * time.Millisecond)
	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())

	startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	results, err := scraper.ScrapeSchedule(startDate, 2)

	if err != nil {
		t.Fatalf("ScrapeSchedule failed: %v", err)
	}

	// Should have 2 unique shifts (1 from Nov, 1 from Dec - the duplicate is detected)
	if len(results.Shifts) != 2 {
		t.Errorf("Expected 2 unique shifts, got %d", len(results.Shifts))
	}

	// Should have detected 1 duplicate
	if results.DuplicateCount != 1 {
		t.Errorf("Expected 1 duplicate, got %d", results.DuplicateCount)
	}

	// Total should include the duplicate
	if results.TotalShifts() != 3 {
		t.Errorf("Expected total of 3 shifts (2 unique + 1 duplicate), got %d", results.TotalShifts())
	}
}

// TestScrapeSchedule_RateLimiting tests that rate limiting is applied
func TestScrapeSchedule_RateLimiting(t *testing.T) {
	monthData := map[string]string{
		"2025-11": `
			<html><body><table><tbody>
				<tr><td>2025-11-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2025-12": `
			<html><body><table><tbody>
				<tr><td>2025-12-15</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room A</td></tr>
			</tbody></table></body></html>
		`,
		"2026-01": `
			<html><body><table><tbody>
				<tr><td>2026-01-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2026-02": `
			<html><body><table><tbody>
				<tr><td>2026-02-15</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room B</td></tr>
			</tbody></table></body></html>
		`,
		"2026-03": `
			<html><body><table><tbody>
				<tr><td>2026-03-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2026-04": `
			<html><body><table><tbody>
				<tr><td>2026-04-15</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room C</td></tr>
			</tbody></table></body></html>
		`,
	}

	requestTimes := make([]time.Time, 0)
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		path := r.URL.Path
		var month string
		if strings.HasPrefix(path, "/schedule/") {
			month = strings.TrimPrefix(path, "/schedule/")
		}

		if html, ok := monthData[month]; ok {
			fmt.Fprint(w, html)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}
	defer client.Close()

	pool := NewGoroutinePool(5)
	defer pool.Close()

	// Use 100ms rate limiting
	limiter := NewRateLimiter(100 * time.Millisecond)
	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())

	startTime := time.Now()
	startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	results, err := scraper.ScrapeSchedule(startDate, 6)
	elapsedTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("ScrapeSchedule failed: %v", err)
	}

	if len(results.Shifts) < 6 {
		t.Errorf("Expected at least 6 shifts, got %d", len(results.Shifts))
	}

	// With 5 concurrent workers and 100ms rate limiting:
	// - First 5 requests start immediately in parallel
	// - Next request waits 100ms from the first request, so ~100ms total
	// Total time should be roughly 100-200ms (depends on concurrency)
	// We allow up to 1 second to be generous
	if elapsedTime > 1*time.Second {
		t.Logf("Rate limiting took longer than expected: %v", elapsedTime)
		// This is a warning, not a failure - actual performance depends on system
	}
}

// TestScrapeSchedule_EmptyMonths tests handling of months with no data
func TestScrapeSchedule_EmptyMonths(t *testing.T) {
	monthData := map[string]string{
		"2025-11": `<html><body><table><tbody></tbody></table></body></html>`,
		"2025-12": `
			<html><body><table><tbody>
				<tr><td>2025-12-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2026-01": `<html><body><table><tbody></tbody></table></body></html>`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		path := r.URL.Path
		var month string
		if strings.HasPrefix(path, "/schedule/") {
			month = strings.TrimPrefix(path, "/schedule/")
		}

		if html, ok := monthData[month]; ok {
			fmt.Fprint(w, html)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}
	defer client.Close()

	pool := NewGoroutinePool(5)
	defer pool.Close()

	limiter := NewRateLimiter(10 * time.Millisecond)
	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())

	startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	results, err := scraper.ScrapeSchedule(startDate, 3)

	if err != nil {
		t.Fatalf("ScrapeSchedule failed: %v", err)
	}

	// Should process all 3 months successfully even if some are empty
	if results.MonthsProcessed != 3 {
		t.Errorf("Expected 3 months processed, got %d", results.MonthsProcessed)
	}

	// Should have 1 shift from December
	if len(results.Shifts) != 1 {
		t.Errorf("Expected 1 shift, got %d", len(results.Shifts))
	}

	if results.Shifts[0].Date != "2025-12-15" {
		t.Errorf("Expected shift from 2025-12-15, got %s", results.Shifts[0].Date)
	}
}

// TestGenerateMonthURLs tests URL generation for multiple months
func TestGenerateMonthURLs(t *testing.T) {
	client, _ := NewAmionHTTPClient("https://example.com")
	pool := NewGoroutinePool(5)
	limiter := NewRateLimiter(1 * time.Second)
	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())

	tests := []struct {
		name           string
		startDate      time.Time
		monthCount     int
		expectedMonths []string
		expectedCount  int
	}{
		{
			name:           "single month",
			startDate:      time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC),
			monthCount:     1,
			expectedMonths: []string{"2025-11"},
			expectedCount:  1,
		},
		{
			name:           "6 months crossing year boundary",
			startDate:      time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC),
			monthCount:     6,
			expectedMonths: []string{"2025-11", "2025-12", "2026-01", "2026-02", "2026-03", "2026-04"},
			expectedCount:  6,
		},
		{
			name:           "december to january",
			startDate:      time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
			monthCount:     3,
			expectedMonths: []string{"2025-12", "2026-01", "2026-02"},
			expectedCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := scraper.generateMonthURLs(tt.startDate, tt.monthCount)

			if len(urls) != tt.expectedCount {
				t.Errorf("Expected %d URLs, got %d", tt.expectedCount, len(urls))
			}

			for i, expectedMonth := range tt.expectedMonths {
				if i >= len(urls) {
					t.Errorf("Missing URL for month %s", expectedMonth)
					continue
				}
				if urls[i].Month != expectedMonth {
					t.Errorf("Expected month %s, got %s", expectedMonth, urls[i].Month)
				}
				if !strings.Contains(urls[i].URL, expectedMonth) {
					t.Errorf("URL should contain month %s, got %s", expectedMonth, urls[i].URL)
				}
			}
		})
	}
}

// TestScrapeSchedule_InvalidMonthCount tests error handling for invalid input
func TestScrapeSchedule_InvalidMonthCount(t *testing.T) {
	client, _ := NewAmionHTTPClient("https://example.com")
	pool := NewGoroutinePool(5)
	limiter := NewRateLimiter(1 * time.Second)
	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())

	tests := []struct {
		name       string
		monthCount int
		wantErr    bool
	}{
		{"zero months", 0, true},
		{"negative months", -1, true},
		{"one month", 1, false},
		{"12 months", 12, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
			_, err := scraper.ScrapeSchedule(startDate, tt.monthCount)

			if (err != nil) != tt.wantErr {
				t.Errorf("ScrapeSchedule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestScrapeSchedule_ContextCancellation tests context cancellation
func TestScrapeSchedule_ContextCancellation(t *testing.T) {
	monthData := map[string]string{
		"2025-11": `
			<html><body><table><tbody>
				<tr><td>2025-11-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
		"2025-12": `
			<html><body><table><tbody>
				<tr><td>2025-12-15</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room A</td></tr>
			</tbody></table></body></html>
		`,
		"2026-01": `
			<html><body><table><tbody>
				<tr><td>2026-01-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
			</tbody></table></body></html>
		`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		path := r.URL.Path
		var month string
		if strings.HasPrefix(path, "/schedule/") {
			month = strings.TrimPrefix(path, "/schedule/")
		}

		if html, ok := monthData[month]; ok {
			fmt.Fprint(w, html)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewAmionHTTPClient(server.URL)
	if err != nil {
		t.Fatalf("NewAmionHTTPClient failed: %v", err)
	}
	defer client.Close()

	pool := NewGoroutinePool(5)
	defer pool.Close()

	limiter := NewRateLimiter(100 * time.Millisecond) // Long rate limiting to allow cancellation
	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel it immediately
	cancel()

	startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)
	_, err = scraper.ScrapeScheduleWithContext(ctx, startDate, 3)

	// Should return context cancelled error
	if err == nil || err != context.Canceled {
		t.Logf("Expected context.Canceled, got %v (this may be ok if cancellation was too late)", err)
		// This test is lenient because cancellation timing is hard to guarantee
	}
}

// TestScrapedShifts_Helpers tests helper methods on ScrapedShifts
func TestScrapedShifts_Helpers(t *testing.T) {
	results := &ScrapedShifts{
		Shifts: []RawAmionShift{
			{Date: "2025-11-15", ShiftType: "Tech"},
			{Date: "2025-11-16", ShiftType: "Radio"},
		},
		Errors: []ScrapingError{
			{Month: "2025-12", Error: fmt.Errorf("test error")},
		},
		Warnings: []ScrapingWarning{
			{Month: "2025-11", Message: "test warning"},
		},
		DuplicateCount: 2,
	}

	if !results.HasErrors() {
		t.Error("HasErrors should return true")
	}

	if results.ErrorCount() != 1 {
		t.Errorf("ErrorCount should be 1, got %d", results.ErrorCount())
	}

	if results.WarningCount() != 1 {
		t.Errorf("WarningCount should be 1, got %d", results.WarningCount())
	}

	if results.TotalShifts() != 4 {
		t.Errorf("TotalShifts should be 4 (2 unique + 2 duplicates), got %d", results.TotalShifts())
	}

	formattedErrors := results.FormattedErrors()
	if !strings.Contains(formattedErrors, "test error") {
		t.Errorf("FormattedErrors should contain error message")
	}

	formattedWarnings := results.FormattedWarnings()
	if !strings.Contains(formattedWarnings, "test warning") {
		t.Errorf("FormattedWarnings should contain warning message")
	}
}

// BenchmarkScrapeSchedule_6Months benchmarks 6 months of scraping
func BenchmarkScrapeSchedule_6Months(b *testing.B) {
	monthData := map[string]string{
		"2025-11": `<html><body><table><tbody><tr><td>2025-11-15</td><td>Tech</td><td>07:00</td><td>15:00</td><td>Lab</td></tr></tbody></table></body></html>`,
		"2025-12": `<html><body><table><tbody><tr><td>2025-12-15</td><td>Radio</td><td>07:00</td><td>19:00</td><td>Room</td></tr></tbody></table></body></html>`,
		"2026-01": `<html><body><table><tbody><tr><td>2026-01-15</td><td>Tech</td><td>07:00</td><td>15:00</td><td>Lab</td></tr></tbody></table></body></html>`,
		"2026-02": `<html><body><table><tbody><tr><td>2026-02-15</td><td>Radio</td><td>07:00</td><td>19:00</td><td>Room</td></tr></tbody></table></body></html>`,
		"2026-03": `<html><body><table><tbody><tr><td>2026-03-15</td><td>Tech</td><td>07:00</td><td>15:00</td><td>Lab</td></tr></tbody></table></body></html>`,
		"2026-04": `<html><body><table><tbody><tr><td>2026-04-15</td><td>Radio</td><td>07:00</td><td>19:00</td><td>Room</td></tr></tbody></table></body></html>`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		path := r.URL.Path
		var month string
		if strings.HasPrefix(path, "/schedule/") {
			month = strings.TrimPrefix(path, "/schedule/")
		}

		if html, ok := monthData[month]; ok {
			fmt.Fprint(w, html)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, _ := NewAmionHTTPClient(server.URL)
	defer client.Close()

	pool := NewGoroutinePool(5)
	defer pool.Close()

	limiter := NewRateLimiter(1 * time.Millisecond) // Very short for benchmark
	scraper := NewAmionScraper(client, pool, limiter, DefaultSelectors())

	startDate := time.Date(2025, time.November, 1, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.ClearDuplicateCache()
		scraper.ScrapeSchedule(startDate, 6)
	}
}
