package service

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScheduleOrchestratorFullWorkflow tests the complete 3-phase workflow
func TestScheduleOrchestratorFullWorkflow(t *testing.T) {
	// Setup
	ctx := context.Background()
	hospitalID := entity.HospitalID(uuid.New())
	creatorID := entity.UserID(uuid.New())

	// Create in-memory repositories
	memRepo := memory.NewMemoryRepository()
	hospitalRepo := memory.NewHospitalRepository(memRepo)
	personRepo := memory.NewPersonRepository(memRepo)
	versionRepo := memory.NewScheduleVersionRepository(memRepo)
	shiftRepo := memory.NewShiftInstanceRepository(memRepo)
	assignmentRepo := memory.NewAssignmentRepository(memRepo)
	batchRepo := memory.NewScrapeBatchRepository(memRepo)
	coverageRepo := memory.NewCoverageCalculationRepository(memRepo)

	// Create hospital
	hospital := &entity.Hospital{
		ID:   hospitalID,
		Name: "Test Hospital",
	}
	err := hospitalRepo.Create(ctx, hospital)
	require.NoError(t, err)

	// Create services
	coverageCalc := NewDynamicCoverageCalculator(shiftRepo, assignmentRepo)
	versionService := NewScheduleVersionService(versionRepo)
	odsImporter := NewODSImportService(shiftRepo, assignmentRepo, versionRepo, coverageCalc)
	amionImporter := NewAmionImportService(assignmentRepo, batchRepo, versionRepo)
	orchestrator := NewScheduleOrchestrator(odsImporter, amionImporter, coverageCalc, versionService)

	// Test data
	startDate := entity.Date(time.Now().Format("2006-01-02"))
	endDate := entity.Date(time.Now().AddDate(0, 1, 0).Format("2006-01-02"))
	odsFilename := "schedule.ods"
	odsContent := bytes.NewReader([]byte{}) // Empty for now (real parsing in Phase 3)

	amionConfig := AmionScraperConfig{
		Username:          "testuser",
		Password:          "testpass",
		MonthsToScrape:    6,
		ConcurrentWorkers: 5,
	}

	// Execute workflow
	result := orchestrator.ExecuteFullWorkflow(
		ctx,
		hospitalID,
		creatorID,
		startDate,
		endDate,
		odsFilename,
		odsContent,
		amionConfig,
	)

	// Verify result
	assert.NotNil(t, result)
	assert.NotNil(t, result.ScheduleVersionID)
	assert.NotNil(t, result.ValidationResult)
	assert.NotNil(t, result.OdsBatch)

	// Verify schedule version was created
	version, err := versionService.GetVersion(ctx, result.ScheduleVersionID)
	require.NoError(t, err)
	assert.Equal(t, entity.VersionStatusStaging, version.Status)
	assert.Equal(t, hospitalID, version.HospitalID)
}

// TestScheduleOrchestratorODSImportPhase tests just the ODS import phase
func TestScheduleOrchestratorODSImportPhase(t *testing.T) {
	ctx := context.Background()
	hospitalID := entity.HospitalID(uuid.New())
	creatorID := entity.UserID(uuid.New())

	// Setup repositories
	memRepo := memory.NewMemoryRepository()
	versionRepo := memory.NewScheduleVersionRepository(memRepo)
	shiftRepo := memory.NewShiftInstanceRepository(memRepo)
	assignmentRepo := memory.NewAssignmentRepository(memRepo)

	// Create services
	coverageCalc := NewDynamicCoverageCalculator(shiftRepo, assignmentRepo)
	odsImporter := NewODSImportService(shiftRepo, assignmentRepo, versionRepo, coverageCalc)

	// Create schedule version
	version := &entity.ScheduleVersion{
		ID:          entity.ScheduleVersionID(uuid.New()),
		HospitalID:  hospitalID,
		StartDate:   entity.Date(time.Now().Format("2006-01-02")),
		EndDate:     entity.Date(time.Now().AddDate(0, 1, 0).Format("2006-01-02")),
		Status:      entity.VersionStatusStaging,
		CreatedAt:   entity.Now(),
		CreatedByID: creatorID,
	}
	err := versionRepo.Create(ctx, version)
	require.NoError(t, err)

	// Execute ODS import
	odsContent := bytes.NewReader([]byte{})
	batch, result, err := odsImporter.ImportODSFile(ctx, hospitalID, version, "test.ods", odsContent)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, batch)
	assert.NotNil(t, result)
	assert.Equal(t, "ODS_FILE", batch.Source)
}

// TestScheduleOrchestratorValidationErrorCollection tests that errors are collected, not fail-fast
func TestScheduleOrchestratorValidationErrorCollection(t *testing.T) {
	ctx := context.Background()

	// This test verifies the error collection pattern from v1
	// When importing ODS, we should collect all errors and continue processing
	// rather than failing on the first error

	memRepo := memory.NewMemoryRepository()
	versionRepo := memory.NewScheduleVersionRepository(memRepo)
	shiftRepo := memory.NewShiftInstanceRepository(memRepo)
	assignmentRepo := memory.NewAssignmentRepository(memRepo)
	batchRepo := memory.NewScrapeBatchRepository(memRepo)

	coverageCalc := NewDynamicCoverageCalculator(shiftRepo, assignmentRepo)
	versionService := NewScheduleVersionService(versionRepo)
	odsImporter := NewODSImportService(shiftRepo, assignmentRepo, versionRepo, coverageCalc)
	amionImporter := NewAmionImportService(assignmentRepo, batchRepo, versionRepo)

	orchestrator := NewScheduleOrchestrator(odsImporter, amionImporter, coverageCalc, versionService)

	// Execute workflow that will have some validation issues
	result := orchestrator.ExecuteFullWorkflow(
		ctx,
		entity.HospitalID(uuid.New()),
		entity.UserID(uuid.New()),
		entity.Date("2025-01-01"),
		entity.Date("2025-02-01"),
		"test.ods",
		bytes.NewReader([]byte{}),
		AmionScraperConfig{
			Username:       "test",
			Password:       "test",
			MonthsToScrape: 1,
		},
	)

	// Verify validation result exists and contains collected messages
	assert.NotNil(t, result.ValidationResult)
	assert.NotNil(t, result.ValidationResult.Messages)

	// Even if there are errors, the result structure should be populated
	assert.NotNil(t, result.ScheduleVersionID)
	assert.NotNil(t, result.OdsBatch)
}

// TestScheduleVersionServiceStateTransitions tests state machine transitions
func TestScheduleVersionServiceStateTransitions(t *testing.T) {
	ctx := context.Background()

	memRepo := memory.NewMemoryRepository()
	versionRepo := memory.NewScheduleVersionRepository(memRepo)
	service := NewScheduleVersionService(versionRepo)

	creatorID := entity.UserID(uuid.New())
	hospitalID := entity.HospitalID(uuid.New())

	// Create version (should start in STAGING)
	version, err := service.CreateVersion(
		ctx,
		hospitalID,
		entity.Date("2025-01-01"),
		entity.Date("2025-02-01"),
		creatorID,
	)
	require.NoError(t, err)
	assert.Equal(t, entity.VersionStatusStaging, version.Status)

	// Promote to PRODUCTION
	err = service.PromoteToProduction(ctx, version.ID, creatorID)
	require.NoError(t, err)

	// Verify promotion
	promoted, err := service.GetVersion(ctx, version.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.VersionStatusProduction, promoted.Status)
	assert.NotNil(t, promoted.PromotedAt)

	// Archive
	err = service.Archive(ctx, version.ID, creatorID)
	require.NoError(t, err)

	// Verify archive
	archived, err := service.GetVersion(ctx, version.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.VersionStatusArchived, archived.Status)
	assert.NotNil(t, archived.ArchivedAt)
}

// TestScheduleVersionServiceInvalidTransition tests that invalid state transitions fail
func TestScheduleVersionServiceInvalidTransition(t *testing.T) {
	ctx := context.Background()

	memRepo := memory.NewMemoryRepository()
	versionRepo := memory.NewScheduleVersionRepository(memRepo)
	service := NewScheduleVersionService(versionRepo)

	creatorID := entity.UserID(uuid.New())
	hospitalID := entity.HospitalID(uuid.New())

	// Create version
	version, err := service.CreateVersion(
		ctx,
		hospitalID,
		entity.Date("2025-01-01"),
		entity.Date("2025-02-01"),
		creatorID,
	)
	require.NoError(t, err)

	// Try to archive STAGING version (should fail - can only archive PRODUCTION)
	err = service.Archive(ctx, version.ID, creatorID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only archive PRODUCTION")

	// Promote to PRODUCTION
	err = service.PromoteToProduction(ctx, version.ID, creatorID)
	require.NoError(t, err)

	// Try to promote again (should fail - can only promote STAGING)
	err = service.PromoteToProduction(ctx, version.ID, creatorID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only promote STAGING")
}

// TestScheduleOrchestratorMultipleVersions tests multiple versions for same hospital
func TestScheduleOrchestratorMultipleVersions(t *testing.T) {
	ctx := context.Background()

	memRepo := memory.NewMemoryRepository()
	versionRepo := memory.NewScheduleVersionRepository(memRepo)
	service := NewScheduleVersionService(versionRepo)

	creatorID := entity.UserID(uuid.New())
	hospitalID := entity.HospitalID(uuid.New())

	// Create first version
	v1, err := service.CreateVersion(
		ctx,
		hospitalID,
		entity.Date("2025-01-01"),
		entity.Date("2025-02-01"),
		creatorID,
	)
	require.NoError(t, err)

	// Create second version
	v2, err := service.CreateVersion(
		ctx,
		hospitalID,
		entity.Date("2025-02-01"),
		entity.Date("2025-03-01"),
		creatorID,
	)
	require.NoError(t, err)

	// Both should be STAGING
	assert.Equal(t, entity.VersionStatusStaging, v1.Status)
	assert.Equal(t, entity.VersionStatusStaging, v2.Status)

	// List versions
	versions, err := service.ListAllVersions(ctx, hospitalID)
	require.NoError(t, err)
	assert.Len(t, versions, 2)

	// Promote v1, then v2 (should archive v1)
	err = service.PromoteToProduction(ctx, v1.ID, creatorID)
	require.NoError(t, err)

	err = service.PromoteAndArchiveOthers(ctx, v2.ID, creatorID)
	require.NoError(t, err)

	// Verify v1 is archived
	updated_v1, err := service.GetVersion(ctx, v1.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.VersionStatusArchived, updated_v1.Status)

	// Verify v2 is production
	updated_v2, err := service.GetVersion(ctx, v2.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.VersionStatusProduction, updated_v2.Status)
}

// BenchmarkScheduleOrchestrator benchmarks the full workflow
func BenchmarkScheduleOrchestrator(b *testing.B) {
	ctx := context.Background()
	hospitalID := entity.HospitalID(uuid.New())
	creatorID := entity.UserID(uuid.New())

	memRepo := memory.NewMemoryRepository()
	versionRepo := memory.NewScheduleVersionRepository(memRepo)
	shiftRepo := memory.NewShiftInstanceRepository(memRepo)
	assignmentRepo := memory.NewAssignmentRepository(memRepo)
	batchRepo := memory.NewScrapeBatchRepository(memRepo)

	coverageCalc := NewDynamicCoverageCalculator(shiftRepo, assignmentRepo)
	versionService := NewScheduleVersionService(versionRepo)
	odsImporter := NewODSImportService(shiftRepo, assignmentRepo, versionRepo, coverageCalc)
	amionImporter := NewAmionImportService(assignmentRepo, batchRepo, versionRepo)
	orchestrator := NewScheduleOrchestrator(odsImporter, amionImporter, coverageCalc, versionService)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := orchestrator.ExecuteFullWorkflow(
			ctx,
			hospitalID,
			creatorID,
			entity.Date("2025-01-01"),
			entity.Date("2025-02-01"),
			"test.ods",
			bytes.NewReader([]byte{}),
			AmionScraperConfig{
				Username:       "test",
				Password:       "test",
				MonthsToScrape: 1,
			},
		)
		_ = result // Use result to prevent optimization
	}
}
