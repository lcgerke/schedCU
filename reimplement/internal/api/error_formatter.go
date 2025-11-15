// Package api provides HTTP API utilities and helpers for status code mapping.
package api

import (
	"fmt"

	"github.com/schedcu/reimplement/internal/validation"
)

// FormatValidationErrors converts a ValidationResult into an ErrorDetail structure
// suitable for API responses. It aggregates validation messages into organized
// field-based error structures.
//
// The structure of returned ErrorDetail:
// - Code: "VALIDATION_ERROR"
// - Message: Summary string like "Validation failed: 2 error(s), 1 warning(s)"
// - Details map contains:
//   - error_count: number of errors
//   - warning_count: number of warnings (if > 0)
//   - info_count: number of infos (if > 0)
//   - errors: map[field]message(s) for each error
//   - warnings: map[field]message(s) for each warning (if present)
//   - infos: map[field]message(s) for each info (if present)
//   - context: copy of validation context (if present)
func FormatValidationErrors(vr *validation.ValidationResult) *ErrorDetail {
	if vr == nil {
		return &ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed: 0 error(s)",
			Details: make(map[string]interface{}),
		}
	}

	errorCount := vr.ErrorCount()
	warningCount := vr.WarningCount()
	infoCount := vr.InfoCount()

	// Create the summary message
	message := formatValidationSummary(errorCount, warningCount, infoCount)

	// Build details map
	details := make(map[string]interface{})
	details["error_count"] = errorCount

	// Add error messages grouped by field
	if errorCount > 0 {
		errorsMap := aggregateMessages(vr.Errors)
		details["errors"] = errorsMap
	}

	// Add warning messages grouped by field if present
	if warningCount > 0 {
		details["warning_count"] = warningCount
		warningsMap := aggregateMessages(vr.Warnings)
		details["warnings"] = warningsMap
	}

	// Add info messages grouped by field if present
	if infoCount > 0 {
		details["info_count"] = infoCount
		infosMap := aggregateMessages(vr.Infos)
		details["infos"] = infosMap
	}

	// Include context if not empty
	if len(vr.Context) > 0 {
		details["context"] = vr.Context
	}

	return &ErrorDetail{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: details,
	}
}

// formatValidationSummary creates a concise summary message of validation results.
func formatValidationSummary(errorCount, warningCount, infoCount int) string {
	message := fmt.Sprintf("Validation failed: %d error(s)", errorCount)

	if warningCount > 0 {
		message += fmt.Sprintf(", %d warning(s)", warningCount)
	}

	if infoCount > 0 {
		message += fmt.Sprintf(", %d info(s)", infoCount)
	}

	return message
}

// aggregateMessages groups validation messages by field name.
// If a field has multiple messages, they are aggregated into a slice.
// If a field has a single message, it's stored as a string.
func aggregateMessages(messages []validation.SimpleValidationMessage) map[string]interface{} {
	result := make(map[string]interface{})

	for _, msg := range messages {
		field := msg.Field
		if field == "" {
			field = "_global_"
		}

		if existing, ok := result[field]; ok {
			// Field already exists - convert to or append to slice
			switch v := existing.(type) {
			case string:
				// Convert string to slice
				result[field] = []interface{}{v, msg.Message}
			case []interface{}:
				// Append to existing slice
				result[field] = append(v, msg.Message)
			}
		} else {
			// First message for this field
			result[field] = msg.Message
		}
	}

	return result
}
