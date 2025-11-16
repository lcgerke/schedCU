package coverage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDescribeStudyType(t *testing.T) {
	tests := []struct {
		name            string
		studyTypeCode   string
		expectedShort   string
		expectedMedium  string
		expectedPatient string
	}{
		{
			name:            "CPMC CT Neuro",
			studyTypeCode:   "CPMC CT Neuro",
			expectedShort:   "Brain and spine CT scans at CPMC",
			expectedMedium:  "Brain and spine CT scans at California Pacific Medical Center",
			expectedPatient: "CT scans of your brain and spine at California Pacific Medical Center",
		},
		{
			name:            "Allen MR Body",
			studyTypeCode:   "Allen MR Body",
			expectedShort:   "Body MRI scans at Allen",
			expectedMedium:  "Body MRI scans at Allen Hospital",
			expectedPatient: "MRI scans of your body at Allen Hospital",
		},
		{
			name:            "NYPLH DX Chest/Abd",
			studyTypeCode:   "NYPLH DX Chest/Abd",
			expectedShort:   "Chest and abdomen X-rays at NYPLH",
			expectedMedium:  "Chest and abdomen X-rays at NewYork-Presbyterian Lower Manhattan Hospital",
			expectedPatient: "X-rays of your chest and abdomen at NewYork-Presbyterian Lower Manhattan Hospital",
		},
		{
			name:            "CPMC US",
			studyTypeCode:   "CPMC US",
			expectedShort:   "general ultrasound scans at CPMC",
			expectedMedium:  "general ultrasound scans at California Pacific Medical Center",
			expectedPatient: "Ultrasounds of the affected area at California Pacific Medical Center",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := DescribeStudyType(tt.studyTypeCode)

			assert.Equal(t, tt.studyTypeCode, desc.OriginalCode)
			assert.Contains(t, desc.ShortDesc, "at")
			assert.Contains(t, desc.MediumDesc, "at")
			assert.Contains(t, desc.PatientFriendly, "at")

			// Check that descriptions are non-empty
			assert.NotEmpty(t, desc.ShortDesc)
			assert.NotEmpty(t, desc.MediumDesc)
			assert.NotEmpty(t, desc.PatientFriendly)

			// Hospital should be extracted
			assert.NotEmpty(t, desc.Hospital)

			t.Logf("Study Type: %s", tt.studyTypeCode)
			t.Logf("  Short: %s", desc.ShortDesc)
			t.Logf("  Medium: %s", desc.MediumDesc)
			t.Logf("  Patient: %s", desc.PatientFriendly)
			if desc.WhatToExpect != "" {
				t.Logf("  What to Expect: %s", desc.WhatToExpect)
			}
		})
	}
}

func TestDescribeStudyType_AllHospitals(t *testing.T) {
	hospitalTests := []struct {
		code         string
		expectedName string
	}{
		{"CPMC CT Neuro", "California Pacific Medical Center"},
		{"Allen MR Body", "Allen Hospital"},
		{"NYPLH DX Bone", "NewYork-Presbyterian Lower Manhattan Hospital"},
		{"CHONY Neuro", "Children's Hospital of New York"},
	}

	for _, tt := range hospitalTests {
		t.Run(tt.code, func(t *testing.T) {
			desc := DescribeStudyType(tt.code)
			assert.Equal(t, tt.expectedName, desc.Hospital)
		})
	}
}

func TestDescribeStudyType_AllModalities(t *testing.T) {
	modalityTests := []struct {
		code             string
		expectedModality string
		expectedName     string
	}{
		{"CPMC CT Neuro", "CT", "CT scan"},
		{"Allen MR Body", "MR", "MRI scan"},
		{"NYPLH DX Bone", "DX", "X-ray"},
		{"CPMC US", "US", "Ultrasound"},
	}

	for _, tt := range modalityTests {
		t.Run(tt.code, func(t *testing.T) {
			desc := DescribeStudyType(tt.code)
			assert.Equal(t, tt.expectedModality, desc.Modality)
			assert.Equal(t, tt.expectedName, desc.ModalityName)
			assert.NotEmpty(t, desc.WhatToExpect) // Should have patient guidance
		})
	}
}

func TestGenerateStudyGlossary(t *testing.T) {
	studyTypes := []string{
		"CPMC CT Neuro",
		"CPMC CT Body",
		"Allen MR Neuro",
		"NYPLH DX Chest/Abd",
	}

	glossary := GenerateStudyGlossary(studyTypes)

	// Should contain all study types
	for _, st := range studyTypes {
		assert.Contains(t, glossary, st)
	}

	// Should contain hospital names
	assert.Contains(t, glossary, "California Pacific Medical Center")
	assert.Contains(t, glossary, "Allen Hospital")
	assert.Contains(t, glossary, "NewYork-Presbyterian")

	// Should have structure
	assert.Contains(t, glossary, "STUDY TYPE GLOSSARY")
	assert.Contains(t, glossary, "🏥")

	t.Logf("Glossary:\n%s", glossary)
}

func TestGenerateQuickReference(t *testing.T) {
	studyTypes := []string{
		"CPMC CT Neuro",
		"Allen MR Body",
		"NYPLH DX Bone",
	}

	reference := GenerateQuickReference(studyTypes)

	// Should contain all study types
	for _, st := range studyTypes {
		assert.Contains(t, reference, st)
	}

	// Should have table structure
	assert.Contains(t, reference, "QUICK REFERENCE")
	assert.Contains(t, reference, "Code")
	assert.Contains(t, reference, "What It Means")

	// Should have descriptions
	assert.Contains(t, reference, "Brain and spine")
	assert.Contains(t, reference, "Body")

	t.Logf("Quick Reference:\n%s", reference)
}

func TestGeneratePatientGuide(t *testing.T) {
	studyTypes := []string{
		"CPMC CT Neuro",
		"Allen MR Body",
		"NYPLH DX Bone",
		"CPMC US",
	}

	guide := GeneratePatientGuide(studyTypes)

	// Should contain patient-friendly language
	assert.Contains(t, guide, "your")
	assert.Contains(t, guide, "What to expect")

	// Should group by modality
	assert.Contains(t, guide, "CT scans")
	assert.Contains(t, guide, "MRI scans")
	assert.Contains(t, guide, "X-rays")
	assert.Contains(t, guide, "Ultrasounds")

	// Should have patient guidance
	assert.Contains(t, guide, "Lie on a table")
	assert.Contains(t, guide, "Gel applied to skin")

	t.Logf("Patient Guide:\n%s", guide)
}

func TestDescribeStudyType_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		studyTypeCode string
	}{
		{"Unknown hospital", "XYZ CT Neuro"},
		{"Unknown modality", "CPMC ABC Neuro"},
		{"Minimal info", "Test"},
		{"Empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := DescribeStudyType(tt.studyTypeCode)

			// Should not panic
			assert.NotNil(t, desc)
			assert.Equal(t, tt.studyTypeCode, desc.OriginalCode)

			// Descriptions should be non-empty (will use defaults)
			assert.NotEmpty(t, desc.ShortDesc)
			assert.NotEmpty(t, desc.MediumDesc)

			t.Logf("Edge case %q: %s", tt.studyTypeCode, desc.MediumDesc)
		})
	}
}
