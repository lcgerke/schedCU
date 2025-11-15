package ods

// RawShiftData represents shift information extracted from an ODS file.
// All fields are kept as strings to allow the service layer to handle
// type conversion, validation, and standardization.
type RawShiftData struct {
	// Date is the shift date as a string (to be parsed by service layer)
	// Expected format: "YYYY-MM-DD" (e.g., "2025-11-15")
	Date string

	// ShiftType is the type of shift (e.g., "DAY", "NIGHT", "WEEKEND")
	ShiftType string

	// RequiredStaffing is the number of staff required, as a string
	// Expected to be parseable as a positive integer
	RequiredStaffing string

	// SpecialtyConstraint is an optional specialty requirement (e.g., "Radiology", "Lab")
	SpecialtyConstraint string

	// StudyType is an optional study type constraint (e.g., "CT", "MRI")
	StudyType string

	// RowMetadata contains information about where this shift data came from
	RowMetadata RowMetadata
}

// RowMetadata contains information about where a row came from in the ODS file.
type RowMetadata struct {
	// Row is the 1-based row number in the ODS file
	Row int

	// CellReference is the Excel-style cell reference (e.g., "A5")
	CellReference string

	// RowData is the raw string data from all cells in this row
	RowData []string
}

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
type ODSParserInterface interface {
	// Parse reads an ODS file and extracts schedule data.
	Parse(odsContent []byte) (*ParsedSchedule, error)

	// ParseWithErrorCollection parses with error collection.
	ParseWithErrorCollection(odsContent []byte) (*ParsedSchedule, []error)
}
