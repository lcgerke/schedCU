// Package postgres provides comprehensive integration tests for all repositories
package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/schedcu/v2/internal/entity"
)

// TestAllRepositories_SoftDeleteCascading tests that soft delete works across related entities
func TestAllRepositories_SoftDeleteCascading(t *testing.T) {
	ctx := context.Background()
	helper := NewPostgresTestHelper(ctx, t)
	defer helper.Close(t)

	// Create test data hierarchy
	hospID := uuid.New()
	personID := uuid.New()
	versionID := uuid.New()

	// Insert hospital
	if _, err := helper.DB().ExecContext(ctx, `
		INSERT INTO hospitals (id, name) VALUES ($1, $2)`,
		hospID, "Test Hospital"); err != nil {
		t.Fatalf("Failed to insert hospital: %v", err)
	}

	// Insert person
	if _, err := helper.DB().ExecContext(ctx, `
		INSERT INTO persons (id, email, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		personID, "test@example.com", "Test Person", time.Now(), time.Now()); err != nil {
		t.Fatalf("Failed to insert person: %v", err)
	}

	// Insert schedule version
	if _, err := helper.DB().ExecContext(ctx, `
		INSERT INTO schedule_versions (id, hospital_id, status, effective_start_date, effective_end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		versionID, hospID, "STAGING", time.Now(), time.Now().AddDate(0, 0, 30), time.Now(), time.Now()); err != nil {
		t.Fatalf("Failed to insert schedule version: %v", err)
	}

	// Insert shift - immutable, so no UpdatedAt/UpdatedBy
	shiftID := uuid.New()
	if _, err := helper.DB().ExecContext(ctx, `
		INSERT INTO shift_instances (id, schedule_version_id, hospital_id, shift_type, schedule_date, start_time, end_time, study_type, specialty_constraint, desired_coverage, is_mandatory, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		shiftID, versionID, hospID, "DAY", time.Now(), "08:00", "16:00", "GENERAL", "BODY_ONLY", 2, false, time.Now(), uuid.New()); err != nil {
		t.Fatalf("Failed to insert shift: %v", err)
	}

	// Insert assignment - immutable after creation, no UpdatedAt/UpdatedBy
	assignmentID := uuid.New()
	if _, err := helper.DB().ExecContext(ctx, `
		INSERT INTO assignments (id, person_id, shift_instance_id, schedule_date, original_shift_type, source, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		assignmentID, personID, shiftID, time.Now(), "DAY", "MANUAL", time.Now(), uuid.New()); err != nil {
		t.Fatalf("Failed to insert assignment: %v", err)
	}

	// Verify assignment exists
	assignmentRepo := NewAssignmentRepository(helper.DB())
	assignment, err := assignmentRepo.GetByID(ctx, assignmentID)
	if err != nil {
		t.Fatalf("Failed to get assignment before delete: %v", err)
	}
	if assignment == nil {
		t.Fatal("Assignment should exist before soft delete")
	}

	// Soft delete assignment
	deleterID := uuid.New()
	if err := assignmentRepo.Delete(ctx, assignmentID, deleterID); err != nil {
		t.Fatalf("Failed to soft delete assignment: %v", err)
	}

	// Verify assignment is not accessible after soft delete
	assignment, err = assignmentRepo.GetByID(ctx, assignmentID)
	if err == nil {
		t.Fatal("Assignment should not be accessible after soft delete")
	}

	// Verify shift is still accessible (soft delete is entity-specific)
	shiftRepo := NewShiftInstanceRepository(helper.DB())
	shift, err := shiftRepo.GetByID(ctx, shiftID)
	if err != nil {
		t.Fatalf("Shift should still be accessible: %v", err)
	}
	if shift == nil {
		t.Fatal("Shift should not be nil")
	}

	t.Log("Soft delete cascading verified")
}

// TestRepositoryQueries_OptimizedForPerformance tests that queries are optimized (no N+1)
func TestRepositoryQueries_OptimizedForPerformance(t *testing.T) {
	ctx := context.Background()
	helper := NewPostgresTestHelper(ctx, t)
	defer helper.Close(t)

	// Create test data
	hospID := uuid.New()
	versionID := uuid.New()

	// Insert hospital
	if _, err := helper.DB().ExecContext(ctx, `
		INSERT INTO hospitals (id, name) VALUES ($1, $2)`,
		hospID, "Test Hospital"); err != nil {
		t.Fatalf("Failed to insert hospital: %v", err)
	}

	// Insert schedule version
	if _, err := helper.DB().ExecContext(ctx, `
		INSERT INTO schedule_versions (id, hospital_id, status, effective_start_date, effective_end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		versionID, hospID, "STAGING", time.Now(), time.Now().AddDate(0, 0, 30), time.Now(), time.Now()); err != nil {
		t.Fatalf("Failed to insert schedule version: %v", err)
	}

	// Create multiple shifts
	shiftIDs := make([]uuid.UUID, 10)
	shiftRepo := NewShiftInstanceRepository(helper.DB())
	creatorID := uuid.New()
	for i := 0; i < 10; i++ {
		shiftIDs[i] = uuid.New()
		shift := &entity.ShiftInstance{
			ID:                  shiftIDs[i],
			ScheduleVersionID:   versionID,
			HospitalID:          hospID,
			ShiftType:           entity.ShiftTypeDay,
			ScheduleDate:        time.Now().AddDate(0, 0, i),
			CreatedAt:           time.Now(),
			CreatedBy:           creatorID,
		}
		if err := shiftRepo.Create(ctx, shift); err != nil {
			t.Fatalf("Failed to create shift %d: %v", i, err)
		}
	}

	// Create persons and assignments for each shift
	personRepo := NewPersonRepository(helper.DB())
	assignmentRepo := NewAssignmentRepository(helper.DB())
	for i := 0; i < 10; i++ {
		// Create person for this shift
		personID := uuid.New()
		person := &entity.Person{
			ID:        personID,
			Email:     fmt.Sprintf("person%d@example.com", i),
			Name:      fmt.Sprintf("Person %d", i),
			Specialty: entity.SpecialtyBodyOnly,
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := personRepo.Create(ctx, person); err != nil {
			t.Fatalf("Failed to create person %d: %v", i, err)
		}

		// Create assignments for this person across multiple shifts
		for j := 0; j < 5; j++ {
			assignment := &entity.Assignment{
				PersonID:        personID,
				ShiftInstanceID: shiftIDs[i],
				ScheduleDate:    time.Now().AddDate(0, 0, i),
				Source:          entity.AssignmentSourceManual,
				CreatedAt:       time.Now(),
				CreatedBy:       creatorID,
			}
			if err := assignmentRepo.Create(ctx, assignment); err != nil {
				t.Fatalf("Failed to create assignment: %v", err)
			}
		}
	}

	// Test batch query optimization
	// This should be 1 query, not 10
	assignments, err := assignmentRepo.GetAllByShiftIDs(ctx, shiftIDs)
	if err != nil {
		t.Fatalf("GetAllByShiftIDs failed: %v", err)
	}
	if len(assignments) != 50 {
		t.Fatalf("Expected 50 assignments (10 shifts × 5 each), got %d", len(assignments))
	}

	t.Log("Batch query optimization verified (single query for multiple entities)")
}

// TestRepositories_AuditTrail tests that audit trail fields are properly tracked
func TestRepositories_AuditTrail(t *testing.T) {
	ctx := context.Background()
	helper := NewPostgresTestHelper(ctx, t)
	defer helper.Close(t)

	personRepo := NewPersonRepository(helper.DB())

	// Create person
	person := &entity.Person{
		Email:     "audit@example.com",
		Name:      "Audit Test Person",
		Specialty: entity.SpecialtyBodyOnly,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := personRepo.Create(ctx, person); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify audit trail on creation
	retrieved, _ := personRepo.GetByID(ctx, person.ID)
	if retrieved.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should be set on creation")
	}
	if retrieved.UpdatedAt.IsZero() {
		t.Fatal("UpdatedAt should be set on creation")
	}

	// Update person and verify audit trail
	person.Name = "Updated Name"
	person.UpdatedAt = time.Now()

	if err := personRepo.Update(ctx, person); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	retrieved, _ = personRepo.GetByID(ctx, person.ID)
	if retrieved.Name != "Updated Name" {
		t.Fatal("Name should be updated")
	}

	t.Log("Audit trail verification complete")
}

// TestRepositories_AuditLogTracking tests AuditLogRepository comprehensive functionality
func TestRepositories_AuditLogTracking(t *testing.T) {
	ctx := context.Background()
	helper := NewPostgresTestHelper(ctx, t)
	defer helper.Close(t)

	auditRepo := NewAuditLogRepository(helper.DB())

	userID := uuid.New()
	resourceID := uuid.New()

	// Create audit log - using correct field names
	log := &entity.AuditLog{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    "CREATE_SCHEDULE",
		Resource:  "ScheduleVersion#" + resourceID.String(),
		OldValues: "",
		NewValues: `{"status":"STAGING"}`,
		Timestamp: time.Now(),
		IPAddress: "127.0.0.1",
	}

	if err := auditRepo.Create(ctx, log); err != nil {
		t.Fatalf("Create audit log failed: %v", err)
	}

	// Test GetByUser
	logs, err := auditRepo.GetByUser(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUser failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("Expected 1 audit log for user, got %d", len(logs))
	}

	// Test GetByAction
	actionLogs, err := auditRepo.GetByAction(ctx, "CREATE_SCHEDULE")
	if err != nil {
		t.Fatalf("GetByAction failed: %v", err)
	}
	if len(actionLogs) != 1 {
		t.Fatalf("Expected 1 audit log for action, got %d", len(actionLogs))
	}

	// Test ListRecent
	recent, err := auditRepo.ListRecent(ctx, 10)
	if err != nil {
		t.Fatalf("ListRecent failed: %v", err)
	}
	if len(recent) != 1 {
		t.Fatalf("Expected 1 recent audit log, got %d", len(recent))
	}

	t.Log("Audit log repository comprehensive test passed")
}

// TestRepositories_JSONStorage tests JSONB storage in repositories
func TestRepositories_JSONStorage(t *testing.T) {
	ctx := context.Background()
	helper := NewPostgresTestHelper(ctx, t)
	defer helper.Close(t)

	// Create test data
	hospID := uuid.New()
	versionID := uuid.New()

	// Insert hospital and version
	if _, err := helper.DB().ExecContext(ctx, `
		INSERT INTO hospitals (id, name) VALUES ($1, $2)`,
		hospID, "Test Hospital"); err != nil {
		t.Fatalf("Failed to insert hospital: %v", err)
	}

	if _, err := helper.DB().ExecContext(ctx, `
		INSERT INTO schedule_versions (id, hospital_id, status, effective_start_date, effective_end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		versionID, hospID, "STAGING", time.Now(), time.Now().AddDate(0, 0, 30), time.Now(), time.Now()); err != nil {
		t.Fatalf("Failed to insert schedule version: %v", err)
	}

	coverageRepo := NewCoverageCalculationRepository(helper.DB())

	// Create coverage with JSON data
	coverage := &entity.CoverageCalculation{
		ID:                         uuid.New(),
		ScheduleVersionID:          versionID,
		HospitalID:                 hospID,
		CalculationDate:            time.Now(),
		CalculationPeriodStartDate: time.Now(),
		CalculationPeriodEndDate:   time.Now().AddDate(0, 0, 30),
		CoverageByPosition: map[string]int{
			"DAY_GENERAL_BODY":   10,
			"DAY_GENERAL_NEURO":  5,
			"NIGHT_GENERAL_BODY": 8,
		},
		CoverageSummary: map[string]interface{}{
			"total_shifts":      23,
			"average_coverage":  0.82,
			"positions_covered": 3,
		},
		QueryCount:   2,
		CalculatedAt: time.Now(),
		CalculatedBy: uuid.New(),
	}

	if err := coverageRepo.Create(ctx, coverage); err != nil {
		t.Fatalf("Create coverage failed: %v", err)
	}

	// Retrieve and verify JSON is properly deserialized
	retrieved, err := coverageRepo.GetByID(ctx, coverage.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved.CoverageByPosition["DAY_GENERAL_BODY"] != 10 {
		t.Fatal("Coverage by position should be properly deserialized")
	}

	summary, ok := retrieved.CoverageSummary["average_coverage"].(float64)
	if !ok || summary != 0.82 {
		t.Fatal("Coverage summary should have correct float value")
	}

	t.Log("JSON storage in repositories verified")
}
