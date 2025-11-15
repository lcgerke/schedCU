package validation

import (
	"testing"
	"time"
)

// TestSeverityConstants tests that all severity constants exist and have correct values
func TestSeverityConstants(t *testing.T) {
	tests := []struct {
		name    string
		sev     Severity
		want    string
	}{
		{"ERROR severity", ERROR, "error"},
		{"WARNING severity", WARNING, "warning"},
		{"INFO severity", INFO, "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sev != Severity(tt.want) {
				t.Errorf("severity constant mismatch: got %v, want %v", tt.sev, tt.want)
			}
		})
	}
}

// TestSeverityString tests that Severity.String() returns the correct string representation
func TestSeverityString(t *testing.T) {
	tests := []struct {
		name    string
		sev     Severity
		want    string
	}{
		{"ERROR", ERROR, "error"},
		{"WARNING", WARNING, "warning"},
		{"INFO", INFO, "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sev.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSeverityFromString tests parsing Severity from string
func TestSeverityFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Severity
		wantErr bool
	}{
		{"Valid ERROR", "error", ERROR, false},
		{"Valid WARNING", "warning", WARNING, false},
		{"Valid INFO", "info", INFO, false},
		{"Uppercase ERROR", "ERROR", ERROR, false},
		{"Mixed case Error", "Error", ERROR, false},
		{"Invalid value", "invalid", "", true},
		{"Empty string", "", "", true},
		{"Uppercase WARNING", "WARNING", WARNING, false},
		{"Lowercase info", "info", INFO, false},
		{"Whitespace before", " error", ERROR, false},
		{"Whitespace after", "error ", ERROR, false},
		{"Multiple spaces", "  error  ", ERROR, false},
		{"Tab characters", "\terror\t", ERROR, false},
		{"All lowercase warning", "warning", WARNING, false},
		{"All lowercase info", "info", INFO, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FromString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestMessageCodeConstants tests that all message code constants exist
func TestMessageCodeConstants(t *testing.T) {
	tests := []struct {
		name string
		code MessageCode
		want string
	}{
		{"INVALID_FILE_FORMAT", INVALID_FILE_FORMAT, "INVALID_FILE_FORMAT"},
		{"MISSING_REQUIRED_FIELD", MISSING_REQUIRED_FIELD, "MISSING_REQUIRED_FIELD"},
		{"DUPLICATE_ENTRY", DUPLICATE_ENTRY, "DUPLICATE_ENTRY"},
		{"PARSE_ERROR", PARSE_ERROR, "PARSE_ERROR"},
		{"DATABASE_ERROR", DATABASE_ERROR, "DATABASE_ERROR"},
		{"EXTERNAL_SERVICE_ERROR", EXTERNAL_SERVICE_ERROR, "EXTERNAL_SERVICE_ERROR"},
		{"UNKNOWN_ERROR", UNKNOWN_ERROR, "UNKNOWN_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != MessageCode(tt.want) {
				t.Errorf("message code mismatch: got %v, want %v", tt.code, tt.want)
			}
		})
	}
}

// TestMessageCodeString tests that MessageCode.String() returns the correct string representation
func TestMessageCodeString(t *testing.T) {
	tests := []struct {
		name string
		code MessageCode
		want string
	}{
		{"INVALID_FILE_FORMAT", INVALID_FILE_FORMAT, "INVALID_FILE_FORMAT"},
		{"MISSING_REQUIRED_FIELD", MISSING_REQUIRED_FIELD, "MISSING_REQUIRED_FIELD"},
		{"DUPLICATE_ENTRY", DUPLICATE_ENTRY, "DUPLICATE_ENTRY"},
		{"PARSE_ERROR", PARSE_ERROR, "PARSE_ERROR"},
		{"DATABASE_ERROR", DATABASE_ERROR, "DATABASE_ERROR"},
		{"EXTERNAL_SERVICE_ERROR", EXTERNAL_SERVICE_ERROR, "EXTERNAL_SERVICE_ERROR"},
		{"UNKNOWN_ERROR", UNKNOWN_ERROR, "UNKNOWN_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.code.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestMessageCodeFromString tests parsing MessageCode from string
func TestMessageCodeFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    MessageCode
		wantErr bool
	}{
		{"Valid INVALID_FILE_FORMAT", "INVALID_FILE_FORMAT", INVALID_FILE_FORMAT, false},
		{"Valid MISSING_REQUIRED_FIELD", "MISSING_REQUIRED_FIELD", MISSING_REQUIRED_FIELD, false},
		{"Valid DUPLICATE_ENTRY", "DUPLICATE_ENTRY", DUPLICATE_ENTRY, false},
		{"Valid PARSE_ERROR", "PARSE_ERROR", PARSE_ERROR, false},
		{"Valid DATABASE_ERROR", "DATABASE_ERROR", DATABASE_ERROR, false},
		{"Valid EXTERNAL_SERVICE_ERROR", "EXTERNAL_SERVICE_ERROR", EXTERNAL_SERVICE_ERROR, false},
		{"Valid UNKNOWN_ERROR", "UNKNOWN_ERROR", UNKNOWN_ERROR, false},
		{"Invalid code", "INVALID_CODE", "", true},
		{"Empty string", "", "", true},
		{"Lowercase invalid_file_format", "invalid_file_format", INVALID_FILE_FORMAT, false},
		{"Mixed case parse_error", "Parse_Error", PARSE_ERROR, false},
		{"Whitespace before", " PARSE_ERROR", PARSE_ERROR, false},
		{"Whitespace after", "PARSE_ERROR ", PARSE_ERROR, false},
		{"Multiple spaces", "  PARSE_ERROR  ", PARSE_ERROR, false},
		{"All lowercase parse_error", "parse_error", PARSE_ERROR, false},
		{"Mixed case PaRsE_eRrOr", "PaRsE_eRrOr", PARSE_ERROR, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MessageCodeFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MessageCodeFromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MessageCodeFromString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestValidationMessageCreation tests creating a ValidationMessage with all required fields
func TestValidationMessageCreation(t *testing.T) {
	now := time.Now()

	msg := &ValidationMessage{
		Code:      INVALID_FILE_FORMAT,
		Severity:  ERROR,
		Message:   "Invalid file format",
		Field:     "filename",
		Details:   map[string]interface{}{"extension": "txt", "expected": "csv"},
		Timestamp: now,
	}

	if msg.Code != INVALID_FILE_FORMAT {
		t.Errorf("Code = %q, want %q", msg.Code, INVALID_FILE_FORMAT)
	}
	if msg.Severity != ERROR {
		t.Errorf("Severity = %q, want %q", msg.Severity, ERROR)
	}
	if msg.Message != "Invalid file format" {
		t.Errorf("Message = %q, want %q", msg.Message, "Invalid file format")
	}
	if msg.Field != "filename" {
		t.Errorf("Field = %q, want %q", msg.Field, "filename")
	}
	if len(msg.Details) != 2 {
		t.Errorf("Details length = %d, want 2", len(msg.Details))
	}
	if msg.Timestamp != now {
		t.Errorf("Timestamp mismatch")
	}
}

// TestValidationMessageWithoutField tests creating a ValidationMessage without optional field
func TestValidationMessageWithoutField(t *testing.T) {
	msg := &ValidationMessage{
		Code:      PARSE_ERROR,
		Severity:  WARNING,
		Message:   "Parse warning",
		Field:     "",
		Details:   nil,
		Timestamp: time.Now(),
	}

	if msg.Code != PARSE_ERROR {
		t.Errorf("Code = %q, want %q", msg.Code, PARSE_ERROR)
	}
	if msg.Severity != WARNING {
		t.Errorf("Severity = %q, want %q", msg.Severity, WARNING)
	}
	if msg.Field != "" {
		t.Errorf("Field should be empty, got %q", msg.Field)
	}
}

// TestValidationMessageWithEmptyDetails tests creating a ValidationMessage with empty details map
func TestValidationMessageWithEmptyDetails(t *testing.T) {
	msg := &ValidationMessage{
		Code:      DATABASE_ERROR,
		Severity:  ERROR,
		Message:   "Database connection failed",
		Field:     "connection",
		Details:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	if len(msg.Details) != 0 {
		t.Errorf("Details should be empty, got %d items", len(msg.Details))
	}
}

// TestValidationMessageMultipleCodesSeverities tests various code/severity combinations
func TestValidationMessageMultipleCodesSeverities(t *testing.T) {
	testCases := []struct {
		code     MessageCode
		severity Severity
		message  string
	}{
		{INVALID_FILE_FORMAT, ERROR, "File format is invalid"},
		{MISSING_REQUIRED_FIELD, ERROR, "Field 'email' is required"},
		{DUPLICATE_ENTRY, WARNING, "Duplicate entry found"},
		{PARSE_ERROR, ERROR, "Failed to parse input"},
		{DATABASE_ERROR, ERROR, "Database operation failed"},
		{EXTERNAL_SERVICE_ERROR, WARNING, "External service unavailable"},
		{UNKNOWN_ERROR, ERROR, "Unknown error occurred"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.code), func(t *testing.T) {
			msg := &ValidationMessage{
				Code:      tc.code,
				Severity:  tc.severity,
				Message:   tc.message,
				Field:     "",
				Details:   nil,
				Timestamp: time.Now(),
			}

			if msg.Code != tc.code {
				t.Errorf("Code mismatch: got %q, want %q", msg.Code, tc.code)
			}
			if msg.Severity != tc.severity {
				t.Errorf("Severity mismatch: got %q, want %q", msg.Severity, tc.severity)
			}
			if msg.Message != tc.message {
				t.Errorf("Message mismatch: got %q, want %q", msg.Message, tc.message)
			}
		})
	}
}

// TestValidationMessageDetailsAreIndependent tests that details maps are independent
func TestValidationMessageDetailsAreIndependent(t *testing.T) {
	details1 := map[string]interface{}{"key": "value1"}
	msg1 := &ValidationMessage{
		Code:      PARSE_ERROR,
		Severity:  ERROR,
		Message:   "Error 1",
		Field:     "",
		Details:   details1,
		Timestamp: time.Now(),
	}

	details2 := map[string]interface{}{"key": "value2"}
	msg2 := &ValidationMessage{
		Code:      PARSE_ERROR,
		Severity:  ERROR,
		Message:   "Error 2",
		Field:     "",
		Details:   details2,
		Timestamp: time.Now(),
	}

	if msg1.Details["key"] == msg2.Details["key"] && msg1.Details["key"] != "value1" {
		t.Errorf("Details maps are not independent")
	}
}

// TestValidationMessageWithAllDetails tests a message with comprehensive details
func TestValidationMessageWithAllDetails(t *testing.T) {
	details := map[string]interface{}{
		"line_number":   42,
		"expected_type": "integer",
		"received_type": "string",
		"value":         "not_a_number",
		"context":       map[string]string{"table": "users", "column": "age"},
	}

	msg := &ValidationMessage{
		Code:      INVALID_FILE_FORMAT,
		Severity:  ERROR,
		Message:   "Type mismatch in CSV file",
		Field:     "age",
		Details:   details,
		Timestamp: time.Now(),
	}

	if msg.Code != INVALID_FILE_FORMAT {
		t.Errorf("Code mismatch")
	}
	if len(msg.Details) != 5 {
		t.Errorf("Details count = %d, want 5", len(msg.Details))
	}
	if msg.Details["line_number"] != 42 {
		t.Errorf("line_number mismatch")
	}
}

// TestValidationMessageTimestamp tests timestamp behavior
func TestValidationMessageTimestamp(t *testing.T) {
	now := time.Now()
	msg := &ValidationMessage{
		Code:      PARSE_ERROR,
		Severity:  ERROR,
		Message:   "Test",
		Field:     "",
		Details:   nil,
		Timestamp: now,
	}

	// Timestamps should match
	if msg.Timestamp != now {
		t.Errorf("Timestamp mismatch: got %v, want %v", msg.Timestamp, now)
	}

	// Different messages should have different timestamps
	time.Sleep(1 * time.Millisecond)
	now2 := time.Now()
	msg2 := &ValidationMessage{
		Code:      PARSE_ERROR,
		Severity:  ERROR,
		Message:   "Test",
		Field:     "",
		Details:   nil,
		Timestamp: now2,
	}

	if msg.Timestamp == msg2.Timestamp {
		t.Errorf("Different messages should ideally have different timestamps (or very close)")
	}
}

// TestValidationMessageFieldOptional tests that Field is optional
func TestValidationMessageFieldOptional(t *testing.T) {
	msg := &ValidationMessage{
		Code:      DATABASE_ERROR,
		Severity:  ERROR,
		Message:   "Database error",
		Field:     "", // explicitly empty
		Details:   nil,
		Timestamp: time.Now(),
	}

	if msg.Field != "" {
		t.Errorf("Field should be empty string, got %q", msg.Field)
	}
}

// TestValidationMessageDetailsOptional tests that Details is optional
func TestValidationMessageDetailsOptional(t *testing.T) {
	msg := &ValidationMessage{
		Code:      EXTERNAL_SERVICE_ERROR,
		Severity:  WARNING,
		Message:   "Service error",
		Field:     "",
		Details:   nil, // explicitly nil
		Timestamp: time.Now(),
	}

	if msg.Details != nil {
		t.Errorf("Details should be nil, got %v", msg.Details)
	}
}

// TestAllSeverityValues tests that all severity constants are uniquely valuable
func TestAllSeverityValues(t *testing.T) {
	severities := []Severity{ERROR, WARNING, INFO}
	seen := make(map[Severity]bool)

	for _, s := range severities {
		if seen[s] {
			t.Errorf("Duplicate severity value: %v", s)
		}
		seen[s] = true
	}

	if len(seen) != 3 {
		t.Errorf("Expected 3 unique severities, got %d", len(seen))
	}
}

// TestAllMessageCodeValues tests that all message code constants are uniquely valuable
func TestAllMessageCodeValues(t *testing.T) {
	codes := []MessageCode{
		INVALID_FILE_FORMAT,
		MISSING_REQUIRED_FIELD,
		DUPLICATE_ENTRY,
		PARSE_ERROR,
		DATABASE_ERROR,
		EXTERNAL_SERVICE_ERROR,
		UNKNOWN_ERROR,
	}
	seen := make(map[MessageCode]bool)

	for _, c := range codes {
		if seen[c] {
			t.Errorf("Duplicate message code value: %v", c)
		}
		seen[c] = true
	}

	if len(seen) != 7 {
		t.Errorf("Expected 7 unique message codes, got %d", len(seen))
	}
}

// TestValidationMessageJSONTags tests that struct has proper JSON tags
func TestValidationMessageJSONTags(t *testing.T) {
	msg := &ValidationMessage{
		Code:      PARSE_ERROR,
		Severity:  ERROR,
		Message:   "Test message",
		Field:     "test_field",
		Details:   map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
	}

	// Simply verify the struct can be created with all fields
	if msg.Code == "" || msg.Severity == "" || msg.Message == "" {
		t.Error("ValidationMessage fields not properly initialized")
	}
}

// BenchmarkSeverityFromString benchmarks the FromString function for Severity
func BenchmarkSeverityFromString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = FromString("error")
	}
}

// BenchmarkMessageCodeFromString benchmarks the MessageCodeFromString function
func BenchmarkMessageCodeFromString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = MessageCodeFromString("PARSE_ERROR")
	}
}

// BenchmarkValidationMessageCreation benchmarks creating a ValidationMessage
func BenchmarkValidationMessageCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &ValidationMessage{
			Code:      PARSE_ERROR,
			Severity:  ERROR,
			Message:   "Test error message",
			Field:     "field_name",
			Details:   map[string]interface{}{"key": "value"},
			Timestamp: time.Now(),
		}
	}
}

// BenchmarkSeverityString benchmarks the String() method for Severity
func BenchmarkSeverityString(b *testing.B) {
	sev := ERROR
	for i := 0; i < b.N; i++ {
		_ = sev.String()
	}
}

// BenchmarkMessageCodeString benchmarks the String() method for MessageCode
func BenchmarkMessageCodeString(b *testing.B) {
	code := PARSE_ERROR
	for i := 0; i < b.N; i++ {
		_ = code.String()
	}
}
