package api

import "github.com/schedcu/reimplement/internal/validation"

// Example usage patterns for the ApiResponse type

// ExampleSuccessResponse demonstrates creating a successful response with data
func ExampleSuccessResponse() {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	user := User{ID: 1, Name: "John Doe"}
	response := NewApiResponse(user)

	// response.IsSuccess() returns true
	// response.Data contains the User
	// response.Validation is initialized but empty
	// response.Meta contains timestamp, requestID, version, serverTime
	_ = response
}

// ExampleResponseWithValidation demonstrates adding validation warnings to a response
func ExampleResponseWithValidation() {
	type Product struct {
		Name string `json:"name"`
		SKU  string `json:"sku"`
	}

	product := Product{Name: "Widget", SKU: "WGT-001"}
	response := NewApiResponse(product)

	// Add validation warnings
	vr := validation.NewValidationResult()
	vr.AddWarning("sku", "SKU format is deprecated, please use new format")
	vr.AddInfo("inventory", "Low stock warning from warehouse system")

	response.WithValidation(vr)

	// response.IsSuccess() still returns true because there are no errors
	// response.Validation contains the warnings and infos
	_ = response
}

// ExampleResponseWithError demonstrates creating an error response
func ExampleResponseWithError() {
	type EmptyData struct{}

	response := NewApiResponse(EmptyData{})

	response.WithError("RESOURCE_NOT_FOUND", "The requested resource does not exist")

	// response.IsSuccess() returns false
	// response.Error.Code is "RESOURCE_NOT_FOUND"
	// response.Error.Message is "The requested resource does not exist"
	_ = response
}

// ExampleResponseWithErrorDetails demonstrates error with contextual details
func ExampleResponseWithErrorDetails() {
	type EmptyData struct{}

	response := NewApiResponse(EmptyData{})

	details := map[string]interface{}{
		"field":    "email",
		"received": "invalid-email",
		"expected": "valid email format",
		"pattern":  "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
	}

	response.
		WithError("VALIDATION_FAILED", "Email validation failed").
		WithErrorDetails(details)

	// response.Error.Details contains all the contextual information
	_ = response
}

// ExampleMethodChaining demonstrates fluent API for method chaining
func ExampleMethodChaining() {
	type Item struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}

	item := Item{ID: 42, Title: "Example Item"}
	vr := validation.NewValidationResult()
	vr.AddWarning("title", "Title contains HTML characters, consider escaping")

	response := NewApiResponse(item).
		WithValidation(vr).
		WithError("PARTIAL_SUCCESS", "Operation completed with warnings")

	// All methods return the same response for chaining
	// response contains data, validation warnings, and error info
	_ = response
}

// ExampleSuccessWithMultipleTypes demonstrates generic type flexibility
func ExampleSuccessWithMultipleTypes() {
	// Works with structs
	type User struct {
		Name string `json:"name"`
	}
	response1 := NewApiResponse(User{Name: "Alice"})
	_ = response1

	// Works with slices
	response2 := NewApiResponse([]int{1, 2, 3})
	_ = response2

	// Works with maps
	response3 := NewApiResponse(map[string]string{"key": "value"})
	_ = response3

	// Works with primitives
	response4 := NewApiResponse(42)
	_ = response4

	// Works with pointers
	user := &User{Name: "Bob"}
	response5 := NewApiResponse(user)
	_ = response5

	// Works with nil
	response6 := NewApiResponse[interface{}](nil)
	_ = response6
}

// ExampleJSONSerialization demonstrates JSON output format
func ExampleJSONSerialization() {
	// A successful response JSON output:
	// {
	//   "data": {
	//     "id": 1,
	//     "name": "Example"
	//   },
	//   "validation": {
	//     "errors": [],
	//     "warnings": [],
	//     "infos": [],
	//     "context": {}
	//   },
	//   "meta": {
	//     "timestamp": "2024-11-15T17:30:00Z",
	//     "request_id": "550e8400-e29b-41d4-a716-446655440000",
	//     "version": "1.0",
	//     "server_time": 1731000600
	//   }
	// }

	// An error response JSON output:
	// {
	//   "data": null,
	//   "error": {
	//     "code": "INVALID_REQUEST",
	//     "message": "The request is malformed",
	//     "details": {
	//       "field": "email",
	//       "reason": "invalid format"
	//     }
	//   },
	//   "validation": {
	//     "errors": [],
	//     "warnings": [],
	//     "infos": [],
	//     "context": {}
	//   },
	//   "meta": {
	//     "timestamp": "2024-11-15T17:30:00Z",
	//     "request_id": "550e8400-e29b-41d4-a716-446655440000",
	//     "version": "1.0",
	//     "server_time": 1731000600
	//   }
	// }
}

// ExampleIsSuccessLogic demonstrates IsSuccess() behavior
func ExampleIsSuccessLogic() {
	type Data struct{}

	// Success: no error and no validation errors
	r1 := NewApiResponse(Data{})
	_ = r1.IsSuccess() // true

	// Failure: error is set
	r2 := NewApiResponse(Data{}).WithError("CODE", "message")
	_ = r2.IsSuccess() // false

	// Failure: validation has errors
	r3 := NewApiResponse(Data{})
	vr := validation.NewValidationResult()
	vr.AddError("field", "error message")
	r3.WithValidation(vr)
	_ = r3.IsSuccess() // false

	// Success: validation has only warnings
	r4 := NewApiResponse(Data{})
	vr2 := validation.NewValidationResult()
	vr2.AddWarning("field", "warning message")
	r4.WithValidation(vr2)
	_ = r4.IsSuccess() // true
}
