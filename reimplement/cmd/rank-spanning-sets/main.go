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

// ODS XML structures (reused from spanning-sets)
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
	ID            string
	Name          string
	Description   string
	Dimension     string // "modality", "specialty", "hospital", "cross"
	MemberCount   int
	Members       []string
	HasWeekday    bool
	HasWeekend    bool
	PreferenceWins   int // How many times user preferred this set
	PreferenceLosses int // How many times user didn't prefer this set
}

// Preference record
type Preference struct {
	WinnerID string `json:"winner_id"`
	LoserID  string `json:"loser_id"`
	Winner   string `json:"winner_name"`
	Loser    string `json:"loser_name"`
	Reason   string `json:"reason,omitempty"`
}

// Preferences file structure
type PreferencesData struct {
	Comparisons []Preference `json:"comparisons"`
}

const preferencesFile = "spanning_set_preferences.json"

func main() {
	filepath := "/home/user/schedCU/cuSchedNormalized.ods"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("         SPANNING SET PREFERENCE COLLECTOR")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("This tool helps you rank which spanning sets are most useful")
	fmt.Println("by comparing pairs and recording your preferences.")
	fmt.Println()

	records, err := parseODS(filepath)
	if err != nil {
		fmt.Printf("❌ Failed to parse ODS: %v\n", err)
		os.Exit(1)
	}

	// Build all spanning sets
	allSets := buildAllSpanningSets(records)
	fmt.Printf("✓ Loaded %d spanning sets\n\n", len(allSets))

	// Load existing preferences
	prefs, err := loadPreferences()
	if err != nil {
		prefs = &PreferencesData{Comparisons: []Preference{}}
	} else {
		fmt.Printf("✓ Loaded %d existing preferences\n\n", len(prefs.Comparisons))
	}

	// Apply preferences to sets to calculate scores
	applyPreferences(allSets, prefs)

	// Interactive menu
	interactiveMenu(allSets, prefs)
}

func interactiveMenu(sets []*SpanningSet, prefs *PreferencesData) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(strings.Repeat("=", 70))
		fmt.Println("MENU")
		fmt.Println(strings.Repeat("=", 70))
		fmt.Println()
		fmt.Println("1. Compare two spanning sets (record preference)")
		fmt.Println("2. View current rankings")
		fmt.Println("3. View all spanning sets")
		fmt.Println("4. Show preference statistics")
		fmt.Println("5. Export preferences to file")
		fmt.Println("6. Quit")
		fmt.Println()
		fmt.Print("Choose option: ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			compareSpanningSets(sets, prefs, reader)
		case "2":
			viewRankings(sets)
		case "3":
			viewAllSets(sets)
		case "4":
			showStatistics(sets, prefs)
		case "5":
			exportPreferences(prefs)
		case "6", "quit", "q", "exit":
			fmt.Println("\n👋 Goodbye!")
			return
		default:
			fmt.Println("❌ Invalid option\n")
		}
	}
}

func compareSpanningSets(sets []*SpanningSet, prefs *PreferencesData, reader *bufio.Reader) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("COMPARE SPANNING SETS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Pick two random sets from different dimensions for interesting comparisons
	if len(sets) < 2 {
		fmt.Println("❌ Not enough sets to compare\n")
		return
	}

	// Get two random sets
	set1 := sets[randomIndex(len(sets))]
	set2 := sets[randomIndex(len(sets))]

	// Make sure they're different
	for set1.ID == set2.ID {
		set2 = sets[randomIndex(len(sets))]
	}

	// Display the two sets
	fmt.Println("A) " + set1.Name)
	fmt.Printf("   %s\n", set1.Description)
	fmt.Printf("   Dimension: %s | Members: %d | Coverage: %s\n",
		set1.Dimension, set1.MemberCount, coverageStatus(set1))
	if set1.MemberCount <= 5 {
		for _, m := range set1.Members {
			fmt.Printf("     • %s\n", m)
		}
	}
	fmt.Println()

	fmt.Println("B) " + set2.Name)
	fmt.Printf("   %s\n", set2.Description)
	fmt.Printf("   Dimension: %s | Members: %d | Coverage: %s\n",
		set2.Dimension, set2.MemberCount, coverageStatus(set2))
	if set2.MemberCount <= 5 {
		for _, m := range set2.Members {
			fmt.Printf("     • %s\n", m)
		}
	}
	fmt.Println()

	fmt.Println(strings.Repeat("─", 70))
	fmt.Println("Which spanning set is MORE USEFUL for clinical understanding?")
	fmt.Println("(A) " + set1.Name)
	fmt.Println("(B) " + set2.Name)
	fmt.Println("(S) Skip this comparison")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Print("\nYour choice (A/B/S): ")

	input, _ := reader.ReadString('\n')
	choice := strings.ToUpper(strings.TrimSpace(input))

	if choice == "S" {
		fmt.Println("⏭️  Skipped\n")
		return
	}

	var winner, loser *SpanningSet
	if choice == "A" {
		winner = set1
		loser = set2
	} else if choice == "B" {
		winner = set2
		loser = set1
	} else {
		fmt.Println("❌ Invalid choice\n")
		return
	}

	// Optional: ask for reason
	fmt.Print("\nOptional - Why is this more useful? (press Enter to skip): ")
	reason, _ := reader.ReadString('\n')
	reason = strings.TrimSpace(reason)

	// Record preference
	pref := Preference{
		WinnerID: winner.ID,
		LoserID:  loser.ID,
		Winner:   winner.Name,
		Loser:    loser.Name,
		Reason:   reason,
	}
	prefs.Comparisons = append(prefs.Comparisons, pref)

	// Update scores
	winner.PreferenceWins++
	loser.PreferenceLosses++

	// Save to file
	if err := savePreferences(prefs); err != nil {
		fmt.Printf("⚠️  Warning: could not save preferences: %v\n", err)
	} else {
		fmt.Println("✓ Preference saved!\n")
	}
}

func viewRankings(sets []*SpanningSet) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("CURRENT RANKINGS (by preference score)")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Sort by score (wins - losses)
	sorted := make([]*SpanningSet, len(sets))
	copy(sorted, sets)
	sort.Slice(sorted, func(i, j int) bool {
		scoreI := sorted[i].PreferenceWins - sorted[i].PreferenceLosses
		scoreJ := sorted[j].PreferenceWins - sorted[j].PreferenceLosses
		if scoreI != scoreJ {
			return scoreI > scoreJ
		}
		// Tie-breaker: total comparisons (more data = higher rank)
		totalI := sorted[i].PreferenceWins + sorted[i].PreferenceLosses
		totalJ := sorted[j].PreferenceWins + sorted[j].PreferenceLosses
		return totalI > totalJ
	})

	fmt.Printf("%-4s %-40s %-8s %-10s %s\n", "Rank", "Spanning Set", "Score", "Record", "Dimension")
	fmt.Println(strings.Repeat("─", 70))

	for i, set := range sorted {
		score := set.PreferenceWins - set.PreferenceLosses
		record := fmt.Sprintf("%d-%d", set.PreferenceWins, set.PreferenceLosses)
		rank := i + 1

		// Truncate name if too long
		name := set.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}

		fmt.Printf("%-4d %-40s %-8d %-10s %s\n", rank, name, score, record, set.Dimension)
	}

	fmt.Println()
}

func viewAllSets(sets []*SpanningSet) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("ALL SPANNING SETS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Group by dimension
	byDimension := make(map[string][]*SpanningSet)
	for _, set := range sets {
		byDimension[set.Dimension] = append(byDimension[set.Dimension], set)
	}

	dimensions := []string{"modality", "specialty", "hospital", "cross"}
	for _, dim := range dimensions {
		sets := byDimension[dim]
		if len(sets) == 0 {
			continue
		}

		fmt.Printf("─── %s (%d sets) ───\n", strings.ToUpper(dim), len(sets))
		for _, set := range sets {
			fmt.Printf("  • %s (%d members)\n", set.Name, set.MemberCount)
		}
		fmt.Println()
	}
}

func showStatistics(sets []*SpanningSet, prefs *PreferencesData) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("PREFERENCE STATISTICS")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	totalComparisons := len(prefs.Comparisons)
	setsWithData := 0
	for _, set := range sets {
		if set.PreferenceWins+set.PreferenceLosses > 0 {
			setsWithData++
		}
	}

	fmt.Printf("Total comparisons recorded: %d\n", totalComparisons)
	fmt.Printf("Total spanning sets: %d\n", len(sets))
	fmt.Printf("Sets with preference data: %d (%.1f%%)\n\n",
		setsWithData, float64(setsWithData)/float64(len(sets))*100)

	// Top 5 most preferred
	sorted := make([]*SpanningSet, len(sets))
	copy(sorted, sets)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PreferenceWins > sorted[j].PreferenceWins
	})

	fmt.Println("🏆 Top 5 Most Preferred:")
	for i := 0; i < 5 && i < len(sorted); i++ {
		if sorted[i].PreferenceWins == 0 {
			break
		}
		fmt.Printf("  %d. %s (%d wins)\n", i+1, sorted[i].Name, sorted[i].PreferenceWins)
	}
	fmt.Println()

	// Recent preferences with reasons
	fmt.Println("📝 Recent Comparisons with Reasons:")
	count := 0
	for i := len(prefs.Comparisons) - 1; i >= 0 && count < 5; i-- {
		pref := prefs.Comparisons[i]
		if pref.Reason != "" {
			fmt.Printf("  • %s > %s\n", pref.Winner, pref.Loser)
			fmt.Printf("    Reason: %s\n", pref.Reason)
			count++
		}
	}
	if count == 0 {
		fmt.Println("  (no reasons provided yet)\n")
	}
	fmt.Println()
}

func exportPreferences(prefs *PreferencesData) {
	filename := "spanning_set_preferences_export.json"
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		fmt.Printf("❌ Failed to export: %v\n\n", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("❌ Failed to write file: %v\n\n", err)
		return
	}

	fmt.Printf("✓ Exported %d preferences to %s\n\n", len(prefs.Comparisons), filename)
}

func loadPreferences() (*PreferencesData, error) {
	data, err := os.ReadFile(preferencesFile)
	if err != nil {
		return nil, err
	}

	var prefs PreferencesData
	if err := json.Unmarshal(data, &prefs); err != nil {
		return nil, err
	}

	return &prefs, nil
}

func savePreferences(prefs *PreferencesData) error {
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(preferencesFile, data, 0644)
}

func applyPreferences(sets []*SpanningSet, prefs *PreferencesData) {
	// Build ID to set map
	setMap := make(map[string]*SpanningSet)
	for _, set := range sets {
		setMap[set.ID] = set
	}

	// Apply each preference
	for _, pref := range prefs.Comparisons {
		if winner, ok := setMap[pref.WinnerID]; ok {
			winner.PreferenceWins++
		}
		if loser, ok := setMap[pref.LoserID]; ok {
			loser.PreferenceLosses++
		}
	}
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

// Reuse spanning set builders from spanning-sets tool
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
			set.HasWeekend = true
		} else {
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
			set.HasWeekend = true
		} else {
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
			set.HasWeekend = true
		} else {
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
			set.HasWeekend = true
		} else {
			set.HasWeekday = true
		}
	}

	for _, set := range sets {
		set.MemberCount = len(set.Members)
	}

	return sets
}

// ODS parsing (reused from spanning-sets)
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

func coverageStatus(set *SpanningSet) string {
	if set.HasWeekday && set.HasWeekend {
		return "24/7"
	} else if set.HasWeekday {
		return "Weekday only"
	} else if set.HasWeekend {
		return "Weekend only"
	}
	return "No coverage"
}

func randomIndex(max int) int {
	// Simple pseudo-random using time-based seed
	// For production, use crypto/rand or math/rand with proper seeding
	return os.Getpid() % max
}
