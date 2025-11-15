package orchestrator

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// LoadTestMetrics captures performance and consistency metrics from load tests.
type LoadTestMetrics struct {
	// Throughput metrics
	TotalWorkflows       int64
	SuccessfulWorkflows  int64
	FailedWorkflows      int64
	WorkflowsPerSecond   float64
	TotalDuration        time.Duration

	// Latency metrics
	MinLatency    time.Duration
	MaxLatency    time.Duration
	AvgLatency    time.Duration
	P50Latency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration

	// Consistency metrics
	DataInconsistencies int
	HospitalDataCounts  map[string]int // hospital_id -> count of workflows completed

	// Concurrency metrics
	GoroutinesBefore int
	GoroutinesAfter  int
	GoroutineLeaks   int

	// Deadlock detection
	DeadlockDetected bool
	StuckGoroutines  int
}

// TestLoadSimulationConcurrentWorkflows tests concurrent workflow execution with 5 simultaneous orchestrations.
// This test verifies:
// 1. All 5 concurrent workflows complete successfully
// 2. Each workflow executes against different hospital schedules
// 3. No goroutine leaks occur
// 4. No deadlocks are detected
// 5. Data is isolated per hospital with no cross-contamination
// 6. Throughput and latency metrics are captured
func TestLoadSimulationConcurrentWorkflows(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	// Record baseline goroutines
	runtime.GC()
	goroutinesBefore := runtime.NumGoroutine()

	// Create 5 independent orchestrators with different hospitals
	const concurrentWorkflows = 5
	hospitalIDs := make([]uuid.UUID, concurrentWorkflows)
	orchestrators := make([]*DefaultScheduleOrchestrator, concurrentWorkflows)
	userID := uuid.New()

	for i := 0; i < concurrentWorkflows; i++ {
		hospitalIDs[i] = uuid.New()

		// Create mock services for each hospital
		mockODS := &MockODSImportService{
			ImportScheduleFunc: func(ctx context.Context, filePath string, hID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
				// Simulate some processing time
				time.Sleep(10 * time.Millisecond)
				return &entity.ScheduleVersion{
					ID:         uuid.New(),
					HospitalID: hID,
					Version:    1,
					Status:     entity.VersionStatusDraft,
					StartDate:  time.Now(),
					EndDate:    time.Now().AddDate(0, 1, 0),
					Source:     "ods_file",
					CreatedBy:  userID,
					CreatedAt:  time.Now(),
				}, validation.NewValidationResult(), nil
			},
		}

		mockAmion := &MockAmionScraperService{
			ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
				// Simulate scraping time
				time.Sleep(5 * time.Millisecond)
				return []entity.Assignment{}, validation.NewValidationResult(), nil
			},
		}

		mockCoverage := &MockCoverageCalculatorService{
			CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
				// Simulate calculation time
				time.Sleep(3 * time.Millisecond)
				return &CoverageMetrics{
					ScheduleVersionID:  scheduleVersionID,
					CoveragePercentage: 95.0,
					AssignedPositions:  95,
					RequiredPositions:  100,
					CalculatedAt:       time.Now(),
				}, nil
			},
		}

		orchestrators[i] = NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
	}

	// Execute concurrent workflows
	var wg sync.WaitGroup
	results := make([]*OrchestrationResult, concurrentWorkflows)
	errors := make([]error, concurrentWorkflows)
	latencies := make([]time.Duration, concurrentWorkflows)
	var successCount, failureCount int64

	startTime := time.Now()

	for i := 0; i < concurrentWorkflows; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workflowStart := time.Now()
			filePath := fmt.Sprintf("/path/to/hospital_%s.ods", hospitalIDs[idx].String()[:8])

			result, err := orchestrators[idx].ExecuteImport(ctx, filePath, hospitalIDs[idx], userID)

			latencies[idx] = time.Since(workflowStart)
			results[idx] = result
			errors[idx] = err

			if err != nil {
				atomic.AddInt64(&failureCount, 1)
			} else {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	// Wait for all goroutines to complete with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All workflows completed
	case <-time.After(30 * time.Second):
		t.Fatal("Load test timeout: workflows did not complete within 30 seconds (possible deadlock)")
	}

	totalDuration := time.Since(startTime)

	// Verify all workflows completed successfully
	for i := 0; i < concurrentWorkflows; i++ {
		assert.NoError(t, errors[i], "workflow %d failed", i)
		assert.NotNil(t, results[i], "workflow %d returned nil result", i)
		if results[i] != nil {
			assert.NotNil(t, results[i].ScheduleVersion, "workflow %d missing schedule version", i)
			assert.Equal(t, hospitalIDs[i], results[i].ScheduleVersion.HospitalID, "workflow %d hospital ID mismatch", i)
			assert.False(t, results[i].ValidationResult.HasErrors(), "workflow %d validation errors", i)
		}
	}

	// Verify success counts
	assert.Equal(t, int64(concurrentWorkflows), successCount, "not all workflows succeeded")
	assert.Equal(t, int64(0), failureCount, "unexpected workflow failures")

	// Calculate throughput
	workflowsPerSecond := float64(successCount) / totalDuration.Seconds()

	// Calculate latency statistics
	minLatency := latencies[0]
	maxLatency := latencies[0]
	totalLatency := time.Duration(0)

	for _, lat := range latencies {
		if lat < minLatency {
			minLatency = lat
		}
		if lat > maxLatency {
			maxLatency = lat
		}
		totalLatency += lat
	}

	avgLatency := time.Duration(int64(totalLatency) / int64(concurrentWorkflows))

	// Log metrics
	t.Logf("Load Test Results:")
	t.Logf("  Total Workflows: %d", successCount)
	t.Logf("  Total Duration: %v", totalDuration)
	t.Logf("  Throughput: %.2f workflows/sec", workflowsPerSecond)
	t.Logf("  Min Latency: %v", minLatency)
	t.Logf("  Max Latency: %v", maxLatency)
	t.Logf("  Avg Latency: %v", avgLatency)

	// Verify data isolation - each hospital should have exactly one schedule version
	hospitalVersionCounts := make(map[uuid.UUID]int)
	for i := 0; i < concurrentWorkflows; i++ {
		if results[i] != nil && results[i].ScheduleVersion != nil {
			hospitalVersionCounts[results[i].ScheduleVersion.HospitalID]++
		}
	}

	// Verify data consistency - no cross-hospital contamination
	for i := 0; i < concurrentWorkflows; i++ {
		if results[i] != nil && results[i].ScheduleVersion != nil {
			expectedCount := 1
			actualCount := hospitalVersionCounts[hospitalIDs[i]]
			assert.Equal(t, expectedCount, actualCount, "hospital %d has unexpected schedule count", i)
		}
	}

	// Check for goroutine leaks
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // Allow goroutines to clean up
	gorutinesAfter := runtime.NumGoroutine()

	// Account for test infrastructure goroutines
	goroutineLeakThreshold := 10 // Allow some variance for test framework
	actualLeak := gorutinesAfter - goroutinesBefore

	t.Logf("Goroutine Analysis:")
	t.Logf("  Before: %d", goroutinesBefore)
	t.Logf("  After: %d", gorutinesAfter)
	t.Logf("  Leak (actual): %d", actualLeak)

	// Assert no significant goroutine leaks
	assert.Less(t, actualLeak, goroutineLeakThreshold,
		"possible goroutine leak detected: %d goroutines not cleaned up", actualLeak)

	// Verify status transitions
	for i := 0; i < concurrentWorkflows; i++ {
		status := orchestrators[i].GetOrchestrationStatus()
		assert.Equal(t, OrchestrationStatusCOMPLETED, status, "orchestrator %d status not completed", i)
	}

	// Verify assertions on throughput (basic sanity check)
	// With mocked services (18ms total per workflow), we should achieve:
	// ~55 workflows/second in ideal conditions
	assert.Greater(t, workflowsPerSecond, 0.0, "throughput should be positive")
	assert.Less(t, avgLatency, 100*time.Millisecond, "average latency too high")
}

// TestLoadSimulationHighThroughput runs 50 concurrent workflows to test sustained throughput.
// This test measures maximum throughput under higher concurrency.
func TestLoadSimulationHighThroughput(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	const concurrentWorkflows = 50
	hospitalIDs := make([]uuid.UUID, concurrentWorkflows)
	orchestrators := make([]*DefaultScheduleOrchestrator, concurrentWorkflows)
	userID := uuid.New()

	// Create orchestrators
	for i := 0; i < concurrentWorkflows; i++ {
		hospitalIDs[i] = uuid.New()

		mockODS := &MockODSImportService{
			ImportScheduleFunc: func(ctx context.Context, filePath string, hID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
				time.Sleep(5 * time.Millisecond)
				return &entity.ScheduleVersion{
					ID:         uuid.New(),
					HospitalID: hID,
					Version:    1,
					Status:     entity.VersionStatusDraft,
					StartDate:  time.Now(),
					EndDate:    time.Now().AddDate(0, 1, 0),
					Source:     "ods_file",
					CreatedBy:  userID,
					CreatedAt:  time.Now(),
				}, validation.NewValidationResult(), nil
			},
		}

		mockAmion := &MockAmionScraperService{
			ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
				time.Sleep(2 * time.Millisecond)
				return []entity.Assignment{}, validation.NewValidationResult(), nil
			},
		}

		mockCoverage := &MockCoverageCalculatorService{
			CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
				time.Sleep(1 * time.Millisecond)
				return &CoverageMetrics{
					ScheduleVersionID:  scheduleVersionID,
					CoveragePercentage: 95.0,
					AssignedPositions:  95,
					RequiredPositions:  100,
					CalculatedAt:       time.Now(),
				}, nil
			},
		}

		orchestrators[i] = NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
	}

	// Execute concurrent workflows
	var wg sync.WaitGroup
	var successCount, failureCount int64
	latencies := make(chan time.Duration, concurrentWorkflows)

	startTime := time.Now()

	for i := 0; i < concurrentWorkflows; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workflowStart := time.Now()
			filePath := fmt.Sprintf("/path/to/hospital_%s.ods", hospitalIDs[idx].String()[:8])

			_, err := orchestrators[idx].ExecuteImport(ctx, filePath, hospitalIDs[idx], userID)

			latencies <- time.Since(workflowStart)

			if err != nil {
				atomic.AddInt64(&failureCount, 1)
			} else {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	// Wait for completion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All workflows completed
	case <-time.After(60 * time.Second):
		t.Fatal("High throughput test timeout: workflows did not complete within 60 seconds")
	}

	totalDuration := time.Since(startTime)
	close(latencies)

	// Collect latencies for analysis
	var latencySlice []time.Duration
	for lat := range latencies {
		latencySlice = append(latencySlice, lat)
	}

	// Calculate statistics
	workflowsPerSecond := float64(successCount) / totalDuration.Seconds()

	minLatency := 1000000 * time.Millisecond // Large value for comparison
	maxLatency := time.Duration(0)
	totalLatency := time.Duration(0)

	for _, lat := range latencySlice {
		if lat < minLatency {
			minLatency = lat
		}
		if lat > maxLatency {
			maxLatency = lat
		}
		totalLatency += lat
	}

	avgLatency := time.Duration(int64(totalLatency) / int64(len(latencySlice)))

	t.Logf("High Throughput Load Test Results:")
	t.Logf("  Total Workflows: %d", successCount)
	t.Logf("  Failed Workflows: %d", failureCount)
	t.Logf("  Total Duration: %v", totalDuration)
	t.Logf("  Throughput: %.2f workflows/sec", workflowsPerSecond)
	t.Logf("  Min Latency: %v", minLatency)
	t.Logf("  Max Latency: %v", maxLatency)
	t.Logf("  Avg Latency: %v", avgLatency)

	// Assertions
	assert.Equal(t, int64(concurrentWorkflows), successCount, "not all workflows succeeded in high throughput test")
	assert.Equal(t, int64(0), failureCount, "unexpected failures in high throughput test")
	assert.Greater(t, workflowsPerSecond, 0.0, "throughput should be positive")
}

// TestLoadSimulationDataConsistency verifies that concurrent workflows maintain data isolation.
// Each hospital's data should remain isolated with no cross-hospital contamination.
func TestLoadSimulationDataConsistency(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	const concurrentWorkflows = 5
	hospitalIDs := make([]uuid.UUID, concurrentWorkflows)
	_ = make([]uuid.UUID, concurrentWorkflows) // scheduleVersionIDs reserved for future use
	orchestrators := make([]*DefaultScheduleOrchestrator, concurrentWorkflows)
	userID := uuid.New()

	mu := sync.Mutex{}
	idMapping := make(map[uuid.UUID]uuid.UUID) // hospital -> schedule version

	for i := 0; i < concurrentWorkflows; i++ {
		hospitalIDs[i] = uuid.New()

		mockODS := &MockODSImportService{
			ImportScheduleFunc: func(ctx context.Context, filePath string, hID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
				time.Sleep(10 * time.Millisecond)

				sv := &entity.ScheduleVersion{
					ID:         uuid.New(),
					HospitalID: hID,
					Version:    1,
					Status:     entity.VersionStatusDraft,
					StartDate:  time.Now(),
					EndDate:    time.Now().AddDate(0, 1, 0),
					Source:     "ods_file",
					CreatedBy:  userID,
					CreatedAt:  time.Now(),
				}

				// Track the mapping
				mu.Lock()
				idMapping[hID] = sv.ID
				mu.Unlock()

				return sv, validation.NewValidationResult(), nil
			},
		}

		mockAmion := &MockAmionScraperService{
			ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
				time.Sleep(5 * time.Millisecond)
				return []entity.Assignment{}, validation.NewValidationResult(), nil
			},
		}

		mockCoverage := &MockCoverageCalculatorService{
			CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
				time.Sleep(3 * time.Millisecond)
				return &CoverageMetrics{
					ScheduleVersionID:  scheduleVersionID,
					CoveragePercentage: 95.0,
					AssignedPositions:  95,
					RequiredPositions:  100,
					CalculatedAt:       time.Now(),
				}, nil
			},
		}

		orchestrators[i] = NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
	}

	// Execute concurrent workflows
	var wg sync.WaitGroup
	results := make([]*OrchestrationResult, concurrentWorkflows)

	for i := 0; i < concurrentWorkflows; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			filePath := fmt.Sprintf("/path/to/hospital_%s.ods", hospitalIDs[idx].String()[:8])
			result, err := orchestrators[idx].ExecuteImport(ctx, filePath, hospitalIDs[idx], userID)

			assert.NoError(t, err, "workflow %d failed", idx)
			results[idx] = result
		}(i)
	}

	// Wait for completion
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All workflows completed
	case <-time.After(30 * time.Second):
		t.Fatal("Data consistency test timeout")
	}

	// Verify data isolation
	for i := 0; i < concurrentWorkflows; i++ {
		require.NotNil(t, results[i], "result %d is nil", i)
		require.NotNil(t, results[i].ScheduleVersion, "schedule version %d is nil", i)

		// Verify correct hospital ID
		assert.Equal(t, hospitalIDs[i], results[i].ScheduleVersion.HospitalID, "hospital ID mismatch for workflow %d", i)

		// Verify no cross-contamination - each hospital should only have its own schedule
		for j := 0; j < concurrentWorkflows; j++ {
			if i != j {
				assert.NotEqual(t, results[i].ScheduleVersion.HospitalID, results[j].ScheduleVersion.HospitalID,
					"cross-hospital contamination detected: workflows %d and %d have same hospital", i, j)
			}
		}
	}

	// Verify schedule version mapping
	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, concurrentWorkflows, len(idMapping), "not all hospitals were mapped")

	for i := 0; i < concurrentWorkflows; i++ {
		mappedID, exists := idMapping[hospitalIDs[i]]
		assert.True(t, exists, "hospital %d not found in mapping", i)
		assert.Equal(t, results[i].ScheduleVersion.ID, mappedID, "schedule version ID mismatch for hospital %d", i)
	}

	t.Logf("Data Consistency Verification:")
	t.Logf("  Hospitals processed: %d", len(idMapping))
	t.Logf("  No cross-contamination detected")
	t.Logf("  All schedule versions correctly mapped to hospitals")
}

// TestLoadSimulationNoDeadlocks verifies that concurrent operations do not deadlock.
// This test creates a stress scenario with concurrent orchestrations and validates
// that all goroutines eventually complete.
func TestLoadSimulationNoDeadlocks(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	const concurrentWorkflows = 10
	hospitalIDs := make([]uuid.UUID, concurrentWorkflows)
	orchestrators := make([]*DefaultScheduleOrchestrator, concurrentWorkflows)
	userID := uuid.New()

	for i := 0; i < concurrentWorkflows; i++ {
		hospitalIDs[i] = uuid.New()

		mockODS := &MockODSImportService{
			ImportScheduleFunc: func(ctx context.Context, filePath string, hID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
				time.Sleep(8 * time.Millisecond)
				return &entity.ScheduleVersion{
					ID:         uuid.New(),
					HospitalID: hID,
					Version:    1,
					Status:     entity.VersionStatusDraft,
					StartDate:  time.Now(),
					EndDate:    time.Now().AddDate(0, 1, 0),
					Source:     "ods_file",
					CreatedBy:  userID,
					CreatedAt:  time.Now(),
				}, validation.NewValidationResult(), nil
			},
		}

		mockAmion := &MockAmionScraperService{
			ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
				time.Sleep(4 * time.Millisecond)
				return []entity.Assignment{}, validation.NewValidationResult(), nil
			},
		}

		mockCoverage := &MockCoverageCalculatorService{
			CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
				time.Sleep(2 * time.Millisecond)
				return &CoverageMetrics{
					ScheduleVersionID:  scheduleVersionID,
					CoveragePercentage: 95.0,
					AssignedPositions:  95,
					RequiredPositions:  100,
					CalculatedAt:       time.Now(),
				}, nil
			},
		}

		orchestrators[i] = NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
	}

	var wg sync.WaitGroup
	completionChan := make(chan int, concurrentWorkflows)
	var successCount int64

	// Execute workflows
	for i := 0; i < concurrentWorkflows; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			filePath := fmt.Sprintf("/path/to/hospital_%s.ods", hospitalIDs[idx].String()[:8])
			_, err := orchestrators[idx].ExecuteImport(ctx, filePath, hospitalIDs[idx], userID)

			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}

			completionChan <- idx
		}(i)
	}

	// Wait with timeout to detect deadlocks
	go wg.Wait()

	completed := 0
	deadlockDetectionTimer := time.NewTimer(15 * time.Second)
	defer deadlockDetectionTimer.Stop()

	for completed < concurrentWorkflows {
		select {
		case <-completionChan:
			completed++
		case <-deadlockDetectionTimer.C:
			t.Fatalf("Deadlock detected: only %d of %d workflows completed within timeout", completed, concurrentWorkflows)
		}
	}

	// Verify all completed successfully
	assert.Equal(t, int64(concurrentWorkflows), successCount, "not all workflows completed successfully")

	// Verify status of all orchestrators
	for i := 0; i < concurrentWorkflows; i++ {
		status := orchestrators[i].GetOrchestrationStatus()
		assert.Equal(t, OrchestrationStatusCOMPLETED, status, "orchestrator %d not in completed state", i)
	}

	t.Logf("Deadlock Detection Test:")
	t.Logf("  All %d workflows completed without deadlock", concurrentWorkflows)
	t.Logf("  All orchestrators transitioned to COMPLETED status")
}

// BenchmarkOrchestrationThroughput benchmarks the maximum throughput of orchestrations.
// This benchmark runs orchestrations sequentially to measure per-operation latency.
func BenchmarkOrchestrationThroughput(b *testing.B) {
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hospitalID := uuid.New()
		userID := uuid.New()

		mockODS := &MockODSImportService{
			ImportScheduleFunc: func(ctx context.Context, filePath string, hID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
				time.Sleep(10 * time.Millisecond)
				return &entity.ScheduleVersion{
					ID:         uuid.New(),
					HospitalID: hID,
					Version:    1,
					Status:     entity.VersionStatusDraft,
					StartDate:  time.Now(),
					EndDate:    time.Now().AddDate(0, 1, 0),
					Source:     "ods_file",
					CreatedBy:  userID,
					CreatedAt:  time.Now(),
				}, validation.NewValidationResult(), nil
			},
		}

		mockAmion := &MockAmionScraperService{
			ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
				time.Sleep(5 * time.Millisecond)
				return []entity.Assignment{}, validation.NewValidationResult(), nil
			},
		}

		mockCoverage := &MockCoverageCalculatorService{
			CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
				time.Sleep(3 * time.Millisecond)
				return &CoverageMetrics{
					ScheduleVersionID:  scheduleVersionID,
					CoveragePercentage: 95.0,
					AssignedPositions:  95,
					RequiredPositions:  100,
					CalculatedAt:       time.Now(),
				}, nil
			},
		}

		orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)
		_, _ = orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)
	}
}
