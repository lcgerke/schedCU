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

// Coverage data
type CoverageRecord struct {
	StudyType string
	Hospital  string
	Modality  string
	Specialty string
}

// Shift position
type ShiftPosition struct {
	Name       string
	StudyTypes []string
}

// Spanning set
type SpanningSet struct {
	Name    string
	Members []string
}

// Precise description component
type DescriptionComponent struct {
	SetName    string
	Exclusions []string
}

// Shift description
type ShiftDescription struct {
	ShiftName   string
	StudyCount  int
	Components  []DescriptionComponent
	Description string
}

func main() {
	filepath := "/home/user/schedCU/cuSchedNormalized.ods"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("         SHIFT POSITION DESCRIPTION GENERATOR")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("Generate precise, concise descriptions of shift position coverage")
	fmt.Println("using spanning sets with exact EXCEPT clauses (no vague language).")
	fmt.Println()

	records, assignments, err := parseODS(filepath)
	if err != nil {
		fmt.Printf("❌ Failed to parse ODS: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Parsed ODS file\n")

	// Build shift positions
	positions := buildShiftPositions(assignments)
	fmt.Printf("✓ Found %d shift positions\n", len(positions))

	// Build spanning sets
	spanningSets := buildAllSpanningSets(records)
	fmt.Printf("✓ Built %d spanning sets\n\n", len(spanningSets))

	// Generate descriptions
	descriptions := []ShiftDescription{}
	for _, pos := range positions {
		desc := describeShiftPosition(pos, spanningSets, records)
		descriptions = append(descriptions, desc)
	}

	// Display descriptions
	displayDescriptions(descriptions)

	// Export
	exportDescriptions(descriptions)
}

type CoverageAssignment struct {
	StudyType     string
	ShiftPosition string
}

func parseODS(filepath string) ([]CoverageRecord, []CoverageAssignment, error) {
	r, err := zip.OpenReader(filepath)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, fmt.Errorf("content.xml not found")
	}

	rc, err := contentFile.Open()
	if err != nil {
		return nil, nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, nil, err
	}

	var doc OfficeDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, nil, err
	}

	var records []CoverageRecord
	var assignments []CoverageAssignment

	for _, table := range doc.Body.Spreadsheet.Tables {
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

			studyType := extractCellText(row.Cells[0])
			if studyType == "" {
				continue
			}

			hospital := extractHospital(studyType)
			modality := extractModality(studyType)
			specialty := extractSpecialty(studyType)

			// Add to records (unique study types)
			found := false
			for _, rec := range records {
				if rec.StudyType == studyType {
					found = true
					break
				}
			}
			if !found {
				records = append(records, CoverageRecord{
					StudyType: studyType,
					Hospital:  hospital,
					Modality:  modality,
					Specialty: specialty,
				})
			}

			// Check each shift position column for assignments
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
					})
				}
			}
		}
	}

	return records, assignments, nil
}

func buildShiftPositions(assignments []CoverageAssignment) []*ShiftPosition {
	posMap := make(map[string]*ShiftPosition)

	for _, assignment := range assignments {
		pos, exists := posMap[assignment.ShiftPosition]
		if !exists {
			pos = &ShiftPosition{
				Name:       assignment.ShiftPosition,
				StudyTypes: []string{},
			}
			posMap[assignment.ShiftPosition] = pos
		}

		// Add study type if not already present
		if !contains(pos.StudyTypes, assignment.StudyType) {
			pos.StudyTypes = append(pos.StudyTypes, assignment.StudyType)
		}
	}

	// Convert to list
	var positions []*ShiftPosition
	for _, pos := range posMap {
		sort.Strings(pos.StudyTypes)
		positions = append(positions, pos)
	}

	// Sort by number of studies (descending)
	sort.Slice(positions, func(i, j int) bool {
		return len(positions[i].StudyTypes) > len(positions[j].StudyTypes)
	})

	return positions
}

func buildAllSpanningSets(records []CoverageRecord) map[string]*SpanningSet {
	sets := make(map[string]*SpanningSet)

	// Build hospital sets
	hospitalSets := make(map[string][]string)
	for _, rec := range records {
		if !contains(hospitalSets[rec.Hospital], rec.StudyType) {
			hospitalSets[rec.Hospital] = append(hospitalSets[rec.Hospital], rec.StudyType)
		}
	}

	hospitalNames := map[string]string{
		"CPMC":  "California Pacific Medical Center",
		"Allen": "Allen Hospital",
		"NYPLH": "NewYork-Presbyterian Lower Manhattan",
		"CHONY": "Children's Hospital of New York",
	}

	for hospital, studies := range hospitalSets {
		fullName := hospitalNames[hospital]
		if fullName == "" {
			fullName = hospital
		}
		sets[hospital+"-All"] = &SpanningSet{
			Name:    fmt.Sprintf("%s - All Services", hospital),
			Members: studies,
		}
	}

	// Build modality sets
	modalitySets := make(map[string][]string)
	for _, rec := range records {
		if !contains(modalitySets[rec.Modality], rec.StudyType) {
			modalitySets[rec.Modality] = append(modalitySets[rec.Modality], rec.StudyType)
		}
	}

	for modality, studies := range modalitySets {
		sets["All-"+modality] = &SpanningSet{
			Name:    fmt.Sprintf("All %s", modality),
			Members: studies,
		}
	}

	// Build specialty sets
	specialtySets := make(map[string][]string)
	for _, rec := range records {
		if !contains(specialtySets[rec.Specialty], rec.StudyType) {
			specialtySets[rec.Specialty] = append(specialtySets[rec.Specialty], rec.StudyType)
		}
	}

	for specialty, studies := range specialtySets {
		sets[specialty+"-AllMod"] = &SpanningSet{
			Name:    fmt.Sprintf("%s - All Modalities", specialty),
			Members: studies,
		}
	}

	return sets
}

func describeShiftPosition(pos *ShiftPosition, sets map[string]*SpanningSet, records []CoverageRecord) ShiftDescription {
	desc := ShiftDescription{
		ShiftName:  pos.Name,
		StudyCount: len(pos.StudyTypes),
		Components: []DescriptionComponent{},
	}

	// Try to find spanning sets that cover this position's studies
	// with minimum number of components and EXCEPT clauses

	// Sort sets by size (descending) to try largest first
	sortedSets := make([]*SpanningSet, 0, len(sets))
	for _, set := range sets {
		sortedSets = append(sortedSets, set)
	}
	sort.Slice(sortedSets, func(i, j int) bool {
		return len(sortedSets[i].Members) > len(sortedSets[j].Members)
	})

	covered := make(map[string]bool)
	uncovered := make(map[string]bool)
	for _, study := range pos.StudyTypes {
		uncovered[study] = true
	}

	// Greedy: pick sets that cover the most uncovered studies
	for len(uncovered) > 0 {
		bestSet := (*SpanningSet)(nil)
		bestCoverage := 0
		var bestExclusions []string

		for _, set := range sortedSets {
			// Count how many uncovered studies this set has
			coverCount := 0
			for _, member := range set.Members {
				if uncovered[member] {
					coverCount++
				}
			}

			// Count exclusions needed (studies in set but not in position)
			exclusions := []string{}
			for _, member := range set.Members {
				if !contains(pos.StudyTypes, member) {
					exclusions = append(exclusions, member)
				}
			}

			// Score: coverage minus penalty for exclusions
			score := coverCount - len(exclusions)/2

			if score > bestCoverage {
				bestCoverage = score
				bestSet = set
				bestExclusions = exclusions
			}
		}

		if bestSet == nil || bestCoverage <= 0 {
			break
		}

		// Add this set
		comp := DescriptionComponent{
			SetName:    bestSet.Name,
			Exclusions: bestExclusions,
		}
		desc.Components = append(desc.Components, comp)

		// Mark covered
		for _, member := range bestSet.Members {
			if uncovered[member] && !contains(bestExclusions, member) {
				covered[member] = true
				delete(uncovered, member)
			}
		}
	}

	// Build description string
	parts := []string{}
	for _, comp := range desc.Components {
		if len(comp.Exclusions) == 0 {
			parts = append(parts, comp.SetName)
		} else {
			exclusionStr := strings.Join(comp.Exclusions, ", ")
			parts = append(parts, fmt.Sprintf("%s EXCEPT (%s)", comp.SetName, exclusionStr))
		}
	}

	desc.Description = strings.Join(parts, " + ")

	return desc
}

func displayDescriptions(descriptions []ShiftDescription) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("PRECISE SHIFT POSITION DESCRIPTIONS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	for _, desc := range descriptions {
		fmt.Printf("%s (%d study types):\n", desc.ShiftName, desc.StudyCount)
		fmt.Printf("  %s\n", desc.Description)
		fmt.Println()

		if len(desc.Components) > 1 {
			fmt.Println("  Breakdown:")
			for i, comp := range desc.Components {
				if len(comp.Exclusions) == 0 {
					fmt.Printf("    %d. %s\n", i+1, comp.SetName)
				} else {
					fmt.Printf("    %d. %s\n", i+1, comp.SetName)
					fmt.Printf("       EXCEPT: %s\n", strings.Join(comp.Exclusions, ", "))
				}
			}
			fmt.Println()
		}
	}
}

func exportDescriptions(descriptions []ShiftDescription) {
	data, err := json.MarshalIndent(descriptions, "", "  ")
	if err != nil {
		fmt.Printf("⚠️  Warning: could not export: %v\n", err)
		return
	}

	filename := "shift_position_descriptions.json"
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("⚠️  Warning: could not write file: %v\n", err)
		return
	}

	fmt.Printf("✓ Exported to %s\n\n", filename)
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

func extractHospital(studyType string) string {
	hospitals := []string{"CPMC", "Allen", "NYPLH", "CHONY"}
	for _, h := range hospitals {
		if strings.Contains(studyType, h) {
			return h
		}
	}
	return "Unknown"
}

func extractModality(studyType string) string {
	upper := strings.ToUpper(studyType)
	modalities := []string{"CT", "MR", "MRI", "DX", "US", "NM", "PET"}
	for _, m := range modalities {
		if strings.Contains(upper, m) {
			if m == "MR" || m == "MRI" {
				return "MRI"
			}
			if m == "DX" {
				return "X-Ray"
			}
			return m
		}
	}
	return "Unknown"
}

func extractSpecialty(studyType string) string {
	specialties := []string{"Neuro", "Body", "Chest", "Bone"}
	for _, s := range specialties {
		if strings.Contains(studyType, s) {
			return s
		}
	}
	return "General"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
