// Package api provides HTTP API utilities and helpers for status code mapping.
package api

import (
	"net/http"

	"github.com/schedcu/reimplement/internal/validation"
)

// StatusMapper maps validation results and error codes to HTTP status codes.
type StatusMapper struct{}

// NewStatusMapper creates a new StatusMapper instance.
func NewStatusMapper() *StatusMapper {
	return &StatusMapper{}
}

// MapValidationToStatus maps a ValidationResult to an HTTP status code.
// Precedence rules:
// 1. If errors exist -> 400 Bad Request
// 2. If no errors but warnings exist -> 200 OK (with warning indicator)
// 3. If no errors/warnings but infos exist -> 200 OK (with info indicator)
// 4. If empty -> 200 OK
//
// The status code alone doesn't distinguish between warning/info states;
// the caller should check the ValidationResult fields to determine the message type.
func (sm *StatusMapper) MapValidationToStatus(vr *validation.ValidationResult) int {
	if vr == nil {
		return http.StatusOK
	}

	// Errors take highest precedence
	if vr.HasErrors() {
		return http.StatusBadRequest
	}

	// If no errors, warnings and infos return 200 OK
	// The response body will indicate whether warnings or infos were present
	return http.StatusOK
}

// ErrorCodeToHTTPStatus maps validation error codes to HTTP status codes.
// This function handles common error codes and provides semantic HTTP status mappings.
func (sm *StatusMapper) ErrorCodeToHTTPStatus(code validation.MessageCode) int {
	switch code {
	// 400 Bad Request - Client error in request data
	case validation.INVALID_FILE_FORMAT:
		return http.StatusBadRequest
	case validation.MISSING_REQUIRED_FIELD:
		return http.StatusBadRequest
	case validation.DUPLICATE_ENTRY:
		return http.StatusBadRequest
	case validation.PARSE_ERROR:
		return http.StatusBadRequest

	// 401 Unauthorized - Authentication required
	// Note: Not currently in the validation package, but included for future use
	// case validation.UNAUTHORIZED:
	//	return http.StatusUnauthorized

	// 403 Forbidden - Authenticated but not authorized
	// Note: Not currently in the validation package, but included for future use
	// case validation.FORBIDDEN:
	//	return http.StatusForbidden

	// 404 Not Found - Resource not found
	// Note: Not currently in the validation package, but included for future use
	// case validation.NOT_FOUND:
	//	return http.StatusNotFound

	// 500 Internal Server Error - Server-side issues
	case validation.DATABASE_ERROR:
		return http.StatusInternalServerError
	case validation.EXTERNAL_SERVICE_ERROR:
		return http.StatusInternalServerError
	case validation.UNKNOWN_ERROR:
		return http.StatusInternalServerError

	// Default to 500 for unknown error codes
	default:
		return http.StatusInternalServerError
	}
}

// SeverityToDescription provides a human-readable description of severity levels.
// This is useful for logging and API response documentation.
func (sm *StatusMapper) SeverityToDescription(severity validation.Severity) string {
	switch severity {
	case validation.ERROR:
		return "Error - validation failed, action cannot proceed"
	case validation.WARNING:
		return "Warning - validation passed with caveats, action can proceed"
	case validation.INFO:
		return "Info - informational message about validation"
	default:
		return "Unknown severity"
	}
}

// MessageCodeToDescription provides a human-readable description of message codes.
// This is useful for API response documentation and client error handling.
func (sm *StatusMapper) MessageCodeToDescription(code validation.MessageCode) string {
	switch code {
	case validation.INVALID_FILE_FORMAT:
		return "The uploaded file format is invalid or not supported"
	case validation.MISSING_REQUIRED_FIELD:
		return "A required field is missing from the request"
	case validation.DUPLICATE_ENTRY:
		return "A duplicate entry was detected"
	case validation.PARSE_ERROR:
		return "Failed to parse the input data"
	case validation.DATABASE_ERROR:
		return "An error occurred while accessing the database"
	case validation.EXTERNAL_SERVICE_ERROR:
		return "An external service returned an error"
	case validation.UNKNOWN_ERROR:
		return "An unknown error occurred"
	default:
		return "An error occurred"
	}
}

// IsClientError returns true if the given HTTP status code represents a client error (4xx).
func (sm *StatusMapper) IsClientError(statusCode int) bool {
	return statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError
}

// IsServerError returns true if the given HTTP status code represents a server error (5xx).
func (sm *StatusMapper) IsServerError(statusCode int) bool {
	return statusCode >= http.StatusInternalServerError && statusCode < 600
}

// IsSuccess returns true if the given HTTP status code represents a successful response (2xx).
func (sm *StatusMapper) IsSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
