package ods

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestFixtureMetadata verifies fixture metadata is valid
func TestFixtureMetadata(t *testing.T) {
	metaPath := "fixtures.json"
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read fixtures.json: %v", err)
	}

	var meta struct {
		Fixtures []struct {
			Name          string   `json:"name"`
			Type          string   `json:"type"`
			ShiftCount    int      `json:"shift_count"`
			Columns       []string `json:"columns"`
			ExpectedError int      `json:"expected_errors"`
		} `json:"fixtures"`
	}

	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("Failed to parse fixtures.json: %v", err)
	}

	if len(meta.Fixtures) == 0 {
		t.Fatal("No fixtures defined in fixtures.json")
	}

	for _, fixture := range meta.Fixtures {
		if fixture.Name == "" {
			t.Error("Fixture missing name")
		}
		if fixture.Type == "" {
			t.Error("Fixture missing type")
		}
		if fixture.ShiftCount <= 0 {
			t.Errorf("Fixture %s has invalid shift count: %d", fixture.Name, fixture.ShiftCount)
		}
		if len(fixture.Columns) == 0 {
			t.Errorf("Fixture %s has no columns", fixture.Name)
		}
	}

	t.Logf("Verified %d fixtures in metadata", len(meta.Fixtures))
}

// TestFixtureFilesExist verifies all fixtures are present
func TestFixtureFilesExist(t *testing.T) {
	fixtures := []string{
		"valid_schedule.ods",
		"partial_schedule.ods",
		"invalid_schedule.ods",
		"large_schedule.ods",
	}

	for _, fixture := range fixtures {
		fi, err := os.Stat(fixture)
		if err != nil {
			t.Errorf("Fixture %s not found: %v", fixture, err)
			continue
		}

		if fi.Size() == 0 {
			t.Errorf("Fixture %s is empty", fixture)
		}

		t.Logf("Fixture %s: %d bytes", fixture, fi.Size())
	}
}

// TestFixturesAreValidZIP verifies ODS files are valid ZIP archives
func TestFixturesAreValidZIP(t *testing.T) {
	fixtures := []string{
		"valid_schedule.ods",
		"partial_schedule.ods",
		"invalid_schedule.ods",
		"large_schedule.ods",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			zr, err := zip.OpenReader(fixture)
			if err != nil {
				t.Fatalf("Failed to open %s as ZIP: %v", fixture, err)
			}
			defer zr.Close()

			// Verify required files exist
			requiredFiles := map[string]bool{
				"mimetype":              false,
				"META-INF/manifest.xml": false,
				"content.xml":           false,
				"styles.xml":            false,
				"settings.xml":          false,
			}

			for _, f := range zr.File {
				if _, exists := requiredFiles[f.Name]; exists {
					requiredFiles[f.Name] = true
				}
			}

			for fileName, found := range requiredFiles {
				if !found {
					t.Errorf("Missing required file: %s", fileName)
				}
			}

			t.Logf("%s: Valid ODS structure with %d files", fixture, len(zr.File))
		})
	}
}

// TestFixtureMimetypeCorrect verifies MIME type is correct
func TestFixtureMimetypeCorrect(t *testing.T) {
	fixtures := []string{
		"valid_schedule.ods",
		"partial_schedule.ods",
		"invalid_schedule.ods",
		"large_schedule.ods",
	}

	expectedMIME := "application/vnd.oasis.opendocument.spreadsheet"

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			zr, err := zip.OpenReader(fixture)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", fixture, err)
			}
			defer zr.Close()

			var mimeFile *zip.File
			for _, f := range zr.File {
				if f.Name == "mimetype" {
					mimeFile = f
					break
				}
			}

			if mimeFile == nil {
				t.Fatal("mimetype file not found")
			}

			rc, err := mimeFile.Open()
			if err != nil {
				t.Fatalf("Failed to open mimetype: %v", err)
			}
			defer rc.Close()

			mimeBytes := make([]byte, 1024)
			n, err := rc.Read(mimeBytes)
			if err != nil && err.Error() != "EOF" {
				t.Fatalf("Failed to read mimetype: %v", err)
			}

			mime := string(mimeBytes[:n])
			if mime != expectedMIME {
				t.Errorf("Wrong MIME type. Expected %q, got %q", expectedMIME, mime)
			}

			t.Logf("MIME type correct: %s", mime)
		})
	}
}

// TestFixtureContentXMLValid verifies content.xml is valid XML
func TestFixtureContentXMLValid(t *testing.T) {
	fixtures := []string{
		"valid_schedule.ods",
		"partial_schedule.ods",
		"invalid_schedule.ods",
		"large_schedule.ods",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			zr, err := zip.OpenReader(fixture)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", fixture, err)
			}
			defer zr.Close()

			var contentFile *zip.File
			for _, f := range zr.File {
				if f.Name == "content.xml" {
					contentFile = f
					break
				}
			}

			if contentFile == nil {
				t.Fatal("content.xml not found")
			}

			// Verify it can be read
			rc, err := contentFile.Open()
			if err != nil {
				t.Fatalf("Failed to open content.xml: %v", err)
			}
			defer rc.Close()

			// Read first bytes to verify it's XML
			header := make([]byte, 5)
			n, _ := rc.Read(header)
			headerStr := string(header[:n])

			if headerStr != "<?xml" {
				t.Errorf("content.xml doesn't start with XML declaration: %q", headerStr)
			}

			t.Logf("content.xml is valid XML (starts with %q)", headerStr)
		})
	}
}

// TestFixtureSizeReasonable verifies file sizes are reasonable
func TestFixtureSizeReasonable(t *testing.T) {
	fixtures := map[string]struct {
		minSize int64
		maxSize int64
	}{
		"valid_schedule.ods":   {minSize: 1000, maxSize: 10000},    // 1-10 KB
		"partial_schedule.ods": {minSize: 1000, maxSize: 10000},    // 1-10 KB
		"invalid_schedule.ods": {minSize: 1000, maxSize: 10000},    // 1-10 KB
		"large_schedule.ods":   {minSize: 10000, maxSize: 1000000}, // 10 KB - 1 MB
	}

	for fixture, sizeRange := range fixtures {
		fi, err := os.Stat(fixture)
		if err != nil {
			t.Errorf("Cannot stat %s: %v", fixture, err)
			continue
		}

		size := fi.Size()
		if size < sizeRange.minSize || size > sizeRange.maxSize {
			t.Errorf("%s size %d bytes outside expected range [%d, %d]",
				fixture, size, sizeRange.minSize, sizeRange.maxSize)
		}

		t.Logf("%s: %d bytes (expected %d-%d)", fixture, size, sizeRange.minSize, sizeRange.maxSize)
	}
}

// TestFixtureDataCoverage verifies fixture contains expected data coverage
func TestFixtureDataCoverage(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		minShifts     int
		minColumns    int
		expectedCells int
	}{
		{"valid_schedule", "valid_schedule.ods", 150, 5, 0},     // 150 * 5 + header row
		{"partial_schedule", "partial_schedule.ods", 50, 4, 0},  // 50 * 4 + header row
		{"invalid_schedule", "invalid_schedule.ods", 30, 5, 0},  // ~30 * 5 + header row
		{"large_schedule", "large_schedule.ods", 1000, 5, 0},    // 1000+ * 5 + header row
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zr, err := zip.OpenReader(tt.fixture)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", tt.fixture, err)
			}
			defer zr.Close()

			// Find content.xml
			var contentFile *zip.File
			for _, f := range zr.File {
				if f.Name == "content.xml" {
					contentFile = f
					break
				}
			}

			if contentFile == nil {
				t.Fatal("content.xml not found")
			}

			// Verify file has reasonable size
			uncompSize := int64(contentFile.UncompressedSize)
			if uncompSize < 100 {
				t.Errorf("content.xml too small: %d bytes", uncompSize)
			}

			// Rough check: expect at least minShifts * minColumns * 10 bytes
			minExpected := int64(tt.minShifts * tt.minColumns * 10)
			if uncompSize < minExpected {
				t.Errorf("content.xml size %d seems too small for %d shifts * %d columns",
					uncompSize, tt.minShifts, tt.minColumns)
			}

			t.Logf("%s: content.xml %d bytes (expected >= %d)", tt.fixture, uncompSize, minExpected)
		})
	}
}

// TestFixtureStructureConsistency verifies all fixtures have consistent structure
func TestFixtureStructureConsistency(t *testing.T) {
	fixtures := []string{
		"valid_schedule.ods",
		"partial_schedule.ods",
		"invalid_schedule.ods",
		"large_schedule.ods",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			zr, err := zip.OpenReader(fixture)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", fixture, err)
			}
			defer zr.Close()

			// All must have exactly 5 files
			if len(zr.File) != 5 {
				t.Errorf("Expected 5 files in archive, got %d", len(zr.File))
			}

			// Check file presence
			hasContent := false
			hasStyles := false
			hasSettings := false
			hasManifest := false
			hasMimetype := false

			for _, f := range zr.File {
				switch f.Name {
				case "content.xml":
					hasContent = true
				case "styles.xml":
					hasStyles = true
				case "settings.xml":
					hasSettings = true
				case "META-INF/manifest.xml":
					hasManifest = true
				case "mimetype":
					hasMimetype = true
				}
			}

			if !hasContent || !hasStyles || !hasSettings || !hasManifest || !hasMimetype {
				t.Error("Missing required files")
			}

			t.Logf("%s: Structure consistent", fixture)
		})
	}
}

// TestReadmeExists verifies documentation is present
func TestReadmeExists(t *testing.T) {
	readmePath := filepath.Join(".", "README.md")
	fi, err := os.Stat(readmePath)
	if err != nil {
		t.Fatalf("README.md not found: %v", err)
	}

	if fi.Size() == 0 {
		t.Fatal("README.md is empty")
	}

	t.Logf("README.md present: %d bytes", fi.Size())
}
