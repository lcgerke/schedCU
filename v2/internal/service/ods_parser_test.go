package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/schedcu/v2/internal/validation"
)

// TestODSParserWithRealFile tests parsing the actual cuSchedNormalized.ods file
func TestODSParserWithRealFile(t *testing.T) {
	parser := NewODSParser()
	result := validation.NewResult()

	odsData, err := parser.ParseFile("/home/lcgerke/schedCU/cuSchedNormalized.ods", result)

	require.NoError(t, err)
	require.NotNil(t, odsData)

	// Verify file name is set
	assert.Equal(t, "/home/lcgerke/schedCU/cuSchedNormalized.ods", odsData.FileName)

	// Verify we found sheets
	assert.Greater(t, len(odsData.Sheets), 0, "Should find at least one sheet")

	// Verify sheet metadata was parsed
	for _, sheet := range odsData.Sheets {
		assert.NotEmpty(t, sheet.Name, "Sheet should have a name")
		assert.NotEmpty(t, sheet.ShiftCategory, "Sheet should have shift category (MID or ON)")
		assert.NotEmpty(t, sheet.DayType, "Sheet should have day type (WEEKDAY or WEEKEND)")
		assert.NotEmpty(t, sheet.SpecialtyScenario, "Sheet should have specialty scenario (BODY or NEURO)")
	}

	// Verify we found coverage data
	totalCoverage := 0
	for _, sheet := range odsData.Sheets {
		totalCoverage += len(sheet.CoverageGrid)
	}
	assert.Greater(t, totalCoverage, 0, "Should find coverage assignments")

	// Verify no parsing errors (should be warnings or info only)
	errorCount := 0
	for _, msg := range result.Messages {
		if msg.Severity == validation.SeverityError {
			t.Logf("Parsing error: %s: %s", msg.Code, msg.Text)
			errorCount++
		}
	}
	assert.Equal(t, 0, errorCount, "Should have no parsing errors")
}

// TestSheetNameParser tests parsing of sheet names
func TestSheetNameParser(t *testing.T) {
	parser := NewSheetNameParser()

	tests := []struct {
		name     string
		input    string
		wantSC   string // ShiftCategory
		wantDT   string // DayType
		wantSS   string // SpecialtyScenario
		wantHour int    // TimeStart hour (or -1 if not set)
	}{
		{
			name:     "Mid Weekday Body with time",
			input:    "Mid Weekday Body 5 - 6 pm",
			wantSC:   "MID",
			wantDT:   "WEEKDAY",
			wantSS:   "BODY",
			wantHour: 17, // 5 pm
		},
		{
			name:     "ON Weekend Neuro without time",
			input:    "ON Weekend Neuro",
			wantSC:   "ON",
			wantDT:   "WEEKEND",
			wantSS:   "NEURO",
			wantHour: -1, // No time
		},
		{
			name:     "Mid Weekday Neuro 5 pm - 6 pm",
			input:    "Mid Weekday Neuro 5 pm - 6 pm",
			wantSC:   "MID",
			wantDT:   "WEEKDAY",
			wantSS:   "NEURO",
			wantHour: 17, // 5 pm
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			require.NotNil(t, result, "Should parse sheet name successfully")

			assert.Equal(t, tt.wantSC, result.ShiftCategory)
			assert.Equal(t, tt.wantDT, result.DayType)
			assert.Equal(t, tt.wantSS, result.SpecialtyScenario)

			if tt.wantHour > 0 {
				assert.Equal(t, tt.wantHour, result.TimeStart.Hour())
			}
		})
	}
}

// TestODSParserCoverageGrid tests coverage grid extraction
func TestODSParserCoverageGrid(t *testing.T) {
	parser := NewODSParser()

	// Mock grid data similar to real ODS structure
	grid := [][]string{
		{"", "Mid Body", "Mid Neuro", "Mid3", ""}, // Header row
		{"CPMC CT Neuro", "", "x", "", ""},         // Data row
		{"CPMC CT Body", "x", "", "", ""},          // Data row
	}

	result := validation.NewResult()
	coverage := parser.extractCoverageGridFromTable(grid, result)

	// Should find coverage assignments
	assert.Greater(t, len(coverage), 0, "Should find coverage assignments")

	// Verify coverage data structure
	for _, cell := range coverage {
		assert.NotEmpty(t, cell.Hospital, "Hospital should be set")
		assert.NotEmpty(t, cell.ShiftType, "ShiftType should be set")
		assert.Equal(t, "x", cell.Assignment, "Assignment should be 'x'")
	}
}

// TestSheetNameParserVariations tests different sheet name formats
func TestSheetNameParserVariations(t *testing.T) {
	parser := NewSheetNameParser()

	invalidNames := []string{
		"Random Sheet Name",
		"_System Sheet",
		"",
		"12345",
	}

	for _, name := range invalidNames {
		t.Run("Invalid: "+name, func(t *testing.T) {
			result := parser.Parse(name)
			assert.Nil(t, result, "Should not parse invalid sheet name: %s", name)
		})
	}
}
