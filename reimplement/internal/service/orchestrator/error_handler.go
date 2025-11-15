// Package orchestrator provides error handling and validation result merging for orchestration workflows.
package orchestrator

import (
	"strings"

	"github.com/schedcu/reimplement/internal/validation"
)

// ErrorPropagator collects and merges validation results from all three orchestration phases.
// It implements decision logic to determine whether to continue or stop processing based on
// error severity, type, and phase context.
type ErrorPropagator struct {
	phaseNames map[Phase]string
}

// NewErrorPropagator creates a new error propagator instance.
func NewErrorPropagator() *ErrorPropagator {
	return &ErrorPropagator{
		phaseNames: map[Phase]string{
			PhaseODSImport:            "ODS_IMPORT",
			PhaseAmionScrape:          "AMION_SCRAPE",
			PhaseCoverageCalculation:  "COVERAGE_CALCULATION",
		},
	}
}

// MergeValidationResults combines multiple ValidationResult objects into a single result.
func (ep *ErrorPropagator) MergeValidationResults(results ...*validation.ValidationResult) *validation.ValidationResult {
	merged := validation.NewValidationResult()

	for _, vr := range results {
		if vr == nil {
			continue
		}

		// Merge errors
		for _, err := range vr.Errors {
			merged.AddError(err.Field, err.Message)
		}

		// Merge warnings
		for _, warn := range vr.Warnings {
			merged.AddWarning(warn.Field, warn.Message)
		}

		// Merge infos
		for _, info := range vr.Infos {
			merged.AddInfo(info.Field, info.Message)
		}

		// Merge context (later values overwrite earlier ones)
		for key, value := range vr.Context {
			merged.SetContext(key, value)
		}
	}

	return merged
}

// MergeValidationResultsWithContext combines multiple ValidationResult objects while preserving phase context information.
func (ep *ErrorPropagator) MergeValidationResultsWithContext(
	phaseResults map[int]*validation.ValidationResult,
) *validation.ValidationResult {
	merged := validation.NewValidationResult()

	// Track which phases had errors for context
	phasesWithErrors := make([]string, 0)

	for phaseNum, vr := range phaseResults {
		if vr == nil {
			continue
		}

		phase := Phase(phaseNum)
		phaseName := ep.phaseNames[phase]

		// Track errors by phase
		if vr.HasErrors() {
			phasesWithErrors = append(phasesWithErrors, phaseName)
		}

		// Merge errors with phase context
		for _, err := range vr.Errors {
			field := err.Field
			if field == "" {
				field = phaseName
			}
			merged.AddError(field, err.Message)
		}

		// Merge warnings
		for _, warn := range vr.Warnings {
			merged.AddWarning(warn.Field, warn.Message)
		}

		// Merge infos
		for _, info := range vr.Infos {
			merged.AddInfo(info.Field, info.Message)
		}

		// Merge context
		for key, value := range vr.Context {
			merged.SetContext(key, value)
		}
	}

	// Record which phases had errors
	if len(phasesWithErrors) > 0 {
		merged.SetContext("phases_with_errors", strings.Join(phasesWithErrors, ","))
	}

	return merged
}

// ShouldContinue determines whether to continue processing to the next phase.
func (ep *ErrorPropagator) ShouldContinue(vr *validation.ValidationResult, phase Phase) bool {
	if vr == nil {
		return true
	}

	// No errors means we can continue
	if !vr.HasErrors() {
		return true
	}

	// Check each error to determine if it's critical
	for _, err := range vr.Errors {
		if ep.isCriticalErrorMessage(err.Message) {
			return false
		}

		// Major errors in Phase 1 (ODS import) must stop
		// In later phases, major errors allow skipping the entity but continuing
		if phase == PhaseODSImport && ep.isMajorErrorMessage(err.Message) {
			return false
		}
	}

	// No critical errors found - we can continue
	return true
}

// IsCriticalError determines if a validation result contains critical errors.
func (ep *ErrorPropagator) IsCriticalError(vr *validation.ValidationResult, phase Phase) bool {
	if vr == nil {
		return false
	}

	if !vr.HasErrors() {
		return false
	}

	for _, err := range vr.Errors {
		if ep.isCriticalErrorMessage(err.Message) {
			return true
		}
	}

	return false
}

// IsMajorError determines if a validation result contains major errors.
func (ep *ErrorPropagator) IsMajorError(vr *validation.ValidationResult) bool {
	if vr == nil {
		return false
	}

	if !vr.HasErrors() {
		return false
	}

	for _, err := range vr.Errors {
		// Skip critical errors
		if ep.isCriticalErrorMessage(err.Message) {
			continue
		}

		// Check if this is a major error
		if ep.isMajorErrorMessage(err.Message) {
			return true
		}
	}

	return false
}

// IsMinorError determines if a validation result contains only minor errors/warnings.
func (ep *ErrorPropagator) IsMinorError(vr *validation.ValidationResult) bool {
	if vr == nil {
		return true
	}

	// Check if there are any errors at all
	if !vr.HasErrors() {
		return true
	}

	// If there are errors, check if they're all minor
	for _, err := range vr.Errors {
		// If any error is critical or major, it's not minor
		if ep.isCriticalErrorMessage(err.Message) {
			return false
		}
		if ep.isMajorErrorMessage(err.Message) {
			return false
		}
	}

	return true
}

// isCriticalErrorMessage checks if an error message indicates a critical error.
func (ep *ErrorPropagator) isCriticalErrorMessage(msg string) bool {
	msgLower := strings.ToLower(msg)

	criticalPatterns := []string{
		"invalid file format", "invalid zip", "parse error", "parsing failed", "cannot parse",
		"malformed", "corrupted", "duplicate", "constraint violation", "unique constraint",
		"foreign key", "integrity constraint", "no space left", "disk full", "out of memory",
		"permission denied", "connection refused", "connection timeout",
	}

	for _, pattern := range criticalPatterns {
		if strings.Contains(msgLower, pattern) {
			return true
		}
	}

	return false
}

// isMajorErrorMessage checks if an error message indicates a major error.
func (ep *ErrorPropagator) isMajorErrorMessage(msg string) bool {
	msgLower := strings.ToLower(msg)

	majorPatterns := []string{
		"invalid date", "invalid format", "invalid shift", "unsupported", "not supported",
		"missing required", "missing field", "required field", "invalid type", "type mismatch",
	}

	for _, pattern := range majorPatterns {
		if strings.Contains(msgLower, pattern) {
			return true
		}
	}

	return false
}
