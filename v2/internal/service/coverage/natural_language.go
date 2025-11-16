// Package coverage provides natural language summary generation for coverage metrics
package coverage

import (
	"fmt"
	"strings"

	"github.com/schedcu/v2/internal/entity"
)

// NaturalLanguageSummary provides human-readable descriptions of coverage metrics
type NaturalLanguageSummary struct {
	ExecutiveSummary string   // High-level overview for executives
	PlainEnglish     string   // Non-technical explanation
	DetailedSummary  string   // Technical details
	Recommendations  []string // Actionable recommendations
	KeyInsights      []string // Important observations
}

// GenerateNaturalLanguageSummary creates a comprehensive human-readable summary
// from coverage metrics suitable for administrators, schedulers, and stakeholders.
//
// The summary includes:
// - Executive overview (for administrators)
// - Plain English explanation (for non-technical readers)
// - Detailed technical summary (for schedulers)
// - Recommendations and next steps
// - Key insights and talking points
//
// Example usage:
//
//	metrics := ResolveCoverage(assignments, requirements)
//	summary := GenerateNaturalLanguageSummary(metrics, "November 2025 Schedule")
//	fmt.Println(summary.PlainEnglish)
func GenerateNaturalLanguageSummary(metrics CoverageMetrics, scheduleName string) NaturalLanguageSummary {
	summary := NaturalLanguageSummary{
		Recommendations: []string{},
		KeyInsights:     []string{},
	}

	// Generate executive summary
	summary.ExecutiveSummary = generateExecutiveSummary(metrics, scheduleName)

	// Generate plain English explanation
	summary.PlainEnglish = generatePlainEnglishSummary(metrics, scheduleName)

	// Generate detailed technical summary
	summary.DetailedSummary = generateDetailedSummary(metrics)

	// Generate recommendations
	summary.Recommendations = generateRecommendations(metrics)

	// Generate key insights
	summary.KeyInsights = generateKeyInsights(metrics)

	return summary
}

// generateExecutiveSummary creates a high-level overview for administrators
func generateExecutiveSummary(metrics CoverageMetrics, scheduleName string) string {
	var lines []string

	lines = append(lines, fmt.Sprintf("EXECUTIVE COVERAGE SUMMARY: %s", scheduleName))
	lines = append(lines, "")

	// Overall status
	if metrics.OverallCoveragePercentage >= 100.0 {
		lines = append(lines, "✅ COVERAGE STATUS: FULLY OPERATIONAL")
		lines = append(lines, "")
		lines = append(lines, "All shifts are fully staffed with qualified personnel.")
		lines = append(lines, "No coverage gaps detected - schedule is production-ready.")
	} else if metrics.OverallCoveragePercentage >= 80.0 {
		lines = append(lines, "⚠️  COVERAGE STATUS: OPERATIONAL WITH GAPS")
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Overall coverage: %.1f%% (target: 100%%)", metrics.OverallCoveragePercentage))
		lines = append(lines, fmt.Sprintf("%d shift(s) require additional staffing.", len(metrics.UnderStaffedShifts)))
	} else {
		lines = append(lines, "❌ COVERAGE STATUS: CRITICAL GAPS")
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Overall coverage: %.1f%% (target: 100%%)", metrics.OverallCoveragePercentage))
		lines = append(lines, "Immediate action required to fill staffing gaps.")
	}

	lines = append(lines, "")

	// Key metrics
	totalShifts := len(metrics.CoverageByShiftType)
	fullyStaffed := totalShifts - len(metrics.UnderStaffedShifts)

	lines = append(lines, "KEY METRICS:")
	lines = append(lines, fmt.Sprintf("• Total shift types: %d", totalShifts))
	lines = append(lines, fmt.Sprintf("• Fully staffed: %d", fullyStaffed))
	lines = append(lines, fmt.Sprintf("• Under-staffed: %d", len(metrics.UnderStaffedShifts)))
	lines = append(lines, fmt.Sprintf("• Over-staffed: %d", len(metrics.OverStaffedShifts)))
	lines = append(lines, fmt.Sprintf("• Overall coverage: %.1f%%", metrics.OverallCoveragePercentage))

	return strings.Join(lines, "\n")
}

// generatePlainEnglishSummary creates a non-technical explanation
func generatePlainEnglishSummary(metrics CoverageMetrics, scheduleName string) string {
	var lines []string

	lines = append(lines, fmt.Sprintf("PLAIN ENGLISH SUMMARY: %s", scheduleName))
	lines = append(lines, "")

	totalShifts := len(metrics.CoverageByShiftType)

	if metrics.OverallCoveragePercentage >= 100.0 {
		lines = append(lines, "WHAT THIS MEANS:")
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("All %d shift positions in the schedule have assigned staff members.", totalShifts))
		lines = append(lines, "Every shift that needs coverage has someone scheduled to work it.")
		lines = append(lines, "There are no empty shifts or gaps in the schedule.")
		lines = append(lines, "")
		lines = append(lines, "BOTTOM LINE:")
		lines = append(lines, "The schedule is complete and ready to use. All positions are filled.")
	} else {
		uncoveredCount := 0
		partialCount := 0
		for _, detail := range metrics.CoverageByShiftType {
			if detail.Status == StatusUncovered {
				uncoveredCount++
			} else if detail.Status == StatusPartial {
				partialCount++
			}
		}

		lines = append(lines, "WHAT THIS MEANS:")
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Out of %d shift positions, some still need staff assigned:", totalShifts))

		if uncoveredCount > 0 {
			lines = append(lines, fmt.Sprintf("• %d shift(s) have NO staff assigned (completely empty)", uncoveredCount))
		}
		if partialCount > 0 {
			lines = append(lines, fmt.Sprintf("• %d shift(s) have SOME staff but need more", partialCount))
		}

		lines = append(lines, "")
		lines = append(lines, "BOTTOM LINE:")
		lines = append(lines, "The schedule needs more staff assignments before it can be used.")
		lines = append(lines, "See the detailed summary below for which specific shifts need coverage.")
	}

	return strings.Join(lines, "\n")
}

// generateDetailedSummary creates a technical breakdown
func generateDetailedSummary(metrics CoverageMetrics) string {
	var lines []string

	lines = append(lines, "DETAILED COVERAGE BREAKDOWN:")
	lines = append(lines, "")

	// Fully covered shifts
	lines = append(lines, "✅ FULLY COVERED SHIFTS:")
	fullyStaffedCount := 0
	for shiftType, detail := range metrics.CoverageByShiftType {
		if detail.Status == StatusFull {
			fullyStaffedCount++
			lines = append(lines, fmt.Sprintf("   • %s: %d/%d staff (%.1f%%)",
				shiftType, detail.Assigned, detail.Required, detail.CoveragePercentage))
		}
	}
	if fullyStaffedCount == 0 {
		lines = append(lines, "   (none)")
	}
	lines = append(lines, "")

	// Partially covered shifts
	lines = append(lines, "⚠️  PARTIALLY COVERED SHIFTS:")
	partialCount := 0
	for shiftType, detail := range metrics.CoverageByShiftType {
		if detail.Status == StatusPartial {
			partialCount++
			missing := detail.Required - detail.Assigned
			lines = append(lines, fmt.Sprintf("   • %s: %d/%d staff (%.1f%%) - NEEDS %d MORE",
				shiftType, detail.Assigned, detail.Required, detail.CoveragePercentage, missing))
		}
	}
	if partialCount == 0 {
		lines = append(lines, "   (none)")
	}
	lines = append(lines, "")

	// Uncovered shifts
	lines = append(lines, "❌ UNCOVERED SHIFTS:")
	uncoveredCount := 0
	for shiftType, detail := range metrics.CoverageByShiftType {
		if detail.Status == StatusUncovered {
			uncoveredCount++
			lines = append(lines, fmt.Sprintf("   • %s: 0/%d staff - NEEDS %d STAFF",
				shiftType, detail.Required, detail.Required))
		}
	}
	if uncoveredCount == 0 {
		lines = append(lines, "   (none)")
	}
	lines = append(lines, "")

	// Over-staffed shifts
	if len(metrics.OverStaffedShifts) > 0 {
		lines = append(lines, "ℹ️  OVER-STAFFED SHIFTS:")
		for shiftType, detail := range metrics.CoverageByShiftType {
			for _, overStaffed := range metrics.OverStaffedShifts {
				if shiftType == overStaffed {
					extra := detail.Assigned - detail.Required
					lines = append(lines, fmt.Sprintf("   • %s: %d/%d staff (+%d extra)",
						shiftType, detail.Assigned, detail.Required, extra))
				}
			}
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// generateRecommendations creates actionable recommendations
func generateRecommendations(metrics CoverageMetrics) []string {
	recommendations := []string{}

	if metrics.OverallCoveragePercentage >= 100.0 {
		recommendations = append(recommendations, "Schedule is production-ready and can be implemented immediately")
		recommendations = append(recommendations, "Distribute schedule to all staff members")
		recommendations = append(recommendations, "Monitor utilization and adjust as needed")

		if len(metrics.OverStaffedShifts) > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Consider redistributing %d over-staffed shift(s) to optimize coverage", len(metrics.OverStaffedShifts)))
		}
	} else {
		// Critical gaps
		uncoveredCount := 0
		for _, detail := range metrics.CoverageByShiftType {
			if detail.Status == StatusUncovered {
				uncoveredCount++
			}
		}

		if uncoveredCount > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("URGENT: Fill %d completely uncovered shift(s) before schedule activation", uncoveredCount))
		}

		// Partial gaps
		partialCount := len(metrics.UnderStaffedShifts) - uncoveredCount
		if partialCount > 0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Address %d partially covered shift(s) to meet staffing requirements", partialCount))
		}

		// Specific shift recommendations
		for shiftType, detail := range metrics.CoverageByShiftType {
			if detail.Status == StatusUncovered {
				recommendations = append(recommendations,
					fmt.Sprintf("Assign %d staff to %s shift (currently uncovered)", detail.Required, shiftType))
			} else if detail.Status == StatusPartial {
				missing := detail.Required - detail.Assigned
				recommendations = append(recommendations,
					fmt.Sprintf("Add %d more staff to %s shift (currently at %.1f%%)", missing, shiftType, detail.CoveragePercentage))
			}
		}

		recommendations = append(recommendations, "Re-validate schedule after making assignments")
	}

	return recommendations
}

// generateKeyInsights creates important observations
func generateKeyInsights(metrics CoverageMetrics) []string {
	insights := []string{}

	totalShifts := len(metrics.CoverageByShiftType)
	fullyStaffed := totalShifts - len(metrics.UnderStaffedShifts)

	// Coverage percentage insight
	if metrics.OverallCoveragePercentage >= 100.0 {
		insights = append(insights, "100% coverage achieved - all shifts fully staffed")
	} else if metrics.OverallCoveragePercentage >= 90.0 {
		insights = append(insights, fmt.Sprintf("%.1f%% coverage - near complete, minor gaps remain", metrics.OverallCoveragePercentage))
	} else if metrics.OverallCoveragePercentage >= 75.0 {
		insights = append(insights, fmt.Sprintf("%.1f%% coverage - substantial gaps require attention", metrics.OverallCoveragePercentage))
	} else {
		insights = append(insights, fmt.Sprintf("%.1f%% coverage - critical staffing shortfall", metrics.OverallCoveragePercentage))
	}

	// Shift coverage distribution
	insights = append(insights, fmt.Sprintf("%d of %d shift types fully covered", fullyStaffed, totalShifts))

	// Over-staffing insight
	if len(metrics.OverStaffedShifts) > 0 {
		insights = append(insights,
			fmt.Sprintf("%d shift(s) over-staffed - consider rebalancing resources", len(metrics.OverStaffedShifts)))
	}

	// Under-staffing severity
	if len(metrics.UnderStaffedShifts) > 0 {
		totalMissing := 0
		for _, detail := range metrics.CoverageByShiftType {
			if detail.Assigned < detail.Required {
				totalMissing += (detail.Required - detail.Assigned)
			}
		}
		insights = append(insights,
			fmt.Sprintf("Total staffing shortage: %d positions across %d shifts", totalMissing, len(metrics.UnderStaffedShifts)))
	}

	// Balance insight
	if len(metrics.OverStaffedShifts) > 0 && len(metrics.UnderStaffedShifts) > 0 {
		insights = append(insights, "Opportunity: Redistribute over-staffed positions to under-staffed shifts")
	}

	return insights
}

// FormatAsText returns the complete summary as formatted text
func (nls NaturalLanguageSummary) FormatAsText() string {
	var sections []string

	sections = append(sections, nls.ExecutiveSummary)
	sections = append(sections, "")
	sections = append(sections, strings.Repeat("─", 70))
	sections = append(sections, "")

	sections = append(sections, nls.PlainEnglish)
	sections = append(sections, "")
	sections = append(sections, strings.Repeat("─", 70))
	sections = append(sections, "")

	sections = append(sections, nls.DetailedSummary)

	if len(nls.KeyInsights) > 0 {
		sections = append(sections, strings.Repeat("─", 70))
		sections = append(sections, "")
		sections = append(sections, "KEY INSIGHTS:")
		sections = append(sections, "")
		for i, insight := range nls.KeyInsights {
			sections = append(sections, fmt.Sprintf("%d. %s", i+1, insight))
		}
		sections = append(sections, "")
	}

	if len(nls.Recommendations) > 0 {
		sections = append(sections, strings.Repeat("─", 70))
		sections = append(sections, "")
		sections = append(sections, "RECOMMENDATIONS:")
		sections = append(sections, "")
		for i, rec := range nls.Recommendations {
			sections = append(sections, fmt.Sprintf("%d. %s", i+1, rec))
		}
		sections = append(sections, "")
	}

	return strings.Join(sections, "\n")
}

// FormatShiftTypeDescription returns a human-readable description of a shift type
func FormatShiftTypeDescription(shiftType entity.ShiftType) string {
	descriptions := map[entity.ShiftType]string{
		"ON1":  "Overnight Shift 1 (primary overnight coverage)",
		"ON2":  "Overnight Shift 2 (backup overnight coverage)",
		"MidC": "Mid-level Coverage (evening hours)",
		"MidL": "Mid-level Late (late evening)",
		"DAY":  "Day Shift (standard daytime hours)",
	}

	if desc, exists := descriptions[shiftType]; exists {
		return desc
	}

	return string(shiftType)
}
