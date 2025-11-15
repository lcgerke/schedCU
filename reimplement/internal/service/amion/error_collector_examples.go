package amion

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExampleErrorCollector_BasicUsage demonstrates basic error collection.
func ExampleErrorCollector_BasicUsage() {
	collector := NewAmionErrorCollector()

	// Simulate parsing a table row that's missing the shift date
	collector.AddError(MissingCell, 5, 1, "Expected shift date in column 1")

	// Simulate invalid time format
	collector.AddError(InvalidValue, 5, 3, "Invalid time format '25:00', expected HH:MM")

	// Check results
	fmt.Printf("Total errors: %d\n", collector.ErrorCount())
	fmt.Printf("Has errors: %v\n", collector.HasErrors())

	result := collector.ToValidationResult()
	fmt.Printf("Result valid: %v\n", result.IsValid())
	fmt.Printf("Error count in result: %d\n", result.ErrorCount())

	// Output:
	// Total errors: 2
	// Has errors: true
	// Result valid: false
	// Error count in result: 2
}

// ExampleErrorCollector_ParsingIntegration demonstrates integration with HTML parsing.
func ExampleErrorCollector_ParsingIntegration() {
	html := `
	<table>
		<tbody>
			<tr>
				<td>2025-11-15</td>
				<td>Technologist</td>
				<td>07:00</td>
				<td>15:00</td>
			</tr>
			<tr>
				<td></td>
				<td>Radiologist</td>
				<td>14:00</td>
				<td>22:00</td>
			</tr>
		</tbody>
	</table>
	`

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	collector := NewAmionErrorCollector()

	rowNum := 0
	doc.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
		rowNum++

		// Extract date
		dateText := strings.TrimSpace(row.Find("td").Eq(0).Text())
		if dateText == "" {
			collector.AddError(MissingCell, rowNum, 1,
				"Expected shift date in column 1")
			return // Skip this row
		}

		// Extract shift type
		typeText := strings.TrimSpace(row.Find("td").Eq(1).Text())
		if typeText == "" {
			collector.AddError(MissingCell, rowNum, 2,
				"Expected shift type in column 2")
			return
		}

		fmt.Printf("Row %d: %s - %s\n", rowNum, dateText, typeText)
	})

	result := collector.ToValidationResult()
	if result.HasErrors() {
		for _, err := range result.Errors {
			fmt.Printf("Error [%s]: %s\n", err.Field, err.Message)
		}
	}

	// Output:
	// Row 1: 2025-11-15 - Technologist
	// Error [R2C1]: [MissingCell@R2C1] Expected shift date in column 1
}

// ExampleErrorCollector_GroupingErrors demonstrates error grouping by type.
func ExampleErrorCollector_GroupingErrors() {
	collector := NewAmionErrorCollector()

	// Add various types of errors
	collector.AddError(MissingCell, 5, 1, "Missing date")
	collector.AddError(MissingCell, 6, 1, "Missing date")
	collector.AddError(InvalidValue, 5, 3, "Invalid time format")
	collector.AddError(InvalidHTML, 0, 0, "Table element not found")

	// Group by type
	groups := collector.GroupErrorsByType()

	for errType, errors := range groups {
		fmt.Printf("%s: %d errors\n", errType, len(errors))
		for _, err := range errors {
			if err.Row > 0 {
				fmt.Printf("  Row %d, Col %d: %s\n", err.Row, err.Col, err.Details)
			} else {
				fmt.Printf("  %s\n", err.Details)
			}
		}
	}

	// Output:
	// INVALID_HTML: 1 errors
	//   Table element not found
	// INVALID_VALUE: 1 errors
	//   Row 5, Col 3: Invalid time format
	// MISSING_CELL: 2 errors
	//   Row 5, Col 1: Missing date
	//   Row 6, Col 1: Missing date
}

// ExampleErrorCollector_ValidationIntegration shows integration with ValidationResult.
func ExampleErrorCollector_ValidationIntegration() {
	collector := NewAmionErrorCollector()

	collector.AddError(MissingCell, 5, 3, "Missing shift date")
	collector.AddError(InvalidValue, 10, 4, "Invalid staff count")

	// Convert to ValidationResult
	result := collector.ToValidationResult()

	// Check validity
	if !result.IsValid() {
		fmt.Printf("Validation failed with %d errors:\n", result.ErrorCount())
		for _, msg := range result.Errors {
			fmt.Printf("  [%s] %s\n", msg.Field, msg.Message)
		}
	}

	// Access context information
	totalErrors, _ := result.GetContext("total_errors")
	fmt.Printf("\nContext - Total errors: %v\n", totalErrors)

	typeBreakdown, _ := result.GetContext("errors_by_type")
	fmt.Printf("Context - Error breakdown: %v\n", typeBreakdown)

	// Output:
	// Validation failed with 2 errors:
	//   [R5C3] [MissingCell@R5C3] Missing shift date
	//   [R10C4] [InvalidValue@R10C4] Invalid staff count
	//
	// Context - Total errors: 2
	// Context - Error breakdown: map[INVALID_VALUE:1 MISSING_CELL:1]
}

// ExampleErrorCollector_ClearAndReset demonstrates clear functionality.
func ExampleErrorCollector_ClearAndReset() {
	collector := NewAmionErrorCollector()

	collector.AddError(MissingCell, 5, 3, "Missing date")
	collector.AddError(InvalidValue, 5, 4, "Invalid time")

	fmt.Printf("Before clear: %d errors\n", collector.ErrorCount())
	fmt.Printf("Has errors: %v\n", collector.HasErrors())

	collector.Clear()

	fmt.Printf("After clear: %d errors\n", collector.ErrorCount())
	fmt.Printf("Has errors: %v\n", collector.HasErrors())

	// Output:
	// Before clear: 2 errors
	// Has errors: true
	// After clear: 0 errors
	// Has errors: false
}

// ExampleErrorCollector_RetrievingErrors shows how to retrieve errors safely.
func ExampleErrorCollector_RetrievingErrors() {
	collector := NewAmionErrorCollector()

	collector.AddError(MissingCell, 5, 3, "Missing date")
	collector.AddError(InvalidValue, 5, 4, "Invalid time")

	// GetErrors returns a copy, safe for external use
	errors := collector.GetErrors()

	fmt.Printf("Retrieved %d errors\n", len(errors))
	for _, err := range errors {
		fmt.Printf("  Type: %v, Row: %d, Col: %d, Details: %s\n",
			err.ErrorType, err.Row, err.Col, err.Details)
	}

	// Output:
	// Retrieved 2 errors
	//   Type: MISSING_CELL, Row: 5, Col: 3, Details: Missing date
	//   Type: INVALID_VALUE, Row: 5, Col: 4, Details: Invalid time
}
