package ods

import (
	"testing"
)

// TestNewODSErrorCollectorInitialization verifies initialization
func TestNewODSErrorCollectorInitialization(t *testing.T) {
	collector := NewODSErrorCollector()

	if collector == nil {
		t.Fatal("expected non-nil collector")
	}

	if collector.HasErrors() {
		t.Error("expected HasErrors to return false for empty collector")
	}

	if collector.HasWarnings() {
		t.Error("expected HasWarnings to return false for empty collector")
	}
}

// TestAddMissingRequiredFieldError tests adding MISSING_REQUIRED_FIELD error
func TestAddMissingRequiredFieldError(t *testing.T) {
	collector := NewODSErrorCollector()

	err := NewParsingError(ErrorTypeMissingRequired, "StudentID is required").
		WithLocation(5, "StudentID").
		WithCellReference("B5")

	collector.AddError(err)

	if !collector.HasErrors() {
		t.Error("expected HasErrors to return true after adding error")
	}

	if collector.ErrorCountParsing() != 1 {
		t.Errorf("expected 1 error, got %d", collector.ErrorCountParsing())
	}

	errors := collector.Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error in Errors slice, got %d", len(errors))
	}

	parsedErr := errors[0]
	if parsedErr.Type != ErrorTypeMissingRequired {
		t.Errorf("expected type %s, got %s", ErrorTypeMissingRequired, parsedErr.Type)
	}
	if parsedErr.Row != 5 {
		t.Errorf("expected row 5, got %d", parsedErr.Row)
	}
	if parsedErr.Column != "StudentID" {
		t.Errorf("expected column StudentID, got %s", parsedErr.Column)
	}
	if parsedErr.CellReference != "B5" {
		t.Errorf("expected cell reference B5, got %s", parsedErr.CellReference)
	}
}

// TestAddInvalidValueError tests adding INVALID_VALUE error
func TestAddInvalidValueError(t *testing.T) {
	collector := NewODSErrorCollector()

	err := NewParsingError(ErrorTypeInvalidValue, "Grade must be an integer").
		WithLocation(10, "Grade").
		WithCellReference("C10").
		WithDetail("expected_type", "integer").
		WithDetail("actual_value", "ABC")

	collector.AddError(err)

	errors := collector.Errors()
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	parsedErr := errors[0]
	if parsedErr.Type != ErrorTypeInvalidValue {
		t.Errorf("expected type %s, got %s", ErrorTypeInvalidValue, parsedErr.Type)
	}
	if parsedErr.Details["expected_type"] != "integer" {
		t.Errorf("expected detail expected_type to be 'integer', got %v", parsedErr.Details["expected_type"])
	}
}

// TestMultipleErrorsDifferentTypes tests adding errors of different types
func TestMultipleErrorsDifferentTypes(t *testing.T) {
	collector := NewODSErrorCollector()

	err1 := NewParsingError(ErrorTypeMissingRequired, "Missing Name").WithLocation(3, "Name")
	err2 := NewParsingError(ErrorTypeInvalidValue, "Invalid Age").WithLocation(4, "Age")
	err3 := NewParsingError(ErrorTypeInvalidFormat, "Invalid Date").WithLocation(6, "Date")
	err4 := NewParsingError(ErrorTypeDuplicate, "Duplicate ID").WithLocation(5, "ID")

	collector.AddError(err1)
	collector.AddError(err2)
	collector.AddError(err3)
	collector.AddError(err4)

	if collector.ErrorCountParsing() != 4 {
		t.Errorf("expected 4 errors, got %d", collector.ErrorCountParsing())
	}

	expectedTypes := []ErrorType{
		ErrorTypeMissingRequired,
		ErrorTypeInvalidValue,
		ErrorTypeInvalidFormat,
		ErrorTypeDuplicate,
	}

	errors := collector.Errors()
	for i, expected := range expectedTypes {
		if errors[i].Type != expected {
			t.Errorf("at index %d, expected type %s, got %s", i, expected, errors[i].Type)
		}
	}
}

// TestAddWarning tests adding a warning
func TestAddWarning(t *testing.T) {
	collector := NewODSErrorCollector()

	warn := NewParsingError(ErrorTypeMissingRow, "File size is unusually large").
		WithLocation(0, "file")

	collector.AddWarning(warn)

	if !collector.HasWarnings() {
		t.Error("expected HasWarnings to return true")
	}

	if collector.WarningCountParsing() != 1 {
		t.Errorf("expected 1 warning, got %d", collector.WarningCountParsing())
	}

	warnings := collector.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
}

// TestGroupErrorsByType tests grouping errors by type
func TestGroupErrorsByType(t *testing.T) {
	collector := NewODSErrorCollector()

	err1 := NewParsingError(ErrorTypeMissingRequired, "Missing Name").WithLocation(3, "Name")
	err2 := NewParsingError(ErrorTypeInvalidValue, "Invalid Age").WithLocation(4, "Age")
	err3 := NewParsingError(ErrorTypeMissingRequired, "Missing Email").WithLocation(5, "Email")
	err4 := NewParsingError(ErrorTypeInvalidValue, "Invalid Score").WithLocation(6, "Score")

	collector.AddError(err1)
	collector.AddError(err2)
	collector.AddError(err3)
	collector.AddError(err4)

	grouped := collector.GroupErrorsByType()

	if len(grouped[ErrorTypeMissingRequired]) != 2 {
		t.Errorf("expected 2 missing required errors, got %d", len(grouped[ErrorTypeMissingRequired]))
	}

	if len(grouped[ErrorTypeInvalidValue]) != 2 {
		t.Errorf("expected 2 invalid value errors, but got %d", len(grouped[ErrorTypeInvalidValue]))
	}
}

// TestGroupErrorsByRow tests grouping errors by row
func TestGroupErrorsByRow(t *testing.T) {
	collector := NewODSErrorCollector()

	err1 := NewParsingError(ErrorTypeMissingRequired, "Missing Name").WithLocation(5, "Name")
	err2 := NewParsingError(ErrorTypeInvalidValue, "Invalid Age").WithLocation(5, "Age")
	err3 := NewParsingError(ErrorTypeMissingRequired, "Missing Email").WithLocation(6, "Email")

	collector.AddError(err1)
	collector.AddError(err2)
	collector.AddError(err3)

	grouped := collector.GroupErrorsByRow()

	if len(grouped[5]) != 2 {
		t.Errorf("expected 2 errors in row 5, got %d", len(grouped[5]))
	}

	if len(grouped[6]) != 1 {
		t.Errorf("expected 1 error in row 6, got %d", len(grouped[6]))
	}
}

// TestToValidationResultEmpty tests converting empty collector
func TestToValidationResultEmpty(t *testing.T) {
	collector := NewODSErrorCollector()

	result := collector.ToValidationResult()

	if result == nil {
		t.Error("expected non-nil ValidationResult")
	}

	if !result.IsValid() {
		t.Error("expected valid result for empty collector")
	}

	if result.ErrorCount() != 0 {
		t.Errorf("expected 0 errors, got %d", result.ErrorCount())
	}
}

// TestToValidationResultWithErrors tests converting with errors
func TestToValidationResultWithErrors(t *testing.T) {
	collector := NewODSErrorCollector()

	err1 := NewParsingError(ErrorTypeMissingRequired, "Missing StudentID").WithLocation(5, "StudentID")
	err2 := NewParsingError(ErrorTypeInvalidValue, "Invalid Grade").WithLocation(10, "Grade")

	collector.AddError(err1)
	collector.AddError(err2)

	result := collector.ToValidationResult()

	if result.IsValid() {
		t.Error("expected invalid result when errors exist")
	}

	if result.ErrorCount() != 2 {
		t.Errorf("expected 2 errors, got %d", result.ErrorCount())
	}
}

// TestMixedErrorsAndWarnings tests with both errors and warnings
func TestMixedErrorsAndWarnings(t *testing.T) {
	collector := NewODSErrorCollector()

	err := NewParsingError(ErrorTypeMissingRequired, "Missing field").WithLocation(5, "Name")
	warn := NewParsingError(ErrorTypeMissingRow, "Warning message")

	collector.AddError(err)
	collector.AddWarning(warn)

	result := collector.ToValidationResult()

	if result.ErrorCount() != 1 {
		t.Errorf("expected 1 error, got %d", result.ErrorCount())
	}

	if result.WarningCount() != 1 {
		t.Errorf("expected 1 warning, got %d", result.WarningCount())
	}

	if result.Count() != 2 {
		t.Errorf("expected total count 2, got %d", result.Count())
	}
}

// TestLargeNumberOfErrors tests with many errors
func TestLargeNumberOfErrors(t *testing.T) {
	collector := NewODSErrorCollector()

	// Add 100 errors
	for i := 0; i < 100; i++ {
		err := NewParsingError(ErrorTypeMissingRequired, "Missing field").
			WithLocation(i+1, "Field")
		collector.AddError(err)
	}

	if collector.ErrorCountParsing() != 100 {
		t.Errorf("expected 100 errors, got %d", collector.ErrorCountParsing())
	}

	result := collector.ToValidationResult()
	if result.ErrorCount() != 100 {
		t.Errorf("expected 100 errors in ValidationResult, got %d", result.ErrorCount())
	}
}

// TestNilErrorHandling tests that nil errors are ignored
func TestNilErrorHandling(t *testing.T) {
	collector := NewODSErrorCollector()

	collector.AddError(nil)

	if collector.HasErrors() {
		t.Error("expected HasErrors to return false when nil is added")
	}

	if collector.ErrorCountParsing() != 0 {
		t.Errorf("expected 0 errors, got %d", collector.ErrorCountParsing())
	}
}

// TestErrorOrderPreservation tests that errors maintain insertion order
func TestErrorOrderPreservation(t *testing.T) {
	collector := NewODSErrorCollector()

	expectedMessages := []string{"error1", "error2", "error3"}

	for i, msg := range expectedMessages {
		err := NewParsingError(ErrorTypeMissingRequired, msg).WithLocation(i+1, "Field")
		collector.AddError(err)
	}

	errors := collector.Errors()

	for i, expected := range expectedMessages {
		if errors[i].Message != expected {
			t.Errorf("at index %d, expected message '%s', got '%s'", i, expected, errors[i].Message)
		}
	}
}

// TestParsingErrorString tests the Error() method output
func TestParsingErrorString(t *testing.T) {
	err := NewParsingError(ErrorTypeMissingRequired, "Missing field").
		WithLocation(5, "StudentID").
		WithCellReference("B5")

	errStr := err.Error()

	if errStr == "" {
		t.Error("expected non-empty error string")
	}

	// Should contain error type
	if !contains(errStr, string(ErrorTypeMissingRequired)) {
		t.Errorf("expected error string to contain error type")
	}
}

// helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
