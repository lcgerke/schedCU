package coverage

import (
	"testing"

	"github.com/schedcu/v2/internal/entity"
	"github.com/stretchr/testify/assert"
)

// TestGenerateNaturalLanguageSummary_FullCoverage tests summary generation for full coverage
func TestGenerateNaturalLanguageSummary_FullCoverage(t *testing.T) {
	// Create sample assignments (all shifts fully covered)
	assignments := []entity.Assignment{
		{PersonID: "alice", OriginalShiftType: "ON1"},
		{PersonID: "bob", OriginalShiftType: "ON1"},
		{PersonID: "charlie", OriginalShiftType: "ON2"},
		{PersonID: "diana", OriginalShiftType: "ON2"},
		{PersonID: "eve", OriginalShiftType: "DAY"},
		{PersonID: "frank", OriginalShiftType: "DAY"},
		{PersonID: "grace", OriginalShiftType: "DAY"},
	}

	requirements := map[entity.ShiftType]int{
		"ON1": 2,
		"ON2": 2,
		"DAY": 3,
	}

	// Calculate coverage
	metrics := ResolveCoverage(assignments, requirements)

	// Generate natural language summary
	summary := GenerateNaturalLanguageSummary(metrics, "November 2025 Schedule")

	// Assertions on executive summary
	assert.Contains(t, summary.ExecutiveSummary, "FULLY OPERATIONAL")
	assert.Contains(t, summary.ExecutiveSummary, "100.0%")
	assert.Contains(t, summary.ExecutiveSummary, "Total shift types: 3")

	// Assertions on plain English
	assert.Contains(t, summary.PlainEnglish, "All 3 shift positions")
	assert.Contains(t, summary.PlainEnglish, "schedule is complete")

	// Assertions on detailed summary
	assert.Contains(t, summary.DetailedSummary, "FULLY COVERED SHIFTS")

	// Assertions on recommendations
	assert.NotEmpty(t, summary.Recommendations)
	assert.Contains(t, summary.Recommendations[0], "production-ready")

	// Assertions on key insights
	assert.NotEmpty(t, summary.KeyInsights)
	assert.Contains(t, summary.KeyInsights[0], "100% coverage")

	// Print for manual inspection
	t.Logf("\n%s\n", summary.FormatAsText())
}

// TestGenerateNaturalLanguageSummary_PartialCoverage tests summary with gaps
func TestGenerateNaturalLanguageSummary_PartialCoverage(t *testing.T) {
	// Create sample assignments (partial coverage)
	assignments := []entity.Assignment{
		{PersonID: "alice", OriginalShiftType: "ON1"}, // Need 2, have 1
		// ON2 completely uncovered (need 2, have 0)
		{PersonID: "charlie", OriginalShiftType: "DAY"},
		{PersonID: "diana", OriginalShiftType: "DAY"}, // Need 3, have 2
	}

	requirements := map[entity.ShiftType]int{
		"ON1": 2,
		"ON2": 2,
		"DAY": 3,
	}

	// Calculate coverage
	metrics := ResolveCoverage(assignments, requirements)

	// Generate natural language summary
	summary := GenerateNaturalLanguageSummary(metrics, "November 2025 Schedule")

	// Assertions on executive summary
	assert.Contains(t, summary.ExecutiveSummary, "OPERATIONAL WITH GAPS")
	assert.Contains(t, summary.ExecutiveSummary, "Under-staffed: 2") // ON1 and DAY partial, ON2 uncovered

	// Assertions on plain English
	assert.Contains(t, summary.PlainEnglish, "some still need staff assigned")

	// Assertions on detailed summary
	assert.Contains(t, summary.DetailedSummary, "PARTIALLY COVERED SHIFTS")
	assert.Contains(t, summary.DetailedSummary, "UNCOVERED SHIFTS")

	// Assertions on recommendations
	assert.NotEmpty(t, summary.Recommendations)
	// Should recommend filling uncovered shift
	foundUncoveredRecommendation := false
	for _, rec := range summary.Recommendations {
		if assert.Contains(t, rec, "uncovered") || assert.Contains(t, rec, "ON2") {
			foundUncoveredRecommendation = true
			break
		}
	}
	assert.True(t, foundUncoveredRecommendation, "Should recommend filling uncovered shifts")

	// Print for manual inspection
	t.Logf("\n%s\n", summary.FormatAsText())
}

// TestGenerateNaturalLanguageSummary_CriticalGaps tests summary with critical gaps
func TestGenerateNaturalLanguageSummary_CriticalGaps(t *testing.T) {
	// Create sample assignments (minimal coverage, mostly empty)
	assignments := []entity.Assignment{
		{PersonID: "alice", OriginalShiftType: "DAY"}, // Only 1 of 3 shifts covered
	}

	requirements := map[entity.ShiftType]int{
		"ON1": 2, // Uncovered
		"ON2": 2, // Uncovered
		"DAY": 3, // Partially covered (1/3)
	}

	// Calculate coverage
	metrics := ResolveCoverage(assignments, requirements)

	// Generate natural language summary
	summary := GenerateNaturalLanguageSummary(metrics, "November 2025 Schedule")

	// Assertions on executive summary
	assert.Contains(t, summary.ExecutiveSummary, "CRITICAL GAPS")
	assert.Contains(t, summary.ExecutiveSummary, "Immediate action required")

	// Overall coverage should be low
	assert.Less(t, metrics.OverallCoveragePercentage, 50.0)

	// Assertions on plain English
	assert.Contains(t, summary.PlainEnglish, "needs more staff")

	// Assertions on key insights
	assert.NotEmpty(t, summary.KeyInsights)
	// Should mention critical staffing shortfall
	assert.Contains(t, summary.KeyInsights[0], "critical staffing")

	// Print for manual inspection
	t.Logf("\n%s\n", summary.FormatAsText())
}

// TestGenerateNaturalLanguageSummary_OverStaffing tests summary with over-staffed shifts
func TestGenerateNaturalLanguageSummary_OverStaffing(t *testing.T) {
	// Create sample assignments (some shifts over-staffed)
	assignments := []entity.Assignment{
		{PersonID: "alice", OriginalShiftType: "ON1"},
		{PersonID: "bob", OriginalShiftType: "ON1"},
		{PersonID: "charlie", OriginalShiftType: "ON1"}, // 3 assigned, only need 2 (over-staffed)
		{PersonID: "diana", OriginalShiftType: "ON2"},
		{PersonID: "eve", OriginalShiftType: "ON2"}, // Fully covered
		{PersonID: "frank", OriginalShiftType: "DAY"},
		{PersonID: "grace", OriginalShiftType: "DAY"},
		{PersonID: "henry", OriginalShiftType: "DAY"}, // Fully covered
	}

	requirements := map[entity.ShiftType]int{
		"ON1": 2,
		"ON2": 2,
		"DAY": 3,
	}

	// Calculate coverage
	metrics := ResolveCoverage(assignments, requirements)

	// Generate natural language summary
	summary := GenerateNaturalLanguageSummary(metrics, "November 2025 Schedule")

	// Should still be fully operational
	assert.Contains(t, summary.ExecutiveSummary, "FULLY OPERATIONAL")

	// Should mention over-staffing
	assert.Contains(t, summary.ExecutiveSummary, "Over-staffed: 1")

	// Should have detailed info on over-staffed shift
	assert.Contains(t, summary.DetailedSummary, "OVER-STAFFED SHIFTS")
	assert.Contains(t, summary.DetailedSummary, "ON1")
	assert.Contains(t, summary.DetailedSummary, "+1 extra")

	// Recommendations should mention redistribution
	foundRedistributeRecommendation := false
	for _, rec := range summary.Recommendations {
		if assert.Contains(t, rec, "redistributing") || assert.Contains(t, rec, "over-staffed") {
			foundRedistributeRecommendation = true
			break
		}
	}
	assert.True(t, foundRedistributeRecommendation, "Should recommend redistributing over-staffed shifts")

	// Print for manual inspection
	t.Logf("\n%s\n", summary.FormatAsText())
}

// TestFormatShiftTypeDescription tests shift type descriptions
func TestFormatShiftTypeDescription(t *testing.T) {
	tests := []struct {
		shiftType   entity.ShiftType
		shouldContain string
	}{
		{"ON1", "Overnight Shift 1"},
		{"ON2", "Overnight Shift 2"},
		{"MidC", "Mid-level Coverage"},
		{"DAY", "Day Shift"},
		{"CUSTOM", "CUSTOM"}, // Unknown shift type returns the type itself
	}

	for _, tt := range tests {
		t.Run(string(tt.shiftType), func(t *testing.T) {
			desc := FormatShiftTypeDescription(tt.shiftType)
			assert.Contains(t, desc, tt.shouldContain)
		})
	}
}

// TestNaturalLanguageSummary_FormatAsText tests text formatting
func TestNaturalLanguageSummary_FormatAsText(t *testing.T) {
	// Create a simple summary
	assignments := []entity.Assignment{
		{PersonID: "alice", OriginalShiftType: "DAY"},
		{PersonID: "bob", OriginalShiftType: "DAY"},
	}

	requirements := map[entity.ShiftType]int{
		"DAY": 2,
	}

	metrics := ResolveCoverage(assignments, requirements)
	summary := GenerateNaturalLanguageSummary(metrics, "Test Schedule")

	// Format as text
	text := summary.FormatAsText()

	// Should contain all major sections
	assert.Contains(t, text, "EXECUTIVE COVERAGE SUMMARY")
	assert.Contains(t, text, "PLAIN ENGLISH SUMMARY")
	assert.Contains(t, text, "DETAILED COVERAGE BREAKDOWN")
	assert.Contains(t, text, "KEY INSIGHTS")
	assert.Contains(t, text, "RECOMMENDATIONS")

	// Should have separators
	assert.Contains(t, text, "──────")

	// Should be readable
	assert.NotEmpty(t, text)
	assert.Greater(t, len(text), 100) // Should be substantial
}
