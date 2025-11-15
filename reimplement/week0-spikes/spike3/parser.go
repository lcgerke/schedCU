package main

import (
	"encoding/xml"
	"fmt"
	"log"
)

// Sheet represents a spreadsheet within an ODS document.
type Sheet struct {
	Name  string
	Cells []Cell
}

// Cell represents a single cell within a sheet.
type Cell struct {
	Address string // e.g., "A1", "B2"
	Value   string
	Type    string // "string", "float", "date", "formula", etc.
}

// ODSParser handles ODS file parsing with error collection.
type ODSParser struct {
	logger *log.Logger
}

// NewODSParser creates a new ODS parser instance.
func NewODSParser() *ODSParser {
	return &ODSParser{
		logger: log.New(log.Writer(), "[ODS] ", log.LstdFlags),
	}
}

// Parse parses an ODS file and returns sheets with data.
// Fails fast on first error.
func (p *ODSParser) Parse(odsContent []byte) ([]Sheet, error) {
	sheets, errs := p.ParseWithErrorCollection(odsContent)
	if len(errs) > 0 {
		return sheets, errs[0]
	}
	return sheets, nil
}

// ParseWithErrorCollection parses an ODS file and collects all errors.
// Returns sheets even if errors occurred.
func (p *ODSParser) ParseWithErrorCollection(odsContent []byte) ([]Sheet, []error) {
	var doc struct {
		XMLName   xml.Name `xml:"document"`
		Namespace string   `xml:"xmlns"`
		Body      struct {
			Spreadsheet struct {
				Tables []xmlTable `xml:"table"`
			} `xml:"spreadsheet"`
		} `xml:"body"`
	}

	errs := []error{}

	// Attempt XML unmarshal
	if err := xml.Unmarshal(odsContent, &doc); err != nil {
		errs = append(errs, fmt.Errorf("XML parsing failed: %w", err))
		// Continue with best-effort parsing
	}

	sheets := []Sheet{}

	// Process tables
	for tableIdx, table := range doc.Body.Spreadsheet.Tables {
		sheet, tableErrs := p.parseTable(table, tableIdx)
		errs = append(errs, tableErrs...)

		if sheet != nil {
			sheets = append(sheets, *sheet)
		}
	}

	// If no sheets extracted but document was parseable, indicate structural issue
	if len(sheets) == 0 && len(errs) == 0 {
		errs = append(errs, fmt.Errorf("ODS document contains no valid tables"))
	}

	return sheets, errs
}

// parseTable extracts data from a single table element.
func (p *ODSParser) parseTable(table xmlTable, idx int) (*Sheet, []error) {
	errs := []error{}

	sheet := &Sheet{
		Name:  table.Name,
		Cells: []Cell{},
	}

	if sheet.Name == "" {
		sheet.Name = fmt.Sprintf("Sheet%d", idx+1)
	}

	// Process rows
	for rowIdx, row := range table.Rows {
		rowErrs := p.parseRow(row, rowIdx, sheet)
		errs = append(errs, rowErrs...)
	}

	return sheet, errs
}

// parseRow extracts cells from a table row.
func (p *ODSParser) parseRow(row xmlTableRow, rowIdx int, sheet *Sheet) []error {
	errs := []error{}

	for colIdx, xmlCell := range row.Cells {
		// Generate cell address (A1, B1, etc.)
		colLetter := numToColLetter(colIdx)
		rowNum := rowIdx + 1
		address := fmt.Sprintf("%s%d", colLetter, rowNum)

		// Extract cell value and type
		cellValue := ""
		cellType := "string"

		if xmlCell.Type != "" {
			cellType = xmlCell.Type
		}

		// Extract text content
		if len(xmlCell.Paragraphs) > 0 && len(xmlCell.Paragraphs[0].Text) > 0 {
			cellValue = xmlCell.Paragraphs[0].Text
		}

		if xmlCell.NumericValue != "" {
			cellValue = xmlCell.NumericValue
		}

		if xmlCell.DateValue != "" {
			cellValue = xmlCell.DateValue
		}

		// Validate number-columns-repeated attribute if present
		if xmlCell.Repeated != "" {
			if _, err := parseIntAttribute(xmlCell.Repeated); err != nil {
				errs = append(errs, fmt.Errorf(
					"invalid repeat count at %s: %v (invalid XML attribute: %q)",
					address, err, xmlCell.Repeated,
				))
				continue
			}
		}

		cell := Cell{
			Address: address,
			Value:   cellValue,
			Type:    cellType,
		}

		sheet.Cells = append(sheet.Cells, cell)
	}

	return errs
}

// numToColLetter converts a column number to letter(s): 0→A, 1→B, 26→AA, etc.
func numToColLetter(num int) string {
	col := ""
	for num >= 0 {
		col = string(rune('A'+num%26)) + col
		num = num/26 - 1
		if num < 0 {
			break
		}
	}
	return col
}

// parseIntAttribute safely parses integer attributes.
func parseIntAttribute(s string) (int, error) {
	if s == "" {
		return 1, nil // Default to 1 if not specified
	}

	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %q", s)
	}

	if val < 1 {
		return 0, fmt.Errorf("must be >= 1, got %d", val)
	}

	return val, nil
}

// XML unmarshaling structures (internal)

type xmlTable struct {
	XMLName xml.Name     `xml:"table"`
	Name    string       `xml:"name,attr"`
	Rows    []xmlTableRow `xml:"table-row"`
}

type xmlTableRow struct {
	XMLName xml.Name   `xml:"table-row"`
	Cells   []xmlCell  `xml:"table-cell"`
}

type xmlCell struct {
	XMLName       xml.Name        `xml:"table-cell"`
	Type          string          `xml:"value-type,attr"`
	Value         string          `xml:"value,attr"`
	NumericValue  string          `xml:"value,attr"`
	DateValue     string          `xml:"date-value,attr"`
	Repeated      string          `xml:"number-columns-repeated,attr"`
	Paragraphs    []xmlParagraph  `xml:"p"`
}

type xmlParagraph struct {
	XMLName xml.Name `xml:"p"`
	Text    string   `xml:",chardata"`
}
