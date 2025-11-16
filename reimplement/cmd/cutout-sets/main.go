package main

import (
	"archive/zip"
	"bufio"
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
	IsWeekend bool
}

// Cutout set definition
type CutoutSet struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	BaseSet     BaseSetCriteria   `json:"base_set"`
	Exclusions  []ExclusionRule   `json:"exclusions"`
	Members     []string          `json:"members,omitempty"`
	MemberCount int               `json:"member_count"`
	HasWeekday  bool              `json:"has_weekday"`
	HasWeekend  bool              `json:"has_weekend"`
	IsEmpty     bool              `json:"is_empty"`
}

// Base set criteria (what to include initially)
type BaseSetCriteria struct {
	Dimension string `json:"dimension"` // "modality", "specialty", "hospital", "all"
	Value     string `json:"value"`     // e.g., "CT", "Neuro", "CPMC", or "*" for all
}

// Exclusion rule (what to remove from base set)
type ExclusionRule struct {
	Dimension string `json:"dimension"` // "modality", "specialty", "hospital"
	Value     string `json:"value"`     // e.g., "MRI", "Body", "Allen"
}

// Cutout definitions file
type CutoutDefinitions struct {
	Sets []CutoutSet `json:"sets"`
}

const cutoutsFile = "cutout_sets.json"

func main() {
	filepath := "/home/user/schedCU/cuSchedNormalized.ods"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("              CUTOUT SETS ANALYZER")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("Define spanning sets with exclusions (e.g., 'All MSK except MRI')")
	fmt.Println()

	records, err := parseODS(filepath)
	if err != nil {
		fmt.Printf("❌ Failed to parse ODS: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Loaded %d coverage records\n\n", len(records))

	// Load existing cutout definitions
	defs, err := loadCutoutDefinitions()
	if err != nil {
		defs = &CutoutDefinitions{Sets: []CutoutSet{}}
	} else {
		fmt.Printf("✓ Loaded %d cutout set definitions\n\n", len(defs.Sets))
	}

	// Evaluate all cutouts against coverage data
	for i := range defs.Sets {
		evaluateCutout(&defs.Sets[i], records)
	}

	// Interactive menu
	interactiveMenu(defs, records)
}

func interactiveMenu(defs *CutoutDefinitions, records []CoverageRecord) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(strings.Repeat("=", 70))
		fmt.Println("MENU")
		fmt.Println(strings.Repeat("=", 70))
		fmt.Println()
		fmt.Println("1. Create new cutout set")
		fmt.Println("2. View all cutout sets")
		fmt.Println("3. View cutout set details")
		fmt.Println("4. Delete cutout set")
		fmt.Println("5. Show examples")
		fmt.Println("6. Export cutout definitions")
		fmt.Println("7. Quit")
		fmt.Println()
		fmt.Print("Choose option: ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			createCutout(defs, records, reader)
		case "2":
			viewAllCutouts(defs)
		case "3":
			viewCutoutDetails(defs, reader)
		case "4":
			deleteCutout(defs, reader)
		case "5":
			showExamples()
		case "6":
			exportCutouts(defs)
		case "7", "quit", "q", "exit":
			fmt.Println("\n👋 Goodbye!")
			return
		default:
			fmt.Println("❌ Invalid option\n")
		}
	}
}

func createCutout(defs *CutoutDefinitions, records []CoverageRecord, reader *bufio.Reader) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("CREATE CUTOUT SET")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	cutout := CutoutSet{
		Exclusions: []ExclusionRule{},
	}

	// Get name
	fmt.Print("Name (e.g., 'All MSK except MRI'): ")
	name, _ := reader.ReadString('\n')
	cutout.Name = strings.TrimSpace(name)
	if cutout.Name == "" {
		fmt.Println("❌ Name required\n")
		return
	}

	// Get description
	fmt.Print("Description (optional, press Enter to skip): ")
	desc, _ := reader.ReadString('\n')
	cutout.Description = strings.TrimSpace(desc)

	// Define base set
	fmt.Println("\n--- BASE SET (what to include) ---")
	fmt.Println("Dimension: modality, specialty, hospital, or 'all'")
	fmt.Print("Choose dimension: ")
	baseDim, _ := reader.ReadString('\n')
	cutout.BaseSet.Dimension = strings.ToLower(strings.TrimSpace(baseDim))

	if cutout.BaseSet.Dimension == "all" {
		cutout.BaseSet.Value = "*"
	} else {
		fmt.Printf("Value for %s (e.g., CT, Neuro, CPMC): ", cutout.BaseSet.Dimension)
		baseVal, _ := reader.ReadString('\n')
		cutout.BaseSet.Value = strings.TrimSpace(baseVal)
	}

	// Define exclusions
	fmt.Println("\n--- EXCLUSIONS (what to remove) ---")
	fmt.Println("Add exclusion rules (type 'done' when finished)")

	for {
		fmt.Print("\nExclusion dimension (modality/specialty/hospital) or 'done': ")
		exDim, _ := reader.ReadString('\n')
		exDim = strings.ToLower(strings.TrimSpace(exDim))

		if exDim == "done" || exDim == "" {
			break
		}

		if exDim != "modality" && exDim != "specialty" && exDim != "hospital" {
			fmt.Println("❌ Invalid dimension")
			continue
		}

		fmt.Printf("Value to exclude for %s: ", exDim)
		exVal, _ := reader.ReadString('\n')
		exVal = strings.TrimSpace(exVal)

		if exVal == "" {
			fmt.Println("❌ Value required")
			continue
		}

		cutout.Exclusions = append(cutout.Exclusions, ExclusionRule{
			Dimension: exDim,
			Value:     exVal,
		})

		fmt.Printf("✓ Added exclusion: %s != %s\n", exDim, exVal)
	}

	// Evaluate cutout
	evaluateCutout(&cutout, records)

	// Show preview
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("PREVIEW")
	fmt.Println(strings.Repeat("=", 70))
	displayCutoutDetails(&cutout)

	// Confirm
	fmt.Print("\nSave this cutout? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("❌ Cancelled\n")
		return
	}

	// Add to definitions
	defs.Sets = append(defs.Sets, cutout)

	// Save
	if err := saveCutoutDefinitions(defs); err != nil {
		fmt.Printf("⚠️  Warning: could not save: %v\n\n", err)
	} else {
		fmt.Println("✓ Cutout set saved!\n")
	}
}

func viewAllCutouts(defs *CutoutDefinitions) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("ALL CUTOUT SETS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	if len(defs.Sets) == 0 {
		fmt.Println("No cutout sets defined yet.\n")
		return
	}

	fmt.Printf("%-4s %-35s %-10s %-10s %s\n", "#", "Name", "Members", "Coverage", "Status")
	fmt.Println(strings.Repeat("─", 70))

	for i, cutout := range defs.Sets {
		coverage := "None"
		if cutout.HasWeekday && cutout.HasWeekend {
			coverage = "24/7"
		} else if cutout.HasWeekday {
			coverage = "Weekday"
		} else if cutout.HasWeekend {
			coverage = "Weekend"
		}

		status := "✓ Has members"
		if cutout.IsEmpty {
			status = "⚠️  Empty"
		}

		name := cutout.Name
		if len(name) > 33 {
			name = name[:30] + "..."
		}

		fmt.Printf("%-4d %-35s %-10d %-10s %s\n", i+1, name, cutout.MemberCount, coverage, status)
	}

	fmt.Println()
}

func viewCutoutDetails(defs *CutoutDefinitions, reader *bufio.Reader) {
	if len(defs.Sets) == 0 {
		fmt.Println("\n❌ No cutout sets defined yet.\n")
		return
	}

	fmt.Print("\nEnter cutout number to view (1-" + fmt.Sprintf("%d", len(defs.Sets)) + "): ")
	input, _ := reader.ReadString('\n')
	var idx int
	if _, err := fmt.Sscanf(strings.TrimSpace(input), "%d", &idx); err != nil {
		fmt.Println("❌ Invalid number\n")
		return
	}

	if idx < 1 || idx > len(defs.Sets) {
		fmt.Println("❌ Invalid number\n")
		return
	}

	cutout := &defs.Sets[idx-1]
	fmt.Println()
	displayCutoutDetails(cutout)
}

func displayCutoutDetails(cutout *CutoutSet) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println(cutout.Name)
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	if cutout.Description != "" {
		fmt.Printf("Description: %s\n\n", cutout.Description)
	}

	// Base set
	fmt.Println("BASE SET:")
	if cutout.BaseSet.Value == "*" {
		fmt.Printf("  • All study types\n")
	} else {
		fmt.Printf("  • %s = %s\n", cutout.BaseSet.Dimension, cutout.BaseSet.Value)
	}
	fmt.Println()

	// Exclusions
	fmt.Println("EXCLUSIONS:")
	if len(cutout.Exclusions) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, ex := range cutout.Exclusions {
			fmt.Printf("  • Exclude %s = %s\n", ex.Dimension, ex.Value)
		}
	}
	fmt.Println()

	// Results
	fmt.Printf("Members: %d\n", cutout.MemberCount)
	fmt.Printf("Coverage: Weekday=%v, Weekend=%v\n", cutout.HasWeekday, cutout.HasWeekend)
	fmt.Println()

	if cutout.IsEmpty {
		fmt.Println("⚠️  WARNING: This cutout set is EMPTY (no matching study types exist)")
		fmt.Println()
	} else {
		fmt.Println("MATCHING STUDY TYPES:")
		sort.Strings(cutout.Members)
		for _, m := range cutout.Members {
			fmt.Printf("  • %s\n", m)
		}
		fmt.Println()
	}
}

func deleteCutout(defs *CutoutDefinitions, reader *bufio.Reader) {
	if len(defs.Sets) == 0 {
		fmt.Println("\n❌ No cutout sets defined yet.\n")
		return
	}

	viewAllCutouts(defs)

	fmt.Print("Enter cutout number to delete (1-" + fmt.Sprintf("%d", len(defs.Sets)) + ") or 'cancel': ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "cancel" {
		fmt.Println("❌ Cancelled\n")
		return
	}

	var idx int
	if _, err := fmt.Sscanf(input, "%d", &idx); err != nil {
		fmt.Println("❌ Invalid number\n")
		return
	}

	if idx < 1 || idx > len(defs.Sets) {
		fmt.Println("❌ Invalid number\n")
		return
	}

	name := defs.Sets[idx-1].Name
	fmt.Printf("\nDelete '%s'? (y/n): ", name)
	confirm, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("❌ Cancelled\n")
		return
	}

	// Remove from slice
	defs.Sets = append(defs.Sets[:idx-1], defs.Sets[idx:]...)

	// Save
	if err := saveCutoutDefinitions(defs); err != nil {
		fmt.Printf("⚠️  Warning: could not save: %v\n\n", err)
	} else {
		fmt.Println("✓ Cutout set deleted!\n")
	}
}

func showExamples() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("CUTOUT SET EXAMPLES")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	examples := []struct {
		name        string
		baseDim     string
		baseVal     string
		exclusions  []string
		description string
	}{
		{
			name:        "All MSK except MRI",
			baseDim:     "specialty",
			baseVal:     "MSK",
			exclusions:  []string{"modality=MRI"},
			description: "All musculoskeletal imaging except MRI (includes CT MSK, X-Ray MSK, US MSK)",
		},
		{
			name:        "All CT except Neuro",
			baseDim:     "modality",
			baseVal:     "CT",
			exclusions:  []string{"specialty=Neuro"},
			description: "All CT scans except neuroimaging (includes CT Body, CT Chest, etc.)",
		},
		{
			name:        "All CPMC except Ultrasound",
			baseDim:     "hospital",
			baseVal:     "CPMC",
			exclusions:  []string{"modality=US"},
			description: "All imaging at CPMC except ultrasound",
		},
		{
			name:        "All imaging except Neuro and Body",
			baseDim:     "all",
			baseVal:     "*",
			exclusions:  []string{"specialty=Neuro", "specialty=Body"},
			description: "Everything except neuro and body imaging (leaves Chest, Bone, MSK, etc.)",
		},
		{
			name:        "All MRI except CPMC and Allen",
			baseDim:     "modality",
			baseVal:     "MRI",
			exclusions:  []string{"hospital=CPMC", "hospital=Allen"},
			description: "All MRI scans except at CPMC and Allen hospitals",
		},
	}

	for i, ex := range examples {
		fmt.Printf("%d. %s\n", i+1, ex.name)
		fmt.Printf("   %s\n", ex.description)
		fmt.Printf("   Base: %s = %s\n", ex.baseDim, ex.baseVal)
		fmt.Printf("   Exclusions: %s\n", strings.Join(ex.exclusions, ", "))
		fmt.Println()
	}

	fmt.Println("These examples show conceptual cutout sets. Some may be empty if those")
	fmt.Println("study type combinations don't exist in your schedule file.")
	fmt.Println()
}

func exportCutouts(defs *CutoutDefinitions) {
	filename := "cutout_sets_export.json"
	data, err := json.MarshalIndent(defs, "", "  ")
	if err != nil {
		fmt.Printf("❌ Failed to export: %v\n\n", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("❌ Failed to write file: %v\n\n", err)
		return
	}

	fmt.Printf("✓ Exported %d cutout sets to %s\n\n", len(defs.Sets), filename)
}

func evaluateCutout(cutout *CutoutSet, records []CoverageRecord) {
	// Reset state
	cutout.Members = []string{}
	cutout.MemberCount = 0
	cutout.HasWeekday = false
	cutout.HasWeekend = false

	// Build set of unique study types that match criteria
	matchingStudies := make(map[string]bool)
	weekdayStudies := make(map[string]bool)
	weekendStudies := make(map[string]bool)

	for _, rec := range records {
		// Check if matches base set
		if !matchesBaseSet(rec, cutout.BaseSet) {
			continue
		}

		// Check if excluded
		excluded := false
		for _, ex := range cutout.Exclusions {
			if matchesExclusion(rec, ex) {
				excluded = true
				break
			}
		}

		if !excluded {
			matchingStudies[rec.StudyType] = true
			if rec.IsWeekend {
				weekendStudies[rec.StudyType] = true
			} else {
				weekdayStudies[rec.StudyType] = true
			}
		}
	}

	// Populate cutout with results
	for study := range matchingStudies {
		cutout.Members = append(cutout.Members, study)
	}
	sort.Strings(cutout.Members)

	cutout.MemberCount = len(cutout.Members)
	cutout.HasWeekday = len(weekdayStudies) > 0
	cutout.HasWeekend = len(weekendStudies) > 0
	cutout.IsEmpty = cutout.MemberCount == 0
}

func matchesBaseSet(rec CoverageRecord, base BaseSetCriteria) bool {
	if base.Value == "*" {
		return true
	}

	switch base.Dimension {
	case "modality":
		return rec.Modality == base.Value
	case "specialty":
		return rec.Specialty == base.Value
	case "hospital":
		return rec.Hospital == base.Value
	default:
		return false
	}
}

func matchesExclusion(rec CoverageRecord, ex ExclusionRule) bool {
	switch ex.Dimension {
	case "modality":
		return rec.Modality == ex.Value
	case "specialty":
		return rec.Specialty == ex.Value
	case "hospital":
		return rec.Hospital == ex.Value
	default:
		return false
	}
}

func loadCutoutDefinitions() (*CutoutDefinitions, error) {
	data, err := os.ReadFile(cutoutsFile)
	if err != nil {
		return nil, err
	}

	var defs CutoutDefinitions
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, err
	}

	return &defs, nil
}

func saveCutoutDefinitions(defs *CutoutDefinitions) error {
	data, err := json.MarshalIndent(defs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cutoutsFile, data, 0644)
}

// ODS parsing (reused from other tools)
func parseODS(filepath string) ([]CoverageRecord, error) {
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

	var records []CoverageRecord

	for _, table := range doc.Body.Spreadsheet.Tables {
		sheetName := table.Name
		isWeekend := strings.Contains(strings.ToLower(sheetName), "weekend")

		if len(table.Rows) == 0 {
			continue
		}

		var headers []string
		for _, cell := range table.Rows[0].Cells {
			headers = append(headers, extractCellText(cell))
		}

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

			for j := 1; j < len(row.Cells) && j < len(headers); j++ {
				cellValue := strings.ToLower(strings.TrimSpace(extractCellText(row.Cells[j])))
				if cellValue == "x" || cellValue == "yes" || cellValue == "1" {
					records = append(records, CoverageRecord{
						StudyType: studyType,
						Hospital:  hospital,
						Modality:  modality,
						Specialty: specialty,
						IsWeekend: isWeekend,
					})
					break // Only need one record per study type per sheet
				}
			}
		}
	}

	return records, nil
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
	specialties := []string{"Neuro", "Body", "Chest", "Bone", "MSK"}
	for _, s := range specialties {
		if strings.Contains(studyType, s) {
			return s
		}
	}
	return "General"
}
