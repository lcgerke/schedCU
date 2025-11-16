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
	IsWeekend bool
}

// Spanning set
type SpanningSet struct {
	ID          string
	Name        string
	Description string
	Dimension   string
	MemberCount int
	Members     []string
}

// Set cover solution
type SetCover struct {
	Sets            []*SpanningSet
	TotalSets       int
	CoveredStudies  []string
	CoveragePercent float64
}

func main() {
	filepath := "/home/user/schedCU/cuSchedNormalized.ods"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("         MINIMUM SPANNING SET COVER CALCULATOR")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("Find the minimum number of spanning sets needed to cover")
	fmt.Println("all study types using greedy set cover algorithm.")
	fmt.Println()

	records, err := parseODS(filepath)
	if err != nil {
		fmt.Printf("❌ Failed to parse ODS: %v\n", err)
		os.Exit(1)
	}

	// Get all unique study types (the universe)
	universe := getUniqueStudyTypes(records)
	fmt.Printf("✓ Found %d unique study types to cover\n", len(universe))

	// Build all spanning sets
	allSets := buildAllSpanningSets(records)
	fmt.Printf("✓ Built %d spanning sets\n\n", len(allSets))

	// Run greedy set cover algorithm
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("GREEDY SET COVER ALGORITHM")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("Algorithm: Repeatedly pick the set covering the most uncovered studies")
	fmt.Println()

	greedyCover := greedySetCover(allSets, universe)

	// Display results
	displaySetCover(greedyCover)

	// Show alternatives by dimension
	fmt.Println()
	analyzeByDimension(allSets, universe)

	// Export results
	exportResults(greedyCover, allSets, universe)
}

func greedySetCover(sets []*SpanningSet, universe []string) *SetCover {
	uncovered := make(map[string]bool)
	for _, study := range universe {
		uncovered[study] = true
	}

	var chosenSets []*SpanningSet
	step := 1

	for len(uncovered) > 0 {
		// Find set that covers the most uncovered studies
		var bestSet *SpanningSet
		bestCoverage := 0

		for _, set := range sets {
			coverage := 0
			for _, member := range set.Members {
				if uncovered[member] {
					coverage++
				}
			}

			if coverage > bestCoverage {
				bestCoverage = coverage
				bestSet = set
			}
		}

		if bestSet == nil || bestCoverage == 0 {
			// No more sets can cover remaining studies
			break
		}

		// Add best set to solution
		chosenSets = append(chosenSets, bestSet)

		// Mark studies as covered
		newlyCovered := []string{}
		for _, member := range bestSet.Members {
			if uncovered[member] {
				delete(uncovered, member)
				newlyCovered = append(newlyCovered, member)
			}
		}

		// Display step
		fmt.Printf("Step %d: %s (%s)\n", step, bestSet.Name, bestSet.Dimension)
		fmt.Printf("        Covers %d new studies (total: %d/%d covered)\n",
			len(newlyCovered), len(universe)-len(uncovered), len(universe))
		if len(newlyCovered) <= 5 {
			fmt.Printf("        New: %s\n", strings.Join(newlyCovered, ", "))
		}
		fmt.Println()

		step++
	}

	// Build covered list
	covered := []string{}
	coveredMap := make(map[string]bool)
	for _, set := range chosenSets {
		for _, member := range set.Members {
			if !coveredMap[member] {
				covered = append(covered, member)
				coveredMap[member] = true
			}
		}
	}
	sort.Strings(covered)

	return &SetCover{
		Sets:            chosenSets,
		TotalSets:       len(chosenSets),
		CoveredStudies:  covered,
		CoveragePercent: float64(len(covered)) / float64(len(universe)) * 100,
	}
}

func displaySetCover(cover *SetCover) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("MINIMUM SPANNING SET COVER RESULT")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	fmt.Printf("✓ Minimum sets needed: %d\n", cover.TotalSets)
	fmt.Printf("✓ Coverage: %d studies (%.1f%%)\n\n", len(cover.CoveredStudies), cover.CoveragePercent)

	fmt.Println("CHOSEN SPANNING SETS:")
	fmt.Println(strings.Repeat("─", 70))
	for i, set := range cover.Sets {
		fmt.Printf("%d. %s (%s)\n", i+1, set.Name, set.Dimension)
		fmt.Printf("   Members: %d\n", set.MemberCount)
	}
	fmt.Println()

	fmt.Println("INTERPRETATION:")
	fmt.Println("This is the MINIMUM number of spanning sets needed to describe all")
	fmt.Println("study types. Use these sets for:")
	fmt.Println("  • Simplest possible summary of coverage")
	fmt.Println("  • Prioritizing which aggregations to show first")
	fmt.Println("  • Understanding natural groupings in your schedule")
	fmt.Println()
}

func analyzeByDimension(sets []*SpanningSet, universe []string) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("ALTERNATIVE COVERS BY DIMENSION")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	dimensions := []string{"modality", "specialty", "hospital", "cross"}

	for _, dim := range dimensions {
		dimSets := []*SpanningSet{}
		for _, set := range sets {
			if set.Dimension == dim {
				dimSets = append(dimSets, set)
			}
		}

		if len(dimSets) == 0 {
			continue
		}

		// Calculate coverage using only this dimension
		cover := greedySetCoverSilent(dimSets, universe)

		fmt.Printf("Using only %s dimension:\n", strings.ToUpper(dim))
		fmt.Printf("  Sets needed: %d\n", cover.TotalSets)
		fmt.Printf("  Coverage: %.1f%%\n", cover.CoveragePercent)
		fmt.Printf("  Sets: ")

		names := []string{}
		for _, set := range cover.Sets {
			names = append(names, set.Name)
		}
		fmt.Printf("%s\n\n", strings.Join(names, ", "))
	}

	fmt.Println("INSIGHT: Compare which dimension provides the most efficient coverage.")
	fmt.Println()
}

func greedySetCoverSilent(sets []*SpanningSet, universe []string) *SetCover {
	uncovered := make(map[string]bool)
	for _, study := range universe {
		uncovered[study] = true
	}

	var chosenSets []*SpanningSet

	for len(uncovered) > 0 {
		var bestSet *SpanningSet
		bestCoverage := 0

		for _, set := range sets {
			coverage := 0
			for _, member := range set.Members {
				if uncovered[member] {
					coverage++
				}
			}

			if coverage > bestCoverage {
				bestCoverage = coverage
				bestSet = set
			}
		}

		if bestSet == nil || bestCoverage == 0 {
			break
		}

		chosenSets = append(chosenSets, bestSet)

		for _, member := range bestSet.Members {
			delete(uncovered, member)
		}
	}

	covered := []string{}
	coveredMap := make(map[string]bool)
	for _, set := range chosenSets {
		for _, member := range set.Members {
			if !coveredMap[member] {
				covered = append(covered, member)
				coveredMap[member] = true
			}
		}
	}

	return &SetCover{
		Sets:            chosenSets,
		TotalSets:       len(chosenSets),
		CoveredStudies:  covered,
		CoveragePercent: float64(len(covered)) / float64(len(universe)) * 100,
	}
}

func exportResults(cover *SetCover, allSets []*SpanningSet, universe []string) {
	type Export struct {
		MinimumCover struct {
			SetsNeeded int      `json:"sets_needed"`
			Coverage   float64  `json:"coverage_percent"`
			Sets       []string `json:"sets"`
		} `json:"minimum_cover"`
		AllStudyTypes  []string `json:"all_study_types"`
		TotalSets      int      `json:"total_sets_available"`
		OptimalPercent float64  `json:"optimal_efficiency_percent"`
	}

	exp := Export{}
	exp.MinimumCover.SetsNeeded = cover.TotalSets
	exp.MinimumCover.Coverage = cover.CoveragePercent
	for _, set := range cover.Sets {
		exp.MinimumCover.Sets = append(exp.MinimumCover.Sets, set.Name)
	}
	exp.AllStudyTypes = universe
	exp.TotalSets = len(allSets)
	exp.OptimalPercent = float64(cover.TotalSets) / float64(len(allSets)) * 100

	data, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		fmt.Printf("⚠️  Warning: could not export results: %v\n", err)
		return
	}

	filename := "minimum_spanning_cover.json"
	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("⚠️  Warning: could not write file: %v\n", err)
		return
	}

	fmt.Printf("✓ Results exported to %s\n", filename)
	fmt.Printf("  Efficiency: Using %d/%d sets (%.1f%% of total)\n\n",
		cover.TotalSets, len(allSets), exp.OptimalPercent)
}

func getUniqueStudyTypes(records []CoverageRecord) []string {
	unique := make(map[string]bool)
	for _, rec := range records {
		unique[rec.StudyType] = true
	}

	result := []string{}
	for study := range unique {
		result = append(result, study)
	}
	sort.Strings(result)
	return result
}

func buildAllSpanningSets(records []CoverageRecord) []*SpanningSet {
	var allSets []*SpanningSet

	// Build modality sets
	modalitySets := buildModalitySpanningSets(records)
	for k, set := range modalitySets {
		set.ID = "modality_" + k
		set.Dimension = "modality"
		allSets = append(allSets, set)
	}

	// Build specialty sets
	specialtySets := buildSpecialtySpanningSets(records)
	for k, set := range specialtySets {
		set.ID = "specialty_" + k
		set.Dimension = "specialty"
		allSets = append(allSets, set)
	}

	// Build hospital sets
	hospitalSets := buildHospitalSpanningSets(records)
	for k, set := range hospitalSets {
		set.ID = "hospital_" + k
		set.Dimension = "hospital"
		allSets = append(allSets, set)
	}

	// Build cross-dimensional sets
	crossSets := buildCrossSpanningSets(records)
	for k, set := range crossSets {
		set.ID = "cross_" + k
		set.Dimension = "cross"
		allSets = append(allSets, set)
	}

	return allSets
}

func buildModalitySpanningSets(records []CoverageRecord) map[string]*SpanningSet {
	sets := make(map[string]*SpanningSet)

	for _, rec := range records {
		key := rec.Modality
		if _, exists := sets[key]; !exists {
			sets[key] = &SpanningSet{
				Name:        fmt.Sprintf("All %s", key),
				Description: fmt.Sprintf("All %s scans across all hospitals and body parts", key),
				Members:     []string{},
			}
		}

		set := sets[key]
		if !contains(set.Members, rec.StudyType) {
			set.Members = append(set.Members, rec.StudyType)
		}
	}

	for _, set := range sets {
		set.MemberCount = len(set.Members)
	}

	return sets
}

func buildSpecialtySpanningSets(records []CoverageRecord) map[string]*SpanningSet {
	sets := make(map[string]*SpanningSet)

	for _, rec := range records {
		key := rec.Specialty
		if _, exists := sets[key]; !exists {
			description := fmt.Sprintf("All %s imaging across all hospitals and modalities", key)
			if key == "Neuro" {
				description = "Brain and spine imaging - all modalities, all hospitals"
			} else if key == "Body" {
				description = "Body imaging (chest, abdomen, pelvis) - all modalities, all hospitals"
			}

			sets[key] = &SpanningSet{
				Name:        fmt.Sprintf("%s - All Modalities", key),
				Description: description,
				Members:     []string{},
			}
		}

		set := sets[key]
		if !contains(set.Members, rec.StudyType) {
			set.Members = append(set.Members, rec.StudyType)
		}
	}

	for _, set := range sets {
		set.MemberCount = len(set.Members)
	}

	return sets
}

func buildHospitalSpanningSets(records []CoverageRecord) map[string]*SpanningSet {
	sets := make(map[string]*SpanningSet)

	hospitalNames := map[string]string{
		"CPMC":  "California Pacific Medical Center",
		"Allen": "Allen Hospital",
		"NYPLH": "NewYork-Presbyterian Lower Manhattan",
		"CHONY": "Children's Hospital of New York",
	}

	for _, rec := range records {
		key := rec.Hospital
		if _, exists := sets[key]; !exists {
			fullName := hospitalNames[key]
			if fullName == "" {
				fullName = key
			}

			sets[key] = &SpanningSet{
				Name:        fmt.Sprintf("%s - All Services", key),
				Description: fmt.Sprintf("All imaging services at %s", fullName),
				Members:     []string{},
			}
		}

		set := sets[key]
		if !contains(set.Members, rec.StudyType) {
			set.Members = append(set.Members, rec.StudyType)
		}
	}

	for _, set := range sets {
		set.MemberCount = len(set.Members)
	}

	return sets
}

func buildCrossSpanningSets(records []CoverageRecord) map[string]*SpanningSet {
	sets := make(map[string]*SpanningSet)

	for _, rec := range records {
		if rec.Specialty == "General" || rec.Specialty == "Unknown" {
			continue
		}

		key := fmt.Sprintf("%s %s", rec.Modality, rec.Specialty)

		if _, exists := sets[key]; !exists {
			sets[key] = &SpanningSet{
				Name:        fmt.Sprintf("%s - All Hospitals", key),
				Description: fmt.Sprintf("All %s %s scans across all hospital locations", rec.Specialty, rec.Modality),
				Members:     []string{},
			}
		}

		set := sets[key]
		if !contains(set.Members, rec.StudyType) {
			set.Members = append(set.Members, rec.StudyType)
		}
	}

	for _, set := range sets {
		set.MemberCount = len(set.Members)
	}

	return sets
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
					break
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
