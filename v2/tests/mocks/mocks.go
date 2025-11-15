package mocks

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/validation"
)

// MockPersonRepository is a mock implementation of PersonRepository for testing
type MockPersonRepository struct {
	mu      sync.RWMutex
	people  map[uuid.UUID]*entity.Person
	getErr  error
	saveErr error
}

// NewMockPersonRepository creates a new mock person repository
func NewMockPersonRepository() *MockPersonRepository {
	return &MockPersonRepository{
		people: make(map[uuid.UUID]*entity.Person),
	}
}

// Create stores a person (mock implementation)
func (m *MockPersonRepository) Create(ctx context.Context, person *entity.Person) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	m.people[person.ID] = person
	return nil
}

// GetByID retrieves a person by ID (mock implementation)
func (m *MockPersonRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Person, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	if person, ok := m.people[id]; ok {
		return person, nil
	}
	return nil, nil
}

// GetByEmail retrieves a person by email (mock implementation)
func (m *MockPersonRepository) GetByEmail(ctx context.Context, email string) (*entity.Person, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, person := range m.people {
		if person.Email == email {
			return person, nil
		}
	}
	return nil, nil
}

// GetAll retrieves all people (mock implementation)
func (m *MockPersonRepository) GetAll(ctx context.Context) ([]*entity.Person, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	var people []*entity.Person
	for _, person := range m.people {
		people = append(people, person)
	}
	return people, nil
}

// SetGetError sets the error to return from Get operations
func (m *MockPersonRepository) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getErr = err
}

// SetSaveError sets the error to return from Create operations
func (m *MockPersonRepository) SetSaveError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveErr = err
}

// Count returns the number of stored people
func (m *MockPersonRepository) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.people)
}

// Clear removes all stored people
func (m *MockPersonRepository) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.people = make(map[uuid.UUID]*entity.Person)
}

// MockScheduleVersionRepository is a mock implementation of ScheduleVersionRepository
type MockScheduleVersionRepository struct {
	mu        sync.RWMutex
	versions  map[uuid.UUID]*entity.ScheduleVersion
	getErr    error
	saveErr   error
	updateErr error
}

// NewMockScheduleVersionRepository creates a new mock schedule version repository
func NewMockScheduleVersionRepository() *MockScheduleVersionRepository {
	return &MockScheduleVersionRepository{
		versions: make(map[uuid.UUID]*entity.ScheduleVersion),
	}
}

// Create stores a schedule version
func (m *MockScheduleVersionRepository) Create(ctx context.Context, version *entity.ScheduleVersion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	m.versions[version.ID] = version
	return nil
}

// GetByID retrieves a schedule version by ID
func (m *MockScheduleVersionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ScheduleVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	if version, ok := m.versions[id]; ok {
		return version, nil
	}
	return nil, nil
}

// GetByStatus retrieves schedule versions by status
func (m *MockScheduleVersionRepository) GetByStatus(ctx context.Context, status entity.VersionStatus) ([]*entity.ScheduleVersion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	var versions []*entity.ScheduleVersion
	for _, version := range m.versions {
		if version.Status == status {
			versions = append(versions, version)
		}
	}
	return versions, nil
}

// Update updates a schedule version
func (m *MockScheduleVersionRepository) Update(ctx context.Context, version *entity.ScheduleVersion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.updateErr != nil {
		return m.updateErr
	}
	m.versions[version.ID] = version
	return nil
}

// SetGetError sets the error to return from Get operations
func (m *MockScheduleVersionRepository) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getErr = err
}

// SetSaveError sets the error to return from Create operations
func (m *MockScheduleVersionRepository) SetSaveError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveErr = err
}

// SetUpdateError sets the error to return from Update operations
func (m *MockScheduleVersionRepository) SetUpdateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateErr = err
}

// Count returns the number of stored versions
func (m *MockScheduleVersionRepository) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.versions)
}

// Clear removes all stored versions
func (m *MockScheduleVersionRepository) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.versions = make(map[uuid.UUID]*entity.ScheduleVersion)
}

// MockAssignmentRepository is a mock implementation of AssignmentRepository
type MockAssignmentRepository struct {
	mu        sync.RWMutex
	assignments map[uuid.UUID]*entity.Assignment
	getErr    error
	saveErr   error
}

// NewMockAssignmentRepository creates a new mock assignment repository
func NewMockAssignmentRepository() *MockAssignmentRepository {
	return &MockAssignmentRepository{
		assignments: make(map[uuid.UUID]*entity.Assignment),
	}
}

// Create stores an assignment
func (m *MockAssignmentRepository) Create(ctx context.Context, assignment *entity.Assignment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveErr != nil {
		return m.saveErr
	}
	m.assignments[assignment.ID] = assignment
	return nil
}

// GetByID retrieves an assignment by ID
func (m *MockAssignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	if assignment, ok := m.assignments[id]; ok {
		return assignment, nil
	}
	return nil, nil
}

// GetByPersonID retrieves all assignments for a person
func (m *MockAssignmentRepository) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]*entity.Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	var assignments []*entity.Assignment
	for _, assignment := range m.assignments {
		if assignment.PersonID == personID {
			assignments = append(assignments, assignment)
		}
	}
	return assignments, nil
}

// GetByShiftInstanceID retrieves all assignments for a shift
func (m *MockAssignmentRepository) GetByShiftInstanceID(ctx context.Context, shiftInstanceID uuid.UUID) ([]*entity.Assignment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	var assignments []*entity.Assignment
	for _, assignment := range m.assignments {
		if assignment.ShiftInstanceID == shiftInstanceID {
			assignments = append(assignments, assignment)
		}
	}
	return assignments, nil
}

// SetGetError sets the error to return from Get operations
func (m *MockAssignmentRepository) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getErr = err
}

// SetSaveError sets the error to return from Create operations
func (m *MockAssignmentRepository) SetSaveError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveErr = err
}

// Count returns the number of stored assignments
func (m *MockAssignmentRepository) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.assignments)
}

// Clear removes all stored assignments
func (m *MockAssignmentRepository) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.assignments = make(map[uuid.UUID]*entity.Assignment)
}

// MockValidationService is a mock implementation of a validation service
type MockValidationService struct {
	mu            sync.RWMutex
	nextResult    *validation.Result
	nextErr       error
	callCount     int
	lastInputName string
}

// NewMockValidationService creates a new mock validation service
func NewMockValidationService() *MockValidationService {
	return &MockValidationService{
		nextResult: validation.NewResult(),
		callCount:  0,
	}
}

// Validate validates something and returns a result
func (m *MockValidationService) Validate(ctx context.Context, name string) (*validation.Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	m.lastInputName = name
	return m.nextResult, m.nextErr
}

// SetNextResult sets the result to return from Validate
func (m *MockValidationService) SetNextResult(result *validation.Result) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextResult = result
}

// SetNextError sets the error to return from Validate
func (m *MockValidationService) SetNextError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextErr = err
}

// GetCallCount returns the number of times Validate was called
func (m *MockValidationService) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount
}

// GetLastInput returns the last input to Validate
func (m *MockValidationService) GetLastInput() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastInputName
}

// Reset resets the mock state
func (m *MockValidationService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount = 0
	m.lastInputName = ""
	m.nextResult = validation.NewResult()
	m.nextErr = nil
}
