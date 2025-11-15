package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidationResultCreation tests creating a new result
func TestValidationResultCreation(t *testing.T) {
	result := NewResult()

	assert.NotNil(t, result)
	assert.Empty(t, result.Messages)
	assert.True(t, result.IsValid())
	assert.True(t, result.CanImport())
	assert.True(t, result.CanPromote())
}

// TestAddError tests adding error messages
func TestAddError(t *testing.T) {
	result := NewResult()

	result.AddError(CodeUnknownShiftType, "Unknown shift type: XRAY_NIGHT on 2024-10-15")

	assert.Len(t, result.Messages, 1)
	assert.False(t, result.IsValid())
	assert.False(t, result.CanImport())
	assert.False(t, result.CanPromote())
	assert.Equal(t, 1, result.ErrorCount())
}

// TestAddWarning tests adding warning messages
func TestAddWarning(t *testing.T) {
	result := NewResult()

	result.AddWarning(CodeMissingMidC, "No MidC assignment on weekday 2024-10-16")

	assert.Len(t, result.Messages, 1)
	assert.True(t, result.IsValid()) // Warnings don't make it invalid
	assert.True(t, result.CanImport()) // Can import with warnings
	assert.False(t, result.CanPromote()) // Cannot promote with warnings
	assert.Equal(t, 1, result.WarningCount())
}

// TestAddInfo tests adding info messages
func TestAddInfo(t *testing.T) {
	result := NewResult()

	result.AddInfo("INFO_CODE", "This is informational")

	assert.Len(t, result.Messages, 1)
	assert.True(t, result.IsValid())
	assert.True(t, result.CanImport())
	assert.True(t, result.CanPromote())
	assert.Equal(t, 1, result.InfoCount())
}

// TestMultipleMessages tests collecting multiple messages
func TestMultipleMessages(t *testing.T) {
	result := NewResult()

	result.
		AddError(CodeUnknownPeople, "Unknown people in Amion: John D, Dr. Smith").
		AddWarning(CodeMissingMidC, "No MidC assignment on 2024-10-16").
		AddInfo("INFO_CODE", "Processing completed with warnings")

	assert.Len(t, result.Messages, 3)
	assert.Equal(t, 1, result.ErrorCount())
	assert.Equal(t, 1, result.WarningCount())
	assert.Equal(t, 1, result.InfoCount())
	assert.False(t, result.IsValid())
	assert.False(t, result.CanImport())
	assert.False(t, result.CanPromote())
}

// TestMessagesByCode tests filtering messages by code
func TestMessagesByCode(t *testing.T) {
	result := NewResult()

	result.
		AddError(CodeUnknownPeople, "Unknown person: John").
		AddError(CodeUnknownPeople, "Unknown person: Jane")

	messages := result.MessagesByCode(CodeUnknownPeople)

	assert.Len(t, messages, 2)
	for _, msg := range messages {
		assert.Equal(t, CodeUnknownPeople, msg.Code)
	}
}

// TestMessagesBySeverity tests filtering messages by severity
func TestMessagesBySeverity(t *testing.T) {
	result := NewResult()

	result.
		AddError(CodeUnknownShiftType, "Error 1").
		AddError(CodeUnknownShiftType, "Error 2").
		AddWarning(CodeMissingMidC, "Warning 1").
		AddInfo("CODE", "Info 1")

	errors := result.MessagesBySeverity(SeverityError)
	warnings := result.MessagesBySeverity(SeverityWarning)
	infos := result.MessagesBySeverity(SeverityInfo)

	assert.Len(t, errors, 2)
	assert.Len(t, warnings, 1)
	assert.Len(t, infos, 1)
}

// TestHasErrorsAndWarnings tests flag methods
func TestHasErrorsAndWarnings(t *testing.T) {
	resultClean := NewResult()
	assert.False(t, resultClean.HasErrors())
	assert.False(t, resultClean.HasWarnings())

	resultWithError := NewResult().AddError("CODE", "Error")
	assert.True(t, resultWithError.HasErrors())
	assert.False(t, resultWithError.HasWarnings())

	resultWithWarning := NewResult().AddWarning("CODE", "Warning")
	assert.False(t, resultWithWarning.HasErrors())
	assert.True(t, resultWithWarning.HasWarnings())

	resultWithBoth := NewResult().
		AddError("ERR", "Error").
		AddWarning("WARN", "Warning")
	assert.True(t, resultWithBoth.HasErrors())
	assert.True(t, resultWithBoth.HasWarnings())
}

// TestWithContext tests messages with additional context
func TestWithContext(t *testing.T) {
	result := NewResult()

	context := map[string]interface{}{
		"shift_type": "XRAY_NIGHT",
		"date":       "2024-10-15",
	}

	result.AddErrorWithContext(CodeUnknownShiftType, "Unknown shift type", context)

	assert.Len(t, result.Messages, 1)
	msg := result.Messages[0]
	assert.Equal(t, context, msg.Context)
	assert.Equal(t, "XRAY_NIGHT", msg.Context["shift_type"])
}

// TestToJSON tests JSON serialization
func TestToJSON(t *testing.T) {
	result := NewResult()

	result.
		AddError(CodeUnknownPeople, "Unknown person: John").
		AddWarning(CodeMissingMidC, "Missing MidC")

	json, err := result.ToJSON()

	assert.NoError(t, err)
	assert.NotEmpty(t, json)
	assert.Contains(t, json, "UNKNOWN_PEOPLE")
	assert.Contains(t, json, "MISSING_MIDC")
	assert.Contains(t, json, "ERROR")
	assert.Contains(t, json, "WARNING")
}

// TestFromJSON tests JSON deserialization
func TestFromJSON(t *testing.T) {
	original := NewResult()
	original.
		AddError(CodeUnknownPeople, "Unknown person: John").
		AddWarning(CodeMissingMidC, "Missing MidC")

	jsonStr, err := original.ToJSON()
	require.NoError(t, err)

	// Deserialize
	restored, err := FromJSON(jsonStr)
	require.NoError(t, err)

	assert.Len(t, restored.Messages, 2)
	assert.Equal(t, original.ErrorCount(), restored.ErrorCount())
	assert.Equal(t, original.WarningCount(), restored.WarningCount())
}

// TestSummary tests human-readable summary
func TestSummary(t *testing.T) {
	result := NewResult()
	result.
		AddError(CodeUnknownPeople, "Unknown person: John").
		AddWarning(CodeMissingMidC, "Missing MidC").
		AddInfo("INFO", "Done")

	summary := result.Summary()

	assert.Contains(t, summary, "1 errors")
	assert.Contains(t, summary, "1 warnings")
	assert.Contains(t, summary, "1 info")
	assert.Contains(t, summary, "UNKNOWN_PEOPLE")
	assert.Contains(t, summary, "MISSING_MIDC")
}

// TestChaining tests method chaining
func TestChaining(t *testing.T) {
	result := NewResult().
		AddError("CODE1", "Error 1").
		AddWarning("CODE2", "Warning 1").
		AddInfo("CODE3", "Info 1")

	assert.Len(t, result.Messages, 3)
	assert.Equal(t, 1, result.ErrorCount())
	assert.Equal(t, 1, result.WarningCount())
	assert.Equal(t, 1, result.InfoCount())
}

// TestRealWorldExample tests a real-world import scenario
func TestRealWorldExample(t *testing.T) {
	// Simulating ODS file import with multiple issues
	result := NewResult()

	// Found unknown shift types
	result.AddErrorWithContext(
		CodeUnknownShiftType,
		"Unknown shift type encountered",
		map[string]interface{}{
			"shift_type": "XRAY_NIGHT",
			"date":       "2024-10-15",
			"count":      3,
		},
	)

	// Found unknown people
	result.AddErrorWithContext(
		CodeUnknownPeople,
		"Unknown people in file",
		map[string]interface{}{
			"people": []string{"John D", "Dr. Smith"},
			"count":  2,
		},
	)

	// Missing coverage on specific date
	result.AddWarning(
		CodeMissingMidC,
		"No MidC coverage on weekday 2024-10-16",
	)

	// Informational: how many records processed
	result.AddInfo(
		"RECORDS_PROCESSED",
		"Processed 150 shift assignments",
	)

	// Cannot import due to errors
	assert.False(t, result.CanImport())
	// Cannot promote due to errors and warnings
	assert.False(t, result.CanPromote())
	// Has both errors and warnings
	assert.True(t, result.HasErrors())
	assert.True(t, result.HasWarnings())
}
