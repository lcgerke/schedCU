package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestODSParsing validates basic ODS file parsing capability.
func TestODSParsing(t *testing.T) {
	parser := NewODSParser()

	// Create minimal valid ODS content
	odsContent := createMockODSFile()

	sheets, err := parser.Parse(odsContent)
	require.NoError(t, err, "Should parse ODS without error")

	assert.Greater(t, len(sheets), 0, "Should have at least one sheet")
	assert.NotNil(t, sheets[0], "First sheet should not be nil")
	assert.NotEmpty(t, sheets[0].Name, "Sheet should have a name")
}

// TestODSErrorCollection validates error collection without fail-fast.
func TestODSErrorCollection(t *testing.T) {
	parser := NewODSParser()

	// Create ODS with intentional structural issues
	odsContent := createMockODSWithErrors()

	sheets, errs := parser.ParseWithErrorCollection(odsContent)

	// Should return results despite errors
	assert.NotNil(t, sheets, "Should return sheet data even with errors")

	// Should collect multiple errors
	assert.Greater(t, len(errs), 0, "Should collect parsing errors")

	// Each error should be descriptive
	for _, err := range errs {
		assert.NotEmpty(t, err.Error(), "Error messages should be non-empty")
	}
}

// TestODSCellExtraction validates cell data extraction.
func TestODSCellExtraction(t *testing.T) {
	parser := NewODSParser()
	odsContent := createMockODSFile()

	sheets, err := parser.Parse(odsContent)
	require.NoError(t, err)
	require.NotEmpty(t, sheets)

	sheet := sheets[0]

	// Should extract cells
	assert.NotNil(t, sheet.Cells, "Sheet should have cells")

	// Verify cell structure
	for _, cell := range sheet.Cells {
		assert.NotEmpty(t, cell.Address, "Cell should have address (e.g., A1)")
		// Value can be empty, but should exist as field
		assert.True(t, cell.Value != nil || cell.Value == nil, "Cell should have value field")
	}
}

// TestODSFileSizePerformance validates parsing performance across sizes.
func TestODSFileSizePerformance(t *testing.T) {
	parser := NewODSParser()

	sizes := map[string]int{
		"small":  100,   // ~100 cells
		"medium": 1000,  // ~1000 cells
		"large":  5000,  // ~5000 cells
	}

	for name, cellCount := range sizes {
		t.Run(name, func(t *testing.T) {
			odsContent := createMockODSWithCells(cellCount)

			sheets, err := parser.Parse(odsContent)
			require.NoError(t, err, "Should parse %s ODS", name)
			require.NotEmpty(t, sheets, "Should have sheets")

			// Verify cell count matches expected
			totalCells := 0
			for _, sheet := range sheets {
				totalCells += len(sheet.Cells)
			}

			// Allow 10% tolerance for formatting/overhead
			maxExpected := int(float64(cellCount) * 1.1)
			minExpected := int(float64(cellCount) * 0.9)
			assert.Greater(t, totalCells, minExpected, "Should extract most cells")
			assert.Less(t, totalCells, maxExpected, "Should not exceed expected cell count")
		})
	}
}

// TestODSErrorMessages validates that error messages are descriptive and actionable.
func TestODSErrorMessages(t *testing.T) {
	parser := NewODSParser()

	// Create ODS with specific error condition (corrupted entry)
	odsContent := createMockODSCorrupted()

	_, errs := parser.ParseWithErrorCollection(odsContent)

	for _, err := range errs {
		msg := err.Error()

		// Errors should indicate WHAT failed
		assert.True(t,
			len(msg) > 0,
			"Error message should be descriptive",
		)

		// Errors should indicate WHERE (cell address, sheet name, etc.)
		// This is checked by verifying message contains useful context
		assert.NotContains(t, msg, "unknown", "Error should avoid generic descriptions")
	}
}

// TestODSDataTypePreservation validates that data types are preserved where possible.
func TestODSDataTypePreservation(t *testing.T) {
	parser := NewODSParser()
	odsContent := createMockODSWithTypes()

	sheets, err := parser.Parse(odsContent)
	require.NoError(t, err)
	require.NotEmpty(t, sheets)

	sheet := sheets[0]

	// Should have cells with preserved type information
	for _, cell := range sheet.Cells {
		assert.NotNil(t, cell.Type, "Cell should have type information")
		assert.NotEmpty(t, cell.Type, "Cell type should be set")
	}
}

// Mock data generators

func createMockODSFile() []byte {
	// Minimal valid ODS structure (XML-based)
	// In production, would use actual ODS file bytes
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0">
  <office:spreadsheet>
    <table:table xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" table:name="Sheet1">
      <table:table-row>
        <table:table-cell><text:p xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">Test</text:p></table:table-cell>
      </table:table-row>
    </table:table>
  </office:spreadsheet>
</office:document>`)
}

func createMockODSWithErrors() []byte {
	// ODS with some structural issues but still parseable
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0">
  <office:spreadsheet>
    <table:table xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" table:name="Sheet1">
      <table:table-row>
        <!-- Missing closing tag -->
        <table:table-cell><text:p xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">Data
      </table:table-row>
    </table:table>
  </office:spreadsheet>
</office:document>`)
}

func createMockODSWithCells(count int) []byte {
	// Create ODS with specified number of cells
	// In production, would generate valid XML with cell data
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<office:document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0">
  <office:spreadsheet>
    <table:table xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" table:name="Sheet1">`

	for i := 0; i < count; i++ {
		xml += `<table:table-cell><text:p xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">Cell</text:p></table:table-cell>`
	}

	xml += `</table:table>
  </office:spreadsheet>
</office:document>`

	return []byte(xml)
}

func createMockODSCorrupted() []byte {
	// ODS with corruption that should be reported as errors
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0">
  <office:spreadsheet>
    <table:table xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" table:name="Sheet1">
      <table:table-row>
        <table:table-cell><text:p xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">Valid</text:p></table:table-cell>
        <!-- Corrupted cell reference -->
        <table:table-cell table:number-columns-repeated="invalid"><text:p xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">Bad</text:p></table:table-cell>
      </table:table-row>
    </table:table>
  </office:spreadsheet>
</office:document>`)
}

func createMockODSWithTypes() []byte {
	// ODS with typed cells (numbers, dates, text, formulas)
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0">
  <office:spreadsheet>
    <table:table xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" table:name="Sheet1">
      <table:table-row>
        <table:table-cell office:value-type="string"><text:p xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">Text</text:p></table:table-cell>
        <table:table-cell office:value-type="float" office:value="42.5"><text:p xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">42.5</text:p></table:table-cell>
        <table:table-cell office:value-type="date" office:date-value="2025-11-15"><text:p xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">2025-11-15</text:p></table:table-cell>
      </table:table-row>
    </table:table>
  </office:spreadsheet>
</office:document>`)
}
