package ods

import (
	"fmt"
	"strings"
	"sync"

	"github.com/schedcu/reimplement/internal/validation"
)

// ErrorSeverity indicates how critical an error is during import.
type ErrorSeverity string

const (
	ErrorSeverityCritical  ErrorSeverity = "critical"  // Fails the entire import
	ErrorSeverityMajor     ErrorSeverity = "major"     // Skips entity but continues
	ErrorSeverityMinor     ErrorSeverity = "minor"     // Logs warning but continues
	ErrorSeverityInfo      ErrorSeverity = "info"      // Informational only
)

// ErrorType represents the category of parsing error.
// Error types are used to categorize validation failures for better error reporting and handling.
type ErrorType string

const (
	// Work package [1.2] error types for ODS validation
	// These provide specific categorization for ODS parsing errors

	// ErrorTypeMissingRequired indicates a required field is missing from a row.
	ErrorTypeMissingRequired ErrorType = "MISSING_REQUIRED_FIELD"

	// ErrorTypeInvalidValue indicates a field contains an invalid value type.
	ErrorTypeInvalidValue ErrorType = "INVALID_VALUE"

	// ErrorTypeMissingRow indicates an expected row is missing from the spreadsheet.
	ErrorTypeMissingRow ErrorType = "MISSING_ROW"

	// ErrorTypeInvalidFormat indicates a field value doesn't match expected format.
	ErrorTypeInvalidFormat ErrorType = "INVALID_FORMAT"

	// ErrorTypeDuplicate indicates the same entry appears multiple times.
	ErrorTypeDuplicate ErrorType = "DUPLICATE_ENTRY"
)

// ParsingError represents a single error encountered during ODS parsing.
// It includes location information (row/column) for precise error reporting.
type ParsingError struct {
	// Type categorizes the error
	Type ErrorType

	// Message is a human-readable description of the error
	Message string

	// Row number where the error occurred (1-indexed), 0 if not applicable
	Row int

	// Column name (e.g., "Date", "ShiftType") where the error occurred, empty if not applicable
	Column string

	// CellReference is the Excel-style cell reference (e.g., "A1", "C5")
	CellReference string

	// Details contains additional context about the error
	Details map[string]interface{}
}

// NewParsingError creates a new ParsingError with the given type and message.
func NewParsingError(errorType ErrorType, message string) *ParsingError {
	return &ParsingError{
		Type:    errorType,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithLocation adds row and column information to the error.
func (pe *ParsingError) WithLocation(row int, column string) *ParsingError {
	pe.Row = row
	pe.Column = column
	return pe
}

// WithCellReference adds Excel-style cell reference to the error.
func (pe *ParsingError) WithCellReference(ref string) *ParsingError {
	pe.CellReference = ref
	return pe
}

// WithDetail adds a key-value detail to the error context.
func (pe *ParsingError) WithDetail(key string, value interface{}) *ParsingError {
	pe.Details[key] = value
	return pe
}

// Error implements the error interface.
func (pe *ParsingError) Error() string {
	parts := []string{string(pe.Type)}

	if pe.CellReference != "" {
		parts = append(parts, fmt.Sprintf("cell %s", pe.CellReference))
	} else if pe.Row > 0 && pe.Column != "" {
		parts = append(parts, fmt.Sprintf("row %d, column '%s'", pe.Row, pe.Column))
	}

	parts = append(parts, pe.Message)

	return strings.Join(parts, ": ")
}

// ImportError represents a single error that occurred during ODS import.
type ImportError struct {
	Severity    ErrorSeverity
	EntityType  string // "shift", "schedule_version", etc.
	EntityID    string // Identifier in the ODS file
	Field       string
	Message     string
	OriginalErr error
	Context     map[string]interface{}
}

// ODSErrorCollector collects and aggregates errors that occur during ODS import.
// It is thread-safe and supports both critical and non-critical error collection.
// It also supports Work Package [1.2] ParsingError collection for fine-grained error reporting.
type ODSErrorCollector struct {
	mu            sync.RWMutex
	errors        []ImportError
	parsingErrors *parsingErrors // Work Package [1.2] parsing errors
	criticalErr   error           // First critical error encountered
	importMetrics
}

// importMetrics tracks various statistics during import.
type importMetrics struct {
	totalShifts      int
	createdShifts    int
	failedShifts     int
	totalSchedules   int
	createdSchedules int
	failedSchedules  int
}

// NewODSErrorCollector creates a new error collector.
// The collector is thread-safe and maintains separate lists for parsing errors and import errors.
func NewODSErrorCollector() *ODSErrorCollector {
	return &ODSErrorCollector{
		errors: make([]ImportError, 0),
	}
}

// ============================================================================
// PARSING ERROR COLLECTION (Work Package [1.2])
// ============================================================================

// parsingErrors stores ParsingError instances for fine-grained error collection
type parsingErrors struct {
	errors   []*ParsingError
	warnings []*ParsingError
}

// AddError adds a ParsingError to the collector without locking (internal use).
// These are collected separately from ImportErrors.
func (ec *ODSErrorCollector) AddError(err *ParsingError) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if err != nil {
		// Initialize parsing errors if needed
		if ec.parsingErrors == nil {
			ec.parsingErrors = &parsingErrors{
				errors:   make([]*ParsingError, 0),
				warnings: make([]*ParsingError, 0),
			}
		}
		ec.parsingErrors.errors = append(ec.parsingErrors.errors, err)
	}
}

// AddWarning adds a ParsingError as a warning to the collector.
// Warnings are non-critical issues that don't prevent import.
func (ec *ODSErrorCollector) AddWarning(warn *ParsingError) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if warn != nil {
		// Initialize parsing errors if needed
		if ec.parsingErrors == nil {
			ec.parsingErrors = &parsingErrors{
				errors:   make([]*ParsingError, 0),
				warnings: make([]*ParsingError, 0),
			}
		}
		ec.parsingErrors.warnings = append(ec.parsingErrors.warnings, warn)
	}
}

// HasErrors returns true if any ParsingErrors have been collected.
func (ec *ODSErrorCollector) HasErrors() bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.parsingErrors != nil && len(ec.parsingErrors.errors) > 0
}

// HasWarnings returns true if any warnings have been collected.
func (ec *ODSErrorCollector) HasWarnings() bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.parsingErrors != nil && len(ec.parsingErrors.warnings) > 0
}

// ErrorCount returns the number of ParsingErrors collected.
func (ec *ODSErrorCollector) ErrorCountParsing() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	if ec.parsingErrors == nil {
		return 0
	}
	return len(ec.parsingErrors.errors)
}

// WarningCount returns the number of warnings collected.
func (ec *ODSErrorCollector) WarningCountParsing() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	if ec.parsingErrors == nil {
		return 0
	}
	return len(ec.parsingErrors.warnings)
}

// Errors returns all collected ParsingErrors.
func (ec *ODSErrorCollector) Errors() []*ParsingError {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	if ec.parsingErrors == nil {
		return make([]*ParsingError, 0)
	}
	result := make([]*ParsingError, len(ec.parsingErrors.errors))
	copy(result, ec.parsingErrors.errors)
	return result
}

// Warnings returns all collected warnings.
func (ec *ODSErrorCollector) Warnings() []*ParsingError {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	if ec.parsingErrors == nil {
		return make([]*ParsingError, 0)
	}
	result := make([]*ParsingError, len(ec.parsingErrors.warnings))
	copy(result, ec.parsingErrors.warnings)
	return result
}

// GroupErrorsByType returns errors grouped by their type.
func (ec *ODSErrorCollector) GroupErrorsByType() map[ErrorType][]*ParsingError {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	grouped := make(map[ErrorType][]*ParsingError)

	if ec.parsingErrors == nil {
		return grouped
	}

	for _, err := range ec.parsingErrors.errors {
		if err != nil {
			grouped[err.Type] = append(grouped[err.Type], err)
		}
	}

	return grouped
}

// GroupErrorsByRow returns errors grouped by row number.
func (ec *ODSErrorCollector) GroupErrorsByRow() map[int][]*ParsingError {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	grouped := make(map[int][]*ParsingError)

	if ec.parsingErrors == nil {
		return grouped
	}

	for _, err := range ec.parsingErrors.errors {
		if err != nil && err.Row > 0 {
			grouped[err.Row] = append(grouped[err.Row], err)
		}
	}

	return grouped
}

// ToValidationResult converts collected ParsingErrors and warnings to a ValidationResult.
// This creates a validation.ValidationResult that can be returned from validation functions
// and serialized to JSON for API responses.
func (ec *ODSErrorCollector) ToValidationResult() *validation.ValidationResult {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	result := validation.NewValidationResult()

	if ec.parsingErrors == nil {
		return result
	}

	// Add all collected ParsingErrors
	for _, err := range ec.parsingErrors.errors {
		if err != nil {
			// Use column name if available, otherwise use cell reference, otherwise use error type
			field := err.Column
			if field == "" && err.CellReference != "" {
				field = err.CellReference
			}
			if field == "" {
				field = string(err.Type)
			}

			// Create detailed error message
			message := ec.formatParsingErrorMessage(err)

			// Add the error to the validation result
			result.AddError(field, message)
		}
	}

	// Add all collected warnings
	for _, warn := range ec.parsingErrors.warnings {
		if warn != nil {
			// Use column name if available, otherwise use cell reference, otherwise use error type
			field := warn.Column
			if field == "" && warn.CellReference != "" {
				field = warn.CellReference
			}
			if field == "" {
				field = string(warn.Type)
			}

			// Create detailed warning message
			message := ec.formatParsingErrorMessage(warn)

			// Add the warning to the validation result
			result.AddWarning(field, message)
		}
	}

	return result
}

// formatParsingErrorMessage creates a detailed error message from a ParsingError.
func (ec *ODSErrorCollector) formatParsingErrorMessage(err *ParsingError) string {
	var parts []string

	// Add the base error message
	parts = append(parts, err.Message)

	// Add location information if available
	if err.CellReference != "" {
		parts = append(parts, fmt.Sprintf("[cell %s]", err.CellReference))
	} else if err.Row > 0 && err.Column != "" {
		parts = append(parts, fmt.Sprintf("[row %d, column '%s']", err.Row, err.Column))
	} else if err.Row > 0 {
		parts = append(parts, fmt.Sprintf("[row %d]", err.Row))
	} else if err.Column != "" {
		parts = append(parts, fmt.Sprintf("[column '%s']", err.Column))
	}

	// Add additional details if available
	if err.Details != nil && len(err.Details) > 0 {
		detailParts := make([]string, 0)
		if expected, ok := err.Details["expected"]; ok {
			detailParts = append(detailParts, fmt.Sprintf("expected %v", expected))
		}
		if actual, ok := err.Details["actual"]; ok {
			detailParts = append(detailParts, fmt.Sprintf("got %v", actual))
		}
		if format, ok := err.Details["format"]; ok {
			detailParts = append(detailParts, fmt.Sprintf("format %v", format))
		}

		if len(detailParts) > 0 {
			parts = append(parts, "("+strings.Join(detailParts, ", ")+")")
		}
	}

	return strings.Join(parts, " ")
}

// ============================================================================
// IMPORT ERROR COLLECTION (Legacy API)
// ============================================================================

// AddCritical records a critical error that should stop the import.
func (ec *ODSErrorCollector) AddCritical(entityType, entityID, field, message string, originalErr error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if ec.criticalErr != nil {
		return
	}

	ec.criticalErr = originalErr
	ec.errors = append(ec.errors, ImportError{
		Severity:    ErrorSeverityCritical,
		EntityType:  entityType,
		EntityID:    entityID,
		Field:       field,
		Message:     message,
		OriginalErr: originalErr,
		Context:     make(map[string]interface{}),
	})
}

// AddMajor records a major error that skips an entity but allows import to continue.
func (ec *ODSErrorCollector) AddMajor(entityType, entityID, field, message string, originalErr error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.errors = append(ec.errors, ImportError{
		Severity:    ErrorSeverityMajor,
		EntityType:  entityType,
		EntityID:    entityID,
		Field:       field,
		Message:     message,
		OriginalErr: originalErr,
		Context:     make(map[string]interface{}),
	})
}

// AddMinor records a minor error that is logged but doesn't affect entity creation.
func (ec *ODSErrorCollector) AddMinor(entityType, entityID, field, message string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.errors = append(ec.errors, ImportError{
		Severity:    ErrorSeverityMinor,
		EntityType:  entityType,
		EntityID:    entityID,
		Field:       field,
		Message:     message,
		OriginalErr: nil,
		Context:     make(map[string]interface{}),
	})
}

// AddInfo records informational messages during import.
func (ec *ODSErrorCollector) AddInfo(entityType, entityID, message string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.errors = append(ec.errors, ImportError{
		Severity:    ErrorSeverityInfo,
		EntityType:  entityType,
		EntityID:    entityID,
		Message:     message,
		OriginalErr: nil,
		Context:     make(map[string]interface{}),
	})
}

// HasCriticalError returns true if any critical error has been recorded.
func (ec *ODSErrorCollector) HasCriticalError() bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.criticalErr != nil
}

// CriticalError returns the first critical error that was recorded, or nil if none.
func (ec *ODSErrorCollector) CriticalError() error {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.criticalErr
}

// AllErrors returns a copy of all collected errors.
func (ec *ODSErrorCollector) AllErrors() []ImportError {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	result := make([]ImportError, len(ec.errors))
	copy(result, ec.errors)
	return result
}

// ErrorCount returns the total number of errors collected.
func (ec *ODSErrorCollector) ErrorCount() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return len(ec.errors)
}

// MajorErrorCount returns the count of major errors only.
func (ec *ODSErrorCollector) MajorErrorCount() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	count := 0
	for _, e := range ec.errors {
		if e.Severity == ErrorSeverityMajor {
			count++
		}
	}
	return count
}

// RecordShiftCreated increments the counter for successfully created shifts.
func (ec *ODSErrorCollector) RecordShiftCreated() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.createdShifts++
}

// RecordShiftFailed increments the counter for failed shifts.
func (ec *ODSErrorCollector) RecordShiftFailed() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.failedShifts++
}

// RecordScheduleCreated increments the counter for successfully created schedules.
func (ec *ODSErrorCollector) RecordScheduleCreated() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.createdSchedules++
}

// RecordScheduleFailed increments the counter for failed schedules.
func (ec *ODSErrorCollector) RecordScheduleFailed() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.failedSchedules++
}

// SetTotalShifts sets the expected total number of shifts to be imported.
func (ec *ODSErrorCollector) SetTotalShifts(count int) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.totalShifts = count
}

// SetTotalSchedules sets the expected total number of schedules to be imported.
func (ec *ODSErrorCollector) SetTotalSchedules(count int) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.totalSchedules = count
}

// BuildValidationResult converts collected errors to a ValidationResult.
func (ec *ODSErrorCollector) BuildValidationResult() *validation.ValidationResult {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	result := validation.NewValidationResult()

	for _, err := range ec.errors {
		field := fmt.Sprintf("%s:%s", err.EntityType, err.EntityID)
		if err.Field != "" {
			field = fmt.Sprintf("%s.%s", field, err.Field)
		}

		switch err.Severity {
		case ErrorSeverityCritical:
			result.AddError(field, err.Message)
		case ErrorSeverityMajor:
			result.AddError(field, fmt.Sprintf("Failed to import: %s", err.Message))
		case ErrorSeverityMinor:
			result.AddWarning(field, err.Message)
		case ErrorSeverityInfo:
			result.AddInfo(field, err.Message)
		}
	}

	// Add metrics to context
	result.SetContext("total_shifts", ec.totalShifts)
	result.SetContext("created_shifts", ec.createdShifts)
	result.SetContext("failed_shifts", ec.failedShifts)
	result.SetContext("total_schedules", ec.totalSchedules)
	result.SetContext("created_schedules", ec.createdSchedules)
	result.SetContext("failed_schedules", ec.failedSchedules)

	return result
}
