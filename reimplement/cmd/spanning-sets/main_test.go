package main

import (
	"testing"
)

func TestExtractHospital(t *testing.T) {
	tests := []struct {
		name      string
		studyType string
		want      string
	}{
		{
			name:      "CPMC hospital",
			studyType: "CPMC CT Neuro",
			want:      "CPMC",
		},
		{
			name:      "Allen hospital",
			studyType: "Allen MR Body",
			want:      "Allen",
		},
		{
			name:      "NYPLH hospital",
			studyType: "NYPLH Body CT",
			want:      "NYPLH",
		},
		{
			name:      "CHONY hospital",
			studyType: "CHONY Neuro",
			want:      "CHONY",
		},
		{
			name:      "Unknown hospital",
			studyType: "Random Study Type",
			want:      "Unknown",
		},
		{
			name:      "Empty string",
			studyType: "",
			want:      "Unknown",
		},
		{
			name:      "Case sensitive - lowercase (case matters)",
			studyType: "cpmc CT Neuro",
			want:      "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHospital(tt.studyType)
			if got != tt.want {
				t.Errorf("extractHospital(%q) = %q, want %q", tt.studyType, got, tt.want)
			}
		})
	}
}

func TestExtractModality(t *testing.T) {
	tests := []struct {
		name      string
		studyType string
		want      string
	}{
		{
			name:      "CT modality",
			studyType: "CPMC CT Neuro",
			want:      "CT",
		},
		{
			name:      "MR modality",
			studyType: "Allen MR Body",
			want:      "MRI",
		},
		{
			name:      "MRI modality",
			studyType: "NYPLH MRI Neuro",
			want:      "MRI",
		},
		{
			name:      "DX modality (X-Ray)",
			studyType: "Allen DX Chest/Abd",
			want:      "X-Ray",
		},
		{
			name:      "US modality",
			studyType: "CPMC US",
			want:      "US",
		},
		{
			name:      "NM modality",
			studyType: "CPMC NM Bone Scan",
			want:      "NM",
		},
		{
			name:      "PET modality",
			studyType: "Allen PET Body",
			want:      "PET",
		},
		{
			name:      "Unknown modality",
			studyType: "Random Study",
			want:      "Unknown",
		},
		{
			name:      "Lowercase CT",
			studyType: "cpmc ct neuro",
			want:      "CT",
		},
		{
			name:      "Empty string",
			studyType: "",
			want:      "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractModality(tt.studyType)
			if got != tt.want {
				t.Errorf("extractModality(%q) = %q, want %q", tt.studyType, got, tt.want)
			}
		})
	}
}

func TestExtractSpecialty(t *testing.T) {
	tests := []struct {
		name      string
		studyType string
		want      string
	}{
		{
			name:      "Neuro specialty",
			studyType: "CPMC CT Neuro",
			want:      "Neuro",
		},
		{
			name:      "Body specialty",
			studyType: "Allen MR Body",
			want:      "Body",
		},
		{
			name:      "Chest specialty",
			studyType: "CPMC CT Chest",
			want:      "Chest",
		},
		{
			name:      "Bone specialty",
			studyType: "Allen DX Bone",
			want:      "Bone",
		},
		{
			name:      "General (no specialty found)",
			studyType: "CPMC US",
			want:      "General",
		},
		{
			name:      "Empty string",
			studyType: "",
			want:      "General",
		},
		{
			name:      "Case sensitive - lowercase (case matters)",
			studyType: "cpmc ct neuro",
			want:      "General",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSpecialty(tt.studyType)
			if got != tt.want {
				t.Errorf("extractSpecialty(%q) = %q, want %q", tt.studyType, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "Item exists in slice",
			slice: []string{"apple", "banana", "cherry"},
			item:  "banana",
			want:  true,
		},
		{
			name:  "Item does not exist in slice",
			slice: []string{"apple", "banana", "cherry"},
			item:  "grape",
			want:  false,
		},
		{
			name:  "Empty slice",
			slice: []string{},
			item:  "apple",
			want:  false,
		},
		{
			name:  "Nil slice",
			slice: nil,
			item:  "apple",
			want:  false,
		},
		{
			name:  "Empty string in slice",
			slice: []string{"", "a", "b"},
			item:  "",
			want:  true,
		},
		{
			name:  "Case sensitive - exact match required",
			slice: []string{"Apple", "Banana"},
			item:  "apple",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, got, tt.want)
			}
		})
	}
}

func TestBuildModalitySpanningSets(t *testing.T) {
	tests := []struct {
		name    string
		records []CoverageRecord
		want    map[string]expectedSet
	}{
		{
			name: "Single modality, single study type",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Hospital:  "CPMC",
					Modality:  "CT",
					Specialty: "Neuro",
					IsWeekend: false,
				},
			},
			want: map[string]expectedSet{
				"CT": {
					name:        "All CT",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
					members:     []string{"CPMC CT Neuro"},
				},
			},
		},
		{
			name: "Multiple modalities with weekday and weekend coverage",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Modality:  "CT",
					IsWeekend: false,
				},
				{
					StudyType: "CPMC CT Neuro",
					Modality:  "CT",
					IsWeekend: true,
				},
				{
					StudyType: "Allen MR Body",
					Modality:  "MRI",
					IsWeekend: false,
				},
			},
			want: map[string]expectedSet{
				"CT": {
					name:        "All CT",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  true,
					members:     []string{"CPMC CT Neuro"},
				},
				"MRI": {
					name:        "All MRI",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
					members:     []string{"Allen MR Body"},
				},
			},
		},
		{
			name: "Same modality, multiple study types",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Modality:  "CT",
					IsWeekend: false,
				},
				{
					StudyType: "Allen CT Body",
					Modality:  "CT",
					IsWeekend: false,
				},
				{
					StudyType: "NYPLH Body CT",
					Modality:  "CT",
					IsWeekend: true,
				},
			},
			want: map[string]expectedSet{
				"CT": {
					name:        "All CT",
					memberCount: 3,
					hasWeekday:  true,
					hasWeekend:  true,
					members:     []string{"CPMC CT Neuro", "Allen CT Body", "NYPLH Body CT"},
				},
			},
		},
		{
			name: "Duplicate records should not duplicate members",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Modality:  "CT",
					IsWeekend: false,
				},
				{
					StudyType: "CPMC CT Neuro",
					Modality:  "CT",
					IsWeekend: false,
				},
				{
					StudyType: "CPMC CT Neuro",
					Modality:  "CT",
					IsWeekend: true,
				},
			},
			want: map[string]expectedSet{
				"CT": {
					name:        "All CT",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  true,
					members:     []string{"CPMC CT Neuro"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildModalitySpanningSets(tt.records)

			// Check number of sets
			if len(got) != len(tt.want) {
				t.Errorf("buildModalitySpanningSets() returned %d sets, want %d", len(got), len(tt.want))
			}

			// Check each set
			for key, expected := range tt.want {
				set, exists := got[key]
				if !exists {
					t.Errorf("buildModalitySpanningSets() missing key %q", key)
					continue
				}

				if set.Name != expected.name {
					t.Errorf("set[%q].Name = %q, want %q", key, set.Name, expected.name)
				}

				if set.MemberCount != expected.memberCount {
					t.Errorf("set[%q].MemberCount = %d, want %d", key, set.MemberCount, expected.memberCount)
				}

				if set.HasWeekday != expected.hasWeekday {
					t.Errorf("set[%q].HasWeekday = %v, want %v", key, set.HasWeekday, expected.hasWeekday)
				}

				if set.HasWeekend != expected.hasWeekend {
					t.Errorf("set[%q].HasWeekend = %v, want %v", key, set.HasWeekend, expected.hasWeekend)
				}

				if len(set.Members) != len(expected.members) {
					t.Errorf("set[%q] has %d members, want %d", key, len(set.Members), len(expected.members))
				}

				// Check all expected members are present
				for _, member := range expected.members {
					if !contains(set.Members, member) {
						t.Errorf("set[%q] missing member %q", key, member)
					}
				}
			}
		})
	}
}

func TestBuildSpecialtySpanningSets(t *testing.T) {
	tests := []struct {
		name    string
		records []CoverageRecord
		want    map[string]expectedSet
	}{
		{
			name: "Single specialty",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Specialty: "Neuro",
					IsWeekend: false,
				},
			},
			want: map[string]expectedSet{
				"Neuro": {
					name:        "Neuro - All Modalities",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
					members:     []string{"CPMC CT Neuro"},
				},
			},
		},
		{
			name: "Multiple specialties",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Specialty: "Neuro",
					IsWeekend: false,
				},
				{
					StudyType: "Allen MR Body",
					Specialty: "Body",
					IsWeekend: true,
				},
				{
					StudyType: "CPMC CT Chest",
					Specialty: "Chest",
					IsWeekend: false,
				},
			},
			want: map[string]expectedSet{
				"Neuro": {
					name:        "Neuro - All Modalities",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
				},
				"Body": {
					name:        "Body - All Modalities",
					memberCount: 1,
					hasWeekday:  false,
					hasWeekend:  true,
				},
				"Chest": {
					name:        "Chest - All Modalities",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
				},
			},
		},
		{
			name: "Same specialty, multiple modalities",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Specialty: "Neuro",
					Modality:  "CT",
					IsWeekend: false,
				},
				{
					StudyType: "Allen MR Neuro",
					Specialty: "Neuro",
					Modality:  "MRI",
					IsWeekend: true,
				},
			},
			want: map[string]expectedSet{
				"Neuro": {
					name:        "Neuro - All Modalities",
					memberCount: 2,
					hasWeekday:  true,
					hasWeekend:  true,
					members:     []string{"CPMC CT Neuro", "Allen MR Neuro"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSpecialtySpanningSets(tt.records)

			if len(got) != len(tt.want) {
				t.Errorf("buildSpecialtySpanningSets() returned %d sets, want %d", len(got), len(tt.want))
			}

			for key, expected := range tt.want {
				set, exists := got[key]
				if !exists {
					t.Errorf("buildSpecialtySpanningSets() missing key %q", key)
					continue
				}

				if set.Name != expected.name {
					t.Errorf("set[%q].Name = %q, want %q", key, set.Name, expected.name)
				}

				if set.MemberCount != expected.memberCount {
					t.Errorf("set[%q].MemberCount = %d, want %d", key, set.MemberCount, expected.memberCount)
				}

				if set.HasWeekday != expected.hasWeekday {
					t.Errorf("set[%q].HasWeekday = %v, want %v", key, set.HasWeekday, expected.hasWeekday)
				}

				if set.HasWeekend != expected.hasWeekend {
					t.Errorf("set[%q].HasWeekend = %v, want %v", key, set.HasWeekend, expected.hasWeekend)
				}

				if len(expected.members) > 0 {
					for _, member := range expected.members {
						if !contains(set.Members, member) {
							t.Errorf("set[%q] missing member %q", key, member)
						}
					}
				}
			}
		})
	}
}

func TestBuildHospitalSpanningSets(t *testing.T) {
	tests := []struct {
		name    string
		records []CoverageRecord
		want    map[string]expectedSet
	}{
		{
			name: "Single hospital",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Hospital:  "CPMC",
					IsWeekend: false,
				},
			},
			want: map[string]expectedSet{
				"CPMC": {
					name:        "CPMC - All Services",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
				},
			},
		},
		{
			name: "Multiple hospitals",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Hospital:  "CPMC",
					IsWeekend: false,
				},
				{
					StudyType: "Allen MR Body",
					Hospital:  "Allen",
					IsWeekend: true,
				},
				{
					StudyType: "NYPLH Body CT",
					Hospital:  "NYPLH",
					IsWeekend: false,
				},
			},
			want: map[string]expectedSet{
				"CPMC": {
					name:        "CPMC - All Services",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
				},
				"Allen": {
					name:        "Allen - All Services",
					memberCount: 1,
					hasWeekday:  false,
					hasWeekend:  true,
				},
				"NYPLH": {
					name:        "NYPLH - All Services",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
				},
			},
		},
		{
			name: "Same hospital, multiple services",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Hospital:  "CPMC",
					IsWeekend: false,
				},
				{
					StudyType: "CPMC MR Body",
					Hospital:  "CPMC",
					IsWeekend: false,
				},
				{
					StudyType: "CPMC US",
					Hospital:  "CPMC",
					IsWeekend: true,
				},
			},
			want: map[string]expectedSet{
				"CPMC": {
					name:        "CPMC - All Services",
					memberCount: 3,
					hasWeekday:  true,
					hasWeekend:  true,
					members:     []string{"CPMC CT Neuro", "CPMC MR Body", "CPMC US"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildHospitalSpanningSets(tt.records)

			if len(got) != len(tt.want) {
				t.Errorf("buildHospitalSpanningSets() returned %d sets, want %d", len(got), len(tt.want))
			}

			for key, expected := range tt.want {
				set, exists := got[key]
				if !exists {
					t.Errorf("buildHospitalSpanningSets() missing key %q", key)
					continue
				}

				if set.Name != expected.name {
					t.Errorf("set[%q].Name = %q, want %q", key, set.Name, expected.name)
				}

				if set.MemberCount != expected.memberCount {
					t.Errorf("set[%q].MemberCount = %d, want %d", key, set.MemberCount, expected.memberCount)
				}

				if set.HasWeekday != expected.hasWeekday {
					t.Errorf("set[%q].HasWeekday = %v, want %v", key, set.HasWeekday, expected.hasWeekday)
				}

				if set.HasWeekend != expected.hasWeekend {
					t.Errorf("set[%q].HasWeekend = %v, want %v", key, set.HasWeekend, expected.hasWeekend)
				}

				if len(expected.members) > 0 {
					for _, member := range expected.members {
						if !contains(set.Members, member) {
							t.Errorf("set[%q] missing member %q", key, member)
						}
					}
				}
			}
		})
	}
}

func TestBuildCrossSpanningSets(t *testing.T) {
	tests := []struct {
		name    string
		records []CoverageRecord
		want    map[string]expectedSet
	}{
		{
			name: "Single cross-dimensional set",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Modality:  "CT",
					Specialty: "Neuro",
					IsWeekend: false,
				},
			},
			want: map[string]expectedSet{
				"CT Neuro": {
					name:        "CT Neuro - All Hospitals",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
				},
			},
		},
		{
			name: "Same modality-specialty across hospitals",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Hospital:  "CPMC",
					Modality:  "CT",
					Specialty: "Neuro",
					IsWeekend: false,
				},
				{
					StudyType: "Allen CT Neuro",
					Hospital:  "Allen",
					Modality:  "CT",
					Specialty: "Neuro",
					IsWeekend: true,
				},
				{
					StudyType: "NYPLH Neuro CT",
					Hospital:  "NYPLH",
					Modality:  "CT",
					Specialty: "Neuro",
					IsWeekend: false,
				},
			},
			want: map[string]expectedSet{
				"CT Neuro": {
					name:        "CT Neuro - All Hospitals",
					memberCount: 3,
					hasWeekday:  true,
					hasWeekend:  true,
					members:     []string{"CPMC CT Neuro", "Allen CT Neuro", "NYPLH Neuro CT"},
				},
			},
		},
		{
			name: "Multiple cross-dimensional sets",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Modality:  "CT",
					Specialty: "Neuro",
					IsWeekend: false,
				},
				{
					StudyType: "Allen MRI Body",
					Modality:  "MRI",
					Specialty: "Body",
					IsWeekend: true,
				},
			},
			want: map[string]expectedSet{
				"CT Neuro": {
					name:        "CT Neuro - All Hospitals",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
				},
				"MRI Body": {
					name:        "MRI Body - All Hospitals",
					memberCount: 1,
					hasWeekday:  false,
					hasWeekend:  true,
				},
			},
		},
		{
			name: "Filters out General and Unknown specialties",
			records: []CoverageRecord{
				{
					StudyType: "CPMC CT Neuro",
					Modality:  "CT",
					Specialty: "Neuro",
					IsWeekend: false,
				},
				{
					StudyType: "CPMC US",
					Modality:  "US",
					Specialty: "General",
					IsWeekend: false,
				},
				{
					StudyType: "Random Study",
					Modality:  "Unknown",
					Specialty: "Unknown",
					IsWeekend: false,
				},
			},
			want: map[string]expectedSet{
				"CT Neuro": {
					name:        "CT Neuro - All Hospitals",
					memberCount: 1,
					hasWeekday:  true,
					hasWeekend:  false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCrossSpanningSets(tt.records)

			if len(got) != len(tt.want) {
				t.Errorf("buildCrossSpanningSets() returned %d sets, want %d", len(got), len(tt.want))
			}

			for key, expected := range tt.want {
				set, exists := got[key]
				if !exists {
					t.Errorf("buildCrossSpanningSets() missing key %q", key)
					continue
				}

				if set.Name != expected.name {
					t.Errorf("set[%q].Name = %q, want %q", key, set.Name, expected.name)
				}

				if set.MemberCount != expected.memberCount {
					t.Errorf("set[%q].MemberCount = %d, want %d", key, set.MemberCount, expected.memberCount)
				}

				if set.HasWeekday != expected.hasWeekday {
					t.Errorf("set[%q].HasWeekday = %v, want %v", key, set.HasWeekday, expected.hasWeekday)
				}

				if set.HasWeekend != expected.hasWeekend {
					t.Errorf("set[%q].HasWeekend = %v, want %v", key, set.HasWeekend, expected.hasWeekend)
				}

				if len(expected.members) > 0 {
					for _, member := range expected.members {
						if !contains(set.Members, member) {
							t.Errorf("set[%q] missing member %q", key, member)
						}
					}
				}
			}
		})
	}
}

// Test helper for edge cases
func TestBuildModalitySpanningSets_EmptyInput(t *testing.T) {
	got := buildModalitySpanningSets([]CoverageRecord{})
	if len(got) != 0 {
		t.Errorf("buildModalitySpanningSets([]) returned %d sets, want 0", len(got))
	}
}

func TestBuildSpecialtySpanningSets_EmptyInput(t *testing.T) {
	got := buildSpecialtySpanningSets([]CoverageRecord{})
	if len(got) != 0 {
		t.Errorf("buildSpecialtySpanningSets([]) returned %d sets, want 0", len(got))
	}
}

func TestBuildHospitalSpanningSets_EmptyInput(t *testing.T) {
	got := buildHospitalSpanningSets([]CoverageRecord{})
	if len(got) != 0 {
		t.Errorf("buildHospitalSpanningSets([]) returned %d sets, want 0", len(got))
	}
}

func TestBuildCrossSpanningSets_EmptyInput(t *testing.T) {
	got := buildCrossSpanningSets([]CoverageRecord{})
	if len(got) != 0 {
		t.Errorf("buildCrossSpanningSets([]) returned %d sets, want 0", len(got))
	}
}

// Helper type for test expectations
type expectedSet struct {
	name        string
	memberCount int
	hasWeekday  bool
	hasWeekend  bool
	members     []string // optional - only check if provided
}
