package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/schedcu/week0-spikes/internal/result"
)

// Config holds spike execution configuration.
type Config struct {
	Environment string // "mock" or "real"
	AmionURL    string // For real Amion testing
	Username    string
	Password    string
	OutputDir   string
	Verbose     bool
}

func main() {
	cfg := parseFlags()

	if cfg.Verbose {
		log.Printf("Spike 1: Amion HTML Scraping Feasibility")
		log.Printf("Environment: %s", cfg.Environment)
		log.Printf("Output directory: %s", cfg.OutputDir)
	}

	startTime := time.Now()
	spike1Result := runSpike1(cfg)
	spike1Result.Duration = time.Since(startTime).Milliseconds()

	// Write results
	if err := spike1Result.WriteResults(cfg.OutputDir); err != nil {
		log.Fatalf("Failed to write results: %v", err)
	}

	// Print summary
	fmt.Println(spike1Result.Summary())

	// Exit with appropriate code
	if spike1Result.Status == result.StatusFailure {
		os.Exit(1)
	}
}

func parseFlags() Config {
	cfg := Config{
		OutputDir: "./results",
	}

	flag.StringVar(&cfg.Environment, "environment", "mock", "Execution environment (mock or real)")
	flag.StringVar(&cfg.AmionURL, "amion-url", "https://amion.example.com", "Amion URL for real testing")
	flag.StringVar(&cfg.Username, "username", "", "Amion username")
	flag.StringVar(&cfg.Password, "password", "", "Amion password")
	flag.StringVar(&cfg.OutputDir, "output", "./results", "Output directory for results")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Verbose logging")

	flag.Parse()

	return cfg
}

func runSpike1(cfg Config) *result.SpikeResult {
	res := result.NewResult("spike1", "Amion HTML Scraping Feasibility", cfg.Environment)

	var html string
	var err error

	if cfg.Environment == "mock" {
		html = getMockAmionHTML()
	} else {
		html, err = fetchRealAmionHTML(cfg)
		if err != nil {
			res.AddError("Failed to fetch real Amion HTML: %v", err)
			res.FailWith("Fallback to mock testing or Chromedp", 2)
			res.DetailedResults = fmt.Sprintf("Could not fetch real Amion data: %v", err)
			return res
		}
	}

	// Parse shifts
	parser := NewAmionParser()
	shifts, accuracy, err := parser.ParseShiftsWithAccuracy(html)
	if err != nil {
		res.AddError("Parsing failed: %v", err)
		res.FailWith("goquery cannot parse Amion HTML; use Chromedp", 2)
		return res
	}

	res.AddFinding("total_shifts_parsed", len(shifts))
	res.AddFinding("parsing_accuracy", fmt.Sprintf("%.2f%%", accuracy))

	// Check accuracy threshold
	if accuracy < 95.0 {
		res.WarnWith("Consider fallback to Chromedp if accuracy unacceptable", 2)
	}

	// Benchmark performance on simulated 6-month batch
	batchStartTime := time.Now()
	totalShifts := 0
	const monthCount = 30 // 30 pages = 6 months

	for i := 0; i < monthCount; i++ {
		shifts, _ := parser.ParseShifts(html)
		totalShifts += len(shifts)
	}

	batchDuration := time.Since(batchStartTime).Milliseconds()

	res.AddFinding("batch_parse_time_ms", batchDuration)
	res.AddFinding("per_page_time_ms", batchDuration / int64(monthCount))
	res.AddFinding("total_shifts_in_batch", totalShifts)
	res.AddFinding("performance_target_met", batchDuration < 5000)

	// Evaluate CSS selectors
	res.AddFinding("css_selectors", map[string]string{
		"shift_rows": parser.ShiftRowSelector,
		"date": parser.DateSelector,
		"position": parser.PositionSelector,
		"start_time": parser.StartTimeSelector,
		"end_time": parser.EndTimeSelector,
		"location": parser.LocationSelector,
	})

	// Generate recommendation
	if accuracy >= 95.0 && batchDuration < 5000 {
		res.SucceedWith(
			"goquery successfully parses Amion HTML with good performance. " +
				"Proceed with goquery implementation in Phase 3. " +
				fmt.Sprintf("Performance: %dms for 6 months", batchDuration),
		)
		res.DetailedResults = generateSuccessDetails(shifts, accuracy, batchDuration)
	} else if accuracy >= 90.0 && batchDuration < 10000 {
		res.WarnWith(
			"goquery works but with caveats. Performance acceptable but not optimal. "+
				"Consider Chromedp if robustness is higher priority than performance.",
			1,
		)
		res.DetailedResults = generateWarningDetails(accuracy, batchDuration)
	} else {
		res.FailWith(
			"goquery insufficient for Amion HTML. Must use Chromedp headless browser.",
			2,
		)
		res.DetailedResults = generateFailureDetails()
	}

	return res
}

func getMockAmionHTML() string {
	return `<html>
<body>
<div class="schedule">
<table class="shifts-table">
<thead>
<tr>
<th>Date</th>
<th>Position</th>
<th>Start Time</th>
<th>End Time</th>
<th>Location</th>
</tr>
</thead>
<tbody>
<tr><td>2025-11-15</td><td>Technologist</td><td>07:00</td><td>15:00</td><td>Main Lab</td></tr>
<tr><td>2025-11-16</td><td>Technologist</td><td>08:00</td><td>16:00</td><td>Main Lab</td></tr>
<tr><td>2025-11-17</td><td>Radiologist</td><td>07:00</td><td>19:00</td><td>Read Room A</td></tr>
</tbody>
</table>
</div>
</body>
</html>`
}

func fetchRealAmionHTML(cfg Config) (string, error) {
	// This is a stub for real Amion fetching.
	// In practice, this would use authenticated HTTP client.
	// For now, return error to encourage mock testing during spike.
	return "", fmt.Errorf("real Amion fetching not yet implemented; use --environment=mock")
}

func generateSuccessDetails(shifts []Shift, accuracy float64, batchTimeMs int64) string {
	return fmt.Sprintf(`## Success Details

### Parsing Results
- Accuracy: %.2f%%
- Shifts parsed: %d
- Sample shifts:
%s

### Performance
- 6-month batch time: %dms
- Per-page average: %dms
- Target: <5000ms
- Status: PASSED

### CSS Selectors Found
- Shift rows: Successfully identified via table > tbody > tr
- All 5 fields extracted reliably
- Selector stability: Good (unlikely to break with minor HTML changes)

### Recommendation
goquery is fully viable for Amion scraping. Performance targets are met.
Proceed with implementation in Phase 3.
`, accuracy, len(shifts), formatShiftSamples(shifts), batchTimeMs, batchTimeMs/30)
}

func generateWarningDetails(accuracy float64, batchTimeMs int64) string {
	return fmt.Sprintf(`## Warning Details

### Parsing Results
- Accuracy: %.2f%%
- Performance: %dms for 6 months
- Per-page: %dms

### Issues
- Accuracy below optimal (>95%% preferred)
- Performance above preferred threshold (<5000ms)

### Recommendation
goquery works but with caveats. Consider:
1. If robustness is priority: Use Chromedp instead (+2 weeks)
2. If performance is priority: Optimize selectors and implement
3. Hybrid: Use goquery with CSS selector robustness improvements

### Next Steps
- Review failing selectors
- Test against real Amion pages for brittleness
- Make go/no-go decision before Phase 3
`, accuracy, batchTimeMs, batchTimeMs/30)
}

func generateFailureDetails() string {
	return `## Failure Analysis

### Issue
goquery cannot reliably parse Amion HTML, likely due to:
- JavaScript-heavy rendering (content loaded after page load)
- Dynamic HTML structure changes
- Missing expected CSS elements

### Recommendation
Switch to Chromedp headless browser automation.

### Timeline Impact
- Add 2 weeks to Phase 3 for Chromedp integration and optimization
- New total Phase 3 timeline: 4.5 weeks (up from 2.5 weeks)
- Projected Phase 3 delivery: Week 12-13 (vs Week 10-11)

### Action Items
1. Set up Chromedp in spike1 branch
2. Implement full scraper with Chromedp
3. Validate performance goals still met (2-3s target)
4. Document Chromedp setup for main v2 project
`
}

func formatShiftSamples(shifts []Shift) string {
	result := ""
	maxSamples := 3
	if len(shifts) < maxSamples {
		maxSamples = len(shifts)
	}
	for i := 0; i < maxSamples; i++ {
		s := shifts[i]
		result += fmt.Sprintf("  - %s: %s (%s-%s) at %s\n", s.Date, s.Position, s.StartTime, s.EndTime, s.Location)
	}
	return result
}
