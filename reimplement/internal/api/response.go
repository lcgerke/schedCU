package api

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/validation"
)

// ApiResponse is a generic response type that combines data, validation results, errors, and metadata.
// Type parameter T represents the actual response data type.
type ApiResponse[T any] struct {
	// Data holds the actual response payload
	Data T `json:"data"`

	// Validation contains validation errors, warnings, and infos from the operation
	Validation *validation.ValidationResult `json:"validation,omitempty"`

	// Error contains error information if the operation failed
	Error *ErrorDetail `json:"error,omitempty"`

	// Meta contains response metadata like timestamp, request ID, and version
	Meta *ResponseMeta `json:"meta"`
}

// ErrorDetail contains detailed information about an error that occurred during an operation.
type ErrorDetail struct {
	// Code is a machine-readable error code for programmatic handling
	Code string `json:"code"`

	// Message is a human-readable description of the error
	Message string `json:"message"`

	// Details contains contextual information about the error (optional)
	Details map[string]interface{} `json:"details,omitempty"`
}

// ResponseMeta contains metadata about the API response.
type ResponseMeta struct {
	// Timestamp is when the response was generated
	Timestamp time.Time `json:"timestamp"`

	// RequestID is a unique identifier for tracing this request
	RequestID string `json:"request_id"`

	// Version is the API version that generated this response
	Version string `json:"version"`

	// ServerTime is the server's current unix timestamp
	ServerTime int64 `json:"server_time"`
}

// NewApiResponse creates a new successful API response with the provided data.
// It initializes validation, metadata, and sets success status.
func NewApiResponse[T any](data T) *ApiResponse[T] {
	now := time.Now()
	return &ApiResponse[T]{
		Data:       data,
		Validation: validation.NewValidationResult(),
		Error:      nil,
		Meta: &ResponseMeta{
			Timestamp: now,
			RequestID: uuid.New().String(),
			Version:   "1.0",
			ServerTime: now.Unix(),
		},
	}
}

// WithValidation adds validation results to the response and returns the response for chaining.
func (ar *ApiResponse[T]) WithValidation(vr *validation.ValidationResult) *ApiResponse[T] {
	ar.Validation = vr
	return ar
}

// WithError sets the error on the response and returns the response for chaining.
func (ar *ApiResponse[T]) WithError(code, message string) *ApiResponse[T] {
	ar.Error = &ErrorDetail{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
	return ar
}

// WithErrorDetails adds details to the error and returns the response for chaining.
func (ar *ApiResponse[T]) WithErrorDetails(details map[string]interface{}) *ApiResponse[T] {
	if ar.Error == nil {
		ar.Error = &ErrorDetail{
			Details: make(map[string]interface{}),
		}
	}
	ar.Error.Details = details
	return ar
}

// IsSuccess returns true if the response represents a successful operation.
// A response is successful if there is no error and validation has no errors.
func (ar *ApiResponse[T]) IsSuccess() bool {
	if ar.Error != nil {
		return false
	}
	if ar.Validation != nil && ar.Validation.HasErrors() {
		return false
	}
	return true
}

// MarshalJSON implements custom JSON marshaling for ApiResponse.
// This ensures that the response is properly serialized with all fields.
func (ar *ApiResponse[T]) MarshalJSON() ([]byte, error) {
	type responseAlias ApiResponse[T]
	return json.Marshal(&struct {
		*responseAlias
	}{
		responseAlias: (*responseAlias)(ar),
	})
}
