package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// FixtureConfig defines a test fixture to generate
type FixtureConfig struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"` // valid, invalid, partial, large
	ShiftCount    int      `json:"shift_count"`
	Columns       []string `json:"columns"`
	ExpectedError int      `json:"expected_errors"`
}

// FixtureMeta holds fixture metadata
type FixtureMeta struct {
	Fixtures []FixtureConfig `json:"fixtures"`
}

// ShiftData represents a single shift row
type ShiftData struct {
	Date                 string
	ShiftType            string
	RequiredStaffing     string
	SpecialtyConstraint  string
	StudyType            string
}

func main() {
	if err := generateAllFixtures(); err != nil {
		log.Fatalf("Failed to generate fixtures: %v", err)
	}
	fmt.Println("All fixtures generated successfully!")
}

func generateAllFixtures() error {
	fixturesDir := "/home/lcgerke/schedCU/reimplement/tests/fixtures/ods"

	// Create directory if it doesn't exist
	if err := os.MkdirAll(fixturesDir, 0755); err != nil {
		return fmt.Errorf("failed to create fixtures directory: %w", err)
	}

	// Generate fixture files
	configs := []FixtureConfig{
		{
			Name:          "valid_schedule.ods",
			Type:          "valid",
			ShiftCount:    150,
			Columns:       []string{"Date", "ShiftType", "RequiredStaffing", "SpecialtyConstraint", "StudyType"},
			ExpectedError: 0,
		},
		{
			Name:          "partial_schedule.ods",
			Type:          "partial",
			ShiftCount:    50,
			Columns:       []string{"Date", "ShiftType", "RequiredStaffing", "SpecialtyConstraint"},
			ExpectedError: 0,
		},
		{
			Name:          "invalid_schedule.ods",
			Type:          "invalid",
			ShiftCount:    30,
			Columns:       []string{"Date", "ShiftType", "RequiredStaffing", "SpecialtyConstraint", "StudyType"},
			ExpectedError: 4,
		},
		{
			Name:          "large_schedule.ods",
			Type:          "large",
			ShiftCount:    1200,
			Columns:       []string{"Date", "ShiftType", "RequiredStaffing", "SpecialtyConstraint", "StudyType"},
			ExpectedError: 0,
		},
	}

	meta := FixtureMeta{Fixtures: configs}

	// Generate each fixture
	for _, config := range configs {
		filePath := filepath.Join(fixturesDir, config.Name)
		fmt.Printf("Generating %s (%d shifts)...\n", config.Name, config.ShiftCount)

		var shifts []ShiftData
		switch config.Type {
		case "valid":
			shifts = generateValidShifts(config.ShiftCount, config.Columns)
		case "partial":
			shifts = generatePartialShifts(config.ShiftCount, config.Columns)
		case "invalid":
			shifts = generateInvalidShifts(config.ShiftCount, config.Columns)
		case "large":
			shifts = generateValidShifts(config.ShiftCount, config.Columns)
		}

		if err := createODSFile(filePath, config.Columns, shifts); err != nil {
			return fmt.Errorf("failed to create %s: %w", config.Name, err)
		}
	}

	// Write metadata file
	metaPath := filepath.Join(fixturesDir, "fixtures.json")
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	fmt.Printf("Created metadata at %s\n", metaPath)
	return nil
}

func generateValidShifts(count int, columns []string) []ShiftData {
	shifts := make([]ShiftData, count)
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	shiftTypes := []string{"Morning", "Afternoon", "Night"}
	specialties := []string{"Emergency", "ICU", "Surgery", "Pediatrics", "Cardiology"}

	for i := 0; i < count; i++ {
		date := startDate.AddDate(0, 0, i)
		shifts[i] = ShiftData{
			Date:                date.Format("2006-01-02"),
			ShiftType:           shiftTypes[i%len(shiftTypes)],
			RequiredStaffing:    fmt.Sprintf("%d", (i%10)+2),
			SpecialtyConstraint: specialties[i%len(specialties)],
			StudyType:           map[int]string{0: "Type-A", 1: "Type-B", 2: "Type-C"}[i%3],
		}
	}

	return shifts
}

func generatePartialShifts(count int, columns []string) []ShiftData {
	shifts := make([]ShiftData, count)
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	shiftTypes := []string{"Morning", "Afternoon", "Night"}
	specialties := []string{"Emergency", "ICU", "Surgery"}

	for i := 0; i < count; i++ {
		date := startDate.AddDate(0, 0, i)
		shifts[i] = ShiftData{
			Date:                date.Format("2006-01-02"),
			ShiftType:           shiftTypes[i%len(shiftTypes)],
			RequiredStaffing:    fmt.Sprintf("%d", (i%5)+1),
			SpecialtyConstraint: specialties[i%len(specialties)],
			StudyType:           "", // Omitted in partial
		}
	}

	return shifts
}

func generateInvalidShifts(count int, columns []string) []ShiftData {
	shifts := make([]ShiftData, 0)
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	shiftTypes := []string{"Morning", "Afternoon"}
	specialties := []string{"Emergency", "ICU"}

	validCount := 0
	for i := 0; i < count; i++ {
		date := startDate.AddDate(0, 0, i)
		shift := ShiftData{
			Date:                date.Format("2006-01-02"),
			ShiftType:           shiftTypes[i%len(shiftTypes)],
			RequiredStaffing:    fmt.Sprintf("%d", (i%5)+1),
			SpecialtyConstraint: specialties[i%len(specialties)],
			StudyType:           "Type-A",
		}

		// Inject specific errors
		switch i % 10 {
		case 0: // Missing RequiredStaffing
			shift.RequiredStaffing = ""
			validCount++
			shifts = append(shifts, shift)
		case 1: // Invalid ShiftType
			shift.ShiftType = "INVALID_TYPE"
			validCount++
			shifts = append(shifts, shift)
		case 2: // Non-numeric RequiredStaffing
			shift.RequiredStaffing = "twenty"
			validCount++
			shifts = append(shifts, shift)
		case 3: // Empty row - skip it to create gap
			continue
		default:
			shifts = append(shifts, shift)
		}

		if validCount >= count {
			break
		}
	}

	return shifts
}

func createODSFile(filePath string, columns []string, shifts []ShiftData) error {
	// Create a buffer for the ZIP file
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	defer zw.Close()

	// Add mimetype file (must be first, uncompressed)
	mimeTypeWriter, err := zw.Create("mimetype")
	if err != nil {
		return err
	}
	if _, err := mimeTypeWriter.Write([]byte("application/vnd.oasis.opendocument.spreadsheet")); err != nil {
		return err
	}

	// Add META-INF/manifest.xml
	manifestPath := "META-INF/manifest.xml"
	manifestFile, err := zw.Create(manifestPath)
	if err != nil {
		return err
	}
	manifestContent := getManifestXML()
	if _, err := manifestFile.Write([]byte(manifestContent)); err != nil {
		return err
	}

	// Add content.xml (main spreadsheet data)
	contentFile, err := zw.Create("content.xml")
	if err != nil {
		return err
	}
	contentXML := generateContentXML(columns, shifts)
	if _, err := contentFile.Write([]byte(contentXML)); err != nil {
		return err
	}

	// Add styles.xml (minimal)
	stylesFile, err := zw.Create("styles.xml")
	if err != nil {
		return err
	}
	if _, err := stylesFile.Write([]byte(getStylesXML())); err != nil {
		return err
	}

	// Add settings.xml (minimal)
	settingsFile, err := zw.Create("settings.xml")
	if err != nil {
		return err
	}
	if _, err := settingsFile.Write([]byte(getSettingsXML())); err != nil {
		return err
	}

	zw.Close()

	// Write to file
	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

func generateContentXML(columns []string, shifts []ShiftData) string {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
    xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0"
    xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
    xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
    xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
    xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0"
    xmlns:number="urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0"
    xmlns:svg="urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0"
    xmlns:chart="urn:oasis:names:tc:opendocument:xmlns:chart:1.0"
    xmlns:dr3d="urn:oasis:names:tc:opendocument:xmlns:dr3d:1.0"
    xmlns:math="http://www.w3.org/1998/Math/MathML"
    xmlns:form="urn:oasis:names:tc:opendocument:xmlns:form:1.0"
    xmlns:script="urn:oasis:names:tc:opendocument:xmlns:script:1.0"
    xmlns:ooo="http://openoffice.org/2004/office"
    xmlns:ooow="http://openoffice.org/2004/writer"
    xmlns:oooc="http://openoffice.org/2004/calc"
    xmlns:dom="http://www.w3.org/2001/xml-events"
    xmlns:xforms="http://www.w3.org/2002/xforms"
    xmlns:xsd="http://www.w3.org/2001/XMLSchema"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    office:version="1.2">
  <office:scripts/>
  <office:font-face-decls/>
  <office:automatic-styles/>
  <office:body>
    <office:spreadsheet>
      <table:table table:name="Sheet1">
`

	// Add header row
	xml += "        <table:table-row>\n"
	for _, col := range columns {
		xml += fmt.Sprintf(`          <table:table-cell office:value-type="string">
            <text:p>%s</text:p>
          </table:table-cell>
`, escapeXML(col))
	}
	xml += "        </table:table-row>\n"

	// Add data rows
	for _, shift := range shifts {
		xml += "        <table:table-row>\n"

		for _, col := range columns {
			var value string
			switch col {
			case "Date":
				value = shift.Date
			case "ShiftType":
				value = shift.ShiftType
			case "RequiredStaffing":
				value = shift.RequiredStaffing
			case "SpecialtyConstraint":
				value = shift.SpecialtyConstraint
			case "StudyType":
				value = shift.StudyType
			}

			if value == "" {
				// Empty cell
				xml += `          <table:table-cell/>
`
			} else {
				// Check if it's numeric
				isNumeric := isNumericValue(value)
				if isNumeric {
					xml += fmt.Sprintf(`          <table:table-cell office:value-type="float" office:value="%s">
            <text:p>%s</text:p>
          </table:table-cell>
`, escapeXML(value), escapeXML(value))
				} else {
					xml += fmt.Sprintf(`          <table:table-cell office:value-type="string">
            <text:p>%s</text:p>
          </table:table-cell>
`, escapeXML(value))
				}
			}
		}

		xml += "        </table:table-row>\n"
	}

	xml += `      </table:table>
    </office:spreadsheet>
  </office:body>
</office:document-content>`

	return xml
}

func isNumericValue(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func escapeXML(s string) string {
	buf := &bytes.Buffer{}
	for _, c := range s {
		switch c {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		case '\'':
			buf.WriteString("&apos;")
		default:
			buf.WriteRune(c)
		}
	}
	return buf.String()
}

func getManifestXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0">
  <manifest:file-entry manifest:media-type="application/vnd.oasis.opendocument.spreadsheet" manifest:full-path="/"/>
  <manifest:file-entry manifest:media-type="text/xml" manifest:full-path="content.xml"/>
  <manifest:file-entry manifest:media-type="text/xml" manifest:full-path="styles.xml"/>
  <manifest:file-entry manifest:media-type="text/xml" manifest:full-path="settings.xml"/>
</manifest:manifest>`
}

func getStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<office:document-styles xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
    xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0"
    xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
    xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
    xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
    xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0"
    xmlns:number="urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0"
    xmlns:svg="urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0"
    xmlns:chart="urn:oasis:names:tc:opendocument:xmlns:chart:1.0"
    xmlns:dr3d="urn:oasis:names:tc:opendocument:xmlns:dr3d:1.0"
    xmlns:math="http://www.w3.org/1998/Math/MathML"
    xmlns:form="urn:oasis:names:tc:opendocument:xmlns:form:1.0"
    xmlns:script="urn:oasis:names:tc:opendocument:xmlns:script:1.0"
    xmlns:ooo="http://openoffice.org/2004/office"
    xmlns:ooow="http://openoffice.org/2004/writer"
    xmlns:oooc="http://openoffice.org/2004/calc"
    xmlns:dom="http://www.w3.org/2001/xml-events"
    xmlns:xforms="http://www.w3.org/2002/xforms"
    xmlns:xsd="http://www.w3.org/2001/XMLSchema"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    office:version="1.2">
  <office:font-face-decls/>
  <office:styles/>
  <office:automatic-styles/>
  <office:master-styles/>
</office:document-styles>`
}

func getSettingsXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<office:document-settings xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
    xmlns:xlink="http://www.w3.org/1999/xlink"
    xmlns:config="urn:oasis:names:tc:opendocument:xmlns:config:1.0"
    office:version="1.2">
  <office:settings/>
</office:document-settings>`
}
