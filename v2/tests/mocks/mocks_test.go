package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/validation"
	"github.com/schedcu/v2/tests/helpers"
)

// TestMockPersonRepository_Create verifies mock can store persons
func TestMockPersonRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPersonRepository()
	person := helpers.CreateValidPerson()

	err := repo.Create(ctx, person)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if repo.Count() != 1 {
		t.Error("expected 1 person in repository")
	}
}

// TestMockPersonRepository_GetByID verifies mock retrieves person by ID
func TestMockPersonRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPersonRepository()
	person := helpers.CreateValidPerson()

	repo.Create(ctx, person)
	retrieved, err := repo.GetByID(ctx, person.ID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved == nil {
		t.Error("expected person to be retrieved")
	}
	if retrieved.Email != person.Email {
		t.Error("expected retrieved person to match")
	}
}

// TestMockPersonRepository_GetByEmail verifies mock retrieves person by email
func TestMockPersonRepository_GetByEmail(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPersonRepository()
	email := "specific@example.com"
	person := helpers.CreateValidPersonWithEmail(email)

	repo.Create(ctx, person)
	retrieved, err := repo.GetByEmail(ctx, email)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved == nil {
		t.Error("expected person to be retrieved")
	}
}

// TestMockPersonRepository_GetAll verifies mock retrieves all persons
func TestMockPersonRepository_GetAll(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPersonRepository()

	people := helpers.BulkCreateValidPeople(5)
	for _, person := range people {
		repo.Create(ctx, person)
	}

	retrieved, err := repo.GetAll(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(retrieved) != 5 {
		t.Errorf("expected 5 persons, got %d", len(retrieved))
	}
}

// TestMockPersonRepository_Error verifies mock returns errors correctly
func TestMockPersonRepository_Error(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPersonRepository()
	testErr := errors.New("database error")

	repo.SetGetError(testErr)
	_, err := repo.GetByID(ctx, uuid.New())

	if !errors.Is(err, testErr) {
		t.Error("expected mock to return set error")
	}
}

// TestMockScheduleVersionRepository_Create verifies mock can store versions
func TestMockScheduleVersionRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewMockScheduleVersionRepository()
	version := helpers.CreateValidScheduleVersion()

	err := repo.Create(ctx, version)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if repo.Count() != 1 {
		t.Error("expected 1 version in repository")
	}
}

// TestMockScheduleVersionRepository_GetByStatus verifies mock retrieves by status
func TestMockScheduleVersionRepository_GetByStatus(t *testing.T) {
	ctx := context.Background()
	repo := NewMockScheduleVersionRepository()

	staging := helpers.CreateValidScheduleVersion()
	production := helpers.CreateValidScheduleVersionProduction()

	repo.Create(ctx, staging)
	repo.Create(ctx, production)

	retrieved, err := repo.GetByStatus(ctx, entity.VersionStatusProduction)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(retrieved) != 1 {
		t.Error("expected 1 production version")
	}
}

// TestMockScheduleVersionRepository_Update verifies mock can update versions
func TestMockScheduleVersionRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := NewMockScheduleVersionRepository()
	version := helpers.CreateValidScheduleVersion()

	repo.Create(ctx, version)

	// Promote the version
	promoterID := uuid.New()
	version.Promote(promoterID)
	err := repo.Update(ctx, version)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	retrieved, _ := repo.GetByID(ctx, version.ID)
	if retrieved.Status != entity.VersionStatusProduction {
		t.Error("expected version to be updated")
	}
}

// TestMockAssignmentRepository_Create verifies mock can store assignments
func TestMockAssignmentRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := NewMockAssignmentRepository()
	assignment := helpers.CreateValidAssignment()

	err := repo.Create(ctx, assignment)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if repo.Count() != 1 {
		t.Error("expected 1 assignment in repository")
	}
}

// TestMockAssignmentRepository_GetByPersonID verifies mock retrieves by person
func TestMockAssignmentRepository_GetByPersonID(t *testing.T) {
	ctx := context.Background()
	repo := NewMockAssignmentRepository()
	personID := uuid.New()

	assignment1 := helpers.NewAssignmentBuilder().WithPersonID(personID).Build()
	assignment2 := helpers.NewAssignmentBuilder().WithPersonID(personID).Build()
	assignment3 := helpers.NewAssignmentBuilder().Build()

	repo.Create(ctx, assignment1)
	repo.Create(ctx, assignment2)
	repo.Create(ctx, assignment3)

	retrieved, err := repo.GetByPersonID(ctx, personID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(retrieved) != 2 {
		t.Errorf("expected 2 assignments for person, got %d", len(retrieved))
	}
}

// TestMockAssignmentRepository_GetByShiftInstanceID verifies mock retrieves by shift
func TestMockAssignmentRepository_GetByShiftInstanceID(t *testing.T) {
	ctx := context.Background()
	repo := NewMockAssignmentRepository()
	shiftID := uuid.New()

	assignment1 := helpers.NewAssignmentBuilder().WithShiftInstanceID(shiftID).Build()
	assignment2 := helpers.NewAssignmentBuilder().WithShiftInstanceID(shiftID).Build()
	assignment3 := helpers.NewAssignmentBuilder().Build()

	repo.Create(ctx, assignment1)
	repo.Create(ctx, assignment2)
	repo.Create(ctx, assignment3)

	retrieved, err := repo.GetByShiftInstanceID(ctx, shiftID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(retrieved) != 2 {
		t.Errorf("expected 2 assignments for shift, got %d", len(retrieved))
	}
}

// TestMockValidationService_Validate verifies mock can validate
func TestMockValidationService_Validate(t *testing.T) {
	ctx := context.Background()
	service := NewMockValidationService()
	testInput := "test_input"

	result, err := service.Validate(ctx, testInput)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected result to be set")
	}
}

// TestMockValidationService_SetNextError verifies mock returns errors
func TestMockValidationService_SetNextError(t *testing.T) {
	ctx := context.Background()
	service := NewMockValidationService()
	testErr := errors.New("validation error")

	service.SetNextError(testErr)
	_, err := service.Validate(ctx, "test")

	if !errors.Is(err, testErr) {
		t.Error("expected mock to return set error")
	}
}

// TestMockValidationService_CallTracking verifies mock tracks calls
func TestMockValidationService_CallTracking(t *testing.T) {
	ctx := context.Background()
	service := NewMockValidationService()

	service.Validate(ctx, "input1")
	service.Validate(ctx, "input2")
	service.Validate(ctx, "input3")

	if service.GetCallCount() != 3 {
		t.Error("expected 3 calls to be tracked")
	}

	if service.GetLastInput() != "input3" {
		t.Error("expected last input to be tracked")
	}
}

// TestMockValidationService_Reset verifies mock can be reset
func TestMockValidationService_Reset(t *testing.T) {
	ctx := context.Background()
	service := NewMockValidationService()

	service.Validate(ctx, "test")
	if service.GetCallCount() != 1 {
		t.Error("expected call to be tracked")
	}

	service.Reset()
	if service.GetCallCount() != 0 {
		t.Error("expected call count to be reset")
	}
	if service.GetLastInput() != "" {
		t.Error("expected last input to be reset")
	}
}

// TestMockValidationService_SetNextResult verifies mock returns custom results
func TestMockValidationService_SetNextResult(t *testing.T) {
	ctx := context.Background()
	service := NewMockValidationService()

	customResult := validation.NewResult().
		AddError("UNKNOWN_PEOPLE", "Test error")
	service.SetNextResult(customResult)

	result, _ := service.Validate(ctx, "test")
	if !result.HasErrors() {
		t.Error("expected result to have errors")
	}
}

// TestMocks_ConcurrentAccess verifies mocks are thread-safe
func TestMocks_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPersonRepository()

	// Create 10 people concurrently
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			person := helpers.CreateValidPerson()
			done <- repo.Create(ctx, person)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	if repo.Count() != 10 {
		t.Errorf("expected 10 people, got %d", repo.Count())
	}
}

// TestMocks_Clear verifies mocks can be cleared
func TestMocks_Clear(t *testing.T) {
	ctx := context.Background()
	repo := NewMockPersonRepository()

	people := helpers.BulkCreateValidPeople(5)
	for _, person := range people {
		repo.Create(ctx, person)
	}

	if repo.Count() != 5 {
		t.Error("expected 5 people")
	}

	repo.Clear()
	if repo.Count() != 0 {
		t.Error("expected 0 people after clear")
	}
}

// BenchmarkMock_PersonRepositoryCreate benchmarks mock create
func BenchmarkMock_PersonRepositoryCreate(b *testing.B) {
	ctx := context.Background()
	repo := NewMockPersonRepository()
	for i := 0; i < b.N; i++ {
		person := helpers.CreateValidPerson()
		repo.Create(ctx, person)
	}
}

// BenchmarkMock_PersonRepositoryGetByID benchmarks mock retrieval
func BenchmarkMock_PersonRepositoryGetByID(b *testing.B) {
	ctx := context.Background()
	repo := NewMockPersonRepository()
	person := helpers.CreateValidPerson()
	repo.Create(ctx, person)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetByID(ctx, person.ID)
	}
}
