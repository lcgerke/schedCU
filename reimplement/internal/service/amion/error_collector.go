// Package amion provides services for scraping and parsing Amion schedule data.
package amion

import (
	"fmt"
	"sync"

	"github.com/schedcu/reimplement/internal/validation"
)

// ErrorType represents the type of parsing error that occurred.
type ErrorType string

const (
	// MissingCell indicates an expected cell in a row was not found.
	MissingCell ErrorType = "MISSING_CELL"

	// InvalidValue indicates a cell contained a value with the wrong type or format.
	InvalidValue ErrorType = "INVALID_VALUE"

	// MissingRow indicates an expected row was not found in the table.
	MissingRow ErrorType = "MISSING_ROW"

	// InvalidHTML indicates the HTML structure does not match the expected selectors.
	InvalidHTML ErrorType = "INVALID_HTML"

	// EmptyTable indicates no data rows were found in the table.
	EmptyTable ErrorType = "EMPTY_TABLE"

	// EncodingError indicates a character encoding issue was encountered.
	EncodingError ErrorType = "ENCODING_ERROR"
)

// String returns the string representation of an ErrorType.
func (et ErrorType) String() string {
	return string(et)
}

// AmionError represents a single parsing error with row and column context.
type AmionError struct {
	// ErrorType categorizes the error
	ErrorType ErrorType

	// Row is the table row index (0-based), or 0 for structural errors
	Row int

	// Col is the table column index (0-based), or 0 for structural errors
	Col int

	// Details contains contextual information about the error
	Details string
}

// AmionErrorCollector collects all parsing errors without fail-fast,
// allowing partial data to be returned alongside detailed error information.
type AmionErrorCollector struct {
	errors []AmionError
	mu     sync.Mutex
}

// NewAmionErrorCollector creates a new error collector.
func NewAmionErrorCollector() *AmionErrorCollector {
	return &AmionErrorCollector{
		errors: make([]AmionError, 0),
	}
}

// AddError records a parsing error without stopping execution.
// This allows the parser to continue and collect all errors at once.
//
// Parameters:
//   - errorType: the type of error (MissingCell, InvalidValue, etc.)
//   - row: the row index (0-based), or 0 for structural errors
//   - col: the column index (0-based), or 0 for structural errors
//   - details: human-readable explanation of what went wrong
func (aec *AmionErrorCollector) AddError(errorType ErrorType, row, col int, details string) {
	aec.mu.Lock()
	defer aec.mu.Unlock()

	aec.errors = append(aec.errors, AmionError{
		ErrorType: errorType,
		Row:       row,
		Col:       col,
		Details:   details,
	})
}

// GetErrors returns a copy of all collected errors.
func (aec *AmionErrorCollector) GetErrors() []AmionError {
	aec.mu.Lock()
	defer aec.mu.Unlock()

	// Return a copy to prevent external modification
	result := make([]AmionError, len(aec.errors))
	copy(result, aec.errors)
	return result
}

// ErrorCount returns the total number of collected errors.
func (aec *AmionErrorCollector) ErrorCount() int {
	aec.mu.Lock()
	defer aec.mu.Unlock()

	return len(aec.errors)
}

// HasErrors returns true if any errors have been collected.
func (aec *AmionErrorCollector) HasErrors() bool {
	aec.mu.Lock()
	defer aec.mu.Unlock()

	return len(aec.errors) > 0
}

// Clear removes all collected errors.
func (aec *AmionErrorCollector) Clear() {
	aec.mu.Lock()
	defer aec.mu.Unlock()

	aec.errors = make([]AmionError, 0)
}

// GroupErrorsByType groups all errors by their type.
// Returns a map where keys are ErrorTypes and values are slices of errors of that type.
func (aec *AmionErrorCollector) GroupErrorsByType() map[ErrorType][]AmionError {
	aec.mu.Lock()
	defer aec.mu.Unlock()

	groups := make(map[ErrorType][]AmionError)

	for _, err := range aec.errors {
		groups[err.ErrorType] = append(groups[err.ErrorType], err)
	}

	return groups
}

// ToValidationResult converts collected errors into a ValidationResult.
// Returns a ValidationResult with all errors translated to the standard format,
// including row/column references for debugging.
func (aec *AmionErrorCollector) ToValidationResult() *validation.ValidationResult {
	aec.mu.Lock()
	defer aec.mu.Unlock()

	result := validation.NewValidationResult()

	// Convert each AmionError to a ValidationMessage
	for _, err := range aec.errors {
		field := formatCellReference(err.Row, err.Col)
		message := formatErrorMessage(err.ErrorType, field, err.Details)

		result.AddError(field, message)
	}

	// Add context information for debugging
	result.SetContext("total_errors", len(aec.errors))

	// Add error type breakdown
	typeBreakdown := make(map[ErrorType]int)
	for _, err := range aec.errors {
		typeBreakdown[err.ErrorType]++
	}
	result.SetContext("errors_by_type", typeBreakdown)

	// Add error type counts as separate context entries for easier access
	for errType, count := range typeBreakdown {
		key := fmt.Sprintf("error_count_%s", errType)
		result.SetContext(key, count)
	}

	return result
}

// formatCellReference creates a human-readable cell reference for error messages.
// For data errors, returns "R5C3" format. For structural errors, returns appropriate description.
func formatCellReference(row, col int) string {
	if row == 0 && col == 0 {
		// Structural error (no specific cell)
		return "table"
	}
	if col == 0 {
		// Row-level error
		return fmt.Sprintf("R%d", row)
	}
	// Cell-level error
	return fmt.Sprintf("R%dC%d", row, col)
}

// formatErrorMessage creates a comprehensive error message that includes
// the error type, cell reference, and contextual details.
func formatErrorMessage(errType ErrorType, cellRef, details string) string {
	switch errType {
	case MissingCell:
		return fmt.Sprintf("[MissingCell@%s] %s", cellRef, details)
	case InvalidValue:
		return fmt.Sprintf("[InvalidValue@%s] %s", cellRef, details)
	case MissingRow:
		return fmt.Sprintf("[MissingRow@%s] %s", cellRef, details)
	case InvalidHTML:
		return fmt.Sprintf("[InvalidHTML] %s", details)
	case EmptyTable:
		return fmt.Sprintf("[EmptyTable] %s", details)
	case EncodingError:
		return fmt.Sprintf("[EncodingError] %s", details)
	default:
		return fmt.Sprintf("[Unknown@%s] %s", cellRef, details)
	}
}
