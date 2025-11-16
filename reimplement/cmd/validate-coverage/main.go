package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Shift represents a parsed shift from the ODS file
type Shift struct {
	Date                string
	ShiftType           string
	Position            string
	Location            string
	StaffMember         string
	SpecialtyConstraint string
	StudyType           string
	RequiredQualification string
	IsWeekend           bool
	DayOfWeek           time.Weekday
}

// CoverageGap represents a gap in coverage
type CoverageGap struct {
	StudyType string
	DayType   string // "Weekday" or "Weekend"
	ShiftType string
	Count     int
	MissingDays []string
}

func main() {
	fmt.Println("=== SchedCU Coverage Validation Tool ===\n")

	// Open the ODS file
	filePath := "/home/user/schedCU/cuSchedNormalized.ods"
	fmt.Printf("Opening file: %s\n", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("File does not exist: %s", filePath)
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		// Try with test fixture instead
		filePath = "/home/user/schedCU/reimplement/tests/fixtures/ods/valid_schedule.ods"
		fmt.Printf("Failed to open main file, trying test fixture: %s\n", filePath)
		f, err = excelize.OpenFile(filePath)
		if err != nil {
			log.Fatalf("Failed to open ODS file: %v", err)
		}
	}
	defer f.Close()

	sheets := f.GetSheetList()
	fmt.Printf("✓ Found %d sheet(s)\n\n", len(sheets))

	if len(sheets) == 0 {
		log.Fatal("No sheets found in ODS file")
	}

	// Parse all shifts
	shifts, err := parseShifts(f, sheets[0])
	if err != nil {
		log.Fatalf("Failed to parse shifts: %v", err)
	}

	fmt.Printf("✓ Parsed %d shifts\n\n", len(shifts))

	// Analyze coverage
	analyzeCoverage(shifts)
}

func parseShifts(f *excelize.File, sheetName string) ([]Shift, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows found in sheet")
	}

	// Parse header to find column indices
	headers := rows[0]
	colMap := make(map[string]int)
	for i, header := range headers {
		colMap[strings.TrimSpace(strings.ToLower(header))] = i
	}

	fmt.Printf("Column mapping:\n")
	for key, idx := range colMap {
		fmt.Printf("  %s -> column %d\n", key, idx)
	}
	fmt.Println()

	// Parse data rows
	var shifts []Shift
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		if len(row) == 0 {
			continue
		}

		shift := Shift{}

		// Extract data from columns
		if idx, ok := colMap["date"]; ok && idx < len(row) {
			shift.Date = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["shift"]; ok && idx < len(row) {
			shift.ShiftType = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["position"]; ok && idx < len(row) {
			shift.Position = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["location"]; ok && idx < len(row) {
			shift.Location = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["staff member"]; ok && idx < len(row) {
			shift.StaffMember = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["specialty constraint"]; ok && idx < len(row) {
			shift.SpecialtyConstraint = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["study type"]; ok && idx < len(row) {
			shift.StudyType = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["required qualification"]; ok && idx < len(row) {
			shift.RequiredQualification = strings.TrimSpace(row[idx])
		}

		// Parse date to determine weekday vs weekend
		if shift.Date != "" {
			// Try multiple date formats
			dateFormats := []string{
				"2006-01-02",
				"01/02/2006",
				"1/2/2006",
				"2006/01/02",
			}

			var parsedDate time.Time
			var parseErr error
			for _, format := range dateFormats {
				parsedDate, parseErr = time.Parse(format, shift.Date)
				if parseErr == nil {
					break
				}
			}

			if parseErr == nil {
				shift.DayOfWeek = parsedDate.Weekday()
				shift.IsWeekend = (parsedDate.Weekday() == time.Saturday || parsedDate.Weekday() == time.Sunday)
			}
		}

		// Only include shifts with valid data
		if shift.Date != "" {
			shifts = append(shifts, shift)
		}
	}

	return shifts, nil
}

func analyzeCoverage(shifts []Shift) {
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("                    COVERAGE ANALYSIS                          ")
	fmt.Println("════════════════════════════════════════════════════════════════\n")

	// Count total shifts
	weekdayCount := 0
	weekendCount := 0
	for _, shift := range shifts {
		if shift.IsWeekend {
			weekendCount++
		} else {
			weekdayCount++
		}
	}

	fmt.Printf("📊 Total Shifts: %d\n", len(shifts))
	fmt.Printf("   Weekday: %d\n", weekdayCount)
	fmt.Printf("   Weekend: %d\n\n", weekendCount)

	// Analyze by study type
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Println("Study Type Coverage")
	fmt.Println("────────────────────────────────────────────────────────────────\n")

	analyzeByStudyType(shifts)

	// Analyze by shift type
	fmt.Println("\n────────────────────────────────────────────────────────────────")
	fmt.Println("Shift Type Coverage")
	fmt.Println("────────────────────────────────────────────────────────────────\n")

	analyzeByShiftType(shifts)

	// Analyze gaps
	fmt.Println("\n────────────────────────────────────────────────────────────────")
	fmt.Println("Coverage Gaps Analysis")
	fmt.Println("────────────────────────────────────────────────────────────────\n")

	findCoverageGaps(shifts)
}

func analyzeByStudyType(shifts []Shift) {
	// Group by study type
	studyTypeWeekday := make(map[string]int)
	studyTypeWeekend := make(map[string]int)
	studyTypeDates := make(map[string]map[string]bool) // studyType -> dates

	for _, shift := range shifts {
		studyType := shift.StudyType
		if studyType == "" {
			studyType = "UNSPECIFIED"
		}

		if shift.IsWeekend {
			studyTypeWeekend[studyType]++
		} else {
			studyTypeWeekday[studyType]++
		}

		// Track unique dates per study type
		if studyTypeDates[studyType] == nil {
			studyTypeDates[studyType] = make(map[string]bool)
		}
		studyTypeDates[studyType][shift.Date] = true
	}

	// Get sorted study types
	var studyTypes []string
	for st := range studyTypeWeekday {
		studyTypes = append(studyTypes, st)
	}
	for st := range studyTypeWeekend {
		found := false
		for _, existing := range studyTypes {
			if existing == st {
				found = true
				break
			}
		}
		if !found {
			studyTypes = append(studyTypes, st)
		}
	}
	sort.Strings(studyTypes)

	// Print coverage by study type
	for _, studyType := range studyTypes {
		weekday := studyTypeWeekday[studyType]
		weekend := studyTypeWeekend[studyType]
		totalDays := len(studyTypeDates[studyType])

		fmt.Printf("📋 %s\n", studyType)
		fmt.Printf("   Weekday shifts: %d\n", weekday)
		fmt.Printf("   Weekend shifts: %d\n", weekend)
		fmt.Printf("   Total unique dates: %d\n", totalDays)

		if weekday == 0 {
			fmt.Printf("   ❌ WARNING: NO WEEKDAY COVERAGE\n")
		}
		if weekend == 0 {
			fmt.Printf("   ❌ WARNING: NO WEEKEND COVERAGE\n")
		}
		fmt.Println()
	}
}

func analyzeByShiftType(shifts []Shift) {
	// Group by shift type
	shiftTypeWeekday := make(map[string]int)
	shiftTypeWeekend := make(map[string]int)

	for _, shift := range shifts {
		shiftType := shift.ShiftType
		if shiftType == "" {
			shiftType = "UNSPECIFIED"
		}

		if shift.IsWeekend {
			shiftTypeWeekend[shiftType]++
		} else {
			shiftTypeWeekday[shiftType]++
		}
	}

	// Get sorted shift types
	var shiftTypes []string
	for st := range shiftTypeWeekday {
		shiftTypes = append(shiftTypes, st)
	}
	for st := range shiftTypeWeekend {
		found := false
		for _, existing := range shiftTypes {
			if existing == st {
				found = true
				break
			}
		}
		if !found {
			shiftTypes = append(shiftTypes, st)
		}
	}
	sort.Strings(shiftTypes)

	// Print coverage by shift type
	for _, shiftType := range shiftTypes {
		weekday := shiftTypeWeekday[shiftType]
		weekend := shiftTypeWeekend[shiftType]

		fmt.Printf("🕐 %s\n", shiftType)
		fmt.Printf("   Weekday: %d shifts\n", weekday)
		fmt.Printf("   Weekend: %d shifts\n", weekend)

		if weekday == 0 {
			fmt.Printf("   ⚠️  No weekday coverage\n")
		}
		if weekend == 0 {
			fmt.Printf("   ⚠️  No weekend coverage\n")
		}
		fmt.Println()
	}
}

func findCoverageGaps(shifts []Shift) {
	// Track unique dates
	uniqueDates := make(map[string]bool)
	weekdayDates := make(map[string]bool)
	weekendDates := make(map[string]bool)

	for _, shift := range shifts {
		if shift.Date != "" {
			uniqueDates[shift.Date] = true
			if shift.IsWeekend {
				weekendDates[shift.Date] = true
			} else {
				weekdayDates[shift.Date] = true
			}
		}
	}

	fmt.Printf("📅 Date Range Summary:\n")
	fmt.Printf("   Total unique dates: %d\n", len(uniqueDates))
	fmt.Printf("   Weekday dates: %d\n", len(weekdayDates))
	fmt.Printf("   Weekend dates: %d\n\n", len(weekendDates))

	// Check for study type + day type combinations
	studyTypeByDay := make(map[string]map[string][]string) // studyType -> dayType -> dates

	for _, shift := range shifts {
		studyType := shift.StudyType
		if studyType == "" {
			studyType = "UNSPECIFIED"
		}

		dayType := "Weekday"
		if shift.IsWeekend {
			dayType = "Weekend"
		}

		if studyTypeByDay[studyType] == nil {
			studyTypeByDay[studyType] = make(map[string][]string)
		}

		// Track dates for this combination
		found := false
		for _, date := range studyTypeByDay[studyType][dayType] {
			if date == shift.Date {
				found = true
				break
			}
		}
		if !found {
			studyTypeByDay[studyType][dayType] = append(studyTypeByDay[studyType][dayType], shift.Date)
		}
	}

	// Find gaps
	gaps := []CoverageGap{}

	for studyType, dayTypes := range studyTypeByDay {
		weekdayDates := dayTypes["Weekday"]
		weekendDates := dayTypes["Weekend"]

		if len(weekdayDates) == 0 {
			gaps = append(gaps, CoverageGap{
				StudyType: studyType,
				DayType:   "Weekday",
				Count:     0,
			})
		}

		if len(weekendDates) == 0 {
			gaps = append(gaps, CoverageGap{
				StudyType: studyType,
				DayType:   "Weekend",
				Count:     0,
			})
		}
	}

	// Report gaps
	if len(gaps) > 0 {
		fmt.Printf("❌ Found %d coverage gaps:\n\n", len(gaps))
		for i, gap := range gaps {
			fmt.Printf("%d. Study Type: %s\n", i+1, gap.StudyType)
			fmt.Printf("   Day Type: %s\n", gap.DayType)
			fmt.Printf("   Status: NO COVERAGE\n\n")
		}
	} else {
		fmt.Printf("✅ NO GAPS FOUND - All study types have both weekday and weekend coverage\n\n")
	}

	// Summary
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("                    VALIDATION SUMMARY                         ")
	fmt.Println("════════════════════════════════════════════════════════════════\n")

	if len(gaps) == 0 {
		fmt.Println("✅ VALIDATION PASSED")
		fmt.Println("   All study types have coverage for both weekdays and weekends")
	} else {
		fmt.Println("❌ VALIDATION FAILED")
		fmt.Printf("   %d study type(s) missing coverage\n", len(gaps))
		fmt.Println("   Review gaps listed above")
	}
	fmt.Println()
}
