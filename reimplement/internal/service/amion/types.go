package amion

// RawAmionShift represents a raw shift extracted from Amion HTML
// before any parsing or validation. It includes cell references
// for error reporting and debugging.
type RawAmionShift struct {
	Date              string // YYYY-MM-DD format
	ShiftType         string // Position/Role (e.g., "Technologist", "Radiologist")
	RequiredStaffing  int    // Number of staff required
	StartTime         string // HH:MM format
	EndTime           string // HH:MM format
	Location          string // Physical location (e.g., "Main Lab", "Read Room A")
	RowIndex          int    // For error reporting: which row in the table
	DateCell          string // Cell reference: row X, column 1
	ShiftTypeCell     string // Cell reference: row X, column 2
	StartTimeCell     string // Cell reference: row X, column 3
	EndTimeCell       string // Cell reference: row X, column 4
	LocationCell      string // Cell reference: row X, column 5
	RequiredStaffCell string // Cell reference: row X, column 6 (if present)
}

// ExtractionError represents an error during shift extraction
type ExtractionError struct {
	RowIndex int
	Field    string
	Value    string
	Reason   string
}

// ExtractionResult holds both successful extractions and errors
type ExtractionResult struct {
	Shifts []RawAmionShift
	Errors []ExtractionError
}
