package coverage

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
)

// Example_naturalLanguageSummary demonstrates the natural language summary generation
func Example_naturalLanguageSummary() {
	// Create sample assignments (full coverage scenario)
	assignments := []entity.Assignment{
		{PersonID: entity.PersonID(uuid.New()), OriginalShiftType: "ON1"},
		{PersonID: entity.PersonID(uuid.New()), OriginalShiftType: "ON1"},
		{PersonID: entity.PersonID(uuid.New()), OriginalShiftType: "ON2"},
		{PersonID: entity.PersonID(uuid.New()), OriginalShiftType: "ON2"},
		{PersonID: entity.PersonID(uuid.New()), OriginalShiftType: "DAY"},
		{PersonID: entity.PersonID(uuid.New()), OriginalShiftType: "DAY"},
		{PersonID: entity.PersonID(uuid.New()), OriginalShiftType: "DAY"},
	}

	requirements := map[entity.ShiftType]int{
		"ON1": 2,
		"ON2": 2,
		"DAY": 3,
	}

	// Calculate coverage
	metrics := ResolveCoverage(assignments, requirements)

	// Generate natural language summary
	summary := GenerateNaturalLanguageSummary(metrics, "November 2025 Schedule")

	// Print formatted summary
	fmt.Println(summary.FormatAsText())
}
