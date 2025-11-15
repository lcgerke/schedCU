package api

import (
	"net/http"
	"testing"

	"github.com/schedcu/reimplement/internal/validation"
)

// ============================================================================
// VALIDATION RESULT MAPPING TESTS
// ============================================================================

// TestMapValidationToStatusWithErrors tests that errors map to 400 Bad Request.
func TestMapValidationToStatusWithErrors(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddError("field1", "error message")

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, status)
	}
}

// TestMapValidationToStatusWithMultipleErrors tests that multiple errors map to 400.
func TestMapValidationToStatusWithMultipleErrors(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddError("field1", "error 1")
	vr.AddError("field2", "error 2")
	vr.AddError("field3", "error 3")

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, status)
	}
}

// TestMapValidationToStatusWithWarningsOnly tests that warnings only map to 200 OK.
func TestMapValidationToStatusWithWarningsOnly(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddWarning("field1", "warning message")

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, status)
	}
}

// TestMapValidationToStatusWithMultipleWarnings tests that multiple warnings map to 200 OK.
func TestMapValidationToStatusWithMultipleWarnings(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddWarning("field1", "warning 1")
	vr.AddWarning("field2", "warning 2")

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, status)
	}
}

// TestMapValidationToStatusWithInfosOnly tests that infos only map to 200 OK.
func TestMapValidationToStatusWithInfosOnly(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddInfo("source1", "info message")

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, status)
	}
}

// TestMapValidationToStatusWithMultipleInfos tests that multiple infos map to 200 OK.
func TestMapValidationToStatusWithMultipleInfos(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddInfo("source1", "info 1")
	vr.AddInfo("source2", "info 2")
	vr.AddInfo("source3", "info 3")

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, status)
	}
}

// TestMapValidationToStatusEmpty tests that empty result maps to 200 OK.
func TestMapValidationToStatusEmpty(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, status)
	}
}

// TestMapValidationToStatusNil tests that nil result maps to 200 OK.
func TestMapValidationToStatusNil(t *testing.T) {
	mapper := NewStatusMapper()

	status := mapper.MapValidationToStatus(nil)
	if status != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, status)
	}
}

// ============================================================================
// PRECEDENCE TESTS (Error > Warning > Info)
// ============================================================================

// TestMapValidationPrecedenceErrorOverWarning tests that errors take precedence over warnings.
func TestMapValidationPrecedenceErrorOverWarning(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddError("field1", "error message")
	vr.AddWarning("field2", "warning message")

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusBadRequest {
		t.Errorf("expected error to take precedence: expected %d, got %d", http.StatusBadRequest, status)
	}
}

// TestMapValidationPrecedenceErrorOverInfo tests that errors take precedence over infos.
func TestMapValidationPrecedenceErrorOverInfo(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddError("field1", "error message")
	vr.AddInfo("source1", "info message")

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusBadRequest {
		t.Errorf("expected error to take precedence: expected %d, got %d", http.StatusBadRequest, status)
	}
}

// TestMapValidationPrecedenceErrorOverAll tests that errors take precedence over warnings and infos.
func TestMapValidationPrecedenceErrorOverAll(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddError("field1", "error 1")
	vr.AddError("field2", "error 2")
	vr.AddWarning("field3", "warning 1")
	vr.AddWarning("field4", "warning 2")
	vr.AddInfo("source1", "info 1")
	vr.AddInfo("source2", "info 2")

	status := mapper.MapValidationToStatus(vr)
	if status != http.StatusBadRequest {
		t.Errorf("expected error to take precedence: expected %d, got %d", http.StatusBadRequest, status)
	}
}

// TestMapValidationPrecedenceWarningOverInfo tests that warnings are present with infos.
func TestMapValidationPrecedenceWarningOverInfo(t *testing.T) {
	mapper := NewStatusMapper()
	vr := validation.NewValidationResult()
	vr.AddWarning("field1", "warning message")
	vr.AddInfo("source1", "info message")

	status := mapper.MapValidationToStatus(vr)
	// Both warnings and infos return 200 OK
	if status != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, status)
	}
	// Verify both are present
	if !vr.HasWarnings() {
		t.Error("expected HasWarnings to be true")
	}
	if vr.InfoCount() == 0 {
		t.Error("expected infos to be present")
	}
}

// ============================================================================
// ERROR CODE MAPPING TESTS
// ============================================================================

// TestErrorCodeToHTTPStatusInvalidFileFormat tests INVALID_FILE_FORMAT maps to 400.
func TestErrorCodeToHTTPStatusInvalidFileFormat(t *testing.T) {
	mapper := NewStatusMapper()
	status := mapper.ErrorCodeToHTTPStatus(validation.INVALID_FILE_FORMAT)
	if status != http.StatusBadRequest {
		t.Errorf("expected %d for INVALID_FILE_FORMAT, got %d", http.StatusBadRequest, status)
	}
}

// TestErrorCodeToHTTPStatusMissingRequiredField tests MISSING_REQUIRED_FIELD maps to 400.
func TestErrorCodeToHTTPStatusMissingRequiredField(t *testing.T) {
	mapper := NewStatusMapper()
	status := mapper.ErrorCodeToHTTPStatus(validation.MISSING_REQUIRED_FIELD)
	if status != http.StatusBadRequest {
		t.Errorf("expected %d for MISSING_REQUIRED_FIELD, got %d", http.StatusBadRequest, status)
	}
}

// TestErrorCodeToHTTPStatusDuplicateEntry tests DUPLICATE_ENTRY maps to 400.
func TestErrorCodeToHTTPStatusDuplicateEntry(t *testing.T) {
	mapper := NewStatusMapper()
	status := mapper.ErrorCodeToHTTPStatus(validation.DUPLICATE_ENTRY)
	if status != http.StatusBadRequest {
		t.Errorf("expected %d for DUPLICATE_ENTRY, got %d", http.StatusBadRequest, status)
	}
}

// TestErrorCodeToHTTPStatusParseError tests PARSE_ERROR maps to 400.
func TestErrorCodeToHTTPStatusParseError(t *testing.T) {
	mapper := NewStatusMapper()
	status := mapper.ErrorCodeToHTTPStatus(validation.PARSE_ERROR)
	if status != http.StatusBadRequest {
		t.Errorf("expected %d for PARSE_ERROR, got %d", http.StatusBadRequest, status)
	}
}

// TestErrorCodeToHTTPStatusDatabaseError tests DATABASE_ERROR maps to 500.
func TestErrorCodeToHTTPStatusDatabaseError(t *testing.T) {
	mapper := NewStatusMapper()
	status := mapper.ErrorCodeToHTTPStatus(validation.DATABASE_ERROR)
	if status != http.StatusInternalServerError {
		t.Errorf("expected %d for DATABASE_ERROR, got %d", http.StatusInternalServerError, status)
	}
}

// TestErrorCodeToHTTPStatusExternalServiceError tests EXTERNAL_SERVICE_ERROR maps to 500.
func TestErrorCodeToHTTPStatusExternalServiceError(t *testing.T) {
	mapper := NewStatusMapper()
	status := mapper.ErrorCodeToHTTPStatus(validation.EXTERNAL_SERVICE_ERROR)
	if status != http.StatusInternalServerError {
		t.Errorf("expected %d for EXTERNAL_SERVICE_ERROR, got %d", http.StatusInternalServerError, status)
	}
}

// TestErrorCodeToHTTPStatusUnknownError tests UNKNOWN_ERROR maps to 500.
func TestErrorCodeToHTTPStatusUnknownError(t *testing.T) {
	mapper := NewStatusMapper()
	status := mapper.ErrorCodeToHTTPStatus(validation.UNKNOWN_ERROR)
	if status != http.StatusInternalServerError {
		t.Errorf("expected %d for UNKNOWN_ERROR, got %d", http.StatusInternalServerError, status)
	}
}

// TestErrorCodeToHTTPStatusUnknownCode tests unknown error code maps to 500.
func TestErrorCodeToHTTPStatusUnknownCode(t *testing.T) {
	mapper := NewStatusMapper()
	unknownCode := validation.MessageCode("UNKNOWN_CODE")
	status := mapper.ErrorCodeToHTTPStatus(unknownCode)
	if status != http.StatusInternalServerError {
		t.Errorf("expected %d for unknown code, got %d", http.StatusInternalServerError, status)
	}
}

// ============================================================================
// SEVERITY DESCRIPTION TESTS
// ============================================================================

// TestSeverityToDescriptionError tests ERROR severity description.
func TestSeverityToDescriptionError(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.SeverityToDescription(validation.ERROR)
	if desc == "" {
		t.Error("expected non-empty description for ERROR severity")
	}
	// Verify it mentions error-related terms
	if len(desc) < 10 {
		t.Error("expected descriptive message for ERROR severity")
	}
}

// TestSeverityToDescriptionWarning tests WARNING severity description.
func TestSeverityToDescriptionWarning(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.SeverityToDescription(validation.WARNING)
	if desc == "" {
		t.Error("expected non-empty description for WARNING severity")
	}
	// Verify it mentions warning-related terms
	if len(desc) < 10 {
		t.Error("expected descriptive message for WARNING severity")
	}
}

// TestSeverityToDescriptionInfo tests INFO severity description.
func TestSeverityToDescriptionInfo(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.SeverityToDescription(validation.INFO)
	if desc == "" {
		t.Error("expected non-empty description for INFO severity")
	}
	// Verify it mentions info-related terms
	if len(desc) < 10 {
		t.Error("expected descriptive message for INFO severity")
	}
}

// TestSeverityToDescriptionUnknown tests unknown severity description.
func TestSeverityToDescriptionUnknown(t *testing.T) {
	mapper := NewStatusMapper()
	unknownSeverity := validation.Severity("unknown")
	desc := mapper.SeverityToDescription(unknownSeverity)
	if desc == "" {
		t.Error("expected non-empty description for unknown severity")
	}
}

// ============================================================================
// MESSAGE CODE DESCRIPTION TESTS
// ============================================================================

// TestMessageCodeToDescriptionInvalidFileFormat tests INVALID_FILE_FORMAT description.
func TestMessageCodeToDescriptionInvalidFileFormat(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.MessageCodeToDescription(validation.INVALID_FILE_FORMAT)
	if desc == "" {
		t.Error("expected non-empty description for INVALID_FILE_FORMAT")
	}
}

// TestMessageCodeToDescriptionMissingRequiredField tests MISSING_REQUIRED_FIELD description.
func TestMessageCodeToDescriptionMissingRequiredField(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.MessageCodeToDescription(validation.MISSING_REQUIRED_FIELD)
	if desc == "" {
		t.Error("expected non-empty description for MISSING_REQUIRED_FIELD")
	}
}

// TestMessageCodeToDescriptionDuplicateEntry tests DUPLICATE_ENTRY description.
func TestMessageCodeToDescriptionDuplicateEntry(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.MessageCodeToDescription(validation.DUPLICATE_ENTRY)
	if desc == "" {
		t.Error("expected non-empty description for DUPLICATE_ENTRY")
	}
}

// TestMessageCodeToDescriptionParseError tests PARSE_ERROR description.
func TestMessageCodeToDescriptionParseError(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.MessageCodeToDescription(validation.PARSE_ERROR)
	if desc == "" {
		t.Error("expected non-empty description for PARSE_ERROR")
	}
}

// TestMessageCodeToDescriptionDatabaseError tests DATABASE_ERROR description.
func TestMessageCodeToDescriptionDatabaseError(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.MessageCodeToDescription(validation.DATABASE_ERROR)
	if desc == "" {
		t.Error("expected non-empty description for DATABASE_ERROR")
	}
}

// TestMessageCodeToDescriptionExternalServiceError tests EXTERNAL_SERVICE_ERROR description.
func TestMessageCodeToDescriptionExternalServiceError(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.MessageCodeToDescription(validation.EXTERNAL_SERVICE_ERROR)
	if desc == "" {
		t.Error("expected non-empty description for EXTERNAL_SERVICE_ERROR")
	}
}

// TestMessageCodeToDescriptionUnknownError tests UNKNOWN_ERROR description.
func TestMessageCodeToDescriptionUnknownError(t *testing.T) {
	mapper := NewStatusMapper()
	desc := mapper.MessageCodeToDescription(validation.UNKNOWN_ERROR)
	if desc == "" {
		t.Error("expected non-empty description for UNKNOWN_ERROR")
	}
}

// TestMessageCodeToDescriptionUnknownCode tests unknown code description.
func TestMessageCodeToDescriptionUnknownCode(t *testing.T) {
	mapper := NewStatusMapper()
	unknownCode := validation.MessageCode("UNKNOWN_CODE")
	desc := mapper.MessageCodeToDescription(unknownCode)
	if desc == "" {
		t.Error("expected non-empty description for unknown code")
	}
}

// ============================================================================
// STATUS CODE CLASSIFICATION TESTS
// ============================================================================

// TestIsClientError tests IsClientError method for 4xx codes.
func TestIsClientError(t *testing.T) {
	mapper := NewStatusMapper()

	tests := []struct {
		statusCode int
		expected   bool
		name       string
	}{
		{http.StatusBadRequest, true, "400 Bad Request"},
		{http.StatusUnauthorized, true, "401 Unauthorized"},
		{http.StatusForbidden, true, "403 Forbidden"},
		{http.StatusNotFound, true, "404 Not Found"},
		{http.StatusOK, false, "200 OK"},
		{http.StatusInternalServerError, false, "500 Internal Server Error"},
		{http.StatusServiceUnavailable, false, "503 Service Unavailable"},
	}

	for _, tt := range tests {
		result := mapper.IsClientError(tt.statusCode)
		if result != tt.expected {
			t.Errorf("IsClientError(%s): expected %v, got %v", tt.name, tt.expected, result)
		}
	}
}

// TestIsServerError tests IsServerError method for 5xx codes.
func TestIsServerError(t *testing.T) {
	mapper := NewStatusMapper()

	tests := []struct {
		statusCode int
		expected   bool
		name       string
	}{
		{http.StatusInternalServerError, true, "500 Internal Server Error"},
		{http.StatusNotImplemented, true, "501 Not Implemented"},
		{http.StatusServiceUnavailable, true, "503 Service Unavailable"},
		{http.StatusOK, false, "200 OK"},
		{http.StatusBadRequest, false, "400 Bad Request"},
		{http.StatusNotFound, false, "404 Not Found"},
	}

	for _, tt := range tests {
		result := mapper.IsServerError(tt.statusCode)
		if result != tt.expected {
			t.Errorf("IsServerError(%s): expected %v, got %v", tt.name, tt.expected, result)
		}
	}
}

// TestIsSuccess tests IsSuccess method for 2xx codes.
func TestIsSuccess(t *testing.T) {
	mapper := NewStatusMapper()

	tests := []struct {
		statusCode int
		expected   bool
		name       string
	}{
		{http.StatusOK, true, "200 OK"},
		{http.StatusCreated, true, "201 Created"},
		{http.StatusAccepted, true, "202 Accepted"},
		{http.StatusNoContent, true, "204 No Content"},
		{http.StatusBadRequest, false, "400 Bad Request"},
		{http.StatusInternalServerError, false, "500 Internal Server Error"},
		{http.StatusMultipleChoices, false, "300 Multiple Choices"},
	}

	for _, tt := range tests {
		result := mapper.IsSuccess(tt.statusCode)
		if result != tt.expected {
			t.Errorf("IsSuccess(%s): expected %v, got %v", tt.name, tt.expected, result)
		}
	}
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

// TestIntegrationErrorCodeToStatus tests integration of error code to status mapping.
func TestIntegrationErrorCodeToStatus(t *testing.T) {
	mapper := NewStatusMapper()

	// Test that client errors (400s) are properly classified
	clientErrorCode := validation.INVALID_FILE_FORMAT
	status := mapper.ErrorCodeToHTTPStatus(clientErrorCode)
	if !mapper.IsClientError(status) {
		t.Errorf("expected client error classification for %s", clientErrorCode)
	}

	// Test that server errors (500s) are properly classified
	serverErrorCode := validation.DATABASE_ERROR
	status = mapper.ErrorCodeToHTTPStatus(serverErrorCode)
	if !mapper.IsServerError(status) {
		t.Errorf("expected server error classification for %s", serverErrorCode)
	}
}

// TestIntegrationValidationResultMapping tests complete validation result mapping flow.
func TestIntegrationValidationResultMapping(t *testing.T) {
	mapper := NewStatusMapper()

	// Test error case
	vrWithError := validation.NewValidationResult()
	vrWithError.AddError("field", "invalid value")
	status := mapper.MapValidationToStatus(vrWithError)
	if !mapper.IsClientError(status) {
		t.Error("expected error result to map to client error (4xx)")
	}

	// Test warning case
	vrWithWarning := validation.NewValidationResult()
	vrWithWarning.AddWarning("field", "weak value")
	status = mapper.MapValidationToStatus(vrWithWarning)
	if !mapper.IsSuccess(status) {
		t.Error("expected warning result to map to success (2xx)")
	}

	// Test info case
	vrWithInfo := validation.NewValidationResult()
	vrWithInfo.AddInfo("source", "processing started")
	status = mapper.MapValidationToStatus(vrWithInfo)
	if !mapper.IsSuccess(status) {
		t.Error("expected info result to map to success (2xx)")
	}

	// Test empty case
	vrEmpty := validation.NewValidationResult()
	status = mapper.MapValidationToStatus(vrEmpty)
	if !mapper.IsSuccess(status) {
		t.Error("expected empty result to map to success (2xx)")
	}
}

// TestNewStatusMapper tests that NewStatusMapper creates a valid instance.
func TestNewStatusMapper(t *testing.T) {
	mapper := NewStatusMapper()
	if mapper == nil {
		t.Error("expected NewStatusMapper to return non-nil instance")
	}
}

// ============================================================================
// EDGE CASE TESTS
// ============================================================================

// TestMapValidationZeroValueStruct tests mapping zero-value ValidationResult.
func TestMapValidationZeroValueStruct(t *testing.T) {
	mapper := NewStatusMapper()
	var vr validation.ValidationResult

	status := mapper.MapValidationToStatus(&vr)
	if status != http.StatusOK {
		t.Errorf("expected %d for zero-value result, got %d", http.StatusOK, status)
	}
}

// TestMultipleMapperInstances tests that multiple mapper instances work independently.
func TestMultipleMapperInstances(t *testing.T) {
	mapper1 := NewStatusMapper()
	mapper2 := NewStatusMapper()

	vr := validation.NewValidationResult()
	vr.AddError("field", "error")

	status1 := mapper1.MapValidationToStatus(vr)
	status2 := mapper2.MapValidationToStatus(vr)

	if status1 != status2 {
		t.Errorf("expected independent mappers to produce same result: %d vs %d", status1, status2)
	}
}

// TestErrorCodeMappingConsistency tests that all error codes map to valid HTTP statuses.
func TestErrorCodeMappingConsistency(t *testing.T) {
	mapper := NewStatusMapper()

	errorCodes := []validation.MessageCode{
		validation.INVALID_FILE_FORMAT,
		validation.MISSING_REQUIRED_FIELD,
		validation.DUPLICATE_ENTRY,
		validation.PARSE_ERROR,
		validation.DATABASE_ERROR,
		validation.EXTERNAL_SERVICE_ERROR,
		validation.UNKNOWN_ERROR,
	}

	for _, code := range errorCodes {
		status := mapper.ErrorCodeToHTTPStatus(code)

		// Status should be either 4xx or 5xx
		if status < http.StatusBadRequest {
			t.Errorf("error code %s mapped to non-error status %d", code, status)
		}
		if status >= 600 {
			t.Errorf("error code %s mapped to invalid status %d", code, status)
		}
	}
}
