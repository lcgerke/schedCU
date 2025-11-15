package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSchedule validates schedule creation with valid inputs.
func TestNewSchedule(t *testing.T) {
	hospitalID := uuid.New()
	userID := uuid.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

	sched := NewSchedule(hospitalID, userID, startDate, endDate, "amion")

	assert.NotNil(t, sched, "Schedule should not be nil")
	assert.NotEmpty(t, sched.ID, "Schedule ID should be generated")
	assert.Equal(t, hospitalID, sched.HospitalID, "Hospital ID should match")
	assert.Equal(t, userID, sched.CreatedBy, "Created by should match")
	assert.Equal(t, "amion", sched.Source, "Source should match")
	assert.WithinDuration(t, time.Now(), sched.CreatedAt, time.Second, "Created at should be now")
	assert.Nil(t, sched.DeletedAt, "Deleted at should be nil")
	assert.Empty(t, sched.Assignments, "Assignments should be empty initially")
}

// TestScheduleValidation validates date range constraints.
func TestScheduleValidation(t *testing.T) {
	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		valid     bool
	}{
		{
			name:      "valid date range",
			startDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
			valid:     true,
		},
		{
			name:      "same day",
			startDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 1, 1, 23, 59, 59, 0, time.UTC),
			valid:     true,
		},
		{
			name:      "end before start",
			startDate: time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 1, 1, 23, 59, 59, 0, time.UTC),
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.startDate, tt.endDate)
			if tt.valid {
				assert.NoError(t, err, "Date range should be valid")
			} else {
				assert.Error(t, err, "Date range should be invalid")
			}
		})
	}
}

// TestScheduleAddAssignment validates adding assignments to a schedule.
func TestScheduleAddAssignment(t *testing.T) {
	hospitalID := uuid.New()
	userID := uuid.New()
	sched := NewSchedule(hospitalID, userID, time.Now(), time.Now().AddDate(0, 0, 1), "amion")

	assignment := map[string]interface{}{
		"id":           uuid.New(),
		"position":    "ER Doctor",
		"start_time":  "08:00",
		"end_time":    "16:00",
		"staff_member": "John Doe",
		"location":    "Main ER",
	}

	sched.AddAssignment(assignment)

	assert.Len(t, sched.Assignments, 1, "Should have one assignment")
}

// TestScheduleSoftDelete validates soft delete functionality.
func TestScheduleSoftDelete(t *testing.T) {
	hospitalID := uuid.New()
	userID := uuid.New()
	sched := NewSchedule(hospitalID, userID, time.Now(), time.Now().AddDate(0, 0, 1), "amion")

	require.Nil(t, sched.DeletedAt, "Should not be deleted initially")

	deleterID := uuid.New()
	sched.SoftDelete(deleterID)

	assert.NotNil(t, sched.DeletedAt, "Should be deleted")
	assert.Equal(t, deleterID, *sched.DeletedBy, "Deleted by should match")
	assert.True(t, sched.IsDeleted(), "IsDeleted should return true")
}

// TestScheduleUpdate validates schedule updates.
func TestScheduleUpdate(t *testing.T) {
	hospitalID := uuid.New()
	userID := uuid.New()
	sched := NewSchedule(hospitalID, userID, time.Now(), time.Now().AddDate(0, 0, 1), "amion")

	originalUpdatedAt := sched.UpdatedAt
	updaterID := uuid.New()
	newEndDate := time.Now().AddDate(0, 0, 7)

	time.Sleep(10 * time.Millisecond) // Ensure time difference
	sched.UpdatedAt = time.Now()
	sched.UpdatedBy = updaterID
	sched.EndDate = newEndDate

	assert.Equal(t, updaterID, sched.UpdatedBy, "Updated by should match")
	assert.Equal(t, newEndDate, sched.EndDate, "End date should be updated")
	assert.True(t, sched.UpdatedAt.After(originalUpdatedAt), "Updated at should be newer")
}
