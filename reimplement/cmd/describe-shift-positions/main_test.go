package main

import (
	"strings"
	"testing"
)

func TestDescribeShiftPosition_ExactMatch(t *testing.T) {
	// Test when shift position exactly matches a spanning set (no exclusions)
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"Allen CT Body",
			"Allen CT Neuro",
			"Allen MR Body",
		},
	}

	sets := map[string]*SpanningSet{
		"Allen-All": {
			Name: "Allen - All Services",
			Members: []string{
				"Allen CT Body",
				"Allen CT Neuro",
				"Allen MR Body",
			},
		},
	}

	records := []CoverageRecord{
		{StudyType: "Allen CT Body", Hospital: "Allen", Modality: "CT", Specialty: "Body"},
		{StudyType: "Allen CT Neuro", Hospital: "Allen", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "Allen MR Body", Hospital: "Allen", Modality: "MRI", Specialty: "Body"},
	}

	desc := describeShiftPosition(pos, sets, records)

	if len(desc.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(desc.Components))
	}

	if len(desc.Components[0].Exclusions) != 0 {
		t.Errorf("Expected no exclusions for exact match, got %d", len(desc.Components[0].Exclusions))
	}

	if desc.Components[0].SetName != "Allen - All Services" {
		t.Errorf("Expected 'Allen - All Services', got '%s'", desc.Components[0].SetName)
	}
}

func TestDescribeShiftPosition_WithExclusions(t *testing.T) {
	// Test when shift position covers most but not all of a spanning set
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"CPMC CT Body",
			"CPMC CT Neuro",
			// Missing: CPMC CT Chest
		},
	}

	sets := map[string]*SpanningSet{
		"CPMC-All": {
			Name: "CPMC - All Services",
			Members: []string{
				"CPMC CT Body",
				"CPMC CT Neuro",
				"CPMC CT Chest",
			},
		},
	}

	records := []CoverageRecord{
		{StudyType: "CPMC CT Body", Hospital: "CPMC", Modality: "CT", Specialty: "Body"},
		{StudyType: "CPMC CT Neuro", Hospital: "CPMC", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "CPMC CT Chest", Hospital: "CPMC", Modality: "CT", Specialty: "Chest"},
	}

	desc := describeShiftPosition(pos, sets, records)

	if len(desc.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(desc.Components))
	}

	if len(desc.Components[0].Exclusions) != 1 {
		t.Errorf("Expected 1 exclusion, got %d", len(desc.Components[0].Exclusions))
	}

	if desc.Components[0].Exclusions[0] != "CPMC CT Chest" {
		t.Errorf("Expected exclusion 'CPMC CT Chest', got '%s'", desc.Components[0].Exclusions[0])
	}
}

func TestDescribeShiftPosition_MultipleSpanningSets(t *testing.T) {
	// Test when shift position requires multiple spanning sets
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"Allen CT Body",
			"Allen CT Neuro",
			"NYPLH Body CT",
			"NYPLH Neuro CT",
		},
	}

	sets := map[string]*SpanningSet{
		"Allen-All": {
			Name: "Allen - All Services",
			Members: []string{
				"Allen CT Body",
				"Allen CT Neuro",
			},
		},
		"NYPLH-All": {
			Name: "NYPLH - All Services",
			Members: []string{
				"NYPLH Body CT",
				"NYPLH Neuro CT",
			},
		},
	}

	records := []CoverageRecord{
		{StudyType: "Allen CT Body", Hospital: "Allen", Modality: "CT", Specialty: "Body"},
		{StudyType: "Allen CT Neuro", Hospital: "Allen", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "NYPLH Body CT", Hospital: "NYPLH", Modality: "CT", Specialty: "Body"},
		{StudyType: "NYPLH Neuro CT", Hospital: "NYPLH", Modality: "CT", Specialty: "Neuro"},
	}

	desc := describeShiftPosition(pos, sets, records)

	if len(desc.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(desc.Components))
	}

	// Both should have no exclusions (exact matches)
	for i, comp := range desc.Components {
		if len(comp.Exclusions) != 0 {
			t.Errorf("Component %d should have no exclusions, got %d", i, len(comp.Exclusions))
		}
	}
}

func TestDescribeShiftPosition_Empty(t *testing.T) {
	// Test with empty shift position
	pos := &ShiftPosition{
		Name:       "EmptyShift",
		StudyTypes: []string{},
	}

	sets := map[string]*SpanningSet{
		"Test": {
			Name:    "Test Set",
			Members: []string{"Study1", "Study2"},
		},
	}

	records := []CoverageRecord{}

	desc := describeShiftPosition(pos, sets, records)

	if len(desc.Components) != 0 {
		t.Errorf("Expected 0 components for empty shift, got %d", len(desc.Components))
	}

	if desc.Description != "" {
		t.Errorf("Expected empty description, got '%s'", desc.Description)
	}
}

func TestDescribeShiftPosition_NoMatchingSets(t *testing.T) {
	// Test when no spanning sets cover the shift position's studies
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"Unique Study A",
			"Unique Study B",
		},
	}

	sets := map[string]*SpanningSet{
		"Different": {
			Name:    "Different Set",
			Members: []string{"Study X", "Study Y"},
		},
	}

	records := []CoverageRecord{
		{StudyType: "Unique Study A", Hospital: "Test", Modality: "CT", Specialty: "Body"},
		{StudyType: "Unique Study B", Hospital: "Test", Modality: "MRI", Specialty: "Neuro"},
	}

	desc := describeShiftPosition(pos, sets, records)

	// Should have no components since no sets match
	if len(desc.Components) != 0 {
		t.Errorf("Expected 0 components when no sets match, got %d", len(desc.Components))
	}
}

func TestDescribeShiftPosition_PreferLargerSets(t *testing.T) {
	// Test that algorithm prefers larger sets over smaller ones
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"CPMC CT Body",
			"CPMC CT Neuro",
			"CPMC MR Body",
		},
	}

	sets := map[string]*SpanningSet{
		"CPMC-All": {
			Name: "CPMC - All Services",
			Members: []string{
				"CPMC CT Body",
				"CPMC CT Neuro",
				"CPMC MR Body",
				"CPMC MR Neuro", // Extra one not in position
			},
		},
		"All-CT": {
			Name: "All CT",
			Members: []string{
				"CPMC CT Body",
				"CPMC CT Neuro",
			},
		},
	}

	records := []CoverageRecord{
		{StudyType: "CPMC CT Body", Hospital: "CPMC", Modality: "CT", Specialty: "Body"},
		{StudyType: "CPMC CT Neuro", Hospital: "CPMC", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "CPMC MR Body", Hospital: "CPMC", Modality: "MRI", Specialty: "Body"},
		{StudyType: "CPMC MR Neuro", Hospital: "CPMC", Modality: "MRI", Specialty: "Neuro"},
	}

	desc := describeShiftPosition(pos, sets, records)

	// Should prefer "CPMC - All Services" even with one exclusion
	// over "All CT" + "CPMC MR Body" separately
	if len(desc.Components) == 0 {
		t.Fatal("Expected at least 1 component")
	}

	// First component should be the larger set
	if desc.Components[0].SetName != "CPMC - All Services" {
		t.Errorf("Expected to prefer 'CPMC - All Services', got '%s'", desc.Components[0].SetName)
	}

	if len(desc.Components[0].Exclusions) != 1 {
		t.Errorf("Expected 1 exclusion, got %d", len(desc.Components[0].Exclusions))
	}
}

func TestDescribeShiftPosition_ComplexScenario(t *testing.T) {
	// Test complex real-world scenario: MidL covering multiple hospitals
	pos := &ShiftPosition{
		Name: "MidL",
		StudyTypes: []string{
			"Allen CT Body",
			"Allen CT Neuro",
			"Allen MR Body",
			"CPMC CT Body",
			"CPMC CT Neuro",
			"CPMC MR Body",
			// Missing: CPMC CT Chest
			"NYPLH Body CT",
			"NYPLH Neuro CT",
		},
	}

	sets := map[string]*SpanningSet{
		"Allen-All": {
			Name: "Allen - All Services",
			Members: []string{
				"Allen CT Body",
				"Allen CT Neuro",
				"Allen MR Body",
			},
		},
		"CPMC-All": {
			Name: "CPMC - All Services",
			Members: []string{
				"CPMC CT Body",
				"CPMC CT Neuro",
				"CPMC CT Chest",
				"CPMC MR Body",
			},
		},
		"NYPLH-All": {
			Name: "NYPLH - All Services",
			Members: []string{
				"NYPLH Body CT",
				"NYPLH Neuro CT",
			},
		},
	}

	records := []CoverageRecord{
		{StudyType: "Allen CT Body", Hospital: "Allen", Modality: "CT", Specialty: "Body"},
		{StudyType: "Allen CT Neuro", Hospital: "Allen", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "Allen MR Body", Hospital: "Allen", Modality: "MRI", Specialty: "Body"},
		{StudyType: "CPMC CT Body", Hospital: "CPMC", Modality: "CT", Specialty: "Body"},
		{StudyType: "CPMC CT Neuro", Hospital: "CPMC", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "CPMC CT Chest", Hospital: "CPMC", Modality: "CT", Specialty: "Chest"},
		{StudyType: "CPMC MR Body", Hospital: "CPMC", Modality: "MRI", Specialty: "Body"},
		{StudyType: "NYPLH Body CT", Hospital: "NYPLH", Modality: "CT", Specialty: "Body"},
		{StudyType: "NYPLH Neuro CT", Hospital: "NYPLH", Modality: "CT", Specialty: "Neuro"},
	}

	desc := describeShiftPosition(pos, sets, records)

	if desc.StudyCount != 8 {
		t.Errorf("Expected study count 8, got %d", desc.StudyCount)
	}

	// Should have 3 hospital sets
	if len(desc.Components) != 3 {
		t.Errorf("Expected 3 components (3 hospitals), got %d", len(desc.Components))
	}

	// Find CPMC component - should have 1 exclusion
	foundCPMC := false
	for _, comp := range desc.Components {
		if comp.SetName == "CPMC - All Services" {
			foundCPMC = true
			if len(comp.Exclusions) != 1 {
				t.Errorf("CPMC component should have 1 exclusion, got %d", len(comp.Exclusions))
			}
			if comp.Exclusions[0] != "CPMC CT Chest" {
				t.Errorf("Expected exclusion 'CPMC CT Chest', got '%s'", comp.Exclusions[0])
			}
		}
	}

	if !foundCPMC {
		t.Error("Expected to find CPMC - All Services component")
	}

	// Description should contain all three hospital sets
	if !strings.Contains(desc.Description, "Allen") {
		t.Error("Description should mention Allen")
	}
	if !strings.Contains(desc.Description, "CPMC") {
		t.Error("Description should mention CPMC")
	}
	if !strings.Contains(desc.Description, "NYPLH") {
		t.Error("Description should mention NYPLH")
	}
	if !strings.Contains(desc.Description, "EXCEPT") {
		t.Error("Description should contain EXCEPT clause for CPMC")
	}
}

func TestDescribeShiftPosition_ModalityBasedCoverage(t *testing.T) {
	// Test when shift position is better described by modality than hospital
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"Allen CT Body",
			"CPMC CT Neuro",
			"NYPLH Body CT",
		},
	}

	sets := map[string]*SpanningSet{
		"All-CT": {
			Name: "All CT",
			Members: []string{
				"Allen CT Body",
				"CPMC CT Neuro",
				"NYPLH Body CT",
				"CPMC CT Chest", // Extra
			},
		},
		"Allen-All": {
			Name: "Allen - All Services",
			Members: []string{
				"Allen CT Body",
				"Allen MR Body",
			},
		},
	}

	records := []CoverageRecord{
		{StudyType: "Allen CT Body", Hospital: "Allen", Modality: "CT", Specialty: "Body"},
		{StudyType: "CPMC CT Neuro", Hospital: "CPMC", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "NYPLH Body CT", Hospital: "NYPLH", Modality: "CT", Specialty: "Body"},
		{StudyType: "CPMC CT Chest", Hospital: "CPMC", Modality: "CT", Specialty: "Chest"},
		{StudyType: "Allen MR Body", Hospital: "Allen", Modality: "MRI", Specialty: "Body"},
	}

	desc := describeShiftPosition(pos, sets, records)

	// Should use "All CT" with one exclusion rather than multiple hospital sets
	if len(desc.Components) == 0 {
		t.Fatal("Expected at least 1 component")
	}

	// First component should be All CT
	if desc.Components[0].SetName != "All CT" {
		t.Errorf("Expected 'All CT' to be chosen, got '%s'", desc.Components[0].SetName)
	}
}

func TestDescribeShiftPosition_MultipleExclusions(t *testing.T) {
	// Test when a set has multiple exclusions
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"CPMC CT Body",
			"CPMC MR Body",
			// Missing: CPMC CT Neuro, CPMC MR Neuro
		},
	}

	sets := map[string]*SpanningSet{
		"CPMC-All": {
			Name: "CPMC - All Services",
			Members: []string{
				"CPMC CT Body",
				"CPMC CT Neuro",
				"CPMC MR Body",
				"CPMC MR Neuro",
			},
		},
	}

	records := []CoverageRecord{
		{StudyType: "CPMC CT Body", Hospital: "CPMC", Modality: "CT", Specialty: "Body"},
		{StudyType: "CPMC CT Neuro", Hospital: "CPMC", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "CPMC MR Body", Hospital: "CPMC", Modality: "MRI", Specialty: "Body"},
		{StudyType: "CPMC MR Neuro", Hospital: "CPMC", Modality: "MRI", Specialty: "Neuro"},
	}

	desc := describeShiftPosition(pos, sets, records)

	if len(desc.Components) == 0 {
		t.Fatal("Expected at least 1 component")
	}

	// Should have 2 exclusions
	if len(desc.Components[0].Exclusions) != 2 {
		t.Errorf("Expected 2 exclusions, got %d", len(desc.Components[0].Exclusions))
	}

	// Check both exclusions are present
	exclusions := desc.Components[0].Exclusions
	hasNeuro := contains(exclusions, "CPMC CT Neuro") || contains(exclusions, "CPMC MR Neuro")
	if !hasNeuro {
		t.Error("Expected to find Neuro exclusions")
	}
}

func TestBuildShiftPositions(t *testing.T) {
	assignments := []CoverageAssignment{
		{StudyType: "Study A", ShiftPosition: "Shift1"},
		{StudyType: "Study B", ShiftPosition: "Shift1"},
		{StudyType: "Study C", ShiftPosition: "Shift2"},
		{StudyType: "Study A", ShiftPosition: "Shift1"}, // Duplicate
	}

	positions := buildShiftPositions(assignments)

	if len(positions) != 2 {
		t.Errorf("Expected 2 shift positions, got %d", len(positions))
	}

	// Find Shift1
	var shift1 *ShiftPosition
	for _, pos := range positions {
		if pos.Name == "Shift1" {
			shift1 = pos
			break
		}
	}

	if shift1 == nil {
		t.Fatal("Expected to find Shift1")
	}

	// Should have 2 unique studies (duplicate should be ignored)
	if len(shift1.StudyTypes) != 2 {
		t.Errorf("Expected Shift1 to have 2 studies, got %d", len(shift1.StudyTypes))
	}

	// Studies should be sorted
	if shift1.StudyTypes[0] > shift1.StudyTypes[1] {
		t.Error("Expected studies to be sorted")
	}
}

func TestBuildAllSpanningSets_HospitalSets(t *testing.T) {
	records := []CoverageRecord{
		{StudyType: "Allen CT Body", Hospital: "Allen", Modality: "CT", Specialty: "Body"},
		{StudyType: "Allen CT Neuro", Hospital: "Allen", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "CPMC CT Body", Hospital: "CPMC", Modality: "CT", Specialty: "Body"},
	}

	sets := buildAllSpanningSets(records)

	// Should have hospital sets
	allenSet, hasAllen := sets["Allen-All"]
	if !hasAllen {
		t.Error("Expected Allen hospital set")
	}

	if len(allenSet.Members) != 2 {
		t.Errorf("Expected Allen set to have 2 members, got %d", len(allenSet.Members))
	}

	cpmcSet, hasCPMC := sets["CPMC-All"]
	if !hasCPMC {
		t.Error("Expected CPMC hospital set")
	}

	if len(cpmcSet.Members) != 1 {
		t.Errorf("Expected CPMC set to have 1 member, got %d", len(cpmcSet.Members))
	}
}

func TestBuildAllSpanningSets_ModalitySets(t *testing.T) {
	records := []CoverageRecord{
		{StudyType: "Allen CT Body", Hospital: "Allen", Modality: "CT", Specialty: "Body"},
		{StudyType: "CPMC CT Neuro", Hospital: "CPMC", Modality: "CT", Specialty: "Neuro"},
		{StudyType: "Allen MR Body", Hospital: "Allen", Modality: "MRI", Specialty: "Body"},
	}

	sets := buildAllSpanningSets(records)

	// Should have modality sets
	ctSet, hasCT := sets["All-CT"]
	if !hasCT {
		t.Error("Expected All CT set")
	}

	if len(ctSet.Members) != 2 {
		t.Errorf("Expected CT set to have 2 members, got %d", len(ctSet.Members))
	}

	mriSet, hasMRI := sets["All-MRI"]
	if !hasMRI {
		t.Error("Expected All MRI set")
	}

	if len(mriSet.Members) != 1 {
		t.Errorf("Expected MRI set to have 1 member, got %d", len(mriSet.Members))
	}
}

func TestBuildAllSpanningSets_SpecialtySets(t *testing.T) {
	records := []CoverageRecord{
		{StudyType: "Allen CT Body", Hospital: "Allen", Modality: "CT", Specialty: "Body"},
		{StudyType: "CPMC MR Body", Hospital: "CPMC", Modality: "MRI", Specialty: "Body"},
		{StudyType: "Allen CT Neuro", Hospital: "Allen", Modality: "CT", Specialty: "Neuro"},
	}

	sets := buildAllSpanningSets(records)

	// Should have specialty sets
	bodySet, hasBody := sets["Body-AllMod"]
	if !hasBody {
		t.Error("Expected Body specialty set")
	}

	if len(bodySet.Members) != 2 {
		t.Errorf("Expected Body set to have 2 members, got %d", len(bodySet.Members))
	}

	neuroSet, hasNeuro := sets["Neuro-AllMod"]
	if !hasNeuro {
		t.Error("Expected Neuro specialty set")
	}

	if len(neuroSet.Members) != 1 {
		t.Errorf("Expected Neuro set to have 1 member, got %d", len(neuroSet.Members))
	}
}

func TestDescribeShiftPosition_GreedyScoring(t *testing.T) {
	// Test that greedy algorithm correctly scores: coverage - (exclusions / 2)
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"Study A",
			"Study B",
			"Study C",
		},
	}

	sets := map[string]*SpanningSet{
		"Set1": {
			Name: "Set 1 - High coverage, some exclusions",
			Members: []string{
				"Study A",
				"Study B",
				"Study C",
				"Study X", // 1 exclusion
			},
			// Score: 3 coverage - 1/2 = 2.5
		},
		"Set2": {
			Name: "Set 2 - Lower coverage, no exclusions",
			Members: []string{
				"Study A",
				"Study B",
			},
			// Score: 2 coverage - 0 = 2.0
		},
	}

	records := []CoverageRecord{
		{StudyType: "Study A", Hospital: "Test", Modality: "CT", Specialty: "Body"},
		{StudyType: "Study B", Hospital: "Test", Modality: "CT", Specialty: "Body"},
		{StudyType: "Study C", Hospital: "Test", Modality: "CT", Specialty: "Body"},
		{StudyType: "Study X", Hospital: "Test", Modality: "CT", Specialty: "Body"},
	}

	desc := describeShiftPosition(pos, sets, records)

	// Should pick Set1 despite having an exclusion, because score is higher
	if len(desc.Components) == 0 {
		t.Fatal("Expected at least 1 component")
	}

	if desc.Components[0].SetName != "Set 1 - High coverage, some exclusions" {
		t.Errorf("Expected Set 1 to be chosen (higher score), got '%s'", desc.Components[0].SetName)
	}
}

func TestDescribeShiftPosition_DescriptionFormat(t *testing.T) {
	// Test description string format
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"Study A",
			"Study B",
		},
	}

	sets := map[string]*SpanningSet{
		"Set1": {
			Name:    "Set One",
			Members: []string{"Study A"},
		},
		"Set2": {
			Name:    "Set Two",
			Members: []string{"Study B"},
		},
	}

	records := []CoverageRecord{
		{StudyType: "Study A", Hospital: "Test", Modality: "CT", Specialty: "Body"},
		{StudyType: "Study B", Hospital: "Test", Modality: "CT", Specialty: "Body"},
	}

	desc := describeShiftPosition(pos, sets, records)

	// Description should use " + " as separator
	if !strings.Contains(desc.Description, " + ") {
		t.Errorf("Expected description to contain ' + ' separator, got '%s'", desc.Description)
	}

	// Should contain both set names
	if !strings.Contains(desc.Description, "Set One") {
		t.Error("Expected description to contain 'Set One'")
	}
	if !strings.Contains(desc.Description, "Set Two") {
		t.Error("Expected description to contain 'Set Two'")
	}
}

func TestDescribeShiftPosition_ExclusionFormat(t *testing.T) {
	// Test EXCEPT clause format in description
	pos := &ShiftPosition{
		Name: "TestShift",
		StudyTypes: []string{
			"Study A",
			"Study B",
		},
	}

	sets := map[string]*SpanningSet{
		"BigSet": {
			Name: "Big Set",
			Members: []string{
				"Study A",
				"Study B",
				"Study X",
				"Study Y",
			},
		},
	}

	records := []CoverageRecord{
		{StudyType: "Study A", Hospital: "Test", Modality: "CT", Specialty: "Body"},
		{StudyType: "Study B", Hospital: "Test", Modality: "CT", Specialty: "Body"},
		{StudyType: "Study X", Hospital: "Test", Modality: "CT", Specialty: "Body"},
		{StudyType: "Study Y", Hospital: "Test", Modality: "CT", Specialty: "Body"},
	}

	desc := describeShiftPosition(pos, sets, records)

	// Description should contain "EXCEPT"
	if !strings.Contains(desc.Description, "EXCEPT") {
		t.Errorf("Expected description to contain 'EXCEPT', got '%s'", desc.Description)
	}

	// Should have parentheses around exclusions
	if !strings.Contains(desc.Description, "(") || !strings.Contains(desc.Description, ")") {
		t.Errorf("Expected description to have parentheses around exclusions, got '%s'", desc.Description)
	}
}

func TestExtractHospital(t *testing.T) {
	tests := []struct {
		studyType string
		want      string
	}{
		{"CPMC CT Neuro", "CPMC"},
		{"Allen MR Body", "Allen"},
		{"NYPLH Body CT", "NYPLH"},
		{"CHONY Neuro", "CHONY"},
		{"Unknown Study", "Unknown"},
		{"", "Unknown"},
	}

	for _, tt := range tests {
		got := extractHospital(tt.studyType)
		if got != tt.want {
			t.Errorf("extractHospital(%q) = %q, want %q", tt.studyType, got, tt.want)
		}
	}
}

func TestExtractModality(t *testing.T) {
	tests := []struct {
		studyType string
		want      string
	}{
		{"CPMC CT Neuro", "CT"},
		{"Allen MR Body", "MRI"},
		{"Allen MRI Body", "MRI"},
		{"CPMC DX Chest", "X-Ray"},
		{"Allen US", "US"},
		{"Unknown Study", "Unknown"},
		{"", "Unknown"},
	}

	for _, tt := range tests {
		got := extractModality(tt.studyType)
		if got != tt.want {
			t.Errorf("extractModality(%q) = %q, want %q", tt.studyType, got, tt.want)
		}
	}
}

func TestExtractSpecialty(t *testing.T) {
	tests := []struct {
		studyType string
		want      string
	}{
		{"CPMC CT Neuro", "Neuro"},
		{"Allen MR Body", "Body"},
		{"CPMC CT Chest", "Chest"},
		{"Allen DX Bone", "Bone"},
		{"CPMC US", "General"},
		{"Unknown Study", "General"},
		{"", "General"},
	}

	for _, tt := range tests {
		got := extractSpecialty(tt.studyType)
		if got != tt.want {
			t.Errorf("extractSpecialty(%q) = %q, want %q", tt.studyType, got, tt.want)
		}
	}
}
