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

// ODS XML structures (reused from validate-coverage-go)
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

// Study type description
type StudyDescription struct {
	OriginalCode    string
	ShortDesc       string
	MediumDesc      string
	LongDesc        string
	PatientFriendly string
	Hospital        string
	HospitalCode    string
	Modality        string
	ModalityName    string
	StudyArea       string
	WhatToExpect    string
}

// Hospital mappings
var hospitals = map[string]string{
	"CPMC":  "California Pacific Medical Center",
	"Allen": "Allen Hospital",
	"NYPLH": "NewYork-Presbyterian Lower Manhattan Hospital",
	"CHONY": "Children's Hospital of New York",
}

// Modality information
type ModalityInfo struct {
	Name        string
	LongName    string
	Description string
	Verb        string
	PatientNote string
}

var modalities = map[string]ModalityInfo{
	"CT": {
		Name:        "CT scan",
		LongName:    "Computed Tomography",
		Description: "detailed cross-sectional X-ray imaging",
		Verb:        "CT scans",
		PatientNote: "Lie on a table that slides through a donut-shaped machine",
	},
	"MR": {
		Name:        "MRI scan",
		LongName:    "Magnetic Resonance Imaging",
		Description: "soft tissue visualization using magnetic fields",
		Verb:        "MRI scans",
		PatientNote: "Lie still in a tube-shaped scanner (can be noisy)",
	},
	"MRI": {
		Name:        "MRI scan",
		LongName:    "Magnetic Resonance Imaging",
		Description: "soft tissue visualization using magnetic fields",
		Verb:        "MRI scans",
		PatientNote: "Lie still in a tube-shaped scanner (can be noisy)",
	},
	"DX": {
		Name:        "X-ray",
		LongName:    "Radiography",
		Description: "standard X-ray imaging",
		Verb:        "X-rays",
		PatientNote: "Stand or lie down while images are taken",
	},
	"US": {
		Name:        "Ultrasound",
		LongName:    "Ultrasonography",
		Description: "real-time imaging using sound waves",
		Verb:        "ultrasound scans",
		PatientNote: "Gel applied to skin, handheld device moved over area",
	},
}

// Study area information
type StudyAreaInfo struct {
	Name        string
	LongName    string
	Description string
}

var studyAreas = map[string]StudyAreaInfo{
	"Neuro": {
		Name:        "Brain and spine",
		LongName:    "Neurological",
		Description: "head, brain, spine, and nervous system",
	},
	"Body": {
		Name:        "Body",
		LongName:    "Body imaging",
		Description: "chest, abdomen, pelvis, and internal organs",
	},
	"Chest/Abd": {
		Name:        "Chest and abdomen",
		LongName:    "Thoracic and abdominal",
		Description: "chest, lungs, heart, and abdominal organs",
	},
	"Chest": {
		Name:        "Chest",
		LongName:    "Thoracic",
		Description: "chest, lungs, and heart",
	},
	"Bone": {
		Name:        "Bone and skeletal",
		LongName:    "Musculoskeletal",
		Description: "bones, joints, and skeletal structure",
	},
}

func main() {
	filepath := "/home/user/schedCU/cuSchedNormalized.ods"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("              STUDY TYPE GLOSSARY")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Printf("Reading from: %s\n\n", filepath)

	studyTypes, err := extractStudyTypes(filepath)
	if err != nil {
		fmt.Printf("❌ Failed to extract study types: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d unique study types\n\n", len(studyTypes))
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	generateStudyGlossary(studyTypes)
	fmt.Println()
	generateQuickReference(studyTypes)
	fmt.Println()
	generatePatientGuide(studyTypes)
}

func extractStudyTypes(filepath string) ([]string, error) {
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

	studyTypeSet := make(map[string]bool)

	for _, table := range doc.Body.Spreadsheet.Tables {
		for i := 1; i < len(table.Rows); i++ {
			row := table.Rows[i]
			if len(row.Cells) == 0 {
				continue
			}

			studyType := extractCellText(row.Cells[0])
			if studyType != "" {
				studyTypeSet[studyType] = true
			}
		}
	}

	var studyTypes []string
	for st := range studyTypeSet {
		studyTypes = append(studyTypes, st)
	}
	sort.Strings(studyTypes)

	return studyTypes, nil
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

func describeStudyType(studyTypeCode string) StudyDescription {
	desc := StudyDescription{
		OriginalCode: studyTypeCode,
	}

	// Extract hospital
	for code, name := range hospitals {
		if strings.Contains(studyTypeCode, code) {
			desc.HospitalCode = code
			desc.Hospital = name
			break
		}
	}

	// Extract modality
	for code, info := range modalities {
		if strings.Contains(strings.ToUpper(studyTypeCode), code) {
			desc.Modality = code
			desc.ModalityName = info.Name
			desc.WhatToExpect = info.PatientNote
			break
		}
	}

	// Extract study area
	for area, info := range studyAreas {
		if strings.Contains(studyTypeCode, area) {
			desc.StudyArea = info.Name
			break
		}
	}

	// Generate descriptions
	desc.ShortDesc = generateShortDescription(desc)
	desc.MediumDesc = generateMediumDescription(desc)
	desc.PatientFriendly = generatePatientFriendlyDescription(desc)

	return desc
}

func generateShortDescription(desc StudyDescription) string {
	hospital := desc.HospitalCode
	if hospital == "" {
		hospital = "Unknown Hospital"
	}

	modalityVerb := "imaging studies"
	if info, ok := modalities[desc.Modality]; ok {
		modalityVerb = info.Verb
	}

	area := desc.StudyArea
	if area == "" {
		area = "general"
	}

	return fmt.Sprintf("%s %s at %s", area, modalityVerb, hospital)
}

func generateMediumDescription(desc StudyDescription) string {
	hospital := desc.Hospital
	if hospital == "" {
		hospital = "Unknown Hospital"
	}

	modalityVerb := "imaging studies"
	if info, ok := modalities[desc.Modality]; ok {
		modalityVerb = info.Verb
	}

	area := desc.StudyArea
	if area == "" {
		area = "general"
	}

	return fmt.Sprintf("%s %s at %s", area, modalityVerb, hospital)
}

func generatePatientFriendlyDescription(desc StudyDescription) string {
	hospital := desc.Hospital
	if hospital == "" {
		hospital = "the hospital"
	}

	modalityName := "imaging"
	if info, ok := modalities[desc.Modality]; ok {
		modalityName = info.Name + "s"
	}

	area := "the affected area"
	if desc.StudyArea != "" {
		area = "your " + strings.ToLower(desc.StudyArea)
	}

	return fmt.Sprintf("%s of %s at %s", modalityName, area, hospital)
}

func generateStudyGlossary(studyTypes []string) {
	// Group by hospital
	byHospital := make(map[string][]string)

	for _, st := range studyTypes {
		desc := describeStudyType(st)
		hospital := desc.Hospital
		if hospital == "" {
			hospital = "Other"
		}
		byHospital[hospital] = append(byHospital[hospital], st)
	}

	// Print grouped by hospital
	var hospitals []string
	for h := range byHospital {
		hospitals = append(hospitals, h)
	}
	sort.Strings(hospitals)

	for _, hospital := range hospitals {
		studies := byHospital[hospital]
		sort.Strings(studies)

		fmt.Printf("🏥 %s\n", hospital)
		fmt.Println(strings.Repeat("─", 70))
		fmt.Println()

		for _, studyType := range studies {
			desc := describeStudyType(studyType)
			fmt.Printf("  • %s\n", studyType)
			fmt.Printf("    → %s\n\n", desc.MediumDesc)
		}

		fmt.Println()
	}
}

func generateQuickReference(studyTypes []string) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("              QUICK REFERENCE GUIDE")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	fmt.Printf("%-30s %s\n", "Code", "What It Means")
	fmt.Println(strings.Repeat("─", 70))

	for _, studyType := range studyTypes {
		desc := describeStudyType(studyType)
		fmt.Printf("%-30s %s\n", studyType, desc.ShortDesc)
	}

	fmt.Println()
}

func generatePatientGuide(studyTypes []string) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("        PATIENT GUIDE: Understanding Your Imaging Study")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Group by modality
	byModality := make(map[string][]string)

	for _, st := range studyTypes {
		desc := describeStudyType(st)
		modality := desc.ModalityName
		if modality == "" {
			modality = "Other"
		}
		byModality[modality] = append(byModality[modality], st)
	}

	// Print grouped by modality
	var modalityNames []string
	for m := range byModality {
		modalityNames = append(modalityNames, m)
	}
	sort.Strings(modalityNames)

	for _, modality := range modalityNames {
		studies := byModality[modality]
		sort.Strings(studies)

		fmt.Printf("📋 %ss\n", modality)
		fmt.Println(strings.Repeat("─", 70))
		fmt.Println()

		for _, studyType := range studies {
			desc := describeStudyType(studyType)
			fmt.Printf("  %s\n", desc.PatientFriendly)

			if desc.WhatToExpect != "" {
				fmt.Printf("     What to expect: %s\n", desc.WhatToExpect)
			}

			fmt.Println()
		}

		fmt.Println()
	}
}
