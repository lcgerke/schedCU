package amion

import (
	"testing"
)

// Test 1: Create new error collector
func TestNewAmionErrorCollector(t *testing.T) {
	collector := NewAmionErrorCollector()
	if collector == nil {
		t.Fatal("NewAmionErrorCollector returned nil")
	}
	if collector.errors == nil {
		t.Fatal("collector.errors is nil")
	}
	if len(collector.errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(collector.errors))
	}
}

// Test 2: Add single MissingCell error
func TestAddError_MissingCell(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Expected shift date cell")

	if len(collector.errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(collector.errors))
	}

	err := collector.errors[0]
	if err.ErrorType != MissingCell {
		t.Errorf("expected error type MissingCell, got %v", err.ErrorType)
	}
	if err.Row != 5 {
		t.Errorf("expected row 5, got %d", err.Row)
	}
	if err.Col != 3 {
		t.Errorf("expected col 3, got %d", err.Col)
	}
	if err.Details != "Expected shift date cell" {
		t.Errorf("expected details 'Expected shift date cell', got '%s'", err.Details)
	}
}

// Test 3: Add multiple errors
func TestAddError_Multiple(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Missing date")
	collector.AddError(InvalidValue, 5, 4, "Invalid time format")
	collector.AddError(MissingCell, 6, 3, "Missing date")

	if len(collector.errors) != 3 {
		t.Errorf("expected 3 errors, got %d", len(collector.errors))
	}
}

// Test 4: Add InvalidValue error
func TestAddError_InvalidValue(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(InvalidValue, 10, 5, "Expected HH:MM format, got 'invalid'")

	err := collector.errors[0]
	if err.ErrorType != InvalidValue {
		t.Errorf("expected InvalidValue, got %v", err.ErrorType)
	}
	if err.Row != 10 {
		t.Errorf("expected row 10, got %d", err.Row)
	}
}

// Test 5: Add MissingRow error
func TestAddError_MissingRow(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingRow, 15, 0, "Expected shift data row not found")

	err := collector.errors[0]
	if err.ErrorType != MissingRow {
		t.Errorf("expected MissingRow, got %v", err.ErrorType)
	}
	if err.Row != 15 {
		t.Errorf("expected row 15, got %d", err.Row)
	}
}

// Test 6: Add InvalidHTML error
func TestAddError_InvalidHTML(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(InvalidHTML, 0, 0, "Table element not found")

	err := collector.errors[0]
	if err.ErrorType != InvalidHTML {
		t.Errorf("expected InvalidHTML, got %v", err.ErrorType)
	}
}

// Test 7: Add EmptyTable error
func TestAddError_EmptyTable(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(EmptyTable, 0, 0, "No data rows in table")

	err := collector.errors[0]
	if err.ErrorType != EmptyTable {
		t.Errorf("expected EmptyTable, got %v", err.ErrorType)
	}
}

// Test 8: Add EncodingError
func TestAddError_EncodingError(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(EncodingError, 0, 0, "Invalid UTF-8 sequence at offset 42")

	err := collector.errors[0]
	if err.ErrorType != EncodingError {
		t.Errorf("expected EncodingError, got %v", err.ErrorType)
	}
}

// Test 9: GroupErrorsByType with single type
func TestGroupErrorsByType_SingleType(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Missing date")
	collector.AddError(MissingCell, 6, 3, "Missing date")
	collector.AddError(MissingCell, 7, 3, "Missing date")

	groups := collector.GroupErrorsByType()
	if len(groups) != 1 {
		t.Errorf("expected 1 error group, got %d", len(groups))
	}

	missingCellErrors, ok := groups[MissingCell]
	if !ok {
		t.Fatal("MissingCell group not found")
	}
	if len(missingCellErrors) != 3 {
		t.Errorf("expected 3 MissingCell errors, got %d", len(missingCellErrors))
	}
}

// Test 10: GroupErrorsByType with multiple types
func TestGroupErrorsByType_MultipleTypes(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Missing date")
	collector.AddError(InvalidValue, 5, 4, "Invalid time")
	collector.AddError(MissingCell, 6, 3, "Missing date")
	collector.AddError(InvalidHTML, 0, 0, "Table not found")

	groups := collector.GroupErrorsByType()
	if len(groups) != 3 {
		t.Errorf("expected 3 error groups, got %d", len(groups))
	}

	if len(groups[MissingCell]) != 2 {
		t.Errorf("expected 2 MissingCell errors, got %d", len(groups[MissingCell]))
	}
	if len(groups[InvalidValue]) != 1 {
		t.Errorf("expected 1 InvalidValue error, got %d", len(groups[InvalidValue]))
	}
	if len(groups[InvalidHTML]) != 1 {
		t.Errorf("expected 1 InvalidHTML error, got %d", len(groups[InvalidHTML]))
	}
}

// Test 11: ErrorCount
func TestErrorCount(t *testing.T) {
	collector := NewAmionErrorCollector()
	if collector.ErrorCount() != 0 {
		t.Errorf("expected 0 errors, got %d", collector.ErrorCount())
	}

	collector.AddError(MissingCell, 5, 3, "Missing date")
	if collector.ErrorCount() != 1 {
		t.Errorf("expected 1 error, got %d", collector.ErrorCount())
	}

	collector.AddError(InvalidValue, 5, 4, "Invalid time")
	collector.AddError(MissingCell, 6, 3, "Missing date")
	if collector.ErrorCount() != 3 {
		t.Errorf("expected 3 errors, got %d", collector.ErrorCount())
	}
}

// Test 12: HasErrors
func TestHasErrors(t *testing.T) {
	collector := NewAmionErrorCollector()
	if collector.HasErrors() {
		t.Fatal("expected no errors")
	}

	collector.AddError(MissingCell, 5, 3, "Missing date")
	if !collector.HasErrors() {
		t.Fatal("expected errors")
	}
}

// Test 13: Convert to ValidationResult with errors
func TestToValidationResult_WithErrors(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Expected shift date")
	collector.AddError(InvalidValue, 5, 4, "Invalid time format")
	collector.AddError(MissingCell, 6, 3, "Expected shift date")

	result := collector.ToValidationResult()

	if result == nil {
		t.Fatal("ToValidationResult returned nil")
	}

	if result.IsValid() {
		t.Fatal("expected result to be invalid (has errors)")
	}

	if result.ErrorCount() != 3 {
		t.Errorf("expected 3 errors in result, got %d", result.ErrorCount())
	}

	// Verify error messages contain row/col information
	if len(result.Errors) > 0 {
		firstMsg := result.Errors[0].Message
		if len(firstMsg) == 0 {
			t.Fatal("expected error message")
		}
		// Message should contain row reference
		if !contains(firstMsg, "5") && !contains(firstMsg, "row") && !contains(firstMsg, "R") {
			t.Errorf("expected row reference in message: %s", firstMsg)
		}
	}
}

// Test 14: Convert to ValidationResult without errors
func TestToValidationResult_NoErrors(t *testing.T) {
	collector := NewAmionErrorCollector()

	result := collector.ToValidationResult()

	if result == nil {
		t.Fatal("ToValidationResult returned nil")
	}

	if !result.IsValid() {
		t.Fatal("expected result to be valid (no errors)")
	}

	if result.ErrorCount() != 0 {
		t.Errorf("expected 0 errors, got %d", result.ErrorCount())
	}
}

// Test 15: Error message quality and format
func TestErrorMessageFormat(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Expected shift date cell")
	collector.AddError(InvalidValue, 10, 4, "Invalid time format 'abc'")

	result := collector.ToValidationResult()

	// Check first error message
	if len(result.Errors) > 0 {
		msg := result.Errors[0]
		// Field should have row/col reference
		if len(msg.Field) == 0 {
			t.Fatal("expected field to contain cell reference")
		}
		if len(msg.Message) == 0 {
			t.Fatal("expected non-empty message")
		}
		t.Logf("Error message format: Field=%q, Message=%q", msg.Field, msg.Message)
	}
}

// Test 16: Error collector clear/reset
func TestClear(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Missing date")
	collector.AddError(InvalidValue, 5, 4, "Invalid time")

	if collector.ErrorCount() != 2 {
		t.Errorf("expected 2 errors before clear, got %d", collector.ErrorCount())
	}

	collector.Clear()

	if collector.ErrorCount() != 0 {
		t.Errorf("expected 0 errors after clear, got %d", collector.ErrorCount())
	}

	if collector.HasErrors() {
		t.Fatal("expected no errors after clear")
	}
}

// Test 17: GetErrors returns a copy
func TestGetErrors(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Missing date")
	collector.AddError(InvalidValue, 5, 4, "Invalid time")

	errors := collector.GetErrors()
	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}

	// Verify we get the right error types
	types := make(map[ErrorType]int)
	for _, err := range errors {
		types[err.ErrorType]++
	}

	if types[MissingCell] != 1 {
		t.Errorf("expected 1 MissingCell error, got %d", types[MissingCell])
	}
	if types[InvalidValue] != 1 {
		t.Errorf("expected 1 InvalidValue error, got %d", types[InvalidValue])
	}
}

// Test 18: Multiple errors same cell
func TestMultipleErrorsSameCell(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Missing date")
	collector.AddError(InvalidValue, 5, 3, "Invalid format")

	if collector.ErrorCount() != 2 {
		t.Errorf("expected 2 errors, got %d", collector.ErrorCount())
	}

	result := collector.ToValidationResult()
	if result.ErrorCount() != 2 {
		t.Errorf("expected 2 errors in result, got %d", result.ErrorCount())
	}
}

// Test 19: Complex error scenario
func TestComplexErrorScenario(t *testing.T) {
	collector := NewAmionErrorCollector()

	// Simulate parsing multiple rows
	collector.AddError(MissingRow, 2, 0, "Expected header row not found")
	collector.AddError(InvalidHTML, 0, 0, "Table structure does not match expected selectors")

	// Errors in first data row
	collector.AddError(MissingCell, 5, 1, "Missing shift date")
	collector.AddError(InvalidValue, 5, 2, "Invalid shift time: '25:00'")

	// Errors in second data row
	collector.AddError(MissingCell, 6, 1, "Missing shift date")
	collector.AddError(InvalidValue, 6, 3, "Invalid staff count: 'abc'")

	// More structural errors
	collector.AddError(EncodingError, 0, 0, "Invalid UTF-8 at position 1024")

	result := collector.ToValidationResult()
	if result.ErrorCount() != 7 {
		t.Errorf("expected 7 errors, got %d", result.ErrorCount())
	}

	if result.IsValid() {
		t.Fatal("expected result to be invalid")
	}

	// Verify grouping works
	groups := collector.GroupErrorsByType()
	if len(groups) < 4 {
		t.Errorf("expected at least 4 error type groups, got %d", len(groups))
	}
}

// Test 20: ValidationResult context information
func TestValidationResultContext(t *testing.T) {
	collector := NewAmionErrorCollector()
	collector.AddError(MissingCell, 5, 3, "Missing date")
	collector.AddError(InvalidValue, 5, 4, "Invalid time")

	result := collector.ToValidationResult()

	// Context should have error type information
	if val, ok := result.GetContext("total_errors"); !ok {
		t.Error("expected 'total_errors' in context")
	} else if val != 2 {
		t.Errorf("expected 2 total errors in context, got %v", val)
	}

	// Context should have error type breakdown
	if val, ok := result.GetContext("errors_by_type"); ok {
		typeMap, isMap := val.(map[ErrorType]int)
		if !isMap {
			t.Errorf("expected map[ErrorType]int for errors_by_type, got %T", val)
		} else if len(typeMap) != 2 {
			t.Errorf("expected 2 error types in context, got %d", len(typeMap))
		}
	}
}

// Helper function for substring checking
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
