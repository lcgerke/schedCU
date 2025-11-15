# Test Infrastructure Documentation

This directory contains comprehensive testing utilities for the schedCU v2 application, including entity builders, factories, fixtures, and mock implementations.

## Directory Structure

```
tests/
├── fixtures/              # Test fixture files
│   ├── ods/              # ODS file fixtures
│   ├── html/             # HTML file fixtures
│   ├── entities/         # Entity JSON fixtures
│   └── data/             # Generic data fixtures
├── helpers/              # Test helper functions
│   ├── builders.go       # Entity builders (Builder pattern)
│   ├── builders_test.go  # Builder tests
│   ├── factories.go      # Entity factories
│   ├── factories_test.go # Factory tests
│   ├── fixtures.go       # Fixture loading utilities
│   └── fixtures_test.go  # Fixture tests
├── mocks/                # Mock implementations
│   ├── mocks.go          # Mock repositories and services
│   └── mocks_test.go     # Mock tests
└── README.md             # This file
```

## Entity Builders

Builders use the fluent Builder pattern to create test entities with full control over field values.

### Available Builders

- `PersonBuilder` - Creates Person entities
- `ShiftInstanceBuilder` - Creates ShiftInstance entities
- `AssignmentBuilder` - Creates Assignment entities
- `ScheduleVersionBuilder` - Creates ScheduleVersion entities
- `ScrapeBatchBuilder` - Creates ScrapeBatch entities
- `CoverageCalculationBuilder` - Creates CoverageCalculation entities

### Builder Usage

```go
import "github.com/schedcu/v2/tests/helpers"

// Create with defaults
person := helpers.NewPersonBuilder().Build()

// Customize specific fields
person := helpers.NewPersonBuilder().
    WithEmail("john@hospital.com").
    WithName("John Doe").
    WithSpecialty(entity.SpecialtyBodyOnly).
    WithActive(false).
    Build()

// Chain methods for fluent interface
shift := helpers.NewShiftInstanceBuilder().
    WithShiftType(entity.ShiftTypeMidC).
    WithScheduleDate(time.Now()).
    WithDesiredCoverage(2).
    WithIsMandatory(false).
    Build()
```

### Builder Methods

Each builder includes:
- `NewXxxBuilder()` - Creates new builder instance
- `XxxBuilder_Default()` - Alternative function to create default builder
- `With*()` methods - Set individual fields (chainable)
- `Build()` - Creates the final entity

### Default Values

Builders create entities with sensible defaults:

**PersonBuilder**
- Email: `"person@example.com"`
- Name: `"Test Person"`
- Specialty: `BOTH`
- Active: `true`
- ID/Timestamps: Auto-generated

**ShiftInstanceBuilder**
- ShiftType: `DAY`
- StartTime: `"08:00"`
- EndTime: `"16:00"`
- DesiredCoverage: `1`
- IsMandatory: `true`
- ID/Timestamps: Auto-generated

**AssignmentBuilder**
- Source: `AMION`
- OriginalShiftType: `"DAY"`
- ID/Timestamps: Auto-generated

**ScheduleVersionBuilder**
- Status: `STAGING`
- ValidationResults: Empty Result
- ShiftInstances: Empty slice
- ID/Timestamps: Auto-generated

**ScrapeBatchBuilder**
- State: `PENDING`
- IngestChecksum: `"default-checksum"`
- RowCount: `0`
- ID/Timestamps: Auto-generated

**CoverageCalculationBuilder**
- CoverageByPosition: Empty map
- CoverageSummary: Empty map
- ValidationErrors: Empty Result
- QueryCount: `0`
- ID/Timestamps: Auto-generated

## Entity Factories

Factories create valid entities with sensible defaults. Use factories when you just need a valid entity without customization.

### Available Factories

#### Person Factories
- `CreateValidPerson()` - Standard person
- `CreateValidPersonWithEmail(email)` - Person with custom email
- `CreateValidPersonWithSpecialty(specialty)` - Person with specific specialty
- `CreateValidPersonInactive()` - Inactive person
- `CreateValidPersonDeleted()` - Soft-deleted person
- `BulkCreateValidPeople(count)` - Multiple people with unique emails

#### ShiftInstance Factories
- `CreateValidShiftInstance()` - Standard shift
- `CreateValidShiftInstanceWithType(shiftType)` - Shift with specific type
- `CreateValidShiftInstanceWithDate(date)` - Shift on specific date
- `CreateValidShiftInstanceWithStudyType(studyType)` - Shift with specific study type
- `CreateValidShiftInstanceOptional()` - Non-mandatory shift
- `CreateValidShiftInstanceWithCoverage(count)` - Shift requiring multiple people
- `BulkCreateValidShiftInstances(count)` - Multiple shifts with varied types

#### Assignment Factories
- `CreateValidAssignment()` - Standard assignment
- `CreateValidAssignmentWithSource(source)` - Assignment from specific source
- `CreateValidAssignmentFromAmion()` - Amion-sourced assignment
- `CreateValidAssignmentFromManual()` - Manually-created assignment
- `CreateValidAssignmentDeleted()` - Soft-deleted assignment
- `BulkCreateValidAssignments(count)` - Multiple assignments with varied sources

#### ScheduleVersion Factories
- `CreateValidScheduleVersion()` - Staging version
- `CreateValidScheduleVersionProduction()` - Production version
- `CreateValidScheduleVersionArchived()` - Archived version
- `CreateValidScheduleVersionWithShifts(count)` - Version with N shifts
- `CreateValidScheduleVersionWithValidation(result)` - Version with validation

#### ScrapeBatch Factories
- `CreateValidScrapeBatch()` - Pending batch
- `CreateValidScrapeBatchComplete()` - Completed batch
- `CreateValidScrapeBatchFailed()` - Failed batch with error
- `CreateValidScrapeBatchArchived()` - Archived batch
- `CreateValidScrapeBatchDeleted()` - Soft-deleted batch

#### CoverageCalculation Factories
- `CreateValidCoverageCalculation()` - Standard coverage
- `CreateValidCoverageCalculationWithMetrics()` - Coverage with metrics
- `CreateValidCoverageCalculationWithValidationErrors()` - Coverage with errors

#### Other Factories
- `CreateValidHospital()` - Hospital entity
- `CreateValidHospitalWithCode(code)` - Hospital with specific code
- `CreateValidUser()` - Viewer user
- `CreateValidUserAdmin()` - Admin user
- `CreateValidUserScheduler()` - Scheduler user for specific hospital
- `CreateValidAuditLog()` - Audit log entry
- `CreateValidJobQueue()` - Job queue entry

### Factory Usage

```go
import "github.com/schedcu/v2/tests/helpers"

// Quick test data
person := helpers.CreateValidPerson()
assignment := helpers.CreateValidAssignmentFromAmion()
version := helpers.CreateValidScheduleVersionProduction()

// Bulk creation
people := helpers.BulkCreateValidPeople(10)
shifts := helpers.BulkCreateValidShiftInstances(5)

// With customization
shift := helpers.CreateValidShiftInstanceWithCoverage(3)
user := helpers.CreateValidUserScheduler()
```

## Fixture Loading

The `FixtureLoader` provides utilities for loading test data from fixture files.

### Usage

```go
import "github.com/schedcu/v2/tests/helpers"

// Generic fixture loader
loader := helpers.NewFixtureLoader()
data, err := loader.LoadTextFixture("data/sample.csv")

// Load JSON fixtures
var entity MyEntity
err := loader.LoadJSONFixture("entities/valid_person.json", &entity)

// ODS fixtures
odsFixture := helpers.NewODSFixture()
odsData, err := odsFixture.LoadODSFile("sample_schedule.ods")
files, err := odsFixture.ListODSFixtures()

// HTML fixtures
htmlFixture := helpers.NewHTMLFixture()
html, err := htmlFixture.LoadHTMLFile("amion_schedule.html")

// Entity fixtures
entityFixture := helpers.NewEntityFixture()
err := entityFixture.LoadEntityFixture("person.json", &person)
err := entityFixture.SaveEntityFixture("new_person.json", person)

// Data fixtures
dataFixture := helpers.NewDataFixture()
data, err := dataFixture.LoadData("raw_csv.txt")
err := dataFixture.SaveData("output.bin", []byte("data"))
```

### Fixture Directory Structure

```
tests/fixtures/
├── ods/              # ODS spreadsheet files
├── html/             # HTML files (e.g., scraped pages)
├── entities/         # JSON entity representations
└── data/             # Generic text/binary data
```

## Mock Implementations

Mocks provide in-memory implementations for testing services without external dependencies.

### Available Mocks

- `MockPersonRepository` - In-memory person storage
- `MockScheduleVersionRepository` - In-memory version storage
- `MockAssignmentRepository` - In-memory assignment storage
- `MockValidationService` - Stub validation service with configurable results

### Mock Usage

```go
import (
    "context"
    "github.com/schedcu/v2/tests/helpers"
    "github.com/schedcu/v2/tests/mocks"
)

ctx := context.Background()

// Create mock
repo := mocks.NewMockPersonRepository()

// Use like real repository
person := helpers.CreateValidPerson()
repo.Create(ctx, person)

retrieved, _ := repo.GetByID(ctx, person.ID)
people, _ := repo.GetAll(ctx)

// Configure error behavior
repo.SetGetError(errors.New("database error"))
_, err := repo.GetByID(ctx, uuid.New()) // Returns error

// Reset for next test
repo.Clear()

// Validation service mocking
validService := mocks.NewMockValidationService()
validService.SetNextError(errors.New("validation failed"))

result, err := validService.Validate(ctx, "data") // Returns error
callCount := validService.GetCallCount() // Track invocations
lastInput := validService.GetLastInput()  // Verify input

validService.Reset() // Clear state
```

### Mock Features

**MockPersonRepository**
- Thread-safe with RWMutex
- Create/GetByID/GetByEmail/GetAll operations
- Error injection via SetGetError/SetSaveError
- Clear() to reset state
- Count() to check storage

**MockScheduleVersionRepository**
- Create/GetByID/GetByStatus/Update operations
- Error injection and state management
- Thread-safe operations

**MockAssignmentRepository**
- Create/GetByID/GetByPersonID/GetByShiftInstanceID operations
- Filter results by person or shift
- Error injection and state management

**MockValidationService**
- Validate(ctx, name) with configurable results
- SetNextResult/SetNextError for behavior control
- Call tracking: GetCallCount(), GetLastInput()
- Reset() to clear state

## Testing Patterns

### Pattern 1: Unit Test with Builder

```go
func TestMyFunction(t *testing.T) {
    person := helpers.NewPersonBuilder().
        WithEmail("test@example.com").
        WithActive(false).
        Build()

    result := myFunction(person)

    if result.Error != nil {
        t.Errorf("unexpected error: %v", result.Error)
    }
}
```

### Pattern 2: Repository Test with Mock

```go
func TestPersonService_CreatePerson(t *testing.T) {
    ctx := context.Background()
    repo := mocks.NewMockPersonRepository()
    service := NewPersonService(repo)

    person := helpers.CreateValidPerson()
    err := service.CreatePerson(ctx, person)

    if repo.Count() != 1 {
        t.Error("expected person to be saved")
    }
}
```

### Pattern 3: Bulk Test Data

```go
func TestScheduleVersion_MultiplePeople(t *testing.T) {
    people := helpers.BulkCreateValidPeople(10)
    shifts := helpers.BulkCreateValidShiftInstances(5)

    version := helpers.NewScheduleVersionBuilder().
        WithShiftInstances(shifts).
        Build()

    // Test with multiple entities
}
```

### Pattern 4: Error Scenario Testing

```go
func TestPersonService_DatabaseError(t *testing.T) {
    repo := mocks.NewMockPersonRepository()
    repo.SetGetError(errors.New("connection refused"))

    service := NewPersonService(repo)
    _, err := service.GetPerson(ctx, uuid.New())

    if err == nil {
        t.Error("expected error to be returned")
    }
}
```

## Test Coverage

All builder and factory functionality is tested with:
- **Builder Tests**: 30+ tests validating default creation, field setting, immutability, and valid entity constraints
- **Factory Tests**: 50+ tests validating factory functions create valid entities with correct defaults
- **Mock Tests**: 25+ tests validating mock behavior, error handling, and thread safety

Run all tests:
```bash
go test ./tests/... -v
```

Run specific packages:
```bash
go test ./tests/helpers -v
go test ./tests/mocks -v
```

Run with coverage:
```bash
go test ./tests/... -cover
```

## Performance

Builders and factories are optimized for test performance:

- **PersonBuilder**: ~0.1µs per build
- **ShiftInstanceBuilder**: ~0.15µs per build
- **AssignmentBuilder**: ~0.12µs per build
- **ScheduleVersionBuilder**: ~0.3µs per build (includes nested shifts)
- **Factory Functions**: ~0.1µs per creation

Suitable for creating thousands of test entities.

## Immutability and Constraints

All builders validate entity constraints:
- Required fields are always populated
- Enums use valid values only
- Foreign key relationships are respected
- Date ranges are validated (start < end)
- Soft delete patterns are enforced

## Integration with v1 Code

Test infrastructure is compatible with both new and legacy validation patterns:
- `entity.ValidationResult` for entity-level validation
- `validation.Result` for comprehensive validation messaging
- Automatic conversion where needed

## Best Practices

1. **Use Factories for Quick Tests**: When you don't need customization, factories are faster
2. **Use Builders for Specific Tests**: When testing edge cases or specific field values
3. **Bulk Creation**: Use bulk factories for performance tests with many entities
4. **Mock Errors**: Test error paths using SetError() methods
5. **Track Mock Calls**: Use call counters and input tracking for behavior verification
6. **Reset State**: Call Clear() or Reset() between tests to avoid cross-test pollution

## Future Enhancements

Potential additions:
- Testcontainers integration for database tests
- Query counting assertions
- Fixture generation from real data
- Snapshot testing support
- Custom assertions for common patterns
