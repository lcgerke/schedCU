package main

import (
	"archive/zip"
	"encoding/json"
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

// Shift position coverage
type ShiftPosition struct {
	Name           string
	StudyTypes     []string
	StudyTypeCount int
	SheetsCovered  []string
	IsWeekday      bool
	IsWeekend      bool
}

// Coverage assignment
type CoverageAssignment struct {
	StudyType     string
	ShiftPosition string
	SheetName     string
	IsWeekend     bool
}

// Set cover solution
type SetCover struct {
	Positions       []*ShiftPosition
	TotalPositions  int
	CoveredStudies  []string
	CoveragePercent float64
}

func main() {
	filepath := "/home/user/schedCU/cuSchedNormalized.ods"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("      MINIMUM SHIFT POSITION COVER CALCULATOR")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("Find the minimum number of SHIFT POSITIONS needed to cover")
	fmt.Println("all study types based on actual ODS assignments.")
	fmt.Println()

	assignments, err := parseODSAssignments(filepath)
	if err != nil {
		fmt.Printf("❌ Failed to parse ODS: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Found %d coverage assignments\n", len(assignments))

	// Build shift positions from assignments
	positions := buildShiftPositions(assignments)
	fmt.Printf("✓ Found %d unique shift positions\n", len(positions))

	// Get universe of study types
	universe := getUniqueStudyTypes(assignments)
	fmt.Printf("✓ Need to cover %d study types\n\n", len(universe))

	// Show top shift positions by coverage
	showTopPositions(positions, 10)

	// Run greedy set cover
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("GREEDY SET COVER ALGORITHM (Shift Positions)")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("Algorithm: Repeatedly pick the shift covering the most uncovered studies")
	fmt.Println()

	greedyCover := greedySetCover(positions, universe)

	// Display results
	displaySetCover(greedyCover, len(positions))

	// Export results
	exportResults(greedyCover, positions, universe)
}

func parseODSAssignments(filepath string) ([]CoverageAssignment, error) {
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

	var assignments []CoverageAssignment

	for _, table := range doc.Body.Spreadsheet.Tables {
		sheetName := table.Name
		isWeekend := strings.Contains(strings.ToLower(sheetName), "weekend")

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

					assignments = append(assignments, CoverageAssignment{
						StudyType:     studyType,
						ShiftPosition: shiftPosition,
						SheetName:     sheetName,
						IsWeekend:     isWeekend,
					})
				}
			}
		}
	}

	return assignments, nil
}

func buildShiftPositions(assignments []CoverageAssignment) []*ShiftPosition {
	posMap := make(map[string]*ShiftPosition)

	for _, assignment := range assignments {
		pos, exists := posMap[assignment.ShiftPosition]
		if !exists {
			pos = &ShiftPosition{
				Name:          assignment.ShiftPosition,
				StudyTypes:    []string{},
				SheetsCovered: []string{},
			}
			posMap[assignment.ShiftPosition] = pos
		}

		// Add study type if not already present
		if !contains(pos.StudyTypes, assignment.StudyType) {
			pos.StudyTypes = append(pos.StudyTypes, assignment.StudyType)
		}

		// Add sheet if not already present
		if !contains(pos.SheetsCovered, assignment.SheetName) {
			pos.SheetsCovered = append(pos.SheetsCovered, assignment.SheetName)
		}

		// Track weekday/weekend
		if assignment.IsWeekend {
			pos.IsWeekend = true
		} else {
			pos.IsWeekday = true
		}
	}

	// Build list and populate counts
	var positions []*ShiftPosition
	for _, pos := range posMap {
		pos.StudyTypeCount = len(pos.StudyTypes)
		sort.Strings(pos.StudyTypes)
		positions = append(positions, pos)
	}

	// Sort by coverage (descending)
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].StudyTypeCount > positions[j].StudyTypeCount
	})

	return positions
}

func getUniqueStudyTypes(assignments []CoverageAssignment) []string {
	unique := make(map[string]bool)
	for _, assignment := range assignments {
		unique[assignment.StudyType] = true
	}

	result := []string{}
	for study := range unique {
		result = append(result, study)
	}
	sort.Strings(result)
	return result
}

func showTopPositions(positions []*ShiftPosition, limit int) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("TOP SHIFT POSITIONS BY COVERAGE")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	fmt.Printf("%-20s %-10s %-15s %s\n", "Shift Position", "Studies", "Coverage", "Example Studies")
	fmt.Println(strings.Repeat("─", 70))

	for i := 0; i < limit && i < len(positions); i++ {
		pos := positions[i]
		coverage := "?"
		if pos.IsWeekday && pos.IsWeekend {
			coverage = "24/7"
		} else if pos.IsWeekday {
			coverage = "Weekday"
		} else if pos.IsWeekend {
			coverage = "Weekend"
		}

		examples := ""
		if len(pos.StudyTypes) <= 2 {
			examples = strings.Join(pos.StudyTypes, ", ")
		} else {
			examples = pos.StudyTypes[0] + ", " + pos.StudyTypes[1] + ", ..."
		}

		if len(examples) > 35 {
			examples = examples[:32] + "..."
		}

		fmt.Printf("%-20s %-10d %-15s %s\n", pos.Name, pos.StudyTypeCount, coverage, examples)
	}

	fmt.Println()
}

func greedySetCover(positions []*ShiftPosition, universe []string) *SetCover {
	uncovered := make(map[string]bool)
	for _, study := range universe {
		uncovered[study] = true
	}

	var chosenPositions []*ShiftPosition
	step := 1

	for len(uncovered) > 0 {
		// Find position that covers the most uncovered studies
		var bestPos *ShiftPosition
		bestCoverage := 0

		for _, pos := range positions {
			coverage := 0
			for _, study := range pos.StudyTypes {
				if uncovered[study] {
					coverage++
				}
			}

			if coverage > bestCoverage {
				bestCoverage = coverage
				bestPos = pos
			}
		}

		if bestPos == nil || bestCoverage == 0 {
			// No more positions can cover remaining studies
			break
		}

		// Add best position to solution
		chosenPositions = append(chosenPositions, bestPos)

		// Mark studies as covered
		newlyCovered := []string{}
		for _, study := range bestPos.StudyTypes {
			if uncovered[study] {
				delete(uncovered, study)
				newlyCovered = append(newlyCovered, study)
			}
		}

		// Display step
		fmt.Printf("Step %d: %s\n", step, bestPos.Name)
		fmt.Printf("        Covers %d new studies (total: %d/%d covered)\n",
			len(newlyCovered), len(universe)-len(uncovered), len(universe))
		if len(newlyCovered) <= 3 {
			fmt.Printf("        New: %s\n", strings.Join(newlyCovered, ", "))
		} else {
			fmt.Printf("        New: %s, ... and %d more\n",
				strings.Join(newlyCovered[:3], ", "), len(newlyCovered)-3)
		}
		fmt.Println()

		step++
	}

	// Build covered list
	covered := []string{}
	coveredMap := make(map[string]bool)
	for _, pos := range chosenPositions {
		for _, study := range pos.StudyTypes {
			if !coveredMap[study] {
				covered = append(covered, study)
				coveredMap[study] = true
			}
		}
	}
	sort.Strings(covered)

	return &SetCover{
		Positions:       chosenPositions,
		TotalPositions:  len(chosenPositions),
		CoveredStudies:  covered,
		CoveragePercent: float64(len(covered)) / float64(len(universe)) * 100,
	}
}

func displaySetCover(cover *SetCover, totalPositions int) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("MINIMUM SHIFT POSITION COVER RESULT")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	fmt.Printf("✓ Minimum shift positions needed: %d\n", cover.TotalPositions)
	fmt.Printf("✓ Coverage: %d studies (%.1f%%)\n", len(cover.CoveredStudies), cover.CoveragePercent)
	fmt.Printf("✓ Efficiency: Using %d/%d positions (%.1f%%)\n\n",
		cover.TotalPositions, totalPositions,
		float64(cover.TotalPositions)/float64(totalPositions)*100)

	fmt.Println("ESSENTIAL SHIFT POSITIONS:")
	fmt.Println(strings.Repeat("─", 70))
	for i, pos := range cover.Positions {
		fmt.Printf("%d. %s\n", i+1, pos.Name)
		fmt.Printf("   Covers: %d study types\n", pos.StudyTypeCount)

		coverage := "?"
		if pos.IsWeekday && pos.IsWeekend {
			coverage = "24/7"
		} else if pos.IsWeekday {
			coverage = "Weekday only"
		} else if pos.IsWeekend {
			coverage = "Weekend only"
		}
		fmt.Printf("   Schedule: %s\n", coverage)

		if pos.StudyTypeCount <= 5 {
			fmt.Printf("   Studies: %s\n", strings.Join(pos.StudyTypes, ", "))
		} else {
			fmt.Printf("   Studies: %s, ... and %d more\n",
				strings.Join(pos.StudyTypes[:3], ", "), pos.StudyTypeCount-3)
		}
		fmt.Println()
	}

	fmt.Println("INTERPRETATION:")
	fmt.Println("These are the MINIMUM shift positions (actual people/shifts) needed")
	fmt.Println("to provide complete coverage. Use this for:")
	fmt.Println("  • Staffing optimization - identify critical positions")
	fmt.Println("  • Cross-training priorities - these positions are essential")
	fmt.Println("  • Backup planning - ensure coverage for these key shifts")
	fmt.Println("  • Schedule simplification - focus on essential positions")
	fmt.Println()
}

func exportResults(cover *SetCover, allPositions []*ShiftPosition, universe []string) {
	type Export struct {
		MinimumCover struct {
			PositionsNeeded int      `json:"positions_needed"`
			Coverage        float64  `json:"coverage_percent"`
			Positions       []string `json:"positions"`
		} `json:"minimum_cover"`
		AllStudyTypes     []string `json:"all_study_types"`
		TotalPositions    int      `json:"total_positions_available"`
		EfficiencyPercent float64  `json:"efficiency_percent"`
	}

	exp := Export{}
	exp.MinimumCover.PositionsNeeded = cover.TotalPositions
	exp.MinimumCover.Coverage = cover.CoveragePercent
	for _, pos := range cover.Positions {
		exp.MinimumCover.Positions = append(exp.MinimumCover.Positions, pos.Name)
	}
	exp.AllStudyTypes = universe
	exp.TotalPositions = len(allPositions)
	exp.EfficiencyPercent = float64(cover.TotalPositions) / float64(len(allPositions)) * 100

	data, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		fmt.Printf("⚠️  Warning: could not export results: %v\n", err)
		return
	}

	filename := "minimum_shift_position_cover.json"
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("⚠️  Warning: could not write file: %v\n", err)
		return
	}

	fmt.Printf("✓ Results exported to %s\n\n", filename)
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
