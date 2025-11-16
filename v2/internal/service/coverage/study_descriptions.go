// Package coverage provides study type description utilities
package coverage

import (
	"fmt"
	"strings"
)

// StudyTypeDescription contains human-readable information about a study type
type StudyTypeDescription struct {
	OriginalCode    string // e.g., "CPMC CT Neuro"
	ShortDesc       string // e.g., "Brain CT scans at CPMC"
	MediumDesc      string // e.g., "Brain and spine CT scans at California Pacific Medical Center"
	LongDesc        string // Full technical description
	PatientFriendly string // e.g., "CT scans of your brain and spine at CPMC"
	Hospital        string // e.g., "California Pacific Medical Center"
	HospitalCode    string // e.g., "CPMC"
	Modality        string // e.g., "CT"
	ModalityName    string // e.g., "CT scan"
	StudyArea       string // e.g., "Brain and spine"
	WhatToExpect    string // Patient guidance
}

// Hospital mappings
var hospitals = map[string]string{
	"CPMC":  "California Pacific Medical Center",
	"Allen": "Allen Hospital",
	"NYPLH": "NewYork-Presbyterian Lower Manhattan Hospital",
	"CHONY": "Children's Hospital of New York",
}

// Modality information
type modalityInfo struct {
	Name        string
	LongName    string
	Description string
	Verb        string
	PatientNote string
}

var modalities = map[string]modalityInfo{
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
type studyAreaInfo struct {
	Name        string
	LongName    string
	Description string
}

var studyAreas = map[string]studyAreaInfo{
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

// DescribeStudyType generates a human-readable description of a study type code
//
// Example:
//
//	desc := DescribeStudyType("CPMC CT Neuro")
//	fmt.Println(desc.MediumDesc)
//	// Output: "Brain and spine CT scans at California Pacific Medical Center"
func DescribeStudyType(studyTypeCode string) StudyTypeDescription {
	desc := StudyTypeDescription{
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
	desc.LongDesc = generateLongDescription(desc, studyTypeCode)
	desc.PatientFriendly = generatePatientFriendlyDescription(desc)

	return desc
}

// generateShortDescription creates a brief description
// Example: "Brain CT scans at CPMC"
func generateShortDescription(desc StudyTypeDescription) string {
	hospital := desc.HospitalCode
	if hospital == "" {
		hospital = "Unknown Hospital"
	}

	modality := "imaging studies"
	if modalityInfo, ok := modalities[desc.Modality]; ok {
		modality = modalityInfo.Verb
	}

	area := "general"
	if desc.StudyArea != "" {
		area = strings.ToLower(desc.StudyArea)
	}

	return fmt.Sprintf("%s %s at %s", desc.StudyArea, modality, hospital)
}

// generateMediumDescription creates a medium-length description
// Example: "Brain and spine CT scans at California Pacific Medical Center"
func generateMediumDescription(desc StudyTypeDescription) string {
	hospital := desc.Hospital
	if hospital == "" {
		hospital = "Unknown Hospital"
	}

	modality := "imaging studies"
	if modalityInfo, ok := modalities[desc.Modality]; ok {
		modality = modalityInfo.Verb
	}

	area := "general"
	if desc.StudyArea != "" {
		area = desc.StudyArea
	}

	return fmt.Sprintf("%s %s at %s", area, modality, hospital)
}

// generateLongDescription creates a detailed technical description
// Example: "Brain and spine CT scans (Computed Tomography) at California Pacific Medical Center - detailed cross-sectional X-ray imaging of head, brain, spine, and nervous system"
func generateLongDescription(desc StudyTypeDescription, originalCode string) string {
	hospital := desc.Hospital
	if hospital == "" {
		hospital = "Unknown Hospital"
	}

	modalityInfo, hasModality := modalities[desc.Modality]
	areaInfo, hasArea := studyAreas[getAreaKey(originalCode)]

	if !hasModality && !hasArea {
		return desc.MediumDesc
	}

	modalityDesc := "imaging studies"
	if hasModality {
		modalityDesc = fmt.Sprintf("%s (%s)", modalityInfo.Verb, modalityInfo.LongName)
	}

	area := "general anatomy"
	if hasArea {
		area = areaInfo.Name
	}

	description := ""
	if hasModality {
		description = modalityInfo.Description
	}

	areaDesc := ""
	if hasArea {
		areaDesc = areaInfo.Description
	}

	if description != "" && areaDesc != "" {
		return fmt.Sprintf("%s %s at %s - %s of %s", area, modalityDesc, hospital, description, areaDesc)
	}

	return desc.MediumDesc
}

// generatePatientFriendlyDescription creates a patient-friendly description
// Example: "CT scans of your brain and spine at California Pacific Medical Center"
func generatePatientFriendlyDescription(desc StudyTypeDescription) string {
	hospital := desc.Hospital
	if hospital == "" {
		hospital = "the hospital"
	}

	modality := "imaging"
	if modalityInfo, ok := modalities[desc.Modality]; ok {
		modality = modalityInfo.Name + "s"
	}

	area := "the affected area"
	if desc.StudyArea != "" {
		area = "your " + strings.ToLower(desc.StudyArea)
	}

	return fmt.Sprintf("%s of %s at %s", modality, area, hospital)
}

// getAreaKey extracts the study area key from the original code
func getAreaKey(code string) string {
	for area := range studyAreas {
		if strings.Contains(code, area) {
			return area
		}
	}
	return ""
}

// GenerateStudyGlossary creates a formatted glossary of study types
func GenerateStudyGlossary(studyTypes []string) string {
	var lines []string

	lines = append(lines, "STUDY TYPE GLOSSARY")
	lines = append(lines, strings.Repeat("=", 70))
	lines = append(lines, "")

	// Group by hospital
	byHospital := make(map[string][]string)
	for _, studyType := range studyTypes {
		desc := DescribeStudyType(studyType)
		hospital := desc.Hospital
		if hospital == "" {
			hospital = "Other"
		}
		byHospital[hospital] = append(byHospital[hospital], studyType)
	}

	// Print grouped by hospital
	for hospital, studies := range byHospital {
		lines = append(lines, fmt.Sprintf("🏥 %s", hospital))
		lines = append(lines, strings.Repeat("─", 70))
		lines = append(lines, "")

		for _, studyType := range studies {
			desc := DescribeStudyType(studyType)
			lines = append(lines, fmt.Sprintf("  • %s", studyType))
			lines = append(lines, fmt.Sprintf("    → %s", desc.MediumDesc))
			lines = append(lines, "")
		}

		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// GenerateQuickReference creates a quick reference table
func GenerateQuickReference(studyTypes []string) string {
	var lines []string

	lines = append(lines, "QUICK REFERENCE GUIDE")
	lines = append(lines, strings.Repeat("=", 70))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%-30s %s", "Code", "What It Means"))
	lines = append(lines, strings.Repeat("─", 70))

	for _, studyType := range studyTypes {
		desc := DescribeStudyType(studyType)
		lines = append(lines, fmt.Sprintf("%-30s %s", studyType, desc.ShortDesc))
	}

	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

// GeneratePatientGuide creates a patient-friendly guide
func GeneratePatientGuide(studyTypes []string) string {
	var lines []string

	lines = append(lines, "PATIENT GUIDE: Understanding Your Imaging Study")
	lines = append(lines, strings.Repeat("=", 70))
	lines = append(lines, "")

	// Group by modality
	byModality := make(map[string][]string)
	for _, studyType := range studyTypes {
		desc := DescribeStudyType(studyType)
		modality := desc.ModalityName
		if modality == "" {
			modality = "Other"
		}
		byModality[modality] = append(byModality[modality], studyType)
	}

	// Print grouped by modality
	for modality, studies := range byModality {
		lines = append(lines, fmt.Sprintf("📋 %ss", modality))
		lines = append(lines, strings.Repeat("─", 70))
		lines = append(lines, "")

		for _, studyType := range studies {
			desc := DescribeStudyType(studyType)
			lines = append(lines, fmt.Sprintf("  %s", desc.PatientFriendly))

			if desc.WhatToExpect != "" {
				lines = append(lines, fmt.Sprintf("     What to expect: %s", desc.WhatToExpect))
			}

			lines = append(lines, "")
		}

		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}
