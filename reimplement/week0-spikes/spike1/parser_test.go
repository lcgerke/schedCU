package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestShiftParsing verifies that we can extract shift data from HTML.
func TestShiftParsing(t *testing.T) {
	html := mockAmionHTML()
	parser := NewAmionParser()

	shifts, err := parser.ParseShifts(html)
	require.NoError(t, err)
	require.NotEmpty(t, shifts)

	// Verify we extracted the expected number of shifts
	assert.Greater(t, len(shifts), 0)

	// Verify shift data structure
	for _, shift := range shifts {
		assert.NotEmpty(t, shift.Date)
		assert.NotEmpty(t, shift.Position)
		assert.NotEmpty(t, shift.StartTime)
	}
}

// TestShiftAccuracy verifies parsing accuracy against known values.
func TestShiftAccuracy(t *testing.T) {
	html := mockAmionHTML()
	parser := NewAmionParser()

	shifts, err := parser.ParseShifts(html)
	require.NoError(t, err)
	require.NotEmpty(t, shifts)

	// We should have specific shifts in our mock data
	expectedDate := "2025-11-15"
	var found bool
	for _, shift := range shifts {
		if shift.Date == expectedDate {
			found = true
			assert.Equal(t, "Technologist", shift.Position)
			assert.NotEmpty(t, shift.StartTime)
			break
		}
	}
	assert.True(t, found, "expected shift date %s not found in parsed results", expectedDate)
}

// TestParsingPerformance benchmarks shift parsing performance.
func BenchmarkShiftParsing(b *testing.B) {
	html := mockAmionHTML()
	parser := NewAmionParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseShifts(html)
	}
}

// TestCSSSelectors verifies that CSS selectors reliably find elements.
func TestCSSSelectors(t *testing.T) {
	html := mockAmionHTML()
	doc, err := parseHTML(html)
	require.NoError(t, err)

	// Test shift row selector
	rows := doc.Find("table tbody tr").Nodes
	assert.Greater(t, len(rows), 0, "expected to find shift rows")

	// Test position cell selector
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		position, _ := s.Find("td:nth-child(2)").Html()
		assert.NotEmpty(t, position)
	})
}

// TestErrorHandling verifies parser handles malformed HTML gracefully.
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantErr bool
	}{
		{
			name:    "empty html",
			html:    "",
			wantErr: false, // Should return empty shifts, not error
		},
		{
			name:    "malformed html",
			html:    "<table><tr><td>incomplete",
			wantErr: false, // Should handle gracefully
		},
		{
			name:    "valid html",
			html:    mockAmionHTML(),
			wantErr: false,
		},
	}

	parser := NewAmionParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shifts, err := parser.ParseShifts(tt.html)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, shifts)
			}
		})
	}
}

// TestBatchPerformance simulates parsing 6 months of schedules.
func TestBatchPerformance(t *testing.T) {
	html := mockAmionHTML()
	parser := NewAmionParser()

	start := time.Now()
	totalShifts := 0

	// Simulate 6 months of parsing (30 concurrent pages)
	for i := 0; i < 30; i++ {
		shifts, err := parser.ParseShifts(html)
		require.NoError(t, err)
		totalShifts += len(shifts)
	}

	duration := time.Since(start)
	durationMs := duration.Milliseconds()

	// Log performance metrics
	t.Logf("Parsed 30 pages in %dms", durationMs)
	t.Logf("Average per page: %dms", durationMs/30)
	t.Logf("Total shifts parsed: %d", totalShifts)

	// Assert performance goal: 6 months (30 pages) in <5 seconds
	assert.Less(t, durationMs, int64(5000), "parsing should complete in <5 seconds")
}

// mockAmionHTML returns sample HTML structure similar to Amion's actual page.
func mockAmionHTML() string {
	return `
	<html>
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
					<tr>
						<td>2025-11-15</td>
						<td>Technologist</td>
						<td>07:00</td>
						<td>15:00</td>
						<td>Main Lab</td>
					</tr>
					<tr>
						<td>2025-11-16</td>
						<td>Technologist</td>
						<td>08:00</td>
						<td>16:00</td>
						<td>Main Lab</td>
					</tr>
					<tr>
						<td>2025-11-17</td>
						<td>Radiologist</td>
						<td>07:00</td>
						<td>19:00</td>
						<td>Read Room A</td>
					</tr>
				</tbody>
			</table>
		</div>
	</body>
	</html>
	`
}
