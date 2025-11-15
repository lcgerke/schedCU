package ods

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// createTestODSFile creates a minimal valid ODS file in memory as bytes.
// Returns the bytes of the ODS file (which is a ZIP archive).
func createTestODSFile(t *testing.T, contentXML string) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	defer zw.Close()

	// Create content.xml entry
	f, err := zw.Create("content.xml")
	if err != nil {
		t.Fatalf("failed to create content.xml in zip: %v", err)
	}

	_, err = f.Write([]byte(contentXML))
	if err != nil {
		t.Fatalf("failed to write content.xml: %v", err)
	}

	// Create mimetype file (required for ODS files)
	f, err = zw.Create("mimetype")
	if err != nil {
		t.Fatalf("failed to create mimetype in zip: %v", err)
	}

	_, err = f.Write([]byte("application/vnd.oasis.opendocument.spreadsheet"))
	if err != nil {
		t.Fatalf("failed to write mimetype: %v", err)
	}

	return buf.Bytes()
}

// writeTestODSFile writes a test ODS file to disk and returns the path.
func writeTestODSFile(t *testing.T, filename string, contentXML string) string {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	odsBytes := createTestODSFile(t, contentXML)

	err := os.WriteFile(filePath, odsBytes, 0644)
	if err != nil {
		t.Fatalf("failed to write test ODS file: %v", err)
	}

	return filePath
}

// simpleODSDocument returns XML for a simple one-sheet ODS document.
func simpleODSDocument() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table name="Sheet1">
				<table-row>
					<table-cell valueType="string">
						<p>Hello</p>
					</table-cell>
					<table-cell valueType="string">
						<p>World</p>
					</table-cell>
				</table-row>
			</table>
		</spreadsheet>
	</body>
</document>`
}

// ============================================================================
// OPENODSFILE TESTS
// ============================================================================

// TestOpenODSFileValidSimpleFile tests opening a valid simple ODS file.
func TestOpenODSFileValidSimpleFile(t *testing.T) {
	filePath := writeTestODSFile(t, "test_simple.ods", simpleODSDocument())

	doc, err := OpenODSFile(filePath)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if doc == nil {
		t.Fatal("expected document, got nil")
	}

	if len(doc.Sheets) == 0 {
		t.Fatal("expected at least one sheet")
	}

	if doc.Sheets[0].Name != "Sheet1" {
		t.Errorf("expected sheet name 'Sheet1', got '%s'", doc.Sheets[0].Name)
	}
}

// TestOpenODSFileNotExist tests opening a non-existent file.
func TestOpenODSFileNotExist(t *testing.T) {
	doc, err := OpenODSFile("/nonexistent/path/file.ods")

	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if doc != nil {
		t.Error("expected nil document for non-existent file")
	}
}

// TestOpenODSFileEmptyPath tests opening with empty path.
func TestOpenODSFileEmptyPath(t *testing.T) {
	doc, err := OpenODSFile("")

	if err == nil {
		t.Fatal("expected error for empty path")
	}

	if doc != nil {
		t.Error("expected nil document for empty path")
	}
}

// TestOpenODSFileInvalidZip tests opening a file that's not a valid ZIP.
func TestOpenODSFileInvalidZip(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.ods")

	// Write invalid ZIP content
	err := os.WriteFile(filePath, []byte("not a zip file"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	doc, err := OpenODSFile(filePath)

	if err == nil {
		t.Fatal("expected error for invalid ZIP file")
	}

	if doc != nil {
		t.Error("expected nil document for invalid file")
	}
}

// TestOpenODSFileMissingContentXML tests ODS file without content.xml.
func TestOpenODSFileMissingContentXML(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// Add only mimetype, no content.xml
	f, err := zw.Create("mimetype")
	if err != nil {
		t.Fatalf("failed to create mimetype: %v", err)
	}
	f.Write([]byte("application/vnd.oasis.opendocument.spreadsheet"))

	zw.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "no_content.ods")
	err = os.WriteFile(filePath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	doc, err := OpenODSFile(filePath)

	if err == nil {
		t.Fatal("expected error for missing content.xml")
	}

	if doc != nil {
		t.Error("expected nil document for missing content.xml")
	}
}

// ============================================================================
// SHEET EXTRACTION TESTS
// ============================================================================

// TestExtractSheetByName tests extracting a sheet by name.
func TestExtractSheetByName(t *testing.T) {
	contentXML := `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table name="Data">
				<table-row>
					<table-cell valueType="string">
						<p>Name</p>
					</table-cell>
					<table-cell valueType="string">
						<p>Age</p>
					</table-cell>
				</table-row>
				<table-row>
					<table-cell valueType="string">
						<p>John</p>
					</table-cell>
					<table-cell valueType="float" value="30">
						<p>30</p>
					</table-cell>
				</table-row>
			</table>
		</spreadsheet>
	</body>
</document>`

	filePath := writeTestODSFile(t, "test_extract.ods", contentXML)
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	data, err := doc.ExtractSheet("Data")
	if err != nil {
		t.Fatalf("failed to extract sheet: %v", err)
	}

	if len(data) != 2 {
		t.Errorf("expected 2 rows, got %d", len(data))
	}

	if len(data[0]) < 2 {
		t.Errorf("expected at least 2 columns, got %d", len(data[0]))
	}

	if data[0][0] != "Name" {
		t.Errorf("expected 'Name' in cell [0][0], got '%s'", data[0][0])
	}
}

// TestExtractSheetNotFound tests extracting a non-existent sheet.
func TestExtractSheetNotFound(t *testing.T) {
	filePath := writeTestODSFile(t, "test_notfound.ods", `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table name="Sheet1">
			</table>
		</spreadsheet>
	</body>
</document>`)

	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	_, err = doc.ExtractSheet("NonExistent")
	if err == nil {
		t.Fatal("expected error for non-existent sheet")
	}
}

// TestExtractSheetByIndex tests extracting a sheet by index.
func TestExtractSheetByIndex(t *testing.T) {
	contentXML := `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table name="First">
				<table-row>
					<table-cell valueType="string">
						<p>Sheet1</p>
					</table-cell>
				</table-row>
			</table>
			<table name="Second">
				<table-row>
					<table-cell valueType="string">
						<p>Sheet2</p>
					</table-cell>
				</table-row>
			</table>
		</spreadsheet>
	</body>
</document>`

	filePath := writeTestODSFile(t, "test_index.ods", contentXML)
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	if len(doc.Sheets) != 2 {
		t.Errorf("expected 2 sheets, got %d", len(doc.Sheets))
	}

	data, err := doc.ExtractSheetByIndex(1)
	if err != nil {
		t.Fatalf("failed to extract sheet by index: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected at least 1 row")
	}

	if data[0][0] != "Sheet2" {
		t.Errorf("expected 'Sheet2', got '%s'", data[0][0])
	}
}

// TestExtractSheetByIndexOutOfRange tests index out of range.
func TestExtractSheetByIndexOutOfRange(t *testing.T) {
	filePath := writeTestODSFile(t, "test_range.ods", simpleODSDocument())
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	_, err = doc.ExtractSheetByIndex(5)
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

// ============================================================================
// MULTIPLE SHEETS TESTS
// ============================================================================

// TestOpenODSFileMultipleSheets tests opening a file with multiple sheets.
func TestOpenODSFileMultipleSheets(t *testing.T) {
	contentXML := `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table name="Sales">
				<table-row>
					<table-cell valueType="string">
						<p>Item</p>
					</table-cell>
				</table-row>
			</table>
			<table name="Inventory">
				<table-row>
					<table-cell valueType="string">
						<p>Stock</p>
					</table-cell>
				</table-row>
			</table>
			<table name="Reports">
				<table-row>
					<table-cell valueType="string">
						<p>Total</p>
					</table-cell>
				</table-row>
			</table>
		</spreadsheet>
	</body>
</document>`

	filePath := writeTestODSFile(t, "test_multi.ods", contentXML)
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	if len(doc.Sheets) != 3 {
		t.Errorf("expected 3 sheets, got %d", len(doc.Sheets))
	}

	names := doc.SheetNames()
	if len(names) != 3 {
		t.Errorf("expected 3 sheet names, got %d", len(names))
	}

	expectedNames := []string{"Sales", "Inventory", "Reports"}
	for i, name := range names {
		if name != expectedNames[i] {
			t.Errorf("expected sheet %d to be '%s', got '%s'", i, expectedNames[i], name)
		}
	}
}

// ============================================================================
// DATA TYPE TESTS
// ============================================================================

// TestODSCellDataTypes tests extraction of various data types.
func TestODSCellDataTypes(t *testing.T) {
	contentXML := `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table name="Types">
				<table-row>
					<table-cell valueType="string">
						<p>Text Value</p>
					</table-cell>
					<table-cell valueType="float" value="42.5">
						<p>42.5</p>
					</table-cell>
					<table-cell valueType="date" dateValue="2024-01-15">
						<p>2024-01-15</p>
					</table-cell>
					<table-cell valueType="boolean" booleanValue="true">
						<p>true</p>
					</table-cell>
				</table-row>
			</table>
		</spreadsheet>
	</body>
</document>`

	filePath := writeTestODSFile(t, "test_types.ods", contentXML)
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	data, err := doc.ExtractSheet("Types")
	if err != nil {
		t.Fatalf("failed to extract sheet: %v", err)
	}

	if len(data[0]) < 4 {
		t.Errorf("expected 4 columns, got %d", len(data[0]))
	}

	if data[0][0] != "Text Value" {
		t.Errorf("expected 'Text Value', got '%s'", data[0][0])
	}
}

// ============================================================================
// STATISTICS TESTS
// ============================================================================

// TestODSDocumentStats tests that statistics are correctly calculated.
func TestODSDocumentStats(t *testing.T) {
	contentXML := `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table name="Stats">
				<table-row>
					<table-cell valueType="string">
						<p>A1</p>
					</table-cell>
					<table-cell valueType="string">
						<p>B1</p>
					</table-cell>
				</table-row>
				<table-row>
					<table-cell valueType="string">
						<p>A2</p>
					</table-cell>
					<table-cell valueType="string">
						<p>B2</p>
					</table-cell>
				</table-row>
			</table>
		</spreadsheet>
	</body>
</document>`

	filePath := writeTestODSFile(t, "test_stats.ods", contentXML)
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	if doc.Stats.TotalSheets != 1 {
		t.Errorf("expected 1 sheet, got %d", doc.Stats.TotalSheets)
	}

	if doc.Stats.TotalRows != 2 {
		t.Errorf("expected 2 rows, got %d", doc.Stats.TotalRows)
	}

	if doc.Stats.FileSizeBytes == 0 {
		t.Error("expected non-zero file size")
	}
}

// ============================================================================
// VALIDATION RESULT TESTS
// ============================================================================

// TestODSToValidationResult tests converting ODS document to ValidationResult.
func TestODSToValidationResult(t *testing.T) {
	filePath := writeTestODSFile(t, "test_validation.ods", simpleODSDocument())
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	result := doc.ToValidationResult()
	if result == nil {
		t.Fatal("expected non-nil ValidationResult")
	}

	// Check context
	if sheets, ok := result.GetContext("sheets_count"); !ok || sheets != 1 {
		t.Error("expected sheets_count in context")
	}
}

// ============================================================================
// EDGE CASES
// ============================================================================

// TestODSEmptySheet tests handling of empty sheets.
func TestODSEmptySheet(t *testing.T) {
	contentXML := `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table name="Empty">
			</table>
		</spreadsheet>
	</body>
</document>`

	filePath := writeTestODSFile(t, "test_empty.ods", contentXML)
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	if len(doc.Sheets) != 1 {
		t.Errorf("expected 1 sheet, got %d", len(doc.Sheets))
	}

	if doc.Sheets[0].RowCount != 0 {
		t.Errorf("expected empty sheet, got %d rows", doc.Sheets[0].RowCount)
	}
}

// TestODSDefaultSheetName tests that unnamed sheets get default names.
func TestODSDefaultSheetName(t *testing.T) {
	contentXML := `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table>
				<table-row>
					<table-cell valueType="string">
						<p>Test</p>
					</table-cell>
				</table-row>
			</table>
		</spreadsheet>
	</body>
</document>`

	filePath := writeTestODSFile(t, "test_default_name.ods", contentXML)
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	if doc.Sheets[0].Name != "Sheet1" {
		t.Errorf("expected default name 'Sheet1', got '%s'", doc.Sheets[0].Name)
	}
}

// TestODSGetSheet tests the GetSheet method.
func TestODSGetSheet(t *testing.T) {
	contentXML := `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
          xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
          xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
	<body>
		<spreadsheet>
			<table name="MySheet">
				<table-row>
					<table-cell valueType="string">
						<p>Data</p>
					</table-cell>
				</table-row>
			</table>
		</spreadsheet>
	</body>
</document>`

	filePath := writeTestODSFile(t, "test_getsheet.ods", contentXML)
	doc, err := OpenODSFile(filePath)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	sheet := doc.GetSheet("MySheet")
	if sheet == nil {
		t.Fatal("expected to find sheet")
	}

	if sheet.Name != "MySheet" {
		t.Errorf("expected sheet name 'MySheet', got '%s'", sheet.Name)
	}

	notFound := doc.GetSheet("DoesNotExist")
	if notFound != nil {
		t.Error("expected nil for non-existent sheet")
	}
}
