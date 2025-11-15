// Package validation provides validation result types and error reporting.
package validation

import (
	"fmt"
	"strings"
	"time"
)

// Severity represents the severity level of a validation message.
type Severity string

const (
	ERROR   Severity = "error"
	WARNING Severity = "warning"
	INFO    Severity = "info"
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	return string(s)
}

// FromString parses a string into a Severity value.
// The input is case-insensitive and trimmed of whitespace.
// Returns an error if the input is not a valid severity.
func FromString(s string) (Severity, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	switch s {
	case "error":
		return ERROR, nil
	case "warning":
		return WARNING, nil
	case "info":
		return INFO, nil
	default:
		return "", fmt.Errorf("invalid severity: %q", s)
	}
}

// MessageCode represents a known validation error type.
type MessageCode string

const (
	INVALID_FILE_FORMAT       MessageCode = "INVALID_FILE_FORMAT"
	MISSING_REQUIRED_FIELD    MessageCode = "MISSING_REQUIRED_FIELD"
	DUPLICATE_ENTRY           MessageCode = "DUPLICATE_ENTRY"
	PARSE_ERROR               MessageCode = "PARSE_ERROR"
	DATABASE_ERROR            MessageCode = "DATABASE_ERROR"
	EXTERNAL_SERVICE_ERROR    MessageCode = "EXTERNAL_SERVICE_ERROR"
	UNKNOWN_ERROR             MessageCode = "UNKNOWN_ERROR"
)

// String returns the string representation of the message code.
func (mc MessageCode) String() string {
	return string(mc)
}

// MessageCodeFromString parses a string into a MessageCode value.
// The input is case-insensitive and trimmed of whitespace.
// Returns an error if the input is not a valid message code.
func MessageCodeFromString(s string) (MessageCode, error) {
	s = strings.TrimSpace(strings.ToUpper(s))

	switch s {
	case "INVALID_FILE_FORMAT":
		return INVALID_FILE_FORMAT, nil
	case "MISSING_REQUIRED_FIELD":
		return MISSING_REQUIRED_FIELD, nil
	case "DUPLICATE_ENTRY":
		return DUPLICATE_ENTRY, nil
	case "PARSE_ERROR":
		return PARSE_ERROR, nil
	case "DATABASE_ERROR":
		return DATABASE_ERROR, nil
	case "EXTERNAL_SERVICE_ERROR":
		return EXTERNAL_SERVICE_ERROR, nil
	case "UNKNOWN_ERROR":
		return UNKNOWN_ERROR, nil
	default:
		return "", fmt.Errorf("invalid message code: %q", s)
	}
}

// ValidationMessage represents a single validation result or error.
// It contains all information about what went wrong and why.
type ValidationMessage struct {
	// Code is the machine-readable error code for this validation message.
	Code MessageCode `json:"code"`

	// Severity indicates how serious this validation failure is.
	Severity Severity `json:"severity"`

	// Message is a human-readable description of the validation failure.
	Message string `json:"message"`

	// Field is the optional name of the field that caused the error.
	// Empty string if the error is not related to a specific field.
	Field string `json:"field,omitempty"`

	// Details contains contextual information about the validation failure.
	// Examples: actual value received, expected value, line number, etc.
	// Can be nil if no additional context is needed.
	Details map[string]interface{} `json:"details,omitempty"`

	// Timestamp records when this validation message was created.
	Timestamp time.Time `json:"timestamp"`
}

// SimpleValidationMessage represents a simple field+message pair.
// This is used internally by ValidationResult for backward compatibility.
// DEPRECATED: Use ValidationMessage for new code.
type SimpleValidationMessage struct {
	// Field is the identifier for the field, source, or context where the message originates
	Field string `json:"field"`

	// Message is the descriptive text of the validation message
	Message string `json:"message"`
}

// ValidationResult aggregates validation messages and contextual information.
// It supports three levels of messages: errors, warnings, and infos.
// Additional debug information can be stored in the Context map.
type ValidationResult struct {
	// Errors contains validation error messages
	Errors []SimpleValidationMessage `json:"errors"`

	// Warnings contains validation warning messages
	Warnings []SimpleValidationMessage `json:"warnings"`

	// Infos contains validation info messages
	Infos []SimpleValidationMessage `json:"infos"`

	// Context stores arbitrary key-value pairs for debug information
	Context map[string]interface{} `json:"context"`
}

// NewValidationResult creates a new ValidationResult with initialized slices and context map.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors:   make([]SimpleValidationMessage, 0),
		Warnings: make([]SimpleValidationMessage, 0),
		Infos:    make([]SimpleValidationMessage, 0),
		Context:  make(map[string]interface{}),
	}
}

// AddError adds an error message to the validation result.
// field is the identifier (e.g., field name, component name) for the error.
// message is the descriptive error text.
func (vr *ValidationResult) AddError(field, message string) {
	vr.Errors = append(vr.Errors, SimpleValidationMessage{
		Field:   field,
		Message: message,
	})
}

// AddWarning adds a warning message to the validation result.
// field is the identifier (e.g., field name, component name) for the warning.
// message is the descriptive warning text.
func (vr *ValidationResult) AddWarning(field, message string) {
	vr.Warnings = append(vr.Warnings, SimpleValidationMessage{
		Field:   field,
		Message: message,
	})
}

// AddInfo adds an info message to the validation result.
// field is the identifier (e.g., source name, component name) for the info.
// message is the descriptive info text.
func (vr *ValidationResult) AddInfo(field, message string) {
	vr.Infos = append(vr.Infos, SimpleValidationMessage{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if the validation result contains any error messages.
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// HasWarnings returns true if the validation result contains any warning messages.
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}

// ErrorCount returns the total number of error messages.
func (vr *ValidationResult) ErrorCount() int {
	return len(vr.Errors)
}

// WarningCount returns the total number of warning messages.
func (vr *ValidationResult) WarningCount() int {
	return len(vr.Warnings)
}

// InfoCount returns the total number of info messages.
func (vr *ValidationResult) InfoCount() int {
	return len(vr.Infos)
}

// Count returns the total number of all messages (errors + warnings + infos).
func (vr *ValidationResult) Count() int {
	return len(vr.Errors) + len(vr.Warnings) + len(vr.Infos)
}

// IsValid returns true if the validation result contains no error messages.
// Warnings and infos do not affect validity.
func (vr *ValidationResult) IsValid() bool {
	return len(vr.Errors) == 0
}

// SetContext stores a value in the context map with the given key.
// If the key already exists, its value is overwritten.
func (vr *ValidationResult) SetContext(key string, value interface{}) {
	vr.Context[key] = value
}

// GetContext retrieves a value from the context map by key.
// It returns the value and a boolean indicating whether the key exists.
// If the key does not exist, it returns nil and false.
func (vr *ValidationResult) GetContext(key string) (interface{}, bool) {
	val, ok := vr.Context[key]
	return val, ok
}
