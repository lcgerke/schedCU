package ods

import (
	"fmt"
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
//
// Example usage:
//
//	collector := NewODSErrorCollector()
//	// ... during import ...
//	if err := repository.Create(ctx, shift); err != nil {
//	    collector.AddMajor("shift", "SHIFT_001", "database", err.Error(), err)
//	}
//	result := collector.BuildValidationResult()
//	if result.HasErrors() {
//	    return fmt.Errorf("import failed with %d errors", result.ErrorCount())
//	}
type ODSErrorCollector struct {
	mu           sync.RWMutex
	errors       []ImportError
	criticalErr  error // First critical error encountered
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
func NewODSErrorCollector() *ODSErrorCollector {
	return &ODSErrorCollector{
		errors: make([]ImportError, 0),
	}
}

// AddCritical records a critical error that should stop the import.
// Only the first critical error is saved; subsequent critical errors are ignored.
func (ec *ODSErrorCollector) AddCritical(entityType, entityID, field, message string, originalErr error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if ec.criticalErr != nil {
		// Only save the first critical error
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
// Critical errors result in ERROR severity; major errors in WARNING; minor in WARNING; info in INFO.
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
