package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
)

// ScheduleRepository is an in-memory implementation for testing.
type ScheduleRepository struct {
	mu        sync.RWMutex
	schedules map[uuid.UUID]*entity.Schedule
	queryCount int
}

// NewScheduleRepository creates a new in-memory schedule repository.
func NewScheduleRepository() *ScheduleRepository {
	return &ScheduleRepository{
		schedules: make(map[uuid.UUID]*entity.Schedule),
	}
}

// CreateSchedule stores a new schedule.
func (r *ScheduleRepository) CreateSchedule(ctx context.Context, schedule *entity.Schedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queryCount++

	if schedule == nil {
		return &repository.NotFoundError{ResourceType: "Schedule", ResourceID: "nil"}
	}

	r.schedules[schedule.ID] = schedule
	return nil
}

// GetScheduleByID retrieves a schedule by ID (excluding soft-deleted).
func (r *ScheduleRepository) GetScheduleByID(ctx context.Context, id uuid.UUID) (*entity.Schedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.queryCount++

	schedule, exists := r.schedules[id]
	if !exists {
		return nil, &repository.NotFoundError{ResourceType: "Schedule", ResourceID: id.String()}
	}

	if schedule.IsDeleted() {
		return nil, &repository.NotFoundError{ResourceType: "Schedule", ResourceID: id.String()}
	}

	return schedule, nil
}

// GetSchedulesByHospital retrieves all active schedules for a hospital.
func (r *ScheduleRepository) GetSchedulesByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.Schedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.queryCount++

	var result []*entity.Schedule
	for _, schedule := range r.schedules {
		if schedule.HospitalID == hospitalID && !schedule.IsDeleted() {
			result = append(result, schedule)
		}
	}

	return result, nil
}

// UpdateSchedule updates an existing schedule.
func (r *ScheduleRepository) UpdateSchedule(ctx context.Context, schedule *entity.Schedule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queryCount++

	if schedule == nil {
		return &repository.NotFoundError{ResourceType: "Schedule", ResourceID: "nil"}
	}

	_, exists := r.schedules[schedule.ID]
	if !exists {
		return &repository.NotFoundError{ResourceType: "Schedule", ResourceID: schedule.ID.String()}
	}

	schedule.UpdatedAt = time.Now().UTC()
	r.schedules[schedule.ID] = schedule

	return nil
}

// DeleteSchedule performs a soft delete.
func (r *ScheduleRepository) DeleteSchedule(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queryCount++

	schedule, exists := r.schedules[id]
	if !exists {
		return &repository.NotFoundError{ResourceType: "Schedule", ResourceID: id.String()}
	}

	schedule.SoftDelete(deleterID)
	r.schedules[id] = schedule

	return nil
}

// GetShiftInstances retrieves all shifts for a schedule.
// For the simplified Schedule entity, shifts are stored in schedule.Assignments
// This method is a stub for compatibility with the repository interface
func (r *ScheduleRepository) GetShiftInstances(ctx context.Context, scheduleID uuid.UUID) ([]*entity.ShiftInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.queryCount++

	// For the simplified Schedule, we don't track shifts separately
	// This is handled by the full v1 schema (ScheduleVersion) in production
	return []*entity.ShiftInstance{}, nil
}

// AddShiftInstance adds a new shift to a schedule.
// For the simplified Schedule entity, use schedule.AddAssignment directly
// This method is a stub for compatibility with the repository interface
func (r *ScheduleRepository) AddShiftInstance(ctx context.Context, shift *entity.ShiftInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queryCount++

	if shift == nil {
		return &repository.NotFoundError{ResourceType: "ShiftInstance", ResourceID: "nil"}
	}

	// For the simplified Schedule, assignments are managed through the Schedule directly
	return nil
}

// Count returns the total number of active schedules.
func (r *ScheduleRepository) Count(ctx context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.queryCount++

	count := int64(0)
	for _, schedule := range r.schedules {
		if !schedule.IsDeleted() {
			count++
		}
	}

	return count, nil
}

// QueryCount returns the number of queries executed (for testing purposes).
func (r *ScheduleRepository) QueryCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.queryCount
}

// Reset clears all data and resets query count.
func (r *ScheduleRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schedules = make(map[uuid.UUID]*entity.Schedule)
	r.queryCount = 0
}
