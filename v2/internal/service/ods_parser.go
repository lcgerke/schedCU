package service

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/schedcu/v2/internal/validation"
)

// odsParser is the concrete implementation of ODSParser
type odsParser struct {
	sheetNameParser *SheetNameParser
}

// NewODSParser creates a new ODS parser
func NewODSParser() ODSParser {
	return &odsParser{
		sheetNameParser: NewSheetNameParser(),
	}
}

// ODSData represents the parsed ODS file structure
type ODSData struct {
	FileName string
	Sheets   []ODSSheet
}

// ODSSheet represents a single sheet in the ODS file
type ODSSheet struct {
	Name               string
	ShiftCategory      string // MID or ON
	DayType            string // WEEKDAY or WEEKEND
	SpecialtyScenario  string // BODY or NEURO
	TimeStart          time.Time
	TimeEnd            time.Time
	CoverageGrid       []CoverageCell
}

// CoverageCell represents a single assignment cell in the coverage grid
type CoverageCell struct {
	Hospital    string // e.g., "CPMC"
	StudyType   string // e.g., "CT Neuro", "DX Bone"
	ShiftType   string // e.g., "Mid Body", "ON1"
	Assignment  string // "x" or "X" marking the assignment
	Row         int
	Column      int
}

// ParseFile parses an ODS file and returns the extracted data
func (p *odsParser) ParseFile(filePath string, result *validation.Result) (*ODSData, error) {
	// Open ZIP file (ODS is a ZIP archive)
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		result.AddError("FILE_READ_ERROR", fmt.Sprintf("Failed to open ODS file: %v", err))
		return nil, err
	}
	defer zipReader.Close()

	// Extract content.xml from ZIP
	var contentXML *zip.File
	for _, file := range zipReader.File {
		if file.Name == "content.xml" {
			contentXML = file
			break
		}
	}

	if contentXML == nil {
		result.AddError("MISSING_CONTENT_XML", "content.xml not found in ODS file")
		return nil, fmt.Errorf("content.xml not found")
	}

	// Parse XML with proper namespace handling
	rc, err := contentXML.Open()
	if err != nil {
		result.AddError("XML_READ_ERROR", fmt.Sprintf("Failed to read content.xml: %v", err))
		return nil, err
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var currentTable *tableElement
	var tables []*tableElement

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch elem := token.(type) {
		case xml.StartElement:
			if elem.Name.Local == "table" && elem.Name.Space == "urn:oasis:names:tc:opendocument:xmlns:table:1.0" {
				currentTable = &tableElement{Rows: [][]string{}}
				for _, attr := range elem.Attr {
					if attr.Name.Local == "name" {
						currentTable.Name = attr.Value
					}
				}
			} else if elem.Name.Local == "table-row" && currentTable != nil {
				currentTable.CurrentRow = []string{}
			} else if elem.Name.Local == "table-cell" && currentTable != nil && len(currentTable.CurrentRow) >= 0 {
				// Handle repeated cells
				repeat := 1
				for _, attr := range elem.Attr {
					if attr.Name.Local == "number-columns-repeated" {
						fmt.Sscanf(attr.Value, "%d", &repeat)
					}
				}
				// We'll capture cell content in CharData
				currentTable.CellRepeat = repeat
			}

		case xml.CharData:
			if currentTable != nil && currentTable.CellRepeat > 0 {
				text := strings.TrimSpace(string(elem))
				for i := 0; i < currentTable.CellRepeat; i++ {
					currentTable.CurrentRow = append(currentTable.CurrentRow, text)
				}
				currentTable.CellRepeat = 0
			}

		case xml.EndElement:
			if elem.Name.Local == "table-row" && currentTable != nil && len(currentTable.CurrentRow) > 0 {
				currentTable.Rows = append(currentTable.Rows, currentTable.CurrentRow)
				currentTable.CurrentRow = []string{}
			} else if elem.Name.Local == "table" && currentTable != nil {
				if currentTable.Name != "" && !strings.HasPrefix(currentTable.Name, "_") {
					tables = append(tables, currentTable)
				}
				currentTable = nil
			}
		}
	}

	// Process sheets
	odsData := &ODSData{
		FileName: filePath,
		Sheets:   []ODSSheet{},
	}

	for _, tbl := range tables {
		sheet := p.parseSheetFromElement(tbl, result)
		if sheet != nil {
			odsData.Sheets = append(odsData.Sheets, *sheet)
		}
	}

	if len(odsData.Sheets) == 0 {
		result.AddError("NO_VALID_SHEETS", "No valid sheets found in ODS file")
		return nil, fmt.Errorf("no valid sheets")
	}

	return odsData, nil
}

// tableElement represents a parsed table from XML
type tableElement struct {
	Name       string
	Rows       [][]string
	CurrentRow []string
	CellRepeat int
}

// parseSheetFromElement parses a sheet from the table element
func (p *odsParser) parseSheetFromElement(table *tableElement, result *validation.Result) *ODSSheet {
	sheet := &ODSSheet{
		Name: table.Name,
	}

	// Parse sheet name to extract metadata
	parsed := p.sheetNameParser.Parse(sheet.Name)
	if parsed == nil {
		result.AddWarning("INVALID_SHEET_NAME", fmt.Sprintf("Cannot parse sheet name: %s", sheet.Name))
		return nil
	}

	sheet.ShiftCategory = parsed.ShiftCategory
	sheet.DayType = parsed.DayType
	sheet.SpecialtyScenario = parsed.SpecialtyScenario
	sheet.TimeStart = parsed.TimeStart
	sheet.TimeEnd = parsed.TimeEnd

	// Extract coverage grid
	sheet.CoverageGrid = p.extractCoverageGridFromTable(table.Rows, result)

	if len(sheet.CoverageGrid) == 0 {
		result.AddWarning("EMPTY_COVERAGE_GRID", fmt.Sprintf("No coverage data in sheet: %s", sheet.Name))
		return nil
	}

	return sheet
}

// extractCoverageGridFromTable parses coverage assignments from the cell grid
func (p *odsParser) extractCoverageGridFromTable(grid [][]string, result *validation.Result) []CoverageCell {
	var coverageGrid []CoverageCell

	// Find header row (first row with shift type names)
	headerRowIdx := p.findHeaderRow(grid)
	if headerRowIdx == -1 {
		result.AddWarning("NO_HEADER_ROW", "Cannot find header row with shift types")
		return coverageGrid
	}

	// Extract shift column positions
	shiftCols := p.extractShiftColumns(grid[headerRowIdx])
	if len(shiftCols) == 0 {
		result.AddWarning("NO_SHIFT_COLUMNS", "No shift columns found")
		return coverageGrid
	}

	// Process data rows
	for rowIdx := headerRowIdx + 1; rowIdx < len(grid); rowIdx++ {
		row := grid[rowIdx]

		// Get hospital + study type from first column
		if len(row) == 0 || strings.TrimSpace(row[0]) == "" {
			continue
		}

		combinedCell := strings.TrimSpace(row[0])
		parts := strings.SplitN(combinedCell, " ", 2)

		hospital := parts[0]
		studyType := ""
		if len(parts) > 1 {
			studyType = parts[1]
		}

		// Process each shift column
		for _, shiftCol := range shiftCols {
			if shiftCol.ColumnIndex < len(row) {
				assignment := strings.TrimSpace(row[shiftCol.ColumnIndex])
				if assignment != "" {
					coverageGrid = append(coverageGrid, CoverageCell{
						Hospital:   hospital,
						StudyType:  studyType,
						ShiftType:  shiftCol.ShiftType,
						Assignment: assignment,
						Row:        rowIdx,
						Column:     shiftCol.ColumnIndex,
					})
				}
			}
		}
	}

	return coverageGrid
}

// findHeaderRow finds the row containing shift type names
func (p *odsParser) findHeaderRow(grid [][]string) int {
	shiftTypePatterns := []string{"Mid", "ON", "Day", "Night", "MidC", "MidL", "ON1", "ON2"}

	for i, row := range grid {
		if len(row) < 2 {
			continue
		}

		// Check if row contains shift type keywords
		for _, cell := range row {
			cell := strings.ToUpper(cell)
			for _, pattern := range shiftTypePatterns {
				if strings.Contains(cell, strings.ToUpper(pattern)) {
					return i
				}
			}
		}
	}

	return -1
}

// ShiftColumn represents a column containing assignments for a specific shift type
type ShiftColumn struct {
	ShiftType   string
	ColumnIndex int
}

// extractShiftColumns identifies columns with shift type headers
func (p *odsParser) extractShiftColumns(headerRow []string) []ShiftColumn {
	var columns []ShiftColumn

	for colIdx, cell := range headerRow {
		cell := strings.TrimSpace(cell)
		if cell == "" {
			continue
		}

		// Normalize shift type names
		shiftType := p.normalizeShiftType(cell)
		if shiftType != "" {
			columns = append(columns, ShiftColumn{
				ShiftType:   shiftType,
				ColumnIndex: colIdx,
			})
		}
	}

	return columns
}

// normalizeShiftType normalizes shift type names found in headers
func (p *odsParser) normalizeShiftType(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	// Handle common shift type patterns
	upper := strings.ToUpper(name)

	// Direct matches
	if upper == "ON1" || upper == "ON2" || upper == "MIDC" || upper == "MIDL" {
		return upper
	}

	// Contains patterns
	if strings.Contains(upper, "ON") && (strings.Contains(upper, "NEURO") || strings.Contains(upper, "BODY")) {
		return strings.ReplaceAll(upper, " ", "")
	}

	if strings.Contains(upper, "MID") {
		return strings.ReplaceAll(upper, " ", "")
	}

	// Default: return the name as-is if it looks like a shift type
	if strings.Contains(upper, "BODY") || strings.Contains(upper, "NEURO") || strings.Contains(upper, "DAY") {
		return name
	}

	return ""
}

// SheetNameParser parses ODS sheet names to extract metadata
type SheetNameParser struct {
	sheetPattern *regexp.Regexp
	timePattern  *regexp.Regexp
}

// NewSheetNameParser creates a new sheet name parser
func NewSheetNameParser() *SheetNameParser {
	return &SheetNameParser{
		sheetPattern: regexp.MustCompile(`^(Mid|ON)\s+(Weekday|Weekend)\s+(Body|Neuro)(?:\s+(.+))?$`),
		timePattern:  regexp.MustCompile(`(\d{1,2})\s*(?:(am|pm))?\s*[-–]\s*(\d{1,2})\s*(am|pm)`),
	}
}

// ParsedSheetName contains extracted sheet name metadata
type ParsedSheetName struct {
	ShiftCategory     string
	DayType           string
	SpecialtyScenario string
	TimeStart         time.Time
	TimeEnd           time.Time
}

// Parse extracts metadata from a sheet name
func (p *SheetNameParser) Parse(sheetName string) *ParsedSheetName {
	if sheetName == "" {
		return nil
	}

	matches := p.sheetPattern.FindStringSubmatch(sheetName)
	if matches == nil {
		return nil
	}

	result := &ParsedSheetName{
		ShiftCategory:     strings.ToUpper(matches[1]),
		DayType:           strings.ToUpper(matches[2]),
		SpecialtyScenario: strings.ToUpper(matches[3]),
	}

	// Parse time range if present
	if len(matches) > 4 && matches[4] != "" {
		p.parseTimeRange(matches[4], result)
	}

	return result
}

// parseTimeRange parses time range string like "5 - 6 pm" or "5 pm - 6 pm"
func (p *SheetNameParser) parseTimeRange(timeStr string, result *ParsedSheetName) {
	matches := p.timePattern.FindStringSubmatch(timeStr)
	if matches == nil {
		return
	}

	// Parse start time hour
	startHour := 0
	fmt.Sscanf(matches[1], "%d", &startHour)

	// Determine start AMPM - if not explicitly set in regex, use end AMPM
	startAMPM := "am"
	if len(matches) > 2 && matches[2] != "" {
		startAMPM = strings.ToLower(matches[2])
	} else if len(matches) > 4 && matches[4] != "" {
		// If start AMPM not found, use end AMPM (e.g., "5 - 6 pm" means 5 pm - 6 pm)
		startAMPM = strings.ToLower(matches[4])
	}

	// Parse end time
	endHour := 0
	fmt.Sscanf(matches[3], "%d", &endHour)
	endAMPM := "am"
	if len(matches) > 4 && matches[4] != "" {
		endAMPM = strings.ToLower(matches[4])
	}

	// Convert to 24-hour format
	startHour = p.to24Hour(startHour, startAMPM)
	endHour = p.to24Hour(endHour, endAMPM)

	result.TimeStart = time.Date(1970, 1, 1, startHour, 0, 0, 0, time.UTC)
	result.TimeEnd = time.Date(1970, 1, 1, endHour, 0, 0, 0, time.UTC)
}

// to24Hour converts 12-hour time to 24-hour format
func (p *SheetNameParser) to24Hour(hour int, ampm string) int {
	ampm = strings.ToLower(ampm)

	if ampm == "am" {
		if hour == 12 {
			return 0 // 12 am = 00:00
		}
		return hour // 1 am = 01:00, etc.
	}

	// PM time
	if hour == 12 {
		return 12 // 12 pm = 12:00
	}
	return hour + 12 // 1 pm = 13:00, 5 pm = 17:00, etc.
}
