package amion

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Helper function to create a document from HTML string
func docFromHTML(html string) (*goquery.Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	return doc, err
}

// Test 1: Extract single shift from minimal HTML
func TestExtractShifts_SingleShift(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 1 {
		t.Errorf("Expected 1 shift, got %d", result.ShiftCount())
	}

	shift := result.Shifts[0]
	if shift.Date != "2025-11-15" {
		t.Errorf("Expected date '2025-11-15', got '%s'", shift.Date)
	}
	if shift.ShiftType != "Technologist" {
		t.Errorf("Expected type 'Technologist', got '%s'", shift.ShiftType)
	}
	if shift.StartTime != "07:00" {
		t.Errorf("Expected start time '07:00', got '%s'", shift.StartTime)
	}
	if shift.EndTime != "15:00" {
		t.Errorf("Expected end time '15:00', got '%s'", shift.EndTime)
	}
	if shift.Location != "Main Lab" {
		t.Errorf("Expected location 'Main Lab', got '%s'", shift.Location)
	}
}

// Test 2: Extract multiple shifts
func TestExtractShifts_MultipleShifts(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
				<tr>
					<td>2025-11-16</td>
					<td>Technologist</td>
					<td>08:00</td>
					<td>16:00</td>
					<td>Main Lab</td>
				</tr>
				<tr>
					<td>2025-11-17</td>
					<td>Radiologist</td>
					<td>07:00</td>
					<td>19:00</td>
					<td>Read Room A</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 3 {
		t.Errorf("Expected 3 shifts, got %d", result.ShiftCount())
	}

	expectedDates := []string{"2025-11-15", "2025-11-16", "2025-11-17"}
	for i, expectedDate := range expectedDates {
		if result.Shifts[i].Date != expectedDate {
			t.Errorf("Shift %d: expected date '%s', got '%s'", i, expectedDate, result.Shifts[i].Date)
		}
	}
}

// Test 3: Handle whitespace in cells
func TestExtractShifts_WhitespaceHandling(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>  2025-11-15  </td>
					<td>
						Technologist
					</td>
					<td>  07:00  </td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 1 {
		t.Errorf("Expected 1 shift, got %d", result.ShiftCount())
	}

	shift := result.Shifts[0]
	if shift.Date != "2025-11-15" {
		t.Errorf("Expected date '2025-11-15', got '%s'", shift.Date)
	}
	if shift.ShiftType != "Technologist" {
		t.Errorf("Expected type 'Technologist', got '%s'", shift.ShiftType)
	}
}

// Test 4: Skip empty rows
func TestExtractShifts_SkipEmptyRows(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
				<tr>
					<td></td>
					<td></td>
					<td></td>
					<td></td>
					<td></td>
				</tr>
				<tr>
					<td>2025-11-16</td>
					<td>Radiologist</td>
					<td>08:00</td>
					<td>16:00</td>
					<td>Read Room A</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 2 {
		t.Errorf("Expected 2 shifts (empty row skipped), got %d", result.ShiftCount())
	}
}

// Test 5: Skip header rows with <th> elements
func TestExtractShifts_SkipHeaderRows(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<thead>
				<tr>
					<th>Date</th>
					<th>Position</th>
					<th>Start</th>
					<th>End</th>
					<th>Location</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 1 {
		t.Errorf("Expected 1 shift (header skipped), got %d", result.ShiftCount())
	}
}

// Test 6: Missing required field (date)
func TestExtractShifts_MissingDate(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td></td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 0 {
		t.Errorf("Expected 0 shifts (missing date), got %d", result.ShiftCount())
	}

	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount())
	}

	if result.Errors[0].Field != "date" {
		t.Errorf("Expected error field 'date', got '%s'", result.Errors[0].Field)
	}
}

// Test 7: Missing required field (shift type)
func TestExtractShifts_MissingShiftType(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td></td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 0 {
		t.Errorf("Expected 0 shifts (missing shift type), got %d", result.ShiftCount())
	}

	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount())
	}
}

// Test 8: Missing required field (start time)
func TestExtractShifts_MissingStartTime(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td></td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 0 {
		t.Errorf("Expected 0 shifts (missing start time), got %d", result.ShiftCount())
	}

	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount())
	}
}

// Test 9: Missing required field (end time)
func TestExtractShifts_MissingEndTime(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td></td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 0 {
		t.Errorf("Expected 0 shifts (missing end time), got %d", result.ShiftCount())
	}

	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount())
	}
}

// Test 10: Missing optional field (location) - should NOT fail
func TestExtractShifts_MissingLocation(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td></td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 1 {
		t.Errorf("Expected 1 shift (location is optional), got %d", result.ShiftCount())
	}

	shift := result.Shifts[0]
	if shift.Location != "" {
		t.Errorf("Expected empty location, got '%s'", shift.Location)
	}
}

// Test 11: Invalid required staffing value (non-integer) - should extract shift but record error
func TestExtractShifts_InvalidStaffing(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
					<td>not-a-number</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 1 {
		t.Errorf("Expected 1 shift (staffing error is non-critical), got %d", result.ShiftCount())
	}

	// Should have error for invalid staffing
	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error (invalid staffing), got %d", result.ErrorCount())
	}

	if result.Errors[0].Field != "required_staffing" {
		t.Errorf("Expected error field 'required_staffing', got '%s'", result.Errors[0].Field)
	}
}

// Test 12: Valid required staffing value
func TestExtractShifts_ValidStaffing(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
					<td>3</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 1 {
		t.Errorf("Expected 1 shift, got %d", result.ShiftCount())
	}

	shift := result.Shifts[0]
	if shift.RequiredStaffing != 3 {
		t.Errorf("Expected required staffing 3, got %d", shift.RequiredStaffing)
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Expected no errors, got %d", result.ErrorCount())
	}
}

// Test 13: Extract shifts for specific month
func TestExtractShiftsForMonth(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
				<tr>
					<td>2025-11-20</td>
					<td>Radiologist</td>
					<td>08:00</td>
					<td>16:00</td>
					<td>Read Room A</td>
				</tr>
				<tr>
					<td>2025-12-01</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
				<tr>
					<td>2025-12-15</td>
					<td>Radiologist</td>
					<td>08:00</td>
					<td>16:00</td>
					<td>Read Room A</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	novemberResult := ExtractShiftsForMonth(doc, "2025-11")
	decemberResult := ExtractShiftsForMonth(doc, "2025-12")

	if novemberResult.ShiftCount() != 2 {
		t.Errorf("Expected 2 November shifts, got %d", novemberResult.ShiftCount())
	}

	if decemberResult.ShiftCount() != 2 {
		t.Errorf("Expected 2 December shifts, got %d", decemberResult.ShiftCount())
	}

	// Verify November shifts
	if novemberResult.Shifts[0].Date != "2025-11-15" {
		t.Errorf("Expected first November shift on 2025-11-15, got %s", novemberResult.Shifts[0].Date)
	}
	if novemberResult.Shifts[1].Date != "2025-11-20" {
		t.Errorf("Expected second November shift on 2025-11-20, got %s", novemberResult.Shifts[1].Date)
	}

	// Verify December shifts
	if decemberResult.Shifts[0].Date != "2025-12-01" {
		t.Errorf("Expected first December shift on 2025-12-01, got %s", decemberResult.Shifts[0].Date)
	}
	if decemberResult.Shifts[1].Date != "2025-12-15" {
		t.Errorf("Expected second December shift on 2025-12-15, got %s", decemberResult.Shifts[1].Date)
	}
}

// Test 14: Empty table
func TestExtractShifts_EmptyTable(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 0 {
		t.Errorf("Expected 0 shifts (empty table), got %d", result.ShiftCount())
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Expected no errors, got %d", result.ErrorCount())
	}
}

// Test 15: Table with no tbody element (should still work with nested tr elements)
func TestExtractShifts_NoTbodyElement(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tr>
				<td>2025-11-15</td>
				<td>Technologist</td>
				<td>07:00</td>
				<td>15:00</td>
				<td>Main Lab</td>
			</tr>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 1 {
		t.Errorf("Expected 1 shift, got %d", result.ShiftCount())
	}
}

// Test 16: Cell references are set correctly
func TestExtractShifts_CellReferences(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	shift := result.Shifts[0]

	if shift.RowIndex != 1 {
		t.Errorf("Expected row index 1, got %d", shift.RowIndex)
	}

	if !strings.Contains(shift.DateCell, "row 1") {
		t.Errorf("Expected DateCell to contain 'row 1', got '%s'", shift.DateCell)
	}

	if !strings.Contains(shift.ShiftTypeCell, "column 2") {
		t.Errorf("Expected ShiftTypeCell to contain 'column 2', got '%s'", shift.ShiftTypeCell)
	}
}

// Test 17: Multiple errors in single extraction
func TestExtractShifts_MultipleErrors(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td></td>
					<td></td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
					<td>bad</td>
				</tr>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 1 {
		t.Errorf("Expected 1 valid shift, got %d", result.ShiftCount())
	}

	// First row has 3 missing critical fields (date, shift type, end time) and 1 invalid staffing
	if result.ErrorCount() != 4 {
		t.Errorf("Expected 4 errors, got %d", result.ErrorCount())
	}

	if result.CriticalErrorCount() != 3 {
		t.Errorf("Expected 3 critical errors, got %d", result.CriticalErrorCount())
	}
}

// Test 18: FormattedErrors provides readable output
func TestExtractShifts_FormattedErrors(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td></td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	formatted := result.FormattedErrors()

	if !strings.Contains(formatted, "Row 1") {
		t.Errorf("Expected formatted errors to contain 'Row 1', got: %s", formatted)
	}

	if !strings.Contains(formatted, "date") {
		t.Errorf("Expected formatted errors to contain 'date', got: %s", formatted)
	}
}

// Test 19: Large batch with 90 shifts (Spike 1 test scenario)
func TestExtractShifts_LargeBatch(t *testing.T) {
	// Generate 90 shifts (6 months × 15 shifts per month)
	var htmlRows strings.Builder
	htmlRows.WriteString(`<html><body><table><tbody>`)

	for i := 0; i < 90; i++ {
		day := (i % 30) + 1
		month := (i / 15) + 11 // November (11) or December (12)
		year := 2025

		date := ""
		if month > 12 {
			month = month - 12
			year = 2026
		}

		date = formatDate(year, month, day)
		position := "Technologist"
		if i%2 == 0 {
			position = "Radiologist"
		}

		startTime := "07:00"
		endTime := "15:00"
		if i%3 == 0 {
			startTime = "08:00"
			endTime = "16:00"
		}

		location := "Main Lab"
		if i%4 == 0 {
			location = "Read Room A"
		}

		htmlRows.WriteString(`<tr><td>` + date + `</td><td>` + position + `</td><td>` + startTime + `</td><td>` + endTime + `</td><td>` + location + `</td></tr>`)
	}

	htmlRows.WriteString(`</tbody></table></body></html>`)

	doc, _ := docFromHTML(htmlRows.String())
	result := ExtractShifts(doc)

	if result.ShiftCount() != 90 {
		t.Errorf("Expected 90 shifts in large batch, got %d", result.ShiftCount())
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Expected no errors in large batch, got %d", result.ErrorCount())
	}
}

// Test 20: Custom selectors work
func TestExtractShifts_CustomSelectors(t *testing.T) {
	html := `
	<html>
	<body>
		<div class="shift-table">
			<div class="shift-row">
				<span class="cell-1">2025-11-15</span>
				<span class="cell-2">Technologist</span>
				<span class="cell-3">07:00</span>
				<span class="cell-4">15:00</span>
				<span class="cell-5">Main Lab</span>
			</div>
		</div>
	</body>
	</html>
	`

	customSelectors := &AmionSelectors{
		ShiftRowSelector:            ".shift-row",
		DateCellSelector:            ".cell-1",
		ShiftTypeCellSelector:       ".cell-2",
		StartTimeCellSelector:       ".cell-3",
		EndTimeCellSelector:         ".cell-4",
		LocationCellSelector:        ".cell-5",
		RequiredStaffingCellSelector: ".cell-6",
	}

	doc, _ := docFromHTML(html)
	result := ExtractShiftsWithSelectors(doc, customSelectors)

	if result.ShiftCount() != 1 {
		t.Errorf("Expected 1 shift with custom selectors, got %d", result.ShiftCount())
	}

	shift := result.Shifts[0]
	if shift.Date != "2025-11-15" {
		t.Errorf("Expected date '2025-11-15', got '%s'", shift.Date)
	}
}

// Test 21: Extract shifts with optional staffing across different valid formats
func TestExtractShifts_VariousStaffingFormats(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
					<td>3</td>
				</tr>
				<tr>
					<td>2025-11-16</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
					<td>0</td>
				</tr>
				<tr>
					<td>2025-11-17</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
					<td></td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 3 {
		t.Errorf("Expected 3 shifts, got %d", result.ShiftCount())
	}

	if result.Shifts[0].RequiredStaffing != 3 {
		t.Errorf("Expected staffing 3 for shift 1, got %d", result.Shifts[0].RequiredStaffing)
	}

	if result.Shifts[1].RequiredStaffing != 0 {
		t.Errorf("Expected staffing 0 for shift 2, got %d", result.Shifts[1].RequiredStaffing)
	}

	if result.Shifts[2].RequiredStaffing != 0 {
		t.Errorf("Expected staffing 0 for shift 3 (empty), got %d", result.Shifts[2].RequiredStaffing)
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Expected no errors, got %d", result.ErrorCount())
	}
}

// Test 22: Real-world Amion HTML structure with thead and tbody
func TestExtractShifts_RealAmionStructure(t *testing.T) {
	html := `
	<html>
	<body>
		<table class="amion-schedule">
			<thead>
				<tr>
					<th>Date</th>
					<th>Position</th>
					<th>Start Time</th>
					<th>End Time</th>
					<th>Location</th>
					<th>Required Staff</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
					<td>2</td>
				</tr>
				<tr>
					<td>2025-11-15</td>
					<td>Radiologist</td>
					<td>08:00</td>
					<td>17:00</td>
					<td>Read Room A</td>
					<td>1</td>
				</tr>
				<tr>
					<td>2025-11-16</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
					<td>2</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if result.ShiftCount() != 3 {
		t.Errorf("Expected 3 shifts, got %d", result.ShiftCount())
	}

	if result.ErrorCount() != 0 {
		t.Errorf("Expected no errors, got %d: %s", result.ErrorCount(), result.FormattedErrors())
	}

	// Verify specific shifts
	if result.Shifts[0].ShiftType != "Technologist" {
		t.Errorf("Expected first shift to be Technologist, got %s", result.Shifts[0].ShiftType)
	}

	if result.Shifts[1].Location != "Read Room A" {
		t.Errorf("Expected second shift location to be Read Room A, got %s", result.Shifts[1].Location)
	}
}

// Helper function to format date
func formatDate(year, month, day int) string {
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

// Test 23: HasErrors and ErrorCount methods
func TestExtractionResult_HelperMethods(t *testing.T) {
	html := `
	<html>
	<body>
		<table>
			<tbody>
				<tr>
					<td></td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
				<tr>
					<td>2025-11-15</td>
					<td>Technologist</td>
					<td>07:00</td>
					<td>15:00</td>
					<td>Main Lab</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>
	`

	doc, _ := docFromHTML(html)
	result := ExtractShifts(doc)

	if !result.HasErrors() {
		t.Error("Expected HasErrors to return true")
	}

	if result.ErrorCount() != 1 {
		t.Errorf("Expected ErrorCount to return 1, got %d", result.ErrorCount())
	}

	if result.ShiftCount() != 1 {
		t.Errorf("Expected ShiftCount to return 1, got %d", result.ShiftCount())
	}

	if result.CriticalErrorCount() != 1 {
		t.Errorf("Expected CriticalErrorCount to return 1, got %d", result.CriticalErrorCount())
	}
}
