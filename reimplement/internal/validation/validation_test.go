package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestEmptyValidationResult tests that a newly created ValidationResult is empty
func TestEmptyValidationResult(t *testing.T) {
	vr := NewValidationResult()

	if vr.ErrorCount() != 0 {
		t.Errorf("expected ErrorCount to be 0, got %d", vr.ErrorCount())
	}

	if vr.WarningCount() != 0 {
		t.Errorf("expected WarningCount to be 0, got %d", vr.WarningCount())
	}

	if vr.InfoCount() != 0 {
		t.Errorf("expected InfoCount to be 0, got %d", vr.InfoCount())
	}

	if vr.Count() != 0 {
		t.Errorf("expected Count to be 0, got %d", vr.Count())
	}

	if !vr.IsValid() {
		t.Error("expected IsValid to be true for empty result")
	}

	if vr.HasErrors() {
		t.Error("expected HasErrors to be false for empty result")
	}

	if vr.HasWarnings() {
		t.Error("expected HasWarnings to be false for empty result")
	}
}

// TestAddError tests adding error messages
func TestAddError(t *testing.T) {
	vr := NewValidationResult()

	vr.AddError("field1", "error message 1")
	vr.AddError("field2", "error message 2")

	if vr.ErrorCount() != 2 {
		t.Errorf("expected ErrorCount to be 2, got %d", vr.ErrorCount())
	}

	if vr.Count() != 2 {
		t.Errorf("expected Count to be 2, got %d", vr.Count())
	}

	if !vr.HasErrors() {
		t.Error("expected HasErrors to be true")
	}

	if vr.IsValid() {
		t.Error("expected IsValid to be false when errors exist")
	}

	if len(vr.Errors) != 2 {
		t.Errorf("expected 2 errors in Errors slice, got %d", len(vr.Errors))
	}

	if vr.Errors[0].Field != "field1" || vr.Errors[0].Message != "error message 1" {
		t.Errorf("error message not set correctly: %+v", vr.Errors[0])
	}

	if vr.Errors[1].Field != "field2" || vr.Errors[1].Message != "error message 2" {
		t.Errorf("error message not set correctly: %+v", vr.Errors[1])
	}
}

// TestAddWarning tests adding warning messages
func TestAddWarning(t *testing.T) {
	vr := NewValidationResult()

	vr.AddWarning("field1", "warning message 1")
	vr.AddWarning("field2", "warning message 2")

	if vr.WarningCount() != 2 {
		t.Errorf("expected WarningCount to be 2, got %d", vr.WarningCount())
	}

	if vr.Count() != 2 {
		t.Errorf("expected Count to be 2, got %d", vr.Count())
	}

	if !vr.HasWarnings() {
		t.Error("expected HasWarnings to be true")
	}

	if !vr.IsValid() {
		t.Error("expected IsValid to be true when only warnings exist")
	}

	if vr.HasErrors() {
		t.Error("expected HasErrors to be false when only warnings exist")
	}

	if len(vr.Warnings) != 2 {
		t.Errorf("expected 2 warnings in Warnings slice, got %d", len(vr.Warnings))
	}

	if vr.Warnings[0].Field != "field1" || vr.Warnings[0].Message != "warning message 1" {
		t.Errorf("warning message not set correctly: %+v", vr.Warnings[0])
	}

	if vr.Warnings[1].Field != "field2" || vr.Warnings[1].Message != "warning message 2" {
		t.Errorf("warning message not set correctly: %+v", vr.Warnings[1])
	}
}

// TestAddInfo tests adding info messages
func TestAddInfo(t *testing.T) {
	vr := NewValidationResult()

	vr.AddInfo("source1", "info message 1")
	vr.AddInfo("source2", "info message 2")
	vr.AddInfo("source3", "info message 3")

	if vr.InfoCount() != 3 {
		t.Errorf("expected InfoCount to be 3, got %d", vr.InfoCount())
	}

	if vr.Count() != 3 {
		t.Errorf("expected Count to be 3, got %d", vr.Count())
	}

	if !vr.IsValid() {
		t.Error("expected IsValid to be true when only infos exist")
	}

	if vr.HasErrors() {
		t.Error("expected HasErrors to be false when only infos exist")
	}

	if vr.HasWarnings() {
		t.Error("expected HasWarnings to be false when only infos exist")
	}

	if len(vr.Infos) != 3 {
		t.Errorf("expected 3 infos in Infos slice, got %d", len(vr.Infos))
	}

	if vr.Infos[0].Field != "source1" || vr.Infos[0].Message != "info message 1" {
		t.Errorf("info message not set correctly: %+v", vr.Infos[0])
	}

	if vr.Infos[2].Field != "source3" || vr.Infos[2].Message != "info message 3" {
		t.Errorf("info message not set correctly: %+v", vr.Infos[2])
	}
}

// TestMixedMessages tests adding errors, warnings, and infos together
func TestMixedMessages(t *testing.T) {
	vr := NewValidationResult()

	vr.AddError("field1", "error 1")
	vr.AddWarning("field2", "warning 1")
	vr.AddInfo("source1", "info 1")
	vr.AddError("field3", "error 2")
	vr.AddWarning("field4", "warning 2")

	if vr.ErrorCount() != 2 {
		t.Errorf("expected ErrorCount to be 2, got %d", vr.ErrorCount())
	}

	if vr.WarningCount() != 2 {
		t.Errorf("expected WarningCount to be 2, got %d", vr.WarningCount())
	}

	if vr.InfoCount() != 1 {
		t.Errorf("expected InfoCount to be 1, got %d", vr.InfoCount())
	}

	if vr.Count() != 5 {
		t.Errorf("expected Count to be 5, got %d", vr.Count())
	}

	if !vr.HasErrors() {
		t.Error("expected HasErrors to be true")
	}

	if !vr.HasWarnings() {
		t.Error("expected HasWarnings to be true")
	}

	if vr.IsValid() {
		t.Error("expected IsValid to be false when errors exist")
	}
}

// TestContextStorage tests storing and retrieving context information
func TestContextStorage(t *testing.T) {
	vr := NewValidationResult()

	vr.SetContext("timestamp", "2024-01-01T00:00:00Z")
	vr.SetContext("version", "1.0.0")
	vr.SetContext("user_id", 42)

	ts, ok := vr.GetContext("timestamp")
	if !ok || ts != "2024-01-01T00:00:00Z" {
		t.Errorf("expected to retrieve timestamp context, got %v, ok=%v", ts, ok)
	}

	version, ok := vr.GetContext("version")
	if !ok || version != "1.0.0" {
		t.Errorf("expected to retrieve version context, got %v, ok=%v", version, ok)
	}

	userID, ok := vr.GetContext("user_id")
	if !ok || userID != 42 {
		t.Errorf("expected to retrieve user_id context, got %v, ok=%v", userID, ok)
	}

	nonExistent, ok := vr.GetContext("does_not_exist")
	if ok || nonExistent != nil {
		t.Errorf("expected to not retrieve non-existent context, got %v, ok=%v", nonExistent, ok)
	}
}

// TestContextOverwrite tests that context values can be overwritten
func TestContextOverwrite(t *testing.T) {
	vr := NewValidationResult()

	vr.SetContext("key", "value1")
	val1, _ := vr.GetContext("key")
	if val1 != "value1" {
		t.Errorf("expected 'value1', got %v", val1)
	}

	vr.SetContext("key", "value2")
	val2, _ := vr.GetContext("key")
	if val2 != "value2" {
		t.Errorf("expected 'value2', got %v", val2)
	}
}

// TestValidationMessageStructure tests ValidationMessage fields
func TestValidationMessageStructure(t *testing.T) {
	vr := NewValidationResult()

	vr.AddError("email", "invalid email format")
	vr.AddWarning("password_strength", "password is weak")
	vr.AddInfo("processing", "validation started")

	// Check error message structure
	if vr.Errors[0].Field != "email" {
		t.Errorf("expected Field 'email', got %q", vr.Errors[0].Field)
	}
	if vr.Errors[0].Message != "invalid email format" {
		t.Errorf("expected Message 'invalid email format', got %q", vr.Errors[0].Message)
	}

	// Check warning message structure
	if vr.Warnings[0].Field != "password_strength" {
		t.Errorf("expected Field 'password_strength', got %q", vr.Warnings[0].Field)
	}
	if vr.Warnings[0].Message != "password is weak" {
		t.Errorf("expected Message 'password is weak', got %q", vr.Warnings[0].Message)
	}

	// Check info message structure
	if vr.Infos[0].Field != "processing" {
		t.Errorf("expected Field 'processing', got %q", vr.Infos[0].Field)
	}
	if vr.Infos[0].Message != "validation started" {
		t.Errorf("expected Message 'validation started', got %q", vr.Infos[0].Message)
	}
}

// TestIsValidWithOnlyWarningsAndInfos tests that IsValid returns true with only warnings/infos
func TestIsValidWithOnlyWarningsAndInfos(t *testing.T) {
	vr := NewValidationResult()

	vr.AddWarning("field1", "warning")
	vr.AddInfo("source1", "info")

	if !vr.IsValid() {
		t.Error("expected IsValid to be true when only warnings and infos exist")
	}

	if vr.HasErrors() {
		t.Error("expected HasErrors to be false")
	}
}

// TestMultipleAddCallsPreserveOrder tests that messages maintain insertion order
func TestMultipleAddCallsPreserveOrder(t *testing.T) {
	vr := NewValidationResult()

	vr.AddError("field1", "error1")
	vr.AddError("field2", "error2")
	vr.AddError("field3", "error3")

	if vr.Errors[0].Message != "error1" || vr.Errors[1].Message != "error2" || vr.Errors[2].Message != "error3" {
		t.Error("expected errors to maintain insertion order")
	}
}

// TestZeroValueValidationResult tests behavior with zero-initialized struct
func TestZeroValueValidationResult(t *testing.T) {
	var vr ValidationResult

	// Should handle gracefully even though nil slices exist
	if !vr.IsValid() {
		t.Error("expected IsValid to be true for zero-value result")
	}

	if vr.HasErrors() {
		t.Error("expected HasErrors to be false for zero-value result")
	}

	if vr.HasWarnings() {
		t.Error("expected HasWarnings to be false for zero-value result")
	}

	if vr.Count() != 0 {
		t.Errorf("expected Count to be 0 for zero-value result, got %d", vr.Count())
	}
}

// TestConcurrentSafeAdditions tests that multiple additions work correctly
func TestConcurrentSafeAdditions(t *testing.T) {
	vr := NewValidationResult()

	// Add many messages to test internal capacity
	for i := 0; i < 100; i++ {
		vr.AddError("field", "error")
		if i%3 == 0 {
			vr.AddWarning("field", "warning")
		}
		if i%5 == 0 {
			vr.AddInfo("field", "info")
		}
	}

	if vr.ErrorCount() != 100 {
		t.Errorf("expected 100 errors, got %d", vr.ErrorCount())
	}

	if vr.WarningCount() != 34 {
		t.Errorf("expected 34 warnings, got %d", vr.WarningCount())
	}

	if vr.InfoCount() != 20 {
		t.Errorf("expected 20 infos, got %d", vr.InfoCount())
	}

	expectedCount := 100 + 34 + 20
	if vr.Count() != expectedCount {
		t.Errorf("expected Count to be %d, got %d", expectedCount, vr.Count())
	}
}

// ============================================================================
// ROUND-TRIP MARSHALING TESTS
// ============================================================================

// TestRoundTripEmptyValidationResult tests marshaling and unmarshaling an empty result
func TestRoundTripEmptyValidationResult(t *testing.T) {
	original := NewValidationResult()

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal empty ValidationResult: %v", err)
	}

	// Unmarshal back
	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal empty ValidationResult: %v", err)
	}

	// Verify equality
	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for empty result\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestRoundTripWithErrors tests marshaling/unmarshaling with multiple errors
func TestRoundTripWithErrors(t *testing.T) {
	original := NewValidationResult()
	original.AddError("field1", "error message 1")
	original.AddError("field2", "error message 2")
	original.AddError("field3", "error message 3")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal ValidationResult with errors: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ValidationResult with errors: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for errors\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestRoundTripWithWarnings tests marshaling/unmarshaling with multiple warnings
func TestRoundTripWithWarnings(t *testing.T) {
	original := NewValidationResult()
	original.AddWarning("field1", "warning 1")
	original.AddWarning("field2", "warning 2")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal ValidationResult with warnings: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ValidationResult with warnings: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for warnings\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestRoundTripWithInfos tests marshaling/unmarshaling with multiple infos
func TestRoundTripWithInfos(t *testing.T) {
	original := NewValidationResult()
	original.AddInfo("source1", "info 1")
	original.AddInfo("source2", "info 2")
	original.AddInfo("source3", "info 3")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal ValidationResult with infos: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ValidationResult with infos: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for infos\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestRoundTripMixedMessages tests marshaling/unmarshaling with mixed message types
func TestRoundTripMixedMessages(t *testing.T) {
	original := NewValidationResult()
	original.AddError("email", "invalid email format")
	original.AddError("password", "password too short")
	original.AddWarning("name", "name has special characters")
	original.AddInfo("source", "validation initiated")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal mixed ValidationResult: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal mixed ValidationResult: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for mixed messages\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestRoundTripWithContext tests marshaling/unmarshaling with context data
func TestRoundTripWithContext(t *testing.T) {
	original := NewValidationResult()
	original.AddError("field1", "error 1")
	original.SetContext("user_id", 42)
	original.SetContext("request_id", "req-12345")
	original.SetContext("source", "api")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal ValidationResult with context: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ValidationResult with context: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for context data\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestRoundTripComplexContext tests context with complex nested objects
func TestRoundTripComplexContext(t *testing.T) {
	original := NewValidationResult()
	original.AddError("parsing", "parse error")

	// Add complex context values
	original.SetContext("metadata", map[string]interface{}{
		"version": "1.0.0",
		"build":   "debug",
		"count":   100,
	})
	original.SetContext("flags", []string{"flag1", "flag2", "flag3"})
	original.SetContext("nested", map[string]interface{}{
		"deep": map[string]interface{}{
			"value": 3.14159,
		},
	})

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal ValidationResult with complex context: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ValidationResult with complex context: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for complex context\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestRoundTripWithAllMessageTypes tests all message types simultaneously
func TestRoundTripWithAllMessageTypes(t *testing.T) {
	original := NewValidationResult()

	// Add multiple of each type
	for i := 1; i <= 3; i++ {
		original.AddError(fmt.Sprintf("error_field_%d", i), fmt.Sprintf("error message %d", i))
	}
	for i := 1; i <= 2; i++ {
		original.AddWarning(fmt.Sprintf("warning_field_%d", i), fmt.Sprintf("warning message %d", i))
	}
	for i := 1; i <= 4; i++ {
		original.AddInfo(fmt.Sprintf("info_source_%d", i), fmt.Sprintf("info message %d", i))
	}

	// Add context
	original.SetContext("execution_time", 123.45)
	original.SetContext("hostname", "validator-1")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal complex ValidationResult: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal complex ValidationResult: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for all message types\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// ============================================================================
// JSON STRUCTURE VALIDATION TESTS
// ============================================================================

// TestMarshaledJSONStructure verifies the JSON structure is correct
func TestMarshaledJSONStructure(t *testing.T) {
	original := NewValidationResult()
	original.AddError("email", "invalid format")
	original.AddWarning("password", "weak")
	original.AddInfo("process", "started")
	original.SetContext("request_id", "req123")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Parse as generic JSON to verify structure
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Verify top-level keys exist
	if _, ok := jsonMap["errors"]; !ok {
		t.Error("missing 'errors' key in JSON")
	}
	if _, ok := jsonMap["warnings"]; !ok {
		t.Error("missing 'warnings' key in JSON")
	}
	if _, ok := jsonMap["infos"]; !ok {
		t.Error("missing 'infos' key in JSON")
	}
	if _, ok := jsonMap["context"]; !ok {
		t.Error("missing 'context' key in JSON")
	}

	// Verify structure of errors
	errors, ok := jsonMap["errors"].([]interface{})
	if !ok {
		t.Errorf("'errors' is not an array: %T", jsonMap["errors"])
	} else if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}

	// Verify structure of context
	context, ok := jsonMap["context"].(map[string]interface{})
	if !ok {
		t.Errorf("'context' is not an object: %T", jsonMap["context"])
	} else if context["request_id"] != "req123" {
		t.Errorf("context value mismatch: expected 'req123', got %v", context["request_id"])
	}
}

// TestValidationMessageMarshaling tests individual message marshaling
func TestValidationMessageMarshaling(t *testing.T) {
	msg := ValidationMessage{
		Field:   "email",
		Message: "invalid email format",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	unmarshaled := &ValidationMessage{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if unmarshaled.Field != msg.Field || unmarshaled.Message != msg.Message {
		t.Errorf("message mismatch\noriginal: %+v\nunmarshaled: %+v", msg, unmarshaled)
	}
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

// TestUnmarshalInvalidJSON tests error handling for invalid JSON
func TestUnmarshalInvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{invalid json}`)
	vr := &ValidationResult{}

	err := json.Unmarshal(invalidJSON, vr)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestUnmarshalMissingFields tests unmarshaling with missing optional fields
func TestUnmarshalMissingFields(t *testing.T) {
	// Minimal valid JSON with only errors
	minimalJSON := []byte(`{"errors": [], "warnings": [], "infos": [], "context": {}}`)
	vr := &ValidationResult{}

	if err := json.Unmarshal(minimalJSON, vr); err != nil {
		t.Fatalf("failed to unmarshal minimal JSON: %v", err)
	}

	if vr.ErrorCount() != 0 || vr.WarningCount() != 0 || vr.InfoCount() != 0 {
		t.Error("expected empty result from minimal JSON")
	}
}

// TestUnmarshalEmptyArrays tests unmarshaling empty message arrays
func TestUnmarshalEmptyArrays(t *testing.T) {
	emptyJSON := []byte(`{
		"errors": [],
		"warnings": [],
		"infos": [],
		"context": {}
	}`)

	vr := &ValidationResult{}
	if err := json.Unmarshal(emptyJSON, vr); err != nil {
		t.Fatalf("failed to unmarshal empty arrays: %v", err)
	}

	if vr.Count() != 0 {
		t.Errorf("expected count 0, got %d", vr.Count())
	}
}

// TestUnmarshalWrongArrayType tests error handling for non-array message fields
func TestUnmarshalWrongArrayType(t *testing.T) {
	wrongJSON := []byte(`{
		"errors": "not an array",
		"warnings": [],
		"infos": [],
		"context": {}
	}`)

	vr := &ValidationResult{}
	err := json.Unmarshal(wrongJSON, vr)
	if err == nil {
		t.Error("expected error for wrong array type, got nil")
	}
}

// TestUnmarshalWrongContextType tests error handling for non-object context
func TestUnmarshalWrongContextType(t *testing.T) {
	wrongJSON := []byte(`{
		"errors": [],
		"warnings": [],
		"infos": [],
		"context": "not an object"
	}`)

	vr := &ValidationResult{}
	err := json.Unmarshal(wrongJSON, vr)
	if err == nil {
		t.Error("expected error for wrong context type, got nil")
	}
}

// TestUnmarshalWithMissingOptionalFields tests unmarshaling with missing field and message fields
// Go's JSON unmarshaling is lenient - missing fields are allowed and left as zero values
func TestUnmarshalWithMissingOptionalFields(t *testing.T) {
	// Missing message field is valid - it will be empty string
	missingMessageJSON := []byte(`{
		"errors": [{"field": "test"}],
		"warnings": [],
		"infos": [],
		"context": {}
	}`)

	vr := &ValidationResult{}
	if err := json.Unmarshal(missingMessageJSON, vr); err != nil {
		t.Fatalf("failed to unmarshal with missing message field: %v", err)
	}

	if vr.ErrorCount() != 1 {
		t.Errorf("expected 1 error, got %d", vr.ErrorCount())
	}
	if vr.Errors[0].Field != "test" {
		t.Errorf("expected field 'test', got %q", vr.Errors[0].Field)
	}
	if vr.Errors[0].Message != "" {
		t.Errorf("expected empty message, got %q", vr.Errors[0].Message)
	}
}

// ============================================================================
// EDGE CASE TESTS
// ============================================================================

// TestEmptyFieldAndMessage tests messages with empty field/message strings
func TestEmptyFieldAndMessage(t *testing.T) {
	original := NewValidationResult()
	original.AddError("", "error with empty field")
	original.AddWarning("warning_field", "")
	original.AddInfo("", "")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal with empty fields: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal with empty fields: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for empty fields\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestSpecialCharactersInMessages tests messages with special characters
func TestSpecialCharactersInMessages(t *testing.T) {
	original := NewValidationResult()
	original.AddError("field", `error with "quotes" and\nescapes and	tabs`)
	original.AddWarning("field", "emoji: ⚠️ warning: ⛔")
	original.AddInfo("field", "unicode: 你好世界 مرحبا العالم")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal with special chars: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal with special chars: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for special chars\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestLargeNumberOfMessages tests marshaling many messages
func TestLargeNumberOfMessages(t *testing.T) {
	original := NewValidationResult()

	// Add 100 of each type
	for i := 0; i < 100; i++ {
		original.AddError(fmt.Sprintf("error_%d", i), fmt.Sprintf("error message %d", i))
		original.AddWarning(fmt.Sprintf("warning_%d", i), fmt.Sprintf("warning message %d", i))
		original.AddInfo(fmt.Sprintf("info_%d", i), fmt.Sprintf("info message %d", i))
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal large result: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal large result: %v", err)
	}

	if !resultEquals(original, unmarshaled) {
		t.Errorf("round-trip failed for large result\noriginal: %+v\nunmarshaled: %+v", original, unmarshaled)
	}
}

// TestNullValuesInContext tests context with null values
func TestNullValuesInContext(t *testing.T) {
	original := NewValidationResult()
	original.SetContext("nullable", nil)
	original.SetContext("value", "test")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal with null context: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal with null context: %v", err)
	}

	// Verify the nullable key exists
	nullable, ok := unmarshaled.GetContext("nullable")
	if !ok {
		t.Error("expected 'nullable' key to exist in context")
	}
	if nullable != nil {
		t.Errorf("expected nil value for 'nullable', got %v", nullable)
	}
}

// TestDuplicateMessagesPreserved tests that duplicate messages are preserved
func TestDuplicateMessagesPreserved(t *testing.T) {
	original := NewValidationResult()
	original.AddError("field", "duplicate error")
	original.AddError("field", "duplicate error")
	original.AddError("field", "duplicate error")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal duplicates: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal duplicates: %v", err)
	}

	if unmarshaled.ErrorCount() != 3 {
		t.Errorf("expected 3 errors, got %d", unmarshaled.ErrorCount())
	}
}

// TestMessageOrderPreserved tests that message order is preserved
func TestMessageOrderPreserved(t *testing.T) {
	original := NewValidationResult()
	messages := []string{"first", "second", "third", "fourth", "fifth"}

	for _, msg := range messages {
		original.AddError("field", msg)
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	unmarshaled := &ValidationResult{}
	if err := json.Unmarshal(data, unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	for i, expected := range messages {
		if unmarshaled.Errors[i].Message != expected {
			t.Errorf("message order not preserved at index %d: expected %q, got %q",
				i, expected, unmarshaled.Errors[i].Message)
		}
	}
}

// TestIndentedJSONFormatting tests that MarshalIndent produces properly formatted JSON
func TestIndentedJSONFormatting(t *testing.T) {
	vr := NewValidationResult()
	vr.AddError("field", "error message")
	vr.SetContext("request_id", "req123")

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(vr); err != nil {
		t.Fatalf("failed to encode with indent: %v", err)
	}

	// Verify the output contains newlines and indentation
	if !bytes.Contains(buf.Bytes(), []byte("\n")) {
		t.Error("expected indented JSON to contain newlines")
	}
	if !bytes.Contains(buf.Bytes(), []byte("  ")) {
		t.Error("expected indented JSON to contain spaces")
	}
}

// ============================================================================
// FIXTURE-BASED TESTS
// ============================================================================

// TestLoadFixtureEmpty tests loading an empty validation result from JSON fixture
func TestLoadFixtureEmpty(t *testing.T) {
	fixtureFile := "../../tests/fixtures/validation_empty.json"
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Skipf("fixture file not found: %s", fixtureFile)
	}

	vr := &ValidationResult{}
	if err := json.Unmarshal(data, vr); err != nil {
		t.Fatalf("failed to unmarshal empty fixture: %v", err)
	}

	if vr.ErrorCount() != 0 || vr.WarningCount() != 0 || vr.InfoCount() != 0 {
		t.Error("expected empty result from fixture")
	}
	if vr.Count() != 0 {
		t.Errorf("expected count 0, got %d", vr.Count())
	}
}

// TestLoadFixtureSimple tests loading a simple validation result from JSON fixture
func TestLoadFixtureSimple(t *testing.T) {
	fixtureFile := "../../tests/fixtures/validation_simple.json"
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Skipf("fixture file not found: %s", fixtureFile)
	}

	vr := &ValidationResult{}
	if err := json.Unmarshal(data, vr); err != nil {
		t.Fatalf("failed to unmarshal simple fixture: %v", err)
	}

	// Verify errors
	if vr.ErrorCount() != 2 {
		t.Errorf("expected 2 errors, got %d", vr.ErrorCount())
	}
	if vr.Errors[0].Field != "email" {
		t.Errorf("expected field 'email', got %q", vr.Errors[0].Field)
	}

	// Verify warnings
	if vr.WarningCount() != 1 {
		t.Errorf("expected 1 warning, got %d", vr.WarningCount())
	}

	// Verify infos
	if vr.InfoCount() != 1 {
		t.Errorf("expected 1 info, got %d", vr.InfoCount())
	}

	// Verify context
	userID, ok := vr.GetContext("user_id")
	if !ok || userID != float64(123) {
		t.Errorf("expected user_id context value 123, got %v", userID)
	}

	reqID, ok := vr.GetContext("request_id")
	if !ok || reqID != "req-abc123" {
		t.Errorf("expected request_id context, got %v", reqID)
	}
}

// TestLoadFixtureComplex tests loading a complex validation result from JSON fixture
func TestLoadFixtureComplex(t *testing.T) {
	fixtureFile := "../../tests/fixtures/validation_complex.json"
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Skipf("fixture file not found: %s", fixtureFile)
	}

	vr := &ValidationResult{}
	if err := json.Unmarshal(data, vr); err != nil {
		t.Fatalf("failed to unmarshal complex fixture: %v", err)
	}

	// Verify message counts
	if vr.ErrorCount() != 3 {
		t.Errorf("expected 3 errors, got %d", vr.ErrorCount())
	}
	if vr.WarningCount() != 2 {
		t.Errorf("expected 2 warnings, got %d", vr.WarningCount())
	}
	if vr.InfoCount() != 2 {
		t.Errorf("expected 2 infos, got %d", vr.InfoCount())
	}

	// Verify context metadata
	if vr.Count() != 7 {
		t.Errorf("expected total count 7, got %d", vr.Count())
	}

	// Verify context contains nested structure
	metadata, ok := vr.GetContext("metadata")
	if !ok {
		t.Error("expected 'metadata' context key")
	}
	metadataMap, ok := metadata.(map[string]interface{})
	if !ok {
		t.Errorf("expected metadata to be a map, got %T", metadata)
	} else {
		if metadataMap["version"] != "1.0.0" {
			t.Errorf("expected metadata.version 1.0.0, got %v", metadataMap["version"])
		}
	}

	// Verify context contains array
	flags, ok := vr.GetContext("flags")
	if !ok {
		t.Error("expected 'flags' context key")
	}
	flagsArray, ok := flags.([]interface{})
	if !ok {
		t.Errorf("expected flags to be an array, got %T", flags)
	} else {
		if len(flagsArray) != 3 {
			t.Errorf("expected 3 flags, got %d", len(flagsArray))
		}
	}
}

// TestRoundTripFixtureSimple tests marshaling and unmarshaling with simple fixture data
func TestRoundTripFixtureSimple(t *testing.T) {
	fixtureFile := "../../tests/fixtures/validation_simple.json"
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Skipf("fixture file not found: %s", fixtureFile)
	}

	// Load from fixture
	original := &ValidationResult{}
	if err := json.Unmarshal(data, original); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	// Marshal back to JSON
	marshaled, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal again
	roundtripped := &ValidationResult{}
	if err := json.Unmarshal(marshaled, roundtripped); err != nil {
		t.Fatalf("failed to unmarshal round-trip: %v", err)
	}

	// Verify round-trip equality
	if !resultEquals(original, roundtripped) {
		t.Errorf("round-trip failed\noriginal: %+v\nround-tripped: %+v", original, roundtripped)
	}
}

// TestRoundTripFixtureComplex tests marshaling and unmarshaling with complex fixture data
func TestRoundTripFixtureComplex(t *testing.T) {
	fixtureFile := "../../tests/fixtures/validation_complex.json"
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Skipf("fixture file not found: %s", fixtureFile)
	}

	// Load from fixture
	original := &ValidationResult{}
	if err := json.Unmarshal(data, original); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	// Marshal back to JSON
	marshaled, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal again
	roundtripped := &ValidationResult{}
	if err := json.Unmarshal(marshaled, roundtripped); err != nil {
		t.Fatalf("failed to unmarshal round-trip: %v", err)
	}

	// Verify round-trip equality
	if !resultEquals(original, roundtripped) {
		t.Errorf("round-trip failed\noriginal: %+v\nround-tripped: %+v", original, roundtripped)
	}
}

// TestFixtureDirectoryExists checks if the fixtures directory exists
func TestFixtureDirectoryExists(t *testing.T) {
	fixtureDir := "../../tests/fixtures"
	info, err := os.Stat(fixtureDir)
	if err != nil {
		t.Skipf("fixtures directory does not exist: %s", fixtureDir)
	}
	if !info.IsDir() {
		t.Fatalf("expected fixtures to be a directory, but it's a file")
	}
}

// TestAllFixturesAreValidJSON checks that all fixture files are valid JSON
func TestAllFixturesAreValidJSON(t *testing.T) {
	fixtureDir := "../../tests/fixtures"
	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Skipf("could not read fixtures directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if filepath.Ext(filename) != ".json" {
			continue
		}

		filePath := filepath.Join(fixtureDir, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("could not read fixture file %s: %v", filename, err)
			continue
		}

		vr := &ValidationResult{}
		if err := json.Unmarshal(data, vr); err != nil {
			t.Errorf("fixture %s is not valid JSON: %v", filename, err)
		}
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// resultEquals compares two ValidationResult objects for equality
func resultEquals(a, b *ValidationResult) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare error counts and content
	if len(a.Errors) != len(b.Errors) {
		return false
	}
	for i := range a.Errors {
		if a.Errors[i].Field != b.Errors[i].Field || a.Errors[i].Message != b.Errors[i].Message {
			return false
		}
	}

	// Compare warning counts and content
	if len(a.Warnings) != len(b.Warnings) {
		return false
	}
	for i := range a.Warnings {
		if a.Warnings[i].Field != b.Warnings[i].Field || a.Warnings[i].Message != b.Warnings[i].Message {
			return false
		}
	}

	// Compare info counts and content
	if len(a.Infos) != len(b.Infos) {
		return false
	}
	for i := range a.Infos {
		if a.Infos[i].Field != b.Infos[i].Field || a.Infos[i].Message != b.Infos[i].Message {
			return false
		}
	}

	// Compare context
	if len(a.Context) != len(b.Context) {
		return false
	}
	for k, v := range a.Context {
		bv, ok := b.Context[k]
		if !ok {
			return false
		}
		// For complex types like maps/slices, use JSON comparison for deep equality
		aJSON, _ := json.Marshal(v)
		bJSON, _ := json.Marshal(bv)
		if !bytes.Equal(aJSON, bJSON) {
			return false
		}
	}

	return true
}
