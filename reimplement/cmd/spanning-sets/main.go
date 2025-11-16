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

// Coverage data
type CoverageRecord struct {
	StudyType     string
	Hospital      string
	Modality      string
	Specialty     string
	ShiftPosition string
	DayType       string
	TimeRange     string
	IsWeekend     bool
}

// Spanning set
type SpanningSet struct {
	Name          string
	Description   string
	MemberCount   int
	Members       []string
	WeekdaySheets int
	WeekendSheets int
	HasWeekday    bool
	HasWeekend    bool
}

func main() {
	filepath := "/home/user/schedCU/cuSchedNormalized.ods"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("              SPANNING SETS GENERATOR")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Printf("Analyzing: %s\n\n", filepath)

	records, err := parseODS(filepath)
	if err != nil {
		fmt.Printf("❌ Failed to parse: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Loaded %d coverage records\n\n", len(records))

	generateSpanningSets(records)
}

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
		dayType := "Weekday"
		if isWeekend {
			dayType = "Weekend"
		}

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

			// Parse study type components
			hospital := extractHospital(studyType)
			modality := extractModality(studyType)
			specialty := extractSpecialty(studyType)

			for j := 1; j < len(row.Cells) && j < len(headers); j++ {
				cellValue := strings.ToLower(strings.TrimSpace(extractCellText(row.Cells[j])))
				if cellValue == "x" || cellValue == "yes" || cellValue == "1" {
					shiftPosition := headers[j]
					if shiftPosition == "" {
						shiftPosition = fmt.Sprintf("Column%d", j)
					}

					records = append(records, CoverageRecord{
						StudyType:     studyType,
						Hospital:      hospital,
						Modality:      modality,
						Specialty:     specialty,
						ShiftPosition: shiftPosition,
						DayType:       dayType,
						IsWeekend:     isWeekend,
					})
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

func generateSpanningSets(records []CoverageRecord) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("                 SPANNING SETS BY DIMENSION")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// 1. Spanning sets by MODALITY (all CT, all MRI, etc)
	modalitySets := buildModalitySpanningSets(records)
	printSpanningSets("BY MODALITY (All hospitals, all body parts)", modalitySets)

	// 2. Spanning sets by SPECIALTY (all Neuro, all Body, etc)
	specialtySets := buildSpecialtySpanningSets(records)
	printSpanningSets("BY SPECIALTY (All hospitals, all modalities)", specialtySets)

	// 3. Spanning sets by HOSPITAL (all Allen, all CPMC, etc)
	hospitalSets := buildHospitalSpanningSets(records)
	printSpanningSets("BY HOSPITAL (All modalities, all body parts)", hospitalSets)

	// 4. Cross-dimensional spanning sets
	crossSets := buildCrossSpanningSets(records)
	printSpanningSets("CROSS-DIMENSIONAL", crossSets)

	// 5. Summary table
	printSummaryTable(modalitySets, specialtySets, hospitalSets)
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

		if rec.IsWeekend {
			set.WeekendSheets++
			set.HasWeekend = true
		} else {
			set.WeekdaySheets++
			set.HasWeekday = true
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

		if rec.IsWeekend {
			set.WeekendSheets++
			set.HasWeekend = true
		} else {
			set.WeekdaySheets++
			set.HasWeekday = true
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

		if rec.IsWeekend {
			set.WeekendSheets++
			set.HasWeekend = true
		} else {
			set.WeekdaySheets++
			set.HasWeekday = true
		}
	}

	for _, set := range sets {
		set.MemberCount = len(set.Members)
	}

	return sets
}

func buildCrossSpanningSets(records []CoverageRecord) map[string]*SpanningSet {
	sets := make(map[string]*SpanningSet)

	// CT Neuro - all hospitals
	// MRI Body - all hospitals
	// etc.

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

		if rec.IsWeekend {
			set.WeekendSheets++
			set.HasWeekend = true
		} else {
			set.WeekdaySheets++
			set.HasWeekday = true
		}
	}

	for _, set := range sets {
		set.MemberCount = len(set.Members)
	}

	return sets
}

func printSpanningSets(title string, sets map[string]*SpanningSet) {
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println(title)
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()

	var keys []string
	for k := range sets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		set := sets[key]

		fmt.Printf("📊 %s\n", set.Name)
		fmt.Printf("   %s\n", set.Description)
		fmt.Printf("   Spans %d study types\n", set.MemberCount)
		fmt.Printf("   Weekday coverage: %s | Weekend coverage: %s\n",
			boolToStatus(set.HasWeekday),
			boolToStatus(set.HasWeekend))

		if set.MemberCount <= 10 {
			fmt.Println("   Members:")
			sort.Strings(set.Members)
			for _, m := range set.Members {
				fmt.Printf("     • %s\n", m)
			}
		} else {
			fmt.Printf("   (showing first 5 of %d members)\n", set.MemberCount)
			sort.Strings(set.Members)
			for i := 0; i < 5 && i < len(set.Members); i++ {
				fmt.Printf("     • %s\n", set.Members[i])
			}
			fmt.Printf("     ... and %d more\n", set.MemberCount-5)
		}

		fmt.Println()
	}

	fmt.Println()
}

func printSummaryTable(modalitySets, specialtySets, hospitalSets map[string]*SpanningSet) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("                  SUMMARY TABLE")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	fmt.Printf("%-30s %-10s %-15s\n", "Spanning Set", "Members", "Coverage")
	fmt.Println(strings.Repeat("─", 70))

	// Print modality sets
	var modalityKeys []string
	for k := range modalitySets {
		modalityKeys = append(modalityKeys, k)
	}
	sort.Strings(modalityKeys)

	for _, k := range modalityKeys {
		set := modalitySets[k]
		coverage := "✅ 24/7"
		if !set.HasWeekday || !set.HasWeekend {
			coverage = "⚠️  Partial"
		}
		fmt.Printf("%-30s %-10d %-15s\n", set.Name, set.MemberCount, coverage)
	}

	fmt.Println()

	// Print specialty sets
	var specialtyKeys []string
	for k := range specialtySets {
		specialtyKeys = append(specialtyKeys, k)
	}
	sort.Strings(specialtyKeys)

	for _, k := range specialtyKeys {
		set := specialtySets[k]
		coverage := "✅ 24/7"
		if !set.HasWeekday || !set.HasWeekend {
			coverage = "⚠️  Partial"
		}
		fmt.Printf("%-30s %-10d %-15s\n", set.Name, set.MemberCount, coverage)
	}

	fmt.Println()

	// Print hospital sets
	var hospitalKeys []string
	for k := range hospitalSets {
		hospitalKeys = append(hospitalKeys, k)
	}
	sort.Strings(hospitalKeys)

	for _, k := range hospitalKeys {
		set := hospitalSets[k]
		coverage := "✅ 24/7"
		if !set.HasWeekday || !set.HasWeekend {
			coverage = "⚠️  Partial"
		}
		fmt.Printf("%-30s %-10d %-15s\n", set.Name, set.MemberCount, coverage)
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

func boolToStatus(b bool) string {
	if b {
		return "✅ Yes"
	}
	return "❌ No"
}
