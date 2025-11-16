package main

import (
	"archive/zip"
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
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

// Coverage data
type CoverageData struct {
	StudyType     string
	ShiftPosition string
	DayType       string
	TimeRange     string
	StartHour     int
	EndHour       int
	Specialty     string
	IsWeekend     bool
	SheetName     string
}

func main() {
	filepath := "/home/user/schedCU/cuSchedNormalized.ods"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	fmt.Println("\n📊 Loading coverage data from:", filepath)
	coverageData, err := parseODSCoverage(filepath)
	if err != nil {
		fmt.Printf("❌ Failed to load coverage data: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Loaded %d coverage assignments\n", len(coverageData))

	// Start interactive mode
	interactiveViewer(coverageData)
}

func parseODSCoverage(filepath string) ([]CoverageData, error) {
	r, err := zip.OpenReader(filepath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var contentFile *zip.File
	for _, f := range r.File {
		if f.Name == "content.xml" {
			contentFile = f
			break
		}
	}
	if contentFile == nil {
		return nil, fmt.Errorf("content.xml not found")
	}

	rc, err := contentFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	var doc OfficeDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	var allCoverage []CoverageData

	for _, table := range doc.Body.Spreadsheet.Tables {
		sheetName := table.Name
		timeInfo := parseTimeFromSheetName(sheetName)

		if len(table.Rows) == 0 {
			continue
		}

		// Get shift position headers from first row
		var shiftPositions []string
		for _, cell := range table.Rows[0].Cells {
			shiftPositions = append(shiftPositions, extractCellText(cell))
		}

		// Parse data rows (study types)
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

			// Check each shift position column for coverage
			for j := 1; j < len(row.Cells) && j < len(shiftPositions); j++ {
				cellValue := strings.ToLower(strings.TrimSpace(extractCellText(row.Cells[j])))
				if cellValue == "x" || cellValue == "yes" || cellValue == "1" {
					shiftPosition := shiftPositions[j]
					if shiftPosition == "" {
						shiftPosition = fmt.Sprintf("Column%d", j)
					}

					allCoverage = append(allCoverage, CoverageData{
						StudyType:     studyType,
						ShiftPosition: shiftPosition,
						DayType:       timeInfo.DayType,
						TimeRange:     timeInfo.TimeRange,
						StartHour:     timeInfo.StartHour,
						EndHour:       timeInfo.EndHour,
						Specialty:     timeInfo.Specialty,
						IsWeekend:     timeInfo.IsWeekend,
						SheetName:     sheetName,
					})
				}
			}
		}
	}

	return allCoverage, nil
}

type TimeInfo struct {
	DayType   string
	TimeRange string
	StartHour int
	EndHour   int
	Specialty string
	IsWeekend bool
}

func parseTimeFromSheetName(sheetName string) TimeInfo {
	nameLower := strings.ToLower(sheetName)

	info := TimeInfo{
		StartHour: -1,
		EndHour:   -1,
	}

	// Determine day type
	info.IsWeekend = strings.Contains(nameLower, "weekend")
	if info.IsWeekend {
		info.DayType = "Weekend"
	} else {
		info.DayType = "Weekday"
	}

	// Extract time range
	if strings.Contains(nameLower, "5") && strings.Contains(nameLower, "6") && strings.Contains(nameLower, "pm") {
		info.TimeRange = "5-6 PM"
		info.StartHour = 17
		info.EndHour = 18
	} else if strings.Contains(nameLower, "6") && strings.Contains(nameLower, "12") && strings.Contains(nameLower, "am") {
		info.TimeRange = "6 PM to Midnight"
		info.StartHour = 18
		info.EndHour = 24
	} else if strings.Contains(nameLower, "5") && strings.Contains(nameLower, "12") && strings.Contains(nameLower, "am") {
		info.TimeRange = "5 PM to Midnight"
		info.StartHour = 17
		info.EndHour = 24
	} else if strings.Contains(nameLower, "10") && strings.Contains(nameLower, "midnight") {
		info.TimeRange = "10 PM to Midnight"
		info.StartHour = 22
		info.EndHour = 24
	} else if strings.Contains(nameLower, "12am") && strings.Contains(nameLower, "1am") {
		info.TimeRange = "Midnight to 1 AM"
		info.StartHour = 0
		info.EndHour = 1
	} else if strings.Contains(nameLower, "1") && strings.Contains(nameLower, "8") && strings.Contains(nameLower, "am") {
		info.TimeRange = "1 AM to 8 AM"
		info.StartHour = 1
		info.EndHour = 8
	} else {
		info.TimeRange = "Unknown"
	}

	// Determine specialty
	if strings.Contains(nameLower, "body") {
		info.Specialty = "Body"
	} else if strings.Contains(nameLower, "neuro") {
		info.Specialty = "Neuro"
	}

	return info
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

func showCoverageAtHour(coverageData []CoverageData, hour int, isWeekend bool) {
	dayType := "Weekday"
	if isWeekend {
		dayType = "Weekend"
	}

	// Filter coverage for this hour
	var relevantCoverage []CoverageData
	for _, c := range coverageData {
		if c.StartHour >= 0 && c.EndHour >= 0 &&
			c.StartHour <= hour && hour < c.EndHour &&
			c.IsWeekend == isWeekend {
			relevantCoverage = append(relevantCoverage, c)
		}
	}

	if len(relevantCoverage) == 0 {
		fmt.Printf("\n❌ No coverage data found for %s at hour %d\n", dayType, hour)
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Printf("         COVERAGE AT %s (%s)\n", formatHour(hour), dayType)
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Group by time period/sheet
	bySheet := make(map[string][]CoverageData)
	for _, c := range relevantCoverage {
		bySheet[c.SheetName] = append(bySheet[c.SheetName], c)
	}

	// Show each time period
	var sheetNames []string
	for name := range bySheet {
		sheetNames = append(sheetNames, name)
	}
	sort.Strings(sheetNames)

	for _, sheetName := range sheetNames {
		items := bySheet[sheetName]
		timeRange := items[0].TimeRange
		specialty := items[0].Specialty

		specialtyLabel := ""
		if specialty != "" {
			specialtyLabel = fmt.Sprintf(" (%s)", specialty)
		}
		fmt.Printf("📅 Time Period: %s%s\n", timeRange, specialtyLabel)
		fmt.Println(strings.Repeat("─", 70))
		fmt.Println()

		// Group by shift position
		byPosition := make(map[string][]string)
		for _, item := range items {
			byPosition[item.ShiftPosition] = append(byPosition[item.ShiftPosition], item.StudyType)
		}

		var positions []string
		for pos := range byPosition {
			positions = append(positions, pos)
		}
		sort.Strings(positions)

		for _, position := range positions {
			studies := byPosition[position]
			sort.Strings(studies)
			fmt.Printf("  👤 %s Position:\n", position)
			for _, study := range studies {
				fmt.Printf("     • %s\n", study)
			}
			fmt.Println()
		}

		fmt.Println()
	}

	// Summary statistics
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("📊 SUMMARY:")
	fmt.Println()

	uniqueStudies := make(map[string]bool)
	uniquePositions := make(map[string]bool)
	for _, c := range relevantCoverage {
		uniqueStudies[c.StudyType] = true
		uniquePositions[c.ShiftPosition] = true
	}

	fmt.Printf("  • Total unique study types covered: %d\n", len(uniqueStudies))
	fmt.Printf("  • Total shift positions staffed: %d\n", len(uniquePositions))
	fmt.Printf("  • Total coverage assignments: %d\n", len(relevantCoverage))
	fmt.Println()

	// List all unique study types
	fmt.Println("  📋 All study types covered during this hour:")
	var allStudies []string
	for study := range uniqueStudies {
		allStudies = append(allStudies, study)
	}
	sort.Strings(allStudies)
	for _, study := range allStudies {
		fmt.Printf("     • %s\n", study)
	}
	fmt.Println()
}

func formatHour(hour int) string {
	if hour == 0 {
		return "Midnight (12:00 AM)"
	} else if hour < 12 {
		return fmt.Sprintf("%d:00 AM", hour)
	} else if hour == 12 {
		return "Noon (12:00 PM)"
	} else {
		return fmt.Sprintf("%d:00 PM", hour-12)
	}
}

func listAvailableHours(coverageData []CoverageData) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("         AVAILABLE COVERAGE HOURS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Collect all hours
	weekdayHours := make(map[int]bool)
	weekendHours := make(map[int]bool)

	for _, c := range coverageData {
		if c.StartHour >= 0 && c.EndHour >= 0 {
			for h := c.StartHour; h < c.EndHour; h++ {
				if c.IsWeekend {
					weekendHours[h] = true
				} else {
					weekdayHours[h] = true
				}
			}
		}
	}

	fmt.Println("📅 WEEKDAY HOURS:")
	var weekdayHoursList []int
	for h := range weekdayHours {
		weekdayHoursList = append(weekdayHoursList, h)
	}
	sort.Ints(weekdayHoursList)
	for _, h := range weekdayHoursList {
		fmt.Printf("  • %s\n", formatHour(h))
	}
	fmt.Println()

	fmt.Println("📅 WEEKEND HOURS:")
	var weekendHoursList []int
	for h := range weekendHours {
		weekendHoursList = append(weekendHoursList, h)
	}
	sort.Ints(weekendHoursList)
	for _, h := range weekendHoursList {
		fmt.Printf("  • %s\n", formatHour(h))
	}
	fmt.Println()
}

func interactiveViewer(coverageData []CoverageData) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("         INTERACTIVE COVERAGE VIEWER")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  <hour> weekday   - Show weekday coverage at hour (0-23)")
	fmt.Println("  <hour> weekend   - Show weekend coverage at hour (0-23)")
	fmt.Println("  list             - List available hours")
	fmt.Println("  examples         - Show example queries")
	fmt.Println("  quit             - Exit")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("📍 Enter command: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nExiting...")
			break
		}

		cmd := strings.ToLower(strings.TrimSpace(input))

		if cmd == "quit" || cmd == "exit" || cmd == "q" {
			break
		}

		if cmd == "list" {
			listAvailableHours(coverageData)
			continue
		}

		if cmd == "examples" {
			fmt.Println("\nExample queries:")
			fmt.Println("  6 weekday    - Show weekday coverage at 6 PM")
			fmt.Println("  18 weekday   - Show weekday coverage at 6 PM (24-hour)")
			fmt.Println("  2 weekend    - Show weekend coverage at 2 AM")
			fmt.Println("  22 weekday   - Show weekday coverage at 10 PM")
			fmt.Println()
			continue
		}

		parts := strings.Fields(cmd)
		if len(parts) == 2 {
			hour, err := strconv.Atoi(parts[0])
			if err != nil {
				fmt.Println("❌ Invalid hour (must be a number 0-23)")
				continue
			}

			if hour < 0 || hour > 23 {
				fmt.Println("❌ Hour must be between 0-23")
				continue
			}

			dayType := parts[1]
			if dayType != "weekday" && dayType != "weekend" {
				fmt.Println("❌ Day type must be 'weekday' or 'weekend'")
				continue
			}

			isWeekend := (dayType == "weekend")
			showCoverageAtHour(coverageData, hour, isWeekend)

			// Prompt for user's summary
			fmt.Println(strings.Repeat("=", 70))
			fmt.Println("💭 HOW WOULD YOU SUMMARIZE THIS COVERAGE?")
			fmt.Println(strings.Repeat("=", 70))
			fmt.Println()
			fmt.Print("Your summary: ")

			summary, err := reader.ReadString('\n')
			if err == nil {
				summary = strings.TrimSpace(summary)
				if summary != "" {
					fmt.Printf("\n✓ Recorded: \"%s\"\n", summary)
					fmt.Println()

					// Save to examples file
					f, err := os.OpenFile("coverage_examples.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err == nil {
						f.WriteString(fmt.Sprintf("Hour: %s (%s)\n", formatHour(hour), dayType))
						f.WriteString(fmt.Sprintf("Summary: %s\n", summary))
						f.WriteString(strings.Repeat("-", 70) + "\n")
						f.Close()
						fmt.Println("✓ Saved to coverage_examples.txt")
					}
					fmt.Println()
				}
			}
		} else {
			fmt.Println("❌ Invalid command. Try 'examples' for help.")
		}
	}
}
