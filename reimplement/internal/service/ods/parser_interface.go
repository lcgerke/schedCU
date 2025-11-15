package ods

// ParsedShift represents shift data extracted from an ODS file.
type ParsedShift struct {
	ShiftType             string
	Position              string
	StartTime             string
	EndTime               string
	Location              string
	StaffMember           string
	SpecialtyConstraint   string
	StudyType             string
	RequiredQualification string
}

// ParsedSchedule represents schedule metadata extracted from an ODS file.
type ParsedSchedule struct {
	StartDate string
	EndDate   string
	Shifts    []*ParsedShift
}

// ODSParserInterface defines the contract for parsing ODS files.
// This allows the importer to work with different parser implementations.
type ODSParserInterface interface {
	// Parse reads an ODS file and extracts schedule data.
	// Returns a ParsedSchedule with all shifts, or an error if parsing fails.
	// Parsing should collect errors but continue processing when possible.
	Parse(odsContent []byte) (*ParsedSchedule, error)

	// ParseWithErrorCollection is like Parse but also returns collected errors.
	// Even if errors occurred, partial data is returned.
	ParseWithErrorCollection(odsContent []byte) (*ParsedSchedule, []error)
}
