package memory

import "sync"

// MemoryRepository is a shared in-memory store for all entity types
type MemoryRepository struct {
	mu sync.RWMutex

	// Data stores for each entity type
	hospitals            map[string]interface{}
	persons              map[string]interface{}
	scheduleVersions     map[string]interface{}
	shiftInstances       map[string]interface{}
	assignments          map[string]interface{}
	scrapeBatches        map[string]interface{}
	coverageCalculations map[string]interface{}
	auditLogs            map[string]interface{}
	users                map[string]interface{}
	jobQueues            map[string]interface{}
}

// NewMemoryRepository creates a new empty in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		hospitals:            make(map[string]interface{}),
		persons:              make(map[string]interface{}),
		scheduleVersions:     make(map[string]interface{}),
		shiftInstances:       make(map[string]interface{}),
		assignments:          make(map[string]interface{}),
		scrapeBatches:        make(map[string]interface{}),
		coverageCalculations: make(map[string]interface{}),
		auditLogs:            make(map[string]interface{}),
		users:                make(map[string]interface{}),
		jobQueues:            make(map[string]interface{}),
	}
}
