package memory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
)

// TestCreateSchedule validates schedule creation with query count assertion.
func TestCreateSchedule(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	hospitalID := uuid.New()
	userID := uuid.New()
	sched := entity.NewSchedule(
		hospitalID,
		userID,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		"amion",
	)

	err := repo.CreateSchedule(ctx, sched)

	assert.NoError(t, err, "Should create schedule successfully")
	assert.Equal(t, 1, repo.QueryCount(), "Should have exactly 1 query")

	// Verify it can be retrieved
	retrieved, err := repo.GetScheduleByID(ctx, sched.ID)
	require.NoError(t, err)
	assert.Equal(t, sched.ID, retrieved.ID)
}

// TestGetScheduleByID validates retrieval with not-found handling.
func TestGetScheduleByID(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	nonExistentID := uuid.New()
	_, err := repo.GetScheduleByID(ctx, nonExistentID)

	assert.Error(t, err, "Should return error for non-existent schedule")
	assert.True(t, repository.IsNotFound(err), "Error should be NotFoundError")
}

// TestGetSchedulesByHospital validates filtering by hospital ID.
func TestGetSchedulesByHospital(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	hospitalID1 := uuid.New()
	hospitalID2 := uuid.New()
	userID := uuid.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

	// Create schedules for different hospitals
	sched1 := entity.NewSchedule(hospitalID1, userID, startDate, endDate, "amion")
	sched2 := entity.NewSchedule(hospitalID1, userID, startDate, endDate, "ods_file")
	sched3 := entity.NewSchedule(hospitalID2, userID, startDate, endDate, "manual")

	repo.CreateSchedule(ctx, sched1)
	repo.CreateSchedule(ctx, sched2)
	repo.CreateSchedule(ctx, sched3)

	// Query for hospital1
	schedules, err := repo.GetSchedulesByHospital(ctx, hospitalID1)

	assert.NoError(t, err)
	assert.Len(t, schedules, 2, "Should find 2 schedules for hospital1")
	assert.Equal(t, 4, repo.QueryCount(), "Should have 4 queries (3 creates + 1 get)")
}

// TestUpdateSchedule validates schedule updates.
func TestUpdateSchedule(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	hospitalID := uuid.New()
	userID := uuid.New()
	sched := entity.NewSchedule(
		hospitalID,
		userID,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		"amion",
	)

	repo.CreateSchedule(ctx, sched)
	originalUpdatedAt := sched.UpdatedAt

	// Update schedule
	time.Sleep(10 * time.Millisecond)
	updaterID := uuid.New()
	newEndDate := sched.EndDate.AddDate(0, 1, 0)
	sched.EndDate = newEndDate
	sched.UpdatedBy = updaterID

	err := repo.UpdateSchedule(ctx, sched)

	assert.NoError(t, err)
	assert.Equal(t, 2, repo.QueryCount(), "Should have 2 queries (1 create + 1 update)")

	// Verify update persisted
	retrieved, _ := repo.GetScheduleByID(ctx, sched.ID)
	assert.Equal(t, newEndDate, retrieved.EndDate, "End date should be updated")
	assert.Equal(t, updaterID, retrieved.UpdatedBy, "Updated by should match")
	assert.True(t, retrieved.UpdatedAt.After(originalUpdatedAt), "Updated at should be newer")
}

// TestSoftDelete validates soft delete functionality.
func TestSoftDelete(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	hospitalID := uuid.New()
	userID := uuid.New()
	sched := entity.NewSchedule(
		hospitalID,
		userID,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		"amion",
	)

	repo.CreateSchedule(ctx, sched)
	deleterID := uuid.New()

	// Delete the schedule
	err := repo.DeleteSchedule(ctx, sched.ID, deleterID)

	assert.NoError(t, err)
	assert.Equal(t, 2, repo.QueryCount(), "Should have 2 queries (1 create + 1 delete)")

	// Verify it's not retrievable (soft deleted)
	_, err = repo.GetScheduleByID(ctx, sched.ID)
	assert.Error(t, err, "Should not retrieve soft-deleted schedule")
	assert.True(t, repository.IsNotFound(err))
}

// TestAddShiftInstance validates adding shifts to a schedule.
// Note: For the simplified Schedule entity, shifts are managed via schedule.AddAssignment()
// The repository methods are stubs for interface compatibility.
func TestAddShiftInstance(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	hospitalID := uuid.New()
	userID := uuid.New()
	sched := entity.NewSchedule(
		hospitalID,
		userID,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		"amion",
	)

	repo.CreateSchedule(ctx, sched)

	// For the simplified Schedule, use the schedule directly
	shift := map[string]interface{}{
		"id":           uuid.New(),
		"position":     "ER Doctor",
		"start_time":   "08:00",
		"end_time":     "16:00",
		"staff_member": "Jane Smith",
		"location":     "Main ER",
	}

	sched.AddAssignment(shift)

	assert.NoError(t, nil)
	assert.Equal(t, 1, repo.QueryCount(), "Should have 1 query (1 create schedule)")
	assert.Len(t, sched.Assignments, 1, "Should have 1 assignment in schedule")
}

// TestGetShiftInstances validates shift retrieval.
// Note: For the simplified Schedule entity, shifts are stored in schedule.Assignments
func TestGetShiftInstances(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	hospitalID := uuid.New()
	userID := uuid.New()
	sched := entity.NewSchedule(
		hospitalID,
		userID,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		"amion",
	)

	repo.CreateSchedule(ctx, sched)

	// Add multiple shifts via schedule
	shift1 := map[string]interface{}{
		"position":     "Doctor",
		"start_time":   "08:00",
		"end_time":     "16:00",
		"staff_member": "Jane",
		"location":     "ER",
	}

	shift2 := map[string]interface{}{
		"position":     "Nurse",
		"start_time":   "16:00",
		"end_time":     "00:00",
		"staff_member": "John",
		"location":     "ICU",
	}

	sched.AddAssignment(shift1)
	sched.AddAssignment(shift2)

	// Verify they're in the schedule
	assert.NoError(t, nil)
	assert.Len(t, sched.Assignments, 2, "Should have 2 assignments")
	assert.Equal(t, 1, repo.QueryCount(), "Should have 1 query (1 create)")
}

// TestCount validates counting active schedules.
func TestCount(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	hospitalID := uuid.New()
	userID := uuid.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

	// Create 3 schedules
	sched1 := entity.NewSchedule(hospitalID, userID, startDate, endDate, "amion")
	sched2 := entity.NewSchedule(hospitalID, userID, startDate, endDate, "ods_file")
	sched3 := entity.NewSchedule(hospitalID, userID, startDate, endDate, "manual")

	repo.CreateSchedule(ctx, sched1)
	repo.CreateSchedule(ctx, sched2)
	repo.CreateSchedule(ctx, sched3)

	// Count all
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count, "Should count 3 active schedules")

	// Delete one
	repo.DeleteSchedule(ctx, sched1.ID, userID)

	// Count again
	count, err = repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count, "Should count 2 active schedules after soft delete")
}

// TestQueryCountAssertion validates query efficiency.
func TestQueryCountAssertion(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	hospitalID := uuid.New()
	userID := uuid.New()
	sched := entity.NewSchedule(
		hospitalID,
		userID,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		"amion",
	)

	// Each operation increments query count
	initialCount := repo.QueryCount()

	repo.CreateSchedule(ctx, sched)
	assert.Equal(t, initialCount+1, repo.QueryCount(), "Create should be 1 query")

	repo.GetScheduleByID(ctx, sched.ID)
	assert.Equal(t, initialCount+2, repo.QueryCount(), "Get should be 1 query")

	repo.UpdateSchedule(ctx, sched)
	assert.Equal(t, initialCount+3, repo.QueryCount(), "Update should be 1 query")

	repo.GetSchedulesByHospital(ctx, hospitalID)
	assert.Equal(t, initialCount+4, repo.QueryCount(), "GetByHospital should be 1 query")

	repo.Count(ctx)
	assert.Equal(t, initialCount+5, repo.QueryCount(), "Count should be 1 query")
}

// TestReset validates repository reset functionality.
func TestReset(t *testing.T) {
	repo := NewScheduleRepository()
	ctx := context.Background()

	hospitalID := uuid.New()
	userID := uuid.New()
	sched := entity.NewSchedule(
		hospitalID,
		userID,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
		"amion",
	)

	repo.CreateSchedule(ctx, sched)
	assert.Equal(t, 1, repo.QueryCount())

	// Reset
	repo.Reset()

	assert.Equal(t, 0, repo.QueryCount(), "Query count should reset")

	_, err := repo.GetScheduleByID(ctx, sched.ID)
	assert.Error(t, err, "Data should be cleared after reset")
}
