package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ODS XML structures
type OfficeDocument struct {
	XMLName xml.Name `xml:"document-content"`
	Body    Body     `xml:"body"`
}

type Body struct {
	Spreadsheet Spreadsheet `xml:"spreadsheet"`
}

type Spreadsheet struct {
	Tables []Table `xml:"table"`
}

type Table struct {
	Name string `xml:"name,attr"`
	Rows []Row  `xml:"table-row"`
}

type Row struct {
	Cells []Cell `xml:"table-cell"`
}

type Cell struct {
	Text []Text `xml:"p"`
}

type Text struct {
	Value string `xml:",chardata"`
}

// Coverage data structures
type CoverageData struct {
	StudyType     string
	ShiftPosition string
	DayType       string
	TimeRange     string
	IsWeekend     bool
	SheetName     string
}

type StudyTypeSummary struct {
	StudyType      string
	WeekdaySheets  []string
	WeekendSheets  []string
	HasWeekday     bool
	HasWeekend     bool
}

func main() {
	filepath := "/home/user/schedCU/cuSchedNormalized.ods"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	fmt.Println("=" + strings.Repeat("=", 69))
	fmt.Println("         SchedCU Coverage Grid Validation Tool (Go)")
	fmt.Println("=" + strings.Repeat("=", 69))
	fmt.Println()
	fmt.Printf("Opening file: %s\n", filepath)

	coverageData, err := parseODS(filepath)
	if err != nil {
		fmt.Printf("❌ Failed to parse ODS file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Parsed %d coverage assignments\n\n", len(coverageData))

	analyzeCoverage(coverageData)
}

func parseODS(filepath string) ([]CoverageData, error) {
	// Open ODS file (which is a ZIP archive)
	r, err := zip.OpenReader(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ODS file: %w", err)
	}
	defer r.Close()

	// Find content.xml
	var contentFile *zip.File
	for _, f := range r.File {
		if f.Name == "content.xml" {
			contentFile = f
			break
		}
	}

	if contentFile == nil {
		return nil, fmt.Errorf("content.xml not found in ODS file")
	}

	// Read content.xml
	rc, err := contentFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open content.xml: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read content.xml: %w", err)
	}

	// Parse XML
	var doc OfficeDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	fmt.Printf("✓ Found %d sheet(s)\n\n", len(doc.Body.Spreadsheet.Tables))

	// Extract coverage data
	var coverageData []CoverageData

	for _, table := range doc.Body.Spreadsheet.Tables {
		sheetName := table.Name
		isWeekend := strings.Contains(strings.ToLower(sheetName), "weekend")
		dayType := "Weekday"
		if isWeekend {
			dayType = "Weekend"
		}

		timeRange := extractTimeRange(sheetName)

		fmt.Printf("Processing sheet: %s\n", sheetName)

		if len(table.Rows) == 0 {
			fmt.Println("  ✓ 0 study types\n")
			continue
		}

		// Get headers from first row
		var headers []string
		if len(table.Rows) > 0 {
			for _, cell := range table.Rows[0].Cells {
				headers = append(headers, extractCellText(cell))
			}
		}

		fmt.Printf("  Shift positions: %v\n", filterNonEmpty(headers))

		// Process data rows
		studyCount := 0
		coverageCount := 0

		for i := 1; i < len(table.Rows); i++ {
			row := table.Rows[i]
			if len(row.Cells) == 0 {
				continue
			}

			// First cell is the study type
			studyType := extractCellText(row.Cells[0])
			if studyType == "" {
				continue
			}

			studyCount++

			// Check each shift position column for coverage ('x' marker)
			for j := 1; j < len(row.Cells) && j < len(headers); j++ {
				cellValue := strings.ToLower(strings.TrimSpace(extractCellText(row.Cells[j])))

				if cellValue == "x" || cellValue == "yes" || cellValue == "1" {
					coverageCount++
					shiftPosition := headers[j]
					if shiftPosition == "" {
						shiftPosition = fmt.Sprintf("Column%d", j)
					}

					coverageData = append(coverageData, CoverageData{
						StudyType:     studyType,
						ShiftPosition: shiftPosition,
						DayType:       dayType,
						TimeRange:     timeRange,
						IsWeekend:     isWeekend,
						SheetName:     sheetName,
					})
				}
			}
		}

		fmt.Printf("  ✓ %d study types, %d coverage markers\n\n", studyCount, coverageCount)
	}

	return coverageData, nil
}

func extractCellText(cell Cell) string {
	var texts []string
	for _, t := range cell.Text {
		if t.Value != "" {
			texts = append(texts, t.Value)
		}
	}
	return strings.TrimSpace(strings.Join(texts, " "))
}

func extractTimeRange(sheetName string) string {
	nameLower := strings.ToLower(sheetName)

	if strings.Contains(nameLower, "5") && strings.Contains(nameLower, "6") && strings.Contains(nameLower, "pm") {
		return "5-6 PM"
	} else if strings.Contains(nameLower, "6") && strings.Contains(nameLower, "12") && strings.Contains(nameLower, "am") {
		return "6 PM to Midnight"
	} else if strings.Contains(nameLower, "5") && strings.Contains(nameLower, "12") && strings.Contains(nameLower, "am") {
		return "5 PM to Midnight"
	} else if strings.Contains(nameLower, "10") && strings.Contains(nameLower, "midnight") {
		return "10 PM to Midnight"
	} else if strings.Contains(nameLower, "12am") && strings.Contains(nameLower, "1am") {
		return "Midnight to 1 AM"
	} else if strings.Contains(nameLower, "1") && strings.Contains(nameLower, "8") && strings.Contains(nameLower, "am") {
		return "1 AM to 8 AM (overnight)"
	}
	return "extended hours"
}

func filterNonEmpty(strs []string) []string {
	var result []string
	for _, s := range strs {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func analyzeCoverage(coverageData []CoverageData) {
	fmt.Println("=" + strings.Repeat("=", 69))
	fmt.Println("                   COVERAGE ANALYSIS")
	fmt.Println("=" + strings.Repeat("=", 69))
	fmt.Println()

	// Build summary by study type
	studyTypeSummaries := make(map[string]*StudyTypeSummary)

	for _, c := range coverageData {
		if _, exists := studyTypeSummaries[c.StudyType]; !exists {
			studyTypeSummaries[c.StudyType] = &StudyTypeSummary{
				StudyType:     c.StudyType,
				WeekdaySheets: []string{},
				WeekendSheets: []string{},
			}
		}

		summary := studyTypeSummaries[c.StudyType]

		if c.IsWeekend {
			if !contains(summary.WeekendSheets, c.SheetName) {
				summary.WeekendSheets = append(summary.WeekendSheets, c.SheetName)
			}
			summary.HasWeekend = true
		} else {
			if !contains(summary.WeekdaySheets, c.SheetName) {
				summary.WeekdaySheets = append(summary.WeekdaySheets, c.SheetName)
			}
			summary.HasWeekday = true
		}
	}

	// Print study type coverage
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("Study Type Coverage (Detailed)")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()

	// Get sorted study types
	var studyTypes []string
	for st := range studyTypeSummaries {
		studyTypes = append(studyTypes, st)
	}
	sort.Strings(studyTypes)

	for _, studyType := range studyTypes {
		summary := studyTypeSummaries[studyType]

		fmt.Printf("📋 %s\n", studyType)
		fmt.Printf("   Weekday coverage: %d time period(s)\n", len(summary.WeekdaySheets))
		fmt.Printf("   Weekend coverage: %d time period(s)\n", len(summary.WeekendSheets))

		if !summary.HasWeekday {
			fmt.Println("   ❌ WARNING: NO WEEKDAY COVERAGE")
		}
		if !summary.HasWeekend {
			fmt.Println("   ❌ WARNING: NO WEEKEND COVERAGE")
		}
		if summary.HasWeekday && summary.HasWeekend {
			fmt.Println("   ✅ Has both weekday and weekend coverage")
		}
		fmt.Println()
	}

	// Find gaps
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("Coverage Gaps")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()

	var gaps []string
	for _, summary := range studyTypeSummaries {
		if !summary.HasWeekday {
			gaps = append(gaps, fmt.Sprintf("%s - Missing WEEKDAY coverage", summary.StudyType))
		}
		if !summary.HasWeekend {
			gaps = append(gaps, fmt.Sprintf("%s - Missing WEEKEND coverage", summary.StudyType))
		}
	}

	if len(gaps) > 0 {
		fmt.Printf("❌ Found %d gap(s):\n\n", len(gaps))
		for i, gap := range gaps {
			fmt.Printf("%d. %s\n", i+1, gap)
		}
		fmt.Println()
	} else {
		fmt.Println("✅ NO GAPS FOUND - All study types have both weekday and weekend coverage\n")
	}

	// Summary
	fmt.Println("=" + strings.Repeat("=", 69))
	fmt.Println("                   VALIDATION SUMMARY")
	fmt.Println("=" + strings.Repeat("=", 69))
	fmt.Println()

	fmt.Println("📊 Statistics:")
	fmt.Printf("   Total study types: %d\n", len(studyTypeSummaries))
	fmt.Printf("   Coverage gaps: %d\n", len(gaps))
	fmt.Println()

	if len(gaps) == 0 {
		fmt.Println("✅ VALIDATION PASSED")
		fmt.Println("   All study types have coverage for both weekdays and weekends")
	} else {
		fmt.Println("❌ VALIDATION FAILED")
		fmt.Printf("   %d coverage gap(s) detected\n", len(gaps))
		fmt.Println("   Review gaps listed above")
	}
	fmt.Println()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
