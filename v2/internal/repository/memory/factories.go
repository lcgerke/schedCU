package memory

// Factory functions that create repositories from a shared MemoryRepository base
// These are thin wrappers around the ScheduleRepository pattern for testing

// NewHospitalRepository creates a new in-memory hospital repository (stub for compatibility)
func NewHospitalRepository(base *MemoryRepository) interface{} {
	// For now, returns a marker; actual implementation uses PostgreSQL or extends ScheduleRepository
	return &repositoryStub{name: "Hospital"}
}

// NewPersonRepository creates a new in-memory person repository (stub for compatibility)
func NewPersonRepository(base *MemoryRepository) interface{} {
	return &repositoryStub{name: "Person"}
}

// NewScheduleVersionRepository creates a new in-memory schedule version repository
func NewScheduleVersionRepository(base *MemoryRepository) *ScheduleRepository {
	// For now, reuse ScheduleRepository as it covers the schedule version pattern
	return NewScheduleRepository()
}

// NewShiftInstanceRepository creates a new in-memory shift instance repository (stub for compatibility)
func NewShiftInstanceRepository(base *MemoryRepository) interface{} {
	return &repositoryStub{name: "ShiftInstance"}
}

// NewAssignmentRepository creates a new in-memory assignment repository (stub for compatibility)
func NewAssignmentRepository(base *MemoryRepository) interface{} {
	return &repositoryStub{name: "Assignment"}
}

// NewScrapeBatchRepository creates a new in-memory scrape batch repository (stub for compatibility)
func NewScrapeBatchRepository(base *MemoryRepository) interface{} {
	return &repositoryStub{name: "ScrapeBatch"}
}

// NewCoverageCalculationRepository creates a new in-memory coverage calculation repository (stub for compatibility)
func NewCoverageCalculationRepository(base *MemoryRepository) interface{} {
	return &repositoryStub{name: "CoverageCalculation"}
}

// repositoryStub is a placeholder for future repository implementations
type repositoryStub struct {
	name string
}
