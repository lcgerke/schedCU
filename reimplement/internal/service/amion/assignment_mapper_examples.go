package amion

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/repository"
)

// ExampleAssignmentMapperBasicUsage demonstrates basic usage of the AssignmentMapper.
func ExampleAssignmentMapperBasicUsage() {
	// Create mapper
	mapper := NewAssignmentMapper()

	// Create mock IDs for this example
	personID := uuid.New()
	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create a shift instance from earlier processing
	shiftInstance := &entity.ShiftInstance{
		ID:                uuid.New(),
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		StaffMember:       "John Doe",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	// Raw data from Amion scraper
	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Senior Radiologist",
		Location:  "Read Room A",
		RowIndex:  1,
	}

	// Create a mock repository (in production, use actual repository)
	mockRepo := &MockExampleRepository{}

	// Map to assignment
	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)
	if err != nil {
		log.Fatalf("Failed to map assignment: %v", err)
	}

	// Use the assignment
	fmt.Printf("Created assignment: ID=%s, PersonID=%s, Source=%s\n",
		assignment.ID.String()[:8],
		assignment.PersonID.String()[:8],
		assignment.Source)
}

// ExampleAssignmentMapperBatchMapping demonstrates mapping multiple shifts from scraper batch.
func ExampleAssignmentMapperBatchMapping() {
	mapper := NewAssignmentMapper()

	scheduleVersionID := uuid.New()
	userID := uuid.New()

	// Create shift instances (from [1.11] batch scraping)
	shiftInstances := []*entity.ShiftInstance{
		{
			ID:                uuid.New(),
			ScheduleVersionID: scheduleVersionID,
			ShiftType:         "Morning",
			Position:          "Senior Doctor",
			Location:          "Main Lab",
			CreatedAt:         time.Now(),
			CreatedBy:         userID,
		},
		{
			ID:                uuid.New(),
			ScheduleVersionID: scheduleVersionID,
			ShiftType:         "Afternoon",
			Position:          "Radiologist",
			Location:          "Read Room A",
			CreatedAt:         time.Now(),
			CreatedBy:         userID,
		},
	}

	// Scraped shifts from Amion
	rawShifts := []RawAmionShift{
		{
			Date:      "2025-11-20",
			ShiftType: "Senior Technologist",
			Location:  "Main Lab",
			RowIndex:  1,
		},
		{
			Date:      "2025-11-21",
			ShiftType: "Radiologist",
			Location:  "Read Room A",
			RowIndex:  2,
		},
	}

	// Person IDs (from [1.3] Person creation)
	personIDs := []uuid.UUID{uuid.New(), uuid.New()}

	mockRepo := &MockExampleRepository{}

	// Map each shift
	assignments := make([]*entity.Assignment, 0)
	for i, raw := range rawShifts {
		assignment, err := mapper.MapToAssignment(
			context.Background(),
			raw,
			personIDs[i],
			shiftInstances[i],
			scheduleVersionID,
			userID,
			mockRepo,
		)
		if err != nil {
			log.Fatalf("Failed to map shift %d: %v", i, err)
		}
		assignments = append(assignments, assignment)
	}

	fmt.Printf("Mapped %d assignments from Amion batch\n", len(assignments))
}

// ExampleAssignmentMapperErrorHandling demonstrates error handling patterns.
func ExampleAssignmentMapperErrorHandling() {
	mapper := NewAssignmentMapper()

	scheduleVersionID := uuid.New()
	userID := uuid.New()

	mockRepo := &MockExampleRepository{}

	// Case 1: Shift instance not found
	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Technologist",
		Location:  "Main Lab",
		RowIndex:  1,
	}

	_, err := mapper.MapToAssignment(context.Background(), raw, uuid.New(), nil, scheduleVersionID, userID, mockRepo)
	if err != nil {
		fmt.Printf("Error Case 1 (shift not found): %v\n", err)
	}

	// Case 2: Invalid date format
	raw.Date = "invalid-date"
	shiftInstance := &entity.ShiftInstance{
		ID:                uuid.New(),
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	_, err = mapper.MapToAssignment(context.Background(), raw, uuid.New(), shiftInstance, scheduleVersionID, userID, mockRepo)
	if err != nil {
		fmt.Printf("Error Case 2 (invalid date): %v\n", err)
	}

	// Case 3: Deleted shift instance
	deletedTime := time.Now()
	shiftInstance.DeletedAt = &deletedTime
	raw.Date = "2025-11-20"

	_, err = mapper.MapToAssignment(context.Background(), raw, uuid.New(), shiftInstance, scheduleVersionID, userID, mockRepo)
	if err != nil {
		fmt.Printf("Error Case 3 (deleted shift): %v\n", err)
	}
}

// ExampleAssignmentMapperWithRepository demonstrates integration with repository.
func ExampleAssignmentMapperWithRepository() {
	// In production, use actual repository implementation
	mapper := NewAssignmentMapper()

	scheduleVersionID := uuid.New()
	personID := uuid.New()
	userID := uuid.New()

	// Create shift instance
	shiftInstance := &entity.ShiftInstance{
		ID:                uuid.New(),
		ScheduleVersionID: scheduleVersionID,
		ShiftType:         "Morning",
		Position:          "Senior Doctor",
		Location:          "Main Lab",
		CreatedAt:         time.Now(),
		CreatedBy:         userID,
	}

	// Raw shift from Amion
	raw := RawAmionShift{
		Date:      "2025-11-20",
		ShiftType: "Senior Radiologist",
		Location:  "Read Room A",
		RowIndex:  1,
	}

	// Map to assignment
	mockRepo := &MockExampleRepository{}
	assignment, err := mapper.MapToAssignment(context.Background(), raw, personID, shiftInstance, scheduleVersionID, userID, mockRepo)
	if err != nil {
		log.Fatalf("Failed to map: %v", err)
	}

	// In production, save to repository
	// savedAssignment, err := assignmentRepo.Create(context.Background(), assignment)
	// if err != nil {
	//     handle constraint violations
	// }

	fmt.Printf("Assignment ready to persist: ID=%s, Source=%s\n", assignment.ID.String()[:8], assignment.Source)
}

// MockExampleRepository is a mock repository for examples.
type MockExampleRepository struct{}

var _ repository.ShiftInstanceRepository = (*MockExampleRepository)(nil)

func (m *MockExampleRepository) Create(ctx context.Context, shift *entity.ShiftInstance) (*entity.ShiftInstance, error) {
	return shift, nil
}

func (m *MockExampleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ShiftInstance, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockExampleRepository) GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.ShiftInstance, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockExampleRepository) CreateBatch(ctx context.Context, shifts []*entity.ShiftInstance) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockExampleRepository) Update(ctx context.Context, shift *entity.ShiftInstance) error {
	return fmt.Errorf("not implemented")
}

func (m *MockExampleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (m *MockExampleRepository) DeleteByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) (int, error) {
	return 0, fmt.Errorf("not implemented")
}
