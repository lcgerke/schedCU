package api

import (
	"time"

	"github.com/schedcu/v2/internal/validation"
)

// APIResponse is the standard response format for all endpoints.
type APIResponse struct {
	Data             interface{}         `json:"data,omitempty"`
	ValidationResult *validation.Result  `json:"validation,omitempty"`
	Error            *ErrorResponse      `json:"error,omitempty"`
	Meta             ResponseMeta        `json:"meta"`
}

// ErrorResponse contains error details.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ResponseMeta contains response metadata.
type ResponseMeta struct {
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
	Version   string    `json:"version,omitempty"`
}

// SuccessResponse returns a successful APIResponse.
func SuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Data:             data,
		ValidationResult: validation.NewResult(), // Empty result means success
		Meta: ResponseMeta{
			Timestamp: time.Now().UTC(),
			Version:   "1.0",
		},
	}
}

// ErrorResponseWithCode returns an error APIResponse.
func ErrorResponseWithCode(code, message string) *APIResponse {
	return &APIResponse{
		Error: &ErrorResponse{
			Code:    code,
			Message: message,
		},
		Meta: ResponseMeta{
			Timestamp: time.Now().UTC(),
			Version:   "1.0",
		},
	}
}

// ValidationErrorResponse returns a validation error APIResponse.
func ValidationErrorResponse(code, message string) *APIResponse {
	result := validation.NewResult()
	result.AddError(code, message)
	return &APIResponse{
		ValidationResult: result,
		Meta: ResponseMeta{
			Timestamp: time.Now().UTC(),
			Version:   "1.0",
		},
	}
}

// ResponseWithValidation returns an APIResponse with custom validation result.
func ResponseWithValidation(data interface{}, validationResult *validation.Result) *APIResponse {
	return &APIResponse{
		Data:             data,
		ValidationResult: validationResult,
		Meta: ResponseMeta{
			Timestamp: time.Now().UTC(),
			Version:   "1.0",
		},
	}
}
