package api

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/schedcu/reimplement/internal/validation"
)

// TestFormatValidationErrors_SimpleError tests formatting a single error
func TestFormatValidationErrors_SimpleError(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("username", "Username is required")

	errorDetail := FormatValidationErrors(vr)

	if errorDetail == nil {
		t.Fatal("expected non-nil ErrorDetail")
	}

	if errorDetail.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", errorDetail.Code)
	}

	if errorDetail.Message != "Validation failed: 1 error(s)" {
		t.Errorf("expected summary message, got %s", errorDetail.Message)
	}

	if len(errorDetail.Details) == 0 {
		t.Fatal("expected non-empty details")
	}

	// Check error count in context
	if errorDetail.Details["error_count"] != 1 {
		t.Errorf("expected error_count=1, got %v", errorDetail.Details["error_count"])
	}

	// Check errors map exists
	if errorDetail.Details["errors"] == nil {
		t.Fatal("expected errors map in details")
	}

	errorsMap := errorDetail.Details["errors"].(map[string]interface{})
	if len(errorsMap) != 1 {
		t.Errorf("expected 1 error in map, got %d", len(errorsMap))
	}

	if errorsMap["username"] == nil {
		t.Fatal("expected username field in errors map")
	}
}

// TestFormatValidationErrors_MultipleErrors tests formatting multiple errors
func TestFormatValidationErrors_MultipleErrors(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("username", "Username is required")
	vr.AddError("email", "Email is required")
	vr.AddError("password", "Password must be at least 8 characters")

	errorDetail := FormatValidationErrors(vr)

	if errorDetail.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", errorDetail.Code)
	}

	if errorDetail.Details["error_count"] != 3 {
		t.Errorf("expected error_count=3, got %v", errorDetail.Details["error_count"])
	}

	errorsMap := errorDetail.Details["errors"].(map[string]interface{})
	if len(errorsMap) != 3 {
		t.Errorf("expected 3 errors in map, got %d", len(errorsMap))
	}

	if errorsMap["username"] == nil || errorsMap["email"] == nil || errorsMap["password"] == nil {
		t.Fatal("expected all three fields in errors map")
	}
}

// TestFormatValidationErrors_WithWarnings tests formatting with both errors and warnings
func TestFormatValidationErrors_WithWarnings(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("username", "Username is required")
	vr.AddWarning("email", "Email not verified")
	vr.AddWarning("phone", "Phone number not verified")

	errorDetail := FormatValidationErrors(vr)

	if errorDetail.Details["error_count"] != 1 {
		t.Errorf("expected error_count=1, got %v", errorDetail.Details["error_count"])
	}

	if errorDetail.Details["warning_count"] != 2 {
		t.Errorf("expected warning_count=2, got %v", errorDetail.Details["warning_count"])
	}

	warningsMap := errorDetail.Details["warnings"].(map[string]interface{})
	if len(warningsMap) != 2 {
		t.Errorf("expected 2 warnings in map, got %d", len(warningsMap))
	}
}

// TestFormatValidationErrors_WithInfos tests formatting with errors, warnings, and infos
func TestFormatValidationErrors_WithInfos(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("field1", "Field1 error")
	vr.AddWarning("field2", "Field2 warning")
	vr.AddInfo("field3", "Field3 info")

	errorDetail := FormatValidationErrors(vr)

	if errorDetail.Details["error_count"] != 1 {
		t.Errorf("expected error_count=1, got %v", errorDetail.Details["error_count"])
	}

	if errorDetail.Details["warning_count"] != 1 {
		t.Errorf("expected warning_count=1, got %v", errorDetail.Details["warning_count"])
	}

	if errorDetail.Details["info_count"] != 1 {
		t.Errorf("expected info_count=1, got %v", errorDetail.Details["info_count"])
	}

	// Verify all maps exist
	if errorDetail.Details["errors"] == nil {
		t.Fatal("expected errors map")
	}
	if errorDetail.Details["warnings"] == nil {
		t.Fatal("expected warnings map")
	}
	if errorDetail.Details["infos"] == nil {
		t.Fatal("expected infos map")
	}
}

// TestFormatValidationErrors_EmptyValidationResult tests formatting an empty validation result
func TestFormatValidationErrors_EmptyValidationResult(t *testing.T) {
	vr := validation.NewValidationResult()

	errorDetail := FormatValidationErrors(vr)

	if errorDetail == nil {
		t.Fatal("expected non-nil ErrorDetail even for empty validation result")
	}

	if errorDetail.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", errorDetail.Code)
	}

	if errorDetail.Details["error_count"] != 0 {
		t.Errorf("expected error_count=0, got %v", errorDetail.Details["error_count"])
	}
}

// TestFormatValidationErrors_ManyErrors tests formatting with large number of errors (100+)
func TestFormatValidationErrors_ManyErrors(t *testing.T) {
	vr := validation.NewValidationResult()
	expectedCount := 150

	for i := 0; i < expectedCount; i++ {
		fieldName := fmt.Sprintf("field_%d", i)
		vr.AddError(fieldName, "Error for field "+fieldName)
	}

	errorDetail := FormatValidationErrors(vr)

	if errorDetail.Details["error_count"] != expectedCount {
		t.Errorf("expected error_count=%d, got %v", expectedCount, errorDetail.Details["error_count"])
	}

	errorsMap := errorDetail.Details["errors"].(map[string]interface{})
	if len(errorsMap) != expectedCount {
		t.Errorf("expected %d errors in map, got %d", expectedCount, len(errorsMap))
	}

	// Verify first and last errors are present
	if errorsMap["field_0"] == nil {
		t.Fatal("expected first error in map")
	}
	if errorsMap["field_149"] == nil {
		t.Fatal("expected last error in map")
	}
}

// TestFormatValidationErrors_SameFieldMultipleErrors tests multiple errors for same field
func TestFormatValidationErrors_SameFieldMultipleErrors(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("email", "Email is required")
	vr.AddError("email", "Email must be valid")

	errorDetail := FormatValidationErrors(vr)

	if errorDetail.Details["error_count"] != 2 {
		t.Errorf("expected error_count=2, got %v", errorDetail.Details["error_count"])
	}

	errorsMap := errorDetail.Details["errors"].(map[string]interface{})
	if len(errorsMap) != 1 {
		t.Errorf("expected 1 field key in errors map (but with multiple messages), got %d", len(errorsMap))
	}

	// Email field should contain a list/array of messages
	emailErrors := errorsMap["email"]
	if emailErrors == nil {
		t.Fatal("expected email field in errors map")
	}

	// Should be a list of messages
	emailList, ok := emailErrors.([]interface{})
	if !ok {
		t.Fatalf("expected email errors to be a list, got type %T", emailErrors)
	}

	if len(emailList) != 2 {
		t.Errorf("expected 2 messages for email field, got %d", len(emailList))
	}
}

// TestFormatValidationErrors_ContextPreservation tests that context is preserved
func TestFormatValidationErrors_ContextPreservation(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("file", "File parse failed")
	vr.SetContext("filename", "schedule.ods")
	vr.SetContext("line_number", 42)

	errorDetail := FormatValidationErrors(vr)

	if errorDetail.Details["context"] == nil {
		t.Fatal("expected context in error details")
	}

	ctx := errorDetail.Details["context"].(map[string]interface{})
	if ctx["filename"] != "schedule.ods" {
		t.Errorf("expected filename context, got %v", ctx["filename"])
	}

	if ctx["line_number"] != 42 {
		t.Errorf("expected line_number context, got %v", ctx["line_number"])
	}
}

// TestFormatValidationErrors_JSONSerialization tests JSON marshaling of formatted errors
func TestFormatValidationErrors_JSONSerialization(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("username", "Username is required")
	vr.AddWarning("email", "Email not verified")
	vr.SetContext("operation", "user_registration")

	errorDetail := FormatValidationErrors(vr)

	jsonBytes, err := json.Marshal(errorDetail)
	if err != nil {
		t.Fatalf("unexpected error marshaling to JSON: %v", err)
	}

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling JSON: %v", err)
	}

	if unmarshaled["code"] != "VALIDATION_ERROR" {
		t.Error("expected code field in JSON")
	}

	if unmarshaled["message"] == nil {
		t.Error("expected message field in JSON")
	}

	if unmarshaled["details"] == nil {
		t.Error("expected details field in JSON")
	}
}

// TestFormatValidationErrors_IntegrationWithApiResponse tests error formatter integration with ApiResponse
func TestFormatValidationErrors_IntegrationWithApiResponse(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("email", "Email is required")
	vr.AddError("password", "Password too short")

	errorDetail := FormatValidationErrors(vr)
	response := NewApiResponse("data").WithError(errorDetail.Code, errorDetail.Message).WithErrorDetails(errorDetail.Details)

	if response.Error == nil {
		t.Fatal("expected error to be set in response")
	}

	if response.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code VALIDATION_ERROR, got %s", response.Error.Code)
	}

	if response.Error.Details == nil {
		t.Fatal("expected error details to be set")
	}

	details := response.Error.Details
	if details["error_count"] != 2 {
		t.Errorf("expected error_count=2 in response details, got %v", details["error_count"])
	}
}

// TestFormatValidationErrors_NestedErrorStructure tests the hierarchy of formatted errors
func TestFormatValidationErrors_NestedErrorStructure(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("field1", "Error 1")
	vr.AddError("field2", "Error 2")
	vr.AddError("field3", "Error 3")
	vr.AddWarning("field4", "Warning 1")
	vr.SetContext("request_id", "req-123")

	errorDetail := FormatValidationErrors(vr)

	// Verify top-level structure
	if errorDetail.Code != "VALIDATION_ERROR" {
		t.Error("expected VALIDATION_ERROR code at top level")
	}

	details := errorDetail.Details
	if details == nil {
		t.Fatal("expected non-nil details map")
	}

	// Verify detail structure
	if details["error_count"] == nil || details["error_count"] != 3 {
		t.Error("expected error_count in details")
	}

	if details["warning_count"] == nil || details["warning_count"] != 1 {
		t.Error("expected warning_count in details")
	}

	if details["errors"] == nil {
		t.Error("expected errors map in details")
	}

	if details["warnings"] == nil {
		t.Error("expected warnings map in details")
	}

	if details["context"] == nil {
		t.Error("expected context in details")
	}

	// Verify errors map
	errorsMap := details["errors"].(map[string]interface{})
	if len(errorsMap) != 3 {
		t.Errorf("expected 3 fields in errors map, got %d", len(errorsMap))
	}
}

// TestFormatValidationErrors_ErrorMessageContent tests the actual error message content
func TestFormatValidationErrors_ErrorMessageContent(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("username", "Username is required")

	errorDetail := FormatValidationErrors(vr)

	errorsMap := errorDetail.Details["errors"].(map[string]interface{})
	usernameError := errorsMap["username"]

	// Should contain the error message
	if usernameError == nil {
		t.Fatal("expected username error")
	}

	// Could be string or array depending on implementation
	switch v := usernameError.(type) {
	case string:
		if v != "Username is required" {
			t.Errorf("expected 'Username is required', got %s", v)
		}
	case []interface{}:
		if len(v) == 0 || v[0] != "Username is required" {
			t.Errorf("expected first message to be 'Username is required', got %v", v)
		}
	default:
		t.Errorf("expected string or array, got %T", v)
	}
}

// TestFormatValidationErrors_DuplicateErrorsAggregation tests how duplicate errors are handled
func TestFormatValidationErrors_DuplicateErrorsAggregation(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("field", "Error message")
	vr.AddError("field", "Error message") // Duplicate

	errorDetail := FormatValidationErrors(vr)

	// Should have 2 errors in count
	if errorDetail.Details["error_count"] != 2 {
		t.Errorf("expected error_count=2, got %v", errorDetail.Details["error_count"])
	}

	// Should have single field key with both messages
	errorsMap := errorDetail.Details["errors"].(map[string]interface{})
	fieldErrors := errorsMap["field"]

	fieldList, ok := fieldErrors.([]interface{})
	if !ok {
		t.Fatalf("expected array for field errors, got %T", fieldErrors)
	}

	if len(fieldList) != 2 {
		t.Errorf("expected 2 messages for field, got %d", len(fieldList))
	}
}

// TestFormatValidationErrors_MessageSummarization tests message summary accuracy
func TestFormatValidationErrors_MessageSummarization(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(*validation.ValidationResult)
		expected string
	}{
		{
			name: "single error",
			setup: func(vr *validation.ValidationResult) {
				vr.AddError("field", "error")
			},
			expected: "Validation failed: 1 error(s)",
		},
		{
			name: "multiple errors",
			setup: func(vr *validation.ValidationResult) {
				vr.AddError("field1", "error1")
				vr.AddError("field2", "error2")
				vr.AddError("field3", "error3")
			},
			expected: "Validation failed: 3 error(s)",
		},
		{
			name: "errors and warnings",
			setup: func(vr *validation.ValidationResult) {
				vr.AddError("field1", "error1")
				vr.AddWarning("field2", "warning1")
			},
			expected: "Validation failed: 1 error(s), 1 warning(s)",
		},
		{
			name: "all types",
			setup: func(vr *validation.ValidationResult) {
				vr.AddError("field1", "error1")
				vr.AddWarning("field2", "warning1")
				vr.AddInfo("field3", "info1")
			},
			expected: "Validation failed: 1 error(s), 1 warning(s), 1 info(s)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vr := validation.NewValidationResult()
			tc.setup(vr)

			errorDetail := FormatValidationErrors(vr)

			if errorDetail.Message != tc.expected {
				t.Errorf("expected message %q, got %q", tc.expected, errorDetail.Message)
			}
		})
	}
}

// TestFormatValidationErrors_EmptyFieldName tests handling of empty field names
func TestFormatValidationErrors_EmptyFieldName(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("", "Global error without field")

	errorDetail := FormatValidationErrors(vr)

	errorsMap := errorDetail.Details["errors"].(map[string]interface{})

	// Should have entry for empty string key or special key
	if len(errorsMap) == 0 {
		t.Fatal("expected at least one error in map")
	}

	// Either empty key or special key like "_global_"
	found := false
	for key := range errorsMap {
		if key == "" || key == "_global_" || key == "global" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected error with empty field name to be preserved")
	}
}
