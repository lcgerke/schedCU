package validation

import (
	"encoding/json"
	"fmt"
)

// Severity levels for validation messages
type Severity string

const (
	SeverityError   Severity = "ERROR"   // Cannot import/promote
	SeverityWarning Severity = "WARNING" // Can import but should review
	SeverityInfo    Severity = "INFO"    // Informational
)

// Result represents structured validation with severity levels
// Collects all errors/warnings, not fail-fast
type Result struct {
	Messages []Message `json:"messages"`
}

// Message represents a single validation message
type Message struct {
	Severity Severity               `json:"severity"`
	Code     string                 `json:"code"`
	Text     string                 `json:"text"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// NewResult creates a new empty validation result
func NewResult() *Result {
	return &Result{
		Messages: []Message{},
	}
}

// AddError adds an error message (cannot import/promote)
func (r *Result) AddError(code, text string) *Result {
	return r.Add(SeverityError, code, text, nil)
}

// AddErrorWithContext adds an error with additional context
func (r *Result) AddErrorWithContext(code, text string, context map[string]interface{}) *Result {
	return r.Add(SeverityError, code, text, context)
}

// AddWarning adds a warning message (can import but should review)
func (r *Result) AddWarning(code, text string) *Result {
	return r.Add(SeverityWarning, code, text, nil)
}

// AddWarningWithContext adds a warning with additional context
func (r *Result) AddWarningWithContext(code, text string, context map[string]interface{}) *Result {
	return r.Add(SeverityWarning, code, text, context)
}

// AddInfo adds an informational message
func (r *Result) AddInfo(code, text string) *Result {
	return r.Add(SeverityInfo, code, text, nil)
}

// Add adds a message with given severity
func (r *Result) Add(severity Severity, code, text string, context map[string]interface{}) *Result {
	r.Messages = append(r.Messages, Message{
		Severity: severity,
		Code:     code,
		Text:     text,
		Context:  context,
	})
	return r
}

// AddMessages adds multiple messages from another result
func (r *Result) AddMessages(messages ...Message) *Result {
	r.Messages = append(r.Messages, messages...)
	return r
}

// IsValid returns true if no ERROR messages
func (r *Result) IsValid() bool {
	for _, msg := range r.Messages {
		if msg.Severity == SeverityError {
			return false
		}
	}
	return true
}

// CanImport returns true if no ERROR messages (can import)
func (r *Result) CanImport() bool {
	return r.IsValid()
}

// CanPromote returns true if no ERROR or WARNING messages (ready for production)
func (r *Result) CanPromote() bool {
	for _, msg := range r.Messages {
		if msg.Severity == SeverityError || msg.Severity == SeverityWarning {
			return false
		}
	}
	return true
}

// ErrorCount returns number of error messages
func (r *Result) ErrorCount() int {
	count := 0
	for _, msg := range r.Messages {
		if msg.Severity == SeverityError {
			count++
		}
	}
	return count
}

// WarningCount returns number of warning messages
func (r *Result) WarningCount() int {
	count := 0
	for _, msg := range r.Messages {
		if msg.Severity == SeverityWarning {
			count++
		}
	}
	return count
}

// InfoCount returns number of info messages
func (r *Result) InfoCount() int {
	count := 0
	for _, msg := range r.Messages {
		if msg.Severity == SeverityInfo {
			count++
		}
	}
	return count
}

// HasErrors returns true if any errors exist
func (r *Result) HasErrors() bool {
	return r.ErrorCount() > 0
}

// HasWarnings returns true if any warnings exist
func (r *Result) HasWarnings() bool {
	return r.WarningCount() > 0
}

// MessagesByCode returns all messages for a given code
func (r *Result) MessagesByCode(code string) []Message {
	var result []Message
	for _, msg := range r.Messages {
		if msg.Code == code {
			result = append(result, msg)
		}
	}
	return result
}

// MessagesBySeverity returns all messages for a given severity
func (r *Result) MessagesBySeverity(severity Severity) []Message {
	var result []Message
	for _, msg := range r.Messages {
		if msg.Severity == severity {
			result = append(result, msg)
		}
	}
	return result
}

// ToJSON marshals the result to JSON
func (r *Result) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON unmarshals JSON into a result
func FromJSON(jsonStr string) (*Result, error) {
	var r Result
	err := json.Unmarshal([]byte(jsonStr), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// Summary returns a human-readable summary
func (r *Result) Summary() string {
	if len(r.Messages) == 0 {
		return "Validation passed: no errors"
	}

	errorCount := r.ErrorCount()
	warningCount := r.WarningCount()
	infoCount := r.InfoCount()

	summary := fmt.Sprintf("Validation result: %d errors, %d warnings, %d info messages",
		errorCount, warningCount, infoCount)

	if errorCount > 0 {
		summary += "\n\nErrors:"
		for _, msg := range r.MessagesBySeverity(SeverityError) {
			summary += fmt.Sprintf("\n  - %s: %s", msg.Code, msg.Text)
		}
	}

	if warningCount > 0 {
		summary += "\n\nWarnings:"
		for _, msg := range r.MessagesBySeverity(SeverityWarning) {
			summary += fmt.Sprintf("\n  - %s: %s", msg.Code, msg.Text)
		}
	}

	return summary
}

// Real-world examples from v1 system
// ValidationResult examples:
// ERROR: "UNKNOWN_SHIFT_TYPE" - "Unknown shift type: XRAY_NIGHT on 2024-10-15"
// WARNING: "MISSING_MIDC" - "No MidC assignment on weekday 2024-10-16"
// ERROR: "UNKNOWN_PEOPLE" - "Unknown people in Amion: John D, Dr. Smith"

// KnownCodes for common validation issues
const (
	CodeUnknownShiftType  = "UNKNOWN_SHIFT_TYPE"
	CodeMissingMidC       = "MISSING_MIDC"
	CodeUnknownPeople     = "UNKNOWN_PEOPLE"
	CodeParseFailed       = "PARSE_FAILED"
	CodeInvalidFileType   = "INVALID_FILE_TYPE"
	CodeFileTooLarge      = "FILE_TOO_LARGE"
	CodeXXEAttack         = "XXE_ATTACK"
	CodeDuplicateAssignment = "DUPLICATE_ASSIGNMENT"
	CodeInvalidDateRange  = "INVALID_DATE_RANGE"
)
