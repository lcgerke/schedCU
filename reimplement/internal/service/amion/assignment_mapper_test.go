package amion

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/repository"
)

// TestAssignmentMapperSuccessfulMapping tests successful mapping when shift instance is found.
func TestAssignmentMapperSuccessfulMapping(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create a shift instance that will be "found"
	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		StartTime:         nil,
		EndTime:           nil,
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if assignment == nil {
		t.Fatalf("Expected assignment to be non-nil")
	}

	if assignment.PersonID != personID {
		t.Errorf("Expected PersonID %s, got %s", personID, assignment.PersonID)
	}

	if assignment.ShiftInstanceID != shiftInstanceID {
		t.Errorf("Expected ShiftInstanceID %s, got %s", shiftInstanceID, assignment.ShiftInstanceID)
	}

	if assignment.Source != entity.AssignmentSourceAmion {
		t.Errorf("Expected Source 'AMION', got %s", assignment.Source)
	}

	if assignment.CreatedBy != userID {
		t.Errorf("Expected CreatedBy %s, got %s", userID, assignment.CreatedBy)
	}

	if assignment.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be set, got zero time")
	}

	if assignment.OriginalShiftType != "Technologist" {
		t.Errorf("Expected OriginalShiftType 'Technologist', got %s", assignment.OriginalShiftType)
	}
}

// TestAssignmentMapperShiftInstanceNotFound tests error when shift instance is nil.
func TestAssignmentMapperShiftInstanceNotFound(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: make(map[uuid.UUID]*entity.ShiftInstance),
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, nil, scheduleVersionID, userID, mockRepo)

	if err == nil {
		t.Errorf("Expected error for missing shift instance, got nil")
	}

	if assignment != nil {
		t.Errorf("Expected assignment to be nil on error, got %v", assignment)
	}
}

// TestAssignmentMapperDeletedShiftInstance tests error when shift instance is soft-deleted.
func TestAssignmentMapperDeletedShiftInstance(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create a deleted shift instance
	deletedTime := time.Now()
	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
		DeletedAt:         &deletedTime,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

	if err == nil {
		t.Errorf("Expected error for deleted shift instance, got nil")
	}

	if assignment != nil {
		t.Errorf("Expected assignment to be nil on error, got %v", assignment)
	}
}

// TestAssignmentMapperNilPersonID tests error when person ID is nil.
func TestAssignmentMapperNilPersonID(t *testing.T) {
	mapper := NewAssignmentMapper()

	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, uuid.Nil, shiftInstance, scheduleVersionID, userID, mockRepo)

	if err == nil {
		t.Errorf("Expected error for nil person ID, got nil")
	}

	if assignment != nil {
		t.Errorf("Expected assignment to be nil on error, got %v", assignment)
	}
}

// TestAssignmentMapperNilScheduleVersionID tests error when schedule version ID is nil.
func TestAssignmentMapperNilScheduleVersionID(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: uuid.Nil,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, uuid.Nil, userID, mockRepo)

	if err == nil {
		t.Errorf("Expected error for nil schedule version ID, got nil")
	}

	if assignment != nil {
		t.Errorf("Expected assignment to be nil on error, got %v", assignment)
	}
}

// TestAssignmentMapperNilUserID tests error when user ID is nil.
func TestAssignmentMapperNilUserID(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         uuid.New(),
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, uuid.Nil, mockRepo)

	if err == nil {
		t.Errorf("Expected error for nil user ID, got nil")
	}

	if assignment != nil {
		t.Errorf("Expected assignment to be nil on error, got %v", assignment)
	}
}

// TestAssignmentMapperTimestampsSet tests that timestamps are set correctly.
func TestAssignmentMapperTimestampsSet(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	beforeTime := time.Now()

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

	afterTime := time.Now()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify CreatedAt is within reasonable time range
	if assignment.CreatedAt.Before(beforeTime) || assignment.CreatedAt.After(afterTime.Add(1*time.Second)) {
		t.Errorf("CreatedAt %v is not within expected range [%v, %v]", assignment.CreatedAt, beforeTime, afterTime)
	}

	// Verify DeletedAt is nil for new assignment
	if assignment.DeletedAt != nil {
		t.Errorf("Expected DeletedAt to be nil, got %v", assignment.DeletedAt)
	}
}

// TestAssignmentMapperSourceSetToAmion tests that source is set to AMION.
func TestAssignmentMapperSourceSetToAmion(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if assignment.Source != entity.AssignmentSourceAmion {
		t.Errorf("Expected Source 'AMION', got %s", assignment.Source)
	}
}

// TestAssignmentMapperDateParsing tests correct date parsing from RawAmionShift.
func TestAssignmentMapperDateParsing(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Parse expected date
	expectedDate, _ := time.Parse("2006-01-02", "2025-11-20")

	if !assignment.ScheduleDate.Equal(expectedDate) {
		t.Errorf("Expected ScheduleDate %v, got %v", expectedDate, assignment.ScheduleDate)
	}
}

// TestAssignmentMapperPreservesOriginalShiftType tests that original shift type is preserved.
func TestAssignmentMapperPreservesOriginalShiftType(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Senior Radiologist",
		Location:  "Read Room A",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if assignment.OriginalShiftType != "Senior Radiologist" {
		t.Errorf("Expected OriginalShiftType 'Senior Radiologist', got %s", assignment.OriginalShiftType)
	}
}

// TestAssignmentMapperMultipleShifts tests mapping multiple shifts independently.
func TestAssignmentMapperMultipleShifts(t *testing.T) {
	mapper := NewAssignmentMapper()

	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create multiple shift instances
	shifts := make(map[uuid.UUID]*entity.ShiftInstance)
	for i := 0; i < 3; i++ {
		shiftID := uuid.New()
		shifts[shiftID] = &entity.ShiftInstance{
			ID:                shiftID,
			ScheduleVersionID: scheduleVersionID,
			ShiftType:         "Morning",
			Position:          "Senior Doctor",
			Location:          "Main Lab",
			StaffMember:       "John Doe",
			CreatedAt:         time.Now(),
			CreatedBy:         userID,
		}
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: shifts,
	}

	// Map each shift
	for shiftID, shiftInstance := range shifts {
		personID := uuid.New()
		raw := RawAmionShift{
			Date:      "2025-11-20",
			ShiftType: "Technologist",
			Location:  "Main Lab",
			RowIndex:  1,
		}

		assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

		if err != nil {
			t.Fatalf("Expected no error for shift %s, got: %v", shiftID, err)
		}

		if assignment.ShiftInstanceID != shiftID {
			t.Errorf("Expected ShiftInstanceID %s, got %s", shiftID, assignment.ShiftInstanceID)
		}
	}
}

// TestAssignmentMapperBatchProcessing tests mapping multiple shifts from scraper batch.
func TestAssignmentMapperBatchProcessing(t *testing.T) {
	mapper := NewAssignmentMapper()

	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create shift instances from batch scraping
	shifts := make(map[uuid.UUID]*entity.ShiftInstance)
	shiftIDs := make([]uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		shiftID := uuid.New()
		shiftIDs[i] = shiftID
		shifts[shiftID] = &entity.ShiftInstance{
			ID:                shiftID,
			ScheduleVersionID: scheduleVersionID,
			ShiftType:         "Morning",
			Position:          "Senior Doctor",
			Location:          "Main Lab",
			StaffMember:       "John Doe",
			CreatedAt:         time.Now(),
			CreatedBy:         userID,
		}
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: shifts,
	}

	// Simulate batch scraping results
	rawShifts := []RawAmionShift{
		{
			Date:      "2025-11-20",
			ShiftType: "Technologist",
			Location:  "Main Lab",
			RowIndex:  1,
		},
		{
			Date:      "2025-11-21",
			ShiftType: "Radiologist",
			Location:  "Read Room A",
			RowIndex:  2,
		},
		{
			Date:      "2025-11-22",
			ShiftType: "Senior Technologist",
			Location:  "Main Lab",
			RowIndex:  3,
		},
	}

	assignments := make([]*entity.Assignment, 0)
	for i, raw := range rawShifts {
		personID := uuid.New()
		shiftInstance := shifts[shiftIDs[i]]

		assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

		if err != nil {
			t.Fatalf("Expected no error for batch item %d, got: %v", i, err)
		}

		assignments = append(assignments, assignment)
	}

	// Verify all assignments were created
	if len(assignments) != len(rawShifts) {
		t.Errorf("Expected %d assignments, got %d", len(rawShifts), len(assignments))
	}

	// Verify each assignment has correct source and original shift type
	for i, assignment := range assignments {
		if assignment.Source != entity.AssignmentSourceAmion {
			t.Errorf("Assignment %d: Expected Source 'AMION', got %s", i, assignment.Source)
		}

		if assignment.OriginalShiftType != rawShifts[i].ShiftType {
			t.Errorf("Assignment %d: Expected OriginalShiftType '%s', got %s", i, rawShifts[i].ShiftType, assignment.OriginalShiftType)
		}
	}
}

// TestAssignmentMapperInvalidDateFormat tests error on invalid date format.
func TestAssignmentMapperInvalidDateFormat(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	invalidDates := []string{
		"2025/11/20",    // Wrong format
		"11-20-2025",    // Wrong format
		"invalid-date",  // Not a date
		"",              // Empty
		"2025-13-01",    // Invalid month
		"2025-02-30",    // Invalid day
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	for _, invalidDate := range invalidDates {
		t.Run("invalid_"+invalidDate, func(t *testing.T) {
			raw := RawAmionShift{
				Date:      invalidDate,
				ShiftType: "Technologist",
				Location:  "Main Lab",
				RowIndex:  1,
			}

			assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

			if err == nil {
				t.Errorf("Expected error for invalid date %q, got nil", invalidDate)
			}

			if assignment != nil {
				t.Errorf("Expected assignment to be nil on error, got %v", assignment)
			}
		})
	}
}

// TestAssignmentMapperIDGeneration tests that new IDs are generated.
func TestAssignmentMapperIDGeneration(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify ID is generated (not nil)
	if assignment.ID == uuid.Nil {
		t.Errorf("Expected ID to be generated, got nil UUID")
	}

	// Verify assignment is valid
	if !assignment.IsValid() {
		t.Errorf("Expected assignment to be valid after mapping")
	}
}

// TestAssignmentMapperValidation tests the IsValid method on created assignments.
func TestAssignmentMapperValidation(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !assignment.IsValid() {
		t.Errorf("Expected assignment to be valid, but IsValid() returned false")
	}

	if assignment.IsDeleted() {
		t.Errorf("Expected assignment to not be deleted, but IsDeleted() returned true")
	}
}

// TestAssignmentMapperScheduleVersionMismatch tests error when shift belongs to different schedule.
func TestAssignmentMapperScheduleVersionMismatch(t *testing.T) {
	mapper := NewAssignmentMapper()

	personID := uuid.New()
	shiftInstanceID := uuid.New()
	scheduleVersionID := uuid.New()
	otherScheduleVersionID := uuid.New()
	userID := uuid.New()

	shiftInstance := &entity.ShiftInstance{
		ID:                shiftInstanceID,
		ScheduleVersionID: otherScheduleVersionID, // Different schedule version
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	mockRepo := &MockShiftInstanceRepository{
		shifts: map[uuid.UUID]*entity.ShiftInstance{
			shiftInstanceID: shiftInstance,
		},
	}

	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)

	if err == nil {
		t.Errorf("Expected error for schedule version mismatch, got nil")
	}

	if assignment != nil {
		t.Errorf("Expected assignment to be nil on error, got %v", assignment)
	}
}

// MockShiftInstanceRepository is a test mock for ShiftInstanceRepository.
type MockShiftInstanceRepository struct {
	shifts map[uuid.UUID]*entity.ShiftInstance
}

var _ repository.ShiftInstanceRepository = (*MockShiftInstanceRepository)(nil)

func (m *MockShiftInstanceRepository) Create(ctx context.Context, shift *entity.ShiftInstance) (*entity.ShiftInstance, error) {
	m.shifts[shift.ID] = shift
	return shift, nil
}

func (m *MockShiftInstanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ShiftInstance, error) {
	if shift, ok := m.shifts[id]; ok {
		return shift, nil
	}
	return nil, fmt.Errorf("shift not found")
}

func (m *MockShiftInstanceRepository) GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.ShiftInstance, error) {
	var shifts []*entity.ShiftInstance
	for _, shift := range m.shifts {
		if shift.ScheduleVersionID == scheduleVersionID {
			shifts = append(shifts, shift)
		}
	}
	return shifts, nil
}

func (m *MockShiftInstanceRepository) CreateBatch(ctx context.Context, shifts []*entity.ShiftInstance) (int, error) {
	for _, shift := range shifts {
		m.shifts[shift.ID] = shift
	}
	return len(shifts), nil
}

func (m *MockShiftInstanceRepository) Update(ctx context.Context, shift *entity.ShiftInstance) error {
	m.shifts[shift.ID] = shift
	return nil
}

func (m *MockShiftInstanceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.shifts[id]; ok {
		delete(m.shifts, id)
		return nil
	}
	return fmt.Errorf("shift not found")
}

func (m *MockShiftInstanceRepository) DeleteByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) (int, error) {
	count := 0
	for _, shift := range m.shifts {
		if shift.ScheduleVersionID == scheduleVersionID {
			count++
		}
	}
	return count, nil
}
