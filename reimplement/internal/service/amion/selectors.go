package amion

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// AmionSelectors contains all CSS selectors for extracting data from Amion HTML.
// These selectors are based on Spike 1 testing and are validated to work
// with the current Amion HTML structure.
type AmionSelectors struct {
	// ShiftTableSelector selects the main table containing all shifts
	ShiftTableSelector string

	// ShiftRowSelector selects individual shift rows within the table body
	ShiftRowSelector string

	// DateCellSelector selects the date column (column 1)
	DateCellSelector string

	// ShiftTypeCellSelector selects the shift type/position column (column 2)
	ShiftTypeCellSelector string

	// StartTimeCellSelector selects the start time column (column 3)
	StartTimeCellSelector string

	// EndTimeCellSelector selects the end time column (column 4)
	EndTimeCellSelector string

	// LocationCellSelector selects the location column (column 5)
	LocationCellSelector string

	// RequiredStaffingCellSelector selects the required staffing column (column 6, optional)
	RequiredStaffingCellSelector string

	// HeaderRowSelector identifies header rows to skip
	HeaderRowSelector string
}

// DefaultSelectors returns the default Amion HTML selectors based on Spike 1 results.
// These selectors are CSS nth-child based and work reliably with standard Amion HTML.
func DefaultSelectors() *AmionSelectors {
	return &AmionSelectors{
		ShiftTableSelector:          "table",
		ShiftRowSelector:            "table tbody tr",
		DateCellSelector:            "td:nth-child(1)",
		ShiftTypeCellSelector:       "td:nth-child(2)",
		StartTimeCellSelector:       "td:nth-child(3)",
		EndTimeCellSelector:         "td:nth-child(4)",
		LocationCellSelector:        "td:nth-child(5)",
		RequiredStaffingCellSelector: "td:nth-child(6)",
		HeaderRowSelector:           "thead tr, tr:has(th)",
	}
}

// ExtractShifts extracts all shifts from a goquery document.
// It returns all successfully extracted shifts plus any errors encountered.
// Does not fail on individual extraction errors - collects and reports them.
func ExtractShifts(doc *goquery.Document) *ExtractionResult {
	return ExtractShiftsWithSelectors(doc, DefaultSelectors())
}

// ExtractShiftsWithSelectors extracts shifts using custom selectors.
// Useful for testing with non-standard HTML structures.
func ExtractShiftsWithSelectors(doc *goquery.Document, sel *AmionSelectors) *ExtractionResult {
	result := &ExtractionResult{
		Shifts: make([]RawAmionShift, 0),
		Errors: make([]ExtractionError, 0),
	}

	rowIndex := 0

	// Select all shift rows
	doc.Find(sel.ShiftRowSelector).Each(func(i int, row *goquery.Selection) {
		rowIndex++

		// Skip header rows
		if row.Find("th").Length() > 0 {
			return
		}

		// Check if row is empty or contains only whitespace
		text := strings.TrimSpace(row.Text())
		if text == "" {
			return
		}

		shift := extractShiftFromRow(row, rowIndex, sel, result)
		if shift != nil {
			result.Shifts = append(result.Shifts, *shift)
		}
	})

	return result
}

// ExtractShiftsForMonth extracts shifts for a specific month from a document.
// The month parameter should be in YYYY-MM format (e.g., "2025-11").
// Returns all shifts matching the specified month.
func ExtractShiftsForMonth(doc *goquery.Document, monthStr string) *ExtractionResult {
	allShifts := ExtractShifts(doc)

	// Filter shifts by month
	filtered := &ExtractionResult{
		Shifts: make([]RawAmionShift, 0),
		Errors: allShifts.Errors, // Keep all errors from extraction
	}

	for _, shift := range allShifts.Shifts {
		// Check if shift date starts with month string (YYYY-MM)
		if strings.HasPrefix(shift.Date, monthStr) {
			filtered.Shifts = append(filtered.Shifts, shift)
		}
	}

	return filtered
}

// extractShiftFromRow extracts a single shift from a table row.
// Returns nil if critical fields are missing; errors are recorded in result.
func extractShiftFromRow(row *goquery.Selection, rowIndex int, sel *AmionSelectors, result *ExtractionResult) *RawAmionShift {
	shift := &RawAmionShift{
		RowIndex: rowIndex,
	}

	// Extract date (required field)
	dateText := strings.TrimSpace(row.Find(sel.DateCellSelector).Text())
	if dateText == "" {
		result.Errors = append(result.Errors, ExtractionError{
			RowIndex: rowIndex,
			Field:    "date",
			Value:    dateText,
			Reason:   "empty or missing date cell",
		})
		return nil
	}
	shift.Date = dateText
	shift.DateCell = fmt.Sprintf("row %d, column 1", rowIndex)

	// Extract shift type/position (required field)
	shiftTypeText := strings.TrimSpace(row.Find(sel.ShiftTypeCellSelector).Text())
	if shiftTypeText == "" {
		result.Errors = append(result.Errors, ExtractionError{
			RowIndex: rowIndex,
			Field:    "shift_type",
			Value:    shiftTypeText,
			Reason:   "empty or missing shift type cell",
		})
		return nil
	}
	shift.ShiftType = shiftTypeText
	shift.ShiftTypeCell = fmt.Sprintf("row %d, column 2", rowIndex)

	// Extract start time (required field)
	startTimeText := strings.TrimSpace(row.Find(sel.StartTimeCellSelector).Text())
	if startTimeText == "" {
		result.Errors = append(result.Errors, ExtractionError{
			RowIndex: rowIndex,
			Field:    "start_time",
			Value:    startTimeText,
			Reason:   "empty or missing start time cell",
		})
		return nil
	}
	shift.StartTime = startTimeText
	shift.StartTimeCell = fmt.Sprintf("row %d, column 3", rowIndex)

	// Extract end time (required field)
	endTimeText := strings.TrimSpace(row.Find(sel.EndTimeCellSelector).Text())
	if endTimeText == "" {
		result.Errors = append(result.Errors, ExtractionError{
			RowIndex: rowIndex,
			Field:    "end_time",
			Value:    endTimeText,
			Reason:   "empty or missing end time cell",
		})
		return nil
	}
	shift.EndTime = endTimeText
	shift.EndTimeCell = fmt.Sprintf("row %d, column 4", rowIndex)

	// Extract location (optional field - don't fail if missing)
	locationText := strings.TrimSpace(row.Find(sel.LocationCellSelector).Text())
	shift.Location = locationText
	shift.LocationCell = fmt.Sprintf("row %d, column 5", rowIndex)

	// Extract required staffing (optional field)
	staffingText := strings.TrimSpace(row.Find(sel.RequiredStaffingCellSelector).Text())
	if staffingText != "" {
		// Try to parse as integer, but don't fail if we can't
		// Just record the error and continue
		if staffNum, err := parseInteger(staffingText); err == nil {
			shift.RequiredStaffing = staffNum
		} else {
			result.Errors = append(result.Errors, ExtractionError{
				RowIndex: rowIndex,
				Field:    "required_staffing",
				Value:    staffingText,
				Reason:   fmt.Sprintf("invalid integer: %v", err),
			})
		}
	}
	shift.RequiredStaffCell = fmt.Sprintf("row %d, column 6", rowIndex)

	return shift
}

// parseInteger safely parses an integer from a string.
func parseInteger(s string) (int, error) {
	// Handle empty string
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	// Try parsing as base 10 integer
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0, fmt.Errorf("failed to parse '%s' as integer: %w", s, err)
	}

	return result, nil
}

// HasErrors returns true if there are any extraction errors
func (er *ExtractionResult) HasErrors() bool {
	return len(er.Errors) > 0
}

// ErrorCount returns the number of extraction errors
func (er *ExtractionResult) ErrorCount() int {
	return len(er.Errors)
}

// ShiftCount returns the number of successfully extracted shifts
func (er *ExtractionResult) ShiftCount() int {
	return len(er.Shifts)
}

// CriticalErrorCount returns the number of rows where critical fields were missing
func (er *ExtractionResult) CriticalErrorCount() int {
	count := 0
	for _, err := range er.Errors {
		// Critical fields: date, shift_type, start_time, end_time
		if err.Field == "date" || err.Field == "shift_type" ||
			err.Field == "start_time" || err.Field == "end_time" {
			count++
		}
	}
	return count
}

// FormattedErrors returns a formatted string of all errors for logging
func (er *ExtractionResult) FormattedErrors() string {
	if !er.HasErrors() {
		return ""
	}

	lines := make([]string, 0, len(er.Errors))
	for _, err := range er.Errors {
		line := fmt.Sprintf("[Row %d] %s: %s (value='%s')",
			err.RowIndex, err.Field, err.Reason, err.Value)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
