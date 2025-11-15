package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/schedcu/reimplement/internal/validation"
)

// Test 1: JSON marshaling of success response with string data
func TestMarshalSuccessResponseWithString(t *testing.T) {
	resp := NewApiResponse("Hello, World!")

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, "Hello, World!", unmarshaled["data"])
	assert.NotNil(t, unmarshaled["meta"])
	assert.NotNil(t, unmarshaled["validation"])
}

// Test 2: JSON marshaling of success response with integer data
func TestMarshalSuccessResponseWithInt(t *testing.T) {
	resp := NewApiResponse(42)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, float64(42), unmarshaled["data"])
	assert.NotNil(t, unmarshaled["meta"])
}

// Test 3: JSON marshaling of success response with float data
func TestMarshalSuccessResponseWithFloat(t *testing.T) {
	resp := NewApiResponse(3.14159)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.InDelta(t, 3.14159, unmarshaled["data"], 0.00001)
}

// Test 4: JSON marshaling of success response with boolean data
func TestMarshalSuccessResponseWithBool(t *testing.T) {
	resp := NewApiResponse(true)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, true, unmarshaled["data"])
}

// Test 5: JSON marshaling of success response with map data
func TestMarshalSuccessResponseWithMap(t *testing.T) {
	data := map[string]interface{}{
		"id":    "123",
		"name":  "Test User",
		"age":   30,
		"email": "test@example.com",
	}
	resp := NewApiResponse(data)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	dataMap := unmarshaled["data"].(map[string]interface{})
	assert.Equal(t, "123", dataMap["id"])
	assert.Equal(t, "Test User", dataMap["name"])
	assert.Equal(t, float64(30), dataMap["age"])
	assert.Equal(t, "test@example.com", dataMap["email"])
}

// Test 6: JSON marshaling of success response with array data
func TestMarshalSuccessResponseWithSlice(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	resp := NewApiResponse(data)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	dataArray := unmarshaled["data"].([]interface{})
	assert.Len(t, dataArray, 3)
	assert.Equal(t, "item1", dataArray[0])
	assert.Equal(t, "item2", dataArray[1])
	assert.Equal(t, "item3", dataArray[2])
}

// Test 7: JSON marshaling of response with nested objects
func TestMarshalSuccessResponseWithNestedObject(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   "123",
			"name": "John Doe",
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "Springfield",
			},
		},
	}
	resp := NewApiResponse(data)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	dataMap := unmarshaled["data"].(map[string]interface{})
	userMap := dataMap["user"].(map[string]interface{})
	addressMap := userMap["address"].(map[string]interface{})
	assert.Equal(t, "Springfield", addressMap["city"])
}

// Test 8: Error response with WithError method
func TestErrorResponseWithMethod(t *testing.T) {
	resp := NewApiResponse("").
		WithError("INVALID_REQUEST", "The request is invalid")

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.NotNil(t, unmarshaled["error"])
	errorObj := unmarshaled["error"].(map[string]interface{})
	assert.Equal(t, "INVALID_REQUEST", errorObj["code"])
	assert.Equal(t, "The request is invalid", errorObj["message"])
}

// Test 9: Error response with error details
func TestErrorResponseWithDetails(t *testing.T) {
	details := map[string]interface{}{
		"field":    "email",
		"expected": "valid email format",
		"actual":   "not-an-email",
	}
	resp := NewApiResponse("").
		WithError("VALIDATION_ERROR", "Field validation failed").
		WithErrorDetails(details)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	errorObj := unmarshaled["error"].(map[string]interface{})
	detailsObj := errorObj["details"].(map[string]interface{})
	assert.Equal(t, "email", detailsObj["field"])
	assert.Equal(t, "not-an-email", detailsObj["actual"])
}

// Test 10: Response with validation result and errors
func TestResponseWithValidationResult(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("name", "Name is required")
	vr.AddWarning("email", "Email format is unusual")
	vr.SetContext("field_count", 2)

	data := map[string]interface{}{"id": "123"}
	resp := NewApiResponse(data).WithValidation(vr)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.NotNil(t, unmarshaled["data"])
	assert.NotNil(t, unmarshaled["validation"])
	validationObj := unmarshaled["validation"].(map[string]interface{})
	assert.NotNil(t, validationObj["errors"])
	assert.NotNil(t, validationObj["warnings"])
}

// Test 11: Roundtrip with simple data types
func TestRoundtripSimpleData(t *testing.T) {
	originalData := map[string]interface{}{"id": "123", "value": 42}
	originalResp := NewApiResponse(originalData)

	// Marshal
	jsonBytes, err := json.Marshal(originalResp)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaledResp ApiResponse[map[string]interface{}]
	err = json.Unmarshal(jsonBytes, &unmarshaledResp)
	require.NoError(t, err)

	// Verify
	assert.NotNil(t, unmarshaledResp.Data)
	assert.NotNil(t, unmarshaledResp.Meta)
	assert.Nil(t, unmarshaledResp.Error)
}

// Test 12: Roundtrip with validation result
func TestRoundtripWithValidationResult(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("field1", "Error message 1")
	vr.AddWarning("field2", "Warning message 1")
	vr.AddInfo("field3", "Info message 1")

	originalResp := NewApiResponse("").WithValidation(vr)

	// Marshal
	jsonBytes, err := json.Marshal(originalResp)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaledResp ApiResponse[string]
	err = json.Unmarshal(jsonBytes, &unmarshaledResp)
	require.NoError(t, err)

	// Verify
	assert.NotNil(t, unmarshaledResp.Validation)
	assert.Equal(t, 1, unmarshaledResp.Validation.ErrorCount())
	assert.Equal(t, 1, unmarshaledResp.Validation.WarningCount())
	assert.Equal(t, 1, unmarshaledResp.Validation.InfoCount())
}

// Test 13: Roundtrip with error response and details
func TestRoundtripErrorResponseWithDetails(t *testing.T) {
	details := map[string]interface{}{
		"attempted_operation": "create_schedule",
		"resource_type":       "Schedule",
		"timestamp":           "2025-11-15T10:30:00Z",
	}
	originalResp := NewApiResponse("").
		WithError("DATABASE_ERROR", "Database operation failed").
		WithErrorDetails(details)

	// Marshal
	jsonBytes, err := json.Marshal(originalResp)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaledResp ApiResponse[string]
	err = json.Unmarshal(jsonBytes, &unmarshaledResp)
	require.NoError(t, err)

	// Verify
	assert.NotNil(t, unmarshaledResp.Error)
	assert.Equal(t, "DATABASE_ERROR", unmarshaledResp.Error.Code)
	assert.NotNil(t, unmarshaledResp.Error.Details)
}

// Test 14: Null error field is omitted in JSON
func TestNullErrorFieldOmitted(t *testing.T) {
	resp := NewApiResponse("data")

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.NotContains(t, jsonStr, "\"error\":null")
}

// Test 15: Null validation errors field
func TestValidationErrorsPresent(t *testing.T) {
	vr := validation.NewValidationResult()
	resp := NewApiResponse("data").WithValidation(vr)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	validationObj := unmarshaled["validation"].(map[string]interface{})
	assert.NotNil(t, validationObj["errors"])
}

// Test 16: IsSuccess returns true for successful response
func TestIsSuccessTrueForSuccess(t *testing.T) {
	resp := NewApiResponse("data")
	assert.True(t, resp.IsSuccess())
}

// Test 17: IsSuccess returns false when error is set
func TestIsSuccessFalseWithError(t *testing.T) {
	resp := NewApiResponse("data").
		WithError("ERROR_CODE", "Error message")
	assert.False(t, resp.IsSuccess())
}

// Test 18: IsSuccess returns false when validation has errors
func TestIsSuccessFalseWithValidationErrors(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("field", "Error message")
	resp := NewApiResponse("data").WithValidation(vr)
	assert.False(t, resp.IsSuccess())
}

// Test 19: IsSuccess returns true with only warnings
func TestIsSuccessTrueWithWarnings(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddWarning("field", "Warning message")
	resp := NewApiResponse("data").WithValidation(vr)
	assert.True(t, resp.IsSuccess())
}

// Test 20: Method chaining works correctly
func TestMethodChaining(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddWarning("deprecation", "Using deprecated endpoint")

	details := map[string]interface{}{"retry_after": 30}
	resp := NewApiResponse("test").
		WithValidation(vr).
		WithError("RATE_LIMITED", "Too many requests").
		WithErrorDetails(details)

	assert.NotNil(t, resp.Error)
	assert.NotNil(t, resp.Validation)
	assert.Equal(t, "RATE_LIMITED", resp.Error.Code)
}

// Test 21: ResponseMeta contains valid timestamp
func TestMetaTimestampValid(t *testing.T) {
	before := time.Now()
	resp := NewApiResponse("data")
	after := time.Now()

	assert.True(t, resp.Meta.Timestamp.After(before) || resp.Meta.Timestamp.Equal(before))
	assert.True(t, resp.Meta.Timestamp.Before(after) || resp.Meta.Timestamp.Equal(after))
}

// Test 22: ResponseMeta contains request ID
func TestMetaRequestIDNotEmpty(t *testing.T) {
	resp := NewApiResponse("data")
	assert.NotEmpty(t, resp.Meta.RequestID)
	assert.Equal(t, 36, len(resp.Meta.RequestID)) // UUID length with hyphens
}

// Test 23: ResponseMeta contains version
func TestMetaVersionPresent(t *testing.T) {
	resp := NewApiResponse("data")
	assert.Equal(t, "1.0", resp.Meta.Version)
}

// Test 24: ResponseMeta contains server time
func TestMetaServerTimePresent(t *testing.T) {
	resp := NewApiResponse("data")
	assert.Greater(t, resp.Meta.ServerTime, int64(0))
}

// Test 25: Empty validation result serialization
func TestEmptyValidationResult(t *testing.T) {
	vr := validation.NewValidationResult()
	resp := NewApiResponse("data").WithValidation(vr)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled ApiResponse[string]
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, 0, unmarshaled.Validation.ErrorCount())
	assert.Equal(t, 0, unmarshaled.Validation.WarningCount())
	assert.Equal(t, 0, unmarshaled.Validation.InfoCount())
}

// Test 26: Complex nested data types
func TestComplexNestedDataTypes(t *testing.T) {
	data := map[string]interface{}{
		"schedules": []interface{}{
			map[string]interface{}{
				"id":   "sched-1",
				"date": "2025-11-15",
				"shifts": []interface{}{
					map[string]interface{}{"position": "Doctor", "start": "08:00"},
					map[string]interface{}{"position": "Nurse", "start": "09:00"},
				},
			},
		},
	}
	resp := NewApiResponse(data)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled ApiResponse[map[string]interface{}]
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	schedules := unmarshaled.Data["schedules"].([]interface{})
	assert.Len(t, schedules, 1)
}

// Test 27: Struct type marshaling
func TestResponseWithCustomStruct(t *testing.T) {
	type Schedule struct {
		ID        string `json:"id"`
		StartDate string `json:"start_date"`
		Duration  int    `json:"duration"`
	}

	schedule := Schedule{
		ID:        "sched-001",
		StartDate: "2025-11-15",
		Duration:  480,
	}

	resp := NewApiResponse(schedule)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled ApiResponse[Schedule]
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, "sched-001", unmarshaled.Data.ID)
	assert.Equal(t, "2025-11-15", unmarshaled.Data.StartDate)
	assert.Equal(t, 480, unmarshaled.Data.Duration)
}

// Test 28: Roundtrip with custom struct preserves all fields
func TestRoundtripCustomStruct(t *testing.T) {
	type User struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	user := User{
		ID:    "user-123",
		Name:  "Alice",
		Email: "alice@example.com",
	}

	originalResp := NewApiResponse(user)

	// Marshal
	jsonBytes, err := json.Marshal(originalResp)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaledResp ApiResponse[User]
	err = json.Unmarshal(jsonBytes, &unmarshaledResp)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, "user-123", unmarshaledResp.Data.ID)
	assert.Equal(t, "Alice", unmarshaledResp.Data.Name)
	assert.Equal(t, "alice@example.com", unmarshaledResp.Data.Email)
}

// Test 29: Validation context preservation
func TestValidationContextPreservation(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("parsing", "Parse error on line 10")
	vr.SetContext("line_number", 10)
	vr.SetContext("column_number", 5)
	vr.SetContext("file_name", "schedule.ods")

	resp := NewApiResponse("").WithValidation(vr)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled ApiResponse[string]
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.NotNil(t, unmarshaled.Validation.Context)
	// JSON numbers become float64 when unmarshaled
	assert.Equal(t, float64(10), unmarshaled.Validation.Context["line_number"])
	assert.Equal(t, float64(5), unmarshaled.Validation.Context["column_number"])
	assert.Equal(t, "schedule.ods", unmarshaled.Validation.Context["file_name"])
}

// Test 30: Response with all fields populated
func TestResponseWithAllFieldsPopulated(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddWarning("deprecation", "Using deprecated endpoint")

	details := map[string]interface{}{"retry_after": 30}
	resp := NewApiResponse("data").
		WithValidation(vr).
		WithError("RATE_LIMITED", "Too many requests").
		WithErrorDetails(details)

	assert.NotNil(t, resp.Error)
	assert.NotNil(t, resp.Validation)
	assert.NotNil(t, resp.Meta)
	assert.Equal(t, "RATE_LIMITED", resp.Error.Code)
	assert.Equal(t, "1.0", resp.Meta.Version)
}

// Test 31: Different error codes produce different responses
func TestErrorCodesProduceDifferentResponses(t *testing.T) {
	resp1 := NewApiResponse("").WithError("INVALID_REQUEST", "msg")
	resp2 := NewApiResponse("").WithError("NOT_FOUND", "msg")
	resp3 := NewApiResponse("").WithError("UNAUTHORIZED", "msg")

	json1, _ := json.Marshal(resp1)
	json2, _ := json.Marshal(resp2)
	json3, _ := json.Marshal(resp3)

	assert.NotEqual(t, json1, json2)
	assert.NotEqual(t, json2, json3)
}

// Test 32: Large payload handling
func TestLargePayloadHandling(t *testing.T) {
	largeData := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeData[i] = map[string]interface{}{
			"id":    i,
			"value": "item",
		}
	}

	resp := NewApiResponse(largeData)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled ApiResponse[[]map[string]interface{}]
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Len(t, unmarshaled.Data, 1000)
}

// Test 33: Special characters handling in strings
func TestSpecialCharactersEscaping(t *testing.T) {
	data := map[string]interface{}{
		"quoted":    "She said \"hello\"",
		"newline":   "line1\nline2",
		"backslash": "path\\to\\file",
	}

	resp := NewApiResponse(data)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled ApiResponse[map[string]interface{}]
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, "She said \"hello\"", unmarshaled.Data["quoted"])
	assert.Equal(t, "line1\nline2", unmarshaled.Data["newline"])
}

// Test 34: WithErrorDetails on response without prior error
func TestWithErrorDetailsWithoutPriorError(t *testing.T) {
	resp := NewApiResponse("data").
		WithErrorDetails(map[string]interface{}{"key": "value"})

	assert.NotNil(t, resp.Error)
	assert.NotNil(t, resp.Error.Details)
	assert.Equal(t, "value", resp.Error.Details["key"])
}

// Test 35: Validation result messages are preserved
func TestValidationMessagesPreserved(t *testing.T) {
	vr := validation.NewValidationResult()
	vr.AddError("email", "Email is invalid")
	vr.AddError("phone", "Phone is invalid")
	vr.AddWarning("age", "Age is unusually high")

	resp := NewApiResponse("").WithValidation(vr)

	jsonBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled ApiResponse[string]
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, 2, len(unmarshaled.Validation.Errors))
	assert.Equal(t, 1, len(unmarshaled.Validation.Warnings))
	assert.Equal(t, "Email is invalid", unmarshaled.Validation.Errors[0].Message)
}

// Benchmark: JSON marshaling performance
func BenchmarkMarshalSuccessResponse(b *testing.B) {
	resp := NewApiResponse(map[string]interface{}{
		"id":   "123",
		"name": "Test",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}

// Benchmark: JSON unmarshaling performance
func BenchmarkUnmarshalSuccessResponse(b *testing.B) {
	resp := NewApiResponse(map[string]interface{}{
		"id":   "123",
		"name": "Test",
	})
	jsonBytes, _ := json.Marshal(resp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var resp ApiResponse[map[string]interface{}]
		_ = json.Unmarshal(jsonBytes, &resp)
	}
}

// Benchmark: Full roundtrip
func BenchmarkRoundtripResponse(b *testing.B) {
	resp := NewApiResponse(map[string]interface{}{
		"id":    "123",
		"name":  "Test",
		"items": []string{"a", "b", "c"},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonBytes, _ := json.Marshal(resp)
		var respResult ApiResponse[map[string]interface{}]
		_ = json.Unmarshal(jsonBytes, &respResult)
	}
}

// Example JSON outputs for documentation
func TestJSONOutputExamples(t *testing.T) {
	// Example 1: Success response with data
	resp1 := NewApiResponse(map[string]interface{}{"id": "123", "name": "Test"})
	json1, _ := json.MarshalIndent(resp1, "", "  ")
	t.Logf("Example 1 - Success Response:\n%s\n", string(json1))

	// Example 2: Error response
	resp2 := NewApiResponse("").
		WithError("INVALID_REQUEST", "Request is invalid")
	json2, _ := json.MarshalIndent(resp2, "", "  ")
	t.Logf("Example 2 - Error Response:\n%s\n", string(json2))

	// Example 3: Validation errors
	vr := validation.NewValidationResult()
	vr.AddError("file_format", "Expected ODS format")
	resp3 := NewApiResponse("").WithValidation(vr)
	json3, _ := json.MarshalIndent(resp3, "", "  ")
	t.Logf("Example 3 - Validation Error Response:\n%s\n", string(json3))

	// Verify all marshal without error
	assert.NotNil(t, json1)
	assert.NotNil(t, json2)
	assert.NotNil(t, json3)
}
